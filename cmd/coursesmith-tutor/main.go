// Command coursesmith-tutor is the adaptive-learning microservice (workstream
// D). It runs alongside the static course site and personalizes practice from
// a student's quiz responses: Bayesian Knowledge Tracing for mastery, a
// spaced-repetition scheduler for review timing, and (stubbed) IRT calibration.
//
// The course itself stays static HTML; this service is opt-in — lesson pages
// POST responses and get back mastery + next-step suggestions.
//
// All the learning-science math lives in internal/adaptive (shared with the Go
// client and the in-process Mock); this file is only the HTTP surface. Scaffold
// status: BKT and the scheduler compute real values; IRT is a stub that
// documents the pyBKT/py-irt handoff (see tools/tutor/). No persistence yet —
// every request is stateless.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/enfec/coursesmith/internal/adaptive"
)

func main() {
	addr := flag.String("addr", "localhost:8765", "listen address")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /bkt/estimate", handleBKT)
	mux.HandleFunc("POST /irt/calibrate", handleIRT)
	mux.HandleFunc("POST /fsrs/schedule", handleFSRS)

	log.Printf("coursesmith-tutor listening on http://%s", *addr)
	if err := http.ListenAndServe(*addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

// withCORS allows the studio (served from a different origin in dev) to call
// this service from the browser. The tutor holds no secrets and no
// credentials, so a permissive policy is fine for the scaffold; tighten the
// allowed origin if this is ever exposed publicly. Pre-flight OPTIONS requests
// are answered here and never reach the handlers.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "600")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return false
	}
	return true
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "coursesmith-tutor"})
}

// handleBKT estimates mastery from a sequence of correct/incorrect responses.
func handleBKT(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Params    *adaptive.BKTParams `json:"params,omitempty"`
		Responses []struct {
			Correct bool `json:"correct"`
		} `json:"responses"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	p := adaptive.DefaultBKTParams()
	if req.Params != nil {
		p = *req.Params
	}
	corrects := make([]bool, len(req.Responses))
	for i, resp := range req.Responses {
		corrects[i] = resp.Correct
	}
	writeJSON(w, http.StatusOK, adaptive.EstimateBKT(p, corrects))
}

// handleFSRS returns the next review interval for a rated review.
func handleFSRS(w http.ResponseWriter, r *http.Request) {
	var req adaptive.FSRSRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	writeJSON(w, http.StatusOK, adaptive.ScheduleFSRS(req))
}

// handleIRT returns per-item difficulty/discrimination (neutral-prior stub).
func handleIRT(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Responses []adaptive.IRTObservation `json:"responses"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	writeJSON(w, http.StatusOK, adaptive.CalibrateIRT(req.Responses))
}
