package pipeline

import (
	"strings"
	"testing"
)

const msNarration = "This one is small enough to finish tonight and real enough to show somebody afterwards."

func missionPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "mission",
		Title:    "Build a hardware inventory report",
		Mission: &MissionSpec{
			Goal: "List every drive on your machine and its size",
			Specs: []string{
				"find each drive on the system",
				"print its size in gigabytes",
				"flag anything under ten percent free",
			},
			Deliverable: "a command-line script",
			Done:        "running it prints one line per drive",
		},
		Beats: []SnippetBeat{
			{ID: "brief", Heading: "What you are building", Narration: msNarration, Mission: &MissionBeat{Show: "brief"}},
			{ID: "find-them", Heading: "Find the drives", Narration: msNarration, Mission: &MissionBeat{Show: "spec", At: 0}},
			{ID: "size-them", Heading: "Print the sizes", Narration: msNarration, Mission: &MissionBeat{Show: "spec", At: 1}},
			{ID: "flag-them", Heading: "Flag the full ones", Narration: msNarration, Mission: &MissionBeat{Show: "spec", At: 2}},
			{ID: "artifact", Heading: "What you hand in", Narration: msNarration, Mission: &MissionBeat{Show: "deliverable"}},
			{ID: "done-when", Heading: "Done when", Narration: msNarration, Mission: &MissionBeat{Show: "done"}},
		},
	}
	p.targetWords = 6 * 40
	return p
}

func TestMissionPlanAccepted(t *testing.T) {
	if err := validateMissionPlan(missionPlan()); err != nil {
		t.Fatalf("a well-formed mission plan was rejected: %v", err)
	}
}

func TestMissionRejectsAMissingGoal(t *testing.T) {
	p := missionPlan()
	p.Mission.Goal = "  "
	if err := validateMissionPlan(p); err == nil {
		t.Fatal("a brief with no goal was accepted, and the checklist is then requirements for nothing")
	}
}

func TestMissionRejectsTooFewSpecs(t *testing.T) {
	p := missionPlan()
	p.Mission.Specs = p.Mission.Specs[:2]
	err := validateMissionPlan(p)
	if err == nil {
		t.Fatal("a two-row checklist was accepted, and under three it is the goal restated")
	}
	if !strings.Contains(err.Error(), "goal restated") {
		t.Fatalf("the error does not explain why: %v", err)
	}
}

func TestMissionRejectsTooManySpecs(t *testing.T) {
	p := missionPlan()
	for i := 0; i < 4; i++ {
		p.Mission.Specs = append(p.Mission.Specs, "one more requirement")
	}
	if err := validateMissionPlan(p); err == nil {
		t.Fatal("a seven-row checklist was accepted, and past six the mission is a sprint")
	}
}

func TestMissionRejectsAMissingDeliverable(t *testing.T) {
	p := missionPlan()
	p.Mission.Deliverable = ""
	err := validateMissionPlan(p)
	if err == nil {
		t.Fatal("a brief with no artifact was accepted, and a brief that assigns an activity produces practice")
	}
	if !strings.Contains(err.Error(), "attach to an email") {
		t.Fatalf("the error does not say what an artifact is: %v", err)
	}
}

func TestMissionRejectsAMissingDoneLine(t *testing.T) {
	p := missionPlan()
	p.Mission.Done = "   "
	err := validateMissionPlan(p)
	if err == nil {
		t.Fatal("a mission with no definition of done was accepted, and a mission whose end cannot be recognised does not end")
	}
	if !strings.Contains(err.Error(), "observable") {
		t.Fatalf("the error does not say what done has to be: %v", err)
	}
}

func TestMissionRequiresOpeningOnTheBrief(t *testing.T) {
	p := missionPlan()
	p.Beats[0].Mission = &MissionBeat{Show: "spec", At: 0}
	p.Beats[1].Mission = &MissionBeat{Show: "brief"}
	if err := validateMissionPlan(p); err == nil {
		t.Fatal("a clip that lands a checklist row before the goal was accepted")
	}
}

func TestMissionRejectsRestatingTheBrief(t *testing.T) {
	p := missionPlan()
	p.Beats[2].Mission = &MissionBeat{Show: "brief"}
	err := validateMissionPlan(p)
	if err == nil {
		t.Fatal("a clip that restates the goal mid-card was accepted")
	}
	if !strings.Contains(err.Error(), "wipes the checklist") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissionRejectsLandingASpecTwice(t *testing.T) {
	p := missionPlan()
	p.Beats[3].Mission = &MissionBeat{Show: "spec", At: 0}
	err := validateMissionPlan(p)
	if err == nil {
		t.Fatal("a checklist row read twice was accepted")
	}
	if !strings.Contains(err.Error(), "padding") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissionRejectsASpecOffTheList(t *testing.T) {
	p := missionPlan()
	p.Beats[1].Mission = &MissionBeat{Show: "spec", At: 9}
	if err := validateMissionPlan(p); err == nil {
		t.Fatal("a beat landing a spec that does not exist was accepted")
	}
}

func TestMissionRejectsTwoDeliverables(t *testing.T) {
	p := missionPlan()
	p.Beats[3].Mission = &MissionBeat{Show: "deliverable"}
	err := validateMissionPlan(p)
	if err == nil {
		t.Fatal("two deliverable beats were accepted, and that is the same chip landing twice")
	}
	if !strings.Contains(err.Error(), "want exactly 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissionRejectsNoDeliverableBeat(t *testing.T) {
	p := missionPlan()
	p.Mission.Specs = append(p.Mission.Specs, "exit with a useful status code")
	p.Beats[4].Mission = &MissionBeat{Show: "spec", At: 3}
	err := validateMissionPlan(p)
	if err == nil {
		t.Fatal("a brief that never names the artifact was accepted")
	}
	if !strings.Contains(err.Error(), "shipping nothing in particular") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissionRejectsTwoDoneStamps(t *testing.T) {
	p := missionPlan()
	p.Beats = append(p.Beats, SnippetBeat{
		ID: "done-again", Heading: "Done again", Narration: msNarration,
		Mission: &MissionBeat{Show: "done"},
	})
	err := validateMissionPlan(p)
	if err == nil {
		t.Fatal("two done beats were accepted, and a mission that is done twice was never clearly done")
	}
	if !strings.Contains(err.Error(), "want exactly 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissionRequiresDoneToBeTheLastBeat(t *testing.T) {
	p := missionPlan()
	p.Beats[4].Mission = &MissionBeat{Show: "done"}
	p.Beats[5].Mission = &MissionBeat{Show: "deliverable"}
	err := validateMissionPlan(p)
	if err == nil {
		t.Fatal("a clip that stamps done and keeps going was accepted")
	}
	if !strings.Contains(err.Error(), "does not exist yet") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissionRejectsASpecNeverLanded(t *testing.T) {
	p := missionPlan()
	p.Mission.Specs = append(p.Mission.Specs, "exit with a useful status code")
	err := validateMissionPlan(p)
	if err == nil {
		t.Fatal("a checklist row with no beat was accepted, and the builder discovers it at the end")
	}
	if !strings.Contains(err.Error(), "never landed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissionNormalizeClampsAndCapsTheCard(t *testing.T) {
	p := missionPlan()
	p.Mission.Goal = "list out every single drive attached to this machine along with how big it is"
	p.Mission.Specs[0] = "find each and every drive attached to the running system"
	p.Mission.Specs = append(p.Mission.Specs, "  ", "a fourth", "a fifth", "a sixth", "a seventh")
	p.Mission.Deliverable = "a small command line script that anybody could run"
	p.Mission.Done = "running it prints exactly one line for every drive that the machine can see"
	normalizeMissionPlan(p)
	if n := len(strings.Fields(p.Mission.Goal)); n != maxMissionGoalWords {
		t.Fatalf("the goal survived at %d words", n)
	}
	if n := len(strings.Fields(p.Mission.Specs[0])); n != maxMissionSpecWords {
		t.Fatalf("a spec survived at %d words", n)
	}
	if n := len(strings.Fields(p.Mission.Deliverable)); n != maxMissionDeliverableWords {
		t.Fatalf("the deliverable survived at %d words", n)
	}
	if n := len(strings.Fields(p.Mission.Done)); n != maxMissionDoneWords {
		t.Fatalf("the done line survived at %d words", n)
	}
	if n := len(p.Mission.Specs); n != maxMissionSpecs {
		t.Fatalf("want %d specs after normalize, got %d", maxMissionSpecs, n)
	}
	for _, s := range p.Mission.Specs {
		if strings.TrimSpace(s) == "" {
			t.Fatal("an empty checklist row survived normalize")
		}
	}
}

func TestMissionNormalizeClampsASpecOffTheList(t *testing.T) {
	p := missionPlan()
	p.Beats[1].Mission.At = 99
	p.Beats[5].Mission.At = 2
	normalizeMissionPlan(p)
	if at := p.Beats[1].Mission.At; at != len(p.Mission.Specs)-1 {
		t.Fatalf("want the spec clamped to the last row, got %d", at)
	}
	// The done stamp does not index anything, so an index on it is noise.
	if at := p.Beats[5].Mission.At; at != 0 {
		t.Fatalf("the done beat kept its index %d", at)
	}
}

func TestMissionShowDefaultsToSpec(t *testing.T) {
	b := MissionBeat{Show: "ship"}
	if got := b.ResolvedShow(); got != "spec" {
		t.Fatalf("an unknown show resolved to %q, want spec", got)
	}
	b = MissionBeat{Show: " DONE "}
	if got := b.ResolvedShow(); got != "done" {
		t.Fatalf("a shouted done resolved to %q", got)
	}
}

// The card builds downward and nothing is un-built: the closer carries every
// landed row plus both stamps.
func TestMissionScenesBuildTheCardDownward(t *testing.T) {
	p := missionPlan()
	scenes, err := missionScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if len(steps) != len(p.Beats) {
		t.Fatalf("want %d steps, got %d", len(p.Beats), len(steps))
	}

	first := steps[0]
	landed, _ := first["landed"].([]int)
	if first["show"] != "brief" || len(landed) != 0 {
		t.Fatalf("the opening brief already has checklist rows: %v", first)
	}
	if first["deliverableOn"] != false || first["doneOn"] != false {
		t.Fatalf("the brief beat has already stamped the card: %v", first)
	}

	last := steps[len(steps)-1]
	rows, _ := last["landed"].([]int)
	if last["show"] != "done" || len(rows) != len(p.Mission.Specs) {
		t.Fatalf("the done stamp does not carry every landed row: %v", last)
	}
	for i, at := range rows {
		if at != i {
			t.Fatalf("the landed set is not sorted or complete: %v", rows)
		}
	}
	if last["deliverableOn"] != true || last["doneOn"] != true {
		t.Fatalf("the closer is missing a stamp: %v", last)
	}
}
