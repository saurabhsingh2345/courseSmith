// Desks, days and the things people aim at.
//
// The set an explainer reaches for when the subject is a practice rather than a
// system — planning, deciding, agreeing, finishing. Several of these are the
// closest this vocabulary gets to a person without being one, which is exactly
// why they exist: the cast template owns people, and the illustration template
// needs a way to talk about what people *do* without borrowing them.

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

/** A book that opens, with pages that turn. */
export const BookFigure: Figure = ({build, t, palette}) => {
  const covers = stage(build, 0, 2);
  const pages = stage(build, 1, 2);
  const turn = cycle(t, 0.32);
  return (
    <g>
      <g transform={popAt(covers, C, C)} opacity={covers}>
        <path d="M100 58 C78 44 50 42 28 46 L28 150 C50 146 78 148 100 162 Z" fill={palette.primary} />
        <path d="M100 58 C122 44 150 42 172 46 L172 150 C150 146 122 148 100 162 Z" fill={palette.primary} />
        <path d="M100 58 C122 44 150 42 172 46 L172 150 C150 146 122 148 100 162 Z" fill={palette.ink} opacity={0.13} />
      </g>
      <g opacity={pages}>
        {/* A single leaf lifting and falling over: a whole stack flipping at
            once reads as the book closing. */}
        <path
          d={`M100 58 C${100 + 68 * Math.cos(turn * Math.PI)} ${44} ${100 + 70 * Math.cos(turn * Math.PI)} ${44} ${100 + 70 * Math.cos(turn * Math.PI)} 48 L${100 + 70 * Math.cos(turn * Math.PI)} 148 C${100 + 40 * Math.cos(turn * Math.PI)} 146 120 150 100 160 Z`}
          fill={palette.soft}
          opacity={0.85}
        />
        <rect x={98} y={58} width={4} height={104} fill={palette.ink} opacity={0.18} />
      </g>
    </g>
  );
};

/** A ruled pad with lines appearing on it. */
export const NotebookFigure: Figure = ({build, t, palette}) => {
  const pad = stage(build, 0, 2);
  const lines = stage(build, 1, 2);
  const written = swing(t, 6, 0.45, 0.35);
  const rows = [86, 106, 126, 146];
  return (
    <g transform="rotate(-4 100 100)">
      <g transform={popAt(pad, C, C)} opacity={pad}>
        <rect x={42} y={34} width={116} height={140} rx={10} fill={palette.soft} />
        <rect x={42} y={34} width={20} height={140} fill={palette.primary} />
        {[0, 1, 2, 3].map((i) => (
          <circle key={i} cx={52} cy={58 + i * 34} r={6} fill={palette.soft} opacity={0.8} />
        ))}
      </g>
      <g opacity={lines}>
        <rect x={74} y={56} width={64} height={9} rx={4.5} fill={palette.accent} />
        {rows.map((y, i) => {
          const w = [70, 58, 66, 44][i];
          const on = Math.max(0, Math.min(1, written * 4.4 - i));
          return <rect key={i} x={74} y={y} width={w * on} height={7} rx={3.5} fill={palette.ink} opacity={0.22} />;
        })}
      </g>
    </g>
  );
};

/** A pencil that writes a line and lifts. */
export const PencilFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 2);
  const mark = stage(build, 1, 2);
  const run = swing(t, 4.4, 0.4, 0.25);
  const x = 44 + run * 96;
  return (
    <g>
      <path
        d="M44 148 C74 132 104 156 140 138"
        fill="none"
        stroke={palette.accent}
        strokeWidth={6}
        strokeLinecap="round"
        opacity={mark}
        pathLength={1}
        strokeDasharray={1}
        strokeDashoffset={1 - run}
      />
      <g opacity={body} transform={`translate(${x - 100} ${-run * 4}) rotate(34 100 100)`}>
        <rect x={88} y={22} width={26} height={96} fill={palette.primary} />
        <rect x={88} y={22} width={26} height={16} fill={palette.accent} />
        <rect x={88} y={38} width={26} height={9} fill={palette.line} opacity={0.7} />
        <path d="M88 118 L114 118 L101 146 Z" fill={palette.soft} />
        <path d="M95 132 L107 132 L101 146 Z" fill={palette.ink} opacity={0.7} />
      </g>
    </g>
  );
};

/** A folder whose tab lifts to show what is inside. */
export const FolderFigure: Figure = ({build, t, palette}) => {
  const back = stage(build, 0, 3);
  const papers = stage(build, 1, 3);
  const front = stage(build, 2, 3);
  const open = swing(t, 4.6, 0.2, 0.4);
  return (
    <g>
      <g transform={popAt(back, C, C)} opacity={back}>
        <path d="M28 62 L84 62 L96 78 L172 78 L172 152 A8 8 0 0 1 164 160 L36 160 A8 8 0 0 1 28 152 Z" fill={palette.primary} />
      </g>
      <g opacity={papers} transform={`translate(0 ${-open * 16})`}>
        <rect x={54} y={72} width={92} height={78} rx={5} fill={palette.soft} />
        <rect x={68} y={90} width={62} height={7} rx={3.5} fill={palette.ink} opacity={0.2} />
        <rect x={68} y={106} width={44} height={7} rx={3.5} fill={palette.ink} opacity={0.15} />
      </g>
      <g opacity={front} transform={`rotate(${-open * 7} 30 160)`}>
        <path d="M28 92 L172 92 L172 152 A8 8 0 0 1 164 160 L36 160 A8 8 0 0 1 28 152 Z" fill={palette.primary} />
        <path d="M100 92 L172 92 L172 152 A8 8 0 0 1 164 160 L100 160 Z" fill={palette.ink} opacity={0.14} />
      </g>
    </g>
  );
};

/** A month, with today moving through it. */
export const CalendarFigure: Figure = ({build, t, palette}) => {
  const pad = stage(build, 0, 3);
  const head = stage(build, 1, 3);
  const days = stage(build, 2, 3);
  const today = Math.floor(t * 1.6) % 12;
  return (
    <g>
      <g transform={popAt(pad, C, C)} opacity={pad}>
        <rect x={32} y={44} width={136} height={124} rx={12} fill={palette.soft} />
      </g>
      <g opacity={head}>
        <path d="M32 56 A12 12 0 0 1 44 44 L156 44 A12 12 0 0 1 168 56 L168 78 L32 78 Z" fill={palette.primary} />
        <rect x={58} y={30} width={11} height={26} rx={5.5} fill={palette.line} />
        <rect x={131} y={30} width={11} height={26} rx={5.5} fill={palette.line} />
      </g>
      <g opacity={days}>
        {Array.from({length: 12}, (_, i) => {
          const cx = 56 + (i % 4) * 30;
          const cy = 100 + Math.floor(i / 4) * 28;
          const on = i === today;
          return (
            <rect
              key={i}
              x={cx - 11}
              y={cy - 10}
              width={22}
              height={20}
              rx={5}
              fill={on ? palette.accent : palette.ink}
              opacity={on ? 1 : 0.14}
            />
          );
        })}
      </g>
    </g>
  );
};

/** A case that swings by its handle. */
export const BriefcaseFigure: Figure = ({build, t, palette}) => {
  const shell = stage(build, 0, 2);
  const handle = stage(build, 1, 2);
  const swingAngle = bob(t, 4, 1.1);
  return (
    <g transform={`rotate(${swingAngle} 100 46)`}>
      <g opacity={handle}>
        <path d="M74 74 L74 54 A26 14 0 0 1 126 54 L126 74" fill="none" stroke={palette.line} strokeWidth={8} strokeLinecap="round" />
      </g>
      <g transform={popAt(shell, C, 122)} opacity={shell}>
        <rect x={28} y={74} width={144} height={92} rx={12} fill={palette.primary} />
        <path d="M100 74 L172 74 L172 166 L100 166 Z" fill={palette.ink} opacity={0.13} />
        <rect x={28} y={110} width={144} height={12} fill={palette.ink} opacity={0.18} />
        <rect x={86} y={104} width={28} height={24} rx={6} fill={palette.accent} />
      </g>
    </g>
  );
};

/** A cup with steam that rises and breaks up. */
export const CoffeeFigure: Figure = ({build, t, palette}) => {
  const cup = stage(build, 0, 3);
  const handle = stage(build, 1, 3);
  const steam = stage(build, 2, 3);
  return (
    <g>
      <g opacity={steam}>
        {[0, 1, 2].map((i) => {
          const p = cycle(t, 0.42, i * 0.33);
          return (
            <path
              key={i}
              d={`M${80 + i * 20} ${72 - p * 34} q${8 + i * 2} -12 0 -24`}
              fill="none"
              stroke={palette.line}
              strokeWidth={5}
              strokeLinecap="round"
              opacity={fadeTravel(p) * 0.65}
            />
          );
        })}
      </g>
      <g opacity={handle}>
        <path d="M146 100 A24 24 0 0 1 146 142" fill="none" stroke={palette.primary} strokeWidth={11} strokeLinecap="round" />
      </g>
      <g transform={popAt(cup, C, 128)} opacity={cup}>
        <path d="M52 88 L148 88 L140 156 A10 10 0 0 1 130 166 L70 166 A10 10 0 0 1 60 156 Z" fill={palette.soft} />
        <path d="M100 88 L148 88 L140 156 A10 10 0 0 1 130 166 L100 166 Z" fill={palette.ink} opacity={0.11} />
        <ellipse cx={C} cy={88} rx={48} ry={9} fill={palette.accent} />
      </g>
    </g>
  );
};

/** Rings with an arrow that lands in the middle. */
export const TargetFigure: Figure = ({build, t, palette}) => {
  const rings = [68, 46, 24];
  const arrow = stage(build, 3, 4);
  const shot = gesture(t, 3.4, 0.28);
  return (
    <g>
      {rings.map((r, i) => {
        const p = stage(build, i, 4, 0.55);
        return (
          <circle
            key={i}
            cx={C}
            cy={C}
            r={r}
            fill={i % 2 === 0 ? palette.soft : palette.primary}
            opacity={p}
            transform={popAt(p, C, C)}
          />
        );
      })}
      <circle cx={C} cy={C} r={10} fill={palette.accent} opacity={stage(build, 2, 4, 0.55)} />
      <g opacity={arrow}>
        {shot >= 0 ? (
          <g transform={`translate(${(1 - ease(shot)) * 70} ${(1 - ease(shot)) * -70})`}>
            <line x1={C} y1={C} x2={158} y2={42} stroke={palette.line} strokeWidth={6} strokeLinecap="round" />
            <path d="M150 34 L166 34 L166 50 Z" fill={palette.accent} />
          </g>
        ) : (
          <g>
            <line x1={C} y1={C} x2={158} y2={42} stroke={palette.line} strokeWidth={6} strokeLinecap="round" />
            <path d="M150 34 L166 34 L166 50 Z" fill={palette.accent} />
          </g>
        )}
      </g>
    </g>
  );
};

/** A cup with a shine that travels across it. */
export const TrophyFigure: Figure = ({build, t, palette}) => {
  const cup = stage(build, 0, 3);
  const handles = stage(build, 1, 3);
  const base = stage(build, 2, 3);
  const shine = cycle(t, 0.4);
  return (
    <g transform={`translate(0 ${bob(t, 2, 1.2)})`}>
      <g opacity={handles}>
        <path d="M58 56 A22 22 0 0 0 58 100" fill="none" stroke={palette.primary} strokeWidth={10} strokeLinecap="round" />
        <path d="M142 56 A22 22 0 0 1 142 100" fill="none" stroke={palette.primary} strokeWidth={10} strokeLinecap="round" />
      </g>
      <g transform={popAt(cup, C, 88)} opacity={cup}>
        <path d="M58 44 L142 44 L142 84 A42 42 0 0 1 58 84 Z" fill={palette.accent} />
        <path d="M100 44 L142 44 L142 84 A42 42 0 0 1 100 126 Z" fill={palette.ink} opacity={0.13} />
        {/* The shine: a bright band crossing the metal. */}
        <rect x={58 + shine * 84 - 8} y={44} width={12} height={54} fill={palette.soft} opacity={fadeTravel(shine) * 0.5} transform="skewX(-14)" />
      </g>
      <g opacity={base}>
        <rect x={92} y={126} width={16} height={22} fill={palette.primary} />
        <path d="M66 148 L134 148 L142 168 L58 168 Z" fill={palette.primary} />
      </g>
    </g>
  );
};

/** A flag planted and rippling. */
export const FlagFigure: Figure = ({build, t, palette}) => {
  const pole = stage(build, 0, 2);
  const cloth = stage(build, 1, 2);
  const ripple = (x: number) => Math.sin(x * 0.06 + t * 3.4) * 7;
  return (
    <g>
      <g opacity={pole}>
        <rect x={54} y={26} width={9} height={148} rx={4.5} fill={palette.line} opacity={0.75} />
        <ellipse cx={58} cy={172} rx={26} ry={7} fill={palette.line} opacity={0.3} />
      </g>
      <path
        d={`M63 36 C 90 ${30 + ripple(90)}, 118 ${44 + ripple(118)}, 152 ${34 + ripple(152)} L152 ${86 + ripple(152)} C118 ${96 + ripple(118)}, 90 ${82 + ripple(90)}, 63 ${88} Z`}
        fill={palette.accent}
        opacity={cloth}
      />
    </g>
  );
};

/** A key that turns in mid-air, showing its bit. */
export const KeyFigure: Figure = ({build, t, palette}) => {
  const bow = stage(build, 0, 2);
  const blade = stage(build, 1, 2);
  const turn = swing(t, 4.2, 0.2, 0.4) * 40;
  return (
    <g transform={`rotate(${-20 + turn} 68 100)`}>
      <g transform={popAt(bow, 68, C)} opacity={bow}>
        <circle cx={68} cy={C} r={32} fill="none" stroke={palette.primary} strokeWidth={14} />
        <circle cx={68} cy={C} r={11} fill={palette.soft} />
      </g>
      <g opacity={blade}>
        <rect x={98} y={92} width={72} height={16} rx={5} fill={palette.primary} />
        <rect x={140} y={108} width={12} height={18} rx={4} fill={palette.accent} />
        <rect x={160} y={108} width={12} height={13} rx={4} fill={palette.accent} />
      </g>
    </g>
  );
};

/** A door that opens onto light. */
export const DoorFigure: Figure = ({build, t, palette}) => {
  const frame = stage(build, 0, 2);
  const leaf = stage(build, 1, 2);
  const open = swing(t, 5, 0.25, 0.4);
  return (
    <g>
      <g transform={popAt(frame, C, C)} opacity={frame}>
        <rect x={48} y={26} width={104} height={148} rx={10} fill={palette.soft} />
        <rect x={60} y={38} width={80} height={136} fill={palette.accent} opacity={0.35 + 0.5 * open} />
      </g>
      {/* The leaf swings by squashing its width, which at this size reads as a
          door opening and costs no perspective maths. */}
      <g opacity={leaf} transform={`translate(60 0) scale(${1 - 0.86 * open} 1) translate(-60 0)`}>
        <rect x={60} y={38} width={80} height={136} rx={4} fill={palette.primary} />
        <rect x={60} y={38} width={80} height={136} rx={4} fill={palette.ink} opacity={0.1} />
        <circle cx={126} cy={110} r={7} fill={palette.soft} />
      </g>
    </g>
  );
};

/** Two hands meeting and shaking. */
export const HandshakeFigure: Figure = ({build, t, palette}) => {
  const arms = stage(build, 0, 2);
  const grip = stage(build, 1, 2);
  const shake = Math.sin(t * 4.4) * 4;
  return (
    <g transform={`translate(0 ${shake})`}>
      <g opacity={arms}>
        <path d="M14 128 L74 96" stroke={palette.primary} strokeWidth={26} strokeLinecap="round" fill="none" />
        <path d="M186 128 L126 96" stroke={palette.soft} strokeWidth={26} strokeLinecap="round" fill="none" />
      </g>
      <g transform={popAt(grip, C, 100)} opacity={grip}>
        <path d="M70 90 L118 76 A14 14 0 0 1 126 102 L92 116 A14 14 0 0 1 70 90 Z" fill={palette.primary} />
        <path d="M130 90 L82 104 A14 14 0 0 0 90 130 L124 116 A14 14 0 0 0 130 90 Z" fill={palette.soft} />
        <path d="M130 90 L100 99 L92 116 L124 116 A14 14 0 0 0 130 90 Z" fill={palette.ink} opacity={0.1} />
      </g>
    </g>
  );
};

/** A brain with a thought crossing it. */
export const BrainFigure: Figure = ({build, t, palette}) => {
  const mass = stage(build, 0, 2);
  const folds = stage(build, 1, 2);
  return (
    <g>
      <g transform={popAt(mass, C, C)} opacity={mass}>
        <path
          d="M100 32 C72 32 58 46 58 60 C40 62 32 78 40 92 C28 104 34 124 50 128 C52 148 70 158 86 150 C92 164 112 164 118 150 C136 158 152 146 152 128 C168 122 172 102 160 92 C168 76 158 62 142 60 C142 44 126 32 100 32 Z"
          fill={palette.primary}
        />
        <path d="M100 32 C126 32 142 44 142 60 C158 62 168 76 160 92 C172 102 168 122 152 128 C152 146 136 158 118 150 C112 164 100 164 100 150 Z" fill={palette.ink} opacity={0.12} />
      </g>
      <g opacity={folds} fill="none" stroke={palette.soft} strokeWidth={4} strokeLinecap="round">
        <path d="M100 44 L100 150" opacity={0.5} />
        {[0, 1, 2].map((i) => {
          // A signal running the folds, so the brain is thinking rather than
          // sitting there being a brain.
          const p = cycle(t, 0.5, i * 0.33);
          return (
            <g key={i}>
              <path d={`M${74 - i * 6} ${70 + i * 26} q14 ${i % 2 ? 12 : -12} 26 0`} opacity={0.45} />
              <path d={`M${126 + i * 6} ${70 + i * 26} q-14 ${i % 2 ? 12 : -12} -26 0`} opacity={0.45} />
              <circle cx={74 + p * 52} cy={70 + i * 26} r={5} fill={palette.accent} stroke="none" opacity={fadeTravel(p)} />
            </g>
          );
        })}
      </g>
    </g>
  );
};

/** An eye that looks around and blinks. */
export const EyeFigure: Figure = ({build, t, palette}) => {
  const white = stage(build, 0, 3);
  const iris = stage(build, 1, 3);
  const lash = stage(build, 2, 3);
  const look = Math.sin(t * 0.9) * 18;
  // The same trick the character uses: a lid that closes, not a pupil that
  // shrinks.
  const blink = (t % 4.2) < 0.14;
  return (
    <g>
      <g transform={popAt(white, C, C)} opacity={white}>
        <path d="M18 100 C50 56 150 56 182 100 C150 144 50 144 18 100 Z" fill={palette.soft} />
      </g>
      {blink ? (
        <path d="M18 100 C50 92 150 92 182 100" fill="none" stroke={palette.primary} strokeWidth={10} strokeLinecap="round" opacity={iris} />
      ) : (
        <g opacity={iris}>
          <circle cx={C + look} cy={C} r={30} fill={palette.primary} />
          <circle cx={C + look} cy={C} r={13} fill={palette.ink} opacity={0.75} />
          <circle cx={C + look - 10} cy={90} r={7} fill={palette.soft} opacity={0.7} />
        </g>
      )}
      <path
        d="M18 100 C50 56 150 56 182 100 C150 144 50 144 18 100 Z"
        fill="none"
        stroke={palette.primary}
        strokeWidth={8}
        strokeLinejoin="round"
        opacity={lash}
        {...draws(lash)}
      />
    </g>
  );
};

/** A list whose items get ticked off, then start again. */
export const ChecklistFigure: Figure = ({build, t, palette}) => {
  const card = stage(build, 0, 2);
  const items = stage(build, 1, 2);
  const done = (t % 5) / 5;
  const rows = [0, 1, 2, 3];
  return (
    <g>
      <g transform={popAt(card, C, C)} opacity={card}>
        <rect x={34} y={30} width={132} height={144} rx={12} fill={palette.soft} />
        <rect x={76} y={20} width={48} height={20} rx={8} fill={palette.line} />
      </g>
      <g opacity={items}>
        {rows.map((i) => {
          const y = 66 + i * 30;
          const ticked = done * 4.6 > i + 1;
          const tick = Math.max(0, Math.min(1, done * 4.6 - i - 1));
          return (
            <g key={i}>
              <rect x={52} y={y - 12} width={24} height={24} rx={6} fill={ticked ? palette.accent : palette.ink} opacity={ticked ? 1 : 0.14} />
              {ticked && (
                <path
                  d={`M58 ${y} L64 ${y + 6} L72 ${y - 7}`}
                  fill="none"
                  stroke={palette.soft}
                  strokeWidth={4}
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  {...draws(tick)}
                />
              )}
              <rect x={88} y={y - 4} width={[58, 44, 52, 38][i]} height={8} rx={4} fill={palette.ink} opacity={ticked ? 0.14 : 0.24} />
              {ticked && <rect x={88} y={y} width={[58, 44, 52, 38][i] * tick} height={3} rx={1.5} fill={palette.ink} opacity={0.3} />}
            </g>
          );
        })}
      </g>
    </g>
  );
};

/** A bulb whose filament lights and whose rays breathe. */
export const LightbulbFigure: Figure = ({build, t, palette}) => {
  const glass = stage(build, 0, 4);
  const base = stage(build, 1, 4);
  const filament = stage(build, 2, 4);
  const rays = stage(build, 3, 4);
  const glow = 0.55 + 0.45 * pulse(t, 2.6);

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
              x1={C + r0 * Math.cos(a)}
              y1={86 + r0 * Math.sin(a)}
              x2={C + (r0 + len) * Math.cos(a)}
              y2={86 + (r0 + len) * Math.sin(a)}
              stroke={palette.accent}
              strokeWidth={5}
              strokeLinecap="round"
              opacity={0.35 + 0.5 * glow}
            />
          );
        })}
      </g>
      <g transform={popAt(glass, C, 86)} opacity={glass}>
        <circle cx={C} cy={86} r={44} fill={palette.soft} />
        <path d="M100 42 A44 44 0 0 1 144 86 L100 86 Z" fill="#ffffff" opacity={0.12} />
      </g>
      <g transform={popAt(base, C, 142)} opacity={base}>
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
        opacity={0.6 + 0.4 * glow}
        {...draws(filament)}
      />
    </g>
  );
};

/** Two meshed gears, turning against each other for as long as they are up. */
export const GearsFigure: Figure = ({build, t, palette}) => {
  const big = stage(build, 0, 2);
  const small = stage(build, 1, 2);

  const gear = (cx: number, cy: number, r: number, teeth: number, spin: number, fill: string) => (
    <g transform={`rotate(${spin} ${cx} ${cy})`}>
      {ring(teeth, r, cx, cy).map((_, i) => (
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
