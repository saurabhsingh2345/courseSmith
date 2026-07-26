package studio

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// draftingRouter returns a fixed valid lesson draft for any content request.
type draftingRouter struct{}

func (draftingRouter) Complete(_ context.Context, _ config.Pipeline, _ llm.TaskType, _ llm.Request) (*llm.Response, error) {
	return &llm.Response{Content: `{
		"title": "What's a list?",
		"summary": "Hold many values in one name.",
		"outcomes": ["Build a list", "Read an item out of it"],
		"diagrams": [{"id":"list-boxes","kind":"mermaid","prompt":"Boxes in a row: index 0, 1, 2."}],
		"sections": [
			{"heading":"Many values, one name","bullets":["A list holds several values","They stay in order"],
			 "code":"xs = [1, 2]\nprint(xs)","output":"[1, 2]","diagram":"list-boxes"},
			{"heading":"What's next","bullets":["Next: looping over a list"]}
		]
	}`}, nil
}

// draftFixture is `fixture` with the outline prompt and a drafting router
// wired up, so POST /api/drafts can actually run.
func draftFixture(t *testing.T) (*Server, string) {
	t.Helper()
	server, root := fixture(t)
	tmpl := `{{define "system"}}draft{{end}}{{define "user"}}{{.Prompt}}{{.Tone}}{{.Audience}}{{.Language}}` +
		`{{.CourseName}}{{.CourseDescription}}{{range .ExistingLessons}}{{.}}{{end}}{{.Critique}}` +
		`{{.MinSections}}{{.MaxSections}}{{.MinCode}}{{.MaxCode}}{{.DiagramCount}}{{end}}`
	if err := os.WriteFile(filepath.Join(root, "prompts", "outline.tmpl"), []byte(tmpl), 0o644); err != nil {
		t.Fatal(err)
	}
	server.env.Router = draftingRouter{}
	return server, root
}

// The whole point of the flow: draft with no course chosen, then file it.
func TestDraftCreateThenAssignToCourse(t *testing.T) {
	server, root := draftFixture(t)
	h := server.Handler()

	var created DraftDetail
	if rec := doJSON(t, h, "POST", "/api/drafts", `{"prompt":"teach python lists"}`, &created); rec.Code != 201 {
		t.Fatalf("create status %d: %s", rec.Code, rec.Body)
	}
	if created.Title != "What's a list?" {
		t.Errorf("title = %q", created.Title)
	}
	if created.Prompt != "teach python lists" {
		t.Errorf("prompt not recorded: %q", created.Prompt)
	}
	if !strings.Contains(created.Source, "[DIAGRAM: list-boxes]") {
		t.Errorf("source missing diagram marker:\n%s", created.Source)
	}
	// Unfiled: it lives in the state dir, not in any course.
	if _, err := os.Stat(filepath.Join(root, ".coursesmith", "drafts", created.ID, "lesson.md")); err != nil {
		t.Fatalf("draft not stored unfiled: %v", err)
	}

	var list []DraftMeta
	if rec := doJSON(t, h, "GET", "/api/drafts", "", &list); rec.Code != 200 {
		t.Fatalf("list status %d", rec.Code)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("list = %+v", list)
	}

	// File it into the existing course — the "which vault" step.
	rec := doJSON(t, h, "POST", "/api/drafts/"+created.ID+"/assign", `{"course":"test-course"}`, nil)
	if rec.Code != 201 {
		t.Fatalf("assign status %d: %s", rec.Code, rec.Body)
	}

	// It is now a real lesson, numbered after the existing one, and gone from
	// the draft store.
	lessonDir := filepath.Join(root, "courses", "test-course", "lessons", "02-what-s-a-list")
	lesson, err := project.LoadLesson(lessonDir)
	if err != nil {
		t.Fatalf("filed lesson does not load: %v", err)
	}
	if lesson.FrontMatter.Title != "What's a list?" {
		t.Errorf("filed title = %q", lesson.FrontMatter.Title)
	}
	if _, err := os.Stat(filepath.Join(lessonDir, draftMetaFileName)); !os.IsNotExist(err) {
		t.Error("draft sidecar was carried into the course")
	}
	if _, err := os.Stat(filepath.Join(root, ".coursesmith", "drafts", created.ID)); !os.IsNotExist(err) {
		t.Error("draft still present after filing")
	}

	var after []DraftMeta
	doJSON(t, h, "GET", "/api/drafts", "", &after)
	if len(after) != 0 {
		t.Errorf("drafts after filing = %+v", after)
	}
}

func TestDraftAssignToNewCourse(t *testing.T) {
	server, root := draftFixture(t)
	h := server.Handler()

	var created DraftDetail
	doJSON(t, h, "POST", "/api/drafts", `{"prompt":"lists"}`, &created)

	rec := doJSON(t, h, "POST", "/api/drafts/"+created.ID+"/assign", `{"new_course":"Python Deep Dive"}`, nil)
	if rec.Code != 201 {
		t.Fatalf("assign status %d: %s", rec.Code, rec.Body)
	}
	course, err := project.LoadCourse(filepath.Join(root, "courses", "python-deep-dive"))
	if err != nil {
		t.Fatalf("new course: %v", err)
	}
	lessons, err := course.Lessons()
	if err != nil {
		t.Fatal(err)
	}
	// The scaffolded placeholder must not survive alongside the real lesson.
	if len(lessons) != 1 {
		t.Fatalf("lessons = %d, want just the filed draft", len(lessons))
	}
	if lessons[0].FrontMatter.Title != "What's a list?" {
		t.Errorf("lesson title = %q", lessons[0].FrontMatter.Title)
	}
}

func TestDraftAssignRejectsAmbiguousTarget(t *testing.T) {
	server, _ := draftFixture(t)
	h := server.Handler()

	var created DraftDetail
	doJSON(t, h, "POST", "/api/drafts", `{"prompt":"lists"}`, &created)

	for _, body := range []string{`{}`, `{"course":"test-course","new_course":"Other"}`} {
		if rec := doJSON(t, h, "POST", "/api/drafts/"+created.ID+"/assign", body, nil); rec.Code != 400 {
			t.Errorf("assign %s = %d, want 400", body, rec.Code)
		}
	}
}

func TestDraftUpdateRejectsInvalidSource(t *testing.T) {
	server, root := draftFixture(t)
	h := server.Handler()

	var created DraftDetail
	doJSON(t, h, "POST", "/api/drafts", `{"prompt":"lists"}`, &created)

	body, _ := json.Marshal(UpdateDraftRequest{Source: "no front matter here"})
	if rec := doJSON(t, h, "PUT", "/api/drafts/"+created.ID, string(body), nil); rec.Code != 400 {
		t.Errorf("update status = %d, want 400", rec.Code)
	}
	// The bad write must not have landed.
	raw, err := os.ReadFile(filepath.Join(root, ".coursesmith", "drafts", created.ID, "lesson.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "title:") {
		t.Errorf("draft was overwritten by invalid source:\n%s", raw)
	}
}

func TestDraftDelete(t *testing.T) {
	server, root := draftFixture(t)
	h := server.Handler()

	var created DraftDetail
	doJSON(t, h, "POST", "/api/drafts", `{"prompt":"lists"}`, &created)

	if rec := doJSON(t, h, "DELETE", "/api/drafts/"+created.ID, "", nil); rec.Code != 204 {
		t.Fatalf("delete status %d", rec.Code)
	}
	if _, err := os.Stat(filepath.Join(root, ".coursesmith", "drafts", created.ID)); !os.IsNotExist(err) {
		t.Error("draft survived delete")
	}
}

// A draft id from the URL must never escape the drafts directory. Plain "../"
// is normalized away by the mux (307) before the handler sees it; the encoded
// forms reach draftDir and must be rejected there.
func TestDraftIDIsNotAPath(t *testing.T) {
	server, _ := draftFixture(t)
	h := server.Handler()

	for _, id := range []string{"..", "%2e%2e", "%2e%2e%2f%2e%2e", "nope", "..%2fcourses"} {
		rec := doJSON(t, h, "GET", "/api/drafts/"+id, "", nil)
		if rec.Code == 200 {
			t.Errorf("GET /api/drafts/%s served content: %s", id, rec.Body)
		}
	}
	// Deleting through a traversal must not remove anything either.
	if rec := doJSON(t, h, "DELETE", "/api/drafts/%2e%2e", "", nil); rec.Code == 204 {
		t.Error("DELETE via traversal reported success")
	}
}
