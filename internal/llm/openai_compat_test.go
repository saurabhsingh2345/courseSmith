package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPProviderSendsWellFormedRequest(t *testing.T) {
	var got chatRequest
	var gotAuth, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"model": "llama-3.3-70b-versatile",
			"choices": [{"message": {"role": "assistant", "content": "Hello!"}, "finish_reason": "stop"}],
			"usage": {"prompt_tokens": 20, "completion_tokens": 5, "total_tokens": 25}
		}`))
	}))
	defer server.Close()

	p := NewGroq("gsk_test", WithBaseURL(server.URL))
	req := testRequest()
	req.Temperature = 0.3
	req.MaxTokens = 800
	req.JSONMode = true

	resp, err := p.Complete(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	if gotPath != "/chat/completions" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer gsk_test" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if got.Model != "test-model" || len(got.Messages) != 2 {
		t.Errorf("wire request = %+v", got)
	}
	if got.Temperature == nil || *got.Temperature != 0.3 {
		t.Errorf("temperature = %v, want explicit 0.3", got.Temperature)
	}
	if got.MaxTokens != 800 {
		t.Errorf("max_tokens = %d", got.MaxTokens)
	}
	if got.ResponseFormat == nil || got.ResponseFormat.Type != "json_object" {
		t.Errorf("response_format = %+v, want json_object", got.ResponseFormat)
	}

	if resp.Content != "Hello!" || resp.Model != "llama-3.3-70b-versatile" || resp.Usage.TotalTokens != 25 {
		t.Errorf("response = %+v", resp)
	}
	if resp.FromCache {
		t.Error("fresh HTTP response claims FromCache")
	}
}

func TestHTTPProviderAPIErrors(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		body          string
		retryAfter    string
		wantMsg       string
		wantRetryable bool
		wantDelay     time.Duration
	}{
		{
			name:          "rate limited with retry-after",
			status:        429,
			body:          `{"error": {"message": "Rate limit reached for llama-3.3-70b-versatile"}}`,
			retryAfter:    "12",
			wantMsg:       "Rate limit reached",
			wantRetryable: true,
			wantDelay:     12 * time.Second,
		},
		{
			name:          "server error",
			status:        503,
			body:          `{"error": {"message": "overloaded"}}`,
			wantMsg:       "overloaded",
			wantRetryable: true,
		},
		{
			name:          "bad request not retryable",
			status:        400,
			body:          `{"error": {"message": "unknown model"}}`,
			wantMsg:       "unknown model",
			wantRetryable: false,
		},
		{
			name:          "non-json error body still surfaces",
			status:        502,
			body:          "Bad Gateway",
			wantMsg:       "Bad Gateway",
			wantRetryable: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.retryAfter != "" {
					w.Header().Set("Retry-After", tt.retryAfter)
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			p := NewOpenAI("sk-test", WithBaseURL(server.URL))
			_, err := p.Complete(context.Background(), testRequest())

			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("error = %v, want *APIError", err)
			}
			if apiErr.StatusCode != tt.status || apiErr.Provider != "openai" {
				t.Errorf("APIError = %+v", apiErr)
			}
			if !strings.Contains(apiErr.Message, tt.wantMsg) {
				t.Errorf("message = %q, want it to contain %q", apiErr.Message, tt.wantMsg)
			}
			if apiErr.Retryable() != tt.wantRetryable {
				t.Errorf("Retryable() = %v, want %v", apiErr.Retryable(), tt.wantRetryable)
			}
			if apiErr.RetryAfter != tt.wantDelay {
				t.Errorf("RetryAfter = %v, want %v", apiErr.RetryAfter, tt.wantDelay)
			}
		})
	}
}

func TestHTTPProviderVisionMessages(t *testing.T) {
	var raw map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		_, _ = w.Write([]byte(`{"choices": [{"message": {"role": "assistant", "content": "seen"}, "finish_reason": "stop"}]}`))
	}))
	defer server.Close()

	p := NewOpenAI("sk-test", WithBaseURL(server.URL))
	req := testRequest()
	req.Images = []string{"aGVsbG8="} // base64 "hello"
	if _, err := p.Complete(context.Background(), req); err != nil {
		t.Fatal(err)
	}

	messages := raw["messages"].([]any)
	system := messages[0].(map[string]any)
	if _, isString := system["content"].(string); !isString {
		t.Errorf("system content should stay a plain string: %v", system["content"])
	}
	user := messages[1].(map[string]any)
	parts, ok := user["content"].([]any)
	if !ok || len(parts) != 2 {
		t.Fatalf("user content = %v, want [text, image_url] parts", user["content"])
	}
	text := parts[0].(map[string]any)
	if text["type"] != "text" || text["text"] != "Explain variables." {
		t.Errorf("text part = %v", text)
	}
	img := parts[1].(map[string]any)
	url := img["image_url"].(map[string]any)["url"].(string)
	if img["type"] != "image_url" || url != "data:image/png;base64,aGVsbG8=" {
		t.Errorf("image part = %v", img)
	}
}

func TestHTTPProviderTruncatedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices": [{"message": {"role": "assistant", "content": "cut off"}, "finish_reason": "length"}]}`))
	}))
	defer server.Close()

	p := NewGroq("gsk_test", WithBaseURL(server.URL))
	req := testRequest()
	req.MaxTokens = 50
	_, err := p.Complete(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "truncated") {
		t.Errorf("error = %v, want truncation error", err)
	}
}

func TestHTTPProviderEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices": []}`))
	}))
	defer server.Close()

	p := NewGroq("gsk_test", WithBaseURL(server.URL))
	_, err := p.Complete(context.Background(), testRequest())
	if err == nil || !strings.Contains(err.Error(), "no choices") {
		t.Errorf("error = %v, want no-choices error", err)
	}
}
