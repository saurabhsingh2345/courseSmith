package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// TranscriptSegment is one timed span of transcribed speech.
type TranscriptSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// Transcription is the result of a speech-to-text request.
type Transcription struct {
	Text     string              `json:"text"`
	Segments []TranscriptSegment `json:"segments"`
}

// GroqTranscriber calls Groq's hosted Whisper via the OpenAI-compatible
// audio transcriptions endpoint.
type GroqTranscriber struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewGroqTranscriber returns a transcriber for Groq's API. baseURL "" uses
// the production endpoint; client nil uses a default with a generous timeout
// (audio uploads are slow).
func NewGroqTranscriber(apiKey, baseURL string, client *http.Client) *GroqTranscriber {
	if baseURL == "" {
		baseURL = GroqBaseURL
	}
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}
	return &GroqTranscriber{baseURL: baseURL, apiKey: apiKey, client: client}
}

// Transcribe uploads an audio file and returns the timed transcription
// (response_format=verbose_json, which carries segments).
func (t *GroqTranscriber) Transcribe(ctx context.Context, audioPath, model string) (*Transcription, error) {
	f, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", audioPath, err)
	}
	defer f.Close()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("building transcription request: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("reading %s: %w", audioPath, err)
	}
	for field, value := range map[string]string{
		"model":           model,
		"response_format": "verbose_json",
	} {
		if err := mw.WriteField(field, value); err != nil {
			return nil, fmt.Errorf("building transcription request: %w", err)
		}
	}
	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("building transcription request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/audio/transcriptions", &body)
	if err != nil {
		return nil, fmt.Errorf("building transcription request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling groq transcription API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return nil, fmt.Errorf("reading transcription response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		msg := ""
		var parsed apiErrorBody
		if json.Unmarshal(respBody, &parsed) == nil && parsed.Error.Message != "" {
			msg = parsed.Error.Message
		} else {
			msg = truncate(string(respBody), 300)
		}
		return nil, &APIError{Provider: "groq", StatusCode: resp.StatusCode, Message: msg}
	}

	// verbose_json carries many extra fields; parse leniently.
	var out Transcription
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("parsing transcription response: %w", err)
	}
	if out.Text == "" && len(out.Segments) == 0 {
		return nil, fmt.Errorf("transcription came back empty (is %s a valid audio file?)", audioPath)
	}
	return &out, nil
}
