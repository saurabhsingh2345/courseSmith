package llm

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCacheKey(t *testing.T) {
	c := NewCache(t.TempDir())
	base := testRequest()

	key := c.Key("groq", base)
	if key != c.Key("groq", testRequest()) {
		t.Error("identical requests produce different keys")
	}
	if len(key) != 64 {
		t.Errorf("key %q is not a sha256 hex digest", key)
	}

	variants := map[string]Request{}
	for name, mutate := range map[string]func(*Request){
		"different model":       func(r *Request) { r.Model = "other-model" },
		"different prompt":      func(r *Request) { r.Messages[1].Content = "Explain loops." },
		"different system":      func(r *Request) { r.Messages[0].Content = "You are a pirate." },
		"different temperature": func(r *Request) { r.Temperature = 0.7 },
		"different max tokens":  func(r *Request) { r.MaxTokens = 500 },
		"json mode":             func(r *Request) { r.JSONMode = true },
	} {
		req := testRequest()
		mutate(&req)
		variants[name] = req
	}
	for name, req := range variants {
		if c.Key("groq", req) == key {
			t.Errorf("%s: key collision with base request", name)
		}
	}
	if c.Key("openai", base) == key {
		t.Error("same request on different providers must not share a key")
	}
}

func TestCacheRoundTrip(t *testing.T) {
	c := NewCache(filepath.Join(t.TempDir(), "cache"))
	req := testRequest()
	key := c.Key("groq", req)

	if _, ok := c.Get(key); ok {
		t.Fatal("Get() hit on empty cache")
	}

	resp := okResponse("hello")
	resp.FromCache = true // must be stripped when persisted
	if err := c.Put(key, "groq", req, resp); err != nil {
		t.Fatalf("Put(): %v", err)
	}

	got, ok := c.Get(key)
	if !ok {
		t.Fatal("Get() miss after Put()")
	}
	if got.Content != "hello" || got.Model != "fake-model" || got.Usage.TotalTokens != 10 {
		t.Errorf("Get() = %+v", got)
	}
	if got.FromCache {
		t.Error("FromCache leaked into the persisted entry")
	}
}

func TestCacheCorruptEntryIsMiss(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)
	key := c.Key("groq", testRequest())
	if err := os.WriteFile(filepath.Join(dir, key+".json"), []byte("{corrupt"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Get(key); ok {
		t.Error("corrupt entry served as a hit")
	}
}

func TestCachedProviderServesRepeatsWithoutInnerCalls(t *testing.T) {
	inner := &fakeProvider{
		name: "groq",
		fn: func(call int, req Request) (*Response, error) {
			return okResponse("generated"), nil
		},
	}
	p := withCache(inner, NewCache(t.TempDir()))

	first, err := p.Complete(context.Background(), testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if first.FromCache {
		t.Error("first call reported FromCache")
	}

	second, err := p.Complete(context.Background(), testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if !second.FromCache {
		t.Error("second identical call was not served from cache")
	}
	if second.Content != "generated" {
		t.Errorf("cached content = %q", second.Content)
	}
	if inner.calls != 1 {
		t.Errorf("inner called %d times, want 1", inner.calls)
	}

	// A different request must miss.
	other := testRequest()
	other.Messages[1].Content = "Explain functions."
	if _, err := p.Complete(context.Background(), other); err != nil {
		t.Fatal(err)
	}
	if inner.calls != 2 {
		t.Errorf("inner called %d times after distinct request, want 2", inner.calls)
	}
}

func TestCachedProviderDoesNotCacheErrors(t *testing.T) {
	inner := &fakeProvider{
		name: "groq",
		fn: func(call int, req Request) (*Response, error) {
			if call == 1 {
				return nil, errors.New("boom")
			}
			return okResponse("recovered"), nil
		},
	}
	p := withCache(inner, NewCache(t.TempDir()))

	if _, err := p.Complete(context.Background(), testRequest()); err == nil {
		t.Fatal("first call should fail")
	}
	resp, err := p.Complete(context.Background(), testRequest())
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if resp.FromCache {
		t.Error("error was cached")
	}
	if inner.calls != 2 {
		t.Errorf("inner called %d times, want 2", inner.calls)
	}
}
