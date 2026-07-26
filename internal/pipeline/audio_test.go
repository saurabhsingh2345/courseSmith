package pipeline

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/enfec/coursesmith/internal/project"
)

// makeWAV builds a valid silent PCM16 mono WAV of the given duration.
func makeWAV(seconds float64) []byte {
	const sampleRate = 8000
	const byteRate = sampleRate * 2 // 16-bit mono
	n := int(math.Round(seconds*byteRate/2)) * 2

	buf := make([]byte, 0, 44+n)
	le := binary.LittleEndian
	u32 := func(v uint32) []byte { b := make([]byte, 4); le.PutUint32(b, v); return b }
	u16 := func(v uint16) []byte { b := make([]byte, 2); le.PutUint16(b, v); return b }

	buf = append(buf, "RIFF"...)
	buf = append(buf, u32(uint32(36+n))...)
	buf = append(buf, "WAVE"...)
	buf = append(buf, "fmt "...)
	buf = append(buf, u32(16)...)
	buf = append(buf, u16(1)...) // PCM
	buf = append(buf, u16(1)...) // mono
	buf = append(buf, u32(sampleRate)...)
	buf = append(buf, u32(byteRate)...)
	buf = append(buf, u16(2)...)  // block align
	buf = append(buf, u16(16)...) // bits per sample
	buf = append(buf, "data"...)
	buf = append(buf, u32(uint32(n))...)
	buf = append(buf, make([]byte, n)...) // silence
	return buf
}

func requireFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
}

func TestWavDuration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.wav")
	if err := os.WriteFile(path, makeWAV(2.5), 0o644); err != nil {
		t.Fatal(err)
	}
	dur, err := wavDuration(path)
	if err != nil {
		t.Fatal(err)
	}
	if diff := (dur - 2500*time.Millisecond).Abs(); diff > 10*time.Millisecond {
		t.Errorf("duration = %v, want ~2.5s", dur)
	}

	bad := filepath.Join(t.TempDir(), "bad.wav")
	if err := os.WriteFile(bad, []byte("not a wav"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := wavDuration(bad); err == nil {
		t.Error("wavDuration accepted a non-WAV file")
	}
}

func TestAudioStageSynthesizesAndConcats(t *testing.T) {
	requireFFmpeg(t)
	course, lesson := testCourse(t)
	seedScript(t, lesson) // two sections

	var requests []map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audio/speech" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding TTS request: %v", err)
		}
		requests = append(requests, body)
		_, _ = w.Write(makeWAV(0.5))
	}))
	defer server.Close()

	env, out := runEnv(t, &fakeRouter{})
	env.TTSBaseURL = server.URL + "/v1"

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageAudio}); err != nil {
		t.Fatal(err)
	}

	if len(requests) != 2 {
		t.Fatalf("TTS requests = %d, want 2 (one per section)", len(requests))
	}
	if requests[0]["voice"] != "af_heart" || requests[0]["model"] != "kokoro" {
		t.Errorf("TTS request = %+v", requests[0])
	}
	if !strings.Contains(requests[0]["input"], "Python reads code line by line.") {
		t.Errorf("first section narration not sent: %+v", requests[0])
	}

	for _, name := range []string{"first-idea.wav", "second-idea.wav"} {
		if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), AudioDirName, name)); err != nil {
			t.Errorf("per-section wav missing: %v", err)
		}
	}
	voiceover := filepath.Join(lesson.GeneratedDir(), VoiceoverFileName)
	dur, err := wavDuration(voiceover)
	if err != nil {
		t.Fatalf("voiceover.wav: %v", err)
	}
	// 2 × 0.5s sections + 0.7s section pause − 2 × 50ms crossfade overlap.
	want := 1600 * time.Millisecond
	if diff := (dur - want).Abs(); diff > 100*time.Millisecond {
		t.Errorf("voiceover duration = %v, want ~%v", dur, want)
	}
	if !strings.Contains(out.String(), VoiceoverFileName) {
		t.Errorf("output missing result line:\n%s", out.String())
	}

	// The spoken-text artifact records what was sent to the TTS.
	var tts TTSScript
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), TTSScriptFileName))
	if err != nil {
		t.Fatalf("tts_script.json missing: %v", err)
	}
	if err := json.Unmarshal(data, &tts); err != nil {
		t.Fatal(err)
	}
	if len(tts.Sections) != 2 || tts.Sections[0].ID != "first-idea" || len(tts.Sections[0].Paragraphs) == 0 {
		t.Errorf("tts_script = %+v", tts)
	}

	// Mastering report with before/after loudness persisted.
	if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), ReviewsDirName, LoudnessFileName)); err != nil {
		t.Errorf("loudness report missing: %v", err)
	}
}

func TestAudioStageAppliesSpeechPrep(t *testing.T) {
	requireFFmpeg(t)
	course, lesson := testCourse(t)
	if err := os.MkdirAll(lesson.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lesson.GeneratedDir(), ScriptFileName),
		[]byte(scriptJSON("Python 3.11 calls __init__ for you.")), 0o644); err != nil {
		t.Fatal(err)
	}

	var inputs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		inputs = append(inputs, body["input"])
		_, _ = w.Write(makeWAV(0.5))
	}))
	defer server.Close()

	env, _ := runEnv(t, &fakeRouter{})
	env.TTSBaseURL = server.URL + "/v1"
	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageAudio}); err != nil {
		t.Fatal(err)
	}
	if len(inputs) == 0 || inputs[0] != "Python three point eleven calls dunder init for you." {
		t.Errorf("TTS input not prepped: %q", inputs)
	}
}

func TestAudioStageAppliesTTSFixes(t *testing.T) {
	requireFFmpeg(t)
	course, lesson := testCourse(t)
	seedScript(t, lesson) // narration mentions "Python"
	if err := os.WriteFile(filepath.Join(lesson.GeneratedDir(), TTSFixesFileName),
		[]byte(`{"Python": "pie thon"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var inputs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		inputs = append(inputs, body["input"])
		_, _ = w.Write(makeWAV(0.5))
	}))
	defer server.Close()

	env, out := runEnv(t, &fakeRouter{})
	env.TTSBaseURL = server.URL + "/v1"
	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageAudio}); err != nil {
		t.Fatal(err)
	}
	if len(inputs) == 0 || !strings.Contains(inputs[0], "pie thon") {
		t.Errorf("tts fix not applied: %q", inputs)
	}
	if !strings.Contains(out.String(), "pronunciation fix") {
		t.Errorf("output missing fix notice:\n%s", out.String())
	}
}

func TestAudioStageUnreachableServer(t *testing.T) {
	requireFFmpeg(t)
	course, lesson := testCourse(t)
	seedScript(t, lesson)

	env, _ := runEnv(t, &fakeRouter{})
	env.TTSBaseURL = "http://127.0.0.1:1/v1" // nothing listens on port 1
	env.HTTPClient = &http.Client{Timeout: 2 * time.Second}

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageAudio})
	if err == nil || !strings.Contains(err.Error(), "docker run") {
		t.Errorf("error = %v, want Kokoro start instructions", err)
	}
}

func TestAudioStageRequiresScript(t *testing.T) {
	requireFFmpeg(t)
	course, lesson := testCourse(t)
	env, _ := runEnv(t, &fakeRouter{})
	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageAudio})
	if err == nil || !strings.Contains(err.Error(), "script stage must run first") {
		t.Errorf("error = %v, want script-first error", err)
	}
}
