package pipeline

// Footage: metadata for a clip of real captured screen, and the honest timing
// of the moments inside it.
//
// The Python course's quality moat is that every code block really ran. A
// no-code course has no Python, and its moat is the same shape with a different
// engine: *the tool really did that*. What makes that checkable rather than
// merely asserted is this file — a clip carries the binary it recorded, the
// version of that binary observed at capture time, and when it was shot. None
// of those three are written by a model.
//
// == Marks, and why they are allowed to admit defeat ==
//
// A raw recording is not editable by a machine. "Cut to the moment the deploy
// went green" needs that moment as a number, and VHS emits no timestamps at
// all — it renders a tape and hands back an mp4.
//
// So a mark's offset is *computed* from the tape: typing speed times
// characters, plus the explicit Sleeps. For an ordinary scripted demo that
// model is exact, because VHS advances tape time by exactly those amounts.
//
// It stops being exact at `Wait`, which blocks on the real screen until a
// pattern matches, for however long the real command really takes. That is
// precisely what the interesting captures do: `claude` thinking, `vercel
// deploy` building. A tape with a Wait in it runs longer than its tape time,
// and every mark after that Wait is late by an amount the model cannot know.
//
// It can know it once, though. With a single Wait, the whole discrepancy
// between computed tape time and the real mp4 duration belongs to that Wait —
// so marks before it are exact, and marks after it shift by a measured amount.
// Two Waits and the discrepancy cannot be attributed, so the marks after the
// first one are flagged `approximate`, which disables mark-accurate cutting and
// speed-ramping for them rather than silently mistiming the cut.
//
// That is the same posture as the WER gate on alignment: measure the thing, and
// say so when it drifted. A mark that is quietly 8 seconds out is worse than no
// mark, because nothing downstream can tell.

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/project"
)

// FootageFileSuffix names the sidecar written beside each clip. It lives with
// the clip rather than in a shared library because the renderer bundles assets
// out of the lesson's generated dir; lifting clips into `.coursesmith/footage/`
// is phase 2, and moving them is what makes it phase 2.
const FootageFileSuffix = ".footage.json"

// CaptureKind is how a clip was produced, which decides where it may run.
type CaptureKind string

const (
	// CaptureKindPython is the original demos path: course code executed in
	// the Docker sandbox with networking off.
	CaptureKindPython CaptureKind = "python"
	// CaptureKindTool records a real CLI session — `claude`, `vercel`, `gh`.
	// These are network clients holding credentials, so they cannot run in the
	// isolation the Python path depends on. See capture.go.
	CaptureKindTool CaptureKind = "tool"
	// CaptureKindWeb records real frames of somebody else's web product, from
	// a checked-in take script. Stills rather than video — see capture_web.go
	// for why that is the right default and not a limitation.
	CaptureKindWeb CaptureKind = "web"
	// CaptureKindDesktop records a native application — Cursor, Figma desktop —
	// with an operator working through a checked-in list of beats. The only
	// path with a person in it, because a native app has no selectors to drive.
	CaptureKindDesktop CaptureKind = "desktop"
)

// FootageMark is one named moment inside a clip.
type FootageMark struct {
	// Name is the mark's id, as written in the tape's `# MARK <name>` comment.
	Name string `json:"name"`
	// AtMs is the offset into the clip.
	AtMs int `json:"atMs"`
	// Approximate is set when the tape-time model could not account for this
	// mark's position — more than one Wait ran before it. Consumers must not
	// cut or speed-ramp on an approximate mark.
	Approximate bool `json:"approximate,omitempty"`
}

// FocusBox is the part of a frame the shot is actually about, normalized to
// 0..1 of the viewport.
//
// Normalized rather than pixels because the renderer scales the frame to the
// stage, and a pixel box would be correct only at the resolution it was
// captured at — which is exactly the kind of thing that goes wrong silently
// when somebody changes the capture viewport a year from now.
type FocusBox struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

// FootageFrame is one still from a web capture.
type FootageFrame struct {
	// Mark names the frame, in the same vocabulary a tape's marks use.
	Mark string `json:"mark"`
	// Path is relative to the lesson's generated dir.
	Path string `json:"path"`
	// Focus is what to push in on; nil means hold the whole frame.
	Focus *FocusBox `json:"focus,omitempty"`
}

// Footage is the sidecar for one captured clip.
type Footage struct {
	// ID matches the clip's id (demo-1, capture-2).
	ID   string      `json:"id"`
	Kind CaptureKind `json:"kind"`
	// Tool is the capture tool's key from the allowlist ("claude"), empty for
	// the python kind.
	Tool string `json:"tool,omitempty"`
	// ToolVersion is what the binary reported at capture time, observed by
	// running its own version flag rather than declared by anybody. Empty when
	// the tool could not be asked.
	ToolVersion string `json:"toolVersion,omitempty"`
	// CapturedAt is RFC3339. This is the freshness clock: these tools redesign
	// quarterly and nothing else in the build can tell that a clip has aged.
	CapturedAt string `json:"capturedAt"`
	// DurationMs is the real measured duration of the rendered clip.
	DurationMs int `json:"durationMs"`
	// TapeTimeMs is what the mark model predicted the clip would run to. The
	// gap between this and DurationMs is the Wait time, and is what makes the
	// single-Wait case exact.
	TapeTimeMs int `json:"tapeTimeMs"`
	// Waits is how many Wait directives the tape contained. 0 or 1 means every
	// mark is exact.
	Waits int `json:"waits"`
	// Marks are the named moments, in tape order. Terminal captures only —
	// a still has no offset into anything.
	Marks []FootageMark `json:"marks,omitempty"`
	// Frames are the stills of a web capture, in take order. The scene divides
	// its screen time across them, so a frame carries no duration of its own.
	Frames []FootageFrame `json:"frames,omitempty"`
	// Origin is the host a web capture was taken at, written by the driver.
	// This is the provenance record: a clip claiming to show Lovable carries
	// the URL it was really captured from, and nothing a model writes can put
	// a different one here.
	Origin string `json:"origin,omitempty"`
	// Take is the take file a web capture was driven by, so a re-shoot is
	// "run this again" rather than an act of archaeology.
	Take string `json:"take,omitempty"`
}

// Exact reports whether every mark in this clip is trustworthy enough to cut
// on. Callers wanting per-mark granularity should read Marks directly.
func (f Footage) Exact() bool {
	for _, m := range f.Marks {
		if m.Approximate {
			return false
		}
	}
	return true
}

// Mark finds a mark by name.
func (f Footage) Mark(name string) (FootageMark, bool) {
	for _, m := range f.Marks {
		if m.Name == name {
			return m, true
		}
	}
	return FootageMark{}, false
}

// markCommentRe matches the `# MARK <name>` comments the tape carries. VHS
// ignores `#` lines, which is what lets a mark ride inside a tape the tool
// still validates — the alternative was a second sidecar file the model would
// have to keep in step with the tape by hand.
var markCommentRe = regexp.MustCompile(`^#\s*MARK\s+([a-z0-9][a-z0-9-]*)\s*$`)

// typeDirectiveRe matches `Type "..."` and `Type@50ms "..."`, capturing the
// per-command speed override and the typed text.
var typeDirectiveRe = regexp.MustCompile(`^Type(?:@(\d+)(ms|s))?\s+(.*)$`)

// keyDirectiveRe matches a keypress with an optional speed override and an
// optional repeat count: `Enter`, `Enter 3`, `Backspace@20ms 5`.
var keyDirectiveRe = regexp.MustCompile(`^(Enter|Backspace|Delete|Tab|Space|Up|Down|Left|Right|Escape|Ctrl\+[A-Za-z]|Alt\+[A-Za-z])(?:@(\d+)(ms|s))?(?:\s+(\d+))?\s*$`)

// sleepDirectiveRe matches `Sleep 2s`, `Sleep 500ms`, `Sleep 1.5s`.
var sleepDirectiveRe = regexp.MustCompile(`^Sleep\s+([0-9.]+)(ms|s|m)?\s*$`)

// waitDirectiveRe matches any form of VHS's Wait, whose real duration is
// unknowable from the tape. That is the whole reason marks can be approximate.
var waitDirectiveRe = regexp.MustCompile(`^Wait\b`)

// parseDurationUnit turns a VHS numeric+unit pair into milliseconds.
func parseDurationUnit(num, unit string) (int, error) {
	v, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a number", num)
	}
	switch unit {
	case "", "s":
		return int(v * 1000), nil
	case "ms":
		return int(v), nil
	case "m":
		return int(v * 60000), nil
	}
	return 0, fmt.Errorf("unknown unit %q", unit)
}

// typedTextLen is the number of characters VHS will actually type for a Type
// directive's argument, which may be single- or double-quoted or backticked.
func typedTextLen(arg string) int {
	arg = strings.TrimSpace(arg)
	if len(arg) >= 2 {
		if q := arg[0]; (q == '"' || q == '\'' || q == '`') && arg[len(arg)-1] == q {
			return len([]rune(arg[1 : len(arg)-1]))
		}
	}
	return len([]rune(arg))
}

// tapeScan is the result of walking a tape body's timeline.
type tapeScan struct {
	// marks are in tape order, with AtMs holding the *tape-time* offset —
	// before any Wait discrepancy has been distributed onto them.
	marks []FootageMark
	// waitsBefore[i] is how many Wait directives preceded marks[i].
	waitsBefore []int
	// totalMs is the tape time of the whole body.
	totalMs int
	// waits is the total Wait count.
	waits int
}

// scanTape walks a tape body accumulating tape time and recording where the
// marks and Waits fell. It deliberately does not error on a directive it does
// not recognise: VHS validated this tape already, and a timing model that
// refuses to run because it met an unfamiliar keyword would take the marks down
// with it. An unknown directive contributes no time, which is the direction
// that shows up as drift rather than as a confidently wrong number.
func scanTape(body string, defaultTypingMs int) tapeScan {
	scan := tapeScan{}
	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if m := markCommentRe.FindStringSubmatch(line); m != nil {
			scan.marks = append(scan.marks, FootageMark{Name: m[1], AtMs: scan.totalMs})
			scan.waitsBefore = append(scan.waitsBefore, scan.waits)
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if waitDirectiveRe.MatchString(line) {
			scan.waits++
			continue
		}
		if m := sleepDirectiveRe.FindStringSubmatch(line); m != nil {
			if ms, err := parseDurationUnit(m[1], m[2]); err == nil {
				scan.totalMs += ms
			}
			continue
		}
		if m := typeDirectiveRe.FindStringSubmatch(line); m != nil {
			speed := defaultTypingMs
			if m[1] != "" {
				if ms, err := parseDurationUnit(m[1], m[2]); err == nil {
					speed = ms
				}
			}
			scan.totalMs += typedTextLen(m[3]) * speed
			continue
		}
		if m := keyDirectiveRe.FindStringSubmatch(line); m != nil {
			speed := defaultTypingMs
			if m[2] != "" {
				if ms, err := parseDurationUnit(m[2], m[3]); err == nil {
					speed = ms
				}
			}
			repeat := 1
			if m[4] != "" {
				if n, err := strconv.Atoi(m[4]); err == nil && n > 0 {
					repeat = n
				}
			}
			scan.totalMs += speed * repeat
			continue
		}
	}
	return scan
}

// markDriftToleranceMs is how far a mark may sit from its computed position
// before it stops being worth cutting on. Half a second is about where a cut
// starts landing on the wrong beat; VHS also trims a couple of hundred
// milliseconds off the tail of every clip, so anything tighter would flag
// perfectly good marks on the encoder's behalf.
const markDriftToleranceMs = 500

// resolveMarks places the scanned marks on the real clip's timeline.
//
// The discrepancy between the tape's computed time and the clip's real duration
// is Wait time. Three cases, in the order they are checked:
//
// **The whole clip ran to its tape time.** Then no Wait blocked for a
// meaningful period, and the mark count stops mattering: the drift is the *sum*
// of every wait's overrun, so a small total is a proof that each one was small.
// This is not a corner case — it is the common one. A model told to wait for a
// slow command puts `Wait` after the fast ones too, and `Wait` after `git init`
// returns instantly. Flagging those marks would throw away good timing to
// punish a stylistic habit.
//
// **Exactly one Wait, and it took real time.** The whole discrepancy is that
// Wait's, so marks before it are still exact and marks after it move by the
// measured amount. This is the case worth engineering for: "run the agent, wait
// for it to finish, show the result" is one Wait.
//
// **Several Waits, and between them they took real time.** Now the discrepancy
// cannot be split, so every mark after the first Wait is flagged rather than
// guessed.
func resolveMarks(scan tapeScan, realDurationMs int) []FootageMark {
	if len(scan.marks) == 0 {
		return nil
	}
	drift := realDurationMs - scan.totalMs
	if drift < 0 {
		drift = 0
	}
	cheap := drift <= markDriftToleranceMs
	out := make([]FootageMark, len(scan.marks))
	for i, m := range scan.marks {
		out[i] = m
		switch {
		case cheap:
			// Every Wait was cheap, so tape time is real time throughout.
		case scan.waitsBefore[i] == 0:
			// Nothing unknowable has happened yet; tape time is real time.
		case scan.waits == 1:
			// The single Wait is entirely responsible for the drift.
			out[i].AtMs = m.AtMs + drift
		default:
			out[i].AtMs = m.AtMs + drift
			out[i].Approximate = true
		}
		// A mark can never sit past the last frame, whatever the model said.
		// This runs on every path — the cheap-waits case is the one where a
		// short clip and a long tape can still put a mark off the end.
		if out[i].AtMs > realDurationMs {
			out[i].AtMs = realDurationMs
			out[i].Approximate = true
		}
	}
	return out
}

// PacingSegment is one stretch of a recording and how fast to play it.
type PacingSegment struct {
	FromMs int     `json:"fromMs"`
	ToMs   int     `json:"toMs"`
	Rate   float64 `json:"rate"`
}

const (
	// maxPacingRate is the fastest a stretch may be played before it reads as a
	// glitch rather than a deliberate fast-forward. Dead air takes a lot of
	// compression before anybody minds; this is where it stops looking intended.
	maxPacingRate = 12.0
	// deadAirMs is the shortest gap worth compressing. Below this it is a beat
	// in the demo rather than time nobody wants back.
	deadAirMs = 2500
)

// PlanTerminalPacing fits a recording of clipMs into a slot of slotMs.
//
// A scene's length comes from the narration and a capture's length comes from
// how long a real tool really took, and those two numbers have no reason to
// agree. The first real agent capture was 53 seconds of recording in a 21
// second slot: the video cut away ten seconds in, and the viewer never saw the
// result — which was the entire point of the shot.
//
// Playing the whole thing faster would make the typing look ridiculous. What
// the clip actually holds is a few moments separated by dead air, and the marks
// say where those moments are. So the gaps get compressed and the moments stay
// at real time. This is what the marks were built for and the first thing to
// use them.
//
// Callers pass only exact marks. An approximate mark is not a cut point, so a
// clip whose timing could not be attributed falls back to a uniform rate rather
// than confidently cutting in the wrong place.
//
// A clip that already fits returns a single real-time segment, which is the
// behaviour every existing python demo has always had.
func PlanTerminalPacing(clipMs, slotMs int, marks []FootageMark) []PacingSegment {
	whole := []PacingSegment{{FromMs: 0, ToMs: clipMs, Rate: 1}}
	if clipMs <= 0 || slotMs <= 0 || clipMs <= slotMs {
		return whole
	}

	points := make([]int, 0, len(marks))
	for _, m := range marks {
		if m.AtMs > 0 && m.AtMs < clipMs {
			points = append(points, m.AtMs)
		}
	}
	sort.Ints(points)
	if len(points) == 0 {
		// Nothing says which part is dead air. Too fast everywhere still beats
		// losing the ending.
		return []PacingSegment{{FromMs: 0, ToMs: clipMs, Rate: math.Min(maxPacingRate, float64(clipMs)/float64(slotMs))}}
	}

	bounds := append([]int{0}, points...)
	bounds = append(bounds, clipMs)
	segs := make([]PacingSegment, 0, len(bounds)-1)
	for i := 0; i < len(bounds)-1; i++ {
		segs = append(segs, PacingSegment{FromMs: bounds[i], ToMs: bounds[i+1], Rate: 1})
	}

	// Spend the overflow on the longest gaps first, none of them going past
	// maxPacingRate. Longest-first is what keeps the short segments — the
	// typing, the output landing — at real time for as long as possible.
	order := make([]int, len(segs))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		return (segs[order[a]].ToMs - segs[order[a]].FromMs) > (segs[order[b]].ToMs - segs[order[b]].FromMs)
	})
	over := float64(clipMs - slotMs)
	for _, i := range order {
		if over <= 0 {
			break
		}
		length := float64(segs[i].ToMs - segs[i].FromMs)
		if length < deadAirMs {
			continue
		}
		canSave := length - length/maxPacingRate
		save := math.Min(canSave, over)
		segs[i].Rate = length / (length - save)
		over -= save
	}

	// Every gap has given what it can and it is still too long: the clip is
	// mostly moments, so the only honest option left is to speed all of it up
	// rather than lose the end.
	if over > 0 {
		played := 0.0
		for _, s := range segs {
			played += float64(s.ToMs-s.FromMs) / s.Rate
		}
		extra := played / float64(slotMs)
		for i := range segs {
			segs[i].Rate = math.Min(maxPacingRate, segs[i].Rate*extra)
		}
	}
	return segs
}

// PlayedMs is how long a pacing plan takes to play.
func PlayedMs(segs []PacingSegment) int {
	total := 0.0
	for _, s := range segs {
		total += float64(s.ToMs-s.FromMs) / s.Rate
	}
	return int(total)
}

// CaptureCredit is what a capture scene states on screen about itself.
//
// It exists because of something the pacing work introduced. Compressing 53
// seconds of real agent work into a 21-second slot and saying nothing is a
// quiet misrepresentation of how long the tool took — in a course whose entire
// moat is "the tool really did that". Every other claim here is defended by a
// recording; the length of the recording cannot be the one claim we let drift.
//
// So a paced clip says so, in the frame, with both numbers. And since the chip
// is there anyway it carries the tool's observed version, which makes it the
// first *rendered* captured fact: every field is measured by the recorder, none
// is written by a model.
type CaptureCredit struct {
	// Tool is the display name — "Claude Code", "Lovable".
	Tool string `json:"tool,omitempty"`
	// Version is what the binary reported at capture time; empty for a web or
	// desktop capture, where there is no version to ask for.
	Version string `json:"version,omitempty"`
	// RealMs is how long the recording really is.
	RealMs int `json:"realMs,omitempty"`
	// ShownMs is how long it is on screen. Set only when the two differ, so a
	// clip that plays at real time makes no claim about its own speed.
	ShownMs int `json:"shownMs,omitempty"`
}

// versionNumberRe finds the dotted number inside a tool's version banner.
var versionNumberRe = regexp.MustCompile(`\d+(?:\.\d+)+`)

// shortVersion reduces a version banner to the number in it.
//
// Tools print their name alongside their version — `claude --version` gives
// "2.1.220 (Claude Code)" and git gives "git version 2.50.1 (Apple Git-155)" —
// so a credit built from the raw string reads "Claude Code 2.1.220 (Claude
// Code)". The full banner stays in footage.json, because that file is the
// evidence and evidence should not be tidied; only the display is shortened.
func shortVersion(banner string) string {
	if m := versionNumberRe.FindString(banner); m != "" {
		return m
	}
	return strings.TrimSpace(banner)
}

// captureCreditFor builds the on-screen credit for a clip.
func captureCreditFor(f Footage, display string, realMs, slotMs int) CaptureCredit {
	p := CaptureCredit{Tool: display, Version: shortVersion(f.ToolVersion), RealMs: realMs}
	// Only claim a compression that actually happened. A second either way is
	// rounding, not a fast-forward, and a chip that says "12s real, shown in
	// 12s" is noise pretending to be rigour.
	if realMs > 0 && slotMs > 0 && realMs-slotMs > 1000 {
		p.ShownMs = slotMs
	}
	return p
}

// loadFootageFor reads a clip's sidecar.
//
// A missing or unreadable sidecar is (zero, false) rather than an error: the
// scene graph is built long after the capture ran, and a lesson resumed from a
// later stage on a machine that never recorded anything is an ordinary thing
// rather than a broken build. The caller decides what an absent clip means,
// and for footage the answer is "draw nothing", because empty browser chrome
// asserts that something was recorded when nothing was.
func loadFootageFor(l *project.Lesson, id string) (Footage, bool) {
	var f Footage
	path := filepath.Join(l.GeneratedDir(), DemosDirName, id+FootageFileSuffix)
	data, err := os.ReadFile(path)
	if err != nil {
		return Footage{}, false
	}
	if err := json.Unmarshal(data, &f); err != nil {
		return Footage{}, false
	}
	return f, true
}

// LessonFootage reads every capture sidecar a lesson has, in id order.
//
// Sidecars rather than the manifest, because the manifest records what the
// scene graph needs and this needs what the *freshness* question needs — when
// it was shot, of what version, from what origin. A clip whose sidecar is
// unreadable is skipped: this is a reporting path, and failing the whole
// listing over one bad file helps nobody.
func LessonFootage(l *project.Lesson) []Footage {
	dir := filepath.Join(l.GeneratedDir(), DemosDirName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []Footage
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), FootageFileSuffix) {
			continue
		}
		id := strings.TrimSuffix(e.Name(), FootageFileSuffix)
		if f, ok := loadFootageFor(l, id); ok {
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// CaptureBriefing describes, in plain lines, what this lesson's recordings
// actually caught — for the script stage to write narration around.
//
// This is what the capture stage running before the script buys. Every line is
// measured: the tool's own version, the real duration, the moments the recorder
// stamped. A writer given this can lead into the clip properly and can say how
// long the tool took; a writer without it is guessing at footage that does not
// exist yet, which is what the pipeline did until the stages were split.
//
// Deliberately prose rather than JSON. The consumer is a prompt, and a schema
// dropped into a prompt gets copied into the output as often as it gets read.
func CaptureBriefing(l *project.Lesson) []string {
	var out []string
	for _, f := range LessonFootage(l) {
		if f.Kind == CaptureKindPython || f.Kind == "" {
			continue // our own code running; the outline already describes it
		}
		display := f.Tool
		switch f.Kind {
		case CaptureKindTool:
			if t, ok := captureTools[f.Tool]; ok {
				display = t.Display
			}
		case CaptureKindWeb:
			if s, ok := captureSites[f.Tool]; ok {
				display = s.Display
			}
		case CaptureKindDesktop:
			if a, ok := captureApps[f.Tool]; ok {
				display = a.Display
			}
		}
		line := fmt.Sprintf("A real recording of %s", display)
		if v := shortVersion(f.ToolVersion); v != "" {
			line += " " + v
		}
		if f.DurationMs > 0 {
			line += fmt.Sprintf(", %d seconds long", (f.DurationMs+500)/1000)
		}
		if len(f.Frames) > 0 {
			line += fmt.Sprintf(", %d captured screen(s)", len(f.Frames))
		}
		if names := markNames(f); names != "" {
			line += ". It shows, in order: " + names
		}
		out = append(out, line)
	}
	return out
}

// markNames lists a clip's moments in order.
func markNames(f Footage) string {
	names := make([]string, 0, len(f.Marks)+len(f.Frames))
	for _, m := range f.Marks {
		names = append(names, strings.ReplaceAll(m.Name, "-", " "))
	}
	for _, fr := range f.Frames {
		names = append(names, strings.ReplaceAll(fr.Mark, "-", " "))
	}
	return strings.Join(names, ", ")
}

// buildFootage assembles the sidecar for a clip that has just been rendered.
func buildFootage(id string, kind CaptureKind, tool, toolVersion, body string, realDurationMs int, now time.Time) Footage {
	scan := scanTape(body, tapeTypingSpeedMs)
	return Footage{
		ID:          id,
		Kind:        kind,
		Tool:        tool,
		ToolVersion: toolVersion,
		CapturedAt:  now.UTC().Format(time.RFC3339),
		DurationMs:  realDurationMs,
		TapeTimeMs:  scan.totalMs,
		Waits:       scan.waits,
		Marks:       resolveMarks(scan, realDurationMs),
	}
}

// LessonFootageByID reads one clip's sidecar.
func LessonFootageByID(l *project.Lesson, id string) (Footage, bool) {
	return loadFootageFor(l, id)
}

// FootageMarkNames lists a clip's moments in order — what a narration is
// allowed to talk about.
func FootageMarkNames(f Footage) []string {
	return markNamesOf(f)
}
