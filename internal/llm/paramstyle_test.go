package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// These tests exist because the pipeline shipped its entire life on
// gpt-4o-mini, and not by choice: the provider layer sent `max_tokens` and an
// explicit `temperature: 0`, and every GPT-5-class model rejects BOTH with a
// 400. Changing the config string alone broke every run, so nobody changed it,
// so the content was written by the cheapest model available for months.
//
// The wire shape is therefore not an implementation detail here. It is the
// thing that decides which models are reachable at all, and it is asserted
// against the real errors the API returns:
//
//	"Unsupported parameter: 'max_tokens' is not supported with this model.
//	 Use 'max_completion_tokens' instead."
//	"Unsupported value: 'temperature' does not support 0 with this model.
//	 Only the default (1) value is supported."

// captureBody stands up a server that records the JSON body of one request.
func captureBody(t *testing.T) (*httptest.Server, *map[string]any) {
	t.Helper()
	raw := map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &raw)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"m","choices":[{"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}],"usage":{}}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &raw
}

func TestParamStyleGPT5SendsMaxCompletionTokensAndNoTemperature(t *testing.T) {
	for _, model := range []string{"gpt-5-mini", "gpt-5.5", "gpt-5", "o3", "o4-mini"} {
		t.Run(model, func(t *testing.T) {
			srv, raw := captureBody(t)

			req := testRequest()
			req.Model = model
			req.Temperature = 0
			req.MaxTokens = 4096
			if _, err := NewOpenAI("sk_test", WithBaseURL(srv.URL)).Complete(context.Background(), req); err != nil {
				t.Fatal(err)
			}

			if _, present := (*raw)["max_tokens"]; present {
				t.Error("max_tokens was sent — this model rejects it outright and the whole call 400s")
			}
			if got := (*raw)["max_completion_tokens"]; got != float64(4096) {
				t.Errorf("max_completion_tokens = %v, want 4096 — the budget must survive the rename, not be dropped", got)
			}
			if _, present := (*raw)["temperature"]; present {
				t.Error("temperature was sent — these models accept only the default, so any explicit value 400s")
			}
		})
	}
}

func TestParamStyleLegacyModelsKeepTheOldContract(t *testing.T) {
	for _, model := range []string{"gpt-4o-mini", "gpt-4o", "gpt-4.1", "gpt-3.5-turbo"} {
		t.Run(model, func(t *testing.T) {
			srv, raw := captureBody(t)

			req := testRequest()
			req.Model = model
			req.Temperature = 0
			req.MaxTokens = 2048
			if _, err := NewOpenAI("sk_test", WithBaseURL(srv.URL)).Complete(context.Background(), req); err != nil {
				t.Fatal(err)
			}

			if got := (*raw)["max_tokens"]; got != float64(2048) {
				t.Errorf("max_tokens = %v, want 2048", got)
			}
			if _, present := (*raw)["max_completion_tokens"]; present {
				t.Error("max_completion_tokens was sent to a gpt-4-era model, which does not know the parameter")
			}
			if got, present := (*raw)["temperature"]; !present || got != float64(0) {
				t.Errorf("temperature = %v (present=%v), want an explicit 0", got, present)
			}
		})
	}
}

// Groq speaks the classic contract regardless of what its models are called, so
// the provider gates the question before the model name is consulted. Getting
// this wrong would send max_completion_tokens to every llama build at once.
func TestParamStyleNonOpenAIProvidersAreAlwaysClassic(t *testing.T) {
	srv, raw := captureBody(t)

	req := testRequest()
	req.Model = "llama-3.3-70b-versatile"
	req.Temperature = 0.2
	req.MaxTokens = 1024
	if _, err := NewGroq("gsk_test", WithBaseURL(srv.URL)).Complete(context.Background(), req); err != nil {
		t.Fatal(err)
	}

	if got := (*raw)["max_tokens"]; got != float64(1024) {
		t.Errorf("max_tokens = %v, want 1024", got)
	}
	if _, present := (*raw)["max_completion_tokens"]; present {
		t.Error("max_completion_tokens was sent to Groq")
	}
	if got := (*raw)["temperature"]; got != 0.2 {
		t.Errorf("temperature = %v, want 0.2", got)
	}
}

// An unknown model name must get the CURRENT contract, not the old one. This is
// the direction the whole closed-list design exists to protect: a model
// released after this code was written has to work on the day it ships, because
// the alternative is what already happened once — the pipeline frozen on an
// obsolete model by its own provider layer.
func TestParamStyleUnknownModelsDefaultToCurrent(t *testing.T) {
	style := paramStyleFor("openai", "gpt-7-something-that-does-not-exist-yet")
	if !style.maxCompletionTokens {
		t.Error("an unknown OpenAI model was given the pre-GPT-5 contract; unknown must mean current")
	}
	if style.temperature {
		t.Error("an unknown OpenAI model was sent an explicit temperature")
	}
}

func TestReasoningEffortOnlyReachesReasoningModels(t *testing.T) {
	t.Run("sent to a gpt-5 model", func(t *testing.T) {
		srv, raw := captureBody(t)
		req := testRequest()
		req.Model = "gpt-5-mini"
		req.ReasoningEffort = "low"
		if _, err := NewOpenAI("sk_test", WithBaseURL(srv.URL)).Complete(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if got := (*raw)["reasoning_effort"]; got != "low" {
			t.Errorf("reasoning_effort = %v, want \"low\" — this is the pipeline's cost dial and it must reach the wire", got)
		}
	})

	t.Run("dropped for a gpt-4 model", func(t *testing.T) {
		srv, raw := captureBody(t)
		req := testRequest()
		req.Model = "gpt-4o-mini"
		req.ReasoningEffort = "low"
		if _, err := NewOpenAI("sk_test", WithBaseURL(srv.URL)).Complete(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if _, present := (*raw)["reasoning_effort"]; present {
			t.Error("reasoning_effort was sent to a non-reasoning model, which 400s")
		}
	})
}

// The substance stage is the one call that both grounds in real sources AND
// inherits the course's reasoning_effort. Search-capable models take a narrower
// parameter set than the rest — this code already drops temperature for them —
// and a 400 here does not degrade gracefully: it fails the stage a course's
// factual grounding depends on.
func TestWebSearchDropsBothNarrowParameters(t *testing.T) {
	srv, raw := captureBody(t)

	req := testRequest()
	req.Model = "gpt-5-search-api"
	req.Temperature = 0
	req.MaxTokens = 3000
	req.ReasoningEffort = "medium"
	req.WebSearch = true
	// The provider errors on a search reply carrying no citations, which this
	// fake server cannot produce; the request body is what is under test.
	_, _ = NewOpenAI("sk_test", WithBaseURL(srv.URL)).Complete(context.Background(), req)

	if _, present := (*raw)["temperature"]; present {
		t.Error("temperature was sent on a web-search request")
	}
	if _, present := (*raw)["reasoning_effort"]; present {
		t.Error("reasoning_effort was sent on a web-search request")
	}
	// The budget still has to survive, under the name this model accepts.
	if got := (*raw)["max_completion_tokens"]; got != float64(3000) {
		t.Errorf("max_completion_tokens = %v, want 3000", got)
	}
	if _, present := (*raw)["max_tokens"]; present {
		t.Error("max_tokens was sent to a gpt-5-class search model — the exact 400 that broke the substance stage")
	}
}

// Effort changes the answer, so it must change the cache key. Otherwise turning
// thinking up would be served the cheap answer back and look like a no-op.
func TestReasoningEffortParticipatesInTheCacheKey(t *testing.T) {
	c := &Cache{}
	low := testRequest()
	low.ReasoningEffort = "low"
	high := testRequest()
	high.ReasoningEffort = "high"

	if c.Key("openai", low) == c.Key("openai", high) {
		t.Fatal("two different reasoning efforts share a cache key, so raising effort would replay the cheaper answer")
	}
}
