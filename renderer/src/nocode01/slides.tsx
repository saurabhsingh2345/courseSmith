// Segment A of nocode01 — the opening nine slides, ported from the No-Code deck.
//
// The deck was authored as HTML with CSS keyframes. CSS animation does not step
// deterministically under a frame-by-frame renderer, so every entrance here is
// driven off useCurrentFrame() instead. The look is unchanged: same tokens, same
// type, same delays, same slam and fill curves.

import React from 'react';
import {
  AbsoluteFill, Audio, Img, Sequence, interpolate, staticFile, useCurrentFrame,
} from 'remotion';
import {loadFont as loadArchivo} from '@remotion/google-fonts/ArchivoBlack';
import {loadFont as loadPlex} from '@remotion/google-fonts/IBMPlexMono';

const {fontFamily: ARCHIVO} = loadArchivo();
const {fontFamily: PLEX} = loadPlex();

export const FPS = 30;
export const GAP = 0.28;                       // silence the deck leaves between slides

// Kokoro durations, in order. Kept here so the cut is readable in one place.
export const VO = [6.775, 8.661, 12.459, 11.947, 14.472, 6.048, 10.173, 11.785, 9.341];

export const SLIDE_FRAMES = VO.map((d, i) =>
  Math.round((d + (i < VO.length - 1 ? GAP : 0)) * FPS));
export const TOTAL_FRAMES = SLIDE_FRAMES.reduce((a, b) => a + b, 0);

const C = {
  bg: '#070707',
  ink: '#f4f1ea',
  muted: '#8a867c',
  yellow: '#f5c518',
  red: '#e53935',
  blue: '#3b82f6',
};

const MONO = `${PLEX}, ui-monospace, monospace`;
const DISPLAY = `${ARCHIVO}, sans-serif`;

// ---------------------------------------------------------------- easing ---
// cubic-bezier(.16,.84,.32,1) — the deck's entrance curve, sampled.
const bez = (p1x: number, p1y: number, p2x: number, p2y: number) => (t: number) => {
  if (t <= 0) return 0;
  if (t >= 1) return 1;
  let lo = 0, hi = 1, u = t;
  const bx = (u: number) =>
    3 * (1 - u) * (1 - u) * u * p1x + 3 * (1 - u) * u * u * p2x + u * u * u;
  const by = (u: number) =>
    3 * (1 - u) * (1 - u) * u * p1y + 3 * (1 - u) * u * u * p2y + u * u * u;
  for (let i = 0; i < 24; i++) {
    const x = bx(u);
    if (Math.abs(x - t) < 1e-4) break;
    if (x < t) lo = u; else hi = u;
    u = (lo + hi) / 2;
  }
  return by(u);
};

const easeEnter = bez(0.16, 0.84, 0.32, 1);
const easeFill = bez(0.2, 0.8, 0.2, 1);
const easeGrow = bez(0.16, 0.9, 0.24, 1);
const easeSlam = bez(0.15, 1.4, 0.3, 1);

/** progress 0..1 for an animation starting at `delay` s running `dur` s */
const prog = (frame: number, delay: number, dur: number) =>
  Math.max(0, Math.min(1, (frame / FPS - delay) / dur));

// ------------------------------------------------------------- primitives --

const Rise: React.FC<{d?: number; children: React.ReactNode; style?: React.CSSProperties}> =
({d = 0, children, style}) => {
  const f = useCurrentFrame();
  const p = easeEnter(prog(f, d, 0.6));
  return (
    <div style={{opacity: p, transform: `translateY(${22 * (1 - p)}px)`, ...style}}>
      {children}
    </div>
  );
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
    fontFamily: DISPLAY, fontWeight: 400, lineHeight: 0.9, letterSpacing: '-0.03em',
    textTransform: 'uppercase', margin: 0, fontSize: size, ...style,
  }}>{children}</div>
);

const Rule: React.FC = () => {
  const f = useCurrentFrame();
  const p = prog(f, 0.7, 0.55);
  return <div style={{height: 3, width: `${42 * p}%`, background: C.red, margin: '18px 0 0'}} />;
};

const Stamp: React.FC<{color?: string; children: React.ReactNode}> = ({color = C.yellow, children}) => {
  const f = useCurrentFrame();
  const p = easeSlam(prog(f, 1.05, 0.42));
  const rot = interpolate(p, [0, 0.7, 1], [-24, -8, -6]);
  const sc = interpolate(p, [0, 0.7, 1], [1.6, 0.94, 1]);
  return (
    <span style={{
      fontFamily: DISPLAY, textTransform: 'uppercase', display: 'inline-block',
      padding: '10px 16px', border: `2px solid ${color}`, color,
      opacity: p === 0 ? 0 : Math.min(1, p / 0.35),
      transform: `rotate(${rot}deg) scale(${sc})`, fontSize: 22,
    }}>{children}</span>
  );
};

const Badge: React.FC<{children: React.ReactNode}> = ({children}) => {
  const f = useCurrentFrame();
  const p = easeSlam(prog(f, 1.15, 0.5));
  return (
    <span style={{
      fontFamily: DISPLAY, textTransform: 'uppercase', display: 'inline-block',
      background: C.yellow, color: '#111', border: '2px solid #fff',
      boxShadow: `8px 8px 0 ${C.red}`, padding: '12px 18px', lineHeight: 0.95,
      fontSize: 30, opacity: p,
      transform: `rotate(${interpolate(p, [0, 1], [18, 4])}deg) translateY(${40 * (1 - p)}px) scale(${interpolate(p, [0, 1], [0.7, 1])})`,
    }}>{children}</span>
  );
};

/** the whole slide drifts from 1.03 to 1 over eight seconds, as the deck did */
const Slide: React.FC<{center?: boolean; shake?: boolean; children: React.ReactNode}> =
({center, shake, children}) => {
  const f = useCurrentFrame();
  const ken = interpolate(easeEnter(prog(f, 0, 8)), [0, 1], [1.03, 1]);
  let dx = 0;
  if (shake) {
    const p = prog(f, 1.05, 0.35);
    if (p > 0 && p < 1) dx = Math.sin(p * Math.PI * 2) * -6;
  }
  const flash = 0.55 * (1 - prog(f, 0, 0.28));
  return (
    <AbsoluteFill style={{background: C.bg}}>
      <AbsoluteFill style={{
        padding: '6.5% 7%', display: 'flex', flexDirection: 'column',
        transform: `scale(${ken}) translateX(${dx}px)`,
        alignItems: center ? 'center' : undefined,
        textAlign: center ? 'center' : undefined,
        color: C.ink, fontFamily: MONO,
      }}>{children}</AbsoluteFill>
      <AbsoluteFill style={{background: '#fff', opacity: flash, pointerEvents: 'none'}} />
    </AbsoluteFill>
  );
};

const Grow = <div style={{flex: 1}} />;

// ----------------------------------------------------------------- slides --

const S01 = () => (
  <Slide>
    <Kicker>01 · the missing manual</Kicker>
    <Display size={128}>
      <Rise d={0.12}>AI</Rise>
      <Rise d={0.38}><span style={{color: C.red}}>CODER</span></Rise>
      <Rise d={0.64}>PROGRAM</Rise>
    </Display>
    <Rule />
    {Grow}
    <div><Badge>3 weeks<br /><span style={{fontSize: '0.55em'}}>the ride starts now</span></Badge></div>
  </Slide>
);

const GridA: React.FC = () => {
  const f = useCurrentFrame();
  const lit = Math.max(0, Math.floor((f / FPS - 0.22) / 0.012));
  return (
    <div style={{display: 'grid', gridTemplateColumns: 'repeat(18, 14px)', gap: 5}}>
      {Array.from({length: 72}).map((_, i) => {
        const on = i < lit;
        return <div key={i} style={{
          width: 14, height: 14, borderRadius: 2,
          background: on ? C.yellow : 'rgba(255,255,255,0.08)',
          opacity: on ? 1 : 0.25, transform: on ? 'none' : 'scale(0.4)',
        }} />;
      })}
    </div>
  );
};

const S02 = () => (
  <Slide>
    <div style={{display: 'flex', alignItems: 'center', gap: '4%', height: '100%'}}>
      <GridA />
      <div>
        <Kicker>over the three-week program</Kicker>
        <Display size={90}>
          <Rise d={0.12}>Expert at</Rise>
          <Rise d={0.38}><span style={{color: C.yellow}}>agentic</span></Rise>
          <Rise d={0.64}>engineering</Rise>
        </Display>
        <Kicker d={1.2} style={{marginTop: 22, color: '#aaa'}}>an extraordinary journey</Kicker>
      </div>
    </div>
  </Slide>
);

const S03 = () => (
  <Slide center shake>
    <Kicker>a word of caution</Kicker>
    <Display size={128}>
      <Rise d={0.12}>You need to</Rise>
      <Rise d={0.38}><span style={{color: C.red}}>buckle up</span></Rise>
    </Display>
    {Grow}
    <div><Stamp color={C.yellow}>roller coaster</Stamp></div>
  </Slide>
);

const XRow: React.FC<{d: number; children: React.ReactNode}> = ({d, children}) => (
  <Rise d={d} style={{
    display: 'flex', alignItems: 'center', gap: 14, border: '1px solid #3a3a3a',
    borderRadius: 8, padding: '12px 16px', color: '#9a9a9a', marginTop: 10, fontSize: 22,
  }}>{children}</Rise>
);

const S04 = () => (
  <Slide>
    <Kicker>standard fare with these kinds of programs</Kicker>
    <Display size={56} style={{maxWidth: '16ch'}}>
      <Rise d={0.12}>Begin with objectives.</Rise>
      <Rise d={0.64}>I don’t.</Rise>
    </Display>
    {Grow}
    <div>
      <Kicker d={0.38} style={{marginBottom: 8}}>many programs start with</Kicker>
      <XRow d={0.64}><span style={{color: C.red}}>×</span> Introduce myself</XRow>
      <XRow d={0.9}><span style={{color: C.red}}>×</span> Curriculum and logistics</XRow>
      <XRow d={1.2}><span style={{color: C.red}}>×</span> The usual welcome tour</XRow>
    </div>
    <div style={{marginTop: 22}}><Stamp color={C.red}>except I don’t</Stamp></div>
  </Slide>
);

const Meter: React.FC = () => {
  const f = useCurrentFrame();
  const w = 78 * easeFill(prog(f, 0.7, 1.1));
  return (
    <div style={{height: 60, border: '2px solid #fff', display: 'flex', overflow: 'hidden'}}>
      <div style={{
        width: `${w}%`, background: C.yellow, color: '#111', display: 'flex',
        alignItems: 'center', padding: '0 16px', fontWeight: 700,
        letterSpacing: '0.08em', textTransform: 'uppercase', whiteSpace: 'nowrap', fontSize: 20,
      }}>build first</div>
      <div style={{
        flex: 1, background: '#1b1b1b', display: 'flex', alignItems: 'center',
        justifyContent: 'flex-end', padding: '0 16px', color: '#8a8a8a',
        letterSpacing: '0.08em', textTransform: 'uppercase', whiteSpace: 'nowrap', fontSize: 20,
      }}>talk later</div>
    </div>
  );
};

const S05 = () => (
  <Slide>
    <Kicker>if you have been on any of our programs</Kicker>
    <Display size={90}>
      <Rise d={0.12}>Instant</Rise>
      <Rise d={0.38}><span style={{color: C.yellow}}>gratification</span></Rise>
    </Display>
    <Kicker d={0.64} style={{margin: '18px 0 28px', color: '#c8c4ba'}}>
      Roll up sleeves. Build something together.
    </Kicker>
    <Meter />
  </Slide>
);

const Col: React.FC<{h: string; bg: string; shadow: string; delay: number; label: string; dark?: boolean}> =
({h, bg, shadow, delay, label, dark}) => {
  const f = useCurrentFrame();
  const p = easeGrow(prog(f, delay, 0.7));
  return (
    <div style={{
      flex: 1, height: h, border: '2px solid #fff', display: 'flex',
      alignItems: 'flex-end', justifyContent: 'center', paddingBottom: 12,
      fontFamily: DISPLAY, color: dark ? '#111' : '#fff', background: bg,
      boxShadow: `7px 7px 0 ${shadow}`, transform: `scaleY(${p})`, transformOrigin: 'bottom',
      fontSize: 26,
    }}>{label}</div>
  );
};

const S06 = () => (
  <Slide>
    <Kicker>next three weeks</Kicker>
    <Display size={90}>
      <Rise d={0.12}>AI products</Rise>
      <Rise d={0.38}>that <span style={{color: C.blue}}>generate</span></Rise>
      <Rise d={0.64}>code</Rise>
    </Display>
    {Grow}
    <div style={{display: 'flex', alignItems: 'flex-end', gap: 18, height: '42%'}}>
      <Col h="46%" bg={C.red} shadow={C.blue} delay={0} label="01" />
      <Col h="68%" bg={C.blue} shadow={C.yellow} delay={0.18} label="02" />
      <Col h="88%" bg={C.yellow} shadow={C.red} delay={0.36} label="03" dark />
    </div>
  </Slide>
);

const S07 = () => (
  <Slide center>
    <Kicker>but for today</Kicker>
    <Rise d={0.12}><Display size={90}>We start with</Display></Rise>
    <Rise d={0.38} style={{width: '100%'}}>
      <Img src={staticFile('nocode01/img/cursor-lockup.png')}
           style={{width: 'min(72%, 1000px)', margin: '28px auto 0', display: 'block'}} />
    </Rise>
    <Kicker d={0.9} style={{marginTop: 22, color: '#c8c4ba'}}>
      Super popular. Easy to use. Free trial.
    </Kicker>
  </Slide>
);

const S08 = () => (
  <Slide>
    <div style={{display: 'flex', alignItems: 'center', gap: '4%', height: '100%'}}>
      <Rise d={0.12}>
        <Img src={staticFile('nocode01/img/cursor-icon.png')}
             style={{width: 340, borderRadius: '22%', flexShrink: 0, display: 'block'}} />
      </Rise>
      <div>
        <Kicker>already used the free trial</Kicker>
        <Display size={56}>
          <Rise d={0.12}>Watch.</Rise>
          <Rise d={0.38}>Or follow along</Rise>
          <Rise d={0.64}><span style={{color: C.yellow}}>your way.</span></Rise>
        </Display>
        <Kicker d={0.9} style={{marginTop: 22, color: '#aaa'}}>
          You do not need a fresh Cursor account.
        </Kicker>
      </div>
    </div>
  </Slide>
);

const S09 = () => (
  <Slide>
    <Kicker>then we get on with the program</Kicker>
    <Display size={90}>
      <Rise d={0.12}>Something</Rise>
      <Rise d={0.38}><span style={{color: C.yellow}}>super simple</span></Rise>
    </Display>
    <Kicker d={0.64} style={{margin: '18px 0 0', color: '#c8c4ba'}}>
      Use AI to generate code. Right away.
    </Kicker>
    {Grow}
    <div><Badge>Let’s go do it</Badge></div>
  </Slide>
);

const SLIDES = [S01, S02, S03, S04, S05, S06, S07, S08, S09];
const ENTER = ['hit', 'whoosh', 'hit', 'whoosh', 'whoosh', 'rise', 'hit', 'whoosh', 'rise'];

export const Nocode01Slides: React.FC = () => {
  let at = 0;
  const out: React.ReactNode[] = [];
  SLIDES.forEach((S, i) => {
    const len = SLIDE_FRAMES[i];
    out.push(
      <Sequence key={`s${i}`} from={at} durationInFrames={len}>
        <S />
        <Audio src={staticFile(`nocode01/vo/0${i + 1}.mp3`)} />
        <Audio src={staticFile(`nocode01/vo/sfx-${ENTER[i]}.mp3`)} volume={0.38} />
      </Sequence>,
    );
    at += len;
  });
  return <AbsoluteFill style={{background: C.bg}}>{out}</AbsoluteFill>;
};
