package pipeline

// Screenshot tech choice: rod (Go CDP driver) over playwright-in-tools/.
// Rationale: the engine core stays a single Go binary with no second
// toolchain required (the tools/ Python env remains optional, for whisperX
// only); rod auto-provisions a compatible headless Chromium on first use and
// caches it; and a screenshot failure degrades gracefully to the v1
// text-based SVG review, so the browser is an enhancement, not a dependency.

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Screenshotter rasterizes an SVG so a vision model can inspect the pixels.
type Screenshotter interface {
	Name() string
	ScreenshotSVG(ctx context.Context, svg []byte) ([]byte, error)
}

// RodScreenshotter renders SVGs in headless Chromium via rod. The browser
// is launched lazily on first use and reused.
type RodScreenshotter struct {
	mu      sync.Mutex
	browser *rod.Browser
}

func (r *RodScreenshotter) Name() string { return "headless-chromium (rod)" }

func (r *RodScreenshotter) getBrowser() (*rod.Browser, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.browser != nil {
		return r.browser, nil
	}
	path, err := launcher.New().Headless(true).Launch()
	if err != nil {
		return nil, fmt.Errorf("launching headless Chromium for diagram QA: %w", err)
	}
	browser := rod.New().ControlURL(path)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connecting to headless Chromium: %w", err)
	}
	r.browser = browser
	return browser, nil
}

// Close shuts the shared browser down (call once per process, optional).
func (r *RodScreenshotter) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.browser != nil {
		_ = r.browser.Close()
		r.browser = nil
	}
}

func (r *RodScreenshotter) ScreenshotSVG(ctx context.Context, svg []byte) (png []byte, err error) {
	browser, err := r.getBrowser()
	if err != nil {
		return nil, err
	}
	// rod panics on protocol errors; convert to error returns.
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("screenshotting diagram: %v", rec)
		}
	}()

	// The page background matches the dark video stage so the vision model
	// judges contrast in the context viewers actually see (diagrams are now
	// transparent, dark-themed SVGs).
	html := `<!DOCTYPE html><html><head><meta charset="utf-8"><style>
		html,body{margin:0;padding:0;background:#0c1524}
		svg{display:block;width:800px;height:auto}
	</style></head><body>` + string(svg) + `</body></html>`

	page, err := browser.Page(proto.TargetCreateTarget{
		URL: "data:text/html;charset=utf-8," + url.PathEscape(html),
	})
	if err != nil {
		return nil, fmt.Errorf("opening diagram page: %w", err)
	}
	defer page.Close()

	page = page.Timeout(20 * time.Second).Context(ctx)
	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("loading diagram page: %w", err)
	}
	el, err := page.Element("svg")
	if err != nil {
		return nil, fmt.Errorf("finding rendered svg: %w", err)
	}
	shot, err := el.Screenshot(proto.PageCaptureScreenshotFormatPng, 0)
	if err != nil {
		return nil, fmt.Errorf("capturing diagram screenshot: %w", err)
	}
	return shot, nil
}
