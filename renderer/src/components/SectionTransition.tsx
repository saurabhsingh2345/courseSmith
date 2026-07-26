import {AbsoluteFill, interpolate, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {MotionTokens, bezierEasing, resolveMotion, secondsToFrames} from '../theme/motion';

// SectionTransition wraps every scene with its entrance/exit: a fade paired
// with a gentle rise and settle-scale on the way in, and a fade + drift on
// the way out. Curves and durations come from the motion tokens.
//
// The exit runs *past* durationInFrames, inside the cross-dissolve overlap
// LessonVideo grants each sequence. That is what makes it a dissolve rather
// than a blackout: the scene holds full opacity for every frame it actually
// owns, then fades out underneath the next scene, which is fading in on top
// over the same window. (Previously the exit ate the scene's own final frames
// while the next sequence had not started yet, so every boundary flashed the
// bare background.)

const FALLBACK_FRAMES = Math.round(FPS / 2);

export const SectionTransition: React.FC<{
  /** Frames the scene actually owns — the exit begins after this. */
  durationInFrames: number;
  /** Overlap window granted by LessonVideo; the exit fades across it. */
  crossFrames?: number;
  motion?: MotionTokens;
  isLast?: boolean;
  children: React.ReactNode;
}> = ({durationInFrames, crossFrames, motion, isLast, children}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const m = resolveMotion(motion);
  const inFrames = Math.max(1, secondsToFrames(fps, m.timing.normal)) || FALLBACK_FRAMES;
  const outFrames = Math.max(1, crossFrames ?? secondsToFrames(fps, m.timing.normal));
  const entrance = bezierEasing(m.easing.entrance);
  const exit = bezierEasing(m.easing.exit);

  const inP = interpolate(frame, [0, inFrames], [0, 1], {
    easing: entrance,
    extrapolateRight: 'clamp',
  });
  // The very last scene holds instead of fading to nothing.
  const outP = isLast
    ? 1
    : interpolate(frame, [durationInFrames, durationInFrames + outFrames], [1, 0], {
        easing: exit,
        extrapolateLeft: 'clamp',
        extrapolateRight: 'clamp',
      });

  const opacity = Math.min(inP, outP);
  const rise = (1 - inP) * 34 - (1 - outP) * 22;
  const scale = 0.985 + 0.015 * inP;

  return (
    <AbsoluteFill style={{opacity, transform: `translateY(${rise}px) scale(${scale})`}}>
      {children}
    </AbsoluteFill>
  );
};
