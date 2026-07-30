package pipeline

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// A prompt may only name vocabulary that exists.
//
// The pose and expression lists are injected as {{.Poses}} and {{.Expressions}}
// from the live vocabularies, which is right and was not enough: the surrounding
// PROSE named poses too, and it kept naming them after they were deleted from the
// renderer. story_shots.tmpl advised that "the payoff wants `celebrate`" and its
// worked JSON example used `defeated`; neither has existed in POSES since the
// character stopped being a stick figure. Both silently fell back to `idle`,
// which is why every character in a finished reel stood in the same neutral
// stance through its punchline — and why the same grinning man appeared six shots
// running.
//
// The existing drift test compares the Go vocabulary against the renderer's. This
// one compares the PROMPTS against the Go vocabulary, which is the direction that
// was unguarded. A model follows an example far more closely than a list.
func TestPromptsOnlyNameVocabularyThatExists(t *testing.T) {
	// Backtick-quoted words and JSON string values are where a vocabulary term
	// gets named. Prose mentions of ordinary English are not matched, because both
	// forms are how this codebase writes a term of art.
	quoted := regexp.MustCompile("`([a-z][a-z-]{2,})`")
	jsonPose := regexp.MustCompile(`"(?:pose|expression|staging|camera)":\s*"([a-z-]+)"`)

	// Every term any of these fields may legally hold.
	legal := map[string]bool{}
	for _, group := range [][]string{
		CastPoseNames(), CastExpressionNames(),
		StoryStagingNames(), StoryCameraNames(), ArtFigureNames(),
	} {
		for _, name := range group {
			legal[name] = true
		}
	}

	// Words that look like vocabulary but are not — template names, field names,
	// and the handful of ordinary words this codebase backticks.
	exempt := map[string]bool{}
	for _, name := range SnippetTemplateNames() {
		exempt[name] = true
	}
	for _, w := range []string{
		"pose", "expression", "staging", "camera", "prop", "prop_b", "caption",
		"beat_id", "shots", "beats", "narration", "heading", "code", "title",
		"subtitle", "print", "true", "false", "null", "and", "not", "the",
		"from", "breaks", "covers", "material", "why", "show", "kind", "svg",
		"mermaid", "python", "javascript", "bash", "json", "yaml",
	} {
		exempt[w] = true
	}

	// Only the prompts that direct a character. Checking every prompt would flag
	// each template's own vocabularies, which are legitimately theirs.
	for _, file := range []string{"story_shots.tmpl", "snippet_cast.tmpl"} {
		t.Run(file, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(repoPromptsDir, file))
			if err != nil {
				t.Fatalf("reading %s: %v", file, err)
			}
			text := string(raw)

			// A JSON example naming a dead term is the worst case: models copy
			// examples. Checked first and reported as such.
			for _, m := range jsonPose.FindAllStringSubmatch(text, -1) {
				if !legal[m[1]] {
					t.Errorf("the worked example in %s uses %q, which is not in any vocabulary — a model will copy it, and it will silently fall back",
						file, m[1])
				}
			}

			// The advice around the example.
			for _, m := range quoted.FindAllStringSubmatch(text, -1) {
				term := m[1]
				if legal[term] || exempt[term] {
					continue
				}
				// Only complain about terms that read like a pose or a mood, which
				// is what actually drifted. A backticked word this test has never
				// heard of is more likely a field name than a dead pose.
				if looksLikeDirection(term) {
					t.Errorf("%s names `%s`, which is not in the pose, expression, staging, camera or prop vocabularies (poses: %v; expressions: %v)",
						file, term, CastPoseNames(), CastExpressionNames())
				}
			}
		})
	}
}

// looksLikeDirection is the heuristic that keeps this test from flagging field
// names. These are the terms that were deleted from the renderer and left behind
// in the prompts, plus the ones most likely to be invented next.
func looksLikeDirection(term string) bool {
	for _, dead := range []string{
		"celebrate", "defeated", "wave", "walk", "think", "explain", "coffee",
		"phone", "typing", "excited", "angry", "confused", "proud", "worried",
		"smile", "frown", "nod", "shake", "jump", "run", "sit", "stand",
	} {
		if term == dead {
			return true
		}
	}
	return false
}

// And the vocabularies must not be empty, or every assertion above is vacuous.
func TestCharacterVocabulariesAreNotEmpty(t *testing.T) {
	if len(CastPoseNames()) == 0 {
		t.Error("there are no poses, so the drift test proves nothing")
	}
	if len(CastExpressionNames()) == 0 {
		t.Error("there are no expressions")
	}
	// The prompts inject these lists; if the injection point vanished, the prompt
	// would hard-code a list instead and drift immediately.
	for _, file := range []string{"story_shots.tmpl", "snippet_cast.tmpl"} {
		raw, err := os.ReadFile(filepath.Join(repoPromptsDir, file))
		if err != nil {
			t.Fatalf("reading %s: %v", file, err)
		}
		if !strings.Contains(string(raw), "{{.Poses}}") {
			t.Errorf("%s no longer injects the live pose list, so its vocabulary can drift freely", file)
		}
	}
}
