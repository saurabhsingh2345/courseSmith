package pipeline

// Quiz overrides: human edits to generated quizzes live in the lesson dir
// as quiz-overrides.yaml — SEPARATE from quiz.json — so a regeneration can
// never clobber human work. The published quiz is the generated one with
// overrides merged on top at build time (hugo stage) and in the studio API.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/enfec/coursesmith/internal/project"
)

// QuizOverridesFileName lives in the LESSON dir (human-owned, next to
// lesson.md), not in generated/.
const QuizOverridesFileName = "quiz-overrides.yaml"

// QuestionOverride edits one generated question, matched by id. Only
// non-nil / non-empty fields override; Drop removes the question.
type QuestionOverride struct {
	ID          string   `yaml:"id" json:"id"`
	Drop        bool     `yaml:"drop,omitempty" json:"drop,omitempty"`
	Prompt      *string  `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	Options     []string `yaml:"options,omitempty" json:"options,omitempty"`
	AnswerIndex *int     `yaml:"answer_index,omitempty" json:"answer_index,omitempty"`
	Explanation *string  `yaml:"explanation,omitempty" json:"explanation,omitempty"`
}

// QuizOverrides is the parsed quiz-overrides.yaml.
type QuizOverrides struct {
	Questions []QuestionOverride `yaml:"questions" json:"questions"`
}

func quizOverridesPath(l *project.Lesson) string {
	return filepath.Join(l.Dir, QuizOverridesFileName)
}

// LoadQuizOverrides reads the lesson's overrides; a missing file returns nil.
func LoadQuizOverrides(l *project.Lesson) (*QuizOverrides, error) {
	data, err := os.ReadFile(quizOverridesPath(l))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", QuizOverridesFileName, err)
	}
	var out QuizOverrides
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", QuizOverridesFileName, err)
	}
	return &out, nil
}

// SaveQuizOverrides writes the overrides file (empty overrides delete it).
func SaveQuizOverrides(l *project.Lesson, o *QuizOverrides) error {
	path := quizOverridesPath(l)
	if o == nil || len(o.Questions) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	data, err := yaml.Marshal(o)
	if err != nil {
		return fmt.Errorf("encoding quiz overrides: %w", err)
	}
	header := []byte("# Human edits to the generated quiz. Merged over quiz.json at build\n# time; regeneration never touches this file.\n")
	return writeFileAtomic(path, append(header, data...))
}

// MergeQuizOverrides applies overrides to a copy of the generated quiz.
// Unknown ids are ignored (the generated quiz may have changed since the
// edit was made).
func MergeQuizOverrides(quiz *Quiz, o *QuizOverrides) *Quiz {
	if quiz == nil {
		return nil
	}
	merged := *quiz
	merged.Questions = append([]Question(nil), quiz.Questions...)
	if o == nil {
		return &merged
	}
	byID := make(map[string]QuestionOverride, len(o.Questions))
	for _, ov := range o.Questions {
		byID[ov.ID] = ov
	}
	out := merged.Questions[:0]
	for _, q := range merged.Questions {
		ov, ok := byID[q.ID]
		if !ok {
			out = append(out, q)
			continue
		}
		if ov.Drop {
			continue
		}
		if ov.Prompt != nil {
			q.Prompt = *ov.Prompt
		}
		if len(ov.Options) > 0 {
			q.Options = ov.Options
		}
		if ov.AnswerIndex != nil {
			q.AnswerIndex = *ov.AnswerIndex
		}
		if ov.Explanation != nil {
			q.Explanation = *ov.Explanation
		}
		out = append(out, q)
	}
	merged.Questions = out
	return &merged
}

// LoadQuizWithOverrides returns the merged quiz, the raw generated JSON,
// and the overrides. A lesson without a generated quiz returns nils.
func LoadQuizWithOverrides(l *project.Lesson) (*Quiz, json.RawMessage, *QuizOverrides, error) {
	raw, err := os.ReadFile(filepath.Join(l.GeneratedDir(), QuizFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, fmt.Errorf("reading %s: %w", QuizFileName, err)
	}
	var quiz Quiz
	if err := json.Unmarshal(raw, &quiz); err != nil {
		return nil, nil, nil, fmt.Errorf("parsing %s: %w", QuizFileName, err)
	}
	overrides, err := LoadQuizOverrides(l)
	if err != nil {
		return nil, nil, nil, err
	}
	return MergeQuizOverrides(&quiz, overrides), raw, overrides, nil
}
