package adaptive

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// tutorTestServer stands up the same HTTP contract as coursesmith-tutor,
// delegating to the shared engine, so the Client is exercised against a real
// round-trip (marshal → HTTP → engine → unmarshal).
func tutorTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /bkt/estimate", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Params    *BKTParams `json:"params"`
			Responses []struct {
				Correct bool `json:"correct"`
			} `json:"responses"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"bad"}`, http.StatusBadRequest)
			return
		}
		p := DefaultBKTParams()
		if req.Params != nil {
			p = *req.Params
		}
		corrects := make([]bool, len(req.Responses))
		for i, x := range req.Responses {
			corrects[i] = x.Correct
		}
		json.NewEncoder(w).Encode(EstimateBKT(p, corrects))
	})
	mux.HandleFunc("POST /fsrs/schedule", func(w http.ResponseWriter, r *http.Request) {
		var req FSRSRequest
		json.NewDecoder(r.Body).Decode(&req)
		json.NewEncoder(w).Encode(ScheduleFSRS(req))
	})
	mux.HandleFunc("POST /irt/calibrate", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Responses []IRTObservation `json:"responses"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		json.NewEncoder(w).Encode(CalibrateIRT(req.Responses))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestClientRoundTrip(t *testing.T) {
	srv := tutorTestServer(t)
	c := NewClient(srv.URL)
	ctx := context.Background()

	if err := c.Health(ctx); err != nil {
		t.Fatalf("health: %v", err)
	}

	est, err := c.EstimateBKT(ctx, nil, []bool{true, true, true})
	if err != nil {
		t.Fatalf("bkt: %v", err)
	}
	want := EstimateBKT(DefaultBKTParams(), []bool{true, true, true})
	if est != want {
		t.Errorf("client BKT = %+v, want engine %+v", est, want)
	}

	fsrs, err := c.ScheduleFSRS(ctx, FSRSRequest{Rating: "good"})
	if err != nil {
		t.Fatalf("fsrs: %v", err)
	}
	if fsrs.IntervalDays < 1 {
		t.Errorf("fsrs interval = %d, want >= 1", fsrs.IntervalDays)
	}

	irt, err := c.CalibrateIRT(ctx, []IRTObservation{{QuestionID: "q1", Correct: true}})
	if err != nil {
		t.Fatalf("irt: %v", err)
	}
	if len(irt.Calibrated) != 1 || irt.Calibrated[0].QuestionID != "q1" {
		t.Errorf("irt calibrated = %+v", irt.Calibrated)
	}
}

func TestClientHealthErrorWhenDown(t *testing.T) {
	// A client pointed at a closed port should surface an error, not hang.
	c := NewClient("http://127.0.0.1:1")
	if err := c.Health(context.Background()); err == nil {
		t.Fatal("expected health error for an unreachable tutor")
	}
}

func TestMockMatchesEngine(t *testing.T) {
	var m Tutor = Mock{}
	ctx := context.Background()
	est, _ := m.EstimateBKT(ctx, nil, []bool{true, false, true})
	if est != EstimateBKT(DefaultBKTParams(), []bool{true, false, true}) {
		t.Error("mock BKT diverged from engine")
	}
	if err := m.Health(ctx); err != nil {
		t.Errorf("mock health: %v", err)
	}
}
