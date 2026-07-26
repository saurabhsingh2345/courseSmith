package llm

import (
	"context"
	"strings"
	"testing"
)

// fakeProvider is a scriptable Provider for wrapper tests.
type fakeProvider struct {
	name  string
	calls int
	fn    func(call int, req Request) (*Response, error)
}

func (f *fakeProvider) Name() string { return f.name }

func (f *fakeProvider) Complete(_ context.Context, req Request) (*Response, error) {
	f.calls++
	return f.fn(f.calls, req)
}

func okResponse(content string) *Response {
	return &Response{Content: content, Model: "fake-model", Usage: Usage{TotalTokens: 10}}
}

func testRequest() Request {
	return Request{
		Model: "test-model",
		Messages: []Message{
			{Role: RoleSystem, Content: "You are a teacher."},
			{Role: RoleUser, Content: "Explain variables."},
		},
	}
}

func TestRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Request)
		wantErr string
	}{
		{name: "valid", mutate: func(r *Request) {}},
		{name: "no model", mutate: func(r *Request) { r.Model = "" }, wantErr: "no model"},
		{name: "no messages", mutate: func(r *Request) { r.Messages = nil }, wantErr: "no messages"},
		{
			name:    "empty message content",
			mutate:  func(r *Request) { r.Messages[1].Content = "" },
			wantErr: "message 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := testRequest()
			tt.mutate(&req)
			err := req.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestFromEnvMissingKeys(t *testing.T) {
	t.Setenv(EnvGroqKey, "")
	t.Setenv(EnvOpenAIKey, "")

	if _, err := GroqFromEnv(); err == nil || !strings.Contains(err.Error(), EnvGroqKey) {
		t.Errorf("GroqFromEnv() error = %v, want mention of %s", err, EnvGroqKey)
	}
	if _, err := OpenAIFromEnv(); err == nil || !strings.Contains(err.Error(), EnvOpenAIKey) {
		t.Errorf("OpenAIFromEnv() error = %v, want mention of %s", err, EnvOpenAIKey)
	}

	t.Setenv(EnvGroqKey, "gsk_test")
	p, err := GroqFromEnv()
	if err != nil {
		t.Fatalf("GroqFromEnv() with key set: %v", err)
	}
	if p.Name() != "groq" {
		t.Errorf("Name() = %q, want groq", p.Name())
	}
}
