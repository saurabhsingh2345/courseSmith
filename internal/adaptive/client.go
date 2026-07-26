package adaptive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is an HTTP Tutor talking to a coursesmith-tutor service.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient returns a Client for the tutor at baseURL (e.g. DefaultBaseURL).
func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{}}
}

var _ Tutor = (*Client)(nil)

// bktRequestBody is the wire shape the /bkt/estimate handler decodes.
type bktRequestBody struct {
	Params    *BKTParams        `json:"params,omitempty"`
	Responses []bktResponseItem `json:"responses"`
}

type bktResponseItem struct {
	Correct bool `json:"correct"`
}

type irtRequestBody struct {
	Responses []IRTObservation `json:"responses"`
}

func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("tutor unreachable at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tutor health check failed: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) EstimateBKT(ctx context.Context, params *BKTParams, corrects []bool) (BKTEstimate, error) {
	body := bktRequestBody{Params: params, Responses: make([]bktResponseItem, len(corrects))}
	for i, ok := range corrects {
		body.Responses[i] = bktResponseItem{Correct: ok}
	}
	var out BKTEstimate
	err := c.post(ctx, "/bkt/estimate", body, &out)
	return out, err
}

func (c *Client) ScheduleFSRS(ctx context.Context, req FSRSRequest) (FSRSResult, error) {
	var out FSRSResult
	err := c.post(ctx, "/fsrs/schedule", req, &out)
	return out, err
}

func (c *Client) CalibrateIRT(ctx context.Context, obs []IRTObservation) (IRTResult, error) {
	var out IRTResult
	err := c.post(ctx, "/irt/calibrate", irtRequestBody{Responses: obs}, &out)
	return out, err
}

// post marshals body, POSTs it to path, and decodes a 2xx JSON reply into out.
// A non-2xx status carries the server's {"error": ...} message when present.
func (c *Client) post(ctx context.Context, path string, body, out any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("tutor request to %s failed: %w", path, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var e struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &e) == nil && e.Error != "" {
			return fmt.Errorf("tutor %s: %s", path, e.Error)
		}
		return fmt.Errorf("tutor %s: HTTP %d", path, resp.StatusCode)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decoding tutor %s reply: %w", path, err)
	}
	return nil
}
