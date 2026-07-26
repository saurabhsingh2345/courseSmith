package pipeline

// Human review notes: a reviewer writes courses/<slug>/review-notes.yaml
// (lesson → section → note); the next pipeline run injects unresolved
// notes into the script regeneration prompt and marks them resolved.
//
// Idempotency: the script stage's inputs include a hash of the lesson's
// UNRESOLVED notes. New notes make the stage stale; the stage resolves
// them before the runner records its post-run input hashes, so state
// stabilizes again immediately.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ReviewNotesFileName sits in the course directory, next to course.yaml.
const ReviewNotesFileName = "review-notes.yaml"

// NotesAppliedFileName records which notes a run injected, under
// generated/reviews/.
const NotesAppliedFileName = "notes-applied.json"

// ReviewNote is one piece of reviewer feedback.
type ReviewNote struct {
	Note     string `yaml:"note" json:"note"`
	Resolved bool   `yaml:"resolved,omitempty" json:"resolved,omitempty"`
}

// LessonNotes groups a lesson's notes: lesson-wide and per section.
type LessonNotes struct {
	Notes    []ReviewNote            `yaml:"notes,omitempty"`
	Sections map[string][]ReviewNote `yaml:"sections,omitempty"`
}

// ReviewNotes is the parsed review-notes.yaml.
type ReviewNotes struct {
	Lessons map[string]*LessonNotes `yaml:"lessons"`

	path string
}

// courseDirOf returns the course directory for a lesson dir
// (courses/<slug>/lessons/<id> → courses/<slug>).
func courseDirOf(lessonDir string) string {
	return filepath.Dir(filepath.Dir(lessonDir))
}

// LoadReviewNotes reads <courseDir>/review-notes.yaml; a missing file
// returns an empty (but savable) set.
func LoadReviewNotes(courseDir string) (*ReviewNotes, error) {
	path := filepath.Join(courseDir, ReviewNotesFileName)
	notes := &ReviewNotes{Lessons: map[string]*LessonNotes{}, path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return notes, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, notes); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if notes.Lessons == nil {
		notes.Lessons = map[string]*LessonNotes{}
	}
	notes.path = path
	return notes, nil
}

// UnresolvedText renders a lesson's unresolved notes for prompt injection;
// "" when there are none.
func (r *ReviewNotes) UnresolvedText(lessonID string) string {
	ln := r.Lessons[lessonID]
	if ln == nil {
		return ""
	}
	var b strings.Builder
	for _, n := range ln.Notes {
		if !n.Resolved && strings.TrimSpace(n.Note) != "" {
			fmt.Fprintf(&b, "- %s\n", strings.TrimSpace(n.Note))
		}
	}
	sections := make([]string, 0, len(ln.Sections))
	for s := range ln.Sections {
		sections = append(sections, s)
	}
	sort.Strings(sections)
	for _, s := range sections {
		for _, n := range ln.Sections[s] {
			if !n.Resolved && strings.TrimSpace(n.Note) != "" {
				fmt.Fprintf(&b, "- [section %s] %s\n", s, strings.TrimSpace(n.Note))
			}
		}
	}
	return strings.TrimSpace(b.String())
}

// MarkResolved flags every note of one lesson as resolved. Returns how
// many notes changed state.
func (r *ReviewNotes) MarkResolved(lessonID string) int {
	ln := r.Lessons[lessonID]
	if ln == nil {
		return 0
	}
	changed := 0
	for i := range ln.Notes {
		if !ln.Notes[i].Resolved {
			ln.Notes[i].Resolved = true
			changed++
		}
	}
	for s := range ln.Sections {
		for i := range ln.Sections[s] {
			if !ln.Sections[s][i].Resolved {
				ln.Sections[s][i].Resolved = true
				changed++
			}
		}
	}
	return changed
}

// Save writes the notes back to review-notes.yaml. Comments in the
// original file are not preserved (the file is data, re-marshaled).
func (r *ReviewNotes) Save() error {
	var buf bytes.Buffer
	buf.WriteString("# Reviewer notes. The next pipeline run injects unresolved notes into\n")
	buf.WriteString("# script regeneration and sets resolved: true on them.\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("encoding review notes: %w", err)
	}
	if err := enc.Close(); err != nil {
		return err
	}
	return writeFileAtomic(r.path, buf.Bytes())
}
