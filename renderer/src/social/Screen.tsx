// The app, on camera.
//
// Lumen is not screenshotted. Its real markup is mounted inside the composition
// and the camera is a CSS transform over it, which buys three things a still
// could not:
//
//  1. Sharpness at any zoom. Chrome re-rasterises text at the scaled size, so a
//     2x push into the caption is as crisp as the wide. This machine's display
//     caps a browser capture at ~1440x790 — under the render's own 1920x1080 —
//     so every screenshot would have been an upscale.
//  2. Real state changes. A like, a new comment, a posted photo, a filter — the
//     film asks for the state and the app renders it, so a transition between
//     two states is a genuine before and after rather than a dissolve.
//  3. Scrolling. `.main` is translated while the sidebar stays put, which is
//     exactly what the page does under a real scroll.
//
// Coordinates in a shot are Lumen's own logical pixels (1440x900), so a callout
// is placed once and rides every zoom for free.

import React from 'react';
import {
  AbsoluteFill,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
} from 'remotion';
import {staticFile} from 'remotion';
import {C, MONO, BODY, EASE, ramp, FPS, GRAD} from './kit';
import {CSS, pageHTML, setBase} from './lumen.gen';

/** Lumen's logical viewport — a 1440x900 browser window. */
export const APP_W = 1440;
export const APP_H = 900;

/** The window box in the 1920x1080 frame. 1500/1440 ≈ 1.04, so near 1:1. */
const BOX_W = 1500;
const BOX_H = Math.round((APP_H * BOX_W) / APP_W); // 938
const CHROME_H = 52;

export type Rect = {x: number; y: number; w: number; h: number};

export type Spot = {
  rect: Rect;
  label?: string;
  side?: 'top' | 'bottom' | 'left' | 'right';
  /** Frames into the shot. */
  at: number;
  until?: number;
};

/** Where the pointer is, in app pixels, at a given frame of the shot. */
export type Move = {at: number; to: [number, number]; click?: boolean};

export type ScreenSpec = {
  /** Any subset of Lumen's state: view, likes, draft, open, posted… */
  state?: Record<string, unknown>;
  /** Page scroll, in app px. A pair animates over the shot. */
  scroll?: number | [number, number];
  /** Zoom at the top of the shot (1 = whole window). */
  start?: number;
  /** Zoom to push to, and the normalised point it closes on. */
  push?: {to: number; at: [number, number]};
  spots?: Spot[];
  cursor?: Move[];
  /** Ring the address pill in the browser bar, at this frame. */
  urlSpot?: {at: number; label?: string};
  /** The address shown in the browser bar. */
  url?: string;
};

// --------------------------------------------------------------------- camera

const view = (f: number, at: [number, number]): Rect => {
  const w = APP_W / f;
  const h = APP_H / f;
  const cx = at[0] * APP_W;
  const cy = at[1] * APP_H;
  const x = Math.min(Math.max(cx - w / 2, 0), APP_W - w);
  const y = Math.min(Math.max(cy - h / 2, 0), APP_H - h);
  return {x, y, w, h};
};

const place = (r: Rect) => {
  const sc = BOX_W / r.w;
  return {sc, left: -r.x * sc, top: -r.y * sc, width: APP_W * sc, height: APP_H * sc};
};

const lerpRect = (a: Rect, b: Rect, t: number): Rect => ({
  x: a.x + (b.x - a.x) * t,
  y: a.y + (b.y - a.y) * t,
  w: a.w + (b.w - a.w) * t,
  h: a.h + (b.h - a.h) * t,
});

// --------------------------------------------------------------------- callout

const SpotBox: React.FC<{spot: Spot; map: ReturnType<typeof place>; scrollY: number}> = ({
  spot,
  map,
  scrollY,
}) => {
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
  const y = (spot.rect.y - scrollY) * map.sc + map.top;
  const w = spot.rect.w * map.sc;
  const h = spot.rect.h * map.sc;

  // Two pulses, then it settles. A ring that keeps breathing is glitter.
  const age = frame - spot.at;
  const pulse = age < 80 ? (age % 40) / 40 : 1;

  const side = spot.side ?? 'top';
  const label: React.CSSProperties = {
    position: 'absolute',
    fontFamily: MONO,
    fontSize: 20,
    fontWeight: 500,
    letterSpacing: '0.04em',
    color: '#100A12',
    backgroundImage: GRAD,
    padding: '8px 16px',
    borderRadius: 10,
    whiteSpace: 'nowrap',
    boxShadow: '0 12px 30px rgba(0,0,0,0.55)',
    opacity: a,
    zIndex: 4,
  };
  const clampY = (v: number) => Math.max(10, Math.min(v, BOX_H - 52));
  const clampX = (v: number) => Math.max(120, Math.min(v, BOX_W - 120));
  if (side === 'top')
    Object.assign(label, {left: clampX(x + w / 2), top: clampY(y - 52), transform: 'translateX(-50%)'});
  if (side === 'bottom')
    Object.assign(label, {left: clampX(x + w / 2), top: clampY(y + h + 16), transform: 'translateX(-50%)'});
  if (side === 'left')
    Object.assign(label, {left: x - 18, top: clampY(y + h / 2), transform: 'translate(-100%,-50%)'});
  if (side === 'right')
    Object.assign(label, {left: x + w + 18, top: clampY(y + h / 2), transform: 'translateY(-50%)'});

  return (
    <>
      <div
        style={{
          position: 'absolute',
          left: x - 6,
          top: y - 6,
          width: w + 12,
          height: h + 12,
          border: `3px solid ${C.rose}`,
          borderRadius: 12,
          opacity: a,
          transform: `scale(${0.96 + a * 0.04})`,
          zIndex: 3,
        }}
      />
      <div
        style={{
          position: 'absolute',
          left: x - 6,
          top: y - 6,
          width: w + 12,
          height: h + 12,
          border: `2px solid ${C.violet}`,
          borderRadius: 12,
          opacity: a * (1 - pulse) * 0.6,
          transform: `scale(${1 + pulse * 0.06})`,
          zIndex: 3,
        }}
      />
      {spot.label ? <div style={label}>{spot.label}</div> : null}
    </>
  );
};

// --------------------------------------------------------------------- cursor

const Cursor: React.FC<{moves: Move[]; map: ReturnType<typeof place>; scrollY: number}> = ({
  moves,
  map,
  scrollY,
}) => {
  const frame = useCurrentFrame();
  if (!moves.length || frame < moves[0].at) return null;

  // Where the pointer is: the last move it has reached, eased toward the next.
  let i = 0;
  for (let k = 0; k < moves.length; k++) if (frame >= moves[k].at) i = k;
  const from = moves[Math.max(0, i - 1)];
  const cur = moves[i];
  const t = i === 0 ? 1 : ramp(frame, from.at, Math.max(8, cur.at - from.at));
  const px = from.to[0] + (cur.to[0] - from.to[0]) * t;
  const py = from.to[1] + (cur.to[1] - from.to[1]) * t;

  const x = px * map.sc + map.left;
  const y = (py - scrollY) * map.sc + map.top;

  // The click ring, on the frame the pointer lands.
  const clickAge = cur.click ? frame - cur.at : -1;
  const ring = clickAge >= 0 && clickAge < 22 ? clickAge / 22 : -1;
  const press = clickAge >= 0 && clickAge < 8 ? 1 - clickAge / 8 : 0;

  return (
    <>
      {ring >= 0 ? (
        <div
          style={{
            position: 'absolute',
            left: x,
            top: y,
            width: 20 + ring * 58,
            height: 20 + ring * 58,
            marginLeft: -(10 + ring * 29),
            marginTop: -(10 + ring * 29),
            borderRadius: '50%',
            border: `2.5px solid ${C.rose}`,
            opacity: (1 - ring) * 0.85,
            zIndex: 5,
          }}
        />
      ) : null}
      <svg
        width={30}
        height={38}
        viewBox="0 0 30 38"
        style={{
          position: 'absolute',
          left: x,
          top: y,
          zIndex: 6,
          transform: `translate(-3px,-2px) scale(${1 - press * 0.18})`,
          filter: 'drop-shadow(0 3px 7px rgba(0,0,0,0.55))',
        }}
      >
        <path d="M3 2 L3 27.5 L9.6 21.4 L14.2 32.6 L18.6 30.7 L14.1 19.8 L23 19.2 Z"
              fill="#FFFFFF" stroke="#0D0F12" strokeWidth={1.7} strokeLinejoin="round" />
      </svg>
    </>
  );
};

// --------------------------------------------------------------------- chrome

const UrlRing: React.FC<{at: number; label?: string}> = ({at, label}) => {
  const frame = useCurrentFrame();
  const a = ramp(frame, at, 12);
  if (frame < at) return null;
  return (
    <>
      <div
        style={{
          position: 'absolute',
          left: 60,
          top: 8,
          width: 300,
          height: 36,
          border: `2.5px solid ${C.rose}`,
          borderRadius: 9,
          opacity: a,
          zIndex: 9,
        }}
      />
      {label ? (
        <div
          style={{
            position: 'absolute',
            left: 380,
            top: 9,
            fontFamily: MONO,
            fontSize: 19,
            color: '#100A12',
            backgroundImage: GRAD,
            padding: '7px 14px',
            borderRadius: 9,
            whiteSpace: 'nowrap',
            opacity: a,
            zIndex: 9,
          }}
        >
          {label}
        </div>
      ) : null}
    </>
  );
};

const BrowserBar: React.FC<{url: string}> = ({url}) => (
  <div
    style={{
      height: CHROME_H,
      background: 'linear-gradient(180deg,#1B1826,#15121F)',
      display: 'flex',
      alignItems: 'center',
      gap: 9,
      paddingLeft: 18,
      paddingRight: 18,
      borderBottom: '1px solid rgba(255,255,255,0.07)',
    }}
  >
    {['#FF5F57', '#FEBC2E', '#28C840'].map((c) => (
      <div key={c} style={{width: 11, height: 11, borderRadius: 6, backgroundColor: c}} />
    ))}
    <div
      style={{
        flex: 1,
        margin: '0 18px',
        height: 30,
        borderRadius: 8,
        background: 'rgba(255,255,255,0.06)',
        border: '1px solid rgba(255,255,255,0.06)',
        display: 'flex',
        alignItems: 'center',
        paddingLeft: 13,
        gap: 9,
      }}
    >
      <svg width={13} height={13} viewBox="0 0 24 24" fill="none" stroke={C.ok} strokeWidth={2.4}>
        <rect x="4" y="10.5" width="16" height="10.5" rx="2.4" />
        <path d="M8 10.5V7.6a4 4 0 0 1 8 0v2.9" />
      </svg>
      <span style={{fontFamily: MONO, fontSize: 16, color: 'rgba(255,255,255,0.5)'}}>{url}</span>
    </div>
    <div style={{width: 46}} />
  </div>
);

// --------------------------------------------------------------------- shot

export const Screen: React.FC<{spec: ScreenSpec; durationInFrames: number}> = ({
  spec,
  durationInFrames,
}) => {
  const frame = useCurrentFrame();

  setBase(staticFile('lumen/'));

  const at = spec.push?.at ?? [0.5, 0.5];
  const f0 = spec.start ?? 1;
  const f1 = spec.push?.to ?? f0;
  const a = view(f0, spec.push ? at : [0.5, 0.5]);
  const b = view(f1, at);

  // The move lands early and then the camera stops dead. Interpolating across
  // the whole shot means the frame is never still — a creep that reads as drift
  // and is exhausting to sit through, and it also stops the viewer reading the
  // screen, which is the only reason to push in.
  const moveLen = Math.min(Math.max(Math.round(durationInFrames * 0.32), 28), 68);
  const t = ramp(frame, 5, moveLen);
  const map = place(lerpRect(a, b, t));

  // Scroll travels over the whole shot: it *is* the content, not a move onto it.
  const sc = spec.scroll ?? 0;
  const scrollY = Array.isArray(sc)
    ? interpolate(frame, [0, durationInFrames], sc, {
        extrapolateLeft: 'clamp',
        extrapolateRight: 'clamp',
        easing: EASE,
      })
    : sc;

  const html = React.useMemo(() => pageHTML(spec.state ?? {}), [spec.state]);
  const lift = ramp(frame, 0, 10);

  return (
    <AbsoluteFill style={{alignItems: 'center', justifyContent: 'center'}}>
      {/* the light the window throws on the ground behind it */}
      <div
        style={{
          position: 'absolute',
          width: BOX_W * 1.1,
          height: (BOX_H + CHROME_H) * 1.12,
          borderRadius: 60,
          background:
            'radial-gradient(closest-side, rgba(255,77,109,0.16), rgba(139,92,246,0.10) 55%, transparent 76%)',
          opacity: lift,
        }}
      />
      <div
        style={{
          width: BOX_W,
          height: BOX_H + CHROME_H,
          borderRadius: 18,
          overflow: 'hidden',
          border: '1px solid rgba(255,255,255,0.11)',
          boxShadow:
            '0 60px 140px rgba(0,0,0,0.72), 0 0 0 1px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.10)',
          opacity: lift,
          transform: `translateY(${(1 - lift) * 18}px) scale(${0.988 + lift * 0.012})`,
        }}
      >
        <div style={{position: 'relative'}}>
          <BrowserBar url={spec.url ?? 'localhost:8000'} />
          {spec.urlSpot ? <UrlRing {...spec.urlSpot} /> : null}
        </div>
        <div style={{position: 'relative', width: BOX_W, height: BOX_H, overflow: 'hidden'}}>
          <div
            style={{
              position: 'absolute',
              left: map.left,
              top: map.top,
              width: map.width,
              height: map.height,
              transformOrigin: '0 0',
            }}
          >
            {/* The app, at its own logical size, scaled into the box. */}
            <div
              style={{
                width: APP_W,
                height: APP_H,
                transform: `scale(${map.sc})`,
                transformOrigin: '0 0',
                overflow: 'hidden',
                background: '#FBFBFC',
              }}
            >
              <style>
                {CSS}
                {`
                  .app { --ui: ${BODY}; --ui-display: ${BODY}; }
                  .app, .app * { animation: none !important; }
                  .side { height: ${APP_H}px; }
                  .main { transform: translateY(${-scrollY}px); }
                `}
              </style>
              <div
                className="app"
                style={{display: 'flex', width: APP_W, minHeight: APP_H}}
                dangerouslySetInnerHTML={{__html: html}}
              />
            </div>
          </div>
          {(spec.spots ?? []).map((sp, i) => (
            <SpotBox key={i} spot={sp} map={map} scrollY={scrollY} />
          ))}
          {spec.cursor?.length ? (
            <Cursor moves={spec.cursor} map={map} scrollY={scrollY} />
          ) : null}
        </div>
      </div>
    </AbsoluteFill>
  );
};

/** A pill under a shot naming what the viewer is looking at. */
export const ScreenNote: React.FC<{text: string; delay?: number}> = ({text, delay = 10}) => {
  const frame = useCurrentFrame();
  const a = ramp(frame, delay, 12);
  return (
    <div
      style={{
        position: 'absolute',
        left: 0,
        right: 0,
        bottom: 24,
        display: 'flex',
        justifyContent: 'center',
        opacity: a,
        transform: `translateY(${(1 - a) * 10}px)`,
        zIndex: 8,
      }}
    >
      <div
        style={{
          fontFamily: BODY,
          fontWeight: 400,
          fontSize: 25,
          color: C.text,
          background: 'rgba(10,8,18,0.86)',
          border: `1px solid ${C.line}`,
          padding: '11px 26px',
          borderRadius: 999,
        }}
      >
        {text}
      </div>
    </div>
  );
};
