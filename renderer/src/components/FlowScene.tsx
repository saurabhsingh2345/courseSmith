import {useMemo} from 'react';
import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_H, STAGE_W} from './Stage';
import {FIGURE_BOX, figureFor, type FigurePalette} from './artwork';
import {curveHead, flowCurve, flowLayout, nodeShape, type Curve, type Rect} from './flow';

// FlowScene is a layered systems diagram that keeps moving.
//
// Three things happen on top of "draw a graph", and each is the reason this is a
// video rather than a screenshot:
//
//  1. Nodes and edges arrive on the beat that introduces them, so the system is
//     assembled in the order the narrator explains it.
//  2. Tokens travel the edges continuously. A static diagram can only imply
//     direction with an arrowhead; moving traffic states it.
//  3. Beats declare a focus. Everything outside it dims and its traffic stops,
//     so one diagram narrates the happy path, then the slow path, then the
//     failure, without ever being redrawn.
//
// Layout is a pure function of the ranks Go assigned (flow.ts) — nothing here
// decides graph shape, and nothing here reads the clock except the animation.

const HEADER_H = 116;
const BOARD_W = STAGE_W;
const BOARD_H = STAGE_H - HEADER_H;

const ARRIVE = {
  /** The edge draws in first, then the box it feeds lands. */
  edgeFrames: 12,
  nodeDelay: 7,
  nodeFrames: 14,
} as const;

const TOKEN = {
  /** Frames for one token to traverse an edge. */
  periodFrames: 52,
  radius: 6,
  /** Tokens start only once the graph has had a moment to settle. */
  startDelay: 10,
  /** Pixels of edge per token in flight. */
  spacing: 190,
  maxPerEdge: 3,
} as const;

/**
 * Tokens in flight on an edge, from its length. A fixed count put two dots
 * within 40px of each other on a short hop, which read as a dotted line rather
 * than as traffic.
 */
const tokensOn = (length: number): number =>
  Math.max(1, Math.min(TOKEN.maxPerEdge, Math.round(length / TOKEN.spacing)));

const FOCUS = {
  /** How sharply the dim/undim transition runs. */
  frames: 10,
  // Dim enough to push the rest of the system back, not so far that it stops
  // being readable. At 0.22 an unfocused node was a ghost, which defeats the
  // point of drawing the whole system: the context is what makes the focused
  // path mean anything.
  dimNode: 0.46,
  dimEdge: 0.3,
} as const;

type FlowNodeProps = {
  id: string;
  label: string;
  kind: string;
  icon: string;
  rank: number;
  order: number;
  atMs: number;
};
type FlowEdgeProps = {from: number; to: number; atMs: number};
type FocusWindow = {startMs: number; endMs: number; nodes: number[]};

/** Kind → accent role. Kind never changes a node's shape: uniform boxes are
 *  what let the ranking own the layout completely. */
const kindColor = (theme: ResolvedTheme, kind: string): string => {
  switch (kind) {
    case 'store':
      return '#5eb0ef';
    case 'cache':
      return theme.accent;
    case 'queue':
      return '#a78bfa';
    case 'external':
      return '#94a3b8';
    case 'decision':
      return '#f472b6';
    case 'client':
      return '#4ade80';
    default:
      return theme.primary;
  }
};

export const FlowScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const nowMs = sceneStartMs + (frame / FPS) * 1000;

  const title = String(props.title ?? '');
  const nodes = (Array.isArray(props.nodes) ? props.nodes : []) as FlowNodeProps[];
  const edges = (Array.isArray(props.edges) ? props.edges : []) as FlowEdgeProps[];
  const focusWindows = (Array.isArray(props.focus) ? props.focus : []) as FocusWindow[];
  const ranks = Number(props.ranks ?? 1);

  const geometry = useMemo(() => {
    const rects = flowLayout(nodes, Math.max(1, ranks), BOARD_W, BOARD_H);
    const curves = edges.map((e) =>
      rects[e.from] && rects[e.to] ? flowCurve(rects[e.from], rects[e.to]) : null,
    );
    return {rects, curves};
  }, [nodes, edges, ranks]);

  if (nodes.length === 0) {
    return null;
  }

  // The active focus set, if any beat's window covers this moment. Nodes
  // outside it dim; so do edges with an endpoint outside it.
  const active = focusWindows.find((w) => nowMs >= w.startMs && nowMs < w.endMs);
  const focusSet = active ? new Set(active.nodes) : null;
  // Ease the dim so a focus change is a shift of attention, not a cut.
  const focusP = active
    ? interpolate(nowMs - active.startMs, [0, (FOCUS.frames / FPS) * 1000], [0, 1], {
        extrapolateLeft: 'clamp',
        extrapolateRight: 'clamp',
      })
    : 0;

  const nodeDim = (i: number): number => {
    if (!focusSet) return 1;
    const lit = focusSet.has(i);
    return lit ? 1 : 1 - (1 - FOCUS.dimNode) * focusP;
  };
  const edgeLit = (e: FlowEdgeProps): boolean =>
    !focusSet || (focusSet.has(e.from) && focusSet.has(e.to));
  const edgeDim = (e: FlowEdgeProps): number =>
    edgeLit(e) ? 1 : 1 - (1 - FOCUS.dimEdge) * focusP;

  const labelSize = Math.max(22, Math.min(32, geometry.rects[0].w / 11));
  // The figure gets real room. At the 30px the flat glyph used, an animated
  // figure is mush — its parts are drawn against a 200-unit box, so a third of
  // that leaves a 2px bar as half a pixel. At ~0.55 of the node's height every
  // mechanism still reads, and the node is wide enough to afford it.
  const figureSize = Math.round(Math.min(geometry.rects[0].h * 0.72, 132));

  const figurePalette: FigurePalette = {
    accent: theme.accent,
    primary: theme.primary,
    ink: theme.ink,
    soft: theme.mass,
    line: theme.line,
  };

  return (
    <Stage justify="flex-start">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={30} />
      <svg
        width={BOARD_W}
        height={BOARD_H}
        viewBox={`0 0 ${BOARD_W} ${BOARD_H}`}
        style={{overflow: 'visible'}}
      >
        <defs>
          {/* A single soft shadow gives the boxes somewhere to sit. Without it
              every node is the same flat fill and the only thing separating a
              lit node from a dim one is its border colour. */}
          <filter id="flow-elev" x="-25%" y="-25%" width="150%" height="150%">
            <feDropShadow
              dx="0"
              dy="10"
              stdDeviation="14"
              floodColor={theme.ink}
              floodOpacity={theme.mode === 'light' ? 0.18 : 0.55}
            />
          </filter>
          {/* Top-edge light, so a box has a lit face rather than a flat one.
              On paper the card is already lighter than the stage, so the same
              white wash flattens it — the light mode leans on the shadow for
              elevation and only shades the bottom edge. */}
          <linearGradient id="flow-sheen" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#ffffff" stopOpacity={theme.mode === 'light' ? 0 : 0.07} />
            <stop offset="55%" stopColor="#ffffff" stopOpacity={theme.mode === 'light' ? 0 : 0.015} />
            <stop offset="100%" stopColor={theme.ink} stopOpacity={theme.mode === 'light' ? 0.04 : 0.06} />
          </linearGradient>
        </defs>

        {/* Edges under the boxes, so a curve entering a box is tucked behind it. */}
        {edges.map((e, i) => {
          const curve = geometry.curves[i];
          if (!curve) return null;
          const since = frame - Math.round(((e.atMs - sceneStartMs) / 1000) * FPS);
          if (since < 0) return null;
          const p = interpolate(since, [0, ARRIVE.edgeFrames], [0, 1], {
            extrapolateLeft: 'clamp',
            extrapolateRight: 'clamp',
          });
          const lit = edgeLit(e);
          const dim = edgeDim(e);
          return (
            <g key={`edge-${i}`} opacity={dim}>
              <path
                d={curve.d}
                fill="none"
                stroke={lit && focusSet ? theme.accent : theme.textMuted}
                strokeWidth={lit && focusSet ? 2.6 : 2}
                strokeLinecap="round"
                strokeDasharray={curve.length}
                strokeDashoffset={curve.length * (1 - p)}
                opacity={0.82}
              />
              {p >= 1 && (
                <path
                  d={curveHead(curve)}
                  fill="none"
                  stroke={lit && focusSet ? theme.accent : theme.textMuted}
                  strokeWidth={2.4}
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  opacity={0.75}
                />
              )}
              {/* Traffic. Only on lit edges: stopping it everywhere else is
                  what makes a focused path read as the one in play. */}
              {p >= 1 &&
                lit &&
                since > ARRIVE.edgeFrames + TOKEN.startDelay &&
                Array.from({length: tokensOn(curve.length)}, (_, k) => {
                  const phase = k / tokensOn(curve.length);
                  const t = ((since / TOKEN.periodFrames + phase) % 1 + 1) % 1;
                  const pt = curve.at(t);
                  // Fade in and out at the ends so a token does not pop into
                  // existence on the edge of a box.
                  const edgeFade = Math.min(1, Math.sin(t * Math.PI) * 2.6);
                  return (
                    <circle
                      key={k}
                      cx={pt.x}
                      cy={pt.y}
                      r={TOKEN.radius}
                      fill={theme.accent}
                      opacity={0.85 * edgeFade}
                    />
                  );
                })}
            </g>
          );
        })}

        {/* Nodes */}
        {nodes.map((node, i) => {
          const since = frame - Math.round(((node.atMs - sceneStartMs) / 1000) * FPS);
          if (since < 0) return null;
          const rect: Rect = geometry.rects[i];
          const enter = spring({
            frame: since - ARRIVE.nodeDelay,
            fps,
            config: {damping: 200, mass: 0.6},
            durationInFrames: ARRIVE.nodeFrames,
          });
          const accent = kindColor(theme, node.kind);
          const lit = !focusSet || focusSet.has(i);
          const shape = nodeShape(node.kind, rect.w, rect.h);
          const Figure = figureFor(node.icon);
          // The figure's own palette leans on the node's kind colour rather than
          // the theme primary, so a store's cylinder and the figure inside it
          // are the same blue and the node reads as one object.
          const nodePalette: FigurePalette = {...figurePalette, primary: accent};

          return (
            <g
              key={`node-${i}`}
              opacity={enter * nodeDim(i)}
              transform={`translate(${rect.x} ${rect.y}) scale(${0.92 + enter * 0.08})`}
              style={{transformOrigin: `${rect.w / 2}px ${rect.h / 2}px`}}
            >
              <path
                d={shape.body}
                fill={theme.surface}
                stroke={lit && focusSet ? accent : theme.surfaceBorder}
                // A dim box needs a *heavier* border than a lit one to survive
                // the group opacity; at 1.5px it faded out from under its own
                // kind stripe and left the stripe floating.
                strokeWidth={lit && focusSet ? 2.4 : 2.2}
                // External systems are not ours; a dashed outline says so
                // without needing a legend.
                strokeDasharray={shape.dashed ? '7 6' : undefined}
                filter="url(#flow-elev)"
              />
              <path d={shape.body} fill="url(#flow-sheen)" />
              {/* Whatever the silhouette needs on top of its outline — a
                  cylinder's far cap, a queue's slots, a window's title bar. */}
              {shape.decor?.map((d, di) => (
                <path
                  key={di}
                  d={d}
                  fill="none"
                  stroke={accent}
                  strokeWidth={2}
                  opacity={lit ? 0.55 : 0.32}
                />
              ))}
              {/* The figure. It is the same animated artwork the illustration
                  and story templates use, not a flat glyph — a queue whose items
                  actually drain says more about what a queue is than any label
                  beside it, and it costs nothing because the drawing already
                  existed. */}
              <g transform={`translate(${shape.padLeft} ${rect.h / 2 - figureSize / 2})`}>
                <svg
                  width={figureSize}
                  height={figureSize}
                  viewBox={`0 0 ${FIGURE_BOX} ${FIGURE_BOX}`}
                  style={{overflow: 'visible'}}
                >
                  <Figure build={enter} t={frame / FPS} palette={nodePalette} />
                </svg>
              </g>
              <text
                x={shape.padLeft + figureSize + 18}
                y={rect.h / 2}
                dominantBaseline="central"
                fontFamily={theme.fontDisplay}
                fontSize={labelSize}
                fontWeight={600}
                fill={theme.text}
              >
                {node.label}
              </text>
            </g>
          );
        })}
      </svg>
    </Stage>
  );
};
