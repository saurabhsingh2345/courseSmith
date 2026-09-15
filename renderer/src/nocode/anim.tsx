// The no-code animation template library.
//
// Every template here exists because a specific lecture beat needed it and a
// bullet list would have been worse. They are all frame-driven off
// useCurrentFrame so they step deterministically under the renderer — CSS
// keyframes do not, which is why the client's HTML deck could never be filmed.

import React from 'react';
import {interpolate, useCurrentFrame} from 'remotion';
import {C, MONO, DISPLAY, easeIn, easeSlam, easeOut, pr, accent, useCues} from './motion';

/* ------------------------------------------------------------------ tokens */
/** text breaking into tokens, each chip landing in turn. The point of the beat
 *  is that a model does not see words, so the chips must visibly disagree with
 *  the word boundaries. */
export const Tokens: React.FC<{
  chips: string[]; d?: number; step?: number; hot?: number[];
}> = ({chips, d = 0.3, step = 0.11, hot = []}) => {
  const f = useCurrentFrame();
  // When the deck supplies cues, a chip lands on the word being spoken — this
  // template's beat IS the list being read out loud, so the default 0.11s
  // cascade would empty the whole list onto the screen before he says the
  // second item.
  const cue = useCues(d, step);
  return (
    <div style={{display: 'flex', flexWrap: 'wrap', gap: 9, justifyContent: 'center'}}>
      {chips.map((t, i) => {
        const p = easeSlam(pr(f, cue(i, chips.length), 0.42));
        const on = hot.includes(i);
        return (
          <span key={i} style={{
            fontFamily: MONO, fontSize: 46, padding: '13px 18px', borderRadius: 8,
            border: `1.5px solid ${on ? C.yellow : '#3a3a3a'}`,
            background: on ? 'rgba(245,197,24,0.14)' : '#131313',
            color: on ? C.yellow : C.dim,
            whiteSpace: 'pre', opacity: Math.min(1, p / 0.4),
            transform: `scale(${interpolate(p, [0, 0.6, 1], [1.35, 0.97, 1])})`,
          }}>{t}</span>
        );
      })}
    </div>
  );
};

/** the candidate next tokens with their probabilities — bars grow, the winner
 *  stays lit. This is the beat where "it predicts text" stops being a slogan. */
export const Probs: React.FC<{
  items: {t: string; p: number}[]; d?: number;
}> = ({items, d = 0.4}) => {
  const f = useCurrentFrame();
  const max = Math.max(...items.map((i) => i.p));
  return (
    <div style={{display: 'flex', flexDirection: 'column', gap: 13, width: '100%'}}>
      {items.map((it, i) => {
        const a = easeIn(pr(f, d + i * 0.14, 0.5));
        const g = easeOut(pr(f, d + 0.25 + i * 0.14, 0.85));
        const win = it.p === max;
        return (
          <div key={i} style={{display: 'flex', alignItems: 'center', gap: 18, opacity: a}}>
            <span style={{
              fontFamily: MONO, fontSize: 30, width: '5.4em', textAlign: 'right',
              color: win ? C.yellow : C.dim, whiteSpace: 'pre',
            }}>{it.t}</span>
            <div style={{flex: 1, height: 34, background: '#131313', borderRadius: 5,
                         border: '1px solid #262626', overflow: 'hidden'}}>
              <div style={{
                width: `${(it.p / max) * 100 * g}%`, height: '100%',
                background: win ? C.yellow : '#3a3a3a',
              }} />
            </div>
            <span style={{
              fontFamily: MONO, fontSize: 24, width: '4.6em',
              color: win ? C.yellow : C.muted,
            }}>{it.p >= 1 ? `${it.p}%` : `${it.p}%`}</span>
          </div>
        );
      })}
    </div>
  );
};

/* -------------------------------------------------------------------- chat */
/** a conversation that is re-sent in full on every turn. `resend` is the turn
 *  index whose send lights the WHOLE stack — which is the entire trick behind
 *  the illusion of memory, and impossible to show in a still. */
export const Chat: React.FC<{
  turns: {who: 'you' | 'model'; t: string}[]; d?: number; resend?: number;
}> = ({turns, d = 0.3, resend}) => {
  const f = useCurrentFrame();
  const flash = resend === undefined ? 0
    : easeIn(pr(f, d + resend * 0.75 + 0.5, 0.4)) * (1 - easeIn(pr(f, d + resend * 0.75 + 1.9, 0.6)));
  return (
    <div style={{
      display: 'flex', flexDirection: 'column', gap: 11, width: '100%',
      padding: 18, borderRadius: 12,
      border: `2px solid ${flash > 0.05 ? C.yellow : 'transparent'}`,
      background: flash > 0.05 ? `rgba(245,197,24,${0.08 * flash})` : 'transparent',
    }}>
      {turns.map((t, i) => {
        const a = easeIn(pr(f, d + i * 0.75, 0.45));
        const mine = t.who === 'you';
        return (
          <div key={i} style={{
            alignSelf: mine ? 'flex-end' : 'flex-start', maxWidth: '76%',
            fontFamily: MONO, fontSize: 25, lineHeight: 1.4,
            padding: '13px 18px', borderRadius: 12,
            border: `1px solid ${mine ? '#3d3d3d' : '#2a2a2a'}`,
            background: mine ? '#1c1c1c' : '#101010',
            color: mine ? C.ink : C.dim,
            opacity: a, transform: `translateY(${14 * (1 - a)}px)`,
          }}>{t.t}</div>
        );
      })}
    </div>
  );
};

/* -------------------------------------------------------------------- flow */
/** boxes joined by arrows that draw in, one hop at a time. Built for the beat
 *  everyone gets wrong: the model never runs the tool, our code does. */
export const Flow: React.FC<{
  nodes: {t: string; sub?: string; tone?: string}[]; d?: number; loop?: boolean;
}> = ({nodes, d = 0.3, loop}) => {
  const f = useCurrentFrame();
  const cue = useCues(d, 0.62);
  return (
    <div style={{display: 'flex', alignItems: 'stretch', gap: 0, width: '100%',
                 position: 'relative', marginBottom: loop ? 92 : 0}}>
      {nodes.map((n, i) => {
        const at = cue(i, nodes.length);
        const a = easeIn(pr(f, at, 0.5));
        const arrow = easeOut(pr(f, at + 0.42, 0.34));
        const tone = n.tone ?? '#3a3a3a';
        return (
          <React.Fragment key={i}>
            <div style={{
              flex: 1, borderRadius: 11, border: `2px solid ${tone}`,
              background: tone === '#3a3a3a' ? '#111' : `${tone}1c`,
              padding: '22px 18px', display: 'flex', flexDirection: 'column',
              justifyContent: 'center', gap: 9, minHeight: 190,
              opacity: a, transform: `translateY(${18 * (1 - a)}px)`,
            }}>
              <div style={{fontFamily: DISPLAY, fontSize: 27, textTransform: 'uppercase',
                           color: tone === '#3a3a3a' ? C.ink : tone, lineHeight: 1.05}}>
                {n.t}
              </div>
              {n.sub ? (
                <div style={{fontFamily: MONO, fontSize: 18, color: C.muted, lineHeight: 1.35}}>
                  {n.sub}
                </div>
              ) : null}
            </div>
            {i < nodes.length - 1 ? (
              <div style={{width: 66, display: 'flex', alignItems: 'center',
                           justifyContent: 'center', flexShrink: 0}}>
                <svg width="66" height="26" viewBox="0 0 66 26">
                  <line x1="6" y1="13" x2={6 + 44 * arrow} y2="13"
                        stroke={C.muted} strokeWidth="2.4" />
                  <path d={`M${6 + 44 * arrow} 13 l-9 -6 v12 z`} fill={C.muted}
                        opacity={arrow > 0.85 ? 1 : 0} />
                </svg>
              </div>
            ) : null}
          </React.Fragment>
        );
      })}
      {loop ? (() => {
        const a = easeOut(pr(f, d + nodes.length * 0.62 + 0.3, 0.7));
        return (
          <svg style={{position: 'absolute', left: 0, right: 0, bottom: -74, height: 74}}
               width="100%" height="74" preserveAspectRatio="none" viewBox="0 0 100 74">
            <path d="M97 2 V40 H3 V2" fill="none" stroke={C.yellow} strokeWidth="0.7"
                  strokeDasharray="200" strokeDashoffset={200 - 200 * a}
                  vectorEffect="non-scaling-stroke" />
            <path d="M3 2 l-1.4 5 h2.8 z" fill={C.yellow} opacity={a > 0.9 ? 1 : 0} />
          </svg>
        );
      })() : null}
    </div>
  );
};

/* ---------------------------------------------------------------- timeline */
/** dated entries arriving in order along a spine — for "how the definition
 *  changed" beats, where the order is the argument. */
export const Timeline: React.FC<{
  items: {when: string; t: string; sub?: string; hot?: boolean}[];
  d?: number; upto?: number;
}> = ({items, d = 0.3, upto}) => {
  const f = useCurrentFrame();
  const spine = easeOut(pr(f, d, 1.1));
  return (
    <div style={{position: 'relative', width: '100%', paddingLeft: 46}}>
      <div style={{
        position: 'absolute', left: 11, top: 6, width: 2, background: '#2c2c2c',
        height: `${spine * 100}%`,
      }} />
      {items.map((it, i) => {
        const shown = upto ?? items.length;
        if (i >= shown) return null;
        // entries already standing are simply there; only the newest arrives
        const a = i === shown - 1 ? easeIn(pr(f, d + 0.35, 0.5)) : 1;
        return (
          <div key={i} style={{position: 'relative', marginBottom: 26, opacity: a,
                               transform: `translateX(${-16 * (1 - a)}px)`}}>
            {/* Anchored to the item, not to the frame: with no `top` of its own
                the marker fell at its static position, which is exactly where
                the `when` label sits, and printed the dot over the day name.
                left: -43 puts its centre on the spine (container left 11 + 1,
                less the container's 46px of padding). */}
            <span style={{
              position: 'absolute', left: -43, top: 3, width: 18, height: 18,
              borderRadius: 99, background: it.hot ? C.yellow : '#1a1a1a',
              border: `2px solid ${it.hot ? C.yellow : '#3a3a3a'}`,
            }} />
            <div style={{fontFamily: MONO, fontSize: 15, letterSpacing: '0.18em',
                         textTransform: 'uppercase', color: C.muted, marginBottom: 6}}>
              {it.when}
            </div>
            <div style={{fontFamily: DISPLAY, fontSize: 33, textTransform: 'uppercase',
                         color: it.hot ? C.yellow : C.ink, lineHeight: 1.05}}>
              {accent(it.t)}
            </div>
            {it.sub ? (
              <div style={{fontFamily: MONO, fontSize: 20, color: C.dim, marginTop: 7,
                           lineHeight: 1.4}}>{it.sub}</div>
            ) : null}
          </div>
        );
      })}
    </div>
  );
};

/* -------------------------------------------------------------------- nest */
/** a box inside a box — the LLM inside the application around it. */
export const Nest: React.FC<{
  outer: string; inner: string; ring?: string[]; d?: number;
}> = ({outer, inner, ring = [], d = 0.3}) => {
  const f = useCurrentFrame();
  const cue = useCues(d, 0.5);
  const o = easeIn(pr(f, cue(0, 2), 0.6));
  const iP = easeSlam(pr(f, cue(1, 2), 0.5));
  return (
    <div style={{
      width: '78%', borderRadius: 18, border: `2px solid ${C.blue}`,
      background: 'rgba(59,130,246,0.07)', padding: '30px 30px 34px',
      opacity: o, transform: `scale(${0.96 + 0.04 * o})`,
    }}>
      <div style={{fontFamily: DISPLAY, fontSize: 30, textTransform: 'uppercase',
                   color: C.blue, marginBottom: 22}}>{outer}</div>
      <div style={{display: 'flex', gap: 20, alignItems: 'stretch'}}>
        <div style={{
          flex: '0 0 40%', borderRadius: 13, border: `2px solid ${C.yellow}`,
          background: 'rgba(245,197,24,0.12)', padding: '30px 22px',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          fontFamily: DISPLAY, fontSize: 40, textTransform: 'uppercase',
          color: C.yellow, textAlign: 'center',
          opacity: Math.min(1, iP / 0.4),
          transform: `scale(${interpolate(iP, [0, 0.65, 1], [1.25, 0.97, 1])})`,
        }}>{inner}</div>
        <div style={{flex: 1, display: 'flex', flexDirection: 'column', gap: 10}}>
          {ring.map((r, i) => {
            const a = easeIn(pr(f, d + 1.15 + i * 0.22, 0.45));
            return (
              <div key={i} style={{
                flex: 1, borderRadius: 9, border: '1px solid #333', background: '#111',
                display: 'flex', alignItems: 'center', padding: '0 18px',
                fontFamily: MONO, fontSize: 22, color: C.dim,
                opacity: a, transform: `translateX(${18 * (1 - a)}px)`,
              }}>{r}</div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

/* -------------------------------------------------------------------- tree */
/** an outcome tree, branches drawing then the surviving leaves lighting.
 *  Built for the two-coins beat, where the whole point is which rows survive. */
export const Tree: React.FC<{
  leaves: {t: string; keep?: boolean; note?: string}[]; d?: number;
}> = ({leaves, d = 0.3}) => {
  const f = useCurrentFrame();
  const cue = useCues(d, 0.3);
  return (
    <div style={{display: 'flex', gap: 16, width: '100%'}}>
      {leaves.map((l, i) => {
        const a = easeIn(pr(f, cue(i, leaves.length), 0.45));
        const cull = l.keep === false ? easeIn(pr(f, d + leaves.length * 0.3 + 0.7, 0.5)) : 0;
        const keep = l.keep === true ? easeIn(pr(f, d + leaves.length * 0.3 + 0.7, 0.5)) : 0;
        return (
          <div key={i} style={{
            flex: 1, borderRadius: 11, padding: '26px 16px', textAlign: 'center',
            border: `2px solid ${keep ? C.yellow : cull ? '#242424' : '#3a3a3a'}`,
            background: keep ? 'rgba(245,197,24,0.13)' : '#111',
            opacity: a * (1 - 0.62 * cull),
            transform: `translateY(${18 * (1 - a)}px) scale(${1 - 0.05 * cull})`,
          }}>
            <div style={{fontFamily: DISPLAY, fontSize: 40, textTransform: 'uppercase',
                         color: keep ? C.yellow : cull ? '#4a4a4a' : C.ink}}>{l.t}</div>
            {l.note ? (
              <div style={{fontFamily: MONO, fontSize: 17, color: C.muted, marginTop: 9}}>
                {l.note}
              </div>
            ) : null}
          </div>
        );
      })}
    </div>
  );
};

/* ---------------------------------------------------------------- terminal */
/** a mono panel whose lines type in. Cheaper and cleaner than filming a
 *  terminal, and it never leaks a username. */
export const Term: React.FC<{
  lines: {t: string; kind?: 'cmd' | 'out' | 'ok' | 'warn'}[];
  d?: number; cps?: number; title?: string; prompt?: string;
}> = ({lines, d = 0.3, cps = 42, title = 'arena', prompt = 'arena % '}) => {
  const f = useCurrentFrame();
  const t = f / 30 - d;
  let clock = 0;
  const col = {cmd: C.ink, out: C.dim, ok: '#5fd07a', warn: C.yellow};
  return (
    <div style={{
      width: '100%', borderRadius: 12, overflow: 'hidden',
      border: '1px solid rgba(255,255,255,0.10)', background: '#0c0c0c',
      boxShadow: '0 30px 80px rgba(0,0,0,0.55)',
    }}>
      <div style={{height: 44, background: '#151515', display: 'flex', alignItems: 'center',
                   gap: 8, padding: '0 16px',
                   borderBottom: '1px solid rgba(255,255,255,0.06)'}}>
        {[C.red, C.yellow, '#3ba55d'].map((c, i) => (
          <span key={i} style={{width: 10, height: 10, borderRadius: 99,
                                background: c, opacity: 0.5}} />
        ))}
        <span style={{fontFamily: MONO, fontSize: 14, color: C.muted, marginLeft: 12}}>
          {title}
        </span>
      </div>
      <div style={{padding: '20px 22px', fontFamily: MONO, fontSize: 23, lineHeight: 1.55}}>
        {lines.map((ln, i) => {
          const start = clock;
          const dur = ln.t.length / cps;
          clock += dur + 0.22;
          const shown = Math.max(0, Math.min(ln.t.length, Math.round((t - start) * cps)));
          if (t < start) return null;
          return (
            <div key={i} style={{color: col[ln.kind ?? 'out'], whiteSpace: 'pre-wrap'}}>
              {ln.kind === 'cmd' && prompt
                ? <span style={{color: C.yellow}}>{prompt}</span> : null}
              {ln.t.slice(0, shown)}
              {shown < ln.t.length ? (
                <span style={{background: C.ink, color: '#0c0c0c'}}>&nbsp;</span>
              ) : null}
            </div>
          );
        })}
      </div>
    </div>
  );
};

/* ------------------------------------------------------------------- meter */
/** the deck's own meter — a filled proportion with a label at each end.
 *  Ported from No-Code.zip's slides.html, which had it and we never used it. */
export const Meter: React.FC<{
  pct: number; fill: string; rest: string; d?: number; tone?: string;
}> = ({pct, fill, rest, d = 0.3, tone = C.yellow}) => {
  const f = useCurrentFrame();
  const cue = useCues(d, 1.0);
  // fill across the phrase that makes the claim, so the bar is still moving
  // while the words are still arriving
  const g = easeOut(pr(f, cue(0, 2), Math.max(1.0, cue(1, 2) - cue(0, 2))));
  return (
    <div style={{width: '100%', height: 74, border: `2px solid ${C.ink}`, display: 'flex',
                 overflow: 'hidden'}}>
      <div style={{
        width: `${pct * g}%`, background: tone, color: '#111', display: 'flex',
        alignItems: 'center', padding: '0 22px', fontFamily: MONO, fontWeight: 700,
        fontSize: 22, letterSpacing: '0.08em', textTransform: 'uppercase',
        whiteSpace: 'nowrap', overflow: 'hidden',
      }}>{fill}</div>
      <div style={{
        flex: 1, background: '#1b1b1b', display: 'flex', alignItems: 'center',
        justifyContent: 'flex-end', padding: '0 22px', fontFamily: MONO, fontSize: 22,
        color: C.muted, letterSpacing: '0.08em', textTransform: 'uppercase',
        whiteSpace: 'nowrap',
      }}>{rest}</div>
    </div>
  );
};

/* ------------------------------------------------------------------- gauge */
/** "does it fit?" — a bar filling toward a marked line, which either clears it
 *  or runs past. Ported from the Go catalog's `gauge`, whose header calls it
 *  "the single most repeated device in the reference clips". `meter` shows a
 *  proportion; only this one shows a proportion against a LIMIT, which is the
 *  whole question in every context-window beat. */
export const Gauge: React.FC<{
  pct: number; limit: number; label: string; limitLabel?: string;
  d?: number; over?: boolean;
}> = ({pct, limit, label, limitLabel = 'the limit', d = 0.35, over}) => {
  const f = useCurrentFrame();
  const g = easeOut(pr(f, d, 1.25));
  const w = pct * g;
  const busted = over ?? pct > limit;
  const hit = busted && w >= limit;
  const tone = hit ? C.red : C.yellow;
  const shake = hit ? Math.sin((f / 30 - d - 1.0) * 44) * Math.max(0, 1 - pr(f, d + 1.15, 0.5)) * 3 : 0;
  return (
    <div style={{width: '100%', transform: `translateX(${shake}px)`}}>
      <div style={{position: 'relative', height: 96, border: `2px solid ${C.ink}`,
                   background: '#131313', overflow: 'visible'}}>
        <div style={{
          position: 'absolute', inset: 0, width: `${Math.min(w, 100)}%`,
          background: tone, display: 'flex', alignItems: 'center', padding: '0 24px',
          fontFamily: DISPLAY, fontSize: 30, textTransform: 'uppercase', color: '#111',
          whiteSpace: 'nowrap', overflow: 'hidden',
        }}>{label}</div>
        {/* the line it has to clear */}
        <div style={{
          position: 'absolute', top: -16, bottom: -16, left: `${limit}%`, width: 3,
          background: C.ink,
        }} />
        <div style={{
          position: 'absolute', top: -46, left: `${limit}%`, transform: 'translateX(-50%)',
          fontFamily: MONO, fontSize: 16, letterSpacing: '0.16em',
          textTransform: 'uppercase', color: C.muted, whiteSpace: 'nowrap',
        }}>{limitLabel}</div>
      </div>
      <div style={{
        marginTop: 22, textAlign: 'center', fontFamily: DISPLAY, fontSize: 34,
        textTransform: 'uppercase', color: tone,
        opacity: easeIn(pr(f, d + 1.25, 0.4)),
      }}>{busted ? 'it does not fit' : 'it fits'}</div>
    </div>
  );
};

/* --------------------------------------------------------------------- myth */
/** what everyone believes, struck through, and what is actually true.
 *  Half of teaching is replacing a wrong model, not filling a gap — and a
 *  bullet list cannot show something being removed. */
export const Myth: React.FC<{
  wrong: string; right: string; note?: string; d?: number;
}> = ({wrong, right, note, d = 0.3}) => {
  const f = useCurrentFrame();
  // Three beats - the wrong model, striking it out, the true one - so they ride
  // three phrases of the narration when the deck knows them.
  const cue = useCues(d, 1.15);
  const c0 = cue(0, 3), c1 = cue(1, 3), c2 = cue(2, 3);
  const a = easeIn(pr(f, c0, 0.5));
  const strike = easeOut(pr(f, c1, 0.45));
  const fade = 1 - 0.55 * easeIn(pr(f, c1 + 0.35, 0.5));
  const b = easeSlam(pr(f, c2, 0.5));
  return (
    <div style={{width: '100%', display: 'flex', flexDirection: 'column', gap: 30}}>
      <div style={{position: 'relative', opacity: a * fade,
                   transform: `translateY(${16 * (1 - a)}px)`}}>
        <div style={{fontFamily: MONO, fontSize: 15, letterSpacing: '0.2em',
                     textTransform: 'uppercase', color: C.muted, marginBottom: 12}}>
          what everyone says
        </div>
        <div style={{fontFamily: DISPLAY, fontSize: 46, textTransform: 'uppercase',
                     color: '#8d8a83', lineHeight: 1.06}}>{wrong}</div>
        <div style={{
          position: 'absolute', left: 0, top: '64%', height: 5, background: C.red,
          width: `${strike * 100}%`,
        }} />
      </div>
      <div style={{
        opacity: Math.min(1, b / 0.4),
        transform: `translateY(${20 * (1 - b)}px)`,
      }}>
        <div style={{fontFamily: MONO, fontSize: 15, letterSpacing: '0.2em',
                     textTransform: 'uppercase', color: C.yellow, marginBottom: 12}}>
          what is actually happening
        </div>
        <div style={{fontFamily: DISPLAY, fontSize: 52, textTransform: 'uppercase',
                     color: C.ink, lineHeight: 1.06}}>{accent(right)}</div>
        {note ? (
          <div style={{fontFamily: MONO, fontSize: 22, color: C.dim, marginTop: 16,
                       lineHeight: 1.45}}>{note}</div>
        ) : null}
      </div>
    </div>
  );
};

/* ----------------------------------------------------------------- stepper */
/** watch it think: one step at a time, with what is carried forward growing.
 *  Pseudocode on a slide removes both the time and the data — this puts both
 *  back, which is the only way the inference beat lands. */
export const Stepper: React.FC<{
  seed: string; steps: string[]; d?: number; hold?: number;
}> = ({seed, steps, d = 0.4, hold = 1.05}) => {
  const f = useCurrentFrame();
  const shown = Math.max(0, Math.min(steps.length, Math.floor((f / 30 - d) / hold) + 1));
  return (
    <div style={{width: '100%', display: 'flex', flexDirection: 'column', gap: 16}}>
      {steps.map((_, i) => {
        if (i >= shown) return null;
        const local = easeIn(pr(f, d + i * hold, 0.4));
        const last = i === shown - 1;
        const carried = [seed, ...steps.slice(0, i + 1)].join('');
        return (
          <div key={i} style={{
            display: 'flex', alignItems: 'baseline', gap: 16,
            fontFamily: MONO, fontSize: 30, lineHeight: 1.35,
            padding: '13px 18px', borderRadius: 9,
            border: `1px solid ${last ? C.yellow : '#242424'}`,
            background: last ? 'rgba(245,197,24,0.07)' : 'transparent',
            color: last ? C.ink : '#7d7a73',
            opacity: local, transform: `translateY(${12 * (1 - local)}px)`,
          }}>
            <span style={{color: C.muted, fontSize: 18, minWidth: 34}}>
              {String(i + 1).padStart(2, '0')}
            </span>
            <span>
              <span style={{color: last ? '#9d9a93' : '#5e5b55'}}>{carried.slice(0, carried.length - steps[i].length)}</span>
              <span style={{color: last ? C.yellow : '#7d7a73'}}>{steps[i]}</span>
            </span>
          </div>
        );
      })}
    </div>
  );
};

/* --------------------------------------------------------------- occupancy */
/** a population of identical units, all visible at once, with part of it
 *  claimed. The deck had this as a lit-cell grid and we never used it — it is
 *  the honest picture of a context window, because it shows the whole thing. */
export const Occupancy = ({
  cols = 40, rows = 12, used, d = 0.3, tone = C.yellow, legend,
}: {cols?: number; rows?: number; used: number; d?: number; tone?: string;
   legend?: string}) => {
  const f = useCurrentFrame();
  const total = cols * rows;
  const fillP = easeOut(pr(f, d, 1.3));
  const lit = Math.round(total * (used / 100) * fillP);
  return (
    <div style={{width: '100%'}}>
      <div style={{
        display: 'grid', gridTemplateColumns: `repeat(${cols}, 1fr)`, gap: 5,
      }}>
        {Array.from({length: total}, (_, i) => {
          const on = i < lit;
          const a = easeIn(pr(f, d + (i % cols) * 0.006 + Math.floor(i / cols) * 0.02, 0.3));
          return (
            <span key={i} style={{
              aspectRatio: '1', borderRadius: 2.5,
              background: on ? tone : 'rgba(255,255,255,0.07)',
              opacity: on ? 1 : 0.32 * a,
              transform: on ? 'none' : 'scale(0.62)',
            }} />
          );
        })}
      </div>
      {legend ? (
        <div style={{
          marginTop: 24, display: 'flex', justifyContent: 'space-between',
          fontFamily: MONO, fontSize: 19, letterSpacing: '0.1em',
          textTransform: 'uppercase', color: C.muted,
          opacity: easeIn(pr(f, d + 1.0, 0.5)),
        }}>
          <span style={{color: tone}}>{legend}</span>
          <span>{`${Math.round(used)}% of the window`}</span>
        </div>
      ) : null}
    </div>
  );
};

/* ------------------------------------------------------------------- stack */
/** THE context stack: the layers a model actually reads on one call, revealed
 *  one at a time into a fixed container so the empty remainder stays visible.
 *  Band heights are proportional, which is the whole argument — by the time the
 *  conversation layer lands it owns most of the frame and almost none of it is
 *  anything you typed. Reused wherever context, compaction or an agents file
 *  comes up, so keep it general. */
export const Stack: React.FC<{
  layers: {t: string; sub?: string; h: number; tone?: string}[];
  upto?: number; hot?: number[]; d?: number; foot?: string;
}> = ({layers, upto, hot = [], d = 0.3, foot}) => {
  const f = useCurrentFrame();
  const shown = upto ?? layers.length;
  const total = layers.reduce((a, l) => a + l.h, 0);
  const H = 500;
  let top = 0;
  const boxes = layers.map((l, i) => {
    const y = top;
    const h = (l.h / total) * H;
    top += h;
    return {l, i, y, h};
  });
  return (
    <div style={{width: '100%'}}>
      <div style={{
        position: 'relative', height: H, border: `2px solid ${C.line}`,
        borderRadius: 10, background: '#0d0d0d', overflow: 'hidden',
      }}>
        {boxes.map(({l, i, y, h}) => {
          if (i >= shown) return null;
          // standing layers are already there; only the newest one arrives
          const p = i === shown - 1 ? easeSlam(pr(f, d, 0.5)) : 1;
          const on = hot.includes(i);
          const tone = l.tone ?? (on ? C.yellow : C.line);
          return (
            <div key={i} style={{
              position: 'absolute', left: 0, right: 0, top: y, height: h,
              borderTop: i ? `1px solid #202020` : 'none',
              background: on ? 'rgba(245,197,24,0.13)' : '#141414',
              display: 'flex', alignItems: 'center', gap: 20,
              padding: '0 26px',
              opacity: Math.min(1, p / 0.4),
              transform: `scaleY(${Math.min(1, p)})`, transformOrigin: 'top',
            }}>
              <span style={{width: 6, alignSelf: 'stretch', margin: '7px 0',
                            borderRadius: 3, background: tone, flexShrink: 0}} />
              <span style={{
                fontFamily: DISPLAY, fontSize: Math.min(30, h * 0.42),
                textTransform: 'uppercase', color: on ? C.yellow : C.ink,
                whiteSpace: 'nowrap',
              }}>{l.t}</span>
              {l.sub ? (
                <span style={{
                  marginLeft: 'auto', fontFamily: MONO,
                  fontSize: Math.min(21, h * 0.3), color: C.muted,
                  textAlign: 'right', lineHeight: 1.25, whiteSpace: 'pre-line',
                }}>{l.sub}</span>
              ) : null}
            </div>
          );
        })}
        {/* the remainder is not dead space — it is a countdown, so an early
            slide with one band in it still gives the eye something to read */}
        {shown < layers.length ? (
          <div style={{
            position: 'absolute', left: 0, right: 0,
            top: boxes[shown].y, bottom: 0, margin: 3,
            border: `2px dashed #262626`, borderRadius: 8,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontFamily: MONO, fontSize: 20, letterSpacing: '0.16em',
            textTransform: 'uppercase', color: '#4a4a4a',
            opacity: easeIn(pr(f, d + 0.5, 0.6)),
          }}>
            {layers.length - shown === 1 ? 'one more layer'
              : `${layers.length - shown} more layers`}
          </div>
        ) : null}
      </div>
      {foot ? (
        <div style={{
          marginTop: 20, fontFamily: MONO, fontSize: 19, letterSpacing: '0.1em',
          textTransform: 'uppercase', color: C.muted, textAlign: 'center',
          opacity: easeIn(pr(f, d + 0.7, 0.5)),
        }}>{foot}</div>
      ) : null}
    </div>
  );
};

/* ------------------------------------------------------------------- curve */
/** quality against how full the context is: a line that sags long before the
 *  wall, then the wall itself. He only ever says this out loud — the whole
 *  point is that the decline starts nowhere near the limit, and a sentence
 *  cannot show you where. */
export const Curve: React.FC<{
  bands?: {from: number; to: number; label: string; tone?: string}[];
  cliff?: boolean; mark?: number; d?: number; yLabel?: string; xLabel?: string;
  rising?: boolean; accel?: boolean; step?: number; stepLabel?: string;
}> = ({bands = [], cliff, mark, d = 0.35, yLabel = 'quality', xLabel = 'context used',
       rising, accel, step, stepLabel}) => {
  const f = useCurrentFrame();
  const W = 1020, H = 470, PAD = 54;
  const BASE = H - 98;            // axis line; the 98px below carries two label rows
  const TOP = 36;
  const x = (t: number) => PAD + t * (W - PAD * 2);
  // three shapes off one chart: the context decline, a frontier score that
  // steps, and a frontier score that steepens
  const q = (t: number) =>
      accel ? 0.05 + 0.90 * Math.pow(t, 2.6)
    : rising ? Math.min(0.96, 0.10 + 0.44 * t
        + 0.30 / (1 + Math.exp(-(t - (step ?? 0.55)) * 26)))
    : 1 - 0.86 * Math.pow(t, 2.1);
  const y = (v: number) => BASE - v * (BASE - TOP);
  const draw = easeOut(pr(f, d, 1.5));
  const pts = Array.from({length: 121}, (_, i) => {
    const t = i / 120;
    return `${x(t).toFixed(1)},${y(q(t)).toFixed(1)}`;
  }).join(' ');
  const mp = mark === undefined ? null : {mx: x(mark), my: y(q(mark))};
  const mA = easeSlam(pr(f, d + 1.35, 0.5));
  return (
    <div style={{width: '100%', maxWidth: 1150, margin: '0 auto'}}>
      <svg viewBox={`0 0 ${W} ${H}`} style={{width: '100%', display: 'block'}}>
        <defs>
          <clipPath id="curveClip">
            <rect x={0} y={0} width={PAD + draw * (W - PAD * 2)} height={H} />
          </clipPath>
        </defs>
        {bands.map((b, i) => {
          const a = easeIn(pr(f, d + 1.5 + i * 0.22, 0.5));
          return (
            <g key={i} opacity={a}>
              <rect x={x(b.from)} y={TOP} width={x(b.to) - x(b.from)}
                    height={BASE - TOP} fill={b.tone ?? C.yellow} opacity={0.08} />
              <text x={(x(b.from) + x(b.to)) / 2} y={BASE + 34}
                    textAnchor="middle" fill={b.tone ?? C.yellow}
                    fontFamily={MONO} fontSize={19} letterSpacing="0.08em">
                {b.label.toUpperCase()}
              </text>
            </g>
          );
        })}
        {/* axes */}
        <line x1={PAD} y1={BASE} x2={W - PAD} y2={BASE} stroke={C.line} strokeWidth={2} />
        <line x1={PAD} y1={TOP} x2={PAD} y2={BASE} stroke={C.line} strokeWidth={2} />
        <text x={PAD - 14} y={TOP - 10} textAnchor="start" fill={C.muted}
              fontFamily={MONO} fontSize={18} letterSpacing="0.14em">
          {yLabel.toUpperCase()}
        </text>
        <text x={W / 2} y={BASE + 74} textAnchor="middle" fill={C.muted}
              fontFamily={MONO} fontSize={18} letterSpacing="0.14em">
          {xLabel.toUpperCase()}
        </text>
        <polyline points={pts} fill="none" stroke={C.yellow} strokeWidth={5}
                  strokeLinecap="round" clipPath="url(#curveClip)" />
        {cliff ? (
          <g opacity={easeIn(pr(f, d + 1.6, 0.45))}>
            <line x1={W - PAD} y1={TOP} x2={W - PAD} y2={BASE}
                  stroke={C.red} strokeWidth={5} />
            <text x={W - PAD - 14} y={TOP + 30} textAnchor="end" fill={C.red}
                  fontFamily={DISPLAY} fontSize={26}>THE LIMIT</text>
          </g>
        ) : null}
        {rising && step !== undefined ? (
          <g opacity={easeIn(pr(f, d + 1.5, 0.5))}>
            <line x1={x(step)} y1={TOP} x2={x(step)} y2={BASE}
                  stroke={C.blue} strokeWidth={3} strokeDasharray="8 8" />
            <text x={x(step) - 14} y={TOP + 30} textAnchor="end" fill={C.blue}
                  fontFamily={DISPLAY} fontSize={26}>
              {(stepLabel ?? 'THE STEP').toUpperCase()}
            </text>
          </g>
        ) : null}
        {mp ? (
          <g opacity={Math.min(1, mA / 0.4)}>
            <circle cx={mp.mx} cy={mp.my} r={11 * Math.min(1.6, mA)}
                    fill={C.ink} stroke={C.bg} strokeWidth={4} />
          </g>
        ) : null}
      </svg>
    </div>
  );
};

/* ----------------------------------------------------------------- compact */
/** compaction: a history that fills the window, then folds into one summary
 *  block and hands the room back. The fold is the entire idea and a still
 *  frame of either state tells you nothing. */
export const Compact: React.FC<{
  blocks?: number; keep?: number; summary?: string; at?: number; d?: number;
  freedLabel?: string;
}> = ({blocks = 11, keep = 4, summary = 'summary of everything above',
       at = 2.4, d = 0.3, freedLabel = 'room recovered'}) => {
  const f = useCurrentFrame();
  const H = 470, GAPY = 6;
  const unit = (H - GAPY * (blocks - 1)) / blocks;
  const c = easeIn(pr(f, at, 0.95));
  const sumH = unit * 1.9;
  const old = blocks - keep;
  // after the fold: summary sits on top, the kept blocks follow it
  const keptTop = (i: number) => (sumH + GAPY) + (i - old) * (unit + GAPY);
  return (
    <div style={{width: '100%'}}>
      <div style={{
        position: 'relative', height: H, border: `2px solid ${C.line}`,
        borderRadius: 10, background: '#0d0d0d', padding: 0, overflow: 'hidden',
      }}>
        {/* the freed room, revealed underneath as the fold completes */}
        <div style={{
          position: 'absolute', left: 0, right: 0,
          top: keptTop(blocks - 1) + unit + GAPY, bottom: 0,
          border: `2px dashed ${C.blue}`, borderRadius: 8, margin: '0 2px 2px',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          fontFamily: MONO, fontSize: 20, letterSpacing: '0.14em',
          textTransform: 'uppercase', color: C.blue,
          opacity: easeIn(pr(f, at + 0.75, 0.5)),
        }}>{freedLabel}</div>
        {Array.from({length: blocks}, (_, i) => {
          const inP = easeIn(pr(f, d + i * 0.055, 0.4));
          const isOld = i < old;
          const y0 = i * (unit + GAPY);
          const y = isOld
            ? y0 * (1 - c)
            : y0 + (keptTop(i) - y0) * c;
          const h = isOld ? unit * (1 - c) : unit;
          return (
            <div key={i} style={{
              position: 'absolute', left: 2, right: 2, top: y, height: Math.max(0, h),
              borderRadius: 6, background: isOld ? '#1a1a1a' : '#242424',
              border: `1px solid ${isOld ? '#242424' : '#3a3a3a'}`,
              opacity: inP * (isOld ? 1 - c : 1),
            }} />
          );
        })}
        {/* the summary that replaces all of it */}
        <div style={{
          position: 'absolute', left: 2, right: 2, top: 0, height: sumH * c,
          borderRadius: 6, background: 'rgba(245,197,24,0.16)',
          border: `2px solid ${C.yellow}`,
          display: 'flex', alignItems: 'center', padding: '0 22px',
          fontFamily: MONO, fontSize: 21, color: C.yellow,
          opacity: c, overflow: 'hidden', whiteSpace: 'nowrap',
        }}>{summary}</div>
      </div>
    </div>
  );
};

/* ----------------------------------------------------------------- windows */
/** context window sizes across models, with the honest part drawn on: the bar
 *  is what you are sold, the lit portion is what stays reliable. Draws his
 *  own conclusion for him — do not shop on the big number. */
export const Windows: React.FC<{
  items: {t: string; tokens: number; note?: string}[];
  usable?: number; d?: number; unit?: string; usableLabel?: string;
}> = ({items, usable, d = 0.4, unit = 'tokens', usableLabel = 'stays reliable'}) => {
  const f = useCurrentFrame();
  const max = Math.max(...items.map((i) => i.tokens));
  const fmt = (n: number) =>
    n >= 1e6 ? `${(n / 1e6).toFixed(n % 1e6 ? 1 : 0)}M` : `${Math.round(n / 1e3)}k`;
  return (
    <div style={{width: '100%', display: 'flex', flexDirection: 'column', gap: 20}}>
      {items.map((it, i) => {
        const a = easeIn(pr(f, d + i * 0.2, 0.45));
        const g = easeOut(pr(f, d + 0.2 + i * 0.2, 1.0));
        const w = (it.tokens / max) * 100 * g;
        const uw = usable === undefined ? 0 : w * usable;
        return (
          <div key={i} style={{opacity: a}}>
            <div style={{
              display: 'flex', justifyContent: 'space-between', alignItems: 'baseline',
              marginBottom: 8,
            }}>
              <span style={{fontFamily: DISPLAY, fontSize: 27,
                            textTransform: 'uppercase', color: C.ink}}>{it.t}</span>
              <span style={{fontFamily: MONO, fontSize: 25, color: C.yellow,
                            fontVariantNumeric: 'tabular-nums'}}>
                {`${fmt(it.tokens * g)} ${unit}`}
              </span>
            </div>
            <div style={{position: 'relative', height: 40, background: '#131313',
                         border: '1px solid #262626', borderRadius: 6, overflow: 'hidden'}}>
              <div style={{position: 'absolute', inset: 0, width: `${w}%`,
                           background: usable === undefined ? C.yellow : '#3a3a3a'}} />
              {usable !== undefined ? (
                <div style={{position: 'absolute', top: 0, bottom: 0, left: 0,
                             width: `${uw}%`, background: C.yellow}} />
              ) : null}
            </div>
            {it.note ? (
              <div style={{marginTop: 7, fontFamily: MONO, fontSize: 18, color: C.muted}}>
                {it.note}
              </div>
            ) : null}
          </div>
        );
      })}
      {usable !== undefined ? (
        <div style={{
          display: 'flex', gap: 26, justifyContent: 'center', marginTop: 4,
          fontFamily: MONO, fontSize: 18, letterSpacing: '0.1em',
          textTransform: 'uppercase', color: C.muted,
          opacity: easeIn(pr(f, d + 1.5, 0.5)),
        }}>
          <span><span style={{color: C.yellow}}>■</span> {usableLabel}</span>
          <span><span style={{color: '#3a3a3a'}}>■</span> sold to you</span>
        </div>
      ) : null}
    </div>
  );
};

/* --------------------------------------------------------------- hierarchy */
/** the nested agents-file rule, which is the one idea in Day 2 a static slide
 *  genuinely cannot carry. The tree draws, the agent lands on a file deep
 *  inside it, and then the agents files light up in the order they are
 *  collected — innermost first, walking out to the root. The badge number is
 *  the merge order, so "inner overrides outer" is readable off the frame. */
export const Hierarchy: React.FC<{
  nodes: {t: string; depth: number; kind?: 'dir' | 'file' | 'agents'}[];
  active?: number; collect?: number[]; d?: number; step?: number;
}> = ({nodes, active, collect = [], d = 0.3, step = 0.55}) => {
  const f = useCurrentFrame();
  const treeDone = d + nodes.length * 0.07 + 0.4;
  const actA = active === undefined ? 0 : easeSlam(pr(f, treeDone, 0.45));
  return (
    <div style={{width: '100%', display: 'flex', justifyContent: 'center'}}>
      <div style={{
        minWidth: '72%', border: `2px solid ${C.line}`, borderRadius: 10,
        background: '#0d0d0d', padding: '22px 26px',
      }}>
        {nodes.map((n, i) => {
          const a = easeIn(pr(f, d + i * 0.07, 0.4));
          const ci = collect.indexOf(i);
          const lit = ci >= 0 ? easeSlam(pr(f, treeDone + 0.5 + ci * step, 0.45)) : 0;
          const isAct = i === active;
          const on = lit > 0.05;
          const col = n.kind === 'agents' ? (on ? C.yellow : C.dim)
                    : n.kind === 'dir' ? C.blue : C.muted;
          return (
            <div key={i} style={{
              display: 'flex', alignItems: 'center', gap: 14,
              padding: '8px 12px', marginLeft: n.depth * 40, borderRadius: 6,
              background: on ? 'rgba(245,197,24,0.13)'
                        : isAct && actA > 0.05 ? 'rgba(59,130,246,0.15)' : 'transparent',
              border: `1px solid ${on ? C.yellow
                        : isAct && actA > 0.05 ? C.blue : 'transparent'}`,
              opacity: a, transform: `translateX(${-18 * (1 - a)}px)`,
            }}>
              <span style={{fontFamily: MONO, fontSize: 26, color: col,
                            whiteSpace: 'pre'}}>
                {n.kind === 'dir' ? `${n.t}/` : n.t}
              </span>
              {isAct ? (
                <span style={{
                  marginLeft: 'auto', fontFamily: MONO, fontSize: 18,
                  letterSpacing: '0.12em', textTransform: 'uppercase',
                  color: C.blue, opacity: Math.min(1, actA / 0.4),
                }}>the agent is working here</span>
              ) : null}
              {on ? (
                <span style={{
                  marginLeft: isAct ? 14 : 'auto',
                  fontFamily: DISPLAY, fontSize: 19, color: '#111',
                  background: C.yellow, borderRadius: 99,
                  width: 30, height: 30, display: 'flex',
                  alignItems: 'center', justifyContent: 'center',
                  transform: `scale(${Math.min(1.4, lit)})`, flexShrink: 0,
                }}>{ci + 1}</span>
              ) : null}
            </div>
          );
        })}
      </div>
    </div>
  );
};

/* ------------------------------------------------------------------ editor */
/** a real file in an editor pane, with the lines under discussion lit. He shows
 *  this as a static image — the win here is that the highlight moves, so a
 *  seventy-line file can be walked section by section without a new slide for
 *  every section. Rendered, never filmed. */
export const Editor: React.FC<{
  file: string;
  lines: {t: string; kind?: 'h1' | 'h2' | 'bullet' | 'code' | 'text' | 'important'}[];
  focus?: [number, number]; d?: number; start?: number;
}> = ({file, lines, focus, d = 0.3, start = 1}) => {
  const f = useCurrentFrame();
  const fa = focus ? easeIn(pr(f, d + 0.9, 0.5)) : 0;
  const tone = (k?: string) =>
      k === 'h1' ? C.yellow : k === 'h2' ? C.blue
    : k === 'code' ? '#7fd88f' : k === 'important' ? C.red
    : k === 'bullet' ? C.ink : C.dim;
  return (
    <div style={{width: '100%', display: 'flex', justifyContent: 'center'}}>
      <div style={{
        width: '86%', border: `2px solid ${C.line}`, borderRadius: 10,
        background: '#0d0d0d', overflow: 'hidden',
      }}>
        <div style={{
          display: 'flex', alignItems: 'center', gap: 12, padding: '11px 18px',
          background: '#161616', borderBottom: `1px solid ${C.line}`,
          fontFamily: MONO, fontSize: 18, color: C.muted,
        }}>
          <span style={{width: 9, height: 9, borderRadius: 99, background: '#3a3a3a'}} />
          <span style={{color: C.ink}}>{file}</span>
        </div>
        <div style={{padding: '14px 0'}}>
          {lines.map((ln, i) => {
            const a = easeIn(pr(f, d + i * 0.035, 0.35));
            const inF = !!focus && i >= focus[0] && i <= focus[1];
            return (
              <div key={i} style={{
                display: 'flex', gap: 18, padding: '3px 20px',
                background: inF ? `rgba(245,197,24,${0.14 * fa})` : 'transparent',
                borderLeft: `3px solid ${inF ? `rgba(245,197,24,${fa})` : 'transparent'}`,
                opacity: focus ? a * (inF ? 1 : 1 - 0.55 * fa) : a,
              }}>
                <span style={{
                  fontFamily: MONO, fontSize: 17, color: '#3d3d3d',
                  minWidth: 34, textAlign: 'right', flexShrink: 0,
                }}>{start + i}</span>
                <span style={{
                  fontFamily: MONO, fontSize: ln.kind === 'h1' ? 25 : 21,
                  color: tone(ln.kind), whiteSpace: 'pre-wrap',
                  fontWeight: ln.kind === 'important' ? 700 : 400,
                }}>{ln.t}</span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

/* ------------------------------------------------------------------- loops */
/** the Ralph Loop, which is the best animation opportunity in Week 1 and
 *  completely inert as a still. The inner ring is what an agent already is —
 *  a model going round with tools. Then the outer ring draws itself around
 *  the whole thing and starts turning at its own, much slower rate, with a
 *  pass counter. "A loop wrapped in a loop" stops being a sentence and
 *  becomes something you can watch. */
export const Loops: React.FC<{
  inner: string[]; outer?: string[]; passes?: number; d?: number;
  innerRpm?: number; label?: string;
}> = ({inner, outer, passes = 10, d = 0.3, innerRpm = 26, label}) => {
  const f = useCurrentFrame();
  const t = Math.max(0, f / 30 - d);
  const CX = 500, CY = 348, R1 = 148, R2 = 292;
  const ringIn = easeIn(pr(f, d, 0.6));
  const outA = outer ? easeIn(pr(f, d + 1.5, 0.7)) : 0;
  // the inner loop spins fast; the outer one advances once per inner cycle
  const spin = t * (innerRpm / 60) * 360;
  const outerSpin = outer ? (t * (innerRpm / 60) / inner.length) * 360 : 0;
  const pass = outer
    ? Math.min(passes, 1 + Math.floor((t * (innerRpm / 60)) / inner.length))
    : 0;
  const at = (r: number, deg: number) => [
    CX + r * Math.cos((deg - 90) * Math.PI / 180),
    CY + r * Math.sin((deg - 90) * Math.PI / 180),
  ];
  const ring = (r: number, items: string[], rot: number, tone: string, a: number,
                fs: number) => (
    <g opacity={a}>
      <circle cx={CX} cy={CY} r={r} fill="none" stroke={tone} strokeWidth={2}
              strokeDasharray="7 9" opacity={0.5} />
      {items.map((s, i) => {
        const deg = (i / items.length) * 360;
        const [px, py] = at(r, deg);
        const [mx, my] = at(r, deg + rot);
        return (
          <g key={i}>
            <circle cx={px} cy={py} r={7} fill={tone} opacity={0.75} />
            <text x={px} y={py + (py < CY ? -20 : 30)} textAnchor="middle"
                  fill={tone} fontFamily={MONO} fontSize={fs}>{s}</text>
            {i === 0 ? <circle cx={mx} cy={my} r={12} fill={tone} /> : null}
          </g>
        );
      })}
    </g>
  );
  return (
    <div style={{width: '100%', maxWidth: 1180, margin: '0 auto'}}>
      <svg viewBox="0 0 1000 740" style={{width: '100%', display: 'block'}}>
        {outer ? ring(R2, outer, outerSpin, C.blue, outA, 21) : null}
        {ring(R1, inner, spin, C.yellow, ringIn, 20)}
        <text x={CX} y={CY - 4} textAnchor="middle" fill={C.ink}
              fontFamily={DISPLAY} fontSize={30} opacity={ringIn}>
          {outer ? 'THE AGENT' : 'THE AGENT'}
        </text>
        {outer ? (
          <text x={CX} y={CY + 34} textAnchor="middle" fill={C.blue}
                fontFamily={MONO} fontSize={24} opacity={outA}>
            {`pass ${pass} of ${passes}`}
          </text>
        ) : null}
        {label ? (
          <text x={CX} y={726} textAnchor="middle" fill={C.muted}
                fontFamily={MONO} fontSize={20} letterSpacing="0.14em"
                opacity={easeIn(pr(f, d + (outer ? 2.4 : 1.0), 0.6))}>
            {label.toUpperCase()}
          </text>
        ) : null}
      </svg>
    </div>
  );
};

/* ------------------------------------------------------------------- board */
/** a benchmark leaderboard, drawn rather than filmed. Filming the real site
 *  ages the frame the moment it renders and buries the actual point, which is
 *  how *close together* the top is. `spread` brackets the top rows and puts the
 *  delta on screen; `dateline` is mandatory in practice — every number here has
 *  a shelf life measured in weeks. */
export const Board: React.FC<{
  items: {t: string; score: number}[]; max?: number; d?: number;
  spread?: boolean; cost?: boolean; dateline?: string;
}> = ({items, max = 100, d = 0.35, spread, cost, dateline}) => {
  const f = useCurrentFrame();
  const hi = Math.max(...items.map((i) => i.score));
  const lo = Math.min(...items.map((i) => i.score));
  const pick = cost ? items.length - 2 : -1;
  return (
    <div style={{width: '100%', maxWidth: 1220, margin: '0 auto'}}>
      {items.map((it, i) => {
        const a = easeIn(pr(f, d + i * 0.18, 0.45));
        const g = easeOut(pr(f, d + 0.15 + i * 0.18, 0.9));
        const dim = cost && i !== pick;
        const on = i === pick;
        return (
          <div key={i} style={{
            display: 'flex', alignItems: 'center', gap: 20, marginBottom: 14,
            opacity: a * (dim ? 0.42 : 1),
          }}>
            <span style={{fontFamily: DISPLAY, fontSize: 22, color: C.muted,
                          width: 34}}>{String(i + 1).padStart(2, '0')}</span>
            <span style={{fontFamily: MONO, fontSize: 25,
                          color: on ? C.yellow : C.ink, width: '13.5em',
                          whiteSpace: 'nowrap', overflow: 'hidden'}}>{it.t}</span>
            <div style={{flex: 1, height: 36, background: '#131313',
                         border: '1px solid #262626', borderRadius: 5,
                         overflow: 'hidden'}}>
              <div style={{width: `${(it.score / max) * 100 * g}%`, height: '100%',
                           background: on ? C.yellow : dim ? '#3a3a3a' : C.yellow}} />
            </div>
            <span style={{fontFamily: MONO, fontSize: 27, width: '2.4em',
                          textAlign: 'right', color: on ? C.yellow : C.ink,
                          fontVariantNumeric: 'tabular-nums'}}>
              {Math.round(it.score * g)}
            </span>
          </div>
        );
      })}
      {spread ? (
        <div style={{
          marginTop: 18, textAlign: 'center', fontFamily: DISPLAY,
          fontSize: 32, textTransform: 'uppercase', color: C.blue,
          opacity: easeIn(pr(f, d + 1.5, 0.6)),
        }}>{`${hi} down to ${lo} — ${hi - lo} points across the whole top`}</div>
      ) : null}
      {cost ? (
        <div style={{
          marginTop: 18, textAlign: 'center', fontFamily: MONO, fontSize: 22,
          color: C.yellow, opacity: easeIn(pr(f, d + 1.5, 0.6)),
        }}>{'two points down the index is very often the right call'}</div>
      ) : null}
      {dateline ? (
        <div style={{
          marginTop: 22, textAlign: 'center', fontFamily: MONO, fontSize: 18,
          letterSpacing: '0.14em', textTransform: 'uppercase', color: C.muted,
          opacity: easeIn(pr(f, d + 1.8, 0.6)),
        }}>{dateline}</div>
      ) : null}
    </div>
  );
};

/* ----------------------------------------------------------------- matrix */
/** The payload of L21, and the reason it is a template and not a `cols` slide.
 *
 *  The day builds the same brief four times in four harnesses. A verdict slide
 *  that simply appears at the end asks the viewer to take the comparison on
 *  trust. This one fills in a column at a time, so by the time the fourth lands
 *  the grid is a record of what they just watched rather than a new claim.
 *
 *  `upto` is how many columns are in yet — the same slide is reused across the
 *  day with upto 1, 2, 3, 4 — and `win` crowns one column without hiding that
 *  the others are close, which is the honest reading of the result. */
export const Matrix: React.FC<{
  cols: {head: string; sub?: string}[];
  rows: {label: string; cells: string[]}[];
  upto?: number; win?: number; dateline?: string; d?: number;
}> = ({cols, rows, upto, win, dateline, d = 0.3}) => {
  const f = useCurrentFrame();
  const shown = upto ?? cols.length;
  const LABEL = 250;
  return (
    <div style={{width: '100%', maxWidth: 1560, margin: '0 auto'}}>
      {/* heads */}
      <div style={{display: 'flex', alignItems: 'flex-end', gap: 12,
                   borderBottom: `1px solid ${C.line}`, paddingBottom: 12}}>
        <div style={{width: LABEL, flexShrink: 0}} />
        {cols.map((c, i) => {
          const on = i < shown;
          const a = on ? easeIn(pr(f, d + i * 0.42, 0.5)) : 0;
          const crowned = win === i && on;
          return (
            <div key={i} style={{
              flex: 1, textAlign: 'center', opacity: on ? 0.25 + a * 0.75 : 0.12,
              transform: `translateY(${(1 - a) * 10}px)`,
            }}>
              <div style={{
                fontFamily: DISPLAY, fontSize: 27, textTransform: 'uppercase',
                letterSpacing: '0.02em',
                color: crowned ? C.yellow : on ? C.ink : C.muted,
              }}>{c.head}</div>
              {c.sub ? (
                <div style={{fontFamily: MONO, fontSize: 17, color: C.muted,
                             marginTop: 5}}>{c.sub}</div>
              ) : null}
            </div>
          );
        })}
      </div>
      {/* rows */}
      {rows.map((r, ri) => (
        <div key={ri} style={{
          display: 'flex', alignItems: 'center', gap: 12,
          borderBottom: ri === rows.length - 1 ? 'none' : `1px solid #1e1e1e`,
          minHeight: 62,
        }}>
          <div style={{
            width: LABEL, flexShrink: 0, fontFamily: DISPLAY, fontSize: 20,
            textTransform: 'uppercase', letterSpacing: '0.08em', color: C.muted,
            opacity: easeIn(pr(f, d + 0.1 + ri * 0.09, 0.45)),
          }}>{r.label}</div>
          {r.cells.map((cell, ci) => {
            const on = ci < shown;
            // a column arrives head-first, then its cells rain down the grid
            const a = on ? easeIn(pr(f, d + ci * 0.42 + 0.2 + ri * 0.07, 0.45)) : 0;
            const crowned = win === ci && on;
            return (
              <div key={ci} style={{
                flex: 1, textAlign: 'center', fontFamily: MONO, fontSize: 21,
                lineHeight: 1.32, padding: '10px 8px',
                color: crowned ? C.yellow : C.ink,
                opacity: on ? a : 0,
                background: crowned ? 'rgba(245,197,24,0.05)' : undefined,
              }}>{on ? cell : ''}</div>
            );
          })}
        </div>
      ))}
      {dateline ? (
        <div style={{
          marginTop: 24, textAlign: 'center', fontFamily: MONO, fontSize: 18,
          letterSpacing: '0.14em', textTransform: 'uppercase', color: C.muted,
          opacity: easeIn(pr(f, d + cols.length * 0.42 + 0.5, 0.6)),
        }}>{dateline}</div>
      ) : null}
    </div>
  );
};
