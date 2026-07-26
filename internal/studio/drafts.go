package studio

// Draft endpoints: the prompt-first authoring flow.
//
// Creating a lesson used to require picking a course first and then
// hand-writing lesson.md. These endpoints invert that: you describe a lesson,
// the model drafts a complete lesson.md, and only afterwards do you decide
// which course it joins — or that it joins none yet.
//
// A draft is an unfiled lesson directory under .coursesmith/drafts/<id>/. It
// holds the same lesson.md a course lesson holds, so filing it is a directory
// move plus a numbering prefix, and a draft can be previewed, edited and
// validated with the loaders the pipeline already uses.

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/pipeline"
	"github.com/enfec/coursesmith/internal/project"
)

// draftsDirName is the unfiled-lesson store under the state dir.
const draftsDirName = "drafts"

// draftMetaFileName records what a draft was asked for, alongside its
// lesson.md. Kept separate so lesson.md stays a plain lesson.
const draftMetaFileName = "draft.json"

// DraftMeta is the sidecar for one draft.
type DraftMeta struct {
	ID        string    `json:"id"`
	Prompt    string    `json:"prompt"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"created_at"`
	// Course is the slug the draft was written *for*, when the request named
	// one. It is a hint for the assign step, not a commitment.
	Course string `json:"course,omitempty"`
}

// DraftDetail is a draft plus its editable source.
type DraftDetail struct {
	DraftMeta
	Source string `json:"source"`
}

// CreateDraftRequest is the whole authoring input: one line of prose.
type CreateDraftRequest struct {
	Prompt string `json:"prompt"`
	// Course optionally scopes the draft to a course's audience and tone, and
	// tells the model what that course already covers. Purely advisory.
	Course string `json:"course,omitempty"`
}

// UpdateDraftRequest replaces a draft's lesson.md source.
type UpdateDraftRequest struct {
	Source string `json:"source"`
}

// AssignDraftRequest files a draft into a course — the "which vault does this
// go in" step. Exactly one of Course (existing slug) or NewCourse (name of a
// course to scaffold) must be set.
type AssignDraftRequest struct {
	Course    string `json:"course,omitempty"`
	NewCourse string `json:"new_course,omitempty"`
}

func (s *Server) draftsDir() string { return filepath.Join(s.stateDir, draftsDirName) }

func (s *Server) draftDir(id string) (string, error) {
	clean := filepath.Base(strings.TrimSpace(id))
	if clean == "" || clean == "." || clean == ".." {
		return "", fmt.Errorf("invalid draft id %q", id)
	}
	dir := filepath.Join(s.draftsDir(), clean)
	if _, err := os.Stat(filepath.Join(dir, project.LessonFileName)); err != nil {
		return "", fmt.Errorf("no draft %q", clean)
	}
	return dir, nil
}

func readDraftMeta(dir string) (DraftMeta, error) {
	var m DraftMeta
	raw, err := os.ReadFile(filepath.Join(dir, draftMetaFileName))
	if err != nil {
		// A draft whose sidecar is missing is still a usable lesson; fall back
		// to the directory name rather than hiding it from the list.
		if errors.Is(err, os.ErrNotExist) {
			return DraftMeta{ID: filepath.Base(dir)}, nil
		}
		return m, err
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return m, err
	}
	return m, nil
}

// handleDraftsList returns every unfiled draft, newest first.
func (s *Server) handleDraftsList(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(s.draftsDir())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]DraftMeta, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(s.draftsDir(), e.Name())
		if _, statErr := os.Stat(filepath.Join(dir, project.LessonFileName)); statErr != nil {
			continue
		}
		m, readErr := readDraftMeta(dir)
		if readErr != nil {
			continue
		}
		m.ID = e.Name()
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	writeJSON(w, http.StatusOK, out)
}

// handleDraftCreate drafts a lesson from a prompt and stores it unfiled.
func (s *Server) handleDraftCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("prompt is required"))
		return
	}

	// Course context is optional. When given it supplies the audience/tone the
	// draft should match and the titles it should not duplicate.
	cfg := config.Defaults()
	var opts pipeline.AuthorOptions
	if req.Course != "" {
		course, err := s.resolveCourse(req.Course)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		cfg = config.Resolve(course.Config, config.Config{}, config.Config{})
		opts.CourseName = course.Name
		opts.CourseDescription = course.Description
		if lessons, err := course.Lessons(); err == nil {
			for _, l := range lessons {
				opts.ExistingLessons = append(opts.ExistingLessons, l.FrontMatter.Title)
			}
		}
	}

	draft, err := s.env.AuthorLesson(r.Context(), cfg, prompt, opts)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	id, dir, err := s.newDraftDir(draft.Title)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	source := draft.Markdown()
	if err := os.WriteFile(filepath.Join(dir, project.LessonFileName), []byte(source), 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	meta := DraftMeta{
		ID:        id,
		Prompt:    prompt,
		Title:     draft.Title,
		Summary:   draft.Summary,
		CreatedAt: time.Now().UTC(),
		Course:    req.Course,
	}
	if err := writeDraftMeta(dir, meta); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, DraftDetail{DraftMeta: meta, Source: source})
}

// newDraftDir reserves a unique directory for a draft titled title.
func (s *Server) newDraftDir(title string) (id, dir string, err error) {
	if err := os.MkdirAll(s.draftsDir(), 0o755); err != nil {
		return "", "", err
	}
	base := project.Slugify(title)
	if base == "" {
		base = "draft"
	}
	for i := 0; ; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i+1)
		}
		dir := filepath.Join(s.draftsDir(), candidate)
		if err := os.Mkdir(dir, 0o755); err == nil {
			return candidate, dir, nil
		} else if !errors.Is(err, os.ErrExist) {
			return "", "", err
		}
	}
}

func writeDraftMeta(dir string, m DraftMeta) error {
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(dir, draftMetaFileName), append(raw, '\n'), 0o644)
}

// handleDraftDetail returns one draft with its source.
func (s *Server) handleDraftDetail(w http.ResponseWriter, r *http.Request) {
	dir, err := s.draftDir(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	meta, err := readDraftMeta(dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	source, err := os.ReadFile(filepath.Join(dir, project.LessonFileName))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	meta.ID = filepath.Base(dir)
	writeJSON(w, http.StatusOK, DraftDetail{DraftMeta: meta, Source: string(source)})
}

// handleDraftUpdate replaces a draft's source, validating it first.
func (s *Server) handleDraftUpdate(w http.ResponseWriter, r *http.Request) {
	dir, err := s.draftDir(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req UpdateDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if err := validateDraftSource(filepath.Base(dir), req.Source); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := writeFileAtomic(filepath.Join(dir, project.LessonFileName), []byte(req.Source), 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	meta, err := readDraftMeta(dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	meta.ID = filepath.Base(dir)
	writeJSON(w, http.StatusOK, DraftDetail{DraftMeta: meta, Source: req.Source})
}

// handleDraftDelete discards a draft.
func (s *Server) handleDraftDelete(w http.ResponseWriter, r *http.Request) {
	dir, err := s.draftDir(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err := os.RemoveAll(dir); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDraftAssign files a draft into a course: the draft directory becomes
// the next numbered lesson directory. The move is the commit point — a draft
// is either unfiled or a lesson, never both.
func (s *Server) handleDraftAssign(w http.ResponseWriter, r *http.Request) {
	dir, err := s.draftDir(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req AssignDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if (req.Course == "") == (req.NewCourse == "") {
		writeError(w, http.StatusBadRequest, fmt.Errorf("set exactly one of course or new_course"))
		return
	}

	// Refuse to file a draft that would not load as a lesson, so a course can
	// never acquire a broken lesson through this path.
	source, err := os.ReadFile(filepath.Join(dir, project.LessonFileName))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := validateDraftSource(filepath.Base(dir), string(source)); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("draft is not a valid lesson: %w", err))
		return
	}

	var course *project.Course
	if req.NewCourse != "" {
		res, scaffoldErr := project.ScaffoldCourse(s.coursesDir, req.NewCourse, "")
		if scaffoldErr != nil {
			writeError(w, http.StatusBadRequest, scaffoldErr)
			return
		}
		// ScaffoldCourse seeds an example lesson; the draft being filed is the
		// real first lesson, so drop the placeholder.
		if lessons, lerr := project.LoadCourse(res.CourseDir); lerr == nil {
			if existing, xerr := lessons.Lessons(); xerr == nil {
				for _, l := range existing {
					os.RemoveAll(l.Dir)
				}
			}
		}
		course, err = project.LoadCourse(res.CourseDir)
	} else {
		course, err = s.resolveCourse(req.Course)
	}
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	lessonsDir := filepath.Join(course.Dir, "lessons")
	if err := os.MkdirAll(lessonsDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	meta, _ := readDraftMeta(dir)
	slug := project.Slugify(meta.Title)
	if slug == "" {
		slug = filepath.Base(dir)
	}
	target := filepath.Join(lessonsDir, fmt.Sprintf("%02d-%s", nextLessonPrefix(lessonsDir), slug))
	if _, err := os.Stat(target); err == nil {
		writeError(w, http.StatusConflict, fmt.Errorf("lesson %q already exists", filepath.Base(target)))
		return
	}
	if err := os.Rename(dir, target); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("filing draft: %w", err))
		return
	}
	// The sidecar described an unfiled draft; it has no meaning inside a course.
	os.Remove(filepath.Join(target, draftMetaFileName))

	lesson, err := project.LoadLesson(target)
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

// validateDraftSource checks that source loads as a lesson before it is
// written or filed, so a course can never acquire a broken lesson this way.
// It reuses the lesson editor's validator against a throwaway directory.
func validateDraftSource(id, source string) error {
	if strings.TrimSpace(source) == "" {
		return fmt.Errorf("lesson source is empty")
	}
	return validateLessonSource(id, source)
}
