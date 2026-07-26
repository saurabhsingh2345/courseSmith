package studio

// Human feedback endpoints: notes land in review-notes.yaml (the same file
// the export-review loop uses), regeneration is a note plus a forced run of
// the stages that own the artifact, and quiz edits are stored as OVERRIDES
// in quiz-overrides.yaml so regeneration never clobbers human work.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/enfec/coursesmith/internal/pipeline"
)

// FeedbackRequest is the POST /api/feedback payload.
type FeedbackRequest struct {
	Course string `json:"course"`
	Lesson string `json:"lesson"`
	// Section scopes the note to a script section ("" = lesson-wide).
	Section string `json:"section,omitempty"`
	Note    string `json:"note"`
	// TimestampMs marks a video position ("f" key in the review screen);
	// it is prefixed onto the note as [at M:SS].
	TimestampMs int `json:"timestamp_ms,omitempty"`
}

func (s *Server) handleFeedback(w http.ResponseWriter, r *http.Request) {
	var req FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if strings.TrimSpace(req.Note) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("note is empty"))
		return
	}
	course, lesson, err := s.resolveLesson(req.Course, req.Lesson)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err := addReviewNote(course.Dir, lesson.ID, req.Section, req.Note, req.TimestampMs); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "saved",
		"file":   pipeline.ReviewNotesFileName,
		"hint":   "the note is injected into the next script generation and marked resolved",
	})
}

// addReviewNote appends one unresolved note to the course's
// review-notes.yaml.
func addReviewNote(courseDir, lessonID, section, note string, timestampMs int) error {
	notes, err := pipeline.LoadReviewNotes(courseDir)
	if err != nil {
		return err
	}
	if timestampMs > 0 {
		total := timestampMs / 1000
		note = fmt.Sprintf("[at %d:%02d] %s", total/60, total%60, note)
	}
	if notes.Lessons == nil {
		notes.Lessons = map[string]*pipeline.LessonNotes{}
	}
	ln := notes.Lessons[lessonID]
	if ln == nil {
		ln = &pipeline.LessonNotes{}
		notes.Lessons[lessonID] = ln
	}
	entry := pipeline.ReviewNote{Note: note}
	if section == "" {
		ln.Notes = append(ln.Notes, entry)
	} else {
		if ln.Sections == nil {
			ln.Sections = map[string][]pipeline.ReviewNote{}
		}
		ln.Sections[section] = append(ln.Sections[section], entry)
	}
	return notes.Save()
}

// RegenerateRequest is the POST /api/regenerate payload.
type RegenerateRequest struct {
	Course string `json:"course"`
	Lesson string `json:"lesson"`
	// Artifact names what to regenerate: script, quiz, visuals, mistakes,
	// exercises, audio.
	Artifact string `json:"artifact"`
	// Section scopes a script note to one section.
	Section string `json:"section,omitempty"`
	Note    string `json:"note,omitempty"`
}

// regenerateStages maps an artifact to the forced stages that rebuild it.
// Downstream staleness cascades on the next full run.
var regenerateStages = map[string][]string{
	"script":    {"script", "verify", "review"},
	"quiz":      {"quiz"},
	"visuals":   {"visuals"},
	"mistakes":  {"mistakes"},
	"exercises": {"exercises"},
	"audio":     {"audio", "align"},
}

func (s *Server) handleRegenerate(w http.ResponseWriter, r *http.Request) {
	var req RegenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	stages, ok := regenerateStages[req.Artifact]
	if !ok {
		keys := make([]string, 0, len(regenerateStages))
		for k := range regenerateStages {
			keys = append(keys, k)
		}
		writeError(w, http.StatusBadRequest, fmt.Errorf("unknown artifact %q (want one of: %s)", req.Artifact, strings.Join(keys, ", ")))
		return
	}
	course, lesson, err := s.resolveLesson(req.Course, req.Lesson)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	// The note rides the same review-notes channel the script generator
	// already injects; for non-script artifacts it still lands in the file
	// as an audit trail.
	if strings.TrimSpace(req.Note) != "" {
		if err := addReviewNote(course.Dir, lesson.ID, req.Section, req.Note, 0); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	runID, err := s.runs.StartStages(course, lesson, stages, true)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"run_id": runID, "stages": stages})
}

// --- quiz overrides ---

// QuizWithOverrides is the GET /api/quiz response: the generated quiz, the
// stored overrides, and the merged result the site will publish.
type QuizWithOverrides struct {
	Generated json.RawMessage         `json:"generated,omitempty"`
	Overrides *pipeline.QuizOverrides `json:"overrides,omitempty"`
	Merged    *pipeline.Quiz          `json:"merged,omitempty"`
}

func (s *Server) handleQuizGet(w http.ResponseWriter, r *http.Request) {
	_, lesson, err := s.resolveLesson(r.PathValue("course"), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	out := QuizWithOverrides{}
	merged, generatedRaw, overrides, err := pipeline.LoadQuizWithOverrides(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out.Generated = generatedRaw
	out.Overrides = overrides
	out.Merged = merged
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleQuizOverridesPut(w http.ResponseWriter, r *http.Request) {
	_, lesson, err := s.resolveLesson(r.PathValue("course"), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var overrides pipeline.QuizOverrides
	if err := json.NewDecoder(r.Body).Decode(&overrides); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid overrides body: %w", err))
		return
	}
	if err := pipeline.SaveQuizOverrides(lesson, &overrides); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	merged, _, _, err := pipeline.LoadQuizWithOverrides(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, QuizWithOverrides{Overrides: &overrides, Merged: merged})
}
