package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLesson(t *testing.T, content string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "01-test-lesson")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, LessonFileName), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLoadLesson(t *testing.T) {
	tests := []struct {
		name    string
		md      string
		wantErr string
		check   func(t *testing.T, l *Lesson)
	}{
		{
			name: "full front-matter with overrides",
			md: `---
title: What is Python?
diagrams:
  - id: memory-model
    prompt: "3 variables in memory"
  - id: flow-chart
    prompt: "if/else flow"
style:
  tone: extra playful
pipeline:
  review_threshold: 9
---

# Outline
- point one
`,
			check: func(t *testing.T, l *Lesson) {
				if l.ID != "01-test-lesson" {
					t.Errorf("ID = %q", l.ID)
				}
				if l.FrontMatter.Title != "What is Python?" {
					t.Errorf("Title = %q", l.FrontMatter.Title)
				}
				if len(l.FrontMatter.Diagrams) != 2 || l.FrontMatter.Diagrams[1].ID != "flow-chart" {
					t.Errorf("Diagrams = %+v", l.FrontMatter.Diagrams)
				}
				ov := l.FrontMatter.Overrides()
				if ov.Style.Tone != "extra playful" || ov.Pipeline.ReviewThreshold != 9 {
					t.Errorf("Overrides() = %+v", ov)
				}
				if !strings.Contains(l.Body, "# Outline") {
					t.Errorf("Body = %q, want markdown after front-matter", l.Body)
				}
				if strings.Contains(l.Body, "review_threshold") {
					t.Errorf("Body contains front-matter content: %q", l.Body)
				}
			},
		},
		{
			name:    "missing title",
			md:      "---\ndiagrams: []\n---\n# Body\n",
			wantErr: "title is required",
		},
		{
			name:    "no front-matter at all",
			md:      "# Just markdown\n",
			wantErr: "title is required",
		},
		{
			name:    "empty body",
			md:      "---\ntitle: X\n---\n\n   \n",
			wantErr: "body is empty",
		},
		{
			name:    "unclosed front-matter",
			md:      "---\ntitle: X\n# Body\n",
			wantErr: "never closed",
		},
		{
			name:    "diagram missing prompt",
			md:      "---\ntitle: X\ndiagrams:\n  - id: foo\n---\n# Body\n",
			wantErr: `diagram "foo": prompt is required`,
		},
		{
			name:    "duplicate diagram ids",
			md:      "---\ntitle: X\ndiagrams:\n  - id: foo\n    prompt: a\n  - id: foo\n    prompt: b\n---\n# Body\n",
			wantErr: "duplicate diagram id",
		},
		{
			name:    "unknown front-matter field",
			md:      "---\ntitle: X\ndiagram:\n  - id: typo\n---\n# Body\n",
			wantErr: "field diagram not found",
		},
		{
			name: "horizontal rule in body is not a closing delimiter trap",
			md: `---
title: X
---
# Body

----

more body
`,
			check: func(t *testing.T, l *Lesson) {
				if !strings.Contains(l.Body, "----") || !strings.Contains(l.Body, "more body") {
					t.Errorf("Body = %q, want full body preserved", l.Body)
				}
			},
		},
		{
			name: "crlf line endings",
			md:   "---\r\ntitle: X\r\n---\r\n# Body\r\n",
			check: func(t *testing.T, l *Lesson) {
				if l.FrontMatter.Title != "X" {
					t.Errorf("Title = %q", l.FrontMatter.Title)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeLesson(t, tt.md)
			l, err := LoadLesson(dir)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("LoadLesson() succeeded, want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.check != nil {
				tt.check(t, l)
			}
		})
	}
}

func TestLessonPaths(t *testing.T) {
	dir := writeLesson(t, "---\ntitle: X\n---\n# Body\n")
	l, err := LoadLesson(dir)
	if err != nil {
		t.Fatal(err)
	}
	if l.SourcePath() != filepath.Join(dir, LessonFileName) {
		t.Errorf("SourcePath() = %q", l.SourcePath())
	}
	if l.GeneratedDir() != filepath.Join(dir, GeneratedDirName) {
		t.Errorf("GeneratedDir() = %q", l.GeneratedDir())
	}
}
