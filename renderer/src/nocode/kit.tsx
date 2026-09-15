// Shared slide kit for the no-code lecture films.
//
// Same visual language as the No-Code deck the client supplied: near-black
// ground, Archivo Black display type, IBM Plex Mono for everything else, and
// red / yellow / blue as the only accents. Every entrance is frame-driven so it
// renders deterministically.
//
// A lecture is a list of Slide specs plus the measured length of its narration
// clips. Deck lays them out and drops the voice under each one.

import React from 'react';
import {
  AbsoluteFill, Audio, OffthreadVideo, Sequence, interpolate, staticFile,
  useCurrentFrame,
} from 'remotion';
import {
  C, CueCtx, DISPLAY, FPS, MONO, Rise as MRise, easeIn, easeOut, easeSlam, pr,
  useCues, useJoin, useSlideSecs,
} from './motion';
import {Tokens, Probs, Chat, Flow, Timeline, Nest, Tree, Term, Meter,
        Gauge, Myth, Stepper, Occupancy, Stack, Curve, Compact,
        Windows, Hierarchy, Editor, Loops, Board, Matrix} from './anim';
import {BrowserWindow, BrowserProps} from './chrome';

export {C, DISPLAY, FPS, MONO};
export const Rise = MRise;




const Kicker: React.FC<{children: React.ReactNode; d?: number}> = ({children, d = 0}) => (
  <Rise d={d} style={{
    letterSpacing: '0.22em', textTransform: 'uppercase', fontSize: 16,
    color: C.muted, marginBottom: 20,
  }}>{children}</Rise>
);

/** headline lines; wrap a word in *stars* to accent it */
const Head: React.FC<{lines: string[]; size: number; accent?: string}> =
({lines, size, accent = C.yellow}) => {
  // A headline is the slide's own sentence, so its lines stay on the tight
  // cascade; only the body templates below take their timing from the voice.
  return (
    <div style={{
      fontFamily: DISPLAY, lineHeight: 0.92, letterSpacing: '-0.03em',
      textTransform: 'uppercase', fontSize: size,
    }}>
      {lines.map((ln, i) => (
        <Rise key={i} d={0.12 + i * 0.26}>
          {ln.split(/(\*[^*]+\*)/).map((bit, j) =>
            bit.startsWith('*') && bit.endsWith('*')
              ? <span key={j} style={{color: accent}}>{bit.slice(1, -1)}</span>
              : <span key={j}>{bit}</span>)}
        </Rise>
      ))}
    </div>
  );
};

const Stamp: React.FC<{children: React.ReactNode; color?: string; d?: number}> =
({children, color = C.yellow, d = 1.05}) => {
  const f = useCurrentFrame();
  const p = easeSlam(pr(f, d, 0.42));
  return (
    <span style={{
      fontFamily: DISPLAY, textTransform: 'uppercase', display: 'inline-block',
      padding: '11px 17px', border: `2px solid ${color}`, color, fontSize: 23,
      opacity: Math.min(1, p / 0.35),
      transform: `rotate(${interpolate(p, [0, 0.7, 1], [-22, -8, -6])}deg) `
               + `scale(${interpolate(p, [0, 0.7, 1], [1.5, 0.95, 1])})`,
    }}>{children}</span>
  );
};

const Row: React.FC<{
  n?: string; d: number; hot?: boolean; live?: boolean; children: React.ReactNode;
}> = ({n, d, hot, live, children}) => {
  const f = useCurrentFrame();
  const p = easeIn(pr(f, d, 0.5));
  // `hot` is the author's permanent emphasis. `live` is the row the voice is on
  // right now: it lifts as it arrives and settles back as the voice moves on,
  // so a list of six reads as six moments instead of one block of text.
  const l = live ? easeOut(pr(f, d, 0.75)) : 0;
  const lit = hot || live;
  return (
    <div style={{
      display: 'flex', alignItems: 'center', gap: 20, marginTop: 13,
      border: `1px solid ${hot ? C.yellow : live ? '#5a5a5a' : '#343434'}`,
      borderRadius: 8,
      padding: '15px 19px', fontSize: 26,
      color: lit ? C.ink : '#b2aea6',
      background: hot ? 'rgba(245,197,24,0.10)'
                : live ? 'rgba(244,241,234,0.045)' : 'transparent',
      opacity: p,
      transform: `translateX(${-24 * (1 - p) + 9 * l}px)`,
    }}>
      {n ? <span style={{fontFamily: DISPLAY, fontSize: 19, minWidth: 36,
                         color: lit ? C.yellow : C.muted}}>{n}</span> : null}
      <span>{children}</span>
    </div>
  );
};

/**
 * A chapter card for the stitched course modules (tools/nocode/modules).
 *
 * `rows` was the obvious fit and the wrong one: it bottom-anchors its list, so
 * a two-chapter module left the middle of the frame empty. This centres the
 * whole block instead, which reads the same whether the module has two chapters
 * or five, and marks the ones already watched - the card appears between every
 * lecture, so "where am I" is the only job it has.
 */
const Chapter: React.FC<{idx: number; items: string[]}> = ({idx, items}) => {
  const f = useCurrentFrame();
  const grow = easeIn(pr(f, 0.55, 0.75));
  const pct = ((idx + 1) / items.length) * 100;
  return (
    <div style={{
      flex: 1, width: '100%', display: 'flex', flexDirection: 'column',
      justifyContent: 'center',
    }}>
      <div style={{display: 'flex', alignItems: 'baseline', gap: 30}}>
        <Rise d={0.1}>
          <span style={{
            fontFamily: DISPLAY, fontSize: 128, lineHeight: 0.82,
            color: C.yellow, letterSpacing: '-0.04em',
          }}>{String(idx + 1).padStart(2, '0')}</span>
        </Rise>
        <Rise d={0.3}>
          <span style={{
            fontFamily: DISPLAY, fontSize: 50, textTransform: 'uppercase',
            letterSpacing: '-0.02em', lineHeight: 1.02,
          }}>{items[idx]}</span>
        </Rise>
      </div>

      <div style={{marginTop: 46, width: '100%'}}>
        {items.map((t, i) => {
          const p = easeIn(pr(f, 0.55 + i * 0.09, 0.5));
          const hot = i === idx;
          const done = i < idx;
          return (
            <div key={i} style={{
              display: 'flex', alignItems: 'center', gap: 20, marginTop: 12,
              border: `1px solid ${hot ? C.yellow : '#2e2e2e'}`, borderRadius: 8,
              padding: '13px 19px', fontSize: 25,
              color: hot ? C.ink : done ? '#6f6a63' : '#9d9890',
              background: hot ? 'rgba(245,197,24,0.10)' : 'transparent',
              opacity: p, transform: `translateX(${-22 * (1 - p)}px)`,
            }}>
              <span style={{
                fontFamily: DISPLAY, fontSize: 18, minWidth: 34,
                color: hot ? C.yellow : done ? '#565149' : C.muted,
              }}>{String(i + 1).padStart(2, '0')}</span>
              <span style={{flex: 1}}>{t}</span>
              <span style={{fontSize: 19, color: done ? '#565149' : 'transparent'}}>
                done
              </span>
            </div>
          );
        })}
      </div>

      <div style={{
        marginTop: 34, height: 3, width: '100%', background: '#242424',
      }}>
        <div style={{height: '100%', width: `${pct * grow}%`, background: C.yellow}} />
      </div>
    </div>
  );
};

/** a horizontal ladder of rungs: grows to `upto`, highlights `picks` */
const Ladder: React.FC<{items: string[]; pick?: number; picks?: number[];
                        upto?: number; tone?: string}> =
({items, pick, picks, upto, tone = C.yellow}) => {
  const f = useCurrentFrame();
  const hot = new Set(picks ?? (pick === undefined ? [] : [pick]));
  const shown = upto ?? items.length;
  return (
    <div style={{display: 'flex', gap: 12, alignItems: 'flex-end', height: '68%'}}>
      {items.map((s, i) => {
        const live = i < shown;
        // rungs already standing snap in at once; the newest one grows
        const p = live ? easeIn(pr(f, i === shown - 1 ? 0.3 : 0, 0.55)) : 1;
        const on = hot.has(i);
        const h = 34 + (i / Math.max(1, items.length - 1)) * 62;
        return (
          <div key={i} style={{
            flex: 1, height: `${h}%`,
            border: `2px solid ${on ? tone : live ? '#fff' : '#2a2a2a'}`,
            background: on ? tone : live ? '#141414' : 'transparent',
            color: on ? '#111' : live ? C.ink : '#3d3d3d',
            display: 'flex', flexDirection: 'column',
            justifyContent: 'flex-end', padding: '10px 8px',
            transform: `scaleY(${p})`, transformOrigin: 'bottom',
            fontFamily: MONO, fontSize: 19, lineHeight: 1.18,
            whiteSpace: 'pre-line', transition: 'none',
          }}>
            <span style={{fontFamily: DISPLAY, fontSize: 24, opacity: 0.85}}>
              {String(i + 1).padStart(2, '0')}
            </span>
            <span>{live ? s : ''}</span>
          </div>
        );
      })}
    </div>
  );
};

/** the program at a glance: three weeks of five days, filled in as we go */
const Grid: React.FC<{done: number; now?: number}> = ({done, now}) => {
  const f = useCurrentFrame();
  return (
    <div style={{display: 'flex', flexDirection: 'column', gap: 14, marginTop: 10}}>
      {[0, 1, 2].map((w) => (
        <div key={w} style={{display: 'flex', gap: 14, alignItems: 'center'}}>
          <Rise d={0.15 + w * 0.18} style={{
            width: 132, fontFamily: DISPLAY, fontSize: 21, textTransform: 'uppercase',
            color: C.muted, letterSpacing: '0.06em',
          }}>{`Week ${w + 1}`}</Rise>
          {[0, 1, 2, 3, 4].map((d) => {
            const i = w * 5 + d;
            const isNow = i === now;
            const filled = i < done;
            const p = easeIn(pr(f, 0.3 + w * 0.18 + d * 0.07, 0.45));
            const slam = isNow ? easeSlam(pr(f, 1.5, 0.5)) : 1;
            return (
              <div key={d} style={{
                flex: 1, height: 152, borderRadius: 10,
                border: `2px solid ${isNow ? C.yellow : filled ? '#4a4a4a' : '#262626'}`,
                background: isNow ? C.yellow : filled ? '#1c1c1c' : 'transparent',
                color: isNow ? '#111' : filled ? C.ink : '#3d3d3d',
                display: 'flex', flexDirection: 'column',
                alignItems: 'center', justifyContent: 'center', gap: 3,
                opacity: p, transform: `scale(${isNow ? 0.94 + 0.06 * slam : 1})`,
              }}>
                <span style={{fontFamily: DISPLAY, fontSize: 44}}>{`D${d + 1}`}</span>
                {filled || isNow ? (
                  <span style={{fontSize: 15, opacity: 0.8}}>done</span>
                ) : null}
              </div>
            );
          })}
        </div>
      ))}
    </div>
  );
};

/** the progress beat: a number that counts up and stops */
const Pct: React.FC<{pct: number; sub?: string}> = ({pct, sub}) => {
  const f = useCurrentFrame();
  const p = easeIn(pr(f, 0.35, 1.5));
  const n = Math.round(pct * p);
  return (
    <div style={{display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
      <div style={{
        fontFamily: DISPLAY, fontSize: 300, lineHeight: 0.86, color: C.yellow,
        letterSpacing: '-0.05em', fontVariantNumeric: 'tabular-nums',
      }}>{n}%</div>
      <div style={{
        width: '62%', height: 12, marginTop: 40, border: `2px solid #343434`,
        borderRadius: 99, overflow: 'hidden',
      }}>
        <div style={{
          width: `${pct * p}%`, height: '100%', background: C.yellow,
        }} />
      </div>
      {sub ? <Rise d={1.9} style={{
        marginTop: 34, letterSpacing: '0.2em', textTransform: 'uppercase',
        fontSize: 17, color: C.muted,
      }}>{sub}</Rise> : null}
    </div>
  );
};

const Cols: React.FC<{cols: {head: string; items: string[]}[]}> = ({cols}) => (
  <div style={{display: 'flex', gap: 26, marginTop: 6}}>
    {cols.map((c, i) => (
      <div key={i} style={{flex: 1}}>
        <Rise d={0.2 + i * 0.2} style={{
          fontFamily: DISPLAY, fontSize: 27, textTransform: 'uppercase',
          color: i === 0 ? C.yellow : i === 1 ? C.blue : C.red, marginBottom: 12,
        }}>{c.head}</Rise>
        {c.items.map((s, j) => (
          <Row key={j} d={0.45 + i * 0.2 + j * 0.16}>{s}</Row>
        ))}
      </div>
    ))}
  </div>
);

const Quote: React.FC<{text: string; who: string}> = ({text, who}) => (
  <>
    <Rise d={0.15} style={{
      fontFamily: DISPLAY, fontSize: 54, lineHeight: 1.05,
      letterSpacing: '-0.02em', maxWidth: '20ch',
    }}>“{text}”</Rise>
    <Rise d={0.9} style={{
      marginTop: 26, letterSpacing: '0.18em', textTransform: 'uppercase',
      fontSize: 17, color: C.yellow,
    }}>{who}</Rise>
  </>
);

/** real screen footage, with the slide's point stated over it and then got out
 *  of the way. The narration and the window length are unchanged, so swapping a
 *  slide for a shot costs no timing anywhere else in the film. */
const Shot: React.FC<{src: string; kicker?: string; lines?: string[]; size?: number}> =
({src, kicker, lines, size = 62}) => {
  const f = useCurrentFrame();
  const inP = easeIn(pr(f, 0.25, 0.6));
  const out = 1 - easeIn(pr(f, 5.0, 0.7));      // say it, then clear the frame
  const a = inP * out;
  const join = useJoin();
  // NO push on footage, deliberately. A slow zoom was tried and rejected: on a
  // screen recording it crops the edges, and the edges are where the thing
  // being talked about often is - a tab bar, a status line, the button he is
  // about to click. A tutorial screen has to be shown whole. If a shot sits
  // still because the screen itself is still, that is honest; the motion in
  // this cut comes from the footage and the cards, not from a fake camera.
  return (
    <AbsoluteFill style={{background: C.bg, opacity: join}}>
      {/* `contain`, not `cover`: the clip is already 1920x1080 so the two are
          identical here, but `contain` can never crop if a source ever differs.
          The grade is kept light - enough that a white IDE does not glare next
          to the near-black cards, not so much that screen text dims. */}
      <OffthreadVideo src={staticFile(src)} muted
        style={{width: '100%', height: '100%', objectFit: 'contain',
                filter: 'brightness(0.94) contrast(1.02)'}} />
      {lines ? (
        <AbsoluteFill style={{
          justifyContent: 'flex-end', padding: '5% 6%', opacity: a,
          background: `linear-gradient(to top, rgba(7,7,7,${0.92 * a}) 0%, `
                    + `rgba(7,7,7,${0.55 * a}) 22%, rgba(7,7,7,0) 46%)`,
        }}>
          <div style={{color: C.ink, fontFamily: MONO,
                       transform: `translateY(${16 * (1 - inP)}px)`}}>
            {kicker ? (
              <div style={{letterSpacing: '0.22em', textTransform: 'uppercase',
                           fontSize: 16, color: C.muted, marginBottom: 16}}>{kicker}</div>
            ) : null}
            <div style={{fontFamily: DISPLAY, lineHeight: 0.94,
                         letterSpacing: '-0.03em', textTransform: 'uppercase',
                         fontSize: size}}>
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
        </AbsoluteFill>
      ) : null}
    </AbsoluteFill>
  );
};

export type Slide = ({trans?: Trans}) & (
  | {k: 'head'; kicker?: string; lines: string[]; size?: number; accent?: string;
     stamp?: string; stampColor?: string}
  | {k: 'rows'; kicker?: string; lines: string[]; size?: number;
     rows: {t: string; hot?: boolean}[]; numbered?: boolean; stamp?: string}
  | {k: 'ladder'; kicker?: string; lines: string[]; size?: number;
     items: string[]; pick?: number;
     picks?: number[]; upto?: number; tone?: string}
  | {k: 'grid'; kicker?: string; lines: string[]; done: number; now?: number}
  | {k: 'pct'; kicker?: string; pct: number; sub?: string}
  | {k: 'cols'; kicker?: string; lines: string[]; size?: number;
     cols: {head: string; items: string[]}[]}
  | {k: 'chapter'; kicker?: string; idx: number; items: string[]}
  | {k: 'quote'; kicker?: string; text: string; who: string}
  | {k: 'shot'; src: string; kicker?: string; lines?: string[]; size?: number}
  | {k: 'browser'; kicker?: string; lines?: string[]; size?: number} & BrowserProps
  | {k: 'tokens'; kicker?: string; lines?: string[]; size?: number;
     chips: string[]; hot?: number[]; caption?: string}
  | {k: 'probs'; kicker?: string; lines?: string[]; size?: number;
     items: {t: string; p: number}[]; caption?: string}
  | {k: 'chat'; kicker?: string; lines?: string[]; size?: number;
     turns: {who: 'you' | 'model'; t: string}[]; resend?: number; caption?: string}
  | {k: 'flow'; kicker?: string; lines?: string[]; size?: number;
     nodes: {t: string; sub?: string; tone?: string}[]; loop?: boolean; caption?: string}
  | {k: 'timeline'; kicker?: string; lines?: string[]; size?: number;
     items: {when: string; t: string; sub?: string; hot?: boolean}[]; upto?: number;
     caption?: string}
  | {k: 'nest'; kicker?: string; lines?: string[]; size?: number;
     outer: string; inner: string; ring?: string[]; caption?: string}
  | {k: 'tree'; kicker?: string; lines?: string[]; size?: number;
     leaves: {t: string; keep?: boolean; note?: string}[]; caption?: string}
  | {k: 'term'; kicker?: string; lines?: string[]; size?: number;
     term: {t: string; kind?: 'cmd' | 'out' | 'ok' | 'warn'}[]; title?: string;
     prompt?: string}
  | {k: 'meter'; kicker?: string; lines?: string[]; size?: number;
     pct: number; fill: string; rest: string; tone?: string; caption?: string}
  | {k: 'gauge'; kicker?: string; lines?: string[]; size?: number;
     pct: number; limit: number; label: string; limitLabel?: string;
     over?: boolean; caption?: string}
  | {k: 'myth'; kicker?: string; lines?: string[]; size?: number;
     wrong: string; right: string; note?: string}
  | {k: 'stepper'; kicker?: string; lines?: string[]; size?: number;
     seed: string; steps: string[]; hold?: number; caption?: string}
  | {k: 'occupancy'; kicker?: string; lines?: string[]; size?: number;
     used: number; cols?: number; rows?: number; legend?: string; caption?: string}
  | {k: 'stack'; kicker?: string; lines?: string[]; size?: number;
     layers: {t: string; sub?: string; h: number; tone?: string}[];
     upto?: number; hot?: number[]; foot?: string; caption?: string}
  | {k: 'curve'; kicker?: string; lines?: string[]; size?: number;
     bands?: {from: number; to: number; label: string; tone?: string}[];
     cliff?: boolean; mark?: number; yLabel?: string; xLabel?: string;
     rising?: boolean; accel?: boolean; step?: number; stepLabel?: string;
     caption?: string}
  | {k: 'compact'; kicker?: string; lines?: string[]; size?: number;
     blocks?: number; keep?: number; summary?: string; at?: number;
     freedLabel?: string; caption?: string}
  | {k: 'windows'; kicker?: string; lines?: string[]; size?: number;
     items: {t: string; tokens: number; note?: string}[]; usable?: number;
     unit?: string; usableLabel?: string; caption?: string}
  | {k: 'hierarchy'; kicker?: string; lines?: string[]; size?: number;
     nodes: {t: string; depth: number; kind?: 'dir' | 'file' | 'agents'}[];
     active?: number; collect?: number[]; step?: number; caption?: string}
  | {k: 'editor'; kicker?: string; lines?: string[]; size?: number;
     file: string; start?: number;
     rows: {t: string; kind?: 'h1' | 'h2' | 'bullet' | 'code' | 'text' | 'important'}[];
     focus?: [number, number]; caption?: string}
  | {k: 'loops'; kicker?: string; lines?: string[]; size?: number;
     inner: string[]; outer?: string[]; passes?: number; innerRpm?: number;
     label?: string; caption?: string}
  | {k: 'board'; kicker?: string; lines?: string[]; size?: number;
     items: {t: string; score: number}[]; max?: number; spread?: boolean;
     cost?: boolean; dateline?: string; caption?: string}
  | {k: 'matrix'; kicker?: string; lines?: string[]; size?: number;
     cols: {head: string; sub?: string}[];
     rows: {label: string; cells: string[]}[];
     upto?: number; win?: number; dateline?: string; caption?: string});

/** how a slide arrives. `flash` is the original deck behaviour and stays the
 *  default so the shipped films are untouched; the others are for new decks. */
export type Trans = 'flash' | 'fade' | 'push' | 'rise' | 'none';

const Frame: React.FC<{
  children: React.ReactNode; center?: boolean; trans?: Trans;
}> = ({children, center, trans = 'flash'}) => {
  const f = useCurrentFrame();
  // The push-in used to finish at 8s and leave everything after that pixel-
  // identical, which is most of a long slide. Run it across the whole slide so
  // the frame is never completely still, and pair it with a slow counter-drift
  // so the movement reads as a camera rather than a zoom.
  //
  // The amount is deliberate. At the old 3% over the whole slide the move was
  // under a pixel between sampled frames, so it rasterised to identical frames
  // and the picture was still dead however long it ran - the measurement and
  // the eye agreed. This is ~40px of travel over a typical slide: a gentle
  // documentary push, well under a pixel per frame, but never the same frame
  // twice.
  // No zoom, on his instruction - a scale reads as a camera push and he does
  // not want one. What is left is a slow lateral drift: it keeps a card from
  // being pixel-identical for ten seconds (which is what made the first cut
  // feel broken) without changing the size of anything on screen, so nothing
  // is ever cropped or grows while you read it.
  const secs = useSlideSecs(8);
  const join = useJoin();
  const k = easeOut(pr(f, 0, secs));
  const ken = 1;
  const drift = interpolate(k, [0, 1], [-16, 16]);
  const flash = trans === 'flash' ? 0.5 * (1 - pr(f, 0, 0.26)) : 0;
  const t = easeIn(pr(f, 0, 0.5));
  // The entrance has to COMPOSE with the camera move, not replace it. Spreading
  // an `enter` that carried its own `transform` silently overwrote the push-in
  // on every 'push' and 'rise' slide - which is most of them - and left those
  // slides pixel-identical from the moment the entrance settled.
  const entering =
    trans === 'push' ? `translateX(${58 * (1 - t)}px)`
  : trans === 'rise' ? `translateY(${52 * (1 - t)}px)`
  : '';
  const enter = trans === 'push' || trans === 'rise' || trans === 'fade'
    ? {opacity: t} : {};
  return (
    <AbsoluteFill style={{background: C.bg, opacity: join}}>
      <AbsoluteFill style={{
        padding: '6.5% 7%', display: 'flex', flexDirection: 'column',
        transform: `${entering} scale(${ken}) `
                 + `translate3d(${drift}px, ${drift * 0.42}px, 0)`,
        color: C.ink, fontFamily: MONO,
        alignItems: center ? 'center' : undefined,
        textAlign: center ? 'center' : undefined,
        ...enter,
      }}>{children}</AbsoluteFill>
      <AbsoluteFill style={{background: '#fff', opacity: flash}} />
    </AbsoluteFill>
  );
};

export const SlideView: React.FC<{s: Slide}> = ({s}) => {
  // every panel template shares one frame: kicker, headline, the thing, a note
  if (s.k === 'tokens' || s.k === 'probs' || s.k === 'chat' || s.k === 'flow'
   || s.k === 'timeline' || s.k === 'nest' || s.k === 'tree' || s.k === 'term'
   || s.k === 'meter' || s.k === 'browser' || s.k === 'gauge' || s.k === 'myth'
   || s.k === 'stepper' || s.k === 'occupancy' || s.k === 'stack'
   || s.k === 'curve' || s.k === 'compact' || s.k === 'windows'
   || s.k === 'hierarchy' || s.k === 'editor' || s.k === 'loops'
   || s.k === 'board' || s.k === 'matrix') {
    const body =
      s.k === 'tokens' ? <Tokens chips={s.chips} hot={s.hot} />
    : s.k === 'probs' ? <Probs items={s.items} />
    : s.k === 'chat' ? <Chat turns={s.turns} resend={s.resend} />
    : s.k === 'flow' ? <Flow nodes={s.nodes} loop={s.loop} />
    : s.k === 'timeline' ? <Timeline items={s.items} upto={s.upto} />
    : s.k === 'nest' ? <Nest outer={s.outer} inner={s.inner} ring={s.ring} />
    : s.k === 'tree' ? <Tree leaves={s.leaves} />
    : s.k === 'term' ? <Term lines={s.term} title={s.title} prompt={s.prompt} />
    : s.k === 'meter' ? <Meter pct={s.pct} fill={s.fill} rest={s.rest} tone={s.tone} />
    : s.k === 'gauge' ? <Gauge pct={s.pct} limit={s.limit} label={s.label}
        limitLabel={s.limitLabel} over={s.over} />
    : s.k === 'myth' ? <Myth wrong={s.wrong} right={s.right} note={s.note} />
    : s.k === 'stepper' ? <Stepper seed={s.seed} steps={s.steps} hold={s.hold} />
    : s.k === 'occupancy' ? <Occupancy used={s.used} cols={s.cols} rows={s.rows}
        legend={s.legend} />
    : s.k === 'stack' ? <Stack layers={s.layers} upto={s.upto} hot={s.hot}
        foot={s.foot} />
    : s.k === 'curve' ? <Curve bands={s.bands} cliff={s.cliff} mark={s.mark}
        yLabel={s.yLabel} xLabel={s.xLabel} rising={s.rising} accel={s.accel}
        step={s.step} stepLabel={s.stepLabel} />
    : s.k === 'compact' ? <Compact blocks={s.blocks} keep={s.keep}
        summary={s.summary} at={s.at} freedLabel={s.freedLabel} />
    : s.k === 'windows' ? <Windows items={s.items} usable={s.usable}
        unit={s.unit} usableLabel={s.usableLabel} />
    : s.k === 'hierarchy' ? <Hierarchy nodes={s.nodes} active={s.active}
        collect={s.collect} step={s.step} />
    : s.k === 'editor' ? <Editor file={s.file} lines={s.rows} focus={s.focus}
        start={s.start} />
    : s.k === 'loops' ? <Loops inner={s.inner} outer={s.outer} passes={s.passes}
        innerRpm={s.innerRpm} label={s.label} />
    : s.k === 'board' ? <Board items={s.items} max={s.max} spread={s.spread}
        cost={s.cost} dateline={s.dateline} />
    : s.k === 'matrix' ? <Matrix cols={s.cols} rows={s.rows} upto={s.upto}
        win={s.win} dateline={s.dateline} />
    : <BrowserWindow {...(s as any)} />;
    const caption = (s as any).caption as string | undefined;
    const wide = s.k === 'timeline' || s.k === 'myth' || s.k === 'stack'
               || s.k === 'curve' || s.k === 'compact' || s.k === 'windows'
               || s.k === 'hierarchy' || s.k === 'editor' || s.k === 'term'
               || s.k === 'loops' || s.k === 'board' || s.k === 'matrix';
    const tr = (s as any).trans as Trans | undefined;
    return (
      <Frame center={!wide} trans={tr}>
        {/* Headline and panel travel together as one centred block. Separated
            by a flex spacer the title floated at the top of the frame with the
            diagram stranded in the middle, which is what made these read as
            slides rather than as a composed frame. */}
        <div style={{flex: 0.86}} />
        {s.kicker ? <Kicker>{s.kicker}</Kicker> : null}
        {s.lines ? <Head lines={s.lines} size={s.size ?? 58} /> : null}
        <div style={{height: 32}} />
        <div style={{width: '100%', display: 'flex', justifyContent: 'center'}}>{body}</div>
        {caption ? (
          <Rise d={2.4} style={{
            marginTop: 30, fontSize: 25, color: C.dim, textAlign: 'center',
          }}>{caption}</Rise>
        ) : null}
        <div style={{flex: 1}} />
      </Frame>
    );
  }
  if (s.k === 'shot') return (
    <Shot src={s.src} kicker={s.kicker} lines={s.lines} size={s.size} />
  );
  if (s.k === 'head') return (
    // Centred as a block. The stamp used to be pushed to the very bottom of the
    // frame by a flex spacer, which left a statement card reading as a line of
    // type at the top and a sticker at the floor with nothing between them.
    <Frame center trans={(s as any).trans}>
      <div style={{flex: 0.86}} />
      {s.kicker ? <Kicker>{s.kicker}</Kicker> : null}
      <Head lines={s.lines} size={s.size ?? 96} accent={s.accent} />
      {s.stamp ? <div style={{marginTop: 40}}>
        <Stamp color={s.stampColor ?? C.yellow} d={1.15}>{s.stamp}</Stamp>
      </div> : null}
      <div style={{flex: 1}} />
    </Frame>
  );
  if (s.k === 'chapter') return (
    <Frame trans={(s as any).trans}>
      {s.kicker ? <Kicker>{s.kicker}</Kicker> : null}
      <Chapter idx={s.idx} items={s.items} />
    </Frame>
  );
  if (s.k === 'quote') return (
    <Frame trans={(s as any).trans}>
      {s.kicker ? <Kicker>{s.kicker}</Kicker> : null}
      <Quote text={s.text} who={s.who} />
    </Frame>
  );
  if (s.k === 'ladder') return (
    <Frame trans={(s as any).trans}>
      {s.kicker ? <Kicker>{s.kicker}</Kicker> : null}
      <Head lines={s.lines} size={s.size ?? 64} />
      <div style={{flex: 1}} />
      <Ladder items={s.items} pick={s.pick} picks={s.picks} upto={s.upto} tone={s.tone} />
    </Frame>
  );
  if (s.k === 'grid') return (
    <Frame trans={(s as any).trans}>
      {s.kicker ? <Kicker>{s.kicker}</Kicker> : null}
      <Head lines={s.lines} size={58} />
      <div style={{flex: 1}} />
      <Grid done={s.done} now={s.now} />
      <div style={{flex: 1}} />
    </Frame>
  );
  if (s.k === 'pct') return (
    <Frame center trans={(s as any).trans}>
      {s.kicker ? <Kicker>{s.kicker}</Kicker> : null}
      <div style={{flex: 1}} />
      <Pct pct={s.pct} sub={s.sub} />
      <div style={{flex: 1}} />
    </Frame>
  );
  if (s.k === 'cols') return (
    <Frame trans={(s as any).trans}>
      {s.kicker ? <Kicker>{s.kicker}</Kicker> : null}
      <Head lines={s.lines} size={s.size ?? 58} />
      <div style={{height: 26}} />
      <Cols cols={s.cols} />
    </Frame>
  );
  return <RowsSlide s={s} />;
};

/** The list slide. Its rows used to cascade in over 1.65s and then hold the
 *  same frame for the rest of a seventeen-second slide; now each row lands on
 *  the sentence that says it, and the row being spoken is the lit one. */
const RowsSlide: React.FC<{s: Extract<Slide, {k: 'rows'}>}> = ({s}) => {
  const cue = useCues(0.55, 0.22);
  const f = useCurrentFrame();
  const now = f / FPS;
  // The last row whose cue has passed — that is the one being talked about.
  let live = -1;
  s.rows.forEach((_, i) => {if (now >= cue(i, s.rows.length)) live = i;});
  return (
    <Frame trans={(s as any).trans}>
      {/* Headline at the top and the list pinned to the bottom leaves the
          middle of the frame empty, which on a card that holds for eight
          seconds reads as an unfinished slide. The block is optically centred
          instead - a little above true centre, which is where the eye expects
          a title block to sit. */}
      <div style={{flex: 0.86}} />
      {s.kicker ? <Kicker>{s.kicker}</Kicker> : null}
      <Head lines={s.lines} size={s.size ?? 62} />
      <div style={{height: 34}} />
      <div>
        {s.rows.map((r, i) => (
          <Row key={i} n={s.numbered ? String(i + 1).padStart(2, '0') : undefined}
               d={cue(i, s.rows.length)} hot={r.hot ?? false}
               live={i === live}>{r.t}</Row>
        ))}
      </div>
      {s.stamp ? <div style={{marginTop: 26}}>
        <Stamp d={cue(s.rows.length - 1, s.rows.length) + 0.9}>{s.stamp}</Stamp>
      </div> : null}
      <div style={{flex: 1}} />
    </Frame>
  );
};

/** one lecture: slides paired with the measured length of each narration clip */
export const Deck: React.FC<{
  slides: Slide[]; durs: number[]; voDir: string; files: string[]; gap?: number;
  /** per slide, the second each spoken sentence starts (adam_deck cues.json).
   *  Omitted, every slide falls back to the original fixed cascade, so decks
   *  that shipped before this existed render exactly as they shipped. */
  cues?: number[][];
}> = ({slides, durs, voDir, files, gap = 0.3, cues}) => {
  let at = 0;
  const out: React.ReactNode[] = [];
  slides.forEach((s, i) => {
    const len = Math.round((durs[i] + (i < slides.length - 1 ? gap : 0)) * FPS);
    out.push(
      <Sequence key={i} from={at} durationInFrames={len}>
        <CueCtx.Provider value={{
          cues: cues?.[i], secs: durs[i],
          // A dip only where the KIND changes. Two consecutive shots are one
          // continuous piece of footage cut to two narration lengths - dipping
          // between them would invent a cut that is not there.
          joinIn: i > 0 && (slides[i - 1].k === 'shot') !== (s.k === 'shot'),
          joinOut: i < slides.length - 1
            && (slides[i + 1].k === 'shot') !== (s.k === 'shot'),
        }}>
          <SlideView s={s} />
        </CueCtx.Provider>
        <Audio src={staticFile(`${voDir}/${files[i]}`)} />
      </Sequence>,
    );
    at += len;
  });
  return <AbsoluteFill style={{background: C.bg}}>{out}</AbsoluteFill>;
};

export const deckFrames = (durs: number[], gap = 0.3) =>
  durs.reduce((a, d, i) => a + Math.round((d + (i < durs.length - 1 ? gap : 0)) * FPS), 0);
