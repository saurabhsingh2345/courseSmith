import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// OccupancyScene is a population of identical units with part of it claimed.
//
// The composition is one grid, drawn once, held for the whole clip. Cells start
// unclaimed and bands light them in place. Nothing moves, nothing re-lays out,
// and the grid never redraws — which is the only way the picture reads as "the
// same set of things, seen again" rather than as a new chart each beat.
//
// Three decisions carry it.
//
// A band's cells are a contiguous run, assigned in Go (see occupancyScenes).
// Scattering them would be more literally true of, say, memory fragmentation,
// but the eye reads a block as a quantity and a sprinkle as noise — and the
// whole job of this frame is making a quantity legible at a glance.
//
// The cell size is derived from the grid's own shape rather than fixed, because
// the template's range runs from twelve cells to twelve hundred. One size cannot
// serve both: at twelve the cells would be dots lost in the frame, at twelve
// hundred they would not fit. The gap shrinks with the cell so a dense grid
// still reads as separate units rather than as a filled rectangle.
//
// Unclaimed cells are drawn, not omitted. The argument this template makes is
// always partly about the part that ISN'T active — the experts you still pay to
// keep in memory — so the remainder has to be visibly present and visibly
// unlit.

/** The box the grid is fitted into. */
const GRID_W = Math.min(STAGE_W, 1360);
const GRID_H = 470;

type Band = {count: number; from: number; label: string; note?: string; role: string};
type Step = {startMs: number; endMs: number; show: 'grid' | 'fill' | 'read'; at?: number};

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
 * How bright a lit cell of this role gets.
 *
 * A neutral band is context, and context that renders at full strength wins the
 * frame on area alone: a hundred and twenty neutral cells at full opacity are
 * visually louder than sixteen gold ones, which inverts the argument the clip is
 * making. So neutral lights enough to be legibly *claimed* and no further, and
 * the roles that carry meaning keep the top of the range to themselves.
 */
const roleLit = (role: string): number => (role === 'neutral' ? 0.4 : 1);

/**
 * Cell geometry for a grid of `cols`×`rows` inside the box.
 *
 * The gap is a fraction of the cell rather than a constant: at twelve cells a
 * 4px gap is invisible and at twelve hundred it is most of the picture.
 */
const cellGeometry = (cols: number, rows: number) => {
  const byWidth = GRID_W / cols;
  const byHeight = GRID_H / rows;
  const pitch = Math.min(byWidth, byHeight);
  const gap = Math.max(1, Math.min(6, pitch * 0.16));
  const size = Math.max(2, pitch - gap);
  return {size, gap, pitch};
};

export const OccupancyScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const total = Number(props.total ?? 0);
  const cols = Number(props.cols ?? 1);
  const rows = Number(props.rows ?? 1);
  const unit = String(props.unit ?? '');
  const label = String(props.label ?? '');
  const bands = (Array.isArray(props.bands) ? props.bands : []) as Band[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (total <= 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // Which bands are lit by now, and when each was lit — a band keeps its colour
  // once shown, so the grid accumulates and the final frame carries the whole
  // argument.
  const litAt = new Map<number, number>();
  steps.slice(0, idx + 1).forEach((s) => {
    if (s.show === 'fill' && s.at !== undefined && !litAt.has(s.at)) {
      litAt.set(s.at, s.startMs);
    }
  });
  const current = step.show === 'fill' ? (step.at ?? -1) : -1;

  const {size, gap, pitch} = cellGeometry(cols, rows);

  // The grid arriving. Cells fade in over a short stagger read from their index,
  // so a large population assembles as a sweep rather than as a hard cut.
  const gridIn = spring({
    frame: ((nowMs - steps[0].startMs) / 1000) * FPS,
    fps,
    config: {damping: 200, mass: 0.7},
    durationInFrames: 26,
  });

  /** Which band owns a cell, and how lit that cell is right now. */
  const cellState = (i: number): {colour: string; alpha: number} => {
    for (const [bandIndex, startMs] of litAt) {
      const band = bands[bandIndex];
      if (!band) continue;
      if (i < band.from || i >= band.from + band.count) continue;
      const since = ((nowMs - startMs) / 1000) * FPS;
      // Cells within a band light left-to-right, fast: the sweep is a gesture,
      // not an animation to watch.
      const offset = ((i - band.from) / Math.max(1, band.count)) * 10;
      const on = interpolate(since - offset, [0, 9], [0, 1], {
        extrapolateLeft: 'clamp',
        extrapolateRight: 'clamp',
      });
      const top = roleLit(band.role);
      return {colour: roleColour(theme, band.role), alpha: 0.16 + on * (top - 0.16)};
    }
    return {colour: theme.textMuted, alpha: 0.16};
  };

  // The reading under the grid: the current band's count, its label and note.
  // On the closing "read" beat the last band stays up, because a frame that
  // empties its caption at the end is a frame that ends on a picture nobody can
  // name.
  const spoken = current >= 0 ? current : [...litAt.keys()].pop() ?? -1;
  const band = spoken >= 0 ? bands[spoken] : undefined;
  const captionIn = interpolate(sinceStep, [2, 16], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // The count runs up with the cells rather than snapping, so the number and the
  // sweep land together.
  const countT = current >= 0
    ? interpolate(sinceStep, [0, 20], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : 1;

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={34}
      />

      {/* The population's own name and size, above the grid. It is set in the
          mono face at caption size because it is a label on the picture rather
          than a second headline competing with the first. */}
      <div
        style={{
          fontFamily: theme.fontMono,
          fontSize: 16,
          letterSpacing: 3.2,
          textTransform: 'uppercase',
          color: theme.textMuted,
          opacity: gridIn * 0.85,
          marginBottom: 16,
          textAlign: 'center',
        }}
      >
        {/* The label already names the population in the plural ("the model's
            experts"), so the unit is only set when there is no label to carry
            it. Rendering both gave "896 EXPERT · THE MODEL'S EXPERTS", and
            pluralising the unit here would mean a naive -s rule deciding how to
            spell somebody's subject. */}
        {total.toLocaleString()} {label || unit}
      </div>

      <div
        style={{
          width: cols * pitch,
          display: 'flex',
          flexWrap: 'wrap',
          justifyContent: 'flex-start',
          opacity: gridIn,
        }}
      >
        {Array.from({length: total}, (_, i) => {
          const {colour, alpha} = cellState(i);
          return (
            <div
              key={i}
              style={{
                width: size,
                height: size,
                margin: gap / 2,
                borderRadius: Math.min(3, size * 0.18),
                background: withAlpha(colour, alpha),
                // A hairline only while the cells are big enough to carry one;
                // on a dense grid it would be most of the cell.
                border: size >= 14 ? `1px solid ${withAlpha(colour, alpha * 0.7)}` : undefined,
              }}
            />
          );
        })}
      </div>

      {band ? (
        <div style={{marginTop: 34, textAlign: 'center', opacity: captionIn}}>
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 62,
              fontWeight: 800,
              lineHeight: 1,
              color: roleColour(theme, band.role),
              fontVariantNumeric: 'tabular-nums',
            }}
          >
            {Math.round(band.count * countT).toLocaleString()}
            <span
              style={{
                fontFamily: theme.fontMono,
                fontSize: 22,
                fontWeight: 400,
                color: theme.textMuted,
                marginLeft: 10,
              }}
            >
              {band.label}
            </span>
          </div>
          {band.note ? (
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 25,
                color: theme.textMuted,
                marginTop: 14,
                maxWidth: 1080,
              }}
            >
              {band.note}
            </div>
          ) : null}
        </div>
      ) : null}
    </Stage>
  );
};
