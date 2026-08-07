// Building things without writing code: the assistant, the blocks, the board
// you wire together and the thing you publish at the end.
//
// This module exists for the same reason `learn` did. The drawer was built to
// explain *systems* — there is a figure for a load balancer, a cache and a
// firewall — and the courses this renders are mostly about somebody assembling
// a working product out of tools that already exist. A narrator saying "drag
// the block onto the canvas" had `puzzle` and a narrator saying "now publish
// it" had `rocket`, which is a metaphor standing in for the literal thing on
// the learner's screen.
//
// Same rule as the rest of the drawer: whatever the object *does* is the thing
// that moves. The block snaps, the plug seats, the belt carries, the caret
// types and the deploy button fires once and resets.

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

/** The assistant: a machine head whose antenna thinks and whose eyes blink. */
export const RobotFigure: Figure = ({build, t, palette}) => {
  const head = stage(build, 0, 3);
  const face = stage(build, 1, 3);
  const ant = stage(build, 2, 3);
  // Blinking is the whole difference between a robot and a rounded rectangle.
  // A long open period with a fast close reads as an eye; a sine reads as a
  // lamp being dimmed.
  const blink = gesture(t, 4.1, 0.07);
  const lid = blink >= 0 ? Math.sin(blink * Math.PI) : 0;
  const think = pulse(t, 2.6);
  const sway = bob(t, 2, 1.3);
  return (
    <g transform={`translate(0 ${sway})`}>
      <g opacity={ant}>
        <line
          x1={C}
          y1={62}
          x2={C}
          y2={38}
          stroke={palette.line}
          strokeWidth={5}
          strokeLinecap="round"
          {...draws(ant)}
        />
        <circle cx={C} cy={32} r={9} fill={palette.accent} opacity={0.5 + 0.5 * think} />
        <circle
          cx={C}
          cy={32}
          r={9 + 7 * think}
          fill="none"
          stroke={palette.accent}
          strokeWidth={2}
          opacity={(1 - think) * 0.5}
        />
      </g>
      <g transform={popAt(head, C, 112)} opacity={head}>
        <rect x={40} y={62} width={120} height={98} rx={26} fill={palette.soft} />
        <rect x={40} y={62} width={120} height={98} rx={26} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.16} />
        {/* Ears, which are what stop the head reading as a phone. */}
        <rect x={28} y={96} width={12} height={30} rx={6} fill={palette.primary} />
        <rect x={160} y={96} width={12} height={30} rx={6} fill={palette.primary} />
      </g>
      <g opacity={face}>
        <rect x={62} y={88} width={76} height={44} rx={14} fill={palette.ink} opacity={0.72} />
        {[80, 120].map((x) => (
          <ellipse
            key={x}
            cx={x}
            cy={110}
            rx={9}
            ry={9 * (1 - lid)}
            fill={palette.accent}
          />
        ))}
        <rect x={84} y={142} width={32} height={5} rx={2.5} fill={palette.primary} opacity={0.6} />
      </g>
    </g>
  );
};

/** The "make it for me" gesture: a wand that throws sparks on a beat. */
export const WandFigure: Figure = ({build, t, palette}) => {
  const stick = stage(build, 0, 2);
  const tip = stage(build, 1, 2);
  // The flourish happens and then stops. A wand sparkling continuously is a
  // sparkler; one that fires on a beat is somebody casting with it.
  const cast = gesture(t, 3.0, 0.4);
  const flick = cast >= 0 ? Math.sin(cast * Math.PI) : 0;
  const sparks = [
    {a: -0.5, d: 34},
    {a: -1.3, d: 46},
    {a: 0.15, d: 40},
    {a: -2.0, d: 30},
  ];
  return (
    <g transform={`rotate(${-8 + flick * 14} 100 130)`}>
      <line
        x1={62}
        y1={162}
        x2={132}
        y2={78}
        stroke={palette.primary}
        strokeWidth={13}
        strokeLinecap="round"
        opacity={stick}
        {...draws(stick)}
      />
      <line
        x1={62}
        y1={162}
        x2={86}
        y2={133}
        stroke={palette.ink}
        strokeWidth={13}
        strokeLinecap="round"
        opacity={0.28 * stick}
      />
      <g opacity={tip}>
        <circle cx={138} cy={72} r={11} fill={palette.accent} />
        {sparks.map((s, i) => {
          const p = flick > 0 ? Math.max(0, Math.min(1, (flick - i * 0.08) / 0.6)) : 0;
          if (p <= 0.01) return null;
          const x = 138 + Math.cos(s.a) * s.d * ease(p);
          const y = 72 + Math.sin(s.a) * s.d * ease(p);
          const r = 5 * (1 - p) + 1.5;
          return <circle key={i} cx={x} cy={y} r={r} fill={palette.accent} opacity={1 - p} />;
        })}
      </g>
    </g>
  );
};

/** Snap-together building: the block above drops and clicks into the stack. */
export const BlocksFigure: Figure = ({build, t, palette}) => {
  const base = stage(build, 0, 2);
  const rest = stage(build, 1, 2);
  // The top block is carried in, seats, and holds — the literal gesture of
  // every no-code builder there is.
  const drop = swing(t, 4.4, 0.22, 0.5);
  const y = 54 - (1 - drop) * 44;
  const seated = drop > 0.96;
  const Stud = ({x, top}: {x: number; top: number}) => (
    <rect x={x} y={top} width={20} height={9} rx={4} fill={palette.ink} opacity={0.2} />
  );
  return (
    <g>
      <g opacity={base}>
        <rect x={38} y={128} width={124} height={40} rx={9} fill={palette.primary} />
        <Stud x={58} top={122} />
        <Stud x={122} top={122} />
      </g>
      <g opacity={rest}>
        <rect x={54} y={84} width={92} height={40} rx={9} fill={palette.soft} />
        <rect x={54} y={84} width={92} height={40} rx={9} fill="none" stroke={palette.ink} strokeWidth={2} opacity={0.16} />
        <Stud x={70} top={78} />
        <Stud x={110} top={78} />
      </g>
      <g opacity={rest * (drop > 0.02 ? 1 : 0.35)}>
        <rect x={66} y={y} width={68} height={38} rx={9} fill={palette.accent} />
        <rect x={82} y={y - 6} width={20} height={9} rx={4} fill={palette.ink} opacity={0.2} />
      </g>
      {/* The click: a flash across the seam, only in the frames it lands. */}
      {seated && (
        <rect x={62} y={80} width={76} height={4} rx={2} fill={palette.accent} opacity={0.7} />
      )}
    </g>
  );
};

/** Connecting one tool to another: a plug that travels and seats in a socket. */
export const PlugFigure: Figure = ({build, t, palette}) => {
  const socket = stage(build, 0, 2);
  const plug = stage(build, 1, 2);
  const seat = swing(t, 3.8, 0.2, 0.42);
  const x = 40 - (1 - seat) * 28;
  return (
    <g>
      <g transform={popAt(socket, 140, C)} opacity={socket}>
        <rect x={122} y={62} width={48} height={76} rx={12} fill={palette.soft} />
        <rect x={122} y={62} width={48} height={76} rx={12} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.18} />
        <rect x={112} y={80} width={14} height={12} rx={4} fill={palette.ink} opacity={0.35} />
        <rect x={112} y={108} width={14} height={12} rx={4} fill={palette.ink} opacity={0.35} />
      </g>
      <g opacity={plug} transform={`translate(${x} 0)`}>
        <rect x={30} y={70} width={54} height={60} rx={12} fill={palette.primary} />
        <rect x={82} y={80} width={30} height={12} rx={4} fill={palette.accent} />
        <rect x={82} y={108} width={30} height={12} rx={4} fill={palette.accent} />
        <line x1={12} y1={C} x2={32} y2={C} stroke={palette.line} strokeWidth={6} strokeLinecap="round" opacity={0.6} />
      </g>
      {/* Contact. Two tools that are merely adjacent are not integrated. */}
      {seat > 0.95 && (
        <circle cx={C + 22} cy={C} r={16 + 22 * (seat - 0.95) * 20} fill="none" stroke={palette.accent} strokeWidth={3} opacity={(1 - seat) * 18} />
      )}
    </g>
  );
};

/** The control panel of the thing you built: an arc and two bars, live. */
export const DashboardFigure: Figure = ({build, t, palette}) => {
  const panel = stage(build, 0, 3);
  const dial = stage(build, 1, 3);
  const bars = stage(build, 2, 3);
  // The needle drifts inside a band rather than sweeping the whole arc: a
  // dashboard whose gauge swings end to end is broken, not busy.
  const a = Math.PI * (0.18 + 0.34 * pulse(t, 0.9));
  const nx = 74 + Math.cos(Math.PI - a) * 30;
  const ny = 104 - Math.sin(Math.PI - a) * 30;
  return (
    <g>
      <g transform={popAt(panel, C, C)} opacity={panel}>
        <rect x={26} y={50} width={148} height={104} rx={14} fill={palette.soft} />
        <rect x={26} y={50} width={148} height={104} rx={14} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.16} />
      </g>
      <g opacity={dial}>
        <path d="M44 104 A30 30 0 0 1 104 104" fill="none" stroke={palette.ink} strokeWidth={7} strokeLinecap="round" opacity={0.2} />
        <path d="M44 104 A30 30 0 0 1 104 104" fill="none" stroke={palette.primary} strokeWidth={7} strokeLinecap="round" {...draws(dial * 0.7)} />
        <line x1={74} y1={104} x2={nx} y2={ny} stroke={palette.accent} strokeWidth={4} strokeLinecap="round" />
        <circle cx={74} cy={104} r={5} fill={palette.accent} />
      </g>
      <g opacity={bars}>
        {[0, 1, 2].map((i) => {
          // Each bar breathes on its own phase, so the panel reads as three
          // readings rather than one animation.
          const h = 14 + 22 * pulse(t, 1.1 + i * 0.5, i * 2.1);
          return (
            <rect key={i} x={120 + i * 17} y={130 - h} width={11} height={h} rx={4} fill={i === 1 ? palette.accent : palette.primary} opacity={i === 1 ? 1 : 0.65} />
          );
        })}
        <rect x={118} y={134} width={48} height={3} rx={1.5} fill={palette.ink} opacity={0.2} />
      </g>
    </g>
  );
};

/** Shipping it: a button that fires once and sends the build up and away. */
export const DeployFigure: Figure = ({build, t, palette}) => {
  const pad = stage(build, 0, 2);
  const btn = stage(build, 1, 2);
  // Most of the cycle is spent in flight rather than at rest. Parked, the shot
  // is a yellow pill on a grey slab and reads as nothing in particular; the
  // payload leaving is the only frame that says "deploy", so it is the frame
  // the figure should mostly be on.
  const fire = gesture(t, 3.6, 0.78);
  const press = fire >= 0 && fire < 0.1 ? 1 : 0;
  // The payload leaves after the press, not with it.
  const flight = fire >= 0.1 ? ease((fire - 0.1) / 0.9) : 0;
  return (
    <g>
      <g opacity={pad}>
        <rect x={44} y={132} width={112} height={34} rx={10} fill={palette.soft} />
        <rect x={44} y={132} width={112} height={34} rx={10} fill="none" stroke={palette.ink} strokeWidth={2} opacity={0.16} />
      </g>
      <g opacity={btn} transform={`translate(0 ${press * 4})`}>
        <rect x={62} y={112} width={76} height={28} rx={14} fill={palette.accent} />
        <rect x={78} y={123} width={44} height={6} rx={3} fill={palette.ink} opacity={0.35} />
      </g>
      {flight > 0.01 && (
        <g opacity={Math.min(1, (1 - flight) * 2.2)} transform={`translate(0 ${-88 * flight})`}>
          <path d="M100 60 L120 96 L80 96 Z" fill={palette.primary} />
          <rect x={90} y={96} width={20} height={16} rx={5} fill={palette.primary} opacity={0.75} />
          <circle cx={C} cy={118} r={7 * (1 - flight)} fill={palette.accent} opacity={0.8} />
        </g>
      )}
    </g>
  );
};

/** The finished product: an app tile with a live badge that pops. */
export const AppFigure: Figure = ({build, t, palette}) => {
  const tile = stage(build, 0, 3);
  const glyph = stage(build, 1, 3);
  const badge = stage(build, 2, 3);
  const ping = gesture(t, 3.4, 0.3);
  const pop = ping >= 0 ? Math.sin(ping * Math.PI) : 0;
  const float = bob(t, 3, 1.4);
  return (
    <g transform={`translate(0 ${float})`}>
      <g transform={popAt(tile, C, C)} opacity={tile}>
        <rect x={48} y={48} width={104} height={104} rx={26} fill={palette.primary} />
        <rect x={48} y={48} width={104} height={104} rx={26} fill={palette.ink} opacity={0.1} />
      </g>
      <g opacity={glyph}>
        {/* A rounded chevron: the generic mark every app icon has, drawn so it
            is unmistakably an icon rather than a window. */}
        <path
          d="M78 86 L100 68 L122 86"
          fill="none"
          stroke={palette.soft}
          strokeWidth={11}
          strokeLinecap="round"
          strokeLinejoin="round"
          {...draws(glyph)}
        />
        <rect x={92} y={92} width={16} height={44} rx={8} fill={palette.soft} opacity={ease(glyph)} />
      </g>
      <g opacity={badge}>
        <circle cx={150} cy={54} r={16 + 4 * pop} fill={palette.accent} />
        <circle cx={150} cy={54} r={16 + 18 * pop} fill="none" stroke={palette.accent} strokeWidth={2} opacity={(1 - pop) * 0.6} />
        <rect x={144} y={46} width={12} height={12} rx={3} fill={palette.ink} opacity={0.4} />
      </g>
    </g>
  );
};

/** Where no-code data actually lives: a grid with a row filling and a total. */
export const SpreadsheetFigure: Figure = ({build, t, palette}) => {
  const sheet = stage(build, 0, 3);
  const grid = stage(build, 1, 3);
  const total = stage(build, 2, 3);
  // One row fills left to right and the total lands after it, which is the one
  // thing a spreadsheet does that a table does not.
  const fill = swing(t, 5.2, 0.34, 0.4);
  const cols = [50, 88, 126];
  const rows = [78, 100, 122];
  return (
    <g>
      <g transform={popAt(sheet, C, C)} opacity={sheet}>
        <rect x={34} y={52} width={132} height={100} rx={10} fill={palette.soft} />
        <rect x={34} y={52} width={132} height={100} rx={10} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.16} />
        <rect x={34} y={52} width={132} height={18} rx={10} fill={palette.primary} />
        <rect x={34} y={62} width={132} height={8} fill={palette.primary} />
      </g>
      <g opacity={grid}>
        {rows.map((y) =>
          cols.map((x) => {
            const lit = y === rows[1] ? Math.max(0, Math.min(1, (fill - (x - 50) / 160) * 3)) : 0;
            return (
              <rect
                key={`${x}-${y}`}
                x={x}
                y={y}
                width={30}
                height={12}
                rx={3}
                fill={lit > 0.1 ? palette.accent : palette.ink}
                opacity={lit > 0.1 ? 0.35 + 0.65 * lit : 0.16}
              />
            );
          }),
        )}
      </g>
      <g opacity={total * (fill > 0.85 ? 1 : 0.25)}>
        <rect x={126} y={122} width={30} height={12} rx={3} fill={palette.accent} opacity={fill > 0.85 ? 1 : 0.3} />
        <line x1={126} y1={118} x2={156} y2={118} stroke={palette.accent} strokeWidth={2} opacity={0.7} />
      </g>
    </g>
  );
};

/** It runs without you: a belt carrying finished work past a driving gear. */
export const AutomationFigure: Figure = ({build, t, palette}) => {
  const belt = stage(build, 0, 2);
  const load = stage(build, 1, 2);
  const roll = cycle(t, 0.32);
  const spin = t * 90;
  return (
    <g>
      <g opacity={belt}>
        <rect x={26} y={112} width={148} height={16} rx={8} fill={palette.soft} />
        <rect x={26} y={112} width={148} height={16} rx={8} fill="none" stroke={palette.ink} strokeWidth={2} opacity={0.16} />
        {/* Tread marks, so the belt is visibly moving even between parcels. */}
        {[0, 1, 2, 3, 4, 5].map((i) => {
          const x = 30 + (((i * 26 + roll * 26) % 150) | 0);
          return <rect key={i} x={x} y={117} width={9} height={6} rx={3} fill={palette.ink} opacity={0.18} />;
        })}
        {/* The drive. Teeth rather than a cross through a ring — a circle with
            two lines across it reads as a "no entry" sign, which is the exact
            opposite of what a belt that keeps running should say. */}
        <g transform={`rotate(${spin} 148 146)`}>
          {ring(8, 17, 148, 146).map((p, i) => (
            <rect key={i} x={p.x - 4} y={p.y - 4} width={8} height={8} rx={2} fill={palette.primary} transform={`rotate(${(i * 360) / 8} ${p.x} ${p.y})`} />
          ))}
          <circle cx={148} cy={146} r={13} fill={palette.primary} />
          <circle cx={148} cy={146} r={5} fill={palette.ink} opacity={0.4} />
        </g>
      </g>
      <g opacity={load}>
        {[0, 1, 2].map((i) => {
          const p = ((roll + i / 3) % 1 + 1) % 1;
          const x = 26 + p * 148;
          return (
            <rect
              key={i}
              x={x - 15}
              y={80}
              width={30}
              height={30}
              rx={7}
              fill={i === 1 ? palette.accent : palette.primary}
              opacity={fadeTravel(p)}
            />
          );
        })}
      </g>
    </g>
  );
};

/** The shape of what you are building: pages branching from one root. */
export const SitemapFigure: Figure = ({build, t, palette}) => {
  const root = stage(build, 0, 3);
  const wires = stage(build, 1, 3);
  const kids = stage(build, 2, 3);
  const xs = [52, 100, 148];
  // A pulse walks out from the root along one branch at a time, so the tree
  // reads as a structure something travels rather than as an org chart.
  const at = cycle(t, 0.4);
  const lane = Math.floor(at * 3) % 3;
  const along = (at * 3) % 1;
  return (
    <g>
      <g transform={popAt(root, C, 58)} opacity={root}>
        <rect x={72} y={38} width={56} height={40} rx={9} fill={palette.primary} />
        <rect x={80} y={50} width={40} height={5} rx={2.5} fill={palette.soft} opacity={0.8} />
        <rect x={80} y={62} width={26} height={5} rx={2.5} fill={palette.soft} opacity={0.5} />
      </g>
      <g opacity={wires}>
        {xs.map((x, i) => (
          <path
            key={i}
            d={`M100 78 V100 H${x} V126`}
            fill="none"
            stroke={palette.line}
            strokeWidth={3}
            strokeLinecap="round"
            strokeLinejoin="round"
            opacity={0.45}
            {...draws(wires)}
          />
        ))}
        <circle
          cx={lane === 1 ? 100 : xs[lane]}
          cy={78 + along * 48}
          r={5}
          fill={palette.accent}
          opacity={fadeTravel(along)}
        />
      </g>
      <g opacity={kids}>
        {xs.map((x, i) => (
          <g key={i} transform={popAt(stage(kids, i, 3, 0.5), x, 142)}>
            <rect x={x - 22} y={126} width={44} height={34} rx={8} fill={palette.soft} />
            <rect x={x - 22} y={126} width={44} height={34} rx={8} fill="none" stroke={palette.ink} strokeWidth={2} opacity={0.16} />
            <rect x={x - 13} y={138} width={26} height={4} rx={2} fill={palette.ink} opacity={0.3} />
            <rect x={x - 13} y={147} width={16} height={4} rx={2} fill={palette.ink} opacity={0.2} />
          </g>
        ))}
      </g>
    </g>
  );
};

/** The instruction itself: a field being typed into, and then sent. */
export const PromptFigure: Figure = ({build, t, palette}) => {
  const field = stage(build, 0, 2);
  const send = stage(build, 1, 2);
  // Type, hold, send, clear — the loop the learner is being taught, drawn.
  const typed = swing(t, 4.6, 0.4, 0.28);
  const fired = typed > 0.99;
  const caretOn = pulse(t, 6) > 0.5;
  const width = 88 * ease(typed);
  return (
    <g>
      <g transform={popAt(field, C, C)} opacity={field}>
        <rect x={28} y={74} width={144} height={52} rx={14} fill={palette.soft} />
        <rect x={28} y={74} width={144} height={52} rx={14} fill="none" stroke={fired ? palette.accent : palette.ink} strokeWidth={3} opacity={fired ? 0.9 : 0.18} />
      </g>
      <g opacity={field}>
        <rect x={44} y={95} width={width} height={9} rx={4.5} fill={palette.primary} />
        {caretOn && !fired && (
          <rect x={46 + width} y={90} width={3} height={19} rx={1.5} fill={palette.accent} />
        )}
      </g>
      <g opacity={send}>
        <circle cx={150} cy={100} r={15} fill={fired ? palette.accent : palette.primary} opacity={fired ? 1 : 0.5} />
        <path d="M144 100 h11 M151 95 l5 5 -5 5" fill="none" stroke={palette.soft} strokeWidth={3} strokeLinecap="round" strokeLinejoin="round" />
      </g>
      {/* The send, as a ring leaving the button. */}
      {fired && (
        <circle cx={150} cy={100} r={15 + 30 * (typed - 0.99) * 100} fill="none" stroke={palette.accent} strokeWidth={2} opacity={(1 - typed) * 90} />
      )}
    </g>
  );
};

/** The plan before the build: a sheet whose outline draws itself with ticks. */
export const BlueprintFigure: Figure = ({build, t, palette}) => {
  const sheet = stage(build, 0, 3);
  const plan = stage(build, 1, 3);
  const dims = stage(build, 2, 3);
  // The plan redraws on a loop, which is what a plan does: it is never
  // finished, it is revised. Kept short because `swing` spends its first
  // quarter at rest — at a slower period the outline is simply absent for the
  // first second and a half, which on a three-second shot is most of it.
  const drawn = swing(t, 4.2, 0.34, 0.4);
  return (
    <g>
      <g transform={popAt(sheet, C, C)} opacity={sheet}>
        <rect x={30} y={46} width={140} height={112} rx={10} fill={palette.primary} />
        <rect x={30} y={46} width={140} height={112} rx={10} fill={palette.ink} opacity={0.28} />
        {[62, 82, 102, 122, 142].map((y) => (
          <line key={y} x1={30} y1={y} x2={170} y2={y} stroke={palette.soft} strokeWidth={1} opacity={0.12} />
        ))}
      </g>
      <g opacity={plan}>
        <path
          d="M56 128 V78 H108 V104 H144 V128 Z"
          fill="none"
          stroke={palette.soft}
          strokeWidth={4}
          strokeLinejoin="round"
          {...draws(drawn)}
        />
      </g>
      <g opacity={dims * ease(drawn)}>
        <line x1={56} y1={140} x2={144} y2={140} stroke={palette.accent} strokeWidth={2} />
        <line x1={56} y1={135} x2={56} y2={145} stroke={palette.accent} strokeWidth={2} />
        <line x1={144} y1={135} x2={144} y2={145} stroke={palette.accent} strokeWidth={2} />
      </g>
    </g>
  );
};

/** The records behind it: stacked slabs with a scan running down them. */
export const DatasetFigure: Figure = ({build, t, palette}) => {
  const slabs = [0, 1, 2, 3];
  const at = cycle(t, 0.4);
  const lit = Math.floor(at * slabs.length);
  return (
    <g>
      {slabs.map((i) => {
        const p = stage(build, i, slabs.length, 0.55);
        const y = 56 + i * 28;
        const on = i === lit;
        return (
          <g key={i} transform={popAt(p, C, y + 11)} opacity={p}>
            <rect x={40} y={y} width={120} height={22} rx={6} fill={on ? palette.accent : palette.soft} opacity={on ? 0.9 : 1} />
            <rect x={40} y={y} width={120} height={22} rx={6} fill="none" stroke={palette.ink} strokeWidth={2} opacity={0.14} />
            {/* Field marks, so a slab reads as a record and not a bar. */}
            <rect x={52} y={y + 8} width={26} height={6} rx={3} fill={palette.ink} opacity={on ? 0.4 : 0.22} />
            <rect x={86} y={y + 8} width={38} height={6} rx={3} fill={palette.ink} opacity={on ? 0.3 : 0.16} />
            <rect x={130} y={y + 8} width={16} height={6} rx={3} fill={palette.primary} opacity={on ? 0.9 : 0.4} />
          </g>
        );
      })}
    </g>
  );
};

/** Constraints that keep it on the road: rails a traveller cannot leave. */
export const GuardrailFigure: Figure = ({build, t, palette}) => {
  const rails = stage(build, 0, 2);
  const car = stage(build, 1, 2);
  const at = cycle(t, 0.34);
  const x = 34 + at * 132;
  // It drifts towards a rail and is pushed back, which is the whole point: the
  // rail is doing something, not decorating the road.
  const drift = Math.sin(at * Math.PI * 4);
  const y = C + drift * 17;
  const nudged = Math.abs(drift) > 0.86;
  return (
    <g>
      <g opacity={rails}>
        {[70, 130].map((ry) => (
          <g key={ry}>
            <line x1={26} y1={ry} x2={174} y2={ry} stroke={palette.primary} strokeWidth={6} strokeLinecap="round" {...draws(rails)} />
            {/* Uprights, drawn away from the road on both sides. A negative
                height is not a rect that points the other way — it is a rect
                the renderer drops, which is how the lower rail lost its posts. */}
            {[44, 100, 156].map((px) => (
              <rect key={px} x={px - 3} y={ry === 70 ? ry - 16 : ry + 2} width={6} height={16} rx={3} fill={palette.primary} opacity={0.55} />
            ))}
          </g>
        ))}
      </g>
      <g opacity={car}>
        <rect x={x - 16} y={y - 11} width={32} height={22} rx={8} fill={palette.accent} />
        <rect x={x - 6} y={y - 5} width={14} height={10} rx={4} fill={palette.ink} opacity={0.3} />
      </g>
      {/* The correction, only in the frames the rail is actually doing it. */}
      {nudged && (
        <circle cx={x} cy={drift > 0 ? 130 : 70} r={13} fill="none" stroke={palette.accent} strokeWidth={3} opacity={0.55} />
      )}
    </g>
  );
};
