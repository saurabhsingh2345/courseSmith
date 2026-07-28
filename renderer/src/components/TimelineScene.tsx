import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_W} from './Stage';
import {FIGURE_BOX, figureFor, type FigurePalette} from './artwork';

// TimelineScene is one spine that fills in as the narration walks it.
//
// The spine is on screen for the whole clip and every milestone keeps its dot
// from the first frame — what moves is how far the line has *filled*, and which
// stop is current. Revealing the dots one at a time was the obvious first
// design and it is wrong: a timeline whose future is invisible is a list, and
// the reason to draw an axis at all is that you can see where you are going.
//
// Milestones alternate above and below the line. On a six-stop run the labels
// are 200px apart, and stacking them all on one side put a title and the next
// mark in the same 40px of frame.

const COL_W = Math.min(STAGE_W, 1620);
const SPINE_Y = 250;
const DOT_R = 13;

type Milestone = {mark: string; title: string; note: string; figure?: string};
type Step = {startMs: number; endMs: number; at?: number; whole?: boolean};

export const TimelineScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const title = String(props.title ?? '');
  const milestones = (Array.isArray(props.milestones) ? props.milestones : []) as Milestone[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];

  if (milestones.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const whole = step.whole === true;
  const current = whole ? milestones.length - 1 : (step.at ?? 0);
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // Dot positions, inset so the first and last labels have room either side.
  const inset = COL_W * 0.08;
  const span = COL_W - inset * 2;
  const xOf = (i: number) =>
    inset + (milestones.length === 1 ? span / 2 : (span * i) / (milestones.length - 1));

  // How far the spine has filled. It eases *within* the current beat rather
  // than snapping, so the walk reads as travel rather than as a cursor jumping
  // between stops.
  const reach = spring({
    frame: sinceStep,
    fps,
    config: {damping: 200, mass: 0.8},
    durationInFrames: 22,
  });
  const prevAt = idx > 0 ? (steps[idx - 1].whole ? milestones.length - 1 : (steps[idx - 1].at ?? 0)) : 0;
  const filledX = xOf(prevAt) + (xOf(current) - xOf(prevAt)) * reach;

  const palette: FigurePalette = {
    accent: theme.accent,
    primary: theme.primary,
    ink: theme.ink,
    soft: theme.mass,
    line: theme.line,
  };

  const active = milestones[current];

  return (
    <Stage justify="center">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={18} />

      <div style={{width: COL_W, height: 520, position: 'relative'}}>
        {/* The spine: the whole run in muted, the walked part in accent. */}
        <div
          style={{
            position: 'absolute',
            left: inset,
            top: SPINE_Y,
            width: span,
            height: 4,
            borderRadius: 2,
            backgroundColor: theme.textMuted,
            opacity: 0.26,
          }}
        />
        <div
          style={{
            position: 'absolute',
            left: inset,
            top: SPINE_Y,
            width: Math.max(0, filledX - inset),
            height: 4,
            borderRadius: 2,
            backgroundColor: theme.accent,
          }}
        />

        {milestones.map((m, i) => {
          const x = xOf(i);
          const reached = i <= current;
          const isCurrent = !whole && i === current;
          const above = i % 2 === 0;
          const Figure = figureFor(m.figure);
          // Every figure is fully assembled, reached or not. Building only the
          // reached ones left the stops ahead showing a label with an empty
          // space where their figure should be — which reads as a bug, and
          // contradicts the whole reason this scene draws the future at all.
          // What separates past from future is opacity; what separates the
          // current stop from the rest is that its idle is the only one running.

          return (
            <div key={i}>
              {/* Figure, on the opposite side to the text. */}
              <div
                style={{
                  position: 'absolute',
                  left: x - 46,
                  top: above ? SPINE_Y + 46 : SPINE_Y - 138,
                  opacity: reached ? (isCurrent ? 1 : 0.5) : 0.18,
                }}
              >
                <svg width={92} height={92} viewBox={`0 0 ${FIGURE_BOX} ${FIGURE_BOX}`} style={{overflow: 'visible'}}>
                  <Figure build={1} t={isCurrent ? frame / FPS : 0} palette={palette} />
                </svg>
              </div>

              {/* The dot. Present from the first frame whether reached or not —
                  a timeline whose future is invisible is a list. */}
              <div
                style={{
                  position: 'absolute',
                  left: x - DOT_R,
                  top: SPINE_Y + 2 - DOT_R,
                  width: DOT_R * 2,
                  height: DOT_R * 2,
                  borderRadius: DOT_R,
                  backgroundColor: reached ? theme.accent : theme.bgBottom,
                  border: `3px solid ${reached ? theme.accent : theme.textMuted}`,
                  opacity: reached ? 1 : 0.4,
                  transform: `scale(${isCurrent ? 1.45 : 1})`,
                }}
              />

              {/* Mark and title, alternating sides of the line. */}
              <div
                style={{
                  position: 'absolute',
                  left: x - 130,
                  width: 260,
                  top: above ? SPINE_Y - 116 : SPINE_Y + 44,
                  textAlign: 'center',
                  opacity: reached ? 1 : 0.34,
                }}
              >
                <div
                  style={{
                    fontFamily: theme.fontMono,
                    fontSize: 26,
                    fontWeight: 700,
                    color: isCurrent ? theme.accentText : theme.textMuted,
                  }}
                >
                  {m.mark}
                </div>
                <div
                  style={{
                    marginTop: 6,
                    fontFamily: theme.fontDisplay,
                    fontSize: 28,
                    fontWeight: 600,
                    lineHeight: 1.2,
                    color: theme.text,
                  }}
                >
                  {m.title}
                </div>
              </div>
            </div>
          );
        })}

        {/* The current stop's note, on its own line under everything. Only one
            is ever up, so the frame never holds two sentences to read. */}
        {!whole && active?.note && (
          <div
            style={{
              position: 'absolute',
              left: 0,
              top: SPINE_Y + 210,
              width: COL_W,
              textAlign: 'center',
              fontFamily: theme.fontBody,
              fontSize: 30,
              lineHeight: 1.35,
              color: theme.textMuted,
              opacity: interpolate(sinceStep, [0, 12], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            {active.note}
          </div>
        )}
      </div>
    </Stage>
  );
};
