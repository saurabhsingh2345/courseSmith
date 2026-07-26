package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/project"
)

func TestBuildConceptGraph(t *testing.T) {
	order := []string{"01-a", "02-b", "03-c"}
	perLesson := map[string]*LessonConcepts{
		"01-a": {
			LessonID:   "01-a",
			Introduced: []ConceptRef{{Name: "variable", Section: "s1"}},
			Used:       []ConceptRef{{Name: "function", Section: "s2", Quote: "call the function"}},
		},
		"02-b": {
			LessonID:   "02-b",
			Introduced: []ConceptRef{{Name: "function", Section: "s1"}},
			Used:       []ConceptRef{{Name: "variable", Section: "s1"}},
		},
		"03-c": {
			LessonID: "03-c",
			Used: []ConceptRef{
				{Name: "variable", Section: "s1"},
				{Name: "loop", Section: "s2", Quote: "loop over items"},
			},
		},
	}

	nodes, violations := BuildConceptGraph(order, perLesson)

	byName := map[string]ConceptNode{}
	for _, n := range nodes {
		byName[n.Name] = n
	}
	if got := byName["variable"]; got.IntroducedIn != "01-a" || strings.Join(got.RequiredBy, ",") != "02-b,03-c" {
		t.Errorf("variable node = %+v", got)
	}
	if got := byName["function"]; got.IntroducedIn != "02-b" {
		t.Errorf("function node = %+v", got)
	}

	if len(violations) != 2 {
		t.Fatalf("violations = %+v, want 2", violations)
	}
	// "function" used in 01-a but introduced in 02-b.
	if violations[0].Concept != "function" || violations[0].UsedIn != "01-a" || violations[0].IntroducedIn != "02-b" {
		t.Errorf("violation[0] = %+v", violations[0])
	}
	// "loop" used in 03-c, never introduced.
	if violations[1].Concept != "loop" || violations[1].UsedIn != "03-c" || violations[1].IntroducedIn != "" {
		t.Errorf("violation[1] = %+v", violations[1])
	}
	if !strings.Contains(violations[1].String(), "never introduced") {
		t.Errorf("violation string = %q", violations[1].String())
	}
}

// multiLessonCourse builds a course with n minimal lessons on disk.
func multiLessonCourse(t *testing.T, n int) *project.Course {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "test-course")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, project.CourseFileName), []byte(testCourseYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= n; i++ {
		lessonDir := filepath.Join(dir, "lessons", fmt.Sprintf("%02d-lesson", i))
		if err := os.MkdirAll(lessonDir, 0o755); err != nil {
			t.Fatal(err)
		}
		md := fmt.Sprintf("---\ntitle: Lesson %d\n---\n\n## Topic %d\n- a point\n", i, i)
		if err := os.WriteFile(filepath.Join(lessonDir, project.LessonFileName), []byte(md), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	course, err := project.LoadCourse(dir)
	if err != nil {
		t.Fatal(err)
	}
	return course
}

func conceptsResponse(introduced, used string) string {
	intro := "[]"
	if introduced != "" {
		intro = fmt.Sprintf(`[{"name":%q,"term":%q,"section":"topic","quote":"..."}]`, introduced, introduced)
	}
	uses := "[]"
	if used != "" {
		uses = fmt.Sprintf(`[{"name":%q,"term":%q,"section":"topic","quote":"..."}]`, used, used)
	}
	return fmt.Sprintf(`{"introduced":%s,"used":%s}`, intro, uses)
}

func TestAnalyzeCourse(t *testing.T) {
	course := multiLessonCourse(t, 2)
	fake := &fakeRouter{
		content: []string{
			conceptsResponse("variable", ""),         // lesson 01
			conceptsResponse("function", "variable"), // lesson 02
		},
		review: []string{
			`{"issues":[]}`,                       // terminology
			`{"score":8.5,"suggestion":"Works."}`, // bridge 01→02
		},
	}
	env, out := runEnv(t, fake)

	report, err := AnalyzeCourse(context.Background(), env, course)
	if err != nil {
		t.Fatalf("AnalyzeCourse: %v", err)
	}
	if len(report.Concepts) != 2 || len(report.Violations) != 0 {
		t.Errorf("report = %+v", report)
	}
	if len(report.Bridges) != 1 || report.Bridges[0].Score != 8.5 {
		t.Errorf("bridges = %+v", report.Bridges)
	}

	outDir := filepath.Join(course.Dir, CourseGeneratedDirName)
	for _, name := range []string{ConceptsFileName, ConceptsSVGFileName, AnalysisFileName} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Errorf("missing artifact %s: %v", name, err)
		}
	}
	svg, err := os.ReadFile(filepath.Join(outDir, ConceptsSVGFileName))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"01-lesson", "02-lesson", "variable", "function"} {
		if !strings.Contains(string(svg), want) {
			t.Errorf("concepts.svg missing %q", want)
		}
	}
	if !strings.Contains(out.String(), "zero dependency violations") {
		t.Errorf("output missing clean verdict:\n%s", out.String())
	}

	// Second run: extractions are cached (no new content requests), but the
	// terminology and bridge checks re-run.
	fake.content = nil
	fake.review = []string{`{"issues":[]}`, `{"score":8.5,"suggestion":"Works."}`}
	if _, err := AnalyzeCourse(context.Background(), env, course); err != nil {
		t.Fatalf("cached re-run: %v", err)
	}
}

func TestAnalyzeCourseFlagsViolations(t *testing.T) {
	course := multiLessonCourse(t, 2)
	fake := &fakeRouter{
		content: []string{
			conceptsResponse("variable", "loop"), // lesson 01 uses "loop" — taught in 02
			conceptsResponse("loop", ""),         // lesson 02
		},
		review: []string{
			`{"issues":[{"concept":"loop","variants":["loop","cycle"],"canonical":"loop","reason":"drift"}]}`,
			`{"score":5,"suggestion":"Open lesson 2 by recalling variables."}`,
		},
	}
	env, out := runEnv(t, fake)

	report, err := AnalyzeCourse(context.Background(), env, course)
	if err == nil || !strings.Contains(err.Error(), "dependency violation") {
		t.Fatalf("error = %v, want dependency violation", err)
	}
	if len(report.Violations) != 1 || report.Violations[0].Concept != "loop" || report.Violations[0].IntroducedIn != "02-lesson" {
		t.Errorf("violations = %+v", report.Violations)
	}
	if !strings.Contains(out.String(), "VIOLATION") || !strings.Contains(out.String(), "terminology drift") {
		t.Errorf("output:\n%s", out.String())
	}
	// A weak bridge (< 7) surfaces its suggestion.
	if !strings.Contains(out.String(), "recalling variables") {
		t.Errorf("weak bridge suggestion missing:\n%s", out.String())
	}

	// Artifacts still written despite the violation error.
	var payload struct {
		Violations []DependencyViolation `json:"violations"`
	}
	data, err := os.ReadFile(filepath.Join(course.Dir, CourseGeneratedDirName, ConceptsFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Violations) != 1 {
		t.Errorf("persisted violations = %+v", payload.Violations)
	}
}
