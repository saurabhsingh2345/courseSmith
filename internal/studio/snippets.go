package studio

// Snippet endpoints: the short-form surface.
//
// A snippet is a prompt plus a visual template, so the API is deliberately
// thin — there is no document to edit and no course to pick. Create returns
// immediately with a run id and the SSE stream carries the rest, because
// planning, synthesizing and rendering a clip takes longer than a request
// should live.
//
// Everything about *reading* a finished snippet already works through the
// lesson routes: a snippet resolves as course "snippets", so its artifacts,
// stage statuses, and video all come back from GET /api/lessons/snippets/{id}.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/pipeline"
	"github.com/enfec/coursesmith/internal/project"
)

// SnippetTemplateInfo is one card in the template gallery.
type SnippetTemplateInfo struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Example     string `json:"example"`
	// ShowsCode marks templates whose clips execute code for real, which is
	// the difference the gallery needs to communicate.
	ShowsCode bool `json:"shows_code"`
}

// SnippetSummary is one row of the snippet list.
type SnippetSummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Prompt   string `json:"prompt"`
	Template string `json:"template"`
	// Ready is true once final.mp4 exists.
	Ready bool `json:"ready"`
	// VideoURL is the artifact URL of the finished clip ("" until ready).
	VideoURL  string `json:"video_url,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// SnippetDetail is a snippet's request plus the plan the model produced.
type SnippetDetail struct {
	SnippetSummary
	TargetSec int             `json:"target_sec"`
	Plan      json.RawMessage `json:"plan,omitempty"`
}

// CreateSnippetRequest is the whole input: a prompt and a template.
type CreateSnippetRequest struct {
	Prompt   string `json:"prompt"`
	Template string `json:"template"`
	Title    string `json:"title,omitempty"`
	// TargetSec is the runtime to aim for (0 = the default).
	TargetSec int `json:"target_sec,omitempty"`
	// CodeLanguage applies to code-bearing templates ("" = python).
	CodeLanguage string `json:"code_language,omitempty"`
	Voice        string `json:"voice,omitempty"`
	// PlanOnly stops after planning, for reviewing the design before paying
	// for TTS and a render.
	PlanOnly bool `json:"plan_only,omitempty"`
}

// CreateSnippetResponse hands back the new snippet and the run watching it.
type CreateSnippetResponse struct {
	SnippetSummary
	RunID string `json:"run_id,omitempty"`
}

func (s *Server) handleSnippetTemplates(w http.ResponseWriter, _ *http.Request) {
	out := make([]SnippetTemplateInfo, 0, len(pipeline.SnippetTemplates))
	for _, t := range pipeline.SnippetTemplateList() {
		out = append(out, SnippetTemplateInfo{
			Name:        t.Name,
			Title:       t.Title,
			Description: t.Description,
			Example:     t.Example,
			ShowsCode:   t.NeedsCode,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// snippetSummary describes one snippet on disk.
func snippetSummary(l *project.Lesson) (SnippetSummary, error) {
	spec, err := pipeline.LoadSnippetSpec(l.Dir)
	if err != nil {
		return SnippetSummary{}, err
	}
	out := SnippetSummary{
		ID:       l.ID,
		Title:    l.FrontMatter.Title,
		Prompt:   spec.Prompt,
		Template: spec.Template,
	}
	if !spec.CreatedAt.IsZero() {
		out.CreatedAt = spec.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if _, err := os.Stat(filepath.Join(l.GeneratedDir(), pipeline.FinalVideoName)); err == nil {
		out.Ready = true
		out.VideoURL = "/artifacts/" + pipeline.SnippetsCourseSlug + "/" + l.ID + "/" + pipeline.FinalVideoName
	}
	return out, nil
}

func (s *Server) handleSnippetsList(w http.ResponseWriter, _ *http.Request) {
	snippets, err := pipeline.ListSnippets(s.projectRoot())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]SnippetSummary, 0, len(snippets))
	for _, l := range snippets {
		summary, err := snippetSummary(l)
		if err != nil {
			continue // a half-written snippet should not break the list
		}
		out = append(out, summary)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleSnippetDetail(w http.ResponseWriter, r *http.Request) {
	_, lesson, err := pipeline.FindSnippet(s.projectRoot(), filepath.Base(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	summary, err := snippetSummary(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	spec, err := pipeline.LoadSnippetSpec(lesson.Dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	detail := SnippetDetail{SnippetSummary: summary, TargetSec: spec.ResolvedTargetSec()}
	// The plan is passed through raw: it is the pipeline's own artifact and the
	// UI shows it as data.
	if raw, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), pipeline.SnippetPlanFileName)); err == nil && json.Valid(raw) {
		detail.Plan = raw
	}
	writeJSON(w, http.StatusOK, detail)
}

// handleSnippetCreate writes a new snippet and starts its pipeline. It returns
// as soon as the run is queued; progress arrives on /api/events.
func (s *Server) handleSnippetCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateSnippetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	spec := pipeline.SnippetSpec{
		Prompt:       strings.TrimSpace(req.Prompt),
		Template:     req.Template,
		Title:        strings.TrimSpace(req.Title),
		TargetSec:    req.TargetSec,
		CodeLanguage: req.CodeLanguage,
	}
	if req.Voice != "" {
		spec.Config = config.Config{Style: config.Style{Voice: req.Voice}}
	}
	if err := spec.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	course, lesson, err := pipeline.CreateSnippet(s.projectRoot(), spec)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	summary, err := snippetSummary(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	stages, err := pipeline.SnippetStages(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if req.PlanOnly {
		stages = []string{project.StagePlan}
	}
	runID, err := s.runs.StartStages(course, lesson, stages, false)
	if err != nil {
		// The snippet exists and is re-runnable; a busy pipeline is not a
		// reason to throw the request away.
		writeJSON(w, http.StatusAccepted, CreateSnippetResponse{SnippetSummary: summary})
		return
	}
	writeJSON(w, http.StatusCreated, CreateSnippetResponse{SnippetSummary: summary, RunID: runID})
}

func (s *Server) handleSnippetDelete(w http.ResponseWriter, r *http.Request) {
	_, lesson, err := pipeline.FindSnippet(s.projectRoot(), filepath.Base(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err := os.RemoveAll(lesson.Dir); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
