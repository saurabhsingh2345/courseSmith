import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// StepperScene: the data on screen, with the rule happening to it.
//
// One row of tiles across the middle of the frame, flags above, indices below,
// and two small pieces of standing chrome — what is being looked for, and how
// much work has been done. That is the whole design, and the restraint is the
// point: the only thing that should ever move is the thing the algorithm just
// did, so a change of any kind reads as the operation rather than as the scene
// rearranging itself.
//
// The tiles are absolutely positioned rather than laid out in a flex row, which
// looks like the harder way to draw a row of boxes and is the only way to draw
// a swap. Two values trading places has to be two tiles CROSSING — one arcing
// over the other — because that arc is the entire visual content of the word
// "swap". In a flex row the same swap is two numerals silently changing, which
// teaches nothing. Each tile's x is interpolated from where its value was, so
// what the eye follows is the value, not the slot.
//
// The array state is never computed here. Go tracks the row through every swap
// and ships the full contents of every cell at every step, so this component
// draws what it is handed. The one thing it derives is where a tile is coming
// FROM during a swap, which is geometry rather than algorithm.
//
// Flags slide between cells on a spring and stack so that low, high and mid on
// the same cell are three readable labels rather than one illegible pile. The
// op counter pops each time it changes, which is the only motion in the frame
// that is not on the row itself, and it is there because the cost is the lesson
// in half the algorithms this template will ever draw.

const ROW_W = Math.min(STAGE_W, 1520);
const CELL_GAP = 14;
const CELL_MAX = 148;
// Three stacked flags plus the stem that points at the cell.
const FLAG_H = 116;
const FLAG_ROW = 30;
const INDEX_H = 44;

type Step = {
  startMs: number;
  endMs: number;
  show: 'array' | 'point' | 'compare' | 'swap' | 'found' | 'done';
  at?: number[];
  values: number[];
  ptr: Record<string, number>;
  ops: number;
  touched: number[];
};

export const StepperScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const pointers = (Array.isArray(props.pointers) ? props.pointers : []) as string[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const target = Number(props.target ?? -1);
  if (steps.length === 0 || !Array.isArray(steps[0].values) || steps[0].values.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const prev = idx > 0 ? steps[idx - 1] : step;
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const values = Array.isArray(step.values) ? step.values : [];
  const n = values.length;
  const cellW = Math.min(CELL_MAX, Math.floor((ROW_W - CELL_GAP * (n - 1)) / n));
  const cellH = Math.round(cellW * 0.84);
  const rowW = cellW * n + CELL_GAP * (n - 1);
  const xOf = (i: number) => i * (cellW + CELL_GAP);

  const acting = Array.isArray(step.at) ? step.at : [];
  const touched = Array.isArray(step.touched) ? step.touched : [];
  const isSwap = step.show === 'swap' && acting.length === 2;
  const isCompare = step.show === 'compare';
  const isFound = step.show === 'found';

  const reveal = step.show === 'array'
    ? (i: number) =>
        interpolate(sinceStep, [3 + i * 3, 15 + i * 3], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : () => 1;

  // The swap arc. Both tiles use the same progress so they cross at the middle.
  const swapT = isSwap
    ? spring({frame: sinceStep - 5, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 26})
    : 1;
  // A slow, shallow pulse on the compared cells. One cycle is about a second,
  // which is a heartbeat rather than a flash.
  const pulse = isCompare ? 0.5 + 0.5 * Math.sin((sinceStep / fps) * Math.PI * 2 * 0.9) : 0;

  const opsPop = step.ops !== prev.ops
    ? spring({frame: sinceStep - 4, fps, config: {damping: 12, mass: 0.5}, durationInFrames: 22})
    : 1;

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

      <div style={{width: rowW, display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 18}}>
        <span style={{fontFamily: theme.fontMono, fontSize: 19, letterSpacing: 2.6, color: theme.textMuted}}>
          {target >= 0 ? `looking for ${target}` : 'sorting in place'}
        </span>
        <span
          style={{
            fontFamily: theme.fontMono,
            fontSize: 19,
            letterSpacing: 2.6,
            color: theme.textMuted,
            transform: `scale(${0.94 + 0.06 * opsPop})`,
            transformOrigin: 'right center',
          }}
        >
          ops <span style={{color: theme.accentText, fontWeight: 700, fontSize: 24}}>{step.ops ?? 0}</span>
        </span>
      </div>

      <div style={{position: 'relative', width: rowW, height: FLAG_H + cellH + INDEX_H}}>
        {/* The flags. They slide between cells rather than cutting. */}
        {pointers.map((name, p) => {
          const cur = step.ptr?.[name];
          const was = prev.ptr?.[name];
          if (cur === undefined || cur < 0) return null;
          const from = was === undefined || was < 0 ? cur : was;
          const slide = spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.5}, durationInFrames: 24});
          const cell = from + (cur - from) * slide;
          const appear = was === undefined || was < 0 ? slide : 1;
          const cx = xOf(cell) + cellW / 2;
          const y = FLAG_H - FLAG_ROW * (pointers.length - p) - 6;
          return (
            <div
              key={name}
              style={{
                position: 'absolute',
                left: cx,
                top: y,
                transform: 'translateX(-50%)',
                opacity: appear,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
              }}
            >
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 17,
                  fontWeight: 700,
                  letterSpacing: 1.2,
                  color: theme.ink,
                  background: theme.accent,
                  borderRadius: 5,
                  padding: '2px 9px',
                  whiteSpace: 'nowrap',
                }}
              >
                {name}
              </span>
              <div style={{width: 1.5, height: FLAG_ROW * (pointers.length - p) - 16, background: withAlpha(theme.accent, 0.5)}} />
            </div>
          );
        })}

        {/* The row. */}
        {values.map((v, i) => {
          const active = acting.includes(i);
          // During a swap the tile at i holds the value that was at the other
          // named cell, so it travels from there.
          let x = xOf(i);
          let lift = 0;
          if (isSwap && active) {
            const other = acting[0] === i ? acting[1] : acting[0];
            x = xOf(other) + (xOf(i) - xOf(other)) * swapT;
            const arc = Math.sin(swapT * Math.PI);
            lift = (acting[0] === i ? -1 : 1) * arc * (cellH * 0.62);
          }
          const found = isFound && active;
          const compared = isCompare && active;
          const seen = touched.includes(i);
          const border = found
            ? theme.accentQuantity
            : compared
              ? withAlpha(theme.accent, 0.55 + 0.45 * pulse)
              : isSwap && active
                ? theme.accent
                : seen
                  ? withAlpha(theme.surfaceBorder, 1)
                  : theme.surfaceBorder;
          const scale = found
            ? 1 + 0.1 * spring({frame: sinceStep - 4, fps, config: {damping: 12, mass: 0.5}, durationInFrames: 24})
            : compared
              ? 1 + 0.035 * pulse
              : 1;
          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: x,
                top: FLAG_H + lift,
                width: cellW,
                height: cellH,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                borderRadius: 12,
                background: found ? theme.accentQuantity : withAlpha(theme.surface, seen ? 0.95 : 0.72),
                border: `2px solid ${border}`,
                // The one glow: the cell the search lands on.
                boxShadow: found ? `0 0 34px ${withAlpha(theme.accentQuantity, 0.45)}` : undefined,
                transform: `scale(${scale})`,
                opacity: reveal(i),
                fontFamily: theme.fontMono,
                fontSize: Math.round(cellW * 0.4),
                fontWeight: 700,
                letterSpacing: -1,
                color: found ? theme.ink : seen ? theme.text : theme.textMuted,
              }}
            >
              {v}
            </div>
          );
        })}

        {/* The index rail, so a beat that names cell 4 can be checked. */}
        {values.map((_, i) => (
          <div
            key={`ix-${i}`}
            style={{
              position: 'absolute',
              left: xOf(i),
              top: FLAG_H + cellH + 14,
              width: cellW,
              textAlign: 'center',
              fontFamily: theme.fontMono,
              fontSize: 17,
              color: acting.includes(i) ? theme.accentText : withAlpha(theme.line, 0.75),
              opacity: reveal(i),
            }}
          >
            {i}
          </div>
        ))}
      </div>
    </Stage>
  );
};
