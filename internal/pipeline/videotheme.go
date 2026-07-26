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

// deriveVideoTheme fills the rich theme tokens from the course branding.
// The base hue comes from the primary colour so the dark background is
// tinted toward the course brand instead of a generic black.
func deriveVideoTheme(colors config.Colors, fonts config.Fonts, courseName string) SceneTheme {
	h, _, _ := hexToHSL(colors.Primary)

	t := SceneTheme{
		Primary:    colors.Primary,
		Accent:     colors.Accent,
		Background: colors.Background,
		CourseName: courseName,

		Mode:          "dark",
		BgTop:         hslToHex(h, 0.42, 0.11),
		BgBottom:      hslToHex(math.Mod(h+14, 360), 0.48, 0.055),
		Surface:       hslToHex(h, 0.30, 0.15),
		SurfaceBorder: hslToHex(h, 0.24, 0.26),
		Text:          hslToHex(h, 0.30, 0.96),
		TextMuted:     hslToHex(h, 0.16, 0.70),
		FontDisplay:   DefaultFontDisplay,
		FontBody:      DefaultFontBody,
		FontMono:      DefaultFontMono,
		Grain:         defaultGrain,
	}
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
	return deriveVideoTheme(colors, cfg.Branding.Fonts, courseName)
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
