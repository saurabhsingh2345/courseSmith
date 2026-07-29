package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	// ErrCourseExists means a course with the derived slug already exists.
	ErrCourseExists = errors.New("course already exists")
	// ErrEmptySlug means the name has no slug-able characters.
	ErrEmptySlug = errors.New("course name produces an empty slug")
)

var nonSlugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts "Python Basics!" → "python-basics".
func Slugify(name string) string {
	s := strings.ToLower(name)
	s = nonSlugRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// ScaffoldResult reports the files ScaffoldCourse created.
type ScaffoldResult struct {
	Slug       string
	CourseDir  string
	CourseFile string
	LessonFile string
}

const defaultDescription = "Describe your course in one or two sentences."

// ScaffoldCourse creates courses/<slug>/course.yaml plus an example first
// lesson, then verifies the scaffold parses. An empty description falls back
// to a placeholder. Returns ErrCourseExists or ErrEmptySlug (wrapped) for
// those cases so callers can map them to the right status.
func ScaffoldCourse(coursesDir, name, description string) (ScaffoldResult, error) {
	slug := Slugify(name)
	if slug == "" {
		return ScaffoldResult{}, fmt.Errorf("%w: %q", ErrEmptySlug, name)
	}
	desc := strings.TrimSpace(description)
	if desc == "" {
		desc = defaultDescription
	}

	courseDir := filepath.Join(coursesDir, slug)
	courseFile := filepath.Join(courseDir, CourseFileName)
	if _, err := os.Stat(courseFile); err == nil {
		return ScaffoldResult{}, fmt.Errorf("%w at %s", ErrCourseExists, courseDir)
	}

	lessonDir := filepath.Join(courseDir, "lessons", "01-welcome")
	if err := os.MkdirAll(lessonDir, 0o755); err != nil {
		return ScaffoldResult{}, fmt.Errorf("creating %s: %w", lessonDir, err)
	}
	lessonFile := filepath.Join(lessonDir, LessonFileName)

	courseYAML := fmt.Sprintf(`name: %q
slug: %s
description: %s

style:
  voice: af_heart
  tone: friendly, conversational teacher
  pace_wpm: 150
  audience: absolute beginners with no programming experience
  language: en

branding:
  colors:
    primary: "#2563eb"
    accent: "#f59e0b"
    background: "#ffffff"
  diagram_style: clean, flat, rounded corners, generous whitespace

pipeline:
  llm_content: openai/gpt-4o-mini
  llm_review: openai/gpt-4o-mini
  review_threshold: 8
  captions_model: whisper-large-v3
  # The course is its videos: skip quiz/mistakes/exercises/site stages.
  # Set false to generate the companion material too.
  video_only: true
`, name, slug, strconv.Quote(desc))

	lessonMD := `---
title: Welcome to the course
diagrams:
  - id: course-map
    prompt: "A simple roadmap of this course: 3 boxes left to right labeled
      'Basics', 'Practice', 'Projects', connected by arrows."
---

# Welcome to the course

## What you'll learn
- Replace this outline with your lesson's key points.
- Each heading becomes a section of narration.

[DIAGRAM: course-map]

## Your first step
- Tell learners what to do right after watching.

[DEMO: show the finished result of the course so learners know where they're headed]
`

	files := map[string]string{
		courseFile: courseYAML,
		lessonFile: lessonMD,
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return ScaffoldResult{}, fmt.Errorf("writing %s: %w", path, err)
		}
	}

	// Prove the scaffold parses before declaring success.
	if _, err := LoadCourse(courseDir); err != nil {
		return ScaffoldResult{}, fmt.Errorf("scaffold failed validation (this is a bug): %w", err)
	}

	return ScaffoldResult{
		Slug:       slug,
		CourseDir:  courseDir,
		CourseFile: courseFile,
		LessonFile: lessonFile,
	}, nil
}
