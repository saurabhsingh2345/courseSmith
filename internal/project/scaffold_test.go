package project

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldCourseCreatesAndValidates(t *testing.T) {
	dir := t.TempDir()
	res, err := ScaffoldCourse(dir, "Python Basics!", "A friendly intro.")
	if err != nil {
		t.Fatalf("ScaffoldCourse: %v", err)
	}
	if res.Slug != "python-basics" {
		t.Errorf("slug = %q, want python-basics", res.Slug)
	}
	for _, f := range []string{res.CourseFile, res.LessonFile} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected file %s: %v", f, err)
		}
	}
	// The custom description is embedded (not the placeholder).
	yaml, err := os.ReadFile(res.CourseFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(yaml), "A friendly intro.") {
		t.Errorf("course.yaml missing custom description:\n%s", yaml)
	}
	// It parses as a real course.
	if _, err := LoadCourse(res.CourseDir); err != nil {
		t.Errorf("LoadCourse on scaffold: %v", err)
	}
}

func TestScaffoldCourseErrors(t *testing.T) {
	dir := t.TempDir()

	if _, err := ScaffoldCourse(dir, "!!!", ""); !errors.Is(err, ErrEmptySlug) {
		t.Errorf("empty slug: got %v, want ErrEmptySlug", err)
	}

	if _, err := ScaffoldCourse(dir, "Dupe Course", ""); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if _, err := ScaffoldCourse(dir, "Dupe Course", ""); !errors.Is(err, ErrCourseExists) {
		t.Errorf("duplicate: got %v, want ErrCourseExists", err)
	}
}

func TestScaffoldCourseDefaultDescription(t *testing.T) {
	dir := t.TempDir()
	res, err := ScaffoldCourse(filepath.Join(dir, "courses"), "No Desc", "   ")
	if err != nil {
		t.Fatalf("ScaffoldCourse: %v", err)
	}
	yaml, err := os.ReadFile(res.CourseFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(yaml), defaultDescription) {
		t.Errorf("blank description should fall back to placeholder:\n%s", yaml)
	}
}
