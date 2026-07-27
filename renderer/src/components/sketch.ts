import {random} from 'remotion';

// Hand-drawn stroke geometry for the whiteboard scene.
//
// Everything here is analytic and seeded: a stroke is a polyline of sampled
// points, so the same seed always produces the same wobble (Remotion renders
// frames out of order and in parallel — anything else flickers), and because we
// own the points we also get two things for free that an imported SVG cannot
// give us: the exact path length, for a stroke-dashoffset draw-on, and the
// position of the pen at any moment, for the marker that leads it.
//
// The wobble is low-frequency on purpose. Per-point white noise reads as a
// jagged mistake; two seeded sine harmonics along the path read as a hand.

export type Pt = {x: number; y: number};

export type Stroke = {
  /** SVG path data. */
  d: string;
  /** Total length, for stroke-dasharray. */
  length: number;
  /** The sampled points, so callers can find the pen. */
  points: Pt[];
};

/**
 * Smooth seeded displacement at position t (0-1) along a stroke.
 *
 * The harmonics are whole cycles, which makes the function periodic in t. That
 * matters for closed shapes: the stroke is sampled slightly past t=1 so it
 * overlaps its own start, and with a non-periodic wobble the two passes landed
 * a few pixels apart and read as a stray floating dash rather than a pen set
 * down twice.
 */
const wobble = (t: number, seed: string, amp: number): number => {
  const p1 = random(`${seed}-w1`) * Math.PI * 2;
  const p2 = random(`${seed}-w2`) * Math.PI * 2;
  const tau = Math.PI * 2;
  return amp * (0.62 * Math.sin(tau * 3 * t + p1) + 0.38 * Math.sin(tau * 7 * t + p2));
};

/** Builds a Stroke from points, computing its path data and length. */
const strokeFrom = (points: Pt[]): Stroke => {
  let length = 0;
  for (let i = 1; i < points.length; i++) {
    length += Math.hypot(points[i].x - points[i - 1].x, points[i].y - points[i - 1].y);
  }
  const d = points
    .map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(2)} ${p.y.toFixed(2)}`)
    .join(' ');
  return {d, length: Math.max(1, length), points};
};

/** The point on a stroke at progress p (0-1) — where the pen is. */
export const penAt = (stroke: Stroke, p: number): Pt => {
  const clamped = Math.max(0, Math.min(1, p));
  const i = Math.round(clamped * (stroke.points.length - 1));
  return stroke.points[i];
};

/**
 * A hand-drawn rounded rectangle.
 *
 * Sampling runs slightly past both ends so the stroke crosses itself at the
 * start corner — the tell that a person drew it rather than a rect element.
 */
export const roughRect = (
  w: number,
  h: number,
  radius: number,
  seed: string,
  amp = 5,
): Stroke => {
  const r = Math.min(radius, w / 2, h / 2);
  // Perimeter segments as parametric functions, with their outward normals.
  const arc = (cx: number, cy: number, a0: number, a1: number) => (u: number) => {
    const a = a0 + (a1 - a0) * u;
    return {p: {x: cx + r * Math.cos(a), y: cy + r * Math.sin(a)}, n: {x: Math.cos(a), y: Math.sin(a)}};
  };
  const line = (x0: number, y0: number, x1: number, y1: number, nx: number, ny: number) => (u: number) => ({
    p: {x: x0 + (x1 - x0) * u, y: y0 + (y1 - y0) * u},
    n: {x: nx, y: ny},
  });

  const segments: {f: (u: number) => {p: Pt; n: Pt}; len: number}[] = [
    {f: line(r, 0, w - r, 0, 0, -1), len: w - 2 * r},
    {f: arc(w - r, r, -Math.PI / 2, 0), len: (Math.PI / 2) * r},
    {f: line(w, r, w, h - r, 1, 0), len: h - 2 * r},
    {f: arc(w - r, h - r, 0, Math.PI / 2), len: (Math.PI / 2) * r},
    {f: line(w - r, h, r, h, 0, 1), len: w - 2 * r},
    {f: arc(r, h - r, Math.PI / 2, Math.PI), len: (Math.PI / 2) * r},
    {f: line(0, h - r, 0, r, -1, 0), len: h - 2 * r},
    {f: arc(r, r, Math.PI, (3 * Math.PI) / 2), len: (Math.PI / 2) * r},
  ];
  const total = segments.reduce((a, s) => a + s.len, 0);

  // The stroke starts and ends mid-top-edge rather than at a corner, so the
  // overshoot crosses itself along a straight run — which reads as a pen laid
  // down twice. Crossing *at* the corner produced a stray tick that looked like
  // a rendering glitch instead of a hand.
  const phase = segments[0].len / 2 / total;

  // at(t) walks the whole perimeter with t in 0-1, wrapping.
  const at = (t: number) => {
    const wrapped = (((t + phase) % 1) + 1) % 1;
    let target = wrapped * total;
    for (const seg of segments) {
      if (target <= seg.len || seg === segments[segments.length - 1]) {
        return seg.f(Math.max(0, Math.min(1, target / seg.len)));
      }
      target -= seg.len;
    }
    return segments[0].f(0);
  };

  const samples = Math.max(48, Math.round(total / 9));
  const from = 0;
  const to = 1.022;
  const points: Pt[] = [];
  for (let i = 0; i <= samples; i++) {
    const t = from + ((to - from) * i) / samples;
    const {p, n} = at(t);
    const off = wobble(t, seed, amp);
    points.push({x: p.x + n.x * off, y: p.y + n.y * off});
  }
  return strokeFrom(points);
};

/**
 * A hand-drawn ellipse filling w x h.
 *
 * Circling something is what a person at a board does to the thing that is not
 * a component — an actor, a moment, a question. Same overshoot as roughRect so
 * the pen visibly crosses its own start.
 */
export const roughEllipse = (w: number, h: number, seed: string, amp = 5): Stroke => {
  const rx = w / 2;
  const ry = h / 2;
  const samples = Math.max(56, Math.round((rx + ry) / 5));
  const to = 1.03;
  const points: Pt[] = [];
  for (let i = 0; i <= samples; i++) {
    const t = (to * i) / samples;
    const a = t * Math.PI * 2 - Math.PI / 2;
    // The outward normal of an ellipse is not the radial direction, but at
    // these aspect ratios the difference is under a pixel and the radial one
    // keeps the wobble periodic, which is what stops the overshoot ticking.
    const off = wobble(t, seed, amp);
    points.push({
      x: rx + Math.cos(a) * (rx + off),
      y: ry + Math.sin(a) * (ry + off),
    });
  }
  return strokeFrom(points);
};

/**
 * A hand-drawn cloud: bumps around an ellipse.
 *
 * The shape for something deliberately vague — "the internet", "everything
 * else", a system nobody is explaining today. A box says the thing has edges;
 * a cloud says the opposite, which is often the honest drawing.
 */
export const roughCloud = (w: number, h: number, seed: string, amp = 3): Stroke => {
  const rx = w / 2;
  const ry = h / 2;
  const bumps = 7;
  const samples = Math.max(96, Math.round((rx + ry) / 3));
  const to = 1.02;
  const points: Pt[] = [];
  for (let i = 0; i <= samples; i++) {
    const t = (to * i) / samples;
    const a = t * Math.PI * 2 - Math.PI / 2;
    // A scalloped radius, bulging *outward only*. A plain sine alternates in
    // and out and draws a star: the inward halves cut back past the body and
    // every bump ends in a point. Rectifying it leaves rounded lobes with flat
    // valleys between them, which is the shape of a cloud. The bump count is a
    // whole number of cycles so the scallop still closes on itself, for the
    // same reason the wobble harmonics are.
    const lobe = Math.max(0, Math.sin(bumps * t * Math.PI * 2));
    const scallop = 1 + 0.13 * Math.pow(lobe, 0.65);
    const off = wobble(t, seed, amp);
    points.push({
      x: rx + Math.cos(a) * (rx * scallop + off),
      y: ry + Math.sin(a) * (ry * scallop + off),
    });
  }
  return strokeFrom(points);
};

/**
 * A hand-drawn sticky note: a square-ish card with one corner turned up.
 *
 * Returns the outline and the fold separately, because the fold is a filled
 * triangle rather than part of the perimeter — a note whose corner is only an
 * outline reads as a box with a scratch on it.
 */
export const roughSticky = (
  w: number,
  h: number,
  seed: string,
  amp = 4,
): {outline: Stroke; fold: Stroke; foldPath: string} => {
  const cut = Math.min(34, w * 0.16, h * 0.24);
  const corners: Pt[] = [
    {x: 0, y: 0},
    {x: w, y: 0},
    {x: w, y: h - cut},
    {x: w - cut, y: h},
    {x: 0, y: h},
  ];
  // Walk the perimeter by *length*, not by segment index.
  //
  // Sampling each edge with the same number of points regardless of how long it
  // is leaves the points non-uniform in arc length — and penAt() finds the pen
  // by point index while the draw-on advances by arc length, so the two
  // disagree and the marker floats off the end of its own stroke. It is only
  // visible while a shape is being drawn, which is the one moment the marker
  // exists for.
  const segs = corners.map((a, i) => {
    const b = corners[(i + 1) % corners.length];
    const dx = b.x - a.x;
    const dy = b.y - a.y;
    const len = Math.hypot(dx, dy) || 1;
    // Outward normal of this edge, so the wobble pushes off the paper rather
    // than along it.
    return {a, dx, dy, len, nx: dy / len, ny: -dx / len};
  });
  const total = segs.reduce((sum, s) => sum + s.len, 0);
  const samples = Math.max(64, Math.round(total / 9));
  const perimeter: Pt[] = [];
  for (let i = 0; i <= samples; i++) {
    const t = (1.015 * i) / samples;
    let target = (((t % 1) + 1) % 1) * total;
    let seg = segs[0];
    let local = 0;
    for (const s of segs) {
      if (target <= s.len || s === segs[segs.length - 1]) {
        seg = s;
        local = Math.max(0, Math.min(1, target / s.len));
        break;
      }
      target -= s.len;
    }
    const off = wobble(t, seed, amp);
    perimeter.push({
      x: seg.a.x + seg.dx * local + seg.nx * off,
      y: seg.a.y + seg.dy * local + seg.ny * off,
    });
  }
  const fold = strokeFrom([
    {x: w - cut, y: h},
    {x: w - cut, y: h - cut},
    {x: w, y: h - cut},
  ]);
  return {
    outline: strokeFrom(perimeter),
    fold,
    foldPath: `M${w - cut} ${h} L${w - cut} ${h - cut} L${w} ${h - cut} Z`,
  };
};

/**
 * A highlighter swipe: a thick band with ragged ends, drawn behind text.
 *
 * Not a rounded rect — a marker laid down and lifted leaves a stroke that is
 * thicker in the middle and never quite level, and that unevenness is the only
 * thing separating "highlighted" from "has a coloured rectangle behind it".
 */
export const roughHighlight = (w: number, h: number, seed: string): string => {
  const wob = (t: number) => wobble(t, seed, h * 0.12);
  const steps = 12;
  const top: string[] = [];
  const bottom: string[] = [];
  for (let i = 0; i <= steps; i++) {
    const t = i / steps;
    const x = w * t;
    // Thicker through the middle, the way pressure actually falls off.
    const swell = Math.sin(t * Math.PI) * h * 0.1;
    top.push(`${x.toFixed(1)} ${(wob(t) - swell).toFixed(1)}`);
    bottom.push(`${x.toFixed(1)} ${(h + wob(1 - t) + swell).toFixed(1)}`);
  }
  return `M${top.join(' L')} L${bottom.reverse().join(' L')} Z`;
};

/**
 * A hand-drawn connector between two points, bowed gently perpendicular to its
 * run so a grid of boxes reads as a drawn diagram rather than a wiring harness.
 */
export const roughArrow = (a: Pt, b: Pt, seed: string, amp = 3): {shaft: Stroke; head: Stroke} => {
  const dx = b.x - a.x;
  const dy = b.y - a.y;
  const len = Math.hypot(dx, dy) || 1;
  const ux = dx / len;
  const uy = dy / len;
  // Perpendicular, for the bow and the wobble.
  const px = -uy;
  const py = ux;
  const bow = (random(`${seed}-bow`) - 0.5) * Math.min(46, len * 0.16);
  // The wobble's harmonics are fixed cycles over the whole stroke, so on a
  // short connector they crowd into a zigzag that reads as a lightning bolt.
  // Fading the amplitude out below a couple of hundred pixels keeps short hops
  // nearly straight, which is what a hand actually draws over that distance.
  const localAmp = amp * Math.min(1, len / 240);

  const samples = Math.max(24, Math.round(len / 10));
  const points: Pt[] = [];
  for (let i = 0; i <= samples; i++) {
    const t = i / samples;
    // The sin envelope zeroes both the bow and the wobble at the ends, so the
    // shaft still meets the boxes it connects instead of floating off them.
    const envelope = Math.sin(t * Math.PI);
    const curve = envelope * (bow + wobble(t, seed, localAmp));
    points.push({x: a.x + dx * t + px * curve, y: a.y + dy * t + py * curve});
  }
  const shaft = strokeFrom(points);

  // Arrowhead: two strokes swept back from the tip along the incoming
  // direction, which is the shaft's last segment rather than the straight line
  // (the bow means those differ).
  const last = points[points.length - 1];
  const prev = points[Math.max(0, points.length - 4)];
  const hx = last.x - prev.x;
  const hy = last.y - prev.y;
  const hlen = Math.hypot(hx, hy) || 1;
  const ax = hx / hlen;
  const ay = hy / hlen;
  // Sized to sit alongside a 4px shaft. A 20px head on a board of 3.4px box
  // outlines read as a tick rather than as an arrow.
  const size = 25;
  const spread = 0.42;
  const wing = (sign: number): Pt => ({
    x: last.x - size * (ax * Math.cos(spread) - sign * ay * Math.sin(spread)),
    y: last.y - size * (ay * Math.cos(spread) + sign * ax * Math.sin(spread)),
  });
  const head = strokeFrom([wing(1), last, wing(-1)]);
  return {shaft, head};
};

/** A hand-drawn underline, for emphasis beneath a heading. */
export const roughUnderline = (w: number, seed: string, amp = 2): Stroke => {
  const samples = Math.max(16, Math.round(w / 12));
  const points: Pt[] = [];
  for (let i = 0; i <= samples; i++) {
    const t = i / samples;
    points.push({x: w * t, y: wobble(t, seed, amp)});
  }
  return strokeFrom(points);
};

export type BoxRect = {x: number; y: number; w: number; h: number};

/**
 * Lays `n` boxes out on the board.
 *
 * The grid shape is chosen by count rather than by the content, so a clip can
 * never come out lopsided: rows are balanced (five items are 3 + 2, not 3 + 1 +
 * 1) and each row is centred, which is what makes a partly-filled last row look
 * deliberate.
 *
 * Rows alternate direction — item 3 sits *below* item 2, not back at the left
 * margin. Reading order snakes, which is unusual for a grid and exactly right
 * here: these boxes are almost always a chain, and a straight left-to-right
 * second row put the connector on a long diagonal straight through the boxes in
 * between. Serpentine makes every consecutive pair adjacent, so every arrow is
 * a short hop.
 */
export const boardLayout = (n: number, boardW: number, boardH: number): BoxRect[] => {
  const cols = n <= 3 ? n : n === 4 ? 2 : n <= 6 ? 3 : 4;
  const rows = Math.ceil(n / cols);
  const gap = rows > 1 ? 84 : 92;

  // Box size caps scale with the grid, not a constant. A fixed cap made a
  // three-item board three small cards adrift in a 1920x1080 frame, and a
  // four-item 2x2 a narrow block with a third of the width unused. Fewer
  // columns means each box gets to be bigger.
  const maxW = cols <= 2 ? 540 : cols === 3 ? (rows === 1 ? 500 : 450) : 390;
  const maxH = rows === 1 ? 330 : 272;

  const w = Math.min((boardW - gap * (cols - 1)) / cols, maxW);
  const h = Math.min(w * 0.62, (boardH - gap * (rows - 1)) / rows, maxH);
  const gridH = rows * h + gap * (rows - 1);
  const top = (boardH - gridH) / 2;

  // Balanced row counts: the remainder is spread over the first rows.
  const counts: number[] = [];
  let left = n;
  for (let r = 0; r < rows; r++) {
    const take = Math.ceil(left / (rows - r));
    counts.push(take);
    left -= take;
  }

  const out: BoxRect[] = [];
  counts.forEach((count, r) => {
    const rowW = count * w + gap * (count - 1);
    const x0 = (boardW - rowW) / 2;
    // Odd rows are laid right to left so the chain snakes back.
    const cells: BoxRect[] = [];
    for (let c = 0; c < count; c++) {
      cells.push({x: x0 + c * (w + gap), y: top + r * (h + gap), w, h});
    }
    // Right-align the reversed row so the snake turns at the same edge the
    // previous row ended on, keeping the vertical hop short.
    if (r % 2 === 1) {
      const prevRowW = counts[r - 1] * w + gap * (counts[r - 1] - 1);
      const shift = (boardW - prevRowW) / 2 + prevRowW - (x0 + rowW);
      cells.forEach((cell) => {
        cell.x += shift;
      });
      cells.reverse();
    }
    out.push(...cells);
  });
  return out;
};

/** The midpoint of the box edge facing `toward`, plus a small outward margin. */
export const edgeAnchor = (box: BoxRect, toward: BoxRect, margin = 10): Pt => {
  const cx = box.x + box.w / 2;
  const cy = box.y + box.h / 2;
  const tx = toward.x + toward.w / 2;
  const ty = toward.y + toward.h / 2;
  const dx = tx - cx;
  const dy = ty - cy;
  // Compare in box-relative units so a wide box still prefers its side edges
  // for a mostly-horizontal run.
  if (Math.abs(dx) / box.w >= Math.abs(dy) / box.h) {
    return {x: dx >= 0 ? box.x + box.w + margin : box.x - margin, y: cy};
  }
  return {x: cx, y: dy >= 0 ? box.y + box.h + margin : box.y - margin};
};
