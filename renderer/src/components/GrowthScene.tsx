import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// GrowthScene: the shape of a complexity class, before the symbol.
//
// A wide chart, low on the frame, with a lot of empty space above the flat
// curves. The emptiness is the composition: the whole argument of this clip is
// that one line stays down there while another leaves the top of the frame, and
// that comparison only lands if there is a visible top of the frame for a curve
// to leave. A chart scaled so every curve fits neatly would be a chart that
// argues the opposite of what the narration is saying.
//
// Curves draw on from the origin with a strokeDashoffset sweep, one per beat, in
// growth order. Watching them arrive slowest-first means the viewer sees the
// hierarchy assemble rather than being handed it, and each new line is read
// against the ones already there. The class notation sits at the end of its own
// curve in mono — at the point where the curve LEAVES the chart, for the ones
// that do, because that exit point is the most informative place on the line.
//
// The x axis is deliberately unlabelled in numbers. The visual domain is small
// (two dozen items) so the curves separate immediately, while the probe reads at
// a real input size that would be nowhere near it — a million. Putting numbers
// on the axis would make those two facts contradict each other on screen. The
// axis says "n, growing", the probe says "at a million", and the readings under
// the chart carry every figure that matters. All of them arrive precomputed and
// preformatted from Go.
//
// One glow, on the closer: the fastest-growing curve and its reading are lifted
// into accentLimit while everything else steps back, which is the verdict the
// moral beat is speaking over.

const CHART_W = Math.min(STAGE_W, 1420);
const CHART_H = 470;
const PAD_L = 92;
const PAD_R = 168;
const PAD_T = 34;
const PAD_B = 64;
// Where the drop-line stands. Not at the right edge: the readings need the
// curves to still be climbing beside them, and a line flush with the axis
// arrow reads as the end of the chart rather than as a measurement.
const PROBE_X_FRAC = 0.78;

type Curve = {class: string; label: string; notation: string; points: number[]; reading?: string};
type Step = {startMs: number; endMs: number; show: 'axes' | 'curve' | 'probe' | 'moral'; at?: number; drawn: number[]};

const PLOT_W = CHART_W - PAD_L - PAD_R;
const PLOT_H = CHART_H - PAD_T - PAD_B;
const PLOT_BOTTOM = PAD_T + PLOT_H;

const xOf = (i: number, n: number) => PAD_L + (n <= 1 ? 0 : (i / (n - 1)) * PLOT_W);
const yOf = (v: number) => PLOT_BOTTOM - Math.max(0, Math.min(1, v)) * PLOT_H;

const pathOf = (points: number[]): string =>
  points.map((v, i) => `${i === 0 ? 'M' : 'L'} ${xOf(i, points.length).toFixed(1)} ${yOf(v).toFixed(1)}`).join(' ');

/** Rough length of the polyline, for the draw-on sweep. */
const lengthOf = (points: number[]): number => {
  let total = 0;
  for (let i = 1; i < points.length; i++) {
    const dx = xOf(i, points.length) - xOf(i - 1, points.length);
    const dy = yOf(points[i]) - yOf(points[i - 1]);
    total += Math.sqrt(dx * dx + dy * dy);
  }
  return Math.max(1, total);
};

/** Where the notation sits: the exit point for a curve that leaves the frame. */
const labelIndex = (points: number[]): number => {
  const exit = points.findIndex((v) => v >= 0.999);
  return exit < 0 ? points.length - 1 : exit;
};

export const GrowthScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const curves = (Array.isArray(props.curves) ? props.curves : []) as Curve[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const probeLabel = String(props.probeLabel ?? '');
  const worst = Number(props.worst ?? curves.length - 1);
  if (curves.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const drawn = Array.isArray(step.drawn) ? step.drawn : [];
  const drawing = step.show === 'curve' ? (step.at ?? -1) : -1;
  const isAxes = step.show === 'axes';
  const isMoral = step.show === 'moral';
  // The probe stays up once dropped: the readings are the evidence the moral
  // is spoken over.
  const probeUp = step.show === 'probe' || isMoral;
  const probeIn = step.show === 'probe'
    ? spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 26})
    : isMoral
      ? 1
      : 0;

  const axesIn = isAxes
    ? interpolate(sinceStep, [2, 26], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : 1;
  const probeX = PAD_L + PLOT_W * PROBE_X_FRAC;

  // The fastest-growing curve is the limit; the rest are the field.
  const palette = [theme.accent, theme.accentQuantity, theme.accentRival];
  const colourOf = (i: number) => (i === worst ? theme.accentLimit : palette[i % palette.length]);
  const dimmed = (i: number) => isMoral && i !== worst;

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

      <svg width={CHART_W} height={CHART_H} viewBox={`0 0 ${CHART_W} ${CHART_H}`} style={{flexShrink: 0, overflow: 'visible'}}>
        {/* The grid the curves are read against. */}
        {[0.25, 0.5, 0.75].map((g) => (
          <line
            key={g}
            x1={PAD_L}
            y1={yOf(g)}
            x2={PAD_L + PLOT_W * axesIn}
            y2={yOf(g)}
            stroke={withAlpha(theme.line, 0.12)}
            strokeWidth={1}
          />
        ))}

        {/* Axes, with arrowheads, drawn from the origin outward. */}
        <line
          x1={PAD_L}
          y1={PLOT_BOTTOM}
          x2={PAD_L + (PLOT_W + 46) * axesIn}
          y2={PLOT_BOTTOM}
          stroke={withAlpha(theme.line, 0.6)}
          strokeWidth={1.8}
        />
        <line
          x1={PAD_L}
          y1={PLOT_BOTTOM}
          x2={PAD_L}
          y2={PLOT_BOTTOM - (PLOT_H + 22) * axesIn}
          stroke={withAlpha(theme.line, 0.6)}
          strokeWidth={1.8}
        />
        <path
          d={`M ${PAD_L + PLOT_W + 36} ${PLOT_BOTTOM - 6} L ${PAD_L + PLOT_W + 46} ${PLOT_BOTTOM} L ${PAD_L + PLOT_W + 36} ${PLOT_BOTTOM + 6}`}
          fill="none"
          stroke={withAlpha(theme.line, 0.6)}
          strokeWidth={1.8}
          strokeLinecap="round"
          strokeLinejoin="round"
          opacity={axesIn}
        />
        <path
          d={`M ${PAD_L - 6} ${PLOT_BOTTOM - PLOT_H - 12} L ${PAD_L} ${PLOT_BOTTOM - PLOT_H - 22} L ${PAD_L + 6} ${PLOT_BOTTOM - PLOT_H - 12}`}
          fill="none"
          stroke={withAlpha(theme.line, 0.6)}
          strokeWidth={1.8}
          strokeLinecap="round"
          strokeLinejoin="round"
          opacity={axesIn}
        />
        <text
          x={PAD_L + PLOT_W + 52}
          y={PLOT_BOTTOM + 30}
          textAnchor="end"
          fill={theme.textMuted}
          opacity={axesIn}
          style={{fontFamily: theme.fontMono, fontSize: 18, letterSpacing: 2.4}}
        >
          n, growing
        </text>
        <text
          x={-(PLOT_BOTTOM - PLOT_H / 2)}
          y={PAD_L - 34}
          transform="rotate(-90)"
          textAnchor="middle"
          fill={theme.textMuted}
          opacity={axesIn}
          style={{fontFamily: theme.fontMono, fontSize: 18, letterSpacing: 2.4}}
        >
          operations
        </text>

        {curves.map((c, i) => {
          const points = Array.isArray(c.points) ? c.points : [];
          if (points.length < 2) return null;
          const on = drawn.includes(i);
          const isDrawing = i === drawing;
          if (!on && !isDrawing) return null;
          const len = lengthOf(points);
          const sweep = isDrawing
            ? interpolate(sinceStep, [3, 33], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
            : 1;
          const colour = colourOf(i);
          const li = labelIndex(points);
          const lx = xOf(li, points.length);
          const ly = yOf(points[li]);
          const labelOn = isDrawing
            ? interpolate(sinceStep, [24, 36], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
            : 1;
          const exits = points[li] >= 0.999;
          return (
            <g key={i} opacity={dimmed(i) ? 0.3 : 1}>
              <path
                d={pathOf(points)}
                fill="none"
                stroke={colour}
                strokeWidth={i === worst && isMoral ? 4 : 3}
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeDasharray={len}
                strokeDashoffset={len * (1 - sweep)}
              />
              <g opacity={labelOn}>
                <text
                  x={exits ? lx + 14 : lx + 16}
                  y={exits ? ly + 4 : ly + 6}
                  fill={colour}
                  style={{fontFamily: theme.fontMono, fontSize: 25, fontWeight: 700, letterSpacing: -0.4}}
                >
                  {c.notation}
                </text>
                <text
                  x={exits ? lx + 14 : lx + 16}
                  y={exits ? ly + 30 : ly + 32}
                  fill={theme.textMuted}
                  style={{fontFamily: theme.fontBody, fontSize: 19}}
                >
                  {c.label}
                </text>
              </g>
            </g>
          );
        })}

        {/* The probe: one hairline dropped through everything on the chart. */}
        {probeUp && probeLabel ? (
          <g opacity={probeIn}>
            <line
              x1={probeX}
              y1={PLOT_BOTTOM}
              x2={probeX}
              y2={PLOT_BOTTOM - PLOT_H * probeIn}
              stroke={withAlpha(theme.accentLimit, 0.8)}
              strokeWidth={2}
              strokeDasharray="6 6"
            />
            <circle cx={probeX} cy={PLOT_BOTTOM} r={5} fill={theme.accentLimit} />
            <text
              x={probeX}
              y={PLOT_BOTTOM + 30}
              textAnchor="middle"
              fill={theme.accentLimit}
              style={{fontFamily: theme.fontMono, fontSize: 19, fontWeight: 700, letterSpacing: 1}}
            >
              {`n = ${probeLabel}`}
            </text>
          </g>
        ) : null}
      </svg>

      {/* The readings, right-aligned under the chart. */}
      <div
        style={{
          width: CHART_W,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'flex-end',
          gap: 10,
          marginTop: 24,
          minHeight: 110,
          opacity: probeIn,
          transform: `translateY(${(1 - probeIn) * 12}px)`,
        }}
      >
        {curves.map((c, i) =>
          c.reading ? (
            <div
              key={i}
              style={{
                display: 'flex',
                alignItems: 'baseline',
                gap: 18,
                padding: '7px 18px',
                borderRadius: 10,
                background: withAlpha(theme.surface, i === worst ? 0.95 : 0.6),
                border: `1px solid ${i === worst ? withAlpha(theme.accentLimit, 0.6) : theme.surfaceBorder}`,
                opacity: dimmed(i) ? 0.45 : 1,
              }}
            >
              <span style={{fontFamily: theme.fontMono, fontSize: 21, fontWeight: 700, color: colourOf(i)}}>{c.notation}</span>
              <span style={{fontFamily: theme.fontBody, fontSize: 19, color: theme.textMuted}}>{c.label}</span>
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 30,
                  fontWeight: 700,
                  letterSpacing: -0.8,
                  color: i === worst ? theme.accentLimit : theme.text,
                  minWidth: 260,
                  textAlign: 'right',
                }}
              >
                {c.reading}
              </span>
            </div>
          ) : null,
        )}
      </div>
    </Stage>
  );
};
