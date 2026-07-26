import {AbsoluteFill, interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {fitText} from '@remotion/layout-utils';
import {ResolvedTheme} from '../theme/theme';
import {MotionTokens, bezierEasing, resolveMotion, secondsToFrames} from '../theme/motion';
import {CAPTION_SAFE} from './Stage';

// TitleCard renders the animated lesson intro (kicker, display heading,
// accent rule, learning outcomes staggering in as icon rows) and section
// heading cards (kicker + per-word heading reveal). Editorial left-aligned
// layout on the shared scene background; timing from the motion tokens.

const CANVAS_W = 1920;
const CONTENT_W = 1460;

const CheckIcon: React.FC<{color: string}> = ({color}) => (
  <svg width="30" height="30" viewBox="0 0 24 24" fill="none">
    <path
      d="M4.5 12.5l5 5 10-11"
      stroke={color}
      strokeWidth="2.6"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

export const TitleCard: React.FC<{
  theme: ResolvedTheme;
  motion?: MotionTokens;
  props: Record<string, unknown>;
}> = ({theme, motion, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const m = resolveMotion(motion);

  const heading = String(props.heading ?? '');
  const subtitle = String(props.subtitle ?? '');
  const intro = Boolean(props.intro);
  const outcomes = Array.isArray(props.outcomes) ? (props.outcomes as string[]) : [];

  const entrance = bezierEasing(m.easing.entrance);
  const itemStagger = secondsToFrames(fps, m.stagger.items);
  const itemDur = secondsToFrames(fps, m.timing.fast);

  // Display size: fill the content column, capped so short titles stay huge
  // and long ones never wrap awkwardly.
  const {fontSize} = fitText({
    text: heading,
    withinWidth: CONTENT_W,
    fontFamily: theme.fontDisplay,
    fontWeight: 700,
  });
  const headingSize = Math.min(intro ? 132 : 108, fontSize);

  const kicker = intro ? theme.courseName : subtitle;
  const kickerOpacity = interpolate(frame, [4, 16], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  // Accent rule draws on after the heading lands.
  const ruleP = interpolate(frame, [16, 16 + secondsToFrames(fps, m.timing.normal)], [0, 1], {
    easing: entrance,
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const words = heading.split(' ');

  return (
    // Left-aligned by design — the one scene that owns the full frame — but it
    // still yields the bottom band to CaptionTrack like every other scene.
    <AbsoluteFill
      style={{
        justifyContent: 'center',
        padding: `0 ${(CANVAS_W - CONTENT_W) / 2}px`,
        paddingBottom: CAPTION_SAFE,
      }}
    >
      <div
        style={{
          fontFamily: theme.fontBody,
          fontSize: 26,
          letterSpacing: 7,
          textTransform: 'uppercase',
          color: theme.accent,
          fontWeight: 600,
          opacity: kickerOpacity,
          marginBottom: 30,
        }}
      >
        {kicker}
      </div>
      <h1
        style={{
          fontFamily: theme.fontDisplay,
          fontSize: headingSize,
          fontWeight: 700,
          letterSpacing: -1.5,
          lineHeight: 1.06,
          margin: 0,
          color: theme.text,
          maxWidth: CONTENT_W,
        }}
      >
        {words.map((w, i) => {
          const s = spring({
            frame: frame - 6 - i * 3,
            fps,
            config: {damping: 200, stiffness: 90},
          });
          return (
            <span
              key={i}
              style={{
                display: 'inline-block',
                opacity: s,
                transform: `translateY(${(1 - s) * 46}px)`,
                marginRight: '0.26em',
              }}
            >
              {w}
            </span>
          );
        })}
      </h1>
      <div
        style={{
          marginTop: 34,
          width: 220 * ruleP,
          height: 7,
          borderRadius: 4,
          background: `linear-gradient(90deg, ${theme.accent}, ${theme.primary})`,
        }}
      />
      {intro && outcomes.length > 0 ? (
        <div style={{marginTop: 58, display: 'flex', flexDirection: 'column', gap: 22}}>
          {outcomes.map((outcome, i) => {
            const delay = 30 + i * itemStagger;
            const p = interpolate(frame, [delay, delay + itemDur], [0, 1], {
              easing: entrance,
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            });
            return (
              <div
                key={i}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 24,
                  opacity: p,
                  transform: `translateX(${(1 - p) * 32}px)`,
                }}
              >
                <div
                  style={{
                    width: 54,
                    height: 54,
                    borderRadius: 15,
                    flexShrink: 0,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    backgroundColor: theme.surface,
                    border: `1.5px solid ${theme.surfaceBorder}`,
                  }}
                >
                  <CheckIcon color={theme.accent} />
                </div>
                <div style={{fontFamily: theme.fontBody, fontSize: 38, fontWeight: 500, color: theme.textMuted}}>
                  {outcome}
                </div>
              </div>
            );
          })}
        </div>
      ) : null}
      {!intro && subtitle ? null : null}
    </AbsoluteFill>
  );
};
