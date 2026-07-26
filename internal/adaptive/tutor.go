package adaptive

import "context"

// DefaultBaseURL is where coursesmith-tutor listens by default.
const DefaultBaseURL = "http://localhost:8765"

// Tutor is the adaptive-learning API. Both the HTTP Client (talks to the
// coursesmith-tutor service) and the in-process Mock (runs the engine locally)
// implement it, so callers and tests can swap one for the other.
type Tutor interface {
	// Health returns nil when the tutor is reachable and ready.
	Health(ctx context.Context) error
	// EstimateBKT returns posterior mastery for a concept given a sequence of
	// correct/incorrect responses. Pass nil params to use DefaultBKTParams.
	EstimateBKT(ctx context.Context, params *BKTParams, corrects []bool) (BKTEstimate, error)
	// ScheduleFSRS returns the next review interval for a rated review.
	ScheduleFSRS(ctx context.Context, req FSRSRequest) (FSRSResult, error)
	// CalibrateIRT returns per-item difficulty/discrimination for pooled
	// responses (currently neutral-prior stubs).
	CalibrateIRT(ctx context.Context, obs []IRTObservation) (IRTResult, error)
}
