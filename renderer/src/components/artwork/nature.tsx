// The world, weather and growing things.
//
// These are the figures a lesson reaches for when the subject is not a machine
// — scale, cycles, energy, decay. They are also the set where the idle motion
// is most obviously right: nothing in nature is still, so a tree that does not
// sway and a flame that does not lick are the two most obviously fake drawings
// this vocabulary could contain.

import {
  C,
  bob,
  cycle,
  ease,
  fadeTravel,
  popAt,
  pulse,
  ring,
  stage,
  type Figure,
} from './kit';

/** A globe with its meridians turning. */
export const GlobeFigure: Figure = ({build, t, palette}) => {
  const ball = stage(build, 0, 2);
  const lines = stage(build, 1, 2);
  const spin = t * 0.55;
  return (
    <g>
      <g transform={popAt(ball, C, C)} opacity={ball}>
        <circle cx={C} cy={C} r={68} fill={palette.primary} />
        <path d="M100 32 A68 68 0 0 1 100 168 Z" fill={palette.ink} opacity={0.12} />
      </g>
      <g opacity={lines} clipPath="url(#globeClip)">
        <defs>
          <clipPath id="globeClip">
            <circle cx={C} cy={C} r={68} />
          </clipPath>
        </defs>
        {/* Latitudes are fixed; longitudes ride the spin, which is the whole
            reason a flat circle reads as a sphere. */}
        {[-34, 0, 34].map((dy, i) => (
          <ellipse key={i} cx={C} cy={C + dy} rx={Math.sqrt(Math.max(0, 68 * 68 - dy * dy))} ry={5} fill="none" stroke={palette.soft} strokeWidth={3} opacity={0.4} />
        ))}
        {[0, 1, 2, 3].map((i) => {
          const rx = Math.abs(Math.cos(spin + (i * Math.PI) / 4)) * 68;
          return <ellipse key={i} cx={C} cy={C} rx={Math.max(1, rx)} ry={68} fill="none" stroke={palette.soft} strokeWidth={3} opacity={0.4} />;
        })}
      </g>
    </g>
  );
};

/** A tree whose canopy sways off its trunk. */
export const TreeFigure: Figure = ({build, t, palette}) => {
  const trunk = stage(build, 0, 2);
  const canopy = stage(build, 1, 2);
  const sway = bob(t, 2.6, 0.9);
  return (
    <g>
      <g opacity={trunk}>
        <path d="M92 168 L92 108 L108 108 L108 168 Z" fill={palette.line} opacity={0.7} />
        <path d="M100 128 L78 108 M100 118 L122 98" stroke={palette.line} strokeWidth={5} strokeLinecap="round" fill="none" opacity={0.6} />
      </g>
      <g transform={`${popAt(canopy, C, 82)} rotate(${sway} 100 118)`} opacity={canopy}>
        <circle cx={C} cy={72} r={40} fill={palette.primary} />
        <circle cx={68} cy={96} r={30} fill={palette.primary} />
        <circle cx={132} cy={96} r={30} fill={palette.primary} />
        <circle cx={118} cy={64} r={24} fill={palette.accent} opacity={0.45} />
      </g>
    </g>
  );
};

/** A leaf with a vein that draws and an edge that curls. */
export const LeafFigure: Figure = ({build, t, palette}) => {
  const blade = stage(build, 0, 2);
  const veins = stage(build, 1, 2);
  const curl = bob(t, 4, 1.1);
  return (
    <g transform={`rotate(${curl} 100 150)`}>
      <g transform={popAt(blade, C, 100)} opacity={blade}>
        <path d="M100 150 C40 130 34 62 100 30 C166 62 160 130 100 150 Z" fill={palette.primary} />
        <path d="M100 30 C166 62 160 130 100 150 Z" fill={palette.ink} opacity={0.13} />
      </g>
      <g opacity={veins} stroke={palette.soft} strokeWidth={3} strokeLinecap="round" fill="none">
        <path d="M100 150 L100 34" opacity={0.7} />
        {[0, 1, 2].map((i) => (
          <g key={i}>
            <path d={`M100 ${120 - i * 26} L${74 - i * 4} ${100 - i * 26}`} opacity={0.5} />
            <path d={`M100 ${120 - i * 26} L${126 + i * 4} ${100 - i * 26}`} opacity={0.5} />
          </g>
        ))}
      </g>
    </g>
  );
};

/** Peaks with a snow cap and a sun behind them. */
export const MountainFigure: Figure = ({build, t, palette}) => {
  const back = stage(build, 0, 3);
  const front = stage(build, 1, 3);
  const sun = stage(build, 2, 3);
  const rise = bob(t, 4, 0.6);
  return (
    <g>
      <circle cx={142} cy={58 + rise} r={20} fill={palette.accent} opacity={sun * 0.9} />
      <g opacity={back}>
        <path d="M96 156 L140 74 L184 156 Z" fill={palette.primary} opacity={0.55} />
      </g>
      <g opacity={front}>
        <path d="M16 156 L76 52 L136 156 Z" fill={palette.primary} />
        <path d="M76 52 L136 156 L76 156 Z" fill={palette.ink} opacity={0.14} />
        <path d="M56 86 L76 52 L96 86 L84 80 L74 88 L64 80 Z" fill={palette.soft} />
      </g>
      <rect x={10} y={154} width={180} height={7} rx={3.5} fill={palette.line} opacity={0.4 * front} />
    </g>
  );
};

/** A sun whose rays turn and breathe. */
export const SunFigure: Figure = ({build, t, palette}) => {
  const disc = stage(build, 0, 2);
  const rays = stage(build, 1, 2);
  const breathe = pulse(t, 1.6);
  return (
    <g>
      <g opacity={rays} transform={`rotate(${t * 6} 100 100)`}>
        {ring(12, 66).map((p, i) => {
          const len = i % 2 === 0 ? 22 : 13;
          const r0 = 66 + breathe * 4;
          return (
            <line
              key={i}
              x1={C + r0 * Math.cos(p.a)}
              y1={C + r0 * Math.sin(p.a)}
              x2={C + (r0 + len) * Math.cos(p.a)}
              y2={C + (r0 + len) * Math.sin(p.a)}
              stroke={palette.accent}
              strokeWidth={6}
              strokeLinecap="round"
              opacity={0.5 + 0.5 * breathe}
            />
          );
        })}
      </g>
      <g transform={popAt(disc, C, C)} opacity={disc}>
        <circle cx={C} cy={C} r={52} fill={palette.accent} />
        <circle cx={84} cy={84} r={16} fill={palette.soft} opacity={0.22} />
      </g>
    </g>
  );
};

/** A crescent moon with stars that come and go around it. */
export const MoonFigure: Figure = ({build, t, palette}) => {
  const moon = stage(build, 0, 2);
  const stars = stage(build, 1, 2);
  return (
    <g>
      <g opacity={stars}>
        {ring(5, 74, C, C, 0.9).map((p, i) => (
          <path
            key={i}
            d={`M${p.x} ${p.y - 7} L${p.x + 2} ${p.y - 2} L${p.x + 7} ${p.y} L${p.x + 2} ${p.y + 2} L${p.x} ${p.y + 7} L${p.x - 2} ${p.y + 2} L${p.x - 7} ${p.y} L${p.x - 2} ${p.y - 2} Z`}
            fill={palette.accent}
            opacity={0.25 + 0.75 * pulse(t, 1.8 + i * 0.5, i * 1.7)}
          />
        ))}
      </g>
      <g transform={`${popAt(moon, C, C)} rotate(${bob(t, 3, 0.7)} 100 100)`} opacity={moon}>
        {/* A crescent as one disc with a second punched out of it, rather than
            as two arcs meeting. Two arcs have to agree about their radii and
            sweep flags to close at all, and the pair that looked right in the
            path string closed onto itself and rendered nothing. A hole cannot
            fail that way, and it does not need the stage's colour to fake. */}
        <path
          fillRule="evenodd"
          d="M94 100 m-60 0 a60 60 0 1 0 120 0 a60 60 0 1 0 -120 0
             M128 88 m-54 0 a54 54 0 1 0 108 0 a54 54 0 1 0 -108 0"
          fill={palette.soft}
        />
        <circle cx={72} cy={82} r={8} fill={palette.ink} opacity={0.14} />
        <circle cx={62} cy={122} r={5} fill={palette.ink} opacity={0.12} />
      </g>
    </g>
  );
};

/** A drop that falls, lands, and rings out. */
export const DropFigure: Figure = ({build, t, palette}) => {
  const drop = stage(build, 0, 2);
  const ripple = stage(build, 1, 2);
  const fall = cycle(t, 0.55);
  const land = fall > 0.7 ? (fall - 0.7) / 0.3 : -1;
  // Two drops half a cycle apart. With one, the figure is empty for the third
  // of every cycle the ripple takes — long enough that a still lands on a bare
  // line more often than on a drop, which is not what "drop" should look like.
  const next = cycle(t, 0.55, 0.5);
  const body = (p: number, opacity: number) => (
    <g opacity={opacity} transform={`translate(0 ${p * 84})`}>
      <path d="M100 24 C122 52 132 66 132 80 A32 32 0 0 1 68 80 C68 66 78 52 100 24 Z" fill={palette.primary} />
      <path d="M100 24 C122 52 132 66 132 80 A32 32 0 0 1 100 112 Z" fill={palette.ink} opacity={0.12} />
      <circle cx={86} cy={82} r={7} fill={palette.soft} opacity={0.35} />
    </g>
  );
  return (
    <g>
      {land >= 0 && next < 0.7 && (
        <g transform="scale(0.62) translate(62 -8)">{body(next, drop * 0.85)}</g>
      )}
      {body(fall, drop * (land < 0 ? 1 : 0))}
      <g opacity={ripple}>
        <line x1={40} y1={158} x2={160} y2={158} stroke={palette.line} strokeWidth={4} strokeLinecap="round" opacity={0.45} />
        {land >= 0 &&
          [0, 1].map((i) => (
            <ellipse
              key={i}
              cx={C}
              cy={158}
              rx={12 + (land + i * 0.3) * 54}
              ry={4 + (land + i * 0.3) * 9}
              fill="none"
              stroke={palette.accent}
              strokeWidth={3}
              opacity={Math.max(0, 0.7 - land - i * 0.3)}
            />
          ))}
      </g>
    </g>
  );
};

/** A flame that never sits still. */
export const FireFigure: Figure = ({build, t, palette}) => {
  const outer = stage(build, 0, 2);
  const inner = stage(build, 1, 2);
  // A slow roll plus a faster flicker, so it reads as combustion rather than
  // as a pulsing shape.
  const lick = 1 + 0.1 * Math.sin(t * 13) + 0.06 * Math.sin(t * 23.7);
  const lean = Math.sin(t * 4.1) * 5;
  return (
    <g>
      <g transform={popAt(outer, C, 130)} opacity={outer}>
        <path
          d={`M100 ${164 - 128 * lick} C${138 + lean} ${96} 146 118 140 136 C134 156 118 168 100 168 C82 168 66 156 60 136 C54 118 ${62 - lean} 96 100 ${164 - 128 * lick} Z`}
          fill={palette.accent}
        />
      </g>
      <g opacity={inner}>
        <path
          d={`M100 ${152 - 74 * lick} C${120 + lean} 116 124 130 120 142 C116 154 110 162 100 162 C90 162 84 154 80 142 C76 130 ${80 - lean} 116 100 ${152 - 74 * lick} Z`}
          fill={palette.soft}
          opacity={0.75}
        />
      </g>
      <ellipse cx={C} cy={168} rx={44} ry={7} fill={palette.line} opacity={0.3 * outer} />
    </g>
  );
};

/** Gusts crossing the frame, one after another. */
export const WindFigure: Figure = ({build, t, palette}) => {
  const lines = [
    {y: 66, w: 96, curl: 1},
    {y: 100, w: 128, curl: -1},
    {y: 134, w: 78, curl: 1},
  ];
  return (
    <g>
      {lines.map((l, i) => {
        const p = stage(build, i, lines.length, 0.55);
        const drift = cycle(t, 0.4, i * 0.28);
        const x = 20 + drift * 40;
        return (
          <g key={i} opacity={p * fadeTravel(drift)}>
            <path
              d={`M${x} ${l.y} L${x + l.w} ${l.y} A${11} ${11} 0 1 ${l.curl > 0 ? 1 : 0} ${x + l.w + (l.curl > 0 ? 4 : -4)} ${l.y + l.curl * 20}`}
              fill="none"
              stroke={i === 1 ? palette.accent : palette.line}
              strokeWidth={7}
              strokeLinecap="round"
              opacity={i === 1 ? 1 : 0.7}
            />
          </g>
        );
      })}
    </g>
  );
};

/** A seed under soil, sending up a shoot. */
export const SeedFigure: Figure = ({build, t, palette}) => {
  const soil = stage(build, 0, 3);
  const seed = stage(build, 1, 3);
  const shoot = stage(build, 2, 3);
  // The shoot grows, holds, and starts over — a sprout at full height forever
  // is just a small plant.
  const g = Math.min(1, ((t % 6) / 6) * 1.6);
  return (
    <g>
      <g opacity={shoot} transform={`rotate(${bob(t, 2, 1.2)} 100 130)`}>
        <path d={`M100 130 L100 ${130 - 70 * g}`} stroke={palette.primary} strokeWidth={7} strokeLinecap="round" fill="none" />
        <ellipse cx={82} cy={130 - 52 * g} rx={19 * g} ry={11 * g} fill={palette.primary} transform={`rotate(-24 82 ${130 - 52 * g})`} />
        <ellipse cx={118} cy={130 - 66 * g} rx={17 * g} ry={10 * g} fill={palette.accent} transform={`rotate(24 118 ${130 - 66 * g})`} />
      </g>
      <g opacity={seed}>
        <ellipse cx={C} cy={140} rx={13} ry={17} fill={palette.line} opacity={0.75} />
      </g>
      <g opacity={soil}>
        {/* A mound, not a slab. Drawn full width in `soft` the soil was the
            biggest, lightest thing in the frame and the figure read as a
            basin with a plant in it. */}
        <path d="M42 158 C42 138 66 128 100 128 C134 128 158 138 158 158 A6 6 0 0 1 152 164 L48 164 A6 6 0 0 1 42 158 Z" fill={palette.line} opacity={0.55} />
        <path d="M100 128 C134 128 158 138 158 158 A6 6 0 0 1 152 164 L100 164 Z" fill={palette.ink} opacity={0.14} />
      </g>
    </g>
  );
};

/** A wave that actually travels. */
export const WaveFigure: Figure = ({build, t, palette}) => {
  const rows = [
    {y: 78, amp: 16, speed: 1.4, color: 'line' as const},
    {y: 108, amp: 20, speed: 1.1, color: 'primary' as const},
    {y: 138, amp: 14, speed: 1.7, color: 'accent' as const},
  ];
  const path = (y: number, amp: number, phase: number) => {
    const pts = Array.from({length: 25}, (_, i) => {
      const x = 16 + i * 7;
      return `${i === 0 ? 'M' : 'L'}${x} ${y + Math.sin(i * 0.42 + phase) * amp}`;
    });
    return pts.join(' ');
  };
  return (
    <g>
      {rows.map((r, i) => {
        const p = stage(build, i, rows.length, 0.5);
        return (
          <path
            key={i}
            d={path(r.y, r.amp, t * r.speed)}
            fill="none"
            stroke={palette[r.color]}
            strokeWidth={7}
            strokeLinecap="round"
            opacity={p * (r.color === 'line' ? 0.6 : 1)}
          />
        );
      })}
    </g>
  );
};

/** Three arrows chasing each other round a loop. */
export const RecycleFigure: Figure = ({build, t, palette}) => {
  const arms = [0, 120, 240];
  return (
    <g transform={`rotate(${t * 24} 100 100)`}>
      {arms.map((deg, i) => {
        const p = stage(build, i, arms.length, 0.5);
        return (
          <g key={i} transform={`rotate(${deg} 100 100)`} opacity={p}>
            <path
              d="M100 40 A60 60 0 0 1 152 70"
              fill="none"
              stroke={i === 0 ? palette.accent : palette.primary}
              strokeWidth={14}
              strokeLinecap="round"
              pathLength={1}
              strokeDasharray={1}
              strokeDashoffset={1 - ease(p)}
            />
            <path d="M156 56 L164 78 L140 80 Z" fill={i === 0 ? palette.accent : palette.primary} opacity={ease(p)} />
          </g>
        );
      })}
    </g>
  );
};
