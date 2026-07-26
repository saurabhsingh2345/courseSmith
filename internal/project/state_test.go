package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testLesson(t *testing.T) *Lesson {
	t.Helper()
	dir := writeLesson(t, "---\ntitle: X\n---\n# Body\n")
	l, err := LoadLesson(dir)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestStateRoundTrip(t *testing.T) {
	l := testLesson(t)

	s, err := l.LoadState()
	if err != nil {
		t.Fatalf("LoadState() on fresh lesson: %v", err)
	}
	if len(s.Stages) != 0 {
		t.Fatalf("fresh state has %d stages, want 0", len(s.Stages))
	}

	inputs := map[string]string{"lesson.md": "sha256:aaa", "config": "sha256:bbb"}
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	s.MarkDone(StageScript, inputs, now)
	if err := l.SaveState(s); err != nil {
		t.Fatalf("SaveState(): %v", err)
	}

	loaded, err := l.LoadState()
	if err != nil {
		t.Fatalf("LoadState() after save: %v", err)
	}
	rec, ok := loaded.Stages[StageScript]
	if !ok {
		t.Fatalf("saved stage missing; state = %+v", loaded)
	}
	if !rec.CompletedAt.Equal(now) {
		t.Errorf("CompletedAt = %v, want %v", rec.CompletedAt, now)
	}
	if rec.InputHashes["lesson.md"] != "sha256:aaa" {
		t.Errorf("InputHashes = %+v", rec.InputHashes)
	}
}

func TestStageStatusInvalidation(t *testing.T) {
	inputs := map[string]string{"lesson.md": "sha256:aaa", "config": "sha256:bbb"}
	s := &State{}
	s.MarkDone(StageScript, inputs, time.Now())

	tests := []struct {
		name    string
		stage   string
		current map[string]string
		want    StageStatus
	}{
		{
			name:    "same inputs → done",
			stage:   StageScript,
			current: map[string]string{"lesson.md": "sha256:aaa", "config": "sha256:bbb"},
			want:    StatusDone,
		},
		{
			name:    "changed hash → stale",
			stage:   StageScript,
			current: map[string]string{"lesson.md": "sha256:CHANGED", "config": "sha256:bbb"},
			want:    StatusStale,
		},
		{
			name:    "new input key → stale",
			stage:   StageScript,
			current: map[string]string{"lesson.md": "sha256:aaa", "config": "sha256:bbb", "script.json": "sha256:ccc"},
			want:    StatusStale,
		},
		{
			name:    "removed input key → stale",
			stage:   StageScript,
			current: map[string]string{"lesson.md": "sha256:aaa"},
			want:    StatusStale,
		},
		{
			name:    "never ran → pending",
			stage:   StageVisuals,
			current: map[string]string{"lesson.md": "sha256:aaa"},
			want:    StatusPending,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.StageStatus(tt.stage, tt.current); got != tt.want {
				t.Errorf("StageStatus(%s) = %s, want %s", tt.stage, got, tt.want)
			}
		})
	}
}

func TestMarkDoneCopiesInputs(t *testing.T) {
	inputs := map[string]string{"lesson.md": "sha256:aaa"}
	s := &State{}
	s.MarkDone(StageScript, inputs, time.Now())
	inputs["lesson.md"] = "sha256:mutated"
	if s.Stages[StageScript].InputHashes["lesson.md"] != "sha256:aaa" {
		t.Error("MarkDone stored the caller's map instead of a copy")
	}
}

func TestInvalidate(t *testing.T) {
	s := &State{}
	s.MarkDone(StageScript, map[string]string{"a": "b"}, time.Now())
	s.Invalidate(StageScript)
	if got := s.StageStatus(StageScript, map[string]string{"a": "b"}); got != StatusPending {
		t.Errorf("after Invalidate, status = %s, want pending", got)
	}
}

func TestLoadStateCorruptFile(t *testing.T) {
	l := testLesson(t)
	if err := os.MkdirAll(l.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(l.StatePath(), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := l.LoadState()
	if err == nil || !strings.Contains(err.Error(), "delete it to reset") {
		t.Errorf("error = %v, want actionable parse error", err)
	}
}

func TestHashHelpers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	h1, err := HashFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != HashBytes([]byte("hello")) {
		t.Errorf("HashFile and HashBytes disagree: %s vs %s", h1, HashBytes([]byte("hello")))
	}
	if !strings.HasPrefix(h1, "sha256:") {
		t.Errorf("hash %q missing sha256: prefix", h1)
	}

	a := HashStrings(map[string]string{"x": "1", "y": "2"})
	b := HashStrings(map[string]string{"y": "2", "x": "1"})
	if a != b {
		t.Error("HashStrings is not order-independent")
	}
	c := HashStrings(map[string]string{"x": "1", "y": "3"})
	if a == c {
		t.Error("HashStrings collides on different values")
	}

	if _, err := HashFile(filepath.Join(dir, "missing")); err == nil {
		t.Error("HashFile on missing file should error")
	}
}
