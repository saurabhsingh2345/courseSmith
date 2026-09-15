// Our own browser and window chrome.
//
// The point: record ONLY the page, never the browser. Then mount that recording
// in a frame we designed. No stray toolbars, no bookmark bars, no extension
// icons, no profile avatar, no OS hints — and every site in the program sits in
// the same frame, so the films read as one piece.
//
// Same tokens as the deck: near-black ground, IBM Plex Mono, and the red /
// yellow / blue accents used for nothing else.

import React from 'react';
import {AbsoluteFill, Img, OffthreadVideo, interpolate, staticFile, useCurrentFrame} from 'remotion';
import {C, MONO, DISPLAY, FPS} from './kit';

const bez = (ax: number, ay: number, bx: number, by: number) => (t: number) => {
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
const ease = bez(0.16, 0.84, 0.32, 1);
const pr = (f: number, d: number, len: number) =>
  Math.max(0, Math.min(1, (f / FPS - d) / len));

const Lock: React.FC<{size?: number}> = ({size = 13}) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none"
       style={{flexShrink: 0, opacity: 0.5}}>
    <rect x="5" y="10.5" width="14" height="10" rx="2.4" fill={C.muted} />
    <path d="M8.4 10.5V7.8a3.6 3.6 0 0 1 7.2 0v2.7" stroke={C.muted}
          strokeWidth="2.1" strokeLinecap="round" />
  </svg>
);

export type BrowserProps = {
  /** page recording in public/, e.g. 'shots/cursor.mp4' */
  src?: string;
  /** or a still of the page */
  img?: string;
  url: string;
  /** optional tab label; omitted gives the cleaner single-bar frame */
  tab?: string;
  /** 0..1 of the frame width the window occupies */
  scale?: number;
  /** delay before the window arrives, seconds */
  d?: number;
  /** page content scrolls up by this many px over the shot */
  drift?: number;
  /** frame aspect — match the recording so nothing is cropped */
  aspect?: string;
};

/** the window itself, with no caption — compose it into any slide */
export const BrowserWindow: React.FC<BrowserProps> = ({
  src, img, url, tab, scale = 0.78, d = 0.15, drift = 0, aspect = '16 / 10',
}) => {
  const f = useCurrentFrame();
  const p = ease(pr(f, d, 0.75));
  const bar = tab ? 92 : 58;
  const y = drift ? interpolate(pr(f, d + 0.6, 8), [0, 1], [0, -drift]) : 0;
  return (
    <div style={{
      width: `${scale * 100}%`,
      aspectRatio: aspect,
      borderRadius: 16,
      overflow: 'hidden',
      background: '#0c0c0c',
      border: '1px solid rgba(255,255,255,0.10)',
      boxShadow: `0 ${44 * p}px ${110 * p}px rgba(0,0,0,${0.62 * p}), `
               + `0 2px 0 rgba(255,255,255,0.05) inset`,
      opacity: p,
      transform: `translateY(${34 * (1 - p)}px) scale(${0.965 + 0.035 * p})`,
      display: 'flex',
      flexDirection: 'column',
    }}>
      {/* --- chrome ------------------------------------------------------ */}
      <div style={{
        height: bar, flexShrink: 0, background: '#141414',
        borderBottom: '1px solid rgba(255,255,255,0.07)',
        display: 'flex', flexDirection: 'column', justifyContent: 'flex-end',
      }}>
        {tab ? (
          <div style={{display: 'flex', alignItems: 'flex-end', padding: '0 14px', gap: 8, height: 40}}>
            <div style={{
              display: 'flex', alignItems: 'center', gap: 9,
              background: '#0c0c0c', borderRadius: '9px 9px 0 0',
              padding: '9px 18px 10px', fontFamily: MONO, fontSize: 13.5,
              color: '#c8c4bb', maxWidth: '46%', whiteSpace: 'nowrap',
              overflow: 'hidden', textOverflow: 'ellipsis',
            }}>
              <span style={{width: 7, height: 7, borderRadius: 99, background: C.yellow,
                            flexShrink: 0, opacity: 0.85}} />
              {tab}
            </div>
          </div>
        ) : null}
        <div style={{display: 'flex', alignItems: 'center', gap: 16, padding: '0 18px', height: 52}}>
          <div style={{display: 'flex', gap: 8, flexShrink: 0}}>
            {[C.red, C.yellow, '#3ba55d'].map((c, i) => (
              <span key={i} style={{
                width: 11, height: 11, borderRadius: 99, background: c, opacity: 0.55,
              }} />
            ))}
          </div>
          <div style={{
            flex: 1, height: 32, borderRadius: 99, background: '#0b0b0b',
            border: '1px solid rgba(255,255,255,0.07)',
            display: 'flex', alignItems: 'center', gap: 9, padding: '0 15px',
            fontFamily: MONO, fontSize: 14.5, color: '#9b968c',
            letterSpacing: '0.01em', overflow: 'hidden', whiteSpace: 'nowrap',
          }}>
            <Lock />
            {url}
          </div>
        </div>
      </div>
      {/* --- the page ---------------------------------------------------- */}
      <div style={{flex: 1, overflow: 'hidden', background: '#000', position: 'relative'}}>
        {src ? (
          <OffthreadVideo src={staticFile(src)} muted style={{
            width: '100%', height: '100%', objectFit: 'cover',
            transform: `translateY(${y}px)`,
          }} />
        ) : img ? (
          <Img src={staticFile(img)} style={{
            width: '100%', display: 'block',
            transform: `translateY(${y}px)`,
          }} />
        ) : null}
      </div>
    </div>
  );
};

/** a full slide: the window, centred, over the deck ground */
export const BrowserSlide: React.FC<BrowserProps & {
  kicker?: string; lines?: string[]; size?: number;
}> = ({kicker, lines, size = 54, ...b}) => {
  const f = useCurrentFrame();
  const p = ease(pr(f, 0.05, 0.6));
  return (
    <AbsoluteFill style={{background: C.bg, alignItems: 'center',
                          justifyContent: 'center', padding: '4% 5%'}}>
      {lines ? (
        <div style={{
          width: '100%', textAlign: 'center', marginBottom: 30, opacity: p,
          transform: `translateY(${14 * (1 - p)}px)`,
        }}>
          {kicker ? (
            <div style={{
              fontFamily: MONO, letterSpacing: '0.22em', textTransform: 'uppercase',
              fontSize: 15, color: C.muted, marginBottom: 13,
            }}>{kicker}</div>
          ) : null}
          <div style={{
            fontFamily: DISPLAY, textTransform: 'uppercase', color: C.ink,
            letterSpacing: '-0.03em', lineHeight: 0.95, fontSize: size,
          }}>
            {lines.map((ln, i) => (
              <div key={i}>
                {ln.split(/(\*[^*]+\*)/).map((bit, j) =>
                  bit.startsWith('*') && bit.endsWith('*')
                    ? <span key={j} style={{color: C.yellow}}>{bit.slice(1, -1)}</span>
                    : <span key={j}>{bit}</span>)}
              </div>
            ))}
          </div>
        </div>
      ) : null}
      <BrowserWindow {...b} />
    </AbsoluteFill>
  );
};
