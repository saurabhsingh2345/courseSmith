import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W, STAGE_H} from './Stage';
import {SceneHeader} from './SceneHeader';

// HistoryScene: the commit graph, drawn as a graph.
//
// The layout is the one git users already have in their heads from `git log
// --graph`, straightened out and given room: lanes are horizontal hairlines
// with their ref names on a left rail, time runs left to right, and every
// commit is a disc sitting on its lane at its own column. Index order is time
// order, so a commit's column is simply its index — no layout pass, no
// heuristics, and the eye can read "later means further right" without being
// told.
//
// Everything the clip claims is a visible fact rather than a caption. A branch
// is two edges leaving one disc. A merge is one disc with two edges arriving
// from two different hairlines. Because those are geometric, they survive being
// paused, screenshotted and argued with — which is the whole reason this
// template exists instead of a bulleted explanation of what a branch is.
//
// Edges draw rather than appear. An edge belongs to its CHILD, so it enters on
// the beat that lands the child, stroked along its own length with a dash
// offset. Lane-changing edges are cubic curves and same-lane edges are straight
// runs, so the shape of the line already says whether the history moved.
//
// HEAD is the only bright chip in the frame and it travels: when a commit
// lands, the chip slides from the previous commit to the new one rather than
// cutting. That movement is what makes HEAD read as a pointer — the thing it
// actually is — instead of a badge painted on the newest disc.

const BODY_W = Math.min(STAGE_W, 1520);
const RAIL_W = 190;
const TRACK_PAD = 70;
const LANE_GAP = 126;
const TOP_PAD = 86;
const LABEL_ROOM = 96;
const DISC_R = 18;

type Commit = {col: number; lane: number; label: string; parents: number[]; children: number[]; merge: boolean};
type Edge = {from: number; to: number; fromLane: number; toLane: number; curved: boolean};
type Step = {
  startMs: number;
  endMs: number;
  show: 'graph' | 'commit' | 'branch' | 'merge' | 'log';
  at?: number;
  kids?: number[];
  landed: number[];
  head: number;
};

export const HistoryScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const lanes = (Array.isArray(props.lanes) ? props.lanes : []) as string[];
  const commits = (Array.isArray(props.commits) ? props.commits : []) as Commit[];
  const edges = (Array.isArray(props.edges) ? props.edges : []) as Edge[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (lanes.length === 0 || commits.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const bodyH = Math.min(STAGE_H - 210, TOP_PAD + (lanes.length - 1) * LANE_GAP + LABEL_ROOM + 40);
  const trackW = BODY_W - RAIL_W;
  const span = Math.max(1, commits.length - 1);
  const stepX = (trackW - TRACK_PAD * 2) / span;
  const x = (col: number) => RAIL_W + TRACK_PAD + col * stepX;
  const y = (lane: number) => TOP_PAD + lane * LANE_GAP;

  const landed = new Set(Array.isArray(step.landed) ? step.landed : []);
  const arriving = step.show === 'commit' || step.show === 'merge' ? (step.at ?? -1) : -1;
  const forking = step.show === 'branch' ? (step.at ?? -1) : -1;
  const forkKids = new Set(step.show === 'branch' && Array.isArray(step.kids) ? step.kids : []);
  const lit = step.show === 'log';

  // The edge that belongs to the commit landing right now strokes itself in;
  // every other visible edge is already whole.
  const drawing = interpolate(sinceStep, [6, 26], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  // HEAD travels rather than cutting, so it reads as a pointer being moved.
  const prevHead = idx > 0 ? steps[idx - 1].head : -1;
  const move = interpolate(sinceStep, [4, 22], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const headNow = step.head;
  const headFrom = prevHead >= 0 && prevHead < commits.length ? commits[prevHead] : undefined;
  const headTo = headNow >= 0 && headNow < commits.length ? commits[headNow] : undefined;
  const headX = headTo
    ? (headFrom ? x(headFrom.col) + (x(headTo.col) - x(headFrom.col)) * move : x(headTo.col))
    : 0;
  const headY = headTo
    ? (headFrom ? y(headFrom.lane) + (y(headTo.lane) - y(headFrom.lane)) * move : y(headTo.lane))
    : 0;

  const edgePath = (e: Edge): string => {
    const x1 = x(commits[e.from].col);
    const y1 = y(e.fromLane);
    const x2 = x(commits[e.to].col);
    const y2 = y(e.toLane);
    if (!e.curved) {
      return `M ${x1} ${y1} L ${x2} ${y2}`;
    }
    const bend = Math.max(30, (x2 - x1) * 0.55);
    return `M ${x1} ${y1} C ${x1 + bend} ${y1}, ${x2 - bend} ${y2}, ${x2} ${y2}`;
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

      <div style={{width: BODY_W, height: bodyH, position: 'relative'}}>
        {/* The lanes: hairlines with their ref names on the rail. */}
        {lanes.map((name, l) => (
          <div key={l}>
            <div
              style={{
                position: 'absolute',
                left: RAIL_W,
                top: y(l) - 1,
                width: trackW,
                height: 2,
                background: withAlpha(theme.line, lit ? 0.45 : 0.25),
              }}
            />
            <div
              style={{
                position: 'absolute',
                left: 0,
                top: y(l) - 15,
                width: RAIL_W - 26,
                textAlign: 'right',
                fontFamily: theme.fontMono,
                fontSize: 22,
                letterSpacing: 0.4,
                color: lit ? theme.accentText : theme.textMuted,
              }}
            >
              {name}
            </div>
          </div>
        ))}

        {/* The edges. An edge belongs to its child and enters when the child does. */}
        <svg
          width={BODY_W}
          height={bodyH}
          viewBox={`0 0 ${BODY_W} ${bodyH}`}
          style={{position: 'absolute', left: 0, top: 0, overflow: 'visible'}}
        >
          {edges.map((e, i) => {
            if (!landed.has(e.to)) return null;
            const isNew = e.to === arriving;
            const highlighted = forking >= 0 && e.from === forking && forkKids.has(e.to);
            const stroke = highlighted ? theme.accent : withAlpha(theme.line, lit ? 0.85 : 0.6);
            return (
              <path
                key={i}
                d={edgePath(e)}
                fill="none"
                stroke={stroke}
                strokeWidth={highlighted ? 4 : 2.5}
                strokeLinecap="round"
                pathLength={1}
                strokeDasharray={1}
                strokeDashoffset={isNew ? 1 - drawing : 0}
              />
            );
          })}
        </svg>

        {/* The commits. Discs on their lane, message underneath. */}
        {commits.map((c, i) => {
          if (!landed.has(i)) return null;
          const pop =
            i === arriving
              ? spring({frame: sinceStep - 8, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 24})
              : 1;
          const highlighted = i === forking;
          const ring = c.merge ? theme.accentRival : highlighted ? theme.accent : theme.accentQuantity;
          return (
            <div key={i}>
              <div
                style={{
                  position: 'absolute',
                  left: x(c.col) - DISC_R,
                  top: y(c.lane) - DISC_R,
                  width: DISC_R * 2,
                  height: DISC_R * 2,
                  borderRadius: DISC_R,
                  background: theme.surface,
                  border: `4px solid ${ring}`,
                  transform: `scale(${0.5 + 0.5 * pop})`,
                  opacity: pop,
                  // The one glow: the commit arriving right now.
                  boxShadow: i === arriving && pop < 0.999 ? `0 0 30px ${withAlpha(ring, 0.55)}` : undefined,
                }}
              />
              <div
                style={{
                  position: 'absolute',
                  left: x(c.col) - 84,
                  top: y(c.lane) + DISC_R + 14,
                  width: 168,
                  textAlign: 'center',
                  fontFamily: theme.fontBody,
                  fontSize: 19,
                  lineHeight: 1.25,
                  color: highlighted || lit ? theme.text : theme.textMuted,
                  opacity: interpolate(pop, [0.3, 1], [0, 1], {
                    extrapolateLeft: 'clamp',
                    extrapolateRight: 'clamp',
                  }),
                }}
              >
                {c.label}
              </div>
            </div>
          );
        })}

        {/* HEAD: the pointer, and the only bright chip in the frame. */}
        {headTo ? (
          <div
            style={{
              position: 'absolute',
              left: headX - 42,
              top: headY - DISC_R - 52,
              width: 84,
              height: 34,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              borderRadius: 17,
              background: theme.accent,
              color: theme.ink,
              fontFamily: theme.fontMono,
              fontSize: 16,
              fontWeight: 700,
              letterSpacing: 1.6,
            }}
          >
            HEAD
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
