import {useMemo} from 'react';
import {AbsoluteFill, interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {FIGURE_BOX, figureFor, type FigurePalette} from './artwork';
import {CAST_FEET, CAST_VIEW, Character, livePose, poseByName, type CastPalette, type Expression} from './cast';
import {CAPTION_SAFE, FRAME_H, SAFE_X} from './Stage';
import {WORLD, camAt, camTransform, parallaxCam} from './camera';

// StoryScene is one shot of a directed piece.
//
// Unlike every other scene here it does not compose into the frame. It lays its
// content out on a world larger than the frame (camera.ts) and points a moving
// camera at part of it, which is what lets a shot keep moving while nothing in
// it changes.
//
// The director chose a staging and a camera move from closed vocabularies. This
// file owns every coordinate that follows from those choices — which is the
// same trade the whiteboard and the flow diagram make, and for the same reason:
// a model handed x/y produces characters standing on their own captions.
//
// The layout scheme is one rule: figures live in the upper two thirds, type
// lives in a lower third, centred. Holding that constant across every staging
// is what makes fourteen different shots read as one film instead of fourteen
// designs.

const BAND = {
  /** Where figures stand: their feet land here. */
  floorY: 0.63,
  /** Vertical centre of a floating object. */
  objectY: 0.36,
  /** Top of the type block. */
  typeY: 0.66,
} as const;

const ENTER = {
  wordStagger: 3,
  figureFrames: 18,
  poseFrames: 14,
  captionDelay: 14,
  captionFrames: 12,
} as const;

/** Backdrop depth. Below ~0.3 the stage detaches; above ~0.6 it reads as flat. */
const PARALLAX = 0.42;

type Placed = {x: number; y: number; size: number};

/**
 * Where things stand, per staging. All in world coordinates.
 *
 * Sizes are fractions of world height so a figure keeps its scale relative to
 * the stage rather than to whatever the frame happens to be.
 */
const stagePlan = (
  staging: string,
): {cast?: Placed; propA?: Placed; propB?: Placed; typeCentred: boolean} => {
  const W = WORLD.w;
  const H = WORLD.h;
  switch (staging) {
    case 'hero':
      return {cast: {x: W * 0.5, y: H * BAND.floorY, size: H * 0.52}, typeCentred: false};
    case 'duo':
      return {
        cast: {x: W * 0.34, y: H * BAND.floorY, size: H * 0.46},
        propA: {x: W * 0.67, y: H * BAND.objectY, size: H * 0.3},
        typeCentred: false,
      };
    case 'object':
      return {propA: {x: W * 0.5, y: H * BAND.objectY, size: H * 0.42}, typeCentred: false};
    case 'pair':
      return {
        propA: {x: W * 0.33, y: H * BAND.objectY, size: H * 0.26},
        propB: {x: W * 0.67, y: H * BAND.objectY, size: H * 0.26},
        typeCentred: false,
      };
    case 'empty':
    default:
      // Nothing on stage, so the type takes the middle instead of the lower
      // third — a bare stage with a lower-third caption reads as a shot whose
      // subject failed to load.
      return {typeCentred: true};
  }
};

export const StoryScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const t = frame / FPS;

  const headline = String(props.headline ?? '');
  const caption = String(props.caption ?? '');
  const staging = String(props.staging ?? 'empty');
  const camera = String(props.camera ?? 'hold');
  const poseName = String(props.pose ?? 'idle');
  const prevPoseName = String(props.prevPose ?? poseName);
  const expression = (String(props.expression ?? 'neutral') || 'neutral') as Expression;
  const propAName = props.prop ? String(props.prop) : '';
  const propBName = props.propB ? String(props.propB) : '';
  const durationMs = Number(props.durationMs ?? 5000);

  const words = useMemo(() => headline.split(/\s+/).filter(Boolean), [headline]);
  const plan = useMemo(() => stagePlan(staging), [staging]);

  // The camera runs on the shot's own length, so a long beat gets a slow move
  // and a short one a brisk one. Tying it to a fixed frame count instead made
  // every move finish early and every shot end locked off.
  const shotFrames = Math.max(1, (durationMs / 1000) * FPS);
  const cam = camAt(camera, frame / shotFrames);
  const bgCam = parallaxCam(cam, PARALLAX);

  const castPalette: CastPalette = {
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

  const enter = spring({frame, fps, config: {damping: 200, mass: 0.8}, durationInFrames: 18});

  // The character carries its pose across the cut, so a change between shots is
  // a move rather than a jump.
  const poseP = interpolate(frame, [0, ENTER.poseFrames], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const a = poseByName(prevPoseName);
  const b = poseByName(poseName);
  const m = (x: number, y: number) => x + (y - x) * poseP;
  const pose = livePose(
    {
      lShoulder: m(a.lShoulder, b.lShoulder), lElbow: m(a.lElbow, b.lElbow),
      rShoulder: m(a.rShoulder, b.rShoulder), rElbow: m(a.rElbow, b.rElbow),
      lHip: m(a.lHip, b.lHip), lKnee: m(a.lKnee, b.lKnee),
      rHip: m(a.rHip, b.rHip), rKnee: m(a.rKnee, b.rKnee),
      headTilt: m(a.headTilt, b.headTilt), torsoLean: m(a.torsoLean, b.torsoLean),
      lift: m(a.lift, b.lift),
    },
    t,
    poseName === 'walk',
    'story-a',
  );

  const lastWordFrame = Math.max(0, (words.length - 1) * ENTER.wordStagger);
  const captionP = interpolate(
    frame,
    [lastWordFrame + ENTER.captionDelay, lastWordFrame + ENTER.captionDelay + ENTER.captionFrames],
    [0, 1],
    {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
  );

  const headlineSize = plan.typeCentred
    ? headline.length <= 26 ? 104 : headline.length <= 44 ? 84 : 68
    : headline.length <= 26 ? 78 : headline.length <= 44 ? 64 : 54;

  const figure = (p: Placed, name: string, key: string, delay: number) => {
    const Fig = figureFor(name || undefined);
    const local = interpolate(frame, [delay, delay + ENTER.figureFrames], [0, 1], {
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp',
    });
    return (
      <div
        key={key}
        style={{
          position: 'absolute',
          left: p.x - p.size / 2,
          top: p.y - p.size / 2,
          width: p.size,
          height: p.size,
          opacity: local,
          transform: `translateY(${(1 - local) * 18}px)`,
        }}
      >
        <svg width={p.size} height={p.size} viewBox={`0 0 ${FIGURE_BOX} ${FIGURE_BOX}`}>
          <Fig build={local} t={t} palette={figurePalette} />
        </svg>
      </div>
    );
  };

  const castW = plan.cast ? plan.cast.size * (CAST_VIEW.w / CAST_VIEW.h) : 0;

  return (
    <AbsoluteFill>
      {/* Backdrop, tracking the camera only partly. Depth on a flat stage is
          entirely differential motion — matched to the camera it is painted on
          the front element, unmoving it is wallpaper behind a cutout. */}
      <AbsoluteFill style={{transform: camTransform(bgCam), transformOrigin: '0 0'}}>
        <div
          style={{
            position: 'absolute',
            left: WORLD.w * 0.12,
            top: WORLD.h * 0.1,
            width: WORLD.w * 0.42,
            height: WORLD.w * 0.42,
            borderRadius: '50%',
            background: `radial-gradient(circle, ${theme.primary}1c 0%, transparent 66%)`,
          }}
        />
        <div
          style={{
            position: 'absolute',
            left: WORLD.w * 0.5,
            top: WORLD.h * 0.3,
            width: WORLD.w * 0.38,
            height: WORLD.w * 0.38,
            borderRadius: '50%',
            background: `radial-gradient(circle, ${theme.accent}14 0%, transparent 64%)`,
          }}
        />
        {/* A horizon line gives the floor somewhere to be, so a standing
            character is standing on something. */}
        <div
          style={{
            position: 'absolute',
            left: 0,
            top: WORLD.h * BAND.floorY,
            width: WORLD.w,
            height: 1,
            // Faint, and fading well before the edges. At full width and any
            // real opacity it stops reading as a floor and starts reading as a
            // horizontal rule drawn across the video.
            background: `linear-gradient(90deg, transparent 8%, ${theme.line}1c 34%, ${theme.line}1c 66%, transparent 92%)`,
          }}
        />
      </AbsoluteFill>

      {/* The stage itself. */}
      <AbsoluteFill style={{transform: camTransform(cam), transformOrigin: '0 0'}}>
        {plan.propA && propAName ? figure(plan.propA, propAName, 'a', 4) : null}
        {plan.propB && propBName ? figure(plan.propB, propBName, 'b', 10) : null}

        {plan.cast && (
          <div
            style={{
              position: 'absolute',
              left: plan.cast.x - castW / 2,
              // Stand the soles on the floor line, not the bottom of the
              // drawing box — the box carries headroom for raised arms, and
              // using it directly leaves the character hovering.
              top: plan.cast.y - plan.cast.size * CAST_FEET,
              width: castW,
              height: plan.cast.size,
              opacity: enter,
              transform: `translateY(${(1 - enter) * 22}px)`,
            }}
          >
            <svg
              width={castW}
              height={plan.cast.size}
              viewBox={`${CAST_VIEW.x} ${CAST_VIEW.y} ${CAST_VIEW.w} ${CAST_VIEW.h}`}
            >
              <Character
                pose={pose}
                expression={expression}
                palette={castPalette}
                t={t}
                // The character always faces stage right: on `duo` that is
                // where the object is, and on `hero` it keeps the presenter
                // consistent from shot to shot rather than flipping on a cut.
                facing={1}
                seed="story-a"
              />
            </svg>
          </div>
        )}

      </AbsoluteFill>

      {/* Type is a lower third, and a lower third does not ride the camera.
          Keeping it in world space had it drift and — worse — scale outward
          under a push, straight through the band the burned-in captions live
          in. It is frame-space now, inside the caption-safe margin, which is
          both correct film grammar and the only way it can be guaranteed not
          to collide. */}
      <AbsoluteFill style={{pointerEvents: 'none'}}>
        <div
          style={{
            position: 'absolute',
            left: SAFE_X,
            right: SAFE_X,
            ...(plan.typeCentred
              ? {top: '38%'}
              : {bottom: CAPTION_SAFE + 28}),
            textAlign: 'center',
          }}
        >
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: headlineSize,
              fontWeight: 700,
              lineHeight: 1.1,
              letterSpacing: '-0.02em',
              color: theme.text,
              display: 'flex',
              flexWrap: 'wrap',
              justifyContent: 'center',
              columnGap: '0.28em',
              rowGap: '0.1em',
            }}
          >
            {words.map((word, i) => {
              const wordEnter = spring({
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
                    opacity: wordEnter,
                    transform: `translateY(${(1 - wordEnter) * 24}px)`,
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
                marginTop: 22,
                fontFamily: theme.fontBody,
                fontSize: plan.typeCentred ? 38 : 32,
                lineHeight: 1.45,
                fontWeight: 500,
                color: theme.textMuted,
                opacity: captionP,
                transform: `translateY(${(1 - captionP) * 12}px)`,
              }}
            >
              {caption}
            </div>
          )}
        </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};
