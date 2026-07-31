package pipeline

// Desktop captures: Cursor, Figma desktop — the apps with no DOM to drive.
//
// == Why this one has a person in it ==
//
// The terminal path generates its own script. The web path runs a script
// somebody wrote once against selectors they can inspect. Neither is available
// here: a native app exposes no selectors, and driving it by simulated
// keystrokes means encoding pixel positions and menu paths that break on the
// next release without anything noticing.
//
// So this path is honest about needing a person. The engine does the parts a
// person is bad at — putting the window in exactly the same place every time,
// recording, cropping, timing, encoding — and the person does the part they are
// good at, which is using the application. A take is a list of beats; the
// engine shows one at a time and stamps a mark when the operator says they have
// done it.
//
// That makes it slower than the other two paths and no less repeatable. The
// beats are checked in, the window geometry is fixed, and a re-shoot is running
// the same list again — which is what §4.5 needs and what a hand-recorded video
// can never offer.
//
// == Why the whole screen, then a crop ==
//
// avfoundation captures a *display*, not a window; window-level capture needs
// ScreenCaptureKit and a second toolchain. Positioning the window to a known
// rectangle and cropping the display to it gets the same frames with tools that
// are already here.
//
// The one trap is Retina. AppleScript works in points and the capture comes out
// in pixels, so the crop has to be scaled by the display's backing factor —
// which is *measured* from the recording rather than assumed to be 2, because
// it is 1 on an external monitor and the failure mode is a video cropped to a
// quarter of the window with nothing to indicate why.

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// captureApp is one recordable desktop application.
type captureApp struct {
	// Key is what a lesson writes in `tool=`.
	Key string
	// Display is how the app is named to a reader.
	Display string
	// Bundle is the macOS application name AppleScript addresses it by.
	Bundle string
}

var captureApps = map[string]captureApp{
	"cursor":        {Key: "cursor", Display: "Cursor", Bundle: "Cursor"},
	"figma-desktop": {Key: "figma-desktop", Display: "Figma", Bundle: "Figma"},
}

func captureAppKeys() []string {
	keys := make([]string, 0, len(captureApps))
	for k := range captureApps {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

// DesktopBeat is one thing the operator does, and the mark it earns.
type DesktopBeat struct {
	// Mark names the moment, in the shared mark vocabulary.
	Mark string `yaml:"mark"`
	// Prompt is what the operator is asked to do, in the imperative. It is
	// shown on its own while the recording runs, so it has to be doable
	// without reading anything else.
	Prompt string `yaml:"prompt"`
}

// DesktopTake is a course's `takes/<name>.yaml` for a desktop app.
type DesktopTake struct {
	// App is the captureApps key.
	App string `yaml:"app"`
	// Window is the size the window is set to before recording. Fixed rather
	// than "whatever it was", because a re-shoot has to frame identically or
	// the new clip cannot be cut into the old lesson.
	Window struct {
		Width  int `yaml:"width"`
		Height int `yaml:"height"`
	} `yaml:"window"`
	Beats []DesktopBeat `yaml:"beats"`
}

const (
	defaultDesktopWindowW = 1440
	defaultDesktopWindowH = 900
	// desktopWindowOrigin keeps the window clear of the menu bar. Recording
	// the menu bar would put the operator's own clock, battery and unrelated
	// application names into the course.
	desktopWindowX = 40
	desktopWindowY = 60
)

// LoadDesktopTake reads and validates a desktop take.
func LoadDesktopTake(path string) (*DesktopTake, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading take %s: %w", path, err)
	}
	var t DesktopTake
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parsing take %s: %w", path, err)
	}
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &t, nil
}

// Validate checks the take before anything is launched or recorded.
func (t *DesktopTake) Validate() error {
	if _, ok := captureApps[t.App]; !ok {
		return fmt.Errorf("app %q is not recordable. Available: %s", t.App, strings.Join(captureAppKeys(), ", "))
	}
	if len(t.Beats) == 0 {
		return fmt.Errorf("take has no beats — a desktop capture is a list of things to do, and an empty list records somebody looking at a window")
	}
	seen := map[string]bool{}
	for i, b := range t.Beats {
		n := i + 1
		if b.Mark == "" {
			return fmt.Errorf("beat %d has no mark", n)
		}
		if !markNameRe.MatchString(b.Mark) {
			return fmt.Errorf("beat %d: mark %q must be lowercase letters, digits and dashes", n, b.Mark)
		}
		if seen[b.Mark] {
			return fmt.Errorf("beat %d: mark %q is used twice", n, b.Mark)
		}
		seen[b.Mark] = true
		if strings.TrimSpace(b.Prompt) == "" {
			return fmt.Errorf("beat %d (%s) tells the operator nothing to do. The prompt is read while recording, so it has to stand alone", n, b.Mark)
		}
	}
	return nil
}

// size returns the window size, defaults applied.
func (t *DesktopTake) size() (int, int) {
	w, h := t.Window.Width, t.Window.Height
	if w == 0 {
		w = defaultDesktopWindowW
	}
	if h == 0 {
		h = defaultDesktopWindowH
	}
	return w, h
}

// avScreenIndexRe matches the screen entries in ffmpeg's device listing.
var avScreenIndexRe = regexp.MustCompile(`\[(\d+)\]\s+Capture screen \d+`)

// parseAVScreenIndex finds the first screen-capture device index in ffmpeg's
// `-list_devices` output.
//
// It is parsed rather than hard-coded because the index counts *all* video
// devices: on a machine with no webcam the screen is 0, and on this one it is 1
// because the FaceTime camera took 0. Hard-coding it records somebody's face
// instead of their screen, which is a memorable way to fail.
func parseAVScreenIndex(listing string) (int, error) {
	m := avScreenIndexRe.FindStringSubmatch(listing)
	if m == nil {
		return 0, fmt.Errorf("ffmpeg listed no screen-capture device. On macOS this usually means Screen Recording permission has not been granted to this terminal (System Settings → Privacy & Security → Screen Recording)")
	}
	return strconv.Atoi(m[1])
}

// cropFilter builds the ffmpeg crop for a window rectangle given in points,
// against a recording made in pixels.
//
// scale is measured, never assumed: it is 2 on a Retina display and 1 on an
// external monitor, and getting it wrong crops to a quarter of the window with
// nothing on screen to say why.
func cropFilter(x, y, w, h int, scale float64) string {
	px := func(v int) int {
		s := int(float64(v)*scale + 0.5)
		if s%2 != 0 {
			s-- // h264 refuses odd dimensions
		}
		return s
	}
	return fmt.Sprintf("crop=%d:%d:%d:%d", px(w), px(h), px(x), px(y))
}

// DesktopRecorder starts and stops a screen recording. The real one shells out
// to ffmpeg; tests supply their own.
type DesktopRecorder interface {
	// Start begins recording the whole screen to path.
	Start(ctx context.Context, path string) error
	// Stop ends the recording and waits for the file to be finalised.
	Stop() error
}

// avfoundationRecorder records the display with ffmpeg.
type avfoundationRecorder struct {
	env   *Env
	index int
	cmd   *exec.Cmd
	stdin io.WriteCloser
}

func (r *avfoundationRecorder) Start(ctx context.Context, path string) error {
	// -framerate 30 rather than the default, because avfoundation's default is
	// whatever the display reports and a 120Hz panel produces an enormous file
	// nobody needs for a screen recording.
	args := []string{
		"-y", "-f", "avfoundation",
		"-capture_cursor", "1",
		"-framerate", "30",
		"-i", fmt.Sprintf("%d:none", r.index),
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "18",
		"-pix_fmt", "yuv420p",
		path,
	}
	cmd := exec.CommandContext(ctx, r.env.ffmpegBin(), args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting the screen recording: %w", err)
	}
	r.cmd, r.stdin = cmd, stdin
	return nil
}

func (r *avfoundationRecorder) Stop() error {
	if r.cmd == nil {
		return nil
	}
	// "q" is ffmpeg's clean stop; killing it leaves an unfinalised container
	// with no moov atom, which plays nowhere.
	_, _ = io.WriteString(r.stdin, "q\n")
	_ = r.stdin.Close()
	done := make(chan error, 1)
	go func() { done <- r.cmd.Wait() }()
	select {
	case <-done:
		return nil
	case <-time.After(20 * time.Second):
		_ = r.cmd.Process.Kill()
		return fmt.Errorf("ffmpeg did not stop cleanly; the recording may be unusable")
	}
}

// osascript runs an AppleScript and returns its output.
func osascript(ctx context.Context, script string) (string, error) {
	out, err := exec.CommandContext(ctx, "osascript", "-e", script).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("osascript: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// placeWindow brings the app to the front and puts its window at a fixed
// rectangle, so every take of the same app frames identically.
func placeWindow(ctx context.Context, app captureApp, w, h int) error {
	script := fmt.Sprintf(`
tell application %q to activate
delay 0.6
tell application "System Events"
  tell process %q
    set frontmost to true
    if (count of windows) is 0 then error "no window is open"
    set position of front window to {%d, %d}
    set size of front window to {%d, %d}
  end tell
end tell`, app.Bundle, app.Bundle, desktopWindowX, desktopWindowY, w, h)
	if _, err := osascript(ctx, script); err != nil {
		return fmt.Errorf("positioning the %s window: %w\n\nThis needs Accessibility permission for your terminal (System Settings → Privacy & Security → Accessibility), and %s has to be running with a window open", app.Display, err, app.Display)
	}
	return nil
}

// screenPointWidth asks macOS how wide the display is in points, which is what
// the measured Retina scale is derived against.
func screenPointWidth(ctx context.Context) (int, error) {
	out, err := osascript(ctx, `tell application "Finder" to get bounds of window of desktop`)
	if err != nil {
		return 0, err
	}
	// "0, 0, 1512, 982"
	parts := strings.Split(out, ",")
	if len(parts) != 4 {
		return 0, fmt.Errorf("unexpected desktop bounds %q", out)
	}
	return strconv.Atoi(strings.TrimSpace(parts[2]))
}

// DesktopCaptureResult is what one desktop take produced.
type DesktopCaptureResult struct {
	ClipPath string
	ClipMs   int
	Marks    []FootageMark
}

// RunDesktopTake records an operator working through the take's beats.
//
// in/out are the operator's console. Every mark is stamped when they press
// Enter, so — like the web video path and unlike the terminal one — the marks
// are measured rather than modelled and are never approximate.
func RunDesktopTake(ctx context.Context, e *Env, take *DesktopTake, rec DesktopRecorder, outDir, clipName string, in io.Reader, out io.Writer) (*DesktopCaptureResult, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("desktop capture is macOS-only (it uses avfoundation and AppleScript); this is %s", runtime.GOOS)
	}
	app := captureApps[take.App]
	w, h := take.size()

	if err := placeWindow(ctx, app, w, h); err != nil {
		return nil, err
	}
	pointW, err := screenPointWidth(ctx)
	if err != nil {
		return nil, fmt.Errorf("measuring the display: %w", err)
	}

	raw := filepath.Join(outDir, clipName+"-raw.mp4")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	defer os.Remove(raw)

	fmt.Fprintf(out, "\nRecording %s. Work through these in the app; press Enter here after each.\n\n", app.Display)
	if err := rec.Start(ctx, raw); err != nil {
		return nil, err
	}
	started := time.Now()

	marks := make([]FootageMark, 0, len(take.Beats))
	reader := bufio.NewReader(in)
	for i, b := range take.Beats {
		fmt.Fprintf(out, "  %d/%d  %s\n", i+1, len(take.Beats), b.Prompt)
		if _, err := reader.ReadString('\n'); err != nil && err != io.EOF {
			_ = rec.Stop()
			return nil, fmt.Errorf("reading the operator's confirmation: %w", err)
		}
		marks = append(marks, FootageMark{Name: b.Mark, AtMs: int(time.Since(started).Milliseconds())})
	}
	// A moment of tail so the last beat is not cut off at the instant it lands.
	time.Sleep(1200 * time.Millisecond)
	if err := rec.Stop(); err != nil {
		return nil, err
	}
	fmt.Fprintf(out, "\nRecording stopped. Cropping to the window...\n")

	// The scale is measured from what was actually recorded.
	pixW, err := videoWidth(raw)
	if err != nil {
		return nil, fmt.Errorf("measuring the recording: %w", err)
	}
	scale := 1.0
	if pointW > 0 {
		scale = float64(pixW) / float64(pointW)
	}

	clip := filepath.Join(outDir, clipName+".mp4")
	if err := e.runFFmpeg(ctx, "-y", "-i", raw,
		"-vf", cropFilter(desktopWindowX, desktopWindowY, w, h, scale),
		"-c:v", "libx264", "-preset", "medium", "-crf", "20", "-pix_fmt", "yuv420p",
		clip,
	); err != nil {
		return nil, fmt.Errorf("cropping the recording to the window: %w", err)
	}
	durMs, err := mediaDurationMs(clip)
	if err != nil {
		return nil, err
	}
	return &DesktopCaptureResult{ClipPath: clipName + ".mp4", ClipMs: durMs, Marks: marks}, nil
}

// videoWidth probes a file's pixel width.
func videoWidth(path string) (int, error) {
	out, err := exec.Command("ffprobe", "-v", "error",
		"-select_streams", "v:0", "-show_entries", "stream=width",
		"-of", "csv=p=0", path).Output()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(out)))
}

// NewDesktopRecorder builds the real recorder, discovering the screen device.
func NewDesktopRecorder(ctx context.Context, e *Env) (DesktopRecorder, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("desktop capture is macOS-only; this is %s", runtime.GOOS)
	}
	// ffmpeg exits non-zero after listing devices, which is expected.
	out, _ := exec.CommandContext(ctx, e.ffmpegBin(),
		"-hide_banner", "-f", "avfoundation", "-list_devices", "true", "-i", "").CombinedOutput()
	idx, err := parseAVScreenIndex(string(out))
	if err != nil {
		return nil, err
	}
	return &avfoundationRecorder{env: e, index: idx}, nil
}

// DesktopCaptureReadiness reports whether a desktop capture could run here,
// naming the screen device it would use.
//
// It is deliberately a separate check from the other two capture kinds. The
// permission this needs is granted to the *terminal application*, not to
// coursesmith, so somebody reading "no screen device" has no reason to look in
// System Settings — and the failure without it is silent: ffmpeg simply lists no
// screen rather than saying it was refused.
func DesktopCaptureReadiness(ctx context.Context, e *Env) (string, error) {
	if runtime.GOOS != "darwin" {
		return "", fmt.Errorf("desktop capture is macOS-only; this is %s", runtime.GOOS)
	}
	out, _ := exec.CommandContext(ctx, e.ffmpegBin(),
		"-hide_banner", "-f", "avfoundation", "-list_devices", "true", "-i", "").CombinedOutput()
	idx, err := parseAVScreenIndex(string(out))
	if err != nil {
		return "", err
	}
	missing := make([]string, 0, len(captureApps))
	for _, k := range captureAppKeys() {
		if _, err := os.Stat(filepath.Join("/Applications", captureApps[k].Bundle+".app")); err != nil {
			missing = append(missing, k)
		}
	}
	name := fmt.Sprintf("avfoundation screen %d", idx)
	if len(missing) > 0 {
		name += "; not installed: " + strings.Join(missing, ", ")
	}
	return name, nil
}
