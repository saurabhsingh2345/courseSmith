// A continuous run of the source recording.
//
// The footage is the whole laptop screen at 3456 by 2234, which is taller than
// 16:9. Stretching it to fill the film would squash every circle in the UI, and
// letterboxing it would waste a quarter of the frame, so instead each run names
// a `win` — the rectangle of the source it wants on screen — and that rectangle
// is fitted to the film. The default window is a 16:9 band that keeps the
// conversation and the composer and drops only the menu bar and the model row,
// which is what you want in almost every shot.
//
// Spots are declared in SOURCE pixels and live inside the same transformed
// layer as the video, so they ride the crop for free. Their labels are
// counter-scaled, because a 21px label inside a layer drawn at 0.55 would
// arrive as 12px and be unreadable.

import React from 'react';
import {
  AbsoluteFill,
  OffthreadVideo,
  staticFile,
  useCurrentFrame,
  useVideoConfig,
  interpolate,
  spring,
} from 'remotion';
import {C, MONO, FPS} from './kit';

export const SRC_W = 3456;
export const SRC_H = 2234;

export type Rect = {x: number; y: number; w: number; h: number};

/**
 * Bottom-anchored: y + h lands exactly on the bottom edge of the frame, so the
 * composer, the folder chip and the model row are ALWAYS fully in shot. Lesson
 * one used a band starting at y 150, which ended at 2094 and sliced the bottom
 * off the prompt box — the one thing a viewer most needs to read.
 */
export const DEFAULT_WIN: Rect = {x: 0, y: 66, w: 3456, h: 2168};

export type Spot = {
  rect: Rect;
  label?: string;
  side?: 'top' | 'bottom' | 'left' | 'right';
  /** Seconds into the run. */
  at: number;
  until?: number;
};

export type RunSpec = {
  /** Source timestamps, in seconds. The run plays every frame between them. */
  from: number;
  to: number;
  /**
   * Play faster than life. A rate of 8 turns eighty seconds of waiting into ten
   * seconds of film. The beat's length is (to - from) / rate, computed in the
   * voice builder so picture and narration agree.
   */
  rate?: number;
  /** Region of the source frame to show. Defaults to DEFAULT_WIN. */
  win?: Rect;
  spots?: Spot[];
};

const SpotBox: React.FC<{spot: Spot; k: number}> = ({spot, k}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const a0 = spot.at * FPS;
  const a1 = (spot.until ?? spot.at + 6) * FPS;
  if (frame < a0 - 2 || frame > a1 + 16) return null;

  const enter = spring({
    frame: frame - a0,
    fps,
    config: {damping: 20, stiffness: 130, mass: 0.9},
  });
  const exit = interpolate(frame, [a1, a1 + 14], [1, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const a = Math.min(enter, exit);
  const {x, y, w, h} = spot.rect;

  const side = spot.side ?? 'top';
  const label: React.CSSProperties = {
    position: 'absolute',
    fontFamily: MONO,
    fontSize: 21 * k,
    letterSpacing: '0.05em',
    color: '#FFFFFF',
    backgroundColor: C.brand,
    padding: `${8 * k}px ${16 * k}px`,
    borderRadius: 9 * k,
    whiteSpace: 'nowrap',
    boxShadow: `0 ${8 * k}px ${22 * k}px rgba(0,0,0,0.28)`,
    opacity: a,
  };
  const gap = 52 * k;
  if (side === 'top') Object.assign(label, {left: x + w / 2, top: y - gap, transform: 'translateX(-50%)'});
  if (side === 'bottom') Object.assign(label, {left: x + w / 2, top: y + h + 16 * k, transform: 'translateX(-50%)'});
  if (side === 'left') Object.assign(label, {left: x - 16 * k, top: y + h / 2, transform: 'translate(-100%,-50%)'});
  if (side === 'right') Object.assign(label, {left: x + w + 16 * k, top: y + h / 2, transform: 'translateY(-50%)'});

  return (
    <>
      <div
        style={{
          position: 'absolute',
          left: x - 5 * k,
          top: y - 5 * k,
          width: w + 10 * k,
          height: h + 10 * k,
          border: `${3 * k}px solid ${C.brand}`,
          borderRadius: 10 * k,
          opacity: a,
        }}
      />
      {spot.label ? <div style={label}>{spot.label}</div> : null}
    </>
  );
};

export const Run: React.FC<{spec: RunSpec}> = ({spec}) => {
  const {width, height} = useVideoConfig();
  const win = spec.win ?? DEFAULT_WIN;
  const scale = Math.min(width / win.w, height / win.h);
  const dispW = win.w * scale;
  const dispH = win.h * scale;

  return (
    <AbsoluteFill style={{backgroundColor: C.ink0}}>
      <div
        style={{
          position: 'absolute',
          left: (width - dispW) / 2,
          top: (height - dispH) / 2,
          width: dispW,
          height: dispH,
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            position: 'absolute',
            width: SRC_W,
            height: SRC_H,
            transformOrigin: '0 0',
            transform: `scale(${scale}) translate(${-win.x}px, ${-win.y}px)`,
          }}
        >
          <OffthreadVideo
            src={staticFile('tutorial3/source.mp4')}
            startFrom={Math.round(spec.from * FPS)}
            endAt={Math.round(spec.to * FPS)}
            playbackRate={spec.rate ?? 1}
            muted
            style={{width: SRC_W, height: SRC_H, display: 'block'}}
          />
          {(spec.spots ?? []).map((s, i) => (
            <SpotBox key={i} spot={s} k={1 / scale} />
          ))}
        </div>
      </div>
    </AbsoluteFill>
  );
};
