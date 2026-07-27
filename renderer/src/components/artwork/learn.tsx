// Teaching, and the shapes a lesson is made of.
//
// This module exists because the vocabulary was built for explaining *systems*
// and courseSmith mostly explains *ideas to a beginner*. There was a figure for
// a load balancer and none for a question; one for a queue and none for the
// moment something clicks. The whiteboard and flow templates draw from this set
// directly now, so a gap here is a box on a finished board illustrating nothing.
//
// Same rule as everywhere else in the drawer: whatever the thing *does* is the
// thing that moves. A question mark tilts because it is asking, a timer's hand
// sweeps, the wrong answer is struck out and the right one ticks. A figure that
// assembles and then holds still is a slide.

import {
  C,
  bob,
  cycle,
  draws,
  ease,
  gesture,
  popAt,
  pulse,
  ring,
  stage,
  swing,
  type Figure,
} from './kit';

/** A question mark on a disc, tilting as if asked. */
export const QuestionFigure: Figure = ({build, t, palette}) => {
  const disc = stage(build, 0, 2);
  const mark = stage(build, 1, 2);
  // It leans into the question and back. A mark held upright is punctuation;
  // one that tilts is somebody asking.
  const lean = Math.sin(t * 1.6) * 5;
  return (
    <g transform={popAt(disc, C, C)} opacity={disc}>
      <circle cx={C} cy={C} r={62} fill={palette.primary} />
      <circle cx={C} cy={C} r={62} fill={palette.ink} opacity={0.08} />
      <g transform={`rotate(${lean} 100 108)`} opacity={mark}>
        <path
          d="M78 78 C78 58 122 56 122 80 C122 98 100 100 100 118"
          fill="none"
          stroke={palette.soft}
          strokeWidth={13}
          strokeLinecap="round"
          {...draws(mark)}
        />
        <circle cx={100} cy={140} r={8} fill={palette.soft} opacity={ease(mark)} />
      </g>
    </g>
  );
};

/** A board on legs with a line of chalk written across it. */
export const ChalkboardFigure: Figure = ({build, t, palette}) => {
  const board = stage(build, 0, 3);
  const legs = stage(build, 1, 3);
  const chalk = stage(build, 2, 3);
  // The writing appears, holds long enough to be read, then wipes.
  const written = swing(t, 6.5, 0.28, 0.44);
  const rows = [
    {y: 84, w: 78},
    {y: 106, w: 96},
    {y: 128, w: 58},
  ];
  return (
    <g>
      <g transform={popAt(board, C, 100)} opacity={board}>
        <rect x={26} y={50} width={148} height={98} rx={8} fill={palette.ink} opacity={0.82} />
        <rect x={26} y={50} width={148} height={98} rx={8} fill="none" stroke={palette.primary} strokeWidth={6} />
      </g>
      <g opacity={chalk}>
        {rows.map((r, i) => {
          // Each line is written after the one above it.
          const p = Math.max(0, Math.min(1, (written - i * 0.14) / 0.5));
          return (
            <rect
              key={i}
              x={46}
              y={r.y}
              width={r.w * ease(p)}
              height={7}
              rx={3.5}
              fill={palette.soft}
              opacity={0.9}
            />
          );
        })}
      </g>
      <g opacity={legs}>
        <rect x={54} y={148} width={7} height={28} rx={3} fill={palette.line} />
        <rect x={139} y={148} width={7} height={28} rx={3} fill={palette.line} />
      </g>
    </g>
  );
};

/** A lightbulb over a head-shape: the moment it clicks. */
export const InsightFigure: Figure = ({build, t, palette}) => {
  const head = stage(build, 0, 2);
  const bulb = stage(build, 1, 2);
  // The idea arrives on a beat rather than glowing constantly — a bulb always
  // on is a lamp, not a realisation.
  const spark = gesture(t, 3.4, 0.45);
  const lit = spark >= 0 ? Math.sin(spark * Math.PI) : 0;
  return (
    <g>
      <g transform={popAt(head, C, 132)} opacity={head}>
        <path d="M62 176 C62 138 78 118 100 118 C122 118 138 138 138 176 Z" fill={palette.soft} />
        <circle cx={C} cy={112} r={26} fill={palette.soft} />
        <circle cx={C} cy={112} r={26} fill={palette.ink} opacity={0.08} />
      </g>
      <g transform={popAt(bulb, C, 56)} opacity={bulb}>
        {/* Rays, only while it is lit. */}
        {lit > 0.02 &&
          ring(8, 40, C, 52).map((p, i) => (
            <line
              key={i}
              x1={C + (p.x - C) * 0.62}
              y1={52 + (p.y - 52) * 0.62}
              x2={p.x}
              y2={p.y}
              stroke={palette.accent}
              strokeWidth={4}
              strokeLinecap="round"
              opacity={lit * 0.9}
            />
          ))}
        {/* The bulb is always plainly a bulb; the *gesture* is the rays and the
            lift in brightness. Resting it at 0.28 made the figure read as a
            headless blob for the three seconds between beats — which is most of
            the time it is on screen. */}
        <circle cx={C} cy={52} r={24} fill={palette.accent} opacity={0.66 + lit * 0.34} />
        <path
          d="M88 40 C88 26 112 26 112 40 C112 50 104 54 104 64 L96 64 C96 54 88 50 88 40 Z"
          fill={palette.soft}
          opacity={0.5}
        />
        <rect x={92} y={74} width={16} height={12} rx={3} fill={palette.line} />
      </g>
    </g>
  );
};

/** A stopwatch whose hand sweeps. */
export const TimerFigure: Figure = ({build, t, palette}) => {
  const body = stage(build, 0, 2);
  const hand = stage(build, 1, 2);
  const sweep = cycle(t, 0.22) * Math.PI * 2;
  return (
    <g transform={popAt(body, C, 108)} opacity={body}>
      <rect x={88} y={30} width={24} height={16} rx={5} fill={palette.line} />
      <circle cx={C} cy={112} r={58} fill={palette.soft} />
      <circle cx={C} cy={112} r={58} fill="none" stroke={palette.primary} strokeWidth={7} />
      {/* Quarter ticks only: twelve on a 116px dial is a grey ring. */}
      {ring(4, 44, C, 112, -Math.PI / 2).map((p, i) => (
        <circle key={i} cx={p.x} cy={p.y} r={3.4} fill={palette.ink} opacity={0.3} />
      ))}
      <g opacity={hand}>
        <line
          x1={C}
          y1={112}
          x2={C + Math.sin(sweep) * 38}
          y2={112 - Math.cos(sweep) * 38}
          stroke={palette.accent}
          strokeWidth={5}
          strokeLinecap="round"
        />
        <circle cx={C} cy={112} r={6} fill={palette.accent} />
      </g>
    </g>
  );
};

/** A rosette with a ribbon: the thing you get for finishing. */
export const CertificateFigure: Figure = ({build, t, palette}) => {
  const sheet = stage(build, 0, 3);
  const seal = stage(build, 1, 3);
  const ribbon = stage(build, 2, 3);
  const shine = pulse(t, 1.8);
  return (
    <g>
      <g transform={popAt(sheet, C, 96)} opacity={sheet}>
        <rect x={34} y={40} width={132} height={104} rx={8} fill={palette.soft} />
        <rect x={34} y={40} width={132} height={104} rx={8} fill="none" stroke={palette.primary} strokeWidth={4} />
        {[64, 82, 100].map((y, i) => (
          <rect key={i} x={54} y={y} width={i === 2 ? 52 : 92} height={6} rx={3} fill={palette.ink} opacity={0.2} />
        ))}
      </g>
      <g transform={popAt(seal, 138, 132)} opacity={seal}>
        <circle cx={138} cy={132} r={24} fill={palette.accent} opacity={0.55 + shine * 0.45} />
        <circle cx={138} cy={132} r={14} fill="none" stroke={palette.ink} strokeWidth={3} opacity={0.35} />
      </g>
      <g opacity={ribbon}>
        <path d="M128 152 L126 182 L138 172 L150 182 L148 152 Z" fill={palette.primary} />
      </g>
    </g>
  );
};

/** Two answers: the wrong one struck out, the right one ticked. */
export const AnswerFigure: Figure = ({build, t, palette}) => {
  const rows = stage(build, 0, 2);
  const marks = stage(build, 1, 2);
  // The verdict lands, holds, and resets — so the figure keeps teaching rather
  // than freezing on the answer.
  const verdict = swing(t, 5.4, 0.22, 0.5);
  return (
    <g>
      <g opacity={rows}>
        <rect x={34} y={58} width={132} height={38} rx={9} fill={palette.soft} />
        <rect x={34} y={110} width={132} height={38} rx={9} fill={palette.soft} />
        <rect x={34} y={110} width={132} height={38} rx={9} fill="none" stroke={palette.primary} strokeWidth={3} opacity={verdict} />
        <rect x={54} y={73} width={68} height={8} rx={4} fill={palette.ink} opacity={0.22} />
        <rect x={54} y={125} width={84} height={8} rx={4} fill={palette.ink} opacity={0.22} />
      </g>
      <g opacity={marks}>
        {/* Struck through, not deleted: the wrong answer has to stay legible or
            there is nothing to have been wrong about. */}
        <line
          x1={48}
          y1={77}
          x2={48 + 104 * ease(verdict)}
          y2={77}
          stroke={palette.line}
          strokeWidth={4}
          strokeLinecap="round"
        />
        <path
          d="M52 130 L64 141 L86 118"
          fill="none"
          stroke={palette.accent}
          strokeWidth={7}
          strokeLinecap="round"
          strokeLinejoin="round"
          {...draws(verdict)}
        />
      </g>
    </g>
  );
};

/** A stack of books, the top one lifting off. */
export const LibraryFigure: Figure = ({build, t, palette}) => {
  const shelf = [
    {y: 148, w: 118, fill: 'primary'},
    {y: 122, w: 104, fill: 'soft'},
    {y: 96, w: 112, fill: 'primary'},
  ] as const;
  const lift = swing(t, 5.8, 0.2, 0.3);
  return (
    <g>
      {shelf.map((b, i) => {
        const p = stage(build, i, shelf.length + 1, 0.5);
        return (
          <g key={i} transform={popAt(p, C, b.y + 11)} opacity={p}>
            <rect
              x={C - b.w / 2}
              y={b.y}
              width={b.w}
              height={22}
              rx={5}
              fill={b.fill === 'primary' ? palette.primary : palette.soft}
            />
            <rect x={C - b.w / 2 + 8} y={b.y} width={5} height={22} fill={palette.ink} opacity={0.2} />
          </g>
        );
      })}
      {/* The top book lifts away and settles back — a shelf you take from. */}
      <g
        transform={`translate(0 ${-lift * 26})`}
        opacity={stage(build, shelf.length, shelf.length + 1, 0.5)}
      >
        <rect x={C - 46} y={66} width={92} height={22} rx={5} fill={palette.accent} />
        <rect x={C - 38} y={66} width={5} height={22} fill={palette.ink} opacity={0.22} />
      </g>
    </g>
  );
};

/** A highlighter drawing a band of colour over a line of text. */
export const HighlighterFigure: Figure = ({build, t, palette}) => {
  const text = stage(build, 0, 2);
  const pen = stage(build, 1, 2);
  const swipe = swing(t, 4.6, 0.3, 0.4);
  const bandW = 112 * ease(swipe);
  return (
    <g>
      <g opacity={text}>
        {[70, 96, 122].map((y, i) => (
          <rect key={i} x={40} y={y} width={i === 1 ? 118 : 96} height={9} rx={4.5} fill={palette.ink} opacity={0.24} />
        ))}
      </g>
      {/* The band sits *behind* the middle line, which is what a highlighter
          does — over the top it would read as redaction. */}
      <rect x={40} y={90} width={bandW} height={21} rx={4} fill={palette.accent} opacity={0.5} />
      <g opacity={pen} transform={`translate(${40 + bandW} 0)`}>
        <g transform="rotate(38 0 100)">
          <rect x={-11} y={62} width={22} height={44} rx={5} fill={palette.primary} />
          <path d="M-11 106 L11 106 L4 124 L-4 124 Z" fill={palette.accent} />
        </g>
      </g>
    </g>
  );
};

/** A signpost with two arms: the point where you choose. */
export const SignpostFigure: Figure = ({build, t, palette}) => {
  const post = stage(build, 0, 3);
  const armA = stage(build, 1, 3);
  const armB = stage(build, 2, 3);
  // A slow sway, so the arms have weight.
  const sway = bob(t, 1.6, 1.1);
  return (
    <g transform={`rotate(${sway} 100 176)`}>
      <g opacity={post}>
        <rect x={94} y={54} width={12} height={122} rx={5} fill={palette.line} />
      </g>
      <g transform={popAt(armA, 104, 82)} opacity={armA}>
        <path d="M104 68 L166 68 L180 82 L166 96 L104 96 Z" fill={palette.primary} />
      </g>
      <g transform={popAt(armB, 96, 122)} opacity={armB}>
        <path d="M96 108 L34 108 L20 122 L34 136 L96 136 Z" fill={palette.accent} />
      </g>
    </g>
  );
};

/** A brick wall with one brick sliding into the gap. */
export const FoundationFigure: Figure = ({build, t, palette}) => {
  const courses = [
    {y: 148, offset: 0},
    {y: 120, offset: 22},
    {y: 92, offset: 0},
  ];
  const place = swing(t, 5.2, 0.26, 0.34);
  return (
    <g>
      {courses.map((row, r) => {
        const p = stage(build, r, courses.length + 1, 0.5);
        return (
          <g key={r} opacity={p}>
            {[0, 1, 2].map((c) => (
              <rect
                key={c}
                x={30 + row.offset + c * 48}
                y={row.y}
                width={42}
                height={22}
                rx={4}
                fill={r === 1 ? palette.soft : palette.primary}
              />
            ))}
          </g>
        );
      })}
      {/* The last brick arrives from the left and seats itself in the top row. */}
      <g opacity={stage(build, courses.length, courses.length + 1, 0.5)}>
        <rect
          x={30 + 3 * 48 - (1 - ease(place)) * 60}
          y={92}
          width={42}
          height={22}
          rx={4}
          fill={palette.accent}
          opacity={0.4 + ease(place) * 0.6}
        />
      </g>
    </g>
  );
};

/** A progress ring filling, for a course or a skill. */
export const ProgressFigure: Figure = ({build, t, palette}) => {
  const track = stage(build, 0, 2);
  const fill = stage(build, 1, 2);
  const r = 56;
  const circ = 2 * Math.PI * r;
  // Fills, holds full, and resets. A ring parked at 100% says nothing; the
  // filling is the whole content of the figure.
  const done = swing(t, 7, 0.42, 0.3);
  return (
    <g transform={popAt(track, C, C)} opacity={track}>
      <circle cx={C} cy={C} r={r} fill="none" stroke={palette.soft} strokeWidth={16} />
      <circle
        cx={C}
        cy={C}
        r={r}
        fill="none"
        stroke={palette.accent}
        strokeWidth={16}
        strokeLinecap="round"
        strokeDasharray={circ}
        strokeDashoffset={circ * (1 - ease(done) * fill)}
        transform={`rotate(-90 ${C} ${C})`}
      />
      <circle cx={C} cy={C} r={34} fill={palette.primary} opacity={0.18} />
    </g>
  );
};

/** A speech bubble with a reply arriving under it. */
export const DiscussionFigure: Figure = ({build, t, palette}) => {
  const first = stage(build, 0, 2);
  const second = stage(build, 1, 2);
  // The second bubble answers on a beat, so the pair reads as a conversation
  // rather than as two shapes.
  const reply = gesture(t, 4.2, 0.5);
  const replyIn = reply >= 0 ? ease(Math.min(1, reply * 2)) : 1;
  return (
    <g>
      <g transform={popAt(first, 84, 74)} opacity={first}>
        <rect x={24} y={44} width={120} height={62} rx={16} fill={palette.primary} />
        <path d="M46 106 L46 128 L72 106 Z" fill={palette.primary} />
        {[62, 82].map((y, i) => (
          <rect key={i} x={44} y={y} width={i ? 58 : 78} height={7} rx={3.5} fill={palette.soft} opacity={0.75} />
        ))}
      </g>
      <g transform={popAt(second, 124, 148)} opacity={second * replyIn}>
        <rect x={62} y={118} width={112} height={54} rx={16} fill={palette.soft} />
        <path d="M152 172 L152 190 L128 172 Z" fill={palette.soft} />
        {[132, 150].map((y, i) => (
          <rect key={i} x={80} y={y} width={i ? 46 : 72} height={7} rx={3.5} fill={palette.ink} opacity={0.28} />
        ))}
      </g>
    </g>
  );
};

/** A magnifier over a document, with the glass travelling across it. */
export const StudyFigure: Figure = ({build, t, palette}) => {
  const sheet = stage(build, 0, 2);
  const glass = stage(build, 1, 2);
  const across = Math.sin(t * 0.9) * 30;
  const down = Math.cos(t * 0.62) * 16;
  return (
    <g>
      <g transform={popAt(sheet, C, C)} opacity={sheet}>
        <rect x={44} y={34} width={112} height={140} rx={10} fill={palette.soft} />
        {[62, 84, 106, 128, 150].map((y, i) => (
          <rect key={i} x={62} y={y} width={i % 2 ? 60 : 76} height={7} rx={3.5} fill={palette.ink} opacity={0.22} />
        ))}
      </g>
      <g transform={`translate(${across} ${down})`} opacity={glass}>
        <circle cx={112} cy={100} r={34} fill={palette.accent} opacity={0.16} />
        <circle cx={112} cy={100} r={34} fill="none" stroke={palette.primary} strokeWidth={8} />
        <line x1={136} y1={124} x2={158} y2={148} stroke={palette.primary} strokeWidth={10} strokeLinecap="round" />
      </g>
    </g>
  );
};

/** Steps climbing to a flag. */
export const StepsFigure: Figure = ({build, t, palette}) => {
  const steps = [0, 1, 2, 3];
  const climb = cycle(t, 0.32);
  const on = Math.floor(climb * steps.length);
  return (
    <g>
      {steps.map((i) => {
        const p = stage(build, i, steps.length + 1, 0.55);
        const h = 30 + i * 28;
        const x = 30 + i * 38;
        return (
          <g key={i} opacity={p}>
            <rect
              x={x}
              y={176 - h}
              width={34}
              height={h}
              rx={5}
              fill={i === on ? palette.accent : palette.primary}
              opacity={i === on ? 1 : 0.85}
            />
          </g>
        );
      })}
      <g opacity={stage(build, steps.length, steps.length + 1, 0.55)}>
        <rect x={180} y={48} width={5} height={62} rx={2.5} fill={palette.line} />
        <path d={`M185 52 L${185 + 26} 62 L185 72 Z`} fill={palette.accent} />
      </g>
    </g>
  );
};

/** A mortarboard with its tassel swinging. */
export const GraduateFigure: Figure = ({build, t, palette}) => {
  const cap = stage(build, 0, 2);
  const tassel = stage(build, 1, 2);
  const swingAngle = Math.sin(t * 1.5) * 9;
  return (
    <g transform={popAt(cap, C, C)} opacity={cap}>
      <path d="M100 58 L182 92 L100 126 L18 92 Z" fill={palette.primary} />
      <path d="M100 58 L182 92 L100 126 Z" fill={palette.ink} opacity={0.12} />
      <path d="M62 106 L62 140 C62 154 138 154 138 140 L138 106 L100 122 Z" fill={palette.soft} />
      <g opacity={tassel} transform={`rotate(${swingAngle} 152 96)`}>
        <line x1={152} y1={96} x2={152} y2={134} stroke={palette.accent} strokeWidth={4} strokeLinecap="round" />
        <circle cx={152} cy={140} r={8} fill={palette.accent} />
      </g>
      <circle cx={152} cy={96} r={5} fill={palette.accent} />
    </g>
  );
};

/** A bookmark sliding into a page. */
export const BookmarkFigure: Figure = ({build, t, palette}) => {
  const page = stage(build, 0, 2);
  const mark = stage(build, 1, 2);
  const slide = swing(t, 5, 0.24, 0.42);
  return (
    <g>
      <g transform={popAt(page, C, C)} opacity={page}>
        <rect x={44} y={34} width={112} height={140} rx={10} fill={palette.soft} />
        {[70, 92, 114, 136].map((y, i) => (
          <rect key={i} x={62} y={y} width={i % 2 ? 56 : 74} height={7} rx={3.5} fill={palette.ink} opacity={0.2} />
        ))}
      </g>
      <g opacity={mark} transform={`translate(0 ${-26 + ease(slide) * 26})`}>
        <path d="M114 26 L146 26 L146 108 L130 94 L114 108 Z" fill={palette.accent} />
      </g>
    </g>
  );
};
