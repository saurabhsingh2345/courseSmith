import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_H} from './Stage';
import {SceneHeader} from './SceneHeader';

// RegionsScene: the address space as one tall column.
//
// It is vertical for a reason that is not taste. The two interesting segments
// MOVE, and they move toward each other: the heap upward, the stack downward,
// both spending the same unclaimed band between them. A horizontal strip can
// show that a program has segments; only a column oriented the way addresses
// are conventionally drawn — low at the bottom, high at the top — can show two
// fronts closing on one budget. Everything else here is in service of that.
//
// The blocks stack FLUSH, with hairline boundaries rather than gutters. An
// address space has no space between its segments, and a gutter would read as
// free room, which is exactly the thing the gap block is supposed to be the
// only instance of. The gap is drawn hatched rather than filled for the same
// reason: it must look like absence, not like another tenant.
//
// The ticks down the left are unlabelled on purpose. A plan gives labels and
// roles, never addresses, so printing "0x7FFF..." beside a band would be the
// component inventing the one kind of content this catalog refuses to invent.
// What the rail carries instead is the only thing that is actually known — the
// direction of the axis — anchored at both ends in mono.
//
// Growth is a block edge advancing, with an arrowhead riding it, and it is
// permanent: once a front has moved it stays moved, because a diagram that
// snaps back has said the allocation was undone. Each front eats 40% of the
// gap, so two growths leave a visible sliver — and the collide beat is the one
// state where both go to 50% and the sliver closes. That closing line is the
// single glow in the frame, in accentLimit, because it is a ceiling being hit.

const COL_W = 380;
const RAIL_W = 118;
const NOTE_W = 560;
const COL_H = Math.min(STAGE_H - 190, 660);
// The gap is the stage the growth plays on, so it is given the most height; the
// two segments that move get a little more than the fixed ones so their edges
// have somewhere to leave from.
const WEIGHT: Record<string, number> = {gap: 2.4, heap: 1.35, stack: 1.35, code: 1, static: 1};
// How much of the gap one front takes on a normal growth beat, and how much it
// takes when the two are made to meet.
const GROW_FRAC = 0.4;
const MEET_FRAC = 0.5;

type Region = {label: string; role: string; note: string; grows: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'map' | 'region' | 'grow' | 'collide' | 'whole';
  at?: number;
  seen: number[];
  grown: number[];
  collided: boolean;
};

type Band = {top: number; h: number};

// Blocks are laid out from the BOTTOM up, because index 0 is the lowest
// address. The returned tops are in the column's own coordinates, which run
// downward, so the last region ends up at top 0.
const layout = (regions: Region[]): Band[] => {
  const weights = regions.map((r) => WEIGHT[r.role] ?? 1);
  const total = weights.reduce((a, b) => a + b, 0) || 1;
  const bands: Band[] = [];
  let bottom = COL_H;
  regions.forEach((_, i) => {
    const h = (weights[i] / total) * COL_H;
    bands[i] = {top: bottom - h, h};
    bottom -= h;
  });
  return bands;
};

export const RegionsScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const regions = (Array.isArray(props.regions) ? props.regions : []) as Region[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const heapAt = Number(props.heapAt ?? -1);
  const stackAt = Number(props.stackAt ?? -1);
  const gapAt = Number(props.gapAt ?? -1);
  if (regions.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const bands = layout(regions);
  const seen = new Set(step.seen ?? []);
  const grown = new Set(step.grown ?? []);
  const focused = step.show === 'region' || step.show === 'grow' ? step.at ?? -1 : -1;
  const lit = step.show === 'whole';

  // The arrival spring, reused by whatever this beat is doing: a band coming
  // forward, an edge advancing, the fronts meeting.
  const arrive = spring({frame: sinceStep - 2, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 26});

  // How far each front has advanced into the gap, as a fraction of the gap's
  // height. Already-grown fronts hold their ground; the one growing right now
  // springs out; a collision pulls both to the meeting line.
  const advance = (at: number): number => {
    if (at < 0) return 0;
    const settled = grown.has(at) ? GROW_FRAC : 0;
    if (step.collided) {
      const base = step.show === 'collide' ? settled + (MEET_FRAC - settled) * arrive : MEET_FRAC;
      return Math.max(settled, base);
    }
    if (step.show === 'grow' && step.at === at) {
      return GROW_FRAC * arrive;
    }
    return settled;
  };

  const gapBand = gapAt >= 0 ? bands[gapAt] : null;
  const heapAdvance = advance(heapAt);
  const stackAdvance = advance(stackAt);
  const meeting = step.collided && gapBand ? gapBand.top + gapBand.h / 2 : 0;
  const flash = step.show === 'collide' ? interpolate(sinceStep % 26, [0, 13, 26], [1, 0.35, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'}) : 1;

  const noteFor = focused >= 0 ? regions[focused] : null;
  const noteBand = focused >= 0 ? bands[focused] : null;

  const fillFor = (role: string, active: boolean): string => {
    if (role === 'gap') return 'transparent';
    if (role === 'heap') return withAlpha(theme.accentQuantity, active ? 0.22 : 0.12);
    if (role === 'stack') return withAlpha(theme.accentRival, active ? 0.22 : 0.12);
    return withAlpha(theme.mass, active ? 0.16 : 0.08);
  };
  const edgeFor = (role: string): string =>
    role === 'heap' ? theme.accentQuantity : role === 'stack' ? theme.accentRival : theme.accent;

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

      <div style={{display: 'flex', alignItems: 'flex-start', gap: 0}}>
        {/* The address rail: direction of the axis, and a tick per boundary. */}
        <div style={{width: RAIL_W, height: COL_H, position: 'relative', flexShrink: 0}}>
          <div
            style={{
              position: 'absolute',
              right: 0,
              top: 0,
              width: 2,
              height: COL_H,
              background: withAlpha(theme.line, 0.35),
            }}
          />
          {bands.map((b, i) => (
            <div
              key={i}
              style={{
                position: 'absolute',
                right: 0,
                top: b.top,
                width: seen.has(i) ? 26 : 16,
                height: 2,
                background: withAlpha(theme.line, seen.has(i) ? 0.7 : 0.35),
              }}
            />
          ))}
          <div
            style={{
              position: 'absolute',
              right: 36,
              top: -4,
              fontFamily: theme.fontMono,
              fontSize: 14,
              letterSpacing: 2.2,
              textTransform: 'uppercase',
              color: theme.textMuted,
              textAlign: 'right',
              width: RAIL_W - 40,
            }}
          >
            high
          </div>
          <div
            style={{
              position: 'absolute',
              right: 36,
              top: COL_H - 16,
              fontFamily: theme.fontMono,
              fontSize: 14,
              letterSpacing: 2.2,
              textTransform: 'uppercase',
              color: theme.textMuted,
              textAlign: 'right',
              width: RAIL_W - 40,
            }}
          >
            low
          </div>
        </div>

        {/* The column itself. */}
        <div style={{width: COL_W, height: COL_H, position: 'relative', flexShrink: 0}}>
          {regions.map((r, i) => {
            const b = bands[i];
            const active = lit || i === focused;
            const isGap = r.role === 'gap';
            return (
              <div
                key={i}
                style={{
                  position: 'absolute',
                  left: 0,
                  top: b.top,
                  width: COL_W,
                  height: b.h,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: isGap
                    ? `repeating-linear-gradient(135deg, ${withAlpha(theme.line, 0.16)} 0 2px, transparent 2px 12px)`
                    : fillFor(r.role, active),
                  borderTop: `1.5px solid ${withAlpha(theme.line, 0.45)}`,
                  borderLeft: `2px solid ${active ? edgeFor(r.role) : withAlpha(theme.surfaceBorder, 0.9)}`,
                  borderRight: `2px solid ${withAlpha(theme.surfaceBorder, 0.9)}`,
                  borderBottom: i === 0 ? `1.5px solid ${withAlpha(theme.line, 0.45)}` : undefined,
                  opacity: lit ? 1 : active ? 1 : seen.has(i) ? 0.72 : 0.5,
                  transform: `translateX(${active && !lit ? 8 * arrive : 0}px)`,
                }}
              >
                <span
                  style={{
                    fontFamily: theme.fontDisplay,
                    fontSize: b.h < 80 ? 24 : 30,
                    fontWeight: 700,
                    letterSpacing: -0.4,
                    color: isGap ? theme.textMuted : active ? theme.text : theme.textMuted,
                    zIndex: 2,
                  }}
                >
                  {r.label}
                </span>
              </div>
            );
          })}

          {/* The heap's front, advancing upward out of its own top edge. */}
          {gapBand && heapAt >= 0 && heapAdvance > 0 ? (
            <div
              style={{
                position: 'absolute',
                left: 2,
                width: COL_W - 4,
                top: gapBand.top + gapBand.h * (1 - heapAdvance),
                height: gapBand.h * heapAdvance,
                background: withAlpha(theme.accentQuantity, 0.3),
                borderTop: `3px solid ${theme.accentQuantity}`,
              }}
            >
              <svg width={26} height={16} style={{position: 'absolute', left: '50%', top: -18, marginLeft: -13, overflow: 'visible'}}>
                <polygon points="13,0 26,15 0,15" fill={theme.accentQuantity} />
              </svg>
            </div>
          ) : null}

          {/* The stack's front, advancing downward out of its own bottom edge. */}
          {gapBand && stackAt >= 0 && stackAdvance > 0 ? (
            <div
              style={{
                position: 'absolute',
                left: 2,
                width: COL_W - 4,
                top: gapBand.top,
                height: gapBand.h * stackAdvance,
                background: withAlpha(theme.accentRival, 0.3),
                borderBottom: `3px solid ${theme.accentRival}`,
              }}
            >
              <svg width={26} height={16} style={{position: 'absolute', left: '50%', bottom: -18, marginLeft: -13, overflow: 'visible'}}>
                <polygon points="13,16 0,1 26,1" fill={theme.accentRival} />
              </svg>
            </div>
          ) : null}

          {/* The meeting line. The one glow, and only when they have met. */}
          {step.collided && gapBand ? (
            <div
              style={{
                position: 'absolute',
                left: -14,
                width: COL_W + 28,
                top: meeting - 2,
                height: 4,
                background: theme.accentLimit,
                opacity: flash,
                boxShadow: `0 0 26px ${withAlpha(theme.accentLimit, 0.65)}`,
              }}
            />
          ) : null}
        </div>

        {/* The note rail: whatever the focused band has to say, or the verdict. */}
        <div style={{width: NOTE_W, height: COL_H, position: 'relative', flexShrink: 0}}>
          {noteFor && noteBand ? (
            <div
              style={{
                position: 'absolute',
                left: 34,
                top: Math.max(0, Math.min(COL_H - 110, noteBand.top + noteBand.h / 2 - 52)),
                width: NOTE_W - 40,
                opacity: arrive,
                transform: `translateX(${(1 - arrive) * -16}px)`,
              }}
            >
              <div
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 15,
                  letterSpacing: 3,
                  textTransform: 'uppercase',
                  color: theme.accentText,
                  marginBottom: 10,
                }}
              >
                {noteFor.role}
                {noteFor.grows === 'up' ? ' / grows up' : noteFor.grows === 'down' ? ' / grows down' : ''}
              </div>
              <div
                style={{
                  fontFamily: theme.fontBody,
                  fontSize: 30,
                  lineHeight: 1.34,
                  color: theme.text,
                }}
              >
                {noteFor.note || noteFor.label}
              </div>
            </div>
          ) : null}

          {step.show === 'collide' && gapBand ? (
            <div
              style={{
                position: 'absolute',
                left: 34,
                top: Math.max(0, gapBand.top + gapBand.h / 2 - 44),
                width: NOTE_W - 40,
                opacity: arrive,
                transform: `translateX(${(1 - arrive) * -16}px)`,
              }}
            >
              <div
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 15,
                  letterSpacing: 3,
                  textTransform: 'uppercase',
                  color: theme.accentLimit,
                  marginBottom: 10,
                }}
              >
                out of room
              </div>
              <div style={{fontFamily: theme.fontBody, fontSize: 30, lineHeight: 1.34, color: theme.text}}>
                The two fronts have spent the same free space, and there is none left between them.
              </div>
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
