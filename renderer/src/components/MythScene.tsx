import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';

// MythScene is a belief crossed out and replaced.
//
// The single gesture this component exists to draw is the strike: a rule that
// sweeps left to right through the claim while the claim itself desaturates and
// shrinks, and the truth rises into the space it vacates. Everything else on
// screen is subordinate to that half-second.
//
// Three details make it read as a retraction rather than as a transition.
//
// The strike is drawn as a growing element, not toggled on. A line that appears
// fully formed reads as a design flourish; one that travels across the words
// reads as somebody crossing them out, which is the meaning being conveyed.
//
// The claim stays on screen afterwards, struck and dimmed, for the whole rest
// of the clip. Removing it would let the viewer's original belief quietly
// return — the point of the frame is that the wrong idea is visibly still
// there, and visibly cancelled.
//
// The claim is set in quotation marks and the truth is not. The claim is
// somebody's words being reported; the truth is the video speaking. That single
// typographic difference does more than any label would.

const COL_W = Math.min(STAGE_W, 1500);

type Step = {startMs: number; endMs: number; show: 'claim' | 'strike' | 'evidence' | 'why'; at?: number};

export const MythScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const claim = String(props.claim ?? '');
  const truth = String(props.truth ?? '');
  const why = String(props.why ?? '');
  const evidence = (Array.isArray(props.evidence) ? props.evidence : []) as string[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (!claim || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // How far through the strike we are. Before the strike beat it is 0; on it,
  // it travels; after it, it is done and stays done.
  const strikeIdx = steps.findIndex((s) => s.show === 'strike');
  const struck = strikeIdx >= 0 && idx > strikeIdx;
  const striking = step.show === 'strike';
  const sweep = striking
    ? interpolate(sinceStep, [4, 22], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : struck
      ? 1
      : 0;
  // The truth follows the rule across rather than arriving with it, so the eye
  // reads "cancelled" before it reads "instead".
  const reveal = striking
    ? interpolate(sinceStep, [18, 36], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : struck
      ? 1
      : 0;

  const enter = spring({
    frame: ((nowMs - steps[0].startMs) / 1000) * FPS,
    fps,
    config: {damping: 200, mass: 0.7},
    durationInFrames: 18,
  });

  const currentEvidence = step.show === 'evidence' ? (step.at ?? 0) : -1;
  const onWhy = step.show === 'why';

  return (
    <Stage justify="center">
      <div style={{width: COL_W, opacity: enter, textAlign: 'center'}}>
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 18,
            letterSpacing: 5,
            textTransform: 'uppercase',
            color: theme.textMuted,
            marginBottom: 20,
          }}
        >
          What everyone says
        </div>

        {/* The claim, with the strike drawn through it. */}
        <div style={{position: 'relative', display: 'inline-block'}}>
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 74,
              fontWeight: 700,
              letterSpacing: -1.6,
              lineHeight: 1.14,
              // Desaturates and settles back as it is cancelled, rather than
              // simply fading — a claim that fades looks forgotten, one that
              // greys out looks overruled.
              color: sweep > 0 ? theme.textMuted : theme.text,
              opacity: 1 - sweep * 0.42,
              transform: `scale(${1 - sweep * 0.06})`,
            }}
          >
            &ldquo;{claim}&rdquo;
          </div>
          {/* The rule itself: drawn as width, so it travels. */}
          <div
            style={{
              position: 'absolute',
              left: 0,
              top: '52%',
              height: 6,
              borderRadius: 3,
              width: `${sweep * 100}%`,
              background: theme.accentLimit,
              opacity: sweep > 0 ? 1 : 0,
            }}
          />
        </div>

        {/* The truth, rising into the space the claim vacated. Unquoted: the
            claim is somebody's words being reported, this is the video
            speaking. */}
        {reveal > 0 ? (
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 86,
              fontWeight: 800,
              letterSpacing: -2,
              lineHeight: 1.08,
              color: theme.accentQuantity,
              marginTop: 40,
              opacity: reveal,
              transform: `translateY(${(1 - reveal) * 34}px)`,
            }}
          >
            {truth}
          </div>
        ) : null}

        {/* Evidence, and the concession. Only one is up at a time — they are
            different moves and stacking them would blunt both. */}
        {currentEvidence >= 0 && evidence[currentEvidence] ? (
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 18,
              marginTop: 46,
              padding: '20px 32px',
              borderRadius: 12,
              background: withAlpha(theme.accentQuantity, 0.1),
              border: `1px solid ${withAlpha(theme.accentQuantity, 0.34)}`,
              opacity: interpolate(sinceStep, [3, 17], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
              transform: `translateY(${
                (1 -
                  interpolate(sinceStep, [3, 17], [0, 1], {
                    extrapolateLeft: 'clamp',
                    extrapolateRight: 'clamp',
                  })) *
                14
              }px)`,
            }}
          >
            <span style={{fontFamily: theme.fontMono, fontSize: 22, color: theme.accentQuantity}}>
              {currentEvidence + 1}/{evidence.length}
            </span>
            <span style={{fontFamily: theme.fontBody, fontSize: 32, color: theme.text}}>
              {evidence[currentEvidence]}
            </span>
          </div>
        ) : null}

        {onWhy && why ? (
          <div
            style={{
              marginTop: 46,
              opacity: interpolate(sinceStep, [3, 18], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 16,
                letterSpacing: 3.6,
                textTransform: 'uppercase',
                color: theme.accentRival,
                marginBottom: 14,
              }}
            >
              Why everyone thought so
            </div>
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 30,
                lineHeight: 1.4,
                color: theme.textMuted,
                maxWidth: 1080,
                margin: '0 auto',
              }}
            >
              {why}
            </div>
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
