import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, seat, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';
import {AppWindow, windowDim, windowText} from './AppWindow';
import {iconFor} from './icons';

// ApprovalScene draws a permission prompt and what each answer hands over.
//
// The composition is a window over a set of rows, and the split is the argument:
// what the tool is ASKING is a literal, so it goes inside a terminal window in
// mono, exactly as the viewer will meet it; what each answer COSTS is editorial, so
// it goes on paper cards below, in the house type. Putting the consequences inside
// the window too would have been more faithful to the tool and much worse teaching
// — it would read as more output to skim, which is what people already do with
// these prompts.
//
// The risk mark is the one piece of colour, and it is on the row rather than on the
// text: a red word among three grey ones reads as emphasis, while a red edge down
// the side of a card reads as a warning about the whole option. The reference for
// this frame is a presenter saying "be careful with that one" out loud, which is
// exactly what a video cannot show and a frame can.
//
// The pick is marked with a check and a rim, and it may be the risky row. That
// combination looks like a mistake and is the honest answer surprisingly often —
// see the validator, which deliberately allows it.

const BLOCK_W = Math.min(STAGE_W, 1480);
const WINDOW_H = 300;

type Answer = {label: string; consequence: string; risk?: boolean};

type Step = {
  startMs: number;
  endMs: number;
  show: 'ask' | 'answer' | 'pick';
  at?: number;
  read?: number;
};

export const ApprovalScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const tool = String(props.tool ?? 'Claude Code');
  const context = String(props.context ?? '');
  const ask = String(props.ask ?? '');
  const closer = String(props.closer ?? '');
  const pick = Number(props.pick ?? 0);
  const answers = (Array.isArray(props.answers) ? props.answers : []) as Answer[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (answers.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const since = ((nowMs - step.startMs) / 1000) * FPS;

  const lit = step.show === 'answer' ? (step.at ?? -1) : -1;
  const picked = step.show === 'pick';
  const land = interpolate(since, [2, 16], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const up = interpolate(frame, [0, 20], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const Check = iconFor('check');

  return (
    <Stage justify="center">
      <div style={{width: BLOCK_W}}>
        <SceneHeader
          theme={theme}
          title={String(props.title ?? '')}
          emphasis={props.emphasis as string | undefined}
          emphasisRole={props.emphasisRole as string | undefined}
          size="compact"
          marginBottom={34}
        />

        <AppWindow theme={theme} title={tool} badge="✳" width={BLOCK_W} height={WINDOW_H}>
          <div style={{padding: '30px 40px', display: 'flex', flexDirection: 'column', gap: 18}}>
            {context ? (
              <div style={{...windowText(theme, 24), color: windowDim()}}>{context}</div>
            ) : null}
            <div style={{display: 'flex', alignItems: 'center', gap: 16}}>
              {/* The prompt's own caret, so the ask reads as a line the tool
                  printed rather than as a caption somebody wrote about it. */}
              <div style={{...windowText(theme, 34), color: theme.accent}}>❯</div>
              <div
                style={{
                  ...windowText(theme, 34),
                  color: '#f0f2f5',
                  fontWeight: 500,
                  // A faint band behind the literal. It is the one thing in the
                  // window the viewer must actually read, and on a dark field at
                  // this size type alone does not hold the eye.
                  background: withAlpha('#ffffff', 0.06),
                  padding: '10px 18px',
                  borderRadius: 8,
                }}
              >
                {ask}
              </div>
            </div>
            <div style={{...windowText(theme, 22), color: windowDim(), marginTop: 4}}>
              Do you want to allow this?
            </div>
          </div>
        </AppWindow>

        {/* The answers, on paper. */}
        <div style={{marginTop: 30, display: 'flex', flexDirection: 'column', gap: 16}}>
          {answers.map((a, i) => {
            const isLit = lit === i;
            const isPick = picked && pick === i;
            const seenAlready = (step.read ?? 0) > i;
            const enter = interpolate(frame, [6 + i * 6, 24 + i * 6], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            });
            const edge = a.risk ? theme.accentLimit : isPick ? theme.accentText : theme.surfaceBorder;
            return (
              <div
                key={i}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 26,
                  padding: '24px 30px',
                  borderRadius: 18,
                  background: theme.surface,
                  border: `1px solid ${theme.surfaceBorder}`,
                  // The risk edge, down the side of the whole row. See the header:
                  // a coloured word is emphasis, a coloured edge is a warning about
                  // the option.
                  borderLeft: `6px solid ${a.risk ? theme.accentLimit : isPick ? theme.accentText : 'transparent'}`,
                  boxShadow: [
                    seat(theme, isLit || isPick ? 'lifted' : 'resting'),
                    isLit || isPick ? `0 0 0 4px ${withAlpha(edge, 0.16)}` : '',
                  ]
                    .filter(Boolean)
                    .join(', '),
                  opacity: enter,
                  transform: `translateY(${(1 - enter) * 12 - (isLit ? land * 6 : 0)}px)`,
                }}
              >
                <div
                  style={{
                    // 430, not 320. "Yes, and don't ask again" is the realistic
                    // label — it is the whole point of the template — and at 320 it
                    // wrapped to two lines, which made that one row taller than its
                    // neighbours. Uneven rows in a list of options read as one
                    // option being more important, which is a claim the frame did
                    // not mean to make.
                    width: 430,
                    flex: 'none',
                    fontFamily: theme.fontDisplay,
                    fontSize: 30,
                    fontWeight: 700,
                    letterSpacing: -0.5,
                    color: isLit || isPick || seenAlready ? theme.text : theme.textMuted,
                  }}
                >
                  {a.label}
                </div>
                <div
                  style={{
                    flex: 1,
                    fontFamily: theme.fontBody,
                    fontSize: 28,
                    lineHeight: 1.35,
                    // The consequence is the content, so it is only fully inked
                    // once the voice has got to it. Before that it is present but
                    // quiet — the viewer can see there IS a consequence, which is
                    // the tension the frame runs on.
                    color: isLit || isPick || seenAlready ? theme.text : withAlpha(theme.text, 0.4),
                  }}
                >
                  {a.consequence}
                </div>
                {a.risk ? (
                  <div
                    style={{
                      flex: 'none',
                      fontFamily: theme.fontBody,
                      fontSize: 19,
                      fontWeight: 700,
                      letterSpacing: 2.4,
                      textTransform: 'uppercase',
                      color: theme.accentLimit,
                    }}
                  >
                    hands over most
                  </div>
                ) : null}
                {isPick ? (
                  <Check size={38} strokeWidth={3} color={theme.accentText} style={{flex: 'none'}} />
                ) : null}
              </div>
            );
          })}
        </div>

        <div style={{minHeight: 76, marginTop: 26, textAlign: 'center', opacity: up}}>
          {picked && closer ? (
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 36,
                fontWeight: 700,
                letterSpacing: -0.5,
                color: theme.accentText,
                opacity: land,
                transform: `translateY(${(1 - land) * 10}px)`,
              }}
            >
              {closer}
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
