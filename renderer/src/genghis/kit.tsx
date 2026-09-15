// Visual kit for "Genghis Khan and the Mongol Empire".
//
// The film cuts three registers against each other — 1920s archival footage of
// the steppe, 14th-century Persian and Chinese painting, and drawn cartography.
// They only read as one film if they share a ground, a grain and a type scale,
// so every scene composes out of this file. Nothing below knows which register
// it is dressing.

import React from 'react';
import {
  AbsoluteFill,
  interpolate,
  useCurrentFrame,
  Easing,
  random,
} from 'remotion';
import {loadFont as loadCormorant} from '@remotion/google-fonts/CormorantGaramond';
import {loadFont as loadSpectral} from '@remotion/google-fonts/Spectral';

const cormorant = loadCormorant('normal', {
  weights: ['300', '400', '600'],
  subsets: ['latin'],
});
const spectral = loadSpectral('normal', {
  weights: ['300', '400', '500'],
  subsets: ['latin'],
});

export const DISPLAY = cormorant.fontFamily;
export const BODY = spectral.fontFamily;

/**
 * Numerals get their own font, and it is NOT the display face.
 *
 * Cormorant Garamond ships old-style figures: its `1` is a short stroke with
 * serifs that reads as a capital I, and its `0` sits below the cap line. On the
 * QA stills "1215" read as "I2I5" and "1,000" as "I,OOO" — unusable for a film
 * whose spine is a year counter. Spectral has lining figures at a matching
 * weight, so every number in the film is set in it.
 */
export const NUM = spectral.fontFamily;

/**
 * One palette, named by role rather than by colour, so the "cost" section can
 * repoint the accent without every scene knowing about it.
 *
 *   ink0   the ground behind everything      ink1  raised plate
 *   line   hairlines and rules               paper primary text
 *   muted  secondary text                    dim   tertiary / captions
 *   gold   the accent (dates, borders, the empire fill)
 *   blood  reserved for the section on what the conquest cost
 */
export const C = {
  ink0: '#0B0A08',
  ink1: '#141210',
  ink2: '#1C1916',
  line: '#3A322A',
  lineSoft: '#241F1A',
  paper: '#EDE3D4',
  muted: '#A99B87',
  dim: '#6E6357',
  gold: '#C8A055',
  goldBright: '#E8C67E',
  goldDeep: '#8A6A2E',
  blood: '#9C3722',
  bloodSoft: '#C4573C',
  steppe: '#6F7A46',
};

export const FPS = 30;

/** The film's one motion curve. Everything eases the same way or the cut jars. */
export const EASE = Easing.bezier(0.22, 0.61, 0.24, 1);

/** 0 → 1 over `dur` frames starting at `from`, on the house curve. */
export const ramp = (
  frame: number,
  from: number,
  dur: number,
  ease: (n: number) => number = EASE,
) =>
  interpolate(frame, [from, from + dur], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: ease,
  });

/** Fade in at the head and out at the tail of a scene of `total` frames. */
export const holdFade = (frame: number, total: number, head = 14, tail = 14) =>
  Math.min(
    ramp(frame, 0, head),
    1 - ramp(frame, total - tail, tail, Easing.linear),
  );

// ---------------------------------------------------------------- film ground

/**
 * Grain. Drawn as a fixed lattice of soft dots whose brightness is reseeded
 * every frame, which reads as emulsion rather than as the crawling static you
 * get from an SVG turbulence filter re-rendered per frame.
 */
export const Grain: React.FC<{amount?: number; cell?: number}> = ({
  amount = 0.06,
  cell = 3,
}) => {
  const frame = useCurrentFrame();
  // One tile is generated and repeated: a full-frame lattice at cell=3 is
  // 230k nodes and will not render in reasonable time.
  const tile = 96;
  const seed = Math.floor(frame / 2);
  const dots: string[] = [];
  for (let i = 0; i < (tile / cell) * (tile / cell); i++) {
    const r = random(`${seed}-${i}`);
    if (r > 0.55) {
      const x = (i % (tile / cell)) * cell;
      const y = Math.floor(i / (tile / cell)) * cell;
      dots.push(
        `<circle cx="${x}" cy="${y}" r="${cell / 2}" fill="white" opacity="${(
          (r - 0.55) * 2
        ).toFixed(2)}"/>`,
      );
    }
  }
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${tile}" height="${tile}">${dots.join(
    '',
  )}</svg>`;
  return (
    <AbsoluteFill
      style={{
        backgroundImage: `url("data:image/svg+xml;utf8,${encodeURIComponent(svg)}")`,
        backgroundRepeat: 'repeat',
        opacity: amount,
        mixBlendMode: 'overlay',
        pointerEvents: 'none',
      }}
    />
  );
};

/** Corner-to-centre falloff. Sits above picture, below type. */
export const Vignette: React.FC<{strength?: number}> = ({strength = 0.72}) => (
  <AbsoluteFill
    style={{
      background: `radial-gradient(115% 105% at 50% 46%, rgba(0,0,0,0) 38%, rgba(0,0,0,${
        strength * 0.55
      }) 76%, rgba(0,0,0,${strength}) 100%)`,
      pointerEvents: 'none',
    }}
  />
);

/**
 * Gate weave — the sub-pixel drift a mechanical projector puts on the frame.
 * Applied to archival registers only; the drawn maps stay locked, which is
 * what separates "footage" from "diagram" without a caption saying so.
 */
export const useWeave = (amp = 2.2) => {
  const frame = useCurrentFrame();
  const s = Math.floor(frame / 2);
  return {
    x: (random(`wx${s}`) - 0.5) * amp,
    y: (random(`wy${s}`) - 0.5) * amp,
    r: (random(`wr${s}`) - 0.5) * amp * 0.06,
  };
};

// ------------------------------------------------------------------- type

/** Spaced small-caps eyebrow. Used for dates and section labels. */
export const Eyebrow: React.FC<{
  children: React.ReactNode;
  color?: string;
  size?: number;
  style?: React.CSSProperties;
}> = ({children, color = C.gold, size = 20, style}) => (
  <div
    style={{
      fontFamily: BODY,
      fontSize: size,
      fontWeight: 500,
      letterSpacing: size * 0.34,
      textTransform: 'uppercase',
      color,
      ...style,
    }}
  >
    {children}
  </div>
);

export const Display: React.FC<{
  children: React.ReactNode;
  size?: number;
  weight?: number;
  color?: string;
  style?: React.CSSProperties;
}> = ({children, size = 96, weight = 300, color = C.paper, style}) => (
  <div
    style={{
      fontFamily: DISPLAY,
      fontSize: size,
      fontWeight: weight,
      lineHeight: 1.06,
      letterSpacing: -size * 0.012,
      color,
      ...style,
    }}
  >
    {children}
  </div>
);

export const Body: React.FC<{
  children: React.ReactNode;
  size?: number;
  color?: string;
  style?: React.CSSProperties;
}> = ({children, size = 30, color = C.muted, style}) => (
  <div
    style={{
      fontFamily: BODY,
      fontSize: size,
      fontWeight: 300,
      lineHeight: 1.5,
      color,
      ...style,
    }}
  >
    {children}
  </div>
);

/** A hairline that draws itself left-to-right. */
export const Rule: React.FC<{
  progress?: number;
  width?: number | string;
  color?: string;
  thickness?: number;
}> = ({progress = 1, width = 220, color = C.gold, thickness = 1}) => (
  <div style={{width, height: thickness, background: C.lineSoft}}>
    <div
      style={{
        width: `${Math.max(0, Math.min(1, progress)) * 100}%`,
        height: '100%',
        background: color,
      }}
    />
  </div>
);

/**
 * Drawn line-art glyph set. Emoji is the fastest way to make a film look like a
 * template, and this cut needs a horse and a bow more than it needs a rocket.
 */
export const Glyph: React.FC<{
  name: 'horse' | 'bow' | 'ger' | 'arrow' | 'rider' | 'seal';
  size?: number;
  color?: string;
  weight?: number;
}> = ({name, size = 48, color = C.gold, weight = 1.5}) => {
  const p: Record<string, string> = {
    // A standing horse, reduced to one stroke of back, neck and four legs.
    horse:
      'M4 15 C7 13 9 13 12 13 L16 13 C17 11 18 9 19 8 L20 5 L21.5 7 L20.5 9 L20 12 C20 14 19 15 18 15.5 M6 15 L5.5 20 M9.5 14.5 L9 20 M14.5 13.5 L14 20 M17.5 15 L17.5 20 M4 15 C3 14.5 2.5 13.5 3 12.5',
    // Composite recurve, drawn upright: the limbs bend away from the archer and
    // the tips curl back toward them, which is the whole point of the object.
    // The string is the straight chord, and the arrow is nocked on it.
    bow: 'M7 2.5 C4.2 3.2 3.4 4.6 4.2 6.2 M4.2 6.2 C1.6 9.4 1.6 14.6 4.2 17.8 M4.2 17.8 C3.4 19.4 4.2 20.8 7 21.5 M6.2 3.4 L6.2 20.6 M6.2 12 L19.5 12 M15.8 9 L19.5 12 L15.8 15 M6.2 12 L8.6 10.2 M6.2 12 L8.6 13.8',
    ger: 'M3 16 L5 9 C8 7.5 16 7.5 19 9 L21 16 M3 16 L21 16 M12 7.8 L12 5.5 M10 16 L10 12 L14 12 L14 16 M6.5 9.6 C9.5 8.6 14.5 8.6 17.5 9.6',
    arrow: 'M2 12 L20 12 M15 7 L20 12 L15 17 M2 12 L5 10 M2 12 L5 14',
    rider:
      // horse: back, neck, head, four legs — then the rider seated on it.
      'M4 15 C7 13.4 9.5 13.2 12.5 13.2 L16.5 13.2 C17.4 11.6 18.2 10 19 8.8 L20 6.2 '
      + 'L21.4 8 L20.6 10 L20 12.6 M6 15.2 L5.4 20 M10 14 L9.6 20 M15 13.4 L14.6 20 '
      + 'M17.8 13.6 L18.2 20 M4 15 C3.1 14.4 2.7 13.5 3.1 12.6 '
      + 'M11 5.4 A1.5 1.5 0 1 0 11 8.4 A1.5 1.5 0 1 0 11 5.4 '
      + 'M11 8.6 L11.4 12 M11.1 9.8 L14.6 10.6 M11.1 9.8 L8.2 11.2 '
      + 'M11.4 12 L13.2 13.2',
    seal: 'M12 2.5 L14.6 8 L20.5 8.8 L16.2 13 L17.3 19 L12 16.1 L6.7 19 L7.8 13 L3.5 8.8 L9.4 8 Z',
  };
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <path
        d={p[name]}
        stroke={color}
        strokeWidth={weight}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};

/** The ground every scene sits on. */
export const Ground: React.FC<{children?: React.ReactNode; tone?: string}> = ({
  children,
  tone = C.ink0,
}) => (
  <AbsoluteFill style={{background: tone}}>
    {children}
    <Grain />
  </AbsoluteFill>
);
