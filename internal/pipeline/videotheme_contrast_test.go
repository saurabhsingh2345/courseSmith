package pipeline

import (
	"math"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

// The theme tokens are derived by formula from one branding colour, which means
// nobody ever looks at the result for most courses — the first time a bad pair
// is seen is in a finished video. Worse, the derivation now has two branches,
// and the light one is the branch nobody's default config exercises.
//
// So the readable pairs are asserted rather than eyeballed, across the hue
// circle and in both modes. This is the gate that makes "light mode" a claim
// rather than a hope.

// relativeLuminance implements the WCAG 2.1 definition.
func relativeLuminance(hex string) float64 {
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

// hueDistance is the shorter way round the colour circle between two hues.
func hueDistance(a, b float64) float64 {
	d := math.Abs(a - b)
	if d > 180 {
		d = 360 - d
	}
	return d
}

// contrastRatio is the WCAG ratio between two colours, 1..21.
func contrastRatio(a, b string) float64 {
	la, lb := relativeLuminance(a), relativeLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func TestThemeContrastBothModes(t *testing.T) {
	// Hues right round the circle, plus a desaturated grey, because the
	// derivation takes its hue from whatever the course happens to brand with.
	primaries := []string{
		"#e11d48", "#f97316", "#eab308", "#22c55e",
		"#06b6d4", "#2563eb", "#7c3aed", "#db2777", "#64748b",
	}
	for _, mode := range []string{ThemeModeDark, ThemeModeLight} {
		for _, primary := range primaries {
			th := deriveVideoTheme(
				config.Colors{Primary: primary, Accent: "#ffd43b", Background: "#ffffff"},
				config.Fonts{}, "Contrast", mode)

			// Body text has to clear AAA on everything it is ever set on: the
			// two ends of the background gradient and a card.
			for _, bg := range []struct{ name, hex string }{
				{"bgTop", th.BgTop}, {"bgBottom", th.BgBottom}, {"surface", th.Surface},
			} {
				if got := contrastRatio(th.Text, bg.hex); got < 7 {
					t.Errorf("%s/%s: text %s on %s %s = %.2f:1, want >= 7 (WCAG AAA)",
						mode, primary, th.Text, bg.name, bg.hex, got)
				}
			}
			// Muted text carries captions and secondary lines, so AA is the
			// floor — below it the supporting line stops being readable at
			// distance, which is the whole point of a video.
			for _, bg := range []struct{ name, hex string }{
				{"bgTop", th.BgTop}, {"surface", th.Surface},
			} {
				if got := contrastRatio(th.TextMuted, bg.hex); got < 4.5 {
					t.Errorf("%s/%s: textMuted %s on %s %s = %.2f:1, want >= 4.5 (WCAG AA)",
						mode, primary, th.TextMuted, bg.name, bg.hex, got)
				}
			}
			// A card's hairline is not text, but it has to be *visible* against
			// the card it outlines or the surface has no edge at all.
			if got := contrastRatio(th.SurfaceBorder, th.Surface); got < 1.12 {
				t.Errorf("%s/%s: surfaceBorder %s on surface %s = %.2f:1, want >= 1.12 (a visible edge)",
					mode, primary, th.SurfaceBorder, th.Surface, got)
			}
			// And the card has to be visible against the PAGE. This pair was the
			// one nobody checked, and it is the pair that decides whether a frame
			// reads as composed or as empty: light mode shipped Surface at
			// lightness 1.0 on a page at 0.985, so every card was invisible and a
			// four-card frame looked like labels floating on nothing. Every other
			// token passed its assertion the whole time.
			//
			// 1.06 rather than the hairline's 1.12, because a card is a large
			// filled area and needs far less separation than a one-pixel line to
			// register — but it does need some.
			if got := contrastRatio(th.Surface, th.BgTop); got < 1.06 {
				t.Errorf("%s/%s: surface %s on bgTop %s = %.2f:1, want >= 1.06 — a card the eye cannot separate from the page is not a card",
					mode, primary, th.Surface, th.BgTop, got)
			}
			// The gradient must not become a hue shift. Rotating the hue for the
			// bottom stop is invisible on a dark stage and blatant on paper: +14
			// degrees off a blue primary landed in violet, and the lavender wash it
			// put across every light frame is most of why they read as stock.
			if mode == ThemeModeLight {
				topH, topS, _ := hexToHSL(th.BgTop)
				botH, botS, _ := hexToHSL(th.BgBottom)
				// Rotation alone is the wrong measure, and asserting on it directly
				// fails for the wrong reason: at 0.16 saturation and near-white
				// lightness the hue barely survives the round trip to hex, so a
				// cyan primary reads back 11 degrees apart when 6 were asked for.
				// That is quantisation, not a colour cast.
				//
				// What the eye actually sees is rotation TIMES saturation — a big
				// turn at no saturation is invisible, and the old palette was
				// visible because it had both. So the two are judged together, with
				// the +14-at-0.30 that shipped (4.2) failing and the +6-at-0.16
				// that replaced it (~1.0) passing comfortably.
				if cast := hueDistance(topH, botH) * botS; cast > 2.5 {
					t.Errorf("light/%s: the gradient turns %.0f degrees at %.2f saturation (%s to %s) — that product reads as a colour cast rather than as light",
						primary, hueDistance(topH, botH), botS, th.BgTop, th.BgBottom)
				}
				if botS > topS+0.02 {
					t.Errorf("light/%s: bgBottom is more saturated (%.2f) than bgTop (%.2f), which tints the page toward its own hue",
						primary, botS, topS)
				}
			}
			// The accent set as type has to be readable, whatever the course
			// branded with. A saturated yellow is very nearly the luminance of
			// paper, so this is the token that light mode lives or dies on.
			for _, bg := range []struct{ name, hex string }{
				{"bgTop", th.BgTop}, {"surface", th.Surface},
			} {
				if got := contrastRatio(th.AccentText, bg.hex); got < 4.5 {
					t.Errorf("%s/%s: accentText %s on %s %s = %.2f:1, want >= 4.5",
						mode, primary, th.AccentText, bg.name, bg.hex, got)
				}
			}
			// A figure's body has to read as a shape against the stage...
			if got := contrastRatio(th.Mass, th.BgTop); got < 1.7 {
				t.Errorf("%s/%s: mass %s on bgTop %s = %.2f:1, want >= 1.7 (a visible shape)",
					mode, primary, th.Mass, th.BgTop, got)
			}
			// ...and its shading has to actually darken it. Ink lighter than
			// Mass turns every shaded face into a highlight.
			if relativeLuminance(th.Ink) >= relativeLuminance(th.Mass) {
				t.Errorf("%s/%s: ink %s is not darker than mass %s — shading would lighten",
					mode, primary, th.Ink, th.Mass)
			}
		}
	}
}

// The two modes must emit the same token *names* — a scene asks for `surface`
// and must get something usable whichever mode it renders in. A mode that left
// a token empty would fall back to the renderer's dark defaults and put a dark
// card on a paper background.
func TestLightModeFillsEveryToken(t *testing.T) {
	light := deriveVideoTheme(
		config.Colors{Primary: "#2563eb", Accent: "#ffd43b", Background: "#ffffff"},
		config.Fonts{}, "Light", ThemeModeLight)
	for name, v := range map[string]string{
		"bgTop": light.BgTop, "bgBottom": light.BgBottom,
		"surface": light.Surface, "surfaceBorder": light.SurfaceBorder,
		"text": light.Text, "textMuted": light.TextMuted,
		"fontDisplay": light.FontDisplay, "fontBody": light.FontBody, "fontMono": light.FontMono,
	} {
		if v == "" {
			t.Errorf("light mode leaves %s empty; the renderer would fall back to its dark default", name)
		}
	}
	if light.Mode != ThemeModeLight {
		t.Errorf("mode = %q, want %q", light.Mode, ThemeModeLight)
	}
	// Light mode has to actually be light, or the name is a lie.
	if relativeLuminance(light.BgTop) < 0.6 {
		t.Errorf("light bgTop %s has luminance %.3f — that is not a light background",
			light.BgTop, relativeLuminance(light.BgTop))
	}
	dark := deriveVideoTheme(
		config.Colors{Primary: "#2563eb", Accent: "#ffd43b", Background: "#ffffff"},
		config.Fonts{}, "Dark", "")
	if relativeLuminance(dark.BgTop) > 0.15 {
		t.Errorf("default bgTop %s has luminance %.3f — the default must stay dark",
			dark.BgTop, relativeLuminance(dark.BgTop))
	}
}

func TestNormalizeThemeMode(t *testing.T) {
	for _, in := range []string{"", "dark", "DARK", "nonsense", "  "} {
		if got := normalizeThemeMode(in); got != ThemeModeDark {
			t.Errorf("normalizeThemeMode(%q) = %q, want dark", in, got)
		}
	}
	for _, in := range []string{"light", "Light", "  LIGHT  "} {
		if got := normalizeThemeMode(in); got != ThemeModeLight {
			t.Errorf("normalizeThemeMode(%q) = %q, want light", in, got)
		}
	}
}
