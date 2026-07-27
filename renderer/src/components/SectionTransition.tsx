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

/**
 * How one beat gives way to the next.
 *
 * Only three, and only because each answers a different question about what the
 * cut *means*. Most templates never see more than `rise`: the accumulating ones
 * — whiteboard, flow, workspace, data — are a single scene for the whole clip,
 * so the only boundary they have is the title card handing over.
 *
 *   rise  the house dissolve: fade, lift and settle. The default, and right
 *         whenever the next shot continues the last one.
 *   push  the outgoing shot leaves and the incoming one arrives from the other
 *         side. For `illustration` and `cast`, where the figure already
 *         alternates sides every beat — the movement reinforces the alternation
 *         instead of ignoring it.
 *   cut   short, flat, no movement. `story` is directed shot by shot with a
 *         real camera; a rising cross-dissolve between two framed shots is the
 *         one transition a film would never use there.
 */
export type CutStyle = 'rise' | 'push' | 'cut';

export const SectionTransition: React.FC<{
  /** Frames the scene actually owns — the exit begins after this. */
  durationInFrames: number;
  /** Overlap window granted by LessonVideo; the exit fades across it. */
  crossFrames?: number;
  motion?: MotionTokens;
  isLast?: boolean;
  cutStyle?: CutStyle;
  children: React.ReactNode;
}> = ({durationInFrames, crossFrames, motion, isLast, cutStyle = 'rise', children}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const m = resolveMotion(motion);
  const inFrames = Math.max(1, secondsToFrames(fps, m.timing.normal)) || FALLBACK_FRAMES;
  const outFrames = Math.max(1, crossFrames ?? secondsToFrames(fps, m.timing.normal));
  const entrance = bezierEasing(m.easing.entrance);
  const exit = bezierEasing(m.easing.exit);

  // A cut is a cut: it takes the shortest window the motion language offers
  // rather than the normal one, so the two shots are not both half-visible for
  // a third of a second.
  const inSpan = cutStyle === 'cut' ? Math.max(1, secondsToFrames(fps, m.timing.fast)) : inFrames;

  const inP = interpolate(frame, [0, inSpan], [0, 1], {
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

  // Displacement is deliberately small in every style. The frame is 1920 wide
  // and a push that actually crosses it reads as a slideshow control; what sells
  // the direction is that the two shots move the *same* way, not how far.
  let dx = 0;
  let dy = 0;
  let scale = 1;
  switch (cutStyle) {
    case 'push':
      dx = (1 - inP) * 90 - (1 - outP) * 70;
      scale = 0.99 + 0.01 * inP;
      break;
    case 'cut':
      // Nothing moves. A framed shot that drifts on arrival is a shot the
      // camera has not settled on yet, which contradicts the whole template.
      break;
    default:
      dy = (1 - inP) * 34 - (1 - outP) * 22;
      scale = 0.985 + 0.015 * inP;
  }

  return (
    <AbsoluteFill style={{opacity, transform: `translate(${dx}px, ${dy}px) scale(${scale})`}}>
      {children}
    </AbsoluteFill>
  );
};
