package pipeline

// The quiz template: a question posed, held, and answered.
//
// The one template in the catalog built around the viewer *doing* something.
// Every other one is watched; this one asks, waits, and only then tells — which
// is retrieval practice, the most reliably effective thing known about how
// people learn, and something the catalog could not express at all.
//
// The gap is the feature. A clip that asks and immediately answers teaches
// nothing the answer alone would not have, so the shape below is enforced
// rather than suggested: there must be at least one beat between the question
// and the reveal, and it must be a beat where nothing new appears. That silence
// is the entire mechanism.
//
// The other half is the distractors. A wrong answer nobody would pick is
// decoration; saying *why* each wrong one is tempting is where the learning
// actually is, which is why `why` is required for every option and not just for
// the right one.
//
// Note for later: the course pipeline already generates exactly this data —
// quiz.json carries a question taxonomy with distractors scored as
// misconceptions — and it has never been on screen. This scene is what would
// let a lesson render its own quiz.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "quiz",
		Title:       "Quiz",
		Description: "A question, a moment to think, then the answer — and why the wrong ones were tempting.",
		Example:     "Test whether people know what len() returns for a nested list",
		PromptFile:  snippetQuizTemplateName,
		NeedsCode:   false,
		// A quiz needs room to breathe: ask, hold, reveal, and at least one
		// explanation is four beats, and rushing the hold defeats the template.
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		Validate:         validateQuizPlan,
		Scenes:           quizScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":            strings.Join(QuizShowNames(), ", "),
				"MinOptions":       minQuizOptions,
				"MaxOptions":       maxQuizOptions,
				"MaxQuestionWords": maxQuestionWords,
				"MaxOptionWords":   maxOptionWords,
			}
		},
	})
}

const snippetQuizTemplateName = "snippet_quiz.tmpl"

// Option bounds.
//
// Three is the floor because two is a coin flip and a fifty-fifty guess is not
// retrieval. Five is the ceiling because the options have to be read *and
// held in mind* during the pause, and nobody holds six.
const (
	minQuizOptions = 3
	maxQuizOptions = 5
	// A question long enough to need re-reading has already lost the pause it
	// was going to buy.
	maxQuestionWords = 22
	maxOptionWords   = 12
)

// quizShows is what a beat can do to the question on screen.
var quizShows = map[string]bool{
	// Pose it: the question lands and the options arrive under it.
	"ask": true,
	// Hold. Nothing new appears; the viewer is answering. This is the beat the
	// template exists for and the one a model will try to skip.
	"think": true,
	// Tell them: the right option lights, the wrong ones recede.
	"reveal": true,
	// Take one option and say why it was tempting (or why it is right).
	"explain": true,
}

// QuizShowNames returns the vocabulary sorted, for the prompt.
func QuizShowNames() []string {
	out := make([]string, 0, len(quizShows))
	for s := range quizShows {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func normalizeQuizShow(s string) string {
	n := strings.ToLower(strings.TrimSpace(s))
	if quizShows[n] {
		return n
	}
	return "think"
}

func validateQuizPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Quiz: true}); err != nil {
		return err
	}

	q := p.Quiz
	if q == nil {
		return fmt.Errorf("the plan has no question — this template is one question asked and answered")
	}
	if n := len(strings.Fields(q.Question)); n == 0 {
		return fmt.Errorf("the question is empty")
	} else if n > maxQuestionWords {
		return fmt.Errorf("the question is %d words; at most %d — a question that needs re-reading has already lost the pause it was going to buy",
			n, maxQuestionWords)
	}
	if n := len(q.Options); n < minQuizOptions || n > maxQuizOptions {
		return fmt.Errorf("there are %d options, want %d-%d", n, minQuizOptions, maxQuizOptions)
	}
	if q.Answer < 0 || q.Answer >= len(q.Options) {
		return fmt.Errorf("answer %d is not one of the %d options", q.Answer, len(q.Options))
	}
	seen := map[string]bool{}
	for i, opt := range q.Options {
		trimmed := strings.TrimSpace(opt)
		if trimmed == "" {
			return fmt.Errorf("option %d is empty", i)
		}
		if n := len(strings.Fields(trimmed)); n > maxOptionWords {
			return fmt.Errorf("option %d is %d words; at most %d — every option is read during the pause", i, n, maxOptionWords)
		}
		key := strings.ToLower(trimmed)
		if seen[key] {
			return fmt.Errorf("option %d repeats an earlier one (%q)", i, trimmed)
		}
		seen[key] = true
	}
	// Every option, not just the right one. A distractor nobody would pick
	// teaches nothing, and the only way to know the model built a tempting one
	// is to make it say what the temptation is.
	if len(q.Why) != len(q.Options) {
		return fmt.Errorf("there are %d options but %d explanations — every option needs one, including the wrong ones: saying what makes a wrong answer tempting is where the teaching is",
			len(q.Options), len(q.Why))
	}
	for i, w := range q.Why {
		if strings.TrimSpace(w) == "" {
			return fmt.Errorf("option %d has an empty explanation", i)
		}
	}

	// Beat shape.
	askAt, revealAt := -1, -1
	explained := map[int]bool{}
	for i, b := range p.Beats {
		if b.Quiz == nil {
			return fmt.Errorf("beat %q has no quiz direction — every beat in this template is doing something to the question", b.ID)
		}
		show := normalizeQuizShow(b.Quiz.Show)
		switch show {
		case "ask":
			if askAt >= 0 {
				return fmt.Errorf("beat %q asks the question again; it is asked once", b.ID)
			}
			askAt = i
		case "reveal":
			if revealAt >= 0 {
				return fmt.Errorf("beat %q reveals the answer again; it is revealed once", b.ID)
			}
			revealAt = i
		case "explain":
			if b.Quiz.Option < 0 || b.Quiz.Option >= len(q.Options) {
				return fmt.Errorf("beat %q explains option %d, which does not exist", b.ID, b.Quiz.Option)
			}
			if explained[b.Quiz.Option] {
				return fmt.Errorf("beat %q explains option %d twice", b.ID, b.Quiz.Option)
			}
			explained[b.Quiz.Option] = true
		}
	}
	if askAt < 0 {
		return fmt.Errorf("no beat asks the question — mark it with \"show\": \"ask\"")
	}
	if revealAt < 0 {
		return fmt.Errorf("no beat reveals the answer — mark it with \"show\": \"reveal\"")
	}
	if revealAt < askAt {
		return fmt.Errorf("the answer is revealed before the question is asked")
	}
	// The gap. This is the template's whole reason for existing: a clip that
	// asks and answers in consecutive beats has taught nothing the answer alone
	// would not have, and it is the first rule a model drops.
	if revealAt-askAt < 2 {
		return fmt.Errorf("the answer is revealed in the beat right after the question. Put a \"think\" beat between them where nothing new appears — that silence is the entire point of a quiz, and without it this is a statement with extra steps")
	}
	for i := askAt + 1; i < revealAt; i++ {
		if normalizeQuizShow(p.Beats[i].Quiz.Show) != "think" {
			return fmt.Errorf("beat %q sits between the question and the answer but is a %q beat; everything in that gap must be \"think\"",
				p.Beats[i].ID, normalizeQuizShow(p.Beats[i].Quiz.Show))
		}
	}
	// An explanation before the answer gives the answer away.
	for i, b := range p.Beats {
		if normalizeQuizShow(b.Quiz.Show) == "explain" && i < revealAt {
			return fmt.Errorf("beat %q explains an option before the answer is revealed", b.ID)
		}
	}
	return nil
}

// quizScenes lays the clip out as ONE scene: the question is on screen from the
// moment it is asked until the end, and the beats only change what is lit.
//
// The same shape the data and workspace templates use, for the same reason —
// re-mounting the question every beat would re-animate it, and a question that
// re-lands while the viewer is trying to answer it is a question they have to
// re-read.
func quizScenes(in SnippetSceneInput) ([]Scene, error) {
	q := in.Plan.Quiz
	if q == nil {
		return nil, fmt.Errorf("the plan has no question")
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Quiz == nil {
			return nil, fmt.Errorf("beat %q has no quiz direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    normalizeQuizShow(beat.Quiz.Show),
		}
		if normalizeQuizShow(beat.Quiz.Show) == "explain" {
			step["option"] = beat.Quiz.Option
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneQuiz,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":    in.Plan.Title,
			"question": q.Question,
			"options":  q.Options,
			"answer":   q.Answer,
			"why":      q.Why,
			"steps":    steps,
		},
	}}, nil
}
