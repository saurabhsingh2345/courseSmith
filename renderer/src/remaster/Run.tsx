// A continuous run of the source recording.
//
// This replaces the earlier approach, which cropped the app window out of each
// frame, remounted it in its own chrome, and pushed the camera in over the
// length of every shot. That made the UI readable but it cost the thing that
// mattered more: continuity. Thirty-one shots sampled from all over the source
// timeline read as a slideshow of moments rather than as somebody using a
// computer, and a camera that never stops moving is exhausting to watch.
//
// So the footage is now played exactly as recorded — full frame, 1:1, at real
// speed, for a long unbroken stretch. Nothing is cropped, scaled, or re-timed.
// The only thing drawn on top is the highlight box, and because the frame is
// untransformed its coordinates are just source pixels.

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

export type Rect = {x: number; y: number; w: number; h: number};

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
  spots?: Spot[];
};

const SpotBox: React.FC<{spot: Spot}> = ({spot}) => {
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
    fontSize: 21,
    letterSpacing: '0.05em',
    color: '#1B1310',
    backgroundColor: C.brand,
    padding: '8px 16px',
    borderRadius: 9,
    whiteSpace: 'nowrap',
    boxShadow: '0 8px 22px rgba(0,0,0,0.35)',
    opacity: a,
  };
  if (side === 'top') Object.assign(label, {left: x + w / 2, top: y - 52, transform: 'translateX(-50%)'});
  if (side === 'bottom') Object.assign(label, {left: x + w / 2, top: y + h + 16, transform: 'translateX(-50%)'});
  if (side === 'left') Object.assign(label, {left: x - 16, top: y + h / 2, transform: 'translate(-100%,-50%)'});
  if (side === 'right') Object.assign(label, {left: x + w + 16, top: y + h / 2, transform: 'translateY(-50%)'});

  return (
    <>
      <div
        style={{
          position: 'absolute',
          left: x - 5,
          top: y - 5,
          width: w + 10,
          height: h + 10,
          border: `3px solid ${C.brand}`,
          borderRadius: 10,
          opacity: a,
        }}
      />
      {spot.label ? <div style={label}>{spot.label}</div> : null}
    </>
  );
};

export const Run: React.FC<{spec: RunSpec}> = ({spec}) => (
  <AbsoluteFill>
    <OffthreadVideo
      src={staticFile('remaster/source.mp4')}
      startFrom={Math.round(spec.from * FPS)}
      endAt={Math.round(spec.to * FPS)}
      muted
      style={{width: '100%', height: '100%'}}
    />
    {(spec.spots ?? []).map((s, i) => (
      <SpotBox key={i} spot={s} />
    ))}
  </AbsoluteFill>
);
