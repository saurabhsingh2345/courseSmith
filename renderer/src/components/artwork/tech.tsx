// Software, infrastructure and the machines it runs on.
//
// The rule these all follow: whatever the object *does* is the thing that
// moves. A server's lights blink, a queue drains, a pipeline carries something
// through, a bug crawls. An object drawn accurately and then held still is a
// screenshot, and the whole reason for owning the geometry is that we do not
// have to hold anything still.

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

/** Racked boxes with a status light on each. */
export const ServerFigure: Figure = ({build, t, palette}) => {
  const units = [0, 1, 2];
  return (
    <g>
      {units.map((i) => {
        const p = stage(build, units.length - 1 - i, units.length, 0.5);
        const y = 132 - i * 42;
        // Each unit blinks on its own period, so the rack reads as three
        // machines rather than one thing flashing.
        const on = pulse(t, 2.2 + i * 0.7, i * 1.9) > 0.45;
        return (
          <g key={i} transform={popAt(p, C, y + 16)} opacity={p}>
            <rect x={46} y={y} width={108} height={34} rx={7} fill={palette.soft} />
            <rect x={46} y={y} width={108} height={34} rx={7} fill="none" stroke={palette.ink} strokeWidth={2} opacity={0.2} />
            <circle cx={62} cy={y + 17} r={5} fill={on ? palette.accent : palette.ink} opacity={on ? 1 : 0.25} />
            <rect x={78} y={y + 13} width={54} height={8} rx={4} fill={palette.ink} opacity={0.16} />
          </g>
        );
      })}
    </g>
  );
};

/** A stacked cylinder with a query sweeping down it. */
export const DatabaseFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 2);
  const sweep = stage(build, 1, 2);
  const bands = [70, 104, 138];
  const at = cycle(t, 0.5);
  return (
    <g transform={popAt(body, C, C)} opacity={body}>
      <path d="M46 62 L46 138 A54 20 0 0 0 154 138 L154 62 Z" fill={palette.soft} />
      <path d="M100 42 A54 20 0 0 1 154 62 L154 138 A54 20 0 0 1 100 158 Z" fill={palette.ink} opacity={0.1} />
      <ellipse cx={C} cy={62} rx={54} ry={20} fill={palette.primary} />
      {bands.map((y, i) => (
        <path
          key={i}
          d={`M46 ${y} A54 20 0 0 0 154 ${y}`}
          fill="none"
          stroke={palette.ink}
          strokeWidth={3}
          opacity={0.18}
        />
      ))}
      {/* The read: a band lighting as it passes. */}
      <path
        d={`M46 ${62 + at * 82} A54 20 0 0 0 154 ${62 + at * 82}`}
        fill="none"
        stroke={palette.accent}
        strokeWidth={5}
        strokeLinecap="round"
        opacity={fadeTravel(at) * sweep}
      />
    </g>
  );
};

/** A prompt with a cursor that blinks and a line that types. */
export const TerminalFigure: Figure = ({build, t, palette}) => {
  const frame = stage(build, 0, 2);
  const text = stage(build, 1, 2);
  // A line types, holds, and clears — which is what a terminal does, and what
  // a permanently full one does not.
  const typed = swing(t, 5, 0.3, 0.5);
  const caret = pulse(t, 6) > 0.5;
  return (
    <g transform={popAt(frame, C, C)} opacity={frame}>
      <rect x={30} y={46} width={140} height={108} rx={10} fill={palette.soft} />
      <rect x={30} y={46} width={140} height={24} rx={10} fill={palette.primary} />
      <rect x={30} y={62} width={140} height={8} fill={palette.primary} />
      {ring(3, 0, 48, 58).map((_, i) => (
        <circle key={i} cx={46 + i * 13} cy={58} r={4} fill={palette.soft} opacity={0.7} />
      ))}
      <g opacity={text}>
        <path d="M44 92 L54 100 L44 108" fill="none" stroke={palette.accent} strokeWidth={4} strokeLinecap="round" strokeLinejoin="round" />
        <rect x={62} y={95} width={70 * typed} height={7} rx={3.5} fill={palette.ink} opacity={0.35} />
        {caret && <rect x={64 + 70 * typed} y={92} width={8} height={13} fill={palette.accent} />}
        <rect x={44} y={122} width={54} height={7} rx={3.5} fill={palette.ink} opacity={0.18} />
      </g>
    </g>
  );
};

/** Angle brackets around a slash, with the brackets breathing apart. */
export const CodeFigure: Figure = ({build, t, palette}) => {
  const left = stage(build, 0, 3);
  const right = stage(build, 1, 3);
  const slash = stage(build, 2, 3);
  const open = bob(t, 4, 1.5);
  return (
    <g>
      <path
        d={`M${76 - open} 70 L${40 - open} 100 L${76 - open} 130`}
        fill="none"
        stroke={palette.primary}
        strokeWidth={11}
        strokeLinecap="round"
        strokeLinejoin="round"
        opacity={left}
        {...draws(left)}
      />
      <path
        d={`M${124 + open} 70 L${160 + open} 100 L${124 + open} 130`}
        fill="none"
        stroke={palette.primary}
        strokeWidth={11}
        strokeLinecap="round"
        strokeLinejoin="round"
        opacity={right}
        {...draws(right)}
      />
      <path
        d="M112 62 L88 138"
        fill="none"
        stroke={palette.accent}
        strokeWidth={10}
        strokeLinecap="round"
        opacity={slash}
        {...draws(slash)}
      />
    </g>
  );
};

/** A commit line that forks, with a commit travelling the branch. */
export const BranchFigure: Figure = ({build, t, palette}) => {
  const trunk = stage(build, 0, 3);
  const fork = stage(build, 1, 3);
  const nodes = stage(build, 2, 3);
  const at = cycle(t, 0.42);
  return (
    <g>
      <path d="M70 42 L70 158" fill="none" stroke={palette.line} strokeWidth={6} strokeLinecap="round" opacity={0.55 * trunk} {...draws(trunk)} />
      <path d="M70 96 C70 74 140 82 140 60 L140 118" fill="none" stroke={palette.line} strokeWidth={6} strokeLinecap="round" opacity={0.55 * fork} {...draws(fork)} />
      <g opacity={nodes}>
        {[52, 96, 148].map((y, i) => (
          <circle key={i} cx={70} cy={y} r={11} fill={palette.primary} />
        ))}
        <circle cx={140} cy={62} r={11} fill={palette.accent} />
        <circle cx={140} cy={116} r={11} fill={palette.accent} />
        {/* The commit in flight along the branch. */}
        <circle cx={140} cy={62 + at * 54} r={6} fill={palette.soft} opacity={fadeTravel(at)} />
      </g>
    </g>
  );
};

/** A beetle whose legs actually scuttle. */
export const BugFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 2);
  const legs = stage(build, 1, 2);
  const step = Math.sin(t * 7);
  return (
    <g transform={`translate(0 ${bob(t, 2, 2.6)})`}>
      <g opacity={legs}>
        {[-1, 1].map((s) =>
          [0, 1, 2].map((i) => {
            const y = 86 + i * 22;
            const kick = step * (i % 2 === 0 ? 1 : -1) * 5;
            return (
              <path
                key={`${s}-${i}`}
                d={`M${100 + s * 26} ${y} L${100 + s * 52} ${y - 8 + kick} L${100 + s * 62} ${y + 6 + kick}`}
                fill="none"
                // `line`, not `ink`: these run over open background, where ink
                // is darker than the stage and the beetle loses its legs.
                stroke={palette.line}
                strokeWidth={5}
                strokeLinecap="round"
                strokeLinejoin="round"
                opacity={0.85}
              />
            );
          }),
        )}
      </g>
      <g transform={popAt(body, C, 110)} opacity={body}>
        <ellipse cx={C} cy={112} rx={30} ry={40} fill={palette.primary} />
        <path d="M100 72 A30 40 0 0 1 130 112 A30 40 0 0 1 100 152 Z" fill={palette.ink} opacity={0.14} />
        <line x1={C} y1={76} x2={C} y2={148} stroke={palette.ink} strokeWidth={3} opacity={0.3} />
        <circle cx={C} cy={62} r={17} fill={palette.primary} />
        <circle cx={93} cy={58} r={3.5} fill={palette.soft} />
        <circle cx={107} cy={58} r={3.5} fill={palette.soft} />
        {[-1, 1].map((s) => (
          <path key={s} d={`M${100 + s * 10} 48 L${100 + s * 20} 34`} stroke={palette.line} strokeWidth={4} strokeLinecap="round" fill="none" opacity={0.85} />
        ))}
      </g>
    </g>
  );
};

/** A carton whose lid lifts, over and over. */
export const PackageFigure: Figure = ({build, t, palette}) => {
  const box = stage(build, 0, 2);
  const lid = stage(build, 1, 2);
  const open = swing(t, 4.4, 0.18, 0.35);
  return (
    <g>
      <g transform={popAt(box, C, 124)} opacity={box}>
        <path d="M44 84 L100 108 L100 168 L44 144 Z" fill={palette.primary} opacity={0.85} />
        <path d="M156 84 L100 108 L100 168 L156 144 Z" fill={palette.primary} />
        <path d="M156 84 L100 108 L100 168 L156 144 Z" fill={palette.ink} opacity={0.16} />
      </g>
      <g opacity={lid} transform={`translate(0 ${-open * 22}) rotate(${-open * 12} 100 84)`}>
        <path d="M44 84 L100 60 L156 84 L100 108 Z" fill={palette.soft} />
        <path d="M100 60 L156 84 L100 108 Z" fill={palette.ink} opacity={0.1} />
      </g>
    </g>
  );
};

/** Two sockets and a plug that seats itself. */
export const ApiFigure: Figure = ({build, t, palette}) => {
  const socket = stage(build, 0, 2);
  const plug = stage(build, 1, 2);
  const seat = swing(t, 4, 0.2, 0.4);
  return (
    <g>
      <g transform={popAt(socket, 58, C)} opacity={socket}>
        <rect x={20} y={70} width={44} height={60} rx={10} fill={palette.soft} />
        <rect x={56} y={82} width={16} height={12} rx={4} fill={palette.ink} opacity={0.3} />
        <rect x={56} y={106} width={16} height={12} rx={4} fill={palette.ink} opacity={0.3} />
      </g>
      <g opacity={plug} transform={`translate(${(1 - seat) * 26} 0)`}>
        <rect x={92} y={70} width={88} height={60} rx={10} fill={palette.primary} />
        <rect x={92} y={70} width={88} height={60} rx={10} fill={palette.ink} opacity={0.12} />
        <rect x={76} y={82} width={20} height={12} rx={4} fill={palette.accent} />
        <rect x={76} y={106} width={20} height={12} rx={4} fill={palette.accent} />
      </g>
      {/* The contact spark, only at the moment it lands. */}
      {seat > 0.9 && (
        <circle cx={80} cy={C} r={10 + 8 * (seat - 0.9) * 10} fill="none" stroke={palette.accent} strokeWidth={3} opacity={(1 - seat) * 8} />
      )}
    </g>
  );
};

/** A run of stages with work flowing through them. */
export const PipelineFigure: Figure = ({build, t, palette}) => {
  const rail = stage(build, 0, 2);
  const boxes = stage(build, 1, 2);
  const at = cycle(t, 0.45);
  const xs = [46, 100, 154];
  return (
    <g>
      <line x1={30} y1={C} x2={170} y2={C} stroke={palette.line} strokeWidth={5} strokeLinecap="round" opacity={0.45 * rail} />
      {xs.map((x, i) => {
        const p = stage(boxes, i, xs.length, 0.5);
        return (
          <g key={i} transform={popAt(p, x, C)} opacity={p}>
            <rect x={x - 22} y={78} width={44} height={44} rx={10} fill={palette.soft} />
            <rect x={x - 22} y={78} width={44} height={44} rx={10} fill="none" stroke={palette.primary} strokeWidth={4} />
          </g>
        );
      })}
      <circle cx={30 + at * 140} cy={C} r={8} fill={palette.accent} opacity={fadeTravel(at)} />
    </g>
  );
};

/** A store that fills on a hit and empties again. */
export const CacheFigure: Figure = ({build, t, palette}) => {
  const shell = stage(build, 0, 2);
  const cells = stage(build, 1, 2);
  return (
    <g transform={popAt(shell, C, C)} opacity={shell}>
      <rect x={36} y={52} width={128} height={96} rx={14} fill={palette.soft} />
      <rect x={36} y={52} width={128} height={96} rx={14} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.18} />
      <g opacity={cells}>
        {ring(6, 0).map((_, i) => {
          const cx = 60 + (i % 3) * 40;
          const cy = 82 + Math.floor(i / 3) * 38;
          // Cells warm and cool on their own phases: a cache that fills all at
          // once is a bar chart.
          const hot = pulse(t, 1.6 + i * 0.4, i * 1.3);
          return (
            <rect
              key={i}
              x={cx - 15}
              y={cy - 13}
              width={30}
              height={26}
              rx={7}
              fill={hot > 0.6 ? palette.accent : palette.primary}
              opacity={0.35 + 0.65 * hot}
            />
          );
        })}
      </g>
    </g>
  );
};

/** A line of jobs that shuffles forward as the head is taken. */
export const QueueFigure: Figure = ({build, t, palette}) => {
  const rail = stage(build, 0, 2);
  const jobs = stage(build, 1, 2);
  const shift = cycle(t, 0.5);
  return (
    <g>
      <rect x={24} y={118} width={152} height={7} rx={3.5} fill={palette.line} opacity={0.4 * rail} />
      <g opacity={jobs}>
        {[0, 1, 2, 3].map((i) => {
          // The whole line slides one slot per cycle; the head fades out as it
          // leaves and the tail fades in behind it.
          const slot = i - shift;
          const x = 42 + slot * 38;
          const edge = slot < 0 ? 1 + slot : slot > 2.6 ? Math.max(0, 3.6 - slot) : 1;
          return (
            <rect
              key={i}
              x={x - 16}
              y={82}
              width={32}
              height={32}
              rx={8}
              fill={i === 0 ? palette.accent : palette.primary}
              opacity={Math.max(0, Math.min(1, edge))}
            />
          );
        })}
      </g>
    </g>
  );
};

/** A shipping container whose doors breathe open. */
export const ContainerFigure: Figure = ({build, t, palette}) => {
  const shell = stage(build, 0, 2);
  const doors = stage(build, 1, 2);
  const open = swing(t, 5, 0.2, 0.4) * 10;
  return (
    <g transform={`translate(0 ${bob(t, 2, 1.3)})`}>
      <g transform={popAt(shell, C, C)} opacity={shell}>
        <rect x={34} y={62} width={132} height={82} rx={8} fill={palette.primary} />
        <path d="M100 62 L166 62 L166 144 L100 144 Z" fill={palette.ink} opacity={0.14} />
        {[0, 1, 2, 3, 4, 5].map((i) => (
          <line key={i} x1={44 + i * 22} y1={66} x2={44 + i * 22} y2={140} stroke={palette.ink} strokeWidth={3} opacity={0.16} />
        ))}
      </g>
      <g opacity={doors}>
        <rect x={34 - open} y={70} width={12} height={66} rx={4} fill={palette.soft} />
        <rect x={154 + open} y={70} width={12} height={66} rx={4} fill={palette.soft} />
      </g>
    </g>
  );
};

/** A wall that turns packets away, and lets one through. */
export const FirewallFigure: Figure = ({build, t, palette}) => {
  const wall = stage(build, 0, 2);
  const traffic = stage(build, 1, 2);
  const at = cycle(t, 0.55);
  // One packet in three gets through; the rest bounce at the wall. Everything
  // bouncing reads as a broken link, everything passing reads as no wall.
  const passes = Math.floor(t * 0.55) % 3 === 0;
  const x = passes ? 20 + at * 160 : at < 0.5 ? 20 + at * 130 : 85 - (at - 0.5) * 130;
  return (
    <g>
      <g transform={popAt(wall, C, C)} opacity={wall}>
        {[0, 1, 2, 3, 4].map((r) =>
          [0, 1, 2].map((c) => (
            <rect
              key={`${r}-${c}`}
              // Every other course is offset by half a brick, which is the
              // only thing that makes a grid of rectangles read as a wall.
              x={72 + (r % 2 === 0 ? 0 : 14) + c * 28}
              y={40 + r * 26}
              width={24}
              height={22}
              rx={4}
              fill={palette.primary}
              opacity={0.92}
            />
          )),
        )}
      </g>
      <circle cx={x} cy={C} r={8} fill={passes ? palette.accent : palette.line} opacity={traffic * fadeTravel(at)} />
    </g>
  );
};

/** A die with pins, and a core that lights under load. */
export const CpuFigure: Figure = ({build, t, palette}) => {
  const die = stage(build, 0, 2);
  const pins = stage(build, 1, 2);
  const load = pulse(t, 2.8);
  return (
    <g>
      <g opacity={pins}>
        {[0, 1, 2, 3].map((i) =>
          [-1, 1].map((s) => (
            <g key={`${i}-${s}`}>
              <rect x={62 + i * 22} y={s < 0 ? 34 : 152} width={12} height={16} rx={3} fill={palette.line} opacity={0.6} />
              <rect x={s < 0 ? 34 : 152} y={62 + i * 22} width={16} height={12} rx={3} fill={palette.line} opacity={0.6} />
            </g>
          )),
        )}
      </g>
      <g transform={popAt(die, C, C)} opacity={die}>
        <rect x={50} y={50} width={100} height={100} rx={12} fill={palette.soft} />
        <rect x={72} y={72} width={56} height={56} rx={8} fill={palette.primary} />
        <rect x={72} y={72} width={56} height={56} rx={8} fill={palette.accent} opacity={load * 0.7} />
      </g>
    </g>
  );
};

/** A stick of memory with cells filling and freeing. */
export const MemoryFigure: Figure = ({build, t, palette}) => {
  const stick = stage(build, 0, 2);
  const cells = stage(build, 1, 2);
  return (
    <g transform={`translate(0 ${bob(t, 2, 1.4)})`}>
      <g transform={popAt(stick, C, C)} opacity={stick}>
        <rect x={30} y={72} width={140} height={56} rx={8} fill={palette.soft} />
        {[0, 1, 2, 3, 4, 5, 6].map((i) => (
          <rect key={i} x={38 + i * 19} y={128} width={11} height={10} fill={palette.line} opacity={0.5} />
        ))}
      </g>
      <g opacity={cells}>
        {[0, 1, 2, 3, 4].map((i) => {
          const held = pulse(t, 1.3 + i * 0.5, i * 2.1) > 0.5;
          return (
            <rect
              key={i}
              x={42 + i * 25}
              y={84}
              width={19}
              height={32}
              rx={5}
              fill={held ? palette.accent : palette.primary}
              opacity={held ? 1 : 0.4}
            />
          );
        })}
      </g>
    </g>
  );
};

/** Platters with a head that seeks across them. */
export const DiskFigure: Figure = ({build, t, palette}) => {
  const platter = stage(build, 0, 2);
  const head = stage(build, 1, 2);
  const seek = 0.5 + 0.5 * Math.sin(t * 0.9);
  return (
    <g>
      <g transform={popAt(platter, C, C)} opacity={platter}>
        <circle cx={C} cy={C} r={68} fill={palette.soft} />
        <circle cx={C} cy={C} r={68} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.15} />
        {[54, 40, 26].map((r, i) => (
          <circle key={i} cx={C} cy={C} r={r} fill="none" stroke={palette.ink} strokeWidth={2} opacity={0.12} />
        ))}
        <circle cx={C} cy={C} r={13} fill={palette.primary} />
      </g>
      <g opacity={head} transform={`rotate(${-24 + seek * 34} 168 44)`}>
        <line x1={168} y1={44} x2={112} y2={92} stroke={palette.primary} strokeWidth={9} strokeLinecap="round" />
        <circle cx={112} cy={92} r={7} fill={palette.accent} />
        <circle cx={168} cy={44} r={9} fill={palette.line} />
      </g>
    </g>
  );
};

/** A laptop whose lid rises and whose screen wakes. */
export const LaptopFigure: Figure = ({build, t, palette}) => {
  const base = stage(build, 0, 2);
  const lid = stage(build, 1, 2);
  const wake = swing(t, 5.2, 0.22, 0.5);
  return (
    <g>
      <g opacity={lid} transform={`translate(100 108) scale(1 ${0.24 + 0.76 * wake}) translate(-100 -108)`}>
        <rect x={44} y={44} width={112} height={70} rx={7} fill={palette.soft} />
        <rect x={52} y={52} width={96} height={54} rx={4} fill={palette.primary} />
        <rect x={52} y={52} width={96} height={54} rx={4} fill={palette.accent} opacity={0.35 * wake} />
      </g>
      <g transform={popAt(base, C, 132)} opacity={base}>
        <path d="M34 118 L166 118 L178 140 L22 140 Z" fill={palette.soft} />
        <path d="M22 140 L178 140 L178 148 L22 148 Z" fill={palette.ink} opacity={0.18} />
      </g>
    </g>
  );
};

/** A handset with a notification arriving. */
export const PhoneFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 2);
  const screen = stage(build, 1, 2);
  const ping = gesture(t, 3.6, 0.3);
  return (
    <g transform={`translate(0 ${bob(t, 2.5, 1.6)})`}>
      <g transform={popAt(body, C, C)} opacity={body}>
        <rect x={64} y={30} width={72} height={140} rx={14} fill={palette.soft} />
        <rect x={72} y={46} width={56} height={108} rx={5} fill={palette.primary} />
        <rect x={90} y={36} width={20} height={5} rx={2.5} fill={palette.ink} opacity={0.3} />
      </g>
      <g opacity={screen}>
        {ping >= 0 && (
          <rect
            x={76}
            y={54 - 12 * (1 - ease(ping))}
            width={48}
            height={22}
            rx={6}
            fill={palette.accent}
            opacity={Math.min(1, (1 - ping) * 3)}
          />
        )}
        <rect x={76} y={92} width={48} height={7} rx={3.5} fill={palette.soft} opacity={0.5} />
        <rect x={76} y={108} width={32} height={7} rx={3.5} fill={palette.soft} opacity={0.35} />
      </g>
    </g>
  );
};

/** A display on a stand, with a chart drawing itself on it. */
export const MonitorFigure: Figure = ({build, t, palette}) => {
  const panel = stage(build, 0, 3);
  const stand = stage(build, 1, 3);
  const line = stage(build, 2, 3);
  const redraw = swing(t, 6, 0.35, 0.45);
  return (
    <g>
      <g transform={popAt(panel, C, 92)} opacity={panel}>
        <rect x={28} y={42} width={144} height={100} rx={10} fill={palette.soft} />
        <rect x={38} y={52} width={124} height={80} rx={5} fill={palette.primary} />
      </g>
      <path
        d="M52 116 L82 92 L108 104 L148 68"
        fill="none"
        stroke={palette.accent}
        strokeWidth={5}
        strokeLinecap="round"
        strokeLinejoin="round"
        opacity={line}
        pathLength={1}
        strokeDasharray={1}
        strokeDashoffset={1 - redraw}
      />
      <g opacity={stand}>
        <rect x={90} y={142} width={20} height={20} fill={palette.line} opacity={0.6} />
        <rect x={64} y={160} width={72} height={9} rx={4.5} fill={palette.line} opacity={0.6} />
      </g>
    </g>
  );
};

/** A cloud with traffic moving in both directions. */
export const CloudFigure: Figure = ({build, t, palette}) => {
  const cloud = stage(build, 0, 2);
  const traffic = stage(build, 1, 2);
  // Two packets on opposite phases, so something is always in flight.
  const up = cycle(t, 0.62);
  const down = cycle(t, 0.62, 0.5);
  return (
    <g>
      <g transform={popAt(cloud, C, 78)} opacity={cloud}>
        <path d="M62 104 A26 26 0 0 1 68 54 A32 32 0 0 1 128 48 A24 24 0 0 1 140 104 Z" fill={palette.soft} />
        <path d="M100 48 A32 32 0 0 1 128 48 A24 24 0 0 1 140 104 L100 104 Z" fill={palette.ink} opacity={0.1} />
      </g>
      <g opacity={traffic}>
        {/* The lane the packets ride, so they are not floating in space. */}
        <line x1={78} y1={116} x2={78} y2={172} stroke={palette.line} strokeWidth={3} strokeDasharray="5 7" opacity={0.5} />
        <line x1={122} y1={116} x2={122} y2={172} stroke={palette.line} strokeWidth={3} strokeDasharray="5 7" opacity={0.5} />
        <rect x={68} y={166 - 50 * up} width={20} height={20} rx={5} fill={palette.accent} opacity={fadeTravel(up)} />
        <rect x={112} y={116 + 50 * down} width={20} height={20} rx={5} fill={palette.primary} opacity={fadeTravel(down)} />
      </g>
    </g>
  );
};

/** A hub with packets running out along its spokes. */
export const NetworkFigure: Figure = ({build, t, palette}) => {
  const links = stage(build, 0, 3);
  const hub = stage(build, 1, 3);
  const leaves = stage(build, 2, 3);
  const nodes = ring(5, 62, C, C, -Math.PI / 2);

  return (
    <g>
      {nodes.map((p, i) => {
        const lp = ease(stage(links, i, nodes.length, 0.7));
        const travel = cycle(t, 0.55, i / nodes.length);
        return (
          <g key={i}>
            <line
              x1={C}
              y1={C}
              x2={C + (p.x - C) * lp}
              y2={C + (p.y - C) * lp}
              stroke={palette.line}
              strokeWidth={3.5}
              strokeLinecap="round"
              opacity={0.55}
            />
            {lp >= 1 && (
              <circle
                cx={C + (p.x - C) * travel}
                cy={C + (p.y - C) * travel}
                r={5}
                fill={palette.accent}
                opacity={fadeTravel(travel)}
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
      <g transform={popAt(hub, C, C)} opacity={hub}>
        <circle cx={C} cy={C} r={24} fill={palette.accent} />
        <circle cx={C} cy={C} r={24} fill="none" stroke={palette.soft} strokeWidth={4} opacity={0.5} />
      </g>
    </g>
  );
};

/** A padlock whose key turns in it, holds, and releases. */
export const LockFigure: Figure = ({build, t, palette}) => {
  const lock = stage(build, 0, 2);
  const key = stage(build, 1, 2);
  // The key seats itself, then turns a quarter and holds — a key that spins
  // forever is a fidget, not a mechanism.
  const turn = swing(t, 4, 0.15, 0.45) * 90;

  return (
    <g>
      <g transform={popAt(lock, C, 116)} opacity={lock}>
        <path d="M76 92 L76 74 A24 24 0 0 1 124 74 L124 92" fill="none" stroke={palette.soft} strokeWidth={12} strokeLinecap="round" />
        <rect x={58} y={92} width={84} height={68} rx={12} fill={palette.primary} />
        <path d="M100 92 L142 92 L142 160 L100 160 Z" fill={palette.ink} opacity={0.15} />
      </g>
      <g opacity={key} transform={`rotate(${turn} 100 122)`}>
        <circle cx={C} cy={122} r={13} fill="none" stroke={palette.accent} strokeWidth={7} />
        <rect x={97} y={130} width={6} height={26} rx={3} fill={palette.accent} />
        <rect x={97} y={144} width={16} height={6} rx={3} fill={palette.accent} />
      </g>
    </g>
  );
};
