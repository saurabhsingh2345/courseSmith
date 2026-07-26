package main

// The studio server's behavior is tested in internal/studio; this file
// only keeps the cmd-level project fixture helper used by other tests.

import (
	"os"
	"path/filepath"
	"testing"
)

// setupProject creates a project root with one course and chdirs into it.
func setupProject(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	lessonDir := filepath.Join(root, "courses", "python-basics", "lessons", "01-what-is-python")
	if err := os.MkdirAll(lessonDir, 0o755); err != nil {
		t.Fatal(err)
	}
	courseYAML := "name: Python Basics\nslug: python-basics\ndescription: Learn Python.\n"
	if err := os.WriteFile(filepath.Join(root, "courses", "python-basics", "course.yaml"), []byte(courseYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	lessonMD := "---\ntitle: What is Python?\n---\n\n## Intro\n- a point\n"
	if err := os.WriteFile(filepath.Join(lessonDir, "lesson.md"), []byte(lessonMD), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
}

func TestResolveCourseFromProjectRoot(t *testing.T) {
	setupProject(t)
	course, err := resolveCourse("python-basics")
	if err != nil {
		t.Fatal(err)
	}
	if course.Slug != "python-basics" {
		t.Errorf("slug = %q", course.Slug)
	}
}
