import {spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// MachineScene: inside the box, one part at a time.
//
// The frame is a chassis plate holding stylized part blocks, and it never
// leaves the screen: the whole clip is the same machine with one part or
// another pulled toward the camera. That persistence is the argument — a part
// explained on its own is a vocabulary card, a part lifted out of a box the
// viewer has watched since the first frame is a component of something.
//
// The depth is faked flat, on purpose. Real isometric projection would tilt
// the text with the blocks, and a label a viewer has to read at an angle is a
// label that loses to the voiceover. Instead every block carries a solid
// offset shadow — the ink of the stage showing under its lower-right edge —
// which reads as thickness without rotating a single glyph, and the focused
// part "lifts" by translating up-left while its shadow grows: the block moves
// toward the camera and its shadow stays on the plate, which is exactly what
// a real object does.
//
// Layout is computed here, deterministically, rather than left to flexbox:
// the connector from the lifted part down to its job caption needs the
// block's real coordinates, and a wrap layout does not surrender them. Blocks
// pack greedily into rows by their declared size weight, so a large part is
// visibly the big one — the box reads as a machine, not a grid of equal
// rectangles.
//
// One glow maximum: the lifted part's accent edge. Everything else is opacity
// and offset — dim to 40% for the parts waiting their turn, full and lit for
// the closing refit.

const CHASSIS_W = Math.min(STAGE_W, 1180);
const CHASSIS_PAD = 30;
const INNER_W = CHASSIS_W - CHASSIS_PAD * 2;
const BLOCK_GAP = 16;
const CAPTION_GAP = 56;

type Part = {label: string; job: string; size: 'large' | 'medium' | 'small'};
type Step = {
  startMs: number;
  endMs: number;
  show: 'whole' | 'part' | 'fit';
  at?: number;
  visited: number[];
};

const sizeW = (size: string): number =>
  size === 'large' ? Math.floor(INNER_W * 0.44) : size === 'small' ? Math.floor(INNER_W * 0.2) : Math.floor(INNER_W * 0.29);
const sizeH = (size: string): number => (size === 'large' ? 168 : size === 'small' ? 108 : 134);

type Box = {x: number; y: number; w: number; h: number};

// Greedy row packing: blocks keep their plan order, a row takes blocks until
// the next one would not fit, and every row is centred in the chassis.
const packParts = (parts: Part[]): {boxes: Box[]; height: number} => {
  const boxes: Box[] = [];
  let row: number[] = [];
  let rowW = 0;
  let y = CHASSIS_PAD;
  const flush = () => {
    if (row.length === 0) return;
    const used = rowW - BLOCK_GAP;
    let x = CHASSIS_PAD + (INNER_W - used) / 2;
    const rowH = Math.max(...row.map((i) => sizeH(parts[i].size)));
    for (const i of row) {
      const w = sizeW(parts[i].size);
      const h = sizeH(parts[i].size);
      boxes[i] = {x, y: y + (rowH - h) / 2, w, h};
      x += w + BLOCK_GAP;
    }
    y += rowH + BLOCK_GAP;
    row = [];
    rowW = 0;
  };
  parts.forEach((p, i) => {
    const w = sizeW(p.size);
    if (rowW + w > INNER_W + 1) flush();
    row.push(i);
    rowW += w + BLOCK_GAP;
  });
  flush();
  return {boxes, height: y - BLOCK_GAP + CHASSIS_PAD};
};

export const MachineScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const chassis = String(props.chassis ?? '');
  const parts = (Array.isArray(props.parts) ? props.parts : []) as Part[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (parts.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const focused = step.show === 'part' ? step.at ?? -1 : -1;
  const fit = step.show === 'fit';
  const visited = new Set(step.visited ?? []);

  // The lift, sprung so the part arrives rather than teleports; the same value
  // fades the job caption in underneath it.
  const lift = spring({frame: sinceStep - 2, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 26});
  // The refit settles everything at once, slightly softer.
  const settle = fit
    ? spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 28})
    : 0;

  const {boxes, height: chassisH} = packParts(parts);
  const focusBox = focused >= 0 ? boxes[focused] : null;
  const captionY = chassisH + CAPTION_GAP;

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={30}
      />

      <div style={{position: 'relative', width: CHASSIS_W, height: captionY + 120}}>
        {/* The chassis plate. Its offset shadow is the "thickness" of the box. */}
        <div
          style={{
            position: 'absolute',
            left: 0,
            top: 0,
            width: CHASSIS_W,
            height: chassisH,
            borderRadius: 22,
            background: withAlpha(theme.surface, 0.55),
            border: `2px solid ${fit ? withAlpha(theme.accent, 0.35 + 0.35 * settle) : theme.surfaceBorder}`,
            boxShadow: `12px 14px 0 ${withAlpha(theme.ink, 0.4)}`,
          }}
        />

        {/* The chassis nameplate, riding the top edge. */}
        <div
          style={{
            position: 'absolute',
            left: CHASSIS_PAD,
            top: -14,
            paddingInline: 14,
            paddingBlock: 4,
            borderRadius: 7,
            background: theme.surface,
            border: `1.5px solid ${theme.surfaceBorder}`,
            fontFamily: theme.fontMono,
            fontSize: 17,
            letterSpacing: 2.6,
            textTransform: 'uppercase',
            color: theme.textMuted,
          }}
        >
          {chassis}
        </div>

        {/* The connector: from the lifted part's underside to its job line. */}
        {focusBox ? (
          <svg
            width={CHASSIS_W}
            height={captionY + 10}
            style={{position: 'absolute', left: 0, top: 0, overflow: 'visible', pointerEvents: 'none'}}
          >
            <line
              x1={focusBox.x + focusBox.w / 2}
              y1={focusBox.y + focusBox.h - 6}
              x2={CHASSIS_W / 2}
              y2={captionY - 6}
              stroke={withAlpha(theme.accent, 0.75 * lift)}
              strokeWidth={2.5}
              strokeDasharray="2 7"
              strokeLinecap="round"
            />
            <circle cx={CHASSIS_W / 2} cy={captionY - 6} r={4.5} fill={withAlpha(theme.accent, lift)} />
          </svg>
        ) : null}

        {/* The parts. */}
        {parts.map((part, i) => {
          const box = boxes[i];
          const isFocus = i === focused;
          const wasVisited = visited.has(i) && !isFocus;
          // Dimmed while waiting, full when lifted or refit; a visited part
          // keeps a little more light than an unvisited one, so progress reads.
          const dimmed = fit ? 1 : isFocus ? 1 : wasVisited ? 0.62 : step.show === 'whole' ? 0.55 : 0.4;
          const rise = isFocus ? lift : 0;
          const pop = fit ? 1 + 0.015 * settle : 1 + 0.08 * rise;
          const edge = isFocus
            ? withAlpha(theme.accent, 0.5 + 0.5 * rise)
            : fit
              ? withAlpha(theme.accent, 0.3 + 0.25 * settle)
              : wasVisited
                ? withAlpha(theme.accent, 0.3)
                : theme.surfaceBorder;
          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: box.x,
                top: box.y,
                width: box.w,
                height: box.h,
                borderRadius: 14,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: isFocus ? withAlpha(theme.accent, 0.12) : withAlpha(theme.mass, 0.08),
                border: `2px solid ${edge}`,
                opacity: dimmed,
                transform: `translate(${-6 * rise}px, ${-16 * rise}px) scale(${pop})`,
                // The block moves toward the camera; its shadow stays on the
                // plate, so the offset grows with the lift.
                boxShadow: isFocus
                  ? `${6 + 8 * rise}px ${7 + 10 * rise}px 0 ${withAlpha(theme.ink, 0.45)}, 0 0 ${30 * rise}px ${withAlpha(theme.accent, 0.3)}`
                  : `6px 7px 0 ${withAlpha(theme.ink, 0.35)}`,
                zIndex: isFocus ? 2 : 1,
              }}
            >
              <span
                style={{
                  fontFamily: theme.fontDisplay,
                  fontSize: part.size === 'large' ? 34 : part.size === 'small' ? 22 : 27,
                  fontWeight: 700,
                  letterSpacing: -0.4,
                  color: isFocus || fit ? theme.text : theme.textMuted,
                  textAlign: 'center',
                  paddingInline: 12,
                }}
              >
                {part.label}
              </span>
            </div>
          );
        })}

        {/* The job line, or the refit line. Centred under the chassis. */}
        <div
          style={{
            position: 'absolute',
            left: 0,
            top: captionY,
            width: CHASSIS_W,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: 6,
            opacity: focusBox ? lift : fit ? settle : 0,
            transform: `translateY(${(1 - (focusBox ? lift : settle)) * 10}px)`,
          }}
        >
          {focusBox && focused >= 0 ? (
            <>
              <div
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 15,
                  letterSpacing: 3,
                  textTransform: 'uppercase',
                  color: theme.accentText,
                }}
              >
                {parts[focused].label}
              </div>
              <div
                style={{
                  fontFamily: theme.fontBody,
                  fontSize: 30,
                  color: theme.text,
                  textAlign: 'center',
                  maxWidth: CHASSIS_W - 80,
                }}
              >
                {parts[focused].job}
              </div>
            </>
          ) : fit ? (
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 26,
                color: theme.textMuted,
                textAlign: 'center',
              }}
            >
              Every part back in place, one machine.
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
