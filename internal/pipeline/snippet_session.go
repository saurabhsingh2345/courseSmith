package pipeline

// The session template: an agent session, held on screen and read.
//
// This is the picture a tutorial about working with an agent is actually made of,
// and nothing in the catalog could draw it. `shell` draws a terminal CARD — a
// small window in the middle of a composition, showing a command and what it
// printed — which is right for "here is what this flag does" and wrong here for
// three reasons.
//
// A SESSION IS THE SUBJECT, NOT AN ILLUSTRATION. It fills the frame. The viewer
// is not glancing at a terminal beside an explanation, they are reading the
// session while somebody talks them through it, the way you read over a
// colleague's shoulder. That changes the size, the type scale, and what else is
// allowed on the frame: nothing.
//
// A SESSION GROWS AND SCROLLS. A shell card appends two commands and stops. Four
// minutes of session runs past the bottom of any window, so the transcript has to
// move — new work arrives at the bottom and the top slides away, which is the
// single strongest signal that this is a real tool and not a diagram of one.
//
// A SESSION ASKS QUESTIONS. This is the part with no precedent at all. An agent
// stops and offers you numbered choices, and the whole lesson of a steering
// tutorial is which one to pick and what it costs. So a menu is a first-class
// event here rather than a separate template you cut away to: the question
// appears IN the session, where a viewer will meet it.
//
// What it is not: a screen recording. Nothing here is captured — the events are
// authored, the way `shell` authors its output, and narration must not claim
// otherwise. `footage` is the template for a real recording, and it earns a trust
// this one does not.
//
// The composer line at the bottom is the detail that sells it. A terminal card
// ends at its last line of output; a session always has somewhere to type, with a
// caret waiting in it, so the frame reads as a conversation in progress rather
// than a transcript of one that finished.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "session",
		Category: CatCode,
		Since:    SinceV10,
		Family:   FamilyAtelier,
		Title:    "An agent session",
		Description: "A full-frame agent session that grows as you walk it: prompts typed at the composer, tool work landing under them, and the numbered questions the agent stops to ask. " +
			"Reach for it when the lesson is how to READ and STEER a session; `shell` is the small card for one command.",
		Example:    "what plan mode shows you before it edits",
		PromptFile: snippetSessionTemplateName,
		NeedsCode:  false,
		// A session needs room. Under about 25 seconds the window arrives, one
		// thing happens and it cuts, which wastes the one shot in the catalog
		// that can hold a frame for a minute without tiring.
		MinTargetSec:     25,
		DefaultTargetSec: 55,
		// Long, deliberately: this is the template a four-minute tutorial spends
		// most of its running time inside, and the beat ceiling rather than the
		// clock is what keeps it honest.
		MaxTargetSec: 200,
		// The window, up to six events, the chip, the pull-back.
		MaxBeats: 9,
		// Higher than most. A beat here is narration over a held window with one
		// thing changing in it, which affords a real sentence or two — the failure
		// mode this fights is a caption per beat over a picture that deserves
		// paragraphs.
		IdealWordsPerBeat: 30,
		Owns:              beatFields{Session: true},
		OwnsPlan:          planFields{Session: true},
		Normalize:         normalizeSessionPlan,
		Validate:          validateSessionPlan,
		Scenes:            sessionScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":          strings.Join(SessionShows(), ", "),
				"Kinds":          strings.Join(SessionKinds(), ", "),
				"MinEvents":      minSessionEvents,
				"MaxEvents":      maxSessionEvents,
				"MaxTextChars":   maxSessionTextChars,
				"MaxLines":       maxSessionLines,
				"MaxLineChars":   maxSessionLineChars,
				"MaxOptions":     maxSessionOptions,
				"MaxOptionChars": maxSessionOptionChars,
				"MaxHeaderLines": maxSessionHeaderLines,
				"MaxNoteWords":   maxSessionNoteWords,
				"MaxChipChars":   maxSessionChipChars,
			}
		},
	})
}

const snippetSessionTemplateName = "snippet_session.tmpl"

const (
	// Two events is a session doing something; past six the window is scrolling
	// faster than a viewer reads and the narration cannot cover them all.
	minSessionEvents = 2
	maxSessionEvents = 6

	// The window is nearly frame-width, so these are generous compared with the
	// shell card's — but a line that wraps three times is still a paragraph
	// pretending to be terminal output.
	maxSessionTextChars   = 78
	maxSessionLines       = 8
	// A prose block — a written plan, a set of acceptance criteria — earns more.
	maxSessionProseLines = 16
	maxSessionLineChars   = 84
	maxSessionOptions     = 5
	maxSessionOptionChars = 52
	// The header block: app, model, working directory. Four lines is what a real
	// one carries and more is furniture.
	maxSessionHeaderLines = 4
	maxSessionHeaderChars = 58
	// The mark under a menu — "6 files changed  +34 -34".
	maxSessionMarkChars = 34
	// The margin annotation, and the status chip in the corner.
	maxSessionNoteWords = 12
	maxSessionChipChars = 28
	maxSessionAppWords  = 3
	maxSessionHintWords = 8
)

// sessionShows is the closed vocabulary of what a beat does.
var sessionShows = map[string]bool{
	// The window arrives with its header and an empty composer. The first beat.
	"open": true,
	// Event At lands in the transcript.
	"event": true,
	// The status chip lights in the corner.
	"chip": true,
	// The camera settles and the whole session is read at once. The last beat.
	"whole": true,
}

// sessionKinds is what one event in the transcript is.
var sessionKinds = map[string]bool{
	// A line typed at the composer by the person driving.
	"ask": true,
	// The agent talking back in prose.
	"say": true,
	// Tool work: a bulleted headline with what it did indented under it.
	"tool": true,
	// The agent stopping to ask, with numbered answers.
	"menu": true,
	// The box a session prints on start: a greeting, the account and model, and
	// the two columns of tips and recent activity beside it.
	"welcome": true,
	// The working line: a verb, how long it has been going, and the tokens it
	// has moved. Nothing in a transcript says "this is running" like one.
	"spin": true,
	// Several subagents working at once, each with its own progress under a
	// shared headline.
	"agents": true,
}

// SessionShows returns the beat vocabulary sorted.
func SessionShows() []string {
	out := make([]string, 0, len(sessionShows))
	for k := range sessionShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// SessionKinds returns the event vocabulary sorted.
func SessionKinds() []string {
	out := make([]string, 0, len(sessionKinds))
	for k := range sessionKinds {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// SessionSpec is the window and everything that happens in it.
type SessionSpec struct {
	// App is the label in the window chrome.
	App string `json:"app,omitempty"`
	// Header is the block the session prints when it starts: what is running,
	// which model, which directory. Optional.
	Header []string `json:"header,omitempty"`
	// Hint is the footer line under the composer — the mode reminder a real
	// session keeps in front of you.
	Hint string `json:"hint,omitempty"`
	// Chip is the status marker in the corner, shown by a "chip" beat.
	Chip string `json:"chip,omitempty"`
	// Branch is the working branch, shown as a pill above the footer.
	Branch string `json:"branch,omitempty"`
	// Pr is the pull request the footer names, once one exists.
	Pr string `json:"pr,omitempty"`
	// Status is the right-hand footer marker — the reasoning effort, the mode.
	Status string `json:"status,omitempty"`
	// Events are what happens, in order.
	Events []SessionEvent `json:"events"`
}

// SessionEvent is one thing that happens in the session.
type SessionEvent struct {
	// Kind is a sessionKinds name.
	Kind string `json:"kind"`
	// Text is the line: what was typed, what the agent said, the tool's headline,
	// or the question a menu asks.
	Text string `json:"text"`
	// Lines are what comes under it — output, sub-steps, continuation.
	Lines []string `json:"lines,omitempty"`
	// More is the collapsed-output line a real session prints when it has cut
	// something short: "+26 lines (ctrl+o to expand)".
	//
	// Small, and the single most convincing detail available. Output that simply
	// stops looks authored; output that says how much of itself is hidden is a
	// tool with more to show, and every viewer who has used one recognises it.
	More string `json:"more,omitempty"`
	// Aside is the right-hand column of a welcome box — tips, recent activity.
	Aside []string `json:"aside,omitempty"`
	// Foot is the key hint under a menu: "Enter to select · up/down to navigate".
	Foot string `json:"foot,omitempty"`
	// Options are a menu's numbered answers.
	Options []string `json:"options,omitempty"`
	// Pick is the option the cursor is on, counting from 0.
	Pick int `json:"pick,omitempty"`
	// Mark is the one-line summary under a menu — what answering would move.
	Mark string `json:"mark,omitempty"`
	// Track draws the restore timeline under a menu: a line with a marker at the
	// point being restored to. Only meaningful on a menu.
	Track bool `json:"track,omitempty"`
	// Note is the annotation in the margin beside this event.
	Note string `json:"note,omitempty"`
}

// ResolvedKind defaults the unknown to prose, which is the safest thing to draw:
// an unrecognised kind rendered as a tool bullet would assert that work happened.
func (e SessionEvent) ResolvedKind() string {
	k := strings.ToLower(strings.TrimSpace(e.Kind))
	if sessionKinds[k] {
		return k
	}
	return "say"
}

// SessionBeat is one shot of the session.
type SessionBeat struct {
	// Show is a sessionShows name.
	Show string `json:"show"`
	// At is which event, counting from 0. Read on an "event" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults the unknown to an event landing, which is what almost
// every beat in this template is.
func (b SessionBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if sessionShows[s] {
		return s
	}
	return "event"
}

// keepIndent collapses a line's interior whitespace but preserves the spaces it
// starts with.
//
// collapseSpaces is right for a headline and wrong for a transcript. A written
// plan indents its detail under the file it belongs to, and a list of acceptance
// criteria indents its continuations — strip that and every line lands flush
// left, so the structure the author wrote is gone and the block reads as
// undifferentiated prose. Observed on the first render of a plan: four numbered
// files and their bullets all at the same margin.
func keepIndent(s string, max int) string {
	trimmed := strings.TrimLeft(s, " \t")
	indent := len(s) - len(trimmed)
	if indent > 8 {
		indent = 8
	}
	body := clampCodeLine(collapseSpaces(trimmed), max-indent)
	if body == "" {
		return ""
	}
	return strings.Repeat(" ", indent) + body
}

func normalizeSessionPlan(p *SnippetPlan) {
	s := p.Session
	if s == nil {
		return
	}
	s.App = clampWords(collapseSpaces(s.App), maxSessionAppWords)
	s.Hint = clampWords(collapseSpaces(s.Hint), maxSessionHintWords)
	s.Chip = clampCodeLine(collapseSpaces(s.Chip), maxSessionChipChars)
	s.Branch = clampCodeLine(collapseSpaces(s.Branch), maxSessionChipChars)
	s.Pr = clampCodeLine(collapseSpaces(s.Pr), maxSessionChipChars)
	s.Status = clampCodeLine(collapseSpaces(s.Status), maxSessionChipChars)

	header := make([]string, 0, maxSessionHeaderLines)
	for _, l := range s.Header {
		if l = clampCodeLine(collapseSpaces(l), maxSessionHeaderChars); l != "" && len(header) < maxSessionHeaderLines {
			header = append(header, l)
		}
	}
	s.Header = header

	events := make([]SessionEvent, 0, len(s.Events))
	for _, e := range s.Events {
		e.Kind = e.ResolvedKind()
		e.Text = clampCodeLine(collapseSpaces(e.Text), maxSessionTextChars)
		e.Note = clampWords(collapseSpaces(e.Note), maxSessionNoteWords)
		e.Mark = clampCodeLine(collapseSpaces(e.Mark), maxSessionMarkChars)
		e.More = clampCodeLine(collapseSpaces(e.More), maxSessionLineChars)
		e.Foot = clampCodeLine(collapseSpaces(e.Foot), maxSessionLineChars)

		aside := make([]string, 0, 4)
		for _, l := range e.Aside {
			if l = clampCodeLine(collapseSpaces(l), maxSessionHeaderChars); l != "" && len(aside) < 4 {
				aside = append(aside, l)
			}
		}
		e.Aside = aside

		// A plan or a prose block is allowed more lines than a command's output:
		// the whole point of the plan step is that the plan is LONG enough to be
		// worth reading before anything is edited.
		cap := maxSessionLines
		if e.Kind == "say" {
			cap = maxSessionProseLines
		}
		lines := make([]string, 0, cap)
		for _, l := range e.Lines {
			if l = keepIndent(l, maxSessionLineChars); l != "" && len(lines) < cap {
				lines = append(lines, l)
			}
		}
		e.Lines = lines

		opts := make([]string, 0, maxSessionOptions)
		for _, o := range e.Options {
			if o = clampCodeLine(collapseSpaces(o), maxSessionOptionChars); o != "" && len(opts) < maxSessionOptions {
				opts = append(opts, o)
			}
		}
		e.Options = opts
		// Options only mean something on a menu, and a stray set on a tool event
		// would draw numbered rows under a bullet.
		if e.Kind != "menu" {
			e.Options = nil
			e.Track = false
			e.Foot = ""
			if e.Kind != "spin" {
				// spin carries its elapsed-and-tokens in Mark; nothing else but
				// a menu has a use for it.
				e.Mark = ""
			}
		}
		if e.Kind != "welcome" {
			e.Aside = nil
		}
		if e.Pick < 0 || e.Pick >= len(e.Options) {
			e.Pick = 0
		}
		if e.Text != "" && len(events) < maxSessionEvents {
			events = append(events, e)
		}
	}
	s.Events = events

	for i := range p.Beats {
		if b := p.Beats[i].Session; b != nil {
			b.Show = b.ResolvedShow()
			if b.Show != "event" {
				b.At = 0
			}
		}
	}
}

func validateSessionPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Session: true}); err != nil {
		return err
	}

	s := p.Session
	if s == nil {
		return fmt.Errorf("the plan has no session — this template is one window held on screen, so the session is the clip")
	}
	if n := len(s.Events); n < minSessionEvents || n > maxSessionEvents {
		return fmt.Errorf("the session has %d event(s) and wants %d-%d: fewer is a still frame, more scrolls past faster than anyone reads",
			n, minSessionEvents, maxSessionEvents)
	}
	for i, e := range s.Events {
		if strings.TrimSpace(e.Text) == "" {
			return fmt.Errorf("event %d has no text — every event is a line somebody could point at", i)
		}
		if e.ResolvedKind() == "menu" {
			if len(e.Options) < 2 {
				return fmt.Errorf("event %d is a menu with %d option(s). A menu is the moment the viewer has to CHOOSE, and one answer is not a choice — give it at least two",
					i, len(e.Options))
			}
			if e.Pick < 0 || e.Pick >= len(e.Options) {
				return fmt.Errorf("event %d picks option %d of %d", i, e.Pick, len(e.Options))
			}
		}
	}

	// The shape: open, walk every event in order exactly once, settle. Same rule
	// the shell template enforces and for the same reason — a session happens in
	// order, and a transcript that jumps is one no tool has ever produced.
	var (
		next   int
		chips  int
		opens  int
		wholes int
	)
	for i, b := range p.Beats {
		if b.Session == nil {
			return fmt.Errorf("beat %q has no session direction", b.ID)
		}
		show := b.Session.ResolvedShow()
		switch show {
		case "open":
			opens++
			if i != 0 {
				return fmt.Errorf("beat %q opens the window, but it is beat %d — the window arrives once, first", b.ID, i+1)
			}
		case "event":
			if b.Session.At != next {
				return fmt.Errorf("beat %q shows event %d but event %d has not happened yet. The transcript only ever grows downward, so the beats walk the events in order with none skipped and none repeated",
					b.ID, b.Session.At, next)
			}
			next++
		case "chip":
			chips++
			if strings.TrimSpace(s.Chip) == "" {
				return fmt.Errorf("beat %q lights the status chip but the session has no chip text", b.ID)
			}
		case "whole":
			wholes++
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q settles on the whole session, but %d beat(s) follow it — that is the closing frame",
					b.ID, len(p.Beats)-1-i)
			}
		}
	}
	if opens != 1 {
		return fmt.Errorf("there are %d \"open\" beats and there must be exactly one: the window arrives once", opens)
	}
	if wholes != 1 {
		return fmt.Errorf("there are %d \"whole\" beats and there must be exactly one, last: it is what the clip ends on", wholes)
	}
	if chips > 1 {
		return fmt.Errorf("the status chip lights %d times; it lights once", chips)
	}
	if next != len(s.Events) {
		return fmt.Errorf("%d of the %d events are never shown. An event in the session that the voice never reaches is one the viewer watched go past and heard nothing about",
			len(s.Events)-next, len(s.Events))
	}
	return nil
}

func sessionScenes(in SnippetSceneInput) ([]Scene, error) {
	s := in.Plan.Session
	if s == nil {
		return nil, fmt.Errorf("the plan has no session")
	}

	events := make([]map[string]any, 0, len(s.Events))
	for _, e := range s.Events {
		events = append(events, map[string]any{
			"kind":    e.ResolvedKind(),
			"text":    e.Text,
			"lines":   e.Lines,
			"more":    e.More,
			"aside":   e.Aside,
			"options": e.Options,
			"pick":    e.Pick,
			"mark":    e.Mark,
			"foot":    e.Foot,
			"track":   e.Track,
			"note":    e.Note,
		})
	}

	// shown is latched: the transcript keeps everything it has printed, which is
	// the one property that makes this a session rather than a slideshow.
	shown, chip := 0, false
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Session == nil {
			return nil, fmt.Errorf("beat %q has no session direction", beat.ID)
		}
		show := beat.Session.ResolvedShow()
		switch show {
		case "event":
			shown = beat.Session.At + 1
		case "chip":
			chip = true
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"at":      beat.Session.At,
			"shown":   shown,
			"chip":    chip,
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneSession,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"app":    s.App,
			"header": s.Header,
			"hint":   s.Hint,
			"chip":   s.Chip,
			"branch": s.Branch,
			"pr":     s.Pr,
			"status": s.Status,
			"events": events,
			"steps":  steps,
		},
	}}, nil
}
