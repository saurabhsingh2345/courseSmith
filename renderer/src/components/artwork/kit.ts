// The shared kit every figure is built from.
//
// This exists because the vocabulary is ~100 objects and the interesting part
// of each one is its *mechanism* — the flame that flickers, the gear that
// meshes, the key that turns. Everything around that is the same in all of
// them: parts that arrive in a staggered pop, a stroke that draws itself, an
// idle motion that keeps running once the entrance is over. Written out per
// figure that boilerplate would be four fifths of the file and the mechanism
// would be buried in it.
//
// Every figure draws into the same 200x200 box and takes the same two clocks:
//
//   build  0-1, the figure assembling itself as its beat arrives. Parts are
//          staggered along it, so a figure lands as one gesture.
//   t      seconds since the scene started, for the idle motion that runs for
//          as long as the figure is on screen. The idle is the whole trick —
//          a figure that finishes assembling and then freezes reads as a
//          static image no matter how good the entrance was.
//
// Nothing here reads a clock itself and nothing is random, so frames stay
// independent and reproducible (Remotion renders them out of order).

export type FigurePalette = {
  accent: string;
  primary: string;
  /**
   * The shading colour, used to darken a face of a light mass. It is a *dark*
   * value, so it only ever works as a fill over something lighter — a stroke
   * painted with it on the stage is invisible. Strokes drawn on open
   * background want `line`.
   */
  ink: string;
  /** A light neutral for the main masses, already blended against the stage. */
  soft: string;
  /** A muted light, for rules and connectors drawn on the open stage. */
  line: string;
};

export type FigureProps = {
  build: number;
  t: number;
  palette: FigurePalette;
};

export type Figure = React.FC<FigureProps>;

/** The box every figure is drawn in. */
export const FIGURE_BOX = 200;

/** Centre of that box, which most figures are built around. */
export const C = 100;

export const clamp01 = (v: number): number => Math.max(0, Math.min(1, v));

/** Cubic ease-out: fast arrival, soft landing. */
export const ease = (p: number): number => 1 - Math.pow(1 - clamp01(p), 3);

/**
 * Part `i` of `n`'s own progress along `build`.
 *
 * The parts overlap heavily rather than arriving strictly one after another. A
 * clean queue reads as a loading sequence; overlapping them reads as one thing
 * appearing, which is what a figure should do in the half-second it gets.
 */
export const stage = (build: number, i: number, n: number, overlap = 0.6): number => {
  if (n <= 1) return clamp01(build);
  const span = 1 / (n - (n - 1) * overlap);
  return clamp01((clamp01(build) - i * span * (1 - overlap)) / span);
};

/** Scale for a part popping in, with a small overshoot at the midpoint. */
export const popScale = (p: number): number =>
  0.55 + 0.45 * ease(p) + 0.09 * Math.sin(Math.PI * clamp01(p));

/** A part's group transform: pop in around a centre. */
export const popAt = (p: number, cx = C, cy = C): string =>
  `translate(${cx} ${cy}) scale(${popScale(p)}) translate(${-cx} ${-cy})`;

/** Points of a regular polygon, for gear teeth, orbits and starbursts. */
export const ring = (n: number, r: number, cx = C, cy = C, phase = 0) =>
  Array.from({length: n}, (_, i) => {
    const a = phase + (i / n) * Math.PI * 2;
    return {x: cx + r * Math.cos(a), y: cy + r * Math.sin(a), a};
  });

/**
 * Props that make a stroke draw itself over `p`.
 *
 * `pathLength={1}` normalises the dash maths so the same three numbers work on
 * any path without measuring it — which is the only reason this is one call
 * instead of a per-figure length constant that goes stale the moment somebody
 * nudges a control point.
 */
export const draws = (p: number) => ({
  pathLength: 1,
  strokeDasharray: 1,
  strokeDashoffset: 1 - ease(p),
});

/** A slow sine, for a float or a breath. */
export const bob = (t: number, amp = 3, speed = 1.7, phase = 0): number =>
  Math.sin(t * speed + phase) * amp;

/** A 0-1 sine, for a glow or a pulse. */
export const pulse = (t: number, speed = 2.4, phase = 0): number =>
  0.5 + 0.5 * Math.sin(t * speed + phase);

/**
 * A value that runs 0-1 and restarts, for anything that travels a path.
 *
 * `phase` offsets one traveller from another so a set of them is never in
 * lockstep, which is what stops three packets on three wires reading as one
 * bar moving.
 */
export const cycle = (t: number, speed = 0.6, phase = 0): number =>
  ((t * speed + phase) % 1 + 1) % 1;

/**
 * Opacity for something travelling a path: up at the start, down at the end,
 * and fully on for the middle. A traveller that pops in and out at full
 * opacity reads as two different objects.
 */
export const fadeTravel = (p: number): number => Math.min(1, Math.sin(p * Math.PI) * 2.6);

/**
 * A periodic gesture that happens and then *stops* — a key turning, a shield
 * confirming, a lid opening.
 *
 * Returns 0-1 over the gesture and -1 while it is at rest, so a caller can skip
 * drawing the moving state entirely rather than leaving it parked at zero. The
 * distinction matters: a pulse ring left drawn at rest opacity between pulses
 * puts a permanent dirty outline around the figure, which is exactly the bug
 * this helper was pulled out of.
 */
export const gesture = (t: number, period = 3.2, span = 0.5): number => {
  const beat = (t % period) / period;
  return beat < span ? beat / span : -1;
};

/**
 * A gesture that goes, holds, and comes back — for a mechanism that latches.
 * `hold` is the fraction of the cycle spent at the far end.
 */
export const swing = (t: number, period = 4, rise = 0.15, hold = 0.45): number => {
  const p = (t % period) / period;
  if (p < 0.25) return 0;
  if (p < 0.25 + rise) return ease((p - 0.25) / rise);
  if (p < 0.25 + rise + hold) return 1;
  const fall = 1 - 0.25 - rise - hold;
  return 1 - ease(Math.min(1, (p - 0.25 - rise - hold) / fall));
};
