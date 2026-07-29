package pipeline

import (
	"encoding/json"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

// The promise this file exists to keep is that skins are additive. A course
// that never mentions `skin` must get exactly the video it got before skins
// existed — not "close enough", byte-identical — because the alternative is
// that adding a house style silently restyles every course already published.

var skinTestPrimaries = []string{
	"#e11d48", "#f97316", "#eab308", "#22c55e",
	"#06b6d4", "#2563eb", "#7c3aed", "#db2777", "#64748b",
}

func TestDefaultSkinIsUnchanged(t *testing.T) {
	for _, mode := range []string{ThemeModeDark, ThemeModeLight} {
		for _, primary := range skinTestPrimaries {
			colors := config.Colors{Primary: primary, Accent: "#ffd43b", Background: "#ffffff"}
			// The pre-skin derivation, reached through the original signature.
			base := deriveBaseVideoTheme(colors, config.Fonts{}, "Course", mode)
			// What a course with no skin setting gets today.
			got := deriveVideoThemeSkinned(colors, config.Fonts{}, "Course", mode, "", "")

			// Every token the old derivation produced must survive untouched.
			// Compared field by field rather than with reflect.DeepEqual so a
			// failure names the token that moved.
			for name, pair := range map[string][2]string{
				"bgTop": {base.BgTop, got.BgTop}, "bgBottom": {base.BgBottom, got.BgBottom},
				"surface": {base.Surface, got.Surface}, "surfaceBorder": {base.SurfaceBorder, got.SurfaceBorder},
				"text": {base.Text, got.Text}, "textMuted": {base.TextMuted, got.TextMuted},
				"mass": {base.Mass, got.Mass}, "ink": {base.Ink, got.Ink},
				"accent": {base.Accent, got.Accent}, "accentText": {base.AccentText, got.AccentText},
				"mode": {base.Mode, got.Mode},
			} {
				if pair[0] != pair[1] {
					t.Errorf("%s/%s: default skin moved %s from %s to %s — skins must be additive",
						mode, primary, name, pair[0], pair[1])
				}
			}
			if base.Grain != got.Grain {
				t.Errorf("%s/%s: default skin moved grain from %v to %v", mode, primary, base.Grain, got.Grain)
			}
			// And it must not switch on chrome nobody asked for.
			if got.Watermark != "" {
				t.Errorf("%s/%s: default skin set a watermark %q", mode, primary, got.Watermark)
			}
			if got.Air != 0 {
				t.Errorf("%s/%s: default skin set air %v, want 0 (fill the stage)", mode, primary, got.Air)
			}
			// The field stays empty so `omitempty` drops it; the resolved value
			// is still the default.
			if got.Skin != "" {
				t.Errorf("%s/%s: skin field = %q, want empty so omitempty drops it", mode, primary, got.Skin)
			}
			if got.ResolvedSkin() != SkinDefault {
				t.Errorf("%s/%s: resolved skin = %q, want %q", mode, primary, got.ResolvedSkin(), SkinDefault)
			}
		}
	}
}

// A scene graph from a course with no skin must not grow a `skin` key. This is
// the serialisation half of the additive promise: config fingerprints and
// golden files are taken over this JSON.
func TestDefaultSkinAddsNoJSONKeys(t *testing.T) {
	th := deriveVideoThemeSkinned(
		config.Colors{Primary: "#2563eb", Accent: "#ffd43b", Background: "#fff"},
		config.Fonts{}, "Course", "", "", "")
	raw, err := json.Marshal(th)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"skin", "air", "watermark"} {
		if _, present := m[key]; present {
			t.Errorf("an unskinned theme serialises %q — omitempty is not holding, and every recorded fingerprint moves", key)
		}
	}
}

// Every skin has to derive in both modes and right round the hue circle. A skin
// that only reads on slate is a hardcoded look wearing a token's name.
func TestSkinsDeriveEveryTokenInBothModes(t *testing.T) {
	for _, skin := range SkinNames() {
		for _, mode := range []string{ThemeModeDark, ThemeModeLight} {
			for _, primary := range skinTestPrimaries {
				th := deriveVideoThemeSkinned(
					config.Colors{Primary: primary, Accent: "#ffd43b", Background: "#fff"},
					config.Fonts{}, "Course", mode, skin, "")

				for name, v := range map[string]string{
					"bgTop": th.BgTop, "bgBottom": th.BgBottom, "surface": th.Surface,
					"surfaceBorder": th.SurfaceBorder, "text": th.Text, "textMuted": th.TextMuted,
					"mass": th.Mass, "ink": th.Ink, "accentText": th.AccentText,
					"accentQuantity": th.AccentQuantity, "accentLimit": th.AccentLimit,
					"accentRival": th.AccentRival,
				} {
					if v == "" {
						t.Errorf("%s/%s/%s: %s is empty — the renderer would fall back to a dark default",
							skin, mode, primary, name)
					}
				}

				// Body text clears AAA on both ends of the backdrop and on a card.
				for _, bg := range []struct{ name, hex string }{
					{"bgTop", th.BgTop}, {"bgBottom", th.BgBottom}, {"surface", th.Surface},
				} {
					if got := contrastRatio(th.Text, bg.hex); got < 7 {
						t.Errorf("%s/%s/%s: text %s on %s %s = %.2f:1, want >= 7 (AAA)",
							skin, mode, primary, th.Text, bg.name, bg.hex, got)
					}
				}
				// Muted text carries the eyebrow and the takeaway line, which in
				// the broadcast skin are the only words on screen besides the
				// headline. AA is the floor.
				for _, bg := range []struct{ name, hex string }{
					{"bgTop", th.BgTop}, {"surface", th.Surface},
				} {
					if got := contrastRatio(th.TextMuted, bg.hex); got < 4.5 {
						t.Errorf("%s/%s/%s: textMuted %s on %s %s = %.2f:1, want >= 4.5 (AA)",
							skin, mode, primary, th.TextMuted, bg.name, bg.hex, got)
					}
				}
				if got := contrastRatio(th.SurfaceBorder, th.Surface); got < 1.12 {
					t.Errorf("%s/%s/%s: surfaceBorder %s on surface %s = %.2f:1 — the card has no edge",
						skin, mode, primary, th.SurfaceBorder, th.Surface, got)
				}
				if got := contrastRatio(th.Mass, th.BgTop); got < 1.7 {
					t.Errorf("%s/%s/%s: mass %s on bgTop %s = %.2f:1 — artwork is not a shape",
						skin, mode, primary, th.Mass, th.BgTop, got)
				}
				if relativeLuminance(th.Ink) >= relativeLuminance(th.Mass) {
					t.Errorf("%s/%s/%s: ink %s is not darker than mass %s — shading would lighten",
						skin, mode, primary, th.Ink, th.Mass)
				}
			}
		}
	}
}

// The semantic accents are the load-bearing part of the new palette: a template
// draws a quantity in one, the ceiling it hits in another, and the alternative
// in a third. If any of them fails to read, the picture stops saying anything.
func TestSemanticAccentContrast(t *testing.T) {
	for _, skin := range SkinNames() {
		for _, mode := range []string{ThemeModeDark, ThemeModeLight} {
			for _, primary := range skinTestPrimaries {
				th := deriveVideoThemeSkinned(
					config.Colors{Primary: primary, Accent: "#ffd43b", Background: "#fff"},
					config.Fonts{}, "Course", mode, skin, "")
				for name, hex := range map[string]string{
					"accentQuantity": th.AccentQuantity,
					"accentLimit":    th.AccentLimit,
					"accentRival":    th.AccentRival,
				} {
					if got := contrastRatio(hex, th.BgTop); got < 4.5 {
						t.Errorf("%s/%s/%s: %s %s on bgTop %s = %.2f:1, want >= 4.5 (AA)",
							skin, mode, primary, name, hex, th.BgTop, got)
					}
				}
				// The three roles have to be distinguishable from each other, or
				// colouring by meaning conveys no meaning. Compared by hue
				// distance, since a colourblind-safe check is a different and
				// larger question than "these are not the same colour".
				for _, pair := range [][2]string{
					{th.AccentQuantity, th.AccentLimit},
					{th.AccentQuantity, th.AccentRival},
					{th.AccentLimit, th.AccentRival},
				} {
					if pair[0] == pair[1] {
						t.Errorf("%s/%s/%s: two semantic accents are both %s — the roles are indistinguishable",
							skin, mode, primary, pair[0])
					}
				}
			}
		}
	}
}

// The semantic accents must not follow the branding hue. This is the rule that
// keeps a red "does not fit" bar red on a course branded green.
func TestSemanticAccentsIgnoreBranding(t *testing.T) {
	var first SceneTheme
	for i, primary := range skinTestPrimaries {
		th := deriveVideoThemeSkinned(
			config.Colors{Primary: primary, Accent: "#ffd43b", Background: "#fff"},
			config.Fonts{}, "Course", ThemeModeDark, SkinBroadcast, "")
		if i == 0 {
			first = th
			continue
		}
		if th.AccentLimit != first.AccentLimit {
			t.Errorf("branding %s moved accentLimit to %s (was %s) — a failure state is red whatever the course is branded with",
				primary, th.AccentLimit, first.AccentLimit)
		}
		if th.AccentQuantity != first.AccentQuantity {
			t.Errorf("branding %s moved accentQuantity to %s (was %s)",
				primary, th.AccentQuantity, first.AccentQuantity)
		}
	}
}

func TestBroadcastSkinIsNearBlack(t *testing.T) {
	th := deriveVideoThemeSkinned(
		config.Colors{Primary: "#2563eb", Accent: "#ffd43b", Background: "#fff"},
		config.Fonts{}, "Course", ThemeModeDark, SkinBroadcast, "")
	if got := relativeLuminance(th.BgTop); got > 0.02 {
		t.Errorf("broadcast bgTop %s has luminance %.4f — the whole look is that the backdrop does nothing", th.BgTop, got)
	}
	if th.Air <= 0 {
		t.Errorf("broadcast air = %v, want > 0 — the diagram sits in space, it does not fill the stage", th.Air)
	}
	if th.Watermark != "Course" {
		t.Errorf("watermark = %q, want the course name as the fallback", th.Watermark)
	}
	custom := deriveVideoThemeSkinned(
		config.Colors{Primary: "#2563eb", Accent: "#ffd43b", Background: "#fff"},
		config.Fonts{}, "Course", ThemeModeDark, SkinBroadcast, "<kai>")
	if custom.Watermark != "<kai>" {
		t.Errorf("watermark = %q, want the configured mark", custom.Watermark)
	}
}

func TestNormalizeSkin(t *testing.T) {
	for _, in := range []string{"", "default", "DEFAULT", "nonsense", "  "} {
		if got := normalizeSkin(in); got != SkinDefault {
			t.Errorf("normalizeSkin(%q) = %q, want %q", in, got, SkinDefault)
		}
	}
	for _, in := range []string{"broadcast", "Broadcast", "  BROADCAST "} {
		if got := normalizeSkin(in); got != SkinBroadcast {
			t.Errorf("normalizeSkin(%q) = %q, want %q", in, got, SkinBroadcast)
		}
	}
	if got := normalizeSkin("minimal"); got != SkinMinimal {
		t.Errorf("normalizeSkin(minimal) = %q, want %q", got, SkinMinimal)
	}
}
