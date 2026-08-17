// Theme resolution: the Go pipeline derives rich design tokens
// (internal/pipeline/videotheme.go) and embeds them in lesson-video.json.
// Scene graphs generated before the design system carry only the three course
// colours — resolveTheme() fills the same dark-editorial defaults so old
// graphs render in the new look too.

import {Theme} from '../types';
import {bodyFamily, displayFamily, monoFamily} from './fonts';

/** The house styles a video can be cut in. Mirrors videoskin.go. */
export type Skin = 'default' | 'broadcast' | 'minimal' | 'editorial' | 'showroom';

export type ResolvedTheme = {
  primary: string;
  accent: string;
  background: string;
  courseName: string;
  mode: 'dark' | 'light';
  /** House style. 'default' is the look every pre-skin scene graph was
   *  recorded against, and what an absent field resolves to. */
  skin: Skin;
  /** Fraction of the drawing box a skin leaves empty on every side. 0 fills
   *  the stage, which is the default. */
  air: number;
  /** Standing corner mark. Empty leaves the corner clean. */
  watermark: string;
  /** The measured number. */
  accentQuantity: string;
  /** The ceiling it hits, the thing that does not fit. */
  accentLimit: string;
  /** The alternative being weighed against the subject. */
  accentRival: string;
  bgTop: string;
  bgBottom: string;
  surface: string;
  surfaceBorder: string;
  text: string;
  textMuted: string;
  /** Body fill for drawn artwork; flips with the mode so a mass is always a
   *  shape against the stage. */
  mass: string;
  /** Shading laid over a mass. Always darker than `mass`, in both modes. */
  ink: string;
  /** The colour a cast shadow is drawn in on this background. */
  shadow: string;
  /** How opaque that shadow is. High on paper, where a shadow is the only thing
   *  holding a card off the page; low on the stage, where black on near-black is
   *  black and the rim does the seating instead. */
  shadowStrength: number;
  /** The hairline highlight along a lit object's top edge. */
  rim: string;
  /** A muted mark on the open stage: rules, connectors, axes. */
  line: string;
  /** The accent set as *type*. Use this wherever the accent is a word; use
   *  `accent` for fills and strokes, which sit on their own shape. */
  accentText: string;
  /** Ready-to-use CSS font stacks (bundled family + fallback). */
  fontDisplay: string;
  fontBody: string;
  fontMono: string;
  grain: number;
};

// Neutral dark defaults for pre-design-system scene graphs (slate-tinted).
const DEFAULTS = {
  bgTop: '#101827',
  bgBottom: '#080d18',
  surface: '#182236',
  surfaceBorder: '#2b3a55',
  text: '#f2f5fa',
  textMuted: '#a2aec4',
  mass: '#dbe4f2',
  ink: '#0a1220',
  shadow: '#000000',
  shadowStrength: 0.5,
  rim: '#ffffff',
  grain: 0.04,
};

/**
 * Overlay a hex colour at a given alpha.
 *
 * Design tokens are opaque hex by construction — Go derives them and proves
 * their contrast against the background — so any scene that wants a tint, a
 * glow or a hairline of an existing token has to build it here rather than
 * inventing a literal, which would not flip with the mode.
 */
export const withAlpha = (hex: string, alpha: number): string => {
  const h = hex.replace('#', '');
  if (h.length !== 6) return hex;
  const r = parseInt(h.slice(0, 2), 16);
  const g = parseInt(h.slice(2, 4), 16);
  const b = parseInt(h.slice(4, 6), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
};

/** How far off its background an object is sitting. */
export type Elevation = 'resting' | 'lifted';

/**
 * The box-shadow that seats an object on this background.
 *
 * The interesting property is that it does NOT ask which polarity it is in, and
 * it does not need to. All three layers are emitted every time and the tokens
 * turn off the ones that do not apply: on paper the rim is white over a white
 * card and is simply not there, while the two cast layers do all the work; on
 * the near-black stage the rim is the one-pixel top-edge highlight that actually
 * seats the card, and the cast layers become the soft darkening of the ground
 * underneath that stops a row of five reading as one long panel.
 *
 * That is the whole reason elevation is three tokens rather than a mode check in
 * ninety scenes. A scene says how far off the surface a thing is; the design
 * system decides what that means where the thing is standing.
 */
export const seat = (t: ResolvedTheme, level: Elevation = 'resting'): string => {
  const s = t.shadowStrength;
  const [contactY, contactBlur, ambientY, ambientBlur, k] =
    level === 'lifted' ? [3, 6, 24, 56, 1.35] : [1, 2, 10, 28, 1];
  return [
    `inset 0 1px 0 ${withAlpha(t.rim, 0.07)}`,
    // The contact shadow: the hard, close darkening right under an object, which
    // is what says it is touching the surface rather than floating above it.
    // Kept at a fraction of the ambient strength — at parity it reads as a
    // border, which is the one thing this look does not want.
    `0 ${contactY}px ${contactBlur}px ${withAlpha(t.shadow, s * 0.5 * k)}`,
    `0 ${ambientY}px ${ambientBlur}px ${withAlpha(t.shadow, s * k)}`,
  ].join(', ');
};

/**
 * Relative luminance of a hex colour, 0..1 — the WCAG definition, matching
 * relLuminance in videotheme.go.
 *
 * The renderer needs its own copy for one reason: fetched brand colours. Every
 * other colour on the frame was derived in Go and proven legible there, but a
 * logo's own hex arrives from the open web at render time and nothing has
 * checked it against the card it is about to sit on. See CardsScene.
 */
export const luminance = (hex: string): number => {
  const h = hex.replace('#', '');
  if (h.length !== 6) return 0;
  const lin = (c: number) => (c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4));
  const r = lin(parseInt(h.slice(0, 2), 16) / 255);
  const g = lin(parseInt(h.slice(2, 4), 16) / 255);
  const b = lin(parseInt(h.slice(4, 6), 16) / 255);
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
};

/** The WCAG contrast ratio between two hex colours, 1..21. */
export const contrastRatio = (a: string, b: string): number => {
  const la = luminance(a);
  const lb = luminance(b);
  return (Math.max(la, lb) + 0.05) / (Math.min(la, lb) + 0.05);
};

export const resolveTheme = (t: Theme): ResolvedTheme => ({
  primary: t.primary,
  accent: t.accent,
  background: t.background || '#ffffff',
  courseName: t.courseName,
  mode: t.mode ?? 'dark',
  skin: t.skin ?? 'default',
  air: t.air ?? 0,
  watermark: t.watermark ?? '',
  // Semantic accents predate no scene graph that used them, so a missing one
  // falling back to the brand accent is safe: the frame loses the *distinction*
  // between the three roles rather than losing colour altogether, which is the
  // right way for an old graph to degrade.
  accentQuantity: t.accentQuantity ?? t.accent,
  accentLimit: t.accentLimit ?? t.accent,
  accentRival: t.accentRival ?? t.primary,
  bgTop: t.bgTop ?? DEFAULTS.bgTop,
  bgBottom: t.bgBottom ?? DEFAULTS.bgBottom,
  surface: t.surface ?? DEFAULTS.surface,
  surfaceBorder: t.surfaceBorder ?? DEFAULTS.surfaceBorder,
  text: t.text ?? DEFAULTS.text,
  textMuted: t.textMuted ?? DEFAULTS.textMuted,
  mass: t.mass ?? DEFAULTS.mass,
  ink: t.ink ?? DEFAULTS.ink,
  // Every scene graph written before these existed is dark, where the defaults
  // are exactly what those scenes had hardcoded — so an old graph keeps the
  // seating it already drew rather than degrading to none.
  shadow: t.shadow ?? DEFAULTS.shadow,
  shadowStrength: t.shadowStrength ?? DEFAULTS.shadowStrength,
  rim: t.rim ?? DEFAULTS.rim,
  // `line` is not derived separately in Go: a muted mark on the stage and
  // secondary text want the same value, and both modes already prove that
  // value readable against the background (videotheme_contrast_test.go).
  line: t.textMuted ?? DEFAULTS.textMuted,
  // Scene graphs written before this token existed are all dark mode, where
  // the accent already passes as type — so falling back to it is correct
  // rather than merely safe.
  accentText: t.accentText ?? t.accent,
  fontDisplay: displayFamily(t.fontDisplay),
  fontBody: bodyFamily(t.fontBody),
  fontMono: monoFamily(t.fontMono),
  grain: t.grain ?? DEFAULTS.grain,
});
