import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// PipelineScene: a grid of stages with work moving through it, one tick at a
// time.
//
// This component draws an occupancy and nothing else. Go simulated the machine
// — which item sits in which stage at every tick — and shipped the full grid
// for each step, so there is exactly one implementation of what a pipeline
// does. Everything here is presentation: where the chip for item 3 was last
// tick, where it is now, and the spring between the two. If this file ever
// starts working out where a chip *should* go, the diagram and the validator
// have become two different machines, and the one on screen is the one nobody
// checked.
//
// The layout is a single row of cells under mono column headers, rather than
// the cycle-by-instruction chart a textbook draws. The chart is better for
// study and worse for film: it is a static table, and the claim being made here
// is about motion — that on one tick, every chip moves and a new one walks in.
// One row makes that a single readable gesture. The cells stay visible and
// empty-looking when nothing is in them, so the grid reads as a machine with
// slots rather than as a row of labels.
//
// The bubble is drawn as a dashed ghost in the shape of a chip, in the limit
// accent. That is the one piece of iconography worth insisting on: a gap drawn
// as nothing reads as a rendering fault, and a gap drawn as a chip-shaped
// absence reads as what it is — a slot doing no work this cycle, which is
// precisely what a stall costs.

const GRID_W = Math.min(STAGE_W, 1500);
const CELL_H = 116;
const CELL_GAP = 16;
const CHIP_INSET = 10;

type Named = {name: string};
type Labelled = {label: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'empty' | 'fill' | 'stall' | 'flow';
  occ: number[];
  bubble: number;
  tick: number;
  retired: number;
};

export const PipelineScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const stages = (Array.isArray(props.stages) ? props.stages : []) as Named[];
  const items = (Array.isArray(props.items) ? props.items : []) as Labelled[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const stall = String(props.stall ?? '');
  const sequentialTicks = Number(props.sequentialTicks ?? 0);
  const pipelinedTicks = Number(props.pipelinedTicks ?? 0);
  if (stages.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;
  const prev = idx > 0 ? steps[idx - 1] : null;

  const n = stages.length;
  const cellW = Math.floor((GRID_W - CELL_GAP * (n - 1)) / n);
  const xOf = (col: number) => col * (cellW + CELL_GAP);
  const chipW = cellW - CHIP_INSET * 2;

  // The tick. Every chip's motion this beat is this one spring.
  const advance = spring({
    frame: sinceStep,
    fps,
    config: {damping: 14, mass: 0.55},
    durationInFrames: 26,
  });

  const occ = Array.isArray(step.occ) ? step.occ : [];
  const prevOcc = prev && Array.isArray(prev.occ) ? prev.occ : [];
  const onStall = step.show === 'stall';
  const onFlow = step.show === 'flow';

  const gridIn = interpolate(frame, [0, 18], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const chips = items.map((it, i) => {
    const cur = occ.indexOf(i);
    const was = prevOcc.indexOf(i);
    if (cur < 0 && was < 0) return null;
    let x: number;
    let opacity: number;
    if (cur >= 0) {
      const from = was >= 0 ? xOf(was) : xOf(cur) - (cellW + CELL_GAP);
      x = interpolate(advance, [0, 1], [from, xOf(cur)]);
      opacity = was >= 0 ? 1 : advance;
    } else {
      // Retired on this tick: it leaves the last stage and the frame.
      x = interpolate(advance, [0, 1], [xOf(was), xOf(was) + cellW + CELL_GAP]);
      opacity = 1 - advance;
    }
    return (
      <div
        key={i}
        style={{
          position: 'absolute',
          left: x + CHIP_INSET,
          top: CHIP_INSET,
          width: chipW,
          height: CELL_H - CHIP_INSET * 2,
          borderRadius: 10,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: withAlpha(theme.accent, 0.2),
          border: `2px solid ${theme.accent}`,
          fontFamily: theme.fontMono,
          fontSize: cellW > 240 ? 26 : 21,
          fontWeight: 700,
          color: theme.accentText,
          opacity,
        }}
      >
        {it.label}
      </div>
    );
  });

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

      <div style={{width: GRID_W, opacity: gridIn}}>
        {/* Column headers. Mono, because a stage name is a machine's word. */}
        <div style={{display: 'flex', gap: CELL_GAP, marginBottom: 14}}>
          {stages.map((s, i) => (
            <div
              key={i}
              style={{
                width: cellW,
                textAlign: 'center',
                fontFamily: theme.fontMono,
                fontSize: 17,
                letterSpacing: 2.6,
                textTransform: 'uppercase',
                color: theme.textMuted,
              }}
            >
              {s.name}
            </div>
          ))}
        </div>

        {/* The slots, and the work moving through them. */}
        <div style={{position: 'relative', height: CELL_H, width: GRID_W}}>
          {stages.map((_, i) => (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: xOf(i),
                top: 0,
                width: cellW,
                height: CELL_H,
                borderRadius: 12,
                background: withAlpha(theme.surface, 0.4),
                border: `1px solid ${theme.surfaceBorder}`,
              }}
            />
          ))}

          {chips}

          {/* The bubble: a chip-shaped absence. */}
          {step.bubble >= 0 ? (
            <div
              style={{
                position: 'absolute',
                left: xOf(step.bubble) + CHIP_INSET,
                top: CHIP_INSET,
                width: chipW,
                height: CELL_H - CHIP_INSET * 2,
                borderRadius: 10,
                border: `2px dashed ${theme.accentLimit}`,
                background: withAlpha(theme.accentLimit, 0.08),
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontFamily: theme.fontMono,
                fontSize: cellW > 240 ? 22 : 18,
                letterSpacing: 2,
                color: theme.accentLimit,
                opacity: advance,
                transform: `scale(${0.94 + 0.06 * advance})`,
              }}
            >
              bubble
            </div>
          ) : null}
        </div>

        {/* The read-out: the clock, then whatever this beat has to say. */}
        <div
          style={{
            marginTop: 34,
            display: 'flex',
            alignItems: 'center',
            gap: 28,
            minHeight: 76,
          }}
        >
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 17,
              letterSpacing: 2.4,
              textTransform: 'uppercase',
              color: theme.textMuted,
              paddingRight: 28,
              borderRight: `1px solid ${theme.line}`,
            }}
          >
            tick {String(step.tick).padStart(2, '0')}
          </div>

          {onStall && stall ? (
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 26,
                color: theme.accentLimit,
                opacity: interpolate(sinceStep, [4, 18], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                }),
              }}
            >
              {stall}
            </div>
          ) : null}

          {onFlow ? (
            <div style={{display: 'flex', alignItems: 'baseline', gap: 34}}>
              <span
                style={{
                  fontFamily: theme.fontBody,
                  fontSize: 24,
                  color: theme.textMuted,
                  opacity: interpolate(sinceStep, [2, 16], [0, 1], {
                    extrapolateLeft: 'clamp',
                    extrapolateRight: 'clamp',
                  }),
                }}
              >
                one at a time,{' '}
                <span style={{fontFamily: theme.fontMono, color: theme.accentRival}}>{sequentialTicks}</span> ticks
              </span>
              <span
                style={{
                  fontFamily: theme.fontBody,
                  fontSize: 30,
                  color: theme.text,
                  opacity: interpolate(sinceStep, [10, 26], [0, 1], {
                    extrapolateLeft: 'clamp',
                    extrapolateRight: 'clamp',
                  }),
                  transform: `translateY(${
                    (1 -
                      spring({
                        frame: sinceStep - 10,
                        fps,
                        config: {damping: 13, mass: 0.55},
                        durationInFrames: 24,
                      })) *
                    8
                  }px)`,
                }}
              >
                overlapped,{' '}
                <span style={{fontFamily: theme.fontMono, fontWeight: 700, color: theme.accentQuantity}}>
                  {pipelinedTicks}
                </span>{' '}
                ticks
              </span>
            </div>
          ) : null}

          {!onStall && !onFlow && step.retired > 0 ? (
            <div style={{fontFamily: theme.fontMono, fontSize: 20, color: theme.textMuted}}>
              {step.retired} finished
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
