import {useMemo} from 'react';
import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {arc, pie} from 'd3-shape';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_H, STAGE_W} from './Stage';
import {COUNTRY_BY_NAME, MAP_BOX, MAP_OFFSET_Y} from './geo';

// DataScene is one chart, held for the whole clip, with the emphasis moving
// around it.
//
// That is the entire design and it is worth being explicit about: a chart per
// beat would mean a viewer re-reading a new set of axes every eight seconds and
// never getting past reading them. One chart that stays put lets the second
// mention of a bar mean something, because it is the same bar.
//
// So the only thing that changes frame to frame is which points are lit — the
// same mechanism the flow diagram uses for its focus, and it carries the same
// meaning. Dimming is deliberately gentle for the same reason it is there:
// the unlit data is the context that makes the lit data a fact rather than a
// number.

const HEADER_H = 112;
const BOARD_W = STAGE_W;
const BOARD_H = STAGE_H - HEADER_H;

const ENTER = {
  /** Frames for one datum to arrive. */
  frames: 16,
  /** Frames between one datum and the next. */
  stagger: 4,
  /** How sharply the highlight transition runs. */
  focusFrames: 9,
} as const;

/** Dim floor for unlit data — readable, but clearly behind. */
const DIM = 0.42;

type Point = {label: string; value: number};
type Window = {startMs: number; endMs: number; labels: string[]};
type Caption = {startMs: number; endMs: number; text: string};

/** Value formatted for display: no trailing .0, and the unit attached. */
const fmt = (v: number, unit: string): string => {
  const n = Number.isInteger(v) ? String(v) : v.toFixed(1);
  return unit ? `${n}${unit.startsWith('%') ? '' : ' '}${unit}` : n;
};

export const DataScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const nowMs = sceneStartMs + (frame / FPS) * 1000;

  const title = String(props.title ?? '');
  const kind = String(props.kind ?? 'bars');
  const unit = String(props.unit ?? '');
  const points = (Array.isArray(props.points) ? props.points : []) as Point[];
  const windows = (Array.isArray(props.highlight) ? props.highlight : []) as Window[];
  const captions = (Array.isArray(props.captions) ? props.captions : []) as Caption[];

  const active = windows.find((w) => nowMs >= w.startMs && nowMs < w.endMs);
  const lit = useMemo(
    () => (active ? new Set(active.labels.map((l) => l.toLowerCase())) : null),
    [active],
  );
  const focusP = active
    ? interpolate(nowMs - active.startMs, [0, (ENTER.focusFrames / FPS) * 1000], [0, 1], {
        extrapolateLeft: 'clamp',
        extrapolateRight: 'clamp',
      })
    : 0;

  const caption = captions.find((c) => nowMs >= c.startMs && nowMs < c.endMs);

  if (points.length === 0) {
    return null;
  }

  const isLit = (label: string): boolean => !lit || lit.has(label.toLowerCase());
  /** Opacity for a datum: 1 when lit or when nothing is, eased down when not. */
  const dim = (label: string): number =>
    isLit(label) ? 1 : 1 - (1 - DIM) * focusP;
  /** The colour a datum takes: accent while lit, the primary otherwise. */
  const tint = (label: string): string =>
    lit && isLit(label) ? theme.accent : theme.primary;

  /** Arrival progress for datum i. */
  const grow = (i: number): number =>
    spring({
      frame: frame - i * ENTER.stagger,
      fps,
      config: {damping: 200, mass: 0.7},
      durationInFrames: ENTER.frames,
    });

  const maxValue = Math.max(...points.map((p) => p.value), 1);
  const chartH = BOARD_H - (caption ? 74 : 24);

  const bars = () => {
    // Horizontal, because it is the only orientation where a real label fits
    // without being rotated — and a rotated axis label is unreadable at a
    // glance, which is the only kind of reading a video gets.
    const rowH = Math.min(chartH / points.length, 92);
    const gap = Math.min(rowH * 0.28, 20);
    const barH = rowH - gap;
    const labelW = Math.min(BOARD_W * 0.28, 420);
    const valueW = 150;
    const trackW = BOARD_W - labelW - valueW - 40;
    const top = (chartH - rowH * points.length) / 2;

    return (
      <svg width={BOARD_W} height={chartH} style={{overflow: 'visible'}}>
        {points.map((p, i) => {
          const g = grow(i);
          const y = top + i * rowH;
          const w = trackW * (p.value / maxValue) * g;
          return (
            <g key={p.label} opacity={dim(p.label)}>
              <text
                x={labelW}
                y={y + barH / 2}
                textAnchor="end"
                dominantBaseline="central"
                fontFamily={theme.fontBody}
                fontSize={Math.min(30, barH * 0.52)}
                fontWeight={600}
                fill={theme.text}
              >
                {p.label}
              </text>
              {/* The track shows what the bar is a fraction *of*. Without it a
                  short bar reads as missing data rather than as a small value. */}
              <rect
                x={labelW + 22}
                y={y}
                width={trackW}
                height={barH}
                rx={barH / 2}
                fill={theme.surface}
                opacity={0.55}
              />
              <rect
                x={labelW + 22}
                y={y}
                width={Math.max(0, w)}
                height={barH}
                rx={barH / 2}
                fill={tint(p.label)}
              />
              <text
                x={labelW + 22 + Math.max(0, w) + 16}
                y={y + barH / 2}
                dominantBaseline="central"
                fontFamily={theme.fontDisplay}
                fontSize={Math.min(30, barH * 0.54)}
                fontWeight={700}
                fill={lit && isLit(p.label) ? theme.accentText : theme.textMuted}
                opacity={g}
              >
                {fmt(p.value, unit)}
              </text>
            </g>
          );
        })}
      </svg>
    );
  };

  const line = () => {
    const padL = 96;
    const padR = 60;
    const padB = 62;
    const w = BOARD_W - padL - padR;
    const h = chartH - padB - 28;
    const x = (i: number) => padL + (w * i) / Math.max(1, points.length - 1);
    const y = (v: number) => 28 + h - (h * v) / maxValue;

    const drawn = interpolate(frame, [0, ENTER.frames + ENTER.stagger * points.length], [0, 1], {
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp',
    });
    const d = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${x(i)} ${y(p.value)}`).join(' ');

    return (
      <svg width={BOARD_W} height={chartH} style={{overflow: 'visible'}}>
        <line x1={padL} y1={28 + h} x2={padL + w} y2={28 + h} stroke={theme.line} strokeWidth={3} opacity={0.45} />
        <path
          d={d}
          fill="none"
          stroke={theme.primary}
          strokeWidth={5}
          strokeLinecap="round"
          strokeLinejoin="round"
          pathLength={1}
          strokeDasharray={1}
          strokeDashoffset={1 - drawn}
        />
        {points.map((p, i) => {
          const g = grow(i);
          const on = lit && isLit(p.label);
          return (
            <g key={p.label}>
              {/* The axis label is not data and must not dim with it — an axis
                  you cannot read is worse than no axis, and dimming one is how
                  a chart loses the very thing that says what it is of. */}
              <text
                x={x(i)}
                y={28 + h + 34}
                textAnchor="middle"
                fontFamily={theme.fontBody}
                fontSize={24}
                fontWeight={600}
                fill={on ? theme.text : theme.textMuted}
                opacity={g}
              >
                {p.label}
              </text>
              <g opacity={dim(p.label) * g}>
              <circle cx={x(i)} cy={y(p.value)} r={on ? 11 : 7} fill={tint(p.label)} />
              {on && (
                <text
                  x={x(i)}
                  y={y(p.value) - 26}
                  textAnchor="middle"
                  fontFamily={theme.fontDisplay}
                  fontSize={30}
                  fontWeight={700}
                  fill={theme.accentText}
                >
                  {fmt(p.value, unit)}
                </text>
              )}
              </g>
            </g>
          );
        })}
      </svg>
    );
  };

  const donut = () => {
    const r = Math.min(chartH * 0.46, 260);
    const cx = BOARD_W / 2;
    const cy = chartH / 2;
    const layout = pie<Point>().value((p) => p.value).sort(null);
    const slices = layout(points);
    // innerRadius is fixed; outerRadius is left as an accessor so a lit slice
    // can step outward per-datum without rebuilding the generator.
    type Slice = {startAngle: number; endAngle: number; outerRadius: number};
    const shape = arc<Slice>().innerRadius(r * 0.58).cornerRadius(3);

    return (
      <svg width={BOARD_W} height={chartH} style={{overflow: 'visible'}}>
        {slices.map((s, i) => {
          const g = grow(i);
          const on = lit && isLit(s.data.label);
          // A lit slice steps outward as well as changing colour: on a ring,
          // colour alone is easy to lose next to its neighbours.
          const d =
            shape({
              startAngle: s.startAngle,
              endAngle: s.startAngle + (s.endAngle - s.startAngle) * g,
              outerRadius: r * (on ? 1.06 : 1),
            }) ?? '';
          const mid = (s.startAngle + s.endAngle) / 2 - Math.PI / 2;
          const lx = cx + Math.cos(mid) * (r * 1.24);
          const ly = cy + Math.sin(mid) * (r * 1.24);
          return (
            <g key={s.data.label} opacity={dim(s.data.label)}>
              <path d={d} transform={`translate(${cx} ${cy})`} fill={tint(s.data.label)} />
              <text
                x={lx}
                y={ly}
                textAnchor={Math.cos(mid) > 0 ? 'start' : 'end'}
                dominantBaseline="central"
                fontFamily={theme.fontBody}
                fontSize={25}
                fontWeight={600}
                fill={on ? theme.accentText : theme.textMuted}
                opacity={g}
              >
                {s.data.label} {fmt(s.data.value, unit)}
              </text>
            </g>
          );
        })}
      </svg>
    );
  };

  const map = () => {
    const scale = Math.min(BOARD_W / MAP_BOX.w, chartH / MAP_BOX.h);
    const byLabel = new Map(points.map((p) => [p.label.toLowerCase(), p]));
    const reveal = interpolate(frame, [0, ENTER.frames], [0, 1], {
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp',
    });

    return (
      <svg
        width={MAP_BOX.w * scale}
        height={MAP_BOX.h * scale}
        viewBox={`0 0 ${MAP_BOX.w} ${MAP_BOX.h}`}
        style={{overflow: 'visible'}}
      >
        <g transform={`translate(0 ${MAP_OFFSET_Y})`}>
        {/* Every country is drawn, and the ones with data are filled. The rest
            are the world the data sits in — a map of only the highlighted
            countries is a set of floating shapes nobody can place. */}
        {Object.values(COUNTRY_BY_NAME).map((c) => {
          const p = byLabel.get(c.name.toLowerCase());
          if (!p) {
            return (
              <path
                key={c.name}
                d={c.d}
                fill={theme.surface}
                stroke={theme.bgBottom}
                strokeWidth={0.4}
                opacity={0.75}
              />
            );
          }
          // Shaded by value within the data's own range, so a set of similar
          // numbers still reads as a gradient rather than as one flat colour.
          // The floor matters more than the range: a country carrying a small
          // value is still a country with data, and at a low enough opacity it
          // is indistinguishable from one with none.
          const t = p.value / maxValue;
          return (
            <path
              key={c.name}
              d={c.d}
              fill={tint(p.label)}
              fillOpacity={(0.5 + 0.5 * t) * reveal}
              stroke={lit && isLit(p.label) ? theme.accent : theme.bgBottom}
              strokeWidth={lit && isLit(p.label) ? 1.4 : 0.4}
              opacity={dim(p.label)}
            />
          );
        })}
        </g>
      </svg>
    );
  };

  const chart =
    kind === 'line' ? line() : kind === 'donut' ? donut() : kind === 'map' ? map() : bars();

  return (
    <Stage justify="flex-start">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={22} />
      <div style={{display: 'flex', justifyContent: 'center', width: BOARD_W}}>{chart}</div>
      {caption && (
        <div
          style={{
            marginTop: 16,
            fontFamily: theme.fontBody,
            fontSize: 30,
            fontWeight: 500,
            color: theme.textMuted,
            textAlign: 'center',
            maxWidth: BOARD_W * 0.8,
          }}
        >
          {caption.text}
        </div>
      )}
    </Stage>
  );
};
