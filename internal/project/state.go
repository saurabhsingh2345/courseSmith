package project

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"
)

// StateFileName is the per-lesson pipeline state file, stored in the
// lesson's generated/ directory.
const StateFileName = "state.json"

// Stage names, in pipeline order. The pipeline package executes these;
// the status command reports on them.
const (
	StagePlan         = "plan" // snippet-only: prompt + template → snippet-plan.json
	StageScript       = "script"
	StageVerify       = "verify"        // execute code blocks, capture real output
	StageTrace        = "trace"         // step-by-step execution trace (code-viz)
	StageReview       = "review"        // LLM quality gate on the script
	StageStoryboard   = "storyboard"    // per-section visual beats (no dead screens)
	StageVisuals      = "visuals"       // SVG diagrams
	StageQuiz         = "quiz"          // multiple-choice quiz
	StageQuizStrategy = "quiz-strategy" // interleave/sequence the quiz
	StageMistakes     = "mistakes"      // common mistakes with real tracebacks
	StageExercises    = "exercises"     // practice exercises, proven solvable
	StageDemos        = "demos"         // VHS terminal demo recordings
	StageAudio        = "audio"         // Kokoro TTS voiceover
	StageAlign        = "align"         // word-level timestamps + WER QA
	StageCaptions     = "captions"      // WebVTT from the alignment
	StageChapters     = "chapters"      // chapters.json/.txt + transcript.md
	StageScenegraph   = "scenegraph"    // lesson-video.json for the renderer
	StageRender       = "render"        // Remotion (or fallback) → final.mp4
	StageHugo         = "hugo"          // website page bundle
)

// StageOrder lists all pipeline stages in execution order.
var StageOrder = []string{
	StageScript, StageVerify, StageTrace, StageReview, StageStoryboard,
	StageVisuals, StageQuiz,
	StageQuizStrategy, StageMistakes, StageExercises,
	StageDemos, StageAudio, StageAlign, StageCaptions, StageChapters,
	StageScenegraph, StageRender, StageHugo,
}

// videoStageSkip marks the stages that produce companion material rather
// than the video itself: quiz generation and sequencing, the mistakes and
// exercises documents, and the Hugo site page. A video-only course skips them.
var videoStageSkip = map[string]bool{
	StageQuiz:         true,
	StageQuizStrategy: true,
	StageMistakes:     true,
	StageExercises:    true,
	StageHugo:         true,
}

// VideoStageOrder is StageOrder minus the companion-material stages —
// everything needed to go from lesson.md to final.mp4, nothing else.
var VideoStageOrder = func() []string {
	out := make([]string, 0, len(StageOrder))
	for _, s := range StageOrder {
		if !videoStageSkip[s] {
			out = append(out, s)
		}
	}
	return out
}()

// SnippetStageOrder is the pipeline for a snippet: a short, standalone video
// built from one prompt plus one visual template.
//
// A snippet skips authoring entirely — no lesson.md to write, no script
// generation, no review rounds, no storyboard. The plan stage does all the
// thinking in one LLM call (narration *and* the template's visual spec), and
// everything after it is the ordinary video path, reused unchanged. Verify
// still runs, so a snippet that shows code shows output the interpreter
// actually produced.
var SnippetStageOrder = []string{
	StagePlan, StageVerify, StageAudio, StageAlign, StageCaptions,
	StageChapters, StageScenegraph, StageRender,
}

// StagesFor returns the stage list a course actually runs.
func StagesFor(videoOnly bool) []string {
	if videoOnly {
		return VideoStageOrder
	}
	return StageOrder
}

// AllStages lists every stage name that can appear in any run mode; used to
// validate an explicit --stage request.
func AllStages() []string {
	out := append([]string{}, StageOrder...)
	for _, s := range SnippetStageOrder {
		if !slices.Contains(out, s) {
			out = append(out, s)
		}
	}
	return out
}

// StageStatus is the resolved status of one stage for reporting.
type StageStatus string

const (
	StatusDone    StageStatus = "done"    // completed, input hashes match
	StatusStale   StageStatus = "stale"   // completed, but inputs changed since
	StatusPending StageStatus = "pending" // never completed
)

// State records which stages have completed for a lesson and the input
// hashes they saw, so unchanged stages can be skipped on re-runs.
type State struct {
	// Stages maps stage name → completion record.
	Stages map[string]StageRecord `json:"stages"`
}

// StageRecord is one completed stage run.
type StageRecord struct {
	// InputHashes maps a stable input key (usually a filename relative to
	// the lesson dir, or a config fingerprint key) to its sha256 hash.
	InputHashes map[string]string `json:"input_hashes"`
	CompletedAt time.Time         `json:"completed_at"`
}

// StatePath returns where a lesson's state.json lives.
func (l *Lesson) StatePath() string {
	return filepath.Join(l.GeneratedDir(), StateFileName)
}

// LoadState reads the lesson's state.json. A missing file is not an error:
// it returns an empty state, meaning no stage has run.
func (l *Lesson) LoadState() (*State, error) {
	data, err := os.ReadFile(l.StatePath())
	if err != nil {
		if os.IsNotExist(err) {
			return &State{Stages: map[string]StageRecord{}}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", l.StatePath(), err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing %s (delete it to reset pipeline state): %w", l.StatePath(), err)
	}
	if s.Stages == nil {
		s.Stages = map[string]StageRecord{}
	}
	return &s, nil
}

// SaveState writes state.json atomically (write temp file, then rename).
func (l *Lesson) SaveState(s *State) error {
	path := l.StatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding state: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), StateFileName+".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp state file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}

// StageStatus resolves the status of one stage given the hashes its inputs
// have right now.
func (s *State) StageStatus(stage string, currentInputs map[string]string) StageStatus {
	rec, ok := s.Stages[stage]
	if !ok {
		return StatusPending
	}
	if !hashesEqual(rec.InputHashes, currentInputs) {
		return StatusStale
	}
	return StatusDone
}

// MarkDone records a completed stage run with the inputs it consumed.
func (s *State) MarkDone(stage string, inputs map[string]string, at time.Time) {
	if s.Stages == nil {
		s.Stages = map[string]StageRecord{}
	}
	cp := make(map[string]string, len(inputs))
	maps.Copy(cp, inputs)
	s.Stages[stage] = StageRecord{InputHashes: cp, CompletedAt: at.UTC()}
}

// Invalidate removes a stage's completion record (used by --force).
func (s *State) Invalidate(stage string) {
	delete(s.Stages, stage)
}

func hashesEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// HashFile returns "sha256:<hex>" of a file's contents.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("hashing %s: %w", path, err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hashing %s: %w", path, err)
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// HashBytes returns "sha256:<hex>" of a byte slice — used to fingerprint
// non-file inputs like resolved config or prompt templates.
func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// HashStrings fingerprints an ordered set of key=value pairs deterministically.
// Useful for hashing a resolved config's relevant fields as one input.
func HashStrings(kv map[string]string) string {
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(h, "%s=%s\n", k, kv[k])
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}
