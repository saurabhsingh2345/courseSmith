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

const COL_W = Math.min(STAGE_W, 1560);

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
  if (step.show === 'call') {
    const land = spring({
      frame: sinceStep,
      fps,
      config: {damping: 200, mass: 0.8},
      durationInFrames: 22,
    });
    return (
      <Stage justify="center">
        <div style={{width: COL_W, textAlign: 'center'}}>
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 20,
              letterSpacing: 5.5,
              textTransform: 'uppercase',
              color: theme.textMuted,
              marginBottom: 34,
              opacity: land,
            }}
          >
            The verdict
          </div>
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 84,
              fontWeight: 800,
              letterSpacing: -1.8,
              lineHeight: 1.1,
              color: theme.text,
              opacity: land,
              transform: `translateY(${(1 - land) * 20}px)`,
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

  return (
    <Stage justify="center">
      <div style={{width: COL_W, opacity: enter}}>
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 19,
            letterSpacing: 5,
            textTransform: 'uppercase',
            color: theme.textMuted,
            textAlign: 'center',
            marginBottom: 12,
          }}
        >
          Ruling on
        </div>
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 54,
            fontWeight: 800,
            letterSpacing: -1,
            color: theme.text,
            textAlign: 'center',
            marginBottom: 58,
          }}
        >
          {subject}
        </div>

        <div style={{display: 'flex', gap: 60, alignItems: 'flex-start'}}>
          <Column
            theme={theme}
            heading="Where it holds"
            mark="✓"
            color={theme.accentQuantity}
            points={holds}
            lit={litHold}
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
  /** Whether the whole column has receded because the clip moved past it. */
  dim: boolean;
  frame: number;
  flex: number;
}> = ({theme, heading, mark, color, points, lit, dim, frame, flex}) => (
  <div style={{flex, opacity: dim ? 0.4 : 1}}>
    <div
      style={{
        fontFamily: theme.fontMono,
        fontSize: 17,
        letterSpacing: 3.6,
        textTransform: 'uppercase',
        color,
        marginBottom: 22,
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
            gap: 16,
            padding: '16px 20px',
            marginBottom: 12,
            borderRadius: 10,
            // The lit line gets a tinted plate rather than a colour change, so
            // the text stays the same weight and only the ground moves.
            background: isLit ? withAlpha(color, 0.13) : 'transparent',
            border: `1px solid ${isLit ? withAlpha(color, 0.4) : 'transparent'}`,
            opacity: on * (isLit || lit < 0 ? 1 : 0.55),
            transform: `translateY(${(1 - on) * 10}px)`,
          }}
        >
          <div style={{fontFamily: theme.fontMono, fontSize: 22, color, lineHeight: 1.35}}>{mark}</div>
          <div
            style={{
              fontFamily: theme.fontBody,
              fontSize: 27,
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
