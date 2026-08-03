import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// ToggleScene is a question, a switch that answers it, and the asterisks piling
// up underneath.
//
// The switch is a physical object rather than two words with one highlighted: a
// track, a knob, and the knob travelling. That matters because the claim is that
// the answer MOVED — from what people assume to what is true — and a colour
// change on a word is a state, not a movement. The knob is thrown once, on the
// first beat, with a spring that overshoots slightly, and then it never moves
// again however many qualifiers land.
//
// Both answers stay legible for the whole clip, the abandoned one dimmed rather
// than removed. Removing it would leave a clip that merely asserts; keeping it
// means the frame is still visibly arguing with the assumption it started from,
// which is what the qualifiers are doing.
//
// The qualifiers stack downward and accumulate. They are the reason this is not
// a title card, so they get the lower two thirds of the frame and the switch
// shrinks into a header once the first one lands — the answer is the setup, and
// after three seconds it is no longer the content.

const LAYOUT_W = Math.min(STAGE_W, 1120);
const TRACK_W = 190;
const TRACK_H = 78;
const KNOB = 62;

type Qualifier = {label: string; note?: string; role: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'answer' | 'qualify' | 'settle';
  at?: number;
  raised: number[];
};

const roleColour = (theme: ResolvedTheme, role: string): string => {
  switch (role) {
    case 'quantity':
      return theme.accentQuantity;
    case 'limit':
      return theme.accentLimit;
    case 'rival':
      return theme.accentRival;
    default:
      return theme.textMuted;
  }
};

export const ToggleScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const question = String(props.question ?? '');
  const from = String(props.from ?? '');
  const to = String(props.to ?? '');
  const qualifiers = (Array.isArray(props.qualifiers) ? props.qualifiers : []) as Qualifier[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (!question || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const raised = new Set(step.raised ?? []);
  const answered = steps[0].startMs <= nowMs;

  // The throw. Sprung, and only ever on the first beat: a switch that keeps
  // moving is a clip that has not decided.
  const sinceStart = ((nowMs - steps[0].startMs) / 1000) * FPS;
  const thrown = answered
    ? spring({frame: sinceStart - 8, fps, config: {damping: 12, mass: 0.6}, durationInFrames: 30})
    : 0;

  // Once a qualifier has landed the switch is history, so it recedes and gives
  // the frame to the asterisks.
  const compact = raised.size > 0 ? 1 : 0;
  const scale = 1 - 0.28 * compact;

  const answerColour = theme.accentRival;

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

      {/* The question, in the viewer's words. It stays for the whole clip: the
          asterisks are asterisks ON something, and dropping it would leave a
          list of conditions attached to nothing. */}
      <div
        style={{
          fontFamily: theme.fontBody,
          fontSize: 30,
          color: theme.textMuted,
          marginBottom: 22,
          textAlign: 'center',
          maxWidth: 1000,
        }}
      >
        {question}
      </div>

      {/* The switch. */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 30,
          transform: `scale(${scale})`,
          marginBottom: 10 + 18 * (1 - compact),
        }}
      >
        <span
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 58,
            fontWeight: 800,
            textTransform: 'uppercase',
            letterSpacing: -1,
            // The abandoned answer stays legible: the frame is still arguing
            // with the assumption it started from.
            color: theme.textMuted,
            opacity: 0.35,
          }}
        >
          {from}
        </span>

        <div
          style={{
            width: TRACK_W,
            height: TRACK_H,
            borderRadius: TRACK_H / 2,
            position: 'relative',
            background: withAlpha(answerColour, 0.1 + 0.12 * thrown),
            border: `2px solid ${withAlpha(answerColour, 0.35 + 0.4 * thrown)}`,
          }}
        >
          <div
            style={{
              position: 'absolute',
              top: (TRACK_H - KNOB) / 2 - 2,
              left: 6 + thrown * (TRACK_W - KNOB - 16),
              width: KNOB,
              height: KNOB,
              borderRadius: KNOB / 2,
              background: answerColour,
              boxShadow: `0 0 26px ${withAlpha(answerColour, 0.5)}`,
            }}
          />
        </div>

        <span
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 58,
            fontWeight: 800,
            textTransform: 'uppercase',
            letterSpacing: -1,
            color: answerColour,
            opacity: 0.3 + 0.7 * thrown,
          }}
        >
          {to}
        </span>
      </div>

      {/* The asterisks. */}
      <div style={{width: LAYOUT_W}}>
        {qualifiers.map((q, i) => {
          if (!raised.has(i)) return null;
          const isCurrent = step.at === i;
          const colour = roleColour(theme, q.role);
          const landing = isCurrent
            ? interpolate(sinceStep, [2, 18], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              })
            : 1;
          return (
            <div
              key={i}
              style={{
                display: 'flex',
                gap: 18,
                alignItems: 'baseline',
                marginTop: 16,
                paddingInline: 20,
                paddingBlock: 14,
                borderRadius: 10,
                background: isCurrent ? withAlpha(colour, 0.1) : undefined,
                borderLeft: `3px solid ${withAlpha(colour, isCurrent ? 0.85 : 0.4)}`,
                opacity: landing,
                transform: `translateY(${(1 - landing) * 10}px)`,
              }}
            >
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 17,
                  letterSpacing: 2.2,
                  textTransform: 'uppercase',
                  color: colour,
                  whiteSpace: 'nowrap',
                }}
              >
                {q.label}
              </span>
              {q.note ? (
                <span
                  style={{
                    fontFamily: theme.fontBody,
                    fontSize: 24,
                    color: isCurrent ? theme.text : theme.textMuted,
                  }}
                >
                  {q.note}
                </span>
              ) : null}
            </div>
          );
        })}
      </div>
    </Stage>
  );
};
