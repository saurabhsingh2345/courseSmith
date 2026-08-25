import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';
import {CardName, CardPanel, CardTile} from './showroomCard';

// ChoicesScene is a question and the answers, and its whole job is to not say
// which one is right.
//
// Every other card row in this renderer has a lit card and quiet ones, and the
// lighting is how the voice points. Here there is nothing to point at — the
// answer is in the viewer's head — so all the cards are drawn in the same state,
// permanently. That is a constraint on the component rather than on the plan:
// there is no `selected` anywhere below and no branch that could introduce one,
// which is why no future plan can accidentally light one card and hand the answer
// over. See snippet_choices.go.
//
// So the motion has to do the work the lighting normally does. Two things move.
//
// The question rises alone and then moves up out of the middle when the row
// arrives, which is the frame reorganising itself around the answers — the
// closest thing to a camera move available on a still composition, and the only
// signal that the clip has moved on to a new beat.
//
// The cards arrive on a spring, staggered left to right, and the stagger is
// timed to the voice reading the list. A row that appears all at once is read as
// a block, and a block of four names is a paragraph. Arriving one at a time makes
// the eye land on each name in the order it is spoken, which is the only pointing
// this frame is allowed to do.
//
// Nothing pulses, nothing counts down, nothing recedes. A held beat here is
// genuinely held: the shape of the pause is the viewer thinking, and an animation
// running through it is the frame telling them to hurry.

const BLOCK_W = Math.min(STAGE_W, 1560);

/** Card metrics: fewer options get a wider measure and a bigger mark. */
const metrics = (n: number) =>
  n <= 3
    ? {tile: 148, label: 50, pad: 52, gap: 44, minH: 400}
    : n === 4
      ? {tile: 132, label: 44, pad: 44, gap: 36, minH: 380}
      : {tile: 112, label: 38, pad: 34, gap: 28, minH: 340};

type Option = {label: string; icon?: string; mark?: string; tint?: string; image?: string};

type Step = {startMs: number; endMs: number; show: 'ask' | 'options' | 'hold'};

export const ChoicesScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const options = (Array.isArray(props.options) ? props.options : []) as Option[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (options.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;

  // Which beat is current never has to be worked out, and that is a property of
  // the template rather than a shortcut: the only thing that moves is whether
  // the row is up, so the row's own beat is the only timestamp needed. A held
  // beat is the same frame as the one before it, not a smaller one.
  const rowAt = steps.findIndex((s) => s.show === 'options');
  const rowStartMs = rowAt >= 0 ? steps[rowAt].startMs : Infinity;
  const sinceRow = ((nowMs - rowStartMs) / 1000) * FPS;
  const rowUp = sinceRow >= 0;

  const m = metrics(options.length);

  // The question's own entrance, and its move out of the middle. Both are
  // measured from the scene rather than from the beat, because this is one
  // continuous frame with one question on it.
  const asked = interpolate(frame, [2, 20], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const lift = rowUp
    ? interpolate(sinceRow, [0, 22], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : 0;

  return (
    <Stage justify="center">
      <div
        style={{
          width: BLOCK_W,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
        }}
      >
        <div style={{opacity: asked, transform: `translateY(${(1 - asked) * 24}px)`}}>
          <SceneHeader
            theme={theme}
            title={String(props.title ?? '')}
            emphasis={props.emphasis as string | undefined}
            emphasisRole={props.emphasisRole as string | undefined}
            marginBottom={56}
          />
        </div>

        {/* The row's height GROWS as the cards land, and that is the camera
            move. Reserving it from the first frame instead leaves the question
            floating above an empty half-frame for the whole asking beat, which
            reads as a missing element rather than as a question. Setting it
            instantly at the row's arrival would snap the question upward. So the
            space the row will occupy opens on the same curve the cards ride in
            on, and the block — centred by the stage — lifts the question by
            exactly half of it. */}
        <div
          style={{
            display: 'flex',
            gap: m.gap,
            width: '100%',
            alignItems: 'stretch',
            justifyContent: 'center',
            height: lift * m.minH,
            overflow: 'hidden',
          }}
        >
          {rowUp
            ? options.map((o, i) => {
                // Staggered left to right, on the voice reading the list. Four
                // frames apart is about an eighth of a second — enough for the
                // eye to land on each name in turn, short enough that the row
                // still reads as one object by the end of the sentence.
                const on = spring({
                  frame: sinceRow - i * 4,
                  fps,
                  config: {damping: 200, mass: 0.7},
                });
                return (
                  <CardPanel
                    key={i}
                    theme={theme}
                    colour={theme.accent}
                    // Never selected. Every card is drawn in the same state for
                    // the whole clip — see the header. A row where one card is
                    // rimmed in the accent has been answered before it was read.
                    selected={false}
                    style={{
                      flex: 1,
                      minHeight: m.minH,
                      padding: m.pad,
                      gap: 24,
                      justifyContent: 'center',
                      alignItems: 'center',
                      textAlign: 'center',
                      opacity: on,
                      transform: `translateY(${(1 - on) * 26}px)`,
                    }}
                  >
                    <CardTile
                      theme={theme}
                      subject={{title: o.label, icon: o.icon, mark: o.mark, tint: o.tint, image: o.image}}
                      size={m.tile}
                      lit
                    />
                    <CardName theme={theme} size={m.label} lit>
                      {o.label}
                    </CardName>
                  </CardPanel>
                );
              })
            : null}
        </div>
      </div>
    </Stage>
  );
};
