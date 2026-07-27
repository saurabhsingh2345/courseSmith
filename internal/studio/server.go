package studio

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/pipeline"
	"github.com/enfec/coursesmith/internal/project"
)

// Server is the Studio backend.
type Server struct {
	env        *pipeline.Env
	coursesDir string
	stateDir   string // .coursesmith (llm cache + ratelimit state)
	hub        *eventHub
	runs       *runManager
	// uiDir serves the built studio frontend when present ("" disables).
	uiDir string
}

// New builds a Server around a pipeline environment.
func New(env *pipeline.Env, coursesDir, stateDir, uiDir string) *Server {
	hub := newEventHub()
	return &Server{
		env:        env,
		coursesDir: coursesDir,
		stateDir:   stateDir,
		hub:        hub,
		runs:       newRunManager(env, hub),
		uiDir:      uiDir,
	}
}

// Handler returns the full route table.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/courses", s.handleCourses)
	mux.HandleFunc("POST /api/courses", s.handleCourseCreate)
	mux.HandleFunc("GET /api/courses/{slug}", s.handleCourseDetail)
	mux.HandleFunc("PUT /api/courses/{slug}", s.handleCourseUpdate)
	mux.HandleFunc("DELETE /api/courses/{slug}", s.handleCourseDelete)
	mux.HandleFunc("GET /api/lessons/{course}/{id}", s.handleLessonDetail)
	mux.HandleFunc("POST /api/lessons/{course}", s.handleLessonCreate)
	mux.HandleFunc("PUT /api/lessons/{course}/{id}", s.handleLessonUpdate)
	mux.HandleFunc("DELETE /api/lessons/{course}/{id}", s.handleLessonDelete)
	// Prompt-first authoring: draft a lesson unfiled, then choose its course.
	mux.HandleFunc("GET /api/drafts", s.handleDraftsList)
	mux.HandleFunc("POST /api/drafts", s.handleDraftCreate)
	mux.HandleFunc("GET /api/drafts/{id}", s.handleDraftDetail)
	mux.HandleFunc("PUT /api/drafts/{id}", s.handleDraftUpdate)
	mux.HandleFunc("DELETE /api/drafts/{id}", s.handleDraftDelete)
	mux.HandleFunc("POST /api/drafts/{id}/assign", s.handleDraftAssign)
	// Short-form: prompt + visual template → one standalone clip.
	mux.HandleFunc("GET /api/snippet-templates", s.handleSnippetTemplates)
	mux.HandleFunc("GET /api/snippets", s.handleSnippetsList)
	mux.HandleFunc("POST /api/snippets", s.handleSnippetCreate)
	mux.HandleFunc("GET /api/snippets/{id}", s.handleSnippetDetail)
	mux.HandleFunc("DELETE /api/snippets/{id}", s.handleSnippetDelete)
	mux.HandleFunc("GET /api/archetypes", s.handleArchetypes)
	mux.HandleFunc("GET /api/library/diagrams", s.handleLibraryDiagramsList)
	mux.HandleFunc("POST /api/library/diagrams", s.handleLibraryDiagramCreate)
	mux.HandleFunc("DELETE /api/library/diagrams/{id}", s.handleLibraryDiagramDelete)
	mux.HandleFunc("GET /api/library/questions", s.handleLibraryQuestionsList)
	mux.HandleFunc("POST /api/library/questions", s.handleLibraryQuestionCreate)
	mux.HandleFunc("DELETE /api/library/questions/{id}", s.handleLibraryQuestionDelete)
	mux.HandleFunc("GET /api/events", s.handleEvents)
	mux.HandleFunc("GET /api/run", s.handleRunStatus)
	mux.HandleFunc("POST /api/run", s.handleRun)
	mux.HandleFunc("DELETE /api/run", s.handleRunCancel)
	mux.HandleFunc("POST /api/feedback", s.handleFeedback)
	mux.HandleFunc("POST /api/regenerate", s.handleRegenerate)
	mux.HandleFunc("GET /api/quiz/{course}/{id}", s.handleQuizGet)
	mux.HandleFunc("PUT /api/quiz-overrides/{course}/{id}", s.handleQuizOverridesPut)
	mux.HandleFunc("GET /api/ledger", s.handleLedger)
	mux.HandleFunc("GET /api/openapi.json", s.handleOpenAPI)
	mux.HandleFunc("GET /artifacts/{course}/{lesson}/", s.handleArtifact)
	if s.uiDir != "" {
		mux.Handle("GET /", spaHandler(s.uiDir))
	} else {
		mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{
				"name": "coursesmith studio API",
				"ui":   "studio frontend not built — cd studio && npm install && npm run build",
			})
		})
	}
	return mux
}

// spaHandler serves the built frontend, falling back to index.html for
// client-side routes.
func spaHandler(dir string) http.Handler {
	fileServer := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// resolveCourse loads a course by slug from the courses dir.
//
// The snippets slug is special: snippets live in a synthetic course under the
// state dir, not in courses/. Resolving it here means the lesson-detail, run,
// and artifact routes all serve snippets with no snippet-specific code — a
// snippet is a lesson, and this is what makes that true for the API too.
func (s *Server) resolveCourse(slug string) (*project.Course, error) {
	dir := filepath.Join(s.coursesDir, filepath.Base(slug))
	if _, err := os.Stat(filepath.Join(dir, project.CourseFileName)); err != nil {
		if filepath.Base(slug) == pipeline.SnippetsCourseSlug {
			return pipeline.EnsureSnippetsCourse(s.projectRoot())
		}
		return nil, fmt.Errorf("no course %q", slug)
	}
	return project.LoadCourse(dir)
}

// projectRoot is the directory the snippets store hangs off. The state dir is
// always <root>/.coursesmith, so its parent is the root.
func (s *Server) projectRoot() string {
	return filepath.Dir(s.stateDir)
}

func (s *Server) resolveLesson(courseSlug, lessonID string) (*project.Course, *project.Lesson, error) {
	course, err := s.resolveCourse(courseSlug)
	if err != nil {
		return nil, nil, err
	}
	lesson, err := course.FindLesson(lessonID)
	if err != nil {
		return nil, nil, err
	}
	return course, lesson, nil
}

// --- courses & matrix ---

// Course is the list-view shape.
type Course struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	LessonCount int    `json:"lesson_count"`
}

// LessonSummary is one row of the dashboard matrix.
type LessonSummary struct {
	ID     string            `json:"id"`
	Title  string            `json:"title"`
	Stages map[string]string `json:"stages"` // stage → done|stale|pending
}

// CourseMeta is the editable style selection surfaced to the course editor.
type CourseMeta struct {
	Archetype      string        `json:"archetype"`
	AnimationStyle string        `json:"animation_style"`
	ColorPalette   string        `json:"color_palette"`
	Colors         config.Colors `json:"colors"`
}

func courseMeta(c *project.Course) CourseMeta {
	return CourseMeta{
		Archetype:      c.Config.Style.Archetype,
		AnimationStyle: c.Config.Style.AnimationStyle,
		ColorPalette:   c.Config.Style.ColorPalette,
		Colors:         c.Config.Branding.Colors,
	}
}

// CourseDetail is the dashboard payload.
type CourseDetail struct {
	Course
	StageOrder []string        `json:"stage_order"`
	Lessons    []LessonSummary `json:"lessons"`
	Meta       CourseMeta      `json:"meta"`
}

func (s *Server) handleCourses(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(s.coursesDir)
	if err != nil && !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := []Course{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		c, err := project.LoadCourse(filepath.Join(s.coursesDir, e.Name()))
		if err != nil {
			continue
		}
		lessons, _ := c.Lessons()
		out = append(out, Course{Name: c.Name, Slug: c.Slug, Description: c.Description, LessonCount: len(lessons)})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCourseDetail(w http.ResponseWriter, r *http.Request) {
	course, err := s.resolveCourse(r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
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
			writeError(w, http.StatusInternalServerError, fmt.Errorf("lesson %s: %w", l.ID, err))
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

// --- lesson detail ---

// Artifact is one generated file exposed to the UI.
type Artifact struct {
	Name string `json:"name"` // path relative to generated/
	Size int64  `json:"size"`
	URL  string `json:"url"`
	// DownloadName is what this file should be called once it is saved — see
	// downloadName. The server already sends it as Content-Disposition and
	// browsers prefer that over an anchor's own `download` attribute, so this
	// is belt and braces; it is here so the UI never has to reimplement the
	// rule and drift from it.
	DownloadName string `json:"download_name"`
}

// LessonDetail is the full workbench payload: everything the studio needs
// to review one lesson. Raw JSON artifacts pass through as-is so the
// frontend types stay in lock-step with the pipeline formats.
type LessonDetail struct {
	Course     string            `json:"course"`
	ID         string            `json:"id"`
	Title      string            `json:"title"`
	Source     string            `json:"source"` // lesson.md content
	Stages     map[string]string `json:"stages"`
	StageOrder []string          `json:"stage_order"`
	Artifacts  []Artifact        `json:"artifacts"`

	Script    json.RawMessage            `json:"script,omitempty"`
	Quiz      json.RawMessage            `json:"quiz,omitempty"`
	Mistakes  json.RawMessage            `json:"mistakes,omitempty"`
	Exercises json.RawMessage            `json:"exercises,omitempty"`
	Chapters  json.RawMessage            `json:"chapters,omitempty"`
	Alignment json.RawMessage            `json:"alignment,omitempty"`
	Reviews   map[string]json.RawMessage `json:"reviews,omitempty"`
}

func (s *Server) handleLessonDetail(w http.ResponseWriter, r *http.Request) {
	course, lesson, err := s.resolveLesson(r.PathValue("course"), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	detail, err := s.buildLessonDetail(course, lesson)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// buildLessonDetail assembles the full workbench payload for one lesson. It is
// shared by GET /api/lessons/... and the lesson edit handlers so an edit can
// return the same shape the UI already renders.
func (s *Server) buildLessonDetail(course *project.Course, lesson *project.Lesson) (LessonDetail, error) {
	cfg := config.Resolve(course.Config, lesson.FrontMatter.Overrides(), config.Config{})
	// A snippet runs its own short pipeline; reporting the lesson stages it
	// never runs would show a wall of permanently-pending rows and hide its
	// plan stage entirely.
	stageOrder := project.StagesFor(cfg.Pipeline.VideoOnly)
	if pipeline.IsSnippet(lesson) {
		if snippetStages, err := pipeline.SnippetStages(lesson); err == nil {
			stageOrder = snippetStages
		}
	}
	statuses, err := s.env.LessonStatusFor(lesson, cfg, stageOrder)
	if err != nil {
		return LessonDetail{}, err
	}
	source, _ := os.ReadFile(lesson.SourcePath())

	detail := LessonDetail{
		Course:     course.Slug,
		ID:         lesson.ID,
		Title:      lesson.FrontMatter.Title,
		Source:     string(source),
		Stages:     map[string]string{},
		StageOrder: stageOrder,
		Reviews:    map[string]json.RawMessage{},
	}
	for stage, status := range statuses {
		detail.Stages[stage] = string(status)
	}

	gen := lesson.GeneratedDir()
	rawFile := func(name string) json.RawMessage {
		data, err := os.ReadFile(filepath.Join(gen, name))
		if err != nil || !json.Valid(data) {
			return nil
		}
		return data
	}
	detail.Script = rawFile(pipeline.ScriptFileName)
	detail.Quiz = rawFile(pipeline.QuizFileName)
	detail.Mistakes = rawFile(pipeline.MistakesFileName)
	detail.Exercises = rawFile(filepath.Join(pipeline.ExercisesDirName, pipeline.ExercisesManifestName))
	detail.Chapters = rawFile(pipeline.ChaptersJSONFileName)
	detail.Alignment = rawFile(pipeline.AlignmentFileName)

	reviewsDir := filepath.Join(gen, pipeline.ReviewsDirName)
	if entries, err := os.ReadDir(reviewsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
				continue
			}
			if data := rawFile(filepath.Join(pipeline.ReviewsDirName, e.Name())); data != nil {
				detail.Reviews[strings.TrimSuffix(e.Name(), ".json")] = data
			}
		}
	}

	// Artifact inventory (players, thumbnails, downloads).
	_ = filepath.WalkDir(gen, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(gen, path)
		if err != nil {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		detail.Artifacts = append(detail.Artifacts, Artifact{
			Name:         filepath.ToSlash(rel),
			Size:         info.Size(),
			URL:          "/artifacts/" + course.Slug + "/" + lesson.ID + "/" + filepath.ToSlash(rel),
			DownloadName: downloadName(course.Name, course.Slug, lesson.ID, lesson.FrontMatter.Title, filepath.ToSlash(rel)),
		})
		return nil
	})
	sort.Slice(detail.Artifacts, func(i, j int) bool { return detail.Artifacts[i].Name < detail.Artifacts[j].Name })

	return detail, nil
}

// --- runs ---

func (s *Server) handleRunStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.runs.Status())
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	course, lesson, err := s.resolveLesson(req.Course, req.Lesson)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	runID, err := s.runs.Start(course, lesson, req.Stage, req.Force)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"run_id": runID})
}

func (s *Server) handleRunCancel(w http.ResponseWriter, r *http.Request) {
	if !s.runs.Cancel() {
		writeError(w, http.StatusConflict, fmt.Errorf("no run in progress"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "canceling"})
}

// --- artifact serving ---

// downloadName is what the browser should call an artifact once it leaves the
// studio.
//
// On disk every lesson's video is `final.mp4`, and that name is a contract:
// compile concatenates it, the Hugo page embeds it, chapters splits it. What it
// is not is a name anybody wants in their Downloads folder, where six lessons
// arrive as final.mp4, final(1).mp4, final(2).mp4 and the only way to tell them
// apart is to play them.
//
// So the name is rebuilt on the way out, from the **titles** — the words the
// user actually sees in the studio — rather than from the directory names.
// That distinction is the whole point for snippets: a snippet's directory id is
// a slug of its *prompt* (uniqueSnippetID), so the clip titled "Applications of
// Python" lives in a folder called
// "hand-drawn-whiteboard-animation-illustrating-pyt-2". Naming the download
// after the directory would swap one unhelpful filename for another.
//
//	Python Fundamentals 01 The print function.mp4
//	Python Fundamentals 01 The print function - 03 printing multiple items.mp4
//	Applications of Python.mp4        (a snippet: its title is the whole name —
//	                                   the course is always "snippets" and the
//	                                   lesson number does not exist)
func downloadName(courseName, courseSlug, lessonID, lessonTitle, rel string) string {
	ext := filepath.Ext(rel)
	stem := strings.TrimSuffix(filepath.Base(rel), ext)

	title := strings.TrimSpace(lessonTitle)
	if title == "" {
		title = lessonID
	}

	var name string
	if courseSlug == pipeline.SnippetsCourseSlug {
		// A snippet is one clip with one title. Prefixing it with the course
		// would name every file after the one thing they all share.
		name = title
	} else {
		// The lesson number leads, so a folder sorts into teaching order — it
		// is the reason the id is worth reading at all. The course name comes
		// first so two courses' lesson 01 do not collide.
		name = strings.TrimSpace(courseName + " " + lessonNumber(lessonID) + " " + title)
	}

	// `final` is the lesson itself rather than a part of it, so it contributes
	// nothing to the name; anything else does.
	if stem != "final" {
		name += " - " + strings.ReplaceAll(stem, "-", " ")
	}

	return sanitizeFilename(shortenWords(name, 96)) + ext
}

// lessonNumber is the ordering prefix a lesson id carries ("01-what-is-python"
// → "01"), or "" for an id that has none.
func lessonNumber(lessonID string) string {
	head, _, ok := strings.Cut(lessonID, "-")
	if !ok || head == "" {
		return ""
	}
	for _, r := range head {
		if r < '0' || r > '9' {
			return ""
		}
	}
	return head
}

// shortenWords trims a name to `max` characters by dropping whole words from
// the middle.
//
// Cutting at a character offset instead is a line shorter and produces "the
// print function making say things on screen", which does not read as a
// shortened name — it reads as a typo. Words go from the middle because the
// head says which course and lesson this is and the tail says which part of it,
// and those are the two things a filename is for.
func shortenWords(name string, max int) string {
	if len(name) <= max {
		return name
	}
	words := strings.Fields(name)
	// The head that says which course and lesson, and the tail that says which
	// part, are both protected. Four at the tail rather than two because a
	// section name is routinely four words ("03 printing multiple items") and
	// losing half of it is losing the only thing that identifies the file.
	const keepHead, keepTail = 2, 4
	for len(words) > keepHead+keepTail && len(strings.Join(words, " ")) > max {
		at := keepHead + (len(words)-keepHead-keepTail)/2
		words = append(words[:at], words[at+1:]...)
	}
	// Still too long means one word is doing it. Cutting the end of the whole
	// name would take the tail with it, so the offending word is shortened.
	for len(strings.Join(words, " ")) > max {
		longest := 0
		for i, w := range words {
			if len(w) > len(words[longest]) {
				longest = i
			}
		}
		if len(words[longest]) <= 3 {
			break
		}
		over := len(strings.Join(words, " ")) - max
		cut := len(words[longest]) - 3
		if over < cut {
			cut = over
		}
		words[longest] = words[longest][:len(words[longest])-cut]
	}
	return strings.Join(words, " ")
}

// sanitizeFilename keeps a name to characters that are safe in a
// Content-Disposition header and on every filesystem we might land on.
//
// Spaces and capitals survive: the point of naming a download after its title
// is that it reads as the title, and "Applications of Python.mp4" is the whole
// improvement over "applications-of-python.mp4". What does not survive is the
// set Windows rejects outright, plus quotes and control characters — a quote
// reaching the header is injection, not a cosmetic problem, and titles are
// model-written text so they genuinely contain question marks and colons.
func sanitizeFilename(s string) string {
	const illegal = `/\:*?"<>|`
	var b strings.Builder
	for _, r := range s {
		switch {
		case r < 0x20 || r == 0x7f, strings.ContainsRune(illegal, r):
			// Dropped rather than replaced: "language?.mp4" wants to become
			// "language.mp4", not "language-.mp4".
		default:
			b.WriteRune(r)
		}
	}
	// Collapse the runs of spaces that dropping characters leaves behind.
	out := strings.Join(strings.Fields(b.String()), " ")
	// A trailing dot is a filename Windows silently mangles.
	out = strings.Trim(out, " .")
	if out == "" {
		return "artifact"
	}
	return out
}

// artifactMIME maps extensions the default detector gets wrong or generic.
var artifactMIME = map[string]string{
	".vtt":  "text/vtt; charset=utf-8",
	".md":   "text/markdown; charset=utf-8",
	".json": "application/json",
	".svg":  "image/svg+xml",
	".wav":  "audio/wav",
	".mp4":  "video/mp4",
	".txt":  "text/plain; charset=utf-8",
	".py":   "text/plain; charset=utf-8",
}

func (s *Server) handleArtifact(w http.ResponseWriter, r *http.Request) {
	course, lesson, err := s.resolveLesson(r.PathValue("course"), r.PathValue("lesson"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	prefix := "/artifacts/" + course.Slug + "/" + lesson.ID + "/"
	rel := strings.TrimPrefix(r.URL.Path, prefix)
	if rel == "" || rel == r.URL.Path {
		writeError(w, http.StatusNotFound, fmt.Errorf("no artifact path"))
		return
	}
	gen := lesson.GeneratedDir()
	path := filepath.Join(gen, filepath.FromSlash(rel))
	// Path traversal guard: the resolved path must stay inside generated/.
	if cleaned, err := filepath.Rel(gen, path); err != nil || strings.HasPrefix(cleaned, "..") {
		writeError(w, http.StatusForbidden, fmt.Errorf("path escapes the artifact dir"))
		return
	}
	if mime, ok := artifactMIME[strings.ToLower(filepath.Ext(path))]; ok {
		w.Header().Set("Content-Type", mime)
	}
	// `inline`, not `attachment`: the same URL is the <video> element's src and
	// the download link's href, and `attachment` would stop the player from
	// ever showing a frame. Browsers still take the filename from here when the
	// file is saved, and it beats the anchor's own `download` attribute — so
	// this one line names the file on every path out of the studio, including
	// curl and a plain right-click.
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("inline; filename=%q", downloadName(course.Name, course.Slug, lesson.ID, lesson.FrontMatter.Title, rel)))
	http.ServeFile(w, r, path)
}
