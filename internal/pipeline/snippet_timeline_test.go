package pipeline

import (
	"strings"
	"testing"
)

const tlNarration = "The browser has a string at this point and nothing else at all."

func timelinePlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "timeline",
		Title:    "What happens when you press enter",
		Timeline: &TimelineSpec{
			Milestones: []Milestone{
				{Mark: "0ms", Title: "You press enter", Note: "Just a string so far.", Figure: "cursor"},
				{Mark: "2ms", Title: "DNS lookup", Note: "The name becomes an address.", Figure: "search"},
				{Mark: "30ms", Title: "TCP and TLS", Note: "A connection opens.", Figure: "lock"},
				{Mark: "80ms", Title: "The server answers", Note: "Something replies.", Figure: "server"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "press", Heading: "You press enter", Narration: tlNarration, Timeline: &TimelineBeat{At: 0}},
			{ID: "dns", Heading: "Finding the address", Narration: tlNarration, Timeline: &TimelineBeat{At: 1}},
			{ID: "shake", Heading: "Opening the pipe", Narration: tlNarration, Timeline: &TimelineBeat{At: 2}},
			{ID: "answer", Heading: "The server replies", Narration: tlNarration, Timeline: &TimelineBeat{At: 3}},
			{ID: "whole", Heading: "All of it", Narration: tlNarration, Timeline: &TimelineBeat{Whole: true}},
		},
	}
}

func TestTimelinePlanAccepted(t *testing.T) {
	if err := validateTimelinePlan(timelinePlan()); err != nil {
		t.Fatalf("a well-formed timeline was rejected: %v", err)
	}
}

// The rule the template rests on. A clip that walks back is not narrating a
// timeline — it is narrating a diagram, and the error says so, because the fix
// is to use a different template rather than to reorder the beats.
func TestTimelineOnlyMovesForward(t *testing.T) {
	p := timelinePlan()
	p.Beats[2].Timeline = &TimelineBeat{At: 3}
	p.Beats[3].Timeline = &TimelineBeat{At: 2}
	err := validateTimelinePlan(p)
	if err == nil {
		t.Fatal("a timeline that walks backwards was accepted")
	}
	if !strings.Contains(err.Error(), "only moves forward") {
		t.Errorf("the error should name the rule; got: %v", err)
	}
	if !strings.Contains(err.Error(), "diagram") {
		t.Errorf("the error should point at the template that would tell it properly; got: %v", err)
	}
}

func TestTimelineReachesEveryMilestone(t *testing.T) {
	p := timelinePlan()
	p.Beats = append(p.Beats[:3], p.Beats[4]) // "The server answers" never reached
	err := validateTimelinePlan(p)
	if err == nil {
		t.Fatal("a milestone with no beat was accepted")
	}
	if !strings.Contains(err.Error(), "never reached") {
		t.Errorf("the error should name the unreached stop; got: %v", err)
	}
}

func TestTimelineVisitsEachStopOnce(t *testing.T) {
	p := timelinePlan()
	p.Beats[3].Timeline = &TimelineBeat{At: 2}
	if err := validateTimelinePlan(p); err == nil {
		t.Error("standing on the same milestone twice was accepted")
	}
}

// The finished run is what the viewer leaves with; ending on the last milestone
// alone leaves the line looking mid-walk.
func TestTimelineMustEndOnTheWholeRun(t *testing.T) {
	p := timelinePlan()
	p.Beats = p.Beats[:4]
	err := validateTimelinePlan(p)
	if err == nil {
		t.Fatal("a timeline that never shows the finished line was accepted")
	}
	if !strings.Contains(err.Error(), "finished line") {
		t.Errorf("the error should ask for the closing beat; got: %v", err)
	}

	// Present but not last is still wrong: the clip ends mid-walk.
	p = timelinePlan()
	p.Beats[0], p.Beats[4] = p.Beats[4], p.Beats[0]
	if err := validateTimelinePlan(p); err == nil {
		t.Error("a clip ending on a single milestone was accepted")
	}
}

func TestTimelineMilestoneCountBounds(t *testing.T) {
	p := timelinePlan()
	p.Timeline.Milestones = p.Timeline.Milestones[:2]
	p.Beats = []SnippetBeat{
		{ID: "a", Heading: "h", Narration: tlNarration, Timeline: &TimelineBeat{At: 0}},
		{ID: "b", Heading: "h", Narration: tlNarration, Timeline: &TimelineBeat{At: 1}},
		{ID: "w", Heading: "h", Narration: tlNarration, Timeline: &TimelineBeat{Whole: true}},
	}
	err := validateTimelinePlan(p)
	if err == nil {
		t.Fatal("a two-milestone timeline was accepted")
	}
	// Two points is a before-and-after, and the catalog already has a template
	// that does that better — the error should say which.
	if !strings.Contains(err.Error(), "compare") {
		t.Errorf("the error should point at the compare template; got: %v", err)
	}
}

func TestTimelineRejectsRepeatedTitles(t *testing.T) {
	p := timelinePlan()
	p.Timeline.Milestones[2].Title = "dns LOOKUP"
	if err := validateTimelinePlan(p); err == nil {
		t.Error("two milestones with the same title were accepted")
	}
}

func TestNormalizeTimelineClampsAndFixesFigures(t *testing.T) {
	p := timelinePlan()
	p.Timeline.Milestones[0].Mark = "  zero   milliseconds  exactly  now  "
	p.Timeline.Milestones[0].Figure = "  NOT-A-FIGURE "
	p.Beats[0].Timeline = &TimelineBeat{At: -1}
	normalizeTimelinePlan(p)

	if n := len(strings.Fields(p.Timeline.Milestones[0].Mark)); n > maxMilestoneMarkWords {
		t.Errorf("mark was not clamped: %d words", n)
	}
	if p.Timeline.Milestones[0].Figure != "spark" {
		t.Errorf("figure = %q, want the spark fallback", p.Timeline.Milestones[0].Figure)
	}
	// A negative index is how a model says "none of them"; the template spells
	// that `whole`.
	if !p.Beats[0].Timeline.Whole {
		t.Error("a negative `at` should become whole")
	}
}

func TestTimelineScenesIsOneSceneWithSteps(t *testing.T) {
	plan := timelinePlan()
	scenes, err := timelineScenes(sceneInput(t, plan, 5000))
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneTimeline {
		t.Fatalf("want one timeline scene, got %d", len(scenes))
	}
	ms, ok := scenes[0].Props["milestones"].([]map[string]any)
	if !ok || len(ms) != 4 {
		t.Fatalf("want four milestones, got %#v", scenes[0].Props["milestones"])
	}
	// Figures reach the renderer normalized: figureFor otherwise falls back to
	// a burst that has nothing to do with the stop.
	if ms[0]["figure"] != "cursor" {
		t.Errorf("milestone 0 figure = %v, want cursor", ms[0]["figure"])
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if len(steps) != len(plan.Beats) {
		t.Fatalf("want one step per beat, got %d", len(steps))
	}
	if steps[len(steps)-1]["whole"] != true {
		t.Error("the closing step should be marked whole")
	}
	if _, present := steps[len(steps)-1]["at"]; present {
		t.Error("a whole step carries a milestone index it does not mean")
	}
}
