// Geometry for the layered systems diagram.
//
// The graph's *shape* is decided in Go (each node arrives with a rank and an
// order within it, from a longest-path layering); everything here is pixels.
// The split matters: ranking is graph logic that wants unit tests, placement is
// a function of the stage box, and keeping them apart means neither can quietly
// depend on the other.

export type Pt = {x: number; y: number};
export type Rect = {x: number; y: number; w: number; h: number};

/** A cubic bezier, with the sampling helpers the animation needs. */
export type Curve = {
  d: string;
  length: number;
  /** Point at parameter t (0-1). Not arc-length parametrized; see below. */
  at: (t: number) => Pt;
};

/**
 * Places `n` nodes given their column (rank) and row (order within the column).
 *
 * Columns are evenly spaced across the board and each column's rows are centred
 * vertically, so a column holding one node sits level with the middle of a
 * column holding three. That vertical centring is what stops a branch from
 * reading as if it were dangling off the bottom.
 */
export const flowLayout = (
  nodes: {rank: number; order: number}[],
  ranks: number,
  boardW: number,
  boardH: number,
): Rect[] => {
  const perRank: number[] = new Array(ranks).fill(0);
  nodes.forEach((n) => {
    perRank[n.rank] = Math.max(perRank[n.rank], n.order + 1);
  });
  const tallest = Math.max(1, ...perRank);

  const colGap = ranks > 1 ? 92 : 0;
  const rowGap = 44;
  // Width cap by column count, for the same reason the whiteboard's scales by
  // count: a fixed cap leaves a three-column diagram of single boxes with a
  // third of the frame unused.
  const maxW = ranks <= 2 ? 560 : ranks === 3 ? 470 : 410;
  const w = Math.min((boardW - colGap * (ranks - 1)) / ranks, maxW);
  // The aspect and absolute caps were tight enough that a two-row diagram used
  // barely half the board's height, leaving a wide gap under the header and an
  // equal one below the graph. Bigger boxes close it from both ends at once.
  const h = Math.min(
    Math.max(96, (boardH - rowGap * (tallest - 1)) / tallest),
    w * 0.7,
    250,
  );

  // Centre the whole set of columns, so a two-column diagram is not pinned left.
  const gridW = ranks * w + colGap * (ranks - 1);
  const left = (boardW - gridW) / 2;

  return nodes.map((n) => {
    const count = perRank[n.rank];
    const colH = count * h + rowGap * (count - 1);
    return {
      x: left + n.rank * (w + colGap),
      y: (boardH - colH) / 2 + n.order * (h + rowGap),
      w,
      h,
    };
  });
};

/**
 * An S-curve from the right edge of `a` to the left edge of `b`.
 *
 * Control points are pulled horizontally, which gives the flat entry and exit
 * that make a layered diagram legible: every edge leaves a box going right and
 * arrives going right, so the eye never has to work out which end is which.
 * Same-column edges (a rank the layering could not separate) fall back to a
 * vertical bow.
 */
export const flowCurve = (a: Rect, b: Rect): Curve => {
  const sameColumn = Math.abs(a.x - b.x) < 1;
  // Small gaps at both ends: the arrowhead needs to sit clear of the target's
  // border, or it lands on top of the kind stripe just inside it.
  const START_GAP = 4;
  const END_GAP = 10;
  const p0 = sameColumn
    ? {x: a.x + a.w / 2, y: a.y + a.h + START_GAP}
    : {x: a.x + a.w + START_GAP, y: a.y + a.h / 2};
  const p3 = sameColumn
    ? {x: b.x + b.w / 2, y: b.y - END_GAP}
    : {x: b.x - END_GAP, y: b.y + b.h / 2};

  const pull = sameColumn
    ? Math.max(28, Math.abs(p3.y - p0.y) * 0.45)
    : Math.max(46, (p3.x - p0.x) * 0.5);
  const p1 = sameColumn ? {x: p0.x, y: p0.y + pull} : {x: p0.x + pull, y: p0.y};
  const p2 = sameColumn ? {x: p3.x, y: p3.y - pull} : {x: p3.x - pull, y: p3.y};

  const at = (t: number): Pt => {
    const u = 1 - t;
    const b0 = u * u * u;
    const b1 = 3 * u * u * t;
    const b2 = 3 * u * t * t;
    const b3 = t * t * t;
    return {
      x: b0 * p0.x + b1 * p1.x + b2 * p2.x + b3 * p3.x,
      y: b0 * p0.y + b1 * p1.y + b2 * p2.y + b3 * p3.y,
    };
  };

  // Polyline approximation of the length, for stroke-dasharray.
  let length = 0;
  let prev = at(0);
  for (let i = 1; i <= 24; i++) {
    const next = at(i / 24);
    length += Math.hypot(next.x - prev.x, next.y - prev.y);
    prev = next;
  }

  const d =
    `M${p0.x.toFixed(2)} ${p0.y.toFixed(2)} ` +
    `C${p1.x.toFixed(2)} ${p1.y.toFixed(2)} ` +
    `${p2.x.toFixed(2)} ${p2.y.toFixed(2)} ` +
    `${p3.x.toFixed(2)} ${p3.y.toFixed(2)}`;

  return {d, length: Math.max(1, length), at};
};

/** Arrowhead path at the curve's end, aimed along its final direction. */
export const curveHead = (curve: Curve, size = 13): string => {
  const tip = curve.at(1);
  const back = curve.at(0.93);
  const dx = tip.x - back.x;
  const dy = tip.y - back.y;
  const len = Math.hypot(dx, dy) || 1;
  const ux = dx / len;
  const uy = dy / len;
  const spread = 0.5;
  const wing = (sign: number) => ({
    x: tip.x - size * (ux * Math.cos(spread) - sign * uy * Math.sin(spread)),
    y: tip.y - size * (uy * Math.cos(spread) + sign * ux * Math.sin(spread)),
  });
  const w1 = wing(1);
  const w2 = wing(-1);
  return `M${w1.x.toFixed(2)} ${w1.y.toFixed(2)} L${tip.x.toFixed(2)} ${tip.y.toFixed(2)} L${w2.x.toFixed(2)} ${w2.y.toFixed(2)}`;
};
