package adaptive

import "context"

// Mock is an in-process Tutor that runs the engine directly — no HTTP, always
// healthy. Use it in tests and for callers that want adaptive routing without
// standing up the coursesmith-tutor service. Because it delegates to the same
// engine functions as the service, its answers match the real endpoints.
type Mock struct{}

var _ Tutor = Mock{}

func (Mock) Health(context.Context) error { return nil }

func (Mock) EstimateBKT(_ context.Context, params *BKTParams, corrects []bool) (BKTEstimate, error) {
	p := DefaultBKTParams()
	if params != nil {
		p = *params
	}
	return EstimateBKT(p, corrects), nil
}

func (Mock) ScheduleFSRS(_ context.Context, req FSRSRequest) (FSRSResult, error) {
	return ScheduleFSRS(req), nil
}

func (Mock) CalibrateIRT(_ context.Context, obs []IRTObservation) (IRTResult, error) {
	return CalibrateIRT(obs), nil
}
