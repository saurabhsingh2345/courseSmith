import {useMemo} from 'react';
import {AbsoluteFill, useCurrentFrame} from 'remotion';
import {noise2D} from '@remotion/noise';
import {ResolvedTheme} from '../theme/theme';

// SceneBackground is the canvas behind every scene: the theme's brand-tinted
// gradient, two softly drifting accent glows (seeded noise — deterministic per
// frame), a patterned field, film grain to mask H.264 banding, and a vignette.
// Everything derives from useCurrentFrame().
//
// **It has surfaces now.** One backdrop behind everything meant a hand-drawn
// board, a systems diagram and a character all stood on the same softly glowing
// dot grid — which is a house style, but it is also the same shot every time,
// and it was fighting half of them. A whiteboard wants to look like a surface
// you draw on; a systems diagram wants a technical grid; a presenter wants a
// pool of light and nothing else competing; a chart wants to be the only ink on
// screen.
//
// The variant is chosen from the scene types in the video (see surfaceFor in
// LessonVideo), not asked for, so nothing downstream has to remember to set it.

const GRAIN_SEED = 7;

export type Surface = 'default' | 'paper' | 'blueprint' | 'spotlight' | 'clean' | 'void';

/**
 * The knobs each surface turns. Kept as data rather than as five components
 * because they differ by degree, not in kind — and a variant that forgot the
 * grain or the vignette would read as a different design system rather than as
 * the same room lit differently.
 */
const SURFACES: Record<
  Surface,
  {
    /** Multipliers on the two drifting accent glows. */
    glow: number;
    /** The repeating field: dots, a technical grid, faint rules, or nothing. */
    field: 'dots' | 'grid' | 'rules' | 'none';
    /** Vignette strength multiplier. */
    vignette: number;
    /** Grain multiplier — paper has tooth, a lit stage does not. */
    grain: number;
  }
> = {
  default: {glow: 1, field: 'dots', vignette: 1, grain: 1},
  // Flatter light and a horizontal rule, so the board reads as a surface with a
  // top and a bottom rather than as open space with boxes floating in it.
  paper: {glow: 0.45, field: 'rules', vignette: 0.75, grain: 1.5},
  // The drawing-office grid. Strong enough to say "this is a diagram", faint
  // enough that a 2px edge still wins against it.
  blueprint: {glow: 0.5, field: 'grid', vignette: 0.9, grain: 0.8},
  // A stage: one pool of light, no pattern, a deeper vignette to close the
  // corners down around whoever is standing in it.
  spotlight: {glow: 1.35, field: 'none', vignette: 1.5, grain: 0.9},
  // Charts are dense and thin. Anything repeating behind them competes with the
  // gridlines they draw themselves.
  clean: {glow: 0.55, field: 'none', vignette: 0.7, grain: 0.7},
  // Nothing at all. The backdrop is not lit, not patterned and barely graded,
  // so every bit of luminance in the frame belongs to the one small diagram in
  // the middle of it. This is what the broadcast skin stands on, and it is the
  // only surface where the *absence* of the glows is the design rather than a
  // reduction of it — a 0.05 glow here would read as a smudge on black.
  void: {glow: 0, field: 'none', vignette: 0.5, grain: 1},
};

const Field: React.FC<{theme: ResolvedTheme; kind: Surface}> = ({theme, kind}) => {
  const {field} = SURFACES[kind];
  if (field === 'none') return null;
  if (field === 'dots') {
    return (
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: `radial-gradient(${theme.text}14 1.4px, transparent 1.4px)`,
          backgroundSize: '56px 56px',
        }}
      />
    );
  }
  if (field === 'rules') {
    // Horizontal only. A full grid on a hand-drawn board reads as graph paper,
    // which is a different and much fussier object than a whiteboard.
    return (
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: `repeating-linear-gradient(0deg, ${theme.text}0d 0 1px, transparent 1px 84px)`,
        }}
      />
    );
  }
  // Blueprint: a fine grid with a heavier line every fifth square, which is what
  // makes it read as measured rather than as wallpaper.
  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        backgroundImage:
          `repeating-linear-gradient(0deg, ${theme.text}0a 0 1px, transparent 1px 40px),` +
          `repeating-linear-gradient(90deg, ${theme.text}0a 0 1px, transparent 1px 40px),` +
          `repeating-linear-gradient(0deg, ${theme.text}12 0 1.5px, transparent 1.5px 200px),` +
          `repeating-linear-gradient(90deg, ${theme.text}12 0 1.5px, transparent 1.5px 200px)`,
      }}
    />
  );
};

export const SceneBackground: React.FC<{
  theme: ResolvedTheme;
  surface?: Surface;
  children?: React.ReactNode;
}> = ({theme, surface = 'default', children}) => {
  const frame = useCurrentFrame();
  const s = SURFACES[surface];

  // Slow orbital drift: ±90px over ~20s, unique per blob.
  const t = frame * 0.0035;
  const b1x = noise2D('bg-blob1-x', t, 0) * 90;
  const b1y = noise2D('bg-blob1-y', 0, t) * 70;
  const b2x = noise2D('bg-blob2-x', t + 50, 0) * 110;
  const b2y = noise2D('bg-blob2-y', 0, t + 50) * 80;

  // The grain rect is static (fixed seed): cheaper than re-seeding per frame
  // and still breaks up gradient banding.
  const grainOpacity = theme.grain * s.grain;
  const grain = useMemo(
    () => (
      <svg width="100%" height="100%" style={{position: 'absolute', inset: 0, opacity: grainOpacity}}>
        <filter id="cs-grain">
          <feTurbulence type="fractalNoise" baseFrequency="0.8" numOctaves="2" seed={GRAIN_SEED} stitchTiles="stitch" />
          <feColorMatrix type="saturate" values="0" />
        </filter>
        <rect width="100%" height="100%" filter="url(#cs-grain)" />
      </svg>
    ),
    [grainOpacity],
  );

  const hex = (alpha: number): string =>
    Math.round(Math.max(0, Math.min(255, alpha * 255)))
      .toString(16)
      .padStart(2, '0');

  return (
    <AbsoluteFill style={{background: `linear-gradient(168deg, ${theme.bgTop} 0%, ${theme.bgBottom} 100%)`}}>
      {/* Accent glows. On a spotlight surface the first one is recentred over
          the stage instead of hanging off the top-left corner — the whole point
          of that surface is that the light is *on* the subject. */}
      <div
        style={{
          position: 'absolute',
          width: surface === 'spotlight' ? 1600 : 1300,
          height: surface === 'spotlight' ? 1600 : 1300,
          left: (surface === 'spotlight' ? 160 : -420) + b1x,
          top: (surface === 'spotlight' ? -420 : -560) + b1y,
          borderRadius: '50%',
          background: `radial-gradient(circle, ${theme.primary}${hex(0.2 * s.glow)} 0%, transparent 62%)`,
        }}
      />
      <div
        style={{
          position: 'absolute',
          width: 1100,
          height: 1100,
          right: -420 + b2x,
          bottom: -520 + b2y,
          borderRadius: '50%',
          background: `radial-gradient(circle, ${theme.accent}${hex(0.12 * s.glow)} 0%, transparent 60%)`,
        }}
      />
      <Field theme={theme} kind={surface} />
      {grain}
      {/* Vignette. A black one is right on the dark stage and wrong on paper —
          it reads as a dirty scan rather than as a lens. Light mode gets a much
          weaker one in the brand ink instead, which keeps the corners settled
          without smudging them. */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          background:
            theme.mode === 'light'
              ? `radial-gradient(ellipse at center, transparent 62%, ${theme.ink}${hex(0.07 * s.vignette)} 100%)`
              : `radial-gradient(ellipse at center, transparent ${58 - (s.vignette - 1) * 10}%, rgba(0,0,0,${(0.32 * s.vignette).toFixed(2)}) 100%)`,
        }}
      />
      {children}
    </AbsoluteFill>
  );
};
