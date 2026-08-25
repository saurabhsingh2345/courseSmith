// The graphic scenes.
//
// Every scene here composes out of kit.tsx, so a slide and a shot of the app
// read as the same film. The animation vocabulary is deliberately not the
// reference cut's: where that one leaned on cards sliding up and a sweep
// highlighting the phrase being spoken, this one leans on things being *drawn* —
// rules that extend, connectors that grow between labelled parts, rings that
// close, counters that run. A lesson about assembling something should look like
// assembly.

import React from 'react';
import {AbsoluteFill, interpolate, useCurrentFrame} from 'remotion';
import {
  C,
  GRAD,
  GRAD_WIDE,
  grad,
  DISPLAY,
  SERIF,
  BODY,
  MONO,
  Eyebrow,
  Headline,
  Body,
  Rule,
  Glass,
  Halo,
  Glyph,
  GlyphName,
  Motes,
  ramp,
  lin,
  useEnter,
  FPS,
} from './kit';

const Pad: React.FC<{children: React.ReactNode; style?: React.CSSProperties}> = ({
  children,
  style,
}) => (
  <AbsoluteFill
    style={{padding: '92px 128px', display: 'flex', flexDirection: 'column', ...style}}
  >
    {children}
  </AbsoluteFill>
);

// ---------------------------------------------------------------- title

export const Title: React.FC<{title: string; kicker: string; sub: string}> = ({
  title,
  kicker,
  sub,
}) => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill>
      <Motes n={40} seed="title" />
      <Pad style={{justifyContent: 'center', gap: 30}}>
        <Eyebrow delay={2}>{kicker}</Eyebrow>
        <Headline
          text={title}
          size={132}
          delay={10}
          serif={['without']}
          accent={['code']}
          width={1500}
        />
        <div style={{marginTop: 6}}>
          <Rule delay={30} width={220} />
        </div>
        <Body delay={38} size={36} width={1080}>
          {sub}
        </Body>
      </Pad>
      {/* a slow horizon line that draws across under the title */}
      <div
        style={{
          position: 'absolute',
          left: 0,
          bottom: 0,
          height: 3,
          width: `${ramp(frame, 20, 90) * 100}%`,
          backgroundImage: GRAD_WIDE,
          opacity: 0.5,
        }}
      />
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------- chapter

export const Chapter: React.FC<{n: string; title: string}> = ({n, title}) => {
  const frame = useCurrentFrame();
  const a = ramp(frame, 4, 22);
  const ring = ramp(frame, 6, 34);
  const R = 96;
  const circ = 2 * Math.PI * R;
  return (
    <AbsoluteFill>
      <Motes n={26} seed={`ch${n}`} />
      <AbsoluteFill
        style={{alignItems: 'center', justifyContent: 'center', flexDirection: 'row', gap: 62}}
      >
        {/* the number, with a ring that closes around it */}
        <div style={{position: 'relative', width: 232, height: 232, flex: 'none'}}>
          <Halo size={300} color={C.violet} opacity={0.26} />
          <svg width={232} height={232} style={{position: 'absolute', inset: 0}}>
            <defs>
              <linearGradient id={`g${n}`} x1="0" y1="0" x2="1" y2="1">
                <stop offset="0%" stopColor={C.rose} />
                <stop offset="100%" stopColor={C.violet} />
              </linearGradient>
            </defs>
            <circle
              cx={116}
              cy={116}
              r={R}
              fill="none"
              stroke="rgba(255,255,255,0.08)"
              strokeWidth={3}
            />
            <circle
              cx={116}
              cy={116}
              r={R}
              fill="none"
              stroke={`url(#g${n})`}
              strokeWidth={4}
              strokeLinecap="round"
              strokeDasharray={circ}
              strokeDashoffset={circ * (1 - ring)}
              transform="rotate(-90 116 116)"
            />
          </svg>
          <div
            style={{
              position: 'absolute',
              inset: 0,
              display: 'grid',
              placeItems: 'center',
              fontFamily: DISPLAY,
              fontWeight: 800,
              fontSize: 92,
              letterSpacing: '-0.04em',
              ...grad,
              opacity: a,
            }}
          >
            {n}
          </div>
        </div>
        <div style={{maxWidth: 1020}}>
          <Headline text={title} size={82} delay={12} width={1020} />
        </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------- statement

export const Statement: React.FC<{
  kicker?: string;
  text: string;
  accent?: string[];
  serif?: string[];
  body?: string;
  size?: number;
}> = ({kicker, text, accent, serif, body, size = 96}) => (
  <Pad style={{justifyContent: 'center', gap: 30}}>
    {kicker ? <Eyebrow delay={0}>{kicker}</Eyebrow> : null}
    <Headline
      text={text}
      size={size}
      delay={kicker ? 8 : 2}
      accent={accent}
      serif={serif}
      width={1560}
    />
    {body ? (
      <>
        <div style={{marginTop: 8}}>
          <Rule delay={26} width={150} />
        </div>
        <Body delay={32} width={1180}>
          {body}
        </Body>
      </>
    ) : null}
  </Pad>
);

// ---------------------------------------------------------------- before / after

/** Two panels: the old way, then the way this lesson is about. */
export const BeforeAfter: React.FC<{
  title: string;
  left: {label: string; items: string[]};
  right: {label: string; items: string[]};
}> = ({title, left, right}) => {
  const frame = useCurrentFrame();
  const panel = (
    side: {label: string; items: string[]},
    delay: number,
    hot: boolean
  ) => (
    <Glass delay={delay} active={hot} style={{flex: 1, padding: '40px 44px'}}>
      <div
        style={{
          fontFamily: MONO,
          fontSize: 20,
          letterSpacing: '0.2em',
          textTransform: 'uppercase',
          color: hot ? C.rose : C.dim,
          marginBottom: 26,
        }}
      >
        {side.label}
      </div>
      {side.items.map((it, i) => {
        const a = ramp(frame, delay + 12 + i * 7, 16);
        return (
          <div
            key={it}
            style={{
              display: 'flex',
              gap: 18,
              alignItems: 'flex-start',
              padding: '13px 0',
              borderTop: i ? `1px solid ${C.line2}` : 'none',
              opacity: a,
              transform: `translateX(${(1 - a) * (hot ? 16 : -16)}px)`,
            }}
          >
            <div
              style={{
                marginTop: 12,
                width: 8,
                height: 8,
                borderRadius: hot ? '50%' : 2,
                backgroundImage: hot ? GRAD : 'none',
                backgroundColor: hot ? undefined : C.dim,
                flex: 'none',
              }}
            />
            <div
              style={{
                fontFamily: BODY,
                fontWeight: 300,
                fontSize: 31,
                lineHeight: 1.4,
                color: hot ? C.text : C.muted,
                textDecoration: hot ? 'none' : 'none',
              }}
            >
              {it}
            </div>
          </div>
        );
      })}
    </Glass>
  );

  return (
    <Pad style={{justifyContent: 'center', gap: 46}}>
      <Headline text={title} size={68} delay={0} width={1500} />
      <div style={{display: 'flex', gap: 34, alignItems: 'stretch'}}>
        {panel(left, 14, false)}
        <div
          style={{
            width: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flex: 'none',
          }}
        >
          <svg width={52} height={52} viewBox="0 0 52 52" style={{opacity: ramp(frame, 40, 16)}}>
            <path
              d="M8 26 H40 M30 15 L41 26 L30 37"
              fill="none"
              stroke={C.rose}
              strokeWidth={3.4}
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </div>
        {panel(right, 44, true)}
      </div>
    </Pad>
  );
};

// ---------------------------------------------------------------- numbers

export const Numbers: React.FC<{
  title: string;
  stats: {value: string; label: string; note?: string}[];
}> = ({title, stats}) => {
  const frame = useCurrentFrame();
  return (
    <Pad style={{justifyContent: 'center', gap: 56}}>
      <Headline text={title} size={68} delay={0} width={1440} />
      <div style={{display: 'flex', gap: 30}}>
        {stats.map((st, i) => {
          const d = 16 + i * 10;
          const a = ramp(frame, d, 20);
          return (
            <Glass key={st.label} delay={d} style={{flex: 1, padding: '38px 36px 34px'}}>
              <div
                style={{
                  fontFamily: DISPLAY,
                  fontWeight: 800,
                  fontSize: 94,
                  letterSpacing: '-0.045em',
                  lineHeight: 1,
                  ...grad,
                  opacity: a,
                }}
              >
                {st.value}
              </div>
              <div
                style={{
                  marginTop: 18,
                  fontFamily: BODY,
                  fontWeight: 500,
                  fontSize: 31,
                  color: C.text,
                  opacity: ramp(frame, d + 8, 16),
                }}
              >
                {st.label}
              </div>
              {st.note ? (
                <div
                  style={{
                    marginTop: 8,
                    fontFamily: BODY,
                    fontWeight: 300,
                    fontSize: 25,
                    lineHeight: 1.4,
                    color: C.muted,
                    opacity: ramp(frame, d + 14, 16),
                  }}
                >
                  {st.note}
                </div>
              ) : null}
            </Glass>
          );
        })}
      </div>
    </Pad>
  );
};

// ---------------------------------------------------------------- checklist

export const Checklist: React.FC<{
  title: string;
  items: {name: string; note: string}[];
  sweep?: number[];
}> = ({title, items, sweep}) => {
  const frame = useCurrentFrame();
  return (
    <Pad style={{justifyContent: 'center', gap: 44}}>
      <Headline text={title} size={70} delay={0} width={1440} />
      <div style={{display: 'flex', flexDirection: 'column', gap: 16}}>
        {items.map((it, i) => {
          const d = 14 + i * 9;
          // The step ticks over when the narration reaches its phrase.
          const hitAt = sweep?.[i];
          const done = hitAt !== undefined ? frame >= hitAt : frame >= d + 40;
          const tick = hitAt !== undefined ? ramp(frame, hitAt, 14) : ramp(frame, d + 40, 14);
          const nextAt = sweep?.[i + 1];
          const current = done && (nextAt === undefined || frame < nextAt);
          return (
            <Glass key={it.name} delay={d} active={current} style={{padding: '22px 32px'}}>
              <div style={{display: 'flex', alignItems: 'center', gap: 26}}>
                <div
                  style={{
                    width: 44,
                    height: 44,
                    borderRadius: 12,
                    flex: 'none',
                    border: `2px solid ${done ? 'transparent' : C.line}`,
                    backgroundImage: done ? GRAD : 'none',
                    display: 'grid',
                    placeItems: 'center',
                  }}
                >
                  {done ? (
                    <svg width={22} height={22} viewBox="0 0 24 24">
                      <polyline
                        points="4,12.5 9.5,18 20,6.5"
                        fill="none"
                        stroke="#140A12"
                        strokeWidth={3}
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeDasharray={30}
                        strokeDashoffset={30 * (1 - tick)}
                      />
                    </svg>
                  ) : (
                    <span
                      style={{
                        fontFamily: MONO,
                        fontSize: 19,
                        color: C.dim,
                      }}
                    >
                      {i + 1}
                    </span>
                  )}
                </div>
                <div
                  style={{
                    fontFamily: DISPLAY,
                    fontWeight: 600,
                    fontSize: 38,
                    letterSpacing: '-0.02em',
                    color: C.text,
                    minWidth: 380,
                  }}
                >
                  {it.name}
                </div>
                <div
                  style={{
                    fontFamily: BODY,
                    fontWeight: 300,
                    fontSize: 28,
                    color: C.muted,
                    flex: 1,
                  }}
                >
                  {it.note}
                </div>
              </div>
            </Glass>
          );
        })}
      </div>
    </Pad>
  );
};

// ---------------------------------------------------------------- define

export const Define: React.FC<{
  word: string;
  pos: string;
  meaning: string;
  also: string;
}> = ({word, pos, meaning, also}) => {
  const frame = useCurrentFrame();
  return (
    <Pad style={{justifyContent: 'center'}}>
      <Glass delay={0} style={{padding: '62px 72px', maxWidth: 1500}}>
        <div style={{display: 'flex', alignItems: 'baseline', gap: 26}}>
          <div
            style={{
              fontFamily: DISPLAY,
              fontWeight: 800,
              fontSize: 116,
              letterSpacing: '-0.045em',
              color: C.text,
              opacity: ramp(frame, 4, 18),
            }}
          >
            {word}
          </div>
          <div
            style={{
              fontFamily: SERIF,
              fontStyle: 'italic',
              fontSize: 46,
              color: C.rose,
              opacity: ramp(frame, 14, 18),
            }}
          >
            {pos}
          </div>
        </div>
        <div style={{margin: '26px 0 30px'}}>
          <Rule delay={20} width={200} height={3} />
        </div>
        <div
          style={{
            fontFamily: BODY,
            fontWeight: 300,
            fontSize: 42,
            lineHeight: 1.42,
            color: C.text,
            opacity: ramp(frame, 24, 20),
          }}
        >
          {meaning}
        </div>
        <div
          style={{
            marginTop: 30,
            paddingTop: 26,
            borderTop: `1px solid ${C.line2}`,
            fontFamily: BODY,
            fontWeight: 300,
            fontSize: 30,
            lineHeight: 1.45,
            color: C.muted,
            opacity: ramp(frame, 40, 20),
          }}
        >
          <span style={{fontFamily: MONO, fontSize: 20, letterSpacing: '0.16em', color: C.dim}}>
            IN PRACTICE&nbsp;&nbsp;
          </span>
          {also}
        </div>
      </Glass>
    </Pad>
  );
};

// ---------------------------------------------------------------- anatomy

/**
 * A prompt taken apart.
 *
 * The chunks arrive in order and a connector grows from each one down to its
 * label, so the sentence is visibly *made of* the four things being named. This
 * is the scene that has to carry the single most useful idea in the lesson, so
 * it gets the most drawing.
 */
export const Anatomy: React.FC<{
  title: string;
  chunks: {text: string; tag: string}[];
  sweep?: number[];
}> = ({title, chunks, sweep}) => {
  const frame = useCurrentFrame();
  return (
    <Pad style={{justifyContent: 'center', gap: 58}}>
      <Headline text={title} size={64} delay={0} width={1400} />
      <div>
        <div style={{display: 'flex', flexWrap: 'wrap', gap: '18px 14px', alignItems: 'flex-end'}}>
          {chunks.map((ch, i) => {
            const d = 14 + i * 12;
            const a = ramp(frame, d, 18);
            const hot = sweep?.[i] !== undefined ? frame >= (sweep[i] as number) : frame >= d + 30;
            const line = ramp(frame, d + 10, 16);
            return (
              <div key={ch.tag} style={{opacity: a, transform: `translateY(${(1 - a) * 18}px)`}}>
                <div
                  style={{
                    fontFamily: BODY,
                    fontWeight: 400,
                    fontSize: 46,
                    lineHeight: 1.28,
                    color: hot ? C.text : C.muted,
                    padding: '10px 16px',
                    borderRadius: 12,
                    background: hot ? 'rgba(255,77,109,0.10)' : 'transparent',
                    border: `1px solid ${hot ? 'rgba(255,120,150,0.34)' : 'transparent'}`,
                  }}
                >
                  {ch.text}
                </div>
                {/* connector down to the tag */}
                <div style={{display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
                  <div
                    style={{
                      width: 2,
                      height: 22 * line,
                      backgroundColor: hot ? C.rose : C.line,
                    }}
                  />
                  <div
                    style={{
                      fontFamily: MONO,
                      fontSize: 19,
                      letterSpacing: '0.14em',
                      textTransform: 'uppercase',
                      color: hot ? C.rose : C.dim,
                      opacity: line,
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {ch.tag}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </Pad>
  );
};

// ---------------------------------------------------------------- compare

export const Compare: React.FC<{
  title: string;
  good: {label: string; items: string[]};
  bad: {label: string; items: string[]};
}> = ({title, good, bad}) => {
  const frame = useCurrentFrame();
  const col = (
    side: {label: string; items: string[]},
    ok: boolean,
    delay: number
  ) => (
    <div style={{flex: 1}}>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 14,
          marginBottom: 22,
          opacity: ramp(frame, delay, 14),
        }}
      >
        <div
          style={{
            width: 34,
            height: 34,
            borderRadius: 10,
            display: 'grid',
            placeItems: 'center',
            backgroundColor: ok ? 'rgba(52,211,153,0.16)' : 'rgba(255,77,109,0.14)',
          }}
        >
          <svg width={19} height={19} viewBox="0 0 24 24" fill="none"
               stroke={ok ? C.ok : C.rose} strokeWidth={3} strokeLinecap="round">
            {ok ? <polyline points="4,12.5 9.5,18 20,6.5" /> : <path d="M6 6 L18 18 M18 6 L6 18" />}
          </svg>
        </div>
        <div
          style={{
            fontFamily: MONO,
            fontSize: 21,
            letterSpacing: '0.16em',
            textTransform: 'uppercase',
            color: ok ? C.ok : C.rose,
          }}
        >
          {side.label}
        </div>
      </div>
      {side.items.map((it, i) => {
        const a = ramp(frame, delay + 12 + i * 8, 16);
        return (
          <div
            key={it}
            style={{
              fontFamily: BODY,
              fontWeight: 300,
              fontSize: 32,
              lineHeight: 1.42,
              color: ok ? C.text : C.muted,
              padding: '15px 0',
              borderTop: `1px solid ${C.line2}`,
              opacity: a,
              transform: `translateY(${(1 - a) * 12}px)`,
            }}
          >
            {it}
          </div>
        );
      })}
    </div>
  );
  return (
    <Pad style={{justifyContent: 'center', gap: 48}}>
      <Headline text={title} size={66} delay={0} width={1400} />
      <div style={{display: 'flex', gap: 78}}>
        {col(bad, false, 14)}
        <div style={{width: 1, backgroundColor: C.line, opacity: ramp(frame, 20, 20)}} />
        {col(good, true, 34)}
      </div>
    </Pad>
  );
};

// ---------------------------------------------------------------- ladder

/** A scope ladder: how big a thing to ask for, and which rung this lesson is on. */
export const Ladder: React.FC<{
  title: string;
  rungs: {name: string; note: string}[];
  pick: number;
}> = ({title, rungs, pick}) => {
  const frame = useCurrentFrame();
  const n = rungs.length;
  return (
    <Pad style={{justifyContent: 'center', gap: 44}}>
      <Headline text={title} size={66} delay={0} width={1400} />
      <div style={{display: 'flex', flexDirection: 'column-reverse', gap: 14}}>
        {rungs.map((r, i) => {
          const d = 14 + (n - 1 - i) * 9;
          const a = ramp(frame, d, 18);
          const on = i === pick;
          const w = 46 + ((i + 1) / n) * 54; // each rung is wider than the last
          return (
            <div
              key={r.name}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 28,
                opacity: a,
                transform: `translateX(${(1 - a) * -20}px)`,
              }}
            >
              <div
                style={{
                  width: `${w}%`,
                  padding: '20px 30px',
                  borderRadius: 14,
                  background: on
                    ? 'linear-gradient(100deg, rgba(255,77,109,0.20), rgba(139,92,246,0.14))'
                    : C.ink1,
                  border: `1px solid ${on ? 'rgba(255,120,150,0.44)' : C.line}`,
                  display: 'flex',
                  alignItems: 'baseline',
                  gap: 22,
                }}
              >
                <span
                  style={{
                    fontFamily: DISPLAY,
                    fontWeight: 600,
                    fontSize: 36,
                    letterSpacing: '-0.02em',
                    color: on ? C.text : C.muted,
                  }}
                >
                  {r.name}
                </span>
                <span
                  style={{
                    fontFamily: BODY,
                    fontWeight: 300,
                    fontSize: 26,
                    color: on ? C.muted : C.dim,
                  }}
                >
                  {r.note}
                </span>
              </div>
              {on ? (
                <div
                  style={{
                    fontFamily: MONO,
                    fontSize: 20,
                    letterSpacing: '0.16em',
                    textTransform: 'uppercase',
                    ...grad,
                    opacity: ramp(frame, d + 16, 16),
                  }}
                >
                  ← start here
                </div>
              ) : null}
            </div>
          );
        })}
      </div>
    </Pad>
  );
};

// ---------------------------------------------------------------- tiles

export const Tiles: React.FC<{
  title: string;
  items: {name: string; note: string; icon: GlyphName}[];
  sweep?: number[];
}> = ({title, items, sweep}) => {
  const frame = useCurrentFrame();
  return (
    <Pad style={{justifyContent: 'center', gap: 48}}>
      <Headline text={title} size={66} delay={0} width={1400} />
      <div style={{display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 24}}>
        {items.map((it, i) => {
          const d = 14 + i * 7;
          // current item only: lit from its own cue until the next one
          const next = sweep?.[i + 1];
          const hot =
            sweep?.[i] !== undefined &&
            frame >= (sweep[i] as number) &&
            (next === undefined || frame < next);
          const draw = ramp(frame, d + 6, 20);
          return (
            <Glass key={it.name} delay={d} active={hot} style={{padding: '32px 30px 30px'}}>
              <div style={{position: 'relative', height: 74, marginBottom: 12}}>
                <div style={{position: 'absolute', left: -6, top: -4}}>
                  <Glyph
                    name={it.icon}
                    size={62}
                    color={hot ? C.rose : C.violet}
                    on={draw}
                  />
                </div>
              </div>
              <div
                style={{
                  fontFamily: DISPLAY,
                  fontWeight: 600,
                  fontSize: 34,
                  letterSpacing: '-0.02em',
                  color: C.text,
                  marginBottom: 9,
                }}
              >
                {it.name}
              </div>
              <div
                style={{
                  fontFamily: BODY,
                  fontWeight: 300,
                  fontSize: 25,
                  lineHeight: 1.44,
                  color: C.muted,
                }}
              >
                {it.note}
              </div>
            </Glass>
          );
        })}
      </div>
    </Pad>
  );
};

// ---------------------------------------------------------------- loop

/** describe → look → change → repeat, drawn as a closing ring. */
export const Loop: React.FC<{title: string; steps: {name: string; note: string}[]}> = ({
  title,
  steps,
}) => {
  const frame = useCurrentFrame();
  const R = 250;
  const cx = 460;
  const cy = 300;
  const n = steps.length;
  const arc = ramp(frame, 16, 60);
  const circ = 2 * Math.PI * R;
  return (
    <Pad style={{justifyContent: 'center', gap: 30}}>
      <Headline text={title} size={64} delay={0} width={1400} />
      <div style={{display: 'flex', alignItems: 'center', gap: 80, marginTop: 8}}>
        <svg width={920} height={600} style={{flex: 'none', overflow: 'visible'}}>
          <defs>
            <linearGradient id="loopg" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stopColor={C.rose} />
              <stop offset="100%" stopColor={C.violet} />
            </linearGradient>
          </defs>
          <circle cx={cx} cy={cy} r={R} fill="none" stroke="rgba(255,255,255,0.07)" strokeWidth={2} />
          <circle
            cx={cx}
            cy={cy}
            r={R}
            fill="none"
            stroke="url(#loopg)"
            strokeWidth={3.4}
            strokeLinecap="round"
            strokeDasharray={circ}
            strokeDashoffset={circ * (1 - arc)}
            transform={`rotate(-90 ${cx} ${cy})`}
          />
          {steps.map((st, i) => {
            const ang = (i / n) * Math.PI * 2 - Math.PI / 2;
            const x = cx + Math.cos(ang) * R;
            const y = cy + Math.sin(ang) * R;
            const a = ramp(frame, 24 + i * 12, 16);
            return (
              <g key={st.name} opacity={a}>
                <circle cx={x} cy={y} r={19} fill={C.ink0} stroke={C.rose} strokeWidth={3} />
                <text
                  x={x}
                  y={y + 7}
                  textAnchor="middle"
                  style={{fontFamily: MONO, fontSize: 19, fontWeight: 500, fill: C.rose}}
                >
                  {String(i + 1).padStart(2, '0')}
                </text>
              </g>
            );
          })}
          <text
            x={cx}
            y={cy - 6}
            textAnchor="middle"
            style={{
              fontFamily: SERIF,
              fontStyle: 'italic',
              fontSize: 46,
              fill: C.muted,
              opacity: ramp(frame, 50, 20),
            }}
          >
            and again
          </text>
          <text
            x={cx}
            y={cy + 42}
            textAnchor="middle"
            style={{
              fontFamily: MONO,
              fontSize: 22,
              letterSpacing: '0.2em',
              fill: C.dim,
              opacity: ramp(frame, 56, 20),
            }}
          >
            NO CODE
          </text>
        </svg>
        <div style={{flex: 1, display: 'flex', flexDirection: 'column', gap: 20, marginLeft: -120}}>
          {steps.map((st, i) => {
            const a = ramp(frame, 26 + i * 12, 18);
            return (
              <div
                key={st.name}
                style={{opacity: a, transform: `translateX(${(1 - a) * 20}px)`}}
              >
                <div style={{display: 'flex', alignItems: 'baseline', gap: 16}}>
                  <span
                    style={{
                      fontFamily: MONO,
                      fontSize: 20,
                      color: C.rose,
                    }}
                  >
                    {String(i + 1).padStart(2, '0')}
                  </span>
                  <span
                    style={{
                      fontFamily: DISPLAY,
                      fontWeight: 600,
                      fontSize: 40,
                      letterSpacing: '-0.02em',
                      color: C.text,
                    }}
                  >
                    {st.name}
                  </span>
                </div>
                <div
                  style={{
                    marginLeft: 44,
                    fontFamily: BODY,
                    fontWeight: 300,
                    fontSize: 27,
                    lineHeight: 1.42,
                    color: C.muted,
                  }}
                >
                  {st.note}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </Pad>
  );
};

// ---------------------------------------------------------------- quiz

export const Quiz: React.FC<{
  n: string;
  question: string;
  options: string[];
  answer: number;
  revealAt: number;
}> = ({n, question, options, answer, revealAt}) => {
  const frame = useCurrentFrame();
  const shown = frame >= revealAt;
  return (
    <Pad style={{justifyContent: 'center', gap: 42}}>
      <div style={{display: 'flex', alignItems: 'center', gap: 20}}>
        <Eyebrow delay={0}>Checkpoint {n}</Eyebrow>
        <div style={{flex: 1, height: 1, backgroundColor: C.line, opacity: ramp(frame, 6, 24)}} />
      </div>
      <Headline text={question} size={62} delay={6} width={1520} />
      <div style={{display: 'flex', flexDirection: 'column', gap: 15, marginTop: 6}}>
        {options.map((op, i) => {
          const d = 20 + i * 8;
          const right = shown && i === answer;
          const wrong = shown && i !== answer;
          return (
            <Glass
              key={op}
              delay={d}
              active={right}
              style={{
                padding: '22px 30px',
                opacity: wrong ? 0.4 : 1,
                display: 'flex',
                alignItems: 'center',
                gap: 24,
              }}
            >
              <div
                style={{
                  width: 42,
                  height: 42,
                  borderRadius: 11,
                  flex: 'none',
                  display: 'grid',
                  placeItems: 'center',
                  backgroundImage: right ? GRAD : 'none',
                  backgroundColor: right ? undefined : 'rgba(255,255,255,0.05)',
                  fontFamily: MONO,
                  fontSize: 20,
                  color: right ? '#140A12' : C.dim,
                }}
              >
                {'ABCD'[i]}
              </div>
              <div
                style={{
                  fontFamily: BODY,
                  fontWeight: right ? 400 : 300,
                  fontSize: 34,
                  color: right ? C.text : C.muted,
                }}
              >
                {op}
              </div>
            </Glass>
          );
        })}
      </div>
    </Pad>
  );
};

// ---------------------------------------------------------------- timeline

export const Timeline: React.FC<{
  title: string;
  nodes: {name: string; state: 'done' | 'here' | 'next'}[];
}> = ({title, nodes}) => {
  const frame = useCurrentFrame();
  const grow = ramp(frame, 14, 60);
  const hereIdx = nodes.findIndex((x) => x.state === 'here');
  return (
    <Pad style={{justifyContent: 'center', gap: 70}}>
      <Headline text={title} size={68} delay={0} align="center" width="100%" />
      <div style={{position: 'relative', height: 260}}>
        {/* the spine */}
        <div
          style={{
            position: 'absolute',
            left: 0,
            right: 0,
            top: 128,
            height: 2,
            backgroundColor: C.line,
          }}
        />
        <div
          style={{
            position: 'absolute',
            left: 0,
            top: 127,
            height: 4,
            width: `${grow * ((hereIdx + 0.5) / nodes.length) * 100}%`,
            backgroundImage: GRAD,
            borderRadius: 2,
          }}
        />
        {nodes.map((nd, i) => {
          const x = ((i + 0.5) / nodes.length) * 100;
          const d = 20 + i * 9;
          const a = ramp(frame, d, 18);
          const up = i % 2 === 0;
          const done = nd.state === 'done';
          const here = nd.state === 'here';
          return (
            <div key={nd.name}>
              <div
                style={{
                  position: 'absolute',
                  left: `${x}%`,
                  top: 128,
                  width: here ? 20 : 13,
                  height: here ? 20 : 13,
                  marginLeft: here ? -10 : -6.5,
                  marginTop: here ? -10 : -6.5,
                  borderRadius: '50%',
                  backgroundImage: done || here ? GRAD : 'none',
                  backgroundColor: done || here ? undefined : C.ink0,
                  border: done || here ? 'none' : `2px solid ${C.line}`,
                  opacity: a,
                  boxShadow: here ? `0 0 0 8px rgba(255,77,109,0.16)` : 'none',
                }}
              />
              <div
                style={{
                  position: 'absolute',
                  left: `${x}%`,
                  top: up ? 128 - 74 : 128 + 34,
                  transform: 'translateX(-50%)',
                  opacity: a,
                  textAlign: 'center',
                  width: 250,
                }}
              >
                {here ? (
                  <div
                    style={{
                      fontFamily: MONO,
                      fontSize: 17,
                      letterSpacing: '0.18em',
                      ...grad,
                      marginBottom: 6,
                    }}
                  >
                    YOU ARE HERE
                  </div>
                ) : null}
                <div
                  style={{
                    fontFamily: BODY,
                    fontWeight: here ? 500 : 300,
                    fontSize: 27,
                    lineHeight: 1.3,
                    color: here ? C.text : done ? C.muted : C.dim,
                  }}
                >
                  {nd.name}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </Pad>
  );
};

// ---------------------------------------------------------------- outro

export const Outro: React.FC<{lines: string[]; sign: string}> = ({lines, sign}) => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill>
      <Motes n={44} seed="out" />
      <Pad style={{justifyContent: 'center', gap: 26}}>
        {lines.map((l, i) => (
          <Headline
            key={l}
            text={l}
            size={i === 0 ? 92 : 62}
            delay={6 + i * 22}
            accent={i === 0 ? ['built'] : []}
            serif={i === 0 ? ['something'] : []}
            width={1560}
          />
        ))}
        <div style={{marginTop: 22}}>
          <Rule delay={70} width={260} />
        </div>
        <div
          style={{
            fontFamily: MONO,
            fontSize: 24,
            letterSpacing: '0.22em',
            textTransform: 'uppercase',
            color: C.dim,
            opacity: ramp(frame, 82, 24),
          }}
        >
          {sign}
        </div>
      </Pad>
    </AbsoluteFill>
  );
};
