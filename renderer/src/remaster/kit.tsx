// Shared visual kit for the no-code remaster.
//
// One ground, one window chrome, one type scale, one motion curve. Every scene
// in this cut composes out of these so the graphic slides and the captured
// footage read as the same film rather than as slides cut against screenshots.

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
import {loadFont as loadSora} from '@remotion/google-fonts/Sora';
import {loadFont as loadInter} from '@remotion/google-fonts/Inter';
import {loadFont as loadMono} from '@remotion/google-fonts/JetBrainsMono';

const sora = loadSora('normal', {weights: ['400', '600', '700'], subsets: ['latin']});
const inter = loadInter('normal', {weights: ['400', '500', '600'], subsets: ['latin']});
const mono = loadMono('normal', {weights: ['400', '500'], subsets: ['latin']});

export const DISPLAY = sora.fontFamily;
export const BODY = inter.fontFamily;
export const MONO = mono.fontFamily;

/**
 * Two skins, one set of roles.
 *
 * The cut this kit was written for was dark. The lesson it now sits inside is
 * light throughout — the screen recordings, the quiz cards and the closing
 * roadmap are all on the pale "showroom" ground — so a dark graphic slide reads
 * as a clip borrowed from another film. LIGHT repoints the same role names at
 * the showroom values; nothing downstream needs to know which skin is on.
 *
 * Roles, so a new value lands in the right place:
 *   ink0            the ground behind everything
 *   ink1 / ink2     raised surfaces (cards, panels)
 *   line            hairline borders and dividers
 *   cream           primary text        muted  secondary      dim  tertiary
 *   brand           the accent, on text / borders / fills
 *   brandSoft       the accent at eyebrow weight (needs contrast, not lift)
 */
export const LIGHT = true;

const DARK_C = {
  ink0: '#0A0908',
  ink1: '#131010',
  ink2: '#1D1917',
  line: '#312A26',
  paper: '#FFFFFF',
  cream: '#F7F1EA',
  muted: '#9E938B',
  dim: '#6B615B',
  brand: '#F97316',
  brandDeep: '#C2410C',
  brandSoft: '#FDBA74',
  blue: '#3B82F6',
  ok: '#34D399',
};

const LIGHT_C = {
  ink0: '#F1F2F4',
  ink1: '#FFFFFF',
  ink2: '#F7F8F9',
  line: '#E3E5E8',
  paper: '#16181B',
  cream: '#16181B',
  muted: '#565C63',
  dim: '#8B9198',
  // The accent has to carry text on a near-white ground, so it is the deeper
  // amber the quiz cards already use rather than the dark cut's hot orange.
  brand: '#B4651F',
  brandDeep: '#8A4A12',
  brandSoft: '#9A5A1C',
  blue: '#2563EB',
  ok: '#15803D',
};

export const C = LIGHT ? LIGHT_C : DARK_C;

/** Shadow ink. On light, black at card-shadow strength turns everything grey. */
const SH = (a: number) => (LIGHT ? `rgba(23,27,33,${a * 0.42})` : `rgba(0,0,0,${a})`);

export const FPS = 30;

/** Seconds → frames. */
export const s = (sec: number) => Math.round(sec * FPS);

/** The house ease for anything that travels. */
export const EASE = Easing.bezier(0.22, 1, 0.36, 1);

/** A spring that settles without wobble — used for every entrance. */
export const useEnter = (delay = 0, damping = 22) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  return spring({frame: frame - delay, fps, config: {damping, stiffness: 120, mass: 0.9}});
};

/** 0 → 1 over `len` frames starting at `at`, eased, clamped. */
export const ramp = (frame: number, at: number, len: number) =>
  interpolate(frame, [at, at + len], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: EASE,
  });

// ---------------------------------------------------------------------------
// Ground

/**
 * Grain sits on top of everything at very low opacity.
 *
 * A flat dark fill on a 1080p H.264 encode bands badly in the corners where the
 * bloom falls off; a little noise breaks the gradient into something the encoder
 * can carry. It is generated once as a data URI so no frame pays for it.
 */
const GRAIN =
  "url(\"data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='200' height='200'><filter id='n'><feTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='3'/></filter><rect width='200' height='200' filter='url(%23n)' opacity='0.5'/></svg>\")";

export const Ground: React.FC<{drift?: boolean; children?: React.ReactNode}> = ({
  drift = true,
  children,
}) => {
  // Fixed, deliberately. A bloom that wanders and a light band that sweeps are
  // invisible on a still and maddening over eleven minutes — they put every
  // frame in slow motion even when nothing is happening.
  const bx = 20;
  const by = 11;
  // Same three lights in both skins, but on a pale ground they have to work by
  // *removing* light at the edges rather than adding it in the corner: a warm
  // bloom that reads on black is invisible at 4% over near-white, and the dark
  // cut's inset vignette turns the whole panel grey.
  return (
    <AbsoluteFill style={{backgroundColor: C.ink0}}>
      {/* the key light, top left */}
      <AbsoluteFill
        style={{
          background: LIGHT
            ? 'radial-gradient(120% 92% at 18% 6%, #FFFFFF 0%, #FDFBF8 34%, transparent 72%)'
            : `radial-gradient(115% 88% at ${bx}% ${by}%, #E2590Cbb 0%, #B3450A55 26%, #6B2A0722 46%, transparent 68%)`,
        }}
      />
      {/* a counter-bounce, bottom right, so the frame is not one flat colour */}
      <AbsoluteFill
        style={{
          background: LIGHT
            ? 'radial-gradient(92% 76% at 94% 96%, #E7E4DF88 0%, #EEEDEA33 42%, transparent 68%)'
            : `radial-gradient(88% 72% at 92% 94%, #8C2D0E66 0%, #47180722 40%, transparent 66%)`,
        }}
      />
      {/* a fixed diagonal lift through the middle band */}
      <AbsoluteFill
        style={{
          background: LIGHT
            ? 'linear-gradient(104deg, transparent 24%, #FFFFFFAA 50%, transparent 76%)'
            : 'linear-gradient(104deg, transparent 24%, #FFB2771A 50%, transparent 76%)',
        }}
      />
      <AbsoluteFill
        style={{
          backgroundImage: GRAIN,
          opacity: LIGHT ? 0.022 : 0.04,
          mixBlendMode: LIGHT ? 'multiply' : 'overlay',
        }}
      />
      {children}
      <AbsoluteFill
        style={{
          boxShadow: LIGHT
            ? 'inset 0 0 220px rgba(150,142,130,0.16)'
            : 'inset 0 0 300px rgba(0,0,0,0.6)',
          pointerEvents: 'none',
        }}
      />
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------
// Type

export const Eyebrow: React.FC<{children: React.ReactNode; delay?: number}> = ({
  children,
  delay = 0,
}) => {
  const e = useEnter(delay);
  return (
    <div
      style={{
        fontFamily: MONO,
        fontSize: 22,
        letterSpacing: '0.24em',
        textTransform: 'uppercase',
        color: C.brandSoft,
        opacity: e,
        transform: `translateY(${(1 - e) * 14}px)`,
      }}
    >
      {children}
    </div>
  );
};

/** A headline that lifts in word by word. */
export const Headline: React.FC<{
  text: string;
  size?: number;
  delay?: number;
  accent?: string[];
  align?: 'left' | 'center';
  width?: number | string;
}> = ({text, size = 92, delay = 0, accent = [], align = 'left', width}) => {
  const frame = useCurrentFrame();
  const words = text.split(' ');
  const hot = new Set(accent.map((w) => w.toLowerCase().replace(/[.,?!]/g, '')));
  return (
    <div
      style={{
        fontFamily: DISPLAY,
        fontWeight: 700,
        fontSize: size,
        lineHeight: 1.08,
        letterSpacing: '-0.025em',
        color: C.cream,
        textAlign: align,
        width,
        display: 'flex',
        flexWrap: 'wrap',
        gap: `0 ${size * 0.26}px`,
        justifyContent: align === 'center' ? 'center' : 'flex-start',
      }}
    >
      {words.map((w, i) => {
        const a = ramp(frame, delay + i * 2.2, 16);
        const isHot = hot.has(w.toLowerCase().replace(/[.,?!]/g, ''));
        return (
          <span
            key={i}
            style={{
              display: 'inline-block',
              opacity: a,
              transform: `translateY(${(1 - a) * 26}px)`,
              color: isHot ? C.brand : undefined,
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
}> = ({children, delay = 0, size = 34, width = 900}) => {
  const e = useEnter(delay);
  return (
    <div
      style={{
        fontFamily: BODY,
        fontSize: size,
        lineHeight: 1.5,
        color: C.muted,
        maxWidth: width,
        opacity: e,
        transform: `translateY(${(1 - e) * 16}px)`,
      }}
    >
      {children}
    </div>
  );
};

/** The thin brand rule that opens most slides. */
export const Rule: React.FC<{delay?: number; width?: number}> = ({delay = 0, width = 120}) => {
  const frame = useCurrentFrame();
  const a = ramp(frame, delay, 18);
  return (
    <div
      style={{
        width: width * a,
        height: 4,
        borderRadius: 2,
        backgroundColor: C.brand,
      }}
    />
  );
};

// ---------------------------------------------------------------------------
// Surfaces

export const Card: React.FC<{
  children: React.ReactNode;
  delay?: number;
  active?: boolean;
  style?: React.CSSProperties;
}> = ({children, delay = 0, active = false, style}) => {
  const e = useEnter(delay);
  return (
    <div
      style={{
        backgroundColor: active ? (LIGHT ? '#FDF4E9' : '#241C17') : C.ink1,
        border: `1px solid ${active ? C.brand : C.line}`,
        borderRadius: 22,
        boxShadow: active
          ? `0 30px 70px ${SH(0.55)}, 0 0 0 1px ${C.brand}${LIGHT ? '66' : '40'}, 0 0 60px ${C.brand}${LIGHT ? '1f' : '22'}`
          : `0 22px 54px ${SH(0.45)}`,
        opacity: e,
        transform: `translateY(${(1 - e) * 30}px) scale(${0.965 + e * 0.035})`,
        ...style,
      }}
    >
      {children}
    </div>
  );
};

/** A soft glow disc, used behind icons so they are not floating on flat black. */
export const Halo: React.FC<{size: number; color?: string; opacity?: number}> = ({
  size,
  color = C.brand,
  opacity = LIGHT ? 0.14 : 0.22,
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
      background: `radial-gradient(circle, ${color}${Math.round(opacity * 255).toString(16).padStart(2, '0')} 0%, transparent 68%)`,
    }}
  />
);

// ---------------------------------------------------------------------------
// Progress rail — a 5px bar at the very bottom of every frame.

export const Rail: React.FC<{progress: number; chapter?: string}> = ({progress, chapter}) => (
  <AbsoluteFill style={{pointerEvents: 'none'}}>
    <div
      style={{
        position: 'absolute',
        left: 0,
        right: 0,
        bottom: 0,
        height: 5,
        backgroundColor: LIGHT ? 'rgba(23,27,33,0.09)' : '#ffffff10',
      }}
    >
      <div
        style={{
          width: `${Math.max(0, Math.min(1, progress)) * 100}%`,
          height: '100%',
          background: `linear-gradient(90deg, ${C.brandDeep}, ${C.brand})`,
        }}
      />
    </div>
    {chapter ? (
      <div
        style={{
          position: 'absolute',
          left: 56,
          bottom: 26,
          fontFamily: MONO,
          fontSize: 17,
          letterSpacing: '0.18em',
          textTransform: 'uppercase',
          color: LIGHT ? 'rgba(23,27,33,0.30)' : '#ffffff35',
        }}
      >
        {chapter}
      </div>
    ) : null}
  </AbsoluteFill>
);

// ---------------------------------------------------------------------------
// Sparkle field — a handful of drifting motes, only on title/chapter cards.

export const Motes: React.FC<{n?: number; seed?: string}> = ({n = 26, seed = 'm'}) => {
  // Static. Drifting them was the "glitter": a field of specks all crawling at
  // slightly different rates, on cards that are only on screen for two seconds.
  return (
    <AbsoluteFill style={{pointerEvents: 'none'}}>
      {Array.from({length: n}).map((_, i) => {
        const rx = random(`${seed}x${i}`);
        const ry = random(`${seed}y${i}`);
        const rs = random(`${seed}s${i}`);
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
              backgroundColor: LIGHT ? C.brand : C.brandSoft,
              opacity: (LIGHT ? 0.05 : 0.1) + rs * (LIGHT ? 0.09 : 0.2),
            }}
          />
        );
      })}
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------------
// Glyphs
//
// Drawn rather than typed. Emoji were the first thing tried and they are the
// single fastest way to make a deck look like a template: they carry another
// vendor's colour palette, they render at a different weight from everything
// around them, and a couple of them (the eight-spoked asterisk, notably) come
// out as a solid green tile on macOS. These are line art on the brand stroke,
// so a row of them reads as one set.

type GlyphName =
  | 'web' | 'mobile' | 'server' | 'loop' | 'sprout' | 'keys'
  | 'python' | 'bolt' | 'braces' | 'star' | 'folder';

export const Glyph: React.FC<{
  name: GlyphName;
  size?: number;
  color?: string;
  /** 0 → 1, for drawing the icon on as it enters. */
  on?: number;
}> = ({name, size = 72, color = C.brand, on = 1}) => {
  const p = {
    fill: 'none',
    stroke: color,
    strokeWidth: 4.4,
    strokeLinecap: 'round' as const,
    strokeLinejoin: 'round' as const,
  };
  const soft = {...p, opacity: 0.45};
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 96 96"
      style={{opacity: on, transform: `scale(${0.86 + on * 0.14})`, overflow: 'visible'}}
    >
      {name === 'web' && (
        <>
          <circle cx={48} cy={48} r={32} {...p} />
          <ellipse cx={48} cy={48} rx={14} ry={32} {...soft} />
          <line x1={16} y1={48} x2={80} y2={48} {...soft} />
          <path d="M22 30 Q48 42 74 30" {...soft} />
          <path d="M22 66 Q48 54 74 66" {...soft} />
        </>
      )}
      {name === 'mobile' && (
        <>
          <rect x={30} y={12} width={36} height={72} rx={8} {...p} />
          <line x1={42} y1={22} x2={54} y2={22} {...soft} />
          <circle cx={48} cy={73} r={3.2} {...p} />
        </>
      )}
      {name === 'server' && (
        <>
          <rect x={16} y={20} width={64} height={22} rx={6} {...p} />
          <rect x={16} y={54} width={64} height={22} rx={6} {...p} />
          <circle cx={28} cy={31} r={3.2} {...p} />
          <circle cx={28} cy={65} r={3.2} {...p} />
          <line x1={44} y1={31} x2={68} y2={31} {...soft} />
          <line x1={44} y1={65} x2={68} y2={65} {...soft} />
        </>
      )}
      {name === 'loop' && (
        <>
          <path d="M22 42 A26 26 0 0 1 74 42" {...p} />
          <path d="M74 54 A26 26 0 0 1 22 54" {...p} />
          <polyline points="66,32 76,42 66,50" {...p} />
          <polyline points="30,64 20,54 30,46" {...p} />
        </>
      )}
      {name === 'sprout' && (
        <>
          <path d="M48 80 L48 44" {...p} />
          <path d="M48 52 C34 52 26 44 26 32 C40 32 48 40 48 52 Z" {...p} />
          <path d="M48 46 C60 46 68 39 68 28 C56 28 48 35 48 46 Z" {...soft} />
          <path d="M30 80 Q48 72 66 80" {...soft} />
        </>
      )}
      {name === 'keys' && (
        <>
          <rect x={12} y={28} width={72} height={44} rx={8} {...p} />
          {[24, 38, 52, 66].map((x) => (
            <line key={x} x1={x} y1={41} x2={x} y2={41} {...p} strokeWidth={7} />
          ))}
          {[24, 38, 52, 66].map((x) => (
            <line key={`b${x}`} x1={x} y1={53} x2={x} y2={53} {...soft} strokeWidth={7} />
          ))}
          <line x1={34} y1={62} x2={62} y2={62} {...p} strokeWidth={6} />
        </>
      )}
      {name === 'python' && (
        <>
          <path d="M48 14 C34 14 30 20 30 28 L30 40 L48 40 L48 44 L24 44 C16 44 12 52 12 62 C12 72 16 80 24 80 L30 80 L30 66 C30 58 34 54 42 54 L58 54" {...p} />
          <path d="M48 82 C62 82 66 76 66 68 L66 56 L48 56 L48 52 L72 52 C80 52 84 44 84 34 C84 24 80 16 72 16 L66 16 L66 30 C66 38 62 42 54 42 L38 42" {...soft} />
        </>
      )}
      {name === 'bolt' && (
        <polygon points="54,10 26,52 44,52 38,86 70,42 52,42" {...p} />
      )}
      {name === 'braces' && (
        <>
          <path d="M38 16 C28 16 30 40 20 48 C30 56 28 80 38 80" {...p} />
          <path d="M58 16 C68 16 66 40 76 48 C66 56 68 80 58 80" {...p} />
          <circle cx={48} cy={48} r={3.4} {...p} />
        </>
      )}
      {name === 'star' && (
        <>
          {Array.from({length: 10}).map((_, i) => {
            const a = (i / 10) * Math.PI * 2;
            const r0 = 12;
            const r1 = i % 2 === 0 ? 34 : 27;
            return (
              <line
                key={i}
                x1={48 + Math.cos(a) * r0}
                y1={48 + Math.sin(a) * r0}
                x2={48 + Math.cos(a) * r1}
                y2={48 + Math.sin(a) * r1}
                {...p}
                strokeWidth={5.2}
              />
            );
          })}
        </>
      )}
      {name === 'folder' && (
        <>
          <path d="M14 30 L14 74 C14 78 17 80 20 80 L76 80 C79 80 82 78 82 74 L82 38 C82 34 79 32 76 32 L48 32 L41 24 L20 24 C17 24 14 26 14 30 Z" {...p} />
          <line x1={14} y1={44} x2={82} y2={44} {...soft} />
        </>
      )}
    </svg>
  );
};
