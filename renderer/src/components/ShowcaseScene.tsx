import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_W} from './Stage';
import {iconFor} from './icons';

// ShowcaseScene is one tool's card, and then a door out of it.
//
// The card is laid out in the order people actually decide in: what it is, then
// the four things they choose on, then the honest two columns. Everything is on
// screen from the first frame and the beats only light parts of it, so a viewer
// who joins late still sees a whole product rather than a fragment.
//
// The last beat is the reason this template exists. The card pushes back and a
// framed play glyph takes the screen — a designed out-point, the same frame
// every time, so a screen recording of the tool can be cut on without matching
// anything. Ending on the card instead would make every clip need a different
// edit, which is exactly the friction that stops a course having demos in it.

const CARD_W = Math.min(STAGE_W, 1700);
const TILE = 116;
const FACTS_H = 128;
const PLATE_W = 980;
const PLATE_H = 552;

type Fact = {label: string; value: string};
type Step = {startMs: number; endMs: number; show: string; at?: number};

export const ShowcaseScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const title = String(props.title ?? '');
  const name = String(props.name ?? '');
  const category = String(props.category ?? '');
  const tagline = String(props.tagline ?? '');
  const facts = (Array.isArray(props.facts) ? props.facts : []) as Fact[];
  const strengths = (Array.isArray(props.strengths) ? props.strengths : []) as string[];
  const limits = (Array.isArray(props.limits) ? props.limits : []) as string[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];

  if (steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const show = step.show;
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const handoff = show === 'handoff';
  const litFact = show === 'fact' ? (step.at ?? 0) : -1;
  const enter = spring({frame: sinceStep, fps, config: {damping: 200, mass: 0.7}, durationInFrames: 20});
  // The card does not vanish on the hand-off; it recedes behind the plate, so
  // the cut reads as leaving the card rather than as a scene change.
  const recede = handoff ? enter : 0;

  const Icon = iconFor(String(props.icon ?? 'box'));

  return (
    <Stage justify="center">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={18} />

      <div style={{position: 'relative', width: CARD_W}}>
        <div
          style={{
            opacity: 1 - recede * 0.84,
            transform: `scale(${1 - recede * 0.06})`,
            filter: recede > 0 ? `saturate(${1 - recede * 0.7})` : undefined,
          }}
        >
          {/* Identity: the tile, the name, and what the thing actually is. */}
          <div style={{display: 'flex', alignItems: 'center', gap: 26, marginBottom: 26}}>
            <div
              style={{
                width: TILE,
                height: TILE,
                flexShrink: 0,
                borderRadius: 30,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: `linear-gradient(150deg, ${withAlpha(theme.accent, 0.34)}, ${withAlpha(
                  theme.primary,
                  0.2,
                )})`,
                border: `1px solid ${withAlpha(theme.accent, 0.42)}`,
                boxShadow: `0 16px 40px ${withAlpha(theme.ink, 0.5)}`,
                transform: `scale(${show === 'intro' ? 0.94 + enter * 0.06 : 1})`,
              }}
            >
              <Icon size={56} color={theme.accent} strokeWidth={2} />
            </div>
            <div style={{minWidth: 0}}>
              <div style={{display: 'flex', alignItems: 'center', gap: 16}}>
                <span
                  style={{
                    fontFamily: theme.fontDisplay,
                    fontSize: 58,
                    fontWeight: 700,
                    letterSpacing: -1.2,
                    color: theme.text,
                  }}
                >
                  {name}
                </span>
                {category ? (
                  <span
                    style={{
                      padding: '7px 16px',
                      borderRadius: 999,
                      border: `1px solid ${withAlpha(theme.accent, 0.4)}`,
                      backgroundColor: withAlpha(theme.accent, 0.1),
                      fontFamily: theme.fontMono,
                      fontSize: 17,
                      letterSpacing: 1.6,
                      textTransform: 'uppercase',
                      color: theme.accentText,
                    }}
                  >
                    {category}
                  </span>
                ) : null}
              </div>
              <div
                style={{
                  marginTop: 8,
                  fontFamily: theme.fontBody,
                  fontSize: 29,
                  lineHeight: 1.3,
                  color: theme.textMuted,
                }}
              >
                {tagline}
              </div>
            </div>
          </div>

          {/* The decision cells. */}
          <div style={{display: 'flex', gap: 20, height: FACTS_H, marginBottom: 22}}>
            {facts.map((f, i) => {
              const lit = i === litFact;
              return (
                <div
                  key={i}
                  style={{
                    flex: 1,
                    minWidth: 0,
                    borderRadius: 18,
                    padding: '20px 22px',
                    display: 'flex',
                    flexDirection: 'column',
                    justifyContent: 'center',
                    gap: 10,
                    backgroundColor: lit ? withAlpha(theme.accent, 0.1) : withAlpha(theme.surface, 0.6),
                    border: `${lit ? 2 : 1}px solid ${lit ? theme.accent : theme.surfaceBorder}`,
                    boxShadow: lit ? `0 0 0 7px ${withAlpha(theme.accent, 0.1)}` : 'none',
                    transform: `translateY(${lit ? -enter * 7 : 0}px)`,
                  }}
                >
                  <span
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: 15,
                      letterSpacing: 2,
                      textTransform: 'uppercase',
                      color: lit ? theme.accentText : theme.textMuted,
                    }}
                  >
                    {f.label}
                  </span>
                  <span
                    style={{
                      fontFamily: theme.fontDisplay,
                      fontSize: 27,
                      fontWeight: 650,
                      lineHeight: 1.2,
                      color: theme.text,
                    }}
                  >
                    {f.value}
                  </span>
                </div>
              );
            })}
          </div>

          {/* The honest half. Both columns are always present — the limits are
              not a reveal, they are half of what the card says.
              The row has no fixed height: flex stretch already makes the two
              columns match each other, and pinning them to a constant left a
              void under whichever list was shorter. */}
          <div style={{display: 'flex', gap: 24}}>
            <PointColumn
              theme={theme}
              heading="Good at"
              iconName="check"
              points={strengths}
              lit={show === 'strengths'}
              sinceStep={sinceStep}
              positive
            />
            <PointColumn
              theme={theme}
              heading="Watch out for"
              iconName="alert"
              points={limits}
              lit={show === 'limits'}
              sinceStep={sinceStep}
              positive={false}
            />
          </div>
        </div>

        {/* The hand-off. This frame is the cut point for a screen recording. */}
        {handoff ? (
          <div
            style={{
              position: 'absolute',
              inset: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              opacity: enter,
            }}
          >
            <div
              style={{
                width: PLATE_W,
                height: PLATE_H,
                borderRadius: 26,
                backgroundColor: theme.bgBottom,
                border: `1px solid ${theme.surfaceBorder}`,
                boxShadow: `0 34px 90px ${withAlpha(theme.ink, 0.72)}`,
                transform: `scale(${0.93 + enter * 0.07})`,
                overflow: 'hidden',
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              <div
                style={{
                  height: 52,
                  flexShrink: 0,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 9,
                  padding: '0 20px',
                  borderBottom: `1px solid ${theme.surfaceBorder}`,
                }}
              >
                {[0, 1, 2].map((d) => (
                  <span
                    key={d}
                    style={{
                      width: 11,
                      height: 11,
                      borderRadius: 6,
                      backgroundColor: theme.line,
                      opacity: 0.45,
                    }}
                  />
                ))}
              </div>
              <div
                style={{
                  flex: 1,
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: 26,
                }}
              >
                <div
                  style={{
                    width: 130,
                    height: 130,
                    borderRadius: 65,
                    backgroundColor: theme.accent,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    boxShadow: `0 0 0 18px ${withAlpha(theme.accent, 0.13)}`,
                  }}
                >
                  <svg width={52} height={56} viewBox="0 0 52 56">
                    <path d="M8 4 L46 28 L8 52 Z" fill={theme.ink} />
                  </svg>
                </div>
                <span
                  style={{
                    fontFamily: theme.fontDisplay,
                    fontSize: 42,
                    fontWeight: 700,
                    letterSpacing: -0.6,
                    color: theme.text,
                  }}
                >
                  {name}
                </span>
                <span
                  style={{
                    fontFamily: theme.fontMono,
                    fontSize: 19,
                    letterSpacing: 3,
                    textTransform: 'uppercase',
                    color: theme.textMuted,
                  }}
                >
                  Demo
                </span>
              </div>
            </div>
          </div>
        ) : null}
      </div>
    </Stage>
  );
};

/**
 * One of the two columns.
 *
 * The two are drawn identically apart from their marker, which is the point: a
 * limitation set in smaller, quieter type than a strength is a limitation the
 * layout is apologising for. They carry the same weight because the template's
 * whole argument is that they are worth the same.
 */
const PointColumn: React.FC<{
  theme: ResolvedTheme;
  heading: string;
  iconName: string;
  points: string[];
  lit: boolean;
  sinceStep: number;
  positive: boolean;
}> = ({theme, heading, iconName, points, lit, sinceStep, positive}) => {
  const Icon = iconFor(iconName);
  const mark = positive ? theme.accent : theme.line;
  return (
    <div
      style={{
        flex: 1,
        minWidth: 0,
        borderRadius: 20,
        padding: '20px 24px',
        backgroundColor: lit ? withAlpha(theme.surface, 0.85) : withAlpha(theme.surface, 0.45),
        border: `${lit ? 2 : 1}px solid ${lit ? theme.accent : theme.surfaceBorder}`,
        opacity: lit ? 1 : 0.52,
        display: 'flex',
        flexDirection: 'column',
        gap: 14,
      }}
    >
      <div style={{display: 'flex', alignItems: 'center', gap: 11}}>
        <Icon size={22} color={lit ? theme.accent : theme.textMuted} strokeWidth={2.3} />
        <span
          style={{
            fontFamily: theme.fontMono,
            fontSize: 16,
            letterSpacing: 2.2,
            textTransform: 'uppercase',
            color: lit ? theme.accentText : theme.textMuted,
          }}
        >
          {heading}
        </span>
      </div>
      {points.map((p, i) => (
        <div
          key={i}
          style={{
            display: 'flex',
            alignItems: 'flex-start',
            gap: 14,
            opacity: lit
              ? interpolate(sinceStep, [4 + i * 6, 16 + i * 6], [0.25, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                })
              : 1,
          }}
        >
          <span
            style={{
              width: 22,
              height: 22,
              marginTop: 5,
              flexShrink: 0,
              borderRadius: 7,
              backgroundColor: withAlpha(mark, 0.2),
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <svg width={12} height={12} viewBox="0 0 12 12">
              <path
                d={positive ? 'M6 1.5 L6 10.5 M1.5 6 L10.5 6' : 'M1.5 6 L10.5 6'}
                stroke={mark}
                strokeWidth={2.4}
                strokeLinecap="round"
              />
            </svg>
          </span>
          <span
            style={{
              fontFamily: theme.fontBody,
              fontSize: 25,
              lineHeight: 1.28,
              color: theme.text,
            }}
          >
            {p}
          </span>
        </div>
      ))}
    </div>
  );
};
