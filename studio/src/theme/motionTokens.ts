// The studio's copy of the course motion language. The values are owned by Go
// (internal/pipeline/motion.go) and mirrored into motion.defaults.json by a
// drift-guarded test — identical to the renderer's copy so a video preview in
// the studio moves exactly like the rendered video. Never edit the JSON.

import motionDefaults from './motion.defaults.json';

export type MotionTiming = {fast: number; normal: number; slow: number; verySlow: number};
export type MotionEasing = {entrance: string; exit: string; subtle: string};
export type MotionStagger = {words: number; items: number; connections: number};
export type MotionTokens = {
  timing: MotionTiming;
  easing: MotionEasing;
  stagger: MotionStagger;
};

export const motion: MotionTokens = motionDefaults;

/** Seconds → milliseconds, for Framer Motion / CSS transitions in the UI. */
export const ms = (seconds: number): number => Math.round(seconds * 1000);
