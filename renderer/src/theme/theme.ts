// Theme resolution: the Go pipeline derives rich design tokens
// (internal/pipeline/videotheme.go) and embeds them in lesson-video.json.
// Scene graphs generated before the design system carry only the three course
// colours — resolveTheme() fills the same dark-editorial defaults so old
// graphs render in the new look too.

import {Theme} from '../types';
import {bodyFamily, displayFamily, monoFamily} from './fonts';

export type ResolvedTheme = {
  primary: string;
  accent: string;
  background: string;
  courseName: string;
  mode: 'dark' | 'light';
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

export const resolveTheme = (t: Theme): ResolvedTheme => ({
  primary: t.primary,
  accent: t.accent,
  background: t.background || '#ffffff',
  courseName: t.courseName,
  mode: t.mode ?? 'dark',
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
