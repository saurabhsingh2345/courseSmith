package studio

// Run orchestration: one pipeline run at a time, executed stage by stage so
// the SSE stream gets precise stage-started / stage-finished / stage-failed
// events, with every progress line forwarded as a log event.

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/pipeline"
	"github.com/enfec/coursesmith/internal/project"
)

// RunRequest is the POST /api/run (and /api/regenerate) payload.
type RunRequest struct {
	Course string `json:"course"`
	Lesson string `json:"lesson"`
	// Stage restricts the run to one stage; "" runs the full pipeline.
	Stage string `json:"stage,omitempty"`
	Force bool   `json:"force,omitempty"`
}

// RunStatus is the current run state exposed at GET /api/run.
type RunStatus struct {
	Running bool   `json:"running"`
	RunID   string `json:"run_id,omitempty"`
	Course  string `json:"course,omitempty"`
	Lesson  string `json:"lesson,omitempty"`
	Stage   string `json:"stage,omitempty"`
}

// runManager executes at most one pipeline run at a time.
type runManager struct {
	mu      sync.Mutex
	running bool
	status  RunStatus
	cancel  context.CancelFunc
	counter int

	env *pipeline.Env
	hub *eventHub
}

func newRunManager(env *pipeline.Env, hub *eventHub) *runManager {
	return &runManager{env: env, hub: hub}
}

// Status returns a snapshot of the current run.
func (m *runManager) Status() RunStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

// Cancel stops the current run, if any.
func (m *runManager) Cancel() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
		return true
	}
	return false
}

// hubWriter forwards pipeline progress output to SSE log events, line by
// line.
type hubWriter struct {
	hub    *eventHub
	runID  string
	course string
	lesson string
	buf    strings.Builder
}

func (w *hubWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	for {
		s := w.buf.String()
		i := strings.IndexByte(s, '\n')
		if i < 0 {
			break
		}
		line := strings.TrimRight(s[:i], " ")
		w.buf.Reset()
		w.buf.WriteString(s[i+1:])
		if strings.TrimSpace(line) == "" {
			continue
		}
		w.hub.Publish(Event{
			Type: "log", RunID: w.runID, Course: w.course, Lesson: w.lesson, Line: line,
		})
	}
	return len(p), nil
}

// Start launches a run of one stage ("" = full pipeline) in the background.
func (m *runManager) Start(course *project.Course, lesson *project.Lesson, stage string, force bool) (string, error) {
	cfg := config.Resolve(course.Config, lesson.FrontMatter.Overrides(), config.Config{})
	stages := project.StagesFor(cfg.Pipeline.VideoOnly)
	if pipeline.IsSnippet(lesson) {
		snippetStages, err := pipeline.SnippetStages(lesson)
		if err != nil {
			return "", err
		}
		stages = snippetStages
	}
	if stage != "" {
		// Naming a companion stage explicitly overrides video-only.
		if !slices.Contains(project.AllStages(), stage) {
			return "", fmt.Errorf("unknown stage %q", stage)
		}
		stages = []string{stage}
	}
	return m.StartStages(course, lesson, stages, force)
}

// StartStages launches a run of an explicit stage list in the background.
// It returns the run id, or an error when a run is already in progress.
func (m *runManager) StartStages(course *project.Course, lesson *project.Lesson, stages []string, force bool) (string, error) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return "", fmt.Errorf("a run is already in progress (%s/%s) — the studio runs one pipeline at a time", m.status.Course, m.status.Lesson)
	}
	m.counter++
	runID := fmt.Sprintf("run-%d", m.counter)
	ctx, cancel := context.WithCancel(context.Background())
	m.running = true
	m.cancel = cancel
	m.status = RunStatus{Running: true, RunID: runID, Course: course.Slug, Lesson: lesson.ID}
	m.mu.Unlock()

	go m.execute(ctx, cancel, runID, course, lesson, stages, force)
	return runID, nil
}

func (m *runManager) execute(ctx context.Context, cancel context.CancelFunc, runID string, course *project.Course, lesson *project.Lesson, stages []string, force bool) {
	defer func() {
		cancel()
		m.mu.Lock()
		m.running = false
		m.cancel = nil
		m.status = RunStatus{}
		m.mu.Unlock()
	}()

	// The env is shared with other handlers; runs get a copy wired to the
	// SSE stream.
	env := *m.env
	env.Out = &hubWriter{hub: m.hub, runID: runID, course: course.Slug, lesson: lesson.ID}

	m.hub.Publish(Event{Type: "run-started", RunID: runID, Course: course.Slug, Lesson: lesson.ID})
	for _, stage := range stages {
		if ctx.Err() != nil {
			m.hub.Publish(Event{Type: "run-failed", RunID: runID, Course: course.Slug, Lesson: lesson.ID, Error: "canceled"})
			return
		}
		m.mu.Lock()
		m.status.Stage = stage
		m.mu.Unlock()
		m.hub.Publish(Event{Type: "stage-started", RunID: runID, Course: course.Slug, Lesson: lesson.ID, Stage: stage})

		err := env.RunLesson(ctx, course, lesson, pipeline.RunOptions{Stage: stage, Force: force})
		if err != nil {
			evt := Event{Type: "stage-failed", RunID: runID, Course: course.Slug, Lesson: lesson.ID, Stage: stage, Error: err.Error()}
			if quota, ok := errors.AsType[*llm.QuotaError](err); ok {
				m.hub.Publish(Event{Type: "quota", RunID: runID, Error: quota.Error()})
			}
			m.hub.Publish(evt)
			m.hub.Publish(Event{Type: "run-failed", RunID: runID, Course: course.Slug, Lesson: lesson.ID, Stage: stage, Error: err.Error()})
			return
		}
		m.hub.Publish(Event{Type: "stage-finished", RunID: runID, Course: course.Slug, Lesson: lesson.ID, Stage: stage})
	}
	m.hub.Publish(Event{Type: "run-finished", RunID: runID, Course: course.Slug, Lesson: lesson.ID})
}
