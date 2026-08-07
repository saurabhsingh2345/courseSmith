import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// LadderScene: the memory hierarchy drawn on the only axis that tells the
// truth about it.
//
// A register is about a nanosecond and a disk seek is about ten million, and
// on a linear axis that is one invisible bar and one bar the width of the
// screen — a picture that is technically correct and teaches nothing, because
// every interesting level in the middle collapses to zero. So the axis is
// logarithmic, and the decade ticks are drawn ON it, labelled, so the viewer
// can see that each step right is a factor of ten. Without the ticks a log
// axis is a lie of a different kind: bars that look comparable and are not.
//
// Go precomputes every position (logPos, 0 at the fastest rung and 1 at the
// slowest) and every label, so this component never takes a logarithm. The
// same rule as RadixScene: a diagram that computes something and gets it wrong
// is worse than no diagram, and the arithmetic belongs where it can be tested.
//
// Bars grow rightward down the stack, each with its latency set at the tip and
// its capacity as a chip pinned to the right margin. The capacity column is
// flush-right on purpose — it is the counter-argument to the whole picture
// (slow things are big) and putting it on its own axis lets the eye read the
// trade in one sweep instead of hunting for it inside each row.
//
// A miss is the one moment with real physics. The request dot spring-DROPS
// from the rung that missed to the rung below it, overshooting slightly, with
// the cost of the fall stated beside it. A cache miss is a fall, and the only
// honest way to show a fall is to let something accelerate.
//
// One glow maximum: the request dot mid-drop.

const BOARD_W = Math.min(STAGE_W, 1500);
const LABEL_W = 220;
const CAP_W = 200;
const LANE_W = 46;
const GAP = 18;
const AXIS_W = BOARD_W - LABEL_W - CAP_W - LANE_W - GAP * 3;
// The fastest rung still needs a bar you can see and label.
const MIN_BAR = 118;

type Rung = {label: string; capacity: string; latencyNs: number; latency: string; logPos: number};
type Tick = {pos: number; label: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'ladder' | 'rung' | 'miss' | 'spread';
  at?: number;
  to?: number;
  cost?: string;
};

const barW = (logPos: number): number => MIN_BAR + Math.max(0, Math.min(1, logPos)) * (AXIS_W - MIN_BAR);

export const LadderScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const ratio = String(props.ratio ?? '');
  const rungs = (Array.isArray(props.rungs) ? props.rungs : []) as Rung[];
  const ticks = (Array.isArray(props.ticks) ? props.ticks : []) as Tick[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (rungs.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const focused = step.show === 'rung' ? step.at ?? -1 : -1;
  const missFrom = step.show === 'miss' ? step.at ?? -1 : -1;
  const missTo = step.show === 'miss' ? step.to ?? -1 : -1;
  const spread = step.show === 'spread';

  const rowH = rungs.length > 6 ? 62 : 76;
  const barH = rungs.length > 6 ? 28 : 34;
  const boardH = rungs.length * rowH;
  const rowMid = (i: number): number => i * rowH + rowH / 2;

  // Bars draw in on the opener, then stay.
  const grow = idx === 0 ? interpolate(sinceStep, [2, 26], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'}) : 1;
  const focus = spring({frame: sinceStep - 2, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 26});
  // The drop: a lower damping than anything else in the catalog, because a
  // fall that does not overshoot is an elevator, not a fall.
  const drop =
    missFrom >= 0 ? spring({frame: sinceStep - 4, fps, config: {damping: 12, mass: 0.6}, durationInFrames: 26}) : 0;
  const spreadIn = spread
    ? spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 28})
    : 0;

  const dotY = missFrom >= 0 && missTo >= 0 ? rowMid(missFrom) + (rowMid(missTo) - rowMid(missFrom)) * drop : 0;
  const axisLeft = LABEL_W + GAP + LANE_W + GAP;

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={20}
      />

      <div style={{width: BOARD_W}}>
        {/* Decade ticks, labelled, above the stack. Without these the log axis
            is unreadable and the bars are misleading. */}
        <div style={{position: 'relative', height: 30, marginBottom: 8}}>
          {ticks.map((t, i) => (
            <span
              key={i}
              style={{
                position: 'absolute',
                left: axisLeft + barW(t.pos),
                top: 6,
                transform: 'translateX(-50%)',
                fontFamily: theme.fontMono,
                fontSize: 15,
                letterSpacing: 1,
                color: withAlpha(theme.textMuted, 0.85),
                whiteSpace: 'nowrap',
              }}
            >
              {t.label}
            </span>
          ))}
        </div>

        <div style={{position: 'relative', height: boardH}}>
          {/* Tick rules run the full stack, behind the bars. */}
          <svg
            width={BOARD_W}
            height={boardH}
            style={{position: 'absolute', left: 0, top: 0, overflow: 'visible', pointerEvents: 'none'}}
          >
            {ticks.map((t, i) => (
              <line
                key={i}
                x1={axisLeft + barW(t.pos)}
                y1={0}
                x2={axisLeft + barW(t.pos)}
                y2={boardH}
                stroke={withAlpha(theme.line, 0.16)}
                strokeWidth={1.5}
                strokeDasharray="3 8"
              />
            ))}
            {/* The fall path, so the drop has somewhere to have come from. */}
            {missFrom >= 0 && missTo >= 0 ? (
              <line
                x1={LABEL_W + GAP + LANE_W / 2}
                y1={rowMid(missFrom)}
                x2={LABEL_W + GAP + LANE_W / 2}
                y2={dotY}
                stroke={withAlpha(theme.accentLimit, 0.5)}
                strokeWidth={2.5}
                strokeLinecap="round"
                strokeDasharray="2 7"
              />
            ) : null}
          </svg>

          {rungs.map((r, i) => {
            const isFocus = i === focused;
            const isSource = i === missFrom;
            const isTarget = i === missTo;
            const hot = isFocus || isTarget || spread;
            const w = barW(r.logPos) * grow;
            const lift = isFocus ? focus : 0;
            const fillColour = isTarget
              ? withAlpha(theme.accentLimit, 0.55 + 0.3 * drop)
              : isFocus
                ? withAlpha(theme.accent, 0.5 + 0.4 * lift)
                : spread
                  ? withAlpha(theme.accent, 0.3 + 0.2 * spreadIn)
                  : withAlpha(theme.mass, 0.16);
            return (
              <div key={i} style={{position: 'absolute', left: 0, top: i * rowH, width: BOARD_W, height: rowH}}>
                <div
                  style={{
                    position: 'absolute',
                    left: 0,
                    top: 0,
                    width: LABEL_W,
                    height: rowH,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'flex-end',
                    paddingRight: 4,
                    fontFamily: theme.fontDisplay,
                    fontSize: rungs.length > 6 ? 24 : 27,
                    fontWeight: 700,
                    letterSpacing: -0.4,
                    textAlign: 'right',
                    color: hot ? theme.text : theme.textMuted,
                    opacity: isSource && drop > 0.3 ? 0.55 : 1,
                  }}
                >
                  {r.label}
                </div>

                <div
                  style={{
                    position: 'absolute',
                    left: axisLeft,
                    top: (rowH - barH) / 2,
                    width: w,
                    height: barH,
                    borderRadius: barH / 2,
                    background: fillColour,
                    border: `1.5px solid ${hot ? withAlpha(theme.accent, 0.5) : withAlpha(theme.surfaceBorder, 0.9)}`,
                    transform: `scaleY(${1 + 0.12 * lift})`,
                    transformOrigin: 'center left',
                  }}
                />
                <span
                  style={{
                    position: 'absolute',
                    left: axisLeft + w + 14,
                    top: 0,
                    height: rowH,
                    display: 'flex',
                    alignItems: 'center',
                    fontFamily: theme.fontMono,
                    fontSize: rungs.length > 6 ? 18 : 20,
                    fontWeight: 700,
                    color: hot ? theme.accentText : theme.textMuted,
                    whiteSpace: 'nowrap',
                    opacity: grow,
                  }}
                >
                  {r.latency}
                </span>

                {r.capacity ? (
                  <span
                    style={{
                      position: 'absolute',
                      right: 0,
                      top: (rowH - 36) / 2,
                      height: 36,
                      display: 'flex',
                      alignItems: 'center',
                      paddingInline: 14,
                      borderRadius: 8,
                      background: withAlpha(theme.surface, 0.85),
                      border: `1.5px solid ${hot ? withAlpha(theme.accent, 0.35) : theme.surfaceBorder}`,
                      fontFamily: theme.fontMono,
                      fontSize: 18,
                      color: hot ? theme.text : theme.textMuted,
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {r.capacity}
                  </span>
                ) : null}
              </div>
            );
          })}

          {/* The request, falling. */}
          {missFrom >= 0 && missTo >= 0 ? (
            <div
              style={{
                position: 'absolute',
                left: LABEL_W + GAP + LANE_W / 2 - 13,
                top: dotY - 13,
                width: 26,
                height: 26,
                borderRadius: 13,
                background: theme.accentLimit,
                // The one glow: the thing that is moving.
                boxShadow: `0 0 26px ${withAlpha(theme.accentLimit, 0.6)}`,
              }}
            />
          ) : null}
        </div>

        {/* The cost of the fall, or the whole spread. Same strip, one at a
            time, so the bottom of the frame never moves. */}
        <div style={{height: 62, marginTop: 14, display: 'flex', alignItems: 'center', gap: 14}}>
          {missFrom >= 0 && step.cost ? (
            <span
              style={{
                paddingInline: 20,
                paddingBlock: 11,
                borderRadius: 10,
                background: withAlpha(theme.accentLimit, 0.14),
                border: `2px solid ${withAlpha(theme.accentLimit, 0.5)}`,
                fontFamily: theme.fontMono,
                fontSize: 26,
                fontWeight: 700,
                color: theme.text,
                opacity: drop,
                transform: `translateY(${(1 - drop) * 10}px)`,
              }}
            >
              miss → {step.cost}
            </span>
          ) : null}
          {spread && ratio ? (
            <span
              style={{
                paddingInline: 20,
                paddingBlock: 11,
                borderRadius: 10,
                background: withAlpha(theme.accent, 0.12),
                border: `2px solid ${withAlpha(theme.accent, 0.45)}`,
                fontFamily: theme.fontMono,
                fontSize: 26,
                fontWeight: 700,
                color: theme.accentText,
                opacity: spreadIn,
                transform: `translateY(${(1 - spreadIn) * 10}px)`,
              }}
            >
              top to bottom: {ratio}
            </span>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
