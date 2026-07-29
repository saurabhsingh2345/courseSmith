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
	// Category is the gallery group this template belongs to, and CategoryTitle
	// the heading to show over it. The title is sent alongside the id so the UI
	// does not have to keep its own copy of the vocabulary and drift from Go.
	Category      string `json:"category"`
	CategoryTitle string `json:"category_title"`
	// Since is the catalog release the template arrived in ("" for the original
	// set). The gallery shows it as a chip so a new batch is findable.
	Since string `json:"since,omitempty"`
	// ShowsCode marks templates whose clips execute code for real, which is
	// the difference the gallery needs to communicate.
	ShowsCode bool `json:"shows_code"`
	// MinTargetSec and DefaultTargetSec are the runtimes this template can
	// actually satisfy. The picker needs them: `story` is built from eight or
	// more beats, so a twenty-second story is not a short clip, it is a plan
	// whose own rules contradict each other. Offering the option at all cost a
	// user three correction rounds and a day's token budget to find out.
	MinTargetSec     int `json:"min_target_sec"`
	DefaultTargetSec int `json:"default_target_sec"`
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
	// Captions burns the caption track into the video: "on" | "off".
	// "" takes the snippets course setting.
	Captions string `json:"captions,omitempty"`
	// Mode is the video's polarity: "light" | "dark". "" is dark.
	Mode string `json:"mode,omitempty"`
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
	// Emitted in CATEGORY order, not name order.
	//
	// The gallery groups by first-seen category so it never keeps its own copy
	// of the vocabulary, which means the order templates arrive in *is* the
	// order the groups render in. Sending the flat name-sorted list put
	// "Ideas & mental models" first purely because `analogy` sorts before
	// `costing` — the sequence Go declares in snippet_categories.go would have
	// been silently ignored by the one client that shows headings.
	out := make([]SnippetTemplateInfo, 0, len(pipeline.SnippetTemplates))
	for _, g := range pipeline.SnippetTemplatesByCategory() {
		for _, t := range g.Templates {
			out = append(out, SnippetTemplateInfo{
				Name:             t.Name,
				Title:            t.Title,
				Description:      t.Description,
				Example:          t.Example,
				Category:         t.Category,
				CategoryTitle:    catTitle(t.Category),
				Since:            t.Since,
				ShowsCode:        t.NeedsCode,
				MinTargetSec:     t.MinTargetSec,
				DefaultTargetSec: t.DefaultTargetSec,
			})
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// catTitle is the heading for a category id, so the API ships the display
// string with the value rather than making every client keep its own map.
func catTitle(name string) string {
	if c, ok := pipeline.LookupSnippetCategory(name); ok {
		return c.Title
	}
	return name
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
	// One assignment rather than three: building the override per-field
	// overwrote whichever of them was set first.
	spec.Config = config.Config{Style: config.Style{
		Voice:    req.Voice,
		Captions: req.Captions,
		Mode:     req.Mode,
	}}
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
