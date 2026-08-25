// The visual kit for the Lumen film.
//
// The lesson this cut follows was built on an orange ground with cream cards, a
// single amber accent, and a deliberately frozen background. This one goes the
// other way on all three counts, because the thing on screen is different: the
// app being built is a photo feed, so the film has to be dark enough that a
// bright white UI lands on it like a lightbox, and colourful enough that it does
// not look like a slide deck wrapped around somebody's screenshots.
//
// So: near-black ground, an aurora that actually moves, glass surfaces, and a
// two-stop accent (rose into violet) used as a gradient rather than a fill.
// Headlines are a grotesque with real personality; the one word that matters in
// a headline is set in an italic serif, which is the cheapest way to make a
// title feel written rather than typed.

import React from 'react';
import {
  AbsoluteFill,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
  Easing,
  random,
} from 'remotion';
import {loadFont as loadBricolage} from '@remotion/google-fonts/BricolageGrotesque';
import {loadFont as loadInstrument} from '@remotion/google-fonts/InstrumentSerif';
import {loadFont as loadGeist} from '@remotion/google-fonts/Geist';
import {loadFont as loadGeistMono} from '@remotion/google-fonts/GeistMono';

const bric = loadBricolage('normal', {weights: ['400', '600', '700', '800'], subsets: ['latin']});
const inst = loadInstrument('italic', {weights: ['400'], subsets: ['latin']});
const geist = loadGeist('normal', {weights: ['300', '400', '500', '600'], subsets: ['latin']});
const geistMono = loadGeistMono('normal', {weights: ['400', '500'], subsets: ['latin']});

export const DISPLAY = bric.fontFamily;
export const SERIF = inst.fontFamily;
export const BODY = geist.fontFamily;
export const MONO = geistMono.fontFamily;

export const C = {
  ink0: '#08070F',
  ink1: 'rgba(255,255,255,0.045)',
  ink2: 'rgba(255,255,255,0.075)',
  line: 'rgba(255,255,255,0.10)',
  line2: 'rgba(255,255,255,0.055)',
  text: '#F6F4FC',
  muted: '#ABA5C6',
  dim: '#6F6992',
  rose: '#FF4D6D',
  violet: '#8B5CF6',
  cyan: '#22D3EE',
  lime: '#A3E635',
  ok: '#34D399',
};

/** The accent, as a gradient. Used on rules, fills, and hot words. */
export const GRAD = `linear-gradient(100deg, ${C.rose}, ${C.violet})`;
export const GRAD_WIDE = `linear-gradient(100deg, ${C.rose} 0%, ${C.violet} 62%, ${C.cyan} 100%)`;

/** Paint text with the accent gradient. */
export const grad: React.CSSProperties = {
  backgroundImage: GRAD,
  WebkitBackgroundClip: 'text',
  backgroundClip: 'text',
  color: 'transparent',
};

export const FPS = 30;
export const s = (sec: number) => Math.round(sec * FPS);
export const EASE = Easing.bezier(0.22, 1, 0.36, 1);

export const useEnter = (delay = 0, damping = 22) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  return spring({frame: frame - delay, fps, config: {damping, stiffness: 120, mass: 0.9}});
};

export const ramp = (frame: number, at: number, len: number) =>
  interpolate(frame, [at, at + len], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: EASE,
  });

/** Linear 0→1, for anything that should not ease (drift, sweeps). */
export const lin = (frame: number, at: number, len: number) =>
  interpolate(frame, [at, at + len], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

// ---------------------------------------------------------------------------
// Ground

const GRAIN =
  "url(\"data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='220' height='220'><filter id='n'><feTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='3'/></filter><rect width='220' height='220' filter='url(%23n)' opacity='0.55'/></svg>\")";

/**
 * Three lights on slow, mutually prime orbits.
 *
 * The reference cut froze its background on purpose, and the note it left was
 * right: a bloom that wanders is maddening over twenty minutes. What made it
 * maddening there, though, was that it wandered at roughly the speed of the
 * content — fast enough to notice, too slow to read as motion. The fix is not
 * to stop it but to slow it well past that: these periods are 41, 57 and 73
 * seconds, so nothing ever repeats inside a chapter and no single shot contains
 * a visible move. It reads as depth, not as animation.
 *
 * No `filter: blur()` anywhere. A 1920-wide blurred layer is one of the most
 * expensive things you can ask a compositor for, and a radial gradient with a
 * soft stop is indistinguishable from a blurred disc.
 */
export const Ground: React.FC<{children?: React.ReactNode; still?: boolean}> = ({
  children,
  still = false,
}) => {
  const frame = useCurrentFrame();
  const t = still ? 0 : frame / FPS;

  const orbit = (period: number, phase: number, ax: number, ay: number) => ({
    x: Math.sin((t / period + phase) * Math.PI * 2) * ax,
    y: Math.cos((t / period + phase * 1.7) * Math.PI * 2) * ay,
  });

  const a = orbit(41, 0.0, 7, 5);
  const b = orbit(57, 0.33, 9, 6);
  const c = orbit(73, 0.66, 6, 8);

  return (
    <AbsoluteFill style={{backgroundColor: C.ink0}}>
      <AbsoluteFill
        style={{
          background: `radial-gradient(58% 62% at ${22 + a.x}% ${18 + a.y}%, rgba(255,77,109,0.20) 0%, rgba(255,77,109,0.07) 38%, transparent 68%)`,
        }}
      />
      <AbsoluteFill
        style={{
          background: `radial-gradient(64% 68% at ${82 + b.x}% ${76 + b.y}%, rgba(139,92,246,0.24) 0%, rgba(139,92,246,0.08) 40%, transparent 70%)`,
        }}
      />
      <AbsoluteFill
        style={{
          background: `radial-gradient(52% 56% at ${60 + c.x}% ${8 + c.y}%, rgba(34,211,238,0.13) 0%, rgba(34,211,238,0.04) 42%, transparent 70%)`,
        }}
      />
      {/* a cool floor so the bottom third does not go muddy */}
      <AbsoluteFill
        style={{
          background:
            'linear-gradient(180deg, transparent 42%, rgba(8,7,15,0.55) 78%, rgba(8,7,15,0.85) 100%)',
        }}
      />
      <AbsoluteFill
        style={{backgroundImage: GRAIN, opacity: 0.055, mixBlendMode: 'overlay'}}
      />
      {children}
      <AbsoluteFill
        style={{boxShadow: 'inset 0 0 340px rgba(0,0,0,0.72)', pointerEvents: 'none'}}
      />
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------
// Type

export const Eyebrow: React.FC<{children: React.ReactNode; delay?: number; color?: string}> = ({
  children,
  delay = 0,
  color = C.cyan,
}) => {
  const e = useEnter(delay);
  return (
    <div
      style={{
        fontFamily: MONO,
        fontSize: 21,
        fontWeight: 500,
        letterSpacing: '0.26em',
        textTransform: 'uppercase',
        color,
        opacity: e,
        transform: `translateY(${(1 - e) * 12}px)`,
      }}
    >
      {children}
    </div>
  );
};

/**
 * A headline that lifts in word by word.
 *
 * `accent` words take the gradient. `serif` words are set in italic Instrument
 * Serif at a slightly larger size — the optical size of a serif at the same
 * point size reads smaller, so it needs the bump to sit on the same line.
 */
export const Headline: React.FC<{
  text: string;
  size?: number;
  delay?: number;
  accent?: string[];
  serif?: string[];
  align?: 'left' | 'center';
  width?: number | string;
  weight?: number;
}> = ({text, size = 92, delay = 0, accent = [], serif = [], align = 'left', width, weight = 700}) => {
  const frame = useCurrentFrame();
  const words = text.split(' ');
  const key = (w: string) => w.toLowerCase().replace(/[.,?!:;—]/g, '');
  const hot = new Set(accent.map(key));
  const cursive = new Set(serif.map(key));
  return (
    <div
      style={{
        fontFamily: DISPLAY,
        fontWeight: weight,
        fontSize: size,
        lineHeight: 1.06,
        letterSpacing: '-0.03em',
        color: C.text,
        textAlign: align,
        width,
        display: 'flex',
        flexWrap: 'wrap',
        gap: `0 ${size * 0.245}px`,
        justifyContent: align === 'center' ? 'center' : 'flex-start',
      }}
    >
      {words.map((w, i) => {
        const a = ramp(frame, delay + i * 2.1, 15);
        const k = key(w);
        const isHot = hot.has(k);
        const isSerif = cursive.has(k);
        return (
          <span
            key={i}
            style={{
              display: 'inline-block',
              opacity: a,
              transform: `translateY(${(1 - a) * 24}px)`,
              ...(isHot ? grad : null),
              ...(isSerif
                ? {
                    fontFamily: SERIF,
                    fontStyle: 'italic',
                    fontWeight: 400,
                    fontSize: size * 1.1,
                    letterSpacing: '-0.01em',
                  }
                : null),
            }}
          >
            {w}
          </span>
        );
      })}
    </div>
  );
};

export const Body: React.FC<{
  children: React.ReactNode;
  delay?: number;
  size?: number;
  width?: number;
  color?: string;
}> = ({children, delay = 0, size = 33, width = 900, color = C.muted}) => {
  const e = useEnter(delay);
  return (
    <div
      style={{
        fontFamily: BODY,
        fontWeight: 300,
        fontSize: size,
        lineHeight: 1.52,
        color,
        maxWidth: width,
        opacity: e,
        transform: `translateY(${(1 - e) * 14}px)`,
      }}
    >
      {children}
    </div>
  );
};

/** The gradient rule that opens most slides. */
export const Rule: React.FC<{delay?: number; width?: number; height?: number}> = ({
  delay = 0,
  width = 132,
  height = 4,
}) => {
  const frame = useCurrentFrame();
  const a = ramp(frame, delay, 20);
  return (
    <div
      style={{
        width: width * a,
        height,
        borderRadius: height,
        backgroundImage: GRAD,
      }}
    />
  );
};

// ---------------------------------------------------------------------------
// Surfaces

/**
 * A glass panel.
 *
 * On a dark ground a card cannot just be a lighter rectangle — it has to catch
 * a highlight along its top edge, or it reads as a hole rather than a surface.
 * That is the inset white hairline; the outer shadow does the lifting.
 */
export const Glass: React.FC<{
  children: React.ReactNode;
  delay?: number;
  active?: boolean;
  radius?: number;
  style?: React.CSSProperties;
}> = ({children, delay = 0, active = false, radius = 24, style}) => {
  const e = useEnter(delay);
  return (
    <div
      style={{
        background: active
          ? 'linear-gradient(180deg, rgba(255,77,109,0.14), rgba(139,92,246,0.10))'
          : `linear-gradient(180deg, ${C.ink2}, ${C.ink1})`,
        border: `1px solid ${active ? 'rgba(255,120,150,0.42)' : C.line}`,
        borderRadius: radius,
        boxShadow: active
          ? `0 26px 70px rgba(0,0,0,0.55), inset 0 1px 0 rgba(255,255,255,0.16), 0 0 46px rgba(255,77,109,0.16)`
          : `0 22px 60px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.09)`,
        opacity: e,
        transform: `translateY(${(1 - e) * 26}px) scale(${0.972 + e * 0.028})`,
        ...style,
      }}
    >
      {children}
    </div>
  );
};

/** A soft disc behind an icon so it is not floating on flat black. */
export const Halo: React.FC<{size: number; color?: string; opacity?: number}> = ({
  size,
  color = C.rose,
  opacity = 0.2,
}) => (
  <div
    style={{
      position: 'absolute',
      width: size,
      height: size,
      left: '50%',
      top: '50%',
      transform: 'translate(-50%,-50%)',
      borderRadius: '50%',
      background: `radial-gradient(circle, ${color}${Math.round(opacity * 255)
        .toString(16)
        .padStart(2, '0')} 0%, transparent 66%)`,
    }}
  />
);

/** The progress rail and chapter label at the bottom of every graphic frame. */
export const Rail: React.FC<{progress: number; chapter?: string}> = ({progress, chapter}) => (
  <AbsoluteFill style={{pointerEvents: 'none'}}>
    <div
      style={{
        position: 'absolute',
        left: 0,
        right: 0,
        bottom: 0,
        height: 5,
        backgroundColor: 'rgba(255,255,255,0.07)',
      }}
    >
      <div
        style={{
          width: `${Math.max(0, Math.min(1, progress)) * 100}%`,
          height: '100%',
          backgroundImage: GRAD_WIDE,
        }}
      />
    </div>
    {chapter ? (
      <div
        style={{
          position: 'absolute',
          left: 58,
          bottom: 26,
          fontFamily: MONO,
          fontSize: 17,
          letterSpacing: '0.2em',
          textTransform: 'uppercase',
          color: 'rgba(255,255,255,0.22)',
        }}
      >
        {chapter}
      </div>
    ) : null}
  </AbsoluteFill>
);

// ---------------------------------------------------------------------------
// Glyphs
//
// Line art on a 96-box, one stroke weight, drawn rather than typed — the same
// reasoning as the reference cut: emoji carry another vendor's palette and
// render at another vendor's weight, and one of them ships as a solid green
// tile on macOS. These are the vocabulary this lesson actually needs.

export type GlyphName =
  | 'feed' | 'heart' | 'grid' | 'camera' | 'upload' | 'bell' | 'chat' | 'user'
  | 'sparkle' | 'lock' | 'db' | 'phone' | 'globe' | 'clock' | 'wand' | 'brush'
  | 'eye' | 'split' | 'stack' | 'check' | 'warn' | 'plug';

export const Glyph: React.FC<{
  name: GlyphName;
  size?: number;
  color?: string;
  on?: number;
}> = ({name, size = 68, color = C.rose, on = 1}) => {
  const p = {
    fill: 'none',
    stroke: color,
    strokeWidth: 4.2,
    strokeLinecap: 'round' as const,
    strokeLinejoin: 'round' as const,
  };
  const soft = {...p, opacity: 0.42};
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 96 96"
      style={{opacity: on, transform: `scale(${0.88 + on * 0.12})`, overflow: 'visible'}}
    >
      {name === 'feed' && (
        <>
          <rect x={18} y={12} width={60} height={26} rx={7} {...p} />
          <rect x={18} y={50} width={60} height={34} rx={7} {...soft} />
          <line x1={28} y1={25} x2={52} y2={25} {...p} />
        </>
      )}
      {name === 'heart' && (
        <path d="M48 80 C24 62 14 50 14 38 A17 17 0 0 1 48 30 A17 17 0 0 1 82 38 C82 50 72 62 48 80 Z" {...p} />
      )}
      {name === 'grid' && (
        <>
          <rect x={14} y={14} width={28} height={28} rx={5} {...p} />
          <rect x={54} y={14} width={28} height={28} rx={5} {...soft} />
          <rect x={14} y={54} width={28} height={28} rx={5} {...soft} />
          <rect x={54} y={54} width={28} height={28} rx={5} {...p} />
        </>
      )}
      {name === 'camera' && (
        <>
          <rect x={12} y={26} width={72} height={50} rx={9} {...p} />
          <circle cx={48} cy={51} r={15} {...p} />
          <path d="M36 26 L41 17 L55 17 L60 26" {...soft} />
          <circle cx={71} cy={38} r={2.8} {...p} />
        </>
      )}
      {name === 'upload' && (
        <>
          <path d="M20 60 L20 72 C20 77 23 80 28 80 L68 80 C73 80 76 77 76 72 L76 60" {...soft} />
          <line x1={48} y1={64} x2={48} y2={18} {...p} />
          <polyline points="32,34 48,18 64,34" {...p} />
        </>
      )}
      {name === 'bell' && (
        <>
          <path d="M26 62 C26 44 30 26 48 26 C66 26 70 44 70 62 L76 70 L20 70 Z" {...p} />
          <path d="M40 78 A8 8 0 0 0 56 78" {...p} />
          <line x1={48} y1={26} x2={48} y2={17} {...soft} />
        </>
      )}
      {name === 'chat' && (
        <>
          <path d="M82 44 C82 60 67 73 48 73 A44 44 0 0 1 37 71.6 L18 79 L23 64 A26 26 0 0 1 14 44 C14 28 29 15 48 15 C67 15 82 28 82 44 Z" {...p} />
          <line x1={34} y1={42} x2={62} y2={42} {...soft} />
          <line x1={34} y1={54} x2={52} y2={54} {...soft} />
        </>
      )}
      {name === 'user' && (
        <>
          <circle cx={48} cy={34} r={15} {...p} />
          <path d="M18 82 A30 30 0 0 1 78 82" {...p} />
        </>
      )}
      {name === 'sparkle' && (
        <>
          <path d="M48 12 L54 40 L82 46 L54 52 L48 80 L42 52 L14 46 L42 40 Z" {...p} />
          <path d="M76 14 L79 24 L89 27 L79 30 L76 40 L73 30 L63 27 L73 24 Z" {...soft} />
        </>
      )}
      {name === 'lock' && (
        <>
          <rect x={20} y={44} width={56} height={38} rx={9} {...p} />
          <path d="M33 44 L33 32 A15 15 0 0 1 63 32 L63 44" {...p} />
          <line x1={48} y1={58} x2={48} y2={68} {...soft} />
        </>
      )}
      {name === 'db' && (
        <>
          <ellipse cx={48} cy={24} rx={28} ry={10} {...p} />
          <path d="M20 24 L20 72 C20 78 33 82 48 82 C63 82 76 78 76 72 L76 24" {...p} />
          <path d="M20 48 C20 54 33 58 48 58 C63 58 76 54 76 48" {...soft} />
        </>
      )}
      {name === 'phone' && (
        <>
          <rect x={30} y={10} width={36} height={76} rx={9} {...p} />
          <line x1={42} y1={20} x2={54} y2={20} {...soft} />
          <circle cx={48} cy={76} r={3} {...p} />
        </>
      )}
      {name === 'globe' && (
        <>
          <circle cx={48} cy={48} r={32} {...p} />
          <ellipse cx={48} cy={48} rx={14} ry={32} {...soft} />
          <line x1={16} y1={48} x2={80} y2={48} {...soft} />
          <path d="M22 30 Q48 42 74 30" {...soft} />
          <path d="M22 66 Q48 54 74 66" {...soft} />
        </>
      )}
      {name === 'clock' && (
        <>
          <circle cx={48} cy={48} r={32} {...p} />
          <polyline points="48,28 48,50 65,58" {...p} />
        </>
      )}
      {name === 'wand' && (
        <>
          <line x1={22} y1={76} x2={62} y2={36} {...p} />
          <path d="M62 36 L74 24" {...p} strokeWidth={6} />
          <path d="M32 20 L34 28 L42 30 L34 32 L32 40 L30 32 L22 30 L30 28 Z" {...soft} />
          <circle cx={76} cy={54} r={2.6} {...soft} />
        </>
      )}
      {name === 'brush' && (
        <>
          <path d="M62 20 L76 34 L44 66 L30 52 Z" {...p} />
          <path d="M30 52 L20 76 L44 66" {...soft} />
        </>
      )}
      {name === 'eye' && (
        <>
          <path d="M10 48 C24 28 72 28 86 48 C72 68 24 68 10 48 Z" {...p} />
          <circle cx={48} cy={48} r={11} {...p} />
        </>
      )}
      {name === 'split' && (
        <>
          <line x1={16} y1={48} x2={40} y2={48} {...p} />
          <path d="M40 48 L58 28 L80 28" {...p} />
          <path d="M40 48 L58 68 L80 68" {...soft} />
          <circle cx={40} cy={48} r={4} {...p} />
        </>
      )}
      {name === 'stack' && (
        <>
          <path d="M48 14 L80 30 L48 46 L16 30 Z" {...p} />
          <path d="M16 46 L48 62 L80 46" {...soft} />
          <path d="M16 62 L48 78 L80 62" {...soft} />
        </>
      )}
      {name === 'check' && (
        <>
          <circle cx={48} cy={48} r={31} {...soft} />
          <polyline points="32,49 43,60 66,37" {...p} />
        </>
      )}
      {name === 'warn' && (
        <>
          <path d="M48 16 L82 76 L14 76 Z" {...p} />
          <line x1={48} y1={40} x2={48} y2={56} {...p} />
          <circle cx={48} cy={66} r={2.8} fill={color} stroke="none" />
        </>
      )}
      {name === 'plug' && (
        <>
          <path d="M36 14 L36 34 M60 14 L60 34" {...soft} />
          <rect x={26} y={34} width={44} height={26} rx={8} {...p} />
          <line x1={48} y1={60} x2={48} y2={82} {...p} />
        </>
      )}
    </svg>
  );
};

/** A drifting mote field, used only on title and chapter cards. */
export const Motes: React.FC<{n?: number; seed?: string}> = ({n = 30, seed = 'm'}) => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{pointerEvents: 'none'}}>
      {Array.from({length: n}).map((_, i) => {
        const rx = random(`${seed}x${i}`);
        const ry = random(`${seed}y${i}`);
        const rs = random(`${seed}s${i}`);
        // one very slow vertical drift, so a two-second card is never still
        const dy = Math.sin((frame / FPS / (14 + rs * 10) + rx) * Math.PI * 2) * 7;
        return (
          <div
            key={i}
            style={{
              position: 'absolute',
              left: `${rx * 100}%`,
              top: `${ry * 100}%`,
              width: 2 + rs * 3,
              height: 2 + rs * 3,
              borderRadius: '50%',
              backgroundColor: rs > 0.6 ? C.cyan : C.rose,
              opacity: 0.10 + rs * 0.22,
              transform: `translateY(${dy}px)`,
            }}
          />
        );
      })}
    </AbsoluteFill>
  );
};
