import {useMemo} from 'react';
import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {Stage, STAGE_H, STAGE_W} from './Stage';
import {FIGURE_BOX, figureFor, type FigurePalette} from './artwork';
import {
  Character,
  aspectFor,
  castFor,
  castPaletteFor,
  faceByName,
  poseByName,
  viewBoxFor,
  type CastPalette,
} from './cast';

// CastScene is a person explaining something.
//
// It is the illustration template's sibling: same one-beat-one-shot shape, same
// kinetic headline, but the thing beside the words is a character who reacts
// rather than an object that assembles. That difference is the whole point of
// the template — a pose carries an attitude to the sentence (this is the
// problem / I'm not sure / here it is), and an attitude is something no diagram
// or object can express.
//
// The character holds a pose, but never a still one: cast.tsx keeps breathing
// and blinking running over the artwork, and a pose *change* between beats
// settles in rather than cutting, so the person moves from thinking to pointing
// instead of teleporting between two drawings of themselves.

const ENTER = {
  wordStagger: 3,
  /** The character arrives before the words, so the words look like theirs. */
  castFrames: 16,
  /** How long a pose change takes. */
  poseFrames: 14,
  captionDelay: 14,
  captionFrames: 12,
  /** The prop pops in after the character has landed. */
  propDelay: 10,
  propFrames: 12,
} as const;

const GUTTER = 72;
const CAST_COL = 0.34;

/**
 * The character's drawn height, fixed across every pose.
 *
 * Fixed is the point: the frame each pose is drawn in is as wide as that pose's
 * arms need (viewBoxFor, cast.tsx), so holding the *height* constant is what
 * keeps the head one size from beat to beat. Sizing to width instead would
 * shrink the character every time they shrugged.
 *
 * The value is set so the widest pose in the vocabulary overhangs its column by
 * less than the gutter — a shrug's hands reach into the white space beside the
 * headline, which is where hands should reach, and never into the words.
 */
const CAST_H = 640;

export const CastScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const t = frame / FPS;

  const headline = String(props.headline ?? '');
  const caption = String(props.caption ?? '');
  const poseName = String(props.pose ?? 'idle');
  const prevPoseName = String(props.prevPose ?? poseName);
  const expression = String(props.expression ?? 'neutral') || 'neutral';
  const propName = props.prop ? String(props.prop) : '';
  const flip = Boolean(props.flip);
  // The presenter is the same person for the whole clip, and a different one
  // from clip to clip — so the seed is the headline's neighbour, the plan, not
  // the beat. Go passes it; older scene graphs fall back to one fixed cast.
  const castSeed = String(props.castSeed ?? 'cast-a');

  const words = useMemo(() => headline.split(/\s+/).filter(Boolean), [headline]);

  const presenter = castFor(castSeed);
  const castPalette: CastPalette = castPaletteFor(theme, presenter);
  const figurePalette: FigurePalette = {
    accent: theme.accent,
    primary: theme.primary,
    ink: theme.ink,
    soft: theme.mass,
    line: theme.line,
  };

  // The pose settles in from the previous beat's, so a cut between shots is
  // still a *move* for the character rather than a jump.
  const poseP = interpolate(frame, [0, ENTER.poseFrames], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const pose = poseByName(poseName);
  const prevPose = poseByName(prevPoseName);
  // While the outgoing pose is still dissolving it has to fit too, or a shrug
  // fading into an idle loses its hands to the frame edge on the way out. The
  // frame is symmetric about the head axis, so widening it moves nothing the
  // viewer can see — the head stays exactly where it was.
  const framed =
    poseP < 1 && prevPose !== pose
      ? {...pose, reach: Math.max(pose.reach, prevPose.reach)}
      : pose;

  const castEnter = spring({
    frame,
    fps,
    config: {damping: 200, mass: 0.8},
    durationInFrames: ENTER.castFrames,
  });
  const propP = interpolate(
    frame,
    [ENTER.propDelay, ENTER.propDelay + ENTER.propFrames],
    [0, 1],
    {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
  );
  const lastWordFrame = Math.max(0, (words.length - 1) * ENTER.wordStagger);
  const captionP = interpolate(
    frame,
    [lastWordFrame + ENTER.captionDelay, lastWordFrame + ENTER.captionDelay + ENTER.captionFrames],
    [0, 1],
    {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
  );

  const castH = Math.min(STAGE_H * 0.74, CAST_H);
  const castW = castH * aspectFor(framed);
  // Sized off the height, not the width: the width is the pose's business and
  // changes with it, and a prop that grew every time the character shrugged
  // would be the most distracting thing in the frame.
  const propSize = Math.min(castH * 0.3, 190);
  const Prop = figureFor(propName || undefined);

  const headlineSize =
    headline.length <= 20 ? 92 : headline.length <= 34 ? 78 : headline.length <= 48 ? 66 : 56;

  const castCol = (
    <div
      style={{
        flex: `0 0 ${CAST_COL * 100}%`,
        // The widest poses draw beyond the column and into the gutter on
        // purpose. Without this the flex item's automatic minimum size would be
        // its content's, so a shrug would widen the column and shove the
        // headline sideways for one beat.
        minWidth: 0,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'flex-end',
      }}
    >
      {/* The prop floats above the character — what they are talking about,
          rather than a second thing competing with them for the frame. */}
      {propName ? (
        <div
          style={{
            // Close above the head rather than in its own band: a prop parked a
            // full box-height away reads as an unrelated object, not as the
            // thing they are talking about.
            //
            // It used to overlap by a tenth, which worked against the old rig
            // because that drawing box carried headroom for raised arms. An
            // Open Peeps bust is cropped to the hair, so the same overlap puts
            // the prop on the character's scalp.
            height: propSize * 0.8,
            marginBottom: 10,
            opacity: propP,
            transform: `translateY(${(1 - propP) * 16}px) scale(${0.86 + propP * 0.14})`,
          }}
        >
          <svg width={propSize} height={propSize} viewBox={`0 0 ${FIGURE_BOX} ${FIGURE_BOX}`}>
            <Prop build={propP} t={t} palette={figurePalette} />
          </svg>
        </div>
      ) : (
        <div style={{height: propSize * 0.72}} />
      )}
      <div
        style={{
          opacity: castEnter,
          transform: `translateY(${(1 - castEnter) * 26}px)`,
        }}
      >
        <svg width={castW} height={castH} viewBox={viewBoxFor(framed)}>
          <Character
            pose={pose}
            prevPose={prevPose}
            poseP={poseP}
            face={faceByName(expression)}
            character={presenter}
            palette={castPalette}
            t={t}
            // Face the words. A character addressing the far edge of frame
            // reads as ignoring the thing they are explaining.
            facing={flip ? -1 : 1}
            seed={castSeed}
          />
        </svg>
      </div>
    </div>
  );

  const typeCol = (
    <div style={{flex: '1 1 auto', minWidth: 0}}>
      <div
        style={{
          fontFamily: theme.fontDisplay,
          fontSize: headlineSize,
          fontWeight: 700,
          lineHeight: 1.12,
          letterSpacing: '-0.02em',
          color: theme.text,
          display: 'flex',
          flexWrap: 'wrap',
          columnGap: '0.28em',
          rowGap: '0.12em',
        }}
      >
        {words.map((word, i) => {
          const enter = spring({
            frame: frame - i * ENTER.wordStagger,
            fps,
            config: {damping: 200, mass: 0.7},
            durationInFrames: 15,
          });
          return (
            <span
              key={i}
              style={{
                display: 'inline-block',
                opacity: enter,
                transform: `translateY(${(1 - enter) * 26}px)`,
              }}
            >
              {word}
            </span>
          );
        })}
      </div>
      {caption && (
        <div
          style={{
            marginTop: 28,
            fontFamily: theme.fontBody,
            fontSize: 33,
            lineHeight: 1.45,
            fontWeight: 500,
            color: theme.textMuted,
            opacity: captionP,
            transform: `translateY(${(1 - captionP) * 14}px)`,
            maxWidth: '92%',
          }}
        >
          {caption}
        </div>
      )}
    </div>
  );

  return (
    <Stage>
      <div
        style={{
          display: 'flex',
          width: STAGE_W,
          alignItems: 'center',
          gap: GUTTER,
          flexDirection: flip ? 'row-reverse' : 'row',
        }}
      >
        {castCol}
        {typeCol}
      </div>
    </Stage>
  );
};
