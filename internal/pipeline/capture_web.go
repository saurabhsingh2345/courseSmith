package pipeline

// Web captures: real frames of somebody else's product.
//
// == Why this one is stills and not video ==
//
// A screenshot pipeline has no frame timing, no encoder and no screencast
// stream to lose, and for most of what a course needs to show — here is where
// the prompt box lives, here is what the pricing page says, here is the table
// with rows in it — a zoomed still with a callout teaches better than four
// seconds of video does. Video is for what is inherently temporal: an app
// assembling itself, a deploy going green. That comes later; this is the part
// that carries most of the course.
//
// == Why the take is checked in, and written by a person ==
//
// The terminal path lets a model write the tape, because the tape drives our
// own shell. This path drives *somebody else's DOM*, and a model cannot invent
// `[data-testid=prompt]` — it can only guess, confidently, and produce a take
// that fails at the first step or, worse, screenshots the wrong element.
//
// So a web take is a YAML file in the course, authored once and maintained.
// That is not a workaround, it is the freshness mechanism from the plan: these
// products redesign quarterly, and a checked-in take is the thing you re-run to
// find out that they did. A course full of hand-recorded video has no such
// thing, which is why it silently rots.
//
// The take is deliberately verbose — one explicit `do:` per step rather than a
// clever shorthand — because the failure it has to survive is a stranger
// changing their markup a year from now, and the person fixing it will not be
// the person who wrote it.

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"gopkg.in/yaml.v3"
)

// captureSite is one recordable web product. Separate from captureTools
// because the two gate different things: a terminal tool is a binary we allow
// a generated script to run, a site is an origin we allow a browser to be
// driven against and to hold a login for.
type captureSite struct {
	// Key is what a lesson writes in `tool=`.
	Key string
	// Display is how the product is named to a reader, spelled as its owner
	// spells it.
	Display string
	// Origin is the scheme and host every step must stay within. A take that
	// navigates off it is refused: the clip's provenance is the origin, and a
	// frame captured somewhere else is a frame captioned with a lie.
	Origin string
}

var captureSites = map[string]captureSite{
	"lovable": {Key: "lovable", Display: "Lovable", Origin: "https://lovable.dev"},
	"bolt":    {Key: "bolt", Display: "Bolt", Origin: "https://bolt.new"},
	"v0":      {Key: "v0", Display: "v0", Origin: "https://v0.app"},
	"figma":   {Key: "figma", Display: "Figma", Origin: "https://www.figma.com"},
}

// captureSiteKeys lists the sites in a stable order, for error messages.
func captureSiteKeys() []string {
	keys := make([]string, 0, len(captureSites))
	for k := range captureSites {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

// Selector is one or more CSS selectors, tried in order.
//
// A take drives somebody else's markup, and that markup changes. A single
// selector makes every redesign a broken take; a short list of fallbacks —
// the precise one first, a looser one behind it — usually survives, and when
// it does not the error names everything it tried, which is the difference
// between "fix this line" and "work out what this page looks like now".
//
// It accepts a plain string too, because most steps need only one and a list
// of one is noise:
//
//	selector: "textarea"
//	selector: ["[data-testid=prompt]", "textarea[placeholder*=describe]", "textarea"]
type Selector []string

// UnmarshalYAML accepts either a scalar or a sequence.
func (s *Selector) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var one string
		if err := value.Decode(&one); err != nil {
			return err
		}
		*s = Selector{one}
		return nil
	case yaml.SequenceNode:
		var many []string
		if err := value.Decode(&many); err != nil {
			return err
		}
		*s = Selector(many)
		return nil
	}
	return fmt.Errorf("a selector is a string or a list of strings")
}

// Empty reports whether no selector was given.
func (s Selector) Empty() bool {
	for _, v := range s {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}

// String renders the alternatives for an error message.
func (s Selector) String() string {
	if len(s) == 1 {
		return strconv.Quote(s[0])
	}
	quoted := make([]string, 0, len(s))
	for _, v := range s {
		quoted = append(quoted, strconv.Quote(v))
	}
	return strings.Join(quoted, " or ")
}

// WebStep is one action in a take.
//
// Every step names what it does in `do:` rather than being inferred from which
// field is set. Inference reads better in the common case and is unreadable in
// the case that matters — somebody opening a broken take to work out which step
// stopped matching.
type WebStep struct {
	// Do is goto | wait | click | type | shot | scroll.
	Do string `yaml:"do"`
	// Path is the path to navigate to, for `goto`. Relative to the site's
	// origin; an absolute URL off that origin is refused.
	Path string `yaml:"path,omitempty"`
	// Selector is the element to wait for, click, or type into. One selector,
	// or several tried in order.
	Selector Selector `yaml:"selector,omitempty"`
	// Text is what to type, for `type`.
	Text string `yaml:"text,omitempty"`
	// Mark names the frame, for `shot`. Same vocabulary as a tape's marks.
	Mark string `yaml:"mark,omitempty"`
	// Focus is the element a `shot` is really about. Its bounding box is
	// recorded with the frame, so the scene can push in on the prompt box
	// rather than on the middle of the page.
	Focus Selector `yaml:"focus,omitempty"`
	// Timeout overrides the default for `wait` ("30s", "3m").
	Timeout string `yaml:"timeout,omitempty"`
	// Pixels is how far to scroll, for `scroll`.
	Pixels int `yaml:"pixels,omitempty"`
}

// WebTake is a course's `takes/<name>.yaml`.
type WebTake struct {
	// Site is the captureSites key this take drives.
	Site string `yaml:"site"`
	// Viewport is the browser size; zero uses the default.
	Viewport struct {
		Width  int `yaml:"width"`
		Height int `yaml:"height"`
	} `yaml:"viewport"`
	Steps []WebStep `yaml:"steps"`
}

const (
	defaultWebViewportW = 1440
	defaultWebViewportH = 900
	defaultWebWait      = 30 * time.Second
)

// webStepVerbs is the closed vocabulary, listed for error messages.
var webStepVerbs = []string{"goto", "wait", "click", "type", "shot", "scroll", "record", "mark", "stop"}

// LoadWebTake reads and validates a take file.
func LoadWebTake(path string) (*WebTake, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading take %s: %w", path, err)
	}
	var t WebTake
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parsing take %s: %w", path, err)
	}
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &t, nil
}

// Validate checks a take before a browser is ever launched. Everything here is
// knowable without running anything, and a take that fails at step nine after
// two minutes of real page loads is a take that wasted two minutes.
func (t *WebTake) Validate() error {
	site, ok := captureSites[t.Site]
	if !ok {
		return fmt.Errorf("site %q is not recordable. Available: %s", t.Site, strings.Join(captureSiteKeys(), ", "))
	}
	if len(t.Steps) == 0 {
		return fmt.Errorf("take has no steps")
	}
	marks := map[string]bool{}
	shots := 0
	recording, recorded := false, false
	for i, s := range t.Steps {
		n := i + 1
		known := false
		for _, v := range webStepVerbs {
			if s.Do == v {
				known = true
				break
			}
		}
		if !known {
			return fmt.Errorf("step %d: %q is not a step. One of: %s", n, s.Do, strings.Join(webStepVerbs, ", "))
		}
		switch s.Do {
		case "goto":
			if s.Path == "" {
				return fmt.Errorf("step %d: goto needs a path", n)
			}
			if err := checkSameOrigin(site, s.Path); err != nil {
				return fmt.Errorf("step %d: %w", n, err)
			}
		case "wait", "click":
			if s.Selector.Empty() {
				return fmt.Errorf("step %d: %s needs a selector", n, s.Do)
			}
		case "type":
			if s.Selector.Empty() {
				return fmt.Errorf("step %d: type needs a selector", n)
			}
			if s.Text == "" {
				return fmt.Errorf("step %d: type needs text", n)
			}
		case "shot":
			if s.Mark == "" {
				return fmt.Errorf("step %d: shot needs a mark — a frame nothing can refer to cannot be cut to", n)
			}
			if !markNameRe.MatchString(s.Mark) {
				return fmt.Errorf("step %d: mark %q must be lowercase letters, digits and dashes", n, s.Mark)
			}
			if marks[s.Mark] {
				return fmt.Errorf("step %d: mark %q is used twice; a mark names one frame", n, s.Mark)
			}
			marks[s.Mark] = true
			shots++
		case "scroll":
			if s.Pixels == 0 {
				return fmt.Errorf("step %d: scroll needs pixels", n)
			}
		case "record":
			if recording {
				return fmt.Errorf("step %d: already recording; a take makes one clip", n)
			}
			recording = true
			recorded = true
		case "stop":
			if !recording {
				return fmt.Errorf("step %d: stop without a record", n)
			}
			recording = false
		case "mark":
			if !recording {
				return fmt.Errorf("step %d: a mark is a moment inside a recording — put it between record and stop, or use `shot` for a still", n)
			}
			if s.Mark == "" {
				return fmt.Errorf("step %d: mark needs a name", n)
			}
			if !markNameRe.MatchString(s.Mark) {
				return fmt.Errorf("step %d: mark %q must be lowercase letters, digits and dashes", n, s.Mark)
			}
			if marks[s.Mark] {
				return fmt.Errorf("step %d: mark %q is used twice; a mark names one moment", n, s.Mark)
			}
			marks[s.Mark] = true
		}
		if s.Timeout != "" {
			if _, err := time.ParseDuration(s.Timeout); err != nil {
				return fmt.Errorf("step %d: timeout %q is not a duration", n, s.Timeout)
			}
		}
	}
	if recording {
		return fmt.Errorf("the take starts recording and never stops")
	}
	if shots == 0 && !recorded {
		return fmt.Errorf("take takes no shots and records nothing — a capture that captures nothing is not a capture")
	}
	return nil
}

// markNameRe is the shared mark vocabulary, so a web mark and a tape mark
// cannot drift into two spellings of the same idea. A test pins it against the
// pattern markCommentRe accepts inside a tape.
var markNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// checkSameOrigin refuses a path that would leave the site.
//
// This is the web half of the provenance gate: `footage.json` records the
// origin as evidence, so a take that wanders to another host would produce a
// clip whose stated provenance is false. A relative path is always fine; an
// absolute URL has to match.
func checkSameOrigin(site captureSite, path string) error {
	if !strings.Contains(path, "://") {
		return nil
	}
	u, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("%q is not a URL: %w", path, err)
	}
	origin, err := url.Parse(site.Origin)
	if err != nil {
		return fmt.Errorf("site %s has an unparseable origin %q", site.Key, site.Origin)
	}
	if u.Scheme != origin.Scheme || u.Host != origin.Host {
		return fmt.Errorf("%q leaves %s. A capture of %s must be captured at %s, or the clip's provenance is a claim rather than a record",
			path, site.Origin, site.Display, site.Origin)
	}
	return nil
}

// browserProfileDir is where a site's login is kept between runs.
//
// One profile per site rather than one shared: they are separate logins with
// separate cookies, and a single profile means every capture runs as whoever
// logged in last.
func browserProfileDir(site captureSite) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding your home directory for the browser profile: %w", err)
	}
	return filepath.Join(home, ".coursesmith", "browser", site.Key), nil
}

// WebLogin opens a real browser window against a site so a person can sign in
// once, into the same profile every later capture reuses.
//
// This is the whole credential story and it is deliberately this dull. Nothing
// is stored in the repository, nothing is passed through an environment
// variable, and no automation ever types a password — the session cookie lives
// in a browser profile under the user's home directory, exactly as it would if
// they had opened Chrome themselves.
func WebLogin(ctx context.Context, out io.Writer, in io.Reader, siteKey string) error {
	site, ok := captureSites[siteKey]
	if !ok {
		return fmt.Errorf("%q is not a recordable site. Available: %s", siteKey, strings.Join(captureSiteKeys(), ", "))
	}
	profile, err := browserProfileDir(site)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(profile, 0o700); err != nil {
		return fmt.Errorf("creating browser profile dir: %w", err)
	}

	controlURL, err := launcher.New().UserDataDir(profile).Headless(false).Launch()
	if err != nil {
		return fmt.Errorf("launching a browser window: %w", err)
	}
	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("connecting to the browser: %w", err)
	}
	defer browser.Close()

	page, err := browser.Page(proto.TargetCreateTarget{URL: site.Origin})
	if err != nil {
		return fmt.Errorf("opening %s: %w", site.Origin, err)
	}
	_ = page

	fmt.Fprintf(out, "A browser window is open at %s.\n", site.Origin)
	fmt.Fprintf(out, "Sign in there, then press Enter here to save the session.\n")
	fmt.Fprintf(out, "The profile is kept at %s — captures reuse it and run headless.\n\n", profile)
	buf := make([]byte, 1)
	for {
		n, err := in.Read(buf)
		if err != nil || n == 0 {
			break
		}
		if buf[0] == '\n' {
			break
		}
	}
	fmt.Fprintf(out, "Saved. `coursesmith run <course>/<lesson> --stage demos` can now capture %s.\n", site.Display)
	return nil
}

// WebCaptureReadiness reports whether web captures can run, and which sites
// have a saved session.
//
// "Has a profile directory with something in it" is the only signal available
// without launching a browser and navigating, which would make `doctor` slow
// and online. It is a weak signal — a profile can exist with an expired
// session — so the check is worded as "signed in" rather than "will work", and
// the real failure it prevents is the common one: nobody ever ran `footage
// login` and the take quietly screenshots a login page.
func WebCaptureReadiness() (ready bool, signedIn, missing []string) {
	if _, has := launcher.LookPath(); !has {
		return false, nil, nil
	}
	for _, key := range captureSiteKeys() {
		profile, err := browserProfileDir(captureSites[key])
		if err != nil {
			missing = append(missing, key)
			continue
		}
		entries, err := os.ReadDir(profile)
		if err != nil || len(entries) == 0 {
			missing = append(missing, key)
			continue
		}
		signedIn = append(signedIn, key)
	}
	return true, signedIn, missing
}

// WebCaptureResult is what one take produced.
type WebCaptureResult struct {
	Frames []FootageFrame
	Origin string
	// ClipPath is set when the take recorded video, relative to outDir.
	ClipPath string
	// ClipMs and Marks describe that clip. Unlike a tape's marks these are
	// measured rather than modelled — we hold the clock here — so they are
	// never approximate.
	ClipMs int
	Marks  []FootageMark
}

// RunWebTake drives the take and writes one PNG per shot into outDir.
//
// frameName maps a mark to the file's name; the caller owns the naming so the
// clip's files sit beside the rest of the lesson's assets under its own id.
func RunWebTake(ctx context.Context, e *Env, take *WebTake, outDir, framePrefix string, headless bool) (*WebCaptureResult, error) {
	site := captureSites[take.Site]
	profile, err := browserProfileDir(site)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(profile, 0o700); err != nil {
		return nil, fmt.Errorf("creating browser profile dir: %w", err)
	}

	w, h := take.Viewport.Width, take.Viewport.Height
	if w == 0 {
		w = defaultWebViewportW
	}
	if h == 0 {
		h = defaultWebViewportH
	}

	controlURL, err := launcher.New().
		UserDataDir(profile).
		Headless(headless).
		Set("window-size", fmt.Sprintf("%d,%d", w, h)).
		Launch()
	if err != nil {
		return nil, fmt.Errorf("launching Chromium for the %s capture: %w", site.Display, err)
	}
	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connecting to Chromium: %w", err)
	}
	defer browser.Close()

	page, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, fmt.Errorf("opening a page: %w", err)
	}
	page = page.Context(ctx)
	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width: w, Height: h, DeviceScaleFactor: 2, Mobile: false,
	}); err != nil {
		return nil, fmt.Errorf("setting the viewport: %w", err)
	}

	res := &WebCaptureResult{Origin: site.Origin}
	rec := &webRecorder{page: page}
	defer func() { _ = rec.finish() }()

	for i, step := range take.Steps {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := runWebStep(page, site, step, outDir, framePrefix, res, rec, w, h); err != nil {
			return nil, fmt.Errorf("step %d (%s): %w", i+1, step.Do, err)
		}
	}
	if rec.frameCount() > 0 {
		if e == nil {
			return nil, fmt.Errorf("the take recorded video but no ffmpeg is configured to encode it")
		}
		clip := framePrefix + ".mp4"
		durMs, err := rec.encode(ctx, e, filepath.Join(outDir, clip))
		if err != nil {
			return nil, err
		}
		res.ClipPath, res.ClipMs, res.Marks = clip, durMs, rec.marks
	}
	return res, nil
}

// runWebStep performs one step. rod panics on protocol errors, so every entry
// point into it is wrapped.
func runWebStep(page *rod.Page, site captureSite, step WebStep, outDir, framePrefix string, res *WebCaptureResult, rec *webRecorder, vw, vh int) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("%v", rec)
		}
	}()

	timeout := defaultWebWait
	if step.Timeout != "" {
		if d, perr := time.ParseDuration(step.Timeout); perr == nil {
			timeout = d
		}
	}
	p := page.Timeout(timeout)

	switch step.Do {
	case "goto":
		target := step.Path
		if !strings.Contains(target, "://") {
			target = strings.TrimSuffix(site.Origin, "/") + "/" + strings.TrimPrefix(target, "/")
		}
		if err := p.Navigate(target); err != nil {
			return fmt.Errorf("navigating to %s: %w", target, err)
		}
		return p.WaitLoad()
	case "wait":
		if _, _, err := firstMatch(p, step.Selector); err != nil {
			return fmt.Errorf("waiting for %s: %w — the page may have changed since this take was written", step.Selector, err)
		}
		return nil
	case "click":
		el, _, err := firstMatch(p, step.Selector)
		if err != nil {
			return fmt.Errorf("finding %s: %w", step.Selector, err)
		}
		return el.Click(proto.InputMouseButtonLeft, 1)
	case "type":
		el, _, err := firstMatch(p, step.Selector)
		if err != nil {
			return fmt.Errorf("finding %s: %w", step.Selector, err)
		}
		if err := el.Focus(); err != nil {
			return err
		}
		return el.Input(step.Text)
	case "scroll":
		return p.Mouse.Scroll(0, float64(step.Pixels), 1)
	case "shot":
		shot, err := p.Screenshot(false, &proto.PageCaptureScreenshot{
			Format: proto.PageCaptureScreenshotFormatPng,
		})
		if err != nil {
			return fmt.Errorf("capturing the frame: %w", err)
		}
		name := framePrefix + "-" + step.Mark + ".png"
		if err := writeFileAtomic(filepath.Join(outDir, name), shot); err != nil {
			return err
		}
		frame := FootageFrame{Mark: step.Mark, Path: name}
		if !step.Focus.Empty() {
			if box := focusBoxOf(p, step.Focus, vw, vh); box != nil {
				frame.Focus = box
			}
		}
		res.Frames = append(res.Frames, frame)
		return nil
	case "record":
		return rec.start()
	case "mark":
		rec.mark(step.Mark)
		return nil
	case "stop":
		return rec.finish()
	}
	return fmt.Errorf("unknown step %q", step.Do)
}

// focusBoxOf resolves a focus selector to a normalized box.
//
// Normalized (0..1 of the viewport) rather than pixels, because the renderer
// scales the frame to the stage and a pixel box would be right only at the
// capture resolution. A focus that cannot be resolved is not an error: the
// frame is still worth having, it just gets a plain hold instead of a push-in.
func focusBoxOf(p *rod.Page, selector Selector, vw, vh int) *FocusBox {
	defer func() { _ = recover() }()
	el, _, err := firstMatch(p, selector)
	if err != nil {
		return nil
	}
	shape, err := el.Shape()
	if err != nil || len(shape.Quads) == 0 {
		return nil
	}
	box := shape.Box()
	if box == nil || box.Width <= 0 || box.Height <= 0 {
		return nil
	}
	return &FocusBox{
		X: clamp01(box.X / float64(vw)),
		Y: clamp01(box.Y / float64(vh)),
		W: clamp01(box.Width / float64(vw)),
		H: clamp01(box.Height / float64(vh)),
	}
}

// firstMatch resolves the first selector that matches, and reports which one.
//
// Fallbacks are what let a take outlive a redesign: the precise selector first,
// a looser one behind it. When none match, the error carries every alternative
// tried — the person repairing it a year from now needs to know what was
// already ruled out, not just that something broke.
func firstMatch(p *rod.Page, sel Selector) (*rod.Element, string, error) {
	if sel.Empty() {
		return nil, "", fmt.Errorf("no selector given")
	}
	var last error
	for _, one := range sel {
		one = strings.TrimSpace(one)
		if one == "" {
			continue
		}
		el, err := p.Element(one)
		if err == nil {
			return el, one, nil
		}
		last = err
	}
	if len(sel) > 1 {
		return nil, "", fmt.Errorf("none of %s matched (last error: %v)", sel, last)
	}
	return nil, "", last
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
