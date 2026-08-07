package studio

// Deleting what the pipeline is currently working on used to be allowed, and
// the failure it produced named none of its causes. The run holds a lesson
// handle pointing at a directory that has stopped existing; the stage in
// flight finishes its LLM call, writes its artifact into a generated/ dir that
// os.MkdirAll silently recreates, and the runner then fails hashing lesson.md
// to record the result:
//
//	lesson <id>, stage substance: hashing …/lesson.md: no such file or directory
//
// What is left on disk is a lesson directory holding one generated artifact and
// no source — invisible to the snippet list (which needs snippet.yaml) and
// mistakable for a pipeline bug. The delete handlers refuse instead.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/pipeline"
)

// markRunning puts the run manager into the state it holds mid-stage, without
// running a pipeline: the guard reads exactly these fields, and driving a real
// run to a chosen stage and pausing it there would test the scheduler rather
// than the guard.
func markRunning(s *Server, course, lesson, stage string) {
	s.runs.mu.Lock()
	defer s.runs.mu.Unlock()
	s.runs.running = true
	s.runs.status = RunStatus{Running: true, RunID: "run-1", Course: course, Lesson: lesson, Stage: stage}
}

func TestSnippetDeleteRefusedWhileItsRunIsInFlight(t *testing.T) {
	server, root := snippetFixture(t)
	h := server.Handler()

	var created CreateSnippetResponse
	if rec := doJSON(t, h, "POST", "/api/snippets",
		`{"prompt":"how for loops work in python","template":"vscode","plan_only":true}`, &created); rec.Code != 201 {
		t.Fatalf("create status %d: %s", rec.Code, rec.Body)
	}
	waitForRun(t, server)
	dir := filepath.Join(root, ".coursesmith", "snippets", "lessons", created.ID)

	markRunning(server, pipeline.SnippetsCourseSlug, created.ID, "substance")

	rec := doJSON(t, h, "DELETE", "/api/snippets/"+created.ID, "", nil)
	if rec.Code != 409 {
		t.Fatalf("delete status %d, want 409 conflict: %s", rec.Code, rec.Body)
	}
	// The message has to name the stage and the way out, because the person
	// reading it is looking at a delete button that did nothing.
	for _, want := range []string{"substance", "/api/run"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("refusal does not mention %q: %s", want, rec.Body)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, pipeline.SnippetFileName)); err != nil {
		t.Fatalf("the refused delete removed the snippet anyway: %v", err)
	}

	// A snippet the run has moved off is deletable as before — the guard is
	// about the one lesson in flight, not about the run existing at all.
	markRunning(server, pipeline.SnippetsCourseSlug, "some-other-snippet", "audio")
	if rec := doJSON(t, h, "DELETE", "/api/snippets/"+created.ID, "", nil); rec.Code != 204 {
		t.Fatalf("delete status %d, want 204: %s", rec.Code, rec.Body)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("delete left the snippet directory behind")
	}
}

func TestLessonAndCourseDeleteRefusedWhileRunning(t *testing.T) {
	server, _ := fixture(t)
	h := server.Handler()

	markRunning(server, "test-course", "01-test", "script")

	if rec := doJSON(t, h, "DELETE", "/api/lessons/test-course/01-test", "", nil); rec.Code != 409 {
		t.Fatalf("lesson delete status %d, want 409: %s", rec.Code, rec.Body)
	}
	// The course delete guards on the run being anywhere inside it: removing
	// the course removes the running lesson just as surely.
	rec := doJSON(t, h, "DELETE", "/api/courses/test-course", "", nil)
	if rec.Code != 409 {
		t.Fatalf("course delete status %d, want 409: %s", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), "01-test") {
		t.Errorf("the course refusal should name the lesson being worked on: %s", rec.Body)
	}
}
