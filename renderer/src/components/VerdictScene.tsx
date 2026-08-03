import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';

// VerdictScene is two columns of conditions and, at the end, one line alone.
//
// The columns sit side by side for the whole clip rather than arriving one at a
// time, because the claim is that the recommendation and its asterisk are one
// object — a viewer who sees the caveats only after they have been sold the
// call has been sold the call. Both halves are on screen from the first frame;
// what moves is which line is lit.
//
// The two columns are NOT symmetric, and that asymmetry is deliberate. Where it
// holds is a checklist in the quantity colour; where it breaks is fewer lines in
// the limit colour, set slightly tighter. A verdict that gave equal visual
// weight to both would be saying "it's a wash", which is the one thing a
// verdict must never say.
//
// The last beat pushes the columns back and brings the call forward at headline
// size on an empty stage. That frame is what gets screenshotted and quoted, so
// nothing else is on it.

// Full width, and everything below is set larger than it was.
//
// This scene was the sparsest in the catalog: a subject line, three conditions
// and two exceptions came to about 437 of the 952 vertical pixels available —
// 46%, centred, so it read as a small composition floating in a lot of nothing
// rather than as a frame. It was typeset for a stage a third shorter than the one
// it now gets.
//
// Bigger type rather than more padding, because the text reflows: a wider column
// at a larger size takes more lines and fills the height on its own, where extra
// padding would only push the same small block further apart.
const COL_W = STAGE_W;

type Step = {startMs: number; endMs: number; show: 'subject' | 'holds' | 'breaks' | 'call'; at?: number};

export const VerdictScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const subject = String(props.subject ?? '');
  const call = String(props.call ?? '');
  const holds = (Array.isArray(props.holds) ? props.holds : []) as string[];
  const breaks = (Array.isArray(props.breaks) ? props.breaks : []) as string[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // The closing frame: the columns are gone and the ruling has the stage.
  //
  // The columns do not vanish, they LEAVE. Returning a different tree the instant
  // the beat changed was a jump cut in the middle of a shot — the two things the
  // whole scene has been building sat side by side one frame and were simply not
  // there the next, which reads as a rendering fault rather than as an edit. The
  // file header calls this frame the one that gets screenshotted and quoted, and
  // it deserves to be arrived at.
  //
  // So the outgoing columns fade and fall back over the first third of a second
  // while the ruling rises through them. Same duration as the entrance spring, so
  // the hand-off is one movement.
  if (step.show === 'call') {
    const land = spring({
      frame: sinceStep,
      fps,
      config: {damping: 200, mass: 0.8},
      durationInFrames: 22,
    });
    const leave = spring({
      frame: sinceStep,
      fps,
      config: {damping: 200, mass: 0.6},
      durationInFrames: 14,
    });
    return (
      <Stage justify="center">
        {/* The columns on their way out, behind the ruling and pointer-inert.
            Rendered from the same data rather than from a snapshot: what leaves
            the frame is what was on it. */}
        <div
          style={{
            position: 'absolute',
            width: COL_W,
            opacity: (1 - leave) * 0.5,
            transform: `scale(${1 - leave * 0.06})`,
            transformOrigin: 'center center',
          }}
        >
          <div style={{display: 'flex', gap: 76, alignItems: 'flex-start'}}>
            <Column
              theme={theme}
              heading="Where it holds"
              mark="✓"
              color={theme.accentQuantity}
              points={holds}
              lit={-1}
              litP={1}
              dim
              frame={frame}
              flex={1.25}
            />
            <Column
              theme={theme}
              heading="Where it breaks"
              mark="✕"
              color={theme.accentLimit}
              points={breaks}
              lit={-1}
              litP={1}
              dim
              frame={frame}
              flex={1}
            />
          </div>
        </div>
        <div style={{width: COL_W, textAlign: 'center', position: 'relative'}}>
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 24,
              letterSpacing: 6,
              textTransform: 'uppercase',
              color: theme.textMuted,
              marginBottom: 42,
              opacity: land,
            }}
          >
            The verdict
          </div>
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 108,
              fontWeight: 800,
              letterSpacing: -2.4,
              lineHeight: 1.1,
              color: theme.text,
              opacity: land,
              transform: `translateY(${(1 - land) * 44}px)`,
            }}
          >
            {call}
          </div>
        </div>
      </Stage>
    );
  }

  const enter = spring({
    frame: ((nowMs - steps[0].startMs) / 1000) * FPS,
    fps,
    config: {damping: 200, mass: 0.7},
    durationInFrames: 18,
  });

  // Which line is lit, if any. A `subject` beat lights nothing — it is the
  // moment the question is on screen and neither column has been walked yet.
  const litHold = step.show === 'holds' ? (step.at ?? -1) : -1;
  const litBreak = step.show === 'breaks' ? (step.at ?? -1) : -1;
  // Once the clip has moved into the breaks column, the holds column recedes
  // as a block — the argument has moved on and keeping it bright would fight
  // the asterisk for attention.
  const inBreaks = step.show === 'breaks';
  // How far through the current line's own beat we are, so the plate that marks
  // it grows in over a fifth of a second instead of appearing between frames.
  const litP = spring({
    frame: sinceStep,
    fps,
    config: {damping: 200, mass: 0.5},
    durationInFrames: 12,
  });

  return (
    <Stage justify="center">
      <div style={{width: COL_W, opacity: enter}}>
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 22,
            letterSpacing: 5.5,
            textTransform: 'uppercase',
            color: theme.textMuted,
            textAlign: 'center',
            marginBottom: 16,
          }}
        >
          Ruling on
        </div>
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 76,
            fontWeight: 800,
            letterSpacing: -1.8,
            lineHeight: 1.1,
            color: theme.text,
            textAlign: 'center',
            marginBottom: 68,
          }}
        >
          {subject}
        </div>

        <div style={{display: 'flex', gap: 76, alignItems: 'flex-start'}}>
          <Column
            theme={theme}
            heading="Where it holds"
            mark="✓"
            color={theme.accentQuantity}
            points={holds}
            lit={litHold}
            litP={litP}
            dim={inBreaks}
            frame={frame}
            // The wider half: this is the body of the advice.
            flex={1.25}
          />
          <Column
            theme={theme}
            heading="Where it breaks"
            mark="✕"
            color={theme.accentLimit}
            points={breaks}
            lit={litBreak}
            litP={litP}
            dim={false}
            frame={frame}
            flex={1}
          />
        </div>
      </div>
    </Stage>
  );
};

const Column: React.FC<{
  theme: ResolvedTheme;
  heading: string;
  mark: string;
  color: string;
  points: string[];
  /** Index of the line currently being spoken, or -1. */
  lit: number;
  /** 0..1 through the lit line's own beat, so the plate grows in rather than
   *  snapping on. An instant background swap was the last thing in this scene
   *  that still read as a slide advancing: the plate simply existed on one frame
   *  and not the previous one, with no movement to carry the eye to it. */
  litP: number;
  /** Whether the whole column has receded because the clip moved past it. */
  dim: boolean;
  frame: number;
  flex: number;
}> = ({theme, heading, mark, color, points, lit, litP, dim, frame, flex}) => (
  <div style={{flex, opacity: dim ? 0.4 : 1}}>
    <div
      style={{
        fontFamily: theme.fontMono,
        fontSize: 21,
        letterSpacing: 4,
        textTransform: 'uppercase',
        color,
        marginBottom: 28,
      }}
    >
      {heading}
    </div>
    {points.map((p, i) => {
      const on = interpolate(frame, [4 + i * 4, 18 + i * 4], [0, 1], {
        extrapolateLeft: 'clamp',
        extrapolateRight: 'clamp',
      });
      const isLit = i === lit;
      return (
        <div
          key={i}
          style={{
            display: 'flex',
            alignItems: 'flex-start',
            gap: 20,
            padding: '24px 26px',
            marginBottom: 18,
            borderRadius: 14,
            // The lit line gets a tinted plate rather than a colour change, so
            // the text stays the same weight and only the ground moves.
            background: isLit ? withAlpha(color, 0.13 * litP) : 'transparent',
            border: `1px solid ${isLit ? withAlpha(color, 0.4 * litP) : 'transparent'}`,
            opacity: on * (isLit || lit < 0 ? 1 : 0.55),
            // The lit line also steps very slightly toward the viewer. Two pixels
            // and a percent of scale is not something anybody sees happening; it
            // is what stops the change reading as a static diff between slides.
            transform: `translateY(${(1 - on) * 10 - (isLit ? litP * 2 : 0)}px) scale(${
              isLit ? 1 + litP * 0.012 : 1
            })`,
            transformOrigin: 'left center',
          }}
        >
          <div style={{fontFamily: theme.fontMono, fontSize: 28, color, lineHeight: 1.35}}>{mark}</div>
          <div
            style={{
              fontFamily: theme.fontBody,
              fontSize: 36,
              lineHeight: 1.35,
              color: isLit ? theme.text : theme.textMuted,
            }}
          >
            {p}
          </div>
        </div>
      );
    })}
  </div>
);
