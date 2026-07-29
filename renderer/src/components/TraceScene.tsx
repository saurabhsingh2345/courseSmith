import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';

// TraceScene is a system caught in the act.
//
// Actors down the left, the queue in the middle, the shared value on the right.
// Operations sit in the queue tagged with who issued them, drain one at a time,
// and the value on the right changes as each one lands.
//
// Three decisions make it teach rather than merely animate.
//
// The queue holds everything from the `queue` beat onward, not just the next
// item. That single frame — three operations pending against one value — is the
// entire argument for why order matters, and it cannot be made by a diagram
// that only ever shows the operation currently executing.
//
// A drained operation stays visible, struck and dimmed, above the queue. The
// viewer needs the history to reconstruct how the value got where it is, which
// is precisely the reasoning the clip is teaching.
//
// An operation that does NOT change the value is marked as such rather than
// left to look identical to one that did. "This read changed nothing" is
// usually the whole bug, and Go decides it (see TraceStepSpec.Becomes) so the
// mark can never disagree with the state chain.

const COL_W = Math.min(STAGE_W, 1580);

type Op = {by: number; op: string; becomes: string; note?: string; changes: boolean};
type Step = {startMs: number; endMs: number; show: 'setup' | 'queue' | 'step' | 'outcome'; at?: number};

export const TraceScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const actors = (Array.isArray(props.actors) ? props.actors : []) as string[];
  const resource = String(props.resource ?? '');
  const start = String(props.start ?? '');
  const outcome = String(props.outcome ?? '');
  const broken = props.broken === true;
  const ops = (Array.isArray(props.ops) ? props.ops : []) as Op[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (actors.length === 0 || ops.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const onSetup = step.show === 'setup';
  const onOutcome = step.show === 'outcome';
  // How many operations have landed. On the outcome every one has.
  const landed = onOutcome
    ? ops.length
    : steps.slice(0, idx + 1).filter((s) => s.show === 'step').length;
  const current = step.show === 'step' ? (step.at ?? 0) : -1;

  // The value the resource currently holds, reconstructed from what has landed.
  const value = landed === 0 ? start : ops[Math.min(landed, ops.length) - 1].becomes;
  const enter = spring({
    frame: ((nowMs - steps[0].startMs) / 1000) * FPS,
    fps,
    config: {damping: 200, mass: 0.7},
    durationInFrames: 20,
  });

  const actorColor = (i: number): string =>
    [theme.accentRival, theme.accentQuantity, theme.accentLimit][i % 3];

  return (
    <Stage justify="center">
      <div style={{width: COL_W, opacity: enter}}>
        <div style={{display: 'flex', gap: 40, alignItems: 'stretch'}}>
          {/* The actors. */}
          <div style={{width: 250, display: 'flex', flexDirection: 'column', gap: 16, justifyContent: 'center'}}>
            {actors.map((a, i) => {
              const active = current >= 0 && ops[current]?.by === i;
              const c = actorColor(i);
              return (
                <div
                  key={i}
                  style={{
                    padding: '18px 22px',
                    borderRadius: 10,
                    background: active ? withAlpha(c, 0.16) : withAlpha(theme.text, 0.04),
                    border: `1px solid ${active ? withAlpha(c, 0.5) : withAlpha(theme.text, 0.1)}`,
                    fontFamily: theme.fontMono,
                    fontSize: 20,
                    letterSpacing: 1.6,
                    color: active ? theme.text : theme.textMuted,
                  }}
                >
                  {a}
                </div>
              );
            })}
          </div>

          {/* The queue. Everything pending is here from the queue beat on —
              that frame is the whole argument for why order matters. */}
          <div
            style={{
              flex: 1,
              padding: '20px 22px',
              borderRadius: 12,
              background: withAlpha(theme.text, 0.035),
              border: `1px solid ${withAlpha(theme.text, 0.1)}`,
              minHeight: 320,
            }}
          >
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 15,
                letterSpacing: 3.4,
                textTransform: 'uppercase',
                color: theme.textMuted,
                marginBottom: 16,
              }}
            >
              Queue
            </div>
            {onSetup ? (
              <div
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 19,
                  color: withAlpha(theme.text, 0.28),
                  paddingTop: 24,
                }}
              >
                nothing sent yet
              </div>
            ) : (
              ops.map((o, i) => {
                const done = i < landed;
                const isCurrent = i === current;
                const c = actorColor(o.by);
                return (
                  <div
                    key={i}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 14,
                      padding: '13px 18px',
                      marginBottom: 9,
                      borderRadius: 8,
                      background: isCurrent ? withAlpha(c, 0.16) : 'transparent',
                      border: `1px solid ${isCurrent ? withAlpha(c, 0.45) : withAlpha(theme.text, 0.08)}`,
                      // Drained operations stay visible: the viewer needs the
                      // history to reconstruct how the value got where it is.
                      opacity: done && !isCurrent ? 0.4 : 1,
                    }}
                  >
                    <span style={{fontFamily: theme.fontMono, fontSize: 15, color: c, minWidth: 88}}>
                      {actors[o.by]}
                    </span>
                    <span
                      style={{
                        fontFamily: theme.fontMono,
                        fontSize: 22,
                        color: theme.text,
                        textDecoration: done && !isCurrent ? 'line-through' : 'none',
                      }}
                    >
                      {o.op}
                    </span>
                    {/* An operation that changed nothing is marked. It usually
                        IS the bug, and it must not look like one that landed. */}
                    {done && !o.changes ? (
                      <span
                        style={{
                          marginLeft: 'auto',
                          fontFamily: theme.fontMono,
                          fontSize: 14,
                          letterSpacing: 1.6,
                          color: theme.accentLimit,
                        }}
                      >
                        no change
                      </span>
                    ) : null}
                  </div>
                );
              })
            )}
          </div>

          {/* The shared value. */}
          <div
            style={{
              width: 320,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              padding: '26px 20px',
              borderRadius: 12,
              background: withAlpha(theme.text, 0.05),
              border: `1px solid ${withAlpha(theme.text, 0.14)}`,
            }}
          >
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 15,
                letterSpacing: 3.4,
                textTransform: 'uppercase',
                color: theme.textMuted,
                marginBottom: 14,
              }}
            >
              {resource}
            </div>
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 96,
                fontWeight: 800,
                letterSpacing: -3,
                lineHeight: 1,
                color: onOutcome && broken ? theme.accentLimit : theme.text,
                fontVariantNumeric: 'tabular-nums',
                // A short pop as the value takes its new state, so a change the
                // eye is not looking at still registers in peripheral vision.
                transform: `scale(${
                  current >= 0
                    ? 1 +
                      0.09 *
                        interpolate(sinceStep, [2, 10, 20], [0, 1, 0], {
                          extrapolateLeft: 'clamp',
                          extrapolateRight: 'clamp',
                        })
                    : 1
                })`,
              }}
            >
              {value}
            </div>
          </div>
        </div>

        {/* The line for the current step, or the outcome. */}
        <div style={{minHeight: 96, marginTop: 30, textAlign: 'center'}}>
          {onOutcome ? (
            <div
              style={{
                opacity: interpolate(sinceStep, [3, 18], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                }),
              }}
            >
              <div
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 16,
                  letterSpacing: 4,
                  textTransform: 'uppercase',
                  color: broken ? theme.accentLimit : theme.accentQuantity,
                  marginBottom: 12,
                }}
              >
                {broken ? 'And that is the bug' : 'What the order proved'}
              </div>
              <div
                style={{
                  fontFamily: theme.fontDisplay,
                  fontSize: 44,
                  fontWeight: 700,
                  letterSpacing: -0.8,
                  lineHeight: 1.2,
                  color: theme.text,
                  maxWidth: 1240,
                  margin: '0 auto',
                }}
              >
                {outcome}
              </div>
            </div>
          ) : current >= 0 && ops[current]?.note ? (
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 28,
                color: theme.textMuted,
                maxWidth: 1180,
                margin: '0 auto',
                opacity: interpolate(sinceStep, [4, 18], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                }),
              }}
            >
              {ops[current].note}
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
