// Theme resolution: the Go pipeline derives rich design tokens
// (internal/pipeline/videotheme.go) and embeds them in lesson-video.json.
// Scene graphs generated before the design system carry only the three course
// colours — resolveTheme() fills the same dark-editorial defaults so old
// graphs render in the new look too.

import {Theme} from '../types';
import {bodyFamily, displayFamily, monoFamily} from './fonts';

/** The house styles a video can be cut in. Mirrors videoskin.go. */
export type Skin = 'default' | 'broadcast' | 'minimal';

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
