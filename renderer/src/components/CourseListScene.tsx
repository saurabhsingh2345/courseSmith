import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// The shared picture behind four of the five v5 course-scaffolding scenes.
//
// `objective`, `prereq`, `recap` and `checkpoint` all draw the same thing: a
// short list where one row at a time becomes the subject and the ones already
// spoken stay on screen, dimmed, so the closing frame holds the whole set at
// once. Only the row's ANATOMY differs — an outcome carries its evidence, an
// assumption carries where it came from, a claim carries its lesson.
//
// Writing that four times would have produced four slightly different rhythms
// and four places to fix a spacing bug, which is exactly how a catalog of
// forty-four templates ends up looking like one template with different middles.
// The row is a render prop instead: each scene decides what a row SAYS, and this
// decides how a list behaves.
//
// What "behaves" means here, and it is the whole reason these are not bullets:
// - A row that has not been spoken yet is not drawn. The list builds.
// - The row being spoken is full-strength and sprung in from the left.
// - Rows already spoken persist at reduced weight. They are context, not noise.
// - The final "everything" beat brings them all to equal weight — that is the
//   frame somebody screenshots, and it is the only one where the list is a set
//   rather than a sequence.

export const LIST_W = Math.min(STAGE_W, 1240);

export type ListStep = {
  startMs: number;
  endMs: number;
  show: string;
  at?: number;
  /** Indices revealed so far. Accumulates; the scene's Go half computes it. */
  lit?: number[];
  ticked?: number[];
};

/** Which beat kinds mean "show the whole set at equal weight". */
const WHOLE_SET = new Set(['contract', 'floor', 'standing', 'done']);

export const CourseListScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
  /** How many rows the list has. */
  count: number;
  /** Draws one row. `weight` is 1 for the row being spoken, lower for settled. */
  row: (i: number, weight: number, focused: boolean) => React.ReactNode;
  /** Optional line under the header — the audience, the thread, the task. */
  lead?: string;
  /** Optional line under the list — the done-when, the consequence. */
  tail?: React.ReactNode;
}> = ({theme, sceneStartMs, props, count, row, lead, tail}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const steps = (Array.isArray(props.steps) ? props.steps : []) as ListStep[];
  if (steps.length === 0 || count === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const revealed = new Set(step.lit ?? step.ticked ?? []);
  const whole = WHOLE_SET.has(step.show);
  const focus = step.at;

  // The lead settles once and stays. It is the frame's premise, so it must not
  // re-animate every time a row lands — that reads as the premise changing.
  const leadIn = interpolate(
    ((nowMs - steps[0].startMs) / 1000) * FPS,
    [0, 14],
    [0, 1],
    {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
  );

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={lead ? 16 : 30}
      />

      {lead ? (
        <div
          style={{
            fontFamily: theme.fontBody,
            fontSize: 28,
            color: theme.textMuted,
            marginBottom: 34,
            maxWidth: LIST_W,
            opacity: leadIn,
          }}
        >
          {lead}
        </div>
      ) : null}

      <div style={{display: 'flex', flexDirection: 'column', gap: 18, width: LIST_W}}>
        {Array.from({length: count}, (_, i) => {
          if (!revealed.has(i) && !whole) return null;
          const focused = !whole && focus === i;
          // Sprung from the left on the beat that lands it, and only then.
          const arrive = focused
            ? spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.5}, durationInFrames: 22})
            : 1;
          // Settled rows hold at 0.5 rather than fading out: they are what the
          // viewer has already been given, and the last frame is the whole set.
          const weight = whole ? 0.92 : focused ? 1 : 0.5;
          return (
            <div
              key={i}
              style={{
                opacity: arrive * (0.25 + 0.75 * weight),
                transform: `translateX(${(1 - arrive) * -28}px)`,
              }}
            >
              {row(i, weight, focused)}
            </div>
          );
        })}
      </div>

      {tail ? <div style={{marginTop: 34, width: LIST_W}}>{tail}</div> : null}
    </Stage>
  );
};

/**
 * The standard row: an index rule, a strong line, and a quieter second line.
 *
 * Shared so the four scenes agree on where a row's weight sits. The accent rule
 * on the left is the only thing that changes colour — it is what makes the
 * focused row read as the subject without the type having to shout.
 */
export const CourseRow: React.FC<{
  theme: ResolvedTheme;
  weight: number;
  focused: boolean;
  /** The row's own accent. Defaults to the brand accent. */
  accent?: string;
  primary: string;
  secondary?: string;
  /** Small label above the secondary line — "from", "proof", "where". */
  tag?: string;
  /** Struck through, for a step already done or an assumption being skipped. */
  struck?: boolean;
}> = ({theme, weight, focused, accent, primary, secondary, tag, struck}) => {
  const bar = accent ?? theme.accentText;
  return (
    <div style={{display: 'flex', gap: 22, alignItems: 'stretch'}}>
      <div
        style={{
          width: focused ? 7 : 4,
          borderRadius: 4,
          flexShrink: 0,
          background: focused ? bar : withAlpha(theme.textMuted, 0.45),
        }}
      />
      <div style={{display: 'flex', flexDirection: 'column', gap: 6, paddingTop: 2}}>
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 40,
            fontWeight: 700,
            lineHeight: 1.15,
            letterSpacing: -0.6,
            color: theme.text,
            textDecoration: struck ? 'line-through' : undefined,
            textDecorationColor: struck ? withAlpha(theme.textMuted, 0.7) : undefined,
          }}
        >
          {primary}
        </div>
        {secondary ? (
          <div style={{display: 'flex', gap: 12, alignItems: 'baseline'}}>
            {tag ? (
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 15,
                  letterSpacing: 2.5,
                  textTransform: 'uppercase',
                  color: withAlpha(bar, 0.85),
                  flexShrink: 0,
                }}
              >
                {tag}
              </span>
            ) : null}
            <span style={{fontFamily: theme.fontBody, fontSize: 25, color: theme.textMuted, lineHeight: 1.3}}>
              {secondary}
            </span>
          </div>
        ) : null}
      </div>
    </div>
  );
};
