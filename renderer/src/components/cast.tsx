// The character.
//
// A hand-drawn person, assembled from Open Peeps parts (CC0, Pablo Stanley) via
// the react-peeps package. This replaced a skeleton rig that drew a person from
// joint angles, and the trade it makes is worth stating plainly, because it is
// the opposite of the trade artwork.tsx makes for objects.
//
// The rig could do *anything* — any pose was eleven numbers, and poses
// interpolated, so a character could move continuously from thinking to
// pointing. What it could not do was look like a person somebody drew. Tapered
// capsules with a circle for a head read as a diagram of a human, and next to
// the illustration template's artwork they read as the cheap thing on the
// stage. For an object that is a fine trade: nobody knows what a "correct"
// rocket looks like, so a built-one is just a rocket. Everybody knows what a
// person looks like, and a viewer forgives a stylised drawing of one instantly
// while never quite forgiving a stick figure.
//
// So poses are now artwork, not angles — which means the vocabulary is no
// longer ours to invent. It is exactly what the Open Peeps bust set can do
// *and* can be coloured as, and the table below is that mapping. `wave`,
// `celebrate`, `defeated` and `walk` went because no drawing of them exists;
// `explain`, `coffee` and `phone` went because their drawings cannot be
// coloured without dressing the character in their own skin (see
// CastCharacter). The emotional register the lost poses carried moved to the
// face, which is where it belonged: `defeated` is now `sad`, and it reads
// better as an expression than it ever did as a slump.
//
// What survives from the rig is the thing that actually made it look alive: the
// head still breathes, tilts and blinks over the artwork, and a pose change
// between beats is still a transition rather than a cut.
//
// Nothing here reads a clock or a random source; frames stay independent.

import {random} from 'remotion';
import {BustPose} from 'react-peeps/lib/peeps/pose/bust/z_options';
import {Face} from 'react-peeps/lib/peeps/face/z_options';
import {Hair} from 'react-peeps/lib/peeps/hair/z_options';
import type {BustPoseType, FaceType, HairType} from 'react-peeps';

/**
 * Four colours, because a peep is drawn in three layers and each layer carries
 * its own pair.
 *
 * The first version gave all three layers the same pair, which painted skin,
 * hands and shirt one flat colour. It rendered, it was legible, and it looked
 * like a sticker rather than a person — the single biggest thing wrong with the
 * character, and it was wiring rather than a limit of the artwork.
 */
export type CastPalette = {
  /** The face's features. Nothing else is drawn with it. */
  ink: string;
  /** Skull, neck and hands — every layer's `backgroundColor`. */
  skin: string;
  /** Hair. */
  hair: string;
  /** The garment, and the body's outlines. */
  garment: string;
};

/**
 * One presenter.
 *
 * `backgroundColor` on the body paints the hands *and* any garment the
 * illustrator filled with it, and there is no way to separate the two. That is
 * the whole reason POSES below is restricted to busts that fill their garment
 * with the *stroke*: on those, background is free to be a skin tone and
 * everything lands where it should. On the others the same wiring dresses the
 * character in their own skin, which is exactly as bad as it sounds.
 *
 * Every bust was classified by rendering it in two loud colours and counting
 * pixels, not by eye.
 */
export type CastCharacter = {
  skin: string;
  /**
   * Hair colour. Kept above about 20% lightness on purpose: the hair's outer
   * edge is the character's silhouette against the stage, and near-black hair
   * on the dark stage loses it — most visibly on the big styles, which is
   * exactly where a silhouette matters.
   */
  hair: string;
  style: HairType;
};

/**
 * The cast.
 *
 * Chosen combinations rather than a random product of skin and hair lists: the
 * point is that two clips are visibly presented by two different people, and
 * random pairing produces near-duplicates as happily as it produces contrast.
 */
export const CASTS: CastCharacter[] = [
  {skin: '#f0c9a6', hair: '#4a3226', style: 'Short'},
  {skin: '#c98d5e', hair: '#2f2a3d', style: 'Bun'},
  {skin: '#8c5a3c', hair: '#3b3028', style: 'Afro'},
  {skin: '#f7d9c0', hair: '#8a4a22', style: 'MediumBangs'},
  {skin: '#d9a06b', hair: '#5a4636', style: 'Twists'},
  {skin: '#eab894', hair: '#4d4a58', style: 'ShortWavy'},
  {skin: '#a86f4a', hair: '#33303a', style: 'LongBangs'},
  {skin: '#f2cfae', hair: '#7a5a3a', style: 'ShortCurly'},
];

/**
 * The presenter for a clip.
 *
 * Deterministic, because Remotion renders frames out of order and two frames
 * that disagree about who is on screen is a character who flickers between two
 * people. Seeded off the clip rather than fixed so two snippets in the same
 * course are not presented by the same person.
 */
export const castFor = (seed: string): CastCharacter =>
  CASTS[Math.floor(random(`${seed}-cast`) * CASTS.length) % CASTS.length];

/**
 * A character's colours on a theme.
 *
 * `ink` is safe here in a way it was not before: nothing is drawn with it
 * except the face's features, and those sit on skin, which is light in both
 * modes. The garment carries the brand and every outline on the body, so it
 * has to read on the stage — which is what the primary is for.
 */
export const castPaletteFor = (
  theme: {ink: string; primary: string},
  character: CastCharacter,
): CastPalette => ({
  ink: theme.ink,
  skin: character.skin,
  hair: character.hair,
  garment: theme.primary,
});

/**
 * Where the parts sit inside a peep.
 *
 * These are not free parameters. The first three are react-peeps' own nesting
 * transforms, reproduced here because we compose the head ourselves instead of
 * using its `Head` — that is what lets the head tilt and breathe over a body
 * that does not. The rest were measured off the rendered artwork by rasterising
 * each part and scanning for its ink, so they describe where the drawing
 * actually is rather than where its viewBox claims.
 */
export const PEEP = {
  /** The head group's offset within the peep. */
  headX: 225,
  /** The face's offset within the head group. */
  faceX: 159,
  faceY: 186,
  /** The head's ink bounds, in peep units. */
  head: {x: 270, y: 64, w: 388, h: 458},
  /**
   * The vertical axis every head is drawn on, and the point where it meets the
   * neck. A tilt pivots here: rotating a head about its own centre slides the
   * chin out through the collar.
   */
  axisX: 464,
  neckY: 512,
  /**
   * The vertical crop shared by every pose, so the head is the same size in
   * every shot.
   *
   * The bottom is short of where the artwork ends (~1190) on purpose: an Open
   * Peeps bust is drawn to run off the bottom edge, and cropping inside it is
   * what makes the character read as standing in the frame rather than as a
   * cut-out floating in it.
   */
  top: 34,
  bottom: 1150,
} as const;

export const PEEP_H = PEEP.bottom - PEEP.top;

/**
 * A pose: which drawing it is, how wide that drawing is, and how much the head
 * is allowed to move on it.
 */
export type PeepPose = {
  /** The Open Peeps bust this pose is drawn from. */
  body: BustPoseType;
  /**
   * How far the pose's ink reaches either side of the head axis, in peep
   * units. The frame is built from this rather than from a fixed box, so a
   * shrug gets the width its arms need without every other pose being framed
   * for arms it does not have.
   */
  reach: number;
  /**
   * How freely the head may breathe on this pose, 0–1.
   *
   * A pose that rests a hand against the face has to pin the head, because the
   * hand is painted into the body and cannot follow it. A head drifting a few
   * pixels off its own knuckles is a small error that the eye reads instantly
   * and cannot un-see.
   */
  headPlay: number;
};

/**
 * The pose vocabulary — and the offset table it is built on.
 *
 * Go mirrors the keys (castPoseVocab in snippet_cast.go) and a drift test keeps
 * the two identical. A pose Go allows and this table does not have would fall
 * back to `idle`, so the character would stand there through the beat that was
 * supposed to be its punchline.
 */
export const POSES: Record<string, PeepPose> = {
  // Arms down, nothing happening — the rest position everything else departs
  // from.
  idle: {body: 'ShirtFilled', reach: 438, headPlay: 1},
  // Index finger raised — "here it is", and equally "that's the idea". It
  // carries the beat where something lands, which is why it is the pose the
  // vocabulary can least afford to lose.
  point: {body: 'PointingUp', reach: 440, headPlay: 0.9},
  // Both palms up and out. "I'm not sure", and the widest drawing in the set.
  shrug: {body: 'Whatever', reach: 614, headPlay: 0.8},
  // Arms folded: the confident stance.
  confident: {body: 'ArmsCrossed', reach: 422, headPlay: 0.9},
  // Reading off a sheet — for a beat that quotes something.
  reading: {body: 'Paper', reach: 430, headPlay: 0.45},
  //
  // Four poses are absent for reasons worth keeping written down.
  //
  // `think` had exactly one drawing that read as hand-to-chin, and the reason
  // Open Peeps calls it `Killer` is that the hand is holding a **knife** — the
  // blade is plain at full size and reads as a smudge at thumbnail size, which
  // is how it got picked. Nothing else in the usable set is a thinking
  // gesture, so the pose is gone rather than re-backed; the raised finger
  // above covers the beat where an idea lands, and `thinking` is still
  // available as an *expression*, which is where that register belongs.
  //
  // `explain` (`Explaining`), `coffee` and `phone` are drawn with their
  // garment filled by `backgroundColor`, which also paints the hands. Dressing
  // the character in their own skin tone to keep the hands right, or hiding
  // their hands in their shirt to keep the garment right, are the only two
  // options and both look like a bug.
  //
  // `typing` (`Geek`) has a poop emoji drawn on the laptop lid as a sticker.
  // That is the reason every drawing in this table was looked at rather than
  // picked off the name in its filename.
  //
  // The neckline and the pattern still change from pose to pose — pose and
  // outfit are fused in this artwork and that is not ours to fix. What the
  // restriction buys is that the *colour* no longer changes, which is what the
  // eye actually tracks as "the same person".
};

export const POSE_NAMES = Object.keys(POSES);

export type Expression = keyof typeof FACES;

/**
 * The expression vocabulary.
 *
 * Go mirrors the keys (castExpressionVocab in snippet_cast.go); the same drift
 * test covers it. `sad` and `serious` are here because the poses that used to
 * carry that register (`defeated`) have no drawing — and a face is the better
 * place for it anyway.
 */
export const FACES = {
  neutral: 'Calm',
  happy: 'SmileBig',
  thinking: 'Suspicious',
  surprised: 'Awe',
  concerned: 'Concerned',
  sad: 'Solemn',
  serious: 'Serious',
  talking: 'Explaining',
} satisfies Record<string, FaceType>;

export const EXPRESSION_NAMES = Object.keys(FACES);

/**
 * The blink.
 *
 * Open Peeps has no eyelid layer, so a blink is a whole different face for an
 * eighth of a second. That also swaps the mouth, which sounds like it should
 * be obvious and is not: at 0.12s nobody resolves the mouth, they only resolve
 * that the eyes shut. It is the cheapest thing on this file and the one that
 * stops the character reading as a still.
 */
const BLINK_FACE: FaceType = 'EyesClosed';

export const poseByName = (name: string | undefined): PeepPose =>
  (name && POSES[name]) || POSES.idle;

export const faceByName = (name: string | undefined): FaceType =>
  (name && FACES[name as Expression]) || FACES.neutral;

/**
 * The frame a pose needs, as an svg viewBox.
 *
 * Built around the head axis rather than around the drawing's own bounds, which
 * is the whole reason the reach numbers were measured: it keeps the head in the
 * same place and at the same size in every shot, so a pose change moves the
 * character's arms and not the camera.
 */
export const viewBoxFor = (pose: PeepPose): string =>
  `${PEEP.axisX - pose.reach} ${PEEP.top} ${pose.reach * 2} ${PEEP_H}`;

/** The aspect ratio of that frame, for sizing the svg. */
export const aspectFor = (pose: PeepPose): number => (pose.reach * 2) / PEEP_H;

/** One peep body, drawn in peep coordinates. */
const Body: React.FC<{body: BustPoseType; palette: CastPalette}> = ({body, palette}) => {
  const Piece = BustPose[body];
  // Stroke is the garment, background is the skin. Only true for the busts in
  // POSES — see CastCharacter for why that list is what it is.
  return <Piece strokeColor={palette.garment} backgroundColor={palette.skin} />;
};

/**
 * The head, composed here rather than taken from react-peeps' own `Head`.
 *
 * Its version wraps hair and face in fixed transforms and offers no way in
 * between, so a head drawn with it can only ever sit exactly where the body
 * expects it. Reproducing the two transforms costs four lines and buys the
 * group that everything below animates.
 */
const PeepHead: React.FC<{
  hair: HairType;
  face: FaceType;
  palette: CastPalette;
}> = ({hair, face, palette}) => {
  const HairPiece = Hair[hair];
  const FacePiece = Face[face];
  return (
    <g transform={`translate(${PEEP.headX} 0)`}>
      {/* The hair layer owns the skull as well as the hair: its background is
          the face's fill, its stroke is the hair itself. */}
      <HairPiece strokeColor={palette.hair} backgroundColor={palette.skin} />
      <g transform={`translate(${PEEP.faceX} ${PEEP.faceY})`}>
        <FacePiece strokeColor={palette.ink} backgroundColor={palette.skin} />
      </g>
    </g>
  );
};

export type CharacterProps = {
  pose: PeepPose;
  /** The pose the previous beat left the character in, for the transition. */
  prevPose?: PeepPose;
  /** How far through that transition this frame is, 0–1. */
  poseP?: number;
  face: FaceType;
  /** Who is presenting — supplies the hairstyle its palette was built from. */
  character: CastCharacter;
  palette: CastPalette;
  /** Seconds since the scene started; drives breathing and blinking. */
  t: number;
  /** Which way the character faces. */
  facing?: 1 | -1;
  /** Seeds the breath and blink phases, so a crowd is not in lockstep. */
  seed?: string;
};

export const Character: React.FC<CharacterProps> = ({
  pose,
  prevPose,
  poseP = 1,
  face,
  character,
  palette,
  t,
  facing = 1,
  seed = 'cast',
}) => {
  // Breathing. A slow rise and fall through the head and shoulders, off its own
  // phase so two characters on stage are not in time with each other.
  const off = random(`${seed}-breath`) * Math.PI * 2;
  const breath = Math.sin(t * 1.5 + off);
  const sway = Math.sin(t * 0.9 + off);

  // A pose that holds a hand to the face gets almost none of it — see headPlay.
  const play = pose.headPlay;
  const headDy = breath * 5 * play;
  const headTilt = sway * 1.3 * play;
  // The body gets a fraction of the head's travel. Equal amounts read as the
  // whole drawing sliding up and down, which is a lift, not a breath.
  const bodyDy = breath * 1.6;

  // Blink: shut for an eighth of a second every few seconds, on its own phase.
  const blinkPeriod = 3.4 + random(`${seed}-blinkp`) * 2.2;
  const blinkPhase = (t + random(`${seed}-blink`) * blinkPeriod) % blinkPeriod;
  const shown = blinkPhase < 0.12 ? BLINK_FACE : face;

  // The pose change. Two drawings cannot interpolate, so the outgoing one
  // dissolves while the incoming one settles in under it — and the settle is
  // what sells it. A straight dissolve between two hand-drawn figures reads as
  // a ghost; a dissolve plus a small rise and overshoot reads as somebody
  // moving.
  const settle = 1 - Math.pow(1 - poseP, 3);
  const arriving = prevPose && prevPose !== pose && poseP < 1;

  return (
    <g
      transform={
        facing === -1 ? `translate(${PEEP.axisX * 2} 0) scale(-1 1)` : undefined
      }
    >
      <g transform={`translate(0 ${bodyDy + (1 - settle) * 14})`}>
        {arriving && (
          <g opacity={1 - poseP}>
            <Body body={prevPose.body} palette={palette} />
          </g>
        )}
        <g opacity={arriving ? poseP : 1}>
          <Body body={pose.body} palette={palette} />
        </g>
      </g>
      <g
        transform={`translate(0 ${headDy}) rotate(${headTilt} ${PEEP.axisX} ${PEEP.neckY})`}
      >
        <PeepHead hair={character.style} face={shown} palette={palette} />
      </g>
    </g>
  );
};
