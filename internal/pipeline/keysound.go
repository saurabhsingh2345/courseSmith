package pipeline

// The keystroke track: a click on every character the editor types.
//
// Synthesised rather than sampled, for the same reason the artwork is drawn
// rather than imported. A recording of a keyboard is one keyboard: it has a
// room, a mic and a licence, and every click in the clip is the identical
// waveform however much you vary the gain. Building the click means every one
// is slightly different by construction, there is no asset to ship or attribute,
// and the sound is a function of the schedule — so it cannot drift out of sync
// with the picture, because the picture is a function of the same array.
//
// A key press is a transient, not a tone: a burst of noise shaped by a fast
// attack and a short decay, over a low "thock" that gives it a body. Two
// components and about twenty milliseconds. Anything longer overlaps the next
// keystroke at typing speed and turns into a rattle.
//
// It sits far under the voice on purpose (see keyClickPeak). This is texture —
// the viewer should notice it only if it stops.

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"path/filepath"
)

const (
	// 44.1kHz mono. The voice track is whatever Kokoro produced; these are
	// separate <Audio> elements and the renderer resamples, so this only has to
	// be high enough for a click's transient to survive — a 20ms burst at 22kHz
	// loses the top of the attack, which is the part that reads as a key.
	keySampleRate = 44100

	// Peak amplitude of one click, against full scale.
	//
	// This is the number that decides whether the feature is pleasant or
	// unbearable, and it is deliberately low. The voice is mastered to −16 LUFS
	// and a click at even 0.25 sat *on* it rather than under it. At 0.09 the
	// typing is texture you stop hearing after two seconds, which is what
	// typing in a screencast actually is.
	keyClickPeak = 0.09

	// A newline is a different key: bigger, deeper, and the one a viewer
	// actually registers as punctuation in the rhythm.
	keyEnterPeak = 0.16

	// Click envelope, in milliseconds.
	keyAttackMs     = 0.8
	keyDecayMs      = 17.0
	keyEnterDecayMs = 34.0

	// Tail beyond the last keystroke, so the file does not end mid-decay.
	keyTailMs = 120
)

// KeySoundFileName is the generated track, beside voiceover.wav.
const KeySoundFileName = "keys.wav"

// keyVariant returns the deterministic per-keystroke variation: a pitch
// multiplier and a level multiplier.
//
// Real typing is not one sound repeated — keys differ, and the same key differs
// by how it is struck. Without this the track is a metronome, which is worse
// than silence because it draws attention to itself.
func keyVariant(i int) (pitch, level float64) {
	h := fnv.New32a()
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(i))
	_, _ = h.Write(buf[:])
	v := h.Sum32()
	pitch = 0.82 + 0.36*float64(v%1000)/1000.0
	level = 0.72 + 0.42*float64((v/1000)%1000)/1000.0
	return pitch, level
}

// noiseAt is a deterministic white-noise sample for absolute sample index n.
// A hash rather than a PRNG so the track is reproducible from the schedule
// alone, with no state carried between runs.
func noiseAt(n int) float64 {
	h := fnv.New32a()
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(n))
	_, _ = h.Write(buf[:])
	return float64(h.Sum32()%20001)/10000.0 - 1.0
}

// renderClick mixes one keystroke into buf at sample offset `at`.
func renderClick(buf []float64, at int, isEnter bool, variant int) {
	pitch, level := keyVariant(variant)
	peak := keyClickPeak
	decayMs := keyDecayMs
	// The body's pitch: a low thock under the noise. Higher for ordinary keys,
	// lower for the newline, which is what makes Enter read as a bigger key
	// rather than just a louder one.
	bodyHz := 780.0 * pitch
	if isEnter {
		peak = keyEnterPeak
		decayMs = keyEnterDecayMs
		bodyHz = 300.0 * pitch
	}

	attack := keyAttackMs / 1000 * keySampleRate
	decay := decayMs / 1000 * keySampleRate
	span := int(attack + decay)

	for k := 0; k < span; k++ {
		n := at + k
		if n < 0 || n >= len(buf) {
			continue
		}
		// Fast attack, exponential decay. The attack matters more than it
		// looks: an instant onset clicks digitally (a step in the waveform),
		// and anything slower than about a millisecond stops reading as a key.
		var env float64
		if float64(k) < attack {
			env = float64(k) / attack
		} else {
			env = math.Exp(-3.6 * (float64(k) - attack) / decay)
		}

		t := float64(k) / keySampleRate
		// Noise carries the click; the sine carries the body. The noise decays
		// faster than the envelope so the transient is at the front, which is
		// where a key's is.
		click := noiseAt(n) * math.Exp(-9.0*t*1000/decayMs)
		body := math.Sin(2 * math.Pi * bodyHz * t)
		buf[n] += peak * level * env * (0.62*click + 0.38*body)
	}
}

// WriteKeystrokeTrack synthesises a mono WAV with a click at each millisecond
// in `times`, and returns the number of clicks written.
//
// `newlineAt` marks which indices are newlines; those get the heavier key. The
// track runs from 0 to the last click plus a short tail, so it can be dropped
// straight onto the timeline beside the voiceover with no offset.
func WriteKeystrokeTrack(path string, times []int, newlineAt map[int]bool, totalMs int) (int, error) {
	if len(times) == 0 {
		return 0, nil
	}
	end := totalMs
	if last := times[len(times)-1] + keyTailMs; last > end {
		end = last
	}
	if end <= 0 {
		return 0, nil
	}

	samples := make([]float64, int(float64(end)/1000*keySampleRate)+1)
	for i, ms := range times {
		renderClick(samples, int(float64(ms)/1000*keySampleRate), newlineAt[i], i)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0, err
	}
	f, err := os.Create(path)
	if err != nil {
		return 0, fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()

	if err := writeWavPCM16(f, samples, keySampleRate); err != nil {
		return 0, fmt.Errorf("writing %s: %w", path, err)
	}
	return len(times), nil
}

// collectKeystrokes gathers every scene's typing schedule into one timeline,
// marking which entries are newlines so those get the heavier key.
//
// Reads the schedule off the *scene graph* rather than recomputing it, which is
// the whole point: the numbers the renderer animates from and the numbers the
// clicks land on are then the same numbers, and a video-plan edit that retimed
// a scene retimes the sound with it for free.
func collectKeystrokes(graph *SceneGraph) ([]int, map[int]bool) {
	var times []int
	newlines := map[int]bool{}
	for _, scene := range graph.Scenes {
		sched := intSliceProp(scene.Props["keystrokes"])
		if len(sched) == 0 {
			continue
		}
		// The characters those keystrokes belong to, so newlines can be told
		// apart. A schedule whose code we cannot read still gets clicks — every
		// key the same weight is worse than the right weights, and far better
		// than silence.
		code := firstStepCode(scene.Props)
		runes := []rune(code)
		for i, ms := range sched {
			if i < len(runes) && runes[i] == '\n' {
				newlines[len(times)] = true
			}
			times = append(times, ms)
		}
	}
	return times, newlines
}

// firstStepCode digs the first step's code out of a walkthrough/workspace
// scene's props, tolerating both the Go-native shape and the map shape a
// video-plan patch leaves behind.
func firstStepCode(props map[string]any) string {
	switch steps := props["steps"].(type) {
	case []map[string]any:
		if len(steps) > 0 {
			s, _ := steps[0]["code"].(string)
			return s
		}
	case []any:
		if len(steps) > 0 {
			if m, ok := steps[0].(map[string]any); ok {
				s, _ := m["code"].(string)
				return s
			}
		}
	}
	// The workspace template keys its code by file rather than by step.
	if files, ok := props["files"].([]map[string]any); ok && len(files) > 0 {
		s, _ := files[0]["code"].(string)
		return s
	}
	return ""
}

// intSliceProp coerces a props value that should be a list of milliseconds.
// JSON round-trips turn []int into []any of float64, and a video-plan patch can
// deliver either; a schedule silently dropped here is a silent track.
func intSliceProp(v any) []int {
	switch xs := v.(type) {
	case []int:
		return xs
	case []any:
		out := make([]int, 0, len(xs))
		for _, x := range xs {
			switch n := x.(type) {
			case int:
				out = append(out, n)
			case float64:
				out = append(out, int(n))
			default:
				return nil
			}
		}
		return out
	}
	return nil
}

// writeWavPCM16 writes float samples (nominally −1..1) as a mono 16-bit PCM
// RIFF/WAVE file. Samples outside the range are clipped rather than scaled: a
// track that quietly turned itself down because two clicks overlapped would be
// a track whose level depends on the code being typed.
func writeWavPCM16(f *os.File, samples []float64, rate int) error {
	const bitsPerSample = 16
	const channels = 1
	dataBytes := len(samples) * 2

	hdr := make([]byte, 0, 44)
	hdr = append(hdr, "RIFF"...)
	hdr = binary.LittleEndian.AppendUint32(hdr, uint32(36+dataBytes))
	hdr = append(hdr, "WAVE"...)
	hdr = append(hdr, "fmt "...)
	hdr = binary.LittleEndian.AppendUint32(hdr, 16) // PCM chunk size
	hdr = binary.LittleEndian.AppendUint16(hdr, 1)  // PCM
	hdr = binary.LittleEndian.AppendUint16(hdr, channels)
	hdr = binary.LittleEndian.AppendUint32(hdr, uint32(rate))
	hdr = binary.LittleEndian.AppendUint32(hdr, uint32(rate*channels*bitsPerSample/8))
	hdr = binary.LittleEndian.AppendUint16(hdr, uint16(channels*bitsPerSample/8))
	hdr = binary.LittleEndian.AppendUint16(hdr, bitsPerSample)
	hdr = append(hdr, "data"...)
	hdr = binary.LittleEndian.AppendUint32(hdr, uint32(dataBytes))
	if _, err := f.Write(hdr); err != nil {
		return err
	}

	out := make([]byte, dataBytes)
	for i, s := range samples {
		if s > 1 {
			s = 1
		} else if s < -1 {
			s = -1
		}
		binary.LittleEndian.PutUint16(out[i*2:], uint16(int16(s*32767)))
	}
	_, err := f.Write(out)
	return err
}
