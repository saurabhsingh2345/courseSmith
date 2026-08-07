import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// BitfieldScene: a wall of bits becoming a structure.
//
// The row is drawn once, at the top, and it never moves. That is the whole
// composition: everything else in the frame is an annotation attached to a
// stretch of it. Cells do not reflow when a field lifts, brackets do not push
// the row around, and the meaning line lands in a reserved band underneath
// rather than growing the layout. A picture whose job is "these bits, that
// region" cannot afford the region to be somewhere else each beat.
//
// Structure is carried by colour bands rather than by boxes. Each field gets a
// thick underline in one of three semantic accents, cycled so that adjacent
// fields never share one, plus a bracket and a label below it. Boxes round each
// field would have worked and would have been worse: a border reads as a
// container, and these fields are not containers, they are stretches of the
// same row. An underline reads as "this part of that", which is the claim.
//
// The lift is a scale, not a colour change. A focused field's cells grow by
// about eight percent and its band thickens, which is enough to be unmissable
// while leaving every unfocused cell exactly where it was — so the viewer can
// still see the whole pattern while one part of it is being explained. The one
// glow in the scene is on the focused band.
//
// Nothing is decoded here. Each field arrives from Go with its own bits already
// sliced out and their unsigned value already computed, because the interval
// arithmetic that proves the fields tile the row happens in the validator and
// this component is in no position to disagree with it.

const ROW_W = Math.min(STAGE_W, 1660);
const CELL_GAP = 4;
const BAND_H = 8;
const BRACKET_H = 14;
// A reserved band for the meaning line, so a field lifting never moves the row.
const MEANS_H = 132;

type Cell = {bit: string; index: number};
type Field = {label: string; from: number; to: number; means: string; bits: string; value: number};
type Step = {
  startMs: number;
  endMs: number;
  show: 'row' | 'split' | 'field' | 'read';
  at?: number;
  done: number[];
};

const RANK: Record<string, number> = {row: 0, split: 1, field: 2, read: 3};

export const BitfieldScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const cells = (Array.isArray(props.cells) ? props.cells : []) as Cell[];
  const fields = (Array.isArray(props.fields) ? props.fields : []) as Field[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (cells.length === 0 || fields.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // The furthest state ever reached, so the picture accumulates: once the
  // brackets are down they stay down.
  const reached = Math.max(...steps.slice(0, idx + 1).map((s) => RANK[s.show] ?? 0));
  const split = reached >= RANK.split;
  const reading = step.show === 'read';
  const focus = step.show === 'field' ? (step.at ?? -1) : -1;

  const cellW = Math.floor((ROW_W - CELL_GAP * (cells.length - 1)) / cells.length);
  const rowW = cellW * cells.length + CELL_GAP * (cells.length - 1);
  const cellH = Math.round(Math.min(cellW * 1.5, 74));
  const bitSize = Math.round(Math.min(cellW * 0.66, 32));

  const xOf = (bit: number) => bit * (cellW + CELL_GAP);
  const widthOf = (from: number, to: number) => (to - from + 1) * cellW + (to - from) * CELL_GAP;

  // Three semantic accents, cycled so two neighbouring fields never share one.
  const bandColour = (i: number): string =>
    [theme.accentQuantity, theme.accentRival, theme.accent][i % 3];

  const focusPop =
    focus >= 0 ? spring({frame: sinceStep - 2, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 26}) : 0;
  const splitIn = split
    ? interpolate(step.show === 'split' ? sinceStep : 99, [2, 22], [0, 1], {
        extrapolateLeft: 'clamp',
        extrapolateRight: 'clamp',
      })
    : 0;

  const fieldOf = (bit: number): number => fields.findIndex((f) => bit >= f.from && bit <= f.to);

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

      <div style={{width: rowW, position: 'relative'}}>
        {/* The row itself. It never moves. */}
        <div style={{position: 'relative', height: cellH}}>
          {cells.map((c, i) => {
            const owner = fieldOf(i);
            const lit = split && owner >= 0;
            const isFocus = owner === focus && focus >= 0;
            const scale = isFocus ? 1 + 0.08 * focusPop : 1;
            return (
              <div
                key={i}
                style={{
                  position: 'absolute',
                  left: xOf(i),
                  top: 0,
                  width: cellW,
                  height: cellH,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  borderRadius: 6,
                  fontFamily: theme.fontMono,
                  fontSize: bitSize,
                  fontWeight: 700,
                  color: isFocus || reading ? theme.text : lit ? theme.text : theme.textMuted,
                  background: isFocus
                    ? withAlpha(bandColour(owner), 0.22)
                    : withAlpha(theme.surface, 0.8),
                  border: `1px solid ${isFocus ? bandColour(owner) : theme.surfaceBorder}`,
                  transform: `scale(${scale})`,
                  opacity: focus >= 0 && !isFocus ? 0.42 : 1,
                }}
              >
                {c.bit}
              </div>
            );
          })}
        </div>

        {/* The bands and brackets: the structure the row turns out to have. */}
        <div style={{position: 'relative', height: BAND_H + BRACKET_H + 54, marginTop: 12}}>
          {fields.map((f, i) => {
            const colour = bandColour(i);
            const isFocus = i === focus;
            const grow = split ? splitIn : 0;
            const w = widthOf(f.from, f.to);
            return (
              <div key={i} style={{position: 'absolute', left: xOf(f.from), top: 0, width: w}}>
                <div
                  style={{
                    width: w * grow,
                    height: isFocus ? BAND_H + 4 : BAND_H,
                    borderRadius: 4,
                    background: colour,
                    opacity: focus >= 0 && !isFocus ? 0.3 : 1,
                    // The one glow: the band under the field being explained.
                    boxShadow: isFocus ? `0 0 26px ${withAlpha(colour, 0.55)}` : undefined,
                  }}
                />
                <div
                  style={{
                    width: w,
                    height: BRACKET_H,
                    borderLeft: `2px solid ${withAlpha(colour, grow)}`,
                    borderRight: `2px solid ${withAlpha(colour, grow)}`,
                    borderBottom: `2px solid ${withAlpha(colour, grow)}`,
                    borderRadius: '0 0 6px 6px',
                    marginTop: 6,
                  }}
                />
                <div
                  style={{
                    width: w,
                    marginTop: 10,
                    textAlign: 'center',
                    fontFamily: theme.fontBody,
                    fontSize: 24,
                    fontWeight: 600,
                    letterSpacing: 0.6,
                    color: isFocus ? colour : theme.textMuted,
                    opacity: grow,
                  }}
                >
                  {f.label}
                </div>
              </div>
            );
          })}
        </div>

        {/* The reserved band. A meaning lands here; nothing above it moves. */}
        <div style={{height: MEANS_H, marginTop: 14, display: 'flex', alignItems: 'flex-start'}}>
          {focus >= 0 ? (
            <div
              style={{
                display: 'flex',
                alignItems: 'baseline',
                gap: 24,
                opacity: focusPop,
                transform: `translateY(${(1 - focusPop) * 14}px)`,
              }}
            >
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 40,
                  fontWeight: 700,
                  color: bandColour(focus),
                }}
              >
                {fields[focus].bits}
              </span>
              <span style={{fontFamily: theme.fontMono, fontSize: 28, color: theme.textMuted}}>
                = {fields[focus].value}
              </span>
              <span style={{fontFamily: theme.fontBody, fontSize: 32, color: theme.text}}>
                {fields[focus].means}
              </span>
            </div>
          ) : reading ? (
            <div
              style={{
                display: 'flex',
                flexWrap: 'wrap',
                alignItems: 'baseline',
                gap: 28,
                opacity: interpolate(sinceStep, [2, 20], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                }),
              }}
            >
              {fields.map((f, i) => (
                <span key={i} style={{display: 'flex', alignItems: 'baseline', gap: 10}}>
                  <span
                    style={{
                      fontFamily: theme.fontBody,
                      fontSize: 26,
                      color: theme.textMuted,
                    }}
                  >
                    {f.label}
                  </span>
                  <span
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: 40,
                      fontWeight: 700,
                      color: bandColour(i),
                    }}
                  >
                    {f.value}
                  </span>
                </span>
              ))}
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
