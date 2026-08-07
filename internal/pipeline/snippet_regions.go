package pipeline

// The regions template: the address space, drawn as one tall column.
//
// Ask a career-switcher where a variable lives and the honest answer is "in
// memory", which is not a place. Every explanation they have met draws memory
// as a row of numbered boxes — a picture of an ARRAY, not of a process — and
// then uses words like "the heap" and "the stack" as if the viewer had already
// been shown a map with those names on it. They never were. So this template
// draws the map: one tall column with low addresses at the bottom, the
// segments stacked in address order, and each one labelled with what it holds.
//
// The reason it is vertical and not a row is that the two interesting segments
// MOVE, and they move toward each other. The heap extends upward as you
// allocate; the stack extends downward as you call; the unclaimed space between
// them is the budget they are both spending. That single fact — one gap, two
// customers — is what makes stack overflow, heap exhaustion and "why does
// recursion crash" one idea instead of three, and a horizontal strip cannot
// show it because nothing in a strip grows toward anything.
//
// Which is why the validator is strict about adjacency and nothing else much.
// A plan may name any segments it likes, but if it names both a heap and a
// stack then the gap has to sit BETWEEN them in address order, because a
// diagram with the free space above the stack is not a simplification of how
// memory works, it is a different and wrong claim, and the growth beats would
// animate two blocks extending away from each other while the narration says
// they are closing in. Growth itself is closed to the segments that actually
// grow: a code segment "extending into the gap" is a sentence no operating
// system has ever meant. And the collision — the frame the whole clip is built
// to earn — is only legal when there are two fronts to collide.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "regions",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The address space",
		Description: "A process's memory as one tall column — code, static data, the heap growing up, the stack growing down, and the gap they share. Reach for it when the subject is where things live and what runs out: stack versus heap, why recursion crashes, what a segment actually is.",
		Example:     "Stack and heap grow toward each other — what happens when they meet",
		PromptFile:  snippetRegionsTemplateName,
		NeedsCode:   false,
		// The map, a region or two, a growth, the collision, the whole: under
		// thirty-five seconds the column goes up and comes down before anybody
		// has read a single address tick.
		MinTargetSec:     35,
		DefaultTargetSec: 50,
		// Opener + up to four focused regions + a growth + the collision +
		// closer. Past eight the same column is being re-lit.
		MaxBeats: 8,
		// A beat here is a SHOT — one state of one column — not a step in an
		// argument. Twenty-eight words is about nine seconds, which is as long
		// as a block edge advancing holds anybody's attention.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Regions: true},
		OwnsPlan:          planFields{Regions: true},
		Normalize:         normalizeRegionsPlan,
		Validate:          validateRegionsPlan,
		Scenes:            regionsScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(RegionsShows(), ", "),
				"RegionRoles":   strings.Join(RegionsRoles(), ", "),
				"MinRegions":    minRegionsCount,
				"MaxRegions":    maxRegionsCount,
				"MaxLabelWords": maxRegionsLabelWords,
				"MaxNoteWords":  maxRegionsNoteWords,
			}
		},
	})
}

const snippetRegionsTemplateName = "snippet_regions.tmpl"

const (
	// Below three blocks the column is not a map, it is a pair of boxes with a
	// line between them — and the picture this template exists for (heap, gap,
	// stack) is itself three blocks, so three is the floor by construction.
	minRegionsCount = 3
	// Past six the blocks are shorter than their own address ticks at 1080
	// lines, and a segment too short to hold its label is a coloured band.
	maxRegionsCount = 6

	// A segment label sits inside a block a few hundred pixels wide, set in
	// display type: three words is "static data", four is a sentence.
	maxRegionsLabelWords = 3
	// The note is one line beside the block while it is focused. Ten words is a
	// caption; more is a paragraph the viewer gets nine seconds to read.
	maxRegionsNoteWords = 10
)

// regionsRoles is the closed vocabulary of what a segment IS. Closed because
// the renderer draws each role differently — the heap gets an upward growth
// edge, the stack a downward one, the gap is hatched rather than filled — and
// an invented role has no drawing.
var regionsRoles = map[string]bool{
	// The instructions themselves, at the bottom of the space.
	"code": true,
	// Globals, constants, anything sized before the program runs.
	"static": true,
	// Allocated at runtime; extends upward, toward higher addresses.
	"heap": true,
	// Frames pushed by calls; extends downward, toward lower addresses.
	"stack": true,
	// The unclaimed space the heap and the stack are both spending.
	"gap": true,
}

// RegionsRoles returns the segment vocabulary sorted.
func RegionsRoles() []string {
	out := make([]string, 0, len(regionsRoles))
	for k := range regionsRoles {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// regionsShows is the closed vocabulary of what a beat does.
var regionsShows = map[string]bool{
	// The whole column, every block in place, dimmed. The opener.
	"map": true,
	// Block At comes forward with its note beside it.
	"region": true,
	// Block At — a heap or a stack — extends into the gap, edge advancing.
	"grow": true,
	// The two growth fronts meet: the stack-overflow frame.
	"collide": true,
	// The whole column lit at once. The closer.
	"whole": true,
}

// RegionsShows returns the beat vocabulary sorted.
func RegionsShows() []string {
	out := make([]string, 0, len(regionsShows))
	for k := range regionsShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// RegionsSpec is the address space itself. On the plan because the column is
// one object that stands for the whole clip.
type RegionsSpec struct {
	// Regions are the segments in ADDRESS order, lowest first — so index 0 is
	// drawn at the bottom of the column.
	Regions []RegionsRegion `json:"regions"`
}

// RegionsRegion is one block in the column.
type RegionsRegion struct {
	// Label names the segment — "code", "the heap", "static data".
	Label string `json:"label"`
	// Role is what the segment is: a regionsRoles name.
	Role string `json:"role,omitempty"`
	// Note is the one line that appears beside the block when it is focused.
	Note string `json:"note,omitempty"`
}

// ResolvedRole returns the segment's kind, defaulting the unknown to static.
// Static is the honest default: it is the segment that neither grows nor is
// free, so a role the model did not state cannot accidentally animate.
func (r RegionsRegion) ResolvedRole() string {
	s := strings.ToLower(strings.TrimSpace(r.Role))
	if regionsRoles[s] {
		return s
	}
	return "static"
}

// RegionsBeat is one shot: which state of the column this beat shows.
type RegionsBeat struct {
	// Show is a regionsShows name.
	Show string `json:"show"`
	// At indexes RegionsSpec.Regions, for a "region" or "grow" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a focused
// region — the workhorse state most beats of this template are in.
func (b RegionsBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if regionsShows[s] {
		return s
	}
	return "region"
}

// growsInto reports whether a role is one of the two that move.
func regionsRoleGrows(role string) bool {
	return role == "heap" || role == "stack"
}

// firstRegionWithRole returns the index of the first segment with a role, or
// -1. The validator guarantees there is at most one of each of the moving
// roles, so "first" is "the one" for heap, stack and gap.
func firstRegionWithRole(regions []RegionsRegion, role string) int {
	for i, r := range regions {
		if r.ResolvedRole() == role {
			return i
		}
	}
	return -1
}

func normalizeRegionsPlan(p *SnippetPlan) {
	rs := p.Regions
	if rs == nil {
		return
	}
	regions := make([]RegionsRegion, 0, len(rs.Regions))
	for _, r := range rs.Regions {
		r.Label = clampWords(collapseSpaces(r.Label), maxRegionsLabelWords)
		r.Note = clampWords(collapseSpaces(r.Note), maxRegionsNoteWords)
		r.Role = r.ResolvedRole()
		if len(regions) < maxRegionsCount {
			regions = append(regions, r)
		}
	}
	rs.Regions = regions

	for i := range p.Beats {
		b := p.Beats[i].Regions
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "region" && b.Show != "grow" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(rs.Regions); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateRegionsPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Regions: true}); err != nil {
		return err
	}

	rs := p.Regions
	if rs == nil {
		return fmt.Errorf("the plan has no address space — this template is one tall column of memory segments, so the column is the clip")
	}
	if n := len(rs.Regions); n < minRegionsCount || n > maxRegionsCount {
		return fmt.Errorf("the column holds %d segment(s), want %d-%d. Below %d it is a pair of boxes rather than a map — the picture this template exists for, heap and gap and stack, is itself three blocks — and past %d the blocks are shorter than their own address ticks",
			n, minRegionsCount, maxRegionsCount, minRegionsCount, maxRegionsCount)
	}
	counts := map[string]int{}
	for i, r := range rs.Regions {
		if strings.TrimSpace(r.Label) == "" {
			return fmt.Errorf("segment %d has no label — an unnamed block in an address space is exactly the picture this template exists to replace", i)
		}
		if s := strings.ToLower(strings.TrimSpace(r.Role)); s != "" && !regionsRoles[s] {
			return fmt.Errorf("segment %d (%q) has role %q, which is not one of: %s. The renderer draws each role differently — the heap gets an upward growth edge, the gap is hatched rather than filled — so an invented role has no drawing",
				i, r.Label, r.Role, strings.Join(RegionsRoles(), ", "))
		}
		counts[r.ResolvedRole()]++
	}
	// One of each moving role, at most. A process has one heap and one stack;
	// two of either is not a simplification, it is a different machine, and the
	// growth beats would have no single front to advance.
	for _, role := range []string{"heap", "stack", "gap"} {
		if counts[role] > 1 {
			return fmt.Errorf("the column has %d segments with role %q. There is one %s in this picture — two of them leaves the growth beats with no single edge to advance and the collision with no two fronts to meet",
				counts[role], role, role)
		}
	}

	heapAt := firstRegionWithRole(rs.Regions, "heap")
	stackAt := firstRegionWithRole(rs.Regions, "stack")
	gapAt := firstRegionWithRole(rs.Regions, "gap")

	// THE ADJACENCY, which is the entire picture. Segments are given in address
	// order, so "between" is a fact about slice positions and can be checked.
	if heapAt >= 0 && stackAt >= 0 {
		if gapAt < 0 {
			return fmt.Errorf("the column has a heap at position %d and a stack at position %d but no gap between them. The free space they are BOTH spending is the whole idea — without a gap block the growth beats extend into segments that are already claimed, and the collision has nothing to happen in. Add a segment with role \"gap\" between them",
				heapAt, stackAt)
		}
		lo, hi := heapAt, stackAt
		if lo > hi {
			lo, hi = hi, lo
		}
		if gapAt < lo || gapAt > hi {
			return fmt.Errorf("the heap is at position %d, the stack is at position %d, and the gap is at position %d — outside them. Segments are listed in address order, so the gap has to sit BETWEEN the heap and the stack: that adjacency is the entire picture, because it is what makes them grow toward each other rather than away",
				heapAt, stackAt, gapAt)
		}
	}

	// The shape. The column is seen whole before any block is singled out.
	if p.Beats[0].Regions == nil || p.Beats[0].Regions.ResolvedShow() != "map" {
		return fmt.Errorf("beat %q does not open on the whole map. A block lighting inside a column nobody has been shown is a coloured rectangle — the first beat is {\"show\": \"map\"}",
			p.Beats[0].ID)
	}
	grown, collided := 0, 0
	for i, b := range p.Beats {
		if b.Regions == nil {
			return fmt.Errorf("beat %q has no regions direction — every beat shows one state of the column", b.ID)
		}
		show := b.Regions.ResolvedShow()
		switch show {
		case "map":
			if i != 0 {
				return fmt.Errorf("beat %q shows the bare map again part-way through. The map is the opener; going back to it un-grows an edge the clip already advanced", b.ID)
			}
		case "whole":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q lights the whole column before the end. \"whole\" is the closer — after it there is nothing left to show", b.ID)
			}
		case "region":
			if b.Regions.At < 0 || b.Regions.At >= len(rs.Regions) {
				return fmt.Errorf("beat %q focuses segment %d, which does not exist — the column holds segments 0-%d", b.ID, b.Regions.At, len(rs.Regions)-1)
			}
		case "grow":
			if b.Regions.At < 0 || b.Regions.At >= len(rs.Regions) {
				return fmt.Errorf("beat %q grows segment %d, which does not exist — the column holds segments 0-%d", b.ID, b.Regions.At, len(rs.Regions)-1)
			}
			r := rs.Regions[b.Regions.At]
			if !regionsRoleGrows(r.ResolvedRole()) {
				return fmt.Errorf("beat %q grows segment %d (%q), whose role is %q. Only the heap and the stack move: the code and static segments are sized before the program starts, so a %q segment extending into the gap is a sentence no operating system has ever meant",
					b.ID, b.Regions.At, r.Label, r.ResolvedRole(), r.ResolvedRole())
			}
			if gapAt < 0 {
				return fmt.Errorf("beat %q grows segment %d (%q) but the column has no gap. Growth is a block edge advancing into free space — with nothing free, the edge advances into a neighbour that is already spoken for",
					b.ID, b.Regions.At, r.Label)
			}
			grown++
		case "collide":
			if heapAt < 0 || stackAt < 0 {
				return fmt.Errorf("beat %q shows the collision, but the column has %s. The collision is two growth fronts meeting — with only one of them there is nothing to meet",
					b.ID, regionsMissingFronts(heapAt, stackAt))
			}
			collided++
			if collided > 1 {
				return fmt.Errorf("beat %q collides a second time. The fronts meet once — that frame is the end of the story the column is telling, and a rerun says the first one did not count", b.ID)
			}
			if grown == 0 {
				return fmt.Errorf("beat %q collides before anything has grown. The meeting only reads as a meeting if the viewer has watched an edge advance first — put at least one {\"show\": \"grow\"} beat before it", b.ID)
			}
		}
	}
	return nil
}

// regionsMissingFronts names which growth front the column is missing, for the
// collision rejection.
func regionsMissingFronts(heapAt, stackAt int) string {
	switch {
	case heapAt < 0 && stackAt < 0:
		return "neither a heap nor a stack"
	case heapAt < 0:
		return "no heap"
	default:
		return "no stack"
	}
}

// regionsScenes lays the clip out as ONE scene. The column persists; the steps
// say which block is forward, which edges have advanced, and whether the fronts
// have met. Which way a block grows is decided here, from its role, so the
// component never has to know what a heap is.
func regionsScenes(in SnippetSceneInput) ([]Scene, error) {
	rs := in.Plan.Regions
	if rs == nil {
		return nil, fmt.Errorf("the plan has no address space")
	}

	regions := make([]map[string]any, len(rs.Regions))
	for i, r := range rs.Regions {
		role := r.ResolvedRole()
		dir := ""
		switch role {
		case "heap":
			dir = "up"
		case "stack":
			dir = "down"
		}
		regions[i] = map[string]any{
			"label": r.Label,
			"role":  role,
			"note":  r.Note,
			"grows": dir,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	// Which blocks have been focused and which have advanced, accumulated in Go
	// so the renderer draws a whole frame from one step.
	seen := map[int]bool{}
	grown := map[int]bool{}
	collided := false
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Regions == nil {
			return nil, fmt.Errorf("beat %q has no regions direction", beat.ID)
		}
		show := beat.Regions.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "region":
			seen[beat.Regions.At] = true
			step["at"] = beat.Regions.At
		case "grow":
			seen[beat.Regions.At] = true
			grown[beat.Regions.At] = true
			step["at"] = beat.Regions.At
		case "collide":
			collided = true
		}
		seenIdx := make([]int, 0, len(seen))
		for at := range seen {
			seenIdx = append(seenIdx, at)
		}
		sort.Ints(seenIdx)
		step["seen"] = seenIdx
		grownIdx := make([]int, 0, len(grown))
		for at := range grown {
			grownIdx = append(grownIdx, at)
		}
		sort.Ints(grownIdx)
		step["grown"] = grownIdx
		step["collided"] = collided
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneRegions,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":   in.Plan.Title,
			"regions": regions,
			"heapAt":  firstRegionWithRole(rs.Regions, "heap"),
			"stackAt": firstRegionWithRole(rs.Regions, "stack"),
			"gapAt":   firstRegionWithRole(rs.Regions, "gap"),
			"steps":   steps,
		}),
	}}, nil
}
