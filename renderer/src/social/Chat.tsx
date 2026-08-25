// The assistant, working.
//
// In the reference lesson these stretches were captured video of Claude Desktop
// building a Flask site. That footage cannot be re-recorded here, so this is the
// same content rebuilt as a component: the prompt types itself into the
// composer, sends, and the reply streams in with its tool rows landing one at a
// time. It is drawn in this film's own visual language rather than traced over
// anybody's product chrome — it is a depiction of a conversation, not a forged
// screenshot of an application.
//
// Two things make it read as real rather than as a slideshow:
//
//  * Steps are weighted, not clocked. A beat's length comes from how long the
//    narration takes to say, which is not known until the voice is built, so a
//    chat scored in absolute frames would either overrun or sit finished and
//    still. Weights are normalised against the beat, so the conversation always
//    lands exactly on the last word.
//  * The transcript is bottom-anchored. `justify-content: flex-end` in a clipped
//    column means new content pushes old content off the top, which is what a
//    chat pinned to its own bottom actually looks like. No scroll maths, and it
//    can never show a gap under the newest line.

import React from 'react';
import {AbsoluteFill, useCurrentFrame} from 'remotion';
import {C, MONO, BODY, DISPLAY, ramp, lin, GRAD} from './kit';

export type ChatStep =
  | {k: 'you'; text: string; w?: number}
  | {k: 'say'; text: string; w?: number}
  | {k: 'tool'; text: string; w?: number}
  | {k: 'files'; items: string[]; w?: number}
  | {k: 'wait'; w?: number}
  | {k: 'done'; text: string; w?: number};

export type ChatSpec = {
  title: string;
  steps: ChatStep[];
  /** Zoom the panel a little for a close read of the reply. */
  zoom?: number;
};

const BOX_W = 1420;
const BOX_H = 712;
const CHROME_H = 50;

const weightOf = (s: ChatStep) => {
  if (s.w) return s.w;
  if (s.k === 'you') return Math.max(2.4, s.text.length / 34);
  if (s.k === 'say') return Math.max(2.6, s.text.length / 46);
  if (s.k === 'done') return Math.max(2.6, s.text.length / 46);
  if (s.k === 'files') return 1.1 + s.items.length * 0.5;
  if (s.k === 'tool') return 1.5;
  return 1.2;
};

/** Text revealed a character at a time, with a caret while it is still coming. */
const Typed: React.FC<{
  text: string;
  t: number;
  style?: React.CSSProperties;
  caret?: boolean;
  frame?: number;
}> = ({text, t, style, caret = true, frame = 0}) => {
  const n = Math.round(Math.min(1, Math.max(0, t)) * text.length);
  const going = t < 1;
  return (
    <span style={style}>
      {text.slice(0, n)}
      {caret && going ? (
        <span
          style={{
            display: 'inline-block',
            width: 2,
            height: '1em',
            marginLeft: 2,
            verticalAlign: '-0.13em',
            backgroundColor: C.rose,
            opacity: Math.floor(frame / 8) % 2 ? 0.25 : 1,
          }}
        />
      ) : null}
    </span>
  );
};

const Spinner: React.FC<{frame: number; done: boolean}> = ({frame, done}) =>
  done ? (
    <svg width={17} height={17} viewBox="0 0 24 24" fill="none" stroke={C.ok} strokeWidth={2.8}
         strokeLinecap="round" strokeLinejoin="round">
      <polyline points="4,12.5 9.5,18 20,6.5" />
    </svg>
  ) : (
    <svg width={17} height={17} viewBox="0 0 24 24"
         style={{transform: `rotate(${(frame * 9) % 360}deg)`}}>
      <circle cx="12" cy="12" r="9" fill="none" stroke="rgba(255,255,255,0.16)" strokeWidth="2.6" />
      <path d="M12 3 A9 9 0 0 1 21 12" fill="none" stroke={C.violet} strokeWidth="2.6"
            strokeLinecap="round" />
    </svg>
  );

const Dots: React.FC<{frame: number}> = ({frame}) => (
  <div style={{display: 'flex', gap: 7, alignItems: 'center', padding: '6px 0'}}>
    {[0, 1, 2].map((i) => (
      <div
        key={i}
        style={{
          width: 8,
          height: 8,
          borderRadius: '50%',
          backgroundColor: C.muted,
          opacity: 0.3 + 0.6 * Math.abs(Math.sin((frame / 11 + i * 0.4) * Math.PI)),
        }}
      />
    ))}
  </div>
);

export const Chat: React.FC<{spec: ChatSpec; durationInFrames: number}> = ({
  spec,
  durationInFrames,
}) => {
  const frame = useCurrentFrame();

  // Lay the steps out across the beat by weight, keeping a short tail so the
  // last line is readable after it finishes rather than cutting on its caret.
  const tail = Math.min(26, Math.round(durationInFrames * 0.12));
  const span = Math.max(30, durationInFrames - tail);
  const ws = spec.steps.map(weightOf);
  const total = ws.reduce((a, b) => a + b, 0) || 1;
  let acc = 0;
  const plan = spec.steps.map((st, i) => {
    const from = Math.round((acc / total) * span);
    acc += ws[i];
    const to = Math.round((acc / total) * span);
    return {st, from, to};
  });

  // What is in the composer: whichever `you` step is mid-type, if any.
  const live = plan.find((p) => p.st.k === 'you' && frame >= p.from && frame < p.to);
  const shown = plan.filter((p) => frame >= p.from && !(p.st.k === 'you' && frame < p.to));

  const zoom = spec.zoom ?? 1;
  const lift = ramp(frame, 0, 10);

  const row = (p: (typeof plan)[number], i: number) => {
    const t = lin(frame, p.from, Math.max(6, p.to - p.from));
    const a = ramp(frame, p.from, 9);
    const key = `${i}`;

    if (p.st.k === 'you') {
      return (
        <div key={key} style={{display: 'flex', justifyContent: 'flex-end', margin: '18px 0 6px'}}>
          <div
            style={{
              maxWidth: '76%',
              background: 'rgba(255,255,255,0.07)',
              border: `1px solid ${C.line}`,
              borderRadius: '16px 16px 5px 16px',
              padding: '13px 19px',
              fontFamily: BODY,
              fontSize: 26,
              lineHeight: 1.45,
              color: C.text,
              opacity: a,
            }}
          >
            {p.st.text}
          </div>
        </div>
      );
    }

    if (p.st.k === 'tool') {
      return (
        <div
          key={key}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 12,
            padding: '9px 14px',
            margin: '7px 0',
            borderRadius: 10,
            background: 'rgba(255,255,255,0.035)',
            border: `1px solid ${C.line2}`,
            fontFamily: MONO,
            fontSize: 21,
            color: C.muted,
            opacity: a,
          }}
        >
          <Spinner frame={frame - p.from} done={t >= 1} />
          <span>{p.st.text}</span>
        </div>
      );
    }

    if (p.st.k === 'files') {
      const n = Math.round(t * p.st.items.length);
      return (
        <div
          key={key}
          style={{
            margin: '9px 0',
            borderRadius: 12,
            border: `1px solid ${C.line2}`,
            background: 'rgba(255,255,255,0.03)',
            overflow: 'hidden',
            opacity: a,
          }}
        >
          {p.st.items.map((f, k) => (
            <div
              key={f}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 12,
                padding: '9px 16px',
                fontFamily: MONO,
                fontSize: 20,
                color: k < n ? C.text : 'transparent',
                borderTop: k ? `1px solid ${C.line2}` : 'none',
              }}
            >
              <span style={{color: k < n ? C.ok : 'transparent', fontSize: 17}}>+</span>
              <span>{f}</span>
            </div>
          ))}
        </div>
      );
    }

    if (p.st.k === 'wait') {
      return t >= 1 ? null : (
        <div key={key} style={{opacity: a}}>
          <div style={{fontFamily: MONO, fontSize: 19, color: C.dim, marginBottom: 2}}>
            Thinking
          </div>
          <Dots frame={frame - p.from} />
        </div>
      );
    }

    // 'say' and 'done'
    const isDone = p.st.k === 'done';
    return (
      <div
        key={key}
        style={{
          margin: '11px 0',
          fontFamily: BODY,
          fontWeight: 300,
          fontSize: 26,
          lineHeight: 1.5,
          color: isDone ? C.text : C.muted,
          opacity: a,
          ...(isDone
            ? {
                borderLeft: `3px solid ${C.rose}`,
                paddingLeft: 18,
                fontWeight: 400,
              }
            : null),
        }}
      >
        <Typed text={p.st.text} t={t} frame={frame} />
      </div>
    );
  };

  return (
    <AbsoluteFill style={{alignItems: 'center', justifyContent: 'center'}}>
      <div
        style={{
          position: 'absolute',
          width: BOX_W * 1.08,
          height: (BOX_H + CHROME_H) * 1.1,
          borderRadius: 60,
          background:
            'radial-gradient(closest-side, rgba(139,92,246,0.20), rgba(255,77,109,0.09) 56%, transparent 76%)',
          opacity: lift,
        }}
      />
      <div
        style={{
          width: BOX_W,
          height: BOX_H + CHROME_H,
          borderRadius: 18,
          overflow: 'hidden',
          background: 'linear-gradient(180deg,#12101C,#0C0A14)',
          border: '1px solid rgba(255,255,255,0.11)',
          boxShadow:
            '0 60px 140px rgba(0,0,0,0.72), inset 0 1px 0 rgba(255,255,255,0.10)',
          opacity: lift,
          transform: `translateY(${(1 - lift) * 18}px) scale(${(0.988 + lift * 0.012) * zoom})`,
        }}
      >
        {/* window bar */}
        <div
          style={{
            height: CHROME_H,
            display: 'flex',
            alignItems: 'center',
            gap: 9,
            paddingLeft: 18,
            borderBottom: `1px solid ${C.line2}`,
          }}
        >
          {['#FF5F57', '#FEBC2E', '#28C840'].map((c) => (
            <div key={c} style={{width: 11, height: 11, borderRadius: 6, backgroundColor: c}} />
          ))}
          <div
            style={{
              flex: 1,
              textAlign: 'center',
              fontFamily: BODY,
              fontSize: 20,
              color: C.dim,
              marginRight: 56,
            }}
          >
            {spec.title}
          </div>
        </div>

        <div style={{display: 'flex', height: BOX_H}}>
          {/* a thin rail, so the panel reads as an app and not as a card */}
          <div
            style={{
              width: 78,
              flex: 'none',
              borderRight: `1px solid ${C.line2}`,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              paddingTop: 22,
              gap: 20,
            }}
          >
            <div
              style={{
                width: 34,
                height: 34,
                borderRadius: 10,
                backgroundImage: GRAD,
                display: 'grid',
                placeItems: 'center',
                fontFamily: DISPLAY,
                fontWeight: 800,
                fontSize: 19,
                color: '#0B0810',
              }}
            >
              C
            </div>
            {[0, 1, 2].map((i) => (
              <div
                key={i}
                style={{
                  width: 26,
                  height: 3,
                  borderRadius: 2,
                  backgroundColor: 'rgba(255,255,255,0.10)',
                }}
              />
            ))}
          </div>

          {/* transcript, pinned to its own bottom */}
          <div style={{flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0}}>
            <div
              style={{
                flex: 1,
                overflow: 'hidden',
                padding: '26px 42px 8px',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'flex-end',
              }}
            >
              {shown.map(row)}
            </div>

            {/* composer */}
            <div style={{padding: '14px 42px 22px'}}>
              <div
                style={{
                  border: `1px solid ${live ? 'rgba(255,120,150,0.4)' : C.line}`,
                  borderRadius: 14,
                  background: 'rgba(255,255,255,0.035)',
                  padding: '15px 20px',
                  minHeight: 58,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 14,
                }}
              >
                <div
                  style={{
                    flex: 1,
                    fontFamily: BODY,
                    fontSize: 25,
                    lineHeight: 1.4,
                    color: live ? C.text : C.dim,
                    fontWeight: 300,
                  }}
                >
                  {live ? (
                    <Typed
                      text={live.st.k === 'you' ? live.st.text : ''}
                      t={lin(frame, live.from, Math.max(6, live.to - live.from))}
                      frame={frame}
                    />
                  ) : (
                    'Reply to Claude…'
                  )}
                </div>
                <div
                  style={{
                    width: 34,
                    height: 34,
                    borderRadius: 9,
                    backgroundImage: live ? GRAD : 'none',
                    background: live ? undefined : 'rgba(255,255,255,0.07)',
                    display: 'grid',
                    placeItems: 'center',
                  }}
                >
                  <svg width={17} height={17} viewBox="0 0 24 24" fill="none"
                       stroke={live ? '#100A12' : C.dim} strokeWidth={2.4}
                       strokeLinecap="round" strokeLinejoin="round">
                    <path d="M12 19V5M6 11l6-6 6 6" />
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </AbsoluteFill>
  );
};
