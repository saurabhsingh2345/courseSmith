import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_W} from './Stage';
import {iconFor} from './icons';

// PromptLoopScene is a conversation with something that builds.
//
// Three things are on screen and each has one job. The goal is pinned across the
// top and never changes — it is what lets a viewer judge whether an attempt got
// closer, and pinning it is how the template teaches that a loop needs something
// to converge on. The thread runs down the left and accumulates. The result
// panel on the right holds one attempt at a time, counting up.
//
// It deliberately shows no code. A vibe-coding clip whose payoff is a
// syntax-highlighted buffer has taught the old skill with new tooling in the
// background; what you actually look at is whether it worked and what it
// changed, so that is what the frame holds.
//
// The status is drawn as *form* rather than as colour — a filled tick, a barred
// ring, a crossed one. The design system is three brand colours and what Go
// derives from them, with no semantic red to reach for, and a literal invented
// here would neither flip with the mode nor survive a colour-blind viewer.

const COL_W = Math.min(STAGE_W, 1700);
const GOAL_H = 66;
const BODY_H = 524;
const THREAD_W = 640;
const GUTTER = 28;
const RESULT_W = COL_W - THREAD_W - GUTTER;
/** How many turns of the thread stay on screen. Older ones have been read. */
const THREAD_DEPTH = 3;

type Turn = {
  who: string;
  text: string;
  startMs: number;
  endMs: number;
  attempt?: number;
  status?: string;
  changes?: string[];
};

/** How close to the goal an attempt got. The bar under the panel reads this. */
const REACH: Record<string, number> = {broken: 0.22, partial: 0.6, ok: 1};

/**
 * The verdict in a word.
 *
 * The mark alone was tried and it is not enough: a tick and a bar are only
 * unambiguous once you have seen both, and the first attempt in a clip is the
 * first time the viewer sees either. The word removes the guess, and it sits
 * beside the mark rather than replacing it so the two reinforce.
 */
const STATUS_LABEL: Record<string, string> = {ok: 'Works', partial: 'Almost', broken: 'Broken'};

export const PromptLoopScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const title = String(props.title ?? '');
  const goal = String(props.goal ?? '');
  const turns = (Array.isArray(props.turns) ? props.turns : []) as Turn[];

  if (turns.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = turns.findIndex((t) => nowMs >= t.startMs && nowMs < t.endMs);
  if (idx < 0) idx = nowMs < turns[0].startMs ? 0 : turns.length - 1;
  const sinceTurn = ((nowMs - turns[idx].startMs) / 1000) * FPS;
  const waiting = turns[idx].who !== 'ai';

  // The panel shows the newest answer there has been. On a prompt beat that is
  // the *previous* one, dimmed — which is exactly true of the screen you are
  // typing the next prompt into, and is why the panel never goes blank
  // mid-clip.
  let shownIdx = -1;
  for (let i = idx; i >= 0; i--) {
    if (turns[i].who === 'ai') {
      shownIdx = i;
      break;
    }
  }
  const shown = shownIdx >= 0 ? turns[shownIdx] : undefined;
  const sinceShown = shown ? ((nowMs - shown.startMs) / 1000) * FPS : 0;

  const enter = spring({frame: sinceTurn, fps, config: {damping: 200, mass: 0.7}, durationInFrames: 18});
  const Target = iconFor('target');
  const Spark = iconFor('sparkles');

  // The thread, newest last. Only the tail is drawn; what is off the top has
  // been read already.
  const visible = turns.slice(Math.max(0, idx - (THREAD_DEPTH - 1)), idx + 1);

  return (
    <Stage justify="center">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={18} />

      {/* The goal, pinned. It is the only thing on screen that never changes. */}
      <div
        style={{
          width: COL_W,
          height: GOAL_H,
          display: 'flex',
          alignItems: 'center',
          gap: 16,
          padding: '0 26px',
          borderRadius: 18,
          backgroundColor: withAlpha(theme.accent, 0.08),
          border: `1px solid ${withAlpha(theme.accent, 0.3)}`,
        }}
      >
        <Target size={26} color={theme.accent} strokeWidth={2.2} />
        <span
          style={{
            fontFamily: theme.fontMono,
            fontSize: 15,
            letterSpacing: 2,
            textTransform: 'uppercase',
            color: theme.accentText,
            flexShrink: 0,
          }}
        >
          Goal
        </span>
        <span
          style={{
            width: 1,
            height: 28,
            backgroundColor: withAlpha(theme.accent, 0.35),
            flexShrink: 0,
          }}
        />
        <span
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 27,
            fontWeight: 600,
            color: theme.text,
            whiteSpace: 'nowrap',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
          }}
        >
          {goal}
        </span>
      </div>

      <div style={{display: 'flex', gap: GUTTER, marginTop: 24, width: COL_W, height: BODY_H}}>
        {/* The thread. Turns are bottom-aligned so the newest always lands in
            the same place instead of the column growing downward past it. */}
        <div
          style={{
            width: THREAD_W,
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'flex-end',
            gap: 16,
          }}
        >
          {visible.map((turn, i) => {
            const isNewest = i === visible.length - 1;
            const age = visible.length - 1 - i;
            const you = turn.who !== 'ai';
            return (
              <div
                key={turn.startMs}
                style={{
                  // Older turns recede but stay readable. Fading them to near
                  // nothing left the column looking like it had rendering
                  // artefacts in it rather than a conversation with a history.
                  opacity: isNewest ? enter : age === 1 ? 0.52 : 0.3,
                  transform: `translateY(${isNewest ? (1 - enter) * 22 : 0}px)`,
                  padding: '17px 20px',
                  borderRadius: 18,
                  // A prompt is the solid card and the answer is the quiet one,
                  // because the prompt is the thing this template is teaching
                  // you to write. Tinting the prompt with the accent was tried
                  // first and yellow at 10% over a navy stage is mud — the
                  // separation is carried by weight and by the edge bar instead.
                  backgroundColor: you ? theme.surface : withAlpha(theme.surface, 0.45),
                  border: `1px solid ${theme.surfaceBorder}`,
                  borderLeft: `5px solid ${you ? theme.accent : theme.line}`,
                }}
              >
                <div style={{display: 'flex', alignItems: 'center', gap: 8, marginBottom: 9}}>
                  {you ? (
                    <span
                      style={{
                        width: 20,
                        height: 20,
                        borderRadius: 6,
                        backgroundColor: theme.accent,
                        display: 'inline-block',
                      }}
                    />
                  ) : (
                    <Spark size={20} color={theme.textMuted} strokeWidth={2.2} />
                  )}
                  <span
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: 14,
                      letterSpacing: 2,
                      textTransform: 'uppercase',
                      color: you ? theme.accentText : theme.textMuted,
                    }}
                  >
                    {you ? 'You' : 'AI'}
                  </span>
                </div>
                <div
                  style={{
                    fontFamily: theme.fontBody,
                    fontSize: 24,
                    lineHeight: 1.36,
                    color: you ? theme.text : theme.textMuted,
                  }}
                >
                  {turn.text}
                </div>
              </div>
            );
          })}
        </div>

        {/* The result. One attempt at a time, counting up — which is what makes
            the loop legible without drawing a diagram of a loop. */}
        <div
          style={{
            width: RESULT_W,
            borderRadius: 24,
            backgroundColor: theme.surface,
            border: `1px solid ${theme.surfaceBorder}`,
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              height: 56,
              flexShrink: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              padding: '0 22px',
              borderBottom: `1px solid ${theme.surfaceBorder}`,
              backgroundColor: withAlpha(theme.bgBottom, 0.5),
            }}
          >
            <div style={{display: 'flex', gap: 9}}>
              {[0, 1, 2].map((d) => (
                <span
                  key={d}
                  style={{
                    width: 11,
                    height: 11,
                    borderRadius: 6,
                    backgroundColor: theme.line,
                    opacity: 0.45,
                  }}
                />
              ))}
            </div>
            {shown?.attempt ? (
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 16,
                  letterSpacing: 1.6,
                  textTransform: 'uppercase',
                  color: theme.textMuted,
                  padding: '5px 13px',
                  borderRadius: 999,
                  border: `1px solid ${theme.surfaceBorder}`,
                }}
              >
                {`Attempt ${shown.attempt}`}
              </span>
            ) : null}
          </div>

          {/* The body is centred rather than top-aligned. Two bullets pinned to
              the top of a 400px panel is a void with a verdict floating above
              it, which reads as a layout that lost half its content. */}
          <div
            style={{
              flex: 1,
              padding: '28px 34px',
              position: 'relative',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
            }}
          >
            <div style={{opacity: waiting ? 0.14 : 1}}>
              {shown ? (
                <>
                  <div style={{display: 'flex', alignItems: 'center', gap: 20}}>
                    <StatusMark theme={theme} status={shown.status ?? 'ok'} />
                    <span
                      style={{
                        fontFamily: theme.fontDisplay,
                        fontSize: 38,
                        fontWeight: 700,
                        letterSpacing: -0.4,
                        color: (shown.status ?? 'ok') === 'broken' ? theme.textMuted : theme.text,
                      }}
                    >
                      {STATUS_LABEL[shown.status ?? 'ok'] ?? STATUS_LABEL.ok}
                    </span>
                  </div>
                  <div
                    style={{
                      height: 1,
                      backgroundColor: theme.surfaceBorder,
                      margin: '26px 0 24px',
                    }}
                  />
                  {(shown.changes ?? []).map((c, i) => (
                    <div
                      key={c}
                      style={{
                        display: 'flex',
                        alignItems: 'flex-start',
                        gap: 14,
                        marginBottom: 15,
                        opacity: waiting
                          ? 1
                          : interpolate(sinceShown, [8 + i * 6, 20 + i * 6], [0, 1], {
                              extrapolateLeft: 'clamp',
                              extrapolateRight: 'clamp',
                            }),
                      }}
                    >
                      <span
                        style={{
                          width: 9,
                          height: 9,
                          marginTop: 12,
                          flexShrink: 0,
                          borderRadius: 2,
                          backgroundColor: theme.accent,
                        }}
                      />
                      <span
                        style={{
                          fontFamily: theme.fontBody,
                          fontSize: 27,
                          lineHeight: 1.3,
                          color: theme.text,
                        }}
                      >
                        {c}
                      </span>
                    </div>
                  ))}
                </>
              ) : null}
            </div>

            {/* Waiting. The same three dots whether there is a previous result
                behind them or the panel has never held one — a prompt beat is a
                prompt beat, and the first one does not deserve its own design. */}
            {waiting ? (
              <div
                style={{
                  position: 'absolute',
                  inset: 0,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <div
                  style={{
                    display: 'flex',
                    gap: 16,
                    padding: '22px 32px',
                    borderRadius: 999,
                    backgroundColor: theme.surface,
                    border: `1px solid ${theme.surfaceBorder}`,
                    boxShadow: `0 16px 40px ${withAlpha(theme.ink, 0.5)}`,
                  }}
                >
                  {[0, 1, 2].map((d) => (
                    <span
                      key={d}
                      style={{
                        width: 18,
                        height: 18,
                        borderRadius: 9,
                        backgroundColor: theme.accent,
                        opacity: interpolate(
                          (frame + d * 8) % 36,
                          [0, 12, 24, 36],
                          [0.2, 1, 0.2, 0.2],
                        ),
                      }}
                    />
                  ))}
                </div>
              </div>
            ) : null}
          </div>

          {/* How close that attempt got to the goal. Across a clip this is the
              convergence — the bar reaching further each round is the argument
              the template is making. */}
          <div style={{height: 7, flexShrink: 0, backgroundColor: withAlpha(theme.line, 0.22)}}>
            <div
              style={{
                height: 7,
                width: `${
                  (REACH[shown?.status ?? 'ok'] ?? 1) *
                  100 *
                  (waiting
                    ? 1
                    : interpolate(sinceShown, [4, 26], [0, 1], {
                        extrapolateLeft: 'clamp',
                        extrapolateRight: 'clamp',
                      }))
                }%`,
                backgroundColor: theme.accent,
                opacity: shown ? (waiting ? 0.4 : 1) : 0,
              }}
            />
          </div>
        </div>
      </div>
    </Stage>
  );
};

/** The verdict, drawn as form rather than as colour. */
const StatusMark: React.FC<{theme: ResolvedTheme; status: string}> = ({theme, status}) => {
  const ok = status === 'ok';
  const broken = status === 'broken';
  const stroke = broken ? theme.line : theme.accent;
  return (
    <svg width={62} height={62} viewBox="0 0 62 62">
      <circle
        cx={31}
        cy={31}
        r={27}
        fill={ok ? theme.accent : 'none'}
        stroke={stroke}
        strokeWidth={ok ? 0 : 4}
      />
      {ok ? (
        <path
          d="M18 32 L27 41 L44 22"
          fill="none"
          stroke={theme.ink}
          strokeWidth={5}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      ) : broken ? (
        <path
          d="M21 21 L41 41 M41 21 L21 41"
          fill="none"
          stroke={stroke}
          strokeWidth={5}
          strokeLinecap="round"
        />
      ) : (
        <path d="M20 31 L42 31" fill="none" stroke={stroke} strokeWidth={5} strokeLinecap="round" />
      )}
    </svg>
  );
};
