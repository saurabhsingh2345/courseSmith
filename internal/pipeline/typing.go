package pipeline

// When each character of a self-typing code scene appears.
//
// This used to live in the renderer, as a per-character random jitter
// normalised to fill the window. It moved here for two reasons, and the second
// is the one that forced it.
//
// The first is that "typing" was a uniform stream with noise on it, and people
// do not type that way. They type a word in a burst, stop at the end of a line,
// stop longer before a new block, and never type the indentation at all —
// the editor inserts it. A stream of evenly spaced characters reads as a
// teleprompter; the pauses are what make it read as somebody thinking.
//
// The second is sound. A keystroke click has to land on the exact frame its
// character appears, and Go cannot synthesise a click track for a schedule that
// only exists inside a React component. Owning the schedule here means the
// track and the animation are the same list of numbers by construction, rather
// than two implementations that agree until one of them is edited.
//
// Everything is deterministic: the jitter is a hash of the character's index,
// not a random source, because Remotion renders frames out of order and in
// parallel and the click track is generated in a different process entirely.

import (
	"hash/fnv"
	"strings"
)

// Relative cost of each kind of keystroke. These are weights, not durations —
// the whole sequence is normalised to whatever window the beat actually got, so
// a long file types faster rather than overrunning. What they encode is the
// *shape* of the rhythm, which is what survives that normalisation.
const (
	// An ordinary character inside a word.
	weightChar = 1.0
	// The space between words. A hair longer: it is where a hand resets.
	weightSpace = 1.35
	// A newline. The eye goes to the start of the next line and the hand
	// follows, and it is the most reliable pause in real typing.
	weightNewline = 3.4
	// A newline that ends a block opener — a line finishing in `:` or `{`. The
	// pause before writing the body of something is longer than the pause
	// between two statements, because it is the moment the next thought starts.
	weightBlockOpen = 5.2
	// Indentation. Effectively free: editors insert it, and a clip that types
	// four spaces one at a time is a clip that has never watched anyone code.
	// Not zero, so the caret still visibly lands in the right column.
	weightIndent = 0.12
	// A closing bracket that the editor would have auto-inserted. Nearly free
	// for the same reason.
	weightAutoClose = 0.2
)

// burstRun is how many characters a typing burst covers before the hand
// resets. Inside a burst the per-character weight is scaled down; the effect is
// that words arrive as words rather than as a metronome.
const burstRun = 5

// jitter is a deterministic 0.78–1.22 multiplier for position i. A hash rather
// than a PRNG so any process — the renderer, the click-track synthesiser, a
// test — computes the same value for the same index without sharing state.
func jitter(i int) float64 {
	h := fnv.New32a()
	var buf [4]byte
	buf[0] = byte(i)
	buf[1] = byte(i >> 8)
	buf[2] = byte(i >> 16)
	buf[3] = byte(i >> 24)
	_, _ = h.Write(buf[:])
	return 0.78 + 0.44*(float64(h.Sum32()%1000)/1000.0)
}

// autoClosing reports whether the character at position i is a closing bracket
// that an editor would have inserted when its opener was typed.
func autoClosing(runes []rune, i int) bool {
	switch runes[i] {
	case ')', ']', '}', '"', '\'':
	default:
		return false
	}
	// Only when the matching opener is the character immediately before, which
	// is the empty-pair case an editor actually completes: `()`, `[]`, `""`.
	if i == 0 {
		return false
	}
	switch runes[i] {
	case ')':
		return runes[i-1] == '('
	case ']':
		return runes[i-1] == '['
	case '}':
		return runes[i-1] == '{'
	case '"':
		return runes[i-1] == '"'
	case '\'':
		return runes[i-1] == '\''
	}
	return false
}

// keystrokeWeights returns the relative cost of typing each rune of `code`.
//
// Exposed separately from the schedule so the rhythm can be unit-tested as a
// shape — "the newline before a block body is the longest gap in this file" —
// without pinning it to a millisecond budget that will change.
func keystrokeWeights(code string) []float64 {
	runes := []rune(code)
	out := make([]float64, len(runes))

	// Whether position i is still inside the leading whitespace of its line.
	inIndent := true
	// Distance since the last burst reset.
	sinceReset := 0

	for i, r := range runes {
		var w float64
		switch {
		case r == '\n':
			// Look back at the line just finished: a block opener earns the
			// longer pause.
			line := lineEndingAt(runes, i)
			trimmed := strings.TrimRight(line, " \t")
			if strings.HasSuffix(trimmed, ":") || strings.HasSuffix(trimmed, "{") {
				w = weightBlockOpen
			} else if strings.TrimSpace(trimmed) == "" {
				// A blank line is a beat of its own, but a short one.
				w = weightNewline * 0.6
			} else {
				w = weightNewline
			}
			inIndent = true
			sinceReset = 0
		case inIndent && (r == ' ' || r == '\t'):
			w = weightIndent
		case r == ' ':
			w = weightSpace
			sinceReset = 0
		case autoClosing(runes, i):
			w = weightAutoClose
		default:
			w = weightChar
		}
		if r != '\n' && r != ' ' && r != '\t' {
			inIndent = false
		}

		// The burst: characters early in a run come faster than characters
		// late in one. Only applied to ordinary characters — stretching a
		// newline's pause because it happens to fall early in a run would undo
		// the thing the pause is for.
		if w == weightChar {
			burst := 0.82 + 0.36*(float64(sinceReset%burstRun)/float64(burstRun))
			w *= burst
			sinceReset++
		}

		out[i] = w * jitter(i)
	}
	return out
}

// lineEndingAt returns the text of the line that ends at the newline at index i.
func lineEndingAt(runes []rune, i int) string {
	start := i
	for start > 0 && runes[start-1] != '\n' {
		start--
	}
	return string(runes[start:i])
}

// KeystrokeSchedule returns the offset in milliseconds, from the moment typing
// begins, at which each character of `code` appears.
//
// The result always has one entry per rune and is non-decreasing, and the last
// entry is budgetMs — the schedule fills its window exactly, so the caller does
// not have to reason about a file that "finishes early" and leaves a caret
// blinking on a static screen for four seconds.
//
// A zero or negative budget yields all-zero offsets: everything is already
// there. That is the honest answer for a window too short to type in, and it is
// better than a schedule that types after the beat has ended.
func KeystrokeSchedule(code string, budgetMs int) []int {
	runes := []rune(code)
	if len(runes) == 0 {
		return nil
	}
	out := make([]int, len(runes))
	if budgetMs <= 0 {
		return out
	}

	weights := keystrokeWeights(code)
	total := 0.0
	for _, w := range weights {
		total += w
	}
	if total <= 0 {
		return out
	}

	acc := 0.0
	for i, w := range weights {
		acc += w
		out[i] = int(acc / total * float64(budgetMs))
	}
	return out
}

// KeystrokeTimesMs shifts a schedule to absolute scene-graph milliseconds.
func KeystrokeTimesMs(schedule []int, startMs int) []int {
	out := make([]int, len(schedule))
	for i, ms := range schedule {
		out[i] = startMs + ms
	}
	return out
}
