// The course animation language for the video renderer.
//
// The tokens themselves are owned by Go (internal/pipeline/motion.go) and
// mirrored into motion.defaults.json by a drift-guarded test — never edit the
// JSON by hand. This module adds the TypeScript types plus the frame/easing
// helpers that turn seconds-and-cubic-bezier tokens into the frames-and-Easing
// that Remotion speaks. Components consume tokens through here so retuning the
// Go defaults (or an archetype override in the scene graph) retunes every scene.

import {Easing} from 'remotion';
import motionDefaults from './motion.defaults.json';

export type MotionTiming = {fast: number; normal: number; slow: number; verySlow: number};
export type MotionEasing = {entrance: string; exit: string; subtle: string};
export type MotionStagger = {words: number; items: number; connections: number};
export type MotionTokens = {
  timing: MotionTiming;
  easing: MotionEasing;
  stagger: MotionStagger;
};

/** Baseline tokens, byte-synced to Go's DefaultMotion(). */
export const defaultMotion: MotionTokens = motionDefaults;

/**
 * Merge scene-graph motion (which may be absent on older graphs) over the
 * baseline. The renderer should always call this with the scene graph's
 * `motion` field so archetype overrides flow through.
 */
export const resolveMotion = (m?: Partial<MotionTokens> | null): MotionTokens => ({
  timing: {...defaultMotion.timing, ...m?.timing},
  easing: {...defaultMotion.easing, ...m?.easing},
  stagger: {...defaultMotion.stagger, ...m?.stagger},
});

/** Seconds → whole frames at the given fps (min 1 so nothing is zero-length). */
export const secondsToFrames = (fps: number, seconds: number): number =>
  Math.max(1, Math.round(seconds * fps));

/** Parse "cubic-bezier(a, b, c, d)" → [a,b,c,d]; throws on malformed input. */
export const parseBezier = (s: string): [number, number, number, number] => {
  const m = s.trim().match(/^cubic-bezier\(([^)]+)\)$/);
  if (!m) throw new Error(`not a cubic-bezier(): ${s}`);
  const parts = m[1].split(',').map((p) => Number(p.trim()));
  if (parts.length !== 4 || parts.some((n) => Number.isNaN(n))) {
    throw new Error(`invalid cubic-bezier control points: ${s}`);
  }
  return [parts[0], parts[1], parts[2], parts[3]];
};

/** A cubic-bezier token → a Remotion Easing function for interpolate({easing}). */
export const bezierEasing = (token: string) => {
  const [x1, y1, x2, y2] = parseBezier(token);
  return Easing.bezier(x1, y1, x2, y2);
};

/** Frame at which the i-th staggered element begins, relative to a base frame. */
export const staggerFrame = (fps: number, base: number, step: number, i: number): number =>
  base + Math.round(step * fps) * i;
