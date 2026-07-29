package pipeline

// Auto-pace: closing the pace-report loop. The align stage measures actual
// words-per-minute against style.pace_wpm; when the whole lesson runs
// outside the tolerance band it writes tts_speed.json — an absolute speaking-
// rate correction the audio stage multiplies into the user's voice_speed on
// the next run (the file's appearance re-stales the audio stage, exactly like
// the pronunciation auto-fix loop).

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/enfec/coursesmith/internal/project"
)

// TTSSpeedFileName is the auto-pace correction artifact in generated/.
const TTSSpeedFileName = "tts_speed.json"

// Speed-correction bounds: beyond these the voice sounds wrong, so the pace
// report stays advisory instead.
const (
	minSpeedFix = 0.75
	maxSpeedFix = 1.35
	// minSpeedDelta avoids churn: corrections under 3% aren't worth a
	// re-synthesis round trip.
	minSpeedDelta = 0.03
)

// effectivePaceWPM is the pace target auto-pace actually steers to.
//
// style.pace_wpm describes the voice at its natural rate; style.voice_speed is
// the user saying "read it slower than that". Measuring a 0.9x read against
// the 1.0x target makes every lesson look 10% under pace, and auto-pace would
// answer by writing a speed-up correction that cancels the very setting that
// asked for it. Scaling the target by the same factor keeps the loop closed
// around what was asked for rather than around what was overridden.
func effectivePaceWPM(paceWPM int, voiceSpeed float64) int {
	if paceWPM <= 0 {
		return 0
	}
	if voiceSpeed <= 0 {
		voiceSpeed = 1
	}
	target := int(float64(paceWPM)*voiceSpeed + 0.5)
	if target <= 0 {
		return 0
	}
	return target
}

// TTSSpeedFix is the persisted tts_speed.json.
type TTSSpeedFix struct {
	// Speed is the absolute auto-pace multiplier (composes with the user's
	// style.voice_speed; both default to 1).
	Speed float64 `json:"speed"`
	// MeasuredWPM and TargetWPM document why the correction exists.
	MeasuredWPM float64 `json:"measured_wpm"`
	TargetWPM   int     `json:"target_wpm"`
}

// loadTTSSpeedFix reads tts_speed.json; missing means no correction (1.0).
func loadTTSSpeedFix(l *project.Lesson) (*TTSSpeedFix, error) {
	data, err := os.ReadFile(filepath.Join(l.GeneratedDir(), TTSSpeedFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", TTSSpeedFileName, err)
	}
	var fix TTSSpeedFix
	if err := json.Unmarshal(data, &fix); err != nil {
		return nil, fmt.Errorf("parsing %s (delete it to reset auto-pace): %w", TTSSpeedFileName, err)
	}
	return &fix, nil
}

// computeSpeedFix derives the next absolute correction from the lesson-wide
// measured wpm. oldFix is the correction that produced the measurement (1.0
// when none). Returns nil when no (new) correction is warranted.
func computeSpeedFix(measuredWPM float64, targetWPM int, oldFix float64) *TTSSpeedFix {
	if targetWPM <= 0 || measuredWPM <= 0 {
		return nil
	}
	deviation := (measuredWPM - float64(targetWPM)) / float64(targetWPM)
	if deviation < paceTolerance && deviation > -paceTolerance {
		return nil // inside the band — nothing to fix
	}
	if oldFix <= 0 {
		oldFix = 1
	}
	next := oldFix * float64(targetWPM) / measuredWPM
	if next < minSpeedFix {
		next = minSpeedFix
	}
	if next > maxSpeedFix {
		next = maxSpeedFix
	}
	if diff := next - oldFix; diff < minSpeedDelta && diff > -minSpeedDelta {
		return nil // already as close as the bounds allow
	}
	return &TTSSpeedFix{Speed: next, MeasuredWPM: measuredWPM, TargetWPM: targetWPM}
}
