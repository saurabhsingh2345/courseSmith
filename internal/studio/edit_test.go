package studio

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/enfec/coursesmith/internal/project"
)

func TestCourseUpdatePersistsMetadata(t *testing.T) {
	server, root := fixture(t)
	h := server.Handler()

	var detail CourseDetail
	rec := doJSON(t, h, "PUT", "/api/courses/test-course",
		`{"name":"Renamed Course","description":"New desc","archetype":"concept-first","animation_style":"smooth","color_palette":"cool","colors":{"primary":"#123456"}}`,
		&detail)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if detail.Name != "Renamed Course" || detail.Description != "New desc" {
		t.Errorf("detail = %+v", detail)
	}

	// The edit must land on disk and round-trip through the loader.
	c, err := project.LoadCourse(filepath.Join(root, "courses", "test-course"))
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if c.Config.Style.Archetype != "concept-first" || c.Config.Style.AnimationStyle != "smooth" {
		t.Errorf("style not persisted: %+v", c.Config.Style)
	}
	if c.Config.Style.ColorPalette != "cool" {
		t.Errorf("palette not persisted: %q", c.Config.Style.ColorPalette)
	}
	if c.Config.Branding.Colors.Primary != "#123456" {
		t.Errorf("color not persisted: %q", c.Config.Branding.Colors.Primary)
	}
	// slug is preserved (not in the patch).
	if c.Slug != "test-course" {
		t.Errorf("slug changed to %q", c.Slug)
	}
}

func TestCourseUpdateRejectsUnknownArchetype(t *testing.T) {
	server, root := fixture(t)
	h := server.Handler()

	rec := doJSON(t, h, "PUT", "/api/courses/test-course", `{"archetype":"nonsense"}`, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body)
	}
	// A rejected edit must not have touched course.yaml.
	c, err := project.LoadCourse(filepath.Join(root, "courses", "test-course"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Config.Style.Archetype != "" {
		t.Errorf("rejected edit leaked to disk: %q", c.Config.Style.Archetype)
	}
}

func TestCourseDelete(t *testing.T) {
	server, root := fixture(t)
	h := server.Handler()

	if rec := doJSON(t, h, "DELETE", "/api/courses/test-course", "", nil); rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", rec.Code, rec.Body)
	}
	if _, err := os.Stat(filepath.Join(root, "courses", "test-course")); !os.IsNotExist(err) {
		t.Errorf("course dir still exists: %v", err)
	}
	if rec := doJSON(t, h, "GET", "/api/courses/test-course", "", nil); rec.Code != http.StatusNotFound {
		t.Errorf("get after delete = %d, want 404", rec.Code)
	}
}

func TestLessonCreateAndUpdate(t *testing.T) {
	server, root := fixture(t)
	h := server.Handler()

	var created LessonDetail
	rec := doJSON(t, h, "POST", "/api/lessons/test-course", `{"title":"Second Lesson"}`, &created)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status %d: %s", rec.Code, rec.Body)
	}
	if created.ID != "02-second-lesson" || created.Title != "Second Lesson" {
		t.Errorf("created = %+v", created)
	}
	if _, err := os.Stat(filepath.Join(root, "courses", "test-course", "lessons", "02-second-lesson", "lesson.md")); err != nil {
		t.Errorf("lesson.md not written: %v", err)
	}

	// Update its source.
	var updated LessonDetail
	newSrc := "---\ntitle: Second Lesson (edited)\n---\n\n## New body\n- edited\n"
	rec = doJSON(t, h, "PUT", "/api/lessons/test-course/02", `{"source":`+jsonString(newSrc)+`}`, &updated)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status %d: %s", rec.Code, rec.Body)
	}
	if updated.Title != "Second Lesson (edited)" {
		t.Errorf("title not updated: %q", updated.Title)
	}
}

func TestLessonUpdateRejectsInvalidSource(t *testing.T) {
	server, _ := fixture(t)
	h := server.Handler()

	// Missing front-matter title -> LoadLesson fails -> 400.
	rec := doJSON(t, h, "PUT", "/api/lessons/test-course/01", `{"source":"no front matter here"}`, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body)
	}
	// Original source must be untouched.
	var detail LessonDetail
	doJSON(t, h, "GET", "/api/lessons/test-course/01", "", &detail)
	if detail.Title != "Test Lesson" {
		t.Errorf("original lesson clobbered: %q", detail.Title)
	}
}

func TestArchetypesCatalog(t *testing.T) {
	server, _ := fixture(t)
	h := server.Handler()

	var out struct {
		Archetypes []struct {
			Name             string `json:"name"`
			DefaultAnimation string `json:"default_animation"`
		} `json:"archetypes"`
		AnimationStyles []string `json:"animation_styles"`
		Palettes        []struct {
			Name string `json:"name"`
		} `json:"palettes"`
	}
	if rec := doJSON(t, h, "GET", "/api/archetypes", "", &out); rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if len(out.Archetypes) != 5 {
		t.Errorf("archetypes = %d, want 5", len(out.Archetypes))
	}
	if len(out.AnimationStyles) != 3 {
		t.Errorf("animation styles = %v", out.AnimationStyles)
	}
	if len(out.Palettes) != 4 {
		t.Errorf("palettes = %d, want 4", len(out.Palettes))
	}
}

func TestLibraryDiagramCRUD(t *testing.T) {
	server, _ := fixture(t)
	h := server.Handler()

	var created LibraryDiagram
	rec := doJSON(t, h, "POST", "/api/library/diagrams", `{"name":"Flow","kind":"mermaid","source":"graph TD; A-->B"}`, &created)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status %d: %s", rec.Code, rec.Body)
	}
	if created.ID == "" || created.Name != "Flow" || created.CreatedAt == "" {
		t.Errorf("created = %+v", created)
	}

	var list []LibraryDiagram
	doJSON(t, h, "GET", "/api/library/diagrams", "", &list)
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("list = %+v", list)
	}

	if rec := doJSON(t, h, "DELETE", "/api/library/diagrams/"+created.ID, "", nil); rec.Code != http.StatusNoContent {
		t.Fatalf("delete status %d: %s", rec.Code, rec.Body)
	}
	doJSON(t, h, "GET", "/api/library/diagrams", "", &list)
	if len(list) != 0 {
		t.Errorf("list after delete = %+v", list)
	}
	if rec := doJSON(t, h, "DELETE", "/api/library/diagrams/missing", "", nil); rec.Code != http.StatusNotFound {
		t.Errorf("delete missing = %d, want 404", rec.Code)
	}
}

func TestLibraryQuestionCRUD(t *testing.T) {
	server, _ := fixture(t)
	h := server.Handler()

	var created LibraryQuestion
	rec := doJSON(t, h, "POST", "/api/library/questions",
		`{"prompt":"What is 2+2?","type":"recall","options":["3","4"],"answer_index":1,"explanation":"basic"}`, &created)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status %d: %s", rec.Code, rec.Body)
	}
	if created.ID == "" || created.Prompt != "What is 2+2?" {
		t.Errorf("created = %+v", created)
	}

	// Missing prompt -> 400.
	if rec := doJSON(t, h, "POST", "/api/library/questions", `{"type":"recall"}`, nil); rec.Code != http.StatusBadRequest {
		t.Errorf("empty-prompt status = %d, want 400", rec.Code)
	}
}

// jsonString quotes a Go string as a JSON string literal for inline bodies.
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
