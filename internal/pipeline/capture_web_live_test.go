package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod/lib/launcher"
)

// The web driver against a real browser and a real page.
//
// The page is one we serve ourselves rather than a real product, and that is
// the honest limit of what a test can assert here: nobody's CI should depend on
// Lovable's markup, and a test that did would fail for reasons that are not our
// bug. What this *can* prove is everything on our side of the boundary — that
// rod is driven correctly, that a shot lands as a real PNG of the right size,
// and that a focus selector resolves to a box in the right place. Those are the
// parts that would otherwise be verified by looking at a screenshot and
// squinting.
//
// Skipped unless a Chromium is available.
func requireChromium(t *testing.T) {
	t.Helper()
	if _, has := launcher.LookPath(); !has {
		t.Skip("no Chromium available for the web capture driver")
	}
}

const takeTestPage = `<!doctype html>
<html><head><meta charset="utf-8"><style>
  html,body{margin:0;padding:0;background:#101317;font-family:system-ui}
  #spacer{height:300px}
  #promptbox{display:block;box-sizing:border-box;margin:0 auto;width:600px;height:80px;background:#fff;font-size:24px}
  #result{display:none;color:#9f9;font-size:40px;text-align:center;padding:40px}
  .shown{display:block !important}
</style></head><body>
  <div id="spacer"></div>
  <input id="promptbox" data-testid="prompt" placeholder="describe your app">
  <div id="result">built</div>
  <script>
    document.getElementById('promptbox').addEventListener('input', () => {
      document.getElementById('result').className = 'shown';
    });
  </script>
</body></html>`

// registerTestSite adds a site pointing at a local server for the duration of
// one test. captureSites is package state, so it is removed again.
func registerTestSite(t *testing.T, key, origin string) {
	t.Helper()
	captureSites[key] = captureSite{Key: key, Display: "Test Site", Origin: origin}
	t.Cleanup(func() { delete(captureSites, key) })
}

func TestWebTakeDrivesARealBrowser(t *testing.T) {
	requireChromium(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(takeTestPage))
	}))
	defer srv.Close()
	registerTestSite(t, "testsite", srv.URL)

	take := &WebTake{Site: "testsite"}
	take.Viewport.Width = 1200
	take.Viewport.Height = 800
	take.Steps = []WebStep{
		{Do: "goto", Path: "/"},
		{Do: "wait", Selector: Selector{"[data-testid=prompt]"}, Timeout: "20s"},
		{Do: "shot", Mark: "landing", Focus: Selector{"[data-testid=prompt]"}},
		{Do: "type", Selector: Selector{"[data-testid=prompt]"}, Text: "a habit tracker"},
		{Do: "wait", Selector: Selector{"#result.shown"}, Timeout: "20s"},
		{Do: "shot", Mark: "built"},
	}
	if err := take.Validate(); err != nil {
		t.Fatalf("the take this test drives is itself invalid: %v", err)
	}

	outDir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	res, err := RunWebTake(ctx, nil, take, outDir, "capture-1", true)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Frames) != 2 {
		t.Fatalf("frames = %+v", res.Frames)
	}
	if res.Origin != srv.URL {
		t.Errorf("origin = %q, want the URL the driver really used (%q)", res.Origin, srv.URL)
	}
	if res.Frames[0].Mark != "landing" || res.Frames[1].Mark != "built" {
		t.Errorf("marks = %q, %q", res.Frames[0].Mark, res.Frames[1].Mark)
	}

	// Both shots are real PNGs of the viewport, at the 2x device scale the
	// driver sets — a frame captured at 1x would look soft on a 1080p stage.
	for _, fr := range res.Frames {
		data, err := os.ReadFile(filepath.Join(outDir, fr.Path))
		if err != nil {
			t.Fatalf("frame %s: %v", fr.Mark, err)
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("frame %s is not an image: %v", fr.Mark, err)
		}
		if format != "png" {
			t.Errorf("frame %s is %s", fr.Mark, format)
		}
		if cfg.Width != 2400 || cfg.Height != 1600 {
			t.Errorf("frame %s is %dx%d, want 2400x1600 (1200x800 at 2x)", fr.Mark, cfg.Width, cfg.Height)
		}
	}

	// The focus box is what lets the scene push in on the prompt box rather
	// than the middle of the page. The input is 600px wide in a 1200px
	// viewport, 300px down an 800px one — so half the width, a bit under half
	// the height, centred horizontally.
	focus := res.Frames[0].Focus
	if focus == nil {
		t.Fatal("the focus selector resolved to no box")
	}
	if focus.W < 0.45 || focus.W > 0.55 {
		t.Errorf("focus width = %.3f, want ~0.5 (600px of 1200)", focus.W)
	}
	if focus.X < 0.2 || focus.X > 0.3 {
		t.Errorf("focus x = %.3f, want ~0.25 (centred 600px box)", focus.X)
	}
	if focus.Y < 0.3 || focus.Y > 0.45 {
		t.Errorf("focus y = %.3f, want ~0.375 (300px of 800)", focus.Y)
	}
	// The second shot named no focus, so it must carry none rather than a
	// guess — a frame with an invented focus pushes in on nothing in
	// particular, which reads as a mistake.
	if res.Frames[1].Focus != nil {
		t.Errorf("a shot with no focus selector got a box: %+v", res.Frames[1].Focus)
	}
}

const buildingPage = `<!doctype html>
<html><head><meta charset="utf-8"><style>
  html,body{margin:0;padding:0;background:#0d1117;color:#e6edf3;font-family:system-ui}
  #bar{height:40px;width:0;background:#3fb950;transition:none}
  #done{display:none;font-size:48px;padding:40px}
  .shown{display:block !important}
</style></head><body>
  <button id="go" data-testid="go" style="font-size:32px;margin:40px">build it</button>
  <div id="bar"></div>
  <div id="done">app ready</div>
  <script>
    document.getElementById('go').addEventListener('click', () => {
      let w = 0;
      const t = setInterval(() => {
        w += 4;
        document.getElementById('bar').style.width = w + '%';
        if (w >= 100) {
          clearInterval(t);
          document.getElementById('done').className = 'shown';
        }
      }, 40);
    });
  </script>
</body></html>`

// The video path against a real browser: a page that genuinely changes over
// time, recorded through CDP screencast and muxed by ffmpeg.
//
// This is the shot the course exists for — a sentence in, an application out —
// and none of it is provable from unit tests: the screencast only sends a frame
// when something changed, the frames carry their own arrival times, and the
// concat demuxer has opinions about the last one.
func TestWebTakeRecordsVideo(t *testing.T) {
	requireChromium(t)
	requireFFmpeg(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(buildingPage))
	}))
	defer srv.Close()
	registerTestSite(t, "testsite", srv.URL)

	take := &WebTake{Site: "testsite"}
	take.Viewport.Width = 800
	take.Viewport.Height = 600
	take.Steps = []WebStep{
		{Do: "goto", Path: "/"},
		{Do: "wait", Selector: Selector{"[data-testid=go]"}, Timeout: "20s"},
		{Do: "record"},
		{Do: "mark", Mark: "before"},
		{Do: "click", Selector: Selector{"[data-testid=go]"}},
		{Do: "wait", Selector: Selector{"#done.shown"}, Timeout: "30s"},
		{Do: "mark", Mark: "built"},
		{Do: "stop"},
	}
	if err := take.Validate(); err != nil {
		t.Fatalf("the take this test drives is itself invalid: %v", err)
	}

	outDir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	res, err := RunWebTake(ctx, &Env{}, take, outDir, "capture-1", true)
	if err != nil {
		t.Fatal(err)
	}

	if res.ClipPath != "capture-1.mp4" {
		t.Fatalf("clip = %q", res.ClipPath)
	}
	info, err := os.Stat(filepath.Join(outDir, res.ClipPath))
	if err != nil || info.Size() == 0 {
		t.Fatalf("clip not written: %v", err)
	}
	// The page takes ~1s to fill its bar, so the recording should be around
	// that — not zero, and not the whole test timeout.
	if res.ClipMs < 500 || res.ClipMs > 20000 {
		t.Errorf("clip is %dms, which does not look like the ~1s build it recorded", res.ClipMs)
	}

	// Marks here are *measured* against our own clock rather than modelled from
	// a script, which is the real difference from the terminal path — they are
	// never approximate, and they must be ordered and inside the clip.
	if len(res.Marks) != 2 || res.Marks[0].Name != "before" || res.Marks[1].Name != "built" {
		t.Fatalf("marks = %+v", res.Marks)
	}
	for _, m := range res.Marks {
		if m.Approximate {
			t.Errorf("mark %q is approximate; a measured mark never is", m.Name)
		}
		if m.AtMs < 0 || m.AtMs > res.ClipMs+1500 {
			t.Errorf("mark %q at %dms sits outside the %dms clip", m.Name, m.AtMs, res.ClipMs)
		}
	}
	if res.Marks[1].AtMs <= res.Marks[0].AtMs {
		t.Errorf("marks did not advance: %+v", res.Marks)
	}

	// A clip of the right length that is entirely black is the classic
	// screencast failure, and every assertion above would pass on one. The
	// whole point of recording rather than screenshotting is that something
	// changed, so prove it: the first and last frames must differ.
	clip := filepath.Join(outDir, res.ClipPath)
	first := extractFrame(t, clip, 0.05)
	last := extractFrame(t, clip, float64(res.ClipMs)/1000-0.3)
	if bytes.Equal(first, last) {
		t.Error("the first and last frames are identical — the recording caught no change")
	}
	if meanLuma(t, first) < 4 && meanLuma(t, last) < 4 {
		t.Error("the clip is essentially black; the screencast recorded nothing")
	}
}

// extractFrame pulls a single PNG out of a clip at the given second.
func extractFrame(t *testing.T, clip string, at float64) []byte {
	t.Helper()
	if at < 0 {
		at = 0
	}
	out := filepath.Join(t.TempDir(), "f.png")
	cmd := exec.Command("ffmpeg", "-v", "error", "-ss", fmt.Sprintf("%.3f", at),
		"-i", clip, "-frames:v", "1", "-y", out)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extracting a frame at %.2fs: %v\n%s", at, err, b)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// meanLuma is the average brightness of a PNG, 0-255.
func meanLuma(t *testing.T, png []byte) float64 {
	t.Helper()
	img, _, err := image.Decode(bytes.NewReader(png))
	if err != nil {
		t.Fatalf("decoding a frame: %v", err)
	}
	b := img.Bounds()
	var sum float64
	var n int
	for y := b.Min.Y; y < b.Max.Y; y += 4 {
		for x := b.Min.X; x < b.Max.X; x += 4 {
			r, g, bl, _ := img.At(x, y).RGBA()
			sum += 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(bl>>8)
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

// A take that starts recording and never stops is a mistake with an expensive
// failure mode — the clip runs until the whole capture times out.
func TestWebTakeRejectsAnUnclosedRecording(t *testing.T) {
	take := &WebTake{Site: "lovable", Steps: []WebStep{{Do: "record"}, {Do: "mark", Mark: "a"}}}
	err := take.Validate()
	if err == nil || !strings.Contains(err.Error(), "never stops") {
		t.Errorf("error = %v", err)
	}
}

// A mark outside a recording has no clock to be measured against. `shot` is the
// thing for a still, and saying so is more useful than silently dropping it.
func TestWebTakeRejectsAMarkOutsideARecording(t *testing.T) {
	take := &WebTake{Site: "lovable", Steps: []WebStep{{Do: "mark", Mark: "a"}}}
	err := take.Validate()
	if err == nil || !strings.Contains(err.Error(), "inside a recording") {
		t.Errorf("error = %v", err)
	}
}

// A step that cannot find its element must fail loudly and name the selector.
// This is the failure that will actually happen in production — somebody else's
// markup changed — and the message is the whole repair experience.
func TestWebTakeFailsUsefullyOnAMissingSelector(t *testing.T) {
	requireChromium(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(takeTestPage))
	}))
	defer srv.Close()
	registerTestSite(t, "testsite", srv.URL)

	take := &WebTake{Site: "testsite"}
	take.Steps = []WebStep{
		{Do: "goto", Path: "/"},
		{Do: "wait", Selector: Selector{"#this-was-renamed"}, Timeout: "2s"},
		{Do: "shot", Mark: "never"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, err := RunWebTake(ctx, nil, take, t.TempDir(), "capture-1", true)
	if err == nil {
		t.Fatal("a take waiting for an element that does not exist succeeded")
	}
	msg := err.Error()
	for _, want := range []string{"step 2", "#this-was-renamed", "may have changed"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q does not mention %q", msg, want)
		}
	}
}
