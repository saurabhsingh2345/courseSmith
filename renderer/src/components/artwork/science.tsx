// Measuring, testing and looking closely.
//
// The shared idea in this set is that an instrument is only an instrument while
// it is reading something. A gauge with a still needle, a balance that never
// tips and a magnet with no field are all diagrams of the object rather than
// pictures of it working — so in every one of these the reading is what moves.

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

/** A nucleus with electrons actually orbiting it. */
export const AtomFigure: Figure = ({build, t, palette}) => {
  const shells = stage(build, 0, 2);
  const core = stage(build, 1, 2);
  const tilts = [0, 60, 120];
  return (
    <g>
      {tilts.map((deg, i) => (
        <g key={i} transform={`rotate(${deg} 100 100)`} opacity={shells}>
          <ellipse cx={C} cy={C} rx={68} ry={26} fill="none" stroke={palette.line} strokeWidth={4} opacity={0.55} />
          {/* The electron rides its own shell, each on a different phase, so
              the three never line up into a single spinning shape. */}
          <circle
            cx={C + 68 * Math.cos(t * 1.5 + i * 2.1)}
            cy={C + 26 * Math.sin(t * 1.5 + i * 2.1)}
            r={8}
            fill={palette.accent}
          />
        </g>
      ))}
      <g transform={popAt(core, C, C)} opacity={core}>
        <circle cx={C} cy={C} r={20} fill={palette.primary} />
        <circle cx={94} cy={94} r={6} fill={palette.soft} opacity={0.35} />
      </g>
    </g>
  );
};

/** A flask with a level that rises and bubbles that leave it. */
export const FlaskFigure: Figure = ({build, t, palette}) => {
  const glass = stage(build, 0, 2);
  const liquid = stage(build, 1, 2);
  const level = 0.62 + 0.06 * Math.sin(t * 1.3);
  const top = 156 - 62 * level;
  return (
    <g>
      <g transform={popAt(glass, C, 110)} opacity={glass}>
        <path d="M86 34 L114 34 L114 82 L152 148 A10 10 0 0 1 143 164 L57 164 A10 10 0 0 1 48 148 L86 82 Z" fill={palette.soft} opacity={0.55} />
        <path d="M86 34 L114 34 L114 82 L152 148 A10 10 0 0 1 143 164 L100 164 Z" fill={palette.ink} opacity={0.07} />
        <rect x={80} y={28} width={40} height={9} rx={4.5} fill={palette.line} opacity={0.6} />
      </g>
      <g opacity={liquid}>
        <path
          d={`M${58 + (top - 100) * 0.0} ${top} L142 ${top} L152 148 A10 10 0 0 1 143 164 L57 164 A10 10 0 0 1 48 148 Z`}
          fill={palette.accent}
          opacity={0.85}
        />
        {[0, 1, 2].map((i) => {
          const up = cycle(t, 0.7, i * 0.33);
          return (
            <circle
              key={i}
              cx={84 + i * 16}
              cy={160 - up * (160 - top - 6)}
              r={4 + i}
              fill={palette.soft}
              opacity={fadeTravel(up) * 0.8}
            />
          );
        })}
      </g>
    </g>
  );
};

/** A microscope with a slide under it and a focus that hunts. */
export const MicroscopeFigure: Figure = ({build, t, palette}) => {
  const stand = stage(build, 0, 3);
  const tube = stage(build, 1, 3);
  const slide = stage(build, 2, 3);
  const focus = bob(t, 4, 0.9);
  return (
    <g>
      <g opacity={stand}>
        <path d="M52 164 L148 164 A8 8 0 0 0 156 156 L156 150 L44 150 L44 156 A8 8 0 0 0 52 164 Z" fill={palette.line} opacity={0.65} />
        <path d="M96 150 C96 120 74 116 74 92" fill="none" stroke={palette.primary} strokeWidth={9} strokeLinecap="round" />
      </g>
      <g opacity={slide}>
        <rect x={78} y={122} width={72} height={9} rx={4} fill={palette.soft} />
        <circle cx={114} cy={126} r={5} fill={palette.accent} opacity={pulse(t, 2.6)} />
      </g>
      <g opacity={tube} transform={`translate(0 ${focus})`}>
        <rect x={92} y={34} width={30} height={64} rx={12} fill={palette.primary} transform="rotate(18 107 66)" />
        <rect x={98} y={94} width={22} height={22} rx={6} fill={palette.soft} />
        <circle cx={109} cy={40} r={13} fill={palette.soft} transform="rotate(18 107 66)" />
      </g>
    </g>
  );
};

/** A telescope that sweeps the sky, with a star it finds. */
export const TelescopeFigure: Figure = ({build, t, palette}) => {
  const tripod = stage(build, 0, 3);
  const tube = stage(build, 1, 3);
  const sky = stage(build, 2, 3);
  const sweep = Math.sin(t * 0.55) * 14;
  return (
    <g>
      <g opacity={sky}>
        {ring(4, 58, 132, 52, 0.4).map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r={3 + (i % 2)} fill={palette.accent} opacity={0.3 + 0.7 * pulse(t, 2 + i, i)} />
        ))}
      </g>
      <g opacity={tripod}>
        <path d="M100 122 L74 168 M100 122 L126 168 M100 122 L100 168" stroke={palette.line} strokeWidth={7} strokeLinecap="round" fill="none" opacity={0.65} />
      </g>
      <g opacity={tube} transform={`rotate(${-34 + sweep} 100 122)`}>
        <rect x={82} y={44} width={36} height={84} rx={14} fill={palette.primary} />
        <rect x={82} y={44} width={36} height={84} rx={14} fill={palette.ink} opacity={0.1} />
        <ellipse cx={C} cy={44} rx={20} ry={8} fill={palette.soft} />
        <rect x={112} y={92} width={26} height={12} rx={6} fill={palette.soft} opacity={0.7} />
      </g>
    </g>
  );
};

/** A double helix that turns on its axis. */
export const DnaFigure: Figure = ({build, t, palette}) => {
  const strands = stage(build, 0, 2);
  const rungs = stage(build, 1, 2);
  const rows = Array.from({length: 9}, (_, i) => 34 + i * 16.5);
  const phase = t * 1.5;
  const xAt = (y: number, s: number) => C + Math.sin(y * 0.055 + phase) * 42 * s;
  return (
    <g>
      <g opacity={rungs}>
        {rows.map((y, i) => {
          const a = xAt(y, 1);
          const b = xAt(y, -1);
          // The rung fades as the helix turns edge-on, which is what gives the
          // flat drawing its depth.
          const face = Math.abs(Math.cos(y * 0.055 + phase));
          return (
            <line
              key={i}
              x1={a}
              y1={y}
              x2={b}
              y2={y}
              stroke={i % 2 === 0 ? palette.accent : palette.primary}
              strokeWidth={5}
              strokeLinecap="round"
              opacity={0.25 + 0.6 * face}
            />
          );
        })}
      </g>
      <g opacity={strands} fill="none" strokeWidth={7} strokeLinecap="round">
        <path d={rows.map((y, i) => `${i === 0 ? 'M' : 'L'}${xAt(y, 1)} ${y}`).join(' ')} stroke={palette.soft} />
        <path d={rows.map((y, i) => `${i === 0 ? 'M' : 'L'}${xAt(y, -1)} ${y}`).join(' ')} stroke={palette.line} />
      </g>
    </g>
  );
};

/** A horseshoe magnet with a field that snaps to it. */
export const MagnetFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 2);
  const field = stage(build, 1, 2);
  return (
    <g>
      <g opacity={field}>
        {[0, 1, 2].map((i) => {
          const p = cycle(t, 0.8, i * 0.33);
          return (
            <path
              key={i}
              d={`M68 ${150 - p * 26} A${34} ${18 + p * 14} 0 0 1 132 ${150 - p * 26}`}
              fill="none"
              stroke={palette.accent}
              strokeWidth={4}
              strokeLinecap="round"
              opacity={fadeTravel(p) * 0.8}
            />
          );
        })}
      </g>
      <g transform={popAt(body, C, C)} opacity={body}>
        <path
          d="M50 148 L50 90 A50 50 0 0 1 150 90 L150 148 L118 148 L118 90 A18 18 0 0 0 82 90 L82 148 Z"
          fill={palette.primary}
        />
        <rect x={50} y={130} width={32} height={18} fill={palette.accent} />
        <rect x={118} y={130} width={32} height={18} fill={palette.soft} />
      </g>
    </g>
  );
};

/** A cell charging up and discharging. */
export const BatteryFigure: Figure = ({build, t, palette}) => {
  const shell = stage(build, 0, 2);
  const charge = stage(build, 1, 2);
  const level = cycle(t, 0.28);
  const bars = Math.floor(level * 4) + 1;
  return (
    <g transform={`translate(0 ${bob(t, 2, 1.5)})`}>
      <g transform={popAt(shell, C, C)} opacity={shell}>
        <rect x={34} y={66} width={124} height={68} rx={12} fill={palette.soft} />
        <rect x={34} y={66} width={124} height={68} rx={12} fill="none" stroke={palette.primary} strokeWidth={7} />
        <rect x={162} y={86} width={12} height={28} rx={5} fill={palette.primary} />
      </g>
      <g opacity={charge}>
        {[0, 1, 2, 3].map((i) => (
          <rect
            key={i}
            x={48 + i * 26}
            y={80}
            width={20}
            height={40}
            rx={4}
            fill={i < bars ? palette.accent : palette.ink}
            opacity={i < bars ? 1 : 0.12}
          />
        ))}
      </g>
    </g>
  );
};

/** A prism splitting a beam into a fan. */
export const PrismFigure: Figure = ({build, t, palette}) => {
  const glass = stage(build, 0, 3);
  const beam = stage(build, 1, 3);
  const fan = stage(build, 2, 3);
  const spread = 1 + 0.14 * Math.sin(t * 1.2);
  return (
    <g>
      <line x1={14} y1={98} x2={72} y2={106} stroke={palette.soft} strokeWidth={5} strokeLinecap="round" opacity={beam} {...draws(beam)} />
      <g opacity={fan}>
        {[-1, 0, 1, 2].map((k, i) => (
          <line
            key={i}
            x1={128}
            y1={110}
            x2={188}
            y2={110 + k * 22 * spread}
            stroke={i % 2 === 0 ? palette.accent : palette.primary}
            strokeWidth={5}
            strokeLinecap="round"
            opacity={0.85}
          />
        ))}
      </g>
      <g transform={popAt(glass, C, 110)} opacity={glass}>
        <path d="M100 46 L152 146 L48 146 Z" fill={palette.soft} opacity={0.65} />
        <path d="M100 46 L152 146 L100 146 Z" fill={palette.ink} opacity={0.08} />
        <path d="M100 46 L152 146 L48 146 Z" fill="none" stroke={palette.line} strokeWidth={4} strokeLinejoin="round" opacity={0.7} />
      </g>
    </g>
  );
};

/** A balance that tips, settles, and tips back. */
export const BalanceFigure: Figure = ({build, t, palette}) => {
  const stand = stage(build, 0, 2);
  const beam = stage(build, 1, 2);
  const tip = Math.sin(t * 0.85) * 12;
  const pan = (side: -1 | 1) => {
    const x = C + side * 56;
    const y = 92 + side * (tip * 1.4);
    return (
      <g key={side}>
        <line x1={x} y1={y} x2={x} y2={y + 26} stroke={palette.line} strokeWidth={3} opacity={0.6} />
        <path d={`M${x - 24} ${y + 26} L${x + 24} ${y + 26} L${x + 16} ${y + 42} L${x - 16} ${y + 42} Z`} fill={side < 0 ? palette.primary : palette.accent} />
      </g>
    );
  };
  return (
    <g>
      <g opacity={stand}>
        <rect x={94} y={80} width={12} height={80} rx={5} fill={palette.line} opacity={0.65} />
        <rect x={62} y={158} width={76} height={10} rx={5} fill={palette.line} opacity={0.65} />
      </g>
      <g opacity={beam}>
        <g transform={`rotate(${tip} 100 88)`}>
          <rect x={38} y={84} width={124} height={9} rx={4.5} fill={palette.soft} />
        </g>
        {pan(-1)}
        {pan(1)}
        <circle cx={C} cy={88} r={10} fill={palette.primary} />
      </g>
    </g>
  );
};

/** A compass whose needle hunts and settles on north. */
export const CompassFigure: Figure = ({build, t, palette}) => {
  const face = stage(build, 0, 2);
  const needle = stage(build, 1, 2);
  // Overshoot and damp, then hold — a needle that swings forever never found
  // anything.
  const p = (t % 5) / 5;
  const hunt = p < 0.5 ? Math.sin(p * 22) * 34 * (1 - p * 2) : 0;
  return (
    <g>
      <g transform={popAt(face, C, C)} opacity={face}>
        <circle cx={C} cy={C} r={70} fill={palette.soft} />
        <circle cx={C} cy={C} r={70} fill="none" stroke={palette.primary} strokeWidth={8} />
        {ring(8, 58).map((p2, i) => (
          <circle key={i} cx={p2.x} cy={p2.y} r={i % 2 === 0 ? 4 : 2.5} fill={palette.ink} opacity={0.35} />
        ))}
      </g>
      <g opacity={needle} transform={`rotate(${hunt} 100 100)`}>
        <path d="M100 44 L114 100 L100 116 L86 100 Z" fill={palette.accent} />
        <path d="M100 156 L86 100 L100 84 L114 100 Z" fill={palette.line} opacity={0.75} />
        <circle cx={C} cy={C} r={8} fill={palette.primary} />
      </g>
    </g>
  );
};

/** A rule with a measured span opening across it. */
export const RulerFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 2);
  const marks = stage(build, 1, 2);
  const span = 0.4 + 0.35 * pulse(t, 0.9);
  return (
    <g transform="rotate(-14 100 100)">
      <g transform={popAt(body, C, C)} opacity={body}>
        <rect x={16} y={80} width={168} height={40} rx={8} fill={palette.soft} />
        <rect x={16} y={80} width={168} height={40} rx={8} fill={palette.ink} opacity={0.06} />
      </g>
      <g opacity={marks}>
        {Array.from({length: 13}, (_, i) => (
          <line
            key={i}
            x1={28 + i * 12.8}
            y1={80}
            x2={28 + i * 12.8}
            y2={80 + (i % 4 === 0 ? 22 : 12)}
            stroke={palette.ink}
            strokeWidth={3}
            opacity={0.35}
          />
        ))}
        {/* The measurement being taken. */}
        <line x1={28} y1={134} x2={28 + 152 * span} y2={134} stroke={palette.accent} strokeWidth={5} strokeLinecap="round" />
        <circle cx={28 + 152 * span} cy={134} r={7} fill={palette.accent} />
      </g>
    </g>
  );
};

/** A calculator with keys that press and a total that lands. */
export const CalculatorFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 3);
  const screen = stage(build, 1, 3);
  const keys = stage(build, 2, 3);
  // One key lights at a time, walking the pad — a calculator with every key
  // lit is a keyboard.
  const hit = Math.floor(t * 3.2) % 9;
  return (
    <g>
      <g transform={popAt(body, C, C)} opacity={body}>
        <rect x={48} y={26} width={104} height={148} rx={14} fill={palette.soft} />
      </g>
      <g opacity={screen}>
        <rect x={60} y={38} width={80} height={32} rx={6} fill={palette.primary} />
        <rect x={92} y={48} width={40} height={11} rx={3} fill={palette.accent} opacity={0.85} />
      </g>
      <g opacity={keys}>
        {Array.from({length: 9}, (_, i) => {
          const cx = 72 + (i % 3) * 28;
          const cy = 92 + Math.floor(i / 3) * 28;
          const on = i === hit;
          return <rect key={i} x={cx - 11} y={cy - 11} width={22} height={22} rx={6} fill={on ? palette.accent : palette.ink} opacity={on ? 1 : 0.16} />;
        })}
      </g>
    </g>
  );
};

/** A satellite with panels that track and a signal it drops. */
export const SatelliteFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 3);
  const panels = stage(build, 1, 3);
  const signal = stage(build, 2, 3);
  const tilt = bob(t, 8, 0.7);
  const beam = pulse(t, 2.4);
  return (
    <g transform={`translate(0 ${bob(t, 3, 1.1)})`}>
      <g opacity={panels} transform={`rotate(${tilt} 100 84)`}>
        <rect x={14} y={68} width={54} height={32} rx={5} fill={palette.primary} />
        <rect x={132} y={68} width={54} height={32} rx={5} fill={palette.primary} />
        {[0, 1, 2].map((i) => (
          <g key={i}>
            <line x1={14 + (i + 1) * 13.5} y1={68} x2={14 + (i + 1) * 13.5} y2={100} stroke={palette.ink} strokeWidth={2} opacity={0.25} />
            <line x1={132 + (i + 1) * 13.5} y1={68} x2={132 + (i + 1) * 13.5} y2={100} stroke={palette.ink} strokeWidth={2} opacity={0.25} />
          </g>
        ))}
      </g>
      <g transform={popAt(body, C, 88)} opacity={body}>
        <rect x={76} y={62} width={48} height={54} rx={9} fill={palette.soft} />
        <path d="M84 62 L100 40 L116 62 Z" fill={palette.line} opacity={0.7} />
      </g>
      <g opacity={signal}>
        {[0, 1, 2].map((i) => (
          <path
            key={i}
            d={`M${86 - i * 12} ${128 + i * 12} A${20 + i * 16} ${20 + i * 16} 0 0 0 ${114 + i * 12} ${128 + i * 12}`}
            fill="none"
            stroke={palette.accent}
            strokeWidth={4}
            strokeLinecap="round"
            opacity={Math.max(0, beam - i * 0.25)}
          />
        ))}
      </g>
    </g>
  );
};

/** A dial whose needle sweeps its arc and comes back. */
export const GaugeFigure: Figure = ({build, t, palette}) => {
  const dial = stage(build, 0, 3);
  const band = stage(build, 1, 3);
  const needle = stage(build, 2, 3);
  const v = 0.5 + 0.42 * Math.sin(t * 0.8);
  const angle = -120 + v * 240;
  return (
    <g>
      <g transform={popAt(dial, C, 116)} opacity={dial}>
        <path d="M28 116 A72 72 0 0 1 172 116" fill="none" stroke={palette.soft} strokeWidth={22} strokeLinecap="round" />
      </g>
      <g opacity={band}>
        {/* The hot end of the scale, so a high reading means something. */}
        <path d="M126 56 A72 72 0 0 1 172 116" fill="none" stroke={palette.accent} strokeWidth={22} strokeLinecap="round" opacity={0.85} />
      </g>
      <g opacity={needle}>
        <g transform={`rotate(${angle} 100 116)`}>
          <path d="M100 52 L106 116 L94 116 Z" fill={palette.primary} />
        </g>
        <circle cx={C} cy={116} r={12} fill={palette.primary} />
        <circle cx={C} cy={116} r={5} fill={palette.soft} />
      </g>
    </g>
  );
};

/** Slabs settling into a stack. */
export const StackFigure: Figure = ({build, t, palette}) => {
  const slabs = [0, 1, 2];
  return (
    <g>
      {slabs.map((i) => {
        // Bottom slab first: a stack that assembles top-down looks like it is
        // falling apart played backwards.
        const p = ease(stage(build, slabs.length - 1 - i, slabs.length, 0.45));
        const y = 132 - i * 34;
        // Each slab drifts on its own phase, so the stack breathes instead of
        // sliding as one rigid block.
        const float = bob(t, 2.4, 1.6, i * 0.9);
        const top = i === slabs.length - 1;
        const face = top ? palette.accent : palette.primary;
        return (
          <g key={i} opacity={p} transform={`translate(0 ${(1 - p) * -46 + float})`}>
            {/* An isometric slab: top face plus two side faces for thickness.
                The sides are the slab's own colour held down rather than a wash
                of ink — ink is darker than the stage, so painted over open
                background it removed the thickness instead of shading it. */}
            <path d={`M100 ${y - 22} L156 ${y} L100 ${y + 22} L44 ${y} Z`} fill={face} opacity={top ? 1 : 0.85 - i * 0.06} />
            <path d={`M44 ${y} L100 ${y + 22} L100 ${y + 32} L44 ${y + 10} Z`} fill={face} opacity={0.42} />
            <path d={`M156 ${y} L100 ${y + 22} L100 ${y + 32} L156 ${y + 10} Z`} fill={face} opacity={0.6} />
          </g>
        );
      })}
    </g>
  );
};

/** A clock whose hands actually move. */
export const ClockFigure: Figure = ({build, t, palette}) => {
  const face = stage(build, 0, 3);
  const ticks = stage(build, 1, 3);
  const hands = stage(build, 2, 3);
  // A whole revolution in twelve seconds: fast enough to read as motion in a
  // short clip, slow enough not to read as a stopwatch gone wrong.
  const second = t * 30;
  return (
    <g>
      <g transform={popAt(face, C, C)} opacity={face}>
        <circle cx={C} cy={C} r={68} fill={palette.soft} />
        <circle cx={C} cy={C} r={68} fill="none" stroke={palette.primary} strokeWidth={9} />
      </g>
      <g opacity={ticks}>
        {ring(12, 54).map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r={i % 3 === 0 ? 4.5 : 2.5} fill={palette.ink} opacity={0.55} />
        ))}
      </g>
      <g opacity={hands}>
        <line x1={C} y1={C} x2={C} y2={64} stroke={palette.ink} strokeWidth={8} strokeLinecap="round" transform={`rotate(${second / 12} 100 100)`} opacity={0.75} />
        <line x1={C} y1={C} x2={C} y2={50} stroke={palette.accent} strokeWidth={6} strokeLinecap="round" transform={`rotate(${second} 100 100)`} />
        <circle cx={C} cy={C} r={7} fill={palette.accent} />
      </g>
    </g>
  );
};

/** A sand timer that runs, and flips when it empties. */
export const HourglassFigure: Figure = ({build, t, palette}) => {
  const shell = stage(build, 0, 2);
  const sand = stage(build, 1, 2);
  const period = 5;
  const run = (t % period) / period;
  // The flip happens in the last tenth, so the glass is running for nine tenths
  // of the time rather than tumbling.
  const flip = run > 0.9 ? ease((run - 0.9) / 0.1) * 180 : 0;
  const drained = Math.min(1, run / 0.9);
  return (
    <g transform={`rotate(${flip} 100 100)`}>
      <g transform={popAt(shell, C, C)} opacity={shell}>
        <rect x={54} y={30} width={92} height={11} rx={5} fill={palette.primary} />
        <rect x={54} y={159} width={92} height={11} rx={5} fill={palette.primary} />
        <path d="M66 41 L134 41 L106 100 L134 159 L66 159 L94 100 Z" fill={palette.soft} opacity={0.55} />
        <path d="M66 41 L134 41 L106 100 L134 159 L66 159 L94 100 Z" fill="none" stroke={palette.line} strokeWidth={4} strokeLinejoin="round" opacity={0.7} />
      </g>
      <g opacity={sand}>
        <path d={`M${72 + drained * 22} ${47 + drained * 40} L${128 - drained * 22} ${47 + drained * 40} L100 100 Z`} fill={palette.accent} />
        <path d={`M100 ${153 - (1 - drained) * 34} L${128 - (1 - drained) * 20} 153 L${72 + (1 - drained) * 20} 153 Z`} fill={palette.accent} />
        <line x1={C} y1={100} x2={C} y2={148} stroke={palette.accent} strokeWidth={3} opacity={drained < 1 ? 0.9 : 0} />
      </g>
    </g>
  );
};
