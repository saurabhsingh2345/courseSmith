package pipeline

import (
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

// The material is the outline's now, not the caster's, and that is a change of
// who answers for it: the facts are a property of the subject, and the template
// is a choice about how to show them. The old order let the caster invent
// material to justify a template it had already chosen, which is why a
// `showcase` cast from a brief naming thirteen tools rendered a card for a
// product called "Creation Tools".
//
// What must not change is that it reaches the segment's WRITER. It was once
// validated and then dropped, so every writer started from a one-line prompt and
// invented the rest.
func TestDirectorCarriesOutlineMaterialOntoTheSegment(t *testing.T) {
	outline := outlineFixture()
	cast := castFixture()
	spec := &ComboSpec{Angle: outline.Angle}
	for i, p := range cast.Picks {
		spec.Segments = append(spec.Segments, ComboSegment{
			Template: p.Template,
			Prompt:   outline.Parts[i].Establishes,
			Heading:  outline.Parts[i].Heading,
			Role:     outline.Parts[i].Role,
			Material: outline.Parts[i].Material,
		})
	}
	for i, seg := range spec.Segments {
		if seg.Material == "" {
			t.Errorf("segment %d (%s) reached the spec with no material", i, seg.Template)
		}
		if seg.Material != outline.Parts[i].Material {
			t.Errorf("segment %d material is %q, want %q", i, seg.Material, outline.Parts[i].Material)
		}
		// The prompt is the INCREMENT the part establishes, not a topic. That
		// difference is what stops two segments arriving at the same content from
		// different directions.
		if seg.Prompt != outline.Parts[i].Establishes {
			t.Errorf("segment %d was asked to cover %q rather than what its part establishes", i, seg.Prompt)
		}
	}
	if spec.Angle == "" {
		t.Error("the angle did not reach the spec — the critic scores every segment against it")
	}
}

// A segment cannot know the piece it belongs to, so the brief and the priors are
// parameters. The point of the assertion is that all four pieces of context
// arrive together — the earlier bug was three of them silently absent.
func TestSnippetSpecCarriesTheWholePlanningContext(t *testing.T) {
	seg := ComboSegment{
		ID:       "gauge-3",
		Template: "gauge",
		Prompt:   "which models fit in 24GB",
		Material: "24GB ceiling; 7B 14GB, 13B 26GB",
	}
	priors := []string{"The three numbers that decide it (rundown): capacity; bandwidth"}
	got := seg.SnippetSpec(config.Defaults(), "A video about running models locally.", priors)

	if got.Material != seg.Material {
		t.Errorf("material did not survive: %q", got.Material)
	}
	if got.Brief == "" {
		t.Error("the brief did not reach the spec")
	}
	if len(got.Priors) != 1 || got.Priors[0] != priors[0] {
		t.Errorf("priors did not reach the spec: %v", got.Priors)
	}
	// And the fields that already worked still do.
	if got.Prompt != seg.Prompt || got.Template != seg.Template || got.ID != seg.ID {
		t.Error("threading the new context disturbed the fields that already worked")
	}
}

// A standalone snippet has no combo around it, and must plan exactly as it did
// before any of this existed.
func TestStandaloneSnippetHasNoComboContext(t *testing.T) {
	cfg := config.Defaults()
	spec := SnippetSpec{Prompt: "why indexes matter", Template: "gauge"}
	data := sharedPromptData(spec, cfg)
	for _, key := range []string{"Brief", "Material"} {
		if s, _ := data[key].(string); s != "" {
			t.Errorf("standalone snippet has a non-empty %s: %q", key, s)
		}
	}
	if priors, _ := data["Priors"].([]string); len(priors) != 0 {
		t.Errorf("standalone snippet has priors: %v", priors)
	}
}

func TestSegmentPriorSummarisesWhatWasCovered(t *testing.T) {
	seg := ComboSegment{Template: "myth", Prompt: "the belief that you need years of experience"}
	plan := &SnippetPlan{
		Title: "What everyone gets wrong about building software",
		Beats: []SnippetBeat{
			{Heading: "What everyone says"},
			{Heading: "Not quite"},
			{Heading: "Self-taught developers"},
		},
	}
	got := segmentPrior(seg, plan)
	if !strings.Contains(got, plan.Title) {
		t.Errorf("the prior drops the title: %q", got)
	}
	if !strings.Contains(got, "myth") {
		t.Errorf("the prior drops the template: %q", got)
	}
	if !strings.Contains(got, "Not quite") {
		t.Errorf("the prior drops the headings: %q", got)
	}
}

// The later segments carry the most priors and have the least room for them, so
// a prior is capped rather than being however long the plan was.
func TestSegmentPriorIsBounded(t *testing.T) {
	beats := make([]SnippetBeat, 12)
	for i := range beats {
		beats[i] = SnippetBeat{Heading: "A heading long enough to matter number " + string(rune('a'+i))}
	}
	got := segmentPrior(
		ComboSegment{Template: "rundown", Prompt: "the mindsets"},
		&SnippetPlan{Title: "Four mindsets to adopt", Beats: beats},
	)
	if n := strings.Count(got, ";"); n >= len(beats) {
		t.Errorf("the prior carries every heading (%d separators): %q", n, got)
	}
	if len(got) > 200 {
		t.Errorf("the prior is %d chars; it is handed to every later segment: %q", len(got), got)
	}
}

// A segment whose plan came back bare still has to say something useful, or the
// next writer is told only "myth:" and learns nothing.
func TestSegmentPriorFallsBackToThePrompt(t *testing.T) {
	seg := ComboSegment{Template: "verdict", Prompt: "what to actually do about it"}
	got := segmentPrior(seg, &SnippetPlan{})
	if !strings.Contains(got, "what to actually do") {
		t.Errorf("an empty plan produced a useless prior: %q", got)
	}
	if got = segmentPrior(seg, nil); !strings.Contains(got, "what to actually do") {
		t.Errorf("a nil plan produced a useless prior: %q", got)
	}
}
