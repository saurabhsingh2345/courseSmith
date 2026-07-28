package pipeline

// The promptloop template: the vibe-coding cycle, on screen.
//
// You write the ask, something builds it, you look at what came back, and you
// ask again — sharper, because now you have seen it be wrong in a specific way.
// That loop is the whole skill, and it is the thing every vibe-coding lesson is
// actually teaching whatever tool it is nominally about: the tools change, the
// loop does not.
//
// Nothing in the catalog could show it. `vscode` and `workspace` show an editor
// with code appearing in it, which is the half of the activity that is no longer
// the interesting half; `cast` puts a person beside a headline; `compare` weighs
// two finished things. None of them can hold a *conversation*, and a conversation
// is what this is.
//
// The clip shows no code, and that is the argument rather than an omission. A
// vibe-coding lesson whose payoff is a syntax-highlighted buffer has taught the
// old skill with new tooling in the background. What you actually look at is the
// result — did it work, what did it change, what is still wrong — so that is
// what the frame holds, and the prompts sit beside it as the thing you author.
//
// Two shape rules carry the template. The turns strictly alternate, because that
// is what a conversation is and it makes the layout a consequence of the plan
// rather than of taste. And there must be at least two prompts: one ask and one
// answer is a demo, and calling it a loop is the exact misunderstanding the
// template exists to correct.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "promptloop",
		Title:       "Prompt loop",
		Description: "The vibe-coding cycle: ask, look at what came back, ask again — with the goal pinned throughout.",
		Example:     "Getting an AI to build a landing page that actually converts",
		PromptFile:  snippetPromptLoopTemplateName,
		NeedsCode:   false,
		// Four beats is the shortest real loop — ask, answer, ask again, answer
		// — and four beats of narration will not fit inside twenty seconds.
		MinTargetSec:     25,
		DefaultTargetSec: 50,
		Owns:             beatFields{Loop: true},
		OwnsPlan:         planFields{Loop: true},
		Normalize:        normalizePromptLoopPlan,
		Validate:         validatePromptLoopPlan,
		Scenes:           promptLoopScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Statuses":       strings.Join(PromptLoopStatuses(), ", "),
				"MaxGoalWords":   maxLoopGoalWords,
				"MaxTextWords":   maxLoopTextWords,
				"MaxChangeWords": maxLoopChangeWords,
				"MaxChanges":     maxLoopChanges,
				"MinPrompts":     minLoopPrompts,
			}
		},
	})
}

const snippetPromptLoopTemplateName = "snippet_promptloop.tmpl"

const (
	maxLoopGoalWords   = 12
	maxLoopTextWords   = 22
	maxLoopChangeWords = 7
	maxLoopChanges     = 3
	// Two asks is what makes this a loop rather than a demonstration.
	minLoopPrompts = 2
)

// loopTurns is the closed vocabulary of who is speaking.
const (
	loopTurnYou = "you"
	loopTurnAI  = "ai"
)

// loopStatuses is the closed vocabulary of how an attempt came out.
//
// There is no colour in these names on purpose. The design system is three
// brand colours and what Go derives from them, with no semantic red — so the
// renderer separates them by *form* (a filled tick, a hollow ring, a crossed
// one) rather than by a literal it would have to invent, which is both a
// palette the theme cannot flip and a distinction nobody colour-blind can read.
var loopStatuses = map[string]bool{
	"ok":      true,
	"partial": true,
	"broken":  true,
}

// PromptLoopStatuses returns the vocabulary sorted, for prompts and docs.
func PromptLoopStatuses() []string {
	out := make([]string, 0, len(loopStatuses))
	for k := range loopStatuses {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// PromptLoopSpec is what the whole conversation is trying to produce. On the
// plan rather than on a beat because it is the one thing that does not change
// across the clip — that is what makes it worth pinning on screen, and pinning
// it is how the template teaches that a loop needs something to converge on.
type PromptLoopSpec struct {
	// Goal is the thing being built, in one line. It sits above the thread for
	// the whole clip.
	Goal string `json:"goal"`
}

// PromptLoopBeat is one turn of the conversation.
type PromptLoopBeat struct {
	// Turn is who is speaking: "you" or "ai".
	Turn string `json:"turn"`
	// Text is what appears in the bubble — the prompt you typed, or the one
	// line the model answers with. Not the narration: the voice explains the
	// turn, the bubble *is* the turn.
	Text string `json:"text"`
	// Changes are what this attempt actually did, shown in the result panel.
	// Only meaningful on an "ai" turn.
	Changes []string `json:"changes,omitempty"`
	// Status is how the attempt came out: a loopStatuses name. Only meaningful
	// on an "ai" turn.
	Status string `json:"status,omitempty"`
}

// ResolvedStatus returns the attempt's status, defaulting the unknown to "ok".
func (b PromptLoopBeat) ResolvedStatus() string {
	s := strings.ToLower(strings.TrimSpace(b.Status))
	if loopStatuses[s] {
		return s
	}
	return "ok"
}

func normalizePromptLoopPlan(p *SnippetPlan) {
	if l := p.Loop; l != nil {
		l.Goal = clampWords(collapseSpaces(l.Goal), maxLoopGoalWords)
	}
	for i := range p.Beats {
		b := p.Beats[i].Loop
		if b == nil {
			continue
		}
		// Whose turn it is follows from the position when the model does not
		// say, because the conversation alternates by definition — and an
		// unlabelled turn is a field left blank rather than a claim about who
		// was speaking.
		turn := strings.ToLower(strings.TrimSpace(b.Turn))
		if turn != loopTurnYou && turn != loopTurnAI {
			turn = loopTurnYou
			if i%2 == 1 {
				turn = loopTurnAI
			}
		}
		b.Turn = turn
		b.Text = clampWords(collapseSpaces(b.Text), maxLoopTextWords)

		if b.Turn == loopTurnYou {
			// A prompt has no result. Left in place these draw a status badge
			// over the panel that is still showing the *previous* attempt.
			b.Changes = nil
			b.Status = ""
			continue
		}
		b.Status = b.ResolvedStatus()
		changes := make([]string, 0, len(b.Changes))
		for _, c := range b.Changes {
			c = clampWords(collapseSpaces(c), maxLoopChangeWords)
			if c != "" && len(changes) < maxLoopChanges {
				changes = append(changes, c)
			}
		}
		b.Changes = changes
	}
}

func validatePromptLoopPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Loop: true}); err != nil {
		return err
	}

	l := p.Loop
	if l == nil || strings.TrimSpace(l.Goal) == "" {
		return fmt.Errorf("the plan has no goal — a loop needs something to converge on, and it stays on screen for the whole clip")
	}

	prompts := 0
	for i, b := range p.Beats {
		if b.Loop == nil {
			return fmt.Errorf("beat %q has no turn — every beat of this template is somebody speaking", b.ID)
		}
		if strings.TrimSpace(b.Loop.Text) == "" {
			return fmt.Errorf("beat %q has an empty bubble — write the prompt you would actually type, or the line the model answers with", b.ID)
		}
		// The conversation alternates. This is not a style rule: the layout
		// stacks turns down one column, and two prompts in a row is a thread
		// where nothing answered the first one.
		want := loopTurnYou
		if i%2 == 1 {
			want = loopTurnAI
		}
		if b.Loop.Turn != want {
			if i == 0 {
				return fmt.Errorf("the clip opens with the model speaking. You start the loop — the first beat is the prompt")
			}
			return fmt.Errorf("beat %q is a %q turn where the conversation expects %q. Turns alternate: you ask, it answers, you ask again",
				b.ID, b.Loop.Turn, want)
		}
		if b.Loop.Turn == loopTurnYou {
			prompts++
		}
	}
	// The rule the template is named after.
	if prompts < minLoopPrompts {
		return fmt.Errorf("there is only %d prompt in this clip. One ask and one answer is a demo; the loop is what happens when you look at the result and ask again, so write at least %d",
			prompts, minLoopPrompts)
	}
	// A thread ending on a prompt is one nobody answered.
	if last := p.Beats[len(p.Beats)-1].Loop; last.Turn != loopTurnAI {
		return fmt.Errorf("the clip ends on a prompt nobody answered — close on what came back")
	}
	return nil
}

// promptLoopScenes lays the clip out as ONE scene: the thread and the result
// panel are on screen throughout, and the beats only add turns to them.
func promptLoopScenes(in SnippetSceneInput) ([]Scene, error) {
	l := in.Plan.Loop
	if l == nil {
		return nil, fmt.Errorf("the plan has no goal")
	}

	turns := make([]map[string]any, 0, len(in.Plan.Beats))
	attempt := 0
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Loop == nil {
			return nil, fmt.Errorf("beat %q has no turn", beat.ID)
		}
		turn := map[string]any{
			"who":     beat.Loop.Turn,
			"text":    beat.Loop.Text,
			"startMs": startMs,
			"endMs":   endMs,
		}
		if beat.Loop.Turn == loopTurnAI {
			attempt++
			// The attempt number is what makes the loop legible without a
			// diagram of one: the same panel, filling in again, counting up.
			turn["attempt"] = attempt
			turn["status"] = beat.Loop.ResolvedStatus()
			if len(beat.Loop.Changes) > 0 {
				turn["changes"] = beat.Loop.Changes
			}
		}
		turns = append(turns, turn)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    ScenePromptLoop,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title": in.Plan.Title,
			"goal":  l.Goal,
			"turns": turns,
		},
	}}, nil
}
