// The studio design system — the single source of truth for colour,
// typography, and spacing. Everything is emitted as CSS custom properties by
// applyTheme(), so a component references var(--color-brand) and a token change
// here propagates everywhere. Motion tokens come from motionTokens.ts (owned by
// Go) so the UI and the video share one animation language.

import {motion, type MotionTokens} from './motionTokens';

export type ColorScale = {light: string; dark: string};

export const tokens = {
  colors: {
    // Brand accent adapts per theme; `saturated` is the pressed/active shade.
    brand: {light: '#0066ff', dark: '#00d9ff', saturated: '#0052cc'},
    semantic: {
      success: '#22c55e',
      error: '#ef4444',
      warning: '#f59e0b',
      info: '#3b82f6',
    },
    // Neutral surfaces resolve differently in light vs dark (see applyTheme).
    neutral: {
      bg: {light: '#fafafa', dark: '#09090b'} as ColorScale,
      surface: {light: '#ffffff', dark: '#131316'} as ColorScale,
      border: {light: '#e5e7eb', dark: '#26262c'} as ColorScale,
      text: {light: '#1f2937', dark: '#c6c6cf'} as ColorScale,
      muted: {light: '#6b7280', dark: '#71717c'} as ColorScale,
    },
  },
  typography: {
    heading1: {size: '2rem', lineHeight: 1.2, weight: 700},
    heading2: {size: '1.5rem', lineHeight: 1.3, weight: 600},
    body: {size: '1rem', lineHeight: 1.6, weight: 400},
    mono: {family: '"JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, monospace', size: '0.9rem'},
  },
  spacing: {xs: '0.25rem', sm: '0.5rem', md: '1rem', lg: '2rem', xl: '4rem'} as Record<string, string>,
  radius: {sm: '4px', md: '8px', lg: '14px', full: '9999px'},
  motion,
} as const;

export type Tokens = typeof tokens;
export type {MotionTokens};
