package pipeline

// The approval template: the moment a tool asks permission, and what each answer
// costs you.
//
// This is the one frame that every video about agentic tools talks through and
// none of them draws. The reference clip this came from is titled around exactly
// this decision — approve each edit, or let it run — and it makes the point with a
// presenter talking to camera, because there is no picture for it. There is now.
//
// The shape is a proposal and its answers. At the top, in a window, the thing
// being asked for: a command, an edit, a file being written. Underneath, the
// answers as rows — and the important part, the CONSEQUENCE of each, because that
// is the whole content. "Yes" and "Yes, and stop asking" are one word apart and
// worlds apart in what they hand over, and a viewer who has only ever seen the two
// words has not been taught anything.
//
// It is a decision frame, which is why it sits in this family: `duel` weighs two
// products, `cards` weighs several, and this weighs the answers to one question.
//
// Three rules earn the shape.
//
// The proposal must be a CONCRETE action. "Claude wants to make some changes" is
// the dialog everybody clicks through without reading; "rm -rf build/" is the one
// they stop at. The template's value is entirely in the specificity of the thing
// being asked, so the ask is set in mono and read as a literal.
//
// Every answer carries its consequence, and one of them is marked as the RISK.
// Exactly one — a frame where two answers are both flagged dangerous has not
// helped anybody choose, and a frame where none is flagged is a menu rather than a
// warning. The risky answer is usually the convenient one, which is the lesson.
//
// And there is a RECOMMENDED answer, which may be the risky one. That combination
// is deliberately allowed: "let it run, but only in a repo you can throw away" is
// the honest recommendation for most people most of the time, and a template that
// could not say it would be a safety poster rather than teaching.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "approval",
		Category: CatDecisions,
		Since:    SinceV9,
		Family:   FamilyShowroom,
		Title:    "The permission gate",
		Description: "A tool asking to do something, and the two or three answers with what each one actually hands over — one of them marked as the risk, one as the pick. " +
			"Reach for it wherever a viewer is about to be shown a prompt they will otherwise click through.",
		Example:    "Approve every edit, or let it run: what you are choosing",
		PromptFile: snippetApprovalTemplateName,
		NeedsCode:  false,
		// The ask, a beat per answer, the pick. Two answers is four beats.
		MinTargetSec:      28,
		DefaultTargetSec:  45,
		MaxBeats:          7,
		IdealWordsPerBeat: 24,
		Owns:              beatFields{Approval: true},
		OwnsPlan:          planFields{Approval: true},
		Normalize:         normalizeApprovalPlan,
		Validate:          validateApprovalPlan,
		Scenes:            approvalScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":               strings.Join(ApprovalShows(), ", "),
				"MinAnswers":          minApprovalAnswers,
				"MaxAnswers":          maxApprovalAnswers,
				"MaxAskChars":         maxApprovalAskChars,
				"MaxContextWords":     maxApprovalContextWords,
				"MaxLabelWords":       maxApprovalLabelWords,
				"MaxConsequenceWords": maxApprovalConsequenceWords,
				"MaxCloserWords":      maxApprovalCloserWords,
			}
		},
	})
}

const snippetApprovalTemplateName = "snippet_approval.tmpl"

const (
	// Two answers is a gate; four is a settings screen. Three is the real shape of
	// this prompt in every tool that has one — yes, yes-and-stop-asking, no.
	minApprovalAnswers = 2
	maxApprovalAnswers = 3

	// The ask itself, set in mono as a literal. Characters rather than words
	// because it is a command, not a sentence.
	maxApprovalAskChars = 56
	// The line above the ask: what the tool is in the middle of.
	maxApprovalContextWords = 12
	// An answer's label — what the button says.
	maxApprovalLabelWords = 5
	// What that answer actually hands over. The content of the whole frame.
	maxApprovalConsequenceWords = 14
	// The line under the finished gate.
	maxApprovalCloserWords = 16
)

// approvalShows is the closed vocabulary of what a beat does.
var approvalShows = map[string]bool{
	// The ask up, the answers listed but none singled out. The opener.
	"ask": true,
	// Answer At lit, its consequence read out.
	"answer": true,
	// The pick marked and the closing line. The closer.
	"pick": true,
}

// ApprovalShows returns the beat vocabulary sorted.
func ApprovalShows() []string {
	out := make([]string, 0, len(approvalShows))
	for k := range approvalShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ApprovalSpec is the gate.
type ApprovalSpec struct {
	// Tool is the thing doing the asking — "Claude Code", "the agent".
	Tool string `json:"tool,omitempty"`
	// Context is what it is in the middle of, above the ask.
	Context string `json:"context,omitempty"`
	// Ask is the concrete action, set as a literal. Required and required to be
	// concrete — see the file header.
	Ask string `json:"ask"`
	// Answers are the rows, in the order the prompt offers them.
	Answers []ApprovalAnswer `json:"answers"`
	// Pick indexes Answers: the one being recommended. May be the risky one.
	Pick int `json:"pick"`
	// Closer is the line under the finished gate.
	Closer string `json:"closer,omitempty"`
}

// ApprovalAnswer is one way out of the gate.
type ApprovalAnswer struct {
	// Label is what the answer says — "Yes", "Yes, and don't ask again", "No".
	Label string `json:"label"`
	// Consequence is what it hands over. The reason this template exists.
	Consequence string `json:"consequence"`
	// Risk marks the answer that gives away the most. Exactly one answer carries
	// it; see the file header.
	Risk bool `json:"risk,omitempty"`
}

// ApprovalBeat is one shot.
type ApprovalBeat struct {
	// Show is an approvalShows name.
	Show string `json:"show"`
	// At indexes ApprovalSpec.Answers, for an "answer" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults to an answer being read.
func (b ApprovalBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if approvalShows[s] {
		return s
	}
	return "answer"
}

func normalizeApprovalPlan(p *SnippetPlan) {
	a := p.Approval
	if a == nil {
		return
	}
	a.Tool = clampWords(collapseSpaces(a.Tool), 4)
	a.Context = clampWords(collapseSpaces(a.Context), maxApprovalContextWords)
	// Collapsed but NOT word-clamped: it is a command, and clipping it at a word
	// boundary would silently change what it does.
	a.Ask = clampCodeLine(collapseSpaces(a.Ask), maxApprovalAskChars)
	a.Closer = clampWords(collapseSpaces(a.Closer), maxApprovalCloserWords)

	answers := make([]ApprovalAnswer, 0, len(a.Answers))
	for _, an := range a.Answers {
		an.Label = clampWords(collapseSpaces(an.Label), maxApprovalLabelWords)
		an.Consequence = clampWords(collapseSpaces(an.Consequence), maxApprovalConsequenceWords)
		if an.Label != "" && len(answers) < maxApprovalAnswers {
			answers = append(answers, an)
		}
	}
	a.Answers = answers

	if a.Pick < 0 || a.Pick >= len(a.Answers) {
		a.Pick = 0
	}

	for i := range p.Beats {
		b := p.Beats[i].Approval
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "answer" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(a.Answers); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateApprovalPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Approval: true}); err != nil {
		return err
	}

	a := p.Approval
	if a == nil {
		return fmt.Errorf("the plan has no gate — this template is one permission prompt and its answers, so the ask is the clip")
	}
	if strings.TrimSpace(a.Ask) == "" {
		return fmt.Errorf("the gate asks for nothing. The whole value of this frame is the specificity of the action — a command, a file being written, an edit — because that is what makes a viewer stop and read instead of clicking through")
	}
	if n := len(a.Answers); n < minApprovalAnswers || n > maxApprovalAnswers {
		return fmt.Errorf("the gate has %d answers, want %d-%d. One answer is not a choice; past %d it is a settings screen rather than a prompt",
			n, minApprovalAnswers, maxApprovalAnswers, maxApprovalAnswers)
	}

	risks := 0
	for i, an := range a.Answers {
		if strings.TrimSpace(an.Consequence) == "" {
			return fmt.Errorf("answer %d (%q) says nothing about what it hands over. That consequence IS the content of this frame — \"Yes\" and \"Yes, and stop asking\" are one word apart and worlds apart, and a viewer shown only the words has learned nothing",
				i, an.Label)
		}
		if an.Risk {
			risks++
		}
	}
	if risks == 0 {
		return fmt.Errorf("no answer is marked as the risk, so the frame is a menu rather than a gate. One of these hands over more than the others — usually the convenient one, which is the lesson — so mark it with \"risk\": true")
	}
	if risks > 1 {
		return fmt.Errorf("%d answers are marked as the risk. If every way out is dangerous the frame has not helped anybody choose; mark the ONE that gives away the most", risks)
	}

	if p.Beats[0].Approval == nil || p.Beats[0].Approval.ResolvedShow() != "ask" {
		return fmt.Errorf("beat %q does not open on the ask. The viewer has to know what is being requested before any answer means anything — open with {\"show\": \"ask\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Approval == nil || last.Approval.ResolvedShow() != "pick" {
		return fmt.Errorf("beat %q does not close on the pick. A gate that lays out three answers and recommends none has described a dialog rather than taught a decision — end with {\"show\": \"pick\"}", last.ID)
	}

	next := 0
	for _, b := range p.Beats {
		d := b.Approval
		if d == nil {
			return fmt.Errorf("beat %q has no approval direction — every beat raises the ask, reads one answer, or lands the pick", b.ID)
		}
		if d.ResolvedShow() != "answer" {
			continue
		}
		if d.At < 0 || d.At >= len(a.Answers) {
			return fmt.Errorf("beat %q reads answer %d, which does not exist — the gate has answers 0-%d", b.ID, d.At, len(a.Answers)-1)
		}
		if next >= len(a.Answers) {
			return fmt.Errorf("beat %q reads an answer when all %d have been read. Each answer is read once, in the order the prompt offers them", b.ID, len(a.Answers))
		}
		if d.At != next {
			return fmt.Errorf("beat %q reads answer %d (%q) when answer %d (%q) is next. The rows are read in order — a prompt whose options are discussed out of order is one the viewer cannot follow",
				b.ID, d.At, a.Answers[d.At].Label, next, a.Answers[next].Label)
		}
		next++
	}
	if next != len(a.Answers) {
		return fmt.Errorf("the clip reads %d of %d answers, so %q is on screen with nothing said about what it does — which is exactly the state this template exists to fix",
			next, len(a.Answers), a.Answers[next].Label)
	}
	return nil
}

// approvalScenes lays the clip out as ONE scene: the gate persists and the steps
// move the light between its rows.
func approvalScenes(in SnippetSceneInput) ([]Scene, error) {
	a := in.Plan.Approval
	if a == nil {
		return nil, fmt.Errorf("the plan has no gate")
	}
	if len(a.Answers) == 0 {
		return nil, fmt.Errorf("the gate has no answers")
	}

	answers := make([]map[string]any, len(a.Answers))
	for i, an := range a.Answers {
		answers[i] = map[string]any{
			"label":       an.Label,
			"consequence": an.Consequence,
			"risk":        an.Risk,
		}
	}

	// `read` counts the rows already covered, so the closing frame can show every
	// consequence at once rather than only the last one spoken about.
	read := 0
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Approval == nil {
			return nil, fmt.Errorf("beat %q has no approval direction", beat.ID)
		}
		show := beat.Approval.ResolvedShow()
		step := map[string]any{"startMs": startMs, "endMs": endMs, "show": show, "at": -1}
		switch show {
		case "answer":
			at := beat.Approval.At
			if at < 0 || at >= len(a.Answers) {
				return nil, fmt.Errorf("beat %q reads answer %d, which does not exist", beat.ID, at)
			}
			step["at"] = at
			read = at + 1
		case "pick":
			read = len(a.Answers)
		}
		step["read"] = read
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneApproval,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":   in.Plan.Title,
			"tool":    a.Tool,
			"context": a.Context,
			"ask":     a.Ask,
			"answers": answers,
			"pick":    a.Pick,
			"closer":  a.Closer,
			"steps":   steps,
		}),
	}}, nil
}
