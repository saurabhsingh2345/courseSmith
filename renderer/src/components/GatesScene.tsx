import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// GatesScene: the circuit on the left, its receipt on the right.
//
// The gate is drawn with its real IEEE silhouette rather than as a labelled
// rectangle, and that is the whole reason this template exists. A rectangle
// with "XOR" written in it teaches a word. The distinctive shapes — the flat
// back and round nose of an AND, the curved back of an OR, the doubled back of
// an XOR, the little bubble that means "and then invert" — are a vocabulary the
// viewer will meet again on every circuit diagram for the rest of their life,
// and they are learnable in one clip because each one is a picture of what the
// gate does.
//
// Signal is carried by stroke brightness, not by animation along the wire. Each
// wire is drawn twice: a cold stroke in the line token, and a hot stroke in the
// quantity accent whose opacity comes up as the signal reaches it. Inputs
// first, then the body, then the output, then the table row — a propagation
// delay of about half a second, which is a lie about physics and the truth
// about causation. A dot travelling the wire was the obvious alternative and
// says the wrong thing: it makes the signal an object in transit rather than a
// level that settles.
//
// The truth table fills in from the top as the rows fire, with the live row
// held in the accent. By the closing beat it is complete, and the gate's one
// line rule sits under the circuit — a sentence the table has just proved
// rather than a fact asserted beside it.
//
// One glow maximum: the gate body while it is passing a one. The gate function
// itself was evaluated in Go, so every number on this screen is arithmetic that
// was checked, not a claim this component made up.

const LANE_W = Math.min(STAGE_W, 1320);
const CIRCUIT_W = 780;
const CIRCUIT_H = 390;
const TABLE_W = 420;
const ROW_H = 58;

type Row = {in: number[]; out: number};
type Step = {
  startMs: number;
  endMs: number;
  show: 'circuit' | 'row' | 'law';
  at?: number;
  done: number[];
};

type Shape = {
  body: string;
  /** A stroke-only flourish behind the body — the XOR's second back. */
  extra?: string;
  bubble?: {cx: number; cy: number; r: number};
  /** Where input wires stop, and where the output wire begins. */
  wireInX: number;
  tipX: number;
};

/**
 * The IEEE silhouettes, in one 600x300 space shared by the wires.
 *
 * Written out per gate rather than derived, because the shapes are not
 * variations on a theme — the flat-backed D and the pointed lens are different
 * drawings, and a parameterised body would have produced a compromise that is
 * neither.
 */
const shapeOf = (gate: string): Shape => {
  switch (gate) {
    case 'NOT':
      return {body: 'M250 90 L250 210 L355 150 Z', bubble: {cx: 366, cy: 150, r: 11}, wireInX: 250, tipX: 377};
    case 'OR':
      return {body: 'M240 90 Q285 150 240 210 Q305 210 360 150 Q305 90 240 90 Z', wireInX: 262, tipX: 360};
    case 'NOR':
      return {
        body: 'M240 90 Q285 150 240 210 Q305 210 360 150 Q305 90 240 90 Z',
        bubble: {cx: 372, cy: 150, r: 11},
        wireInX: 262,
        tipX: 383,
      };
    case 'XOR':
      return {
        body: 'M240 90 Q285 150 240 210 Q305 210 360 150 Q305 90 240 90 Z',
        extra: 'M218 90 Q263 150 218 210',
        wireInX: 262,
        tipX: 360,
      };
    case 'NAND':
      return {
        body: 'M240 90 L300 90 A60 60 0 0 1 300 210 L240 210 Z',
        bubble: {cx: 372, cy: 150, r: 11},
        wireInX: 240,
        tipX: 383,
      };
    default:
      return {body: 'M240 90 L300 90 A60 60 0 0 1 300 210 L240 210 Z', wireInX: 240, tipX: 360};
  }
};

const clamp01 = (v: number) => Math.max(0, Math.min(1, v));

export const GatesScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const gate = String(props.gate ?? '');
  const law = String(props.law ?? '');
  const inputs = (Array.isArray(props.inputs) ? props.inputs : []) as string[];
  const rows = (Array.isArray(props.rows) ? props.rows : []) as Row[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (inputs.length === 0 || rows.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const shape = shapeOf(gate);
  const active = step.show === 'row' ? (step.at ?? -1) : -1;
  const firing = active >= 0 && active < rows.length;
  const live = firing ? rows[active] : null;
  const done = new Set(step.done);

  // The propagation: inputs, then body, then output, then the table row.
  const sig = firing
    ? interpolate(sinceStep, [2, 30], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : 0;
  const hotIn = (v: number) => (v === 1 ? clamp01((sig - 0.05) / 0.35) : 0);
  const hotBody = clamp01((sig - 0.4) / 0.3);
  const hotOut = live && live.out === 1 ? clamp01((sig - 0.7) / 0.28) : 0;
  const tick = clamp01((sig - 0.82) / 0.18);

  const pinY = (i: number) => (inputs.length === 1 ? 150 : i === 0 ? 118 : 182);
  const lawIn =
    step.show === 'law'
      ? spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 28})
      : 0;

  const wire = (d: string, hot: number, key: string): React.ReactNode => (
    <g key={key}>
      <path d={d} stroke={theme.line} strokeWidth={3} strokeLinecap="round" fill="none" opacity={0.7} />
      <path
        d={d}
        stroke={theme.accentQuantity}
        strokeWidth={4}
        strokeLinecap="round"
        fill="none"
        opacity={hot}
      />
    </g>
  );

  const chip = (value: number, cx: number, cy: number, hot: number, key: string): React.ReactNode => (
    <g key={key}>
      <rect
        x={cx - 23}
        y={cy - 23}
        width={46}
        height={46}
        rx={12}
        fill={withAlpha(theme.surface, 0.95)}
        stroke={value === 1 ? theme.accentQuantity : theme.surfaceBorder}
        strokeWidth={2}
        opacity={0.35 + 0.65 * Math.max(hot, value === 1 ? 0 : 1)}
      />
      <text
        x={cx}
        y={cy + 11}
        textAnchor="middle"
        fontFamily={theme.fontMono}
        fontSize={30}
        fontWeight={700}
        fill={value === 1 ? theme.accentQuantity : theme.textMuted}
      >
        {value}
      </text>
    </g>
  );

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={18}
      />

      <div style={{width: LANE_W, display: 'flex', alignItems: 'center', gap: 48}}>
        <div style={{flexShrink: 0, width: CIRCUIT_W}}>
          <svg width={CIRCUIT_W} height={CIRCUIT_H} viewBox="0 0 600 300">
            {/* Input pins: label, level chip, wire. */}
            {inputs.map((label, i) => {
              const value = live ? (live.in[i] ?? 0) : 0;
              const y = pinY(i);
              return (
                <g key={`in-${i}`}>
                  <text
                    x={46}
                    y={y + 9}
                    textAnchor="end"
                    fontFamily={theme.fontBody}
                    fontSize={26}
                    fill={theme.textMuted}
                  >
                    {label}
                  </text>
                  {chip(value, 84, y, hotIn(value), `chip-${i}`)}
                  {wire(`M112 ${y} H${shape.wireInX}`, hotIn(value), `wire-${i}`)}
                </g>
              );
            })}

            {/* The output wire and its level. */}
            {wire(`M${shape.tipX} 150 H492`, hotOut, 'wire-out')}
            {chip(live ? live.out : 0, 522, 150, hotOut, 'chip-out')}
            <text
              x={522}
              y={104}
              textAnchor="middle"
              fontFamily={theme.fontBody}
              fontSize={20}
              letterSpacing={2}
              fill={theme.textMuted}
            >
              out
            </text>

            {/* The gate itself. */}
            <g
              style={{
                // The one glow: the body while it is passing a one.
                filter: hotOut > 0.2 ? `drop-shadow(0 0 16px ${withAlpha(theme.accentQuantity, 0.55)})` : undefined,
              }}
            >
              {shape.extra ? (
                <path
                  d={shape.extra}
                  stroke={theme.accent}
                  strokeWidth={4}
                  fill="none"
                  strokeLinecap="round"
                  opacity={0.85}
                />
              ) : null}
              <path
                d={shape.body}
                fill={withAlpha(theme.accent, 0.1 + 0.22 * hotBody)}
                stroke={hotBody > 0.5 ? theme.accentQuantity : theme.accent}
                strokeWidth={4}
                strokeLinejoin="round"
              />
              {shape.bubble ? (
                <circle
                  cx={shape.bubble.cx}
                  cy={shape.bubble.cy}
                  r={shape.bubble.r}
                  fill={withAlpha(theme.surface, 1)}
                  stroke={hotBody > 0.5 ? theme.accentQuantity : theme.accent}
                  strokeWidth={4}
                />
              ) : null}
              <text
                x={shape.body.startsWith('M250') ? 282 : 292}
                y={162}
                textAnchor="middle"
                fontFamily={theme.fontDisplay}
                fontSize={38}
                fontWeight={800}
                letterSpacing={1}
                fill={theme.text}
              >
                {gate}
              </text>
            </g>
          </svg>

          {/* The closer, under the circuit it was proved by. */}
          <div
            style={{
              minHeight: 62,
              marginTop: 8,
              paddingLeft: 40,
              fontFamily: theme.fontDisplay,
              fontSize: 40,
              fontWeight: 700,
              letterSpacing: -0.5,
              color: theme.accentText,
              opacity: lawIn,
              transform: `translateY(${(1 - lawIn) * 14}px)`,
            }}
          >
            {law}
          </div>
        </div>

        {/* The receipt. */}
        <div style={{flexShrink: 0, width: TABLE_W}}>
          <div
            style={{
              display: 'flex',
              paddingBottom: 12,
              borderBottom: `2px solid ${theme.surfaceBorder}`,
              fontFamily: theme.fontMono,
              fontSize: 20,
              letterSpacing: 2.4,
              textTransform: 'uppercase',
              color: theme.textMuted,
            }}
          >
            {inputs.map((label, i) => (
              <div key={i} style={{flex: 1, textAlign: 'center'}}>
                {label}
              </div>
            ))}
            <div style={{flex: 1, textAlign: 'center', color: theme.accentText}}>out</div>
          </div>
          {rows.map((row, i) => {
            const shown = done.has(i);
            const isLive = i === active;
            const opacity = !shown ? 0 : isLive ? tick : 1;
            return (
              <div
                key={i}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  height: ROW_H,
                  borderRadius: 10,
                  background: isLive ? withAlpha(theme.accentQuantity, 0.16) : 'transparent',
                  opacity,
                  fontFamily: theme.fontMono,
                  fontSize: 30,
                  fontWeight: 700,
                }}
              >
                {row.in.map((v, j) => (
                  <div key={j} style={{flex: 1, textAlign: 'center', color: v === 1 ? theme.text : theme.textMuted}}>
                    {v}
                  </div>
                ))}
                <div
                  style={{
                    flex: 1,
                    textAlign: 'center',
                    color: row.out === 1 ? theme.accentQuantity : theme.textMuted,
                  }}
                >
                  {row.out}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </Stage>
  );
};
