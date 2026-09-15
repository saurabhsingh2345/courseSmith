// Week 3's animation layer.
//
// Week 1's lectures are Remotion decks with the screen recordings mounted
// inside them as `shot` slides, so the picture changes every ten seconds or so.
// Week 3 was cut the other way round - one continuous 4K screen capture with
// narration laid over it and the footage ramped to fit - which is why it has no
// animation in it at all, and why 25 of its 35 lectures sit above 70% of their
// runtime on an unchanged picture.
//
// Rather than rebuild those 35 lectures, this renders short animated CUTAWAYS
// that get laid over the stretches where the delivered picture does not move.
// `tools/nocode/qa/deadzones.py` finds those stretches; the plans in
// `tools/nocode/w3/inserts/plans/` say what goes in each one.
//
// Two rules the whole approach depends on:
//
//   1. **The narration is never touched.** An insert is exactly as long as the
//      hole it fills, so the audio never moves and the runtime never changes.
//      That is why `MuteDeck` carries no `<Audio>` and adds no inter-slide gap
//      the way `Deck` does - a gap would push the last slide past the end of
//      the hole and the cut back to footage would land late.
//   2. **Only dead frames get covered.** Live footage is the lesson; a cutaway
//      over a terminal that is actually printing would be vandalism.
//
// Rendered at 1920x1080 like every other deck in this library and taken to 4K
// with Remotion's `--scale=2`, which supersamples rather than upscaling, so the
// Archivo Black headlines stay crisp against a 4K master.

import React from 'react';
import {AbsoluteFill, Sequence} from 'remotion';
import {C, FPS, Slide, SlideView} from './kit';

/** Frame count for one insert. Rounded per slide so the sum is exact. */
export const insertFrames = (secs: number[]) =>
  secs.reduce((a, s) => a + Math.round(s * FPS), 0);

/**
 * A deck with no voice track and no gap between slides.
 *
 * The last slide absorbs any rounding drift so the strip is exactly
 * `insertFrames(secs)` long - the overlay window is computed from the same
 * number, and a one-frame disagreement shows up as a flash of footage.
 */
export const MuteDeck: React.FC<{slides: Slide[]; secs: number[]}> = ({
  slides,
  secs,
}) => {
  const total = insertFrames(secs);
  let at = 0;
  const out: React.ReactNode[] = [];

  slides.forEach((s, i) => {
    const last = i === slides.length - 1;
    const len = last ? total - at : Math.round((secs[i] ?? 0) * FPS);
    if (len <= 0) return;
    out.push(
      <Sequence key={i} from={at} durationInFrames={len}>
        <SlideView s={s} />
      </Sequence>,
    );
    at += len;
  });

  return <AbsoluteFill style={{background: C.bg}}>{out}</AbsoluteFill>;
};

export type InsertProps = {slides: Slide[]; secs: number[]};

/**
 * One insert, driven entirely by `--props`, so the 35 lectures' worth of
 * cutaways live in JSON plans rather than in 35 hand-written compositions.
 */
export const W3Insert: React.FC<InsertProps> = ({slides, secs}) => (
  <MuteDeck slides={slides} secs={secs} />
);

/** A visible default so the composition is openable in the Remotion studio. */
export const INSERT_DEFAULT: InsertProps = {
  slides: [
    {
      k: 'head',
      kicker: 'w3 insert',
      lines: ['Pass *slides* and', '*secs* as props'],
      size: 88,
      trans: 'fade',
    },
  ],
  secs: [4],
};
