import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// RadixScene: one number spelled three ways, with the arithmetic done on screen.
//
// The layout is three labelled rows on a hard left rail — base 10, base 2,
// base 16 — because the claim of the clip is that these are one value, and
// stacking them makes the equality a vertical alignment the eye can check.
// The decimal row never moves: it is the referee, and the other two rows are
// judged against it.
//
// The binary row is the working area. Each bit is a monospaced cell with its
// place value set small above it, and during the "sum" beat the set bits
// light left to right while a running total ticks up at the right-hand end of
// the row. Every number the ticker shows arrived precomputed from Go
// (sumSteps) — this component never does base conversion, because a diagram
// of computing that is wrong is worse than none, and the Go validator is
// where the arithmetic was proven.
//
// The hex row reuses the same cells conceptually: nibble brackets slide in
// under groups of four bits and each bracket hands its group down to one hex
// digit. That physical hand-off — bracket above, digit below — is the whole
// lesson about hex, so it gets the spring.
//
// One glow maximum: the currently-lighting bit during the sum. Everything
// else changes by fill and opacity, calmly.

const ROW_W = Math.min(STAGE_W, 1360);
const LABEL_W = 130;
const CELL_GAP = 8;
const TOTAL_W = 210;

type Cell = {bit: string; weight: number};
type SumStep = {at: number; weight: number; total: number};
type Nibble = {bits: string; hex: string};
type Step = {startMs: number; endMs: number; show: 'value' | 'weights' | 'sum' | 'hex' | 'same'};

const RANK: Record<string, number> = {value: 0, weights: 1, sum: 2, hex: 3, same: 4};

export const RadixScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const decimal = Number(props.decimal ?? 0);
  const story = String(props.story ?? '');
  const hex = String(props.hex ?? '');
  const cells = (Array.isArray(props.cells) ? props.cells : []) as Cell[];
  const sumSteps = (Array.isArray(props.sumSteps) ? props.sumSteps : []) as SumStep[];
  const nibbles = (Array.isArray(props.nibbles) ? props.nibbles : []) as Nibble[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (cells.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // The furthest state ever reached, so the picture accumulates: once the
  // columns are up they stay up, once the bits have summed they stay lit.
  const reached = Math.max(...steps.slice(0, idx + 1).map((s) => RANK[s.show] ?? 0));
  const inSum = step.show === 'sum';

  // How many set bits have lit. During the sum they light one at a time, a
  // fresh spring each; after it, all of them.
  const litInterval = 14;
  const litCount = inSum
    ? Math.min(sumSteps.length, Math.max(0, Math.floor((sinceStep - 6) / litInterval) + 1))
    : reached > RANK.sum
      ? sumSteps.length
      : 0;
  const litBits = new Set(sumSteps.slice(0, litCount).map((s) => s.at));
  const currentLit = inSum && litCount > 0 ? sumSteps[litCount - 1].at : -1;
  const runningTotal = litCount > 0 ? sumSteps[litCount - 1].total : 0;

  const rowReveal = (on: boolean, delay = 0) =>
    on
      ? interpolate(sinceStep, [delay, delay + 14], [0, 1], {
          extrapolateLeft: 'clamp',
          extrapolateRight: 'clamp',
        })
      : 0;

  const cellW = Math.min(74, Math.floor((ROW_W - LABEL_W - TOTAL_W - CELL_GAP * (cells.length - 1)) / cells.length));
  const bitsW = cellW * cells.length + CELL_GAP * (cells.length - 1);

  const weightsUp = reached >= RANK.weights;
  const hexUp = reached >= RANK.hex;
  const allLit = step.show === 'same';

  const rowLabel = (text: string, colour: string): React.ReactNode => (
    <div
      style={{
        width: LABEL_W,
        flexShrink: 0,
        fontFamily: theme.fontMono,
        fontSize: 16,
        letterSpacing: 2.4,
        textTransform: 'uppercase',
        color: colour,
        paddingTop: 6,
      }}
    >
      {text}
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
        marginBottom={34}
      />

      {/* Base 10: the referee. It never moves. */}
      <div style={{width: ROW_W, display: 'flex', alignItems: 'baseline', marginBottom: 40}}>
        {rowLabel('base 10', theme.textMuted)}
        <span
          style={{
            fontFamily: theme.fontMono,
            fontSize: 84,
            fontWeight: 700,
            letterSpacing: -2,
            color: allLit ? theme.accentText : theme.text,
            lineHeight: 1,
          }}
        >
          {decimal}
        </span>
        {story ? (
          <span
            style={{
              fontFamily: theme.fontBody,
              fontSize: 26,
              color: theme.textMuted,
              marginLeft: 30,
            }}
          >
            {story}
          </span>
        ) : null}
      </div>

      {/* Base 2: the working row. Weights above, cells below, ticker right. */}
      <div
        style={{
          width: ROW_W,
          display: 'flex',
          alignItems: 'flex-end',
          marginBottom: 44,
          opacity: weightsUp ? Math.max(rowReveal(step.show === 'weights'), weightsUp ? 1 : 0) : 0,
        }}
      >
        {rowLabel('base 2', weightsUp ? theme.textMuted : 'transparent')}
        {weightsUp ? (
          <div style={{width: bitsW, flexShrink: 0}}>
            <div style={{display: 'flex', gap: CELL_GAP, marginBottom: 8}}>
              {cells.map((c, i) => (
                <div
                  key={i}
                  style={{
                    width: cellW,
                    textAlign: 'center',
                    fontFamily: theme.fontMono,
                    fontSize: cells.length > 8 ? 13 : 16,
                    color: litBits.has(i) ? theme.accentQuantity : theme.textMuted,
                    opacity: interpolate(
                      step.show === 'weights' ? sinceStep : 99,
                      [4 + i * 2, 12 + i * 2],
                      [0, 1],
                      {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
                    ),
                  }}
                >
                  {c.weight}
                </div>
              ))}
            </div>
            <div style={{display: 'flex', gap: CELL_GAP}}>
              {cells.map((c, i) => {
                const lit = litBits.has(i) && c.bit === '1';
                const isCurrent = i === currentLit;
                const pop = isCurrent
                  ? spring({frame: sinceStep - 6 - (litCount - 1) * litInterval, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 24})
                  : 1;
                return (
                  <div
                    key={i}
                    style={{
                      width: cellW,
                      height: cellW + 10,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      borderRadius: 10,
                      fontFamily: theme.fontMono,
                      fontSize: Math.round(cellW * 0.58),
                      fontWeight: 700,
                      color: lit ? theme.ink : c.bit === '1' ? theme.text : theme.textMuted,
                      background: lit ? theme.accentQuantity : withAlpha(theme.surface, 0.8),
                      border: `2px solid ${lit ? theme.accentQuantity : theme.surfaceBorder}`,
                      transform: `scale(${0.9 + 0.1 * pop})`,
                      // The one glow: the bit lighting right now.
                      boxShadow: isCurrent ? `0 0 24px ${withAlpha(theme.accentQuantity, 0.5)}` : undefined,
                    }}
                  >
                    {c.bit}
                  </div>
                );
              })}
            </div>
          </div>
        ) : null}
        {weightsUp && (litCount > 0 || reached > RANK.weights) ? (
          <div
            style={{
              width: TOTAL_W,
              textAlign: 'right',
              fontFamily: theme.fontMono,
              fontSize: 44,
              fontWeight: 700,
              color: runningTotal === decimal ? theme.accentQuantity : theme.text,
              paddingBottom: 14,
              marginLeft: 'auto',
            }}
          >
            {litCount > 0 ? `= ${runningTotal}` : ''}
          </div>
        ) : null}
      </div>

      {/* Base 16: nibble brackets sliding in, hex digits landing beneath. */}
      <div style={{width: ROW_W, display: 'flex', alignItems: 'flex-start', minHeight: 130}}>
        {rowLabel('base 16', hexUp ? theme.textMuted : 'transparent')}
        {hexUp ? (
          <div style={{display: 'flex', gap: 34}}>
            {nibbles.map((nb, i) => {
              const slide =
                step.show === 'hex'
                  ? spring({frame: sinceStep - 4 - i * 10, fps, config: {damping: 14, mass: 0.5}, durationInFrames: 26})
                  : 1;
              return (
                <div key={i} style={{display: 'flex', flexDirection: 'column', alignItems: 'center', opacity: slide}}>
                  <div
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: 24,
                      letterSpacing: 6,
                      color: allLit ? theme.accentText : theme.textMuted,
                      transform: `translateX(${(1 - slide) * -20}px)`,
                    }}
                  >
                    {nb.bits}
                  </div>
                  {/* The bracket: the physical hand-off from four bits to one digit. */}
                  <div
                    style={{
                      width: 120 * slide,
                      height: 12,
                      borderLeft: `2px solid ${theme.accent}`,
                      borderRight: `2px solid ${theme.accent}`,
                      borderBottom: `2px solid ${theme.accent}`,
                      borderRadius: '0 0 8px 8px',
                      marginTop: 4,
                    }}
                  />
                  <div
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: 52,
                      fontWeight: 700,
                      color: allLit ? theme.accentText : theme.text,
                      marginTop: 6,
                      transform: `translateY(${(1 - slide) * -8}px)`,
                    }}
                  >
                    {nb.hex}
                  </div>
                </div>
              );
            })}
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 30,
                color: theme.textMuted,
                alignSelf: 'flex-end',
                paddingBottom: 10,
                marginLeft: 6,
              }}
            >
              = 0x{hex}
            </div>
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
