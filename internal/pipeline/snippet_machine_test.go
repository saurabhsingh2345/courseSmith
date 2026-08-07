package pipeline

import (
	"strings"
	"testing"
)

const mcNarration = "This part has one job and it does that job every single moment the machine is on."

func machinePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "machine",
		Title:    "What is really inside the box",
		Machine: &MachineSpec{
			Chassis: "a desktop PC",
			Parts: []MachinePart{
				{Label: "CPU", Job: "runs every instruction the machine executes", Size: "large"},
				{Label: "RAM", Job: "holds whatever the CPU is working on right now", Size: "medium"},
				{Label: "storage", Job: "keeps files after the power goes off", Size: "medium"},
				{Label: "power supply", Job: "turns wall power into what the parts need", Size: "small"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "whole", Heading: "The box", Narration: mcNarration, Machine: &MachineBeat{Show: "whole"}},
			{ID: "cpu", Heading: "The CPU", Narration: mcNarration, Machine: &MachineBeat{Show: "part", At: 0}},
			{ID: "ram", Heading: "The RAM", Narration: mcNarration, Machine: &MachineBeat{Show: "part", At: 1}},
			{ID: "disk", Heading: "The storage", Narration: mcNarration, Machine: &MachineBeat{Show: "part", At: 2}},
			{ID: "psu", Heading: "The power supply", Narration: mcNarration, Machine: &MachineBeat{Show: "part", At: 3}},
			{ID: "fit", Heading: "One machine", Narration: mcNarration, Machine: &MachineBeat{Show: "fit"}},
		},
	}
	// A beat here is a shot, so the fixture budget is sized at the template's
	// own 28-word ideal — nBeats * 40 would make beatBounds demand more beats
	// than the fixture has.
	p.targetWords = 6 * 28
	return p
}

func TestMachinePlanAccepted(t *testing.T) {
	if err := validateMachinePlan(machinePlan()); err != nil {
		t.Fatalf("a well-formed machine plan was rejected: %v", err)
	}
}

func TestMachineRequiresTheWholeBoxFirst(t *testing.T) {
	p := machinePlan()
	p.Beats[0].Machine = &MachineBeat{Show: "part", At: 0}
	p.Beats[1].Machine = &MachineBeat{Show: "whole"}
	err := validateMachinePlan(p)
	if err == nil {
		t.Fatal("a clip that lifts a part before showing the box whole was accepted")
	}
	if !strings.Contains(err.Error(), "open on the whole machine") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMachineRequiresTheRefitLast(t *testing.T) {
	p := machinePlan()
	p.Beats[5].Machine = &MachineBeat{Show: "part", At: 3}
	p.Beats[4].Machine = &MachineBeat{Show: "fit"}
	err := validateMachinePlan(p)
	if err == nil {
		t.Fatal("a clip that does not close on the refit was accepted")
	}
}

// Every part gets exactly one lift: a part never spoken about is a mystery the
// clip planted on purpose.
func TestMachineRejectsAPartNeverLifted(t *testing.T) {
	p := machinePlan()
	// Drop the power-supply beat entirely, so part 3 is drawn but never lifted.
	p.Beats = append(p.Beats[:4], p.Beats[5])
	p.targetWords = 5 * 28
	err := validateMachinePlan(p)
	if err == nil {
		t.Fatal("a plan with an unlifted part was accepted")
	}
	if !strings.Contains(err.Error(), "never lifted") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMachineRejectsLiftingAPartTwice(t *testing.T) {
	p := machinePlan()
	p.Beats = append(p.Beats[:5], SnippetBeat{
		ID: "again", Heading: "The CPU again", Narration: mcNarration,
		Machine: &MachineBeat{Show: "part", At: 0},
	}, p.Beats[5])
	p.targetWords = 7 * 28
	err := validateMachinePlan(p)
	if err == nil {
		t.Fatal("a part lifted twice was accepted")
	}
	if !strings.Contains(err.Error(), "exactly one lift") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMachineRejectsTooFewParts(t *testing.T) {
	p := machinePlan()
	p.Machine.Parts = p.Machine.Parts[:3]
	if err := validateMachinePlan(p); err == nil {
		t.Fatal("a three-part machine was accepted")
	}
}

func TestMachineRejectsAPartWithNoJob(t *testing.T) {
	p := machinePlan()
	p.Machine.Parts[1].Job = ""
	err := validateMachinePlan(p)
	if err == nil {
		t.Fatal("a part with no job was accepted")
	}
	if !strings.Contains(err.Error(), "no job") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMachineRejectsAnInventedSize(t *testing.T) {
	p := machinePlan()
	p.Machine.Parts[0].Size = "huge"
	err := validateMachinePlan(p)
	if err == nil {
		t.Fatal("an invented part size was accepted")
	}
	if !strings.Contains(err.Error(), "no width") {
		t.Fatalf("the error does not explain why: %v", err)
	}
}

func TestMachineNormalizeCoercesSizes(t *testing.T) {
	p := machinePlan()
	p.Machine.Parts[0].Size = "HUGE"
	p.Machine.Parts[1].Size = ""
	normalizeMachinePlan(p)
	if got := p.Machine.Parts[0].Size; got != "medium" {
		t.Fatalf("an unknown size resolved to %q, want medium", got)
	}
	if got := p.Machine.Parts[1].Size; got != "medium" {
		t.Fatalf("an empty size resolved to %q, want medium", got)
	}
}

func TestMachineNormalizeClampsAnOutOfRangeLift(t *testing.T) {
	p := machinePlan()
	p.Beats[1].Machine.At = 99
	normalizeMachinePlan(p)
	if got := p.Beats[1].Machine.At; got != 3 {
		t.Fatalf("an out-of-range lift clamped to %d, want 3", got)
	}
}

func TestMachineSizeDefaultsToMedium(t *testing.T) {
	part := MachinePart{Label: "CPU", Job: "runs instructions"}
	if got := part.ResolvedSize(); got != "medium" {
		t.Fatalf("an unstated size resolved to %q, want medium", got)
	}
}

// Each step carries the visited set as it stands, so the renderer draws a
// whole frame from one step rather than replaying the beat list.
func TestMachineScenesAccumulateVisitedParts(t *testing.T) {
	p := machinePlan()
	scenes, err := machineScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if len(steps) != 6 {
		t.Fatalf("want 6 steps, got %d", len(steps))
	}
	first, _ := steps[0]["visited"].([]int)
	if len(first) != 0 {
		t.Fatalf("the opening beat already has visited parts: %v", first)
	}
	last, _ := steps[len(steps)-1]["visited"].([]int)
	if len(last) != 4 {
		t.Fatalf("the closing beat has %d visited parts, want all 4: %v", len(last), last)
	}
	if steps[5]["show"] != "fit" {
		t.Fatalf("the last step shows %v, want fit", steps[5]["show"])
	}
}
