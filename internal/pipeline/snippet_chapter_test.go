package pipeline

import (
	"strings"
	"testing"
)

const chNarration = "Two parts down, and the third one is where the work actually starts to repeat itself."

func chapterPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "chapter",
		Title:    "Part three: loops",
		Chapter: &ChapterSpec{
			Path: "The Python course",
			At:   2,
			Stops: []ChapterStop{
				{Label: "Printing", Icon: "terminal", Note: "Getting Python to say something back"},
				{Label: "Variables", Icon: "box", Note: "Names for the things you keep"},
				{Label: "Loops", Icon: "refresh", Note: "The same work without writing it twice"},
				{Label: "Functions", Icon: "puzzle", Note: "Work you can call by name"},
				{Label: "Files", Icon: "folder", Note: "What outlives the program"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "where", Heading: "Where you are", Narration: chNarration, Chapter: &ChapterBeat{Show: "path"}},
			{ID: "printed", Heading: "You can print", Narration: chNarration, Chapter: &ChapterBeat{Show: "done", At: 0}},
			{ID: "stored", Heading: "You can store", Narration: chNarration, Chapter: &ChapterBeat{Show: "done", At: 1}},
			{ID: "loops", Heading: "Loops start now", Narration: chNarration, Chapter: &ChapterBeat{Show: "here"}},
		},
	}
	p.targetWords = 4 * 40
	return p
}

func TestChapterPlanAccepted(t *testing.T) {
	if err := validateChapterPlan(chapterPlan()); err != nil {
		t.Fatalf("a well-formed chapter break was rejected: %v", err)
	}
}

// The rule this template exists for, first half: the clip hands over. A break
// that ends by summarising what closed leaves the viewer at a stopping point.
func TestChapterEndsOnWhatOpensNext(t *testing.T) {
	p := chapterPlan()
	// Same beats, re-pointed rather than removed — dropping one would trip the
	// shared beat-count floor first and prove nothing about this rule.
	p.Beats[3].Chapter = &ChapterBeat{Show: "done", At: 1}
	p.Beats[2].Chapter = &ChapterBeat{Show: "here"}
	err := validateChapterPlan(p)
	if err == nil {
		t.Fatal("a break that closed on what had already finished was accepted")
	}
	if !strings.Contains(err.Error(), "other side of it") {
		t.Errorf("the error does not say why the order matters: %v", err)
	}
}

// The rule this template exists for, second half: the marker only moves
// forward, so nothing may be recalled from at or ahead of the break.
func TestChapterOnlyLooksBackwards(t *testing.T) {
	p := chapterPlan()
	p.Beats[2].Chapter = &ChapterBeat{Show: "done", At: 3}
	err := validateChapterPlan(p)
	if err == nil {
		t.Fatal("a break that recalled a section still ahead of the viewer was accepted")
	}
	if !strings.Contains(err.Error(), "only moves forward") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChapterLookBacksRunForwards(t *testing.T) {
	p := chapterPlan()
	p.Beats[1].Chapter = &ChapterBeat{Show: "done", At: 1}
	p.Beats[2].Chapter = &ChapterBeat{Show: "done", At: 0}
	err := validateChapterPlan(p)
	if err == nil {
		t.Fatal("a break that walked its own path backwards was accepted")
	}
	if !strings.Contains(err.Error(), "recap") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChapterOpensOnThePath(t *testing.T) {
	p := chapterPlan()
	p.Beats[0].Chapter = &ChapterBeat{Show: "done", At: 0}
	p.Beats[1].Chapter = &ChapterBeat{Show: "path"}
	err := validateChapterPlan(p)
	if err == nil {
		t.Fatal("a break that recalled a stop before drawing the path was accepted")
	}
	if !strings.Contains(err.Error(), "position on") {
		t.Errorf("unexpected error: %v", err)
	}
}

// A break in the middle of a course that never names what closed is a title
// card. At the very first stop there is nothing behind, and demanding a look
// back there would be asking the model to invent one.
func TestChapterMustNameWhatClosed(t *testing.T) {
	// The shortest legal shape — draw the path, hand over — which is exactly the
	// plan that never looks back. Budgeted for two beats so the shared
	// beat-count floor is not what rejects it.
	short := func(at int) *SnippetPlan {
		p := chapterPlan()
		p.Chapter.At = at
		p.Beats = []SnippetBeat{
			{ID: "where", Heading: "Where you are", Narration: chNarration, Chapter: &ChapterBeat{Show: "path"}},
			{ID: "next", Heading: "Starting now", Narration: chNarration, Chapter: &ChapterBeat{Show: "here"}},
		}
		p.targetWords = 2 * 40
		return p
	}

	err := validateChapterPlan(short(2))
	if err == nil {
		t.Fatal("a break part-way through a course that never looked back was accepted")
	}
	if !strings.Contains(err.Error(), "title card") {
		t.Errorf("unexpected error: %v", err)
	}

	// At the first stop the very same plan is right, because there is nothing
	// behind the viewer to name and demanding one would be asking the model to
	// invent it.
	if err := validateChapterPlan(short(0)); err != nil {
		t.Errorf("a break at the first stop was told to look back at nothing: %v", err)
	}
}

func TestChapterNeedsAPathName(t *testing.T) {
	p := chapterPlan()
	p.Chapter.Path = ""
	if err := validateChapterPlan(p); err == nil {
		t.Fatal("a break with no run to be a break in was accepted")
	}
}

func TestChapterRejectsDuplicateStops(t *testing.T) {
	p := chapterPlan()
	p.Chapter.Stops[3].Label = "printing"
	err := validateChapterPlan(p)
	if err == nil {
		t.Fatal("two stops with the same name were accepted")
	}
	if !strings.Contains(err.Error(), "counting their way") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChapterDrawsThePathOnce(t *testing.T) {
	p := chapterPlan()
	p.Beats[1].Chapter = &ChapterBeat{Show: "path"}
	if err := validateChapterPlan(p); err == nil {
		t.Fatal("a break that drew its path twice was accepted")
	}
}

func TestNormalizeChapterRepairsMechanicalMistakes(t *testing.T) {
	p := chapterPlan()
	p.Chapter.Path = "  The   Python course that goes on for ever and ever  "
	p.Chapter.Stops[0].Icon = "unicorn"
	p.Chapter.At = 99
	p.Beats[1].Chapter.Show = "recap"
	p.Beats[1].Chapter.At = 42
	normalizeChapterPlan(p)

	if got := p.Chapter.Path; got != "The Python course that goes" {
		t.Errorf("the path name was not collapsed and clamped: %q", got)
	}
	if got := p.Chapter.Stops[0].Icon; got != "" {
		t.Errorf("an invented icon survived normalization: %q", got)
	}
	if got := p.Chapter.At; got != len(p.Chapter.Stops)-1 {
		t.Errorf("a break past the end of its own path was not clamped: %d", got)
	}
	if got := p.Beats[1].Chapter.Show; got != "done" {
		t.Errorf("an invented middle-beat action was not repaired: %q", got)
	}
	if got := p.Beats[1].Chapter.At; got >= len(p.Chapter.Stops) {
		t.Errorf("a look-back at a stop that does not exist was not clamped: %d", got)
	}
}

func TestChapterScenesCarryTheWholePath(t *testing.T) {
	p := chapterPlan()
	scenes, err := chapterScenes(sceneInput(t, p, 7000))
	if err != nil {
		t.Fatalf("laying the break out: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("want one standing scene, got %d — the card is furniture, and a scene per beat would remount it", len(scenes))
	}
	props := scenes[0].Props
	if props["ordinal"] != 3 {
		t.Errorf("the ordinal on screen is %v, want 3 — the picture says part three, not part index two", props["ordinal"])
	}
	stops, _ := props["stops"].([]map[string]any)
	if len(stops) != 5 {
		t.Fatalf("want every stop on the path, got %d", len(stops))
	}
	for i, want := range []string{"done", "done", "here", "ahead", "ahead"} {
		if got := stops[i]["state"]; got != want {
			t.Errorf("stop %d is %v, want %v", i, got, want)
		}
	}
}
