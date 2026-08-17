package pipeline

// Video theme derivation: the scene graph carries a rich, Go-owned set of
// visual tokens (dark editorial background gradient, surfaces, type stack)
// derived deterministically from the course's three branding colours, so
// every course gets a cohesive broadcast look without new required config.
//
// The renderer treats every new field as optional (old lesson-video.json
// files keep rendering); Go is the single source of truth for derivation.

import (
	"fmt"
	"math"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

// Default video type stack (Google Fonts, loaded by the renderer via
// @remotion/google-fonts). Overridable per course via branding.fonts.
const (
	DefaultFontDisplay = "Space Grotesk"
	DefaultFontBody    = "Inter"
	DefaultFontMono    = "JetBrains Mono"
)

// defaultGrain is the film-grain overlay opacity (masks H.264 banding in
// dark gradients; ~4% reads as texture, not noise).
const defaultGrain = 0.04

// The two polarities a video can be rendered in.
const (
	ThemeModeDark  = "dark"
	ThemeModeLight = "light"
)

// normalizeThemeMode maps config input onto a known mode. Anything
// unrecognised — including the empty string — is dark, which is the default
// look and the one every existing scene graph was recorded against.
func normalizeThemeMode(mode string) string {
	if strings.EqualFold(strings.TrimSpace(mode), ThemeModeLight) {
		return ThemeModeLight
	}
	return ThemeModeDark
}

// relLuminance is the WCAG 2.1 relative luminance of a hex colour.
func relLuminance(hex string) float64 {
	r, g, b, err := parseHex(hex)
	if err != nil {
		return 0
	}
	lin := func(c float64) float64 {
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

// contrast is the WCAG ratio between two hex colours, 1..21.
func contrast(a, b string) float64 {
	la, lb := relLuminance(a), relLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// readableOn darkens `hex` until it clears `want` contrast against `bg`.
//
// This exists because a brand accent is chosen to sit on a dark stage, and the
// usual choice — a saturated yellow or orange — is very nearly the luminance of
// paper. Used as *text* in light mode it is simply illegible, and no amount of
// weight or size fixes a 1.4:1 ratio. Rather than forbid such accents or ask
// courses to configure a second one, the readable variant is derived: same hue,
// same saturation, walked down in lightness until it passes.
//
// The accent is still used unmodified for fills and strokes, where it sits on
// its own shape rather than on the background. Only text takes this.
func readableOn(hex, bg string, want float64) string {
	h, s, l := hexToHSL(hex)
	out := hex
	for i := 0; i < 50 && contrast(out, bg) < want; i++ {
		l -= 0.02
		if l < 0.04 {
			l = 0.04
			out = hslToHex(h, s, l)
			break
		}
		out = hslToHex(h, s, l)
	}
	return out
}

// deriveVideoTheme fills the rich theme tokens from the course branding.
// The base hue comes from the primary colour so the background is tinted
// toward the course brand instead of a generic black or a generic white.
//
// Mode picks which set of lightness targets the same hue is run through. It is
// deliberately the *only* thing that varies: both modes emit exactly the same
// token names, so a scene never asks which mode it is in — it asks for
// `surface` and gets something it can draw on either way. Any scene that has
// to branch on mode is a scene with a colour hardcoded in it.
// deriveVideoTheme keeps its original signature and its original result: it is
// the default skin, and every call site that does not care about skins keeps
// working unchanged. deriveVideoThemeSkinned is the same derivation with the
// house style applied on top.
func deriveVideoTheme(colors config.Colors, fonts config.Fonts, courseName, mode string) SceneTheme {
	return deriveVideoThemeSkinned(colors, fonts, courseName, mode, SkinDefault, "")
}

// deriveVideoThemeSkinned derives the theme and then lets the skin override the
// tokens it disagrees with. watermark is the standing corner mark; empty falls
// back to the course name for skins that carry chrome.
func deriveVideoThemeSkinned(colors config.Colors, fonts config.Fonts, courseName, mode, skin, watermark string) SceneTheme {
	t := deriveBaseVideoTheme(colors, fonts, courseName, mode)
	h, _, _ := hexToHSL(colors.Primary)
	applySkin(&t, h, skin, normalizeThemeMode(mode))
	t.Air = roundTo(skinAir(skin), 3)
	// Only a skin that actually draws chrome carries a watermark. `minimal` is
	// defined by the absence of furniture, so setting a mark it never renders
	// would be a token that lies about the frame.
	if normalizeSkin(skin) == SkinBroadcast {
		if strings.TrimSpace(watermark) != "" {
			t.Watermark = strings.TrimSpace(watermark)
		} else {
			t.Watermark = courseName
		}
	}
	// The default skin derives the semantic accents too — a template that draws
	// a limit being crossed needs those colours whichever house style it is cut
	// in, and they are new tokens rather than changed ones, so nothing that
	// rendered before this looks different for having them.
	if normalizeSkin(skin) == SkinDefault {
		deriveSemanticAccents(&t, normalizeThemeMode(mode))
	}
	return t
}

func deriveBaseVideoTheme(colors config.Colors, fonts config.Fonts, courseName, mode string) SceneTheme {
	h, _, _ := hexToHSL(colors.Primary)

	t := SceneTheme{
		Primary:     colors.Primary,
		Accent:      colors.Accent,
		Background:  colors.Background,
		CourseName:  courseName,
		FontDisplay: DefaultFontDisplay,
		FontBody:    DefaultFontBody,
		FontMono:    DefaultFontMono,
	}
	if normalizeThemeMode(mode) == ThemeModeLight {
		t.Mode = ThemeModeLight
		// Paper, not white. A pure #ffffff stage blows out against any
		// saturated brand colour and leaves the artwork looking pasted on; a
		// few points of the brand hue at very high lightness reads as stock.
		//
		// The page sits at 0.955 rather than 0.985, and that is the fix for the
		// single worst thing about light mode. Surface was 1.0 — the top of the
		// range, with nowhere further to go — against a page at 0.985, which is
		// one and a half points of separation. Cards were invisible: a showcase
		// frame read as four labels floating on an empty page, and the whole
		// composition looked like it had failed to load. Dark mode never had this
		// problem because its 0.11-to-0.15 gap sits where the eye is most
		// sensitive, so the same four points read clearly.
		//
		// Lowering the page rather than darkening the card, because in light mode
		// a card should be the bright thing. White-on-grey is what every light
		// interface does, and it is the direction that keeps text contrast rising
		// instead of falling.
		t.BgTop = hslToHex(h, 0.30, 0.955)
		// +6 degrees, not +14, and 0.16 saturation rather than 0.30.
		//
		// The rotation is invisible in dark mode at lightness 0.055 and blatant
		// here: fourteen degrees off a blue primary lands in violet, and 0.30
		// saturation at 0.94 lightness is enough to see it. That lavender wash
		// across every light frame is most of why they read as a stock template
		// rather than as a deliberate palette. A gradient still helps the frame
		// feel lit; it does not have to change hue to do it.
		t.BgBottom = hslToHex(math.Mod(h+6, 360), 0.16, 0.925)
		t.Surface = hslToHex(h, 0.26, 1.0)
		// A hairline that is actually visible on white. 0.86 cleared the 1.12
		// ratio the test asks for and no more, which is a border you can only find
		// if you already know it is there.
		t.SurfaceBorder = hslToHex(h, 0.20, 0.82)
		t.Text = hslToHex(h, 0.42, 0.13)
		// 0.42 was the first guess and it fails AA on green and yellow hues:
		// HSL lightness is not perceptual, so the same number carries much more
		// luminance at h=60-140 than at h=240. The contrast test picks the
		// floor; this is the value that clears it right around the hue circle.
		t.TextMuted = hslToHex(h, 0.20, 0.37)
		// A mass has to be darker than paper to be a shape at all. 0.58 rather
		// than 0.63: with the page lowered to 0.955 the old value kept its ratio
		// but a drawn mass wants to read as solid rather than as a tint, and on
		// paper that means committing to being noticeably darker.
		t.Mass = hslToHex(h, 0.22, 0.58)
		t.Ink = hslToHex(h, 0.45, 0.20)
		// Grain masks H.264 banding across a dark gradient. A light gradient
		// bands far less and the same grain reads as dirt on the paper.
		t.Grain = defaultGrain / 4
	} else {
		t.Mode = ThemeModeDark
		t.BgTop = hslToHex(h, 0.42, 0.11)
		t.BgBottom = hslToHex(math.Mod(h+14, 360), 0.48, 0.055)
		t.Surface = hslToHex(h, 0.30, 0.15)
		t.SurfaceBorder = hslToHex(h, 0.24, 0.26)
		t.Text = hslToHex(h, 0.30, 0.96)
		t.TextMuted = hslToHex(h, 0.16, 0.70)
		// On the dark stage a mass is near-white, so the same Ink darkens it.
		t.Mass = hslToHex(h, 0.28, 0.90)
		t.Ink = hslToHex(h, 0.55, 0.06)
		t.Grain = defaultGrain
	}
	// The accent as *text*, guaranteed readable on this mode's background. On
	// the dark stage a bright accent already clears it and comes back
	// unchanged; on paper it is walked down until it does.
	t.AccentText = readableOn(t.Accent, t.BgTop, 4.5)
	// How an object is seated on this background. Derived in the base rather than
	// per skin so a scene can reach for a shadow without first asking which house
	// style it is in — the same reason the semantic accents are derived for every
	// skin. See deriveElevation.
	deriveElevation(&t, h, normalizeThemeMode(mode))
	if fonts.Display != "" {
		t.FontDisplay = fonts.Display
	}
	if fonts.Body != "" {
		t.FontBody = fonts.Body
	}
	if fonts.Mono != "" {
		t.FontMono = fonts.Mono
	}
	return t
}

// videoThemeForConfig derives the video theme the way the scenegraph stage
// does — branding colours unless the archetype supplies a palette — so
// stages that theme artifacts earlier in the pipeline (visuals) agree with
// the final scene look.
func videoThemeForConfig(cfg config.Config, courseName string) SceneTheme {
	colors := cfg.Branding.Colors
	if arch, err := ResolveArchetype(cfg.Style); err == nil && arch.HasPalette {
		colors = arch.Palette
	}
	return deriveVideoThemeSkinned(colors, cfg.Branding.Fonts, courseName,
		cfg.Style.Mode, cfg.Style.Skin, cfg.Style.Watermark)
}

// shiftLightness returns the hex colour moved by dl in HSL lightness
// (clamped to [0,1]) — used to derive secondary/tertiary diagram surfaces
// from the theme surface colour.
func shiftLightness(hex string, dl float64) string {
	h, s, l := hexToHSL(hex)
	l += dl
	if l < 0 {
		l = 0
	}
	if l > 1 {
		l = 1
	}
	return hslToHex(h, s, l)
}

// hexToHSL parses #rgb/#rrggbb into HSL (h in degrees, s/l in 0..1).
// Unparseable input degrades to a neutral slate hue.
func hexToHSL(hex string) (h, s, l float64) {
	r, g, b, err := parseHex(hex)
	if err != nil {
		return 220, 0.3, 0.4
	}
	maxC := math.Max(r, math.Max(g, b))
	minC := math.Min(r, math.Min(g, b))
	l = (maxC + minC) / 2
	d := maxC - minC
	if d == 0 {
		return 0, 0, l
	}
	if l > 0.5 {
		s = d / (2 - maxC - minC)
	} else {
		s = d / (maxC + minC)
	}
	switch maxC {
	case r:
		h = math.Mod((g-b)/d, 6)
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	return h, s, l
}

// hslToHex converts HSL (h degrees, s/l 0..1) to a #rrggbb string.
func hslToHex(h, s, l float64) string {
	c := (1 - math.Abs(2*l-1)) * s
	hp := math.Mod(h, 360) / 60
	x := c * (1 - math.Abs(math.Mod(hp, 2)-1))
	var r, g, b float64
	switch {
	case hp < 1:
		r, g, b = c, x, 0
	case hp < 2:
		r, g, b = x, c, 0
	case hp < 3:
		r, g, b = 0, c, x
	case hp < 4:
		r, g, b = 0, x, c
	case hp < 5:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	m := l - c/2
	toByte := func(v float64) int {
		n := int(math.Round((v + m) * 255))
		if n < 0 {
			n = 0
		}
		if n > 255 {
			n = 255
		}
		return n
	}
	return fmt.Sprintf("#%02x%02x%02x", toByte(r), toByte(g), toByte(b))
}

// parseHex reads #rgb or #rrggbb into 0..1 channels.
func parseHex(hex string) (r, g, b float64, err error) {
	s := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return 0, 0, 0, fmt.Errorf("bad hex color %q", hex)
	}
	var ri, gi, bi int
	if _, err := fmt.Sscanf(s, "%02x%02x%02x", &ri, &gi, &bi); err != nil {
		return 0, 0, 0, fmt.Errorf("bad hex color %q: %w", hex, err)
	}
	return float64(ri) / 255, float64(gi) / 255, float64(bi) / 255, nil
}
