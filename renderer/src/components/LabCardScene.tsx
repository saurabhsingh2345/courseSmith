import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W, STAGE_H} from './Stage';
import {SceneHeader} from './SceneHeader';

// LabCardScene: the briefing a learner keeps open while their hands are busy.
//
// The composition is two zones on one card, because a lab has two kinds of
// information and they are read at different moments. On the left, the
// BRIEFING: the task set large enough to read from across a room, and under it
// the tools as monospaced chips, because a tool is a thing you type or install
// and mono is the typeface that says so. On the right, the WORK: the numbered
// steps, all of them, all the time.
//
// "All of them, all the time" is the whole design. A lab card that reveals its
// steps one at a time is a slideshow, and a slideshow is exactly what a learner
// with a terminal open cannot use — they need to see step four while they are
// still on step two, so they know what they are heading towards. So every row
// is present from the first frame, at three levels of emphasis: walked (muted,
// with a tick), current (full contrast, accent rail, the one glow in the
// picture), and ahead (dim, but legible). Nothing appears; only the emphasis
// moves, which is also why the motion here is a rail sliding down a list rather
// than rows animating in.
//
// The zones are divided by a single hairline rather than a box, and the whole
// card sits on the stage without a border, because a card drawn as a rectangle
// inside a frame reads as a slide. The one enclosed element is the expected
// result, and it gets a box on purpose: it is styled as a terminal line with a
// check in front of it, so the answer to "how do I know it worked" looks like
// the thing the learner will actually be staring at.

const CARD_W = Math.min(STAGE_W, 1560);
const BRIEF_W = 520;
const GUTTER = 56;
const BODY_H = STAGE_H - 190;
const ROW_GAP = 14;
const NUM_W = 52;

type Tool = {name: string};
type StepRow = {n: number; text: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'task' | 'step' | 'expect';
  at?: number;
  reached: number[];
};

export const LabCardScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const task = String(props.task ?? '');
  const expect = String(props.expect ?? '');
  const tools = (Array.isArray(props.tools) ? props.tools : []) as Tool[];
  const rows = (Array.isArray(props.stepList) ? props.stepList : []) as StepRow[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (rows.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const current = step.show === 'step' && typeof step.at === 'number' ? step.at : -1;
  const walked = new Set(step.reached ?? []);
  const onExpect = step.show === 'expect';

  // The card itself settles once, at the top of the clip, and then holds.
  const cardIn = interpolate(frame, [0, 18], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const rowH = Math.max(54, Math.min(78, Math.floor((BODY_H - 150) / rows.length) - ROW_GAP));

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

      <div style={{width: CARD_W, opacity: cardIn}}>
        <div style={{display: 'flex', alignItems: 'stretch', gap: GUTTER}}>
          {/* Zone one: the briefing. What you are making, and what you need. */}
          <div style={{width: BRIEF_W, flexShrink: 0, paddingTop: 4}}>
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 15,
                letterSpacing: 2.6,
                textTransform: 'uppercase',
                color: theme.textMuted,
                marginBottom: 16,
              }}
            >
              the lab
            </div>
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 52,
                fontWeight: 700,
                lineHeight: 1.12,
                letterSpacing: -1,
                color: theme.text,
                transform: `translateY(${(1 - cardIn) * 10}px)`,
              }}
            >
              {task}
            </div>

            <div
              style={{
                width: 68,
                height: 3,
                background: theme.accent,
                margin: '30px 0 26px',
                transform: `scaleX(${cardIn})`,
                transformOrigin: 'left center',
              }}
            />

            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 14,
                letterSpacing: 2.4,
                textTransform: 'uppercase',
                color: theme.textMuted,
                marginBottom: 14,
              }}
            >
              you will need
            </div>
            <div style={{display: 'flex', flexWrap: 'wrap', gap: 10}}>
              {tools.map((t, i) => {
                const pop = spring({
                  frame: frame - 8 - i * 5,
                  fps,
                  config: {damping: 13, mass: 0.55},
                  durationInFrames: 24,
                });
                return (
                  <div
                    key={i}
                    style={{
                      padding: '9px 16px',
                      borderRadius: 8,
                      background: withAlpha(theme.surface, 0.9),
                      border: `1px solid ${theme.surfaceBorder}`,
                      fontFamily: theme.fontMono,
                      fontSize: 20,
                      color: theme.text,
                      opacity: pop,
                      transform: `translateY(${(1 - pop) * 8}px)`,
                    }}
                  >
                    {t.name}
                  </div>
                );
              })}
            </div>
          </div>

          {/* The hairline that divides briefing from work. */}
          <div style={{width: 1, background: theme.line, flexShrink: 0}} />

          {/* Zone two: the work. Every row present, emphasis moving. */}
          <div style={{flex: 1, display: 'flex', flexDirection: 'column', gap: ROW_GAP, paddingTop: 6}}>
            {rows.map((r, i) => {
              const isCurrent = i === current;
              const isDone = walked.has(i) && !isCurrent;
              const pop = isCurrent
                ? spring({frame: sinceStep, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 24})
                : 1;
              const railH = isCurrent ? interpolate(pop, [0, 1], [0, rowH]) : 0;
              return (
                <div
                  key={i}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 18,
                    height: rowH,
                    paddingLeft: 18,
                    borderRadius: 10,
                    position: 'relative',
                    background: isCurrent ? withAlpha(theme.accent, 0.1) : 'transparent',
                  }}
                >
                  {/* The rail: the one thing that moves down the list. */}
                  <div
                    style={{
                      position: 'absolute',
                      left: 0,
                      top: (rowH - railH) / 2,
                      width: 3,
                      height: railH,
                      background: theme.accent,
                      borderRadius: 2,
                    }}
                  />
                  <div
                    style={{
                      width: NUM_W,
                      height: NUM_W,
                      flexShrink: 0,
                      borderRadius: 10,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontFamily: theme.fontMono,
                      fontSize: 22,
                      fontWeight: 700,
                      color: isCurrent ? theme.ink : isDone ? theme.accentText : theme.textMuted,
                      background: isCurrent ? theme.accent : withAlpha(theme.surface, isDone ? 0.9 : 0.5),
                      border: `1px solid ${isCurrent ? theme.accent : theme.surfaceBorder}`,
                      transform: `scale(${0.94 + 0.06 * pop})`,
                      // The one glow: the step the learner is on right now.
                      boxShadow: isCurrent ? `0 0 26px ${withAlpha(theme.accent, 0.45)}` : undefined,
                    }}
                  >
                    {isDone ? '✓' : r.n}
                  </div>
                  <div
                    style={{
                      fontFamily: theme.fontBody,
                      fontSize: 28,
                      lineHeight: 1.25,
                      color: isCurrent ? theme.text : isDone ? theme.textMuted : withAlpha(theme.text, 0.42),
                    }}
                  >
                    {r.text}
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* The answer to "how do I know it worked", dressed as a terminal line. */}
        <div
          style={{
            marginTop: 34,
            display: 'flex',
            alignItems: 'center',
            gap: 16,
            padding: '18px 24px',
            borderRadius: 10,
            background: withAlpha(theme.surface, onExpect ? 1 : 0.45),
            border: `1px solid ${onExpect ? withAlpha(theme.accentQuantity, 0.5) : theme.surfaceBorder}`,
            borderLeft: `3px solid ${onExpect ? theme.accentQuantity : theme.surfaceBorder}`,
            opacity: onExpect
              ? interpolate(sinceStep, [0, 14], [0.5, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                })
              : 0.5,
            transform: onExpect
              ? `translateY(${
                  (1 -
                    spring({
                      frame: sinceStep,
                      fps,
                      config: {damping: 14, mass: 0.6},
                      durationInFrames: 26,
                    })) *
                  10
                }px)`
              : 'none',
          }}
        >
          <span
            style={{
              fontFamily: theme.fontMono,
              fontSize: 24,
              color: onExpect ? theme.accentQuantity : theme.textMuted,
            }}
          >
            ✓
          </span>
          <span
            style={{
              fontFamily: theme.fontMono,
              fontSize: 24,
              color: onExpect ? theme.text : theme.textMuted,
            }}
          >
            {expect}
          </span>
        </div>
      </div>
    </Stage>
  );
};
