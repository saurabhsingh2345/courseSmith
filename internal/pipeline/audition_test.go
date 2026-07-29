package pipeline

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

func TestFilterVoicesForLanguage(t *testing.T) {
	voices := []string{"af_heart", "am_adam", "bf_emma", "ef_dora", "jf_alpha"}
	tests := []struct {
		lang string
		want []string
	}{
		{"en", []string{"af_heart", "am_adam", "bf_emma"}},
		{"en-US", []string{"af_heart", "am_adam", "bf_emma"}},
		{"es", []string{"ef_dora"}},
		{"ja", []string{"jf_alpha"}},
		{"xx", voices}, // unknown language keeps everything
	}
	for _, tt := range tests {
		got := filterVoicesForLanguage(voices, tt.lang)
		if strings.Join(got, ",") != strings.Join(tt.want, ",") {
			t.Errorf("filterVoicesForLanguage(%q) = %v, want %v", tt.lang, got, tt.want)
		}
	}
}

func TestRunAuditionRendersPageAndSkipsExisting(t *testing.T) {
	course, _ := testCourse(t)

	var speechCalls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/audio/voices":
			_ = json.NewEncoder(w).Encode(map[string]any{"voices": []map[string]string{
				{"id": "af_heart"}, {"id": "am_adam"}, {"id": "jf_alpha"},
			}})
		case "/v1/audio/speech":
			// map[string]any, not map[string]string: the request carries a
			// numeric "speed", and one number fails the whole decode.
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			voice, _ := body["voice"].(string)
			speechCalls = append(speechCalls, voice)
			_, _ = w.Write(makeWAV(0.2))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	env, _ := runEnv(t, &fakeRouter{})
	env.TTSBaseURL = server.URL + "/v1"
	cfg := config.Resolve(course.Config, config.Config{}, config.Config{})

	index, err := RunAudition(context.Background(), env, course, cfg)
	if err != nil {
		t.Fatal(err)
	}
	// jf_alpha is Japanese; the en course must skip it.
	if strings.Join(speechCalls, ",") != "af_heart,am_adam" {
		t.Errorf("synthesized voices = %v", speechCalls)
	}
	page, err := os.ReadFile(index)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "af_heart.wav") || !strings.Contains(string(page), "am_adam.wav") {
		t.Errorf("audition page missing players:\n%s", page)
	}

	// Re-run: recordings exist, so no new speech calls.
	speechCalls = nil
	if _, err := RunAudition(context.Background(), env, course, cfg); err != nil {
		t.Fatal(err)
	}
	if len(speechCalls) != 0 {
		t.Errorf("re-run re-synthesized: %v", speechCalls)
	}
}

func TestChooseVoice(t *testing.T) {
	course, _ := testCourse(t)
	if err := ChooseVoice(course, "am_adam"); err != nil {
		t.Fatal(err)
	}
	reloaded, err := project.LoadCourse(course.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Config.Style.Voice != "am_adam" {
		t.Errorf("voice = %q, want am_adam", reloaded.Config.Style.Voice)
	}
	// The rest of the file survives the edit.
	data, err := os.ReadFile(filepath.Join(course.Dir, project.CourseFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "warm teacher") {
		t.Errorf("course.yaml lost content:\n%s", data)
	}
}
