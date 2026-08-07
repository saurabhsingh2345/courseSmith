import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// RelayScene: a baton travelling down a line of stages.
//
// One horizontal row, one axis, no branching — the composition is an argument
// that this subject has no branches. Boxes stacked or scattered would let the
// eye read the stages as things that coexist; a single line read left to right
// is the same claim the narration is making, made by the layout.
//
// The capsules are equal width and always all present, because the point of the
// opener is that the viewer sees the whole run before any of it happens. What
// changes is contrast: an unfired stage is a dim outline, a fired stage is
// filled and shows the one job it does. Nothing enters, nothing leaves.
//
// The spark is the only travelling element and it carries the entire idea. It
// leaves the previous capsule's right edge and lands on the next one's left
// edge, on a spring, and the capsule fills as it arrives — so "this stage is
// finished, that one now has control" is a single continuous motion rather than
// two states the viewer has to compare. It is also the one glow in the frame:
// nothing else here is allowed to emit light, because a chain of glowing
// capsules would say they are all live at once, which is the misreading this
// template exists to prevent.
//
// The hand-off rides under its chevron rather than inside either capsule, since
// it belongs to neither stage — it is what passes BETWEEN them, and putting it
// in the gap is the cheapest way to say so.

const CHAIN_W = Math.min(STAGE_W, 1660);
const CAP_H = 208;
const SPARK_R = 9;

type Stage_ = {label: string; does: string; hands: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'line' | 'ignite' | 'chain';
  at?: number;
  from?: number;
  lit: number[];
};

export const RelayScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const stages = (Array.isArray(props.stages) ? props.stages : []) as Stage_[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (stages.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const n = stages.length;
  const gap = n >= 6 ? 66 : 94;
  const capW = Math.floor((CHAIN_W - gap * (n - 1)) / n);
  const xOf = (i: number) => i * (capW + gap);

  const labelSize = capW > 260 ? 31 : capW > 200 ? 26 : 22;
  const doesSize = capW > 260 ? 19 : 17;

  const lit = new Set(step.lit ?? []);
  const igniting = step.show === 'ignite' && typeof step.at === 'number' ? step.at : -1;
  const allLit = step.show === 'chain';

  // The spark: it leaves the previous capsule as the beat opens and lands a
  // little before the new capsule finishes filling.
  const travel =
    igniting >= 0 && typeof step.from === 'number'
      ? spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.5}, durationInFrames: 26})
      : -1;
  const sparkFrom = typeof step.from === 'number' ? step.from : -1;
  const sparkX =
    travel >= 0 && sparkFrom >= 0
      ? interpolate(travel, [0, 1], [xOf(sparkFrom) + capW - SPARK_R, xOf(igniting) + SPARK_R])
      : -1;

  // The whole chain settles once, then holds.
  const chainIn = interpolate(frame, [0, 20], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

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

      <div style={{width: CHAIN_W, position: 'relative', height: CAP_H + 96, opacity: chainIn}}>
        {stages.map((s, i) => {
          const on = lit.has(i) || allLit;
          const isNew = i === igniting;
          const arrive = isNew
            ? spring({frame: sinceStep - 8, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 24})
            : 1;
          const fill = on ? (isNew ? arrive : 1) : 0;
          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: xOf(i),
                top: 0,
                width: capW,
                height: CAP_H,
                borderRadius: 18,
                padding: '22px 20px',
                display: 'flex',
                flexDirection: 'column',
                boxSizing: 'border-box',
                background: withAlpha(theme.surface, 0.35 + 0.6 * fill),
                border: `2px solid ${on ? withAlpha(theme.accent, 0.35 + 0.5 * fill) : theme.surfaceBorder}`,
                transform: `scale(${0.98 + 0.02 * fill})`,
                // The one glow, and only while the baton is landing.
                boxShadow: isNew ? `0 0 34px ${withAlpha(theme.accent, 0.34 * arrive)}` : undefined,
              }}
            >
              <div
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 13,
                  letterSpacing: 2,
                  color: withAlpha(theme.textMuted, 0.8),
                  marginBottom: 8,
                }}
              >
                {String(i + 1).padStart(2, '0')}
              </div>
              <div
                style={{
                  fontFamily: theme.fontDisplay,
                  fontSize: labelSize,
                  fontWeight: 700,
                  letterSpacing: -0.5,
                  lineHeight: 1.1,
                  color: on ? theme.text : withAlpha(theme.text, 0.45),
                }}
              >
                {s.label}
              </div>
              <div
                style={{
                  marginTop: 12,
                  fontFamily: theme.fontBody,
                  fontSize: doesSize,
                  lineHeight: 1.3,
                  color: theme.textMuted,
                  opacity: fill,
                  transform: `translateY(${(1 - fill) * 6}px)`,
                }}
              >
                {s.does}
              </div>
            </div>
          );
        })}

        {/* Chevrons and the hand-offs that ride under them. */}
        {stages.slice(0, -1).map((s, i) => {
          const passed = (lit.has(i) && lit.has(i + 1)) || allLit;
          const cx = xOf(i) + capW + gap / 2;
          return (
            <div key={`c${i}`} style={{position: 'absolute', left: cx - gap / 2, top: 0, width: gap}}>
              <div
                style={{
                  position: 'absolute',
                  left: gap / 2 - 8,
                  top: CAP_H / 2 - 9,
                  width: 15,
                  height: 15,
                  borderTop: `3px solid ${passed ? theme.accent : theme.line}`,
                  borderRight: `3px solid ${passed ? theme.accent : theme.line}`,
                  transform: 'rotate(45deg)',
                }}
              />
              <div
                style={{
                  position: 'absolute',
                  left: -34,
                  top: CAP_H + 18,
                  width: gap + 68,
                  textAlign: 'center',
                  fontFamily: theme.fontMono,
                  fontSize: 15,
                  lineHeight: 1.3,
                  color: passed ? theme.accentText : 'transparent',
                  opacity: passed ? 1 : 0,
                }}
              >
                {s.hands}
              </div>
            </div>
          );
        })}

        {/* The baton. */}
        {sparkX >= 0 ? (
          <div
            style={{
              position: 'absolute',
              left: sparkX - SPARK_R,
              top: CAP_H / 2 - SPARK_R,
              width: SPARK_R * 2,
              height: SPARK_R * 2,
              borderRadius: SPARK_R,
              background: theme.accent,
              boxShadow: `0 0 24px ${withAlpha(theme.accent, 0.8)}`,
              opacity: interpolate(travel, [0, 0.85, 1], [1, 1, 0], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          />
        ) : null}
      </div>
    </Stage>
  );
};
