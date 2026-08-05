package pipeline

// The checkpoint template: do this now, and here is how you know it worked.
//
// This closes the loop `objective` opens, and the two share a vocabulary on
// purpose — a checkpoint's task has to be the same KIND of thing a lesson
// promised, or the lesson and its proof are about different subjects. Using
// objectiveVerbs here is not reuse for tidiness; it is the mechanism that keeps
// a course's promise and its evidence in the same language.
//
// Two rules earn it its place.
//
// **The task has a done-when, and it is checkable by the learner alone.** "You
// should now feel comfortable with retries" is not a checkpoint. Neither is
// anything requiring a grader. The viewer is on their own at the end of a video,
// so the condition has to be something they can look at and settle: a command
// that exits zero, a number under a threshold, a page that loads.
//
// **The task is the size of the lesson, not the size of the subject.** A
// checkpoint that asks for a production-grade retry layer after an eight-minute
// lesson is a way of telling people they failed. The step ceiling is the blunt
// instrument for this, and it is deliberately low.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "checkpoint",
		Category:    CatPresenting,
		Since:       SinceV5,
		Title:       "Prove you can do it",
		Description: "The task that closes a lesson: a few steps the viewer does now, and the condition that tells them it worked without anyone marking it. Reach for it as the last clip of a lesson.",
		Example:     "Close a lesson on retries by having them make a flaky call succeed and then give up",
		PromptFile:  snippetCheckpointTemplateName,
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		MaxBeats:         8,
		Owns:             beatFields{Checkpoint: true},
		OwnsPlan:         planFields{Checkpoint: true},
		Normalize:        normalizeCheckpointPlan,
		Validate:         validateCheckpointPlan,
		Scenes:           checkpointScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				// The emphasis roles every headline picks from.
				"Roles": strings.Join(MetricRoles(), ", "),
				// The same list objective uses. See the file comment.
				"Verbs":         strings.Join(ObjectiveVerbs(), ", "),
				"BeliefVerbs":   strings.Join(objectiveBeliefVerbList(), ", "),
				"MinSteps":      minCheckpointSteps,
				"MaxSteps":      maxCheckpointSteps,
				"MaxTaskWords":  maxCheckpointTaskWords,
				"MaxStepWords":  maxCheckpointStepWords,
				"MaxDoneWords":  maxCheckpointDoneWords,
				"MaxStuckWords": maxCheckpointStuckWords,
			}
		},
	})
}

const snippetCheckpointTemplateName = "snippet_checkpoint.tmpl"

const (
	// Two steps is a task; one is an instruction. Four is where a checkpoint
	// stops being the size of a lesson and starts being homework.
	minCheckpointSteps = 2
	maxCheckpointSteps = 4

	maxCheckpointTaskWords  = 12
	maxCheckpointStepWords  = 12
	maxCheckpointDoneWords  = 14
	maxCheckpointStuckWords = 16
)

// CheckpointSpec is the task that proves the lesson landed.
type CheckpointSpec struct {
	// Task is what to do, opening with an observable verb.
	Task string `json:"task"`
	// Steps are the moves, in order.
	Steps []string `json:"steps"`
	// Done is the condition that says it worked — checkable by the learner,
	// alone, without a grader.
	Done string `json:"done"`
	// Stuck is the one thing to check first when it does not work. Optional,
	// and the difference between a checkpoint and a cliff.
	Stuck string `json:"stuck,omitempty"`
}

// Verb returns the task's opening word, lowercased and stripped.
func (c CheckpointSpec) Verb() string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(c.Task)))
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(fields[0], ".,:;\"'")
}

// CheckpointBeat says which part of the task this beat is on.
type CheckpointBeat struct {
	At   int    `json:"at,omitempty"`
	Show string `json:"show"`
}

var checkpointShows = map[string]bool{
	// The task, stated.
	"task": true,
	// One step of it.
	"step": true,
	// The condition that settles it.
	"done": true,
	// What to check when it does not work.
	"stuck": true,
}

// CheckpointShows returns the beat vocabulary, sorted, for the prompt.
func CheckpointShows() []string {
	out := make([]string, 0, len(checkpointShows))
	for s := range checkpointShows {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// ResolvedShow defaults to a step.
func (b CheckpointBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if checkpointShows[s] {
		return s
	}
	return "step"
}

func normalizeCheckpointPlan(p *SnippetPlan) {
	if p.Checkpoint == nil {
		return
	}
	c := p.Checkpoint
	c.Task = strings.TrimSpace(c.Task)
	for _, lead := range []string{"you should ", "you will ", "now ", "try to ", "go and "} {
		if len(c.Task) > len(lead) && strings.EqualFold(c.Task[:len(lead)], lead) {
			c.Task = strings.TrimSpace(c.Task[len(lead):])
			break
		}
	}
	c.Done = strings.TrimSpace(c.Done)
	c.Stuck = strings.TrimSpace(c.Stuck)
	for i := range c.Steps {
		c.Steps[i] = strings.TrimSpace(c.Steps[i])
	}
	for i := range p.Beats {
		if b := p.Beats[i].Checkpoint; b != nil {
			b.Show = b.ResolvedShow()
			if b.At < 0 || b.At >= len(c.Steps) {
				b.At = 0
			}
		}
	}
}

func validateCheckpointPlan(p *SnippetPlan) error {
	c := p.Checkpoint
	if c == nil {
		return fmt.Errorf("the plan has no checkpoint — this template closes a lesson with a task, so the task is the clip")
	}
	verb := c.Verb()
	if verb == "" {
		return fmt.Errorf("the checkpoint has no task")
	}
	for belief, why := range objectiveBeliefVerbs {
		if verb == belief || strings.HasPrefix(verb, belief) {
			return fmt.Errorf("the task opens with %q: %s. A checkpoint is the PROOF a lesson landed, so it has to be something the viewer does and can watch themselves finish. Rewrite it around one of: %s",
				verb, why, strings.Join(ObjectiveVerbs(), ", "))
		}
	}
	if !objectiveVerbs[verb] {
		return fmt.Errorf("the task opens with %q, which is not in the vocabulary this batch shares with `objective`. Use one of: %s. The two templates share a verb list on purpose — a lesson's promise and its proof have to be the same kind of thing, or they are about different subjects",
			verb, strings.Join(ObjectiveVerbs(), ", "))
	}
	if n := len(strings.Fields(c.Task)); n > maxCheckpointTaskWords {
		return fmt.Errorf("the task is %d words; keep it to %d", n, maxCheckpointTaskWords)
	}
	if n := len(c.Steps); n < minCheckpointSteps || n > maxCheckpointSteps {
		return fmt.Errorf("the task has %d step(s); this template takes %d to %d. One step is an instruction rather than a task, and past %d it is homework — a checkpoint is the size of the lesson, not the size of the subject",
			n, minCheckpointSteps, maxCheckpointSteps, maxCheckpointSteps)
	}
	for i, s := range c.Steps {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("step %d is empty", i+1)
		}
		if n := len(strings.Fields(s)); n > maxCheckpointStepWords {
			return fmt.Errorf("step %d is %d words; keep it to %d", i+1, n, maxCheckpointStepWords)
		}
	}
	if strings.TrimSpace(c.Done) == "" {
		return fmt.Errorf("the checkpoint has no done-when. The viewer is alone at the end of a video with nobody to mark their work, so the clip has to hand them the condition: a command that exits zero, a number under a threshold, a page that loads. Without it this is an exercise, and an exercise nobody can grade is a suggestion")
	}
	if n := len(strings.Fields(c.Done)); n > maxCheckpointDoneWords {
		return fmt.Errorf("the done-when is %d words; keep it to %d — it is a condition to check, not a procedure", n, maxCheckpointDoneWords)
	}
	// A done-when that describes a feeling is the failure the field exists to
	// prevent, and it survives the length check comfortably.
	for _, soft := range []string{"feel", "comfortable", "confident", "understand", "familiar", "make sense", "makes sense"} {
		if strings.Contains(strings.ToLower(c.Done), soft) {
			return fmt.Errorf("the done-when (%q) describes a feeling, not a condition. Nobody can check whether they feel ready. Give the viewer something to LOOK at: what does the terminal print, what does the page show, what number comes out",
				c.Done)
		}
	}
	if n := len(strings.Fields(c.Stuck)); n > maxCheckpointStuckWords {
		return fmt.Errorf("the stuck hint is %d words; keep it to %d", n, maxCheckpointStuckWords)
	}

	shown := map[string]bool{}
	walked := map[int]bool{}
	for _, b := range p.Beats {
		if b.Checkpoint == nil {
			return fmt.Errorf("beat %q has no checkpoint direction — say whether it states the task, walks a step, gives the done-when, or handles being stuck", b.ID)
		}
		show := b.Checkpoint.ResolvedShow()
		shown[show] = true
		if show == "step" {
			walked[b.Checkpoint.At] = true
		}
	}
	for _, need := range []string{"task", "done"} {
		if !shown[need] {
			return fmt.Errorf("no beat gives the %s. The task and the done-when are the two halves of a checkpoint; a clip missing either is not one", need)
		}
	}
	for i, s := range c.Steps {
		if !walked[i] {
			return fmt.Errorf("step %d (%q) is never walked by a beat", i+1, s)
		}
	}
	return nil
}

func checkpointScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.Checkpoint
	if c == nil {
		return nil, fmt.Errorf("the plan has no checkpoint")
	}

	done := map[int]bool{}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Checkpoint == nil {
			return nil, fmt.Errorf("beat %q has no checkpoint direction", beat.ID)
		}
		show := beat.Checkpoint.ResolvedShow()
		if show == "step" {
			done[beat.Checkpoint.At] = true
		}
		ticked := make([]int, 0, len(done))
		for at := range done {
			ticked = append(ticked, at)
		}
		sort.Ints(ticked)

		step := map[string]any{"startMs": startMs, "endMs": endMs, "show": show, "ticked": ticked}
		if show == "step" {
			step["at"] = beat.Checkpoint.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneCheckpoint,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title": in.Plan.Title,
			"task":  c.Task,
			"list":  c.Steps,
			"done":  c.Done,
			"stuck": c.Stuck,
			"steps": steps,
		}),
	}}, nil
}
