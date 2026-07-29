import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';

// DecisionScene is one question, an axis under it, and the answer for wherever
// you land.
//
// The axis is a row of segments with a hairline between them and the band
// labels below, and a marker that travels to whichever band is current. What
// makes it read as a *decision* rather than as a chart is that the answer is
// set at headline size beneath the axis and changes as the marker moves — the
// viewer's eye goes marker, then answer, and the sequence is the argument.
//
// Three decisions carry it.
//
// The segments are drawn evenly rather than in proportion to their bounds. A
// proportional axis gives the open-ended final band no width at all and
// squeezes "0-8" against "32-512" into a sliver; the arithmetic lives in the
// labels. This axis is a sequence of cases, and `gauge` is the template where
// distance along the track is the claim.
//
// Every band is on screen from the first frame, unlit. The claim is that the
// question *partitions* the audience, and a partition you reveal one piece at a
// time reads as a list.
//
// The answer swaps with a short vertical wipe rather than a cross-fade, so two
// instructions are never legible at once. A decision guide that appears to give
// two answers for half a second is a decision guide nobody trusts.

const COL_W = Math.min(STAGE_W, 1560);
const AXIS_H = 96;

type Tier = {
  band: string;
  answer: string;
  note?: string;
  role?: 'quantity' | 'limit' | 'rival' | 'neutral';
};
type Step = {startMs: number; endMs: number; show: 'question' | 'tier' | 'rule'; at?: number};

const roleColor = (theme: ResolvedTheme, role?: string): string => {
  switch (role) {
    case 'quantity':
      return theme.accentQuantity;
    case 'limit':
      return theme.accentLimit;
    case 'rival':
      return theme.accentRival;
    default:
      return theme.textMuted;
  }
};

export const DecisionScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const question = String(props.question ?? '');
  const tiers = (Array.isArray(props.tiers) ? props.tiers : []) as Tier[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (tiers.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;
  const rule = step.show === 'rule';
  const current = step.show === 'tier' ? (step.at ?? 0) : -1;

  const enter = spring({
    frame: ((nowMs - steps[0].startMs) / 1000) * FPS,
    fps,
    config: {damping: 200, mass: 0.7},
    durationInFrames: 18,
  });

  // The marker travels rather than jumps, so the eye follows it across the axis
  // and reads the bands it passes as ones it has already been given.
  const markerTarget = current < 0 ? 0 : current;
  const glide = spring({
    frame: sinceStep,
    fps,
    config: {damping: 26, mass: 0.9, stiffness: 90},
    durationInFrames: 30,
  });
  const prevTier = (() => {
    for (let i = idx - 1; i >= 0; i--) {
      if (steps[i].show === 'tier') return steps[i].at ?? 0;
    }
    return markerTarget;
  })();
  const markerAt = current < 0 ? markerTarget : prevTier + (markerTarget - prevTier) * glide;
  const segW = 100 / tiers.length;

  return (
    <Stage justify="center">
      <div style={{width: COL_W, opacity: enter}}>
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 19,
            letterSpacing: 5,
            textTransform: 'uppercase',
            color: theme.textMuted,
            textAlign: 'center',
            marginBottom: 16,
          }}
        >
          The only question
        </div>
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 62,
            fontWeight: 800,
            letterSpacing: -1.4,
            lineHeight: 1.1,
            color: theme.text,
            textAlign: 'center',
            marginBottom: 54,
          }}
        >
          {question}
        </div>

        {/* The axis. */}
        <div style={{position: 'relative', marginBottom: 22}}>
          {/* The travelling marker. Present only while a single band is
              current: on the question beat nothing has been chosen yet, and on
              the rule beat every band is lit at once, so a marker singling one
              of them out would contradict the frame it sits on. */}
          {current >= 0 ? (
            <div
              style={{
                position: 'absolute',
                left: `${(markerAt + 0.5) * segW}%`,
                top: -26,
                transform: 'translateX(-50%)',
                fontSize: 20,
                color: roleColor(theme, tiers[Math.round(markerAt)]?.role),
                lineHeight: 1,
              }}
            >
              ▼
            </div>
          ) : null}

          <div style={{display: 'flex', height: AXIS_H, borderRadius: 8, overflow: 'hidden'}}>
            {tiers.map((t, i) => {
              const lit = rule || i === current;
              const c = roleColor(theme, t.role);
              return (
                <div
                  key={i}
                  style={{
                    flex: 1,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    // A hairline between bands rather than a gap: the axis is
                    // continuous and the divisions are cuts in it, not four
                    // separate objects sitting near each other.
                    borderRight: i === tiers.length - 1 ? 'none' : `1px solid ${withAlpha(theme.text, 0.16)}`,
                    background: lit ? withAlpha(c, 0.2) : withAlpha(theme.text, 0.045),
                    transition: 'none',
                  }}
                >
                  <div
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: 21,
                      letterSpacing: 2.4,
                      textTransform: 'uppercase',
                      color: lit ? c : theme.textMuted,
                      fontWeight: lit ? 600 : 500,
                    }}
                  >
                    {t.band}
                  </div>
                </div>
              );
            })}
          </div>
          {/* The lit band's underline, drawn on top so it reads as a selection
              on the axis rather than as part of the segment's fill. */}
          {current >= 0 || rule ? (
            <div
              style={{
                position: 'absolute',
                left: rule ? '0%' : `${markerAt * segW}%`,
                width: rule ? '100%' : `${segW}%`,
                bottom: -3,
                height: 4,
                borderRadius: 2,
                background: rule
                  ? withAlpha(theme.text, 0.3)
                  : roleColor(theme, tiers[Math.round(markerAt)]?.role),
              }}
            />
          ) : null}
        </div>

        {/* The answer for the current band. On the rule beat every band's
            answer is listed instead, which is the frame people screenshot. */}
        <div style={{minHeight: 210, marginTop: 40}}>
          {rule ? (
            <div style={{display: 'flex', gap: 26}}>
              {tiers.map((t, i) => {
                const on = interpolate(sinceStep, [4 + i * 6, 20 + i * 6], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                });
                const c = roleColor(theme, t.role);
                return (
                  <div
                    key={i}
                    style={{
                      flex: 1,
                      padding: '22px 24px',
                      borderRadius: 12,
                      background: withAlpha(c, 0.09),
                      border: `1px solid ${withAlpha(c, 0.32)}`,
                      opacity: on,
                      transform: `translateY(${(1 - on) * 14}px)`,
                    }}
                  >
                    <div
                      style={{
                        fontFamily: theme.fontMono,
                        fontSize: 15,
                        letterSpacing: 2.4,
                        textTransform: 'uppercase',
                        color: c,
                        marginBottom: 12,
                      }}
                    >
                      {t.band}
                    </div>
                    <div
                      style={{
                        fontFamily: theme.fontDisplay,
                        fontSize: 32,
                        fontWeight: 700,
                        lineHeight: 1.2,
                        color: theme.text,
                      }}
                    >
                      {t.answer}
                    </div>
                  </div>
                );
              })}
            </div>
          ) : current >= 0 ? (
            <Answer theme={theme} tier={tiers[current]} sinceStep={sinceStep} />
          ) : null}
        </div>
      </div>
    </Stage>
  );
};

/**
 * One band's instruction, wiped in from below.
 *
 * A vertical wipe rather than a cross-fade: mid-transition a fade shows two
 * instructions at once, and a decision guide that appears to give two answers
 * for half a second is one nobody trusts.
 */
const Answer: React.FC<{theme: ResolvedTheme; tier: Tier; sinceStep: number}> = ({
  theme,
  tier,
  sinceStep,
}) => {
  const wipe = interpolate(sinceStep, [3, 17], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const c = roleColor(theme, tier.role);
  return (
    <div style={{textAlign: 'center', overflow: 'hidden'}}>
      <div
        style={{
          fontFamily: theme.fontDisplay,
          fontSize: 68,
          fontWeight: 800,
          letterSpacing: -1.4,
          lineHeight: 1.12,
          color: c,
          opacity: wipe,
          transform: `translateY(${(1 - wipe) * 40}px)`,
        }}
      >
        {tier.answer}
      </div>
      {tier.note ? (
        <div
          style={{
            fontFamily: theme.fontBody,
            fontSize: 28,
            lineHeight: 1.4,
            color: theme.textMuted,
            maxWidth: 1040,
            margin: '20px auto 0',
            opacity: interpolate(sinceStep, [14, 28], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            }),
          }}
        >
          {tier.note}
        </div>
      ) : null}
    </div>
  );
};
