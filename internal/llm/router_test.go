package llm

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

func TestParseModelRef(t *testing.T) {
	tests := []struct {
		ref          string
		wantProvider string
		wantModel    string
		wantErr      bool
	}{
		{ref: "groq/llama-3.3-70b-versatile", wantProvider: "groq", wantModel: "llama-3.3-70b-versatile"},
		{ref: "openai/gpt-4o-mini", wantProvider: "openai", wantModel: "gpt-4o-mini"},
		{ref: "groq/meta-llama/llama-4-scout", wantProvider: "groq", wantModel: "meta-llama/llama-4-scout"},
		{ref: "no-slash", wantErr: true},
		{ref: "/model-only", wantErr: true},
		{ref: "provider/", wantErr: true},
		{ref: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			provider, model, err := ParseModelRef(tt.ref)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseModelRef(%q) = %q/%q, want error", tt.ref, provider, model)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if provider != tt.wantProvider || model != tt.wantModel {
				t.Errorf("ParseModelRef(%q) = %q/%q, want %q/%q", tt.ref, provider, model, tt.wantProvider, tt.wantModel)
			}
		})
	}
}

func TestModelFor(t *testing.T) {
	pcfg := config.Defaults().Pipeline

	ref, err := ModelFor(pcfg, TaskContent)
	if err != nil || ref != "openai/gpt-5-mini" {
		t.Errorf("ModelFor(content) = %q, %v", ref, err)
	}
	ref, err = ModelFor(pcfg, TaskReview)
	if err != nil || ref != "openai/gpt-5-mini" {
		t.Errorf("ModelFor(review) = %q, %v", ref, err)
	}
	if _, err := ModelFor(config.Pipeline{}, TaskContent); err == nil {
		t.Error("ModelFor with empty config should error")
	}
	if _, err := ModelFor(pcfg, TaskType("nonsense")); err == nil {
		t.Error("ModelFor with unknown task should error")
	}
}

// newTestRouter returns a router whose base providers are fakes, keyed by
// provider name, with all state in a temp dir.
func newTestRouter(t *testing.T, fakes map[string]*fakeProvider) *Router {
	t.Helper()
	r := NewRouter(t.TempDir())
	built := map[string]int{}
	r.newProvider = func(name string) (Provider, error) {
		built[name]++
		if built[name] > 1 {
			t.Errorf("provider %q built %d times, want memoized", name, built[name])
		}
		f, ok := fakes[name]
		if !ok {
			return nil, fmt.Errorf("unknown LLM provider %q", name)
		}
		return f, nil
	}
	return r
}

func TestRouterRoutesByTask(t *testing.T) {
	var groqModel, openaiModel string
	fakes := map[string]*fakeProvider{
		"groq": {name: "groq", fn: func(_ int, req Request) (*Response, error) {
			groqModel = req.Model
			return okResponse("from groq"), nil
		}},
		"openai": {name: "openai", fn: func(_ int, req Request) (*Response, error) {
			openaiModel = req.Model
			return okResponse("from openai"), nil
		}},
	}
	r := newTestRouter(t, fakes)
	// Two DIFFERENT providers, set explicitly rather than taken from the
	// defaults. This test is about the router sending each task to the model
	// its config names, and borrowing the defaults made it silently stop
	// proving that the moment content and review came to share a model.
	pcfg := config.Pipeline{
		LLMContent: "groq/llama-3.3-70b-versatile",
		LLMReview:  "openai/gpt-4o-mini",
	}
	ctx := context.Background()

	req := testRequest()
	req.Model = "" // router must fill this in

	resp, err := r.Complete(ctx, pcfg, TaskContent, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "from groq" || groqModel != "llama-3.3-70b-versatile" {
		t.Errorf("content task: resp = %q, model = %q", resp.Content, groqModel)
	}

	resp, err = r.Complete(ctx, pcfg, TaskReview, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "from openai" || openaiModel != "gpt-4o-mini" {
		t.Errorf("review task: resp = %q, model = %q", resp.Content, openaiModel)
	}
}

func TestRouterCachesAcrossCalls(t *testing.T) {
	fakes := map[string]*fakeProvider{
		"groq": {name: "groq", fn: func(_ int, req Request) (*Response, error) {
			return okResponse("expensive"), nil
		}},
	}
	r := newTestRouter(t, fakes)
	// Pinned, not defaulted: this test is about the cache, and only the groq
	// fake is registered. Reading the model out of Defaults() coupled it to a
	// choice that has nothing to do with what it checks.
	pcfg := config.Pipeline{LLMContent: "groq/llama-3.3-70b-versatile"}
	ctx := context.Background()

	if _, err := r.Complete(ctx, pcfg, TaskContent, testRequest()); err != nil {
		t.Fatal(err)
	}
	resp, err := r.Complete(ctx, pcfg, TaskContent, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if !resp.FromCache {
		t.Error("second identical request not served from cache")
	}
	if fakes["groq"].calls != 1 {
		t.Errorf("base provider called %d times, want 1", fakes["groq"].calls)
	}
}

func TestRouterUnknownProvider(t *testing.T) {
	r := newTestRouter(t, nil)
	pcfg := config.Pipeline{LLMContent: "anthropic/claude"}
	_, err := r.Complete(context.Background(), pcfg, TaskContent, testRequest())
	if err == nil || !strings.Contains(err.Error(), "unknown LLM provider") {
		t.Errorf("error = %v, want unknown-provider error", err)
	}
}

func TestRouterMissingKeyErrorIsActionable(t *testing.T) {
	t.Setenv(EnvGroqKey, "")
	r := NewRouter(t.TempDir()) // real env-based construction
	// Pinned to groq because that is the key this asserts the message names.
	// The point is that a missing key says which one and how to get it, which
	// is provider-specific, so the test names the provider rather than
	// inheriting whichever one the defaults happen to prefer.
	pcfg := config.Pipeline{LLMContent: "groq/llama-3.3-70b-versatile"}
	_, err := r.Complete(context.Background(), pcfg, TaskContent, testRequest())
	if err == nil || !strings.Contains(err.Error(), EnvGroqKey) {
		t.Errorf("error = %v, want actionable %s message", err, EnvGroqKey)
	}
}
