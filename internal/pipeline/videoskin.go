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
	// SkinEditorial breaks the centre line.
	//
	// The three skins above are all axially symmetric: headline centred, subject
	// centred under it, caption centred under that. It is a good composition and
	// it is the ONLY composition — measured across the catalog, 37 of the 41
	// explicit placements are `justify="center"` and not one of the fifty scenes
	// passes an `align` at all. Two templates as different as a leaderboard and a
	// memory budget come out looking like the same slide with a different middle,
	// which is most of why a catalog of forty-four reads as one template.
	//
	// This is the asymmetric counterpart: the headline sits left against a hard
	// vertical axis with a rule marking it, set larger and tighter than any other
	// skin, and the frame is allowed to be unbalanced. Nothing else changes —
	// which is the point of doing it as a skin. Every existing template inherits
	// the new composition without being touched, because the placement lives in
	// the shared header and the stage rather than in any scene.
	SkinEditorial = "editorial"
	// SkinShowroom is the light one, and it is the only skin that overrides the
	// mode axis rather than deriving in both polarities.
	//
	// That looks like breaking the rule this file opens with, so here is why it
	// is the rule being applied rather than bent. The other three skins are
	// *treatments* — how much light is on the backdrop, how loud the type is,
	// whether there is chrome — and a treatment is orthogonal to polarity, so
	// each one owes a dark version and a paper version. This skin is not a
	// treatment. It is a specific published look: a neutral near-white ground
	// with pure-white cards sitting on it, seated by cast shadow, the way a
	// product page seats a tile. Every part of that IS the paper. A dark
	// "showroom" would be a stage with white cards on it, which is not a darker
	// version of this look, it is one of the three above.
	//
	// What the look buys, and the reason it exists at all: a card here can wear a
	// real brand mark in the brand's own colour. On the dark stage a fetched logo
	// has to be recoloured to be visible, and a recoloured logo is no longer the
	// thing the viewer recognises — which is the entire premise of the templates
	// that fetch one. The cards batch is cut in this skin because on paper the
	// mark can simply be itself.
	SkinShowroom = "showroom"
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
	case SkinEditorial:
		return SkinEditorial
	case SkinShowroom:
		return SkinShowroom
	default:
		return SkinDefault
	}
}

// SkinNames returns the selectable skins, for the CLI and the studio picker.
func SkinNames() []string {
	return []string{SkinDefault, SkinBroadcast, SkinMinimal, SkinEditorial, SkinShowroom}
}

// SkinIsLight reports whether a skin fixes its own polarity to paper regardless
// of the course's mode. Read by the places that have to describe the look before
// a theme exists — the CLI's help, the studio's picker.
func SkinIsLight(skin string) bool { return normalizeSkin(skin) == SkinShowroom }

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

// deriveElevation fills the three tokens that say how an object is seated on
// this background — and they are the tokens that stop a scene from having to ask
// which mode it is in to draw a card.
//
// The problem they solve is that "lift this off the surface" is not one effect,
// it is two opposite ones, and until now every scene picked the dark answer with
// a literal. On the near-black stage a cast shadow does almost nothing: black on
// near-black is black. What seats a card there is the RIM — the one-pixel
// highlight along its top edge that a real object catches from a light above it —
// plus the surface being a shade lighter than the ground. On paper the rim does
// nothing instead (white on white) and the cast shadow does all of it, which is
// why the same card drawn with dark-stage seating on a light skin looks pasted
// on rather than placed.
//
// So the tokens are: what colour a shadow is cast in, how strongly, and what
// colour a rim light is. A scene composes its own shadow from the first two and
// its own hairline from the third, and gets the right physics in either polarity
// without a branch.
func deriveElevation(t *SceneTheme, h float64, mode string) {
	t.Rim = "#ffffff"
	if mode == ThemeModeLight {
		// Not black. A shadow on paper takes the colour of the light it is the
		// absence of, and a neutral-black one under a warm white card is the grey
		// smudge that makes a light design look cheap. The hue is the course's,
		// deep and desaturated enough that nobody would call it coloured.
		t.Shadow = hslToHex(h, 0.30, 0.20)
		// A tenth. On paper this is a real, visible shadow — it is the only thing
		// holding the card off the page — and every value past about 0.14 stops
		// reading as depth and starts reading as a blur filter.
		t.ShadowStrength = 0.10
		return
	}
	t.Shadow = "#000000"
	// Five times the light-mode number, for a shadow that shows five times less.
	// It is not there to be seen: it darkens the ground immediately under the
	// object, and that gradient is what stops a row of five cards reading as one
	// long panel.
	t.ShadowStrength = 0.50
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
	case SkinEditorial:
		t.Skin = SkinEditorial
		applyEditorialSkin(t, h, mode)
	case SkinShowroom:
		t.Skin = SkinShowroom
		applyShowroomSkin(t, h)
	}
}

// applyShowroomSkin puts everything on paper and lets shadow do the seating.
//
// Three numbers carry this look and each one is the opposite of what the dark
// stage wants.
//
// The ground is NEUTRAL. Every other skin keeps some of the brand hue in the
// backdrop — 8% is invisible as colour and visible as warmth, and on a dark
// stage that is free. On paper it is not free: a tint at 96% lightness is a
// visible cast, and a cast under a row of cards each wearing its own brand
// colour is a fourth colour arguing with three. So saturation drops to almost
// nothing and the ground is grey-white. The hue is still in the type and in the
// shadow, where it warms the look without colouring it.
//
// The card is PURE WHITE and the ground is not. Light mode learned this the hard
// way (see deriveBaseVideoTheme): a card can only be the bright thing if there
// is somewhere brighter to go, so the page steps down rather than the card
// stepping in. Here the page goes further down than light mode's 0.955, because
// this skin's cards have no drawn edge worth speaking of.
//
// And the border is a WHISPER — right at the floor the contrast test allows.
// That is the actual difference between this look and light mode, and it is not
// a colour at all: on paper a real object is located by the shadow it casts, not
// by an outline drawn around it. Give the card a visible hairline as well and it
// reads as a UI element pasted onto a photograph. The hairline stays only
// because a shadow alone disappears on a hard cut to a bright frame.
func applyShowroomSkin(t *SceneTheme, h float64) {
	// The mode field is overwritten rather than respected. Every renderer branch
	// that asks about polarity has to get "light" here or it will draw dark-stage
	// seating on paper.
	t.Mode = ThemeModeLight

	t.BgTop = hslToHex(h, 0.05, 0.968)
	// Barely a gradient. The reference ground is one flat tone; a few points of
	// fall-off keeps the frame from looking like a screenshot of a document
	// without ever reading as a lit surface.
	t.BgBottom = hslToHex(h, 0.04, 0.951)
	t.Surface = "#ffffff"
	t.SurfaceBorder = hslToHex(h, 0.10, 0.90)
	t.Text = hslToHex(h, 0.22, 0.10)
	t.TextMuted = hslToHex(h, 0.10, 0.40)
	t.Mass = hslToHex(h, 0.16, 0.56)
	t.Ink = hslToHex(h, 0.26, 0.18)
	// Grain is banding insurance for a dark gradient and there is no dark
	// gradient. At anything visible it reads as dirt on the paper.
	t.Grain = defaultGrain / 8

	t.AccentText = readableOn(t.Accent, t.BgTop, 4.5)
	deriveSemanticAccents(t, ThemeModeLight)
	deriveElevation(t, h, ThemeModeLight)
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
	deriveElevation(t, h, mode)
}

// applyEditorialSkin keeps the brand hue visible and gives the card an edge.
//
// The composition is what makes this skin, and composition is not a colour
// token — it lives in SceneHeader and Stage. What the palette has to do is hold
// up a hard left axis, and that asks for two things the other dark skins do not:
//
// A backdrop with the brand hue actually IN it. Broadcast drops saturation to
// ~10% so that every bit of luminance belongs to the diagram; that is right when
// one small picture floats in the middle of nothing, and wrong here, where the
// frame is deliberately unbalanced and the empty right-hand side is a large area
// of pure backdrop the viewer will look at. At 10% that area reads as a black
// hole beside the type. At 22% it reads as a chosen colour, which is what makes
// the asymmetry look intended rather than broken.
//
// And a surface that steps clearly off it. A centred composition can let a card
// dissolve into the background because its position already says where it is; a
// left-aligned one cannot, because the edge IS the axis.
func applyEditorialSkin(t *SceneTheme, h float64, mode string) {
	if mode == ThemeModeLight {
		t.BgTop = hslToHex(h, 0.16, 0.965)
		t.BgBottom = hslToHex(h, 0.14, 0.935)
		t.Surface = hslToHex(h, 0.12, 0.995)
		t.SurfaceBorder = hslToHex(h, 0.20, 0.84)
		t.Text = hslToHex(h, 0.28, 0.08)
		t.TextMuted = hslToHex(h, 0.14, 0.34)
		t.Mass = hslToHex(h, 0.16, 0.58)
		t.Ink = hslToHex(h, 0.30, 0.14)
		t.Grain = defaultGrain / 4
	} else {
		t.BgTop = hslToHex(h, 0.22, 0.085)
		t.BgBottom = hslToHex(h, 0.26, 0.055)
		t.Surface = hslToHex(h, 0.20, 0.145)
		t.SurfaceBorder = hslToHex(h, 0.24, 0.30)
		t.Text = hslToHex(h, 0.08, 0.97)
		t.TextMuted = hslToHex(h, 0.12, 0.66)
		t.Mass = hslToHex(h, 0.14, 0.84)
		t.Ink = hslToHex(h, 0.34, 0.07)
		t.Grain = defaultGrain / 2
	}
	t.AccentText = readableOn(t.Accent, t.BgTop, 4.5)
	deriveSemanticAccents(t, mode)
	deriveElevation(t, h, mode)
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
	deriveElevation(t, h, mode)
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
	case SkinEditorial:
		// Zero, and a negative value was TRIED and reverted. Scaling the block up
		// to fill the height is the obvious answer to a top-anchored composition
		// leaving the bottom empty, and it does not work: air is applied as a
		// transform about the centre (see Stage.tsx), so scaling past 1.0 grows
		// the content straight through the frame margins — at 1.1x the headline
		// clipped off the top edge and the picture ran to x=20 against a SAFE_X
		// of 110. Stage exists to make that impossible; a skin must not be the
		// thing that breaks it.
		//
		// The empty lower band on a short scene is real and is NOT a skin
		// problem. It is per-scene layout — around thirty scenes capture STAGE_H
		// at module scope — and fixing it properly means giving those scenes a
		// runtime height, which is its own piece of work.
		return 0
	case SkinShowroom:
		// A little, and 0.07 was tried first and was wrong. Air here is not the
		// same trade as on the broadcast stage: there, insetting a small diagram
		// buys composition, because the surrounding near-black is doing nothing
		// either way. On paper the surrounding area is a bright sheet, and pulling
		// a row of cards back from it does not make them look chosen, it makes
		// them look small on a large empty page — the objects are meant to be the
		// frame's subject at close to full size, the way a product page fills the
		// viewport rather than floating in it.
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
