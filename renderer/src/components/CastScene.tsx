import {useMemo} from 'react';
import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {Stage, STAGE_H, STAGE_W} from './Stage';
import {FIGURE_BOX, figureFor, type FigurePalette} from './artwork';
import {CAST_VIEW, Character, livePose, poseByName, type CastPalette, type Expression} from './cast';

// CastScene is a person explaining something.
//
// It is the illustration template's sibling: same one-beat-one-shot shape, same
// kinetic headline, but the thing beside the words is a character who reacts
// rather than an object that assembles. That difference is the whole point of
// the template — a pose carries an attitude to the sentence (this is the
// problem / I'm not sure / here it is), and an attitude is something no diagram
// or object can express.
//
// The character holds a pose, but never a still one: livePose (cast.tsx) keeps
// breathing and blinking running underneath, and a pose *change* between beats
// is interpolated rather than cut, so the person moves from thinking to
// pointing instead of teleporting between two drawings of themselves.

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
  const expression = (String(props.expression ?? 'neutral') || 'neutral') as Expression;
  const propName = props.prop ? String(props.prop) : '';
  const flip = Boolean(props.flip);

  const words = useMemo(() => headline.split(/\s+/).filter(Boolean), [headline]);

  const castPalette: CastPalette = {
    // The character is branded rather than naturalistic: the shirt is the
    // course's primary. A cast that ignores the palette reads as clip art
    // dropped into someone else's video.
    skin: '#f0c8a8',
    hair: theme.ink,
    top: theme.primary,
    bottom: theme.mode === 'light' ? theme.textMuted : '#2f4560',
    ink: theme.ink,
  };
  const figurePalette: FigurePalette = {
    accent: theme.accent,
    primary: theme.primary,
    ink: theme.ink,
    soft: theme.mass,
    line: theme.line,
  };

  // The pose eases from the previous beat's to this one's, so a cut between
  // shots is still a *move* for the character rather than a jump.
  const poseP = interpolate(frame, [0, ENTER.poseFrames], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const target = poseByName(poseName);
  const from = poseByName(prevPoseName);
  const blended = {
    lShoulder: from.lShoulder + (target.lShoulder - from.lShoulder) * poseP,
    lElbow: from.lElbow + (target.lElbow - from.lElbow) * poseP,
    rShoulder: from.rShoulder + (target.rShoulder - from.rShoulder) * poseP,
    rElbow: from.rElbow + (target.rElbow - from.rElbow) * poseP,
    lHip: from.lHip + (target.lHip - from.lHip) * poseP,
    lKnee: from.lKnee + (target.lKnee - from.lKnee) * poseP,
    rHip: from.rHip + (target.rHip - from.rHip) * poseP,
    rKnee: from.rKnee + (target.rKnee - from.rKnee) * poseP,
    headTilt: from.headTilt + (target.headTilt - from.headTilt) * poseP,
    torsoLean: from.torsoLean + (target.torsoLean - from.torsoLean) * poseP,
    lift: from.lift + (target.lift - from.lift) * poseP,
  };
  const pose = livePose(blended, t, poseName === 'walk', 'cast-a');

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

  const castH = Math.min(STAGE_H * 0.86, 690);
  const castW = castH * (CAST_VIEW.w / CAST_VIEW.h);
  const propSize = Math.min(castW * 0.62, 190);
  const Prop = figureFor(propName || undefined);

  const headlineSize =
    headline.length <= 20 ? 92 : headline.length <= 34 ? 78 : headline.length <= 48 ? 66 : 56;

  const castCol = (
    <div
      style={{
        flex: `0 0 ${CAST_COL * 100}%`,
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
            // Overlapping the character's headroom rather than sitting in its
            // own band: a prop parked a full box-height above the head reads as
            // an unrelated object, not as the thing they are talking about.
            height: propSize * 0.72,
            marginBottom: -propSize * 0.1,
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
        <svg
          width={castW}
          height={castH}
          viewBox={`${CAST_VIEW.x} ${CAST_VIEW.y} ${CAST_VIEW.w} ${CAST_VIEW.h}`}
        >
          <Character
            pose={pose}
            expression={expression}
            palette={castPalette}
            t={t}
            // Face the words. A character addressing the far edge of frame
            // reads as ignoring the thing they are explaining.
            facing={flip ? -1 : 1}
            seed="cast-a"
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
