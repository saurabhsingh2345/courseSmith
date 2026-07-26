// Package studio is the local-first backend for the coursesmith Studio UI:
// a JSON API over project state, a run orchestrator with SSE progress, a
// token/cost ledger, and artifact serving. It binds to localhost and has
// no auth by design.
package studio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Event is one SSE message. Type is one of:
//
//	run-started, stage-started, stage-finished, stage-skipped,
//	stage-failed, log, run-finished, run-failed, quota
type Event struct {
	Type   string `json:"type"`
	RunID  string `json:"run_id,omitempty"`
	Course string `json:"course,omitempty"`
	Lesson string `json:"lesson,omitempty"`
	Stage  string `json:"stage,omitempty"`
	Line   string `json:"line,omitempty"`
	Error  string `json:"error,omitempty"`
	// Seq lets clients resume after reconnect (Last-Event-ID).
	Seq int64     `json:"seq"`
	At  time.Time `json:"at"`
}

// eventHub broadcasts events to all connected SSE clients and keeps a
// short replay buffer so reconnecting clients can catch up.
type eventHub struct {
	mu      sync.Mutex
	nextSeq int64
	clients map[chan Event]struct{}
	// replay is a ring of recent events for Last-Event-ID catch-up.
	replay []Event
}

const replayLimit = 512

func newEventHub() *eventHub {
	return &eventHub{clients: map[chan Event]struct{}{}}
}

// Publish stamps and fans out one event. Slow clients are skipped rather
// than blocking the pipeline.
func (h *eventHub) Publish(e Event) {
	h.mu.Lock()
	h.nextSeq++
	e.Seq = h.nextSeq
	e.At = time.Now().UTC()
	h.replay = append(h.replay, e)
	if len(h.replay) > replayLimit {
		h.replay = h.replay[len(h.replay)-replayLimit:]
	}
	for ch := range h.clients {
		select {
		case ch <- e:
		default:
		}
	}
	h.mu.Unlock()
}

// Subscribe registers a client and returns its channel, any replay events
// after lastSeq, and an unsubscribe func.
func (h *eventHub) Subscribe(lastSeq int64) (<-chan Event, []Event, func()) {
	ch := make(chan Event, 64)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	var missed []Event
	if lastSeq > 0 {
		for _, e := range h.replay {
			if e.Seq > lastSeq {
				missed = append(missed, e)
			}
		}
	}
	h.mu.Unlock()
	return ch, missed, func() {
		h.mu.Lock()
		delete(h.clients, ch)
		h.mu.Unlock()
	}
}

// handleEvents is the GET /api/events SSE endpoint.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	var lastSeq int64
	fmt.Sscanf(r.Header.Get("Last-Event-ID"), "%d", &lastSeq)

	ch, missed, unsubscribe := s.hub.Subscribe(lastSeq)
	defer unsubscribe()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	write := func(e Event) bool {
		data, err := json.Marshal(e)
		if err != nil {
			return true
		}
		if _, err := fmt.Fprintf(w, "id: %d\ndata: %s\n\n", e.Seq, data); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}
	for _, e := range missed {
		if !write(e) {
			return
		}
	}
	// Heartbeat keeps proxies from closing the idle stream.
	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case e := <-ch:
			if !write(e) {
				return
			}
		}
	}
}
