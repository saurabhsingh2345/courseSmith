package studio

// Course and lesson editing endpoints (Phase 7 studio-as-source-of-truth).
//
// These make the studio a real authoring surface: edit course metadata and the
// archetype/palette/motion selection, create/edit/delete lessons on disk, and
// browse the archetype catalog. Writes are validated by re-parsing through the
// same project loaders the pipeline uses, and land atomically (temp + rename)
// so a bad edit can never truncate a course.yaml or lesson.md.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/pipeline"
	"github.com/enfec/coursesmith/internal/project"
)

// UpdateCourseRequest patches course.yaml. Every field is optional (pointer):
// nil means "leave as-is", so the UI can PATCH just what changed. Colors maps
// to branding.colors.
type UpdateCourseRequest struct {
	Name           *string        `json:"name,omitempty"`
	Description    *string        `json:"description,omitempty"`
	Archetype      *string        `json:"archetype,omitempty"`
	AnimationStyle *string        `json:"animation_style,omitempty"`
	ColorPalette   *string        `json:"color_palette,omitempty"`
	Colors         *config.Colors `json:"colors,omitempty"`
}

// handleCourseUpdate applies a metadata patch to course.yaml, preserving every
// other key. It validates the result by re-loading through project.LoadCourse
// (KnownFields) and pipeline.ResolveArchetype before writing.
func (s *Server) handleCourseUpdate(w http.ResponseWriter, r *http.Request) {
	course, err := s.resolveCourse(r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req UpdateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	path := filepath.Join(course.Dir, project.CourseFileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("parsing %s: %w", project.CourseFileName, err))
		return
	}
	if doc == nil {
		doc = map[string]any{}
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			writeError(w, http.StatusBadRequest, fmt.Errorf("name cannot be empty"))
			return
		}
		doc["name"] = *req.Name
	}
	if req.Description != nil {
		doc["description"] = *req.Description
	}
	if req.Archetype != nil || req.AnimationStyle != nil || req.ColorPalette != nil {
		style := nestedMap(doc, "style")
		setOrClear(style, "archetype", req.Archetype)
		setOrClear(style, "animation_style", req.AnimationStyle)
		setOrClear(style, "color_palette", req.ColorPalette)
	}
	if req.Colors != nil {
		branding := nestedMap(doc, "branding")
		colors := nestedMap(branding, "colors")
		if req.Colors.Primary != "" {
			colors["primary"] = req.Colors.Primary
		}
		if req.Colors.Accent != "" {
			colors["accent"] = req.Colors.Accent
		}
		if req.Colors.Background != "" {
			colors["background"] = req.Colors.Background
		}
	}

	out, err := yaml.Marshal(doc)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	// Validate the edit before it touches disk: it must parse with KnownFields
	// and resolve to a real archetype/animation/palette.
	var check project.Course
	dec := yaml.NewDecoder(strings.NewReader(string(out)))
	dec.KnownFields(true)
	if err := dec.Decode(&check); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("edit produced invalid course.yaml: %w", err))
		return
	}
	check.Dir = course.Dir
	if err := check.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if _, err := pipeline.ResolveArchetype(check.Config.Style); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := writeFileAtomic(path, out, 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeCourseDetail(w, course.Dir)
}

// handleCourseDelete removes a course directory. Guarded to stay inside the
// courses dir so a crafted slug can't escape it.
func (s *Server) handleCourseDelete(w http.ResponseWriter, r *http.Request) {
	course, err := s.resolveCourse(r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if !withinDir(s.coursesDir, course.Dir) {
		writeError(w, http.StatusBadRequest, fmt.Errorf("refusing to delete outside courses dir"))
		return
	}
	if err := os.RemoveAll(course.Dir); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateLessonRequest is the POST /api/lessons/{course} payload.
type CreateLessonRequest struct {
	Title string `json:"title"`
}

// handleLessonCreate scaffolds a new lesson directory with the next numeric
// prefix and a minimal lesson.md, then returns its detail payload.
func (s *Server) handleLessonCreate(w http.ResponseWriter, r *http.Request) {
	course, err := s.resolveCourse(r.PathValue("course"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req CreateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("title is required"))
		return
	}

	lessonsDir := filepath.Join(course.Dir, "lessons")
	next := nextLessonPrefix(lessonsDir)
	slug := project.Slugify(title)
	if slug == "" {
		slug = "lesson"
	}
	dirName := fmt.Sprintf("%02d-%s", next, slug)
	lessonDir := filepath.Join(lessonsDir, dirName)
	if _, err := os.Stat(lessonDir); err == nil {
		writeError(w, http.StatusConflict, fmt.Errorf("lesson %q already exists", dirName))
		return
	}
	if err := os.MkdirAll(lessonDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	body := fmt.Sprintf("---\ntitle: %s\n---\n\n# %s\n\n## Overview\n- Replace this outline with your lesson's key points.\n", strconv.Quote(title), title)
	if err := os.WriteFile(filepath.Join(lessonDir, project.LessonFileName), []byte(body), 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	lesson, err := project.LoadLesson(lessonDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	detail, err := s.buildLessonDetail(course, lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, detail)
}

// UpdateLessonRequest replaces the lesson.md source wholesale.
type UpdateLessonRequest struct {
	Source string `json:"source"`
}

// handleLessonUpdate writes new lesson.md content after validating that it
// parses as a lesson (front-matter title present). It never writes an invalid
// lesson — validation happens against a temp copy first.
func (s *Server) handleLessonUpdate(w http.ResponseWriter, r *http.Request) {
	course, lesson, err := s.resolveLesson(r.PathValue("course"), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req UpdateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if err := validateLessonSource(lesson.ID, req.Source); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := writeFileAtomic(lesson.SourcePath(), []byte(req.Source), 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	reloaded, err := project.LoadLesson(lesson.Dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	detail, err := s.buildLessonDetail(course, reloaded)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// handleLessonDelete removes a lesson directory (source + generated artifacts).
func (s *Server) handleLessonDelete(w http.ResponseWriter, r *http.Request) {
	_, lesson, err := s.resolveLesson(r.PathValue("course"), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if !withinDir(s.coursesDir, lesson.Dir) {
		writeError(w, http.StatusBadRequest, fmt.Errorf("refusing to delete outside courses dir"))
		return
	}
	if err := os.RemoveAll(lesson.Dir); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleArchetypes exposes the archetype catalog (registry + animation styles +
// palettes) so the studio can offer them without hard-coding the Go registry.
func (s *Server) handleArchetypes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"archetypes":       pipeline.ArchetypeList(),
		"animation_styles": pipeline.AnimationStyleNames(),
		"palettes":         pipeline.ColorPaletteList(),
	})
}

// writeCourseDetail reloads a course from disk and writes its CourseDetail.
func (s *Server) writeCourseDetail(w http.ResponseWriter, dir string) {
	course, err := project.LoadCourse(dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	lessons, err := course.Lessons()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	courseCfg := config.Resolve(course.Config, config.Config{}, config.Config{})
	detail := CourseDetail{
		Course:     Course{Name: course.Name, Slug: course.Slug, Description: course.Description, LessonCount: len(lessons)},
		StageOrder: project.StagesFor(courseCfg.Pipeline.VideoOnly),
		Meta:       courseMeta(course),
	}
	for _, l := range lessons {
		cfg := config.Resolve(course.Config, l.FrontMatter.Overrides(), config.Config{})
		statuses, err := s.env.LessonStatus(l, cfg)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		stages := make(map[string]string, len(statuses))
		for stage, status := range statuses {
			stages[stage] = string(status)
		}
		detail.Lessons = append(detail.Lessons, LessonSummary{ID: l.ID, Title: l.FrontMatter.Title, Stages: stages})
	}
	writeJSON(w, http.StatusOK, detail)
}

// --- helpers ---

// nestedMap returns doc[key] as a map, creating it if missing or the wrong type.
func nestedMap(doc map[string]any, key string) map[string]any {
	if existing, ok := doc[key].(map[string]any); ok {
		return existing
	}
	m := map[string]any{}
	doc[key] = m
	return m
}

// setOrClear sets key to *val, or deletes it when *val is the empty string, so
// the UI can clear a selection back to "inherit".
func setOrClear(m map[string]any, key string, val *string) {
	if val == nil {
		return
	}
	if strings.TrimSpace(*val) == "" {
		delete(m, key)
		return
	}
	m[key] = *val
}

// validateLessonSource confirms new lesson.md content parses as a lesson by
// loading it from a throwaway directory named after the lesson id.
func validateLessonSource(lessonID, source string) error {
	tmp, err := os.MkdirTemp("", "lesson-validate-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	dir := filepath.Join(tmp, lessonID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, project.LessonFileName), []byte(source), 0o644); err != nil {
		return err
	}
	if _, err := project.LoadLesson(dir); err != nil {
		return fmt.Errorf("invalid lesson source: %w", err)
	}
	return nil
}

// nextLessonPrefix returns 1 + the highest NN- prefix in lessonsDir (1 if none).
func nextLessonPrefix(lessonsDir string) int {
	entries, err := os.ReadDir(lessonsDir)
	if err != nil {
		return 1
	}
	max := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		i := strings.IndexByte(name, '-')
		if i <= 0 {
			continue
		}
		if n, err := strconv.Atoi(name[:i]); err == nil && n > max {
			max = n
		}
	}
	return max + 1
}

// withinDir reports whether target is parent or a descendant path under parent.
func withinDir(parent, target string) bool {
	pa, err1 := filepath.Abs(parent)
	ta, err2 := filepath.Abs(target)
	if err1 != nil || err2 != nil {
		return false
	}
	rel, err := filepath.Rel(pa, ta)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// writeFileAtomic writes data to a temp file in the same dir then renames it
// over path, so readers never see a partial write.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
