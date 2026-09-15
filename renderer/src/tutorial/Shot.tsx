// Footage shots.
//
// The source recording is a 1920x1080 export in which the real app window is
// already inset on an orange card with a lot of margin. Playing that back whole
// is why the original reads as unreadable: the UI runs at roughly 55% of native
// size inside a frame that also has to carry a border.
//
// So none of the source framing survives. A shot names `win` — the rectangle in
// source pixels where the actual app window sits — and that rectangle is scaled
// to fill our own chrome. The chrome takes its aspect from `win`, so nothing is
// ever cropped by accident and no orange edge leaks in. `push` then moves the
// camera during the shot, which is what lets a forty-second static screen
// become a slow reveal of the one control being discussed.
//
// Callouts are declared in the same source-pixel space, so they ride the zoom
// for free instead of needing to be re-placed per keyframe.

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
import {C, MONO, BODY, EASE, ramp, FPS} from './kit';

export const SRC_W = 3456;
export const SRC_H = 2234;

/** A rectangle in source-video pixels. */
export type Rect = {x: number; y: number; w: number; h: number};

/** A highlight drawn over the footage, positioned in source pixels. */
export type Spot = {
  rect: Rect;
  label?: string;
  side?: 'top' | 'bottom' | 'left' | 'right';
  /** Frames into the shot. */
  at: number;
  until?: number;
  /** Scrim everything outside the ring. Off by default: a blanket scrim
   * over a white UI reads as a broken screenshot, not as focus. */
  dim?: boolean;
};

export type ShotSpec = {
  /** Source timestamp in seconds for the shot's first frame. */
  from: number;
  /** Where the app window sits in the source frame. */
  win: Rect;
  /** Zoom factor at the start of the shot (1 = whole window). */
  start?: number;
  /** Zoom factor at the end, plus the point it closes on (normalised in `win`). */
  push?: {to: number; at: [number, number]};
  rate?: number;
  spots?: Spot[];
  label?: string;
};

/** The biggest box of this aspect that fits the frame with room to breathe. */
const MAX_W = 1524;
const MAX_H = 864;
const CHROME_H = 44;

export const boxFor = (win: Rect) => {
  const a = win.w / win.h;
  let w = MAX_W;
  let h = w / a;
  if (h > MAX_H) {
    h = MAX_H;
    w = h * a;
  }
  return {w: Math.round(w), h: Math.round(h)};
};

/** The visible source rect at zoom `f` closing on normalised point `at`. */
const view = (win: Rect, f: number, at: [number, number]): Rect => {
  const w = win.w / f;
  const h = win.h / f;
  const cx = win.x + at[0] * win.w;
  const cy = win.y + at[1] * win.h;
  // Keep the view inside the window so no orange ground leaks in at the edges.
  const x = Math.min(Math.max(cx - w / 2, win.x), win.x + win.w - w);
  const y = Math.min(Math.max(cy - h / 2, win.y), win.y + win.h - h);
  return {x, y, w, h};
};

const place = (r: Rect, box: {w: number; h: number}) => {
  const sc = Math.max(box.w / r.w, box.h / r.h);
  return {
    sc,
    width: SRC_W * sc,
    height: SRC_H * sc,
    left: -r.x * sc + (box.w - r.w * sc) / 2,
    top: -r.y * sc + (box.h - r.h * sc) / 2,
  };
};

const lerpRect = (a: Rect, b: Rect, t: number): Rect => ({
  x: a.x + (b.x - a.x) * t,
  y: a.y + (b.y - a.y) * t,
  w: a.w + (b.w - a.w) * t,
  h: a.h + (b.h - a.h) * t,
});

const SpotBox: React.FC<{spot: Spot; map: ReturnType<typeof place>}> = ({spot, map}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const end = spot.until ?? 1e6;
  if (frame < spot.at - 2 || frame > end + 14) return null;

  const enter = spring({
    frame: frame - spot.at,
    fps,
    config: {damping: 18, stiffness: 150, mass: 0.8},
  });
  const exit = interpolate(frame, [end, end + 12], [1, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const a = Math.min(enter, exit);

  const x = spot.rect.x * map.sc + map.left;
  const y = spot.rect.y * map.sc + map.top;
  const w = spot.rect.w * map.sc;
  const h = spot.rect.h * map.sc;

  const side = spot.side ?? 'top';
  const label: React.CSSProperties = {
    position: 'absolute',
    fontFamily: MONO,
    fontSize: 20,
    letterSpacing: '0.05em',
    color: '#1B1310',
    backgroundColor: C.brand,
    padding: '7px 15px',
    borderRadius: 9,
    whiteSpace: 'nowrap',
    boxShadow: '0 10px 26px rgba(0,0,0,0.5)',
    opacity: a,
    zIndex: 3,
  };
  if (side === 'top') Object.assign(label, {left: x + w / 2, top: y - 50, transform: 'translateX(-50%)'});
  if (side === 'bottom') Object.assign(label, {left: x + w / 2, top: y + h + 16, transform: 'translateX(-50%)'});
  if (side === 'left') Object.assign(label, {left: x - 16, top: y + h / 2, transform: 'translate(-100%,-50%)'});
  if (side === 'right') Object.assign(label, {left: x + w + 16, top: y + h / 2, transform: 'translateY(-50%)'});

  // Two pulses to catch the eye, then nothing. A ring that keeps breathing
  // for the length of the shot is just glitter.
  const age = frame - spot.at;
  const pulse = age < 84 ? (age % 42) / 42 : 1;

  return (
    <>
      <div
        style={{
          position: 'absolute',
          left: x - 6,
          top: y - 6,
          width: w + 12,
          height: h + 12,
          border: `3px solid ${C.brand}`,
          borderRadius: 11,
          boxShadow: spot.dim ? `0 0 0 9999px rgba(8,6,5,${0.34 * a})` : 'none',
          opacity: a,
          transform: `scale(${0.95 + a * 0.05})`,
          zIndex: 2,
        }}
      />
      <div
        style={{
          position: 'absolute',
          left: x - 6,
          top: y - 6,
          width: w + 12,
          height: h + 12,
          border: `2px solid ${C.brandSoft}`,
          borderRadius: 11,
          opacity: a * (1 - pulse) * 0.55,
          transform: `scale(${1 + pulse * 0.07})`,
          zIndex: 2,
        }}
      />
      {spot.label ? <div style={label}>{spot.label}</div> : null}
    </>
  );
};

export const Shot: React.FC<{spec: ShotSpec; durationInFrames: number}> = ({
  spec,
  durationInFrames,
}) => {
  const frame = useCurrentFrame();
  const box = boxFor(spec.win);

  const at = spec.push?.at ?? [0.5, 0.5];
  const f0 = spec.start ?? 1;
  const f1 = spec.push?.to ?? f0;
  const a = view(spec.win, f0, spec.push ? at : [0.5, 0.5]);
  const b = view(spec.win, f1, at);

  // The move happens once, at the top of the shot, and then the camera stops
  // dead. Running the interpolation across the shot's whole length — which is
  // what this used to do — means the frame is never still: a slow creep that
  // reads as drift rather than as a camera move, and is exhausting over eleven
  // minutes. Landing the move early and holding also lets the viewer actually
  // read the screen, which is the entire reason for pushing in.
  const moveLen = Math.min(Math.max(Math.round(durationInFrames * 0.3), 26), 62);
  const t = interpolate(frame, [4, 4 + moveLen], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: EASE,
  });
  const map = place(lerpRect(a, b, t), box);

  // The window lifts in over eight frames so a cut from a graphic slide has a
  // beat rather than a jolt.
  const lift = ramp(frame, 0, 9);

  return (
    <AbsoluteFill style={{alignItems: 'center', justifyContent: 'center'}}>
      <div
        style={{
          position: 'absolute',
          width: box.w * 1.06,
          height: (box.h + CHROME_H) * 1.08,
          borderRadius: 40,
          background: 'radial-gradient(closest-side, rgba(255,150,70,0.16), transparent 72%)',
          filter: 'blur(38px)',
          opacity: lift,
        }}
      />
      <div
        style={{
          width: box.w,
          height: box.h + CHROME_H,
          borderRadius: 16,
          overflow: 'hidden',
          border: '1px solid #ffffff18',
          boxShadow: '0 54px 130px rgba(0,0,0,0.7), 0 0 0 1px rgba(0,0,0,0.6)',
          opacity: lift,
          transform: `translateY(${(1 - lift) * 20}px) scale(${0.986 + lift * 0.014})`,
        }}
      >
        <div
          style={{
            height: CHROME_H,
            backgroundColor: '#181413',
            display: 'flex',
            alignItems: 'center',
            paddingLeft: 17,
            gap: 8,
            borderBottom: '1px solid #ffffff0d',
          }}
        >
          {['#FF5F57', '#FEBC2E', '#28C840'].map((c) => (
            <div key={c} style={{width: 11, height: 11, borderRadius: 6, backgroundColor: c}} />
          ))}
          <div
            style={{
              flex: 1,
              textAlign: 'center',
              fontFamily: MONO,
              fontSize: 17,
              letterSpacing: '0.1em',
              color: '#877B73',
              marginRight: 56,
            }}
          >
            {spec.label ?? ''}
          </div>
        </div>
        <div style={{position: 'relative', width: box.w, height: box.h, overflow: 'hidden'}}>
          <OffthreadVideo
            src={staticFile('remaster/source.mp4')}
            startFrom={Math.round(spec.from * FPS)}
            playbackRate={spec.rate ?? 1}
            muted
            style={{
              position: 'absolute',
              width: map.width,
              height: map.height,
              left: map.left,
              top: map.top,
              maxWidth: 'none',
            }}
          />
          {(spec.spots ?? []).map((sp, i) => (
            <SpotBox key={i} spot={sp} map={map} />
          ))}
        </div>
      </div>
    </AbsoluteFill>
  );
};

/** A pill under a footage shot naming what the viewer is looking at. */
export const ShotNote: React.FC<{text: string; delay?: number}> = ({text, delay = 8}) => {
  const frame = useCurrentFrame();
  const a = ramp(frame, delay, 12);
  return (
    <div
      style={{
        position: 'absolute',
        left: 0,
        right: 0,
        bottom: 26,
        display: 'flex',
        justifyContent: 'center',
        opacity: a,
        transform: `translateY(${(1 - a) * 10}px)`,
        zIndex: 5,
      }}
    >
      <div
        style={{
          fontFamily: BODY,
          fontSize: 26,
          color: C.cream,
          backgroundColor: '#0A0908E6',
          border: `1px solid ${C.line}`,
          padding: '11px 26px',
          borderRadius: 999,
          backdropFilter: 'blur(6px)',
        }}
      >
        {text}
      </div>
    </div>
  );
};
