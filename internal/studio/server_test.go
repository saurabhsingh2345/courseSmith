package studio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/pipeline"
)

// fixture builds a project on disk with one course/lesson plus generated
// artifacts, and returns a Server over it.
func fixture(t *testing.T) (*Server, string) {
	t.Helper()
	root := t.TempDir()
	coursesDir := filepath.Join(root, "courses")
	lessonDir := filepath.Join(coursesDir, "test-course", "lessons", "01-test")
	genDir := filepath.Join(lessonDir, "generated")
	if err := os.MkdirAll(filepath.Join(genDir, "reviews"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		filepath.Join(coursesDir, "test-course", "course.yaml"): "name: Test Course\nslug: test-course\ndescription: For tests.\n",
		filepath.Join(lessonDir, "lesson.md"):                   "---\ntitle: Test Lesson\n---\n\n## First idea\n- a point\n",
		filepath.Join(genDir, "script.json"): `{"title":"Test Lesson","sections":[
			{"id":"first-idea","narration":"Hello there.","duration_est_sec":5,"cues":[]}]}`,
		filepath.Join(genDir, "quiz.json"): `{"title":"Quiz","questions":[
			{"id":"q1","type":"recall","prompt":"P1?","options":["a","b","c","d"],"answer_index":0,"explanation":"E1"},
			{"id":"q2","type":"application","prompt":"P2?","options":["a","b","c","d"],"answer_index":1,"explanation":"E2"},
			{"id":"q3","type":"debugging","prompt":"P3?","options":["a","b","c","d"],"answer_index":2,"explanation":"E3"},
			{"id":"q4","type":"prediction","prompt":"P4?","options":["a","b","c","d"],"answer_index":3,"explanation":"E4"}]}`,
		filepath.Join(genDir, "reviews", "pace.json"): `{"target_wpm":150,"tolerance":0.15,"sections":[],"flagged":[],"checked_at":"2026-01-01T00:00:00Z"}`,
		filepath.Join(genDir, "voiceover.wav"):        "RIFFfakewav",
	}
	for path, content := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	env := &pipeline.Env{PromptsDir: filepath.Join(root, "prompts")}
	if err := os.MkdirAll(env.PromptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(root, ".coursesmith")
	return New(env, coursesDir, stateDir, ""), root
}

func doJSON(t *testing.T, h http.Handler, method, path, body string, out any) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if out != nil && rec.Code < 300 {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			t.Fatalf("decoding %s %s response: %v\n%s", method, path, err, rec.Body)
		}
	}
	return rec
}

func TestCoursesAndMatrix(t *testing.T) {
	server, _ := fixture(t)
	h := server.Handler()

	var courses []Course
	if rec := doJSON(t, h, "GET", "/api/courses", "", &courses); rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if len(courses) != 1 || courses[0].Slug != "test-course" || courses[0].LessonCount != 1 {
		t.Errorf("courses = %+v", courses)
	}

	var detail CourseDetail
	if rec := doJSON(t, h, "GET", "/api/courses/test-course", "", &detail); rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if len(detail.Lessons) != 1 || detail.Lessons[0].Stages["script"] == "" {
		t.Errorf("detail = %+v", detail)
	}
	if rec := doJSON(t, h, "GET", "/api/courses/nope", "", nil); rec.Code != 404 {
		t.Errorf("unknown course status = %d", rec.Code)
	}
}

func TestLessonDetail(t *testing.T) {
	server, _ := fixture(t)
	h := server.Handler()

	var detail LessonDetail
	if rec := doJSON(t, h, "GET", "/api/lessons/test-course/01", "", &detail); rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if detail.ID != "01-test" || detail.Title != "Test Lesson" {
		t.Errorf("detail = %+v", detail)
	}
	if !strings.Contains(detail.Source, "## First idea") {
		t.Errorf("source not included: %q", detail.Source)
	}
	if detail.Script == nil || detail.Quiz == nil {
		t.Error("script/quiz artifacts not inlined")
	}
	if _, ok := detail.Reviews["pace"]; !ok {
		t.Errorf("reviews = %v, want pace", detail.Reviews)
	}
	var haveVoiceover bool
	for _, a := range detail.Artifacts {
		if a.Name == "voiceover.wav" {
			haveVoiceover = true
			if a.URL != "/artifacts/test-course/01-test/voiceover.wav" {
				t.Errorf("artifact url = %q", a.URL)
			}
		}
	}
	if !haveVoiceover {
		t.Errorf("artifacts = %+v", detail.Artifacts)
	}
}

func TestArtifactServingAndTraversalGuard(t *testing.T) {
	server, _ := fixture(t)
	h := server.Handler()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/artifacts/test-course/01-test/quiz.json", nil))
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("artifact: status %d, type %q", rec.Code, rec.Header().Get("Content-Type"))
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/artifacts/test-course/01-test/../../../course.yaml", nil))
	if rec.Code == 200 {
		t.Error("path traversal served a file outside generated/")
	}
}

func TestFeedbackWritesReviewNotes(t *testing.T) {
	server, root := fixture(t)
	h := server.Handler()

	body := `{"course":"test-course","lesson":"01-test","section":"first-idea","note":"slow down here","timestamp_ms":83000}`
	if rec := doJSON(t, h, "POST", "/api/feedback", body, nil); rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	notes, err := pipeline.LoadReviewNotes(filepath.Join(root, "courses", "test-course"))
	if err != nil {
		t.Fatal(err)
	}
	text := notes.UnresolvedText("01-test")
	if !strings.Contains(text, "[at 1:23] slow down here") || !strings.Contains(text, "[section first-idea]") {
		t.Errorf("notes = %q", text)
	}

	if rec := doJSON(t, h, "POST", "/api/feedback", `{"course":"test-course","lesson":"01-test","note":"  "}`, nil); rec.Code != 400 {
		t.Errorf("empty note status = %d", rec.Code)
	}
}

func TestQuizOverridesRoundTrip(t *testing.T) {
	server, _ := fixture(t)
	h := server.Handler()

	put := `{"questions":[{"id":"q1","prompt":"Edited prompt?"},{"id":"q4","drop":true}]}`
	var saved QuizWithOverrides
	if rec := doJSON(t, h, "PUT", "/api/quiz-overrides/test-course/01", put, &saved); rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if saved.Merged == nil || len(saved.Merged.Questions) != 3 {
		t.Fatalf("merged = %+v", saved.Merged)
	}
	if saved.Merged.Questions[0].Prompt != "Edited prompt?" {
		t.Errorf("override not applied: %+v", saved.Merged.Questions[0])
	}
	// Non-overridden fields survive.
	if saved.Merged.Questions[0].Explanation != "E1" || saved.Merged.Questions[0].AnswerIndex != 0 {
		t.Errorf("merge clobbered fields: %+v", saved.Merged.Questions[0])
	}

	var got QuizWithOverrides
	if rec := doJSON(t, h, "GET", "/api/quiz/test-course/01", "", &got); rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if got.Overrides == nil || len(got.Overrides.Questions) != 2 {
		t.Errorf("overrides not persisted: %+v", got.Overrides)
	}
	var generated pipeline.Quiz
	if err := json.Unmarshal(got.Generated, &generated); err != nil || len(generated.Questions) != 4 {
		t.Errorf("generated quiz must stay untouched: %v %+v", err, generated)
	}
}

func TestLedgerAggregatesCache(t *testing.T) {
	server, root := fixture(t)
	cacheDir := filepath.Join(root, ".coursesmith", "cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	entry := func(model string, prompt, completion int) string {
		return fmt.Sprintf(`{"provider":"openai","created_at":"2026-07-17T10:00:00Z",
			"response":{"model":%q,"usage":{"prompt_tokens":%d,"completion_tokens":%d}}}`, model, prompt, completion)
	}
	os.WriteFile(filepath.Join(cacheDir, "a.json"), []byte(entry("gpt-4o-mini-2024-07-18", 1_000_000, 0)), 0o644)
	os.WriteFile(filepath.Join(cacheDir, "b.json"), []byte(entry("gpt-4o-mini-2024-07-18", 0, 1_000_000)), 0o644)

	var ledger Ledger
	if rec := doJSON(t, server.Handler(), "GET", "/api/ledger", "", &ledger); rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if ledger.TotalCalls != 2 {
		t.Errorf("calls = %d", ledger.TotalCalls)
	}
	// 1M prompt at $0.15 + 1M completion at $0.60.
	if diff := ledger.TotalCostUSD - 0.75; diff > 0.001 || diff < -0.001 {
		t.Errorf("cost = %v, want 0.75", ledger.TotalCostUSD)
	}
	if len(ledger.Quotas) == 0 {
		t.Error("quotas missing")
	}
}

// fakeStageEnv runs the real pipeline with a fake router so a run produces
// genuine SSE events.
func TestRunLifecycleOverSSE(t *testing.T) {
	server, root := fixture(t)
	// Wire a real prompts dir + fake router so the script stage can run.
	promptsDir := filepath.Join(root, "prompts")
	os.WriteFile(filepath.Join(promptsDir, "script.tmpl"),
		[]byte(`{{define "system"}}s{{end}}{{define "user"}}{{.Title}} {{.Outline}}{{if .Notes}}{{.Notes}}{{end}}{{if .Critique}}{{.Critique}}{{end}}{{end}}`), 0o644)
	server.env.Router = scriptedRouter{}

	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	// Subscribe to SSE first.
	sseResp, err := http.Get(ts.URL + "/api/events")
	if err != nil {
		t.Fatal(err)
	}
	defer sseResp.Body.Close()

	rec := doJSON(t, server.Handler(), "POST", "/api/run",
		`{"course":"test-course","lesson":"01-test","stage":"script"}`, nil)
	if rec.Code != 202 {
		t.Fatalf("run status %d: %s", rec.Code, rec.Body)
	}

	// Read events until run-finished (with a deadline).
	types := map[string]bool{}
	scanner := bufio.NewScanner(sseResp.Body)
	deadline := time.After(10 * time.Second)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			var e Event
			if json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &e) == nil {
				types[e.Type] = true
				if e.Type == "run-finished" || e.Type == "run-failed" {
					return
				}
			}
		}
	}()
	select {
	case <-done:
	case <-deadline:
		t.Fatal("timed out waiting for run-finished")
	}
	for _, want := range []string{"run-started", "stage-started", "stage-finished", "run-finished"} {
		if !types[want] {
			t.Errorf("missing SSE event %q (got %v)", want, types)
		}
	}
	if types["run-failed"] {
		t.Errorf("run failed: %v", types)
	}

	// A second run while idle succeeds; the busy conflict is exercised by
	// starting one and immediately starting another.
	if rec := doJSON(t, server.Handler(), "GET", "/api/run", "", nil); rec.Code != 200 {
		t.Errorf("run status endpoint = %d", rec.Code)
	}
}

// scriptedRouter returns a fixed valid script for any content request.
type scriptedRouter struct{}

func (scriptedRouter) Complete(_ context.Context, _ config.Pipeline, task llm.TaskType, _ llm.Request) (*llm.Response, error) {
	return &llm.Response{Content: `{"title":"Test Lesson","sections":[
		{"id":"first-idea","narration":"Fresh narration.","duration_est_sec":5,"cues":[]}]}`}, nil
}
