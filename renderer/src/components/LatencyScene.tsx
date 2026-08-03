import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// LatencyScene is a log time axis with operations placed along it.
//
// Each operation gets its own row and a bar that runs from the LEFT EDGE of the
// axis to its position, rather than a dot at that position. A dot would be the
// more literal drawing of "this took 12ms" and it would read worse: the eye
// compares lengths far more reliably than it compares positions, so the bar is
// what makes "a hundred times" land without arithmetic.
//
// The bar is honest despite the log scale because the ticks are on screen and
// named. That combination — a length the eye can compare, over a scale the frame
// declares — is the whole reason this is not a bar chart with a rescaled axis.
//
// Two decisions carry the rest.
//
// The decade grid stays visible behind the bars for the whole clip. On a log axis
// the gridlines are not decoration: without them a bar ending two thirds along
// means nothing, and with them it means "between 100ms and a second". They are
// the scale, so they never fade.
//
// Rows already placed stay at full strength rather than dimming. The clip's
// argument is a comparison, and a comparison you have to remember is one nobody
// makes.

const AXIS_W = Math.min(STAGE_W, 1260);
const ROW_H = 62;
const ROW_GAP = 16;
const LABEL_W = 300;

type Tick = {label: string; frac: number};
type Op = {label: string; value: string; note?: string; role: string; frac: number};
type Step = {
  startMs: number;
  endMs: number;
  show: 'axis' | 'place' | 'read';
  at?: number;
  placed: number[];
};

const roleColour = (theme: ResolvedTheme, role: string): string => {
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

/**
 * How solid a bar of this role is.
 *
 * A neutral row is the reference point rather than the argument — the ordinary
 * case the interesting ones are measured against — so it sits below them. Not as
 * far below as in the budget bar, though: these are 26px rows on a dark stage
 * rather than full-height blocks, and 0.4 alpha on a thin bar is a bar nobody
 * sees. The catalog is consistent about the ORDER, not about one number.
 */
const roleSolidity = (role: string): number => (role === 'neutral' ? 0.55 : 0.9);

export const LatencyScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const ticks = (Array.isArray(props.ticks) ? props.ticks : []) as Tick[];
  const ops = (Array.isArray(props.operations) ? props.operations : []) as Op[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (ticks.length === 0 || ops.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const placed = new Set(step.placed ?? []);
  const enter = interpolate(sinceStep, [0, 18], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // The arriving bar runs out to its mark. Slow enough that the eye follows it
  // past the gridlines it crosses, which is where the "different decade" reading
  // comes from.
  const run = interpolate(sinceStep, [3, 26], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const spoken = step.at !== undefined ? ops[step.at] : undefined;
  const note =
    spoken?.note ??
    (step.show === 'read' ? ops[ops.length - 1]?.note : undefined);

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={26}
      />

      <div style={{width: AXIS_W, opacity: enter}}>
        <div style={{position: 'relative'}}>
          {/* The decade grid. Never fades: on a log axis the gridlines ARE the
              scale, and a bar ending two thirds along means nothing without them. */}
          <div
            style={{
              position: 'absolute',
              left: LABEL_W,
              right: 0,
              top: 0,
              height: ops.length * (ROW_H + ROW_GAP),
            }}
          >
            {ticks.map((t, i) => (
              <div
                key={i}
                style={{
                  position: 'absolute',
                  left: `${t.frac * 100}%`,
                  top: 0,
                  bottom: 0,
                  width: 0,
                  borderLeft: `1px ${i === 0 ? 'solid' : 'dashed'} ${withAlpha(
                    theme.textMuted,
                    i === 0 ? 0.45 : 0.24,
                  )}`,
                }}
              />
            ))}
          </div>

          {ops.map((op, i) => {
            const shown = placed.has(i);
            const isCurrent = step.at === i;
            const colour = roleColour(theme, op.role);
            const w = op.frac * (isCurrent ? run : 1);
            return (
              <div
                key={i}
                style={{
                  height: ROW_H,
                  marginBottom: ROW_GAP,
                  display: 'flex',
                  alignItems: 'center',
                  position: 'relative',
                }}
              >
                <div
                  style={{
                    width: LABEL_W,
                    paddingRight: 20,
                    textAlign: 'right',
                    fontFamily: theme.fontBody,
                    fontSize: 22,
                    color: shown ? theme.text : theme.textMuted,
                    opacity: shown ? 1 : 0.38,
                  }}
                >
                  {op.label}
                </div>
                <div style={{flex: 1, position: 'relative', height: 26}}>
                  {shown ? (
                    <>
                      <div
                        style={{
                          position: 'absolute',
                          left: 0,
                          top: 0,
                          bottom: 0,
                          width: `${w * 100}%`,
                          minWidth: 3,
                          borderRadius: 4,
                          background: withAlpha(colour, roleSolidity(op.role)),
                        }}
                      />
                      <div
                        style={{
                          position: 'absolute',
                          left: `${w * 100}%`,
                          top: -6,
                          marginLeft: 14,
                          whiteSpace: 'nowrap',
                          fontFamily: theme.fontDisplay,
                          fontSize: 30,
                          fontWeight: 800,
                          color: colour,
                          fontVariantNumeric: 'tabular-nums',
                        }}
                      >
                        {op.value}
                      </div>
                    </>
                  ) : null}
                </div>
              </div>
            );
          })}
        </div>

        {/* The tick labels, under the track. */}
        <div style={{position: 'relative', height: 30, marginLeft: LABEL_W}}>
          {ticks.map((t, i) => (
            <span
              key={i}
              style={{
                position: 'absolute',
                left: `${t.frac * 100}%`,
                transform: 'translateX(-50%)',
                fontFamily: theme.fontMono,
                fontSize: 15,
                letterSpacing: 1.4,
                color: theme.textMuted,
                opacity: 0.8,
              }}
            >
              {t.label}
            </span>
          ))}
        </div>
      </div>

      {note ? (
        <div
          style={{
            marginTop: 26,
            maxWidth: 1040,
            textAlign: 'center',
            fontFamily: theme.fontBody,
            fontSize: 24,
            color: theme.textMuted,
            opacity: interpolate(sinceStep, [12, 26], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            }),
          }}
        >
          {note}
        </div>
      ) : null}
    </Stage>
  );
};
