package pipeline

import (
	"strings"
	"testing"
)

// The shared beat-shape gate enforces a ten-word floor per beat, so these
// fixtures carry real narration. A one-word placeholder made every quiz test
// fail on that rule instead of on the rule it was written for.
const (
	askNarration    = "Here is a quick one to check you have actually got this."
	thinkNarration  = "Take a second with it before I say anything else at all."
	revealNarration = "It is two, because the list only holds two things directly."
	trapNarration   = "Five is tempting because that is how many numbers are in there."
)

func quizPlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "quiz",
		Title:    "What does len() really count?",
		Quiz: &QuizSpec{
			Question: "What does len() return for a list of lists?",
			Options:  []string{"2", "5", "7", "TypeError"},
			Answer:   0,
			Why: []string{
				"len() counts the top level, and there are two lists inside it.",
				"This counts every number instead.",
				"This adds the two inner counts together.",
				"Nested lists are valid, so nothing raises.",
			},
		},
		Beats: []SnippetBeat{
			{ID: "ask", Heading: "A quick check", Narration: askNarration, Quiz: &QuizBeat{Show: "ask"}},
			{ID: "think", Heading: "Your turn", Narration: thinkNarration, Quiz: &QuizBeat{Show: "think"}},
			{ID: "reveal", Heading: "The answer", Narration: revealNarration, Quiz: &QuizBeat{Show: "reveal"}},
			{ID: "trap", Heading: "Why five tempts", Narration: trapNarration, Quiz: &QuizBeat{Show: "explain", Option: 1}},
		},
	}
}

func TestQuizPlanAccepted(t *testing.T) {
	if err := validateQuizPlan(quizPlan()); err != nil {
		t.Fatalf("a well-formed quiz plan was rejected: %v", err)
	}
}

// The gap between the question and the answer is the whole template. A clip
// that asks and answers back to back has taught nothing the answer alone would
// not have, and it is the first rule a model drops.
func TestQuizRequiresAThinkingGap(t *testing.T) {
	p := quizPlan()
	p.Beats = []SnippetBeat{
		{ID: "ask", Heading: "h", Narration: askNarration, Quiz: &QuizBeat{Show: "ask"}},
		{ID: "reveal", Heading: "h", Narration: revealNarration, Quiz: &QuizBeat{Show: "reveal"}},
	}
	err := validateQuizPlan(p)
	if err == nil {
		t.Fatal("a reveal in the beat right after the ask was accepted")
	}
	if !strings.Contains(err.Error(), "think") {
		t.Errorf("the error should tell the model to add a think beat; got: %v", err)
	}
}

func TestQuizGapMustBeThinkBeats(t *testing.T) {
	p := quizPlan()
	// An explain beat smuggled into the gap gives the answer away.
	p.Beats = []SnippetBeat{
		{ID: "ask", Heading: "h", Narration: askNarration, Quiz: &QuizBeat{Show: "ask"}},
		{ID: "sneaky", Heading: "h", Narration: trapNarration, Quiz: &QuizBeat{Show: "explain", Option: 1}},
		{ID: "reveal", Heading: "h", Narration: revealNarration, Quiz: &QuizBeat{Show: "reveal"}},
	}
	if err := validateQuizPlan(p); err == nil {
		t.Fatal("an explain beat before the reveal was accepted")
	}
}

func TestQuizRejectsMissingAskOrReveal(t *testing.T) {
	for _, drop := range []string{"ask", "reveal"} {
		p := quizPlan()
		kept := p.Beats[:0]
		for _, b := range p.Beats {
			if b.Quiz.Show != drop {
				kept = append(kept, b)
			}
		}
		p.Beats = kept
		if err := validateQuizPlan(p); err == nil {
			t.Errorf("a plan with no %q beat was accepted", drop)
		}
	}
}

func TestQuizRejectsBadAnswerIndex(t *testing.T) {
	for _, idx := range []int{-1, 4, 99} {
		p := quizPlan()
		p.Quiz.Answer = idx
		if err := validateQuizPlan(p); err == nil {
			t.Errorf("answer index %d was accepted for 4 options", idx)
		}
	}
}

// Every option needs an explanation, not just the right one. A distractor
// nobody would pick teaches nothing, and making the model say why a wrong
// answer is tempting is the only way to know it built a real one.
func TestQuizRequiresAnExplanationPerOption(t *testing.T) {
	p := quizPlan()
	p.Quiz.Why = p.Quiz.Why[:2]
	err := validateQuizPlan(p)
	if err == nil {
		t.Fatal("a plan with fewer explanations than options was accepted")
	}
	if !strings.Contains(err.Error(), "tempting") {
		t.Errorf("the error should say why the wrong answers need explaining; got: %v", err)
	}

	p = quizPlan()
	p.Quiz.Why[2] = "   "
	if err := validateQuizPlan(p); err == nil {
		t.Fatal("a blank explanation was accepted")
	}
}

func TestQuizRejectsDuplicateAndOversizedOptions(t *testing.T) {
	p := quizPlan()
	p.Quiz.Options[2] = "2"
	if err := validateQuizPlan(p); err == nil {
		t.Error("a repeated option was accepted")
	}

	p = quizPlan()
	p.Quiz.Options[1] = strings.Repeat("word ", maxOptionWords+3)
	if err := validateQuizPlan(p); err == nil {
		t.Error("an option far over the word cap was accepted")
	}

	p = quizPlan()
	p.Quiz.Question = strings.Repeat("word ", maxQuestionWords+5)
	if err := validateQuizPlan(p); err == nil {
		t.Error("a question far over the word cap was accepted")
	}
}

func TestQuizRejectsOptionCountOutOfRange(t *testing.T) {
	p := quizPlan()
	p.Quiz.Options = []string{"a", "b"}
	p.Quiz.Why = []string{"x", "y"}
	if err := validateQuizPlan(p); err == nil {
		t.Error("two options were accepted — that is a coin flip, not retrieval")
	}
}

func TestQuizRejectsRepeatedExplanationOfOneOption(t *testing.T) {
	p := quizPlan()
	p.Beats = append(p.Beats, SnippetBeat{
		ID: "again", Heading: "h", Narration: trapNarration, Quiz: &QuizBeat{Show: "explain", Option: 1},
	})
	if err := validateQuizPlan(p); err == nil {
		t.Error("the same option was explained twice and accepted")
	}
}

func TestQuizScenesIsOneSceneWithSteps(t *testing.T) {
	plan := quizPlan()
	scenes, err := quizScenes(sceneInput(t, plan, 6000))
	if err != nil {
		t.Fatal(err)
	}
	// One scene for the whole clip: the question must not re-mount and
	// re-animate while somebody is trying to answer it.
	if len(scenes) != 1 {
		t.Fatalf("got %d scenes, want exactly one spanning the clip", len(scenes))
	}
	if scenes[0].Type != SceneQuiz {
		t.Errorf("scene type = %s, want %s", scenes[0].Type, SceneQuiz)
	}
	steps, ok := scenes[0].Props["steps"].([]map[string]any)
	if !ok || len(steps) != len(plan.Beats) {
		t.Fatalf("want one step per beat, got %#v", scenes[0].Props["steps"])
	}
	if got := steps[3]["option"]; got != 1 {
		t.Errorf("the explain step lost its option: %v", got)
	}
	if _, present := steps[0]["option"]; present {
		t.Error("a non-explain step carries an option index it does not mean")
	}
	// The options and their explanations must arrive index-aligned, or the
	// scene shows the wrong reason under the wrong answer.
	opts, _ := scenes[0].Props["options"].([]string)
	whys, _ := scenes[0].Props["why"].([]string)
	if len(opts) != len(whys) {
		t.Errorf("%d options but %d explanations reached the scene", len(opts), len(whys))
	}
}

func TestNormalizeQuizShow(t *testing.T) {
	for _, s := range QuizShowNames() {
		if got := normalizeQuizShow(s); got != s {
			t.Errorf("normalizeQuizShow(%q) = %q, want it preserved", s, got)
		}
	}
	// The fallback is `think` rather than `ask` or `reveal`: an invented name
	// becoming a second ask or a second reveal would break the shape the rest
	// of the validation depends on.
	for _, bad := range []string{"", "answer", "show", "  "} {
		if got := normalizeQuizShow(bad); got != "think" {
			t.Errorf("normalizeQuizShow(%q) = %q, want the think fallback", bad, got)
		}
	}
}
