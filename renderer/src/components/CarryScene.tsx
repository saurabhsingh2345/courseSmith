import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// CarryScene: long addition, drawn the way it is done on paper.
//
// The composition is deliberately the most familiar object in the whole
// catalog — a number over a number, a rule underneath, an answer below it —
// because the argument of the clip is that the viewer already knows this
// procedure. Any cleverness in the layout would work against that. So the grid
// is orthodox: right-aligned columns, a monospaced digit per cell, a horizontal
// rule, and the carries written small above the columns they arrive in, which
// is exactly where a person writes them.
//
// The carry chip is the only thing that moves horizontally, and it is the one
// element allowed to look like an object rather than a character. When a column
// with a carry out is worked, a chip is born at that column and ARCS up and to
// the left into the carry lane above its neighbour, where it then stays. The
// arc matters: a chip that faded in above the next column would be a value
// appearing, and the whole point is that it was HANDED over. Every other state
// change is a fill or an opacity.
//
// Nothing here is computed. Each column's two bits, its incoming carry, its
// result digit and its outgoing carry all arrive from Go, already worked, in
// significance order with entry 0 the rightmost. The grid is drawn reversed,
// which is one line here and keeps the arithmetic order honest on the other
// side. A diagram of computing that is wrong is worse than no diagram, and this
// component is in no position to check anything.
//
// The decimal equivalents sit in muted type off the right-hand end of each row.
// They are not the subject — they are the receipt, and the sum's own decimal
// only arrives on the closing beat, when it can be read as a confirmation
// rather than as a second thing to follow.

const CELL_GAP = 10;
const MAX_CELL = 92;
const CARRY_H = 58;
const ROW_GAP = 14;
const RULE_H = 3;
// Room off the right-hand end for "= 11" in muted type, at the size the rows
// are set in.
const DEC_W = 210;
// Room off the left-hand end for the plus sign.
const SIGN_W = 78;

type Column = {a: string; b: string; carryIn: number; digit: string; carryOut: number};
type Step = {
  startMs: number;
  endMs: number;
  show: 'problem' | 'column' | 'carrychain' | 'answer';
  at?: number;
  done: number[];
};

export const CarryScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const columns = (Array.isArray(props.columns) ? props.columns : []) as Column[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const aDecimal = Number(props.aDecimal ?? 0);
  const bDecimal = Number(props.bDecimal ?? 0);
  const sumDecimal = Number(props.sumDecimal ?? 0);
  if (columns.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const width = columns.length;
  const gridW = Math.min(STAGE_W - SIGN_W - DEC_W, 1180);
  const cellW = Math.min(MAX_CELL, Math.floor((gridW - CELL_GAP * (width - 1)) / width));
  const rowW = cellW * width + CELL_GAP * (width - 1);
  const digitSize = Math.round(cellW * 0.62);

  // Draw position 0 is the LEFTMOST cell, which is the most significant column.
  const xOf = (significance: number) => (width - 1 - significance) * (cellW + CELL_GAP);

  const worked = new Set(step.done);
  const working = step.show === 'column' ? (step.at ?? -1) : -1;
  const finished = step.show === 'answer';
  const chainLit = step.show === 'carrychain';

  const yA = CARRY_H;
  const yB = yA + cellW + ROW_GAP;
  const yRule = yB + cellW + ROW_GAP;
  const ySum = yRule + RULE_H + ROW_GAP;
  const gridH = ySum + cellW;

  // The digit landing in the sum row, and the chip leaving for the next column.
  const land = working >= 0 ? spring({frame: sinceStep - 8, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 24}) : 1;
  const hop =
    working >= 0
      ? spring({frame: sinceStep - 18, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 28})
      : 0;

  const cell = (
    text: string,
    x: number,
    y: number,
    colour: string,
    opacity: number,
    highlight: boolean,
  ): React.ReactNode => (
    <div
      style={{
        position: 'absolute',
        left: x,
        top: y,
        width: cellW,
        height: cellW,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        borderRadius: 12,
        fontFamily: theme.fontMono,
        fontSize: digitSize,
        fontWeight: 700,
        color: colour,
        opacity,
        background: highlight ? withAlpha(theme.accent, 0.14) : withAlpha(theme.surface, 0.75),
        border: `2px solid ${highlight ? theme.accent : theme.surfaceBorder}`,
      }}
    >
      {text}
    </div>
  );

  const decimal = (value: number, y: number, colour: string, opacity: number): React.ReactNode => (
    <div
      style={{
        position: 'absolute',
        left: rowW + 30,
        top: y + cellW / 2 - 20,
        width: DEC_W,
        fontFamily: theme.fontMono,
        fontSize: 30,
        color: colour,
        opacity,
      }}
    >
      = {value}
    </div>
  );

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

      <div style={{position: 'relative', width: rowW + SIGN_W + DEC_W, height: gridH}}>
        <div style={{position: 'relative', left: SIGN_W, width: rowW, height: gridH}}>
          {/* The plus sign, off the left edge of the second operand. */}
          <div
            style={{
              position: 'absolute',
              left: -SIGN_W + 8,
              top: yB,
              width: SIGN_W - 24,
              height: cellW,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'flex-end',
              fontFamily: theme.fontMono,
              fontSize: digitSize,
              fontWeight: 700,
              color: theme.textMuted,
            }}
          >
            +
          </div>

          {/* The carry lane. A chip sits above the column its carry arrived in. */}
          {columns.map((c, i) => {
            if (c.carryIn !== 1) return null;
            // The chip only exists once the column that produced it was worked.
            const born = worked.has(i - 1) || finished || chainLit;
            if (!born) return null;
            const leaving = working === i - 1;
            const x = xOf(i);
            const travel = leaving ? 1 - hop : 0;
            return (
              <div
                key={`carry-${i}`}
                style={{
                  position: 'absolute',
                  left: x + cellW / 2 - 19,
                  top: 4,
                  width: 38,
                  height: 38,
                  borderRadius: 19,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontFamily: theme.fontMono,
                  fontSize: 22,
                  fontWeight: 700,
                  color: theme.ink,
                  background: chainLit ? theme.accentQuantity : theme.accent,
                  opacity: leaving ? hop : 1,
                  // Born on its own column and arcing up and to the left into the
                  // lane above its neighbour: handed over, not appearing.
                  transform: `translate(${travel * (cellW + CELL_GAP)}px, ${travel * (CARRY_H + 6)}px)`,
                  boxShadow: chainLit ? `0 0 22px ${withAlpha(theme.accentQuantity, 0.5)}` : undefined,
                }}
              >
                1
              </div>
            );
          })}

          {/* The two operands. They are up from the first frame: the problem is
              the opener, and nothing about it is revealed piecemeal. */}
          {columns.map((c, i) =>
            cell(c.a, xOf(i), yA, theme.text, 1, working === i),
          )}
          {columns.map((c, i) =>
            cell(c.b, xOf(i), yB, theme.text, 1, working === i),
          )}

          <div
            style={{
              position: 'absolute',
              left: 0,
              top: yRule,
              width: rowW,
              height: RULE_H,
              borderRadius: 2,
              background: theme.line,
            }}
          />

          {/* The answer, one digit at a time. */}
          {columns.map((c, i) => {
            const done = worked.has(i) || finished;
            if (!done) return null;
            const isCurrent = working === i;
            const pop = isCurrent ? land : 1;
            return (
              <div
                key={`sum-${i}`}
                style={{
                  position: 'absolute',
                  left: xOf(i),
                  top: ySum,
                  width: cellW,
                  height: cellW,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  borderRadius: 12,
                  fontFamily: theme.fontMono,
                  fontSize: digitSize,
                  fontWeight: 700,
                  color: theme.ink,
                  background: theme.accentQuantity,
                  opacity: isCurrent ? pop : 1,
                  transform: `scale(${0.86 + 0.14 * pop})`,
                  // The one glow: the digit landing right now.
                  boxShadow: isCurrent ? `0 0 28px ${withAlpha(theme.accentQuantity, 0.55)}` : undefined,
                }}
              >
                {c.digit}
              </div>
            );
          })}

          {decimal(aDecimal, yA, theme.textMuted, 1)}
          {decimal(bDecimal, yB, theme.textMuted, 1)}
          {decimal(
            sumDecimal,
            ySum,
            theme.accentQuantity,
            finished
              ? interpolate(sinceStep, [4, 20], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
              : 0,
          )}
        </div>
      </div>

      {/* The receipt, and only on the closing beat: the same sum in the base the
          viewer has always used. */}
      <div
        style={{
          marginTop: 30,
          fontFamily: theme.fontBody,
          fontSize: 30,
          color: theme.textMuted,
          opacity: finished
            ? interpolate(sinceStep, [12, 28], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
            : 0,
        }}
      >
        {aDecimal} + {bDecimal} = {sumDecimal}
      </div>
    </Stage>
  );
};
