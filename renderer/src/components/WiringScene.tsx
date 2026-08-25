import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, seat, withAlpha} from '../theme/theme';
import {STAGE_W, Stage} from './Stage';

// WiringScene draws a path: named blocks in a row, one hop lit at a time.
//
// THE ROW IS HORIZONTAL EVEN WHEN THE PATH IS A LOOP. A ring is the prettier
// drawing and it costs the one thing this shape is good at: left to right is
// time, so a viewer reads the ORDER before they read a single label. On a ring
// there is no first block, and every loop diagram then needs an arrow explaining
// where to start. The return is an arc underneath instead — which also keeps the
// claim honest, because a loop that comes back changed is not a circle.
//
// LIT HOPS LATCH. The path builds up rather than a spark travelling along it, so
// at any moment the frame shows how far through the mechanism the narration has
// got. Same reasoning as the waypoint's spine: progress the viewer can see
// without being told.
//
// THE UNLIT PATH STAYS VISIBLE. At about a fifth ink, so the shape is legible
// from the first frame and a lit hop means something against it. Hiding the
// remainder would make each hop a surprise, and the surprise is not the lesson —
// the mechanism is.

type Node = {label: string; kind?: 'in' | 'work' | 'store' | 'out'; note?: string};

type Step = {
  startMs: number;
  endMs: number;
  show: 'shape' | 'hop' | 'round' | 'path';
  at?: number;
  walked?: number;
  round?: boolean;
};

type Props = {
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: {
    nodes?: Node[];
    hops?: string[];
    return?: string;
    steps?: Step[];
  };
};

const BLOCK_H = 156;
const ROW_Y = 300;

export const WiringScene = ({theme, sceneStartMs, props}: Props) => {
  const frame = useCurrentFrame();
  const ms = (frame / FPS) * 1000 + sceneStartMs;
  const nodes = props.nodes ?? [];
  const hops = props.hops ?? [];
  const steps = props.steps ?? [];

  const idx = Math.max(
    0,
    steps.findIndex((s) => ms >= s.startMs && ms < s.endMs),
  );
  const step = steps[idx] ?? steps[steps.length - 1];
  const sceneStart = steps[0]?.startMs ?? sceneStartMs;
  const walked = step?.walked ?? 0;
  const roundOn = Boolean(step?.round);

  const arrive = (from: number, dur = 420) =>
    interpolate(ms, [from, from + dur], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});

  // Geometry: fit the row to the stage rather than to a fixed block width, so
  // three blocks are generous and five still fit at a readable size.
  const n = Math.max(nodes.length, 1);
  const gap = n <= 3 ? 150 : 104;
  const blockW = Math.min(320, Math.floor((STAGE_W - gap * (n - 1)) / n));
  const rowW = blockW * n + gap * (n - 1);
  const left = Math.round((STAGE_W - rowW) / 2);

  const tint = (kind: string) => {
    switch (kind) {
      case 'in':
        return theme.accentQuantity;
      case 'store':
        return theme.accentRival;
      case 'out':
        return theme.accent;
      default:
        return theme.mass;
    }
  };

  return (
    <Stage>
      <div style={{position: 'relative', width: STAGE_W, height: 680}}>
        {nodes.map((node, i) => {
          // A block is lit once the hop INTO it has been walked; the first block
          // is lit from the moment the shape is drawn, because the work starts
          // there and a dark origin reads as a step that has not happened.
          const on = i === 0 ? true : walked >= i;
          const x = left + i * (blockW + gap);
          const litAt = i === 0 ? sceneStart : (steps.find((s) => s.show === 'hop' && s.at === i - 1)?.startMs ?? sceneStart);
          const o = on ? 0.28 + 0.72 * arrive(litAt, 380) : 0.28;
          return (
            <div key={i} style={{position: 'absolute', left: x, top: ROW_Y, width: blockW}}>
              <div
                style={{
                  height: BLOCK_H,
                  borderRadius: 14,
                  background: theme.surface,
                  border: `1px solid ${on ? withAlpha(tint(node.kind ?? 'work'), 0.5) : theme.surfaceBorder}`,
                  boxShadow: on ? seat(theme, 'lifted') : seat(theme, 'resting'),
                  opacity: o,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  padding: '0 22px',
                  textAlign: 'center',
                }}
              >
                <span
                  style={{
                    fontFamily: theme.fontBody,
                    fontSize: 31,
                    fontWeight: 700,
                    lineHeight: 1.2,
                    color: theme.text,
                  }}
                >
                  {node.label}
                </span>
              </div>
              {node.note ? (
                <div
                  style={{
                    fontFamily: theme.fontBody,
                    fontSize: 21,
                    lineHeight: 1.35,
                    color: theme.textMuted,
                    marginTop: 18,
                    textAlign: 'center',
                    opacity: o,
                  }}
                >
                  {node.note}
                </div>
              ) : null}
            </div>
          );
        })}

        {/* The hops. */}
        {hops.map((hop, i) => {
          const on = walked > i;
          const x = left + i * (blockW + gap) + blockW;
          const litAt = steps.find((s) => s.show === 'hop' && s.at === i)?.startMs ?? sceneStart;
          const a = on ? arrive(litAt, 420) : 0;
          const ink = on ? theme.accent : withAlpha(theme.text, 0.2);
          return (
            <div key={i} style={{position: 'absolute', left: x, top: ROW_Y, width: gap, height: BLOCK_H}}>
              {/* the wire */}
              <div
                style={{
                  position: 'absolute',
                  top: BLOCK_H / 2 - 1,
                  left: 6,
                  width: gap - 22,
                  height: 2,
                  background: ink,
                }}
              />
              {/* the head, which travels in as the hop lights */}
              <div
                style={{
                  position: 'absolute',
                  top: BLOCK_H / 2 - 6,
                  left: 6 + Math.round((gap - 34) * (on ? a : 0.6)),
                  width: 0,
                  height: 0,
                  borderTop: '6px solid transparent',
                  borderBottom: '6px solid transparent',
                  borderLeft: `10px solid ${ink}`,
                }}
              />
              <div
                style={{
                  position: 'absolute',
                  top: BLOCK_H / 2 - 46,
                  left: 0,
                  width: gap,
                  textAlign: 'center',
                  fontFamily: theme.fontMono,
                  fontSize: 18,
                  lineHeight: 1.25,
                  color: on ? theme.accentText : withAlpha(theme.text, 0.28),
                }}
              >
                {hop}
              </div>
            </div>
          );
        })}

        {/* The return: an arc under the row, drawn only when the path loops. */}
        {props.return ? (
          <svg
            width={STAGE_W}
            height={220}
            style={{position: 'absolute', left: 0, top: ROW_Y + BLOCK_H + 66, overflow: 'visible'}}
          >
            <path
              d={`M ${left + rowW - blockW / 2} 10 C ${left + rowW - blockW / 2} 130, ${left + blockW / 2} 130, ${left + blockW / 2} 10`}
              fill="none"
              stroke={roundOn ? theme.accent : withAlpha(theme.text, 0.18)}
              strokeWidth={2}
              strokeDasharray={roundOn ? 'none' : '7 7'}
            />
            <polygon
              points={`${left + blockW / 2 - 6},18 ${left + blockW / 2 + 6},18 ${left + blockW / 2},4`}
              fill={roundOn ? theme.accent : withAlpha(theme.text, 0.18)}
            />
            <text
              x={left + rowW / 2}
              y={126}
              textAnchor="middle"
              style={{
                fontFamily: theme.fontMono,
                fontSize: 19,
                fill: roundOn ? theme.accentText : withAlpha(theme.text, 0.26),
              }}
            >
              {props.return}
            </text>
          </svg>
        ) : null}
      </div>
    </Stage>
  );
};
