package studio

// The combos API: create a multi-template video, watch it run, and edit it
// afterwards.
//
// The endpoints mirror the snippets ones exactly where they can, because a combo
// is the same object with more than one template in it — same synthetic course,
// same stage machinery, same SSE run stream. The one shape that has no snippet
// equivalent is PATCH on a segment, which is the whole point of the page: a
// combo is watched and then adjusted, and an adjustment has to be addressable to
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

// ComboSegmentInfo is one segment as the page shows it.
type ComboSegmentInfo struct {
	ID       string `json:"id"`
	Template string `json:"template"`
	Prompt   string `json:"prompt"`
	// Material is the concrete facts this segment will be filled with. Surfaced
	// because it is the field most worth correcting by hand: it is what the
	// segment's writer plans from, so a wrong figure here is a wrong figure in
	// the finished video, and fixing it is one edit rather than a re-cast.
	Material string `json:"material,omitempty"`
	// Heading, Role and Why are the director's account of this segment: what
	// part of the argument it is, what job it does in the arc, and why this look
	// was chosen for it.
	//
	// Read-only in the editor, and surfaced for a reason the material's comment
	// already gives in the other direction. The material is what you correct; the
	// role and the heading are what tell you WHETHER to. A segment you are about
	// to rewrite reads very differently once you can see it is the piece's only
	// hook. Empty on a hand-authored combo, which is why nothing depends on them.
	Heading string `json:"heading,omitempty"`
	Role    string `json:"role,omitempty"`
	Why     string `json:"why,omitempty"`
	// TargetSec is 0 when the segment takes its template's default.
	TargetSec int `json:"target_sec,omitempty"`
	// Skip means the segment stays in the file but leaves the cut.
	Skip bool `json:"skip,omitempty"`
	// Title and Category come from the template catalog so the editor can label
	// a segment without holding its own copy of the catalog.
	TemplateTitle    string `json:"template_title"`
	TemplateCategory string `json:"template_category"`
}

// ComboSummary describes one combo in the list.
type ComboSummary struct {
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

// ComboDetail is a combo's request plus the plan, for the editor.
type ComboDetail struct {
	ComboSummary
	SegmentList []ComboSegmentInfo `json:"segment_list"`
	// Plan is combo-plan.json once the plan stage has run, so the editor can
	// show the narration a segment actually produced rather than only the
	// prompt that asked for it.
	Plan json.RawMessage `json:"plan,omitempty"`
}

// CreateComboRequest is the whole input: a brief, and the ordered segments.
type CreateComboRequest struct {
	Title string `json:"title,omitempty"`
	Brief string `json:"brief,omitempty"`
	// Angle is what the piece argues, carried back from the director's proposal.
	//
	// Round-tripped rather than re-derived, because it is what the critic scores
	// every finished segment against — a combo created without one gets a critic
	// that can only ask "is this good", which returns opinions about prose
	// instead of the one judgement this pass exists for.
	Angle    string               `json:"angle,omitempty"`
	Segments []CreateComboSegment `json:"segments"`
	Voice    string               `json:"voice,omitempty"`
	Captions string               `json:"captions,omitempty"`
	Mode     string               `json:"mode,omitempty"`
	// Skin is the theme: "" | "default" | "broadcast" | "minimal" | "editorial".
	Skin string `json:"skin,omitempty"`
	// PlanOnly stops after planning, for reviewing the design before paying
	// for TTS and a render. On a combo that matters more than on a snippet:
	// planning is one call per segment, and rendering is minutes.
	PlanOnly bool `json:"plan_only,omitempty"`
}

// CreateComboSegment is one requested segment.
type CreateComboSegment struct {
	Template string `json:"template"`
	Prompt   string `json:"prompt"`
	// Heading, Role and Why are the director's reasoning, carried through so the
	// page can show WHY the piece is shaped this way rather than only what it
	// chose. All optional: a hand-authored combo supplies none of them and plans
	// exactly as it did before they existed.
	Heading string `json:"heading,omitempty"`
	Role    string `json:"role,omitempty"`
	Why     string `json:"why,omitempty"`
	// Material is optional on a hand-authored combo — a segment without it plans
	// from the prompt alone, exactly as it did before the field existed. Supply
	// it and the writer stops guessing.
	Material  string `json:"material,omitempty"`
	TargetSec int    `json:"target_sec,omitempty"`
}

// CreateComboResponse is the created combo plus the run watching it.
type CreateComboResponse struct {
	ComboSummary
	RunID string `json:"run_id,omitempty"`
}

// PatchComboSegmentRequest edits one segment. Every field is a pointer so the
// editor can change exactly one thing: a plain string could not distinguish
// "set the prompt to empty" from "leave the prompt alone", and the second is
// what almost every request means.
type PatchComboSegmentRequest struct {
	Template  *string `json:"template,omitempty"`
	Prompt    *string `json:"prompt,omitempty"`
	Material  *string `json:"material,omitempty"`
	TargetSec *int    `json:"target_sec,omitempty"`
	Skip      *bool   `json:"skip,omitempty"`
}

func (s *Server) handleCombosList(w http.ResponseWriter, _ *http.Request) {
	combos, err := pipeline.ListCombos(s.projectRoot())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]ComboSummary, 0, len(combos))
	for _, l := range combos {
		summary, err := comboSummary(l)
		if err != nil {
			continue // a half-written combo should not break the list
		}
		out = append(out, summary)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleComboDetail(w http.ResponseWriter, r *http.Request) {
	_, lesson, err := pipeline.FindCombo(s.projectRoot(), filepath.Base(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	spec, err := pipeline.LoadComboSpec(lesson.Dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	summary, err := comboSummary(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	detail := ComboDetail{ComboSummary: summary, SegmentList: segmentInfos(*spec)}
	// The plan is optional: a combo that has not been planned yet is a perfectly
	// good thing to open, and the editor shows the prompts until it has.
	if raw, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), pipeline.ComboPlanFileName)); err == nil {
		detail.Plan = raw
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) handleComboCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateComboRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	spec := pipeline.ComboSpec{
		Title: strings.TrimSpace(req.Title),
		Brief: strings.TrimSpace(req.Brief),
		Angle: strings.TrimSpace(req.Angle),
	}
	for _, seg := range req.Segments {
		spec.Segments = append(spec.Segments, pipeline.ComboSegment{
			Template:  seg.Template,
			Prompt:    strings.TrimSpace(seg.Prompt),
			Heading:   strings.TrimSpace(seg.Heading),
			Role:      strings.TrimSpace(seg.Role),
			Why:       strings.TrimSpace(seg.Why),
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

	course, lesson, err := pipeline.CreateCombo(s.projectRoot(), spec)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	summary, err := comboSummary(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	stages, err := pipeline.ComboStages(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if req.PlanOnly {
		stages = []string{project.StagePlan}
	}
	runID, err := s.runs.StartStages(course, lesson, stages, false)
	if err != nil {
		// The combo exists and is re-runnable; a busy pipeline is not a reason
		// to throw the request away.
		writeJSON(w, http.StatusAccepted, CreateComboResponse{ComboSummary: summary})
		return
	}
	writeJSON(w, http.StatusCreated, CreateComboResponse{ComboSummary: summary, RunID: runID})
}

// handleComboRun re-runs an existing combo, which is what the editor calls after
// a segment has been changed. Nothing is forced: the stage machinery already
// knows which stages the edit invalidated.
func (s *Server) handleComboRun(w http.ResponseWriter, r *http.Request) {
	course, lesson, err := pipeline.FindCombo(s.projectRoot(), filepath.Base(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	stages, err := pipeline.ComboStages(lesson)
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

// handleComboSegmentPatch edits one segment of an existing combo.
//
// It writes combo.yaml and stops. Re-running is a separate call the page makes
// when the user asks for it — batching several edits into one run is the
// difference between one re-render and four, and only the person editing knows
// when they have finished.
func (s *Server) handleComboSegmentPatch(w http.ResponseWriter, r *http.Request) {
	_, lesson, err := pipeline.FindCombo(s.projectRoot(), filepath.Base(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req PatchComboSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	spec, err := pipeline.LoadComboSpec(lesson.Dir)
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
		writeError(w, http.StatusNotFound, fmt.Errorf("combo %q has no segment %q", spec.ID, segID))
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
	if err := pipeline.SaveComboSpec(lesson.Dir, spec); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, segmentInfos(*spec))
}

func (s *Server) handleComboDelete(w http.ResponseWriter, r *http.Request) {
	_, lesson, err := pipeline.FindCombo(s.projectRoot(), filepath.Base(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if s.refuseDeleteWhileRunning(w, pipeline.CombosCourseSlug, lesson.ID) {
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
func segmentInfos(spec pipeline.ComboSpec) []ComboSegmentInfo {
	out := make([]ComboSegmentInfo, 0, len(spec.Segments))
	for _, seg := range spec.Segments {
		info := ComboSegmentInfo{
			ID:        seg.ID,
			Template:  seg.Template,
			Prompt:    seg.Prompt,
			Material:  seg.Material,
			Heading:   seg.Heading,
			Role:      seg.Role,
			Why:       seg.Why,
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

// comboSummary describes one combo on disk.
func comboSummary(l *project.Lesson) (ComboSummary, error) {
	spec, err := pipeline.LoadComboSpec(l.Dir)
	if err != nil {
		return ComboSummary{}, err
	}
	out := ComboSummary{
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
		// generated/generated/final.mp4 and 404'd on every finished combo.
		out.VideoURL = fmt.Sprintf("/artifacts/%s/%s/%s",
			pipeline.CombosCourseSlug, spec.ID, pipeline.FinalVideoName)
	}
	return out, nil
}

// DirectComboRequest is the whole front door: a subject and four choices.
//
// Everything else the piece needs — how it divides, what it argues, which look
// carries which part, how long each one runs — is decided by the director. That
// asymmetry is the point of this surface, and it is why this struct is smaller
// than the one it replaced rather than larger.
type DirectComboRequest struct {
	// Subject is what the video is about, in the creator's words.
	Subject string `json:"subject"`
	Title   string `json:"title,omitempty"`
	// Minutes is how long the piece should run. Zero reads a length out of the
	// subject if one is stated there, and otherwise takes the default.
	Minutes int `json:"minutes,omitempty"`
	// Skin is the theme, which also decides which templates may be cast.
	Skin     string `json:"skin,omitempty"`
	Captions string `json:"captions,omitempty"`
	Mode     string `json:"mode,omitempty"`
}

// DirectComboResponse is the proposed piece, NOT a created one.
//
// Directing returns a proposal rather than writing anything, which is the whole
// shape of the feature: the structure is a decision worth reading before a dozen
// planning calls are spent on it, and the page lets you change a pick before
// committing. Writing it here would make "direct" and "accept this" the same
// irreversible action.
type DirectComboResponse struct {
	Title string `json:"title"`
	// Angle is what the piece argues — the line every segment is judged against,
	// by the director now and by the critic later.
	Angle string `json:"angle"`
	// Pool and Runtime are the two decisions a person most often wants to
	// overrule, said in words rather than left to be inferred from the picks:
	// which catalog the theme narrowed this to, and how the runtime was spread.
	Pool     string               `json:"pool"`
	Runtime  string               `json:"runtime,omitempty"`
	Segments []CreateComboSegment `json:"segments"`
}

func (s *Server) handleComboDirect(w http.ResponseWriter, r *http.Request) {
	var req DirectComboRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if strings.TrimSpace(req.Subject) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("directing needs a subject — say what the video is about"))
		return
	}
	course, err := pipeline.EnsureCombosCourse(s.projectRoot())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	// Layered over the defaults: the combos course records only what it
	// overrides, so its bare manifest names no model.
	cfg := config.Resolve(course.Config, config.Config{}, config.Config{})
	result, err := pipeline.DirectCombo(r.Context(), s.env, pipeline.ComboRequest{
		Subject:  req.Subject,
		Title:    req.Title,
		Minutes:  req.Minutes,
		Skin:     req.Skin,
		Captions: req.Captions,
		Mode:     req.Mode,
	}, cfg)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	out := DirectComboResponse{
		Title:   result.Spec.Title,
		Angle:   result.Spec.Angle,
		Pool:    pipeline.ComboPoolDescribe(req.Skin),
		Runtime: result.Budget.Describe(),
	}
	for _, seg := range result.Spec.Segments {
		// The page POSTs this proposal straight back to create, so anything
		// dropped here is lost on the round trip. That is not hypothetical: the
		// material used to be omitted, which reproduced the original bug one
		// layer up — the CLI path carried the facts and the studio path did not.
		out.Segments = append(out.Segments, CreateComboSegment{
			Template:  seg.Template,
			Prompt:    seg.Prompt,
			Heading:   seg.Heading,
			Role:      seg.Role,
			Material:  seg.Material,
			Why:       seg.Why,
			TargetSec: seg.TargetSec,
		})
	}
	writeJSON(w, http.StatusOK, out)
}
