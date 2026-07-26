import {useEffect, useMemo, useState} from 'react';
import {continueRender, delayRender, staticFile, useCurrentFrame, useVideoConfig} from 'remotion';
import {stratify, tree, type HierarchyPointNode} from 'd3-hierarchy';
import {linkVertical, linkRadial} from 'd3-shape';
import {D3DiagramSpec, D3Node,  assetPath} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {MotionTokens, bezierEasing, resolveMotion, secondsToFrames} from '../theme/motion';
import {SceneHeader} from './SceneHeader';
import {Stage} from './Stage';

// D3Diagram lays out a structured node-link graph and animates it building up:
// nodes pop in (scale + fade) in reveal order, each edge draws on right after
// its child node appears. Layout is computed once (useMemo, not per frame) so it
// is deterministic across the render; only the reveal is frame-driven. Timing
// comes from the course motion tokens.

const VIEW_W = 1000;
const VIEW_H = 620;
const NODE_R = 30;

type Placed = {node: D3Node; x: number; y: number; depth: number; order: number};
type PlacedEdge = {from: string; to: string; label?: string; d: string; order: number};

// A small brand-aware palette for node groups.
const groupColor = (theme: ResolvedTheme, group: number): string => {
  const palette = [theme.primary, theme.accent, '#22a06b', '#a05ec6'];
  return palette[group % palette.length];
};

/** Build parent lookup from edges (child → parent) for hierarchy layouts. */
const parentOf = (spec: D3DiagramSpec): Map<string, string | undefined> => {
  const p = new Map<string, string | undefined>();
  spec.nodes.forEach((n) => p.set(n.id, undefined));
  spec.edges.forEach((e) => p.set(e.to, e.from));
  return p;
};

/** Compute node positions + edge paths + reveal order for the chosen layout. */
const layoutSpec = (spec: D3DiagramSpec): {nodes: Placed[]; edges: PlacedEdge[]} => {
  const byId = new Map(spec.nodes.map((n) => [n.id, n]));

  if (spec.layout === 'tree' || spec.layout === 'radial') {
    const parent = parentOf(spec);
    let root;
    try {
      root = stratify<D3Node>()
        .id((d) => d.id)
        .parentId((d) => parent.get(d.id))(spec.nodes);
    } catch {
      return forceLayout(spec, byId);
    }

    if (spec.layout === 'radial') {
      const radius = Math.min(VIEW_W, VIEW_H) / 2 - NODE_R - 40;
      const laid = tree<D3Node>().size([2 * Math.PI, radius]).separation((a, b) => (a.parent === b.parent ? 1 : 1.6) / a.depth || 1)(root);
      const cx = VIEW_W / 2;
      const cy = VIEW_H / 2;
      const toXY = (n: HierarchyPointNode<D3Node>) => {
        const angle = n.x - Math.PI / 2;
        return {x: cx + n.y * Math.cos(angle), y: cy + n.y * Math.sin(angle)};
      };
      const nodes = laid.descendants().map((n, i) => ({node: n.data, ...toXY(n), depth: n.depth, order: i}));
      const orderOf = new Map(nodes.map((n) => [n.node.id, n.order]));
      const link = linkRadial<unknown, HierarchyPointNode<D3Node>>()
        .angle((n) => n.x)
        .radius((n) => n.y);
      const edges = laid.links().map((l) => {
        const raw = link({source: l.source, target: l.target} as never) ?? '';
        return {from: l.source.data.id, to: l.target.data.id, d: shiftPath(raw, cx, cy), order: orderOf.get(l.target.data.id) ?? 0};
      });
      return {nodes, edges};
    }

    const laid = tree<D3Node>().size([VIEW_W - 2 * NODE_R - 40, VIEW_H - 2 * NODE_R - 80])(root);
    const dx = NODE_R + 20;
    const dy = NODE_R + 50;
    const nodes = laid.descendants().map((n, i) => ({node: n.data, x: n.x + dx, y: n.y + dy, depth: n.depth, order: i}));
    const orderOf = new Map(nodes.map((n) => [n.node.id, n.order]));
    const link = linkVertical<unknown, {x: number; y: number}>()
      .x((n) => n.x)
      .y((n) => n.y);
    const edges = laid.links().map((l) => ({
      from: l.source.data.id,
      to: l.target.data.id,
      d: link({source: {x: l.source.x + dx, y: l.source.y + dy}, target: {x: l.target.x + dx, y: l.target.y + dy}} as never) ?? '',
      order: orderOf.get(l.target.data.id) ?? 0,
    }));
    return {nodes, edges};
  }

  return forceLayout(spec, byId);
};

/** Deterministic ring layout for general graphs (no simulation randomness). */
const forceLayout = (spec: D3DiagramSpec, byId: Map<string, D3Node>): {nodes: Placed[]; edges: PlacedEdge[]} => {
  const cx = VIEW_W / 2;
  const cy = VIEW_H / 2;
  const radius = Math.min(VIEW_W, VIEW_H) / 2 - NODE_R - 30;
  const n = spec.nodes.length;
  const pos = new Map<string, {x: number; y: number}>();
  const nodes: Placed[] = spec.nodes.map((node, i) => {
    // Single node centred; otherwise evenly spaced on a circle.
    const angle = (i / Math.max(1, n)) * 2 * Math.PI - Math.PI / 2;
    const x = n === 1 ? cx : cx + radius * Math.cos(angle);
    const y = n === 1 ? cy : cy + radius * Math.sin(angle);
    pos.set(node.id, {x, y});
    return {node, x, y, depth: 0, order: i};
  });
  const orderOf = new Map(nodes.map((p) => [p.node.id, p.order]));
  const edges: PlacedEdge[] = spec.edges.map((e) => {
    const a = pos.get(e.from)!;
    const b = pos.get(e.to)!;
    return {
      from: e.from,
      to: e.to,
      label: e.label,
      d: `M${a.x},${a.y} L${b.x},${b.y}`,
      order: Math.max(orderOf.get(e.from) ?? 0, orderOf.get(e.to) ?? 0),
    };
  });
  void byId;
  return {nodes, edges};
};

/** d3.linkRadial emits paths around (0,0); translate them to the centre. */
const shiftPath = (d: string, cx: number, cy: number): string =>
  d.replace(/(-?\d+\.?\d*),(-?\d+\.?\d*)/g, (_m, x, y) => `${Number(x) + cx},${Number(y) + cy}`);

export const D3Diagram: React.FC<{
  theme: ResolvedTheme;
  motion?: MotionTokens;
  assetBase?: string;
  props: Record<string, unknown>;
}> = ({theme, motion, assetBase, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const m = resolveMotion(motion);
  const src = String(props.src ?? '');
  const title = String(props.title ?? '');

  const [spec, setSpec] = useState<D3DiagramSpec | null | undefined>(undefined);
  const [handle] = useState(() => delayRender('load-d3-spec'));

  useEffect(() => {
    let cancelled = false;
    fetch(staticFile(assetPath(assetBase, src)))
      .then((res) => (res.ok ? res.json() : Promise.reject(new Error(`HTTP ${res.status}`))))
      .then((json) => {
        if (!cancelled) {
          setSpec(json as D3DiagramSpec);
          continueRender(handle);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setSpec(null);
          continueRender(handle);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [assetBase, src, handle]);

  const laid = useMemo(() => (spec ? layoutSpec(spec) : null), [spec]);

  if (spec === undefined) return null;
  if (spec === null || !laid) {
    return (
      <Stage>
        <div style={{color: theme.text, fontFamily: theme.fontBody, fontSize: 40}}>Diagram unavailable: {src}</div>
      </Stage>
    );
  }

  const nodeStagger = secondsToFrames(fps, m.stagger.items);
  const edgeStagger = secondsToFrames(fps, m.stagger.connections);
  const revealDur = secondsToFrames(fps, m.timing.normal);
  const ease = bezierEasing(m.easing.entrance);

  const heading = title || spec.title || '';

  return (
    <Stage>
      <SceneHeader theme={theme} title={heading} size="compact" marginBottom={26} />
      <svg viewBox={`0 0 ${VIEW_W} ${VIEW_H}`} style={{width: '100%', maxWidth: 1600, flex: 1, minHeight: 0}}>
        <defs>
          <marker id="d3-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
            <path d="M0,0 L10,5 L0,10 z" fill={theme.textMuted} opacity={0.8} />
          </marker>
        </defs>

        {laid.edges.map((e, i) => {
          const start = e.order * edgeStagger + nodeStagger; // draw just after the child node
          const p = clamp01((frame - start) / revealDur);
          const drawn = ease(p);
          return (
            <path
              key={`e${i}`}
              d={e.d}
              fill="none"
              stroke={theme.textMuted}
              strokeWidth={3}
              // Fade the whole edge (stroke + arrowhead) in with the draw so the
              // marker never floats ahead of an undrawn line.
              opacity={0.5 * drawn}
              markerEnd={drawn > 0.6 ? 'url(#d3-arrow)' : undefined}
              pathLength={1}
              strokeDasharray={1}
              strokeDashoffset={1 - drawn}
            />
          );
        })}

        {laid.nodes.map((pn) => {
          const start = pn.order * nodeStagger;
          const p = clamp01((frame - start) / revealDur);
          const e = ease(p);
          const color = groupColor(theme, pn.node.group ?? 0);
          return (
            <g key={pn.node.id} transform={`translate(${pn.x},${pn.y})`} opacity={e}>
              <g transform={`scale(${0.6 + 0.4 * e})`}>
                <circle r={NODE_R} fill="#182236" stroke={color} strokeWidth={4} />
                <circle r={NODE_R} fill={color} opacity={0.2} />
                <text
                  textAnchor="middle"
                  dominantBaseline="central"
                  fontSize={pn.node.label.length > 8 ? 15 : 18}
                  fontFamily="Inter, sans-serif"
                  fontWeight={600}
                  fill="#f2f5fa"
                >
                  {pn.node.label}
                </text>
              </g>
            </g>
          );
        })}
      </svg>
    </Stage>
  );
};

const clamp01 = (x: number): number => Math.max(0, Math.min(1, x));
