package pipeline

// The callstack template: recursion filmed instead of described.
//
// Recursion is the wall a career-switcher hits hardest, and the reason is
// almost never the idea — "a function that calls itself" lands in one
// sentence. What does not land is the SHAPE of the execution: that four calls
// are alive at once, that each one is frozen mid-sentence waiting on the one
// above it, and that nothing at all is computed on the way down. The usual
// teaching aid is a trace printed as text, which shows the frames in the one
// arrangement that hides the point: a flat list, where the fourth call looks
// like a sibling of the first rather than a passenger riding on it.
//
// So this template draws the stack as a stack. Frames are plates that push
// upward with their arguments, the base case lands on top and the recursion
// stops, and then the values fall back down one plate at a time until the
// outermost frame — the one the caller actually asked — is holding the answer
// alone. Push, land, unwind. The clip is a physical process, and the
// narration is spoken over it rather than about it.
//
// The validator enforces stack discipline, because a stack drawn wrong is a
// stack that teaches the wrong thing. Frames must push in call order with no
// skips, since a frame appearing from nowhere is exactly the misconception
// (that the recursion "jumps to the bottom"). The base case may only land once
// every frame is up, because a base case shown early says the descent was
// optional. Returns must pop in reverse order, last in first out — a frame
// returning while another sits above it is not a call stack, it is a list, and
// it is the single most common way a hand-written recursion diagram is wrong.
// And the answer the clip ends on is Frames[0].Returns, computed here rather
// than trusted, so the number the viewer keeps is the number the outermost
// call actually produced.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "callstack",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The stack breathes",
		Description: "Recursion drawn as a stack that grows and shrinks: frames push upward with their arguments, the base case lands, and the values unwind back down into the call below. Reach for it when the subject is recursion itself — why the calls pile up, where the answer actually gets built, what a stack overflow is.",
		Example:     "factorial(4): the stack grows, then the answers fall back down",
		PromptFile:  snippetCallStackTemplateName,
		NeedsCode:   false,
		// Five frames pushing, landing and unwinding is a lot of distinct
		// states; under thirty-five seconds none of them holds long enough to
		// be read, and the picture becomes a flicker.
		MinTargetSec:     35,
		DefaultTargetSec: 55,
		// The worst case the shape allows: five calls, the base case landing,
		// five returns, and the empty closer. Twelve is exactly that clip, and
		// nothing larger is a stack anybody can follow.
		MaxBeats: 12,
		// A beat here is a SHOT — one plate moving — not a step in an argument.
		// Twenty-eight words is about nine seconds, which is how long a single
		// push or pop stays interesting.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{CallStack: true},
		OwnsPlan:          planFields{CallStack: true},
		Normalize:         normalizeCallStackPlan,
		Validate:          validateCallStackPlan,
		Scenes:            callStackScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":        strings.Join(MetricRoles(), ", "),
				"Shows":        strings.Join(CallStackShows(), ", "),
				"MinFrames":    minCallStackFrames,
				"MaxFrames":    maxCallStackFrames,
				"MaxFnWords":   maxCallStackFnWords,
				"MaxArgChars":  maxCallStackArgChars,
				"MaxBaseWords": maxCallStackBaseWords,
			}
		},
	})
}

const snippetCallStackTemplateName = "snippet_callstack.tmpl"

const (
	// Two frames is the smallest thing that is recursion at all: one call and
	// the base case it reaches. One frame is a function call, which every other
	// template in the catalog already draws better.
	minCallStackFrames = 2
	// Five plates is what the stage holds at a size where the arguments stay
	// legible in mono, and five is also the most pushes and pops MaxBeats can
	// fund. Deeper recursions are shown by choosing a smaller input, not by
	// shrinking the type.
	maxCallStackFrames = 5

	// A function name, not a signature — "factorial", "fib".
	maxCallStackFnWords = 2
	// An argument chip is a token the eye reads at a glance: "n=4", "[3,1,2]".
	// Ten characters is the widest string the plate carries without the mono
	// type having to shrink below the label size.
	maxCallStackArgChars = 10
	// The base case is a caption under the top plate, not an explanation.
	maxCallStackBaseWords = 10
)

// callStackShows is the closed vocabulary of what a beat does.
var callStackShows = map[string]bool{
	// Frame At pushes onto the stack, springing in with its arguments.
	"call": true,
	// The last frame lands and the base-case caption shows. The recursion
	// stops here, which is the only reason it ever ends.
	"base": true,
	// Frame At pops, and its value drops into the frame below it.
	"return": true,
	// The closer: the stack gone, the final answer alone centre stage.
	"empty": true,
}

// CallStackShows returns the beat vocabulary sorted.
func CallStackShows() []string {
	out := make([]string, 0, len(callStackShows))
	for k := range callStackShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// CallStackSpec is one recursion, frame by frame. On the plan because the
// frames persist for the whole clip — they are the thing being pushed and
// popped, not something a beat introduces.
type CallStackSpec struct {
	// Fn is the function being traced — "factorial".
	Fn string `json:"fn"`
	// Frames are the calls in CALL ORDER: Frames[0] is the outermost call the
	// caller made, and the last is the base case.
	Frames []CallStackFrame `json:"frames"`
	// Base says why the recursion stops — "zero is where it stops".
	Base string `json:"base"`
}

// CallStackFrame is one live call: what it was asked, and what it will answer.
type CallStackFrame struct {
	// Args is the argument chip on the plate — "n=4".
	Args string `json:"args"`
	// Returns is the value this frame hands to the frame below it — "24".
	Returns string `json:"returns"`
}

// CallStackBeat is one shot: which plate moves, and how.
type CallStackBeat struct {
	// Show is a callStackShows name.
	Show string `json:"show"`
	// At indexes CallStackSpec.Frames — the frame that pushes, lands or pops.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults the unknown to a call, which is what most beats of
// this template are: the stack spends the first half of the clip growing.
func (b CallStackBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if callStackShows[s] {
		return s
	}
	return "call"
}

func normalizeCallStackPlan(p *SnippetPlan) {
	c := p.CallStack
	if c == nil {
		return
	}
	c.Fn = clampWords(collapseSpaces(c.Fn), maxCallStackFnWords)
	c.Base = clampWords(collapseSpaces(c.Base), maxCallStackBaseWords)

	frames := make([]CallStackFrame, 0, len(c.Frames))
	for _, f := range c.Frames {
		f.Args = clampChars(collapseSpaces(f.Args), maxCallStackArgChars)
		f.Returns = clampChars(collapseSpaces(f.Returns), maxCallStackArgChars)
		if len(frames) < maxCallStackFrames {
			frames = append(frames, f)
		}
	}
	c.Frames = frames

	for i := range p.Beats {
		b := p.Beats[i].CallStack
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		n := len(c.Frames)
		switch b.Show {
		case "call", "return":
			if b.At < 0 {
				b.At = 0
			}
			if n > 0 && b.At >= n {
				b.At = n - 1
			}
		case "base":
			// The base case is the last frame by definition, so a stated index
			// is a fact about the plan rather than a choice; snap it.
			if n > 0 {
				b.At = n - 1
			}
		default:
			b.At = 0
		}
	}
}

func validateCallStackPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{CallStack: true}); err != nil {
		return err
	}

	c := p.CallStack
	if c == nil {
		return fmt.Errorf("the plan has no call stack — this template is a recursion drawn as plates pushing and popping, so the frames are the clip")
	}
	if strings.TrimSpace(c.Fn) == "" {
		return fmt.Errorf("the plan does not name the function. Every plate on screen is a call to the same function, and without its name the stack is a pile of unlabelled boxes")
	}
	if n := len(c.Frames); n < minCallStackFrames || n > maxCallStackFrames {
		return fmt.Errorf("the recursion has %d frames, want %d-%d. Below two there is nothing recursive to watch — one call and no descent — and past %d the plates cannot hold their argument chips at a legible size, so pick a smaller input rather than a deeper trace",
			n, minCallStackFrames, maxCallStackFrames, maxCallStackFrames)
	}
	for i, f := range c.Frames {
		if strings.TrimSpace(f.Args) == "" {
			return fmt.Errorf("frame %d has no arguments. The argument chip is what distinguishes one plate from the next — a stack of identical plates says the calls are identical, which is the opposite of the lesson", i)
		}
		if strings.TrimSpace(f.Returns) == "" {
			return fmt.Errorf("frame %d (%s) has no return value. Every frame hands something down when it pops, and the unwind is half the clip", i, f.Args)
		}
	}
	if strings.TrimSpace(c.Base) == "" {
		return fmt.Errorf("the plan does not say why the recursion stops. The base case is the only reason the stack ever comes back down, and a viewer who cannot see it reads the whole picture as an infinite loop")
	}

	if last := p.Beats[len(p.Beats)-1]; last.CallStack == nil || last.CallStack.ResolvedShow() != "empty" {
		return fmt.Errorf("beat %q does not close on the empty stack. The final frame is the one the viewer keeps: every plate gone and %s alone in the middle of the screen — end with {\"show\": \"empty\"}",
			p.Beats[len(p.Beats)-1].ID, c.Frames[0].Returns)
	}

	// THE STACK DISCIPLINE, simulated. A model writing this from memory pops
	// frames in the order it pushed them, which draws a queue and calls it a
	// stack, so the plan is executed here rather than believed.
	stack := make([]int, 0, len(c.Frames))
	pushed, popped, bases := 0, 0, 0
	for _, b := range p.Beats {
		d := b.CallStack
		if d == nil {
			return fmt.Errorf("beat %q has no callstack direction — every beat pushes a frame, lands the base case, pops a frame, or shows the empty stack", b.ID)
		}
		show := d.ResolvedShow()
		if show != "empty" && (d.At < 0 || d.At >= len(c.Frames)) {
			return fmt.Errorf("beat %q acts on frame %d, which does not exist — the recursion has frames 0-%d", b.ID, d.At, len(c.Frames)-1)
		}
		switch show {
		case "call":
			if pushed >= len(c.Frames) {
				return fmt.Errorf("beat %q makes another call, but all %d frames of %s are already on the stack. Add a frame to the plan or make this beat a return", b.ID, len(c.Frames), c.Fn)
			}
			if d.At != pushed {
				return fmt.Errorf("beat %q pushes frame %d (%s) when frame %d (%s) is the next one to be called. Frames go on in call order, 0 first, with no skips — a plate appearing from nowhere teaches that the recursion jumps to the bottom, which is precisely the thing this clip exists to correct",
					b.ID, d.At, c.Frames[d.At].Args, pushed, c.Frames[pushed].Args)
			}
			stack = append(stack, d.At)
			pushed++
		case "base":
			if pushed < len(c.Frames) {
				return fmt.Errorf("beat %q lands the base case with only %d of %d frames on the stack. The base case is the LAST call, reached after the descent — showing it early says the descent was optional",
					b.ID, pushed, len(c.Frames))
			}
			if d.At != len(c.Frames)-1 {
				return fmt.Errorf("beat %q calls frame %d (%s) the base case, but the base case is the deepest frame, %d (%s). That is the call that returns without calling again",
					b.ID, d.At, c.Frames[d.At].Args, len(c.Frames)-1, c.Frames[len(c.Frames)-1].Args)
			}
			bases++
		case "return":
			if len(stack) == 0 {
				return fmt.Errorf("beat %q returns from frame %d (%s) with nothing on the stack. A frame can only pop while it is live", b.ID, d.At, c.Frames[d.At].Args)
			}
			if top := stack[len(stack)-1]; d.At != top {
				return fmt.Errorf("frame %s cannot return while %s is still above it. Returns pop in reverse order, last called first — beat %q pops frame %d when frame %d is on top, and a stack that pops in call order is a queue, which is the wrong picture of what happens when a function calls itself",
					c.Frames[d.At].Args, c.Frames[top].Args, b.ID, d.At, top)
			}
			stack = stack[:len(stack)-1]
			popped++
		case "empty":
			if len(stack) != 0 {
				live := make([]string, 0, len(stack))
				for _, f := range stack {
					live = append(live, c.Frames[f].Args)
				}
				return fmt.Errorf("beat %q shows the empty stack with %d frame(s) still live: %s. Every frame has to pop before the stack is gone, and the answer only lands once the outermost call is the last thing left",
					b.ID, len(stack), strings.Join(live, ", "))
			}
			if popped < len(c.Frames) {
				return fmt.Errorf("beat %q shows the empty stack, but only %d of %d frames ever returned. The unwind is half this clip — every frame that pushed has to pop, in reverse order, before the answer is on screen",
					b.ID, popped, len(c.Frames))
			}
		}
	}
	if bases == 0 {
		return fmt.Errorf("no beat lands the base case. %q is the only reason this stack ever stops growing, and a recursion clip without it shows a descent with no floor — give the deepest frame a {\"show\": \"base\"} beat",
			c.Base)
	}
	return nil
}

// callStackScenes lays the clip out as ONE scene. The stack is simulated here
// so the component never has to reason about push and pop order: every step
// arrives carrying the exact set of frames that are live and the exact set
// that have returned.
func callStackScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.CallStack
	if c == nil {
		return nil, fmt.Errorf("the plan has no call stack")
	}
	if len(c.Frames) == 0 {
		return nil, fmt.Errorf("the recursion has no frames")
	}

	frames := make([]map[string]any, len(c.Frames))
	for i, f := range c.Frames {
		frames[i] = map[string]any{
			"args":    f.Args,
			"returns": f.Returns,
			"base":    i == len(c.Frames)-1,
		}
	}

	stack := make([]int, 0, len(c.Frames))
	returned := make([]int, 0, len(c.Frames))
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.CallStack == nil {
			return nil, fmt.Errorf("beat %q has no callstack direction", beat.ID)
		}
		show := beat.CallStack.ResolvedShow()
		at := beat.CallStack.At
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "call":
			if at < 0 || at >= len(c.Frames) {
				return nil, fmt.Errorf("beat %q pushes frame %d, which does not exist", beat.ID, at)
			}
			stack = append(stack, at)
			step["at"] = at
		case "base":
			step["at"] = len(c.Frames) - 1
		case "return":
			if at < 0 || at >= len(c.Frames) {
				return nil, fmt.Errorf("beat %q pops frame %d, which does not exist", beat.ID, at)
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			returned = append(returned, at)
			step["at"] = at
			// The chip that falls, and the plate it falls into. Precomputed so
			// the component animates a value it was handed rather than one it
			// looked up.
			step["value"] = c.Frames[at].Returns
			if at > 0 {
				step["into"] = at - 1
			}
		case "empty":
			stack = stack[:0]
		}
		onStack := make([]int, len(stack))
		copy(onStack, stack)
		sort.Ints(onStack)
		done := make([]int, len(returned))
		copy(done, returned)
		sort.Ints(done)
		step["onStack"] = onStack
		step["returned"] = done
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneCallStack,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"fn":     c.Fn,
			"base":   c.Base,
			"frames": frames,
			// The answer is the outermost call's return value, by definition —
			// resolved here so the closer cannot show a different number from
			// the one the bottom plate is holding.
			"answer": c.Frames[0].Returns,
			"steps":  steps,
		}),
	}}, nil
}
