package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// httpProvider talks to any OpenAI-compatible chat-completions API.
// Both Groq and OpenAI use this with different base URLs.
type httpProvider struct {
	name    string
	baseURL string
	apiKey  string
	client  *http.Client
}

// Option customizes an HTTP provider (mainly for tests).
type Option func(*httpProvider)

// WithBaseURL overrides the API base URL (no trailing slash).
func WithBaseURL(url string) Option {
	return func(p *httpProvider) { p.baseURL = url }
}

// WithHTTPClient overrides the HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(p *httpProvider) { p.client = c }
}

func newHTTPProvider(name, baseURL, apiKey string, opts ...Option) *httpProvider {
	p := &httpProvider{
		name:    name,
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *httpProvider) Name() string { return p.name }

// Wire types for the OpenAI-compatible chat completions endpoint.
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []wireMessage `json:"messages"`
	// Temperature stays explicit for reproducibility. omitempty on a *pointer*
	// tests nil, not zero, so &0.0 still goes out as `"temperature": 0` and only
	// a deliberate nil is dropped — which is what a web-search request needs,
	// because search-capable models reject the parameter outright.
	Temperature      *float64          `json:"temperature,omitempty"`
	MaxTokens        int               `json:"max_tokens,omitempty"`
	ResponseFormat   *responseFormat   `json:"response_format,omitempty"`
	WebSearchOptions *webSearchOptions `json:"web_search_options,omitempty"`
}

// webSearchOptions turns on the chat-completions search path. Empty is
// deliberate: the defaults are what we want, and every knob it accepts
// (search_context_size, user_location) would either cost more or make the
// result depend on where the machine is — which for a cached, resumable
// pipeline means two developers grounding the same reel differently.
type webSearchOptions struct{}

// wireMessage allows string content or multimodal content parts.
type wireMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type contentPart struct {
	Type     string        `json:"type"` // "text" or "image_url"
	Text     string        `json:"text,omitempty"`
	ImageURL *imagePayload `json:"image_url,omitempty"`
}

type imagePayload struct {
	URL string `json:"url"`
}

// wireMessages converts a request's messages, attaching any images to the
// final user message as data-URL content parts.
func wireMessages(req Request) []wireMessage {
	out := make([]wireMessage, len(req.Messages))
	lastUser := -1
	for i, m := range req.Messages {
		out[i] = wireMessage{Role: m.Role, Content: m.Content}
		if m.Role == RoleUser {
			lastUser = i
		}
	}
	if len(req.Images) > 0 && lastUser >= 0 {
		parts := []contentPart{{Type: "text", Text: req.Messages[lastUser].Content}}
		for _, img := range req.Images {
			parts = append(parts, contentPart{
				Type:     "image_url",
				ImageURL: &imagePayload{URL: "data:image/png;base64," + img},
			})
		}
		out[lastUser].Content = parts
	}
	return out
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message      wireResponseMessage `json:"message"`
		FinishReason string              `json:"finish_reason"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
}

// wireResponseMessage is the assistant turn as it comes back. Separate from
// Message because a search response carries annotations that Message — which is
// also the *request* shape — has no business holding.
type wireResponseMessage struct {
	Role        string           `json:"role"`
	Content     string           `json:"content"`
	Annotations []wireAnnotation `json:"annotations,omitempty"`
}

type wireAnnotation struct {
	Type        string `json:"type"`
	URLCitation *struct {
		URL   string `json:"url"`
		Title string `json:"title"`
	} `json:"url_citation,omitempty"`
}

// citations extracts the sources, de-duplicated by URL and in first-cited order.
//
// The API annotates every *span* of the answer, so one source backing three
// sentences arrives three times. Order is kept rather than sorted because the
// first citation is generally the one the answer leans on hardest.
func (m wireResponseMessage) citations() []Citation {
	if len(m.Annotations) == 0 {
		return nil
	}
	var out []Citation
	seen := map[string]bool{}
	for _, a := range m.Annotations {
		if a.URLCitation == nil || a.URLCitation.URL == "" || seen[a.URLCitation.URL] {
			continue
		}
		seen[a.URLCitation.URL] = true
		out = append(out, Citation{URL: a.URLCitation.URL, Title: a.URLCitation.Title})
	}
	return out
}

type apiErrorBody struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *httpProvider) Complete(ctx context.Context, req Request) (*Response, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	body := chatRequest{
		Model:       req.Model,
		Messages:    wireMessages(req),
		Temperature: &req.Temperature, // always explicit, for reproducibility
		MaxTokens:   req.MaxTokens,
	}
	if req.JSONMode {
		body.ResponseFormat = &responseFormat{Type: "json_object"}
	}
	if req.WebSearch {
		if p.name != "openai" {
			return nil, fmt.Errorf("%s does not support web search — grounding needs an OpenAI search-capable model, so set pipeline.llm_search (or drop the grounding)", p.name)
		}
		body.WebSearchOptions = &webSearchOptions{}
		// Search-capable models reject an explicit temperature. Dropping it is
		// safe because a grounded call's reproducibility comes from the response
		// cache, not from sampling.
		body.Temperature = nil
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encoding %s request: %w", p.name, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("building %s request: %w", p.name, err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling %s API: %w", p.name, err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(httpResp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("reading %s response: %w", p.name, err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, p.apiError(httpResp, respBody)
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("parsing %s response: %w", p.name, err)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("%s returned no choices", p.name)
	}
	choice := parsed.Choices[0]
	if choice.FinishReason == "length" {
		return nil, fmt.Errorf("%s response was truncated at the max_tokens limit (%d) — raise MaxTokens", p.name, req.MaxTokens)
	}
	if req.WebSearch && len(choice.Message.Annotations) == 0 {
		// A search request that came back with no sources did not search. The
		// model answered from memory and the answer looks exactly like a grounded
		// one, which is the failure this whole path exists to prevent — so it is
		// an error rather than a warning. The usual cause is a model that is not
		// search-capable.
		return nil, fmt.Errorf("%s returned no sources for a web-search request (model %q) — it answered from memory, which is what grounding is meant to replace; check that the model is search-capable",
			p.name, parsed.Model)
	}
	return &Response{
		Content:   choice.Message.Content,
		Model:     parsed.Model,
		Usage:     parsed.Usage,
		Citations: choice.Message.citations(),
	}, nil
}

func (p *httpProvider) apiError(resp *http.Response, body []byte) *APIError {
	msg := ""
	var parsed apiErrorBody
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error.Message != "" {
		msg = parsed.Error.Message
	} else {
		msg = truncate(string(body), 300)
	}
	apiErr := &APIError{
		Provider:   p.name,
		StatusCode: resp.StatusCode,
		Message:    msg,
	}
	if s := resp.Header.Get("Retry-After"); s != "" {
		if secs, err := strconv.ParseFloat(s, 64); err == nil && secs > 0 {
			apiErr.RetryAfter = time.Duration(secs * float64(time.Second))
		}
	}
	return apiErr
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
