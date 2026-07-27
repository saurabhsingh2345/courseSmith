// The camera, and the stage it looks at.
//
// Every other scene in this renderer composes directly into the frame. The
// story template does not: it lays its content out on a *world* larger than the
// frame and points a camera at part of it. That indirection buys the one thing
// a run of static compositions cannot have — a shot that is still moving while
// nothing in it changes, which is what separates film from slideshow.
//
// A camera is three numbers: the world point at the centre of frame, and a
// zoom. A move is two of those and an easing between them, evaluated from the
// shot's own progress. Nothing here reads the wall clock or the previous shot,
// so a move is reproducible from (kind, progress) alone and frames stay
// independent.

import {FRAME_H, FRAME_W} from './Stage';

export type Cam = {x: number; y: number; zoom: number};

/**
 * The world the camera moves over.
 *
 * Deliberately larger than the frame. A pan across a world exactly frame-sized
 * would slide the content off its own edge; the margin is what the camera has
 * room to move *within*, and it is why `pan` reveals stage rather than
 * revealing background.
 */
export const WORLD = {w: Math.round(FRAME_W * 1.34), h: Math.round(FRAME_H * 1.28)};

/** The neutral camera: the middle of the world, unzoomed. */
export const restCam = (): Cam => ({x: WORLD.w / 2, y: WORLD.h / 2, zoom: 1});

/**
 * Smoothstep. A camera move that starts and stops abruptly reads as a glitch
 * even at small amplitudes — the eye is far more sensitive to the derivative of
 * a camera than to its position.
 */
const smooth = (p: number): number => {
  const c = Math.max(0, Math.min(1, p));
  return c * c * (3 - 2 * c);
};

/**
 * The endpoints of a named move.
 *
 * Amplitudes are small on purpose. A 14% push over eight seconds is barely
 * perceptible frame to frame and unmistakable across the shot, which is exactly
 * what a camera move should be. The first pass used roughly double these and
 * every shot read as a zoom effect rather than as a camera.
 */
export const cameraMove = (kind: string): {from: Cam; to: Cam} => {
  const base = restCam();
  const at = (dx: number, dy: number, zoom: number): Cam => ({
    x: base.x + dx,
    y: base.y + dy,
    zoom,
  });
  switch (kind) {
    case 'push':
      return {from: at(0, 0, 1), to: at(0, -10, 1.14)};
    case 'pull':
      return {from: at(0, -8, 1.16), to: at(0, 0, 1)};
    case 'pan':
      return {from: at(-52, 0, 1.05), to: at(52, 0, 1.05)};
    case 'rise':
      return {from: at(0, 44, 1.04), to: at(0, -44, 1.04)};
    case 'drift':
      return {from: at(-34, 20, 1.0), to: at(30, -18, 1.06)};
    case 'hold':
    default:
      return {from: base, to: base};
  }
};

/** The camera at progress `p` (0-1) through a move. */
export const camAt = (kind: string, p: number): Cam => {
  const {from, to} = cameraMove(kind);
  const e = smooth(p);
  return {
    x: from.x + (to.x - from.x) * e,
    y: from.y + (to.y - from.y) * e,
    zoom: from.zoom + (to.zoom - from.zoom) * e,
  };
};

/**
 * A camera that follows another one only partly — the parallax backdrop.
 *
 * Depth on a flat stage is entirely a matter of differential motion: if the
 * background tracks the camera exactly it is painted on the front element, and
 * if it does not move at all the stage reads as a cutout floating on wallpaper.
 */
export const parallaxCam = (cam: Cam, factor: number): Cam => {
  const base = restCam();
  return {
    x: base.x + (cam.x - base.x) * factor,
    y: base.y + (cam.y - base.y) * factor,
    zoom: 1 + (cam.zoom - 1) * factor,
  };
};

/**
 * The CSS transform that puts world point (cam.x, cam.y) at the centre of the
 * frame at the given zoom.
 *
 * Read right to left: move the world so the camera's target sits at the origin,
 * scale about that origin, then move the origin to the middle of the frame.
 * Written in the other order it silently zooms about the world's top-left
 * corner, which looks like the shot sliding away as it zooms.
 */
export const camTransform = (cam: Cam): string =>
  `translate(${FRAME_W / 2}px, ${FRAME_H / 2}px) scale(${cam.zoom}) translate(${-cam.x}px, ${-cam.y}px)`;
