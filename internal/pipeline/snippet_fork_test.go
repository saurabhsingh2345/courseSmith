package pipeline

import (
	"strings"
	"testing"
)

const fkNarration = "The parent writes one page, gets its own copy, and everything else is still one page."

func forkPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "fork",
		Title:    "A snapshot that never pauses the database",
		Fork: &ForkSpec{
			Origin:     "redis-server",
			OriginNote: "pid 4271",
			Pages: []ForkPage{
				{Label: "user:42"}, {Label: "cart:42"}, {Label: "session:7"},
				{Label: "score:42"}, {Label: "keys G-M"}, {Label: "keys N-Z"},
			},
			Writes: []ForkWrite{
				{At: 1, By: "parent", Note: "The child still sees the old value", Role: "quantity"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "shared", Heading: "Two processes", Narration: fkNarration, Fork: &ForkBeat{Show: "shared"}},
			{ID: "write", Heading: "Someone writes", Narration: fkNarration, Fork: &ForkBeat{Show: "write", At: 0}},
			{ID: "cost", Heading: "What it cost", Narration: fkNarration, Fork: &ForkBeat{Show: "read"}},
		},
	}
	p.targetWords = 3 * 40
	return p
}

func TestForkPlanAccepted(t *testing.T) {
	if err := validateForkPlan(forkPlan()); err != nil {
		t.Fatalf("a well-formed fork plan was rejected: %v", err)
	}
}

// The rule the template exists for. The saving IS the pages nobody wrote to, so
// a clip that copies all of them has drawn a full duplicate and argued against
// its own subject.
func TestForkRejectsCopyingEveryPage(t *testing.T) {
	p := forkPlan()
	p.Fork.Pages = p.Fork.Pages[:2]
	p.Fork.Writes = []ForkWrite{
		{At: 0, By: "parent"},
		{At: 1, By: "child"},
	}
	p.Beats = append(p.Beats, SnippetBeat{
		ID: "second", Heading: "And again", Narration: fkNarration,
		Fork: &ForkBeat{Show: "write", At: 1},
	})
	err := validateForkPlan(p)
	if err == nil {
		t.Fatal("a clip that copied every page was accepted")
	}
	// Two pages also trips the floor, so assert on the message that matters.
	if !strings.Contains(err.Error(), "pages") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// The same rule with a legal page count, so nothing else can be doing the
// rejecting.
func TestForkRejectsCopyingEveryPageAtFullSize(t *testing.T) {
	p := forkPlan()
	p.Fork.Pages = p.Fork.Pages[:4]
	p.Fork.Writes = []ForkWrite{
		{At: 0, By: "parent"}, {At: 1, By: "child"}, {At: 2, By: "parent"},
	}
	// A fourth write would exceed the cap, so shrink the memory to three pages
	// worth of writes by dropping one page.
	p.Fork.Pages = p.Fork.Pages[:3]
	p.Beats = []SnippetBeat{
		{ID: "shared", Heading: "Two processes", Narration: fkNarration, Fork: &ForkBeat{Show: "shared"}},
		{ID: "w0", Heading: "First write", Narration: fkNarration, Fork: &ForkBeat{Show: "write", At: 0}},
		{ID: "w1", Heading: "Second write", Narration: fkNarration, Fork: &ForkBeat{Show: "write", At: 1}},
		{ID: "w2", Heading: "Third write", Narration: fkNarration, Fork: &ForkBeat{Show: "write", At: 2}},
	}
	p.targetWords = 4 * 40
	if err := validateForkPlan(p); err == nil {
		t.Fatal("a clip that copied every page was accepted")
	}
}

// Copy-on-write copies a page once. After the first write the writer owns its
// copy and writes straight to it — a second copy is the classic misunderstanding.
func TestForkRejectsCopyingTheSamePageTwice(t *testing.T) {
	p := forkPlan()
	p.Fork.Writes = append(p.Fork.Writes, ForkWrite{At: 1, By: "child"})
	p.Beats = append(p.Beats, SnippetBeat{
		ID: "again", Heading: "And again", Narration: fkNarration,
		Fork: &ForkBeat{Show: "write", At: 1},
	})
	err := validateForkPlan(p)
	if err == nil {
		t.Fatal("a page copied twice was accepted")
	}
	if !strings.Contains(err.Error(), "ONCE") {
		t.Fatalf("the error does not state the mechanism: %v", err)
	}
}

func TestForkRequiresTheSharedMemoryFirst(t *testing.T) {
	p := forkPlan()
	p.Beats[0].Fork = &ForkBeat{Show: "write", At: 0}
	p.Beats[1].Fork = &ForkBeat{Show: "shared"}
	if err := validateForkPlan(p); err == nil {
		t.Fatal("a clip that splits a page before showing the memory whole was accepted")
	}
}

// Once a page has diverged it does not go back, so drawing the memory whole
// again would say it did.
func TestForkRejectsShowingTheMemoryWholeTwice(t *testing.T) {
	p := forkPlan()
	p.Beats[2].Fork = &ForkBeat{Show: "shared"}
	if err := validateForkPlan(p); err == nil {
		t.Fatal("a clip that re-shares a diverged memory was accepted")
	}
}

func TestForkRejectsAWriteFromNeitherSide(t *testing.T) {
	p := forkPlan()
	p.Fork.Writes[0].By = "both"
	err := validateForkPlan(p)
	if err == nil {
		t.Fatal("a write from neither side was accepted")
	}
	if !strings.Contains(err.Error(), "nowhere to go") {
		t.Fatalf("the error does not explain why: %v", err)
	}
}

func TestForkRejectsAMemoryTooSmallToShowSharing(t *testing.T) {
	p := forkPlan()
	p.Fork.Pages = p.Fork.Pages[:3]
	if err := validateForkPlan(p); err == nil {
		t.Fatal("a three-page memory was accepted")
	}
}

// A second write to the same page by the same side lands on the copy that side
// already owns, so it is dropped rather than drawn as a second copy.
func TestForkNormalizeDropsARepeatedWriteBySameSide(t *testing.T) {
	p := forkPlan()
	p.Fork.Writes = append(p.Fork.Writes, ForkWrite{At: 1, By: "parent"})
	normalizeForkPlan(p)
	if len(p.Fork.Writes) != 1 {
		t.Fatalf("want 1 write after normalize, got %d", len(p.Fork.Writes))
	}
}

func TestForkNormalizeDropsWritesOffTheEndOfMemory(t *testing.T) {
	p := forkPlan()
	p.Fork.Writes = append(p.Fork.Writes, ForkWrite{At: 99, By: "child"})
	normalizeForkPlan(p)
	if len(p.Fork.Writes) != 1 {
		t.Fatalf("want 1 write after normalize, got %d", len(p.Fork.Writes))
	}
}

func TestForkSideNamesHaveDefaults(t *testing.T) {
	f := &ForkSpec{}
	if f.ResolvedParent() == "" || f.ResolvedChild() == "" {
		t.Fatal("a fork that names neither side renders with blank boxes")
	}
	f.Child = "the snapshot"
	if f.ResolvedChild() != "the snapshot" {
		t.Fatalf("a stated name was ignored: %q", f.ResolvedChild())
	}
}

// Each step carries the divergence as it stands, so the renderer draws a whole
// frame from one step rather than replaying the write list.
func TestForkScenesAccumulateTheCopiedPages(t *testing.T) {
	p := forkPlan()
	scenes, err := forkScenes(sceneInput(t, p, 12000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	first, _ := steps[0]["copied"].(map[string]any)
	if len(first) != 0 {
		t.Fatalf("the opening beat already has copies: %v", first)
	}
	after, _ := steps[1]["copied"].(map[string]any)
	if after["1"] != "parent" {
		t.Fatalf("page 1 is owned by %v after the write, want parent", after["1"])
	}
	// The closing beat still carries it: the divergence does not disappear when
	// the narrator stops talking about it.
	last, _ := steps[2]["copied"].(map[string]any)
	if last["1"] != "parent" {
		t.Fatalf("the copy was lost by the closing beat: %v", last)
	}
}
