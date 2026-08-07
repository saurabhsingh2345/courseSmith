import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// HandshakeScene: a sequence diagram that behaves like one.
//
// The layout is the oldest drawing in computing and it is right for a reason:
// two headed columns, two lifelines dropping from them, and time running down
// the page. Every arrow's meaning is carried by its geometry — an arrow to the
// right is the initiator, an arrow to the left is an answer — so the viewer
// reads WHO is speaking before they read a single word, which is the thing a
// list of message names can never do.
//
// Delivered arrows stay exactly where they landed. That is the second reason
// for this shape: by the last message the middle of the frame is a transcript
// of the whole exchange, readable top to bottom, and the closing shot is that
// transcript with the wire lit behind it. A scene that cleared each arrow to
// make room for the next would have a tidier frame and no record.
//
// The label rides on the arrow and the meaning sits under it, half a size down
// and muted. The hierarchy is deliberate: the label is the protocol's own
// vocabulary, which the viewer will meet again in documentation, and the
// meaning is the part that makes the vocabulary stick. Putting them the same
// size would make the row a sentence instead of a fact with a gloss.
//
// One glow, saved for the end: the open channel fills the gap between the
// lifelines with a soft accent wash. Nothing else in the frame is lit, so the
// state the whole exchange existed to reach is the only thing that arrives
// with light behind it.

const WIRE_W = Math.min(STAGE_W, 1380);
// The headers, their glyphs and the gap before the first arrow.
const HEAD_H = 152;
// One arrow, its label and its meaning. Sized so six rows still clear the
// caption band at the bottom of the stage.
const ROW_H = 84;
const COL_INSET = 190;
const HEAD_W = 300;

type Msg = {dir: string; label: string; means: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'wire' | 'msg' | 'open';
  at?: number;
  delivered: number[];
};

/** The two role glyphs. Same stroke weight, same construction. */
const RoleGlyph: React.FC<{side: 'left' | 'right'; colour: string}> = ({side, colour}) => {
  const s = {fill: 'none', stroke: colour, strokeWidth: 2.2, strokeLinecap: 'round' as const, strokeLinejoin: 'round' as const};
  if (side === 'left') {
    // A window: the thing a person is looking at.
    return (
      <g>
        <rect x={-14} y={-11} width={28} height={22} rx={3} {...s} />
        <path d="M -14 -4 L 14 -4" {...s} />
        <circle cx={-9.5} cy={-7.5} r={1.5} fill={colour} stroke="none" />
      </g>
    );
  }
  // A stack: the thing that answers.
  return (
    <g>
      <rect x={-14} y={-12} width={28} height={10} rx={2.5} {...s} />
      <rect x={-14} y={2} width={28} height={10} rx={2.5} {...s} />
      <circle cx={-8.5} cy={-7} r={1.5} fill={colour} stroke="none" />
      <circle cx={-8.5} cy={7} r={1.5} fill={colour} stroke="none" />
    </g>
  );
};

export const HandshakeScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const left = String(props.left ?? '');
  const right = String(props.right ?? '');
  const msgs = (Array.isArray(props.msgs) ? props.msgs : []) as Msg[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (msgs.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const delivered = Array.isArray(step.delivered) ? step.delivered : [];
  const firing = step.show === 'msg' ? (step.at ?? -1) : -1;
  const isOpen = step.show === 'open';
  const isWire = step.show === 'wire';

  const leftX = COL_INSET;
  const rightX = WIRE_W - COL_INSET;
  const svgH = HEAD_H + msgs.length * ROW_H + 46;
  const rowY = (i: number) => HEAD_H + i * ROW_H + ROW_H / 2;

  const headEnter = isWire
    ? interpolate(sinceStep, [2, 20], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : 1;
  const lifeline = isWire
    ? interpolate(sinceStep, [10, 34], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : 1;
  const openGlow = isOpen
    ? spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 28})
    : 0;

  const column = (x: number, text: string, side: 'left' | 'right') => {
    const colour = isOpen ? theme.accentText : theme.text;
    return (
      <g transform={`translate(${x} 46)`} opacity={headEnter}>
        <rect
          x={-HEAD_W / 2}
          y={-34}
          width={HEAD_W}
          height={68}
          rx={12}
          fill={withAlpha(theme.surface, 0.9)}
          stroke={isOpen ? withAlpha(theme.accent, 0.7) : theme.surfaceBorder}
          strokeWidth={1.6}
        />
        <g transform={`translate(${-HEAD_W / 2 + 42} 0)`}>
          <RoleGlyph side={side} colour={isOpen ? theme.accentText : theme.textMuted} />
        </g>
        <text
          x={-HEAD_W / 2 + 74}
          y={8}
          fill={colour}
          style={{fontFamily: theme.fontBody, fontSize: 25, fontWeight: 600}}
        >
          {text}
        </text>
      </g>
    );
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

      <svg width={WIRE_W} height={svgH} viewBox={`0 0 ${WIRE_W} ${svgH}`} style={{flexShrink: 0, overflow: 'visible'}}>
        {/* The wire itself, lit only at the end. */}
        {openGlow > 0 ? (
          <rect
            x={leftX}
            y={HEAD_H - 26}
            width={rightX - leftX}
            height={msgs.length * ROW_H + 40}
            rx={16}
            fill={withAlpha(theme.accent, 0.1 * openGlow)}
          />
        ) : null}

        {[leftX, rightX].map((x) => (
          <line
            key={x}
            x1={x}
            y1={92}
            x2={x}
            y2={92 + (svgH - 118) * lifeline}
            stroke={isOpen ? withAlpha(theme.accent, 0.75) : withAlpha(theme.line, 0.32)}
            strokeWidth={isOpen ? 2.2 : 1.4}
            strokeDasharray={isOpen ? undefined : '5 8'}
          />
        ))}

        {column(leftX, left, 'left')}
        {column(rightX, right, 'right')}

        {msgs.map((m, i) => {
          const landed = delivered.includes(i);
          const isFiring = i === firing;
          if (!landed && !isFiring) return null;
          const y = rowY(i);
          const ltr = m.dir !== 'rtl';
          const from = ltr ? leftX : rightX;
          const to = ltr ? rightX : leftX;
          const fly = isFiring
            ? spring({frame: sinceStep - 4, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 28})
            : 1;
          const tip = from + (to - from) * fly;
          const head = ltr ? -13 : 13;
          const stroke = isFiring ? theme.accent : withAlpha(theme.accent, 0.55);
          return (
            <g key={i} opacity={isFiring ? Math.min(1, fly * 3) : 1}>
              <line x1={from} y1={y} x2={tip} y2={y} stroke={stroke} strokeWidth={isFiring ? 2.6 : 1.8} strokeLinecap="round" />
              <path
                d={`M ${tip + head} ${y - 7} L ${tip} ${y} L ${tip + head} ${y + 7}`}
                fill="none"
                stroke={stroke}
                strokeWidth={isFiring ? 2.6 : 1.8}
                strokeLinecap="round"
                strokeLinejoin="round"
              />
              <text
                x={(leftX + rightX) / 2}
                y={y - 16}
                textAnchor="middle"
                fill={isFiring ? theme.accentText : theme.text}
                style={{fontFamily: theme.fontMono, fontSize: 25, fontWeight: 600, letterSpacing: 0.4}}
              >
                {m.label}
              </text>
              <text
                x={(leftX + rightX) / 2}
                y={y + 30}
                textAnchor="middle"
                fill={theme.textMuted}
                style={{fontFamily: theme.fontBody, fontSize: 20}}
              >
                {m.means}
              </text>
            </g>
          );
        })}
      </svg>

      <div
        style={{
          width: WIRE_W,
          display: 'flex',
          alignItems: 'baseline',
          gap: 20,
          marginTop: 26,
          opacity: openGlow,
          transform: `translateY(${(1 - openGlow) * 10}px)`,
        }}
      >
        <span
          style={{
            fontFamily: theme.fontMono,
            fontSize: 16,
            letterSpacing: 3.4,
            textTransform: 'uppercase',
            color: theme.textMuted,
          }}
        >
          established
        </span>
        <span
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 38,
            fontWeight: 600,
            letterSpacing: -0.6,
            color: theme.accentText,
          }}
        >
          {left} ⇄ {right}
        </span>
      </div>
    </Stage>
  );
};
