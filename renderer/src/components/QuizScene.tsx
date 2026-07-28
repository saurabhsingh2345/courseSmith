import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {Check, X} from 'lucide-react';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_W} from './Stage';

// QuizScene is one question, on screen from the moment it is asked until the
// end of the clip, with the beats changing only what is lit.
//
// One scene for the whole clip rather than one per beat, for the reason the
// data template uses the same shape: re-mounting the question would re-animate
// it, and a question that re-lands while the viewer is trying to answer it is a
// question they have to read again. The pause only works if the thing being
// held is completely still.
//
// The four states come from Go (snippet_quiz.go) and are enforced there: ask,
// think, reveal, explain. Everything here is a consequence of which one is
// current, so there is no state machine — just a lookup of the step whose span
// contains this frame.

const HEADER_H = 108;
// The question column. Narrower than the stage on purpose: a question set to
// the full 1700px runs to two lines of very long measure, and the eye loses the
// start of the next line — which is the last thing you want from a sentence
// somebody has to hold in mind.
const COL_W = Math.min(STAGE_W, 1320);

// Sized to fill the frame rather than to fit it. The first pass used a 96px
// row and 34px option text, which left the whole bottom third of a 1080 frame
// empty — and this scene is read, not glanced at, so bigger type is not
// decoration.
const ROW_H = 118;
const ROW_GAP = 22;

type Step = {startMs: number; endMs: number; show: string; option?: number};

/** A, B, C… for the option chips. */
const letterFor = (i: number): string => String.fromCharCode(65 + i);

export const QuizScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const title = String(props.title ?? '');
  const question = String(props.question ?? '');
  const options = (Array.isArray(props.options) ? props.options : []) as string[];
  const why = (Array.isArray(props.why) ? props.why : []) as string[];
  const answer = typeof props.answer === 'number' ? props.answer : 0;
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];

  if (options.length === 0 || steps.length === 0) {
    return null;
  }

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  // The current step, and how long we have been in it. `find` rather than an
  // index: the spans are contiguous and a frame past the last one should hold
  // the last state rather than fall off the end.
  let stepIdx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (stepIdx < 0) stepIdx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[stepIdx];
  const sinceStep = Math.max(0, ((nowMs - step.startMs) / 1000) * FPS);

  // Once asked, always asked; once revealed, always revealed. Derived by
  // scanning rather than tracked, because a frame renders on its own.
  const askedAt = steps.find((s) => s.show === 'ask');
  const revealStep = steps.find((s) => s.show === 'reveal');
  const asked = askedAt ? nowMs >= askedAt.startMs : true;
  const revealed = revealStep ? nowMs >= revealStep.startMs : false;
  const framesSinceReveal = revealStep ? ((nowMs - revealStep.startMs) / 1000) * FPS : -1;

  // The option this beat is talking about, if any.
  const explaining = step.show === 'explain' ? step.option : undefined;

  const questionIn = spring({
    frame: askedAt ? ((nowMs - askedAt.startMs) / 1000) * FPS : 0,
    fps,
    config: {damping: 200, mass: 0.6},
    durationInFrames: 18,
  });

  // The thinking pulse. A ring that fills over the whole `think` span, which is
  // the only honest countdown available — it is showing exactly how much time
  // is left, because the span is the beat.
  const thinking = step.show === 'think';
  const thinkP = thinking
    ? Math.max(0, Math.min(1, (nowMs - step.startMs) / Math.max(1, step.endMs - step.startMs)))
    : 0;

  return (
    <Stage justify="flex-start">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={26} />
      <div style={{width: COL_W, display: 'flex', flexDirection: 'column', gap: 26}}>
        {/* The question. */}
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 60,
            lineHeight: 1.2,
            fontWeight: 700,
            color: theme.text,
            opacity: asked ? questionIn : 0,
            transform: `translateY(${(1 - (asked ? questionIn : 0)) * 20}px)`,
          }}
        >
          {question}
        </div>

        {/* The options. */}
        <div style={{display: 'flex', flexDirection: 'column', gap: ROW_GAP}}>
          {options.map((opt, i) => {
            const isAnswer = i === answer;
            // Options stagger in behind the question rather than with it, so the
            // question is read first — which is the order somebody answering
            // would use anyway.
            const rowIn = asked
              ? spring({
                  frame: (askedAt ? ((nowMs - askedAt.startMs) / 1000) * FPS : 0) - 8 - i * 4,
                  fps,
                  config: {damping: 200, mass: 0.5},
                  durationInFrames: 16,
                })
              : 0;

            // After the reveal the right answer holds full strength and the
            // wrong ones recede — except the one currently being explained,
            // which comes back up, because it is the subject of the sentence.
            const settled = revealed
              ? interpolate(framesSinceReveal, [0, 14], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                })
              : 0;
            const lit = !revealed || isAnswer || explaining === i;
            const dim = 1 - settled * (lit ? 0 : 0.62);

            const border = revealed
              ? isAnswer
                ? theme.accent
                : explaining === i
                  ? theme.textMuted
                  : theme.surfaceBorder
              : theme.surfaceBorder;

            return (
              <div key={i} style={{opacity: rowIn * dim}}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 22,
                    minHeight: ROW_H,
                    padding: '0 28px',
                    borderRadius: 16,
                    backgroundColor: theme.surface,
                    border: `2px solid ${border}`,
                    // The right answer lifts rather than just recolouring: a
                    // colour change alone is invisible to a fifth of viewers,
                    // and the lift reads for everyone.
                    transform: `translateX(${revealed && isAnswer ? settled * 14 : 0}px) scale(${
                      1 + (revealed && isAnswer ? settled * 0.02 : 0)
                    })`,
                  }}
                >
                  <div
                    style={{
                      width: 60,
                      height: 60,
                      flexShrink: 0,
                      borderRadius: 12,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      backgroundColor:
                        revealed && isAnswer ? theme.accent : `${theme.textMuted}22`,
                      color: revealed && isAnswer ? theme.bgBottom : theme.textMuted,
                      fontFamily: theme.fontMono,
                      fontSize: 30,
                      fontWeight: 700,
                    }}
                  >
                    {revealed && isAnswer ? (
                      <Check size={28} strokeWidth={3} />
                    ) : revealed && !isAnswer && explaining === i ? (
                      <X size={26} strokeWidth={3} />
                    ) : (
                      letterFor(i)
                    )}
                  </div>
                  <div
                    style={{
                      fontFamily: theme.fontBody,
                      fontSize: 40,
                      fontWeight: revealed && isAnswer ? 700 : 500,
                      color: theme.text,
                      minWidth: 0,
                    }}
                  >
                    {opt}
                  </div>
                </div>

                {/* The explanation, under the option it belongs to. It appears
                    only while that option is the subject, so the frame never
                    holds more than one line of prose to read. */}
                {explaining === i && why[i] && (
                  <div
                    style={{
                      marginTop: 12,
                      marginLeft: 28,
                      paddingLeft: 20,
                      borderLeft: `3px solid ${isAnswer ? theme.accent : theme.textMuted}`,
                      fontFamily: theme.fontBody,
                      fontSize: 31,
                      lineHeight: 1.35,
                      color: theme.textMuted,
                      opacity: interpolate(sinceStep, [0, 10], [0, 1], {
                        extrapolateLeft: 'clamp',
                        extrapolateRight: 'clamp',
                      }),
                    }}
                  >
                    {why[i]}
                  </div>
                )}
              </div>
            );
          })}
        </div>

        {/* The thinking bar. Only during a `think` beat, and it is a real
            measure rather than decoration: it fills across exactly the span the
            viewer has, because the span IS the beat. */}
        {thinking && (
          <div style={{display: 'flex', alignItems: 'center', gap: 16, marginTop: 4}}>
            <div
              style={{
                flex: 1,
                height: 8,
                borderRadius: 4,
                backgroundColor: `${theme.textMuted}26`,
                overflow: 'hidden',
              }}
            >
              <div
                style={{
                  width: `${thinkP * 100}%`,
                  height: '100%',
                  backgroundColor: theme.accent,
                  borderRadius: 4,
                }}
              />
            </div>
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 24,
                color: theme.textMuted,
                letterSpacing: 1,
              }}
            >
              your turn
            </div>
          </div>
        )}
      </div>
    </Stage>
  );
};
