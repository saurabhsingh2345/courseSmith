package pipeline

// How long the creator asked for, and how that gets spread over segments.
//
// A brief that says "create a 25-30 minute introduction video" was read for its
// subject and not for its length. CastCombo never set TargetSec, so every segment
// took its template's default and the finished file ran 9:43 against a 25-minute
// ask — a third of what was requested, with nothing anywhere saying so.
//
// The runtime is not a detail of the request; for a course introduction it is
// most of the request. A creator who asks for 25 minutes and gets 10 does not
// have a slightly short video, they have a video they cannot use.
//
// Two things are needed. The length has to be read out of the brief, because
// that is where people put it. And it has to be turned into per-segment targets
// before the caster runs, because how many parts a piece has is a function of
// how long it is — five segments cannot carry 25 minutes without each running
// five minutes, which is twice any template's ceiling.

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// preferredSegmentSec is the runtime a segment is aimed at when a whole-piece
// length has to be divided up.
//
// It was ninety, on the argument that a segment of a long piece is a chapter and
// can afford to develop an idea. That was written for twenty-five-minute asks,
// where dividing at forty-five would have wanted seventeen segments against a cap
// of twelve — and at that length it was right.
//
// It is wrong at the lengths this surface is actually used at, and wrong in the
// direction that matters. Five minutes came out as THREE segments of a hundred
// seconds each: a longer ask bought longer clips rather than more of them, which
// is the opposite of why anyone reaches for a combo. The whole point of cutting
// between looks is the cut, and a five-minute piece with two cuts in it is three
// snippets in a trenchcoat.
//
// There is a second reason, and it is about quality rather than variety. Every
// template in the catalog is designed around defaultSnippetTargetSec — forty-five
// seconds, seven beats at most, one picture per beat. Stretching one to a hundred
// seconds does not give it more room; it gives it the same seven beats holding
// the frame for fourteen seconds each, which is where a good template starts
// reading as a slideshow. Aiming at the design point keeps each segment inside
// what its own picture can hold, and spends the extra runtime on another look.
//
// So: five minutes is now eight segments of ~42s, and ten minutes is twelve of
// ~50s. The cap does the limiting past that, which is what it is for.
const preferredSegmentSec = defaultSnippetTargetSec

// runtimeRe matches the ways people write a length: "25-30 minute", "2 minutes",
// "90 seconds", "10 min", "1 hour". The unit is required — a bare number in a
// brief is far more often a count of things than a duration.
var runtimeRe = regexp.MustCompile(`(?i)(\d+)\s*(?:[-–—]\s*(\d+))?\s*(hour|hr|minute|min|second|sec)s?\b`)

// ParseRequestedRuntime finds the runtime a brief asks for, in seconds.
//
// Returns ok=false when the brief says nothing about length, which is the common
// case and not a problem: the segments then take their templates' defaults
// exactly as they did before.
//
// A range takes its midpoint. "25-30 minutes" is a request to land inside the
// range, so aiming at an end means half the tolerance is spent before anything
// is written; 27 minutes satisfies the ask and 25 barely does.
//
// The FIRST match wins rather than the largest. A brief mentioning "a 30-second
// hook" before asking for "20 minutes" is unusual, but when a length appears
// early it is nearly always the length of the thing being commissioned.
func ParseRequestedRuntime(brief string) (int, bool) {
	m := runtimeRe.FindStringSubmatch(brief)
	if m == nil {
		return 0, false
	}
	lo, err := strconv.Atoi(m[1])
	if err != nil || lo <= 0 {
		return 0, false
	}
	hi := lo
	if m[2] != "" {
		if h, err := strconv.Atoi(m[2]); err == nil && h >= lo {
			hi = h
		}
	}
	var unit int
	switch strings.ToLower(m[3]) {
	case "hour", "hr":
		unit = 3600
	case "minute", "min":
		unit = 60
	default:
		unit = 1
	}
	// Midpoint, rounded down; for a single value lo == hi and this is identity.
	return (lo + hi) / 2 * unit, true
}

// RuntimeBudget is a requested length spread over segments.
type RuntimeBudget struct {
	// RequestedSec is what the brief asked for.
	RequestedSec int
	// Segments is how many parts to cast.
	Segments int
	// PerSegmentSec is the runtime each one aims at.
	PerSegmentSec int
	// Shortfall is how many seconds short of the request the plan is, when even
	// the maximum number of segments at their maximum runtime cannot reach it.
	// Zero when the ask is satisfiable.
	//
	// Reported rather than silently absorbed, which is the whole bug: the old
	// behaviour was to deliver a third of the ask and say nothing.
	Shortfall int
}

// Achievable is the runtime this budget will actually produce.
func (b RuntimeBudget) Achievable() int { return b.Segments * b.PerSegmentSec }

// BudgetRuntime spreads a requested runtime over segments.
//
// wantSegments is the caller's preference (a CLI flag, the studio's default) and
// is honoured only when the brief asked for no particular length. Once a length
// is known it decides the count, because a segment count that cannot carry the
// runtime is not a preference, it is a contradiction — and the old code resolved
// that contradiction by quietly dropping the runtime.
func BudgetRuntime(requestedSec, wantSegments int) RuntimeBudget {
	if requestedSec <= 0 {
		return RuntimeBudget{}
	}
	b := RuntimeBudget{RequestedSec: requestedSec}

	// Start from the runtime a segment wants to be, then clamp the count.
	b.Segments = (requestedSec + preferredSegmentSec/2) / preferredSegmentSec
	b.Segments = min(max(b.Segments, minComboSegments), maxComboSegments)

	// Divide, then clamp the per-segment runtime to what a template can hold. A
	// long ask hits the ceiling; a very short one hits the floor.
	b.PerSegmentSec = requestedSec / b.Segments
	b.PerSegmentSec = min(max(b.PerSegmentSec, minSnippetTargetSec), maxSnippetTargetSec)

	// With the per-segment runtime pinned, the count may be able to close a gap
	// the first division left — 25 minutes over 12 segments is 125s each, but
	// 17 segments would be needed at 90s and only 12 are allowed, so the segments
	// stretch instead.
	if got := b.Achievable(); got < requestedSec && b.Segments < maxComboSegments {
		needed := (requestedSec + b.PerSegmentSec - 1) / b.PerSegmentSec
		b.Segments = min(needed, maxComboSegments)
	}
	if got := b.Achievable(); got < requestedSec {
		b.Shortfall = requestedSec - got
	}
	return b
}

// Describe renders the budget for a progress log, saying plainly when the ask
// cannot be met.
func (b RuntimeBudget) Describe() string {
	if b.RequestedSec <= 0 {
		return ""
	}
	s := fmt.Sprintf("%s asked for → %d segments of ~%s",
		durationWords(b.RequestedSec), b.Segments, durationWords(b.PerSegmentSec))
	if b.Shortfall > 0 {
		s += fmt.Sprintf("\n    ! that is the most %d segments can carry (%s); the piece will run about %s short of the ask",
			maxComboSegments, durationWords(b.Achievable()), durationWords(b.Shortfall))
	}
	return s
}

// segmentTargetFor raises a budgeted runtime to the floor its template can
// actually satisfy.
//
// A template's MinTargetSec is arithmetic, not preference: `story` needs eight
// beats and eight beats cannot be written inside a twenty-second word budget, so
// asking for one is a plan that cannot come out at all. ComboSpec.Validate
// rejects a TargetSec below the floor, which means an evenly-divided budget could
// produce a spec that fails validation immediately after casting — the caster's
// work thrown away over arithmetic it was never told about.
//
// Raising rather than rejecting: one segment running longer than its share is a
// piece slightly over its ask, which is a far better outcome than no piece.
func segmentTargetFor(template string, perSegmentSec int) int {
	target := min(max(perSegmentSec, minSnippetTargetSec), maxSnippetTargetSec)
	if tpl, ok := SnippetTemplates[template]; ok && tpl.MinTargetSec > target {
		target = min(tpl.MinTargetSec, maxSnippetTargetSec)
	}
	return target
}

// durationWords formats seconds the way somebody would say them.
func durationWords(sec int) string {
	if sec < 60 {
		return fmt.Sprintf("%ds", sec)
	}
	m, s := sec/60, sec%60
	if s == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dm%02ds", m, s)
}
