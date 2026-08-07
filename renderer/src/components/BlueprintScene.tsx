import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// BlueprintScene: the box drawn as a board, so a signal has somewhere to go.
//
// A foundations course spends most of its life naming parts — CPU, RAM, bus,
// disk — and a list of names is exactly the wrong picture, because the whole
// claim is that the parts are WIRED. So the composition is a board: labelled
// blocks pinned to a grid, orthogonal wires between them, and a pulse that
// travels one wire at a time. The pulse is the argument. Everything else is
// scaffolding that holds still while it runs.
//
// Placement is computed here rather than left to flexbox, for the same reason
// MachineScene packs its own rows: the wires need real coordinates, and a wrap
// layout does not surrender them. Blocks take grid slots ordered by ROLE —
// io first, then units, then stores — which turns the default reading order
// into a rough dataflow: things enter at the top, get worked on in the middle,
// come to rest at the bottom. The block's plan index is preserved through the
// slot map, so a wire declared between block 0 and block 4 still lands on the
// right two rectangles no matter where the roles put them.
//
// Wires are orthogonal, never diagonal, because a diagonal on a block diagram
// reads as "roughly connected" and an elbow reads as a bus. Same-row wires run
// straight between facing edges; cross-row wires exit the source's underside,
// travel to the channel between the rows, cross, and enter the target's top.
// The channel gets a small per-wire offset so two wires crossing the same gap
// do not stack into one thick line.
//
// One glow maximum: the focused block's accent edge. The pulse is a moving
// dash, not a light — a lit trail plus a lit block is two competing subjects,
// and the pulse only exists to say WHICH wire.

const BOARD_W = Math.min(STAGE_W, 1300);
const BLOCK_H = 108;
const COL_GAP = 46;
// Deep enough that a cross-row wire has a channel to turn in, and its label
// has somewhere to sit that is not on top of a block.
const ROW_GAP = 104;

type Block = {id: string; label: string; role: 'unit' | 'store' | 'io'};
type Wire = {from: number; to: number; label: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'board' | 'block' | 'path' | 'whole';
  at?: number;
  lit: number[];
};

type Box = {x: number; y: number; w: number; h: number};

const ROLE_RANK: Record<string, number> = {io: 0, unit: 1, store: 2};

/** Two per row for four blocks, three otherwise — never a lonely wide row. */
const colsFor = (n: number): number => (n === 4 ? 2 : Math.min(n, 3));

/**
 * Grid slots in role order, returned indexed by the block's PLAN index so
 * wires — which address blocks by plan index — need no translation.
 */
const layout = (blocks: Block[]): {boxes: Box[]; height: number} => {
  const order = blocks
    .map((b, i) => ({i, rank: ROLE_RANK[b.role] ?? 1}))
    .sort((a, b) => (a.rank === b.rank ? a.i - b.i : a.rank - b.rank))
    .map((e) => e.i);

  const cols = colsFor(blocks.length);
  const rows = Math.ceil(blocks.length / cols);
  const boxes: Box[] = new Array(blocks.length);
  for (let r = 0; r < rows; r++) {
    const slots = order.slice(r * cols, r * cols + cols);
    if (slots.length === 0) continue;
    const w = Math.floor((BOARD_W - COL_GAP * (cols - 1)) / cols);
    const used = w * slots.length + COL_GAP * (slots.length - 1);
    let x = (BOARD_W - used) / 2;
    const y = r * (BLOCK_H + ROW_GAP);
    for (const i of slots) {
      boxes[i] = {x, y, w, h: BLOCK_H};
      x += w + COL_GAP;
    }
  }
  return {boxes, height: rows * BLOCK_H + (rows - 1) * ROW_GAP};
};

type Route = {pts: number[][]; d: string; len: number; mid: number[]};

/** An orthogonal route between two boxes, with its own length precomputed. */
const route = (a: Box, b: Box, nudge: number): Route => {
  let pts: number[][];
  if (Math.abs(a.y - b.y) < 1) {
    const y = a.y + a.h / 2 + nudge;
    const forward = a.x < b.x;
    pts = [
      [forward ? a.x + a.w : a.x, y],
      [forward ? b.x : b.x + b.w, y],
    ];
  } else {
    const down = b.y > a.y;
    const y1 = down ? a.y + a.h : a.y;
    const y2 = down ? b.y : b.y + b.h;
    const midY = (y1 + y2) / 2 + nudge;
    const x1 = a.x + a.w / 2;
    const x2 = b.x + b.w / 2;
    pts = [
      [x1, y1],
      [x1, midY],
      [x2, midY],
      [x2, y2],
    ];
  }
  let len = 0;
  for (let i = 1; i < pts.length; i++) {
    len += Math.abs(pts[i][0] - pts[i - 1][0]) + Math.abs(pts[i][1] - pts[i - 1][1]);
  }
  const d = pts.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p[0]} ${p[1]}`).join(' ');
  // The label rides the longest horizontal run, which on both route shapes is
  // the segment that actually crosses the gap.
  let best = 0;
  let bestRun = -1;
  for (let i = 1; i < pts.length; i++) {
    const run = Math.abs(pts[i][0] - pts[i - 1][0]);
    if (run > bestRun) {
      bestRun = run;
      best = i;
    }
  }
  const mid = [(pts[best][0] + pts[best - 1][0]) / 2, (pts[best][1] + pts[best - 1][1]) / 2];
  return {pts, d, len, mid};
};

export const BlueprintScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const blocks = (Array.isArray(props.blocks) ? props.blocks : []) as Block[];
  const wires = (Array.isArray(props.wires) ? props.wires : []) as Wire[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (blocks.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const focused = step.show === 'block' ? step.at ?? -1 : -1;
  const pulsing = step.show === 'path' ? step.at ?? -1 : -1;
  const whole = step.show === 'whole';
  const lit = new Set(Array.isArray(step.lit) ? step.lit : []);

  const {boxes, height: boardH} = layout(blocks);
  const lift = spring({frame: sinceStep - 2, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 26});
  const settle = whole
    ? spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 28})
    : 0;

  // How far the pulse has travelled along its wire, 0-1 across the beat.
  const stepFrames = Math.max(1, ((step.endMs - step.startMs) / 1000) * FPS);
  const travel = interpolate(sinceStep, [4, stepFrames * 0.82], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  // Blocks a pulsed wire touches read as "on the path" for the rest of the
  // clip, so the board accumulates rather than resetting each beat.
  const onPath = new Set<number>();
  for (const w of lit) {
    const wire = wires[w];
    if (!wire) continue;
    onPath.add(wire.from);
    onPath.add(wire.to);
  }

  const roleFill = (role: string, on: boolean): string => {
    if (role === 'store') return withAlpha(theme.mass, on ? 0.16 : 0.07);
    if (role === 'io') return withAlpha(theme.accentQuantity, on ? 0.14 : 0.06);
    return withAlpha(theme.surface, on ? 0.95 : 0.7);
  };

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

      <div style={{position: 'relative', width: BOARD_W, height: boardH}}>
        {/* Wires first: the blocks sit on top of their own endpoints. */}
        <svg
          width={BOARD_W}
          height={boardH}
          style={{position: 'absolute', left: 0, top: 0, overflow: 'visible', pointerEvents: 'none'}}
        >
          {wires.map((w, i) => {
            const a = boxes[w.from];
            const b = boxes[w.to];
            if (!a || !b) return null;
            const r = route(a, b, ((i % 3) - 1) * 9);
            const isPulse = i === pulsing;
            const wasLit = lit.has(i);
            const strong = isPulse || wasLit || whole;
            const draw = interpolate(sinceStep, [2 + i * 3, 20 + i * 3], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            });
            const drawn = idx === 0 ? draw : 1;
            return (
              <g key={i}>
                <path
                  d={r.d}
                  fill="none"
                  stroke={strong ? withAlpha(theme.accent, whole ? 0.5 + 0.35 * settle : 0.72) : withAlpha(theme.line, 0.34)}
                  strokeWidth={strong ? 3 : 2}
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeDasharray={r.len}
                  strokeDashoffset={r.len * (1 - drawn)}
                />
                {/* The pulse: a short dash walking the exact path length. */}
                {isPulse ? (
                  <path
                    d={r.d}
                    fill="none"
                    stroke={theme.accentQuantity}
                    strokeWidth={5}
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeDasharray={`${Math.min(90, r.len * 0.3)} ${r.len}`}
                    strokeDashoffset={r.len - travel * (r.len + Math.min(90, r.len * 0.3))}
                  />
                ) : null}
                {w.label ? (
                  <g transform={`translate(${r.mid[0]}, ${r.mid[1]})`}>
                    <rect
                      x={-(w.label.length * 4.6 + 12)}
                      y={-14}
                      width={w.label.length * 9.2 + 24}
                      height={28}
                      rx={7}
                      fill={theme.surface}
                      stroke={strong ? withAlpha(theme.accent, 0.45) : theme.surfaceBorder}
                      strokeWidth={1.5}
                      opacity={drawn}
                    />
                    <text
                      x={0}
                      y={5}
                      textAnchor="middle"
                      fill={strong ? theme.accentText : theme.textMuted}
                      opacity={drawn}
                      style={{fontFamily: theme.fontMono, fontSize: 15, letterSpacing: 0.6}}
                    >
                      {w.label}
                    </text>
                  </g>
                ) : null}
              </g>
            );
          })}
        </svg>

        {blocks.map((b, i) => {
          const box = boxes[i];
          if (!box) return null;
          const isFocus = i === focused;
          const touched = onPath.has(i);
          const rise = isFocus ? lift : 0;
          const dim = whole ? 1 : isFocus ? 1 : touched ? 0.82 : step.show === 'board' ? 0.7 : 0.5;
          const edge = isFocus
            ? withAlpha(theme.accent, 0.55 + 0.45 * rise)
            : whole
              ? withAlpha(theme.accent, 0.3 + 0.3 * settle)
              : touched
                ? withAlpha(theme.accent, 0.34)
                : theme.surfaceBorder;
          return (
            <div
              key={b.id || i}
              style={{
                position: 'absolute',
                left: box.x,
                top: box.y,
                width: box.w,
                height: box.h,
                borderRadius: b.role === 'store' ? 8 : b.role === 'io' ? 54 : 14,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 6,
                background: roleFill(b.role, isFocus || whole || touched),
                border: `2px solid ${edge}`,
                opacity: dim,
                transform: `translateY(${-10 * rise}px) scale(${1 + 0.035 * rise})`,
                // The one glow: whichever block the voiceover is naming.
                boxShadow: isFocus ? `0 0 ${34 * rise}px ${withAlpha(theme.accent, 0.34)}` : undefined,
                zIndex: isFocus ? 2 : 1,
              }}
            >
              <span
                style={{
                  fontFamily: theme.fontDisplay,
                  fontSize: 28,
                  fontWeight: 700,
                  letterSpacing: -0.4,
                  color: isFocus || whole ? theme.text : theme.textMuted,
                  textAlign: 'center',
                  paddingInline: 14,
                  lineHeight: 1.1,
                }}
              >
                {b.label}
              </span>
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 12,
                  letterSpacing: 2.4,
                  textTransform: 'uppercase',
                  color: isFocus ? theme.accentText : withAlpha(theme.textMuted, 0.75),
                }}
              >
                {b.role}
              </span>
            </div>
          );
        })}
      </div>
    </Stage>
  );
};
