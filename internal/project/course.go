// Package project models a coursesmith project on disk: courses, lessons,
// and per-lesson pipeline state.
package project

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/enfec/coursesmith/internal/config"
)

// CourseFileName is the manifest expected at the root of every course dir.
const CourseFileName = "course.yaml"

// Course is a parsed course.yaml plus its location on disk.
type Course struct {
	Name        string        `yaml:"name"`
	Slug        string        `yaml:"slug"`
	Description string        `yaml:"description"`
	Config      config.Config `yaml:",inline"`

	// Dir is the absolute path of the course directory (not serialized).
	Dir string `yaml:"-"`
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// LoadCourse reads and validates <dir>/course.yaml.
func LoadCourse(dir string) (*Course, error) {
	path := filepath.Join(dir, CourseFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var c Course
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolving course dir %s: %w", dir, err)
	}
	c.Dir = abs
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", path, err)
	}
	return &c, nil
}

// Validate checks required fields and value sanity.
func (c *Course) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Slug == "" {
		return fmt.Errorf("slug is required")
	}
	if !slugRe.MatchString(c.Slug) {
		return fmt.Errorf("slug %q must be lowercase letters, digits, and hyphens (e.g. %q)", c.Slug, "python-basics")
	}
	if p := c.Config.Style.PaceWPM; p < 0 || p > 400 {
		return fmt.Errorf("style.pace_wpm %d is out of range (0-400)", p)
	}
	if t := c.Config.Pipeline.ReviewThreshold; t < 0 || t > 10 {
		return fmt.Errorf("pipeline.review_threshold %.1f is out of range (0-10)", t)
	}
	return nil
}

// Lessons returns the course's lesson directories (containing lesson.md),
// sorted by directory name so numeric prefixes give pipeline order.
func (c *Course) Lessons() ([]*Lesson, error) {
	lessonsDir := filepath.Join(c.Dir, "lessons")
	entries, err := os.ReadDir(lessonsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("course %q has no lessons/ directory (looked in %s)", c.Slug, lessonsDir)
		}
		return nil, fmt.Errorf("reading %s: %w", lessonsDir, err)
	}
	var lessons []*Lesson
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(lessonsDir, e.Name())
		if _, err := os.Stat(filepath.Join(dir, LessonFileName)); err != nil {
			continue // not a lesson dir
		}
		l, err := LoadLesson(dir)
		if err != nil {
			return nil, fmt.Errorf("lesson %s: %w", e.Name(), err)
		}
		lessons = append(lessons, l)
	}
	sort.Slice(lessons, func(i, j int) bool { return lessons[i].ID < lessons[j].ID })
	if len(lessons) == 0 {
		return nil, fmt.Errorf("course %q has no lessons (each lesson is a directory under %s containing %s)", c.Slug, lessonsDir, LessonFileName)
	}
	return lessons, nil
}

// FindLesson returns the lesson whose directory name equals id, or whose
// directory name starts with id + "-" (so "01" matches "01-what-is-python").
func (c *Course) FindLesson(id string) (*Lesson, error) {
	lessons, err := c.Lessons()
	if err != nil {
		return nil, err
	}
	var prefixMatch *Lesson
	for _, l := range lessons {
		if l.ID == id {
			return l, nil
		}
		if len(l.ID) > len(id) && l.ID[:len(id)] == id && l.ID[len(id)] == '-' {
			if prefixMatch != nil {
				return nil, fmt.Errorf("lesson id %q is ambiguous in course %q (matches %q and %q)", id, c.Slug, prefixMatch.ID, l.ID)
			}
			prefixMatch = l
		}
	}
	if prefixMatch != nil {
		return prefixMatch, nil
	}
	return nil, fmt.Errorf("no lesson %q in course %q", id, c.Slug)
}
