// The three beats of lecture one that cannot be filmed on this machine:
// the Windows installer, the first-run sign-in, and the Windows open-folder
// dialog. Same deck language as the opening slides, so the cut does not
// announce the seam. Each runs exactly as long as its Kokoro line.

import React from 'react';
import {AbsoluteFill, Audio, Sequence, interpolate, staticFile, useCurrentFrame} from 'remotion';
import {loadFont as loadArchivo} from '@remotion/google-fonts/ArchivoBlack';
import {loadFont as loadPlex} from '@remotion/google-fonts/IBMPlexMono';

const {fontFamily: ARCHIVO} = loadArchivo();
const {fontFamily: PLEX} = loadPlex();

export const FPS = 30;
const C = {
  bg: '#070707', ink: '#f4f1ea', muted: '#8a867c',
  yellow: '#f5c518', red: '#e53935', blue: '#3b82f6',
};
const MONO = `${PLEX}, ui-monospace, monospace`;
const DISPLAY = `${ARCHIVO}, sans-serif`;

const bez = (p1x: number, p1y: number, p2x: number, p2y: number) => (t: number) => {
  if (t <= 0) return 0;
  if (t >= 1) return 1;
  let lo = 0, hi = 1, u = t;
  const bx = (v: number) => 3 * (1 - v) * (1 - v) * v * p1x + 3 * (1 - v) * v * v * p2x + v * v * v;
  const by = (v: number) => 3 * (1 - v) * (1 - v) * v * p1y + 3 * (1 - v) * v * v * p2y + v * v * v;
  for (let i = 0; i < 24; i++) {
    const x = bx(u);
    if (Math.abs(x - t) < 1e-4) break;
    if (x < t) lo = u; else hi = u;
    u = (lo + hi) / 2;
  }
  return by(u);
};
const easeEnter = bez(0.16, 0.84, 0.32, 1);
const easeSlam = bez(0.15, 1.4, 0.3, 1);
const prog = (f: number, delay: number, dur: number) =>
  Math.max(0, Math.min(1, (f / FPS - delay) / dur));

const Rise: React.FC<{d?: number; children: React.ReactNode; style?: React.CSSProperties}> =
({d = 0, children, style}) => {
  const f = useCurrentFrame();
  const p = easeEnter(prog(f, d, 0.6));
  return <div style={{opacity: p, transform: `translateY(${22 * (1 - p)}px)`, ...style}}>{children}</div>;
};

const Kicker: React.FC<{d?: number; children: React.ReactNode; style?: React.CSSProperties}> =
({d = 0, children, style}) => (
  <Rise d={d} style={{
    letterSpacing: '0.22em', textTransform: 'uppercase', fontSize: 15,
    color: C.muted, margin: '0 0 18px', ...style,
  }}>{children}</Rise>
);

const Display: React.FC<{size: number; children: React.ReactNode; style?: React.CSSProperties}> =
({size, children, style}) => (
  <div style={{
    fontFamily: DISPLAY, lineHeight: 0.9, letterSpacing: '-0.03em',
    textTransform: 'uppercase', fontSize: size, ...style,
  }}>{children}</div>
);

const Stamp: React.FC<{color?: string; delay?: number; children: React.ReactNode}> =
({color = C.yellow, delay = 1.05, children}) => {
  const f = useCurrentFrame();
  const p = easeSlam(prog(f, delay, 0.42));
  return (
    <span style={{
      fontFamily: DISPLAY, textTransform: 'uppercase', display: 'inline-block',
      padding: '10px 16px', border: `2px solid ${color}`, color, fontSize: 22,
      opacity: Math.min(1, p / 0.35),
      transform: `rotate(${interpolate(p, [0, 0.7, 1], [-24, -8, -6])}deg) `
               + `scale(${interpolate(p, [0, 0.7, 1], [1.6, 0.94, 1])})`,
    }}>{children}</span>
  );
};

/** a numbered step that slides in and, optionally, lights up */
const Step: React.FC<{n: string; d: number; hot?: boolean; children: React.ReactNode}> =
({n, d, hot, children}) => {
  const f = useCurrentFrame();
  const p = easeEnter(prog(f, d, 0.55));
  const glow = hot ? prog(f, d + 0.5, 0.5) : 0;
  return (
    <div style={{
      display: 'flex', alignItems: 'center', gap: 20, marginTop: 14,
      border: `1px solid ${hot ? C.yellow : '#3a3a3a'}`, borderRadius: 8,
      padding: '16px 20px', fontSize: 25,
      color: hot ? C.ink : '#b4b0a8',
      background: hot ? `rgba(245,197,24,${0.12 * glow})` : 'transparent',
      opacity: p, transform: `translateX(${-26 * (1 - p)}px)`,
    }}>
      <span style={{
        fontFamily: DISPLAY, fontSize: 20, minWidth: 34,
        color: hot ? C.yellow : C.muted,
      }}>{n}</span>
      <span>{children}</span>
    </div>
  );
};

const Slide: React.FC<{center?: boolean; children: React.ReactNode}> = ({center, children}) => {
  const f = useCurrentFrame();
  const ken = interpolate(easeEnter(prog(f, 0, 8)), [0, 1], [1.03, 1]);
  const flash = 0.5 * (1 - prog(f, 0, 0.28));
  return (
    <AbsoluteFill style={{background: C.bg}}>
      <AbsoluteFill style={{
        padding: '6.5% 7%', display: 'flex', flexDirection: 'column',
        transform: `scale(${ken})`, color: C.ink, fontFamily: MONO,
        alignItems: center ? 'center' : undefined,
        textAlign: center ? 'center' : undefined,
      }}>{children}</AbsoluteFill>
      <AbsoluteFill style={{background: '#fff', opacity: flash}} />
    </AbsoluteFill>
  );
};

const Grow = <div style={{flex: 1}} />;

// ------------------------------------------------------- D · on Windows ----

const D1 = () => (
  <Slide>
    <Kicker>and just for the pc people</Kicker>
    <Display size={100}>
      <Rise d={0.12}>On</Rise>
      <Rise d={0.38}><span style={{color: C.blue}}>Windows</span></Rise>
    </Display>
    {Grow}
    <div>
      <Step n="01" d={0.7}>Go to cursor.com in your browser</Step>
      <Step n="02" d={1.1}>Download for Windows is the default button</Step>
    </div>
  </Slide>
);

const D2 = () => (
  <Slide>
    <Kicker>once it has downloaded, open the file</Kicker>
    <Display size={64}><Rise d={0.12}>The installer</Rise></Display>
    {Grow}
    <div>
      <Step n="01" d={0.5}>Accept the agreement</Step>
      <Step n="02" d={0.9}>Press Next</Step>
      <Step n="03" d={1.3} hot>Keep the defaults — Add to PATH matters</Step>
      <Step n="04" d={1.9}>Next, then Install</Step>
    </div>
    <div style={{marginTop: 24}}><Stamp color={C.blue} delay={2.4}>installed on your pc</Stamp></div>
  </Slide>
);

// ---------------------------------------------------- F · sign in / up -----

const F1 = () => (
  <Slide center>
    <Kicker>the first time you come into cursor</Kicker>
    <Display size={92}>
      <Rise d={0.12}>Sign in</Rise>
      <Rise d={0.38}>or <span style={{color: C.yellow}}>sign up</span></Rise>
    </Display>
    {Grow}
    <div><Stamp color={C.yellow} delay={1.2}>free for a couple of weeks</Stamp></div>
  </Slide>
);

const F2 = () => (
  <Slide>
    <Kicker>not seeing that screen?</Kicker>
    <Display size={58}>
      <Rise d={0.12}>File</Rise>
      <Rise d={0.38}><span style={{color: C.muted}}>›</span> New Window</Rise>
    </Display>
    {Grow}
    <div>
      <Step n="01" d={0.7}>The welcome screen comes back</Step>
      <Step n="02" d={1.1} hot>Press Sign in</Step>
    </div>
  </Slide>
);

const F3 = () => (
  <Slide>
    <Kicker>it opens cursor in your browser</Kicker>
    <Display size={58}>
      <Rise d={0.12}>Don’t have</Rise>
      <Rise d={0.38}>an account? <span style={{color: C.yellow}}>Sign up.</span></Rise>
    </Display>
    {Grow}
    <div>
      <Step n="01" d={0.6}>Continue to sign in</Step>
      <Step n="02" d={1.0}>Enter your details, answer a few questions</Step>
      <Step n="03" d={1.5}>Come back to Cursor and sign in</Step>
    </div>
    <div style={{marginTop: 22}}>
      <Stamp color={C.blue} delay={2.1}>already have one? so much the better</Stamp>
    </div>
  </Slide>
);

// ------------------------------------------- I · the folder, on Windows ----

const I1 = () => (
  <Slide>
    <Kicker>now the same thing on a pc</Kicker>
    <Display size={72}>
      <Rise d={0.12}>Open</Rise>
      <Rise d={0.38}><span style={{color: C.blue}}>project</span></Rise>
    </Display>
    {Grow}
    <div>
      <Step n="01" d={0.7}>If it says Sign in, click there first</Step>
      <Step n="02" d={1.1}>Otherwise press Open Project</Step>
      <Step n="03" d={1.5}>The Windows file browser comes up, in your Home folder</Step>
    </div>
  </Slide>
);

const I2 = () => (
  <Slide>
    <Kicker>into projects, then a new folder</Kicker>
    <Display size={92}>
      <Rise d={0.12}>Call it</Rise>
      <Rise d={0.38}><span style={{color: C.yellow}}>Instant</span></Rise>
    </Display>
    {Grow}
    <div>
      <Step n="01" d={0.6}>Go into Projects</Step>
      <Step n="02" d={1.0}>New Folder, name it Instant</Step>
      <Step n="03" d={1.4}>Go into Instant, then Select Folder</Step>
      <Step n="04" d={1.9} hot>INSTANT appears on the top left</Step>
    </div>
  </Slide>
);

// --------------------------------------------------------------- wiring ----

const build = (parts: React.FC[], totalFrames: number, vo: string) => {
  const each = Math.floor(totalFrames / parts.length);
  return () => (
    <AbsoluteFill style={{background: C.bg}}>
      {parts.map((P, i) => (
        <Sequence key={i} from={i * each}
                  durationInFrames={i === parts.length - 1 ? totalFrames - i * each : each}>
          <P />
        </Sequence>
      ))}
      <Audio src={staticFile(vo)} />
    </AbsoluteFill>
  );
};

export const D_FRAMES = Math.round(30.898 * FPS);
export const F_FRAMES = Math.round(60.387 * FPS);
export const I_FRAMES = Math.round(43.674 * FPS);

export const InsertD = build([D1, D2], D_FRAMES, 'nocode01/vo/D_windows.mp3');
export const InsertF = build([F1, F2, F3], F_FRAMES, 'nocode01/vo/F_signin.mp3');
export const InsertI = build([I1, I2], I_FRAMES, 'nocode01/vo/I_winfolder.mp3');
