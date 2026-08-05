import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// The trap: a mistake, the symptom it shows up as, and the one move out.
//
// Not a list, which is why this one does not use CourseListScene. The other four
// v5 scenes draw a SET — things that accumulate and are all true at the end.
// This draws a SEQUENCE with a turn in it: the mistake is struck out when the
// fix lands, and the symptom is framed as a thing you saw rather than a thing
// you are told.
//
// The symptom gets a monospace card because that is what it usually IS — a log
// line, an error, a number off a dashboard. Setting it as prose would lose the
// only property that matters about it: that the viewer can match it against
// what is on their own screen right now.

const CARD_W = Math.min(STAGE_W, 1180);

type Step = {startMs: number; endMs: number; show: 'mistake' | 'symptom' | 'why' | 'fix'};

export const PitfallScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const mistake = String(props.mistake ?? '');
  const symptom = String(props.symptom ?? '');
  const why = String(props.why ?? '');
  const fix = String(props.fix ?? '');
  if (steps.length === 0 || !mistake) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // What has been reached, so earlier parts persist instead of swapping out.
  const seen = new Set(steps.slice(0, idx + 1).map((s) => s.show));
  const fixed = seen.has('fix');

  const reveal = (on: boolean) =>
    on ? interpolate(sinceStep, [0, 14], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'}) : 0;

  const strike = fixed
    ? spring({frame: sinceStep - 4, fps, config: {damping: 16, mass: 0.5}, durationInFrames: 24})
    : 0;

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={28}
      />

      {/* The mistake. Struck when the fix lands — the strike IS the turn. */}
      <div style={{width: CARD_W, position: 'relative', marginBottom: 26}}>
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 44,
            fontWeight: 700,
            lineHeight: 1.15,
            letterSpacing: -0.8,
            color: fixed ? theme.textMuted : theme.text,
          }}
        >
          {mistake}
        </div>
        <div
          style={{
            position: 'absolute',
            left: 0,
            top: '50%',
            height: 4,
            borderRadius: 2,
            width: `${strike * 100}%`,
            background: theme.accentLimit,
          }}
        />
      </div>

      {/* The symptom, as the thing you would actually see. */}
      {seen.has('symptom') ? (
        <div
          style={{
            width: CARD_W,
            opacity: seen.has('symptom') ? Math.max(reveal(step.show === 'symptom'), 1) : 0,
            border: `2px solid ${withAlpha(theme.accentLimit, 0.5)}`,
            background: withAlpha(theme.surface, 0.7),
            borderRadius: 12,
            padding: '20px 26px',
            marginBottom: 24,
            display: 'flex',
            flexDirection: 'column',
            gap: 8,
          }}
        >
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 15,
              letterSpacing: 3,
              textTransform: 'uppercase',
              color: theme.accentLimit,
            }}
          >
            what you see
          </div>
          <div style={{fontFamily: theme.fontMono, fontSize: 28, color: theme.text, lineHeight: 1.35}}>
            {symptom}
          </div>
        </div>
      ) : null}

      {seen.has('why') && why ? (
        <div
          style={{
            width: CARD_W,
            fontFamily: theme.fontBody,
            fontSize: 26,
            color: theme.textMuted,
            marginBottom: 24,
            fontStyle: 'italic',
          }}
        >
          {why}
        </div>
      ) : null}

      {seen.has('fix') && fix ? (
        <div style={{width: CARD_W, display: 'flex', gap: 20, alignItems: 'stretch'}}>
          <div style={{width: 7, borderRadius: 4, background: theme.accentQuantity, flexShrink: 0}} />
          <div style={{display: 'flex', flexDirection: 'column', gap: 6}}>
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 15,
                letterSpacing: 3,
                textTransform: 'uppercase',
                color: theme.accentQuantity,
              }}
            >
              the one change
            </div>
            <div style={{fontFamily: theme.fontDisplay, fontSize: 40, fontWeight: 700, color: theme.text, lineHeight: 1.2}}>
              {fix}
            </div>
          </div>
        </div>
      ) : null}
    </Stage>
  );
};
