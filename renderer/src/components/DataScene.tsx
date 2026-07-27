import {useMemo} from 'react';
import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {arc, pie} from 'd3-shape';
import {hierarchy, treemap as d3Treemap} from 'd3-hierarchy';
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
//
// Every kind below is written against the same context object, which is what
// makes "lit" mean the same thing on a treemap tile, a scatter dot and a
// country. A kind that had to invent its own idea of emphasis would be a
// second design sharing a file with the first.

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

type Point = {label: string; value: number; values?: number[]};
type Window = {startMs: number; endMs: number; labels: string[]};
type Caption = {startMs: number; endMs: number; text: string};

/** Value formatted for display: no trailing .0, and the unit attached. */
const fmt = (v: number, unit: string): string => {
  const n = Number.isInteger(v) ? String(v) : v.toFixed(1);
  return unit ? `${n}${unit.startsWith('%') ? '' : ' '}${unit}` : n;
};

/**
 * Everything a chart kind needs, resolved once.
 *
 * Passing a context rather than a dozen props is what keeps thirteen kinds
 * honest with each other: `dim`, `tint` and `grow` are computed in one place,
 * so a highlight looks the same however the data is drawn, and a new kind
 * cannot accidentally invent a second visual language for emphasis.
 */
type Ctx = {
  theme: ResolvedTheme;
  points: Point[];
  series: string[];
  unit: string;
  /** Drawing height available below the header, minus any caption. */
  h: number;
  w: number;
  frame: number;
  /** Largest single value in the data, floored at 1. */
  maxValue: number;
  /** Largest point total, for the kinds that stack. */
  maxTotal: number;
  /** True when nothing is lit, or when this label is. */
  isLit: (label: string) => boolean;
  /** Whether any highlight is active at all. */
  focused: boolean;
  /** Opacity for a datum. */
  dim: (label: string) => number;
  /** The colour a datum takes: accent while lit, the primary otherwise. */
  tint: (label: string) => string;
  /** Arrival progress for datum i. */
  grow: (i: number) => number;
  /** Text colour for a datum's value label. */
  valueColor: (label: string) => string;
  /** Colour for series i of n. */
  seriesFill: (i: number) => string;
  seriesOpacity: (i: number) => number;
};

/**
 * Series colours.
 *
 * Two hues alternating and stepping down in opacity, rather than a ramp of one.
 * A four-step ramp of a single hue is four greys to anybody watching on a phone
 * in daylight, and these charts are read once, at speed, from across a room.
 */
const seriesFillFor = (theme: ResolvedTheme, i: number): string =>
  i % 2 === 0 ? theme.primary : theme.accent;
const seriesOpacityFor = (i: number): number => 1 - Math.floor(i / 2) * 0.34;

/** A legend, for the kinds where a colour means a name. */
const Legend: React.FC<{ctx: Ctx}> = ({ctx}) => (
  <div
    style={{
      display: 'flex',
      justifyContent: 'center',
      flexWrap: 'wrap',
      gap: 28,
      marginTop: 10,
    }}
  >
    {ctx.series.map((name, i) => (
      <div key={name} style={{display: 'flex', alignItems: 'center', gap: 10}}>
        <div
          style={{
            width: 22,
            height: 22,
            borderRadius: 6,
            background: ctx.seriesFill(i),
            opacity: ctx.seriesOpacity(i),
          }}
        />
        <span
          style={{
            fontFamily: ctx.theme.fontBody,
            fontSize: 26,
            fontWeight: 600,
            color: ctx.theme.textMuted,
          }}
        >
          {name}
        </span>
      </div>
    ))}
  </div>
);

// ---------------------------------------------------------------------------
// The kinds
// ---------------------------------------------------------------------------

const bars = (c: Ctx) => {
  // Horizontal, because it is the only orientation where a real label fits
  // without being rotated — and a rotated axis label is unreadable at a
  // glance, which is the only kind of reading a video gets.
  const rowH = Math.min(c.h / c.points.length, 92);
  const gap = Math.min(rowH * 0.28, 20);
  const barH = rowH - gap;
  const labelW = Math.min(c.w * 0.28, 420);
  const valueW = 150;
  const trackW = c.w - labelW - valueW - 40;
  const top = (c.h - rowH * c.points.length) / 2;

  return (
    <svg width={c.w} height={c.h} style={{overflow: 'visible'}}>
      {c.points.map((p, i) => {
        const g = c.grow(i);
        const y = top + i * rowH;
        const w = trackW * (p.value / c.maxValue) * g;
        return (
          <g key={p.label} opacity={c.dim(p.label)}>
            <text
              x={labelW}
              y={y + barH / 2}
              textAnchor="end"
              dominantBaseline="central"
              fontFamily={c.theme.fontBody}
              fontSize={Math.min(30, barH * 0.52)}
              fontWeight={600}
              fill={c.theme.text}
            >
              {p.label}
            </text>
            {/* The track shows what the bar is a fraction *of*. Without it a
                short bar reads as missing data rather than as a small value. */}
            <rect x={labelW + 22} y={y} width={trackW} height={barH} rx={barH / 2} fill={c.theme.surface} opacity={0.55} />
            <rect x={labelW + 22} y={y} width={Math.max(0, w)} height={barH} rx={barH / 2} fill={c.tint(p.label)} />
            <text
              x={labelW + 22 + Math.max(0, w) + 16}
              y={y + barH / 2}
              dominantBaseline="central"
              fontFamily={c.theme.fontDisplay}
              fontSize={Math.min(30, barH * 0.54)}
              fontWeight={700}
              fill={c.valueColor(p.label)}
              opacity={g}
            >
              {fmt(p.value, c.unit)}
            </text>
          </g>
        );
      })}
    </svg>
  );
};

/**
 * Vertical bars split into named parts, or set side by side.
 *
 * Vertical rather than horizontal, unlike plain bars: a stack has to be read
 * against a common baseline for the segments to be comparable at all, and the
 * baseline the eye finds without being told is the floor.
 */
const barsWithSeries = (c: Ctx, grouped: boolean) => {
  const padB = 66;
  const padT = 30;
  const plotH = c.h - padB - padT;
  const slot = c.w / c.points.length;
  const bandW = Math.min(slot * 0.62, 190);
  const n = Math.max(1, c.series.length);
  const barW = grouped ? bandW / n : bandW;
  const scale = plotH / (grouped ? c.maxValue : c.maxTotal);

  return (
    <svg width={c.w} height={c.h} style={{overflow: 'visible'}}>
      <line x1={0} y1={padT + plotH} x2={c.w} y2={padT + plotH} stroke={c.theme.line} strokeWidth={3} opacity={0.45} />
      {c.points.map((p, i) => {
        const g = c.grow(i);
        const cx = slot * (i + 0.5);
        const values = p.values ?? [p.value];
        let stackY = padT + plotH;
        return (
          <g key={p.label}>
            <g opacity={c.dim(p.label)}>
              {values.map((v, s) => {
                const hgt = v * scale * g;
                if (grouped) {
                  const x = cx - bandW / 2 + s * barW;
                  return (
                    <rect
                      key={s}
                      x={x + 3}
                      y={padT + plotH - hgt}
                      width={Math.max(1, barW - 6)}
                      height={Math.max(0, hgt)}
                      rx={6}
                      fill={c.seriesFill(s)}
                      opacity={c.seriesOpacity(s)}
                    />
                  );
                }
                stackY -= hgt;
                return (
                  <rect
                    key={s}
                    x={cx - barW / 2}
                    y={stackY}
                    width={barW}
                    height={Math.max(0, hgt)}
                    // Only the ends are rounded, or every segment reads as its
                    // own free-floating bar and the stack stops being one.
                    rx={s === 0 || s === values.length - 1 ? 6 : 0}
                    fill={c.seriesFill(s)}
                    opacity={c.seriesOpacity(s)}
                  />
                );
              })}
            </g>
            {/* The axis label is not data and must not dim with it. */}
            <text
              x={cx}
              y={padT + plotH + 38}
              textAnchor="middle"
              fontFamily={c.theme.fontBody}
              fontSize={26}
              fontWeight={600}
              fill={c.focused && c.isLit(p.label) ? c.theme.text : c.theme.textMuted}
              opacity={g}
            >
              {p.label}
            </text>
            {!grouped && c.focused && c.isLit(p.label) && (
              <text
                x={cx}
                y={padT + plotH - p.value * scale * g - 16}
                textAnchor="middle"
                fontFamily={c.theme.fontDisplay}
                fontSize={28}
                fontWeight={700}
                fill={c.theme.accentText}
              >
                {fmt(p.value, c.unit)}
              </text>
            )}
          </g>
        );
      })}
    </svg>
  );
};

const line = (c: Ctx, filled: boolean) => {
  const padL = 96;
  const padR = 60;
  const padB = 62;
  const w = c.w - padL - padR;
  const h = c.h - padB - 28;
  const x = (i: number) => padL + (w * i) / Math.max(1, c.points.length - 1);
  const y = (v: number) => 28 + h - (h * v) / c.maxValue;

  const drawn = interpolate(c.frame, [0, ENTER.frames + ENTER.stagger * c.points.length], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const d = c.points.map((p, i) => `${i === 0 ? 'M' : 'L'}${x(i)} ${y(p.value)}`).join(' ');

  return (
    <svg width={c.w} height={c.h} style={{overflow: 'visible'}}>
      <line x1={padL} y1={28 + h} x2={padL + w} y2={28 + h} stroke={c.theme.line} strokeWidth={3} opacity={0.45} />
      {filled && (
        // The fill sweeps in from the left with the stroke rather than fading
        // up underneath it: a wash that appears whole under a line still being
        // drawn reads as two unrelated animations.
        <g clipPath="url(#areaWipe)">
          <defs>
            <clipPath id="areaWipe">
              <rect x={padL} y={0} width={w * drawn} height={c.h} />
            </clipPath>
          </defs>
          <path d={`${d} L${x(c.points.length - 1)} ${28 + h} L${padL} ${28 + h} Z`} fill={c.theme.primary} opacity={0.22} />
        </g>
      )}
      <path
        d={d}
        fill="none"
        stroke={c.theme.primary}
        strokeWidth={5}
        strokeLinecap="round"
        strokeLinejoin="round"
        pathLength={1}
        strokeDasharray={1}
        strokeDashoffset={1 - drawn}
      />
      {c.points.map((p, i) => {
        const g = c.grow(i);
        const on = c.focused && c.isLit(p.label);
        return (
          <g key={p.label}>
            {/* The axis label is not data and must not dim with it — an axis
                you cannot read is worse than no axis, and dimming one is how
                a chart loses the very thing that says what it is of. */}
            <text
              x={x(i)}
              y={28 + h + 34}
              textAnchor="middle"
              fontFamily={c.theme.fontBody}
              fontSize={24}
              fontWeight={600}
              fill={on ? c.theme.text : c.theme.textMuted}
              opacity={g}
            >
              {p.label}
            </text>
            <g opacity={c.dim(p.label) * g}>
              <circle cx={x(i)} cy={y(p.value)} r={on ? 11 : 7} fill={c.tint(p.label)} />
              {on && (
                <text
                  x={x(i)}
                  y={y(p.value) - 26}
                  textAnchor="middle"
                  fontFamily={c.theme.fontDisplay}
                  fontSize={30}
                  fontWeight={700}
                  fill={c.theme.accentText}
                >
                  {fmt(p.value, c.unit)}
                </text>
              )}
            </g>
          </g>
        );
      })}
    </svg>
  );
};

/** Two variables against each other, with every dot labelled. */
const scatter = (c: Ctx) => {
  const padL = 110;
  // Wide, because the label sits above its dot and the dot at the top right of
  // the data is exactly the one worth naming.
  const padR = 170;
  const padB = 76;
  const padT = 72;
  const w = c.w - padL - padR;
  const h = c.h - padB - padT;
  const xs = c.points.map((p) => p.values?.[0] ?? 0);
  const ys = c.points.map((p) => p.values?.[1] ?? 0);
  // Scaled from zero rather than from the data's own minimum. A scatter with a
  // clipped origin exaggerates whatever relationship it has, and these are read
  // in two seconds by somebody who will not check the axis.
  const maxX = Math.max(...xs, 1);
  const maxY = Math.max(...ys, 1);
  const px = (v: number) => padL + (w * v) / maxX;
  const py = (v: number) => padT + h - (h * v) / maxY;

  return (
    <svg width={c.w} height={c.h} style={{overflow: 'visible'}}>
      <line x1={padL} y1={padT} x2={padL} y2={padT + h} stroke={c.theme.line} strokeWidth={3} opacity={0.45} />
      <line x1={padL} y1={padT + h} x2={padL + w} y2={padT + h} stroke={c.theme.line} strokeWidth={3} opacity={0.45} />
      <text x={padL + w / 2} y={padT + h + 54} textAnchor="middle" fontFamily={c.theme.fontBody} fontSize={26} fontWeight={600} fill={c.theme.textMuted}>
        {c.series[0]}
      </text>
      <text
        x={-(padT + h / 2)}
        y={padL - 60}
        transform="rotate(-90)"
        textAnchor="middle"
        fontFamily={c.theme.fontBody}
        fontSize={26}
        fontWeight={600}
        fill={c.theme.textMuted}
      >
        {c.series[1]}
      </text>
      {c.points.map((p, i) => {
        const g = c.grow(i);
        const on = c.focused && c.isLit(p.label);
        const cx = px(p.values?.[0] ?? 0);
        const cy = py(p.values?.[1] ?? 0);
        return (
          <g key={p.label} opacity={c.dim(p.label) * g}>
            <circle cx={cx} cy={cy} r={on ? 20 : 14} fill={c.tint(p.label)} />
            <text
              x={cx}
              y={cy - (on ? 32 : 26)}
              textAnchor="middle"
              fontFamily={c.theme.fontBody}
              fontSize={on ? 26 : 22}
              fontWeight={600}
              fill={on ? c.theme.accentText : c.theme.textMuted}
            >
              {p.label}
            </text>
          </g>
        );
      })}
    </svg>
  );
};

const donut = (c: Ctx) => {
  const r = Math.min(c.h * 0.46, 260);
  const cx = c.w / 2;
  const cy = c.h / 2;
  const layout = pie<Point>().value((p) => p.value).sort(null);
  const slices = layout(c.points);
  // innerRadius is fixed; outerRadius is left as an accessor so a lit slice
  // can step outward per-datum without rebuilding the generator.
  type Slice = {startAngle: number; endAngle: number; outerRadius: number};
  const shape = arc<Slice>().innerRadius(r * 0.58).cornerRadius(3);

  return (
    <svg width={c.w} height={c.h} style={{overflow: 'visible'}}>
      {slices.map((s, i) => {
        const g = c.grow(i);
        const on = c.focused && c.isLit(s.data.label);
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
          <g key={s.data.label} opacity={c.dim(s.data.label)}>
            <path d={d} transform={`translate(${cx} ${cy})`} fill={c.tint(s.data.label)} />
            <text
              x={lx}
              y={ly}
              textAnchor={Math.cos(mid) > 0 ? 'start' : 'end'}
              dominantBaseline="central"
              fontFamily={c.theme.fontBody}
              fontSize={25}
              fontWeight={600}
              fill={on ? c.theme.accentText : c.theme.textMuted}
              opacity={g}
            >
              {s.data.label} {fmt(s.data.value, c.unit)}
            </text>
          </g>
        );
      })}
    </svg>
  );
};

/**
 * A hundred squares, handed out in proportion.
 *
 * The reason to own this as well as the donut: people read angles and areas
 * badly and count squares well, so "eleven in a hundred" lands off a waffle in
 * a way it never does off an eleven-degree slice.
 */
const waffle = (c: Ctx) => {
  const COLS = 10;
  const ROWS = 10;
  const cell = Math.min((c.h - 30) / ROWS, 62);
  const gap = cell * 0.16;
  const gridW = COLS * cell;
  const left = (c.w - gridW) / 2 - 190;
  const total = c.points.reduce((s, p) => s + p.value, 0) || 1;

  // Largest remainder, so the squares always total exactly a hundred. Plain
  // rounding leaves 99 or 101 cells and somebody always counts.
  const exact = c.points.map((p) => (p.value / total) * 100);
  const counts = exact.map(Math.floor);
  let left100 = 100 - counts.reduce((s, v) => s + v, 0);
  exact
    .map((v, i) => ({i, frac: v - Math.floor(v)}))
    .sort((a, b) => b.frac - a.frac)
    .forEach(({i}) => {
      if (left100 > 0) {
        counts[i]++;
        left100--;
      }
    });

  const owner: number[] = [];
  counts.forEach((n, i) => {
    for (let k = 0; k < n; k++) owner.push(i);
  });

  return (
    <svg width={c.w} height={c.h} style={{overflow: 'visible'}}>
      {owner.map((idx, k) => {
        const p = c.points[idx];
        // Filled column by column from the bottom left, which is the order a
        // reader's eye already expects a tally to grow in.
        const col = Math.floor(k / ROWS);
        const row = ROWS - 1 - (k % ROWS);
        const g = c.grow(Math.floor(k / 12));
        return (
          <rect
            key={k}
            x={left + col * cell}
            y={row * cell + 14}
            width={cell - gap}
            height={cell - gap}
            rx={5}
            fill={c.tint(p.label)}
            opacity={c.dim(p.label) * g}
          />
        );
      })}
      {c.points.map((p, i) => {
        const on = c.focused && c.isLit(p.label);
        return (
          <g key={p.label} opacity={c.dim(p.label)}>
            <rect x={left + gridW + 46} y={26 + i * 46} width={26} height={26} rx={6} fill={c.tint(p.label)} />
            <text
              x={left + gridW + 86}
              y={39 + i * 46}
              dominantBaseline="central"
              fontFamily={c.theme.fontBody}
              fontSize={26}
              fontWeight={600}
              fill={on ? c.theme.accentText : c.theme.textMuted}
            >
              {p.label} — {counts[i]} in 100
            </text>
          </g>
        );
      })}
    </svg>
  );
};

/** A row of dials, each filled against the largest value in the set. */
const gauge = (c: Ctx) => {
  const slot = c.w / c.points.length;
  const r = Math.min(slot * 0.34, c.h * 0.32, 150);
  const cy = c.h * 0.52;
  const sweep = 240;

  const dialPath = (cx: number, from: number, to: number) => {
    // Angles run clockwise from the lower left, so the 120° the sweep does not
    // cover sits at the bottom where a dial's gap belongs. Starting anywhere
    // else puts the gap on one side and the dial reads as a letter C.
    const a = (deg: number) => ((180 + (180 - sweep) / 2 + deg) * Math.PI) / 180;
    const x0 = cx + r * Math.cos(a(from));
    const y0 = cy + r * Math.sin(a(from));
    const x1 = cx + r * Math.cos(a(to));
    const y1 = cy + r * Math.sin(a(to));
    return `M${x0} ${y0} A${r} ${r} 0 ${to - from > 180 ? 1 : 0} 1 ${x1} ${y1}`;
  };

  return (
    <svg width={c.w} height={c.h} style={{overflow: 'visible'}}>
      {c.points.map((p, i) => {
        const g = c.grow(i);
        const cx = slot * (i + 0.5);
        const frac = (p.value / c.maxValue) * g;
        return (
          <g key={p.label} opacity={c.dim(p.label)}>
            <path d={dialPath(cx, 0, sweep)} fill="none" stroke={c.theme.surface} strokeWidth={r * 0.3} strokeLinecap="round" />
            {frac > 0.001 && (
              <path
                d={dialPath(cx, 0, sweep * frac)}
                fill="none"
                stroke={c.tint(p.label)}
                strokeWidth={r * 0.3}
                strokeLinecap="round"
              />
            )}
            <text
              x={cx}
              y={cy}
              textAnchor="middle"
              dominantBaseline="central"
              fontFamily={c.theme.fontDisplay}
              fontSize={Math.min(r * 0.5, 56)}
              fontWeight={700}
              fill={c.valueColor(p.label)}
            >
              {fmt(p.value, c.unit)}
            </text>
            <text
              x={cx}
              y={cy + r + 58}
              textAnchor="middle"
              fontFamily={c.theme.fontBody}
              fontSize={26}
              fontWeight={600}
              fill={c.theme.textMuted}
            >
              {p.label}
            </text>
          </g>
        );
      })}
    </svg>
  );
};

/** Rectangles sized by value, for a breakdown with a wide spread. */
const treemap = (c: Ctx) => {
  const w = Math.min(c.w * 0.8, 1200);
  const left = (c.w - w) / 2;
  const h = c.h - 20;
  const root = hierarchy<{children: Point[]} | Point>({children: c.points} as {children: Point[]})
    .sum((d) => ('value' in d ? d.value : 0))
    .sort((a, b) => (b.value ?? 0) - (a.value ?? 0));
  const laid = d3Treemap<{children: Point[]} | Point>().size([w, h]).paddingInner(8).round(true)(root);

  return (
    <svg width={c.w} height={c.h} style={{overflow: 'visible'}}>
      {laid.leaves().map((leaf, i) => {
        const p = leaf.data as Point;
        const g = c.grow(i);
        const tw = leaf.x1 - leaf.x0;
        const th = leaf.y1 - leaf.y0;
        const on = c.focused && c.isLit(p.label);
        // A tile too small for its own label gets none. A clipped word is
        // worse than a blank tile, because it reads as a different word.
        const fits = tw > 130 && th > 74;
        return (
          <g key={p.label} opacity={c.dim(p.label)} transform={`translate(${left + leaf.x0} ${leaf.y0 + 10})`}>
            <rect
              width={tw}
              height={th * g}
              y={th * (1 - g)}
              rx={10}
              fill={c.tint(p.label)}
              opacity={on ? 1 : 0.9}
            />
            {fits && (
              <>
                <text x={18} y={40} fontFamily={c.theme.fontBody} fontSize={Math.min(28, th * 0.22)} fontWeight={600} fill={c.theme.bgBottom} opacity={g}>
                  {p.label}
                </text>
                <text x={18} y={40 + Math.min(38, th * 0.3)} fontFamily={c.theme.fontDisplay} fontSize={Math.min(38, th * 0.28)} fontWeight={700} fill={c.theme.bgBottom} opacity={g}>
                  {fmt(p.value, c.unit)}
                </text>
              </>
            )}
          </g>
        );
      })}
    </svg>
  );
};

/** Stages narrowing to a result. */
const funnel = (c: Ctx) => {
  const n = c.points.length;
  const rowH = Math.min((c.h - 20) / n, 116);
  const top = (c.h - rowH * n) / 2;
  const maxW = Math.min(c.w * 0.46, 700);
  const cx = c.w * 0.5;
  const widthAt = (i: number) => maxW * (c.points[i].value / c.maxValue);

  return (
    <svg width={c.w} height={c.h} style={{overflow: 'visible'}}>
      {c.points.map((p, i) => {
        const g = c.grow(i);
        const y = top + i * rowH;
        const wTop = widthAt(i);
        // The band tapers towards the next stage, so the drop between two
        // stages is a visible slope rather than two stacked bars.
        const wBot = i === n - 1 ? wTop * 0.86 : widthAt(i + 1);
        const gap = 7;
        return (
          <g key={p.label} opacity={c.dim(p.label)}>
            <path
              d={`M${cx - (wTop / 2) * g} ${y} L${cx + (wTop / 2) * g} ${y} L${cx + (wBot / 2) * g} ${y + rowH - gap} L${cx - (wBot / 2) * g} ${y + rowH - gap} Z`}
              fill={c.tint(p.label)}
            />
            {/* Both texts sit outside the band, on opposite sides. Centring
                the value inside it works until the band is narrower than the
                number, which by the last stage of a funnel it always is. */}
            <text
              x={cx - maxW / 2 - 34}
              y={y + (rowH - gap) / 2}
              textAnchor="end"
              dominantBaseline="central"
              fontFamily={c.theme.fontDisplay}
              fontSize={30}
              fontWeight={700}
              fill={c.valueColor(p.label)}
              opacity={g}
            >
              {fmt(p.value, c.unit)}
            </text>
            <text
              x={cx + maxW / 2 + 34}
              y={y + (rowH - gap) / 2}
              dominantBaseline="central"
              fontFamily={c.theme.fontBody}
              fontSize={28}
              fontWeight={600}
              fill={c.theme.text}
              opacity={g}
            >
              {p.label}
            </text>
          </g>
        );
      })}
    </svg>
  );
};

/** Headline figures as cards, for a clip whose data is three numbers. */
const kpi = (c: Ctx) => {
  const gap = 34;
  const cardW = Math.min((c.w - gap * (c.points.length - 1)) / c.points.length, 400);
  const totalW = cardW * c.points.length + gap * (c.points.length - 1);
  const left = (c.w - totalW) / 2;
  const cardH = Math.min(c.h * 0.72, 300);
  const top = (c.h - cardH) / 2;

  return (
    <svg width={c.w} height={c.h} style={{overflow: 'visible'}}>
      {c.points.map((p, i) => {
        const g = c.grow(i);
        const x = left + i * (cardW + gap);
        const on = c.focused && c.isLit(p.label);
        return (
          <g key={p.label} opacity={c.dim(p.label)} transform={`translate(${x} ${top + (1 - g) * 22})`}>
            <rect width={cardW} height={cardH} rx={20} fill={c.theme.surface} opacity={0.85 * g} />
            {/* A lit card gets a rule rather than a fill: filling it would put
                the number on the accent, and a big number needs the highest
                contrast the theme has, not the loudest colour. */}
            <rect
              width={cardW}
              height={cardH}
              rx={20}
              fill="none"
              stroke={on ? c.theme.accent : c.theme.surfaceBorder}
              strokeWidth={on ? 4 : 2}
              opacity={g}
            />
            <text
              x={cardW / 2}
              y={cardH * 0.46}
              textAnchor="middle"
              dominantBaseline="central"
              fontFamily={c.theme.fontDisplay}
              fontSize={Math.min(cardW * 0.3, 92)}
              fontWeight={700}
              fill={on ? c.theme.accentText : c.theme.text}
              opacity={g}
            >
              {fmt(p.value, c.unit)}
            </text>
            <text
              x={cardW / 2}
              y={cardH * 0.76}
              textAnchor="middle"
              fontFamily={c.theme.fontBody}
              fontSize={27}
              fontWeight={600}
              fill={c.theme.textMuted}
              opacity={g}
            >
              {p.label}
            </text>
          </g>
        );
      })}
    </svg>
  );
};

const worldMap = (c: Ctx) => {
  const scale = Math.min(c.w / MAP_BOX.w, c.h / MAP_BOX.h);
  const byLabel = new Map(c.points.map((p) => [p.label.toLowerCase(), p]));
  const reveal = interpolate(c.frame, [0, ENTER.frames], [0, 1], {
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
        {Object.values(COUNTRY_BY_NAME).map((country) => {
          const p = byLabel.get(country.name.toLowerCase());
          if (!p) {
            return (
              <path key={country.name} d={country.d} fill={c.theme.surface} stroke={c.theme.bgBottom} strokeWidth={0.4} opacity={0.75} />
            );
          }
          // Shaded by value within the data's own range, so a set of similar
          // numbers still reads as a gradient rather than as one flat colour.
          // The floor matters more than the range: a country carrying a small
          // value is still a country with data, and at a low enough opacity it
          // is indistinguishable from one with none.
          const on = c.focused && c.isLit(p.label);
          return (
            <path
              key={country.name}
              d={country.d}
              fill={c.tint(p.label)}
              fillOpacity={(0.5 + 0.5 * (p.value / c.maxValue)) * reveal}
              stroke={on ? c.theme.accent : c.theme.bgBottom}
              strokeWidth={on ? 1.4 : 0.4}
              opacity={c.dim(p.label)}
            />
          );
        })}
      </g>
    </svg>
  );
};

/**
 * The kinds this scene can draw.
 *
 * Go mirrors the keys (chartKindVocab in snippet_data.go) and a drift test
 * keeps the two identical: a kind Go allows and this map does not have falls
 * through to bars, so a clip planned as a treemap renders as a bar chart and
 * the plan's own captions stop making sense.
 */
const CHARTS: Record<string, (c: Ctx) => React.ReactNode> = {
  bars: bars,
  stackedbars: (c) => barsWithSeries(c, false),
  groupedbars: (c) => barsWithSeries(c, true),
  line: (c) => line(c, false),
  area: (c) => line(c, true),
  scatter: scatter,
  donut: donut,
  waffle: waffle,
  gauge: gauge,
  treemap: treemap,
  funnel: funnel,
  kpi: kpi,
  map: worldMap,
};

/** The kinds that carry a legend, because a colour stands for a name. */
const LEGEND_KINDS = new Set(['stackedbars', 'groupedbars']);

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
  const series = (Array.isArray(props.series) ? props.series : []) as string[];
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

  const draw = CHARTS[kind] ?? CHARTS.bars;
  const showLegend = LEGEND_KINDS.has(kind) && series.length > 0;
  const chartH = BOARD_H - (caption ? 74 : 24) - (showLegend ? 56 : 0);

  const isLit = (label: string): boolean => !lit || lit.has(label.toLowerCase());
  const ctx: Ctx = {
    theme,
    points,
    series,
    unit,
    h: chartH,
    w: BOARD_W,
    frame,
    maxValue: Math.max(...points.flatMap((p) => (p.values?.length ? p.values : [p.value])), 1),
    maxTotal: Math.max(
      ...points.map((p) => (p.values?.length ? p.values.reduce((s, v) => s + v, 0) : p.value)),
      1,
    ),
    isLit,
    focused: Boolean(lit),
    dim: (label) => (isLit(label) ? 1 : 1 - (1 - DIM) * focusP),
    tint: (label) => (lit && isLit(label) ? theme.accent : theme.primary),
    valueColor: (label) => (lit && isLit(label) ? theme.accentText : theme.textMuted),
    grow: (i) =>
      spring({
        frame: frame - i * ENTER.stagger,
        fps,
        config: {damping: 200, mass: 0.7},
        durationInFrames: ENTER.frames,
      }),
    seriesFill: (i) => seriesFillFor(theme, i),
    seriesOpacity: seriesOpacityFor,
  };

  return (
    <Stage justify="flex-start">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={22} />
      <div style={{display: 'flex', justifyContent: 'center', width: BOARD_W}}>{draw(ctx)}</div>
      {showLegend && <Legend ctx={ctx} />}
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
