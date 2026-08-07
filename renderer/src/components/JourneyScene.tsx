import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// JourneyScene: a route with one thing moving on it.
//
// The composition is a single gently bowed path across the full width, with the
// stops sitting on it as glyph nodes. The bow is doing real work: a straight
// line of evenly spaced circles reads as a list that happens to be horizontal,
// and the whole claim of this clip is that the packet is going SOMEWHERE. A
// curve has a near end and a far end, and the eye travels it.
//
// Everything on the route is drawn from the first frame, dim. That is the
// opener's entire content — the distance — and it is why the map beat exists at
// all. A stop lights only once the packet has been there, so at any moment the
// frame divides itself into what has happened and what has not, which is a
// thing the viewer can read without being told.
//
// The glyphs are geometric and drawn here in inline SVG rather than picked from
// an icon set: five shapes, all built from the same stroke weight and the same
// tokens, so a router and a server read as members of one family rather than as
// two pieces of clip art. Recognition is the job — the viewer should know what
// a node is before they read its label.
//
// The packet is the one glow in the frame. It rides the actual bezier the path
// is drawn from, with a short tail of fading dots behind it, so its speed is
// visibly the speed of a thing crossing a distance. The tail is what makes a
// dot read as travelling rather than as being redrawn.
//
// Each travelled leg keeps a small mono stamp at its midpoint. It is the hop
// ordinal rather than a millisecond figure: the plan carries no timings, and a
// latency this component invented would be the one number on screen that no
// validator had checked. The ordinal is a fact about the picture.

const ROUTE_W = Math.min(STAGE_W, 1580);
const ROUTE_H = 360;
// Half a node's width of breathing room at each end, so the origin glyph and
// its label are not flush against the drawing box.
const ROUTE_PAD = 96;
const NODE_R = 34;
// The bow. Deep enough that the path is unmistakably a curve at a glance,
// shallow enough that the labels under every node share a baseline band.
const BOW = 58;
const MID_Y = 150;
const CAPTION_H = 150;

type Stop = {label: string; kind: string; adds: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'map' | 'hop' | 'reach' | 'return';
  at?: number;
  reached: number;
  legs: number[];
};

type Pt = {x: number; y: number};

/** The node centres, bowed across the route. */
const layout = (n: number): Pt[] => {
  const span = ROUTE_W - ROUTE_PAD * 2;
  return Array.from({length: n}, (_, i) => {
    const t = n === 1 ? 0.5 : i / (n - 1);
    return {x: ROUTE_PAD + span * t, y: MID_Y - Math.sin(t * Math.PI) * BOW};
  });
};

/** Horizontal control handles, so consecutive legs meet without a kink. */
const handles = (a: Pt, b: Pt): [Pt, Pt] => {
  const dx = (b.x - a.x) / 2;
  return [
    {x: a.x + dx, y: a.y},
    {x: b.x - dx, y: b.y},
  ];
};

const legPath = (a: Pt, b: Pt): string => {
  const [c1, c2] = handles(a, b);
  return `M ${a.x} ${a.y} C ${c1.x} ${c1.y}, ${c2.x} ${c2.y}, ${b.x} ${b.y}`;
};

/** A point on the whole route, in units of legs travelled. */
const pointAt = (pts: Pt[], u: number): Pt => {
  const clamped = Math.max(0, Math.min(pts.length - 1, u));
  const leg = Math.min(pts.length - 2, Math.floor(clamped));
  if (leg < 0) return pts[0];
  const t = clamped - leg;
  const a = pts[leg];
  const b = pts[leg + 1];
  const [c1, c2] = handles(a, b);
  const m = 1 - t;
  return {
    x: m * m * m * a.x + 3 * m * m * t * c1.x + 3 * m * t * t * c2.x + t * t * t * b.x,
    y: m * m * m * a.y + 3 * m * m * t * c1.y + 3 * m * t * t * c2.y + t * t * t * b.y,
  };
};

/** The stop glyphs. One stroke weight, one construction, five silhouettes. */
const Glyph: React.FC<{kind: string; colour: string}> = ({kind, colour}) => {
  const s = {fill: 'none', stroke: colour, strokeWidth: 2.4, strokeLinecap: 'round' as const, strokeLinejoin: 'round' as const};
  switch (kind) {
    case 'router':
      return (
        <g>
          <rect x={-14} y={-10} width={28} height={20} rx={4} {...s} />
          <path d="M -7 -3 L 7 -3 M 4 -6 L 7 -3 L 4 0" {...s} />
          <path d="M 7 5 L -7 5 M -4 2 L -7 5 L -4 8" {...s} />
        </g>
      );
    case 'dns':
      return (
        <g>
          <circle cx={0} cy={0} r={13} {...s} />
          <path d="M -13 0 L 13 0" {...s} />
          <path d="M 0 -13 C 7 -6, 7 6, 0 13 C -7 6, -7 -6, 0 -13" {...s} />
        </g>
      );
    case 'server':
      return (
        <g>
          <rect x={-13} y={-13} width={26} height={11} rx={3} {...s} />
          <rect x={-13} y={2} width={26} height={11} rx={3} {...s} />
          <circle cx={-7.5} cy={-7.5} r={1.7} fill={colour} stroke="none" />
          <circle cx={-7.5} cy={7.5} r={1.7} fill={colour} stroke="none" />
        </g>
      );
    case 'cloud':
      return (
        <g>
          <path d="M -14 8 A 6.5 6.5 0 0 1 -11 -4 A 9 9 0 0 1 5 -6 A 7 7 0 0 1 13 8 Z" {...s} />
        </g>
      );
    default:
      return (
        <g>
          <rect x={-14} y={-12} width={28} height={19} rx={3} {...s} />
          <path d="M -8 12 L 8 12" {...s} />
        </g>
      );
  }
};

export const JourneyScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const stops = (Array.isArray(props.stops) ? props.stops : []) as Stop[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const returnText = String(props.return ?? '');
  if (stops.length < 2 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const pts = layout(stops.length);
  const reached = Math.max(0, Math.min(stops.length - 1, step.reached ?? 0));
  const legs = Array.isArray(step.legs) ? step.legs : [];
  const isMap = step.show === 'map';
  const isReturn = step.show === 'return';

  // The route draws itself on during the opener and then simply stays.
  const routeDraw = isMap
    ? interpolate(sinceStep, [2, 30], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : 1;

  // Where the packet is, in legs travelled. The hop springs into its landing
  // stop; the return sweeps the whole way home.
  let packetU = -1;
  if (step.show === 'hop' || step.show === 'reach') {
    const target = Math.max(0, Math.min(stops.length - 1, step.at ?? 0));
    const travel = spring({frame: sinceStep - 3, fps, config: {damping: 13, mass: 0.6}, durationInFrames: 30});
    packetU = target - 1 + travel;
  } else if (isReturn) {
    const sweep = interpolate(sinceStep, [6, 54], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});
    packetU = (stops.length - 1) * (1 - sweep);
  }
  const packet = packetU >= 0 ? pointAt(pts, packetU) : null;
  const tail = packet
    ? [0.09, 0.18, 0.27, 0.36].map((back) => pointAt(pts, isReturn ? packetU + back : packetU - back))
    : [];

  const activeStop = step.show === 'hop' || step.show === 'reach' ? Math.max(0, Math.min(stops.length - 1, step.at ?? 0)) : -1;
  const caption = isReturn ? returnText : activeStop >= 0 ? stops[activeStop].adds : '';
  const captionKicker = isReturn ? 'coming back' : activeStop >= 0 ? stops[activeStop].label : '';
  const captionEnter = interpolate(sinceStep, [10, 24], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={20}
      />

      <svg width={ROUTE_W} height={ROUTE_H} viewBox={`0 0 ${ROUTE_W} ${ROUTE_H}`} style={{flexShrink: 0, overflow: 'visible'}}>
        {/* The grid the whole composition is hung from: one hairline under the
            route, at the baseline the labels share. */}
        <line
          x1={0}
          y1={MID_Y + 96}
          x2={ROUTE_W}
          y2={MID_Y + 96}
          stroke={withAlpha(theme.line, 0.18)}
          strokeWidth={1}
        />

        {pts.slice(0, -1).map((p, i) => {
          const d = legPath(p, pts[i + 1]);
          const travelled = legs.includes(i + 1);
          const mid = pointAt(pts, i + 0.5);
          return (
            <g key={i}>
              <path
                d={d}
                fill="none"
                stroke={travelled ? withAlpha(theme.accent, 0.75) : withAlpha(theme.line, 0.3)}
                strokeWidth={travelled ? 3 : 2}
                strokeDasharray={isMap ? 1400 : undefined}
                strokeDashoffset={isMap ? 1400 * (1 - routeDraw) : undefined}
                strokeLinecap="round"
              />
              {travelled ? (
                <text
                  x={mid.x}
                  y={mid.y - 16}
                  textAnchor="middle"
                  fill={withAlpha(theme.accentText, 0.85)}
                  style={{fontFamily: theme.fontMono, fontSize: 15, letterSpacing: 1.4}}
                >
                  {`hop ${i + 1}`}
                </text>
              ) : null}
            </g>
          );
        })}

        {pts.map((p, i) => {
          const lit = i <= reached && !isMap;
          const arriving = i === activeStop;
          const pop = arriving
            ? spring({frame: sinceStep - 14, fps, config: {damping: 12, mass: 0.5}, durationInFrames: 24})
            : 1;
          const appear = isMap
            ? interpolate(sinceStep, [4 + i * 4, 16 + i * 4], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
            : 1;
          const stroke = lit ? theme.accent : withAlpha(theme.line, 0.55);
          return (
            <g key={i} transform={`translate(${p.x} ${p.y})`} opacity={appear}>
              <circle
                r={NODE_R * (0.94 + 0.06 * pop)}
                fill={lit ? withAlpha(theme.accent, 0.14) : withAlpha(theme.surface, 0.85)}
                stroke={stroke}
                strokeWidth={lit ? 2.4 : 1.6}
              />
              <Glyph kind={String(stops[i].kind ?? 'device')} colour={lit ? theme.accentText : theme.textMuted} />
              <text
                x={0}
                y={NODE_R + 34}
                textAnchor="middle"
                fill={lit ? theme.text : theme.textMuted}
                style={{fontFamily: theme.fontBody, fontSize: 23, fontWeight: 600}}
              >
                {stops[i].label}
              </text>
            </g>
          );
        })}

        {/* The one glow in the frame. */}
        {packet ? (
          <g>
            {tail.map((t, i) => (
              <circle key={i} cx={t.x} cy={t.y} r={9 - i * 1.7} fill={withAlpha(theme.accentQuantity, 0.34 - i * 0.07)} />
            ))}
            <circle cx={packet.x} cy={packet.y} r={11} fill={theme.accentQuantity} />
            <circle cx={packet.x} cy={packet.y} r={22} fill={withAlpha(theme.accentQuantity, 0.18)} />
          </g>
        ) : null}
      </svg>

      <div
        style={{
          width: ROUTE_W,
          minHeight: CAPTION_H,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'flex-start',
          justifyContent: 'flex-start',
          paddingTop: 26,
          opacity: caption ? captionEnter : 0,
          transform: `translateY(${(1 - captionEnter) * 10}px)`,
        }}
      >
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 16,
            letterSpacing: 3.4,
            textTransform: 'uppercase',
            color: isReturn ? theme.accentText : theme.textMuted,
            marginBottom: 12,
          }}
        >
          {captionKicker}
        </div>
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 40,
            fontWeight: 600,
            letterSpacing: -0.6,
            lineHeight: 1.2,
            color: isReturn ? theme.accentText : theme.text,
            maxWidth: 1200,
          }}
        >
          {caption}
        </div>
      </div>
    </Stage>
  );
};
