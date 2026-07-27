package pipeline

import (
	"math"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

func TestHexHSLRoundTrip(t *testing.T) {
	for _, hex := range []string{"#306998", "#ffd43b", "#2563eb", "#f59e0b", "#ffffff", "#000000", "#abc"} {
		h, s, l := hexToHSL(hex)
		if h < 0 || h >= 360 || s < 0 || s > 1 || l < 0 || l > 1 {
			t.Errorf("hexToHSL(%q) out of range: h=%v s=%v l=%v", hex, h, s, l)
		}
		out := hslToHex(h, s, l)
		or, og, ob, err1 := parseHex(hex)
		nr, ng, nb, err2 := parseHex(out)
		if err1 != nil || err2 != nil {
			t.Fatalf("parseHex failed for %q/%q", hex, out)
		}
		// Round-trip within 1/255 per channel.
		for _, d := range []float64{or - nr, og - ng, ob - nb} {
			if math.Abs(d) > 1.5/255 {
				t.Errorf("round trip %q -> %q drifted", hex, out)
			}
		}
	}
}

func TestHexToHSLBadInputDegrades(t *testing.T) {
	h, s, l := hexToHSL("not-a-color")
	if h != 220 || s != 0.3 || l != 0.4 {
		t.Errorf("bad input should degrade to neutral slate, got h=%v s=%v l=%v", h, s, l)
	}
}

func TestDeriveVideoTheme(t *testing.T) {
	colors := config.Colors{Primary: "#306998", Accent: "#ffd43b", Background: "#ffffff"}
	th := deriveVideoTheme(colors, config.Fonts{}, "Python Basics", "")

	if th.Primary != "#306998" || th.Accent != "#ffd43b" || th.Background != "#ffffff" {
		t.Errorf("legacy fields must pass through unchanged: %+v", th)
	}
	if th.CourseName != "Python Basics" {
		t.Errorf("CourseName = %q", th.CourseName)
	}
	if th.Mode != "dark" {
		t.Errorf("Mode = %q, want dark", th.Mode)
	}
	for name, v := range map[string]string{
		"BgTop": th.BgTop, "BgBottom": th.BgBottom, "Surface": th.Surface,
		"SurfaceBorder": th.SurfaceBorder, "Text": th.Text, "TextMuted": th.TextMuted,
	} {
		if !strings.HasPrefix(v, "#") || len(v) != 7 {
			t.Errorf("%s = %q, want #rrggbb", name, v)
		}
	}
	// The background must be dark and the text light.
	if _, _, l := hexToHSL(th.BgTop); l > 0.2 {
		t.Errorf("BgTop lightness %v, want dark (<0.2)", func() float64 { _, _, l := hexToHSL(th.BgTop); return l }())
	}
	if _, _, l := hexToHSL(th.Text); l < 0.85 {
		t.Errorf("Text lightness too dark for a dark theme")
	}
	// Background hue should be tinted toward the primary hue.
	ph, _, _ := hexToHSL(colors.Primary)
	bh, _, _ := hexToHSL(th.BgTop)
	if diff := math.Abs(ph - bh); diff > 20 && diff < 340 {
		t.Errorf("BgTop hue %v not tinted toward primary hue %v", bh, ph)
	}
	if th.FontDisplay != DefaultFontDisplay || th.FontBody != DefaultFontBody || th.FontMono != DefaultFontMono {
		t.Errorf("default fonts wrong: %q/%q/%q", th.FontDisplay, th.FontBody, th.FontMono)
	}
	if th.Grain <= 0 || th.Grain > 0.2 {
		t.Errorf("Grain = %v, want subtle (0, 0.2]", th.Grain)
	}
}

func TestDeriveVideoThemeFontOverrides(t *testing.T) {
	th := deriveVideoTheme(config.Colors{Primary: "#2563eb"}, config.Fonts{Display: "Sora", Mono: "Fira Code"}, "C", "")
	if th.FontDisplay != "Sora" {
		t.Errorf("FontDisplay = %q, want Sora", th.FontDisplay)
	}
	if th.FontBody != DefaultFontBody {
		t.Errorf("FontBody = %q, want default when not overridden", th.FontBody)
	}
	if th.FontMono != "Fira Code" {
		t.Errorf("FontMono = %q, want Fira Code", th.FontMono)
	}
}
