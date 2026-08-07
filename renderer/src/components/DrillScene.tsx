import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// DrillScene: the check, with the wrong answers killed on camera.
//
// A comprehension check that just states the answer teaches nothing, because
// the viewer never got to be wrong. So this composition puts the question up
// large and the options as lettered plates, then spends its middle beats
// ELIMINATING — each dismissed plate takes a strike through its text and
// desaturates in place. It does not disappear. A vanished option is an option
// the viewer can no longer check their own guess against, and the whole value
// of a drill is that they guessed.
//
// The plates are lettered A/B/C/D in mono because that is the grammar of every
// quiz anyone has ever sat, and reusing it means zero explanation. The letter
// chip is also the only part of a plate that keeps full contrast after a
// strike, so the eliminated rows stay countable at a glance.
//
// The answer does not simply survive by attrition — it LIFTS, with a ring
// around it, because arriving at the right answer should read as an event and
// not as the last thing left. Then the why-line lands underneath in accent.
// That order matters: the reveal is the payoff and the reason is the lesson,
// and collapsing them into one beat loses the second.
//
// One glow maximum: the ring on the revealed plate.

const CARD_W = Math.min(STAGE_W, 1280);
const LETTER = ['A', 'B', 'C', 'D', 'E', 'F'];

type Step = {
  startMs: number;
  endMs: number;
  show: 'ask' | 'eliminate' | 'reveal' | 'why';
  at?: number;
  struck: number[];
  revealed: boolean;
  whyOn: boolean;
};

export const DrillScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const question = String(props.question ?? '');
  const why = String(props.why ?? '');
  const answer = Number(props.answer ?? 0);
  const options = (Array.isArray(props.options) ? props.options : []) as string[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (options.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const striking = step.show === 'eliminate' ? step.at ?? -1 : -1;
  const struck = new Set(Array.isArray(step.struck) ? step.struck : []);
  const revealed = Boolean(step.revealed);
  const whyOn = Boolean(step.whyOn);

  // The strike draws left to right rather than appearing, so the elimination
  // is an action the viewer watches happen.
  const strike = interpolate(sinceStep, [2, 18], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});
  const lift = revealed
    ? spring({frame: sinceStep - 2, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 26})
    : 0;
  const whyLand = whyOn
    ? spring({frame: sinceStep - 3, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 28})
    : 0;
  const plateH = options.length > 3 ? 82 : 92;

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

      <div style={{width: CARD_W}}>
        {/* The question, set as the loudest thing on the board. */}
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 46,
            fontWeight: 700,
            letterSpacing: -1,
            lineHeight: 1.16,
            color: theme.text,
            marginBottom: 30,
            paddingBottom: 24,
            borderBottom: `2px solid ${withAlpha(theme.line, 0.22)}`,
          }}
        >
          {question}
        </div>

        <div style={{display: 'flex', flexDirection: 'column', gap: 13}}>
          {options.map((opt, i) => {
            const isStriking = i === striking;
            const dead = struck.has(i);
            const isAnswer = revealed && i === answer;
            const cut = dead ? (isStriking ? strike : 1) : 0;
            const rise = isAnswer ? lift : 0;
            return (
              <div
                key={i}
                style={{
                  position: 'relative',
                  minHeight: plateH,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 22,
                  paddingInline: 22,
                  paddingBlock: 14,
                  borderRadius: 14,
                  background: isAnswer
                    ? withAlpha(theme.accent, 0.14 * rise)
                    : withAlpha(theme.surface, dead ? 0.4 : 0.85),
                  border: `2px solid ${
                    isAnswer ? withAlpha(theme.accent, 0.5 + 0.4 * rise) : dead ? withAlpha(theme.surfaceBorder, 0.6) : theme.surfaceBorder
                  }`,
                  // Desaturated, not hidden: the viewer still has to be able to
                  // read what was ruled out.
                  filter: dead ? `saturate(${1 - 0.85 * cut})` : undefined,
                  opacity: dead ? 1 - 0.42 * cut : 1,
                  transform: `translateX(${8 * rise}px) scale(${1 + 0.02 * rise})`,
                  // The one glow: the ring on the answer.
                  boxShadow: isAnswer
                    ? `0 0 0 ${3 * rise}px ${withAlpha(theme.accent, 0.28)}, 0 0 ${34 * rise}px ${withAlpha(theme.accent, 0.26)}`
                    : undefined,
                }}
              >
                <div
                  style={{
                    width: 48,
                    height: 48,
                    flexShrink: 0,
                    borderRadius: 10,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: isAnswer ? withAlpha(theme.accent, 0.85) : withAlpha(theme.mass, 0.1),
                    border: `1.5px solid ${isAnswer ? withAlpha(theme.accent, 0.9) : theme.surfaceBorder}`,
                    fontFamily: theme.fontMono,
                    fontSize: 22,
                    fontWeight: 700,
                    color: isAnswer ? theme.ink : theme.textMuted,
                  }}
                >
                  {LETTER[i] ?? String(i + 1)}
                </div>
                <div style={{position: 'relative', flex: 1}}>
                  <span
                    style={{
                      fontFamily: theme.fontBody,
                      fontSize: 28,
                      lineHeight: 1.25,
                      color: isAnswer ? theme.text : dead ? theme.textMuted : theme.text,
                    }}
                  >
                    {opt}
                  </span>
                  {/* The strike, drawn as a growing rule rather than a text
                      decoration, so it can be animated. */}
                  <div
                    style={{
                      position: 'absolute',
                      left: 0,
                      top: '50%',
                      width: `${100 * cut}%`,
                      height: 2.5,
                      borderRadius: 2,
                      background: withAlpha(theme.accentRival, 0.75),
                    }}
                  />
                </div>
              </div>
            );
          })}
        </div>

        {/* The reason, under the answer, in accent. */}
        <div
          style={{
            marginTop: 26,
            display: 'flex',
            alignItems: 'flex-start',
            gap: 14,
            opacity: whyLand,
            transform: `translateY(${(1 - whyLand) * 12}px)`,
          }}
        >
          <div
            style={{
              width: 4,
              alignSelf: 'stretch',
              borderRadius: 2,
              background: withAlpha(theme.accent, 0.8),
              flexShrink: 0,
            }}
          />
          <div
            style={{
              fontFamily: theme.fontBody,
              fontSize: 28,
              lineHeight: 1.3,
              color: theme.accentText,
              paddingTop: 2,
            }}
          >
            {why}
          </div>
        </div>
      </div>
    </Stage>
  );
};
