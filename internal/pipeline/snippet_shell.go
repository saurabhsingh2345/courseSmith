package pipeline

// The shell template: a terminal session, watched at reading speed.
//
// The terminal is where a CS foundations course stops being theory, and the
// catalog had no way to put one on screen without the full weight of a code
// template — an editor, a verify stage, a program that has to run. Most of
// what a beginner needs from the terminal is not a program at all: it is
// watching a command get typed, seeing what comes back, and being told which
// flag did the work. That is a picture, not a computation.
//
// So this template draws the session and is honest about what it is: the
// output is AUTHORED, written to illustrate what the command typically prints,
// not captured from a run. That honesty lives in the description and in the
// prompt, because a template that pretended its output was real would poison
// the trust the verify-stage templates spend so much effort earning.
//
// The validators protect the one property a terminal has that a slide does
// not: it is append-only, and it happens in order.
//
// **A command is typed before its output lands.** Output appearing before
// anyone typed is a terminal running backwards, and it is exactly what a model
// shuffling beats produces.
//
// **Output follows the command still at the prompt.** A terminal appends; it
// cannot show an earlier command's output after a later command has been
// typed. Breaking this draws a session no shell has ever produced.
//
// **A command is a command.** A line opening with "#" is a comment — stage
// direction, not something anyone runs — and a session of narrated comments
// is a slide wearing a terminal's chrome.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "shell",
		Category:    CatCode,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "At the prompt",
		Description: "A terminal session: the prompt, a command typed in character by character, the output it prints, and a caption on the flag that matters. The output is authored to illustrate, not executed. Reach for it when the lesson is a command and what it shows you.",
		Example:     "ls -la: reading the long listing",
		PromptFile:  snippetShellTemplateName,
		NeedsCode:   false,
		// A prompt, at least one command with its output, and the recap.
		MinTargetSec:     35,
		DefaultTargetSec: 50,
		MaxBeats:         8,
		Owns:             beatFields{Shell: true},
		OwnsPlan:         planFields{Shell: true},
		Normalize:        normalizeShellPlan,
		Validate:         validateShellPlan,
		Scenes:           shellScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":          strings.Join(MetricRoles(), ", "),
				"Shows":          strings.Join(ShellShows(), ", "),
				"MinEntries":     minShellEntries,
				"MaxEntries":     maxShellEntries,
				"MaxCmdChars":    maxShellCmdChars,
				"MaxOutputLines": maxShellOutputLines,
				"MaxOutputChars": maxShellOutputChars,
				"MaxNoteWords":   maxShellNoteWords,
				"MaxHostWords":   maxShellHostWords,
			}
		},
	})
}

const snippetShellTemplateName = "snippet_shell.tmpl"

const (
	// One command is a perfectly good clip — most terminal lessons are one
	// command examined closely. Past four the session outgrows the frame and
	// each command gets too little of the voice to be examined at all.
	minShellEntries = 1
	maxShellEntries = 4

	// Sixty characters is a command that fits the terminal card at a size a
	// phone can read; longer commands are two lessons pretending to be one.
	maxShellCmdChars = 60

	// Eight lines of output fill the card without scrolling the command that
	// produced them out of view; seventy characters is the width the mono
	// type affords at readable size.
	maxShellOutputLines = 8
	maxShellOutputChars = 70

	// The note is an annotation riding the line, not a paragraph beside it.
	maxShellNoteWords = 12

	// The prompt's host label — "ubuntu", "web-01". Two words is already
	// generous for something that lives in a title bar.
	maxShellHostWords = 2
)

// shellShows is the closed vocabulary of what a beat does.
var shellShows = map[string]bool{
	// The empty terminal, prompt blinking. The first beat.
	"prompt": true,
	// Entry At's command types in, character by character.
	"type": true,
	// Entry At's output lines land, and its note caption shows.
	"output": true,
	// The whole session scrolled into view. The last beat.
	"recap": true,
}

// ShellShows returns the beat vocabulary sorted.
func ShellShows() []string {
	out := make([]string, 0, len(shellShows))
	for k := range shellShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ShellSpec is the session: one terminal, its host, and the commands run in it.
type ShellSpec struct {
	// Host is the prompt's host label ("" takes "ubuntu").
	Host string `json:"host,omitempty"`
	// Entries are the commands, in the order they are typed.
	Entries []ShellEntry `json:"entries"`
}

// ResolvedHost returns the host label, defaulting to the distribution every
// beginner tutorial in the world seems to be running.
func (s *ShellSpec) ResolvedHost() string {
	if h := strings.TrimSpace(s.Host); h != "" {
		return h
	}
	return "ubuntu"
}

// ShellEntry is one command and what it printed.
type ShellEntry struct {
	// Cmd is the command line, exactly as typed.
	Cmd string `json:"cmd"`
	// Output is what came back — illustrative, authored to show the shape of
	// the real thing. May be empty: plenty of commands print nothing.
	Output []string `json:"output,omitempty"`
	// Note is the one thing to notice — usually the flag that matters.
	Note string `json:"note,omitempty"`
}

// ShellBeat is one move: what this beat shows.
type ShellBeat struct {
	// Show is a shellShows name.
	Show string `json:"show"`
	// At indexes ShellSpec.Entries, for "type" and "output" beats.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to typing —
// which is what most beats of this template do.
func (b ShellBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if shellShows[s] {
		return s
	}
	return "type"
}

func normalizeShellPlan(p *SnippetPlan) {
	s := p.Shell
	if s == nil {
		return
	}
	s.Host = clampWords(collapseSpaces(s.Host), maxShellHostWords)

	entries := make([]ShellEntry, 0, len(s.Entries))
	for _, e := range s.Entries {
		e.Cmd = clampChars(strings.TrimSpace(e.Cmd), maxShellCmdChars)
		lines := make([]string, 0, len(e.Output))
		for _, line := range e.Output {
			line = clampChars(strings.TrimRight(line, " \t"), maxShellOutputChars)
			if len(lines) < maxShellOutputLines {
				lines = append(lines, line)
			}
		}
		e.Output = lines
		e.Note = clampWords(collapseSpaces(e.Note), maxShellNoteWords)
		if len(entries) < maxShellEntries {
			entries = append(entries, e)
		}
	}
	s.Entries = entries

	for i := range p.Beats {
		b := p.Beats[i].Shell
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "type" && b.Show != "output" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(s.Entries); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateShellPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Shell: true}); err != nil {
		return err
	}

	s := p.Shell
	if s == nil {
		return fmt.Errorf("the plan has no session — this template is a terminal, the commands typed into it, and what they printed")
	}
	if n := len(s.Entries); n < minShellEntries || n > maxShellEntries {
		return fmt.Errorf("the session runs %d command(s), want %d-%d. One command examined closely is a fine clip; past %d the session outgrows the frame and no command gets enough of the voice to be examined at all",
			n, minShellEntries, maxShellEntries, maxShellEntries)
	}
	for i, e := range s.Entries {
		cmd := strings.TrimSpace(e.Cmd)
		if cmd == "" {
			return fmt.Errorf("entry %d has no command. A terminal with nothing typed into it is an empty window — give it the command line, exactly as a learner would type it", i)
		}
		if strings.HasPrefix(cmd, "#") {
			return fmt.Errorf("entry %d's command %q opens with \"#\", which makes it a comment — stage direction, not something anyone runs. Put the explanation in the narration or the note, and give the entry a real command",
				i, cmd)
		}
		if n := len(e.Output); n > maxShellOutputLines {
			return fmt.Errorf("entry %d prints %d lines of output, want at most %d — more scrolls the command that produced them out of view. Cut to the lines the narration actually reads",
				i, n, maxShellOutputLines)
		}
	}

	if p.Beats[0].Shell == nil || p.Beats[0].Shell.ResolvedShow() != "prompt" {
		return fmt.Errorf("beat %q does not open on the empty prompt. The blank terminal is what makes the first keystroke an event — a session that starts mid-scroll has spent its opening",
			p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Shell == nil || last.Shell.ResolvedShow() != "recap" {
		return fmt.Errorf("beat %q does not close on the recap. The last frame is the whole session in view — it is the frame a learner screenshots, and without it the clip ends mid-command",
			p.Beats[len(p.Beats)-1].ID)
	}

	typed := map[int]bool{}
	shown := map[int]bool{}
	lastTyped := -1
	for i, b := range p.Beats {
		if b.Shell == nil {
			return fmt.Errorf("beat %q has no shell direction — every beat shows the prompt, types a command, lands its output, or recaps the session", b.ID)
		}
		switch b.Shell.ResolvedShow() {
		case "prompt":
			if i != 0 {
				return fmt.Errorf("beat %q shows the empty prompt part-way through. A terminal is append-only — once something has been typed, the blank screen does not come back", b.ID)
			}
		case "recap":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q recaps the session before it is over — the recap is the closing frame, after every command has run", b.ID)
			}
		case "type":
			at := b.Shell.At
			if at < 0 || at >= len(s.Entries) {
				return fmt.Errorf("beat %q types entry %d, which does not exist", b.ID, at)
			}
			if typed[at] {
				return fmt.Errorf("beat %q types entry %d again; a command is typed once", b.ID, at)
			}
			if at != lastTyped+1 {
				return fmt.Errorf("beat %q types entry %d but entry %d is next. A terminal happens in order — the commands are typed exactly as listed, none skipped",
					b.ID, at, lastTyped+1)
			}
			typed[at] = true
			lastTyped = at
		case "output":
			at := b.Shell.At
			if at < 0 || at >= len(s.Entries) {
				return fmt.Errorf("beat %q lands output for entry %d, which does not exist", b.ID, at)
			}
			if !typed[at] {
				return fmt.Errorf("beat %q lands output for entry %d before anyone typed it. A terminal runs forward: the command is typed, then it prints — output first is a session running backwards",
					b.ID, at)
			}
			if at != lastTyped {
				return fmt.Errorf("beat %q lands output for entry %d while entry %d is at the prompt. A terminal appends — output can only follow the command most recently typed",
					b.ID, at, lastTyped)
			}
			if shown[at] {
				return fmt.Errorf("beat %q lands entry %d's output again; a command prints once", b.ID, at)
			}
			shown[at] = true
		}
	}
	if len(typed) != len(s.Entries) {
		return fmt.Errorf("%d of the %d commands are never typed. A command nobody typed is dead weight in the plan — give it a beat or cut it",
			len(s.Entries)-len(typed), len(s.Entries))
	}
	// And whatever a command prints has to be spoken to.
	//
	// The recap lands every entry's output at once, so output the beats never
	// covered still reaches the screen — it simply arrives in the closing frame
	// with nothing said about it. That is the "drawn but never explained"
	// failure the rest of this batch has a coverage rule for, and here it is
	// worse than a silent diagram: terminal output is the evidence the command
	// did what the clip claims, and evidence that flashes up unremarked in the
	// last two seconds has not been shown to anybody.
	//
	// An entry with no output needs no beat — plenty of real commands succeed
	// silently, and demanding a beat for an empty result would be asking the
	// model to narrate a blank.
	for i, e := range s.Entries {
		if len(e.Output) > 0 && !shown[i] {
			return fmt.Errorf("command %d prints %d line(s) that no beat ever lands. The recap will put them on screen in the closing frame regardless, so as written the clip shows output it never explains — give entry %d a beat with {\"show\": \"output\"}, or cut its output",
				i, len(e.Output), i)
		}
	}
	return nil
}

// shellScenes lays the clip out as ONE scene: one terminal, its state per beat.
func shellScenes(in SnippetSceneInput) ([]Scene, error) {
	s := in.Plan.Shell
	if s == nil {
		return nil, fmt.Errorf("the plan has no session")
	}

	entries := make([]map[string]any, len(s.Entries))
	for i, e := range s.Entries {
		out := e.Output
		if out == nil {
			out = []string{}
		}
		entries[i] = map[string]any{
			"cmd":    e.Cmd,
			"output": out,
			"note":   e.Note,
		}
	}

	// Which entries have been typed and which have printed, accumulated in Go
	// so the renderer draws any frame from its own step alone.
	typed := map[int]bool{}
	shown := map[int]bool{}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Shell == nil {
			return nil, fmt.Errorf("beat %q has no shell direction", beat.ID)
		}
		show := beat.Shell.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "type":
			typed[beat.Shell.At] = true
			step["at"] = beat.Shell.At
		case "output":
			shown[beat.Shell.At] = true
			step["at"] = beat.Shell.At
		case "recap":
			// The recap shows everything, including output no beat landed.
			for j := range s.Entries {
				shown[j] = true
			}
		}
		typedList := make([]int, 0, len(typed))
		for j := range typed {
			typedList = append(typedList, j)
		}
		sort.Ints(typedList)
		shownList := make([]int, 0, len(shown))
		for j := range shown {
			shownList = append(shownList, j)
		}
		sort.Ints(shownList)
		step["typed"] = typedList
		step["shown"] = shownList
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneShell,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":   in.Plan.Title,
			"host":    s.ResolvedHost(),
			"entries": entries,
			"steps":   steps,
		}),
	}}, nil
}
