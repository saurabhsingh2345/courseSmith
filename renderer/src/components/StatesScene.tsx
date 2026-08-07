import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// StatesScene: a small machine, and one dot that is only ever in one place.
//
// The graph is drawn quiet and the TOKEN is the only loud thing on the stage.
// That inversion is the whole design. A conventional state diagram gives every
// circle and every arrow the same weight, which is exactly what leaves a
// beginner with a picture of somewhere they could be in all at once; here the
// pills sit back at a third of their weight, one dot is at full brightness
// with the only glow in the frame, and the eye has no choice about what the
// subject is.
//
// Nodes auto-arrange into one row up to three states and two rows beyond,
// because the alternative — a radial layout, which is what a graph library
// would give — makes the arc lengths wildly unequal and the token's slide
// speed read as meaningful when it is not. Two even rows keep every transition
// roughly the same length, so a slide always takes the same time to say the
// same thing.
//
// Arcs are quadratic curves bowed along their own normal, which means an arc
// from A to B and one from B to A bow to opposite sides for free, and a pair
// of states with traffic both ways reads as two roads rather than one road
// drawn twice. Endpoints are pulled back to the pill's boundary with an
// ellipse intersection so the arrowhead lands ON the capsule rather than under
// it. A self-transition gets a loop above its own pill and the token pulses in
// place, because a dot sliding back to where it started is a dot that appears
// not to have moved.
//
// Fired arcs stay at 70% for the rest of the clip. That persistence is the
// closer: by the "walk" beat the route is a visibly brighter subgraph, which
// is the frame somebody screenshots.

const BOARD_W = Math.min(STAGE_W, 1520);
const BOARD_H = 470;
const PILL_W = 236;
const PILL_H = 86;
const BOW = 62;
const ROW_ONE_MAX = 3;

type Node = {id: string; label: string};
type Arc = {from: number; to: number; on: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'machine' | 'fire' | 'state' | 'walk';
  at?: number;
  from: number;
  token: number;
  lit: number[];
};
type Pt = {x: number; y: number};

// One row up to three states, two even rows beyond, so every arc comes out
// roughly the same length.
const placeNodes = (n: number): Pt[] => {
  const rows: number[][] = [];
  if (n <= ROW_ONE_MAX) {
    rows.push([...Array(n).keys()]);
  } else {
    const top = Math.ceil(n / 2);
    rows.push([...Array(top).keys()]);
    rows.push([...Array(n - top).keys()].map((i) => i + top));
  }
  const pts: Pt[] = [];
  rows.forEach((row, r) => {
    const cy = rows.length === 1 ? BOARD_H / 2 : r === 0 ? BOARD_H * 0.26 : BOARD_H * 0.76;
    const span = BOARD_W - PILL_W;
    row.forEach((idx, i) => {
      const x = row.length === 1 ? BOARD_W / 2 : PILL_W / 2 + (span * i) / (row.length - 1);
      pts[idx] = {x, y: cy};
    });
  });
  return pts;
};

const control = (a: Pt, b: Pt): Pt => {
  const dx = b.x - a.x;
  const dy = b.y - a.y;
  const len = Math.hypot(dx, dy) || 1;
  return {x: (a.x + b.x) / 2 - (dy / len) * BOW, y: (a.y + b.y) / 2 + (dx / len) * BOW};
};

// Pull an endpoint back to the pill's boundary, so the arrowhead lands on the
// capsule instead of under it.
const onRim = (centre: Pt, toward: Pt): Pt => {
  const rx = PILL_W / 2 + 8;
  const ry = PILL_H / 2 + 8;
  const dx = toward.x - centre.x;
  const dy = toward.y - centre.y;
  const len = Math.hypot(dx, dy) || 1;
  const ux = dx / len;
  const uy = dy / len;
  const t = 1 / Math.sqrt((ux / rx) ** 2 + (uy / ry) ** 2);
  return {x: centre.x + ux * t, y: centre.y + uy * t};
};

const quad = (p0: Pt, c: Pt, p1: Pt, t: number): Pt => ({
  x: (1 - t) ** 2 * p0.x + 2 * (1 - t) * t * c.x + t ** 2 * p1.x,
  y: (1 - t) ** 2 * p0.y + 2 * (1 - t) * t * c.y + t ** 2 * p1.y,
});

const loopPath = (p: Pt): string =>
  `M ${p.x - 34} ${p.y - PILL_H / 2 - 4} C ${p.x - 96} ${p.y - PILL_H / 2 - 108} ${p.x + 96} ${p.y - PILL_H / 2 - 108} ${p.x + 34} ${p.y - PILL_H / 2 - 4}`;

export const StatesScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const nodes = (Array.isArray(props.nodes) ? props.nodes : []) as Node[];
  const arcs = (Array.isArray(props.arcs) ? props.arcs : []) as Arc[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (nodes.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const slide = spring({frame: sinceStep - 4, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 30});
  const settle = spring({frame: sinceStep - 2, fps, config: {damping: 12, mass: 0.5}, durationInFrames: 24});

  const pts = placeNodes(nodes.length);
  const lit = new Set(step.lit ?? []);
  const firing = step.show === 'fire' ? step.at ?? -1 : -1;
  const dwell = step.show === 'state' ? step.at ?? -1 : -1;
  const route = step.show === 'walk';

  // Which pills the walk has actually touched, derived from the fired arcs —
  // the route is a subgraph, and the states on it should not sit as far back
  // as the ones the token never reached.
  const touched = new Set<number>([step.token, step.from]);
  lit.forEach((a) => {
    const arc = arcs[a];
    if (arc) {
      touched.add(arc.from);
      touched.add(arc.to);
    }
  });

  // Where the dot is right now. On a fire beat it rides its own arc; on every
  // other beat it sits on the state it is in.
  const tokenPt = ((): Pt => {
    const home = pts[step.token] ?? pts[0];
    if (firing < 0 || !arcs[firing]) return home;
    const arc = arcs[firing];
    const a = pts[arc.from];
    const b = pts[arc.to];
    if (!a || !b || arc.from === arc.to) return home;
    return quad(a, control(a, b), b, slide);
  })();
  const pulse = firing >= 0 && arcs[firing] && arcs[firing].from === arcs[firing].to ? 1 + 0.35 * Math.sin(sinceStep / 4) : 1;

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

      <div style={{position: 'relative', width: BOARD_W, height: BOARD_H}}>
        <svg width={BOARD_W} height={BOARD_H} style={{position: 'absolute', left: 0, top: 0, overflow: 'visible'}}>
          {arcs.map((arc, i) => {
            const a = pts[arc.from];
            const b = pts[arc.to];
            if (!a || !b) return null;
            const isFiring = i === firing;
            const wasLit = lit.has(i) && !isFiring;
            const stroke = isFiring
              ? theme.accent
              : wasLit
                ? withAlpha(theme.accent, route ? 0.85 : 0.7)
                : withAlpha(theme.line, 0.28);
            const width = isFiring ? 4 : wasLit ? 3 : 2;

            if (arc.from === arc.to) {
              return (
                <path
                  key={i}
                  d={loopPath(a)}
                  fill="none"
                  stroke={stroke}
                  strokeWidth={width}
                  strokeLinecap="round"
                />
              );
            }
            const c = control(a, b);
            const p0 = onRim(a, c);
            const p1 = onRim(b, c);
            // The arrowhead sits on the rim, pointing the way the token travels.
            const near = quad(p0, c, p1, 0.86);
            const angle = (Math.atan2(p1.y - near.y, p1.x - near.x) * 180) / Math.PI;
            const draw = isFiring
              ? interpolate(sinceStep, [2, 24], [0.2, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
              : 1;
            return (
              <g key={i}>
                <path
                  d={`M ${p0.x} ${p0.y} Q ${c.x} ${c.y} ${p1.x} ${p1.y}`}
                  fill="none"
                  stroke={stroke}
                  strokeWidth={width}
                  strokeLinecap="round"
                  pathLength={1}
                  strokeDasharray={isFiring ? `${draw} 1` : undefined}
                />
                <polygon
                  points="0,-7 15,0 0,7"
                  fill={stroke}
                  transform={`translate(${p1.x} ${p1.y}) rotate(${angle})`}
                  opacity={isFiring ? draw : 1}
                />
              </g>
            );
          })}
        </svg>

        {/* Event captions, riding their own curves. */}
        {arcs.map((arc, i) => {
          const a = pts[arc.from];
          const b = pts[arc.to];
          if (!a || !b) return null;
          const isFiring = i === firing;
          const wasLit = lit.has(i) && !isFiring;
          const at =
            arc.from === arc.to
              ? {x: a.x, y: a.y - PILL_H / 2 - 108}
              : quad(onRim(a, control(a, b)), control(a, b), onRim(b, control(a, b)), 0.5);
          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: at.x - 110,
                top: at.y - 16,
                width: 220,
                textAlign: 'center',
                fontFamily: theme.fontMono,
                fontSize: isFiring ? 17 : 15,
                lineHeight: 1.2,
                letterSpacing: 0.6,
                color: isFiring ? theme.accentText : theme.textMuted,
                opacity: isFiring ? settle : wasLit ? 0.75 : 0.3,
                transform: `scale(${isFiring ? 0.94 + 0.06 * settle : 1})`,
              }}
            >
              {arc.on}
            </div>
          );
        })}

        {/* The states. */}
        {nodes.map((node, i) => {
          const p = pts[i];
          if (!p) return null;
          const here = i === step.token;
          const isDwell = i === dwell;
          const active = here || isDwell;
          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: p.x - PILL_W / 2,
                top: p.y - PILL_H / 2,
                width: PILL_W,
                height: PILL_H,
                borderRadius: PILL_H / 2,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: active ? withAlpha(theme.accent, 0.14) : withAlpha(theme.surface, 0.85),
                border: `2px solid ${active ? theme.accent : touched.has(i) ? withAlpha(theme.accent, 0.35) : theme.surfaceBorder}`,
                opacity: active ? 1 : touched.has(i) ? 0.82 : 0.5,
                transform: `scale(${isDwell ? 0.97 + 0.03 * settle : 1})`,
              }}
            >
              <span
                style={{
                  fontFamily: theme.fontDisplay,
                  fontSize: 30,
                  fontWeight: 700,
                  letterSpacing: -0.4,
                  color: active ? theme.text : theme.textMuted,
                  paddingInline: 14,
                  textAlign: 'center',
                }}
              >
                {node.label}
              </span>
            </div>
          );
        })}

        {/* The token. The one glow in the frame. */}
        <div
          style={{
            position: 'absolute',
            left: tokenPt.x - 13,
            top: tokenPt.y - 13,
            width: 26,
            height: 26,
            borderRadius: 13,
            background: theme.accentQuantity,
            transform: `scale(${pulse})`,
            boxShadow: `0 0 28px ${withAlpha(theme.accentQuantity, 0.7)}`,
            zIndex: 3,
          }}
        />

        {/* The closer's caption: the route, named. */}
        {route ? (
          <div
            style={{
              position: 'absolute',
              left: 0,
              top: BOARD_H + 8,
              width: BOARD_W,
              textAlign: 'center',
              fontFamily: theme.fontMono,
              fontSize: 16,
              letterSpacing: 3,
              textTransform: 'uppercase',
              color: theme.accentText,
              opacity: settle,
            }}
          >
            {`the route it took: ${lit.size} of ${arcs.length} transitions`}
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
