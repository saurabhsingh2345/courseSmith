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
  fontDisplay: displayFamily(t.fontDisplay),
  fontBody: bodyFamily(t.fontBody),
  fontMono: monoFamily(t.fontMono),
  grain: t.grain ?? DEFAULTS.grain,
});
