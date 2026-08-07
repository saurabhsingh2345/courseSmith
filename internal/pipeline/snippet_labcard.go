package pipeline

// The labcard template: the briefing a learner keeps open while their hands
// are busy.
//
// Every foundations course eventually stops explaining and says "now go and do
// it" — install VirtualBox, boot Ubuntu, compile hello.c, watch the process
// table. That moment is where recorded courses lose people, and they lose them
// for a boring reason: the instructions were spoken. A learner with a terminal
// in one window and a video in the other cannot rewind a sentence they half
// heard while typing. So this template is not a narration with a list under it.
// It is a CARD — a task, the tools that must already be installed, the numbered
// steps, and the one line that says what success looks like — held on screen
// for the whole clip while the voice walks it.
//
// The validator's job is to protect the card from becoming prose. Steps are
// capped at ten words because a step that needs a sentence is two steps, and a
// step list that wraps stops being scannable at a glance, which is the only
// thing it is for. Tools are capped at four because a lab that needs five
// separate installs is a lab that will be abandoned at the third.
//
// The one hard structural rule is that every step is lit exactly once and IN
// ORDER. A model asked for a lab will happily narrate steps two and four and
// leave three to the viewer's imagination, or double back to step one to "recap"
// — and a numbered list whose highlight jumps around teaches the viewer that the
// numbers are decoration. They are not: they are the order the commands must be
// typed in, and typing them out of order is how a lab fails silently.
//
// The clip opens on the task because a learner who does not yet know what they
// are building cannot evaluate a tool list, and it closes on the expected result
// because "how do I know it worked" is the question the learner actually has.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "labcard",
		Category:    CatCode,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The lab briefing",
		Description: "A mission card for hands-on lab time: the task, the tools that must be installed, the numbered steps, and what the screen looks like when it worked. Reach for it when the viewer is about to stop watching and start typing.",
		Example:     "Lab: install VirtualBox and boot Ubuntu",
		PromptFile:  snippetLabCardTemplateName,
		NeedsCode:   false,
		// The card has to be read, not glanced at: task, tools, three to six
		// steps and a result line. Under thirty seconds the step highlights
		// arrive faster than a learner can look away and look back.
		MinTargetSec:     30,
		DefaultTargetSec: 45,
		// Opener + six steps + closer is eight, and one spare beat for a step
		// that genuinely needs two. Past nine the card is being re-read.
		MaxBeats: 9,
		// No IdealWordsPerBeat: this is a card, not a diagram. A beat here is a
		// step being talked through, which carries the shared forty words
		// comfortably, and forcing it down to a diagram's twenty-eight would
		// turn "run this and here is why" into "run this".
		Owns:      beatFields{LabCard: true},
		OwnsPlan:  planFields{LabCard: true},
		Normalize: normalizeLabCardPlan,
		Validate:  validateLabCardPlan,
		Scenes:    labCardScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":          strings.Join(MetricRoles(), ", "),
				"Shows":          strings.Join(LabCardShows(), ", "),
				"MaxTaskWords":   maxLabCardTaskWords,
				"MinTools":       minLabCardTools,
				"MaxTools":       maxLabCardTools,
				"MaxToolWords":   maxLabCardToolWords,
				"MinSteps":       minLabCardSteps,
				"MaxSteps":       maxLabCardSteps,
				"MaxStepWords":   maxLabCardStepWords,
				"MaxExpectWords": maxLabCardExpectWords,
			}
		},
	})
}

const snippetLabCardTemplateName = "snippet_labcard.tmpl"

const (
	// The task is a headline for the hands, not a paragraph: "install
	// VirtualBox and boot Ubuntu" is nine words and already at the edge of
	// what reads at 60px across the briefing zone.
	maxLabCardTaskWords = 10

	// A lab with no prerequisites has nothing to check before starting, and
	// the tool row would be an empty box on screen.
	minLabCardTools = 1
	// Five separate installs is a lab that gets abandoned at the third, and
	// four mono chips is what the briefing zone holds on one line.
	maxLabCardTools = 4
	// A tool is a name — "VirtualBox", "Ubuntu 24.04 ISO" — so three words is
	// generous. Longer means a description crept into the chip.
	maxLabCardToolWords = 3

	// Two steps is an instruction, not a procedure; the numbered list only
	// earns its place once there is an order to get wrong.
	minLabCardSteps = 3
	// Six rows at a size that reads across the room fills the step zone, and
	// a seven-step lab is two labs.
	maxLabCardSteps = 6
	// A step is a command or an action. Past ten words it is a sentence, it
	// wraps, and the list stops being scannable, which is the only thing a
	// numbered list is for.
	maxLabCardStepWords = 10

	// The result line is a terminal echo — "a purple Ubuntu desktop, no
	// errors" — so twelve words is a full one plus its qualifier.
	maxLabCardExpectWords = 12
)

// labCardShows is the closed vocabulary of what a beat does.
var labCardShows = map[string]bool{
	// The task and the tool chips. The opener.
	"task": true,
	// The step at index At lights on the numbered list.
	"step": true,
	// The expected-result strip lands in terminal styling. The closer.
	"expect": true,
}

// LabCardShows returns the beat vocabulary sorted.
func LabCardShows() []string {
	out := make([]string, 0, len(labCardShows))
	for k := range labCardShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// LabCardSpec is the briefing itself. On the plan because the card is up for
// the whole clip — the beats light parts of it, they do not replace it.
type LabCardSpec struct {
	// Task is what the learner is about to build or make happen.
	Task string `json:"task"`
	// Tools are what must already be installed before step one.
	Tools []string `json:"tools"`
	// Steps are the numbered actions, in the order they must be done.
	Steps []string `json:"steps"`
	// Expect is what the learner sees when it worked.
	Expect string `json:"expect"`
}

// LabCardBeat is one shot: which part of the card this beat is on.
type LabCardBeat struct {
	// Show is a labCardShows name.
	Show string `json:"show"`
	// At is the zero-based step index, used only when Show is "step".
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a step —
// the state most beats of this template are in.
func (b LabCardBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if labCardShows[s] {
		return s
	}
	return "step"
}

func normalizeLabCardPlan(p *SnippetPlan) {
	c := p.LabCard
	if c == nil {
		return
	}
	c.Task = clampWords(collapseSpaces(c.Task), maxLabCardTaskWords)
	c.Expect = clampWords(collapseSpaces(c.Expect), maxLabCardExpectWords)

	tools := make([]string, 0, len(c.Tools))
	for _, t := range c.Tools {
		t = clampWords(collapseSpaces(t), maxLabCardToolWords)
		if t != "" {
			tools = append(tools, t)
		}
	}
	if len(tools) > maxLabCardTools {
		tools = tools[:maxLabCardTools]
	}
	c.Tools = tools

	steps := make([]string, 0, len(c.Steps))
	for _, s := range c.Steps {
		s = clampWords(collapseSpaces(s), maxLabCardStepWords)
		if s != "" {
			steps = append(steps, s)
		}
	}
	if len(steps) > maxLabCardSteps {
		steps = steps[:maxLabCardSteps]
	}
	c.Steps = steps

	for i := range p.Beats {
		b := p.Beats[i].LabCard
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		// Clamp rather than drop: an index one past the end is a model that
		// counted from one, and the beat still means "the last step".
		if b.At < 0 {
			b.At = 0
		}
		if n := len(c.Steps); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateLabCardPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{LabCard: true}); err != nil {
		return err
	}

	c := p.LabCard
	if c == nil {
		return fmt.Errorf("the plan has no lab card — this template IS the card, so without a task, tools, steps and an expected result there is nothing on screen for the voice to walk")
	}
	if strings.TrimSpace(c.Task) == "" {
		return fmt.Errorf("the card has no task. The task is the first thing the learner reads and the thing they judge every tool and step against — say what they are about to build in at most %d words", maxLabCardTaskWords)
	}
	if n := len(strings.Fields(c.Task)); n > maxLabCardTaskWords {
		return fmt.Errorf("the task %q is %d words, and the briefing zone holds %d at the size it is set. Cut it to the outcome and let the steps carry the detail", c.Task, n, maxLabCardTaskWords)
	}
	if n := len(c.Tools); n < minLabCardTools || n > maxLabCardTools {
		return fmt.Errorf("the card lists %d tools, want %d-%d. A lab with no prerequisites leaves an empty box on screen, and one that needs more than %d separate installs is a lab that gets abandoned at the third",
			n, minLabCardTools, maxLabCardTools, maxLabCardTools)
	}
	for i, t := range c.Tools {
		if strings.TrimSpace(t) == "" {
			return fmt.Errorf("tool %d is blank. Each chip names one thing that must already be installed", i+1)
		}
		if n := len(strings.Fields(t)); n > maxLabCardToolWords {
			return fmt.Errorf("the tool %q is %d words, and a tool chip holds %d. A chip names a thing — \"VirtualBox\", \"Ubuntu ISO\" — the reason it is needed belongs in the narration", t, n, maxLabCardToolWords)
		}
	}
	if n := len(c.Steps); n < minLabCardSteps || n > maxLabCardSteps {
		return fmt.Errorf("the card lists %d steps, want %d-%d. Under %d there is no order to get wrong and the numbered list is decoration; over %d the rows stop reading at the size they are set",
			n, minLabCardSteps, maxLabCardSteps, minLabCardSteps, maxLabCardSteps)
	}
	for i, s := range c.Steps {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("step %d is blank. Every numbered row is one action the learner takes", i+1)
		}
		if n := len(strings.Fields(s)); n > maxLabCardStepWords {
			return fmt.Errorf("step %d, %q, is %d words and the limit is %d. A step that needs a sentence is two steps, and a row that wraps stops being scannable, which is the only thing a numbered list is for",
				i+1, s, n, maxLabCardStepWords)
		}
	}
	if strings.TrimSpace(c.Expect) == "" {
		return fmt.Errorf("the card says nothing about what success looks like. \"How do I know it worked\" is the question the learner actually has, and the closing strip is where it gets answered — give an expect line of at most %d words", maxLabCardExpectWords)
	}
	if n := len(strings.Fields(c.Expect)); n > maxLabCardExpectWords {
		return fmt.Errorf("the expected result %q is %d words, and the terminal strip holds %d. Describe the screen, not the reasoning", c.Expect, n, maxLabCardExpectWords)
	}

	for _, b := range p.Beats {
		if b.LabCard == nil {
			return fmt.Errorf("beat %q has no labcard direction — every beat is on one part of the card, so every beat needs a {\"show\": ...}", b.ID)
		}
	}
	if p.Beats[0].LabCard.ResolvedShow() != "task" {
		return fmt.Errorf("beat %q does not open on the task. A learner who does not yet know what they are building cannot judge a tool list or a step — open with {\"show\": \"task\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.LabCard.ResolvedShow() != "expect" {
		return fmt.Errorf("the clip does not close on the expected result. The last frame is the one the learner leaves on screen while they type, and it has to say what success looks like — end with {\"show\": \"expect\"}")
	}

	// THE ORDER. A numbered list whose highlight jumps around teaches the
	// viewer that the numbers are decoration; they are the order the commands
	// have to be typed in.
	var lit []int
	for i, b := range p.Beats {
		show := b.LabCard.ResolvedShow()
		switch show {
		case "task":
			if i != 0 {
				return fmt.Errorf("beat %q goes back to the task part-way through. The task stays on the card the whole time, so there is nothing to return to — this beat is either a step or the expect closer", b.ID)
			}
		case "expect":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q shows the expected result before the end. \"expect\" is the closer, and once the learner has been told what success looks like there is no reason to walk another step", b.ID)
			}
		case "step":
			at := b.LabCard.At
			if at < 0 || at >= len(c.Steps) {
				return fmt.Errorf("beat %q lights step %d, and the card has %d steps numbered 0 to %d. The index is zero-based",
					b.ID, at, len(c.Steps), len(c.Steps)-1)
			}
			if len(lit) > 0 && at <= lit[len(lit)-1] {
				return fmt.Errorf("beat %q lights step %d after step %d. The steps are walked in order and each is lit exactly once — a highlight that doubles back tells the viewer the numbers are decoration, and they are the order the commands have to be typed in",
					b.ID, at, lit[len(lit)-1])
			}
			if len(lit) == 0 && at != 0 {
				return fmt.Errorf("beat %q starts the walk at step %d, skipping step 0, %q. Start at the first step", b.ID, at, c.Steps[0])
			}
			if len(lit) > 0 && at != lit[len(lit)-1]+1 {
				return fmt.Errorf("beat %q jumps to step %d, skipping step %d, %q. Every step gets its own beat — a step nobody says out loud is a step the learner does not do",
					b.ID, at, lit[len(lit)-1]+1, c.Steps[lit[len(lit)-1]+1])
			}
			lit = append(lit, at)
		}
	}
	if len(lit) != len(c.Steps) {
		missing := len(c.Steps) - len(lit)
		return fmt.Errorf("the clip lights %d of the card's %d steps, leaving %d never walked — starting with step %d, %q. Either give every step a beat or cut the steps the clip does not cover",
			len(lit), len(c.Steps), missing, len(lit), c.Steps[len(lit)])
	}
	return nil
}

// labCardScenes lays the clip out as ONE scene. The card is static; the steps
// array carries which row is lit at each moment and which rows are already
// behind the learner, both computed here so the component only draws.
func labCardScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.LabCard
	if c == nil {
		return nil, fmt.Errorf("the plan has no lab card")
	}

	stepList := make([]map[string]any, len(c.Steps))
	for i, s := range c.Steps {
		stepList[i] = map[string]any{
			"n":    i + 1,
			"text": s,
		}
	}
	tools := make([]map[string]any, len(c.Tools))
	for i, t := range c.Tools {
		tools[i] = map[string]any{"name": t}
	}

	done := map[int]bool{}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.LabCard == nil {
			return nil, fmt.Errorf("beat %q has no labcard direction", beat.ID)
		}
		show := beat.LabCard.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "step" {
			at := beat.LabCard.At
			if at >= 0 && at < len(c.Steps) {
				done[at] = true
				step["at"] = at
			}
		}
		reached := make([]int, 0, len(done))
		for k := range done {
			reached = append(reached, k)
		}
		sort.Ints(reached)
		step["reached"] = reached
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneLabCard,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":    in.Plan.Title,
			"task":     c.Task,
			"tools":    tools,
			"stepList": stepList,
			"expect":   c.Expect,
			"steps":    steps,
		}),
	}}, nil
}
