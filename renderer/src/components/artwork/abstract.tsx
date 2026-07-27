// Shapes for ideas that are not objects.
//
// Direction, choice, constraint, progress. These are the hardest figures in the
// vocabulary to make read, because there is no real thing to look like — an
// "arrow" is fine, but a drawing of *filtering* has to earn its meaning from
// what it does rather than from what it resembles. So the motion in this set is
// not decoration on top of a recognisable object; in most of them it is the
// only thing carrying the meaning at all.

import {
  C,
  bob,
  cycle,
  draws,
  ease,
  fadeTravel,
  popAt,
  pulse,
  ring,
  stage,
  swing,
  type Figure,
} from './kit';

/** An arrow that draws and travels. */
export const ArrowFigure: Figure = ({build, t, palette}) => {
  const shaft = stage(build, 0, 2);
  const head = stage(build, 1, 2);
  const push = bob(t, 7, 1.5);
  return (
    <g transform={`translate(${push} 0)`}>
      <line
        x1={26}
        y1={C}
        x2={132}
        y2={C}
        stroke={palette.primary}
        strokeWidth={16}
        strokeLinecap="round"
        opacity={shaft}
        {...draws(shaft)}
      />
      <g transform={popAt(head, 156, C)} opacity={head}>
        <path d="M120 60 L180 100 L120 140 Z" fill={palette.accent} />
      </g>
    </g>
  );
};

/** A closed circuit with something going round it forever. */
export const LoopFigure: Figure = ({build, t, palette}) => {
  const track = stage(build, 0, 2);
  const runner = stage(build, 1, 2);
  const a = t * 1.3;
  return (
    <g>
      <g opacity={track}>
        <ellipse cx={C} cy={C} rx={64} ry={64} fill="none" stroke={palette.line} strokeWidth={12} opacity={0.45} />
        <path d="M100 36 A64 64 0 0 1 156 68" fill="none" stroke={palette.primary} strokeWidth={12} strokeLinecap="round" {...draws(track)} />
      </g>
      <g opacity={runner}>
        <circle cx={C + 64 * Math.cos(a)} cy={C + 64 * Math.sin(a)} r={13} fill={palette.accent} />
        {/* A short trail, so the direction of travel is legible in a still. */}
        {[1, 2, 3].map((i) => (
          <circle
            key={i}
            cx={C + 64 * Math.cos(a - i * 0.16)}
            cy={C + 64 * Math.sin(a - i * 0.16)}
            r={13 - i * 3}
            fill={palette.accent}
            opacity={0.4 - i * 0.1}
          />
        ))}
      </g>
    </g>
  );
};

/** A funnel: a lot goes in, a little comes out. */
export const FunnelFigure: Figure = ({build, t, palette}) => {
  const cone = stage(build, 0, 2);
  const flow = stage(build, 1, 2);
  return (
    <g>
      <g opacity={flow}>
        {[0, 1, 2, 3, 4].map((i) => {
          const p = cycle(t, 0.55, i * 0.2);
          // Above the neck the items are spread across the mouth; below it they
          // are all on the centre line. Squeezing them together *is* the
          // figure.
          const spread = 1 - Math.min(1, p * 2.1);
          const x = C + (i - 2) * 26 * spread;
          const y = 30 + p * 140;
          return <circle key={i} cx={x} cy={y} r={7} fill={i === 2 ? palette.accent : palette.primary} opacity={fadeTravel(p)} />;
        })}
      </g>
      <g transform={popAt(cone, C, C)} opacity={cone}>
        <path d="M28 52 L172 52 L116 116 L116 168 L84 168 L84 116 Z" fill={palette.soft} opacity={0.6} />
        <path d="M28 52 L172 52 L116 116 L116 168 L84 168 L84 116 Z" fill="none" stroke={palette.line} strokeWidth={6} strokeLinejoin="round" opacity={0.8} />
      </g>
    </g>
  );
};

/** A mesh that lets some things through and holds others. */
export const FilterFigure: Figure = ({build, t, palette}) => {
  const mesh = stage(build, 0, 2);
  const items = stage(build, 1, 2);
  return (
    <g>
      <g opacity={items}>
        {[0, 1, 2, 3].map((i) => {
          const passes = i % 2 === 0;
          const p = cycle(t, 0.5, i * 0.25);
          // What is kept stops at the mesh and stays stopped; what passes
          // carries on. Both moving at the same speed reads as no filter.
          const y = passes ? 26 + p * 148 : 26 + Math.min(p * 148, 68);
          return (
            <rect
              key={i}
              x={52 + i * 26}
              y={y}
              width={16}
              height={16}
              rx={passes ? 8 : 3}
              fill={passes ? palette.accent : palette.line}
              opacity={passes ? fadeTravel(p) : 0.8}
            />
          );
        })}
      </g>
      <g opacity={mesh}>
        {/* A screen with holes punched through it, rather than a bar with a
            few marks on it. The bar version read as a slider — which is the
            figure sitting two entries away in this same file. */}
        <path
          fillRule="evenodd"
          d={`M26 86 h148 v30 h-148 z ${[0, 1, 2, 3, 4, 5]
            .map((i) => `M${44 + i * 22} 93 h11 v16 h-11 z`)
            .join(' ')}`}
          fill={palette.primary}
        />
        <rect x={26} y={86} width={148} height={30} fill="none" stroke={palette.soft} strokeWidth={3} opacity={0.3} />
      </g>
    </g>
  );
};

/** A toggle that flips, and means it. */
export const SwitchFigure: Figure = ({build, t, palette}) => {
  const track = stage(build, 0, 2);
  const knob = stage(build, 1, 2);
  const on = swing(t, 4, 0.14, 0.4);
  return (
    <g>
      <g transform={popAt(track, C, C)} opacity={track}>
        <rect x={26} y={70} width={148} height={60} rx={30} fill={palette.line} opacity={0.4} />
        <rect x={26} y={70} width={148} height={60} rx={30} fill={palette.accent} opacity={on} />
      </g>
      <g opacity={knob}>
        <circle cx={58 + on * 84} cy={C} r={24} fill={palette.soft} />
        <circle cx={58 + on * 84} cy={C} r={24} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.14} />
      </g>
    </g>
  );
};

/** Three sliders finding their settings. */
export const SliderFigure: Figure = ({build, t, palette}) => {
  const rows = [
    {y: 62, speed: 0.8, phase: 0},
    {y: 100, speed: 0.6, phase: 1.9},
    {y: 138, speed: 1.0, phase: 3.4},
  ];
  return (
    <g>
      {rows.map((r, i) => {
        const p = stage(build, i, rows.length, 0.5);
        const v = 0.5 + 0.36 * Math.sin(t * r.speed + r.phase);
        return (
          <g key={i} opacity={p}>
            <rect x={30} y={r.y - 5} width={140} height={10} rx={5} fill={palette.line} opacity={0.35} />
            <rect x={30} y={r.y - 5} width={140 * v} height={10} rx={5} fill={palette.primary} />
            <circle cx={30 + 140 * v} cy={r.y} r={15} fill={i === 1 ? palette.accent : palette.soft} />
          </g>
        );
      })}
    </g>
  );
};

/** A piece that drops into the hole it fits. */
export const PuzzleFigure: Figure = ({build, t, palette}) => {
  const board = stage(build, 0, 2);
  const piece = stage(build, 1, 2);
  const seat = swing(t, 4.4, 0.22, 0.42);
  // The hole and the piece are the same path, which is the only way the fit
  // actually looks like a fit.
  const tile =
    'M0 0 L52 0 A13 13 0 0 1 65 13 A13 13 0 0 0 78 26 L78 78 L26 78 A13 13 0 0 0 13 65 A13 13 0 0 1 0 52 Z';
  return (
    <g>
      <g transform={popAt(board, C, C)} opacity={board}>
        <rect x={34} y={34} width={132} height={132} rx={12} fill={palette.primary} />
        <g transform="translate(61 61)" opacity={0.85}>
          <path d={tile} fill={palette.ink} opacity={0.32} />
        </g>
      </g>
      <g opacity={piece} transform={`translate(61 ${61 - (1 - seat) * 62}) rotate(${(1 - seat) * 14} 39 39)`}>
        <path d={tile} fill={palette.accent} />
      </g>
    </g>
  );
};

/** A ladder with a climb running up it. */
export const LadderFigure: Figure = ({build, t, palette}) => {
  const rails = stage(build, 0, 2);
  const rungs = stage(build, 1, 2);
  const climb = cycle(t, 0.32);
  const ys = [50, 74, 98, 122, 146];
  return (
    <g>
      <g opacity={rails}>
        <line x1={62} y1={30} x2={54} y2={172} stroke={palette.primary} strokeWidth={11} strokeLinecap="round" />
        <line x1={138} y1={30} x2={146} y2={172} stroke={palette.primary} strokeWidth={11} strokeLinecap="round" />
      </g>
      <g opacity={rungs}>
        {ys.map((y, i) => {
          // The lit rung walks upward, so the ladder is being climbed rather
          // than standing there.
          const lit = Math.floor((1 - climb) * ys.length) === i;
          return (
            <line
              key={i}
              x1={60 - (y - 30) * 0.056}
              y1={y}
              x2={140 + (y - 30) * 0.056}
              y2={y}
              stroke={lit ? palette.accent : palette.line}
              strokeWidth={9}
              strokeLinecap="round"
              opacity={lit ? 1 : 0.6}
            />
          );
        })}
      </g>
    </g>
  );
};

/** A span across a gap, with something crossing it. */
export const BridgeFigure: Figure = ({build, t, palette}) => {
  const banks = stage(build, 0, 3);
  const span = stage(build, 1, 3);
  const cables = stage(build, 2, 3);
  const at = cycle(t, 0.35);
  const deckY = 108;
  return (
    <g>
      <g opacity={banks}>
        <path d="M4 108 L44 108 L44 170 L4 170 Z" fill={palette.line} opacity={0.5} />
        <path d="M156 108 L196 108 L196 170 L156 170 Z" fill={palette.line} opacity={0.5} />
      </g>
      <g opacity={cables}>
        <path d="M44 108 Q100 46 156 108" fill="none" stroke={palette.primary} strokeWidth={7} strokeLinecap="round" />
        {[0, 1, 2, 3, 4].map((i) => {
          const x = 60 + i * 20;
          const y = 108 - 62 * (1 - Math.pow((x - 100) / 56, 2));
          return <line key={i} x1={x} y1={y} x2={x} y2={deckY} stroke={palette.line} strokeWidth={3} opacity={0.55} />;
        })}
      </g>
      <g opacity={span}>
        <rect x={40} y={deckY} width={120} height={11} rx={5} fill={palette.soft} />
        <rect x={26 + at * 148} y={deckY - 15} width={22} height={15} rx={4} fill={palette.accent} opacity={fadeTravel(at)} />
      </g>
    </g>
  );
};

/** A maze with a route finding its way through. */
export const MazeFigure: Figure = ({build, t, palette}) => {
  const walls = stage(build, 0, 2);
  const route = stage(build, 1, 2);
  // The path draws, holds, and clears: a solved maze held forever is a
  // diagram of a solution, not the act of solving one.
  const solve = swing(t, 6, 0.4, 0.4);
  return (
    <g>
      <g opacity={walls} fill="none" stroke={palette.primary} strokeWidth={8} strokeLinecap="round">
        <rect x={32} y={32} width={136} height={136} rx={8} />
        <path d="M32 68 L104 68 M136 32 L136 100 M68 100 L136 100 M68 100 L68 136 M100 136 L168 136 M100 136 L100 168" />
      </g>
      <path
        d="M48 48 L48 84 L84 84 L84 118 L120 118 L120 152 L152 152"
        fill="none"
        stroke={palette.accent}
        strokeWidth={7}
        strokeLinecap="round"
        strokeLinejoin="round"
        opacity={route}
        pathLength={1}
        strokeDasharray={1}
        strokeDashoffset={1 - solve}
      />
    </g>
  );
};

/** A body with satellites on real orbits. */
export const OrbitFigure: Figure = ({build, t, palette}) => {
  const paths = stage(build, 0, 2);
  const core = stage(build, 1, 2);
  const shells = [
    {rx: 78, ry: 30, tilt: 0, speed: 1.1},
    {rx: 58, ry: 58, tilt: 34, speed: -0.75},
  ];
  return (
    <g>
      {shells.map((s, i) => (
        <g key={i} transform={`rotate(${s.tilt} 100 100)`} opacity={paths}>
          <ellipse cx={C} cy={C} rx={s.rx} ry={s.ry} fill="none" stroke={palette.line} strokeWidth={3.5} opacity={0.5} />
          <circle
            cx={C + s.rx * Math.cos(t * s.speed + i * 2)}
            cy={C + s.ry * Math.sin(t * s.speed + i * 2)}
            r={i === 0 ? 11 : 8}
            fill={i === 0 ? palette.accent : palette.primary}
          />
        </g>
      ))}
      <g transform={popAt(core, C, C)} opacity={core}>
        <circle cx={C} cy={C} r={26} fill={palette.soft} />
        <circle cx={92} cy={92} r={8} fill={palette.ink} opacity={0.12} />
      </g>
    </g>
  );
};

/** A curve going up and to the right, with the point still climbing. */
export const GrowthFigure: Figure = ({build, t, palette}) => {
  const axes = stage(build, 0, 3);
  const curve = stage(build, 1, 3);
  const head = stage(build, 2, 3);
  const lift = bob(t, 4, 1.4);
  return (
    <g>
      <g opacity={axes}>
        <line x1={36} y1={30} x2={36} y2={166} stroke={palette.line} strokeWidth={4} strokeLinecap="round" opacity={0.5} />
        <line x1={36} y1={166} x2={176} y2={166} stroke={palette.line} strokeWidth={4} strokeLinecap="round" opacity={0.5} />
      </g>
      <path
        d={`M46 150 C80 148 96 120 112 96 C126 74 142 ${56 + lift} 166 ${44 + lift}`}
        fill="none"
        stroke={palette.primary}
        strokeWidth={9}
        strokeLinecap="round"
        opacity={curve}
        {...draws(curve)}
      />
      <g opacity={head}>
        <path d={`M144 ${40 + lift} L172 ${38 + lift} L166 ${64 + lift} Z`} fill={palette.accent} />
      </g>
    </g>
  );
};

/** A rocket climbing, with a flame that never sits still. */
export const RocketFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 4);
  const fins = stage(build, 1, 4);
  const window_ = stage(build, 2, 4);
  const flame = stage(build, 3, 4);
  // A slow climb plus a faster flicker, so the flame reads as combustion
  // rather than as a pulsing shape.
  const lick = 1 + 0.17 * Math.sin(t * 16) + 0.09 * Math.sin(t * 27.3);
  const drift = bob(t, 3, 1.7);

  return (
    <g transform={`translate(0 ${drift})`}>
      {/* Exhaust. Two nested teardrops hanging off the nozzle: a lens shape
          closed at both ends read as a sliver of shadow rather than as fire,
          and at half opacity the accent came out darker than the stage. */}
      <g transform={popAt(flame, C, 150)} opacity={flame}>
        <path d={`M84 146 Q100 ${146 + 62 * lick} 116 146 Z`} fill={palette.accent} opacity={0.9} />
        <path d={`M91 146 Q100 ${146 + 36 * lick} 109 146 Z`} fill="#fff3c4" opacity={0.95} />
      </g>
      <g transform={popAt(fins, C, 140)} opacity={fins}>
        <path d="M84 120 L64 152 L84 146 Z" fill={palette.primary} />
        <path d="M116 120 L136 152 L116 146 Z" fill={palette.primary} />
      </g>
      <g transform={popAt(body, C, C)} opacity={body}>
        <path d="M100 34 C122 56 130 92 130 122 L130 148 L70 148 L70 122 C70 92 78 56 100 34 Z" fill={palette.soft} />
        {/* The shaded side is what stops a flat shape reading as a sticker. */}
        <path d="M100 34 C122 56 130 92 130 122 L130 148 L100 148 Z" fill={palette.ink} opacity={0.12} />
      </g>
      <g transform={popAt(window_, C, 92)} opacity={window_}>
        <circle cx={C} cy={92} r={18} fill={palette.primary} />
        <circle cx={C} cy={92} r={18} fill="none" stroke={palette.accent} strokeWidth={4} />
        <circle cx={94} cy={86} r={5} fill="#ffffff" opacity={0.35} />
      </g>
    </g>
  );
};

/** A shield with a check that draws itself, and a guard pulse. */
export const ShieldFigure: Figure = ({build, t, palette}) => {
  const shield = stage(build, 0, 2);
  const check = stage(build, 1, 2);
  // The pulse restarts every few seconds rather than breathing continuously —
  // a shield should read as periodically confirming, not as inhaling. It runs
  // over the first half of the cycle and is simply absent for the rest; an
  // earlier version left the ring drawn at rest opacity between pulses, which
  // put a permanent dirty outline around the shield.
  const beat = (t % 3.2) / 3.2;
  const expand = beat < 0.5 ? beat / 0.5 : -1;

  return (
    <g>
      {expand >= 0 && (
        <circle
          cx={C}
          cy={C}
          r={62 + 26 * expand}
          fill="none"
          stroke={palette.accent}
          strokeWidth={3}
          opacity={0.4 * (1 - expand) * check}
        />
      )}
      <g transform={popAt(shield, C, C)} opacity={shield}>
        <path d="M100 34 L156 56 L156 104 C156 138 130 158 100 168 C70 158 44 138 44 104 L44 56 Z" fill={palette.primary} />
        <path d="M100 34 L156 56 L156 104 C156 138 130 158 100 168 Z" fill={palette.ink} opacity={0.16} />
      </g>
      <path
        d="M74 100 L94 120 L128 78"
        fill="none"
        stroke={palette.accent}
        strokeWidth={11}
        strokeLinecap="round"
        strokeLinejoin="round"
        {...draws(check)}
      />
    </g>
  );
};

/** Bars growing under a trend line that draws across them. */
export const ChartFigure: Figure = ({build, t, palette}) => {
  const axes = stage(build, 0, 3);
  const bars = stage(build, 1, 3);
  const trend = stage(build, 2, 3);
  const heights = [50, 80, 66, 112];
  // A slight continuous jitter on the last bar reads as live data.
  const live = 1 + 0.05 * Math.sin(t * 3.1);

  return (
    <g>
      {/* Axes take `line`, not `ink` — ink is the dark shading fill and a rule
          drawn with it on the stage is simply not there. */}
      <g opacity={axes}>
        <line x1={38} y1={32} x2={38} y2={168} stroke={palette.line} strokeWidth={4} strokeLinecap="round" opacity={0.5} />
        <line x1={38} y1={168} x2={178} y2={168} stroke={palette.line} strokeWidth={4} strokeLinecap="round" opacity={0.5} />
      </g>
      {heights.map((h, i) => {
        const p = ease(stage(bars, i, heights.length, 0.5));
        const last = i === heights.length - 1;
        const full = h * (last ? live : 1);
        return (
          <rect
            key={i}
            x={54 + i * 32}
            y={168 - full * p}
            width={24}
            height={full * p}
            rx={6}
            fill={last ? palette.accent : palette.primary}
            opacity={last ? 1 : 0.75}
          />
        );
      })}
      <path
        d="M66 110 L98 80 L130 94 L162 48"
        fill="none"
        stroke={palette.soft}
        strokeWidth={5}
        strokeLinecap="round"
        strokeLinejoin="round"
        {...draws(trend)}
      />
      <circle cx={162} cy={48} r={7} fill={palette.soft} opacity={trend >= 1 ? 1 : 0} />
    </g>
  );
};

/** The neutral fallback: a burst with a few things in orbit. */
export const SparkFigure: Figure = ({build, t, palette}) => {
  const star = stage(build, 0, 2);
  const orbit = stage(build, 1, 2);
  const breathe = 1 + 0.06 * pulse(t, 2.4);

  return (
    <g>
      <g transform={`${popAt(star, C, C)} rotate(${t * 8} 100 100)`} opacity={star}>
        {/* The four control points must mirror about the centre in both axes.
            Pinning them all to one corner (which the first pass did) makes a
            star that leans, and a lopsided sparkle reads as a mistake. */}
        <path
          d={`M100 ${100 - 62 * breathe} Q106 94 ${100 + 62 * breathe} 100 Q106 106 100 ${100 + 62 * breathe} Q94 106 ${100 - 62 * breathe} 100 Q94 94 100 ${100 - 62 * breathe} Z`}
          fill={palette.accent}
        />
      </g>
      <g opacity={orbit}>
        {ring(3, 78, C, C, t * 0.7).map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r={7 - i} fill={palette.primary} opacity={0.8} />
        ))}
      </g>
    </g>
  );
};
