import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';

// CostingScene is a bill assembling itself, with the total counting as it goes.
//
// Lines land one at a time and stay. The running total sits beside the sheet
// and climbs with every addition, so by the time the closing figure arrives the
// viewer has watched it be built and has no reason to doubt it — which is the
// entire difference between this and asserting a number at the end.
//
// Two details do most of the work.
//
// Hidden costs are drawn in the limit colour with a small mark. The premise of
// a bill of materials is that the sticker price is not the price, so the lines
// nobody budgets for have to be visually separable from the ones everybody
// does — otherwise the sheet is just a list and the surprise never lands.
//
// The bars are scaled against the largest single line rather than the total.
// Against the total the small lines collapse to nothing, and the small lines
// are usually the surprising ones — the point is not that power costs less than
// the GPU, it is that power costs anything at all.

const COL_W = Math.min(STAGE_W, 1520);

type Line = {
  label: string;
  amount: number;
  note?: string;
  hidden?: boolean;
  running: number;
  frac: number;
};
type Step = {startMs: number; endMs: number; show: 'setup' | 'line' | 'total'; at?: number};

/** Money, grouped, with no trailing decimals on whole amounts. */
const money = (n: number): string =>
  Number.isInteger(n) ? n.toLocaleString('en-US') : n.toLocaleString('en-US', {maximumFractionDigits: 2});

export const CostingScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const subject = String(props.subject ?? '');
  const unit = String(props.unit ?? '');
  const verdict = String(props.verdict ?? '');
  const total = Number(props.total ?? 0);
  const lines = (Array.isArray(props.lines) ? props.lines : []) as Line[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (lines.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;
  const onTotal = step.show === 'total';
  const current = step.show === 'line' ? (step.at ?? 0) : -1;

  // How many lines have landed. On the total every one has.
  const landed = onTotal ? lines.length : current >= 0 ? current + 1 : 0;
  const running = landed === 0 ? 0 : lines[landed - 1].running;

  const enter = spring({
    frame: ((nowMs - steps[0].startMs) / 1000) * FPS,
    fps,
    config: {damping: 200, mass: 0.7},
    durationInFrames: 20,
  });

  return (
    <Stage justify="center">
      <div style={{width: COL_W, opacity: enter}}>
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 18,
            letterSpacing: 4.6,
            textTransform: 'uppercase',
            color: theme.textMuted,
            marginBottom: 12,
          }}
        >
          What it really costs
        </div>
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 46,
            fontWeight: 800,
            letterSpacing: -1,
            color: theme.text,
            marginBottom: 38,
          }}
        >
          {subject}
        </div>

        <div style={{display: 'flex', gap: 44, alignItems: 'flex-start'}}>
          {/* The sheet. */}
          <div style={{flex: 1}}>
            {lines.map((l, i) => {
              const shown = i < landed;
              const isCurrent = i === current;
              const c = l.hidden ? theme.accentLimit : theme.accentQuantity;
              // Each line grows from the frame its own beat started, so lines
              // already on the sheet are not re-animating under the new one.
              const since = isCurrent ? sinceStep : shown ? 999 : 0;
              const grow = interpolate(since, [2, 20], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              });
              return (
                <div
                  key={i}
                  style={{
                    marginBottom: 14,
                    opacity: shown ? (isCurrent || onTotal ? 1 : 0.55) : 0.1,
                  }}
                >
                  <div
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'baseline',
                      marginBottom: 7,
                    }}
                  >
                    <div style={{display: 'flex', alignItems: 'center', gap: 12}}>
                      <span
                        style={{
                          fontFamily: theme.fontMono,
                          fontSize: 19,
                          letterSpacing: 1.6,
                          color: isCurrent || onTotal ? theme.text : theme.textMuted,
                        }}
                      >
                        {l.label}
                      </span>
                      {/* The mark that makes a bill worth drawing. */}
                      {l.hidden ? (
                        <span
                          style={{
                            fontFamily: theme.fontMono,
                            fontSize: 13,
                            letterSpacing: 1.8,
                            textTransform: 'uppercase',
                            color: theme.accentLimit,
                            border: `1px solid ${withAlpha(theme.accentLimit, 0.45)}`,
                            borderRadius: 4,
                            padding: '2px 7px',
                          }}
                        >
                          nobody budgets this
                        </span>
                      ) : null}
                    </div>
                    <span
                      style={{
                        fontFamily: theme.fontDisplay,
                        fontSize: 25,
                        fontWeight: 700,
                        color: c,
                        fontVariantNumeric: 'tabular-nums',
                      }}
                    >
                      {unit}
                      {money(l.amount)}
                    </span>
                  </div>
                  <div style={{height: 12, borderRadius: 3, background: withAlpha(theme.text, 0.05)}}>
                    <div
                      style={{
                        height: '100%',
                        width: `${l.frac * grow * 100}%`,
                        borderRadius: 3,
                        background: c,
                      }}
                    />
                  </div>
                  {isCurrent && l.note ? (
                    <div
                      style={{
                        fontFamily: theme.fontBody,
                        fontSize: 22,
                        color: theme.textMuted,
                        marginTop: 9,
                        opacity: interpolate(sinceStep, [14, 26], [0, 1], {
                          extrapolateLeft: 'clamp',
                          extrapolateRight: 'clamp',
                        }),
                      }}
                    >
                      {l.note}
                    </div>
                  ) : null}
                </div>
              );
            })}
          </div>

          {/* The running total, climbing beside the sheet. */}
          <div
            style={{
              width: 420,
              padding: '32px 28px',
              borderRadius: 14,
              background: withAlpha(theme.text, onTotal ? 0.08 : 0.04),
              border: `1px solid ${onTotal ? withAlpha(theme.accentLimit, 0.5) : withAlpha(theme.text, 0.12)}`,
              textAlign: 'center',
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
              {onTotal ? 'Year one' : 'Running total'}
            </div>
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: onTotal ? 82 : 62,
                fontWeight: 800,
                letterSpacing: -2.4,
                lineHeight: 1,
                color: onTotal ? theme.accentLimit : theme.text,
                fontVariantNumeric: 'tabular-nums',
              }}
            >
              {unit}
              {money(Math.round(onTotal ? total : running))}
            </div>
            {onTotal && verdict ? (
              <div
                style={{
                  fontFamily: theme.fontBody,
                  fontSize: 24,
                  lineHeight: 1.35,
                  color: theme.textMuted,
                  marginTop: 22,
                  opacity: interpolate(sinceStep, [10, 24], [0, 1], {
                    extrapolateLeft: 'clamp',
                    extrapolateRight: 'clamp',
                  }),
                }}
              >
                {verdict}
              </div>
            ) : null}
          </div>
        </div>
      </div>
    </Stage>
  );
};
