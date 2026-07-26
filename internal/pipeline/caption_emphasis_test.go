package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/enfec/coursesmith/internal/project"
)

func TestLoadCaptionEmphasisAbsent(t *testing.T) {
	l := &project.Lesson{Dir: t.TempDir()}
	ce, err := loadCaptionEmphasis(l)
	if err != nil {
		t.Fatalf("absent file must not error: %v", err)
	}
	if ce != nil {
		t.Fatalf("absent file must yield nil, got %+v", ce)
	}
}

func TestLoadCaptionEmphasisRoundTrip(t *testing.T) {
	l := &project.Lesson{Dir: t.TempDir()}
	if err := os.MkdirAll(l.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), CaptionEmphasisFileName), &CaptionEmphasis{Indices: []int{3, 17, 42}}); err != nil {
		t.Fatal(err)
	}
	ce, err := loadCaptionEmphasis(l)
	if err != nil {
		t.Fatal(err)
	}
	if ce == nil || len(ce.Indices) != 3 || ce.Indices[2] != 42 {
		t.Fatalf("round trip mismatch: %+v", ce)
	}
}

func TestLoadCaptionEmphasisCorrupt(t *testing.T) {
	l := &project.Lesson{Dir: t.TempDir()}
	if err := os.MkdirAll(l.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(l.GeneratedDir(), CaptionEmphasisFileName), []byte("{nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadCaptionEmphasis(l); err == nil {
		t.Fatal("corrupt file must error")
	}
}
