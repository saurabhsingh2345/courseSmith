// The situations a narrator points at, rather than the things they name.
//
// Everything else in the drawer is a noun. A narrator explaining something is
// mostly not naming objects — they are saying "here is the catch", "this is
// where people get stuck", "that one knocks over the next one", "you are here".
// Those are the sentences the connective tissue of a course is made of, and
// before this module the only figures for them were a `lightbulb` for every
// realisation and a `warning`-shaped hole.
//
// The distinction that keeps this module from becoming a second icon set: these
// draw a *relation* or a *state*, not a thing. `domino` is causality, `hurdle`
// is difficulty overcome, `handoff` is a transfer, `blocked` is a stop. If a
// figure here could be described as "a picture of an X", it belongs in one of
// the other modules.
//
// Same rule as the rest of the drawer: the situation is what moves.

import {
  C,
  bob,
  cycle,
  draws,
  ease,
  fadeTravel,
  gesture,
  popAt,
  pulse,
  ring,
  stage,
  swing,
  type Figure,
} from './kit';

/** The catch: a triangle that flashes on a beat rather than glowing. */
export const WarningFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 2);
  const mark = stage(build, 1, 2);
  // Two quick flashes then a long rest — the rhythm of an actual alert. A
  // warning that pulses like a heartbeat reads as ambient, not as a warning.
  const beat = (t % 3.4) / 3.4;
  const hot = beat < 0.08 || (beat > 0.16 && beat < 0.24) ? 1 : 0;
  return (
    <g>
      <g transform={popAt(body, C, 118)} opacity={body}>
        <path d="M100 44 L172 158 H28 Z" fill={palette.accent} />
        <path d="M100 44 L172 158 H28 Z" fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.18} />
        {hot > 0 && (
          <path d="M100 44 L172 158 H28 Z" fill={palette.soft} opacity={0.28} />
        )}
      </g>
      <g opacity={mark}>
        <rect x={93} y={82} width={14} height={44} rx={7} fill={palette.ink} opacity={0.75} />
        <circle cx={C} cy={140} r={8} fill={palette.ink} opacity={0.75} />
      </g>
    </g>
  );
};

/** Where people get stuck: a traveller that reaches a barrier and stops. */
export const BlockedFigure: Figure = ({build, t, palette}) => {
  const road = stage(build, 0, 3);
  const bar = stage(build, 1, 3);
  const mover = stage(build, 2, 3);
  // It approaches, hits, and bounces back — then tries again. The retry is the
  // difference between "there is a wall" and "you cannot get past this".
  const run = cycle(t, 0.45);
  const reach = run < 0.62 ? ease(run / 0.62) : 1 - ease((run - 0.62) / 0.38);
  const x = 34 + reach * 58;
  const hit = run > 0.58 && run < 0.68;
  return (
    <g>
      <line x1={26} y1={C} x2={174} y2={C} stroke={palette.line} strokeWidth={5} strokeLinecap="round" strokeDasharray="10 9" opacity={0.35 * road} />
      {/* A full-height barrier with a footing. At sixteen wide it read as a
          post somebody could walk round, which is not what "blocked" means. */}
      <g opacity={bar}>
        <rect x={100} y={44} width={26} height={112} rx={9} fill={palette.primary} />
        {[54, 82, 110, 138].map((y) => (
          <rect key={y} x={96} y={y} width={34} height={11} rx={5} fill={palette.accent} opacity={0.85} />
        ))}
        <rect x={86} y={152} width={54} height={10} rx={5} fill={palette.ink} opacity={0.4} />
      </g>
      <g opacity={mover}>
        <circle cx={x} cy={C} r={17} fill={palette.soft} />
        <circle cx={x} cy={C} r={17} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.18} />
      </g>
      {hit && (
        <g opacity={0.8}>
          {ring(4, 1, 100, C, -0.4).map((_, i) => (
            <line key={i} x1={98} y1={C - 18 + i * 12} x2={88} y2={C - 24 + i * 14} stroke={palette.accent} strokeWidth={3} strokeLinecap="round" />
          ))}
        </g>
      )}
    </g>
  );
};

/** The win: a cone that throws confetti once and then reloads. */
export const CelebrateFigure: Figure = ({build, t, palette}) => {
  const cone = stage(build, 0, 2);
  const burst = stage(build, 1, 2);
  const pop = gesture(t, 3.2, 0.62);
  const bits = [
    {a: -1.9, d: 78, c: 0},
    {a: -1.45, d: 96, c: 1},
    {a: -1.05, d: 82, c: 2},
    {a: -0.66, d: 66, c: 0},
    {a: -2.3, d: 60, c: 1},
    {a: -1.2, d: 108, c: 2},
  ];
  const colours = [palette.accent, palette.primary, palette.soft];
  return (
    <g>
      <g transform={popAt(cone, 62, 148)} opacity={cone}>
        <path d="M40 168 L52 128 L84 142 Z" fill={palette.primary} />
        <path d="M40 168 L52 128 L84 142 Z" fill={palette.ink} opacity={0.14} />
      </g>
      <g opacity={burst}>
        {bits.map((b, i) => {
          const p = pop >= 0 ? Math.max(0, Math.min(1, (pop - i * 0.05) / 0.72)) : 0;
          if (p <= 0.01) return null;
          // Gravity: the arc rises and falls rather than radiating evenly,
          // which is the difference between confetti and a starburst.
          const x = 68 + Math.cos(b.a) * b.d * ease(p);
          const y = 138 + Math.sin(b.a) * b.d * ease(p) + 46 * p * p;
          return (
            <rect
              key={i}
              x={x - 5}
              y={y - 5}
              width={10}
              height={10}
              rx={2}
              fill={colours[b.c]}
              opacity={1 - p * 0.85}
              transform={`rotate(${p * 320} ${x} ${y})`}
            />
          );
        })}
      </g>
    </g>
  );
};

/** You are here: a post on a road with the marker dropping onto it. */
export const MilestoneFigure: Figure = ({build, t, palette}) => {
  const road = stage(build, 0, 3);
  const post = stage(build, 1, 3);
  const pinP = stage(build, 2, 3);
  const drop = swing(t, 4.6, 0.18, 0.52);
  const y = 44 - (1 - drop) * 40;
  return (
    <g>
      <g opacity={road}>
        <path d="M28 152 Q100 128 172 152" fill="none" stroke={palette.line} strokeWidth={6} strokeLinecap="round" opacity={0.45} {...draws(road)} />
      </g>
      <g opacity={post}>
        <rect x={95} y={80} width={10} height={72} rx={5} fill={palette.primary} />
        <rect x={64} y={94} width={62} height={9} rx={4.5} fill={palette.primary} opacity={0.55} />
      </g>
      <g opacity={pinP} transform={`translate(0 ${y})`}>
        <path d="M100 20 C118 20 128 33 128 46 C128 62 100 84 100 84 C100 84 72 62 72 46 C72 33 82 20 100 20 Z" fill={palette.accent} />
        <circle cx={C} cy={46} r={11} fill={palette.ink} opacity={0.35} />
      </g>
      {/* The landing ring, only in the frames it seats. */}
      {drop > 0.97 && (
        <ellipse cx={C} cy={128} rx={14 + 30 * (drop - 0.97) * 33} ry={5 + 10 * (drop - 0.97) * 33} fill="none" stroke={palette.accent} strokeWidth={2} opacity={(1 - drop) * 33} />
      )}
    </g>
  );
};

/** The whole route at once: a folded sheet with a path drawn across it. */
export const MapFigure: Figure = ({build, t, palette}) => {
  const sheet = stage(build, 0, 3);
  const path = stage(build, 1, 3);
  const stops = stage(build, 2, 3);
  const walk = cycle(t, 0.3);
  const d = 'M52 138 C70 104 92 128 106 96 C118 68 136 76 150 58';
  const marks = [
    {x: 52, y: 138},
    {x: 106, y: 96},
    {x: 150, y: 58},
  ];
  return (
    <g>
      <g transform={popAt(sheet, C, C)} opacity={sheet}>
        <path d="M32 56 L76 44 L124 58 L168 44 V150 L124 164 L76 150 L32 162 Z" fill={palette.soft} />
        <path d="M32 56 L76 44 L124 58 L168 44 V150 L124 164 L76 150 L32 162 Z" fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.16} />
        <line x1={76} y1={44} x2={76} y2={150} stroke={palette.ink} strokeWidth={2} opacity={0.12} />
        <line x1={124} y1={58} x2={124} y2={164} stroke={palette.ink} strokeWidth={2} opacity={0.12} />
      </g>
      <g opacity={path}>
        {/* Dashed rather than drawn-on: `draws` owns strokeDasharray, and a
            route that is a solid line is a wire. The reveal is the traveller. */}
        <path d={d} fill="none" stroke={palette.primary} strokeWidth={5} strokeLinecap="round" strokeDasharray="9 8" opacity={ease(path)} />
        {marks.map((m, i) => (
          <circle key={i} cx={m.x} cy={m.y} r={i === 1 ? 7 : 5} fill={i === 1 ? palette.accent : palette.primary} opacity={stops} />
        ))}
      </g>
      {/* A traveller running the route, so the map is a journey not a diagram. */}
      <circle
        cx={52 + (150 - 52) * walk}
        cy={138 - (138 - 58) * walk + Math.sin(walk * Math.PI * 2) * 9}
        r={5}
        fill={palette.accent}
        opacity={fadeTravel(walk) * stops}
      />
    </g>
  );
};

/** The way around: a path that forks past a closure and rejoins. */
export const DetourFigure: Figure = ({build, t, palette}) => {
  const direct = stage(build, 0, 3);
  const around = stage(build, 1, 3);
  const closed = stage(build, 2, 3);
  const at = cycle(t, 0.36);
  // The traveller only ever takes the long way, which is the point.
  const x = 30 + at * 140;
  const y = 118 - Math.sin(at * Math.PI) * 56;
  return (
    <g>
      <g opacity={direct}>
        <line x1={30} y1={118} x2={170} y2={118} stroke={palette.line} strokeWidth={5} strokeLinecap="round" opacity={0.28} />
      </g>
      <g opacity={around}>
        <path d="M30 118 Q100 46 170 118" fill="none" stroke={palette.primary} strokeWidth={5} strokeLinecap="round" {...draws(around)} />
      </g>
      <g opacity={closed}>
        <rect x={78} y={110} width={44} height={9} rx={4.5} fill={palette.accent} transform="rotate(-14 100 114)" />
        <rect x={78} y={110} width={44} height={9} rx={4.5} fill={palette.accent} transform="rotate(14 100 114)" />
      </g>
      <circle cx={x} cy={y} r={7} fill={palette.accent} opacity={fadeTravel(at) * around} />
    </g>
  );
};

/** What it costs, or what it earns: a coin turning on its edge. */
export const CoinFigure: Figure = ({build, t, palette}) => {
  const disc = stage(build, 0, 2);
  const face = stage(build, 1, 2);
  // A full spin means the coin passes edge-on, where its width goes to zero —
  // which is what makes it read as a coin rather than a circle.
  const spin = Math.cos(t * 1.9);
  const w = Math.abs(spin);
  const edge = w < 0.16;
  const lift = bob(t, 4, 1.9);
  return (
    <g transform={`translate(0 ${lift})`}>
      <ellipse cx={C} cy={C} rx={Math.max(5, 58 * w)} ry={58} fill={palette.accent} opacity={disc} />
      <ellipse cx={C} cy={C} rx={Math.max(5, 58 * w)} ry={58} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.2 * disc} />
      {!edge && (
        <g opacity={face * (w - 0.16) * 1.6}>
          <ellipse cx={C} cy={C} rx={44 * w} ry={44} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.28} />
          <rect x={C - 5 * w} y={72} width={Math.max(2, 10 * w)} height={56} rx={4} fill={palette.ink} opacity={0.4} />
        </g>
      )}
      {/* The rim catches the light as it passes edge-on. */}
      {edge && <rect x={C - 4} y={44} width={8} height={112} rx={4} fill={palette.primary} opacity={disc} />}
    </g>
  );
};

/** Settled: a stamp that comes down, marks, and lifts. */
export const StampFigure: Figure = ({build, t, palette}) => {
  const paper = stage(build, 0, 3);
  const head = stage(build, 1, 3);
  const inkP = stage(build, 2, 3);
  const press = swing(t, 4.2, 0.12, 0.14);
  const y = -46 + press * 46;
  // The mark survives the lift — the whole reason to stamp something.
  const marked = (t % 4.2) / 4.2 > 0.29;
  return (
    <g>
      <g transform={popAt(paper, C, 148)} opacity={paper}>
        <rect x={40} y={122} width={120} height={44} rx={8} fill={palette.soft} />
        <rect x={40} y={122} width={120} height={44} rx={8} fill="none" stroke={palette.ink} strokeWidth={2} opacity={0.16} />
      </g>
      <g opacity={inkP * (marked ? 1 : 0)}>
        <circle cx={C} cy={144} r={26} fill="none" stroke={palette.accent} strokeWidth={5} opacity={0.85} />
        <path d="M88 144 l9 10 17 -20" fill="none" stroke={palette.accent} strokeWidth={5} strokeLinecap="round" strokeLinejoin="round" />
      </g>
      <g opacity={head} transform={`translate(0 ${y})`}>
        <rect x={68} y={92} width={64} height={22} rx={6} fill={palette.primary} />
        <rect x={86} y={58} width={28} height={38} rx={9} fill={palette.primary} />
        <rect x={76} y={44} width={48} height={18} rx={9} fill={palette.ink} opacity={0.4} />
      </g>
    </g>
  );
};

/** The one that matters: a cone of light that picks one thing out of a row. */
export const SpotlightFigure: Figure = ({build, t, palette}) => {
  const lamp = stage(build, 0, 3);
  const beam = stage(build, 1, 3);
  const items = stage(build, 2, 3);
  const xs = [56, 100, 144];
  // The beam swings between the three and settles on each in turn, so the
  // figure is about *selection* rather than about a light.
  const at = Math.floor(cycle(t, 0.28) * 3) % 3;
  const target = xs[at];
  const lean = (target - C) * 0.16;
  return (
    <g>
      <g opacity={beam}>
        <path
          d={`M${C - 12} 62 L${target - 30} 142 L${target + 30} 142 L${C + 12} 62 Z`}
          fill={palette.accent}
          opacity={0.2}
        />
      </g>
      <g transform={popAt(lamp, C, 52) + ` rotate(${lean} 100 44)`} opacity={lamp}>
        <rect x={78} y={30} width={44} height={30} rx={8} fill={palette.primary} />
        <rect x={72} y={56} width={56} height={10} rx={5} fill={palette.accent} />
        <rect x={95} y={16} width={10} height={16} rx={5} fill={palette.ink} opacity={0.35} />
      </g>
      <g opacity={items}>
        {xs.map((x, i) => {
          const on = i === at;
          return (
            <rect
              key={x}
              x={x - 17}
              y={on ? 132 : 138}
              width={34}
              height={on ? 34 : 28}
              rx={7}
              fill={on ? palette.accent : palette.soft}
              opacity={on ? 1 : 0.5}
            />
          );
        })}
      </g>
    </g>
  );
};

/** What holds when everything else moves: an anchor swinging on its chain. */
export const AnchorFigure: Figure = ({build, t, palette}) => {
  const chain = stage(build, 0, 2);
  const body = stage(build, 1, 2);
  // It swings and damps rather than swinging forever — an anchor that never
  // settles is a pendulum.
  const swingA = Math.sin(t * 1.4) * 6 * (0.55 + 0.45 * pulse(t, 0.35));
  return (
    <g transform={`rotate(${swingA} 100 44)`}>
      <g opacity={chain}>
        {[34, 48].map((y, i) => (
          <circle key={i} cx={C} cy={y} r={9} fill="none" stroke={palette.line} strokeWidth={4} opacity={0.6} />
        ))}
      </g>
      <g opacity={body}>
        <rect x={94} y={58} width={12} height={92} rx={6} fill={palette.primary} />
        <rect x={62} y={68} width={76} height={12} rx={6} fill={palette.primary} />
        <path
          d="M50 116 C50 152 74 166 100 166 C126 166 150 152 150 116"
          fill="none"
          stroke={palette.accent}
          strokeWidth={12}
          strokeLinecap="round"
          {...draws(body)}
        />
      </g>
    </g>
  );
};

/** Complexity: a tangle that pulls tight and never comes loose. */
export const KnotFigure: Figure = ({build, t, palette}) => {
  const rope = stage(build, 0, 2);
  const tail = stage(build, 1, 2);
  // Tightening rather than untying. The narration this is for is "this is the
  // knot", not "here is how to undo it".
  const tight = 1 - 0.16 * pulse(t, 0.8);
  return (
    <g transform={`translate(${C} ${C}) scale(${tight}) translate(${-C} ${-C})`}>
      <g opacity={rope}>
        <path
          d="M60 78 C104 44 140 88 108 112 C80 132 62 108 88 92 C120 72 148 116 122 140"
          fill="none"
          stroke={palette.primary}
          strokeWidth={12}
          strokeLinecap="round"
          {...draws(rope)}
        />
        <path
          d="M140 74 C104 54 76 104 104 124 C128 141 152 122 134 104"
          fill="none"
          stroke={palette.accent}
          strokeWidth={11}
          strokeLinecap="round"
          {...draws(rope)}
        />
      </g>
      <g opacity={tail}>
        <path d="M122 140 C118 156 106 162 92 160" fill="none" stroke={palette.primary} strokeWidth={11} strokeLinecap="round" {...draws(tail)} />
      </g>
    </g>
  );
};

/** Causality: one tile falls and takes the rest with it. */
export const DominoFigure: Figure = ({build, t, palette}) => {
  const tiles = [0, 1, 2, 3];
  const run = cycle(t, 0.26);
  return (
    <g>
      {tiles.map((i) => {
        const p = stage(build, i, tiles.length, 0.55);
        const x = 46 + i * 36;
        // Each tile starts falling a beat after the one before it, which is
        // the only way a row of rectangles reads as a consequence.
        const own = Math.max(0, Math.min(1, (run - i * 0.13) / 0.16));
        const angle = ease(own) * 74;
        return (
          <g key={i} opacity={p} transform={`rotate(${angle} ${x + 9} 148)`}>
            <rect x={x} y={62} width={19} height={86} rx={5} fill={i === 0 ? palette.accent : palette.soft} />
            <rect x={x} y={62} width={19} height={86} rx={5} fill="none" stroke={palette.ink} strokeWidth={2} opacity={0.18} />
            <circle cx={x + 9.5} cy={88} r={3.5} fill={palette.ink} opacity={0.3} />
            <circle cx={x + 9.5} cy={122} r={3.5} fill={palette.ink} opacity={0.3} />
          </g>
        );
      })}
      <line x1={34} y1={150} x2={172} y2={150} stroke={palette.line} strokeWidth={3} strokeLinecap="round" opacity={0.3} />
    </g>
  );
};

/** Permission to go: a light that runs red, amber, green and holds on green. */
export const TrafficFigure: Figure = ({build, t, palette}) => {
  const box = stage(build, 0, 2);
  const lamps = stage(build, 1, 2);
  // Green holds for most of the cycle. A light that spends equal time on each
  // colour is a decoration; one that mostly says go is a signal.
  const phase = (t % 5.2) / 5.2;
  const on = phase < 0.22 ? 0 : phase < 0.34 ? 1 : 2;
  const colours = [palette.accent, palette.primary, palette.primary];
  return (
    <g>
      <g transform={popAt(box, C, C)} opacity={box}>
        <rect x={66} y={34} width={68} height={132} rx={18} fill={palette.soft} />
        <rect x={66} y={34} width={68} height={132} rx={18} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.18} />
      </g>
      <g opacity={lamps}>
        {[68, 100, 132].map((cy, i) => {
          const lit = i === on;
          return (
            <g key={cy}>
              <circle cx={C} cy={cy} r={17} fill={lit ? colours[i] : palette.ink} opacity={lit ? 1 : 0.22} />
              {lit && <circle cx={C} cy={cy} r={24} fill="none" stroke={colours[i]} strokeWidth={2} opacity={0.35} />}
            </g>
          );
        })}
      </g>
    </g>
  );
};

/** Passing it on: a baton that travels from one hand-post to the next. */
export const HandoffFigure: Figure = ({build, t, palette}) => {
  const posts = stage(build, 0, 2);
  const baton = stage(build, 1, 2);
  const pass = swing(t, 3.6, 0.3, 0.3);
  const x = 62 + pass * 76;
  const y = C - Math.sin(pass * Math.PI) * 26;
  return (
    <g>
      <g opacity={posts}>
        {[62, 138].map((px, i) => (
          <g key={px}>
            <circle cx={px} cy={C} r={22} fill={i === 0 ? palette.primary : palette.soft} opacity={i === 0 ? 1 - pass * 0.5 : 0.55 + 0.45 * pass} />
            <circle cx={px} cy={C} r={22} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.16} />
          </g>
        ))}
        <path d="M62 138 Q100 156 138 138" fill="none" stroke={palette.line} strokeWidth={3} strokeLinecap="round" opacity={0.32} {...draws(posts)} />
      </g>
      <g opacity={baton} transform={`rotate(${-24 + pass * 48} ${x} ${y})`}>
        <rect x={x - 22} y={y - 6} width={44} height={12} rx={6} fill={palette.accent} />
      </g>
    </g>
  );
};

/** Exactly here: a pin that drops onto a place and rings once. */
export const PinFigure: Figure = ({build, t, palette}) => {
  const ground = stage(build, 0, 2);
  const pinP = stage(build, 1, 2);
  const drop = swing(t, 3.8, 0.16, 0.56);
  const y = -54 + drop * 54;
  const landed = drop > 0.96;
  return (
    <g>
      <g opacity={ground}>
        <ellipse cx={C} cy={158} rx={52} ry={13} fill={palette.line} opacity={0.22} />
        <ellipse cx={C} cy={158} rx={16} ry={5} fill={palette.ink} opacity={0.3 * drop} />
      </g>
      <g opacity={pinP} transform={`translate(0 ${y})`}>
        <path
          d="M100 40 C128 40 146 60 146 84 C146 116 100 158 100 158 C100 158 54 116 54 84 C54 60 72 40 100 40 Z"
          fill={palette.primary}
        />
        <circle cx={C} cy={84} r={20} fill={palette.soft} />
        <circle cx={C} cy={84} r={9} fill={palette.accent} />
      </g>
      {landed && (
        <ellipse cx={C} cy={158} rx={18 + 46 * (drop - 0.96) * 25} ry={5 + 13 * (drop - 0.96) * 25} fill="none" stroke={palette.accent} strokeWidth={2.5} opacity={(1 - drop) * 25} />
      )}
    </g>
  );
};

/** The hard part, cleared: a runner that jumps the bar rather than stopping. */
export const HurdleFigure: Figure = ({build, t, palette}) => {
  const track = stage(build, 0, 3);
  const bar = stage(build, 1, 3);
  const mover = stage(build, 2, 3);
  const run = cycle(t, 0.36);
  const x = 26 + run * 148;
  // The arc is centred on the bar, so the jump is visibly *about* the bar.
  const over = Math.max(0, 1 - Math.abs(run - 0.5) * 4.2);
  const y = 138 - ease(over) * 62;
  return (
    <g>
      <line x1={22} y1={152} x2={178} y2={152} stroke={palette.line} strokeWidth={5} strokeLinecap="round" opacity={0.35 * track} />
      <g opacity={bar}>
        <rect x={72} y={92} width={56} height={11} rx={5.5} fill={palette.accent} />
        <rect x={74} y={98} width={8} height={54} rx={4} fill={palette.primary} transform="rotate(-9 78 125)" />
        <rect x={118} y={98} width={8} height={54} rx={4} fill={palette.primary} transform="rotate(9 122 125)" />
      </g>
      <g opacity={mover}>
        <circle cx={x} cy={y} r={16} fill={palette.soft} />
        <circle cx={x} cy={y} r={16} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.18} />
        {/* A motion trail while it is airborne. */}
        {over > 0.1 && (
          <circle cx={x - 22} cy={y + 10} r={10} fill={palette.soft} opacity={0.22 * over} />
        )}
      </g>
    </g>
  );
};
