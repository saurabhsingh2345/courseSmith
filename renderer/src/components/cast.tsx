// The character rig.
//
// A flat-vector person built from a skeleton rather than from artwork, for the
// same reason the figures in artwork.tsx are drawn rather than imported: a
// character has to be able to *do* something. A downloaded illustration of a
// person pointing is a person pointing forever. Here a pose is a set of joint
// angles, poses interpolate, and the same character can point, then think, then
// throw its hands up — which is the difference between a cast and a sticker
// sheet.
//
// Everything is forward kinematics from the hips, drawn as tapered capsules.
// The rig is deliberately small: shoulders, elbows, hips, knees, a head tilt
// and a torso lean. That is enough for every pose an explainer needs, and it is
// few enough joints that a pose can be written by hand and read at a glance.
//
// Angles are degrees from straight down and **outward-positive on both sides**:
// the rig negates the left, so `lShoulder: 40` and `rShoulder: 40` both swing
// the arm away from the body. Elbow and knee angles are relative to their
// parent limb, so bending an elbow is one number that means the same thing
// whatever the shoulder is doing.
//
// The first version used raw screen-clockwise angles for both sides, which
// means a positive left-arm angle swings *across the chest*. Every symmetric
// pose came out with the arms crossed and the walk cycle scissored its own
// legs. Poses are authored by hand, so the convention has to be the one that
// makes a hand-written pose mean what it looks like it means.
//
// Nothing here reads a clock or a random source; frames stay independent.

import {random} from 'remotion';

export type CastPalette = {
  skin: string;
  hair: string;
  top: string;
  bottom: string;
  /** Line colour for features (mouth, brows). */
  ink: string;
};

export type Pose = {
  lShoulder: number;
  lElbow: number;
  rShoulder: number;
  rElbow: number;
  lHip: number;
  lKnee: number;
  rHip: number;
  rKnee: number;
  headTilt: number;
  torsoLean: number;
  /** Vertical offset of the whole body, for a crouch or a jump. */
  lift: number;
};

export type Expression = 'neutral' | 'happy' | 'thinking' | 'surprised' | 'concerned';

/** The box every character is drawn in. */
export const CAST_BOX = {w: 200, h: 260};

/**
 * A tighter crop of that box, for scenes that want the character to fill its
 * column.
 *
 * The drawing box has to be wide enough for `celebrate` and `shrug`, whose arms
 * reach far past the shoulders. A standing figure only occupies the middle
 * ~40% of it, so a scene that sizes an svg to CAST_BOX gets a small person
 * surrounded by air. This crop keeps the widest pose in frame while throwing
 * away the margin an idle one leaves.
 */
export const CAST_VIEW = {x: 24, y: 12, w: 152, h: 244};

/**
 * Where the soles are, as a fraction of CAST_VIEW's height.
 *
 * A scene that stands a character on a floor line has to know this. Treating
 * the bottom of the drawing box as the feet leaves the figure hovering by the
 * margin the box carries for `celebrate`'s raised arms — which reads, exactly,
 * as a character floating.
 */
export const CAST_FEET = (239 - 12) / 244;

// Skeleton proportions, in the 200x260 box.
const RIG = {
  hipY: 152,
  hipSpread: 13,
  shoulderY: 94,
  // Wider than the first pass. Narrow shoulders under a 25px head made the
  // head and neck dominate the figure and read as a bobblehead.
  shoulderSpread: 26,
  headY: 56,
  headR: 24,
  upperArm: 33,
  foreArm: 31,
  thigh: 40,
  shin: 38,
  neckY: 84,
} as const;

const rad = (deg: number): number => (deg * Math.PI) / 180;

/** Endpoint of a bone of length `len` leaving `from` at absolute angle `deg`. */
const bone = (from: {x: number; y: number}, deg: number, len: number) => ({
  x: from.x + len * Math.sin(rad(deg)),
  y: from.y + len * Math.cos(rad(deg)),
});

/**
 * The pose vocabulary.
 *
 * A person standing perfectly symmetrically reads as a mannequin, so every
 * pose here is asymmetric somewhere — even `idle`, whose arms hang at slightly
 * different angles. It is the cheapest thing that makes a rig look alive.
 */
export const POSES: Record<string, Pose> = {
  idle: {
    lShoulder: 11, lElbow: 6, rShoulder: 8, rElbow: 9,
    lHip: 3, lKnee: 1, rHip: 3, rKnee: 2,
    headTilt: -1, torsoLean: 0, lift: 0,
  },
  // Pointing at whatever is beside the character — the workhorse pose for an
  // explainer, since it ties the person to the thing being explained. The
  // pointing arm is the right one so that at facing=1 it points off-frame
  // right, where the template puts the subject.
  point: {
    lShoulder: 12, lElbow: 6, rShoulder: 88, rElbow: 4,
    lHip: 3, lKnee: 1, rHip: 4, rKnee: 1,
    headTilt: 4, torsoLean: 2, lift: 0,
  },
  // One hand to the chin. The elbow has to come up before the forearm can
  // reach the face at all — the forearm is shorter than the shoulder-to-chin
  // distance, so a low elbow simply cannot get there.
  think: {
    lShoulder: 14, lElbow: 8, rShoulder: 148, rElbow: 142,
    lHip: 4, lKnee: 2, rHip: 5, rKnee: 3,
    headTilt: -7, torsoLean: -2, lift: 0,
  },
  wave: {
    lShoulder: 11, lElbow: 6, rShoulder: 142, rElbow: 30,
    lHip: 3, lKnee: 1, rHip: 3, rKnee: 2,
    headTilt: 5, torsoLean: 1, lift: 0,
  },
  celebrate: {
    lShoulder: 152, lElbow: 22, rShoulder: 152, rElbow: 22,
    lHip: 7, lKnee: 3, rHip: 7, rKnee: 3,
    headTilt: 0, torsoLean: 0, lift: -6,
  },
  shrug: {
    lShoulder: 62, lElbow: 78, rShoulder: 62, rElbow: 78,
    lHip: 4, lKnee: 2, rHip: 4, rKnee: 2,
    headTilt: 3, torsoLean: 0, lift: 2,
  },
  // Hands on hips. This started out as "carry" — hands together in front —
  // and the elbow angle that actually reads well put them on the hips instead.
  // A confident stance is the more useful pose for an explainer anyway, so it
  // is named for what it is rather than for what it was aiming at.
  confident: {
    lShoulder: 58, lElbow: -114, rShoulder: 58, rElbow: -114,
    lHip: 4, lKnee: 2, rHip: 4, rKnee: 2,
    headTilt: 2, torsoLean: -2, lift: 0,
  },
  // Slumped: the "this is the problem" pose.
  defeated: {
    lShoulder: 14, lElbow: 3, rShoulder: 12, rElbow: 3,
    lHip: 3, lKnee: 7, rHip: 3, rKnee: 7,
    headTilt: 13, torsoLean: 9, lift: 5,
  },
  // Mid-stride: one leg forward, opposite arm forward. Forward and back are
  // opposite signs on the same side, which is exactly what mirroring swaps.
  walk: {
    lShoulder: -26, lElbow: -12, rShoulder: 26, rElbow: 12,
    lHip: 24, lKnee: 4, rHip: -22, rKnee: 24,
    headTilt: 0, torsoLean: -3, lift: 0,
  },
};

export const POSE_NAMES = Object.keys(POSES);

/** The mirror of a pose, for the second half of a walk cycle. */
const mirrorPose = (p: Pose): Pose => ({
  lShoulder: p.rShoulder, lElbow: p.rElbow, rShoulder: p.lShoulder, rElbow: p.lElbow,
  lHip: p.rHip, lKnee: p.rKnee, rHip: p.lHip, rKnee: p.lKnee,
  headTilt: -p.headTilt, torsoLean: p.torsoLean, lift: p.lift,
});

/** Linear blend of two poses. */
export const blendPose = (a: Pose, b: Pose, t: number): Pose => {
  const m = (x: number, y: number) => x + (y - x) * t;
  return {
    lShoulder: m(a.lShoulder, b.lShoulder), lElbow: m(a.lElbow, b.lElbow),
    rShoulder: m(a.rShoulder, b.rShoulder), rElbow: m(a.rElbow, b.rElbow),
    lHip: m(a.lHip, b.lHip), lKnee: m(a.lKnee, b.lKnee),
    rHip: m(a.rHip, b.rHip), rKnee: m(a.rKnee, b.rKnee),
    headTilt: m(a.headTilt, b.headTilt), torsoLean: m(a.torsoLean, b.torsoLean),
    lift: m(a.lift, b.lift),
  };
};

export const poseByName = (name: string | undefined): Pose =>
  (name && POSES[name]) || POSES.idle;

/**
 * The pose actually drawn on a given frame: the requested pose, plus the idle
 * motion that runs underneath every pose.
 *
 * The breathing is what stops a held pose reading as a freeze-frame — the same
 * lesson the figures taught. A walking character additionally swings through
 * the mirrored stride rather than holding one leg up.
 */
export const livePose = (base: Pose, t: number, walking: boolean, seed: string): Pose => {
  if (walking) {
    // A full stride is two steps, so the cycle runs base -> mirror -> base.
    const phase = (t * 1.6) % 1;
    const swing = Math.sin(phase * Math.PI * 2) * 0.5 + 0.5;
    return blendPose(base, mirrorPose(base), swing);
  }
  // Breathing: a slow rise and fall through the shoulders, plus a barely-there
  // sway. Seeded so two characters on stage are not in lockstep.
  const off = random(`${seed}-breath`) * Math.PI * 2;
  const breath = Math.sin(t * 1.5 + off);
  return {
    ...base,
    lShoulder: base.lShoulder + breath * 1.1,
    rShoulder: base.rShoulder - breath * 1.1,
    headTilt: base.headTilt + breath * 0.8,
    torsoLean: base.torsoLean + Math.sin(t * 0.9 + off) * 0.7,
    lift: base.lift + breath * 0.9,
  };
};

/** Mouth path for an expression, in head-local coordinates. */
const mouthPath = (e: Expression): string => {
  switch (e) {
    case 'happy':
      return 'M-9 4 Q0 13 9 4';
    case 'concerned':
      return 'M-8 8 Q0 1 8 8';
    case 'surprised':
      return 'M-5 3 Q0 3 5 3 Q7 8 0 10 Q-7 8 -5 3 Z';
    case 'thinking':
      return 'M-7 6 Q1 3 8 7';
    default:
      return 'M-7 6 L7 6';
  }
};

/** Brow angle (degrees) for an expression; mirrored for the other brow. */
const browTilt = (e: Expression): number => {
  switch (e) {
    case 'concerned':
      return -16;
    case 'surprised':
      return 6;
    case 'thinking':
      return -9;
    default:
      return 0;
  }
};

export type CharacterProps = {
  pose: Pose;
  expression: Expression;
  palette: CastPalette;
  /** Seconds since the scene started; drives blinking. */
  t: number;
  /** Which way the character faces. */
  facing?: 1 | -1;
  /** Seeds the blink phase, so a crowd does not blink in unison. */
  seed?: string;
};

export const Character: React.FC<CharacterProps> = ({
  pose, expression, palette, t, facing = 1, seed = 'cast',
}) => {
  const lean = pose.torsoLean;
  const lift = pose.lift;

  // Joint chain. The torso lean rotates the shoulders about the hips, so the
  // arms follow the body instead of floating beside it.
  const hipC = {x: 100, y: RIG.hipY + lift};
  const shoulderC = bone(hipC, 180 + lean, RIG.hipY - RIG.shoulderY);
  const neck = bone(hipC, 180 + lean, RIG.hipY - RIG.neckY);
  const headC = bone(hipC, 180 + lean, RIG.hipY - RIG.headY);

  const shoulderL = {x: shoulderC.x - RIG.shoulderSpread, y: shoulderC.y};
  const shoulderR = {x: shoulderC.x + RIG.shoulderSpread, y: shoulderC.y};
  const hipL = {x: hipC.x - RIG.hipSpread, y: hipC.y};
  const hipR = {x: hipC.x + RIG.hipSpread, y: hipC.y};

  // Outward-positive: the left side's angles are negated here, once, so every
  // pose in the table can be written as if both limbs swung the same way.
  const elbowL = bone(shoulderL, -pose.lShoulder + lean, RIG.upperArm);
  const handL = bone(elbowL, -(pose.lShoulder + pose.lElbow) + lean, RIG.foreArm);
  const elbowR = bone(shoulderR, pose.rShoulder + lean, RIG.upperArm);
  const handR = bone(elbowR, pose.rShoulder + pose.rElbow + lean, RIG.foreArm);

  const kneeL = bone(hipL, -pose.lHip, RIG.thigh);
  const footL = bone(kneeL, -(pose.lHip + pose.lKnee), RIG.shin);
  const kneeR = bone(hipR, pose.rHip, RIG.thigh);
  const footR = bone(kneeR, pose.rHip + pose.rKnee, RIG.shin);

  // Blink: shut for a fraction of a second every few seconds, off its own
  // phase so a group does not blink together.
  const blinkPeriod = 3.4 + random(`${seed}-blinkp`) * 2.2;
  const blinkPhase = (t + random(`${seed}-blink`) * blinkPeriod) % blinkPeriod;
  const blinking = blinkPhase < 0.13;

  const limb = (
    a: {x: number; y: number},
    b: {x: number; y: number},
    width: number,
    color: string,
    key: string,
  ) => (
    <line
      key={key}
      x1={a.x} y1={a.y} x2={b.x} y2={b.y}
      stroke={color}
      strokeWidth={width}
      strokeLinecap="round"
    />
  );

  return (
    <g transform={facing === -1 ? 'translate(200 0) scale(-1 1)' : undefined}>
      {/* Legs behind the torso, so the hip join is hidden by the shirt. */}
      {limb(hipL, kneeL, 15, palette.bottom, 'thighL')}
      {limb(kneeL, footL, 13, palette.bottom, 'shinL')}
      {limb(hipR, kneeR, 15, palette.bottom, 'thighR')}
      {limb(kneeR, footR, 13, palette.bottom, 'shinR')}
      {/* Shoes */}
      <ellipse cx={footL.x + 3} cy={footL.y + 3} rx={11} ry={6} fill={palette.ink} opacity={0.85} />
      <ellipse cx={footR.x + 3} cy={footR.y + 3} rx={11} ry={6} fill={palette.ink} opacity={0.85} />

      {/* Torso: a tapered slab from hips to shoulders. */}
      <path
        d={`M${hipL.x} ${hipL.y} L${shoulderL.x - 3} ${shoulderL.y} Q100 ${shoulderC.y - 8} ${shoulderR.x + 3} ${shoulderR.y} L${hipR.x} ${hipR.y} Q100 ${hipC.y + 9} ${hipL.x} ${hipL.y} Z`}
        fill={palette.top}
      />

      {/* Arms over the torso. */}
      {limb(shoulderL, elbowL, 13, palette.top, 'upperL')}
      {limb(elbowL, handL, 11, palette.skin, 'foreL')}
      {limb(shoulderR, elbowR, 13, palette.top, 'upperR')}
      {limb(elbowR, handR, 11, palette.skin, 'foreR')}
      <circle cx={handL.x} cy={handL.y} r={7} fill={palette.skin} />
      <circle cx={handR.x} cy={handR.y} r={7} fill={palette.skin} />

      {/* Neck */}
      {limb(neck, shoulderC, 13, palette.skin, 'neck')}

      {/* Head, tilted about the neck. */}
      <g transform={`rotate(${pose.headTilt} ${neck.x} ${neck.y})`}>
        <circle cx={headC.x} cy={headC.y} r={RIG.headR} fill={palette.skin} />
        {/* Hair as a cap over the top of the skull. */}
        <path
          d={`M${headC.x - RIG.headR} ${headC.y - 2}
              A${RIG.headR} ${RIG.headR} 0 0 1 ${headC.x + RIG.headR} ${headC.y - 2}
              Q${headC.x + RIG.headR - 4} ${headC.y - 13} ${headC.x} ${headC.y - 12}
              Q${headC.x - RIG.headR + 4} ${headC.y - 13} ${headC.x - RIG.headR} ${headC.y - 2} Z`}
          fill={palette.hair}
        />
        <g transform={`translate(${headC.x} ${headC.y})`}>
          {/* Eyes. A blink is a line, not a smaller circle — a shrinking pupil
              reads as a zoom, a flat lid reads as a blink. */}
          {blinking ? (
            <>
              <line x1={-13} y1={0} x2={-5} y2={0} stroke={palette.ink} strokeWidth={2.6} strokeLinecap="round" />
              <line x1={5} y1={0} x2={13} y2={0} stroke={palette.ink} strokeWidth={2.6} strokeLinecap="round" />
            </>
          ) : (
            <>
              <circle cx={-9} cy={0} r={3.2} fill={palette.ink} />
              <circle cx={9} cy={0} r={3.2} fill={palette.ink} />
            </>
          )}
          {browTilt(expression) !== 0 && (
            <>
              <line
                x1={-14} y1={-9} x2={-4} y2={-9}
                stroke={palette.ink} strokeWidth={2.4} strokeLinecap="round"
                transform={`rotate(${browTilt(expression)} -9 -9)`}
              />
              <line
                x1={4} y1={-9} x2={14} y2={-9}
                stroke={palette.ink} strokeWidth={2.4} strokeLinecap="round"
                transform={`rotate(${-browTilt(expression)} 9 -9)`}
              />
            </>
          )}
          <path
            d={mouthPath(expression)}
            fill={expression === 'surprised' ? palette.ink : 'none'}
            stroke={palette.ink}
            strokeWidth={2.4}
            strokeLinecap="round"
          />
        </g>
      </g>
    </g>
  );
};
