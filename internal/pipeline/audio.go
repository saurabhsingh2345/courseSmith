package pipeline

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// Stage 5 outputs, under the lesson's generated dir.
const (
	VoiceoverFileName = "voiceover.wav"
	AudioDirName      = "audio" // per-section wavs, reused by the video stage
)

// DefaultTTSBaseURL is the local Kokoro-FastAPI server's OpenAI-compatible root.
const DefaultTTSBaseURL = "http://localhost:8880/v1"

// ttsStartHelp tells the user how to get a Kokoro server running.
const ttsStartHelp = "start it with:\n" +
	"      docker run -p 8880:8880 ghcr.io/remsky/kokoro-fastapi-cpu:latest\n" +
	"    (or the kokoro-onnx pip alternative — see README), or point KOKORO_URL at a running server"

func (e *Env) ttsBaseURL() string {
	if e.TTSBaseURL != "" {
		return strings.TrimSuffix(e.TTSBaseURL, "/")
	}
	return DefaultTTSBaseURL
}

func (e *Env) httpClient() *http.Client {
	if e.HTTPClient != nil {
		return e.HTTPClient
	}
	return &http.Client{Timeout: 10 * time.Minute} // long narration synthesis is slow on CPU
}

// ttsSpeak synthesizes one piece of narration to WAV bytes via the Kokoro
// server's OpenAI-compatible /audio/speech endpoint. speed is the speaking
// rate multiplier; values <= 0 and 1 mean "server default".
func (e *Env) ttsSpeak(ctx context.Context, voice, text string, speed float64) ([]byte, error) {
	reqBody := map[string]any{
		"model":           "kokoro",
		"voice":           voice,
		"input":           text,
		"response_format": "wav",
	}
	if speed > 0 && speed != 1 {
		reqBody["speed"] = speed
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("encoding TTS request: %w", err)
	}
	url := e.ttsBaseURL() + "/audio/speech"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("building TTS request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient().Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("cannot reach the Kokoro TTS server at %s — %s\n    (%v)", e.ttsBaseURL(), ttsStartHelp, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 500<<20))
	if err != nil {
		return nil, fmt.Errorf("reading TTS response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Kokoro server returned HTTP %d: %s", resp.StatusCode, tailLines(string(body), 3))
	}
	if !bytes.HasPrefix(body, []byte("RIFF")) {
		return nil, fmt.Errorf("Kokoro server at %s returned %d bytes that are not WAV audio", e.ttsBaseURL(), len(body))
	}
	return body, nil
}

// runAudioStage is Stage 5: script.json → prepped spoken text → per-section
// WAVs (paragraph pauses baked in) → crossfaded, mastered voiceover.wav.
//
// Post-production, in order:
//  1. Narration is rewritten for speech (tts_prep.go) and synthesized
//     per paragraph; paragraphs join with cfg.Audio.ParagraphPauseMs of
//     silence and a crossfade at every seam.
//  2. Section WAVs join with cfg.Audio.SectionPauseMs of silence, again
//     crossfaded, so no join is audible.
//  3. The joined narration is mastered (see masterChain) and loudness-
//     normalized to cfg.Audio.TargetLUFS with two-pass ffmpeg loudnorm.
//  4. Optionally a music bed from courses/<slug>/assets/music/ is ducked
//     underneath via sidechain compression.
func runAudioStage(ctx context.Context, e *Env, course *project.Course, l *project.Lesson, cfg config.Config) error {
	if _, err := e.CheckFFmpeg(); err != nil {
		return err
	}
	script, err := loadScript(l)
	if err != nil {
		return err
	}
	audioDir := filepath.Join(l.GeneratedDir(), AudioDirName)
	if err := os.MkdirAll(audioDir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", audioDir, err)
	}

	dict := SpeechDict(cfg.Style.Pronunciations)
	fixes, err := loadTTSFixes(l)
	if err != nil {
		return err
	}
	maps.Copy(dict, fixes)

	// Effective speaking rate: the user's voice_speed times the auto-pace
	// correction the align stage may have written.
	speed := cfg.Style.VoiceSpeed
	if speed <= 0 {
		speed = 1
	}
	if fix, err := loadTTSSpeedFix(l); err != nil {
		return err
	} else if fix != nil && fix.Speed > 0 {
		speed *= fix.Speed
	}

	fmt.Fprintf(e.out(), "  → audio     synthesizing %d section(s) with Kokoro (voice %s, speed %.2f)...\n", len(script.Sections), cfg.Style.Voice, speed)
	if len(fixes) > 0 {
		fmt.Fprintf(e.out(), "    applying %d pronunciation fix(es) from %s\n", len(fixes), TTSFixesFileName)
	}

	ttsScript := TTSScript{}
	wavPaths := make([]string, 0, len(script.Sections))
	var total time.Duration
	for _, sec := range script.Sections {
		if err := ctx.Err(); err != nil {
			return err
		}
		spoken := SpokenSection{ID: sec.ID}
		var paraWavs []string
		for pi, para := range SplitParagraphs(sec.Narration) {
			prepped := PrepForSpeech(para, dict)
			spoken.Paragraphs = append(spoken.Paragraphs, prepped)
			wav, err := e.ttsSpeak(ctx, cfg.Style.Voice, prepped, speed)
			if err != nil {
				return fmt.Errorf("section %q: %w", sec.ID, err)
			}
			paraPath := filepath.Join(audioDir, fmt.Sprintf("%s.p%d.wav", sec.ID, pi))
			if err := writeFileAtomic(paraPath, wav); err != nil {
				return err
			}
			paraWavs = append(paraWavs, paraPath)
		}
		sectionPath := filepath.Join(audioDir, sec.ID+".wav")
		if err := joinWavs(ctx, e, paraWavs, cfg.Audio.ParagraphPauseMs, cfg.Audio.CrossfadeMs, sectionPath); err != nil {
			return fmt.Errorf("section %q: %w", sec.ID, err)
		}
		for _, p := range paraWavs {
			os.Remove(p)
		}
		dur, err := wavDuration(sectionPath)
		if err != nil {
			return fmt.Errorf("section %q: %w", sec.ID, err)
		}
		total += dur
		wavPaths = append(wavPaths, sectionPath)
		ttsScript.Sections = append(ttsScript.Sections, spoken)
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), TTSScriptFileName), ttsScript); err != nil {
		return err
	}

	// Join sections and master.
	raw := filepath.Join(audioDir, "voiceover.raw.wav")
	if err := joinWavs(ctx, e, wavPaths, cfg.Audio.SectionPauseMs, cfg.Audio.CrossfadeMs, raw); err != nil {
		return err
	}
	defer os.Remove(raw)

	voiceover := filepath.Join(l.GeneratedDir(), VoiceoverFileName)
	report, err := masterVoiceover(ctx, e, raw, voiceover, cfg.Audio.TargetLUFS)
	if err != nil {
		return err
	}

	// Optional music bed.
	if cfg.Audio.MusicBed && course != nil {
		music := findMusicBed(course.Dir)
		if music == "" {
			fmt.Fprintf(e.out(), "  ⚠ audio     music_bed is on but no .mp3 found in %s — skipping\n",
				filepath.Join(course.Dir, "assets", "music"))
		} else {
			if err := mixMusicBed(ctx, e, voiceover, music, cfg.Audio.TargetLUFS, cfg.Audio.MusicDuckDB); err != nil {
				return err
			}
			report.MusicBed = filepath.Base(music)
			fmt.Fprintf(e.out(), "    music bed %s mixed in (ducked %.0f dB under voice)\n", filepath.Base(music), cfg.Audio.MusicDuckDB)
		}
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), ReviewsDirName, LoudnessFileName), report); err != nil {
		return err
	}

	fmt.Fprintf(e.out(), "    %s written (%s of narration)\n", VoiceoverFileName, total.Round(time.Second))
	fmt.Fprintf(e.out(), "    loudness %s → %s LUFS (target %.0f), true peak %s dBTP\n",
		lufsString(report.Before.IntegratedLUFS), lufsString(report.After.IntegratedLUFS),
		cfg.Audio.TargetLUFS, lufsString(report.After.TruePeakDB))
	return nil
}

// loadTTSFixes reads generated/tts_fixes.json (written by the align stage's
// WER gate). Missing file → empty map.
func loadTTSFixes(l *project.Lesson) (map[string]string, error) {
	data, err := os.ReadFile(filepath.Join(l.GeneratedDir(), TTSFixesFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", TTSFixesFileName, err)
	}
	fixes := map[string]string{}
	if err := json.Unmarshal(data, &fixes); err != nil {
		return nil, fmt.Errorf("parsing %s (delete it to reset): %w", TTSFixesFileName, err)
	}
	return fixes, nil
}

// ttsSampleRate is the pipeline-wide audio format every segment is
// resampled to before joining (Kokoro's native rate).
const ttsSampleRate = 24000

// joinWavs joins voice segments into one WAV, inserting pauseMs of silence
// between them and crossfading crossfadeMs at every seam so no join can
// click or clip a word.
func joinWavs(ctx context.Context, e *Env, wavPaths []string, pauseMs, crossfadeMs int, outPath string) error {
	if len(wavPaths) == 0 {
		return fmt.Errorf("no audio segments to join")
	}
	norm := fmt.Sprintf("aresample=%d,aformat=sample_fmts=s16:channel_layouts=mono", ttsSampleRate)
	if len(wavPaths) == 1 {
		return e.runFFmpeg(ctx, "-y", "-i", wavPaths[0], "-af", norm, "-c:a", "pcm_s16le", outPath)
	}

	// One reusable silence file; ffmpeg treats each -i occurrence as its own
	// input stream.
	var silence string
	if pauseMs > 0 {
		silence = outPath + ".pause.wav"
		if err := e.runFFmpeg(ctx, "-y", "-f", "lavfi", "-t", fmt.Sprintf("%.3f", float64(pauseMs)/1000),
			"-i", fmt.Sprintf("anullsrc=r=%d:cl=mono", ttsSampleRate), "-c:a", "pcm_s16le", silence); err != nil {
			return fmt.Errorf("generating pause silence: %w", err)
		}
		defer os.Remove(silence)
	}

	args := []string{"-y"}
	inputs := 0
	for i, p := range wavPaths {
		if i > 0 && silence != "" {
			args = append(args, "-i", silence)
			inputs++
		}
		args = append(args, "-i", p)
		inputs++
	}

	var filter strings.Builder
	for i := 0; i < inputs; i++ {
		fmt.Fprintf(&filter, "[%d:a]%s[n%d];", i, norm, i)
	}
	prev := "[n0]"
	if crossfadeMs > 0 {
		d := float64(crossfadeMs) / 1000
		for i := 1; i < inputs; i++ {
			out := fmt.Sprintf("[x%d]", i)
			fmt.Fprintf(&filter, "%s[n%d]acrossfade=d=%.3f:c1=tri:c2=tri%s;", prev, i, d, out)
			prev = out
		}
	} else {
		var labels strings.Builder
		for i := 0; i < inputs; i++ {
			fmt.Fprintf(&labels, "[n%d]", i)
		}
		fmt.Fprintf(&filter, "%sconcat=n=%d:v=0:a=1[cat];", labels.String(), inputs)
		prev = "[cat]"
	}
	graph := strings.TrimSuffix(filter.String(), ";")

	args = append(args, "-filter_complex", graph, "-map", prev, "-c:a", "pcm_s16le", outPath)
	if err := e.runFFmpeg(ctx, args...); err != nil {
		return fmt.Errorf("joining %d audio segments: %w", len(wavPaths), err)
	}
	return nil
}

// LoudnessFileName is the mastering QA report under generated/reviews/.
const LoudnessFileName = "loudness.json"

// masterChain is the voice processing applied before loudness
// normalization. Documented stage by stage:
//
//	highpass=f=70      — remove rumble below the voice band
//	deesser=i=0.4      — tame harsh sibilance ("s"/"sh" spikes)
//	acompressor        — gentle 2.5:1 compression from -18 dBFS
//	                     (15 ms attack keeps consonants, 200 ms release
//	                     breathes with speech)
const masterChain = "highpass=f=70,deesser=i=0.4,acompressor=threshold=0.125:ratio=2.5:attack=15:release=200"

// LoudnessStats is one loudnorm measurement.
type LoudnessStats struct {
	IntegratedLUFS float64 `json:"integrated_lufs"`
	TruePeakDB     float64 `json:"true_peak_db"`
	LoudnessRange  float64 `json:"loudness_range"`
}

// LoudnessReport is the persisted before/after mastering record.
type LoudnessReport struct {
	Before     LoudnessStats `json:"before"`
	After      LoudnessStats `json:"after"`
	TargetLUFS float64       `json:"target_lufs"`
	Chain      string        `json:"chain"`
	TwoPass    bool          `json:"two_pass"`
	MusicBed   string        `json:"music_bed,omitempty"`
}

// masterVoiceover applies masterChain and two-pass loudnorm, writing the
// mastered file to outPath and returning before/after measurements.
func masterVoiceover(ctx context.Context, e *Env, inPath, outPath string, targetLUFS float64) (*LoudnessReport, error) {
	report := &LoudnessReport{TargetLUFS: targetLUFS, Chain: masterChain}

	before, err := measureLoudness(ctx, e, inPath, "", targetLUFS)
	if err != nil {
		return nil, fmt.Errorf("measuring input loudness: %w", err)
	}
	report.Before = before.stats()

	// Pass 1: measure through the master chain (the chain shifts loudness,
	// so loudnorm must see post-chain audio).
	measured, err := measureLoudness(ctx, e, inPath, masterChain, targetLUFS)
	if err != nil {
		return nil, fmt.Errorf("loudnorm measurement pass: %w", err)
	}

	// Pass 2: apply chain + loudnorm with the measured values (linear mode,
	// which two-pass measurement makes possible — no dynamic pumping).
	af := masterChain
	if measured.InputI > -70 { // skip normalization of silence (test fixtures, dry runs)
		af = fmt.Sprintf(
			"%s,loudnorm=I=%.1f:TP=-1.5:LRA=11:measured_I=%.2f:measured_TP=%.2f:measured_LRA=%.2f:measured_thresh=%.2f:offset=%.2f:linear=true",
			masterChain, targetLUFS, measured.InputI, measured.InputTP, measured.InputLRA, measured.InputThresh, measured.TargetOffset,
		)
		report.TwoPass = true
	}
	tmp := outPath + ".master.wav"
	if err := e.runFFmpeg(ctx, "-y", "-i", inPath, "-af", af,
		"-ar", fmt.Sprint(ttsSampleRate), "-ac", "1", "-c:a", "pcm_s16le", tmp); err != nil {
		return nil, fmt.Errorf("mastering voiceover: %w", err)
	}
	if err := os.Rename(tmp, outPath); err != nil {
		os.Remove(tmp)
		return nil, fmt.Errorf("replacing %s: %w", outPath, err)
	}

	after, err := measureLoudness(ctx, e, outPath, "", targetLUFS)
	if err != nil {
		return nil, fmt.Errorf("verifying output loudness: %w", err)
	}
	report.After = after.stats()
	return report, nil
}

// loudnormMeasurement is ffmpeg loudnorm's print_format=json block. ffmpeg
// emits every value as a string.
type loudnormMeasurement struct {
	InputI       float64
	InputTP      float64
	InputLRA     float64
	InputThresh  float64
	TargetOffset float64
}

func (m loudnormMeasurement) stats() LoudnessStats {
	return LoudnessStats{IntegratedLUFS: m.InputI, TruePeakDB: m.InputTP, LoudnessRange: m.InputLRA}
}

// measureLoudness runs a loudnorm analysis pass (optionally through a
// preceding filter chain) and parses the JSON summary from stderr.
func measureLoudness(ctx context.Context, e *Env, path, chain string, targetLUFS float64) (*loudnormMeasurement, error) {
	af := fmt.Sprintf("loudnorm=I=%.1f:TP=-1.5:LRA=11:print_format=json", targetLUFS)
	if chain != "" {
		af = chain + "," + af
	}
	stderr, err := e.runFFmpegCapture(ctx, "-hide_banner", "-i", path, "-af", af, "-f", "null", "-")
	if err != nil {
		return nil, err
	}
	return parseLoudnorm(stderr)
}

// parseLoudnorm extracts the last {...} JSON block from ffmpeg stderr and
// reads the measured input values.
func parseLoudnorm(stderr string) (*loudnormMeasurement, error) {
	start := strings.LastIndex(stderr, "{")
	end := strings.LastIndex(stderr, "}")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("no loudnorm JSON in ffmpeg output:\n%s", tailLines(stderr, 6))
	}
	var raw map[string]string
	if err := json.Unmarshal([]byte(stderr[start:end+1]), &raw); err != nil {
		return nil, fmt.Errorf("parsing loudnorm JSON: %w", err)
	}
	get := func(key string) float64 {
		v := raw[key]
		if v == "" || strings.Contains(v, "inf") || strings.Contains(v, "nan") {
			return -99
		}
		var f float64
		if _, err := fmt.Sscanf(v, "%f", &f); err != nil {
			return -99
		}
		return f
	}
	return &loudnormMeasurement{
		InputI:       get("input_i"),
		InputTP:      get("input_tp"),
		InputLRA:     get("input_lra"),
		InputThresh:  get("input_thresh"),
		TargetOffset: get("target_offset"),
	}, nil
}

func lufsString(v float64) string {
	if v <= -90 {
		return "-inf"
	}
	return fmt.Sprintf("%.1f", v)
}

// findMusicBed returns the first .mp3 (alphabetically) under the course's
// assets/music dir, or "".
func findMusicBed(courseDir string) string {
	entries, err := os.ReadDir(filepath.Join(courseDir, "assets", "music"))
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".mp3") {
			return filepath.Join(courseDir, "assets", "music", entry.Name())
		}
	}
	return ""
}

// mixMusicBed mixes a looped music bed under the mastered voice, in place.
//
// The bed is gained so it sits duckDB below the voice's loudness target,
// then sidechain compression (keyed by the voice) pushes it down further
// whenever the narrator speaks, and it fades in/out over 2 s at the lesson
// bounds. A final limiter guards the -1.5 dBTP ceiling.
func mixMusicBed(ctx context.Context, e *Env, voicePath, musicPath string, targetLUFS, duckDB float64) error {
	voiceDurMs, err := mediaDurationMs(voicePath)
	if err != nil {
		return err
	}
	music, err := measureLoudness(ctx, e, musicPath, "", targetLUFS)
	if err != nil {
		return fmt.Errorf("measuring music bed loudness: %w", err)
	}
	bedGain := (targetLUFS + duckDB) - music.InputI // dB to bring bed to duckDB below voice
	dur := float64(voiceDurMs) / 1000
	fadeOutStart := dur - 2
	if fadeOutStart < 0 {
		fadeOutStart = 0
	}

	filter := fmt.Sprintf(
		"[1:a]aresample=%d,aformat=sample_fmts=s16:channel_layouts=mono,atrim=0:%.3f,"+
			"volume=%.1fdB,afade=t=in:d=2,afade=t=out:st=%.3f:d=2[bed];"+
			"[bed][0:a]sidechaincompress=threshold=0.02:ratio=8:attack=25:release=350[duck];"+
			"[0:a][duck]amix=inputs=2:duration=first:dropout_transition=0:normalize=0[mix];"+
			"[mix]alimiter=limit=0.84:level=false[out]",
		ttsSampleRate, dur, bedGain, fadeOutStart,
	)
	tmp := voicePath + ".music.wav"
	err = e.runFFmpeg(ctx, "-y",
		"-i", voicePath,
		"-stream_loop", "-1", "-i", musicPath,
		"-filter_complex", filter, "-map", "[out]",
		"-t", fmt.Sprintf("%.3f", dur),
		"-ar", fmt.Sprint(ttsSampleRate), "-c:a", "pcm_s16le", tmp)
	if err != nil {
		return fmt.Errorf("mixing music bed: %w", err)
	}
	if err := os.Rename(tmp, voicePath); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("replacing %s: %w", voicePath, err)
	}
	return nil
}

// wavDuration reads a WAV file's header and returns its play time.
func wavDuration(path string) (time.Duration, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("statting %s: %w", path, err)
	}
	fileSize := info.Size()

	header := make([]byte, 12)
	if _, err := io.ReadFull(f, header); err != nil {
		return 0, fmt.Errorf("%s is not a WAV file: %w", path, err)
	}
	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return 0, fmt.Errorf("%s is not a WAV file", path)
	}

	var byteRate uint32
	// Walk the RIFF chunks for "fmt " (byte rate) and "data" (payload size).
	for {
		chunkHeader := make([]byte, 8)
		if _, err := io.ReadFull(f, chunkHeader); err != nil {
			return 0, fmt.Errorf("%s: missing data chunk: %w", path, err)
		}
		chunkID := string(chunkHeader[0:4])
		chunkSize := int64(binary.LittleEndian.Uint32(chunkHeader[4:8]))
		offset, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", path, err)
		}
		switch chunkID {
		case "fmt ":
			fmtChunk := make([]byte, chunkSize)
			if _, err := io.ReadFull(f, fmtChunk); err != nil {
				return 0, fmt.Errorf("%s: truncated fmt chunk: %w", path, err)
			}
			if len(fmtChunk) < 12 {
				return 0, fmt.Errorf("%s: fmt chunk too short", path)
			}
			byteRate = binary.LittleEndian.Uint32(fmtChunk[8:12])
		case "data":
			if byteRate == 0 {
				return 0, fmt.Errorf("%s: data chunk before fmt chunk", path)
			}
			// Streaming writers (Kokoro-FastAPI among them) emit 0 or
			// 0xFFFFFFFF as a placeholder size; the real payload is
			// whatever follows the header.
			if remaining := fileSize - offset; chunkSize <= 0 || chunkSize == 0xFFFFFFFF || chunkSize > remaining {
				chunkSize = remaining
			}
			return time.Duration(float64(chunkSize) / float64(byteRate) * float64(time.Second)), nil
		default:
			if _, err := f.Seek(chunkSize+chunkSize%2, io.SeekCurrent); err != nil {
				return 0, fmt.Errorf("%s: seeking past %s chunk: %w", path, chunkID, err)
			}
		}
	}
}
