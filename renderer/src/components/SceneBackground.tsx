import {useMemo} from 'react';
import {AbsoluteFill, useCurrentFrame} from 'remotion';
import {noise2D} from '@remotion/noise';
import {ResolvedTheme} from '../theme/theme';

// SceneBackground is the shared canvas behind every scene: the theme's
// brand-tinted gradient, two softly drifting accent glows (seeded noise —
// deterministic per frame), a faint dot grid, film grain to mask H.264
// banding, and a vignette. Everything derives from useCurrentFrame().

const GRAIN_SEED = 7;

export const SceneBackground: React.FC<{
  theme: ResolvedTheme;
  children?: React.ReactNode;
}> = ({theme, children}) => {
  const frame = useCurrentFrame();

  // Slow orbital drift: ±90px over ~20s, unique per blob.
  const t = frame * 0.0035;
  const b1x = noise2D('bg-blob1-x', t, 0) * 90;
  const b1y = noise2D('bg-blob1-y', 0, t) * 70;
  const b2x = noise2D('bg-blob2-x', t + 50, 0) * 110;
  const b2y = noise2D('bg-blob2-y', 0, t + 50) * 80;

  // The grain rect is static (fixed seed): cheaper than re-seeding per frame
  // and still breaks up gradient banding.
  const grain = useMemo(
    () => (
      <svg width="100%" height="100%" style={{position: 'absolute', inset: 0, opacity: theme.grain}}>
        <filter id="cs-grain">
          <feTurbulence type="fractalNoise" baseFrequency="0.8" numOctaves="2" seed={GRAIN_SEED} stitchTiles="stitch" />
          <feColorMatrix type="saturate" values="0" />
        </filter>
        <rect width="100%" height="100%" filter="url(#cs-grain)" />
      </svg>
    ),
    [theme.grain],
  );

  return (
    <AbsoluteFill style={{background: `linear-gradient(168deg, ${theme.bgTop} 0%, ${theme.bgBottom} 100%)`}}>
      {/* Accent glows */}
      <div
        style={{
          position: 'absolute',
          width: 1300,
          height: 1300,
          left: -420 + b1x,
          top: -560 + b1y,
          borderRadius: '50%',
          background: `radial-gradient(circle, ${theme.primary}33 0%, transparent 62%)`,
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
          background: `radial-gradient(circle, ${theme.accent}1f 0%, transparent 60%)`,
        }}
      />
      {/* Dot grid, very faint */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: `radial-gradient(${theme.text}14 1.4px, transparent 1.4px)`,
          backgroundSize: '56px 56px',
        }}
      />
      {grain}
      {/* Vignette */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          background: 'radial-gradient(ellipse at center, transparent 58%, rgba(0,0,0,0.32) 100%)',
        }}
      />
      {children}
    </AbsoluteFill>
  );
};
