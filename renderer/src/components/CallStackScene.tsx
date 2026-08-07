import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W, STAGE_H} from './Stage';
import {SceneHeader} from './SceneHeader';

// CallStackScene: recursion drawn as a stack that actually stacks.
//
// The whole composition rests on one decision: frames are plates on a floor,
// and the floor never moves. A new call springs in from the right and settles
// ON TOP of the plates already there, so the pile grows upward and the
// outermost call stays pinned to the ground for the entire clip. That is the
// fact a printed trace cannot show — the first call is still sitting there,
// frozen, underneath everything — and it is worth spending the whole layout on.
//
// The unwind is the mirror image and it is deliberately physical. When a frame
// pops, its plate slides off to the right while a chip carrying its return
// value falls into the slot below. Values move DOWN and calls move UP, always,
// so the two halves of the clip read as opposite motions rather than as the
// same list changing colour. When the outermost frame finally returns there is
// no plate below to catch the chip, so it keeps falling to the middle of the
// stage and becomes the answer — which is the one frame somebody screenshots.
//
// Depth is marked twice, cheaply: a vertical axis on the left rail that the
// plates hang off, and a mono "#n" on each plate. The base case gets a caption
// card beside the top plate rather than under the stack, because it is a claim
// about that one frame ("this call returns without calling again") and not a
// note on the picture.
//
// Colour is doing semantic work, not decoration. Argument chips are quantities
// (accentQuantity) because they are what the call was asked; return values are
// the accent because they are what it answers. Nothing else is tinted, so the
// two colours stay legible as a grammar. One glow only: the plate landing right
// now during a call.

const RAIL_W = 190;
const PLATE_W = 520;
const CAPTION_GAP = 40;
const CAPTION_W = 370;
const BODY_W = Math.min(STAGE_W, RAIL_W + PLATE_W + CAPTION_GAP + CAPTION_W);
const PLATE_H = 102;
const PLATE_GAP = 14;
const SLOT = PLATE_H + PLATE_GAP;
const BODY_H = Math.min(STAGE_H - 230, SLOT * 5 + 70);

type Frame = {args: string; returns: string; base: boolean};
type Step = {
  startMs: number;
  endMs: number;
  show: 'call' | 'base' | 'return' | 'empty';
  at?: number;
  value?: string;
  into?: number;
  onStack: number[];
  returned: number[];
};

export const CallStackScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const fn = String(props.fn ?? '');
  const base = String(props.base ?? '');
  const answer = String(props.answer ?? '');
  const frames = (Array.isArray(props.frames) ? props.frames : []) as Frame[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (frames.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const live = new Set(Array.isArray(step.onStack) ? step.onStack : []);
  const gone = new Set(Array.isArray(step.returned) ? step.returned : []);
  const popping = step.show === 'return' ? (step.at ?? -1) : -1;
  const landing = step.show === 'call' ? (step.at ?? -1) : -1;
  const baseShown = step.show === 'base' || live.has(frames.length - 1) || gone.has(frames.length - 1);
  const emptied = step.show === 'empty';

  // The bottom of plate i, measured from the floor. Slots are fixed, so a
  // plate never moves once it has landed.
  const slotBottom = (i: number) => i * SLOT;

  // The pop is one continuous motion across the beat: the plate leaves to the
  // right while the chip falls a slot.
  const exit = interpolate(sinceStep, [8, 30], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const drop = interpolate(sinceStep, [4, 28], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const chipFrom = popping >= 0 ? slotBottom(popping) : 0;
  const chipTo =
    step.into === undefined
      ? // Nothing below the outermost frame: the value keeps falling to the
        // middle of the stage and becomes the answer.
        Math.round(BODY_H / 2 - PLATE_H / 2)
      : slotBottom(step.into);
  const chipBottom = chipFrom + (chipTo - chipFrom) * drop;

  const answerIn = emptied
    ? spring({frame: sinceStep - 4, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 26})
    : 0;

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

      <div style={{width: BODY_W, height: BODY_H, position: 'relative'}}>
        {/* The rail: what this pile is, and the axis it hangs off. */}
        <div
          style={{
            position: 'absolute',
            left: 0,
            bottom: 0,
            width: RAIL_W,
            height: BODY_H,
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'flex-end',
            alignItems: 'flex-start',
            paddingBottom: 6,
          }}
        >
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 15,
              letterSpacing: 2.6,
              textTransform: 'uppercase',
              color: theme.textMuted,
              marginBottom: 10,
            }}
          >
            call stack
          </div>
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 30,
              fontWeight: 700,
              color: theme.text,
              letterSpacing: -0.5,
            }}
          >
            {fn}()
          </div>
        </div>

        {/* The axis and the floor: the stack stands on something. */}
        <div
          style={{
            position: 'absolute',
            left: RAIL_W - 26,
            bottom: 0,
            width: 2,
            height: BODY_H - 40,
            background: `linear-gradient(0deg, ${withAlpha(theme.line, 0.5)}, ${withAlpha(theme.line, 0)})`,
          }}
        />
        <div
          style={{
            position: 'absolute',
            left: RAIL_W - 26,
            bottom: 0,
            width: BODY_W - RAIL_W + 26,
            height: 2,
            background: withAlpha(theme.line, emptied ? 0.2 : 0.5),
          }}
        />

        {/* The plates. Fixed slots, so landing never shifts what is already up. */}
        {frames.map((f, i) => {
          const isLive = live.has(i);
          const isPopping = i === popping;
          if (!isLive && !isPopping) return null;

          const land =
            i === landing
              ? spring({frame: sinceStep - 3, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 26})
              : 1;
          const leave = isPopping ? exit : 0;
          const isBase = f.base;

          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: RAIL_W,
                bottom: slotBottom(i),
                width: PLATE_W,
                height: PLATE_H,
                display: 'flex',
                alignItems: 'center',
                gap: 20,
                paddingLeft: 22,
                paddingRight: 22,
                borderRadius: 14,
                background: withAlpha(theme.surface, isBase ? 0.98 : 0.82),
                border: `2px solid ${isBase ? withAlpha(theme.accent, 0.65) : theme.surfaceBorder}`,
                opacity: (1 - leave) * (i === landing ? land : 1),
                transform: `translateX(${(1 - land) * 90 + leave * 130}px)`,
                // The one glow: the frame arriving right now.
                boxShadow: i === landing && land < 0.999 ? `0 0 34px ${withAlpha(theme.accent, 0.35)}` : undefined,
              }}
            >
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 15,
                  letterSpacing: 1.6,
                  color: theme.textMuted,
                  width: 34,
                  flexShrink: 0,
                }}
              >
                #{i}
              </span>
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 24,
                  color: theme.textMuted,
                }}
              >
                {fn}
              </span>
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 34,
                  fontWeight: 700,
                  letterSpacing: -0.5,
                  color: theme.accentQuantity,
                }}
              >
                {f.args}
              </span>
              <span
                style={{
                  marginLeft: 'auto',
                  fontFamily: theme.fontBody,
                  fontSize: 17,
                  letterSpacing: 2,
                  textTransform: 'uppercase',
                  color: isBase ? theme.accentText : withAlpha(theme.textMuted, 0.7),
                }}
              >
                {isBase ? 'base case' : 'waiting'}
              </span>
            </div>
          );
        })}

        {/* The return chip: a value falling one slot, always downward. */}
        {popping >= 0 && step.value ? (
          <div
            style={{
              position: 'absolute',
              left: RAIL_W + PLATE_W - 150,
              bottom: chipBottom + PLATE_H / 2 - 27,
              width: 150,
              height: 54,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              borderRadius: 27,
              background: theme.accent,
              color: theme.ink,
              fontFamily: theme.fontMono,
              fontSize: 30,
              fontWeight: 700,
              transform: `scale(${0.86 + 0.14 * drop})`,
              opacity: interpolate(sinceStep, [2, 10], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            {step.value}
          </div>
        ) : null}

        {/* The base-case caption: a claim about the top plate, set beside it. */}
        {baseShown && base ? (
          <div
            style={{
              position: 'absolute',
              left: RAIL_W + PLATE_W + CAPTION_GAP,
              bottom: slotBottom(frames.length - 1) + 4,
              width: CAPTION_W,
              paddingLeft: 20,
              borderLeft: `3px solid ${theme.accent}`,
              opacity: interpolate(step.show === 'base' ? sinceStep : 99, [4, 20], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 15,
                letterSpacing: 2.4,
                textTransform: 'uppercase',
                color: theme.accentText,
                marginBottom: 8,
              }}
            >
              why it stops
            </div>
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 27,
                lineHeight: 1.3,
                color: theme.text,
              }}
            >
              {base}
            </div>
          </div>
        ) : null}

        {/* The closer: the pile gone, the answer alone. */}
        {emptied ? (
          <div
            style={{
              position: 'absolute',
              left: RAIL_W,
              bottom: Math.round(BODY_H / 2 - PLATE_H / 2),
              width: BODY_W - RAIL_W,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              opacity: answerIn,
              transform: `scale(${0.9 + 0.1 * answerIn})`,
            }}
          >
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 18,
                letterSpacing: 3,
                textTransform: 'uppercase',
                color: theme.textMuted,
                marginBottom: 12,
              }}
            >
              {fn}({frames[0].args.replace(/^[^=]*=/, '')}) =
            </div>
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 118,
                fontWeight: 700,
                letterSpacing: -3,
                lineHeight: 1,
                color: theme.accentText,
              }}
            >
              {answer}
            </div>
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
