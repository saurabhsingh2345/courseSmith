// Deterministic font loading for renders. Each bundled family is loaded via
// @remotion/google-fonts (which blocks the render until the font is ready, so
// there is never a system-font first frame). The theme names a family; if it
// isn't bundled here we fall back to the defaults rather than to a system
// font, keeping output consistent across machines.

import {loadFont as loadSpaceGrotesk} from '@remotion/google-fonts/SpaceGrotesk';
import {loadFont as loadInter} from '@remotion/google-fonts/Inter';
import {loadFont as loadJetBrainsMono} from '@remotion/google-fonts/JetBrainsMono';
import {loadFont as loadSora} from '@remotion/google-fonts/Sora';
import {loadFont as loadIBMPlexSans} from '@remotion/google-fonts/IBMPlexSans';

const spaceGrotesk = loadSpaceGrotesk('normal', {
  weights: ['400', '500', '600', '700'],
  subsets: ['latin'],
});
const inter = loadInter('normal', {
  weights: ['400', '500', '600', '700', '800'],
  subsets: ['latin'],
});
const jetBrainsMono = loadJetBrainsMono('normal', {
  weights: ['400', '500', '700'],
  subsets: ['latin'],
});
const sora = loadSora('normal', {weights: ['400', '600', '700'], subsets: ['latin']});
const ibmPlexSans = loadIBMPlexSans('normal', {
  weights: ['400', '500', '600', '700'],
  subsets: ['latin'],
});

/** Bundled family name → the fontFamily string the loader registered. */
const BUNDLED: Record<string, string> = {
  'Space Grotesk': spaceGrotesk.fontFamily,
  Inter: inter.fontFamily,
  'JetBrains Mono': jetBrainsMono.fontFamily,
  Sora: sora.fontFamily,
  'IBM Plex Sans': ibmPlexSans.fontFamily,
};

const FALLBACK_SANS = 'Helvetica, Arial, sans-serif';
const FALLBACK_MONO = 'Menlo, Consolas, monospace';

export const displayFamily = (name?: string): string =>
  `"${BUNDLED[name ?? ''] ?? BUNDLED['Space Grotesk']}", ${FALLBACK_SANS}`;

export const bodyFamily = (name?: string): string =>
  `"${BUNDLED[name ?? ''] ?? BUNDLED.Inter}", ${FALLBACK_SANS}`;

export const monoFamily = (name?: string): string =>
  `"${BUNDLED[name ?? ''] ?? BUNDLED['JetBrains Mono']}", ${FALLBACK_MONO}`;
