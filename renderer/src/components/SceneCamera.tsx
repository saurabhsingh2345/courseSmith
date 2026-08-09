import {AbsoluteFill, interpolate, useCurrentFrame, useVideoConfig} from 'remotion';
import {MotionTokens, bezierEasing, resolveMotion} from '../theme/motion';
import {FRAME_W} from './Stage';

// The camera: a slow move underneath every scene.
//
// == Why this exists ==
//
// The catalog had no camera, and its absence is most of why a finished clip read
// as a slide deck rather than as a film. Every scene animated its entrances —
// things faded and rose into place over the first half second — and then held
// perfectly still for the remaining twenty or fifty seconds, with only opacity
// changing to mark which item was being spoken about. A held frame with a voice
// over it is a slide, however well it is set.
//
// So the frame is never quite still. Over the first eighteen seconds of a scene
// it creeps closer and travels slightly sideways, then holds — which is what
// makes a shot feel photographed rather than printed.
//
// == Why it is central rather than per-scene ==
//
// The same argument the skin's `air` makes. Forty-odd scene components compose
// into the timeline, and a camera each one opted into would be a camera
// thirty-eight of them had and two did not — and the two without it would read as
// broken rather than as still. Wrapping the scene here means a template added
// tomorrow inherits the move without knowing it exists.
//
// == Why so small, and why it settles ==
//
// The effect has to be invisible frame to frame and obvious in aggregate. 3.2%
// over eighteen seconds is about 0.18% a second: no single second reads as
// movement, and a still from the top of a beat and one from the bottom are
// different shots. Past about 4% it becomes a zoom the viewer notices and then
// resents, and content starts crossing the safe margin.
//
// It settles rather than running the length of the scene because the rate is what
// the eye reads. Spread across a seventy-second segment the same 3% is 0.03% a
// second, which measured out as twenty-eight pixels of travel over seventy
// seconds — nothing. A push that reaches its mark and holds is a real camera move;
// a crawl too slow to see is just an expensive no-op.
//
// Direction alternates by scene index. A run of eight segments all pushing in and
// drifting left is a tic the eye picks up by the third one, and a tic is worse
// than stillness because it draws attention to the mechanism. Alternating reads
// as coverage — different shots of the same subject.
//
// == What it must not break ==
//
// The scale is applied to a wrapper, not to any scene's own transforms, so the
// scenes that already move (the editor window scaling up, the terminal drawer
// growing, the memory layout) compose with it rather than fighting it. And it
// only ever scales UP from 1, never below — starting under 1 would letterbox the
// background gradient, which is painted behind this and does not move.

export const SceneCamera: React.FC<{
  /** Frames this scene is on screen for. Only used to cut the move short on a
   *  scene briefer than the settle time — the move's pace comes from the token,
   *  not from the scene length. */
  durationInFrames: number;
  motion?: Partial<MotionTokens> | null;
  /** Scene position in the timeline, used only to alternate direction. */
  index: number;
  children: React.ReactNode;
}> = ({durationInFrames, motion, index, children}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const m = resolveMotion(motion);
  const {push, drift} = m.camera;

  // Nothing to do, and nothing to pay for: a scene graph that zeroes the camera
  // (or an archetype that wants stillness) renders the children untouched rather
  // than wrapped in an identity transform.
  if (!push && !drift) return <>{children}</>;

  // The move has its own duration and then the frame holds — it is NOT paced
  // across the scene.
  //
  // Pacing it across the scene was the first implementation and it produced no
  // motion at all on the scenes that needed it most. A combo segment runs sixty to
  // a hundred seconds, so spreading 3% over one gives 0.03% a second: measured on
  // a rendered frame, content moved twenty-eight pixels in seventy seconds. That
  // is not a camera, it is a rounding error. The rate is what the eye reads, not
  // the total, so the rate is what the token now fixes.
  const settleFrames = Math.max(1, Math.round((m.camera.settleSec || 18) * fps));
  // `subtle` rather than `entrance`: the entrance curve overshoots, which is
  // right for something arriving and wrong for something that must never appear
  // to arrive at all.
  const p = interpolate(frame, [0, Math.min(settleFrames, Math.max(1, durationInFrames))], [0, 1], {
    easing: bezierEasing(m.easing.subtle),
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  // Alternate: even scenes push in, odd scenes pull back from a hair closer. Both
  // stay at or above scale 1 so the gradient behind is never exposed.
  const inward = index % 2 === 0;
  const scale = inward ? 1 + push * p : 1 + push * (1 - p);
  // Sideways travel as a fraction of the frame, so the token reads in pixels of a
  // 1920 frame regardless of what the composition is actually rendered at.
  const dir = index % 4 < 2 ? 1 : -1;
  const x = ((inward ? p : 1 - p) - 0.5) * 2 * (drift / FRAME_W) * 100 * dir;

  return (
    <AbsoluteFill
      style={{
        transform: `scale(${scale}) translateX(${x}%)`,
        transformOrigin: 'center center',
      }}
    >
      {children}
    </AbsoluteFill>
  );
};
