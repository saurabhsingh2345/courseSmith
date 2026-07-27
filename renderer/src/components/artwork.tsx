// The flat-vector figure vocabulary for the illustration scene.
//
// These are drawn here rather than imported. The obvious alternative was to
// bundle a CC0 set (unDraw and friends), and it was rejected for a reason that
// only shows up once you try to animate one: a downloaded illustration is a
// single flat blob of paths, so a rocket's flame cannot flicker and a gear
// cannot turn. Owning the geometry means every figure has *parts*, and parts
// are what make this a video instead of a slideshow of clipart. It also means
// the figures speak the design system's palette instead of being recoloured
// towards it, and there is no third-party asset licence to keep track of.
//
// Every figure draws into the same 200x200 box and takes the same two clocks:
//
//   build  0-1, the figure assembling itself as its beat arrives. Parts are
//          staggered along it, so a figure lands as one gesture.
//   t      seconds since the scene started, for the idle motion that runs for
//          as long as the figure is on screen. The idle is the whole trick —
//          a figure that finishes assembling and then freezes reads as a
//          static image no matter how good the entrance was.
//
// Nothing here reads a clock itself and nothing is random, so frames stay
// independent and reproducible (Remotion renders them out of order).

export type FigurePalette = {
  accent: string;
  primary: string;
  /**
   * The shading colour, used to darken a face of a light mass. It is a *dark*
   * value, so it only ever works as a fill over something lighter — a stroke
   * painted with it on the stage is invisible. Strokes drawn on open
   * background want `line`.
   */
  ink: string;
  /** A light neutral for the main masses, already blended against the stage. */
  soft: string;
  /** A muted light, for rules and connectors drawn on the open stage. */
  line: string;
};

export type FigureProps = {
  build: number;
  t: number;
  palette: FigurePalette;
};

export type Figure = React.FC<FigureProps>;

/** The box every figure is drawn in. */
export const FIGURE_BOX = 200;

const clamp01 = (v: number): number => Math.max(0, Math.min(1, v));

/** Cubic ease-out: fast arrival, soft landing. */
const ease = (p: number): number => 1 - Math.pow(1 - clamp01(p), 3);

/**
 * Part `i` of `n`'s own progress along `build`.
 *
 * The parts overlap heavily rather than arriving strictly one after another. A
 * clean queue reads as a loading sequence; overlapping them reads as one thing
 * appearing, which is what a figure should do in the half-second it gets.
 */
const stage = (build: number, i: number, n: number, overlap = 0.6): number => {
  if (n <= 1) return clamp01(build);
  const span = 1 / (n - (n - 1) * overlap);
  return clamp01((clamp01(build) - i * span * (1 - overlap)) / span);
};

/** Scale for a part popping in, with a small overshoot at the midpoint. */
const popScale = (p: number): number =>
  0.55 + 0.45 * ease(p) + 0.09 * Math.sin(Math.PI * clamp01(p));

/** A part's group transform: pop in around a centre. */
const popAt = (p: number, cx: number, cy: number): string =>
  `translate(${cx} ${cy}) scale(${popScale(p)}) translate(${-cx} ${-cy})`;

/** Points of a regular polygon, for gear teeth and starbursts. */
const ring = (n: number, r: number, cx = 100, cy = 100, phase = 0) =>
  Array.from({length: n}, (_, i) => {
    const a = phase + (i / n) * Math.PI * 2;
    return {x: cx + r * Math.cos(a), y: cy + r * Math.sin(a), a};
  });

// ---------------------------------------------------------------------------
// The figures
// ---------------------------------------------------------------------------

/** A rocket climbing, with a flame that never sits still. */
const RocketFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 4);
  const fins = stage(build, 1, 4);
  const window_ = stage(build, 2, 4);
  const flame = stage(build, 3, 4);
  // A slow climb plus a faster flicker, so the flame reads as combustion
  // rather than as a pulsing shape.
  const lick = 1 + 0.17 * Math.sin(t * 16) + 0.09 * Math.sin(t * 27.3);
  const drift = Math.sin(t * 1.7) * 3;

  return (
    <g transform={`translate(0 ${drift})`}>
      {/* Exhaust. Two nested teardrops hanging off the nozzle: a lens shape
          closed at both ends read as a sliver of shadow rather than as fire,
          and at half opacity the accent came out darker than the stage. */}
      <g transform={popAt(flame, 100, 150)} opacity={flame}>
        <path
          d={`M84 146 Q100 ${146 + 62 * lick} 116 146 Z`}
          fill={palette.accent}
          opacity={0.9}
        />
        <path
          d={`M91 146 Q100 ${146 + 36 * lick} 109 146 Z`}
          fill="#fff3c4"
          opacity={0.95}
        />
      </g>
      {/* Fins */}
      <g transform={popAt(fins, 100, 140)} opacity={fins}>
        <path d="M84 120 L64 152 L84 146 Z" fill={palette.primary} />
        <path d="M116 120 L136 152 L116 146 Z" fill={palette.primary} />
      </g>
      {/* Body */}
      <g transform={popAt(body, 100, 100)} opacity={body}>
        <path
          d="M100 34 C122 56 130 92 130 122 L130 148 L70 148 L70 122 C70 92 78 56 100 34 Z"
          fill={palette.soft}
        />
        {/* The shaded side is what stops a flat shape reading as a sticker. */}
        <path
          d="M100 34 C122 56 130 92 130 122 L130 148 L100 148 Z"
          fill={palette.ink}
          opacity={0.12}
        />
      </g>
      {/* Porthole */}
      <g transform={popAt(window_, 100, 92)} opacity={window_}>
        <circle cx={100} cy={92} r={18} fill={palette.primary} />
        <circle cx={100} cy={92} r={18} fill="none" stroke={palette.accent} strokeWidth={4} />
        <circle cx={94} cy={86} r={5} fill="#ffffff" opacity={0.35} />
      </g>
    </g>
  );
};

/** A bulb whose filament lights and whose rays breathe. */
const LightbulbFigure: Figure = ({build, t, palette}) => {
  const glass = stage(build, 0, 4);
  const base = stage(build, 1, 4);
  const filament = stage(build, 2, 4);
  const rays = stage(build, 3, 4);
  const glow = 0.55 + 0.45 * (0.5 + 0.5 * Math.sin(t * 2.6));

  return (
    <g>
      {/* Rays, over the top arc only. Spaced evenly around the whole circle
          they read as a sun, and the bottom ones fire straight down through
          the screw base — which is the one direction a bulb does not light. */}
      <g opacity={rays}>
        {Array.from({length: 7}, (_, i) => {
          const a = ((-175 + (i * 170) / 6) * Math.PI) / 180;
          // Alternating lengths read as light; seven identical spokes read as
          // a symbol.
          const len = i % 2 === 0 ? 26 : 15;
          const r0 = 56 + 4 * glow;
          return (
            <line
              key={i}
              x1={100 + r0 * Math.cos(a)}
              y1={86 + r0 * Math.sin(a)}
              x2={100 + (r0 + len) * Math.cos(a)}
              y2={86 + (r0 + len) * Math.sin(a)}
              stroke={palette.accent}
              strokeWidth={5}
              strokeLinecap="round"
              opacity={0.35 + 0.5 * glow}
            />
          );
        })}
      </g>
      {/* Glass */}
      <g transform={popAt(glass, 100, 86)} opacity={glass}>
        <circle cx={100} cy={86} r={44} fill={palette.soft} />
        <path d="M100 42 A44 44 0 0 1 144 86 L100 86 Z" fill="#ffffff" opacity={0.12} />
      </g>
      {/* Screw base */}
      <g transform={popAt(base, 100, 142)} opacity={base}>
        <rect x={86} y={126} width={28} height={10} rx={4} fill={palette.primary} />
        <rect x={86} y={140} width={28} height={10} rx={4} fill={palette.primary} />
        {/* The contact tip. Painted in ink it is darker than the stage and
            reads as a hole punched under the bulb. */}
        <rect x={90} y={154} width={20} height={9} rx={4} fill={palette.soft} opacity={0.45} />
      </g>
      {/* Filament: two leads and a single loop between them. Both a zigzag and
          a wave were tried and both read as a handwritten letter sitting in
          the bulb — at this size the eye resolves any three-stroke squiggle as
          glyph before it resolves it as wire. One arc cannot be misread. */}
      <path
        d="M89 110 L89 95 C89 79 111 79 111 95 L111 110"
        fill="none"
        stroke={palette.accent}
        strokeWidth={5}
        strokeLinecap="round"
        strokeLinejoin="round"
        pathLength={1}
        strokeDasharray={1}
        strokeDashoffset={1 - ease(filament)}
        opacity={0.6 + 0.4 * glow}
      />
    </g>
  );
};

/** Two meshed gears, turning against each other for as long as they are up. */
const GearsFigure: Figure = ({build, t, palette}) => {
  const big = stage(build, 0, 2);
  const small = stage(build, 1, 2);

  const gear = (cx: number, cy: number, r: number, teeth: number, spin: number, fill: string) => (
    <g transform={`rotate(${spin} ${cx} ${cy})`}>
      {ring(teeth, r, cx, cy).map((p, i) => (
        <rect
          key={i}
          x={cx - r * 0.17}
          y={cy - r - r * 0.26}
          width={r * 0.34}
          height={r * 0.3}
          rx={r * 0.07}
          fill={fill}
          transform={`rotate(${(i / teeth) * 360} ${cx} ${cy})`}
        />
      ))}
      <circle cx={cx} cy={cy} r={r} fill={fill} />
      <circle cx={cx} cy={cy} r={r * 0.34} fill={palette.ink} opacity={0.55} />
    </g>
  );

  // Meshed teeth must turn in opposite directions at rates in inverse ratio to
  // their radii, or the two gears visibly grind through each other.
  const spin = t * 34;
  return (
    <g>
      <g transform={popAt(big, 76, 84)} opacity={big}>
        {gear(76, 84, 46, 10, spin, palette.primary)}
      </g>
      <g transform={popAt(small, 142, 134)} opacity={small}>
        {gear(142, 134, 32, 7, -spin * (46 / 32), palette.accent)}
      </g>
    </g>
  );
};

/** A cloud with traffic moving in both directions. */
const CloudFigure: Figure = ({build, t, palette}) => {
  const cloud = stage(build, 0, 2);
  const traffic = stage(build, 1, 2);
  // Two packets on opposite phases, so something is always in flight.
  const up = (t * 0.62) % 1;
  const down = ((t * 0.62 + 0.5) % 1 + 1) % 1;
  const fade = (p: number) => Math.min(1, Math.sin(p * Math.PI) * 2.4);

  return (
    <g>
      <g transform={popAt(cloud, 100, 78)} opacity={cloud}>
        <path
          d="M62 104 A26 26 0 0 1 68 54 A32 32 0 0 1 128 48 A24 24 0 0 1 140 104 Z"
          fill={palette.soft}
        />
        <path d="M100 48 A32 32 0 0 1 128 48 A24 24 0 0 1 140 104 L100 104 Z" fill={palette.ink} opacity={0.1} />
      </g>
      <g opacity={traffic}>
        {/* The lane the packets ride, so they are not floating in space. */}
        <line x1={78} y1={116} x2={78} y2={172} stroke={palette.line} strokeWidth={3} strokeDasharray="5 7" opacity={0.5} />
        <line x1={122} y1={116} x2={122} y2={172} stroke={palette.line} strokeWidth={3} strokeDasharray="5 7" opacity={0.5} />
        <rect
          x={68}
          y={166 - 50 * up}
          width={20}
          height={20}
          rx={5}
          fill={palette.accent}
          opacity={fade(up)}
        />
        <rect
          x={112}
          y={116 + 50 * down}
          width={20}
          height={20}
          rx={5}
          fill={palette.primary}
          opacity={fade(down)}
        />
      </g>
    </g>
  );
};

/** A shield with a check that draws itself, and a guard pulse. */
const ShieldFigure: Figure = ({build, t, palette}) => {
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
          cx={100}
          cy={100}
          r={62 + 26 * expand}
          fill="none"
          stroke={palette.accent}
          strokeWidth={3}
          opacity={0.4 * (1 - expand) * check}
        />
      )}
      <g transform={popAt(shield, 100, 100)} opacity={shield}>
        <path
          d="M100 34 L156 56 L156 104 C156 138 130 158 100 168 C70 158 44 138 44 104 L44 56 Z"
          fill={palette.primary}
        />
        <path
          d="M100 34 L156 56 L156 104 C156 138 130 158 100 168 Z"
          fill={palette.ink}
          opacity={0.16}
        />
      </g>
      <path
        d="M74 100 L94 120 L128 78"
        fill="none"
        stroke={palette.accent}
        strokeWidth={11}
        strokeLinecap="round"
        strokeLinejoin="round"
        pathLength={1}
        strokeDasharray={1}
        strokeDashoffset={1 - ease(check)}
      />
    </g>
  );
};

/** Bars growing under a trend line that draws across them. */
const ChartFigure: Figure = ({build, t, palette}) => {
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
        pathLength={1}
        strokeDasharray={1}
        strokeDashoffset={1 - ease(trend)}
      />
      <circle cx={162} cy={48} r={7} fill={palette.soft} opacity={trend >= 1 ? 1 : 0} />
    </g>
  );
};

/** Slabs settling into a stack. */
const StackFigure: Figure = ({build, t, palette}) => {
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
        const float = Math.sin(t * 1.6 + i * 0.9) * 2.4;
        const top = i === slabs.length - 1;
        const face = top ? palette.accent : palette.primary;
        return (
          <g key={i} opacity={p} transform={`translate(0 ${(1 - p) * -46 + float})`}>
            {/* An isometric slab: top face plus two side faces for thickness.
                The sides are the slab's own colour held down rather than a wash
                of ink — ink is darker than the stage, so painted over open
                background it removed the thickness instead of shading it. */}
            <path
              d={`M100 ${y - 22} L156 ${y} L100 ${y + 22} L44 ${y} Z`}
              fill={face}
              opacity={top ? 1 : 0.85 - i * 0.06}
            />
            <path
              d={`M44 ${y} L100 ${y + 22} L100 ${y + 32} L44 ${y + 10} Z`}
              fill={face}
              opacity={0.42}
            />
            <path
              d={`M156 ${y} L100 ${y + 22} L100 ${y + 32} L156 ${y + 10} Z`}
              fill={face}
              opacity={0.6}
            />
          </g>
        );
      })}
    </g>
  );
};

/** A padlock whose key turns in it, holds, and releases. */
const LockFigure: Figure = ({build, t, palette}) => {
  const lock = stage(build, 0, 2);
  const key = stage(build, 1, 2);
  // The key seats itself, then turns a quarter and holds — a key that spins
  // forever is a fidget, not a mechanism.
  const cycle = (t % 4) / 4;
  const turn = cycle < 0.25 ? 0 : cycle < 0.4 ? ease((cycle - 0.25) / 0.15) * 90 : cycle < 0.85 ? 90 : 90 * (1 - ease((cycle - 0.85) / 0.15));

  return (
    <g>
      <g transform={popAt(lock, 100, 116)} opacity={lock}>
        <path
          d="M76 92 L76 74 A24 24 0 0 1 124 74 L124 92"
          fill="none"
          stroke={palette.soft}
          strokeWidth={12}
          strokeLinecap="round"
        />
        <rect x={58} y={92} width={84} height={68} rx={12} fill={palette.primary} />
        <path d="M100 92 L142 92 L142 160 L100 160 Z" fill={palette.ink} opacity={0.15} />
      </g>
      <g opacity={key} transform={`rotate(${turn} 100 122)`}>
        <circle cx={100} cy={122} r={13} fill="none" stroke={palette.accent} strokeWidth={7} />
        <rect x={97} y={130} width={6} height={26} rx={3} fill={palette.accent} />
        <rect x={97} y={144} width={16} height={6} rx={3} fill={palette.accent} />
      </g>
    </g>
  );
};

/** A clock whose hands actually move. */
const ClockFigure: Figure = ({build, t, palette}) => {
  const face = stage(build, 0, 3);
  const ticks = stage(build, 1, 3);
  const hands = stage(build, 2, 3);
  // A whole revolution in twelve seconds: fast enough to read as motion in a
  // short clip, slow enough not to read as a stopwatch gone wrong.
  const second = t * 30;

  return (
    <g>
      <g transform={popAt(face, 100, 100)} opacity={face}>
        <circle cx={100} cy={100} r={68} fill={palette.soft} />
        <circle cx={100} cy={100} r={68} fill="none" stroke={palette.primary} strokeWidth={9} />
      </g>
      <g opacity={ticks}>
        {ring(12, 54).map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r={i % 3 === 0 ? 4.5 : 2.5} fill={palette.ink} opacity={0.55} />
        ))}
      </g>
      <g opacity={hands}>
        <line
          x1={100}
          y1={100}
          x2={100}
          y2={64}
          stroke={palette.ink}
          strokeWidth={8}
          strokeLinecap="round"
          transform={`rotate(${second / 12} 100 100)`}
          opacity={0.75}
        />
        <line
          x1={100}
          y1={100}
          x2={100}
          y2={50}
          stroke={palette.accent}
          strokeWidth={6}
          strokeLinecap="round"
          transform={`rotate(${second} 100 100)`}
        />
        <circle cx={100} cy={100} r={7} fill={palette.accent} />
      </g>
    </g>
  );
};

/** A hub with packets running out along its spokes. */
const NetworkFigure: Figure = ({build, t, palette}) => {
  const links = stage(build, 0, 3);
  const hub = stage(build, 1, 3);
  const leaves = stage(build, 2, 3);
  const nodes = ring(5, 62, 100, 100, -Math.PI / 2);

  return (
    <g>
      {/* Spokes, and a packet running each one on its own phase. */}
      {nodes.map((p, i) => {
        const lp = ease(stage(links, i, nodes.length, 0.7));
        const travel = ((t * 0.55 + i / nodes.length) % 1 + 1) % 1;
        const fade = Math.min(1, Math.sin(travel * Math.PI) * 2.6);
        return (
          <g key={i}>
            <line
              x1={100}
              y1={100}
              x2={100 + (p.x - 100) * lp}
              y2={100 + (p.y - 100) * lp}
              stroke={palette.line}
              strokeWidth={3.5}
              strokeLinecap="round"
              opacity={0.55}
            />
            {lp >= 1 && (
              <circle
                cx={100 + (p.x - 100) * travel}
                cy={100 + (p.y - 100) * travel}
                r={5}
                fill={palette.accent}
                opacity={fade}
              />
            )}
          </g>
        );
      })}
      {nodes.map((p, i) => {
        const np = stage(leaves, i, nodes.length, 0.65);
        return (
          <g key={`n${i}`} transform={popAt(np, p.x, p.y)} opacity={np}>
            <circle cx={p.x} cy={p.y} r={16} fill={palette.primary} />
            <circle cx={p.x - 4} cy={p.y - 5} r={4} fill="#ffffff" opacity={0.25} />
          </g>
        );
      })}
      <g transform={popAt(hub, 100, 100)} opacity={hub}>
        <circle cx={100} cy={100} r={24} fill={palette.accent} />
        <circle cx={100} cy={100} r={24} fill="none" stroke={palette.soft} strokeWidth={4} opacity={0.5} />
      </g>
    </g>
  );
};

/** The neutral fallback: a burst with a few things in orbit. */
const SparkFigure: Figure = ({build, t, palette}) => {
  const star = stage(build, 0, 2);
  const orbit = stage(build, 1, 2);
  const breathe = 1 + 0.06 * Math.sin(t * 2.4);

  return (
    <g>
      <g transform={`${popAt(star, 100, 100)} rotate(${t * 8} 100 100)`} opacity={star}>
        {/* The four control points must mirror about the centre in both axes.
            Pinning them all to one corner (which the first pass did) makes a
            star that leans, and a lopsided sparkle reads as a mistake. */}
        <path
          d={`M100 ${100 - 62 * breathe} Q106 94 ${100 + 62 * breathe} 100 Q106 106 100 ${100 + 62 * breathe} Q94 106 ${100 - 62 * breathe} 100 Q94 94 100 ${100 - 62 * breathe} Z`}
          fill={palette.accent}
        />
      </g>
      <g opacity={orbit}>
        {ring(3, 78, 100, 100, t * 0.7).map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r={7 - i} fill={palette.primary} opacity={0.8} />
        ))}
      </g>
    </g>
  );
};

/**
 * The closed figure vocabulary.
 *
 * Go owns the list (artFigureVocab in snippet_illustration.go) and normalizes
 * anything outside it to "spark"; a drift test keeps the two identical, because
 * a figure Go allows and this map does not have would render as an empty stage
 * and nobody would notice until they watched the clip.
 */
export const FIGURES: Record<string, Figure> = {
  rocket: RocketFigure,
  lightbulb: LightbulbFigure,
  gears: GearsFigure,
  cloud: CloudFigure,
  shield: ShieldFigure,
  chart: ChartFigure,
  stack: StackFigure,
  lock: LockFigure,
  clock: ClockFigure,
  network: NetworkFigure,
  spark: SparkFigure,
};

/** The figure for a vocabulary name, falling back to the neutral burst. */
export const figureFor = (name: string | undefined): Figure =>
  (name && FIGURES[name]) || FIGURES.spark;
