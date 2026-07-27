// The web, and the things people do on it.
//
// These lean harder on gesture than the infrastructure set does, because that
// is what they are about: a cursor clicks, a bell rings, an envelope opens, a
// cart rolls. Where an object's whole meaning is an action somebody takes, the
// idle motion is that action on a loop rather than a float.

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
  stage,
  swing,
  type Figure,
} from './kit';

/** A window with a chrome bar and content that loads in. */
export const BrowserFigure: Figure = ({build, t, palette}) => {
  const frame = stage(build, 0, 3);
  const bar = stage(build, 1, 3);
  const body = stage(build, 2, 3);
  const load = swing(t, 5.4, 0.3, 0.45);
  return (
    <g transform={popAt(frame, C, C)} opacity={frame}>
      <rect x={24} y={44} width={152} height={112} rx={11} fill={palette.soft} />
      <g opacity={bar}>
        <path d="M24 55 A11 11 0 0 1 35 44 L165 44 A11 11 0 0 1 176 55 L176 72 L24 72 Z" fill={palette.primary} />
        {[0, 1, 2].map((i) => (
          <circle key={i} cx={40 + i * 14} cy={58} r={4.5} fill={palette.soft} opacity={0.75} />
        ))}
        <rect x={78} y={51} width={84} height={14} rx={7} fill={palette.soft} opacity={0.5} />
      </g>
      <g opacity={body}>
        <rect x={38} y={86} width={124 * load} height={12} rx={6} fill={palette.accent} />
        <rect x={38} y={110} width={100} height={9} rx={4.5} fill={palette.ink} opacity={0.18} />
        <rect x={38} y={128} width={72} height={9} rx={4.5} fill={palette.ink} opacity={0.14} />
      </g>
    </g>
  );
};

/** An arrow that clicks, with the ripple a click leaves. */
export const CursorFigure: Figure = ({build, t, palette}) => {
  const arrow = stage(build, 0, 1);
  const click = gesture(t, 2.6, 0.35);
  const press = click >= 0 && click < 0.25 ? 1 : 0;
  return (
    <g>
      {click >= 0 && (
        <circle
          cx={C}
          cy={C}
          r={16 + 40 * ease(click)}
          fill="none"
          stroke={palette.accent}
          strokeWidth={4}
          opacity={0.55 * (1 - click)}
        />
      )}
      <g transform={`${popAt(arrow, C, C)} translate(${press * 3} ${press * 3})`} opacity={arrow}>
        <path d="M78 62 L78 148 L100 126 L116 158 L134 148 L118 118 L146 116 Z" fill={palette.soft} />
        <path d="M78 62 L78 148 L100 126 L116 158 L134 148 L118 118 L146 116 Z" fill="none" stroke={palette.primary} strokeWidth={6} strokeLinejoin="round" />
      </g>
    </g>
  );
};

/** A lens that sweeps, with a result landing under it. */
export const SearchFigure: Figure = ({build, t, palette}) => {
  const glass = stage(build, 0, 2);
  const rows = stage(build, 1, 2);
  const sweep = Math.sin(t * 1.1);
  return (
    <g>
      <g opacity={rows}>
        {[0, 1, 2].map((i) => (
          <rect key={i} x={44} y={126 + i * 18} width={i === 0 ? 96 : 68} height={9} rx={4.5} fill={palette.ink} opacity={i === 0 ? 0.28 : 0.15} />
        ))}
      </g>
      <g transform={`${popAt(glass, 92, 80)} translate(${sweep * 12} ${sweep * 5})`} opacity={glass}>
        <circle cx={92} cy={80} r={38} fill={palette.soft} opacity={0.5} />
        <circle cx={92} cy={80} r={38} fill="none" stroke={palette.primary} strokeWidth={10} />
        <line x1={120} y1={108} x2={148} y2={136} stroke={palette.primary} strokeWidth={12} strokeLinecap="round" />
        <path d="M74 66 A24 24 0 0 1 92 58" fill="none" stroke={palette.accent} strokeWidth={5} strokeLinecap="round" opacity={0.8} />
      </g>
    </g>
  );
};

/** Fields that fill one after another, then a button that confirms. */
export const FormFigure: Figure = ({build, t, palette}) => {
  const card = stage(build, 0, 2);
  const fields = stage(build, 1, 2);
  const run = cycle(t, 0.28);
  return (
    <g transform={popAt(card, C, C)} opacity={card}>
      <rect x={40} y={34} width={120} height={132} rx={12} fill={palette.soft} />
      <g opacity={fields}>
        {[0, 1, 2].map((i) => {
          // Each field fills over its own third of the loop, so the form is
          // always being filled rather than always full.
          const fill = Math.max(0, Math.min(1, run * 3.6 - i));
          return (
            <g key={i}>
              <rect x={54} y={52 + i * 30} width={92} height={18} rx={5} fill={palette.ink} opacity={0.12} />
              <rect x={54} y={52 + i * 30} width={92 * fill} height={18} rx={5} fill={palette.primary} opacity={0.8} />
            </g>
          );
        })}
        <rect x={54} y={142} width={92} height={20} rx={10} fill={run > 0.85 ? palette.accent : palette.line} opacity={run > 0.85 ? 1 : 0.5} />
      </g>
    </g>
  );
};

/** A basket that rolls, with an item dropping into it. */
export const CartFigure: Figure = ({build, t, palette}) => {
  const basket = stage(build, 0, 2);
  const wheels = stage(build, 1, 2);
  const drop = cycle(t, 0.5);
  const roll = bob(t, 4, 1.2);
  return (
    <g transform={`translate(${roll} 0)`}>
      {/* The item, falling in before the basket reads as full. */}
      <rect
        x={92}
        y={24 + drop * 52}
        width={22}
        height={22}
        rx={5}
        fill={palette.accent}
        opacity={drop < 0.8 ? fadeTravel(drop * 1.25) : 0}
      />
      <g transform={popAt(basket, C, 104)} opacity={basket}>
        <path d="M52 78 L156 78 L142 132 L66 132 Z" fill={palette.soft} />
        <path d="M104 78 L156 78 L142 132 L104 132 Z" fill={palette.ink} opacity={0.12} />
        <path d="M28 56 L48 56 L52 78" fill="none" stroke={palette.primary} strokeWidth={7} strokeLinecap="round" strokeLinejoin="round" />
      </g>
      <g opacity={wheels}>
        <circle cx={78} cy={150} r={11} fill={palette.primary} />
        <circle cx={132} cy={150} r={11} fill={palette.primary} />
      </g>
    </g>
  );
};

/** A price tag swinging from its hole. */
export const TagFigure: Figure = ({build, t, palette}) => {
  const tag = stage(build, 0, 2);
  const string = stage(build, 1, 2);
  const swingAngle = bob(t, 5, 1.4);
  return (
    <g transform={`rotate(${swingAngle} 62 50)`}>
      <path d="M62 50 L62 30" stroke={palette.line} strokeWidth={4} strokeLinecap="round" opacity={0.6 * string} fill="none" />
      <g transform={popAt(tag, 110, 100)} opacity={tag}>
        <path d="M62 44 L136 44 L168 100 L136 156 L62 156 A16 16 0 0 1 46 140 L46 60 A16 16 0 0 1 62 44 Z" fill={palette.primary} />
        <path d="M110 44 L136 44 L168 100 L136 156 L110 156 Z" fill={palette.ink} opacity={0.14} />
        <circle cx={70} cy={100} r={9} fill={palette.soft} />
      </g>
    </g>
  );
};

/** A wallet whose card slides out and back. */
export const WalletFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 2);
  const card = stage(build, 1, 2);
  const out = swing(t, 4.6, 0.22, 0.4);
  return (
    <g>
      <g opacity={card} transform={`translate(0 ${-out * 26})`}>
        <rect x={62} y={62} width={90} height={54} rx={8} fill={palette.accent} />
        <rect x={72} y={96} width={40} height={7} rx={3.5} fill={palette.ink} opacity={0.3} />
      </g>
      <g transform={popAt(body, C, 128)} opacity={body}>
        <rect x={40} y={90} width={120} height={72} rx={12} fill={palette.soft} />
        <rect x={40} y={90} width={120} height={72} rx={12} fill={palette.ink} opacity={0.08} />
        <rect x={40} y={110} width={120} height={52} rx={12} fill={palette.primary} />
        <rect x={118} y={126} width={30} height={16} rx={8} fill={palette.soft} opacity={0.6} />
      </g>
    </g>
  );
};

/** A receipt unrolling, line by line. */
export const ReceiptFigure: Figure = ({build, t, palette}) => {
  const paper = stage(build, 0, 2);
  const lines = stage(build, 1, 2);
  const unroll = swing(t, 6, 0.4, 0.4);
  return (
    <g transform={popAt(paper, C, C)} opacity={paper}>
      <path
        d="M56 32 L144 32 L144 158 L132 148 L120 158 L108 148 L96 158 L84 148 L72 158 L56 148 Z"
        fill={palette.soft}
      />
      <g opacity={lines}>
        {[0, 1, 2, 3].map((i) => {
          const on = unroll * 4 > i;
          const w = [64, 52, 68, 40][i];
          return (
            <rect
              key={i}
              x={70}
              y={54 + i * 22}
              width={w}
              height={8}
              rx={4}
              fill={i === 3 ? palette.accent : palette.ink}
              opacity={on ? (i === 3 ? 1 : 0.25) : 0}
            />
          );
        })}
      </g>
    </g>
  );
};

/** A star that fills, with the sparkle a rating gets. */
export const StarFigure: Figure = ({build, t, palette}) => {
  const star = stage(build, 0, 2);
  const glint = stage(build, 1, 2);
  const shine = pulse(t, 2.2);
  const d =
    'M100 34 L122 82 L174 89 L136 126 L145 178 L100 153 L55 178 L64 126 L26 89 L78 82 Z';
  return (
    <g transform={`rotate(${bob(t, 3, 1.1)} 100 106)`}>
      <g transform={popAt(star, C, 106)} opacity={star}>
        <path d={d} fill={palette.accent} />
        <path d="M100 34 L122 82 L174 89 L136 126 L145 178 L100 153 Z" fill={palette.ink} opacity={0.12} />
      </g>
      <g opacity={glint * shine}>
        <path d="M66 64 L74 72 M74 64 L66 72" stroke={palette.soft} strokeWidth={4} strokeLinecap="round" />
      </g>
    </g>
  );
};

/** A heart with a beat that is a beat, not a breath. */
export const HeartFigure: Figure = ({build, t, palette}) => {
  const heart = stage(build, 0, 1);
  // Two quick contractions and a rest — a heart on a sine wave reads as a
  // balloon inflating.
  const p = (t % 1.4) / 1.4;
  const thump =
    p < 0.12 ? Math.sin((p / 0.12) * Math.PI) : p < 0.3 ? 0.6 * Math.sin(((p - 0.18) / 0.12) * Math.PI) : 0;
  const s = 1 + 0.1 * Math.max(0, thump);
  return (
    <g transform={`${popAt(heart, C, C)} translate(100 104) scale(${s}) translate(-100 -104)`} opacity={heart}>
      <path
        d="M100 162 C40 122 30 88 46 66 C62 44 92 50 100 74 C108 50 138 44 154 66 C170 88 160 122 100 162 Z"
        fill={palette.accent}
      />
      <path d="M100 74 C108 50 138 44 154 66 C170 88 160 122 100 162 Z" fill={palette.ink} opacity={0.13} />
      <path d="M62 72 A22 22 0 0 1 80 64" fill="none" stroke={palette.soft} strokeWidth={5} strokeLinecap="round" opacity={0.5} />
    </g>
  );
};

/** A bell that rings — clapper and all. */
export const BellFigure: Figure = ({build, t, palette}) => {
  const bell = stage(build, 0, 2);
  const arcs = stage(build, 1, 2);
  const ring = gesture(t, 3, 0.4);
  const shake = ring >= 0 ? Math.sin(ring * Math.PI * 6) * 9 * (1 - ring) : 0;
  return (
    <g transform={`rotate(${shake} 100 44)`}>
      <g transform={popAt(bell, C, C)} opacity={bell}>
        <path d="M56 132 C56 92 62 62 100 56 C138 62 144 92 144 132 Z" fill={palette.soft} />
        <path d="M100 56 C138 62 144 92 144 132 L100 132 Z" fill={palette.ink} opacity={0.12} />
        <rect x={46} y={132} width={108} height={12} rx={6} fill={palette.primary} />
        <circle cx={C} cy={44} r={9} fill={palette.primary} />
        <circle cx={C} cy={154} r={9} fill={palette.primary} />
      </g>
      {/* The sound, drawn only while it is sounding. */}
      {ring >= 0 && (
        <g opacity={arcs * (1 - ring)}>
          {[0, 1].map((i) => (
            <g key={i}>
              <path d={`M${34 - i * 12} ${96 - i * 8} A${34 + i * 14} ${34 + i * 14} 0 0 1 ${34 - i * 12} ${68 - i * 8}`} fill="none" stroke={palette.accent} strokeWidth={4} strokeLinecap="round" />
              <path d={`M${166 + i * 12} ${96 - i * 8} A${34 + i * 14} ${34 + i * 14} 0 0 0 ${166 + i * 12} ${68 - i * 8}`} fill="none" stroke={palette.accent} strokeWidth={4} strokeLinecap="round" />
            </g>
          ))}
        </g>
      )}
    </g>
  );
};

/** An envelope whose flap opens and lets a letter rise. */
export const EnvelopeFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 2);
  const flap = stage(build, 1, 2);
  const open = swing(t, 4.8, 0.2, 0.4);
  return (
    <g>
      {/* The letter, behind the front panel and above the back one. */}
      <g opacity={open} transform={`translate(0 ${-open * 34})`}>
        <rect x={68} y={62} width={64} height={62} rx={5} fill={palette.soft} />
        <rect x={78} y={78} width={44} height={6} rx={3} fill={palette.ink} opacity={0.2} />
        <rect x={78} y={92} width={34} height={6} rx={3} fill={palette.ink} opacity={0.2} />
      </g>
      <g transform={popAt(body, C, C)} opacity={body}>
        <path d="M32 66 L168 66 L168 148 A8 8 0 0 1 160 156 L40 156 A8 8 0 0 1 32 148 Z" fill={palette.primary} />
        <path d="M32 156 L100 108 L168 156 Z" fill={palette.ink} opacity={0.15} />
      </g>
      <g opacity={flap} transform={`translate(100 66) scale(1 ${1 - 1.9 * open}) translate(-100 -66)`}>
        <path d="M32 66 L100 116 L168 66 Z" fill={palette.soft} />
      </g>
    </g>
  );
};

/** Two bubbles taking turns, with the dots of someone typing. */
export const ChatFigure: Figure = ({build, t, palette}) => {
  const a = stage(build, 0, 2);
  const b = stage(build, 1, 2);
  const turn = (t % 3.4) / 3.4;
  return (
    <g>
      <g transform={popAt(a, 84, 72)} opacity={a}>
        <path d="M30 44 L128 44 A12 12 0 0 1 140 56 L140 92 A12 12 0 0 1 128 104 L54 104 L38 120 L40 104 A12 12 0 0 1 30 92 Z" fill={palette.primary} />
        <rect x={46} y={62} width={62} height={7} rx={3.5} fill={palette.soft} opacity={0.6} />
        <rect x={46} y={78} width={40} height={7} rx={3.5} fill={palette.soft} opacity={0.4} />
      </g>
      <g transform={popAt(b, 124, 140)} opacity={b}>
        <path d="M170 104 L84 104 A12 12 0 0 0 72 116 L72 148 A12 12 0 0 0 84 160 L150 160 L166 174 L164 160 A12 12 0 0 0 170 148 Z" fill={palette.soft} />
        {/* Typing dots, only during the second speaker's turn. */}
        {[0, 1, 2].map((i) => {
          const lift = turn > 0.5 ? Math.max(0, Math.sin((turn * 8 - i * 0.6) * Math.PI)) : 0;
          return <circle key={i} cx={104 + i * 18} cy={132 - lift * 5} r={6} fill={palette.accent} opacity={turn > 0.5 ? 0.5 + 0.5 * lift : 0.2} />;
        })}
      </g>
    </g>
  );
};

/** A megaphone with a cone of sound leaving it. */
export const MegaphoneFigure: Figure = ({build, t, palette}) => {
  const horn = stage(build, 0, 2);
  const waves = stage(build, 1, 2);
  const shout = pulse(t, 3.2);
  return (
    <g transform={`rotate(${-8 + bob(t, 2, 2.2)} 70 110)`}>
      <g transform={popAt(horn, 84, 104)} opacity={horn}>
        <path d="M34 88 L34 128 L58 128 L110 158 L110 58 L58 88 Z" fill={palette.primary} />
        <path d="M110 58 L110 158 L58 128 L58 88 Z" fill={palette.ink} opacity={0.14} />
        <rect x={22} y={92} width={16} height={32} rx={6} fill={palette.soft} />
      </g>
      <g opacity={waves}>
        {[0, 1, 2].map((i) => (
          <path
            key={i}
            d={`M${124 + i * 16} ${86 - i * 10} A${26 + i * 16} ${26 + i * 16} 0 0 1 ${124 + i * 16} ${130 + i * 10}`}
            fill="none"
            stroke={palette.accent}
            strokeWidth={5}
            strokeLinecap="round"
            opacity={Math.max(0, shout - i * 0.28)}
          />
        ))}
      </g>
    </g>
  );
};

/** A node handing something to two others. */
export const ShareFigure: Figure = ({build, t, palette}) => {
  const links = stage(build, 0, 2);
  const nodes = stage(build, 1, 2);
  const at = cycle(t, 0.55);
  const from = {x: 58, y: 100};
  const to = [{x: 148, y: 54}, {x: 148, y: 146}];
  return (
    <g>
      <g opacity={links}>
        {to.map((p, i) => (
          <line key={i} x1={from.x} y1={from.y} x2={p.x} y2={p.y} stroke={palette.line} strokeWidth={5} strokeLinecap="round" opacity={0.5} />
        ))}
      </g>
      {to.map((p, i) => {
        const trav = cycle(t, 0.55, i * 0.5);
        return (
          <circle
            key={`p${i}`}
            cx={from.x + (p.x - from.x) * trav}
            cy={from.y + (p.y - from.y) * trav}
            r={6}
            fill={palette.accent}
            opacity={fadeTravel(trav)}
          />
        );
      })}
      <g opacity={nodes}>
        <circle cx={from.x} cy={from.y} r={20} fill={palette.primary} />
        {to.map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r={16} fill={palette.soft} opacity={0.6 + 0.4 * (at > 0.8 ? 1 : 0)} />
        ))}
      </g>
    </g>
  );
};

/** A play button on a strip of film, with a scrub head running. */
export const VideoFigure: Figure = ({build, t, palette}) => {
  const screen = stage(build, 0, 3);
  const play = stage(build, 1, 3);
  const bar = stage(build, 2, 3);
  const at = cycle(t, 0.32);
  return (
    <g>
      <g transform={popAt(screen, C, 96)} opacity={screen}>
        <rect x={30} y={44} width={140} height={100} rx={12} fill={palette.soft} />
        <rect x={30} y={44} width={140} height={100} rx={12} fill={palette.primary} opacity={0.9} />
      </g>
      <g transform={popAt(play, C, 94)} opacity={play}>
        <path d="M84 70 L124 94 L84 118 Z" fill={palette.soft} />
      </g>
      <g opacity={bar}>
        <rect x={40} y={158} width={120} height={7} rx={3.5} fill={palette.line} opacity={0.4} />
        <rect x={40} y={158} width={120 * at} height={7} rx={3.5} fill={palette.accent} />
        <circle cx={40 + 120 * at} cy={161.5} r={8} fill={palette.accent} />
      </g>
    </g>
  );
};
