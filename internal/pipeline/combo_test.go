package pipeline

import (
	"slices"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// comboPlanFixture is a two-segment combo: a rundown followed by a gauge. It
// reuses the per-template plan fixtures, which is the point — a segment is
// planned by exactly the code a snippet is, so the fixtures are the same ones.
func comboPlanFixture() *ComboPlan {
	return &ComboPlan{
		Title: "What actually decides whether a model runs",
		Segments: []ComboPlanSegment{
			{ID: "the-three", Template: "rundown", Plan: rundownPlan()},
			{ID: "does-it-fit", Template: "gauge", Plan: gaugePlan()},
		},
	}
}

func comboSpecFixture() ComboSpec {
	s := ComboSpec{
		ID:    "gpu-combo",
		Title: "What actually decides whether a model runs",
		Brief: "Explain what decides whether a model runs on your machine.",
		Segments: []ComboSegment{
			{Template: "rundown", Prompt: "the three numbers that decide it"},
			{Template: "gauge", Prompt: "which models fit in 24GB"},
		},
	}
	s.EnsureSegmentIDs()
	return s
}

// comboAlignment fakes the aligner: one section per beat across every segment,
// evenly spaced, which is the shape CaptionSections really returns.
func comboAlignment(plan *ComboPlan, beatMs int) *Alignment {
	a := &Alignment{}
	i := 0
	for _, seg := range plan.Segments {
		for _, b := range seg.Plan.Beats {
			a.Sections = append(a.Sections, SectionSpan{
				ID:      seg.ID + "--" + b.ID,
				StartMs: i * beatMs,
				EndMs:   (i + 1) * beatMs,
			})
			i++
		}
	}
	return a
}

// The claim Phase 1 exists to prove: several templates land on one timeline,
// in order, with no gaps and no overlaps.
func TestComboAssemblesTemplatesOntoOneTimeline(t *testing.T) {
	plan := comboPlanFixture()
	spec := comboSpecFixture()
	const beatMs = 6000
	align := comboAlignment(plan, beatMs)
	total := plan.Beats() * beatMs

	graph, err := buildComboSceneGraph(
		&project.Course{Name: "Combos", Slug: "combos"},
		&project.Lesson{ID: "gpu-combo"},
		config.Defaults(), spec, plan, align, nil, total)
	if err != nil {
		t.Fatalf("buildComboSceneGraph: %v", err)
	}
	if len(graph.Scenes) < 2 {
		t.Fatalf("got %d scenes, want at least one per segment", len(graph.Scenes))
	}

	// Both templates are represented, in the order the combo declares them.
	kinds := []string{}
	for _, s := range graph.Scenes {
		if len(kinds) == 0 || kinds[len(kinds)-1] != s.Type {
			kinds = append(kinds, s.Type)
		}
	}
	if len(kinds) != 2 || kinds[0] != SceneRundown || kinds[1] != SceneGauge {
		t.Errorf("scene types run %v, want rundown then gauge", kinds)
	}

	// The timeline is continuous. A gap would show as a black frame mid-video
	// and an overlap would render two templates on top of each other.
	for i, s := range graph.Scenes {
		if s.EndMs <= s.StartMs {
			t.Errorf("scene %d (%s) ends at or before it starts", i, s.Type)
		}
		if i > 0 && s.StartMs != graph.Scenes[i-1].EndMs {
			t.Errorf("scene %d starts at %dms but the one before ends at %dms — the cut has a %dms hole",
				i, s.StartMs, graph.Scenes[i-1].EndMs, s.StartMs-graph.Scenes[i-1].EndMs)
		}
	}
	if graph.Scenes[0].StartMs != 0 {
		t.Errorf("the combo opens at %dms, want 0", graph.Scenes[0].StartMs)
	}
}

// The join between two segments has to land on a word, not on a gap: the last
// beat of one segment holds until the first beat of the next begins.
func TestComboSegmentsMeetExactly(t *testing.T) {
	plan := comboPlanFixture()
	const beatMs = 6000
	first := len(plan.Segments[0].Plan.Beats)

	graph, err := buildComboSceneGraph(
		&project.Course{Name: "Combos"}, &project.Lesson{ID: "r"},
		config.Defaults(), comboSpecFixture(), plan,
		comboAlignment(plan, beatMs), nil, plan.Beats()*beatMs)
	if err != nil {
		t.Fatalf("buildComboSceneGraph: %v", err)
	}
	// Both templates lay their whole segment out as one scene, so the boundary
	// is the boundary between scene 0 and scene 1.
	if got := graph.Scenes[0].EndMs; got != first*beatMs {
		t.Errorf("the first segment ends at %dms, want %dms — the cut is not on the word",
			got, first*beatMs)
	}
	// Only the final beat runs past its span, so the piece does not cut on the
	// last syllable.
	last := graph.Scenes[len(graph.Scenes)-1]
	if last.EndMs <= plan.Beats()*beatMs {
		t.Errorf("the combo ends at %dms, want a tail past the last word at %dms",
			last.EndMs, plan.Beats()*beatMs)
	}
}

// A miscounted alignment is the failure that would otherwise render every
// segment after the mistake against the wrong audio.
func TestComboRejectsMismatchedAlignment(t *testing.T) {
	plan := comboPlanFixture()
	align := comboAlignment(plan, 6000)
	align.Sections = align.Sections[:len(align.Sections)-1]

	_, err := buildComboSceneGraph(
		&project.Course{Name: "Combos"}, &project.Lesson{ID: "r"},
		config.Defaults(), comboSpecFixture(), plan, align, nil, 60000)
	if err == nil {
		t.Fatal("an alignment with the wrong number of sections was accepted")
	}
	if !strings.Contains(err.Error(), "re-run the align stage") {
		t.Errorf("the error does not say how to fix it: %v", err)
	}
}

// One narration, one read: the script is every segment's beats in order, and
// the TTS stage never learns where a segment ends.
func TestComboScriptIsOneContinuousRead(t *testing.T) {
	plan := comboPlanFixture()
	script := plan.Script(150)
	if len(script.Sections) != plan.Beats() {
		t.Fatalf("script has %d sections, want %d (every beat of every segment)",
			len(script.Sections), plan.Beats())
	}
	// Section ids are namespaced by segment. Beat ids are only unique within a
	// template's plan, and the aligner keys sections by id — two segments both
	// opening on "intro" would collide and mistime everything after them.
	seen := map[string]bool{}
	for _, s := range script.Sections {
		if seen[s.ID] {
			t.Errorf("duplicate section id %q — the aligner keys on this", s.ID)
		}
		seen[s.ID] = true
		if !strings.Contains(s.ID, "--") {
			t.Errorf("section id %q is not namespaced to its segment", s.ID)
		}
	}
}

// Skip is how the editor drops a segment without losing the prompt that made
// it, so it must actually leave the cut.
func TestComboSkipRemovesSegmentFromTheCut(t *testing.T) {
	spec := comboSpecFixture()
	spec.Segments = append(spec.Segments, ComboSegment{
		ID: "extra", Template: "metric", Prompt: "the memory arithmetic", Skip: true,
	})
	if got := len(spec.Active()); got != 2 {
		t.Errorf("%d segments in the cut, want 2 — a skipped segment is still rendered", got)
	}
	// And it stays in the file, so putting it back is one flag rather than
	// rewriting the prompt.
	if len(spec.Segments) != 3 {
		t.Errorf("skipping deleted the segment; it should only leave the cut")
	}
}

func TestComboValidate(t *testing.T) {
	ok := comboSpecFixture()
	if err := ok.Validate(); err != nil {
		t.Fatalf("a well-formed combo was rejected: %v", err)
	}

	// One segment is a snippet.
	one := ok
	one.Segments = ok.Segments[:1]
	if err := one.Validate(); err == nil {
		t.Error("a single-segment combo was accepted")
	} else if !strings.Contains(err.Error(), "snippet") {
		t.Errorf("the error does not point at the right tool: %v", err)
	}

	// An unknown template names the catalog rather than failing at render.
	bad := comboSpecFixture()
	bad.Segments[1].Template = "nonsense"
	if err := bad.Validate(); err == nil {
		t.Error("an unknown template was accepted")
	}

	// A segment with nothing to cover cannot be planned.
	empty := comboSpecFixture()
	empty.Segments[0].Prompt = ""
	if err := empty.Validate(); err == nil {
		t.Error("a segment with no prompt was accepted")
	}
}

// Ids are how an edit addresses a segment, so they must survive their
// neighbours moving.
func TestComboSegmentIDsAreStable(t *testing.T) {
	spec := comboSpecFixture()
	before := []string{spec.Segments[0].ID, spec.Segments[1].ID}

	// Insert at the front — the commonest edit after a first watch.
	spec.Segments = append([]ComboSegment{
		{Template: "myth", Prompt: "the belief everyone starts with"},
	}, spec.Segments...)
	spec.EnsureSegmentIDs()

	if spec.Segments[1].ID != before[0] || spec.Segments[2].ID != before[1] {
		t.Errorf("ids moved when a segment was inserted: %q,%q became %q,%q",
			before[0], before[1], spec.Segments[1].ID, spec.Segments[2].ID)
	}
	if spec.Segments[0].ID == "" {
		t.Error("the new segment was not given an id")
	}
	// And the generated ids are unique, or two edits address the same segment.
	seen := map[string]bool{}
	for _, s := range spec.Segments {
		if seen[s.ID] {
			t.Errorf("duplicate segment id %q", s.ID)
		}
		seen[s.ID] = true
	}
}

// A combo mixing a code template with nine that show none still has to verify,
// or it ships a clip claiming output nothing produced.
func TestComboStagesKeepVerifyWhenAnySegmentShowsCode(t *testing.T) {
	dir := t.TempDir()
	spec := &ComboSpec{
		Segments: []ComboSegment{
			{Template: "metric", Prompt: "the numbers"},
			{Template: "vscode", Prompt: "the code that proves it"},
		},
	}
	if err := SaveComboSpec(dir, spec); err != nil {
		t.Fatalf("SaveComboSpec: %v", err)
	}
	stages, err := ComboStages(&project.Lesson{ID: "r", Dir: dir})
	if err != nil {
		t.Fatalf("ComboStages: %v", err)
	}
	if !slices.Contains(stages, project.StageVerify) {
		t.Error("verify was dropped from a combo containing a code template")
	}

	// And a combo with no code at all does not pay for it.
	noCode := &ComboSpec{Segments: []ComboSegment{
		{Template: "metric", Prompt: "the numbers"},
		{Template: "gauge", Prompt: "what fits"},
	}}
	dir2 := t.TempDir()
	if err := SaveComboSpec(dir2, noCode); err != nil {
		t.Fatalf("SaveComboSpec: %v", err)
	}
	stages, err = ComboStages(&project.Lesson{ID: "r", Dir: dir2})
	if err != nil {
		t.Fatalf("ComboStages: %v", err)
	}
	if slices.Contains(stages, project.StageVerify) {
		t.Error("verify runs for a combo where nothing shows code")
	}
}

// Round-tripping combo.yaml is what makes it an edit surface: what a person
// writes has to survive being read back.
func TestComboSpecRoundTrips(t *testing.T) {
	dir := t.TempDir()
	want := comboSpecFixture()
	want.Segments[1].Skip = true
	want.Segments[1].TargetSec = 60
	if err := SaveComboSpec(dir, &want); err != nil {
		t.Fatalf("SaveComboSpec: %v", err)
	}
	got, err := LoadComboSpec(dir)
	if err != nil {
		t.Fatalf("LoadComboSpec: %v", err)
	}
	if got.Title != want.Title || got.Brief != want.Brief {
		t.Errorf("title/brief did not survive the round trip")
	}
	if len(got.Segments) != len(want.Segments) {
		t.Fatalf("got %d segments, want %d", len(got.Segments), len(want.Segments))
	}
	for i := range want.Segments {
		if got.Segments[i] != want.Segments[i] {
			t.Errorf("segment %d changed: %+v -> %+v", i, want.Segments[i], got.Segments[i])
		}
	}
}
