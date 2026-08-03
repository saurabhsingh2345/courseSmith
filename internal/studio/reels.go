package studio

// The reels API: create a multi-template video, watch it run, and edit it
// afterwards.
//
// The endpoints mirror the snippets ones exactly where they can, because a reel
// is the same object with more than one template in it — same synthetic course,
// same stage machinery, same SSE run stream. The one shape that has no snippet
// equivalent is PATCH on a segment, which is the whole point of the page: a
// reel is watched and then adjusted, and an adjustment has to be addressable to
// one segment without disturbing the others.

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

// ReelSegmentInfo is one segment as the page shows it.
type ReelSegmentInfo struct {
	ID       string `json:"id"`
	Template string `json:"template"`
	Prompt   string `json:"prompt"`
	// Material is the concrete facts this segment will be filled with. Surfaced
	// because it is the field most worth correcting by hand: it is what the
	// segment's writer plans from, so a wrong figure here is a wrong figure in
	// the finished video, and fixing it is one edit rather than a re-cast.
	Material string `json:"material,omitempty"`
	// TargetSec is 0 when the segment takes its template's default.
	TargetSec int `json:"target_sec,omitempty"`
	// Skip means the segment stays in the file but leaves the cut.
	Skip bool `json:"skip,omitempty"`
	// Title and Category come from the template catalog so the editor can label
	// a segment without holding its own copy of the catalog.
	TemplateTitle    string `json:"template_title"`
	TemplateCategory string `json:"template_category"`
}

// ReelSummary describes one reel in the list.
type ReelSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Brief string `json:"brief,omitempty"`
	// Segments is how many are in the cut; Skipped how many are parked.
	Segments int `json:"segments"`
	Skipped  int `json:"skipped"`
	// Ready is true once final.mp4 exists.
	Ready bool `json:"ready"`
	// VideoURL is the artifact URL of the finished video ("" until ready).
	VideoURL  string `json:"video_url,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// ReelDetail is a reel's request plus the plan, for the editor.
type ReelDetail struct {
	ReelSummary
	SegmentList []ReelSegmentInfo `json:"segment_list"`
	// Plan is reel-plan.json once the plan stage has run, so the editor can
	// show the narration a segment actually produced rather than only the
	// prompt that asked for it.
	Plan json.RawMessage `json:"plan,omitempty"`
}

// CreateReelRequest is the whole input: a brief, and the ordered segments.
type CreateReelRequest struct {
	Title    string              `json:"title,omitempty"`
	Brief    string              `json:"brief,omitempty"`
	Segments []CreateReelSegment `json:"segments"`
	Voice    string              `json:"voice,omitempty"`
	Captions string              `json:"captions,omitempty"`
	Mode     string              `json:"mode,omitempty"`
	// Skin is the house style: "" | "default" | "broadcast" | "minimal".
	Skin string `json:"skin,omitempty"`
	// PlanOnly stops after planning, for reviewing the design before paying
	// for TTS and a render. On a reel that matters more than on a snippet:
	// planning is one call per segment, and rendering is minutes.
	PlanOnly bool `json:"plan_only,omitempty"`
}

// CreateReelSegment is one requested segment.
type CreateReelSegment struct {
	Template string `json:"template"`
	Prompt   string `json:"prompt"`
	// Material is optional on a hand-authored reel — a segment without it plans
	// from the prompt alone, exactly as it did before the field existed. Supply
	// it and the writer stops guessing.
	Material  string `json:"material,omitempty"`
	TargetSec int    `json:"target_sec,omitempty"`
}

// CreateReelResponse is the created reel plus the run watching it.
type CreateReelResponse struct {
	ReelSummary
	RunID string `json:"run_id,omitempty"`
}

// PatchReelSegmentRequest edits one segment. Every field is a pointer so the
// editor can change exactly one thing: a plain string could not distinguish
// "set the prompt to empty" from "leave the prompt alone", and the second is
// what almost every request means.
type PatchReelSegmentRequest struct {
	Template  *string `json:"template,omitempty"`
	Prompt    *string `json:"prompt,omitempty"`
	Material  *string `json:"material,omitempty"`
	TargetSec *int    `json:"target_sec,omitempty"`
	Skip      *bool   `json:"skip,omitempty"`
}

func (s *Server) handleReelsList(w http.ResponseWriter, _ *http.Request) {
	reels, err := pipeline.ListReels(s.projectRoot())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]ReelSummary, 0, len(reels))
	for _, l := range reels {
		summary, err := reelSummary(l)
		if err != nil {
			continue // a half-written reel should not break the list
		}
		out = append(out, summary)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleReelDetail(w http.ResponseWriter, r *http.Request) {
	_, lesson, err := pipeline.FindReel(s.projectRoot(), filepath.Base(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	spec, err := pipeline.LoadReelSpec(lesson.Dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	summary, err := reelSummary(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	detail := ReelDetail{ReelSummary: summary, SegmentList: segmentInfos(*spec)}
	// The plan is optional: a reel that has not been planned yet is a perfectly
	// good thing to open, and the editor shows the prompts until it has.
	if raw, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), pipeline.ReelPlanFileName)); err == nil {
		detail.Plan = raw
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) handleReelCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateReelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	spec := pipeline.ReelSpec{
		Title: strings.TrimSpace(req.Title),
		Brief: strings.TrimSpace(req.Brief),
	}
	for _, seg := range req.Segments {
		spec.Segments = append(spec.Segments, pipeline.ReelSegment{
			Template:  seg.Template,
			Prompt:    strings.TrimSpace(seg.Prompt),
			Material:  strings.TrimSpace(seg.Material),
			TargetSec: seg.TargetSec,
		})
	}
	// One assignment rather than four, for the reason the snippet handler
	// records: building the override per-field overwrote whichever was set
	// first.
	spec.Config = config.Config{Style: config.Style{
		Voice:    req.Voice,
		Captions: req.Captions,
		Mode:     req.Mode,
		Skin:     req.Skin,
	}}
	spec.EnsureSegmentIDs()
	if err := spec.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	course, lesson, err := pipeline.CreateReel(s.projectRoot(), spec)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	summary, err := reelSummary(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	stages, err := pipeline.ReelStages(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if req.PlanOnly {
		stages = []string{project.StagePlan}
	}
	runID, err := s.runs.StartStages(course, lesson, stages, false)
	if err != nil {
		// The reel exists and is re-runnable; a busy pipeline is not a reason
		// to throw the request away.
		writeJSON(w, http.StatusAccepted, CreateReelResponse{ReelSummary: summary})
		return
	}
	writeJSON(w, http.StatusCreated, CreateReelResponse{ReelSummary: summary, RunID: runID})
}

// handleReelRun re-runs an existing reel, which is what the editor calls after
// a segment has been changed. Nothing is forced: the stage machinery already
// knows which stages the edit invalidated.
func (s *Server) handleReelRun(w http.ResponseWriter, r *http.Request) {
	course, lesson, err := pipeline.FindReel(s.projectRoot(), filepath.Base(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	stages, err := pipeline.ReelStages(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	runID, err := s.runs.StartStages(course, lesson, stages, false)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"run_id": runID})
}

// handleReelSegmentPatch edits one segment of an existing reel.
//
// It writes reel.yaml and stops. Re-running is a separate call the page makes
// when the user asks for it — batching several edits into one run is the
// difference between one re-render and four, and only the person editing knows
// when they have finished.
func (s *Server) handleReelSegmentPatch(w http.ResponseWriter, r *http.Request) {
	_, lesson, err := pipeline.FindReel(s.projectRoot(), filepath.Base(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req PatchReelSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	spec, err := pipeline.LoadReelSpec(lesson.Dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	segID := r.PathValue("segment")
	idx := -1
	for i, seg := range spec.Segments {
		if seg.ID == segID {
			idx = i
			break
		}
	}
	if idx < 0 {
		writeError(w, http.StatusNotFound, fmt.Errorf("reel %q has no segment %q", spec.ID, segID))
		return
	}

	seg := &spec.Segments[idx]
	if req.Template != nil {
		if _, ok := pipeline.SnippetTemplates[*req.Template]; !ok {
			writeError(w, http.StatusBadRequest, fmt.Errorf("unknown template %q", *req.Template))
			return
		}
		seg.Template = *req.Template
	}
	if req.Prompt != nil {
		seg.Prompt = strings.TrimSpace(*req.Prompt)
	}
	if req.Material != nil {
		seg.Material = strings.TrimSpace(*req.Material)
	}
	if req.TargetSec != nil {
		seg.TargetSec = *req.TargetSec
	}
	if req.Skip != nil {
		seg.Skip = *req.Skip
	}
	if err := spec.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := pipeline.SaveReelSpec(lesson.Dir, spec); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, segmentInfos(*spec))
}

func (s *Server) handleReelDelete(w http.ResponseWriter, r *http.Request) {
	_, lesson, err := pipeline.FindReel(s.projectRoot(), filepath.Base(r.PathValue("id")))
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

// segmentInfos decorates the spec's segments with their template's gallery
// copy, so the editor never has to join the two itself.
func segmentInfos(spec pipeline.ReelSpec) []ReelSegmentInfo {
	out := make([]ReelSegmentInfo, 0, len(spec.Segments))
	for _, seg := range spec.Segments {
		info := ReelSegmentInfo{
			ID:        seg.ID,
			Template:  seg.Template,
			Prompt:    seg.Prompt,
			Material:  seg.Material,
			TargetSec: seg.TargetSec,
			Skip:      seg.Skip,
		}
		if tpl, ok := pipeline.SnippetTemplates[seg.Template]; ok {
			info.TemplateTitle = tpl.Title
			info.TemplateCategory = tpl.Category
		}
		out = append(out, info)
	}
	return out
}

// reelSummary describes one reel on disk.
func reelSummary(l *project.Lesson) (ReelSummary, error) {
	spec, err := pipeline.LoadReelSpec(l.Dir)
	if err != nil {
		return ReelSummary{}, err
	}
	out := ReelSummary{
		ID:       spec.ID,
		Title:    spec.Title,
		Brief:    spec.Brief,
		Segments: len(spec.Active()),
		Skipped:  len(spec.Segments) - len(spec.Active()),
	}
	if !spec.CreatedAt.IsZero() {
		out.CreatedAt = spec.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if _, err := os.Stat(filepath.Join(l.GeneratedDir(), pipeline.FinalVideoName)); err == nil {
		out.Ready = true
		// No "generated/" segment: handleArtifact resolves the path relative to
		// the lesson's generated dir already, so including it here asked for
		// generated/generated/final.mp4 and 404'd on every finished reel.
		out.VideoURL = fmt.Sprintf("/artifacts/%s/%s/%s",
			pipeline.ReelsCourseSlug, spec.ID, pipeline.FinalVideoName)
	}
	return out, nil
}

// CastReelRequest asks the model to choose the segments from a brief.
type CastReelRequest struct {
	Brief string `json:"brief"`
	Title string `json:"title,omitempty"`
	// Segments is how many to aim for (0 = the caster's own default).
	Segments int `json:"segments,omitempty"`
}

// CastReelResponse is the proposed structure, NOT a created reel.
//
// Casting returns a proposal rather than writing anything, which is the whole
// shape of the feature: the structure is a decision worth reading before nine
// planning calls are spent on it, and the page lets you change a pick before
// committing. Writing it here would make "cast" and "accept this cast" the same
// irreversible action.
type CastReelResponse struct {
	Title    string              `json:"title"`
	Segments []CreateReelSegment `json:"segments"`
}

func (s *Server) handleReelCast(w http.ResponseWriter, r *http.Request) {
	var req CastReelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if strings.TrimSpace(req.Brief) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("casting needs a brief — say what the whole piece is about"))
		return
	}
	course, err := pipeline.EnsureReelsCourse(s.projectRoot())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	// Layered over the defaults: the reels course records only what it
	// overrides, so its bare manifest names no model.
	cfg := config.Resolve(course.Config, config.Config{}, config.Config{})
	spec, err := pipeline.CastReel(r.Context(), s.env, req.Brief, req.Title, req.Segments, cfg)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	out := CastReelResponse{Title: spec.Title}
	for _, seg := range spec.Segments {
		out.Segments = append(out.Segments, CreateReelSegment{
			Template: seg.Template,
			Prompt:   seg.Prompt,
			// Casting proposes and the page POSTs the proposal back to create,
			// so anything missing here is lost on the round trip. Omitting the
			// material would have reproduced the original bug one layer up: the
			// CLI path would carry the facts and the studio path would not.
			Material: seg.Material,
		})
	}
	writeJSON(w, http.StatusOK, out)
}
