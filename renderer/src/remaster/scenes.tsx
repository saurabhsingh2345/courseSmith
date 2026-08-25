// Graphic scenes for the remaster.
//
// Every one of these replaces a slide from the original cut that was a single
// line of text and a flat clip-art icon held for six to thirty seconds. The rule
// here is that a slide has to do something a sentence could not: show a
// comparison, take something apart, count, or reveal in an order that matches
// the argument.

import React from 'react';
import {AbsoluteFill, useCurrentFrame, useVideoConfig, spring} from 'remotion';
import {
  C,
  DISPLAY,
  BODY,
  MONO,
  Headline,
  Body,
  Eyebrow,
  Rule,
  Card,
  Halo,
  Motes,
  Glyph,
  ramp,
  useEnter,
  LIGHT,
} from './kit';

const PAD = 132;

/**
 * Which item the narration is on right now.
 *
 * `sweep` holds one frame number per item — the moment its phrase is spoken.
 * An item is live from its own mark until the next one's, so the highlight
 * walks the row in step with the voice instead of all of them sitting lit.
 */
const useSweep = (sweep: number[] | undefined, frame: number, fallback: number) => {
  if (!sweep || sweep.length === 0) return fallback;
  let i = -1;
  for (let k = 0; k < sweep.length; k += 1) if (frame >= sweep[k]) i = k;
  return i;
};

/** Spring that fires when an item becomes live, for the pop on highlight. */
const usePop = (at: number | undefined, frame: number, fps: number) =>
  at === undefined || frame < at
    ? 0
    : spring({frame: frame - at, fps, config: {damping: 13, stiffness: 190, mass: 0.6}});



type Icon = React.ComponentProps<typeof Glyph>['name'];

/** An icon sitting inside its own glow, centred on itself. */
const Badge: React.FC<{icon: Icon; size?: number; on?: number; dim?: boolean}> = ({
  icon,
  size = 76,
  on = 1,
  dim = false,
}) => (
  <div
    style={{
      position: 'relative',
      width: size * 1.9,
      height: size * 1.9,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
    }}
  >
    {!dim ? <Halo size={size * 2.6} opacity={0.26} /> : null}
    <div style={{position: 'relative', lineHeight: 0}}>
      <Glyph name={icon} size={size} on={on} color={dim ? C.dim : C.brand} />
    </div>
  </div>
);

// ---------------------------------------------------------------------------

export const Title: React.FC<{title: string; kicker: string; sub: string}> = ({
  title,
  kicker,
  sub,
}) => {
  const frame = useCurrentFrame();
  const line = ramp(frame, 26, 26);
  return (
    <AbsoluteFill>
      <Motes n={34} seed="title" />
      <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 34}}>
        <Eyebrow delay={2}>{kicker}</Eyebrow>
        <Headline
          text={title}
          size={132}
          delay={8}
          accent={['without', 'writing', 'code']}
          width={1500}
        />
        <div style={{width: 1400 * line, height: 3, backgroundColor: C.line, marginTop: 10}} />
        <Body delay={34} size={38} width={1080}>
          {sub}
        </Body>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};

export const Chapter: React.FC<{n: string; title: string}> = ({n, title}) => {
  const frame = useCurrentFrame();
  const wipe = ramp(frame, 4, 22);
  const num = ramp(frame, 0, 18);
  return (
    <AbsoluteFill>
      <Motes n={18} seed={n} />
      <AbsoluteFill style={{padding: PAD, justifyContent: 'center'}}>
        <div style={{display: 'flex', alignItems: 'baseline', gap: 46}}>
          <div
            style={{
              fontFamily: DISPLAY,
              fontWeight: 700,
              fontSize: 230,
              lineHeight: 0.85,
              color: 'transparent',
              WebkitTextStroke: `2px ${C.brand}`,
              opacity: num,
              transform: `translateX(${(1 - num) * -40}px)`,
            }}
          >
            {n}
          </div>
          <div style={{flex: 1}}>
            <div
              style={{height: 3, backgroundColor: C.brand, width: `${wipe * 100}%`, marginBottom: 30}}
            />
            <Headline text={title} size={78} delay={12} width={1200} />
          </div>
        </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------

export const Statement: React.FC<{
  kicker?: string;
  text: string;
  accent?: string[];
  body?: string;
  size?: number;
}> = ({kicker, text, accent = [], body, size = 96}) => (
  <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 30}}>
    {kicker ? <Eyebrow>{kicker}</Eyebrow> : null}
    <Rule delay={4} />
    <Headline text={text} size={size} delay={10} accent={accent} width={1560} />
    {body ? (
      <Body delay={26} width={1180}>
        {body}
      </Body>
    ) : null}
  </AbsoluteFill>
);

// ---------------------------------------------------------------------------
// Swap — the core loop: words in, code out, you judge.

export const Swap: React.FC<{panels: {icon: Icon; who: string; what: string}[]}> = ({panels}) => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{padding: PAD, justifyContent: 'center'}}>
      <div style={{display: 'flex', alignItems: 'stretch', gap: 26}}>
        {panels.map((p, i) => {
          const d = i * 16;
          const on = ramp(frame, d + 8, 18);
          const arrow = ramp(frame, d + 22, 14);
          return (
            <React.Fragment key={i}>
              <Card delay={d} active={i === 0} style={{flex: 1, padding: '46px 42px 50px'}}>
                <div style={{marginLeft: -20, marginBottom: 14}}>
                  <Badge icon={p.icon} size={72} on={on} />
                </div>
                <div
                  style={{
                    fontFamily: MONO,
                    fontSize: 20,
                    letterSpacing: '0.2em',
                    textTransform: 'uppercase',
                    color: C.brandSoft,
                    marginBottom: 14,
                  }}
                >
                  {p.who}
                </div>
                <div
                  style={{
                    fontFamily: DISPLAY,
                    fontWeight: 600,
                    fontSize: 46,
                    lineHeight: 1.15,
                    letterSpacing: '-0.02em',
                    color: C.cream,
                  }}
                >
                  {p.what}
                </div>
              </Card>
              {i < panels.length - 1 ? (
                <div
                  style={{
                    alignSelf: 'center',
                    fontSize: 46,
                    color: C.brand,
                    opacity: arrow,
                    transform: `translateX(${(1 - arrow) * -14}px)`,
                  }}
                >
                  →
                </div>
              ) : null}
            </React.Fragment>
          );
        })}
      </div>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------

export const Cards: React.FC<{
  title: string;
  accent?: string[];
  items: {name: string; note: string; icon?: Icon}[];
  active?: number;
  footer?: string;
  sweepAt?: string[];
  sweep?: number[];
}> = ({title, accent = [], items, active = -1, footer, sweep}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const foot = ramp(frame, 48, 18);
  const wide = items.length <= 2;
  const live = useSweep(sweep, frame, active);
  const swept = Boolean(sweep && sweep.length);
  return (
    <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 54}}>
      <Headline text={title} size={62} delay={0} accent={accent} align="center" width="100%" />
      <div style={{display: 'flex', gap: 26, justifyContent: 'center'}}>
        {items.map((it, i) => {
          const isOn = live === i;
          // In a sweep the cards it has already passed go back to normal rather
          // than staying dimmed — only a fixed `active` greys the rest out.
          const faded = swept ? false : active >= 0 && !isOn;
          const on = ramp(frame, 12 + i * 7, 16);
          const pop = usePop(swept ? sweep![i] : undefined, frame, fps);
          const lift = isOn ? pop : 0;
          return (
            <Card
              key={i}
              delay={10 + i * 7}
              active={isOn}
              style={{
                flex: 1,
                maxWidth: wide ? 520 : 400,
                padding: wide ? '52px 40px 46px' : '40px 30px 38px',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                textAlign: 'center',
                position: 'relative',
                zIndex: isOn ? 2 : 1,
                transition: 'none',
                marginTop: -lift * 14,
                marginBottom: lift * 14,
                boxShadow: isOn
                  ? `0 ${26 + lift * 20}px ${60 + lift * 30}px ${LIGHT ? 'rgba(23,27,33,0.20)' : 'rgba(0,0,0,0.6)'}, 0 0 0 1px ${C.brand}${LIGHT ? '77' : '55'}, 0 0 ${44 * lift}px ${C.brand}${LIGHT ? '24' : '44'}`
                  : `0 22px 54px ${LIGHT ? 'rgba(23,27,33,0.14)' : 'rgba(0,0,0,0.45)'}`,
              }}
            >
              {isOn ? (
                <div
                  style={{
                    position: 'absolute',
                    left: 0,
                    right: 0,
                    top: 0,
                    height: 3,
                    borderRadius: 3,
                    background: `linear-gradient(90deg, transparent, ${C.brand}, transparent)`,
                    opacity: pop,
                    transform: `scaleX(${0.3 + pop * 0.7})`,
                  }}
                />
              ) : null}
              <Badge
                icon={it.icon ?? 'web'}
                size={wide ? 84 : 72}
                on={on}
                dim={faded || (swept && !isOn)}
              />
              <div
                style={{
                  fontFamily: DISPLAY,
                  fontWeight: 650,
                  fontSize: wide ? 42 : 37,
                  color: faded ? C.dim : isOn || !swept ? C.cream : C.muted,
                  margin: '10px 0 12px',
                  letterSpacing: '-0.01em',
                }}
              >
                {it.name}
              </div>
              <div
                style={{
                  fontFamily: BODY,
                  fontSize: wide ? 26 : 24,
                  color: faded ? (LIGHT ? '#AEB4BA' : '#5A524D') : C.muted,
                  lineHeight: 1.4,
                }}
              >
                {it.note}
              </div>
            </Card>
          );
        })}
      </div>
      {footer ? (
        <div
          style={{
            textAlign: 'center',
            fontFamily: BODY,
            fontSize: 32,
            color: C.brandSoft,
            opacity: foot,
            transform: `translateY(${(1 - foot) * 12}px)`,
          }}
        >
          {footer}
        </div>
      ) : null}
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------
// Steps — numbered nodes on one rule, labels hung underneath.

export const Steps: React.FC<{
  title: string;
  steps: {name: string; note: string}[];
  sweepAt?: string[];
  sweep?: number[];
}> = ({title, steps, sweep}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const n = steps.length;
  const live = useSweep(sweep, frame, -1);
  // With a sweep the rail advances with the voice; without one it just draws in.
  const rail = sweep && sweep.length
    ? Math.max(0, live) / Math.max(1, n - 1)
    : ramp(frame, 12, 30);
  return (
    <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 84}}>
      <Headline text={title} size={62} align="center" width="100%" />
      <div style={{position: 'relative'}}>
        {/* the rule runs between the first and last node centres */}
        <div
          style={{
            position: 'absolute',
            top: 48,
            left: `${100 / (n * 2)}%`,
            width: `${100 - 100 / n}%`,
            height: 2,
            backgroundColor: C.line,
          }}
        />
        <div
          style={{
            position: 'absolute',
            top: 48,
            left: `${100 / (n * 2)}%`,
            width: `${(100 - 100 / n) * rail}%`,
            height: 2,
            backgroundColor: C.brand,
          }}
        />
        <div style={{display: 'flex'}}>
          {steps.map((st, i) => {
            const e = useEnter(12 + i * 14);
            const isOn = live === i;
            const pop = usePop(sweep ? sweep[i] : undefined, frame, fps);
            const done = live > i;
            return (
              <div
                key={i}
                style={{
                  flex: 1,
                  textAlign: 'center',
                  opacity: e,
                  transform: `translateY(${(1 - e) * 22}px)`,
                }}
              >
                <div
                  style={{
                    width: 96,
                    height: 96,
                    margin: '0 auto 26px',
                    borderRadius: '50%',
                    border: `2px solid ${isOn || done ? C.brand : C.line}`,
                    backgroundColor: isOn ? (LIGHT ? '#F6E7D4' : '#3A2716') : (LIGHT ? C.ink2 : '#241C17'),
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontFamily: DISPLAY,
                    fontWeight: 700,
                    fontSize: 40,
                    color: isOn || done ? C.brand : C.dim,
                    boxShadow: isOn
                      ? `0 0 ${30 + pop * 40}px ${C.brand}77`
                      : `0 0 44px ${C.brand}22`,
                    position: 'relative',
                    zIndex: 1,
                    transform: `scale(${1 + (isOn ? pop * 0.12 : 0)})`,
                  }}
                >
                  {done ? '✓' : i + 1}
                </div>
                <div
                  style={{
                    fontFamily: DISPLAY,
                    fontWeight: 620,
                    fontSize: 40,
                    color: isOn || done || live < 0 ? C.cream : C.muted,
                    marginBottom: 10,
                  }}
                >
                  {st.name}
                </div>
                <div style={{fontFamily: BODY, fontSize: 25, color: C.muted}}>{st.note}</div>
              </div>
            );
          })}
        </div>
      </div>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------

export const Spotlight: React.FC<{
  title: string;
  accent?: string[];
  name: string;
  tagline: string;
  icon: Icon;
  facts: string[];
}> = ({title, accent = [], name, tagline, icon, facts}) => {
  const frame = useCurrentFrame();
  const on = ramp(frame, 12, 20);
  return (
    <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 50}}>
      <Headline text={title} size={58} accent={accent} width={1500} />
      <div style={{display: 'flex', gap: 44, alignItems: 'center'}}>
        <Card
          delay={8}
          active
          style={{
            width: 440,
            padding: '46px 40px 44px',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            textAlign: 'center',
          }}
        >
          <Badge icon={icon} size={104} on={on} />
          <div
            style={{fontFamily: DISPLAY, fontWeight: 700, fontSize: 50, color: C.cream, marginTop: 8}}
          >
            {name}
          </div>
          <div style={{fontFamily: BODY, fontSize: 25, color: C.muted, marginTop: 12}}>
            {tagline}
          </div>
        </Card>
        <div style={{flex: 1, display: 'flex', flexDirection: 'column', gap: 20}}>
          {facts.map((f, i) => {
            const e = useEnter(22 + i * 12);
            return (
              <div
                key={i}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 24,
                  padding: '28px 34px',
                  borderRadius: 18,
                  backgroundColor: C.ink1,
                  border: `1px solid ${C.line}`,
                  opacity: e,
                  transform: `translateX(${(1 - e) * 40}px)`,
                }}
              >
                <div
                  style={{
                    width: 12,
                    height: 12,
                    borderRadius: 6,
                    backgroundColor: C.brand,
                    flexShrink: 0,
                    boxShadow: `0 0 20px ${C.brand}`,
                  }}
                />
                <div style={{fontFamily: BODY, fontSize: 34, color: C.cream}}>{f}</div>
              </div>
            );
          })}
        </div>
      </div>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------

export const Define: React.FC<{word: string; pos: string; meaning: string; also: string}> = ({
  word,
  pos,
  meaning,
  also,
}) => {
  const frame = useCurrentFrame();
  const under = ramp(frame, 18, 22);
  return (
    <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 22}}>
      <Eyebrow>the word for it</Eyebrow>
      <div style={{display: 'flex', alignItems: 'baseline', gap: 28}}>
        <Headline text={word} size={158} delay={6} accent={[word]} />
        <div
          style={{
            fontFamily: BODY,
            fontStyle: 'italic',
            fontSize: 40,
            color: C.dim,
            opacity: ramp(frame, 16, 14),
          }}
        >
          {pos}
        </div>
      </div>
      <div style={{width: 760 * under, height: 5, backgroundColor: C.brand, borderRadius: 3}} />
      <Body delay={26} size={44} width={1400}>
        {meaning}
      </Body>
      <div
        style={{
          marginTop: 22,
          fontFamily: MONO,
          fontSize: 26,
          color: C.dim,
          opacity: ramp(frame, 44, 16),
        }}
      >
        {also}
      </div>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------

export const Scale: React.FC<{
  title: string;
  rungs: string[];
  pick: number;
  left: string;
  right: string;
  note: string;
}> = ({title, rungs, pick, left, right, note}) => {
  const frame = useCurrentFrame();
  const bar = ramp(frame, 8, 26);
  const n = rungs.length;
  return (
    <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 62}}>
      <Headline text={title} size={62} align="center" width="100%" />
      <div style={{position: 'relative', height: 190}}>
        <div
          style={{
            position: 'absolute',
            top: 62,
            left: `${100 / (n * 2)}%`,
            width: `${(100 - 100 / n) * bar}%`,
            height: 6,
            background: `linear-gradient(90deg, ${C.dim}, ${C.brand})`,
            borderRadius: 3,
          }}
        />
        <div style={{display: 'flex'}}>
          {rungs.map((r, i) => {
            const on = ramp(frame, 20 + i * 6, 14);
            const isPick = i === pick;
            return (
              <div key={i} style={{flex: 1, textAlign: 'center', opacity: on}}>
                <div
                  style={{
                    fontFamily: BODY,
                    fontSize: 30,
                    color: isPick ? C.cream : C.dim,
                    marginBottom: 18,
                    fontWeight: isPick ? 600 : 400,
                  }}
                >
                  {r}
                </div>
                <div
                  style={{
                    width: isPick ? 30 : 16,
                    height: isPick ? 30 : 16,
                    margin: '0 auto',
                    borderRadius: '50%',
                    backgroundColor: isPick ? C.brand : C.line,
                    border: isPick ? `4px solid ${C.ink0}` : 'none',
                    boxShadow: isPick ? `0 0 34px ${C.brand}` : 'none',
                    transform: `translateY(${isPick ? -6 : 0}px)`,
                    position: 'relative',
                    zIndex: 1,
                  }}
                />
                {isPick ? (
                  <div
                    style={{
                      marginTop: 20,
                      fontFamily: MONO,
                      fontSize: 20,
                      letterSpacing: '0.14em',
                      color: C.brand,
                      textTransform: 'uppercase',
                    }}
                  >
                    default
                  </div>
                ) : null}
              </div>
            );
          })}
        </div>
      </div>
      <div style={{display: 'flex', justifyContent: 'space-between', opacity: ramp(frame, 54, 16)}}>
        <div style={{fontFamily: BODY, fontSize: 28, color: C.muted}}>← {left}</div>
        <div style={{fontFamily: BODY, fontSize: 28, color: C.muted}}>{right} →</div>
      </div>
      <div
        style={{
          textAlign: 'center',
          fontFamily: BODY,
          fontSize: 30,
          color: C.brandSoft,
          opacity: ramp(frame, 62, 16),
        }}
      >
        {note}
      </div>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------
// Address — taking localhost:8000 apart, each label hung under its own part.

export const Address: React.FC = () => {
  const frame = useCurrentFrame();
  const parts: {text: string; label?: string; note?: string}[] = [
    {text: 'localhost', label: 'this machine', note: 'not the internet'},
    {text: ':'},
    {text: '8000', label: 'which door', note: 'the port it listens on'},
  ];
  return (
    <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 20}}>
      <div style={{textAlign: 'center'}}>
        <Eyebrow>reading the address</Eyebrow>
      </div>
      <div style={{display: 'flex', justifyContent: 'center', alignItems: 'flex-start', gap: 6}}>
        {parts.map((p, i) => {
          const on = ramp(frame, 8 + i * 10, 16);
          const lab = ramp(frame, 34 + i * 10, 18);
          return (
            <div key={i} style={{display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
              <div
                style={{
                  fontFamily: MONO,
                  fontSize: 124,
                  fontWeight: 700,
                  lineHeight: 1.15,
                  color: p.label ? C.brand : C.dim,
                  opacity: on,
                  transform: `translateY(${(1 - on) * 22}px)`,
                }}
              >
                {p.text}
              </div>
              {p.label ? (
                <div style={{textAlign: 'center', opacity: lab}}>
                  <div style={{width: 2, height: 52 * lab, backgroundColor: C.line, margin: '14px auto 18px'}} />
                  <div style={{fontFamily: DISPLAY, fontWeight: 650, fontSize: 44, color: C.cream}}>
                    {p.label}
                  </div>
                  <div style={{fontFamily: BODY, fontSize: 27, color: C.muted, marginTop: 10}}>
                    {p.note}
                  </div>
                </div>
              ) : null}
            </div>
          );
        })}
      </div>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------

export const PromptBuild: React.FC<{
  title: string;
  chunks: {text: string; tag: string}[];
}> = ({title, chunks}) => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 56}}>
      <Headline text={title} size={58} width={1500} />
      <Card delay={6} style={{padding: '48px 54px 44px'}}>
        <div
          style={{
            fontFamily: MONO,
            fontSize: 22,
            letterSpacing: '0.2em',
            color: C.dim,
            textTransform: 'uppercase',
            marginBottom: 30,
          }}
        >
          your prompt
        </div>
        <div style={{display: 'flex', flexWrap: 'wrap', gap: '22px 14px', alignItems: 'flex-start'}}>
          {chunks.map((c, i) => {
            const on = ramp(frame, 16 + i * 20, 16);
            const tagged = Boolean(c.tag);
            return (
              <div key={i} style={{opacity: on, transform: `translateY(${(1 - on) * 16}px)`}}>
                <div
                  style={{
                    fontFamily: BODY,
                    fontSize: 46,
                    color: C.cream,
                    padding: '10px 16px',
                    borderRadius: 12,
                    backgroundColor: tagged ? `${C.brand}1F` : 'transparent',
                    border: `1px solid ${tagged ? `${C.brand}55` : 'transparent'}`,
                  }}
                >
                  {c.text}
                </div>
                <div
                  style={{
                    marginTop: 12,
                    fontFamily: MONO,
                    fontSize: 20,
                    letterSpacing: '0.12em',
                    color: C.brandSoft,
                    textAlign: 'center',
                    textTransform: 'uppercase',
                    minHeight: 24,
                  }}
                >
                  {c.tag}
                </div>
              </div>
            );
          })}
        </div>
      </Card>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------

export const Devices: React.FC<{title: string}> = ({title}) => {
  const frame = useCurrentFrame();
  const devices = [
    {name: 'Mobile', w: 200, h: 400, cols: 1},
    {name: 'Tablet', w: 360, h: 440, cols: 2},
    {name: 'Desktop', w: 660, h: 420, cols: 3},
  ];
  return (
    <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 60}}>
      <Headline text={title} size={62} align="center" width="100%" />
      <div style={{display: 'flex', gap: 62, justifyContent: 'center', alignItems: 'flex-end'}}>
        {devices.map((d, i) => {
          const e = useEnter(10 + i * 12);
          const cells = d.cols * 2;
          return (
            <div
              key={i}
              style={{textAlign: 'center', opacity: e, transform: `translateY(${(1 - e) * 30}px)`}}
            >
              <div
                style={{
                  width: d.w,
                  height: d.h,
                  borderRadius: 18,
                  border: `2px solid ${C.line}`,
                  backgroundColor: C.ink1,
                  padding: 14,
                  boxShadow: `0 26px 60px ${LIGHT ? 'rgba(23,27,33,0.16)' : 'rgba(0,0,0,0.5)'}`,
                }}
              >
                <div
                  style={{
                    height: 12,
                    borderRadius: 4,
                    backgroundColor: C.brand,
                    opacity: 0.85,
                    marginBottom: 10,
                    width: '55%',
                  }}
                />
                <div
                  style={{
                    height: d.h * 0.24,
                    borderRadius: 10,
                    background: `linear-gradient(135deg, ${C.brandDeep}, ${C.brand})`,
                    marginBottom: 12,
                    opacity: 0.9,
                  }}
                />
                <div style={{display: 'grid', gridTemplateColumns: `repeat(${d.cols}, 1fr)`, gap: 8}}>
                  {Array.from({length: cells}).map((_, k) => {
                    const on = ramp(frame, 24 + i * 12 + k * 3, 12);
                    return (
                      <div
                        key={k}
                        style={{
                          height: 46,
                          borderRadius: 8,
                          backgroundColor: C.ink2,
                          border: `1px solid ${C.line}`,
                          opacity: on,
                        }}
                      />
                    );
                  })}
                </div>
              </div>
              <div
                style={{
                  marginTop: 22,
                  fontFamily: MONO,
                  fontSize: 22,
                  letterSpacing: '0.16em',
                  textTransform: 'uppercase',
                  color: C.muted,
                }}
              >
                {d.name}
              </div>
            </div>
          );
        })}
      </div>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------

export const Quiz: React.FC<{
  n: string;
  question: string;
  options: string[];
  answer: number;
  revealAt: number;
}> = ({n, question, options, answer, revealAt}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  return (
    <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 58}}>
      <div style={{display: 'flex', alignItems: 'center', gap: 22}}>
        <div
          style={{
            fontFamily: MONO,
            fontSize: 22,
            letterSpacing: '0.2em',
            color: C.brandSoft,
            border: `1px solid ${C.brand}66`,
            padding: '8px 16px',
            borderRadius: 999,
          }}
        >
          {n}
        </div>
        <Rule delay={2} width={80} />
      </div>
      <Headline text={question} size={68} delay={4} width={1560} />
      <div style={{display: 'flex', gap: 22}}>
        {options.map((o, i) => {
          const e = useEnter(16 + i * 8);
          const isAns = i === answer;
          const rev = spring({frame: frame - revealAt, fps, config: {damping: 15, stiffness: 170}});
          const revealed = frame >= revealAt;
          const wrongFade = revealed ? 1 - rev * 0.72 : 1;
          return (
            <div
              key={i}
              style={{
                flex: 1,
                padding: '32px 26px',
                borderRadius: 18,
                backgroundColor: revealed && isAns ? (LIGHT ? '#E9F6EC' : '#1B2C1C') : C.ink1,
                border: `1px solid ${revealed && isAns ? C.ok : C.line}`,
                boxShadow:
                  revealed && isAns ? `0 0 0 2px ${C.ok}${LIGHT ? '99' : '55'}, 0 22px 60px ${LIGHT ? 'rgba(23,27,33,0.16)' : 'rgba(0,0,0,0.5)'}` : 'none',
                opacity: (isAns ? 1 : wrongFade) * e,
                transform: `translateY(${(1 - e) * 22}px) scale(${
                  revealed && isAns ? 1 + rev * 0.035 : 1
                })`,
                display: 'flex',
                alignItems: 'center',
                gap: 16,
              }}
            >
              <div
                style={{
                  fontFamily: MONO,
                  fontSize: 24,
                  color: revealed && isAns ? C.ok : C.dim,
                  width: 28,
                  flexShrink: 0,
                }}
              >
                {String.fromCharCode(65 + i)}
              </div>
              <div
                style={{
                  fontFamily: BODY,
                  fontSize: 33,
                  color: revealed && isAns ? (LIGHT ? '#12551F' : '#EAFBEA') : C.cream,
                  fontWeight: revealed && isAns ? 600 : 400,
                  lineHeight: 1.25,
                }}
              >
                {o}
              </div>
              {revealed && isAns ? (
                <div style={{marginLeft: 'auto', fontSize: 32, color: C.ok, opacity: rev}}>✓</div>
              ) : null}
            </div>
          );
        })}
      </div>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------

export const Roadmap: React.FC<{
  title: string;
  nodes: {name: string; state: 'done' | 'here' | 'next'}[];
}> = ({title, nodes}) => {
  const frame = useCurrentFrame();
  const n = nodes.length;
  const doneCount = nodes.filter((x) => x.state === 'done').length;
  const hereIdx = nodes.findIndex((x) => x.state === 'here');
  const target = ((hereIdx >= 0 ? hereIdx : doneCount - 1) + 0.5) / n;
  const line = ramp(frame, 8, 34) * target;
  return (
    <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 92}}>
      <Headline text={title} size={62} width={1400} />
      <div style={{position: 'relative'}}>
        <div
          style={{position: 'absolute', top: 30, left: 0, right: 0, height: 2, backgroundColor: C.line}}
        />
        <div
          style={{
            position: 'absolute',
            top: 30,
            left: 0,
            width: `${line * 100}%`,
            height: 2,
            background: `linear-gradient(90deg, ${C.brand}, ${C.brandSoft})`,
          }}
        />
        <div style={{display: 'flex', position: 'relative'}}>
          {nodes.map((nd, i) => {
            const on = ramp(frame, 14 + i * 8, 16);
            const done = nd.state === 'done';
            const here = nd.state === 'here';
            return (
              <div key={i} style={{flex: 1, textAlign: 'center', opacity: on, padding: '0 8px'}}>
                <div
                  style={{
                    width: here ? 30 : 18,
                    height: here ? 30 : 18,
                    margin: `${here ? 15 : 21}px auto 0`,
                    borderRadius: '50%',
                    backgroundColor: done ? C.brand : here ? C.cream : C.ink2,
                    border: `2px solid ${done ? C.brand : here ? C.cream : C.line}`,
                    boxShadow: here ? `0 0 34px ${C.cream}77` : 'none',
                    position: 'relative',
                    zIndex: 1,
                  }}
                />
                <div
                  style={{
                    marginTop: 26,
                    fontFamily: BODY,
                    fontSize: 27,
                    lineHeight: 1.3,
                    color: done ? C.cream : here ? C.cream : C.dim,
                    fontWeight: here ? 600 : 400,
                  }}
                >
                  {nd.name}
                </div>
                {done ? (
                  <div style={{marginTop: 8, fontSize: 22, color: C.brand}}>✓</div>
                ) : null}
                {here ? (
                  <div
                    style={{
                      marginTop: 10,
                      fontFamily: MONO,
                      fontSize: 18,
                      letterSpacing: '0.16em',
                      color: C.brand,
                      textTransform: 'uppercase',
                    }}
                  >
                    you are here
                  </div>
                ) : null}
              </div>
            );
          })}
        </div>
      </div>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------

export const Outro: React.FC<{lines: string[]; sign: string}> = ({lines, sign}) => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill>
      <Motes n={30} seed="outro" />
      <AbsoluteFill style={{padding: PAD, justifyContent: 'center', gap: 40}}>
        <Rule delay={0} width={160} />
        {lines.map((l, i) => (
          <Headline
            key={i}
            text={l}
            size={72}
            delay={8 + i * 14}
            accent={['alongside', 'me.']}
            width={1500}
          />
        ))}
        <div
          style={{
            marginTop: 40,
            fontFamily: MONO,
            fontSize: 26,
            letterSpacing: '0.2em',
            textTransform: 'uppercase',
            color: C.dim,
            opacity: ramp(frame, 52, 20),
          }}
        >
          {sign}
        </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};
