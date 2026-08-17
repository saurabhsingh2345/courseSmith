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
// The one serif in the stack, and the reason it is here: a display serif is a
// different VOICE, not a different taste. Every other family bundled above is a
// sans, so a title set in any of them reads as the same title at a different
// size — which is why the catalog's intro cards have always looked like its
// diagram labels. Instrument Serif is a high-contrast transitional face, and it
// carries an intro the way a book's title page does.
import {loadFont as loadInstrumentSerif} from '@remotion/google-fonts/InstrumentSerif';

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
// One weight, because the face ships one. A display serif at 400 set very large
// is already heavier on the page than a sans at 700 set small — the contrast in
// the strokes does the work that weight does elsewhere.
const instrumentSerif = loadInstrumentSerif('normal', {weights: ['400'], subsets: ['latin']});

/** Bundled family name → the fontFamily string the loader registered. */
const BUNDLED: Record<string, string> = {
  'Space Grotesk': spaceGrotesk.fontFamily,
  Inter: inter.fontFamily,
  'JetBrains Mono': jetBrainsMono.fontFamily,
  Sora: sora.fontFamily,
  'IBM Plex Sans': ibmPlexSans.fontFamily,
  'Instrument Serif': instrumentSerif.fontFamily,
};

const FALLBACK_SANS = 'Helvetica, Arial, sans-serif';
const FALLBACK_MONO = 'Menlo, Consolas, monospace';
const FALLBACK_SERIF = 'Georgia, "Times New Roman", serif';

export const displayFamily = (name?: string): string =>
  `"${BUNDLED[name ?? ''] ?? BUNDLED['Space Grotesk']}", ${FALLBACK_SANS}`;

export const bodyFamily = (name?: string): string =>
  `"${BUNDLED[name ?? ''] ?? BUNDLED.Inter}", ${FALLBACK_SANS}`;

export const monoFamily = (name?: string): string =>
  `"${BUNDLED[name ?? ''] ?? BUNDLED['JetBrains Mono']}", ${FALLBACK_MONO}`;

export const serifFamily = (name?: string): string =>
  `"${BUNDLED[name ?? ''] ?? BUNDLED['Instrument Serif']}", ${FALLBACK_SERIF}`;
