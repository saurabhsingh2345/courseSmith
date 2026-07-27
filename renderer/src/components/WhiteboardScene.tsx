import {useMemo} from 'react';
import {interpolate, random, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_H, STAGE_W} from './Stage';
import {iconFor} from './icons';
import {
  boardLayout,
  edgeAnchor,
  penAt,
  roughArrow,
  roughRect,
  type BoxRect,
  type Stroke,
} from './sketch';

// WhiteboardScene is one board that fills in as the narrator talks.
//
// It is deliberately one scene for the whole clip rather than a scene per beat.
// The accumulation *is* the form: each box stays, arrows reach from what is
// already drawn to what is arriving, and the last beat lands with the finished
// picture on screen. Cutting between drawings would make it a slideshow.
//
// Nothing here is an imported drawing. The strokes are generated (sketch.ts) so
// they can be drawn on by stroke-dashoffset, so the marker can sit exactly at
// the pen, and so the layout is ours — an LLM authoring freehand SVG produced
// crooked, overlapping boards often enough to need a vision-QA gate.

// Header height reserve, so the board centres in what is left of the stage.
const HEADER_H = 116;
const BOARD_W = STAGE_W;
const BOARD_H = STAGE_H - HEADER_H;

// Per-item choreography, in frames from the item's cue.
const DRAW = {
  /** The connector reaches across first… */
  arrowFrames: 13,
  /** …then the box is sketched. */
  boxStart: 9,
  boxFrames: 17,
  /** The second pass of the outline trails the first. */
  secondPassDelay: 5,
  /** The icon wipes in once the box is most of the way round. */
  iconStart: 20,
  iconFrames: 9,
  /** The label is written after the box closes. */
  labelStart: 26,
  labelWordFrames: 5,
  /** How long the newest item keeps the accent before settling back. */
  settleFrames: 34,
} as const;

type SketchItemProps = {label: string; icon: string; atMs: number; from?: number};

/** A stroke drawn on by dashoffset, with its progress supplied. */
const DrawnPath: React.FC<{
  stroke: Stroke;
  progress: number;
  color: string;
  width: number;
  opacity?: number;
}> = ({stroke, progress, color, width, opacity = 1}) => {
  if (progress <= 0) {
    return null;
  }
  return (
    <path
      d={stroke.d}
      fill="none"
      stroke={color}
      strokeWidth={width}
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeDasharray={stroke.length}
      strokeDashoffset={stroke.length * (1 - Math.min(1, progress))}
      opacity={opacity}
    />
  );
};

export const WhiteboardScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const title = String(props.title ?? '');
  const items = (Array.isArray(props.items) ? props.items : []) as SketchItemProps[];

  // Geometry is content-derived and frame-independent, so it is built once.
  const board = useMemo(() => {
    const boxes = boardLayout(items.length, BOARD_W, BOARD_H);
    // Two passes per box, each with its own seed. This is the whole reason the
    // boxes read as drawn rather than as CSS: a single clean outline, however
    // much you wobble it, still looks like a rect element. A second pass
    // trailing the first is how a pen actually behaves.
    const rects = boxes.map((b, i) => {
      const key = `wb-box-${i}-${items[i]?.label ?? ''}`;
      return {
        rect: b,
        stroke: roughRect(b.w, b.h, 26, `${key}-a`),
        second: roughRect(b.w, b.h, 26, `${key}-b`, 6.5),
        // A degree of seeded tilt. Every box sitting exactly square on a shared
        // baseline is what made a drawn board read as a table; this is small
        // enough that nothing looks crooked and large enough that the grid
        // stops announcing itself.
        tilt: (random(`${key}-tilt`) - 0.5) * 1.5,
      };
    });
    const links = items.map((item, i) => {
      if (item.from === undefined || item.from < 0 || item.from >= boxes.length || item.from === i) {
        return null;
      }
      const a = boxes[item.from] as BoxRect;
      const b = boxes[i] as BoxRect;
      return roughArrow(edgeAnchor(a, b), edgeAnchor(b, a), `wb-link-${item.from}-${i}`);
    });
    return {rects, links};
  }, [items]);

  if (items.length === 0) {
    return null;
  }

  const fontSize = Math.max(26, Math.min(40, board.rects[0].rect.w / 10));
  const iconSize = Math.round(Math.min(58, board.rects[0].rect.h * 0.34));

  return (
    <Stage justify="flex-start">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={30} />
      <svg
        width={BOARD_W}
        height={BOARD_H}
        viewBox={`0 0 ${BOARD_W} ${BOARD_H}`}
        style={{overflow: 'visible'}}
      >
        {items.map((item, i) => {
          const cue = Math.round(((item.atMs - sceneStartMs) / 1000) * FPS);
          const since = frame - cue;
          if (since < 0) {
            return null;
          }
          const {rect, stroke, second, tilt} = board.rects[i];
          const link = board.links[i];

          const arrowP = interpolate(since, [0, DRAW.arrowFrames], [0, 1], {
            extrapolateLeft: 'clamp',
            extrapolateRight: 'clamp',
          });
          const boxP = interpolate(since, [DRAW.boxStart, DRAW.boxStart + DRAW.boxFrames], [0, 1], {
            extrapolateLeft: 'clamp',
            extrapolateRight: 'clamp',
          });
          // The second pass trails the first by a few frames, the way a hand
          // going round twice does.
          const boxP2 = interpolate(
            since,
            [DRAW.boxStart + DRAW.secondPassDelay, DRAW.boxStart + DRAW.secondPassDelay + DRAW.boxFrames],
            [0, 1],
            {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
          );
          const iconP = interpolate(since, [DRAW.iconStart, DRAW.iconStart + DRAW.iconFrames], [0, 1], {
            extrapolateLeft: 'clamp',
            extrapolateRight: 'clamp',
          });

          // The item being drawn owns the accent, then settles back to chalk so
          // the eye always sits where the narration is.
          const heat = interpolate(
            since,
            [0, DRAW.labelStart, DRAW.labelStart + DRAW.settleFrames],
            [1, 1, 0],
            {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
          );
          const ink = heat > 0.5 ? theme.accent : theme.text;
          // The settled floor has to hold its own against a full-white label —
          // at 0.42 the outline read as a faint grey frame around bold text.
          const inkOpacity = 0.58 + heat * 0.42;

          const words = item.label.split(' ');
          const Icon = iconFor(item.icon);
          const pen = boxP > 0 && boxP < 1 ? penAt(stroke, boxP) : null;

          return (
            <g key={i}>
              {link && (
                <>
                  {/* Connectors carry the reading order of the whole board, so
                      they get the same ink and nearly the same weight as the
                      boxes. At muted-grey 3px they were the faintest thing on
                      screen while being the thing that says which way to read. */}
                  <DrawnPath
                    stroke={link.shaft}
                    progress={arrowP}
                    color={theme.text}
                    width={4.2}
                    opacity={0.58 + heat * 0.34}
                  />
                  <DrawnPath
                    stroke={link.head}
                    progress={interpolate(since, [DRAW.arrowFrames - 3, DRAW.arrowFrames + 3], [0, 1], {
                      extrapolateLeft: 'clamp',
                      extrapolateRight: 'clamp',
                    })}
                    color={theme.text}
                    width={4.2}
                    opacity={0.58 + heat * 0.34}
                  />
                </>
              )}
              <g transform={`translate(${rect.x} ${rect.y}) rotate(${tilt} ${rect.w / 2} ${rect.h / 2})`}>
                {/* The card's fill follows the outline round, so the box reads
                    as being filled in rather than appearing behind the stroke.
                    Near-opaque: at 0.55 the card barely separated from the
                    board and the boxes floated instead of sitting on it. */}
                <rect
                  x={0}
                  y={0}
                  width={rect.w}
                  height={rect.h}
                  rx={26}
                  fill={theme.surface}
                  opacity={boxP * 0.92}
                />
                <DrawnPath stroke={stroke} progress={boxP} color={ink} width={3.4} opacity={inkOpacity} />
                <DrawnPath
                  stroke={second}
                  progress={boxP2}
                  color={ink}
                  width={2.4}
                  opacity={inkOpacity * 0.55}
                />
                {/* The marker: a soft accent bead sitting exactly at the pen. */}
                {pen && (
                  <>
                    <circle cx={pen.x} cy={pen.y} r={17} fill={theme.accent} opacity={0.22} />
                    <circle cx={pen.x} cy={pen.y} r={6} fill={theme.accent} />
                  </>
                )}
                {/* Icon, wiped in left to right. */}
                {iconP > 0 && (
                  <g
                    transform={`translate(${rect.w / 2 - iconSize / 2} ${rect.h * 0.21})`}
                    clipPath={`url(#wb-clip-${i})`}
                    opacity={0.85 + heat * 0.15}
                  >
                    {/* The icon settles to the primary rather than back to
                        chalk, so a finished board still has a colour hierarchy
                        instead of being uniformly grey once the heat is gone. */}
                    <Icon
                      size={iconSize}
                      color={heat > 0.5 ? theme.accent : theme.primary}
                      strokeWidth={1.9}
                    />
                  </g>
                )}
                <clipPath id={`wb-clip-${i}`}>
                  <rect x={0} y={-4} width={iconSize * iconP} height={iconSize + 8} />
                </clipPath>
                {/* Label, written word by word. */}
                <text
                  x={rect.w / 2}
                  y={rect.h * 0.74}
                  textAnchor="middle"
                  fontFamily={theme.fontDisplay}
                  fontSize={fontSize}
                  fontWeight={600}
                  fill={theme.text}
                >
                  {words.map((word, wi) => {
                    const start = DRAW.labelStart + wi * DRAW.labelWordFrames;
                    return (
                      <tspan
                        key={wi}
                        opacity={interpolate(since, [start, start + DRAW.labelWordFrames], [0, 1], {
                          extrapolateLeft: 'clamp',
                          extrapolateRight: 'clamp',
                        })}
                      >
                        {wi > 0 ? ' ' : ''}
                        {word}
                      </tspan>
                    );
                  })}
                </text>
              </g>
            </g>
          );
        })}
      </svg>
    </Stage>
  );
};
