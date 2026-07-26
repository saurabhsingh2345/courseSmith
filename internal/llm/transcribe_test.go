package llm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGroqTranscriber(t *testing.T) {
	audioPath := filepath.Join(t.TempDir(), "voiceover.wav")
	if err := os.WriteFile(audioPath, []byte("RIFFfakewavdata"), 0o644); err != nil {
		t.Fatal(err)
	}

	var gotPath, gotAuth, gotModel, gotFormat, gotFileName string
	var gotFileBytes int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("parsing multipart: %v", err)
		}
		gotModel = r.FormValue("model")
		gotFormat = r.FormValue("response_format")
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Errorf("file part missing: %v", err)
		} else {
			defer file.Close()
			gotFileName = header.Filename
			buf := make([]byte, 64)
			n, _ := file.Read(buf)
			gotFileBytes = n
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"task": "transcribe", "language": "en", "duration": 3.2,
			"text": "Hello world.",
			"segments": [
				{"id": 0, "seek": 0, "start": 0.0, "end": 1.5, "text": " Hello", "tokens": [1], "temperature": 0},
				{"id": 1, "seek": 0, "start": 1.5, "end": 3.2, "text": " world.", "tokens": [2], "temperature": 0}
			]
		}`))
	}))
	defer server.Close()

	tr := NewGroqTranscriber("gsk_test", server.URL, nil)
	got, err := tr.Transcribe(context.Background(), audioPath, "whisper-large-v3")
	if err != nil {
		t.Fatal(err)
	}

	if gotPath != "/audio/transcriptions" || gotAuth != "Bearer gsk_test" {
		t.Errorf("path = %q, auth = %q", gotPath, gotAuth)
	}
	if gotModel != "whisper-large-v3" || gotFormat != "verbose_json" {
		t.Errorf("model = %q, response_format = %q", gotModel, gotFormat)
	}
	if gotFileName != "voiceover.wav" || gotFileBytes == 0 {
		t.Errorf("uploaded file = %q (%d bytes read)", gotFileName, gotFileBytes)
	}
	if got.Text != "Hello world." || len(got.Segments) != 2 {
		t.Errorf("transcription = %+v", got)
	}
	if got.Segments[1].Start != 1.5 || got.Segments[1].End != 3.2 {
		t.Errorf("segment timing = %+v", got.Segments[1])
	}
}

func TestGroqTranscriberAPIError(t *testing.T) {
	audioPath := filepath.Join(t.TempDir(), "voiceover.wav")
	if err := os.WriteFile(audioPath, []byte("RIFF"), 0o644); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"error": {"message": "rate limit for whisper-large-v3"}}`))
	}))
	defer server.Close()

	tr := NewGroqTranscriber("gsk_test", server.URL, nil)
	_, err := tr.Transcribe(context.Background(), audioPath, "whisper-large-v3")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 429 {
		t.Fatalf("error = %v, want 429 *APIError", err)
	}
	if !strings.Contains(apiErr.Message, "rate limit") {
		t.Errorf("message = %q", apiErr.Message)
	}
}

func TestGroqTranscriberMissingFile(t *testing.T) {
	tr := NewGroqTranscriber("gsk_test", "http://unused.invalid", nil)
	_, err := tr.Transcribe(context.Background(), filepath.Join(t.TempDir(), "nope.wav"), "whisper-large-v3")
	if err == nil || !strings.Contains(err.Error(), "nope.wav") {
		t.Errorf("error = %v, want mention of the missing file", err)
	}
}
