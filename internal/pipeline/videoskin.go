package pipeline

// Skins: the house style a video is cut in.
//
// A skin is a third, independent axis alongside the two the theme already had.
// Branding decides the *hue* everything is tinted toward; mode decides the
// *polarity*, paper or stage; a skin decides the *house style* — how much light
// is on the backdrop, how loud the type is, and whether the frame carries
// standing chrome. They compose: every skin derives in both modes and around
// the whole hue circle, because a skin that only worked on dark slate would be
// a hardcoded look wearing a token's name.
//
// The catalog's own look is `default`, and it is untouched by design. Skins are
// strictly additive: deriveVideoTheme runs exactly as it always did and a skin
// then overrides the handful of tokens it disagrees with. A course that never
// mentions `skin` gets a scene graph byte-identical to the one it got before
// this file existed (TestDefaultSkinIsUnchanged), so nothing that already
// renders can regress.
//
// What a skin does *not* get to do is invent colour. Semantic accents
// (below) are derived here for every skin, including `default`, because
// "the measured quantity is gold and the ceiling it hits is red" is a claim
// about meaning rather than about branding — a template that draws a bar
// crossing a limit needs the same two colours whatever the course is branded
// with, and needs them to still be legible on paper.

import (
	"math"
	"strings"
)

// The house styles a video can be cut in.
const (
	// SkinDefault is the look every existing course was recorded against: a
	// brand-tinted gradient, drifting accent glows, a display serif-free
	// headline over a gradient rule. Unchanged, and the default.
	SkinDefault = "default"
	// SkinBroadcast is the near-black explainer look: the backdrop drops to
	// almost nothing, the headline goes large and uppercase, small type goes
	// mono, and every frame carries standing chrome — an eyebrow naming where
	// you are, a takeaway line, and a watermark. It is built for clips whose
	// subject is one precise diagram with a lot of air around it.
	SkinBroadcast = "broadcast"
	// SkinMinimal is the quiet counterpart: a flat charcoal stage, one accent,
	// sentence case, no headline furniture and no chrome at all. The diagram is
	// the whole frame and nothing else competes with it.
	SkinMinimal = "minimal"
)

// normalizeSkin maps config input onto a known skin. Anything unrecognised —
// including the empty string — is the default, which is the look every existing
// scene graph was recorded against.
func normalizeSkin(skin string) string {
	switch strings.ToLower(strings.TrimSpace(skin)) {
	case SkinBroadcast:
		return SkinBroadcast
	case SkinMinimal:
		return SkinMinimal
	default:
		return SkinDefault
	}
}

// SkinNames returns the selectable skins, for the CLI and the studio picker.
func SkinNames() []string {
	return []string{SkinDefault, SkinBroadcast, SkinMinimal}
}

// ResolvedSkin is the house style this theme is in. The field is empty on an
// unskinned theme (see applySkin), so read it through here rather than
// directly.
func (t SceneTheme) ResolvedSkin() string { return normalizeSkin(t.Skin) }

// Semantic accent hue anchors.
//
// These are fixed rather than derived from branding, and that is the point. A
// bar that overruns its ceiling is red because red means "this does not fit",
// not because the course happens to be branded red; run the anchor through the
// brand hue and a course branded green would draw its failure state in green,
// which is not a restyling, it is a lie about what the picture says.
//
// So the hue is a constant and only lightness and saturation move — enough to
// stay legible against a near-black stage and against paper, proven by
// TestSemanticAccentContrast rather than eyeballed.
const (
	// hueLimit is the ceiling, the overrun, the thing that does not fit.
	hueLimit = 4
	// hueQuantity is the measured number — the value being counted up to.
	hueQuantity = 45
	// hueRival is the alternative being weighed against the subject.
	hueRival = 217
	// hueQuantityLight is the quantity role on paper. See deriveSemanticAccents:
	// gold at readable contrast is khaki, amber at the same contrast is still
	// amber.
	hueQuantityLight = 32
)

// deriveSemanticAccents fills the three role colours for a mode. They are
// derived for every skin, not only the ones that lean on them, so a template
// can reach for `accentLimit` without first asking which skin it is in.
func deriveSemanticAccents(t *SceneTheme, mode string) {
	// On the stage a role colour is a light mark on darkness; on paper it has
	// to be dark enough to read as ink. Saturation stays high in both — these
	// are signal colours and a desaturated signal is a decoration.
	l := 0.62
	quantityHue := float64(hueQuantity)
	if mode == ThemeModeLight {
		l = 0.42
		// Gold does not survive being darkened. Walked down to AA on paper it
		// lands on khaki — the same colour a bar in mud is — and the role stops
		// reading as "the measured quantity" and starts reading as a mistake.
		// Rotating toward amber keeps chroma as lightness falls, so the role
		// still looks deliberate on white. The hue moves only here: on the dark
		// stage the original gold is already legible and is the better colour.
		quantityHue = hueQuantityLight
	}
	t.AccentLimit = hslToHex(hueLimit, 0.80, l)
	t.AccentQuantity = hslToHex(quantityHue, 0.90, l)
	t.AccentRival = hslToHex(hueRival, 0.80, l)

	// Gold sits at very nearly the luminance of paper, so in light mode the
	// nominal lightness is not enough on its own. Each role is walked down
	// until it clears AA against the background it will be set on, exactly as
	// the brand accent is.
	t.AccentLimit = readableOn(t.AccentLimit, t.BgTop, 4.5)
	t.AccentQuantity = readableOn(t.AccentQuantity, t.BgTop, 4.5)
	t.AccentRival = readableOn(t.AccentRival, t.BgTop, 4.5)
}

// applySkin overrides the tokens a skin disagrees with, on top of an
// already-derived theme. h is the branding hue the base theme was derived from.
//
// The default skin returns without touching anything — that is what makes this
// file additive rather than a rewrite of the design system.
func applySkin(t *SceneTheme, h float64, skin, mode string) {
	switch normalizeSkin(skin) {
	case SkinBroadcast:
		// Only a non-default skin writes the field. The default has to leave it
		// empty so `omitempty` drops it: every config fingerprint and golden
		// scene graph recorded before skins existed is taken over this JSON, and
		// a new key on an unskinned course would move all of them.
		t.Skin = SkinBroadcast
		applyBroadcastSkin(t, h, mode)
	case SkinMinimal:
		t.Skin = SkinMinimal
		applyMinimalSkin(t, h, mode)
	}
}

// applyBroadcastSkin drops the stage to near-black and takes the light off it.
//
// The reference look reads as expensive because the backdrop does nothing at
// all: no gradient to speak of, no glow, no field. Every bit of luminance in
// the frame belongs to the one diagram in the middle. That is the opposite of
// the default stage, which is doing several things at once and doing them well
// when a scene fills the frame — and fighting the picture when it does not.
func applyBroadcastSkin(t *SceneTheme, h float64, mode string) {
	if mode == ThemeModeLight {
		// Paper, flattened. The gradient nearly vanishes so the page reads as
		// one sheet rather than as a lit surface.
		t.BgTop = hslToHex(h, 0.10, 0.975)
		t.BgBottom = hslToHex(h, 0.08, 0.955)
		t.Surface = hslToHex(h, 0.08, 0.995)
		t.SurfaceBorder = hslToHex(h, 0.10, 0.88)
		t.Text = hslToHex(h, 0.20, 0.10)
		t.TextMuted = hslToHex(h, 0.10, 0.38)
		t.Mass = hslToHex(h, 0.12, 0.62)
		t.Ink = hslToHex(h, 0.24, 0.18)
		t.Grain = defaultGrain / 5
	} else {
		// Near-black, with just enough of the brand hue left in it that the
		// stage is not a neutral void — 8% saturation is invisible as colour
		// and visible as warmth, which is the difference between "black" and
		// "this video has a palette".
		t.BgTop = hslToHex(h, 0.10, 0.045)
		t.BgBottom = hslToHex(h, 0.12, 0.028)
		t.Surface = hslToHex(h, 0.10, 0.095)
		t.SurfaceBorder = hslToHex(h, 0.10, 0.20)
		t.Text = hslToHex(h, 0.06, 0.98)
		t.TextMuted = hslToHex(h, 0.06, 0.62)
		t.Mass = hslToHex(h, 0.08, 0.88)
		t.Ink = hslToHex(h, 0.30, 0.05)
		// Less grain than the default: there is barely a gradient left to band,
		// and on a stage this dark the same 4% reads as sensor noise.
		t.Grain = defaultGrain / 2
	}
	// The accent has to be re-proven against the new background — a brand
	// colour that cleared AA on the default slate is being asked to sit on
	// something much darker or much lighter now.
	t.AccentText = readableOn(t.Accent, t.BgTop, 4.5)
	deriveSemanticAccents(t, mode)
}

// applyMinimalSkin flattens the stage to a single tone and keeps one accent.
//
// Where broadcast is loud — huge type, standing chrome, three signal colours —
// this is the same restraint pointed the other way: no gradient, no glow, no
// furniture, one colour used sparingly. It suits subjects whose picture is a
// graph of nodes and edges, where every extra colour is another thing to
// decode.
func applyMinimalSkin(t *SceneTheme, h float64, mode string) {
	if mode == ThemeModeLight {
		t.BgTop = hslToHex(h, 0.06, 0.985)
		t.BgBottom = hslToHex(h, 0.06, 0.975)
		t.Surface = hslToHex(h, 0.05, 0.945)
		t.SurfaceBorder = hslToHex(h, 0.08, 0.86)
		t.Text = hslToHex(h, 0.16, 0.12)
		t.TextMuted = hslToHex(h, 0.08, 0.40)
		t.Mass = hslToHex(h, 0.10, 0.64)
		t.Ink = hslToHex(h, 0.20, 0.20)
		t.Grain = defaultGrain / 6
	} else {
		// Charcoal rather than black: the cards in this look are a shade above
		// the stage rather than outlined on it, and that only works if the
		// stage itself is not already at the floor.
		t.BgTop = hslToHex(h, 0.08, 0.075)
		t.BgBottom = hslToHex(h, 0.08, 0.062)
		t.Surface = hslToHex(h, 0.10, 0.135)
		t.SurfaceBorder = hslToHex(h, 0.10, 0.24)
		t.Text = hslToHex(h, 0.05, 0.96)
		t.TextMuted = hslToHex(h, 0.05, 0.62)
		t.Mass = hslToHex(h, 0.08, 0.86)
		t.Ink = hslToHex(h, 0.28, 0.06)
		t.Grain = defaultGrain / 3
	}
	t.AccentText = readableOn(t.Accent, t.BgTop, 4.5)
	deriveSemanticAccents(t, mode)
}

// skinAir is how much a skin pulls its content in from the stage edges, as a
// fraction of the drawing box: 0 is the default (fill the stage), 0.14 leaves a
// seventh of it empty on every side.
//
// This is the single most visible difference between the reference look and the
// catalog's, and it is not a colour. The default stage is filled deliberately —
// a diagram that fills the frame is legible on a phone. The broadcast look
// instead puts one small precise picture in the middle of a lot of nothing, and
// reads as composed rather than as packed. Both are right, for different clips,
// which is exactly why it is a skin setting and not a global.
// The renderer applies it as a scale on the scene's content rather than as
// padding (see Stage.tsx for why), so the numbers are small: 0.06 sets the
// composition back to 88% and reads as breathing room, while anything near 0.15
// would shrink the headline as much as the diagram and undo the point of
// setting the type larger in the first place.
func skinAir(skin string) float64 {
	switch normalizeSkin(skin) {
	case SkinBroadcast:
		return 0.06
	case SkinMinimal:
		return 0.03
	default:
		return 0
	}
}

// roundTo is a small helper for keeping derived floats stable in JSON.
func roundTo(v float64, places int) float64 {
	p := math.Pow(10, float64(places))
	return math.Round(v*p) / p
}
