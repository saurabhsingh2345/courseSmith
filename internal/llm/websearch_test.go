package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The grounded response the search path expects: prose plus one annotation per
// cited span, which means a source backing two sentences arrives twice.
const searchResponseBody = `{
	"model": "gpt-5-search-api",
	"choices": [{"message": {
		"role": "assistant",
		"content": "Webflow has a free Starter plan; paid site plans begin at $14/month.",
		"annotations": [
			{"type": "url_citation", "url_citation": {"url": "https://webflow.com/pricing", "title": "Webflow Pricing"}},
			{"type": "url_citation", "url_citation": {"url": "https://webflow.com/pricing", "title": "Webflow Pricing"}},
			{"type": "url_citation", "url_citation": {"url": "https://example.com/review", "title": "A review"}}
		]
	}, "finish_reason": "stop"}],
	"usage": {"prompt_tokens": 40, "completion_tokens": 20, "total_tokens": 60}
}`

func TestWebSearchSendsOptionsAndDropsTemperature(t *testing.T) {
	var got chatRequest
	var rawBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		_ = json.Unmarshal(body, &rawBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchResponseBody))
	}))
	defer server.Close()

	req := testRequest()
	req.Temperature = 0.6
	req.WebSearch = true
	if _, err := NewOpenAI("sk_test", WithBaseURL(server.URL)).Complete(context.Background(), req); err != nil {
		t.Fatal(err)
	}

	if got.WebSearchOptions == nil {
		t.Error("web_search_options was not sent, so the request did not search")
	}
	// Asserted against the raw body, not the decoded struct: the point is that
	// the key is ABSENT from the JSON. Search-capable models reject an explicit
	// temperature, so sending 0.6 — or even null — is a 400.
	if _, present := rawBody["temperature"]; present {
		t.Error("temperature was sent on a web-search request; search-capable models reject it")
	}
}

// The inverse, and the one that would break every other stage if it regressed:
// an ordinary request must still pin its temperature, including zero.
//
// Pinned to a gpt-4 model deliberately. This test guards the *float64/omitempty
// mechanic — that &0.0 survives to the wire — and that mechanic only has
// anything to guard on the families which accept an explicit temperature at
// all. The GPT-5 contract is asserted separately in TestParamStyle*.
func TestOrdinaryRequestStillSendsExplicitZeroTemperature(t *testing.T) {
	var rawBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &rawBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"m","choices":[{"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}],"usage":{}}`))
	}))
	defer server.Close()

	req := testRequest()
	req.Model = "gpt-4o-mini"
	req.Temperature = 0
	if _, err := NewOpenAI("sk_test", WithBaseURL(server.URL)).Complete(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	v, present := rawBody["temperature"]
	if !present {
		t.Fatal("temperature 0 was dropped — omitempty on a *float64 must test nil, not zero")
	}
	if v != float64(0) {
		t.Errorf("temperature = %v, want 0", v)
	}
	if _, present := rawBody["web_search_options"]; present {
		t.Error("an ordinary request asked for web search")
	}
}

func TestWebSearchReturnsDeduplicatedCitations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchResponseBody))
	}))
	defer server.Close()

	req := testRequest()
	req.WebSearch = true
	resp, err := NewOpenAI("sk_test", WithBaseURL(server.URL)).Complete(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Citations) != 2 {
		t.Fatalf("got %d citations, want 2 (the duplicate URL collapsed): %+v", len(resp.Citations), resp.Citations)
	}
	if resp.Citations[0].URL != "https://webflow.com/pricing" {
		t.Errorf("first citation is %q; first-cited order is not preserved", resp.Citations[0].URL)
	}
	if resp.Citations[0].Title != "Webflow Pricing" {
		t.Errorf("titles are dropped: %+v", resp.Citations[0])
	}
}

// A search request answered without sources is the failure mode grounding
// exists to prevent: it is indistinguishable from a grounded answer downstream,
// so it must not be allowed to look like success.
func TestWebSearchWithoutSourcesIsAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"gpt-4o-mini","choices":[{"message":{"role":"assistant","content":"Webflow costs about $14."},"finish_reason":"stop"}],"usage":{}}`))
	}))
	defer server.Close()

	req := testRequest()
	req.WebSearch = true
	_, err := NewOpenAI("sk_test", WithBaseURL(server.URL)).Complete(context.Background(), req)
	if err == nil {
		t.Fatal("a search request that returned no sources was accepted as grounded")
	}
	if !strings.Contains(err.Error(), "from memory") {
		t.Errorf("the error does not say what went wrong: %v", err)
	}
}

func TestWebSearchIsRejectedOnNonOpenAIProviders(t *testing.T) {
	req := testRequest()
	req.WebSearch = true
	_, err := NewGroq("gsk_test", WithBaseURL("http://127.0.0.1:1")).Complete(context.Background(), req)
	if err == nil {
		t.Fatal("groq accepted a web-search request")
	}
	// Rejected before the call, not by a connection error to the dead address.
	if !strings.Contains(err.Error(), "does not support web search") {
		t.Errorf("the request reached the network instead of being refused: %v", err)
	}
}

// The reason WebSearch is a Request field and not a side channel: the cache key
// is a hash of the whole request, so grounding participates automatically and a
// resumed run replays the same sources instead of searching again.
//
// This is load-bearing rather than an optimisation. A pipeline that re-grounded
// on every run would produce a different artifact from the same input, which is
// the one thing its resumable contract promises it will not do.
func TestWebSearchParticipatesInTheCacheKey(t *testing.T) {
	c := NewCache(t.TempDir())
	plain := testRequest()
	grounded := testRequest()
	grounded.WebSearch = true

	if c.Key("openai", plain) == c.Key("openai", grounded) {
		t.Fatal("a grounded and an ungrounded request share a cache key, so one would be served the other's answer")
	}
	// And the same grounded request keys stably, or nothing would ever hit.
	if c.Key("openai", grounded) != c.Key("openai", grounded) {
		t.Fatal("the grounded cache key is not stable")
	}
}

func TestWebSearchCitationsSurviveTheCache(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchResponseBody))
	}))
	defer server.Close()

	p := withCache(NewOpenAI("sk_test", WithBaseURL(server.URL)), NewCache(t.TempDir()))
	req := testRequest()
	req.WebSearch = true

	first, err := p.Complete(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	second, err := p.Complete(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("the provider was called %d times; a repeated grounded request must not search again", calls)
	}
	if !second.FromCache {
		t.Error("the second grounded response did not come from the cache")
	}
	// Citations must round-trip through the cache, or a resumed run can no longer
	// say where any of its facts came from.
	if len(second.Citations) != len(first.Citations) {
		t.Fatalf("citations were lost in the cache: %d then %d", len(first.Citations), len(second.Citations))
	}
	if second.Citations[0].URL != first.Citations[0].URL {
		t.Errorf("cached citation is %q, want %q", second.Citations[0].URL, first.Citations[0].URL)
	}
}
