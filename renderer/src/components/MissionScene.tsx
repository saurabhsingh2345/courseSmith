import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// MissionScene: the assignment, as a card you could hand someone.
//
// Practice work in a video course usually arrives as a paragraph the narrator
// reads out, which is the least actionable form it could take: by the time the
// fourth requirement is spoken the first one is gone. So the brief becomes an
// object — one card, goal at the top, a numbered spec list building down it,
// the artifact named in mono, and the definition of done stamped across the
// bottom. A viewer can pause on the last frame and have the entire assignment.
//
// The specs build DOWN, one per beat, each with its number already drawn and
// waiting in the rail. The empty numbers are visible from the first frame,
// deliberately: a viewer who can see there are five slots knows how much is
// coming, and a checklist whose length is a surprise is a checklist people
// abandon. Landing a spec fills its number and slides the text in from the
// rail, so the motion always travels the same direction the list reads.
//
// The deliverable is a mono chip and nothing else, because it is a filename or
// an artifact type — a thing, not a sentence — and mono is how this catalog
// says "thing".
//
// The DONE-WHEN strip is the only element that arrives by stamp: it scales
// down onto the card from slightly oversize with a hard settle. Everything
// else has been building; this one lands. That contrast is what makes it read
// as the closing condition rather than a sixth requirement.
//
// One glow maximum: the done strip on its stamp.

const CARD_W = Math.min(STAGE_W, 1240);
const RAIL_W = 62;

type Step = {
  startMs: number;
  endMs: number;
  show: 'brief' | 'spec' | 'deliverable' | 'done';
  at?: number;
  landed: number[];
  deliverableOn: boolean;
  doneOn: boolean;
};

export const MissionScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const goal = String(props.goal ?? '');
  const deliverable = String(props.deliverable ?? '');
  const done = String(props.done ?? '');
  const specs = (Array.isArray(props.specs) ? props.specs : []) as string[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (specs.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const landing = step.show === 'spec' ? step.at ?? -1 : -1;
  const landed = new Set(Array.isArray(step.landed) ? step.landed : []);
  const deliverableOn = Boolean(step.deliverableOn);
  const doneOn = Boolean(step.doneOn);

  const arrive = spring({frame: sinceStep - 2, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 24});
  const chip = deliverableOn
    ? step.show === 'deliverable'
      ? spring({frame: sinceStep - 2, fps, config: {damping: 13, mass: 0.5}, durationInFrames: 24})
      : 1
    : 0;
  // The stamp: a slightly harder spring than anything else on the card, so it
  // reads as an impact rather than a reveal.
  const stamp = doneOn
    ? spring({frame: sinceStep - 2, fps, config: {damping: 12, mass: 0.5}, durationInFrames: 22})
    : 0;
  const goalIn = interpolate(sinceStep, [2, 18], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});
  const goalUp = idx === 0 ? goalIn : 1;

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

      <div
        style={{
          width: CARD_W,
          padding: 34,
          borderRadius: 20,
          background: withAlpha(theme.surface, 0.75),
          border: `2px solid ${doneOn ? withAlpha(theme.accent, 0.3 + 0.25 * stamp) : theme.surfaceBorder}`,
          boxShadow: `10px 12px 0 ${withAlpha(theme.ink, 0.38)}`,
        }}
      >
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 14,
            letterSpacing: 3.4,
            textTransform: 'uppercase',
            color: theme.accentText,
            marginBottom: 12,
            opacity: goalUp,
          }}
        >
          your mission
        </div>
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 48,
            fontWeight: 700,
            letterSpacing: -1.1,
            lineHeight: 1.13,
            color: theme.text,
            opacity: goalUp,
            transform: `translateY(${(1 - goalUp) * 10}px)`,
          }}
        >
          {goal}
        </div>

        <div
          style={{
            marginTop: 26,
            paddingTop: 22,
            borderTop: `2px solid ${withAlpha(theme.line, 0.2)}`,
            display: 'flex',
            flexDirection: 'column',
            gap: 14,
          }}
        >
          {specs.map((spec, i) => {
            const isLanding = i === landing;
            const on = landed.has(i);
            const pop = isLanding ? arrive : 1;
            return (
              <div key={i} style={{display: 'flex', alignItems: 'center', gap: 20}}>
                <div
                  style={{
                    width: RAIL_W - 16,
                    height: 42,
                    flexShrink: 0,
                    borderRadius: 9,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontFamily: theme.fontMono,
                    fontSize: 19,
                    fontWeight: 700,
                    background: on ? withAlpha(theme.accent, 0.85 * (isLanding ? pop : 1)) : 'transparent',
                    border: `2px solid ${on ? withAlpha(theme.accent, 0.85) : withAlpha(theme.surfaceBorder, 0.85)}`,
                    color: on ? theme.ink : withAlpha(theme.textMuted, 0.7),
                    transform: `scale(${on ? 0.85 + 0.15 * (isLanding ? pop : 1) : 1})`,
                  }}
                >
                  {String(i + 1).padStart(2, '0')}
                </div>
                <div
                  style={{
                    fontFamily: theme.fontBody,
                    fontSize: 29,
                    lineHeight: 1.25,
                    color: on ? theme.text : withAlpha(theme.textMuted, 0.35),
                    opacity: on ? 1 : 0.55,
                    transform: `translateX(${on ? (1 - (isLanding ? pop : 1)) * -18 : -6}px)`,
                  }}
                >
                  {spec}
                </div>
              </div>
            );
          })}
        </div>

        {/* The artifact, named as a thing. */}
        <div
          style={{
            marginTop: 28,
            display: 'flex',
            alignItems: 'center',
            gap: 16,
            opacity: chip,
            transform: `translateY(${(1 - chip) * 10}px)`,
          }}
        >
          <span
            style={{
              fontFamily: theme.fontMono,
              fontSize: 14,
              letterSpacing: 3,
              textTransform: 'uppercase',
              color: theme.textMuted,
            }}
          >
            deliverable
          </span>
          <span
            style={{
              paddingInline: 18,
              paddingBlock: 9,
              borderRadius: 9,
              background: withAlpha(theme.mass, 0.1),
              border: `1.5px solid ${withAlpha(theme.accent, 0.4)}`,
              fontFamily: theme.fontMono,
              fontSize: 25,
              color: theme.text,
            }}
          >
            {deliverable}
          </span>
        </div>

        {/* DONE WHEN — the only thing on this card that stamps. */}
        <div
          style={{
            marginTop: 24,
            display: 'flex',
            alignItems: 'center',
            gap: 18,
            paddingInline: 22,
            paddingBlock: 16,
            borderRadius: 13,
            background: withAlpha(theme.accent, 0.13 * stamp),
            border: `2px solid ${withAlpha(theme.accent, 0.25 + 0.45 * stamp)}`,
            opacity: stamp,
            // Oversize on arrival, settling to true. The one impact.
            transform: `scale(${1.06 - 0.06 * stamp})`,
            boxShadow: `0 0 ${34 * stamp}px ${withAlpha(theme.accent, 0.28)}`,
          }}
        >
          <span
            style={{
              fontFamily: theme.fontMono,
              fontSize: 15,
              fontWeight: 700,
              letterSpacing: 3.2,
              textTransform: 'uppercase',
              color: theme.accentText,
              flexShrink: 0,
            }}
          >
            done when
          </span>
          <span style={{fontFamily: theme.fontBody, fontSize: 28, lineHeight: 1.25, color: theme.text}}>{done}</span>
        </div>
      </div>
    </Stage>
  );
};
