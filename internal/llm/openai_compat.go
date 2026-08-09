package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

// Request deadlines. Two of them, because a reasoning model and a classic one
// fail in opposite directions against a single number.
//
// 120s was the whole budget and it was right for chat completions. A reasoning
// model at medium effort on a long review prompt runs past it, and the shape of
// the failure is the worst kind: the retry stack exhausts, the review stage
// reports "could not run, keeping the plan as-is", and the pipeline carries on
// with the quality gate silently switched off. That is the same class of quiet
// degradation as the 401s that once left the no-code course unreviewed for a
// whole run — right to keep going, wrong to be invisible.
//
// So the classic deadline stays where it was (a gpt-4o call that has not
// answered in two minutes is stuck, and waiting ten does not help), and only
// the models that legitimately think for longer get the longer one.
const (
	classicRequestTimeout   = 120 * time.Second
	reasoningRequestTimeout = 10 * time.Minute
)

func newHTTPProvider(name, baseURL, apiKey string, opts ...Option) *httpProvider {
	p := &httpProvider{
		name:   name,
		apiKey: apiKey,
		// The ceiling; Complete narrows it per request via the context, which is
		// the only place the model — and so the right deadline — is known.
		baseURL: baseURL,
		client:  &http.Client{Timeout: reasoningRequestTimeout},
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
	Temperature *float64 `json:"temperature,omitempty"`
	// MaxTokens and MaxCompletionTokens are the SAME budget under two names;
	// exactly one goes on the wire, chosen by paramStyleFor. See legacyModelPrefixes.
	MaxTokens           int               `json:"max_tokens,omitempty"`
	MaxCompletionTokens int               `json:"max_completion_tokens,omitempty"`
	ReasoningEffort     string            `json:"reasoning_effort,omitempty"`
	ResponseFormat      *responseFormat   `json:"response_format,omitempty"`
	WebSearchOptions    *webSearchOptions `json:"web_search_options,omitempty"`
}

// legacyModelPrefixes are the OpenAI families that still speak the pre-GPT-5
// chat-completions contract: `max_tokens`, and any temperature you like.
//
// This list is CLOSED, and that is the whole design. Every OpenAI model from
// gpt-5 onward renamed the output budget to `max_completion_tokens` and accepts
// only the default temperature, and both are hard 400s rather than parameters
// the API politely ignores:
//
//	"Unsupported parameter: 'max_tokens' is not supported with this model."
//	"'temperature' does not support 0 with this model. Only the default (1)."
//
// So we detect the OLD families and treat everything else as current. The
// inverse — an allow-list of known-new models — is what pinned this pipeline to
// gpt-4o-mini for its whole life: the config string could not be changed
// without the provider layer rejecting the call, so nobody changed it. A list
// of things that already exist cannot go stale; a list of things that don't
// yet, always does.
//
// "gpt-4" covers gpt-4, gpt-4o, gpt-4o-mini, gpt-4.1 and gpt-4-turbo by prefix.
// The o-series (o1/o3/o4) is deliberately absent: it wants the new contract.
var legacyModelPrefixes = []string{"gpt-4", "gpt-3.5", "chatgpt-4o"}

// paramStyle is how one model wants its sampling knobs on the wire.
type paramStyle struct {
	// maxCompletionTokens sends the budget under the newer name.
	maxCompletionTokens bool
	// temperature reports whether an explicit temperature is accepted at all.
	temperature bool
	// reasoning reports whether reasoning_effort is meaningful. Sending it to a
	// non-reasoning model is a 400, so it follows the same closed-list rule.
	reasoning bool
}

// paramStyleFor picks the wire contract for a provider/model pair.
//
// Only OpenAI ever changed the contract. Groq (and anything else routed through
// this same OpenAI-compatible client) still speaks the classic one, so the
// provider name gates the whole question before the model name is consulted.
func paramStyleFor(provider, model string) paramStyle {
	classic := paramStyle{maxCompletionTokens: false, temperature: true, reasoning: false}
	if provider != "openai" {
		return classic
	}
	for _, prefix := range legacyModelPrefixes {
		if strings.HasPrefix(model, prefix) {
			return classic
		}
	}
	return paramStyle{maxCompletionTokens: true, temperature: false, reasoning: true}
}

// webSearchOptions turns on the chat-completions search path. Empty is
// deliberate: the defaults are what we want, and every knob it accepts
// (search_context_size, user_location) would either cost more or make the
// result depend on where the machine is — which for a cached, resumable
// pipeline means two developers grounding the same combo differently.
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
	style := paramStyleFor(p.name, req.Model)
	body := chatRequest{
		Model:    req.Model,
		Messages: wireMessages(req),
	}
	if style.maxCompletionTokens {
		body.MaxCompletionTokens = req.MaxTokens
	} else {
		body.MaxTokens = req.MaxTokens
	}
	if style.temperature {
		body.Temperature = &req.Temperature // always explicit, for reproducibility
	}
	// Determinism for the models that refuse an explicit temperature comes from
	// the response cache instead, exactly as it does for web search below.
	if style.reasoning {
		body.ReasoningEffort = req.ReasoningEffort
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
		// And reasoning_effort goes with it. The search models are a much shorter
		// list than the rest and accept a narrower set of parameters; sending one
		// they do not take is a hard 400 that fails the substance stage, which is
		// the stage a course's factual grounding depends on. There is nothing to
		// gain either way — the thinking on a grounded call is the search.
		body.ReasoningEffort = ""
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encoding %s request: %w", p.name, err)
	}

	// Narrow the client's ceiling to what this particular model should need.
	// A caller that already set a shorter deadline keeps it — context.WithTimeout
	// never extends an existing one.
	deadline := classicRequestTimeout
	if style.reasoning {
		deadline = reasoningRequestTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

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
		// On a reasoning model the budget is shared with thinking that never
		// reaches Content, so this fires with an EMPTY reply and "raise the
		// limit" is only half the advice — lowering ReasoningEffort buys back
		// the same room more cheaply. Say so, because an empty response with a
		// truncation error reads like a broken prompt.
		field := "max_tokens"
		hint := "raise MaxTokens"
		if style.reasoning {
			field = "max_completion_tokens"
			hint = fmt.Sprintf("raise MaxTokens or lower ReasoningEffort (currently %q) — reasoning tokens are spent from this same budget before any content is emitted", req.ReasoningEffort)
		}
		return nil, fmt.Errorf("%s response was truncated at the %s limit (%d), returning %d characters — %s",
			p.name, field, req.MaxTokens, len(choice.Message.Content), hint)
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
