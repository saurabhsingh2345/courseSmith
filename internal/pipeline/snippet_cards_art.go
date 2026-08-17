package pipeline

// Fetching the marks a cards clip wears.
//
// This is the only stage in the pipeline that reaches out to the open web for
// something that ends up on screen, so it is worth saying why, and what it will
// not do.
//
// WHY. The cards template's whole claim is that the viewer identifies the thing
// before they read its name, and that only happens if the mark is the real one.
// Every alternative was worse. A closed vocabulary of hand-drawn logos is a
// vocabulary that is missing whatever anybody wants to talk about this week. An
// LLM asked to emit SVG for a brand mark produces a plausible wrong shape, which
// is the failure mode with no floor — a wrong logo is worse than no logo,
// because it reads as a real one. Asking the creator to supply the files is
// asking them to do the job the template exists to do.
//
// WHAT IT WILL NOT DO. Nothing here can fail a plan. A fetch that times out, a
// slug the model invented, a service that is down on the day — all of it lands
// on the same floor, which is the Lucide glyph every card carries anyway. A clip
// that renders with a generic icon is a clip; a clip that failed to plan because
// a CDN was slow is not.
//
// THE MARKS THEMSELVES are brand assets, fetched from Simple Icons (the SVGs are
// CC0; the trademarks are not, and are used here the way a comparison article
// uses them). Anything that is not a brand falls back to the site's own favicon,
// and then to a drawn glyph. The provenance travels with the card in `markFrom`
// so a wrong mark on a finished frame can be traced to the thing that served it.

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/enfec/coursesmith/internal/config"
)

// planCardsSnippet plans the clip and then goes and gets the art.
//
// The fetch sits after planning rather than inside the correction loop on
// purpose: what the model is being corrected on is the clip's shape, and
// re-fetching four logos on every round would spend real seconds proving the
// network still works. The plan is written to disk with the marks resolved, so a
// re-render never fetches at all.
func planCardsSnippet(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config) (*SnippetPlan, error) {
	plan, err := planSnippetDefault(ctx, e, spec, cfg)
	if err != nil {
		return nil, err
	}
	resolveCardArt(ctx, e, plan.Cards)
	return plan, nil
}

// cardArtTimeout is per fetch. Short on purpose: this is decoration on a clip
// that will render without it, and a creator waiting a minute for a logo has
// been charged more than the logo is worth.
const cardArtTimeout = 8 * time.Second

// cardArtMaxBytes caps what will be pulled down and inlined into the plan. A
// brand SVG is two kilobytes and a favicon is ten; anything past this is not the
// thing that was asked for.
const cardArtMaxBytes = 512 << 10

// cardArtFetch is the GET this file makes. A var so tests can serve bytes
// without a network, the way the router injects its providers.
var cardArtFetch = fetchCardArt

// cardArtCache memoizes fetches for the life of the process. A combo can cut
// twelve segments in one run and several of them may name the same tool.
var cardArtCache = struct {
	sync.Mutex
	got map[string]cardArtResult
}{got: map[string]cardArtResult{}}

type cardArtResult struct {
	body []byte
	mime string
	err  error
}

// resolveCardArt fills in each card's Mark or Image, best effort and in place.
//
// The order is a quality order, not a convenience one. A brand mark is a vector
// that takes the card's colour and stays sharp at any size. A favicon is a
// 32-pixel bitmap somebody made for a browser tab. A drawn glyph is neither, and
// it is always available.
func resolveCardArt(ctx context.Context, e *Env, c *CardsSpec) {
	if c == nil {
		return
	}
	for i := range c.Items {
		it := &c.Items[i]
		if it.Mark != "" || it.Image != "" {
			continue // already resolved — a re-plan of a plan that had its art
		}
		switch {
		case cardArtFromImage(ctx, it):
		case cardArtFromBrand(ctx, it):
		case cardArtFromSite(ctx, it):
		}
		if it.MarkFrom != "" {
			fmt.Fprintf(e.out(), "    mark      %s ← %s\n", it.Title, it.MarkFrom)
			continue
		}
		// Said out loud rather than left to be discovered in the finished frame:
		// a generic glyph where a logo was expected is the one outcome here that
		// looks like a bug and is not.
		fmt.Fprintf(e.out(), "    mark      %s — no logo found, drawing the %s glyph\n", it.Title, it.ResolvedIcon())
	}
}

// cardArtFromImage takes a URL the model supplied directly. Last resort in
// quality terms and first in precedence, because a creator or a model that named
// an exact image meant that image.
func cardArtFromImage(ctx context.Context, it *Card) bool {
	u := it.ResolvedImageURL()
	if u == "" {
		return false
	}
	body, mime, err := cardArtFetch(ctx, u)
	if err != nil || !strings.HasPrefix(mime, "image/") {
		return false
	}
	it.Image = dataURI(mime, body)
	it.MarkFrom = "url:" + u
	return true
}

// cardArtFromBrand fetches the real brand mark: a single-path SVG on a 24x24
// grid, which the renderer paints in the card's own colour.
func cardArtFromBrand(ctx context.Context, it *Card) bool {
	slug := cardsSlug(it.Brand)
	if slug == "" {
		return false
	}
	body, _, err := cardArtFetch(ctx, "https://cdn.simpleicons.org/"+slug)
	if err != nil {
		return false
	}
	d := svgPathData(body)
	if d == "" {
		return false
	}
	it.Mark = d
	it.MarkFrom = "simpleicons:" + slug
	return true
}

// cardArtFromSite falls back to the site's favicon.
//
// A bitmap made for a 16-pixel browser tab, blown up to sit on a card — which is
// exactly as good as it sounds, and still better than a generic box for a tool
// that never made it into an icon set. Google's service is used rather than the
// site's own /favicon.ico because it already did the work of finding whichever
// of the six places a site hides its icon.
func cardArtFromSite(ctx context.Context, it *Card) bool {
	host := cardsHost(it.Site)
	if host == "" {
		return false
	}
	u := "https://www.google.com/s2/favicons?sz=128&domain=" + url.QueryEscape(host)
	body, mime, err := cardArtFetch(ctx, u)
	if err != nil || !strings.HasPrefix(mime, "image/") {
		return false
	}
	it.Image = dataURI(mime, body)
	it.MarkFrom = "favicon:" + host
	return true
}

// cardsSlug normalizes a Simple Icons slug: lowercase, letters and digits only.
// Their slugs are formed that way ("googlegemini", "visualstudiocode"), so a
// model writing "Google Gemini" or "visual-studio-code" resolves rather than
// 404s.
func cardsSlug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	if b.Len() > 40 {
		return ""
	}
	return b.String()
}

// cardsHost normalizes a domain: "https://www.anthropic.com/claude" and
// "anthropic.com" are the same answer.
func cardsHost(s string) string {
	h := strings.ToLower(strings.TrimSpace(s))
	h = strings.TrimPrefix(strings.TrimPrefix(h, "https://"), "http://")
	if i := strings.IndexAny(h, "/?#"); i >= 0 {
		h = h[:i]
	}
	h = strings.TrimPrefix(h, "www.")
	if h == "" || !strings.Contains(h, ".") || strings.ContainsAny(h, " \t") {
		return ""
	}
	return h
}

var (
	svgViewBoxRe = regexp.MustCompile(`viewBox\s*=\s*"([^"]*)"`)
	svgPathRe    = regexp.MustCompile(`<path[^>]*\sd\s*=\s*"([^"]*)"`)
)

// svgPathData pulls the drawable geometry out of a Simple Icons SVG.
//
// The path is taken rather than the whole document because the document arrives
// with the brand's colour baked into a fill attribute, and a black mark on a
// near-black stage is an empty card. Extracting the geometry and letting the
// renderer choose the colour is what makes a fetched logo obey the theme like
// everything else on the frame.
//
// Anything not on the 24x24 grid is refused rather than scaled: the renderer
// draws these at one size on one viewBox, and a mark from a different grid would
// land somewhere off the card with no sign that anything went wrong.
func svgPathData(body []byte) string {
	src := string(body)
	if m := svgViewBoxRe.FindStringSubmatch(src); m == nil || strings.Join(strings.Fields(m[1]), " ") != "0 0 24 24" {
		return ""
	}
	var ds []string
	for _, m := range svgPathRe.FindAllStringSubmatch(src, -1) {
		if d := strings.TrimSpace(m[1]); d != "" {
			ds = append(ds, d)
		}
	}
	return strings.Join(ds, " ")
}

func dataURI(mime string, body []byte) string {
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(body)
}

// fetchCardArt is the real GET, memoized. Errors are cached too: a slug that
// 404s once will 404 for every other segment in the same run, and re-asking is
// twelve round trips to learn the same thing.
func fetchCardArt(ctx context.Context, rawURL string) ([]byte, string, error) {
	cardArtCache.Lock()
	got, hit := cardArtCache.got[rawURL]
	cardArtCache.Unlock()
	if hit {
		return got.body, got.mime, got.err
	}

	body, mime, err := getCardArt(ctx, rawURL)
	cardArtCache.Lock()
	cardArtCache.got[rawURL] = cardArtResult{body: body, mime: mime, err: err}
	cardArtCache.Unlock()
	return body, mime, err
}

func getCardArt(ctx context.Context, rawURL string) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(ctx, cardArtTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	// Some CDNs serve a different asset, or nothing at all, to a client with no
	// agent string.
	req.Header.Set("User-Agent", "coursesmith/1 (+https://github.com/enfec/coursesmith)")
	req.Header.Set("Accept", "image/svg+xml,image/*")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("fetching %s: %s", rawURL, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, cardArtMaxBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(body) > cardArtMaxBytes {
		return nil, "", fmt.Errorf("fetching %s: over %d bytes", rawURL, cardArtMaxBytes)
	}
	mime := resp.Header.Get("Content-Type")
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	if mime == "" {
		mime = http.DetectContentType(body)
	}
	return body, mime, nil
}
