import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// SchedulerScene: lanes, one shared axis, and exactly one of them filled.
//
// The composition is a Gantt chart and it is deliberately not dressed up,
// because the whole misunderstanding this clip corrects is spatial: a beginner
// pictures programs running side by side, and the only thing that dislodges
// that is a chart where the horizontal axis is unambiguously TIME and a
// vertical slice through it hits at most one block. So the lanes share one
// axis, the axis carries counted unit ticks in mono, and no lane is ever
// allowed to overlap another in x. Everything else follows.
//
// Waiting is drawn. Each lane carries a faint rule across every unit that has
// elapsed, and the block sits on top of it, so a lane that has not run yet is
// not empty space — it is a visible stretch of time that process spent not
// running. That rule is most of the pedagogy: the filled blocks are the part
// people already believe.
//
// Blocks extend rather than appear. A turn lands by growing rightward from the
// previous block's right edge, which is what makes the axis feel like time
// passing instead of a bar chart being populated. The extending block is the
// one glow in the frame and it is the only thing moving.
//
// The switch beat draws a narrow band at the boundary, in accentLimit, because
// the point of the beat is that the changeover COSTS — it is a ceiling on how
// finely you can slice, not a free operation. It is drawn over the blocks
// rather than between them so the chart does not reflow to make room for it.
//
// The totals land as chips at the lane ends on the closer, and every number in
// them was summed in Go.

const BOARD_W = Math.min(STAGE_W, 1560);
const LABEL_W = 190;
const TOTALS_W = 150;
const TRACK_W = BOARD_W - LABEL_W - TOTALS_W;
const LANE_H = 92;
const LANE_GAP = 16;
const AXIS_H = 56;
const BLOCK_INSET = 10;

type Proc = {label: string; total: number};
type Slot = {proc: number; len: number; start: number};
type Step = {
  startMs: number;
  endMs: number;
  show: 'queue' | 'run' | 'switch' | 'fair';
  at?: number;
  boundary?: number;
  laid: number;
};

export const SchedulerScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const policy = String(props.policy ?? '');
  const procs = (Array.isArray(props.procs) ? props.procs : []) as Proc[];
  const slots = (Array.isArray(props.slots) ? props.slots : []) as Slot[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const units = Math.max(1, Number(props.units ?? 1));
  if (procs.length === 0 || slots.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const extend = spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 28});
  const settle = spring({frame: sinceStep - 4, fps, config: {damping: 12, mass: 0.5}, durationInFrames: 24});

  const unitW = TRACK_W / units;
  const laneTop = (i: number): number => i * (LANE_H + LANE_GAP);
  const chartH = procs.length * (LANE_H + LANE_GAP) - LANE_GAP;
  const running = step.show === 'run' ? step.at ?? -1 : -1;
  const tally = step.show === 'fair';
  const queued = step.show === 'queue';

  // The lane a block belongs to decides its colour, so a viewer can attribute a
  // block without reading it.
  const laneInk = (i: number): string =>
    [theme.accentQuantity, theme.accentRival, theme.accent, theme.mass][i % 4];

  // How far along the axis the clip has got, including the block extending now.
  const elapsed = ((): number => {
    const done = slots.slice(0, step.laid).reduce((a, s) => a + s.len, 0);
    if (running < 0) return done;
    const cur = slots[running];
    return done - cur.len + cur.len * extend;
  })();

  const tickEvery = units > 12 ? 2 : 1;

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={18}
      />

      {/* The policy nameplate, over the chart and on its left axis. */}
      <div style={{width: BOARD_W, display: 'flex', alignItems: 'baseline', gap: 16, marginBottom: 18}}>
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 15,
            letterSpacing: 3,
            textTransform: 'uppercase',
            color: theme.textMuted,
          }}
        >
          policy
        </div>
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 30,
            fontWeight: 700,
            letterSpacing: -0.4,
            color: theme.accentText,
          }}
        >
          {policy}
        </div>
      </div>

      <div style={{position: 'relative', width: BOARD_W, height: chartH + AXIS_H}}>
        {/* The lanes. */}
        {procs.map((proc, i) => {
          const top = laneTop(i);
          const isRunning = running >= 0 && slots[running].proc === i;
          return (
            <div key={i}>
              <div
                style={{
                  position: 'absolute',
                  left: 0,
                  top,
                  width: LABEL_W - 22,
                  height: LANE_H,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'flex-end',
                  fontFamily: theme.fontDisplay,
                  fontSize: 28,
                  fontWeight: 700,
                  letterSpacing: -0.3,
                  color: isRunning ? theme.text : theme.textMuted,
                  opacity: isRunning ? 1 : 0.75,
                }}
              >
                {proc.label}
              </div>

              {/* The lane's track, and the waiting rule across the time that
                  has already passed. */}
              <div
                style={{
                  position: 'absolute',
                  left: LABEL_W,
                  top,
                  width: TRACK_W,
                  height: LANE_H,
                  borderRadius: 10,
                  background: withAlpha(theme.surface, 0.42),
                  border: `1.5px solid ${withAlpha(theme.surfaceBorder, 0.8)}`,
                }}
              />
              <div
                style={{
                  position: 'absolute',
                  left: LABEL_W,
                  top: top + LANE_H / 2 - 1,
                  width: unitW * elapsed,
                  height: 2,
                  background: withAlpha(theme.line, 0.35),
                }}
              />

              {/* The waiting chip, only while nothing has been scheduled. */}
              {queued ? (
                <div
                  style={{
                    position: 'absolute',
                    left: LABEL_W + 16,
                    top: top + LANE_H / 2 - 17,
                    paddingInline: 14,
                    paddingBlock: 7,
                    borderRadius: 8,
                    border: `1.5px dashed ${withAlpha(laneInk(i), 0.7)}`,
                    fontFamily: theme.fontMono,
                    fontSize: 16,
                    letterSpacing: 2,
                    textTransform: 'uppercase',
                    color: theme.textMuted,
                    opacity: interpolate(sinceStep, [2 + i * 5, 16 + i * 5], [0, 1], {
                      extrapolateLeft: 'clamp',
                      extrapolateRight: 'clamp',
                    }),
                  }}
                >
                  waiting
                </div>
              ) : null}

              {/* The lane's total, landing on the closer. */}
              {tally ? (
                <div
                  style={{
                    position: 'absolute',
                    left: LABEL_W + TRACK_W + 22,
                    top: top + LANE_H / 2 - 22,
                    paddingInline: 14,
                    paddingBlock: 8,
                    borderRadius: 9,
                    background: withAlpha(laneInk(i), 0.16),
                    border: `2px solid ${withAlpha(laneInk(i), 0.7)}`,
                    fontFamily: theme.fontMono,
                    fontSize: 22,
                    fontWeight: 700,
                    color: theme.text,
                    opacity: interpolate(sinceStep, [4 + i * 6, 20 + i * 6], [0, 1], {
                      extrapolateLeft: 'clamp',
                      extrapolateRight: 'clamp',
                    }),
                    transform: `translateX(${(1 - settle) * 14}px)`,
                  }}
                >
                  {`${proc.total}u`}
                </div>
              ) : null}
            </div>
          );
        })}

        {/* The blocks. */}
        {slots.map((slot, i) => {
          if (i >= step.laid) return null;
          const isRunning = i === running;
          const grow = isRunning ? extend : 1;
          const w = Math.max(0, slot.len * unitW * grow - BLOCK_INSET);
          const top = laneTop(slot.proc);
          const ink = laneInk(slot.proc);
          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: LABEL_W + slot.start * unitW + BLOCK_INSET / 2,
                top: top + 12,
                width: w,
                height: LANE_H - 24,
                borderRadius: 8,
                background: withAlpha(ink, isRunning ? 0.9 : 0.72),
                border: `2px solid ${ink}`,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                overflow: 'hidden',
                // The one glow: the turn extending right now.
                boxShadow: isRunning ? `0 0 26px ${withAlpha(ink, 0.55)}` : undefined,
              }}
            >
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 20,
                  fontWeight: 700,
                  color: theme.ink,
                  opacity: grow > 0.6 ? 1 : 0,
                  whiteSpace: 'nowrap',
                }}
              >
                {`${slot.len}u`}
              </span>
            </div>
          );
        })}

        {/* The context switch: a band over the boundary, with its cost. */}
        {step.show === 'switch' && step.boundary !== undefined ? (
          <>
            <div
              style={{
                position: 'absolute',
                left: LABEL_W + step.boundary * unitW - 9,
                top: -8,
                width: 18,
                height: chartH + 16,
                borderRadius: 5,
                background: withAlpha(theme.accentLimit, 0.22),
                border: `2px solid ${theme.accentLimit}`,
                opacity: settle,
                transform: `scaleY(${0.9 + 0.1 * settle})`,
              }}
            />
            <div
              style={{
                position: 'absolute',
                left: Math.min(BOARD_W - 320, LABEL_W + step.boundary * unitW + 22),
                top: chartH + 8,
                width: 320,
                fontFamily: theme.fontMono,
                fontSize: 17,
                lineHeight: 1.3,
                color: theme.accentLimit,
                opacity: settle,
              }}
            >
              context switch: registers saved, caches cold
            </div>
          </>
        ) : null}

        {/* The time axis. */}
        <div
          style={{
            position: 'absolute',
            left: LABEL_W,
            top: chartH + 14,
            width: TRACK_W,
            height: 2,
            background: withAlpha(theme.line, 0.45),
          }}
        />
        {Array.from({length: units + 1}).map((_, u) => (
          <div key={u}>
            <div
              style={{
                position: 'absolute',
                left: LABEL_W + u * unitW,
                top: chartH + 14,
                width: 2,
                height: u % tickEvery === 0 ? 10 : 5,
                background: withAlpha(theme.line, u <= elapsed ? 0.7 : 0.35),
              }}
            />
            {u % tickEvery === 0 ? (
              <div
                style={{
                  position: 'absolute',
                  left: LABEL_W + u * unitW - 20,
                  top: chartH + 28,
                  width: 40,
                  textAlign: 'center',
                  fontFamily: theme.fontMono,
                  fontSize: 15,
                  color: u <= elapsed ? theme.textMuted : withAlpha(theme.textMuted, 0.45),
                }}
              >
                {u}
              </div>
            ) : null}
          </div>
        ))}
      </div>
    </Stage>
  );
};
