package studio

// The no-code API: pieces where every segment stands on a real recording or a
// stated fact.
//
// It mirrors the reels API because the object is the same shape — several
// segments cut onto one timeline — and differs in exactly one place, which is
// the place that matters: a segment carries **evidence**, and the page's job is
// to make that impossible to leave out. So the detail response labels every
// segment with what backs it, and the catalog it offers is the no-code subset
// rather than the whole gallery.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/enfec/coursesmith/internal/pipeline"
	"github.com/enfec/coursesmith/internal/project"
)

// NoCodeEvidenceInfo is what backs a segment, as the page shows it.
type NoCodeEvidenceInfo struct {
	Kind string `json:"kind"`
	// Tool, Of, Take and Fixture describe a capture. Tool is the key; Display
	// is how it is named to a reader, resolved here so the page does not carry
	// its own copy of the registry.
	Tool    string `json:"tool,omitempty"`
	Display string `json:"display,omitempty"`
	Of      string `json:"of,omitempty"`
	Take    string `json:"take,omitempty"`
	Fixture string `json:"fixture,omitempty"`
	// Facts back a fact-kind segment.
	Facts []string `json:"facts,omitempty"`
	// Recorded is true once the capture actually exists on disk. This is the
	// single most useful thing the page can show: a piece cannot be planned
	// until its recordings are made, and "not recorded yet" is a state, not an
	// error.
	Recorded bool `json:"recorded"`
	// Marks are the moments the recorder stamped, once it has run. What the
	// narration is allowed to talk about, and therefore worth seeing.
	Marks []string `json:"marks,omitempty"`
	// DurationMs is the real length of the recording.
	DurationMs int `json:"duration_ms,omitempty"`
}

// NoCodeSegmentInfo is one segment as the page shows it.
type NoCodeSegmentInfo struct {
	ID               string             `json:"id"`
	Template         string             `json:"template"`
	Prompt           string             `json:"prompt"`
	Skip             bool               `json:"skip,omitempty"`
	TemplateTitle    string             `json:"template_title"`
	TemplateCategory string             `json:"template_category"`
	Evidence         NoCodeEvidenceInfo `json:"evidence"`
}

// NoCodeSummary describes one piece in the list.
type NoCodeSummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Brief    string `json:"brief,omitempty"`
	Segments int    `json:"segments"`
	Skipped  int    `json:"skipped"`
	// Captures is how many segments stand on a recording, and Recorded how many
	// of those have actually been shot. The gap between them is what somebody
	// has to do before the piece can be planned.
	Captures  int    `json:"captures"`
	Recorded  int    `json:"recorded"`
	Ready    bool   `json:"ready"`
	VideoURL string `json:"video_url,omitempty"`
	// VideoPath is where the render lands on disk, relative to the project root.
	// Always set, whether or not it exists yet: "it will be written here" is as
	// useful before a run as after one.
	VideoPath string `json:"video_path,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// NoCodeDetail is a piece plus its plan, for the editor.
type NoCodeDetail struct {
	NoCodeSummary
	SegmentList []NoCodeSegmentInfo `json:"segment_list"`
	// Stages is exactly what a run of this piece will do, in order. Sent so the
	// page can show real progress against a real list rather than a spinner:
	// generating a piece takes minutes, and a spinner for minutes reads as
	// broken.
	Stages []string        `json:"stages,omitempty"`
	Plan   json.RawMessage `json:"plan,omitempty"`
}

// noCodeEvidenceInfo describes a segment's evidence, including whether the
// recording it names has actually been made.
func noCodeEvidenceInfo(l *project.Lesson, seg pipeline.NoCodeSegment) NoCodeEvidenceInfo {
	info := NoCodeEvidenceInfo{Kind: string(seg.Evidence.Kind), Facts: seg.Evidence.Facts}
	if seg.Evidence.Kind != pipeline.EvidenceCapture || seg.Evidence.Capture == nil {
		return info
	}
	c := seg.Evidence.Capture
	info.Tool, info.Of, info.Take, info.Fixture = c.Tool, c.Of, c.Take, c.Fixture
	info.Display = pipeline.CaptureDisplayNameFor(c.Tool)
	// A capture's id is its segment's id — see NoCodeSpec.CaptureIDs.
	if f, ok := pipeline.LessonFootageByID(l, seg.ID); ok {
		info.Recorded = true
		info.Marks = pipeline.FootageMarkNames(f)
		info.DurationMs = f.DurationMs
	}
	return info
}

func noCodeSummary(l *project.Lesson) (NoCodeSummary, error) {
	spec, err := pipeline.LoadNoCodeSpec(l.Dir)
	if err != nil {
		return NoCodeSummary{}, err
	}
	live := spec.Live()
	sum := NoCodeSummary{
		ID:       l.ID,
		Title:    spec.Title,
		Brief:    spec.Brief,
		Segments: len(live),
		Skipped:  len(spec.Segments) - len(live),
	}
	if !spec.CreatedAt.IsZero() {
		sum.CreatedAt = spec.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	for _, seg := range live {
		if seg.Evidence.Kind != pipeline.EvidenceCapture {
			continue
		}
		sum.Captures++
		if _, ok := pipeline.LessonFootageByID(l, seg.ID); ok {
			sum.Recorded++
		}
	}
	final := filepath.Join(l.GeneratedDir(), pipeline.FinalVideoName)
	if _, err := os.Stat(final); err == nil {
		sum.Ready = true
		// The artifact route, the same one snippets and reels use. It used to be
		// `/api/lessons/{id}/artifacts/{name}`, which is not a route this server
		// has — so a finished piece advertised a URL that 404'd and the page's
		// player showed nothing. That is the whole of "where did my clip go".
		sum.VideoURL = fmt.Sprintf("/artifacts/%s/%s/%s",
			pipeline.NoCodeCourseSlug, l.ID, pipeline.FinalVideoName)
	}
	// Where the file actually lands, said plainly. A creator who wants to cut it
	// into something else needs the path, and hunting for it under a dot-dir
	// nobody mentioned is how a finished render feels lost.
	sum.VideoPath = filepath.Join(pipeline.NoCodeRoot, "lessons", l.ID, "generated", pipeline.FinalVideoName)
	return sum, nil
}

func (s *Server) handleNoCodeList(w http.ResponseWriter, _ *http.Request) {
	pieces, err := pipeline.ListNoCodePieces(s.projectRoot())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]NoCodeSummary, 0, len(pieces))
	for _, l := range pieces {
		sum, err := noCodeSummary(l)
		if err != nil {
			continue // a half-written piece should not break the list
		}
		out = append(out, sum)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleNoCodeDetail(w http.ResponseWriter, r *http.Request) {
	_, lesson, err := pipeline.FindNoCodePiece(s.projectRoot(), filepath.Base(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	spec, err := pipeline.LoadNoCodeSpec(lesson.Dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sum, err := noCodeSummary(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	detail := NoCodeDetail{NoCodeSummary: sum}
	if stages, err := pipeline.NoCodeStages(lesson); err == nil {
		detail.Stages = stages
	}
	for _, seg := range spec.Segments {
		info := NoCodeSegmentInfo{
			ID:       seg.ID,
			Template: seg.Template,
			Prompt:   seg.Prompt,
			Skip:     seg.Skip,
			Evidence: noCodeEvidenceInfo(lesson, seg),
		}
		if tpl, ok := pipeline.SnippetTemplates[seg.Template]; ok {
			info.TemplateTitle, info.TemplateCategory = tpl.Title, tpl.Category
		}
		detail.SegmentList = append(detail.SegmentList, info)
	}
	if raw, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), pipeline.ReelPlanFileName)); err == nil {
		detail.Plan = raw
	}
	writeJSON(w, http.StatusOK, detail)
}

// noCodeCreateRequest is what the page posts to start a piece.
type noCodeCreateRequest struct {
	Title    string                `json:"title"`
	Brief    string                `json:"brief"`
	Segments []noCodeCreateSegment `json:"segments"`
}

type noCodeCreateSegment struct {
	Template string   `json:"template"`
	Prompt   string   `json:"prompt"`
	Kind     string   `json:"kind"`
	Tool     string   `json:"tool,omitempty"`
	Of       string   `json:"of,omitempty"`
	Take     string   `json:"take,omitempty"`
	Fixture  string   `json:"fixture,omitempty"`
	Facts    []string `json:"facts,omitempty"`
}

func (s *Server) handleNoCodeCreate(w http.ResponseWriter, r *http.Request) {
	var req noCodeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	spec := pipeline.NoCodeSpec{Title: strings.TrimSpace(req.Title), Brief: strings.TrimSpace(req.Brief)}
	for _, in := range req.Segments {
		seg := pipeline.NoCodeSegment{
			Template: in.Template,
			Prompt:   strings.TrimSpace(in.Prompt),
			Evidence: pipeline.NoCodeEvidence{Kind: pipeline.EvidenceKind(in.Kind)},
		}
		switch pipeline.EvidenceKind(in.Kind) {
		case pipeline.EvidenceCapture:
			seg.Evidence.Capture = &pipeline.NoCodeCapture{
				Tool: in.Tool, Of: strings.TrimSpace(in.Of),
				Take: in.Take, Fixture: in.Fixture,
			}
		case pipeline.EvidenceFact:
			for _, f := range in.Facts {
				if f = strings.TrimSpace(f); f != "" {
					seg.Evidence.Facts = append(seg.Evidence.Facts, f)
				}
			}
		}
		spec.Segments = append(spec.Segments, seg)
	}
	// Validation happens in NewNoCodePiece, so a piece with a hollow segment is
	// refused here rather than written and discovered later. The page shows the
	// message as-is: they are written to be read by whoever is editing.
	_, lesson, err := pipeline.NewNoCodePiece(s.projectRoot(), spec)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	sum, err := noCodeSummary(lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, sum)
}

func (s *Server) handleNoCodeRun(w http.ResponseWriter, r *http.Request) {
	course, lesson, err := pipeline.FindNoCodePiece(s.projectRoot(), filepath.Base(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	stages, err := pipeline.NoCodeStages(lesson)
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

// NoCodeCatalogEntry is one template the page may offer.
type NoCodeCatalogEntry struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	// NeedsCapture marks the templates that cannot be filled from a prompt —
	// `footage` is a recording or it is nothing. The page uses it to force the
	// evidence choice rather than letting somebody pick a template that will
	// refuse to plan.
	NeedsCapture bool `json:"needs_capture"`
}

// handleNoCodeCatalog serves the no-code subset of the template gallery.
//
// Separate from /api/snippet-templates on purpose: this surface excludes the
// drawn-figure templates, and a page that filtered the full catalog client-side
// would need its own copy of that rule — which is exactly how the two would
// drift.
func (s *Server) handleNoCodeCatalog(w http.ResponseWriter, _ *http.Request) {
	names := pipeline.NoCodeTemplateNames()
	out := make([]NoCodeCatalogEntry, 0, len(names))
	for _, name := range names {
		tpl, ok := pipeline.SnippetTemplates[name]
		if !ok || tpl.Shelved {
			continue
		}
		out = append(out, NoCodeCatalogEntry{
			Name:         name,
			Title:        tpl.Title,
			Description:  tpl.Description,
			Category:     tpl.Category,
			NeedsCapture: name == "footage",
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// handleNoCodeRecordables lists everything a capture segment may record.
func (s *Server) handleNoCodeRecordables(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, pipeline.Recordables())
}

// handleNoCodeTakes lists the take files a segment can actually name.
//
// `take:` is the one required field on this surface that is a filename rather
// than prose or a choice — spelled without its extension, in a directory nobody
// would think to look in. Serving the real list turns a spelling test into a
// pick, and a name that is not on it is a warning the page can show now rather
// than an error a run finds in five minutes.
func (s *Server) handleNoCodeTakes(w http.ResponseWriter, _ *http.Request) {
	takes := pipeline.AvailableTakes(s.projectRoot())
	if takes == nil {
		takes = []pipeline.AvailableTake{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"dir":   pipeline.NoCodeTakesDir(""),
		"takes": takes,
	})
}
