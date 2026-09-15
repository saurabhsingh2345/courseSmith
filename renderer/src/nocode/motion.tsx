// Shared tokens and easing for every no-code template.
//
// Pulled out of kit.tsx so the template library can grow without that file
// becoming unreadable, and so chrome.tsx / anim.tsx can share the same curves.
// Design tokens are the client's No-Code deck, unchanged: near-black ground,
// Archivo Black display, IBM Plex Mono, and red / yellow / blue used for
// nothing but emphasis.

import React from 'react';
import {useCurrentFrame} from 'remotion';
import {loadFont as loadArchivo} from '@remotion/google-fonts/ArchivoBlack';
import {loadFont as loadPlex} from '@remotion/google-fonts/IBMPlexMono';

const {fontFamily: ARCHIVO} = loadArchivo();
const {fontFamily: PLEX} = loadPlex();

export const FPS = 30;
export const C = {
  bg: '#070707', ink: '#f4f1ea', muted: '#8a867c',
  yellow: '#f5c518', red: '#e53935', blue: '#3b82f6',
  line: '#343434', panel: '#141414', dim: '#b2aea6',
};
export const MONO = `${PLEX}, ui-monospace, monospace`;
export const DISPLAY = `${ARCHIVO}, sans-serif`;

export const bez = (ax: number, ay: number, bx: number, by: number) => (t: number) => {
  if (t <= 0) return 0;
  if (t >= 1) return 1;
  let lo = 0, hi = 1, u = t;
  const fx = (v: number) => 3 * (1 - v) ** 2 * v * ax + 3 * (1 - v) * v * v * bx + v ** 3;
  const fy = (v: number) => 3 * (1 - v) ** 2 * v * ay + 3 * (1 - v) * v * v * by + v ** 3;
  for (let i = 0; i < 22; i++) {
    const x = fx(u);
    if (Math.abs(x - t) < 1e-4) break;
    if (x < t) lo = u; else hi = u;
    u = (lo + hi) / 2;
  }
  return fy(u);
};
export const easeIn = bez(0.16, 0.84, 0.32, 1);
export const easeSlam = bez(0.15, 1.4, 0.3, 1);
export const easeOut = bez(0.4, 0, 0.2, 1);

/** progress 0..1 of an animation starting at `d` seconds running `len` seconds */
export const pr = (f: number, d: number, len: number) =>
  Math.max(0, Math.min(1, (f / FPS - d) / len));

/** fade+rise in, and (optionally) back out before the slide ends */
export const Rise: React.FC<{
  d?: number; out?: number; y?: number;
  children: React.ReactNode; style?: React.CSSProperties;
}> = ({d = 0, out, y = 22, children, style}) => {
  const f = useCurrentFrame();
  const p = easeIn(pr(f, d, 0.6));
  const o = out === undefined ? 1 : 1 - easeIn(pr(f, out, 0.5));
  return (
    <div style={{opacity: p * o, transform: `translateY(${y * (1 - p)}px)`, ...style}}>
      {children}
    </div>
  );
};

/* -------------------------------------------------------------------- cues */
//
// A slide used to reveal everything on a fixed cascade — element i at
// 0.12 + i*0.26 — so a three-row slide finished moving 0.9s in and then held
// the same pixels for the remaining sixteen seconds. That is what read as
// "the audio is talking about something the picture already did".
//
// `adam_deck.py` speaks each slide sentence by sentence, so it knows when each
// sentence starts. It writes those into cues.json and the deck hands them down
// here. Element i now lands on sentence i, which is the sentence that
// describes it.
//
// The fallback is the OLD cascade on purpose: every deck that does not supply
// cues renders exactly as it shipped.

export type SlideTiming = {
  cues?: number[]; secs?: number;
  /** true when the NEXT slide is a different kind (card <-> footage), so this
   *  one should dip out rather than hard-cut. Consecutive shots from one take
   *  are continuous picture and must NOT dip - dipping there would break the
   *  continuity it is meant to create. */
  joinOut?: boolean;
  /** true when the PREVIOUS slide was a different kind */
  joinIn?: boolean;
};
export const CueCtx = React.createContext<SlideTiming>({});

/** Called once per component; returns `cue(i)` = when element i should arrive.
 *  A hook per list item would break the rules of hooks, so the context is read
 *  once and the resolver closes over it. */
export const useCues = (d0 = 0.12, step = 0.26) => {
  const {cues, secs} = React.useContext(CueCtx);
  return (i: number, n = 0) => {
    if (!cues || !cues.length) return d0 + i * step;
    // Fewer things on the slide than phrases in the narration: spread them
    // across the phrases rather than using the first n and going still for the
    // rest of the words.
    if (n > 1 && n < cues.length) {
      return cues[Math.min(cues.length - 1, Math.round((i * (cues.length - 1)) / (n - 1)))];
    }
    if (i < cues.length) return cues[i];
    // More elements than sentences is normal, and the worst case is a slide
    // read as ONE long sentence: cascading the leftovers at `step` would put
    // them all on screen in half a second and leave ten seconds of stillness,
    // which is the defect this whole pass exists to remove. Spread them across
    // the rest of the slide instead, so the last one lands near the end of the
    // words rather than at the start of them.
    const last = cues[cues.length - 1];
    const extra = Math.max(1, n - cues.length + 1);
    const room = Math.max(0, (secs ?? last + extra * step * 2.2) - last - 1.1);
    const gap = room > 0 ? room / extra : step * 2.2;
    return last + (i - cues.length + 1) * gap;
  };
};

/** the slide's own length in seconds, when the deck knows it */
export const useSlideSecs = (fallback = 8) =>
  React.useContext(CueCtx).secs ?? fallback;

/** Opacity for a slide that is joining to a different KIND of slide either
 *  side of it: a short dip through the near-black ground, which reads as a
 *  dissolve and stops a card slamming straight into footage. */
export const useJoin = (fadeIn = 0.26, fadeOut = 0.3) => {
  const f = useCurrentFrame();
  const {secs, joinIn, joinOut} = React.useContext(CueCtx);
  const a = joinIn ? Math.min(1, f / FPS / fadeIn) : 1;
  const b = joinOut && secs ? 1 - Math.max(0, (f / FPS - (secs - fadeOut)) / fadeOut) : 1;
  return Math.max(0, Math.min(1, a) * Math.min(1, b));
};

/** wrap a word in *stars* to accent it */
export const accent = (s: string, color = C.yellow) =>
  s.split(/(\*[^*]+\*)/).map((bit, j) =>
    bit.startsWith('*') && bit.endsWith('*')
      ? <span key={j} style={{color}}>{bit.slice(1, -1)}</span>
      : <span key={j}>{bit}</span>);
