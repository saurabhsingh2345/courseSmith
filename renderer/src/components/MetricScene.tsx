import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {counted} from './counted';

// MetricScene is one figure at a time, set large enough to be the whole frame.
//
// Almost nothing is on screen: a label above, the number, its unit, and one
// line under it. That emptiness is the design and not a shortage of ideas — a
// number competing with a diagram is a number nobody reads, and the whole
// premise of the template is that the figure IS the picture.
//
// Three things carry it.
//
// The count-up. A number that fades in is a caption; a number that runs up to
// its value is an event, and the eye follows it to the end. It is deliberately
// fast (under a second) — a slow counter stops being drama and becomes a
// progress bar.
//
// The role colour. Each figure is set in the accent for what it is doing in the
// argument, and those three colours mean the same thing in every clip in the
// course. By the second video the viewer knows the red one is the ceiling
// before the narrator gets there.
//
// The unit's baseline. It sits tight against the figure and small, the way a
// spec sheet sets it, rather than centred beside it — that one detail is most
// of the difference between "designed" and "two divs in a flexbox".

const COL_W = Math.min(STAGE_W, 1500);

type Figure = {
  value: string;
  unit?: string;
  label: string;
  note?: string;
  role?: 'quantity' | 'limit' | 'rival' | 'neutral';
  countsUp?: boolean;
};
type Step = {startMs: number; endMs: number; show: 'state' | 'recap'; at?: number};

/** The accent a figure's role resolves to. Unknown roles read as context. */
const roleColor = (theme: ResolvedTheme, role?: string): string => {
  switch (role) {
    case 'quantity':
      return theme.accentQuantity;
    case 'limit':
      return theme.accentLimit;
    case 'rival':
      return theme.accentRival;
    default:
      return theme.text;
  }
};

export const MetricScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const figures = (Array.isArray(props.figures) ? props.figures : []) as Figure[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (figures.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // The entrance every beat shares: the block lifts and fades in together, so
  // the label, the figure and the note read as one object arriving rather than
  // as three elements animating.
  const enter = spring({
    frame: sinceStep,
    fps,
    config: {damping: 200, mass: 0.6},
    durationInFrames: 14,
  });

  if (step.show === 'recap') {
    return <Recap theme={theme} figures={figures} enter={enter} />;
  }

  const fig = figures[Math.max(0, Math.min(figures.length - 1, step.at ?? 0))];
  const color = roleColor(theme, fig.role);
  // Under a second, and eased hard at the end so it lands rather than creeps.
  const run = interpolate(sinceStep, [4, 26], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <Stage justify="center">
      <div
        style={{
          width: COL_W,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          opacity: enter,
          transform: `translateY(${(1 - enter) * 22}px)`,
        }}
      >
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 21,
            letterSpacing: 5,
            textTransform: 'uppercase',
            color: theme.textMuted,
            marginBottom: 26,
            textAlign: 'center',
          }}
        >
          {fig.label}
        </div>

        {/* The figure and its unit share a baseline, with the unit small and
            tight against the number the way a spec sheet sets it. */}
        <div style={{display: 'flex', alignItems: 'baseline', gap: 14}}>
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 250,
              fontWeight: 800,
              letterSpacing: -8,
              lineHeight: 0.92,
              color,
              // Tabular figures, or the number jitters horizontally on every
              // frame of the count-up as glyph widths change.
              fontVariantNumeric: 'tabular-nums',
            }}
          >
            {counted(fig.value, fig.countsUp === true, run)}
          </div>
          {fig.unit ? (
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 46,
                fontWeight: 500,
                letterSpacing: 1,
                color: theme.textMuted,
              }}
            >
              {fig.unit}
            </div>
          ) : null}
        </div>

        {fig.note ? (
          <div
            style={{
              fontFamily: theme.fontBody,
              fontSize: 31,
              lineHeight: 1.4,
              color: theme.textMuted,
              textAlign: 'center',
              maxWidth: 1080,
              marginTop: 34,
              // Held back a few frames behind the figure, so the eye reads the
              // number first and the argument second.
              opacity: interpolate(sinceStep, [16, 30], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            {fig.note}
          </div>
        ) : null}
      </div>
    </Stage>
  );
};

/**
 * The closing frame: every figure back at once, in its own role colour.
 *
 * Set as a row rather than restated one by one, because the job here is not to
 * teach the numbers again — it is to let the viewer see the shape of the
 * argument in a single glance, which is also why this is the frame people
 * screenshot.
 */
const Recap: React.FC<{theme: ResolvedTheme; figures: Figure[]; enter: number}> = ({
  theme,
  figures,
  enter,
}) => {
  const frame = useCurrentFrame();
  return (
    <Stage justify="center">
      <div
        style={{
          width: COL_W,
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'flex-start',
          gap: 54,
          flexWrap: 'wrap',
          opacity: enter,
        }}
      >
        {figures.map((f, i) => {
          // Staggered, so the row assembles left to right instead of appearing.
          const on = interpolate(frame, [6 + i * 5, 20 + i * 5], [0, 1], {
            extrapolateLeft: 'clamp',
            extrapolateRight: 'clamp',
          });
          return (
            <div
              key={i}
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                maxWidth: 300,
                opacity: on,
                transform: `translateY(${(1 - on) * 14}px)`,
              }}
            >
              <div style={{display: 'flex', alignItems: 'baseline', gap: 6}}>
                <div
                  style={{
                    fontFamily: theme.fontDisplay,
                    fontSize: 104,
                    fontWeight: 800,
                    letterSpacing: -3,
                    lineHeight: 1,
                    color: roleColor(theme, f.role),
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {f.value}
                </div>
                {f.unit ? (
                  <div
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: 24,
                      color: theme.textMuted,
                    }}
                  >
                    {f.unit}
                  </div>
                ) : null}
              </div>
              <div
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 16,
                  letterSpacing: 2.4,
                  textTransform: 'uppercase',
                  color: theme.textMuted,
                  textAlign: 'center',
                  marginTop: 16,
                  lineHeight: 1.5,
                }}
              >
                {f.label}
              </div>
            </div>
          );
        })}
      </div>
    </Stage>
  );
};
