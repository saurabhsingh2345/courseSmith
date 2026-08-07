package pipeline

// The stepper template: watch the algorithm think.
//
// An algorithm is a thing that happens over time to a piece of data, and the
// standard way of teaching one — pseudocode on a slide — removes both the time
// and the data. The viewer is shown the rule and asked to run it in their head,
// which is exactly the skill they do not have yet. That is why binary search
// feels obvious in the lecture and impossible in the exercise.
//
// So this template puts the data on screen and runs the rule ON it. Eight cells
// at most, pointer flags standing over them, two cells pulsing when they are
// compared, two tiles physically crossing when they swap, and a counter in the
// corner ticking every time the algorithm does work. Nothing is described; the
// array is simply in a different state than it was a second ago, and the
// narration says why.
//
// Which means the array state has to be RIGHT, and that is what this file's
// validators are for. The array is tracked in Go across every swap, so the
// component is handed the contents of every cell at every moment and never
// simulates anything. On top of that the mechanics are checked: a swap names
// exactly two cells because a swap of one cell is not a swap and a swap of
// three is not a thing; every pointer a beat moves must be a pointer the plan
// declared, because a flag appearing over a cell with a name nobody introduced
// is noise; and a "found" beat has to name a cell that ACTUALLY holds the
// target in the tracked array at that instant. That last one is the family's
// signature rule. A model writing a binary search from memory will confidently
// announce the answer at the wrong index, and a search animation that lands on
// the wrong cell has taught the viewer that the picture is decoration.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "stepper",
		Category:    CatConcepts,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "Watch the algorithm think",
		Description: "An array of cells with pointer flags over them, comparisons pulsing, swaps arcing across and an operation counter ticking. Reach for it when the subject is an algorithm's mechanics — how a search narrows, how a sort moves things, what a pointer is actually doing.",
		Example:     "Binary search finds 42 in eight cells in three steps",
		PromptFile:  snippetStepperTemplateName,
		NeedsCode:   false,
		// The array, the pointers, and at least two rounds of compare-and-move:
		// under thirty-five seconds the tiles change faster than a viewer can
		// re-read the row, which is the one thing this template must not do.
		MinTargetSec: 35,
		// The beat count is a property of the ALGORITHM — a three-step search
		// is three compares whether or not the budget wants them — so the
		// default runtime has to fund the steps the prompt asks for.
		DefaultTargetSec: 55,
		// Ten, higher than any other template in this family, because a run is
		// a sequence rather than a set of shots: eight beats of stepping is
		// still one continuous picture, not the same picture eight times.
		MaxBeats: 10,
		// A beat here is a SHOT — one operation on the row — so twenty-eight
		// words, about nine seconds, is how long one comparison holds.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Stepper: true},
		OwnsPlan:          planFields{Stepper: true},
		Normalize:         normalizeStepperPlan,
		Validate:          validateStepperPlan,
		Scenes:            stepperScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":           strings.Join(MetricRoles(), ", "),
				"Shows":           strings.Join(StepperShows(), ", "),
				"MinCells":        minStepperCells,
				"MaxCells":        maxStepperCells,
				"MaxValue":        maxStepperValue,
				"MinPointers":     minStepperPointers,
				"MaxPointers":     maxStepperPointers,
				"MaxPointerWords": maxStepperPointerWords,
			}
		},
	})
}

const snippetStepperTemplateName = "snippet_stepper.tmpl"

const (
	// Three cells is a row a viewer solves by looking, so nothing the
	// algorithm does can be seen to be necessary. Four is the smallest row
	// with a middle to point at.
	minStepperCells = 4
	// Eight tiles fill the drawing box at a size where a three-digit numeral
	// is readable from a phone; a ninth shrinks the type below that.
	maxStepperCells = 8
	// Three digits is the widest numeral the tile holds without the type
	// dropping a size, and it is plenty for a teaching array.
	maxStepperValue = 999

	// A run with no pointer is a row of tiles changing colour with nothing
	// standing over it, which is the picture this template exists to replace.
	minStepperPointers = 1
	// Three flags fit above eight cells without overlapping; low, high and mid
	// is also the most any teaching algorithm actually carries.
	maxStepperPointers = 3
	// A pointer's name is "low" or "left edge", never a phrase.
	maxStepperPointerWords = 2
)

// stepperShows is the closed vocabulary of what a beat does.
var stepperShows = map[string]bool{
	// The row appears with its values. The opener.
	"array": true,
	// One or more pointer flags move to new cells.
	"point": true,
	// The named cells pulse against each other.
	"compare": true,
	// Two named cells trade places, arcing over each other.
	"swap": true,
	// The named cell holds the target and takes the frame.
	"found": true,
	// The run is over and the row rests in its final state.
	"done": true,
}

// StepperShows returns the beat vocabulary sorted.
func StepperShows() []string {
	out := make([]string, 0, len(stepperShows))
	for k := range stepperShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// StepperSpec is the data the algorithm runs on. On the plan because the row
// stands for the whole clip; what changes is which cells are lit and where the
// flags are.
type StepperSpec struct {
	// Values are the cells, left to right, in their starting order.
	Values []int `json:"values"`
	// Pointers are the flag names that will stand over cells — "low", "high",
	// "mid", or "i" and "j".
	Pointers []string `json:"pointers"`
	// Target is the value being searched for, or -1 when the clip is a sort
	// and there is nothing to find.
	Target int `json:"target"`
}

// StepperBeat is one shot: one operation on the row.
type StepperBeat struct {
	// Show is a stepperShows name.
	Show string `json:"show"`
	// At are the cell indices this beat acts on.
	At []int `json:"at,omitempty"`
	// Ptr is where each named pointer stands after this beat. Partial: a
	// pointer left out keeps the cell it was on.
	Ptr map[string]int `json:"ptr,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a compare —
// the workhorse operation, and the one that is safe to draw for a beat whose
// direction did not survive: it lights cells without moving anything.
func (b StepperBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if stepperShows[s] {
		return s
	}
	return "compare"
}

func normalizeStepperPlan(p *SnippetPlan) {
	sp := p.Stepper
	if sp == nil {
		return
	}
	if len(sp.Values) > maxStepperCells {
		sp.Values = sp.Values[:maxStepperCells]
	}
	for i := range sp.Values {
		if sp.Values[i] < 0 {
			sp.Values[i] = 0
		}
		if sp.Values[i] > maxStepperValue {
			sp.Values[i] = maxStepperValue
		}
	}
	// Pointer names are cleaned and de-duplicated here rather than argued
	// about: "low" arriving twice is a copy-paste, not a design.
	names := make([]string, 0, len(sp.Pointers))
	seen := map[string]bool{}
	for _, n := range sp.Pointers {
		n = clampWords(collapseSpaces(n), maxStepperPointerWords)
		if n == "" || seen[strings.ToLower(n)] {
			continue
		}
		seen[strings.ToLower(n)] = true
		names = append(names, n)
		if len(names) == maxStepperPointers {
			break
		}
	}
	sp.Pointers = names
	// -1 is the canonical "no target"; any other negative is the same
	// intention spelled differently.
	if sp.Target < 0 {
		sp.Target = -1
	}

	last := len(sp.Values) - 1
	for i := range p.Beats {
		b := p.Beats[i].Stepper
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		// Cell indices are clamped and de-duplicated: an index off the end is
		// an off-by-one, and the same cell named twice is a slip. The counts
		// the operations need — two for a swap, one or two for a compare — are
		// then checked against what is left, which is the honest order: a
		// "swap" of cells [3, 3] should be rejected as a swap of one cell, not
		// silently drawn as a tile trading places with itself.
		cells := make([]int, 0, len(b.At))
		used := map[int]bool{}
		for _, c := range b.At {
			if c < 0 {
				c = 0
			}
			if last >= 0 && c > last {
				c = last
			}
			if used[c] {
				continue
			}
			used[c] = true
			cells = append(cells, c)
		}
		b.At = cells
	}
}

func validateStepperPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Stepper: true}); err != nil {
		return err
	}

	sp := p.Stepper
	if sp == nil {
		return fmt.Errorf("the plan has no array — this template runs an algorithm ON data, and without the data there is nothing on screen for the rule to happen to")
	}
	if n := len(sp.Values); n < minStepperCells || n > maxStepperCells {
		return fmt.Errorf("the array has %d cells, want %d-%d. Under %d the viewer solves the row by looking at it, so nothing the algorithm does looks necessary, and past %d the numerals shrink below what reads on a phone",
			n, minStepperCells, maxStepperCells, minStepperCells, maxStepperCells)
	}
	for i, v := range sp.Values {
		if v < 0 || v > maxStepperValue {
			return fmt.Errorf("cell %d holds %d, and cells hold 0-%d. A tile is sized for three digits — a wider number drops the whole row a type size and the array stops being readable",
				i, v, maxStepperValue)
		}
	}
	if n := len(sp.Pointers); n < minStepperPointers || n > maxStepperPointers {
		return fmt.Errorf("the plan declares %d pointers, want %d-%d. With none, the row is tiles changing colour with nothing standing over them — which is the picture this template exists to replace — and past %d the flags overlap above the cells",
			n, minStepperPointers, maxStepperPointers, maxStepperPointers)
	}
	declared := map[string]bool{}
	for _, n := range sp.Pointers {
		if strings.TrimSpace(n) == "" {
			return fmt.Errorf("one of the pointers has no name. A flag over a cell says what the algorithm is holding — \"low\", \"mid\", \"i\" — and an unnamed one says only that something is there")
		}
		if w := len(strings.Fields(n)); w > maxStepperPointerWords {
			return fmt.Errorf("the pointer %q is %d words, and a flag holds %d. It sits in a label the width of one cell", n, w, maxStepperPointerWords)
		}
		if declared[n] {
			return fmt.Errorf("the pointer %q is declared twice. Two flags with one name cannot both be moved, because a beat names the pointer it is moving", n)
		}
		declared[n] = true
	}
	if sp.Target > maxStepperValue {
		return fmt.Errorf("the target is %d, and the cells hold 0-%d, so nothing in this array can ever be it. Either lower the target or put it in the row", sp.Target, maxStepperValue)
	}

	if p.Beats[0].Stepper == nil || p.Beats[0].Stepper.ResolvedShow() != "array" {
		return fmt.Errorf("beat %q does not open on the array. The data has to be on screen and read before anything happens to it, or the first comparison is two cells lighting up in a row nobody has looked at — open with {\"show\": \"array\"}",
			p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Stepper == nil {
		return fmt.Errorf("the final beat %q has no stepper direction", last.ID)
	} else if s := last.Stepper.ResolvedShow(); s != "found" && s != "done" {
		return fmt.Errorf("the clip ends on %q, which leaves the algorithm mid-step. A run has to be seen to finish — close with {\"show\": \"found\"} when there is a target, or {\"show\": \"done\"} when the row has simply reached its final order", s)
	}

	// The array as it really stands after each beat, with every swap applied.
	// This is what makes the "found" rule possible: the target has to be in the
	// cell the beat names AFTER everything that moved it.
	states := stepperStates(sp, p.Beats)

	for i, b := range p.Beats {
		sb := b.Stepper
		if sb == nil {
			return fmt.Errorf("beat %q has no stepper direction — every beat is one operation on the row", b.ID)
		}
		show := sb.ResolvedShow()
		if show == "array" && i != 0 {
			return fmt.Errorf("beat %q re-shows the bare array part-way through. The row is up from the first beat onward; a second {\"show\": \"array\"} throws away the pointers and the work", b.ID)
		}
		if show == "done" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q says the run is done and then the clip keeps going. \"done\" is the closer — after it the algorithm has nothing left to do", b.ID)
		}
		for _, c := range sb.At {
			if c < 0 || c >= len(sp.Values) {
				return fmt.Errorf("beat %q acts on cell %d, and the array has cells 0-%d. The indices are drawn under the tiles, so one off the end points at nothing",
					b.ID, c, len(sp.Values)-1)
			}
		}
		// The pointers, checked by name and by cell. A flag with a name nobody
		// declared is a label the viewer has never been introduced to.
		for name, cell := range sb.Ptr {
			if !declared[name] {
				return fmt.Errorf("beat %q moves a pointer called %q, which the plan never declared. The declared pointers are: %s. A flag whose name the viewer has not been introduced to is noise standing over a cell",
					b.ID, name, strings.Join(sp.Pointers, ", "))
			}
			if cell < 0 || cell >= len(sp.Values) {
				return fmt.Errorf("beat %q puts the pointer %q on cell %d, and the array has cells 0-%d. A flag has to stand over a tile that exists",
					b.ID, name, cell, len(sp.Values)-1)
			}
		}
		switch show {
		case "point":
			if len(sb.Ptr) == 0 {
				return fmt.Errorf("beat %q is a \"point\" beat that moves no pointer, so nothing on screen changes for the length of it. Give it a \"ptr\" entry, or say what it says over a compare", b.ID)
			}
		case "compare":
			if n := len(sb.At); n < 1 || n > 2 {
				return fmt.Errorf("beat %q compares %d cells, and a comparison is 1 or 2. One cell is a value being weighed against the target; two are being weighed against each other; three is a decision this picture cannot draw",
					b.ID, n)
			}
		case "swap":
			if n := len(sb.At); n != 2 {
				return fmt.Errorf("beat %q swaps %d cells, and a swap is exactly 2. The animation is two tiles arcing over each other into each other's places — with one there is nowhere to go, and with three there is no such move",
					b.ID, n)
			}
		case "found":
			if len(sb.At) != 1 {
				return fmt.Errorf("beat %q announces the find over %d cells, and a find lands on exactly 1. The whole payoff of the shot is one tile taking the frame", b.ID, len(sb.At))
			}
			if sp.Target < 0 {
				return fmt.Errorf("beat %q says the value was found, but the plan has no target — stepper.target is -1, which means this clip is a sort with nothing to search for. Set the target to the value being hunted, or close on {\"show\": \"done\"} instead",
					b.ID)
			}
			// THE MECHANICS. A model writing a search from memory announces the
			// answer at the wrong index fluently, and a search that lands on the
			// wrong cell teaches the viewer that the picture is decoration.
			if got := states[i][sb.At[0]]; got != sp.Target {
				return fmt.Errorf("beat %q says %d was found at cell %d, which holds %d. The row on screen is the row this plan described, tracked through every swap, so the cell the find lands on has to be the cell the value is actually in",
					b.ID, sp.Target, sb.At[0], got)
			}
		}
	}
	return nil
}

// stepperStates returns the array as it stands after each beat, with every swap
// applied in order.
//
// Deliberately defensive about indices, because its first caller is the
// validator — the one place where the indices have not been checked yet, and
// where a panic would turn a rejectable plan into a crashed run.
func stepperStates(sp *StepperSpec, beats []SnippetBeat) [][]int {
	cur := append([]int(nil), sp.Values...)
	out := make([][]int, len(beats))
	for i, b := range beats {
		if sb := b.Stepper; sb != nil && sb.ResolvedShow() == "swap" && len(sb.At) == 2 {
			a, c := sb.At[0], sb.At[1]
			if a >= 0 && a < len(cur) && c >= 0 && c < len(cur) {
				cur[a], cur[c] = cur[c], cur[a]
			}
		}
		out[i] = append([]int(nil), cur...)
	}
	return out
}

// stepperScenes lays the clip out as ONE scene, and ships the FULL contents of
// the array at every step. The component draws the state it is given; it never
// applies a swap itself, because then there would be two implementations of the
// algorithm's effect on the data and only one of them would be tested.
func stepperScenes(in SnippetSceneInput) ([]Scene, error) {
	sp := in.Plan.Stepper
	if sp == nil {
		return nil, fmt.Errorf("the plan has no array")
	}
	states := stepperStates(sp, in.Plan.Beats)

	// Pointers start off the row: -1 means "not placed yet", so a flag does not
	// appear over cell zero before any beat has put it there.
	pos := make(map[string]int, len(sp.Pointers))
	for _, n := range sp.Pointers {
		pos[n] = -1
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	touched := make([]int, 0, len(sp.Values))
	seen := map[int]bool{}
	ops := 0
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Stepper == nil {
			return nil, fmt.Errorf("beat %q has no stepper direction", beat.ID)
		}
		sb := beat.Stepper
		show := sb.ResolvedShow()
		for name, cell := range sb.Ptr {
			if _, ok := pos[name]; ok {
				pos[name] = cell
			}
		}
		if show == "compare" || show == "swap" {
			ops++
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"values":  append([]int(nil), states[i]...),
			"ops":     ops,
		}
		if len(sb.At) > 0 {
			step["at"] = append([]int(nil), sb.At...)
			for _, c := range sb.At {
				if c >= 0 && c < len(sp.Values) && !seen[c] {
					seen[c] = true
					touched = append(touched, c)
				}
			}
		}
		ptr := make(map[string]any, len(pos))
		for name, cell := range pos {
			ptr[name] = cell
		}
		step["ptr"] = ptr
		everTouched := append([]int(nil), touched...)
		sort.Ints(everTouched)
		step["touched"] = everTouched
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneStepper,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":    in.Plan.Title,
			"values":   append([]int(nil), sp.Values...),
			"pointers": append([]string(nil), sp.Pointers...),
			"target":   sp.Target,
			"steps":    steps,
		}),
	}}, nil
}
