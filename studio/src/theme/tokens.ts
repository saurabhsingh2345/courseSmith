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
    // The `ink` ramp — the studio's working scale, used directly as
    // `bg-ink-900` / `text-ink-100` across every page.
    //
    // It is ordered by DISTANCE FROM THE READER, not by lightness: 950 is the
    // page behind everything and 100 is the brightest thing written on it. That
    // is why light mode is not a different palette but the same ramp inverted —
    // 950 becomes the white page and 100 becomes near-black type. Encoding the
    // ramp's meaning as depth rather than as luminance is what lets one
    // `text-ink-100` be correct in both modes, instead of every page needing a
    // `dark:` variant on 640 class names.
    ink: {
      950: {light: '#ffffff', dark: '#09090b'} as ColorScale,
      900: {light: '#fafafa', dark: '#0e0e11'} as ColorScale,
      850: {light: '#f4f4f6', dark: '#131316'} as ColorScale,
      800: {light: '#ececef', dark: '#18181c'} as ColorScale,
      750: {light: '#e3e3e7', dark: '#1e1e23'} as ColorScale,
      700: {light: '#d4d4da', dark: '#26262c'} as ColorScale,
      600: {light: '#b0b0b8', dark: '#33333b'} as ColorScale,
      // 500 and 400 are the muted/secondary *text* steps — `text-ink-500` alone
      // is the most-used colour in the studio. Their dark values were #4c4c56
      // and #71717c, which measure 2.3:1 and 4.1:1 on the page and so failed AA
      // as body text. Raised to pass; the ink.test.ts assertions are what keep
      // them passing, in both modes.
      500: {light: '#6b6b76', dark: '#7a7a86'} as ColorScale,
      400: {light: '#55555f', dark: '#8e8e9a'} as ColorScale,
      300: {light: '#3f3f48', dark: '#9d9da8'} as ColorScale,
      200: {light: '#2a2a31', dark: '#c6c6cf'} as ColorScale,
      100: {light: '#18181c', dark: '#e6e6ec'} as ColorScale,
    } as Record<string, ColorScale>,
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
