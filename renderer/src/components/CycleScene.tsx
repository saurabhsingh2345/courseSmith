import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage} from './Stage';
import {iconFor} from './icons';

// CycleScene is a closed ring with a light running round it.
//
// The template's whole claim is "this comes back", and a diagram cannot make
// that claim — an arrow returning to the first box is a line the eye reads once
// and forgets. Motion can. So the ring is standing furniture and the only thing
// that ever moves is a comet travelling it, which means the return is something
// the viewer *watches happen* rather than something they infer from an
// arrowhead.
//
// Four decisions carry it.
//
// The comet has a tail, drawn as a run of shrinking dots behind the head. A bare
// dot moving round a circle reads as a loading spinner; a tail gives the motion a
// direction even in a single frame, which matters because most frames of this
// scene are single frames somebody scrubbed to.
//
// The traversed arc stays lit behind it. By the last beat the ring is a complete
// bright circle, which is the picture the clip exists to produce — and it can
// only be complete if the light has actually gone all the way round, so the
// image is a receipt for the journey rather than a decoration.
//
// The stages are discs with icons, and their labels sit OUTSIDE the ring on the
// radius. Labels inside would collide with the hub at four stages and with each
// other at six; on the radius they can never collide, because the ring's own
// geometry spaces them.
//
// And the hub carries the loop's name until the return, when it carries what is
// different next lap. That is the one line this template is built to deliver, so
// it lands in the middle of the frame, inside a ring that has just closed.

const R = 252;
const NODE_R = 50;
const BOX_W = 1180;
const BOX_H = 812;
const CX = BOX_W / 2;
const CY = 366;

type StageSpec = {label: string; icon?: string; note?: string; angle: number};
type Step = {startMs: number; endMs: number; show: 'ring' | 'stage' | 'again'; at?: number};

// Where the comet rests before the first stage runs: a little short of the top,
// so the opening beat has the light *waiting* to set off rather than already
// standing on stage one.
const START_T = -0.34;

const polar = (t: number, n: number, radius: number) => {
  const deg = -90 + (t * 360) / n;
  const rad = (deg * Math.PI) / 180;
  return {x: CX + Math.cos(rad) * radius, y: CY + Math.sin(rad) * radius, deg};
};

/**
 * An SVG arc along the ring, from t0 to t1 in stage units.
 *
 * Emitted in quarter-circle pieces rather than as one command, which is not
 * fussiness: a single `A` cannot express a full circle at all — its start and
 * end points coincide, so the renderer draws nothing — and at the last beat this
 * path IS a full circle. The first version of this drew one arc with the
 * large-arc flag and produced a loop hanging off the top of the ring on exactly
 * the frame the template exists to deliver. Quarter pieces are exact at any
 * length and need no flags.
 */
const arcPath = (t0: number, t1: number, n: number, radius: number): string => {
  // Never past one full turn: the ring cannot show a second lap, and asking it
  // to would draw the opening arc twice.
  const span = Math.min(t1 - t0, n);
  if (span <= 0) return '';
  const pieces = Math.max(1, Math.ceil((span * 360) / n / 90));
  const from = t1 - span;
  const start = polar(from, n, radius);
  let d = `M ${start.x} ${start.y}`;
  for (let k = 1; k <= pieces; k++) {
    const p = polar(from + (span * k) / pieces, n, radius);
    d += ` A ${radius} ${radius} 0 0 1 ${p.x} ${p.y}`;
  }
  return d;
};

/**
 * Where a stage's name sits, and which way it is set.
 *
 * Centring every label on its own radius is the obvious approach and it collides
 * at exactly the two angles a ring always has: a label centred on the horizontal
 * radius runs straight back through the node it belongs to. So a label to the
 * side of the ring is set against the node and pushed outward, and only the ones
 * near the top and bottom are centred over it — which is what somebody
 * lettering this by hand would do without thinking about it.
 */
const labelPlacement = (deg: number, x: number, y: number) => {
  const c = Math.cos((deg * Math.PI) / 180);
  const s = Math.sin((deg * Math.PI) / 180);
  if (c > 0.45) {
    return {left: x + NODE_R + 20, top: y, transform: 'translateY(-50%)', textAlign: 'left' as const};
  }
  if (c < -0.45) {
    return {
      left: x - NODE_R - 20,
      top: y,
      transform: 'translate(-100%, -50%)',
      textAlign: 'right' as const,
    };
  }
  return {
    left: x,
    top: y + (s < 0 ? -NODE_R - 34 : NODE_R + 34),
    transform: 'translate(-50%, -50%)',
    textAlign: 'center' as const,
  };
};

export const CycleScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const name = String(props.name ?? '');
  const changes = String(props.changes ?? '');
  const stages = (Array.isArray(props.stages) ? props.stages : []) as StageSpec[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const n = stages.length;
  if (n === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;
  const sinceStart = ((nowMs - steps[0].startMs) / 1000) * FPS;
  const returning = step.show === 'again';

  // Where the light rests at the end of each beat, in stage units. Travel takes
  // the first two thirds of a beat so every stage gets a moment standing still
  // under its own narration.
  const restOf = (s: Step): number => {
    if (s.show === 'ring') return START_T;
    if (s.show === 'again') return n;
    // A shade short of the stage, so the light waits at the disc's edge rather
    // than under it. The nodes are opaque — they have to be, or the arc runs
    // through the icon — so a comet parked dead on its stage is a comet nobody
    // can see, on the beat that stage is being talked about.
    return (s.at ?? 0) - 0.12;
  };
  const from = idx === 0 ? START_T : restOf(steps[idx - 1]);
  const to = restOf(step);
  const beatFrames = ((step.endMs - step.startMs) / 1000) * FPS;
  const travel = interpolate(sinceStep, [2, Math.max(14, beatFrames * 0.6)], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // Eased, because a light moving round a ring at constant speed is a machine
  // and a light that sets off and arrives is a journey.
  const eased = travel * travel * (3 - 2 * travel);
  const t = from + (to - from) * eased;

  const enter = spring({frame: sinceStart, fps, config: {damping: 200, mass: 0.8}, durationInFrames: 26});
  // The ring draws itself in on the opening beat: a circle that was always there
  // is a shape, one that gets drawn is a claim about a route.
  const ringDraw = interpolate(sinceStart, [4, 34], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const pulse = 0.5 + 0.5 * Math.sin((frame / fps) * 2.4);

  // Which stages the light has already reached.
  const lit = (i: number) => t >= i - 0.13 || returning;
  const current = step.show === 'stage' ? (step.at ?? 0) : -1;
  const note = current >= 0 ? stages[current]?.note : undefined;

  const head = polar(t, n, R);
  const circumference = 2 * Math.PI * R;

  return (
    <Stage justify="center">
      <div style={{width: BOX_W, height: BOX_H, position: 'relative', opacity: enter}}>
        <svg width={BOX_W} height={BOX_H} style={{position: 'absolute', inset: 0, overflow: 'visible'}}>
          <defs>
            <radialGradient id="cycle-hub-glow">
              <stop offset="0%" stopColor={withAlpha(theme.accentQuantity, 0.24)} />
              <stop offset="100%" stopColor={withAlpha(theme.accentQuantity, 0)} />
            </radialGradient>
          </defs>

          {/* An outer ring of ticks, turning very slowly. It is the only thing on
              the frame that never stops, so a still of this scene still says the
              machine is running. */}
          <g transform={`rotate(${(frame / fps) * 3} ${CX} ${CY})`} opacity={enter * 0.5}>
            {Array.from({length: 72}).map((_, i) => {
              const rad = ((i * 360) / 72 - 90) * (Math.PI / 180);
              const r0 = R + 34;
              const r1 = r0 + (i % 6 === 0 ? 12 : 5);
              return (
                <line
                  key={i}
                  x1={CX + Math.cos(rad) * r0}
                  y1={CY + Math.sin(rad) * r0}
                  x2={CX + Math.cos(rad) * r1}
                  y2={CY + Math.sin(rad) * r1}
                  stroke={withAlpha(theme.text, i % 6 === 0 ? 0.22 : 0.1)}
                  strokeWidth={1.5}
                />
              );
            })}
          </g>

          {/* The ring itself, drawing on from the top. */}
          <circle
            cx={CX}
            cy={CY}
            r={R}
            fill="none"
            stroke={withAlpha(theme.text, 0.13)}
            strokeWidth={10}
            strokeDasharray={circumference}
            strokeDashoffset={circumference * (1 - ringDraw)}
            transform={`rotate(-90 ${CX} ${CY})`}
            strokeLinecap="round"
          />

          {/* The part the light has travelled, left lit behind it. By the last
              beat this is the whole circle — the picture is a receipt for the
              journey, not an ornament. */}
          <path
            d={arcPath(START_T, Math.max(t, START_T + 0.001), n, R)}
            fill="none"
            stroke={theme.accentQuantity}
            strokeWidth={10}
            strokeLinecap="round"
            opacity={0.95}
          />

          {/* Direction, stated between every pair of stages rather than once. A
              ring with one arrowhead is a ring somebody has to study. */}
          {stages.map((_, i) => {
            const mid = polar(i + 0.5, n, R);
            const on = t >= i + 0.5;
            return (
              <g
                key={`chev-${i}`}
                transform={`translate(${mid.x} ${mid.y}) rotate(${mid.deg + 90})`}
                opacity={enter * (on ? 1 : 0.4)}
              >
                <path
                  d="M -11 -7 L 0 4 L 11 -7"
                  fill="none"
                  // On the lit run the chevron is cut OUT of the arc, which is
                  // the only way a mark on a 10px stroke stays legible; ahead of
                  // the light it is drawn on the empty ring instead.
                  stroke={on ? theme.bgBottom : withAlpha(theme.text, 0.55)}
                  strokeWidth={3.4}
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </g>
            );
          })}

          {/* The hub's glow. */}
          <circle cx={CX} cy={CY} r={168} fill="url(#cycle-hub-glow)" opacity={returning ? 1 : 0.45} />

        </svg>

        {/* The stages. */}
        {stages.map((s, i) => {
          const p = polar(i, n, R);
          const place = labelPlacement(p.deg, p.x, p.y);
          const on = lit(i);
          const isCurrent = i === current;
          const land = isCurrent
            ? spring({frame: sinceStep - 6, fps, config: {damping: 14, mass: 0.5}, durationInFrames: 22})
            : 0;
          const c = on ? theme.accentQuantity : withAlpha(theme.text, 0.3);
          const Icon = iconFor(s.icon);
          return (
            <div key={i}>
              <div
                style={{
                  position: 'absolute',
                  left: p.x,
                  top: p.y,
                  transform: `translate(-50%, -50%) scale(${1 + land * 0.16})`,
                  width: NODE_R * 2,
                  height: NODE_R * 2,
                  borderRadius: '50%',
                  boxSizing: 'border-box',
                  // Opaque, with the tint painted over it by an inset shadow.
                  // A translucent fill let the 10px lit arc run straight through
                  // the disc and behind the icon, which at 26px left the icon
                  // unreadable on the one frame the ring is complete.
                  background: theme.bgBottom,
                  boxShadow: `inset 0 0 0 999px ${withAlpha(
                    theme.accentQuantity,
                    on ? (isCurrent ? 0.2 : 0.12) : 0,
                  )}${isCurrent ? `, 0 0 32px ${withAlpha(theme.accentQuantity, 0.35)}` : ''}`,
                  border: `3px solid ${c}`,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: on ? theme.accentQuantity : theme.textMuted,
                }}
              >
                <Icon size={30} strokeWidth={2.2} />
              </div>
              <div
                style={{
                  position: 'absolute',
                  ...place,
                  width: 230,
                  fontFamily: theme.fontDisplay,
                  fontSize: 30,
                  fontWeight: 700,
                  lineHeight: 1.12,
                  letterSpacing: -0.4,
                  color: on ? theme.text : theme.textMuted,
                  opacity: isCurrent ? 1 : on ? 0.8 : 0.42,
                }}
              >
                {s.label}
              </div>
            </div>
          );
        })}

        {/* The comet, on its own layer above the stages.
            The tail is what gives a single frame a direction, which matters
            because most frames of this scene are frames somebody paused on. It
            is drawn last, after the nodes, so the one moving object on the stage
            can never end up behind one of the standing ones — which is exactly
            what happened the moment the discs were made opaque. */}
        <svg width={BOX_W} height={BOX_H} style={{position: 'absolute', inset: 0, overflow: 'visible'}}>
          {Array.from({length: 11}).map((_, k) => {
            const p = polar(t - k * 0.032, n, R);
            const f = 1 - k / 11;
            return (
              <circle
                key={`tail-${k}`}
                cx={p.x}
                cy={p.y}
                r={3 + f * 8}
                fill={theme.accentQuantity}
                opacity={f * f * 0.75}
              />
            );
          })}
          <circle cx={head.x} cy={head.y} r={26 + pulse * 5} fill={withAlpha(theme.accentQuantity, 0.22)} />
          <circle cx={head.x} cy={head.y} r={12} fill={theme.accentQuantity} />
          <circle cx={head.x} cy={head.y} r={5} fill={theme.bgBottom} opacity={0.8} />
        </svg>

        {/* The hub. The loop's name until the return, and then the one line this
            template refuses to be planned without. */}
        <div
          style={{
            position: 'absolute',
            left: CX,
            top: CY,
            transform: 'translate(-50%, -50%)',
            width: 320,
            textAlign: 'center',
          }}
        >
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 35,
              fontWeight: 800,
              letterSpacing: -0.8,
              lineHeight: 1.1,
              color: theme.text,
              opacity: returning ? 0.55 : 1,
            }}
          >
            {name}
          </div>
          {returning ? (
            <div
              style={{
                marginTop: 16,
                opacity: interpolate(sinceStep, [Math.max(14, beatFrames * 0.5), Math.max(30, beatFrames * 0.7)], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                }),
                transform: `translateY(${interpolate(
                  sinceStep,
                  [Math.max(14, beatFrames * 0.5), Math.max(30, beatFrames * 0.7)],
                  [10, 0],
                  {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
                )}px)`,
              }}
            >
              <div
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 15,
                  letterSpacing: 3.4,
                  textTransform: 'uppercase',
                  color: theme.accentText,
                }}
              >
                next lap
              </div>
              <div
                style={{
                  fontFamily: theme.fontBody,
                  fontSize: 27,
                  lineHeight: 1.25,
                  color: theme.text,
                  marginTop: 6,
                }}
              >
                {changes}
              </div>
            </div>
          ) : null}
        </div>

        {/* The current stage's line, under the ring. It belongs to the stage
            rather than to the loop, so it lives outside the hub. */}
        {note && !returning ? (
          <div
            style={{
              position: 'absolute',
              left: 0,
              right: 0,
              bottom: 6,
              textAlign: 'center',
              fontFamily: theme.fontBody,
              fontSize: 28,
              color: theme.textMuted,
              opacity: interpolate(sinceStep, [10, 26], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            {note}
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
