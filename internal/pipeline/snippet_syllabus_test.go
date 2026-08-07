package pipeline

import (
	"strings"
	"testing"
)

const sylNarration = "This module is where the course stops being abstract and the machine starts feeling real."

func syllabusPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "syllabus",
		Title:    "Eight stops on the road to fluency",
		Syllabus: &SyllabusSpec{
			Modules: []SyllabusModule{
				{Label: "Hardware", Sub: "What a computer physically is"},
				{Label: "Binary & Data", Sub: "How everything becomes numbers"},
				{Label: "Memory & Storage", Sub: "Where the bytes actually live"},
				{Label: "Operating Systems", Sub: "Who is really in charge"},
				{Label: "Networks", Sub: "How machines talk"},
			},
			Current: 2,
		},
		Beats: []SnippetBeat{
			{ID: "map", Heading: "The whole route", Narration: sylNarration, Syllabus: &SyllabusBeat{Show: "map"}},
			{ID: "hardware", Heading: "Stop one", Narration: sylNarration, Syllabus: &SyllabusBeat{Show: "stop", At: 0}},
			{ID: "binary", Heading: "Stop two", Narration: sylNarration, Syllabus: &SyllabusBeat{Show: "stop", At: 1}},
			{ID: "here", Heading: "Where we are", Narration: sylNarration, Syllabus: &SyllabusBeat{Show: "here"}},
		},
	}
	p.targetWords = 4 * 40
	return p
}

func TestSyllabusPlanAccepted(t *testing.T) {
	if err := validateSyllabusPlan(syllabusPlan()); err != nil {
		t.Fatalf("a well-formed syllabus plan was rejected: %v", err)
	}
}

func TestSyllabusRejectsTooFewModules(t *testing.T) {
	p := syllabusPlan()
	p.Syllabus.Modules = p.Syllabus.Modules[:3]
	err := validateSyllabusPlan(p)
	if err == nil {
		t.Fatal("a three-module map was accepted")
	}
	if !strings.Contains(err.Error(), "journey") {
		t.Fatalf("the error does not explain why: %v", err)
	}
}

func TestSyllabusRejectsACurrentOffTheMap(t *testing.T) {
	p := syllabusPlan()
	p.Syllabus.Current = 9
	err := validateSyllabusPlan(p)
	if err == nil {
		t.Fatal("a current position off the map was accepted")
	}
	if !strings.Contains(err.Error(), "somewhere that exists") {
		t.Fatalf("the error does not explain why: %v", err)
	}
}

func TestSyllabusRequiresTheMapFirst(t *testing.T) {
	p := syllabusPlan()
	p.Beats[0].Syllabus = &SyllabusBeat{Show: "stop", At: 0}
	p.Beats[1].Syllabus = &SyllabusBeat{Show: "map"}
	if err := validateSyllabusPlan(p); err == nil {
		t.Fatal("a clip that lifts a stop before showing the route whole was accepted")
	}
}

func TestSyllabusRejectsASecondHere(t *testing.T) {
	p := syllabusPlan()
	p.Beats[2].Syllabus = &SyllabusBeat{Show: "here"}
	err := validateSyllabusPlan(p)
	if err == nil {
		t.Fatal("a clip that lands twice was accepted")
	}
	if !strings.Contains(err.Error(), "two presents") {
		t.Fatalf("the error does not explain why: %v", err)
	}
}

func TestSyllabusRequiresHereToBeTheLastBeat(t *testing.T) {
	p := syllabusPlan()
	p.Beats[2].Syllabus = &SyllabusBeat{Show: "here"}
	p.Beats[3].Syllabus = &SyllabusBeat{Show: "stop", At: 1}
	err := validateSyllabusPlan(p)
	if err == nil {
		t.Fatal("a clip that lands and then keeps touring was accepted")
	}
	if !strings.Contains(err.Error(), "arrival") {
		t.Fatalf("the error does not explain why: %v", err)
	}
}

func TestSyllabusRejectsVisitingAStopTwice(t *testing.T) {
	p := syllabusPlan()
	p.Beats[2].Syllabus = &SyllabusBeat{Show: "stop", At: 0}
	err := validateSyllabusPlan(p)
	if err == nil {
		t.Fatal("a stop visited twice was accepted")
	}
	if !strings.Contains(err.Error(), "second time") {
		t.Fatalf("the error does not name the repeat: %v", err)
	}
}

func TestSyllabusRejectsAnUnlabelledModule(t *testing.T) {
	p := syllabusPlan()
	p.Syllabus.Modules[3].Label = "   "
	if err := validateSyllabusPlan(p); err == nil {
		t.Fatal("a nameless stop was accepted")
	}
}

func TestSyllabusNormalizeCapsTheModuleList(t *testing.T) {
	p := syllabusPlan()
	for i := 0; i < 6; i++ {
		p.Syllabus.Modules = append(p.Syllabus.Modules, SyllabusModule{Label: "Extra"})
	}
	normalizeSyllabusPlan(p)
	if n := len(p.Syllabus.Modules); n != maxSyllabusModules {
		t.Fatalf("want %d modules after normalize, got %d", maxSyllabusModules, n)
	}
}

func TestSyllabusNormalizeClampsAStopOffTheMap(t *testing.T) {
	p := syllabusPlan()
	p.Beats[1].Syllabus.At = 99
	normalizeSyllabusPlan(p)
	if at := p.Beats[1].Syllabus.At; at != len(p.Syllabus.Modules)-1 {
		t.Fatalf("want the stop clamped to the last module, got %d", at)
	}
}

func TestSyllabusNormalizeClampsLongLabels(t *testing.T) {
	p := syllabusPlan()
	p.Syllabus.Modules[0].Label = "a very long module label indeed truly"
	normalizeSyllabusPlan(p)
	if n := len(strings.Fields(p.Syllabus.Modules[0].Label)); n != maxSyllabusLabelWords {
		t.Fatalf("want the label clamped to %d words, got %d", maxSyllabusLabelWords, n)
	}
}

func TestSyllabusShowDefaultsToStop(t *testing.T) {
	b := SyllabusBeat{Show: "wander"}
	if got := b.ResolvedShow(); got != "stop" {
		t.Fatalf("an unknown show resolved to %q, want stop", got)
	}
	b = SyllabusBeat{Show: " HERE "}
	if got := b.ResolvedShow(); got != "here" {
		t.Fatalf("a shouted here resolved to %q", got)
	}
}

// The ticks belong to the landing: nothing is done until "here", and then
// everything before the current stop is.
func TestSyllabusScenesTickOnlyAtTheLanding(t *testing.T) {
	p := syllabusPlan()
	scenes, err := syllabusScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if len(steps) != len(p.Beats) {
		t.Fatalf("want %d steps, got %d", len(p.Beats), len(steps))
	}
	first, _ := steps[0]["ticked"].([]int)
	if steps[0]["show"] != "map" || len(first) != 0 {
		t.Fatalf("the opening map already has ticks: %v", steps[0])
	}
	last := steps[len(steps)-1]
	ticked, _ := last["ticked"].([]int)
	if last["show"] != "here" || len(ticked) != p.Syllabus.Current {
		t.Fatalf("the landing does not tick everything before module %d: %v", p.Syllabus.Current, last)
	}
	for i, at := range ticked {
		if at != i {
			t.Fatalf("ticks are not the modules before current: %v", ticked)
		}
	}
}
