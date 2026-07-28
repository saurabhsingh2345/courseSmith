import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_W} from './Stage';

// SpecScene is one card: a goal, its constraints, and a checklist.
//
// The clip is in two halves and the scene only has to keep them apart. While the
// list is being written every box is empty, because a box that ticked as its
// line was written would be showing a build that happened to be described
// afterwards — the exact habit the template argues against. Then one beat checks
// the lot, in a cascade rather than all at once, so the eye can follow the list
// down and the misses land as interruptions rather than as a static difference
// nobody looked for.
//
// Rows that have not been written yet are drawn as empty rules rather than left
// out. The card's height is the shape of the spec, and a card that grew a row at
// a time would move its own goal line up the frame on every beat.

const CARD_W = Math.min(STAGE_W, 1300);
const ROW_H = 82;
const BOX = 38;

type Criterion = {text: string; status?: string; note?: string};
type Step = {startMs: number; endMs: number; at?: number; check?: boolean};

export const SpecScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const title = String(props.title ?? '');
  const goal = String(props.goal ?? '');
  const constraints = (Array.isArray(props.constraints) ? props.constraints : []) as string[];
  const criteria = (Array.isArray(props.criteria) ? props.criteria : []) as Criterion[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];

  if (criteria.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const checking = step.check === true;
  const current = checking ? criteria.length - 1 : (step.at ?? 0);
  const written = current + 1;
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // The cascade. Ticks land across the first two thirds of the closing beat, so
  // the verdict has a moment to itself afterwards.
  const checkP = checking
    ? Math.min(1, Math.max(0, (nowMs - step.startMs) / Math.max(1, step.endMs - step.startMs)))
    : 0;
  const markedThrough = checking ? (checkP / 0.66) * criteria.length : 0;
  const marked = (i: number) => markedThrough > i;
  // The tally counts what the cascade has actually reached, so it ticks up with
  // the list instead of appearing finished. It is on screen from the first frame
  // reading zero, which is both true — nothing has been checked while the spec
  // is still being written — and better than reserving the strip's height for a
  // verdict that has not arrived, which left an unexplained gap under the list.
  const metSoFar = criteria.filter((c, i) => marked(i) && (c.status ?? 'met') !== 'missed').length;

  const write = spring({frame: sinceStep, fps, config: {damping: 200, mass: 0.7}, durationInFrames: 16});
  const active = criteria[current];

  return (
    <Stage justify="center">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={18} />

      <div
        style={{
          width: CARD_W,
          borderRadius: 26,
          backgroundColor: theme.surface,
          border: `1px solid ${theme.surfaceBorder}`,
          borderLeft: `6px solid ${theme.accent}`,
          boxShadow: `0 26px 66px ${withAlpha(theme.ink, 0.5)}`,
          overflow: 'hidden',
        }}
      >
        {/* The ask, and what it has to live inside. */}
        <div
          style={{
            padding: '26px 34px 24px',
            borderBottom: `1px solid ${theme.surfaceBorder}`,
            backgroundColor: withAlpha(theme.bgBottom, 0.45),
          }}
        >
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 15,
              letterSpacing: 2.6,
              textTransform: 'uppercase',
              color: theme.accentText,
              marginBottom: 10,
            }}
          >
            Spec
          </div>
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 40,
              fontWeight: 700,
              letterSpacing: -0.6,
              lineHeight: 1.15,
              color: theme.text,
            }}
          >
            {goal}
          </div>
          {constraints.length > 0 ? (
            <div style={{display: 'flex', gap: 12, marginTop: 18, flexWrap: 'wrap'}}>
              {constraints.map((c) => (
                <span
                  key={c}
                  style={{
                    padding: '7px 16px',
                    borderRadius: 999,
                    border: `1px solid ${withAlpha(theme.line, 0.5)}`,
                    fontFamily: theme.fontMono,
                    fontSize: 18,
                    letterSpacing: 0.6,
                    color: theme.textMuted,
                  }}
                >
                  {c}
                </span>
              ))}
            </div>
          ) : null}
        </div>

        {/* The checklist. Every row is present from the first frame; what
            changes is whether it has been written on. */}
        <div style={{padding: '12px 34px 16px'}}>
          {criteria.map((c, i) => {
            const isWritten = i < written;
            const isCurrent = !checking && i === current;
            const isMarked = marked(i);
            const missed = (c.status ?? 'met') === 'missed';
            const land = i === current && !checking ? write : 1;
            return (
              <div
                key={i}
                style={{
                  height: ROW_H,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 20,
                  borderBottom:
                    i < criteria.length - 1 ? `1px solid ${withAlpha(theme.line, 0.16)}` : 'none',
                  opacity: isWritten ? land : 1,
                  transform: `translateY(${(1 - land) * 12}px)`,
                }}
              >
                <Box theme={theme} on={isWritten} marked={isMarked} missed={missed} />
                {isWritten ? (
                  <span
                    style={{
                      flex: 1,
                      fontFamily: theme.fontBody,
                      fontSize: 31,
                      lineHeight: 1.25,
                      fontWeight: isCurrent ? 600 : 400,
                      color: isMarked && missed ? theme.textMuted : theme.text,
                      textDecoration: isMarked && missed ? 'line-through' : 'none',
                      textDecorationColor: withAlpha(theme.line, 0.7),
                    }}
                  >
                    {c.text}
                  </span>
                ) : (
                  // Not written yet: a rule where the sentence will go. The row
                  // still occupies its height, so the card never resizes.
                  <span
                    style={{
                      flex: 1,
                      height: 12,
                      borderRadius: 6,
                      backgroundColor: withAlpha(theme.line, 0.16),
                      maxWidth: `${64 - i * 6}%`,
                    }}
                  />
                )}
              </div>
            );
          })}
        </div>

        {/* The tally. Present throughout, reading zero while the spec is only
            being written — so the payoff is the number moving rather than a
            strip fading in over space that was empty until it did. */}
        <div
          style={{
            height: 74,
            display: 'flex',
            alignItems: 'center',
            gap: 22,
            padding: '0 34px',
            borderTop: `1px solid ${theme.surfaceBorder}`,
            backgroundColor: withAlpha(theme.bgBottom, 0.45),
          }}
        >
          <span
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 32,
              fontWeight: 700,
              color: metSoFar > 0 ? theme.text : theme.textMuted,
            }}
          >
            {`${metSoFar} of ${criteria.length}`}
          </span>
          <div
            style={{
              flex: 1,
              height: 9,
              borderRadius: 5,
              backgroundColor: withAlpha(theme.line, 0.22),
              overflow: 'hidden',
            }}
          >
            <div
              style={{
                width: `${(metSoFar / criteria.length) * 100}%`,
                height: 9,
                borderRadius: 5,
                backgroundColor: theme.accent,
              }}
            />
          </div>
        </div>
      </div>

      <div
        style={{
          marginTop: 22,
          width: CARD_W,
          textAlign: 'center',
          fontFamily: theme.fontBody,
          fontSize: 28,
          lineHeight: 1.35,
          color: theme.textMuted,
          opacity: interpolate(sinceStep, [0, 12], [0, 1], {
            extrapolateLeft: 'clamp',
            extrapolateRight: 'clamp',
          }),
        }}
      >
        {checking ? '' : (active?.note ?? '')}
      </div>
    </Stage>
  );
};

/**
 * The checkbox.
 *
 * Three states and no colour between them beyond the accent: empty while the
 * line is only written down, filled with a tick when it is met, and crossed when
 * it is not. A miss reads as a miss by its *shape*, for the same reason
 * PromptLoopScene's status does — there is no semantic red in the design system
 * to reach for, and inventing one would neither flip with the mode nor survive a
 * viewer who cannot see it.
 */
const Box: React.FC<{theme: ResolvedTheme; on: boolean; marked: boolean; missed: boolean}> = ({
  theme,
  on,
  marked,
  missed,
}) => {
  const filled = marked && !missed;
  return (
    <div
      style={{
        width: BOX,
        height: BOX,
        flexShrink: 0,
        borderRadius: 11,
        backgroundColor: filled ? theme.accent : 'transparent',
        border: `2.5px solid ${
          filled ? theme.accent : marked ? theme.line : withAlpha(theme.line, on ? 0.55 : 0.28)
        }`,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      {filled ? (
        <svg width={22} height={22} viewBox="0 0 22 22">
          <path
            d="M5 11.4 L9 15.4 L17 6.6"
            fill="none"
            stroke={theme.ink}
            strokeWidth={3.2}
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      ) : marked && missed ? (
        <svg width={20} height={20} viewBox="0 0 20 20">
          <path
            d="M5 5 L15 15 M15 5 L5 15"
            fill="none"
            stroke={theme.line}
            strokeWidth={3}
            strokeLinecap="round"
          />
        </svg>
      ) : null}
    </div>
  );
};
