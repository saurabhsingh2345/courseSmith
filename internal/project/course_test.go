package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeCourse creates a course dir under t.TempDir() with the given
// course.yaml content and optional lesson files (path relative to the
// course dir → content).
func writeCourse(t *testing.T, courseYAML string, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, CourseFileName), []byte(courseYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	for rel, content := range files {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

const validCourseYAML = `
name: "Python Basics"
slug: python-basics
description: An intro course.
style:
  tone: warm teacher
  pace_wpm: 145
pipeline:
  llm_content: groq/llama-3.3-70b-versatile
  review_threshold: 8
`

const validLessonMD = `---
title: What is Python?
diagrams:
  - id: memory-model
    prompt: "3 variables in memory"
---

# What is Python?
- Some outline content.
`

func TestLoadCourse(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string // substring of the expected error; empty means success
	}{
		{
			name: "valid course",
			yaml: validCourseYAML,
		},
		{
			name:    "missing name",
			yaml:    "slug: python-basics\n",
			wantErr: "name is required",
		},
		{
			name:    "missing slug",
			yaml:    "name: Python Basics\n",
			wantErr: "slug is required",
		},
		{
			name:    "bad slug",
			yaml:    "name: X\nslug: Python_Basics\n",
			wantErr: "slug",
		},
		{
			name:    "unknown field rejected",
			yaml:    "name: X\nslug: x\nstyl:\n  tone: oops\n",
			wantErr: "field styl not found",
		},
		{
			name:    "pace out of range",
			yaml:    "name: X\nslug: x\nstyle:\n  pace_wpm: 9000\n",
			wantErr: "pace_wpm",
		},
		{
			name:    "threshold out of range",
			yaml:    "name: X\nslug: x\npipeline:\n  review_threshold: 11\n",
			wantErr: "review_threshold",
		},
		{
			name:    "malformed yaml",
			yaml:    "name: [unclosed\n",
			wantErr: "parsing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeCourse(t, tt.yaml, nil)
			c, err := LoadCourse(dir)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("LoadCourse() error = %v, want success", err)
				}
				if c.Name != "Python Basics" || c.Slug != "python-basics" {
					t.Errorf("parsed course = %+v", c)
				}
				if c.Config.Style.Tone != "warm teacher" {
					t.Errorf("inline config not parsed: Style = %+v", c.Config.Style)
				}
				if c.Config.Pipeline.ReviewThreshold != 8 {
					t.Errorf("ReviewThreshold = %v, want 8", c.Config.Pipeline.ReviewThreshold)
				}
				return
			}
			if err == nil {
				t.Fatalf("LoadCourse() succeeded, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadCourseMissingFile(t *testing.T) {
	_, err := LoadCourse(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), CourseFileName) {
		t.Errorf("error = %v, want mention of %s", err, CourseFileName)
	}
}

func TestLessonsSortedAndFiltered(t *testing.T) {
	dir := writeCourse(t, validCourseYAML, map[string]string{
		"lessons/02-variables/lesson.md":      strings.Replace(validLessonMD, "What is Python?", "Variables", 2),
		"lessons/01-what-is-python/lesson.md": validLessonMD,
		"lessons/notes.txt":                   "not a lesson",
		"lessons/drafts/README.md":            "dir without lesson.md is skipped",
	})
	c, err := LoadCourse(dir)
	if err != nil {
		t.Fatal(err)
	}
	lessons, err := c.Lessons()
	if err != nil {
		t.Fatal(err)
	}
	if len(lessons) != 2 {
		t.Fatalf("got %d lessons, want 2", len(lessons))
	}
	if lessons[0].ID != "01-what-is-python" || lessons[1].ID != "02-variables" {
		t.Errorf("lesson order = [%s, %s], want numeric order", lessons[0].ID, lessons[1].ID)
	}
}

func TestFindLesson(t *testing.T) {
	dir := writeCourse(t, validCourseYAML, map[string]string{
		"lessons/01-what-is-python/lesson.md": validLessonMD,
		"lessons/02-variables/lesson.md":      strings.Replace(validLessonMD, "What is Python?", "Variables", 2),
	})
	c, err := LoadCourse(dir)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		id      string
		wantID  string
		wantErr bool
	}{
		{id: "01-what-is-python", wantID: "01-what-is-python"},
		{id: "01", wantID: "01-what-is-python"}, // numeric prefix shorthand
		{id: "02", wantID: "02-variables"},
		{id: "03", wantErr: true},
		{id: "0", wantErr: true}, // must match a full hyphen-delimited prefix
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			l, err := c.FindLesson(tt.id)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("FindLesson(%q) = %s, want error", tt.id, l.ID)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if l.ID != tt.wantID {
				t.Errorf("FindLesson(%q) = %s, want %s", tt.id, l.ID, tt.wantID)
			}
		})
	}
}
