import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// SyllabusScene: the course as a route, and the viewer standing on one stop.
//
// Course maps are almost always drawn as a vertical list of modules, which is
// a table of contents — it says what exists, not where you are. This one is
// horizontal, with numbered markers joined by a single line, because a route
// has a DIRECTION and a list does not. Left is behind you, right is ahead, and
// the viewer's position on it is the only thing this template exists to say.
//
// The line is two lines stacked exactly: a full-width dim rail and an accent
// progress line drawn over it, growing to whichever stop is currently in
// focus. Same geometry, different fill — so progress reads as the route being
// travelled rather than as a second graphic appearing. On the closing "here"
// beat it runs to the current module and stops dead, which is the point: the
// rest of the course is visibly there and visibly not done.
//
// Completed stops stamp a tick INSIDE their marker, replacing the number.
// Losing the number is deliberate — a finished module does not need its
// position in the sequence any more, and the difference between numbered and
// ticked markers is what lets a viewer count what remains at a glance.
//
// The focused stop's card lifts ABOVE the route rather than expanding in
// place, so nothing on the line moves when it appears. Cards that push their
// neighbours around make a map that squirms; a map should be a fixed thing you
// point at. The card is clamped inside the route's own bounds so a first or
// last stop's card never runs off the frame.
//
// One glow maximum: the marker of the stop in focus.

const ROUTE_W = Math.min(STAGE_W, 1620);
const CARD_W = 470;
const CARD_ZONE = 190;
const LABEL_ZONE = 104;

type Module = {label: string; sub: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'map' | 'stop' | 'here';
  at?: number;
  ticked: number[];
};

export const SyllabusScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const current = Number(props.current ?? 0);
  const modules = (Array.isArray(props.modules) ? props.modules : []) as Module[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (modules.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const here = step.show === 'here';
  const focused = here ? Math.max(0, Math.min(modules.length - 1, current)) : step.show === 'stop' ? step.at ?? -1 : -1;
  const ticked = new Set(Array.isArray(step.ticked) ? step.ticked : []);

  const dotR = modules.length > 6 ? 25 : 30;
  const inset = dotR + 8;
  const usable = ROUTE_W - inset * 2;
  const nodeX = (i: number): number => (modules.length === 1 ? ROUTE_W / 2 : inset + (usable * i) / (modules.length - 1));

  const lift = spring({frame: sinceStep - 2, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 26});
  const stamp = spring({frame: sinceStep - 3, fps, config: {damping: 12, mass: 0.5}, durationInFrames: 22});
  // The progress line eases to the focused stop from wherever it last was, so
  // moving between stops is a journey and not a jump cut.
  const targetX = focused >= 0 ? nodeX(focused) : 0;
  const prevStep = idx > 0 ? steps[idx - 1] : undefined;
  const prevFocus = prevStep
    ? prevStep.show === 'here'
      ? Math.max(0, Math.min(modules.length - 1, current))
      : prevStep.show === 'stop'
        ? prevStep.at ?? -1
        : -1
    : -1;
  const fromX = prevFocus >= 0 ? nodeX(prevFocus) : 0;
  const progress = focused < 0 ? 0 : interpolate(sinceStep, [2, 24], [fromX, targetX], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const cardLeft = focused >= 0 ? Math.max(0, Math.min(ROUTE_W - CARD_W, nodeX(focused) - CARD_W / 2)) : 0;

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

      <div style={{position: 'relative', width: ROUTE_W, height: CARD_ZONE + dotR * 2 + LABEL_ZONE}}>
        {/* The lifted card, above the line so the line never moves. */}
        {focused >= 0 && modules[focused] ? (
          <div
            style={{
              position: 'absolute',
              left: cardLeft,
              top: CARD_ZONE - 156 - 26 * lift,
              width: CARD_W,
              padding: 24,
              borderRadius: 16,
              background: withAlpha(theme.surface, 0.95),
              border: `2px solid ${withAlpha(theme.accent, 0.35 + 0.35 * lift)}`,
              boxShadow: `8px 10px 0 ${withAlpha(theme.ink, 0.4)}`,
              opacity: lift,
              zIndex: 3,
            }}
          >
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 13,
                letterSpacing: 3,
                textTransform: 'uppercase',
                color: theme.accentText,
                marginBottom: 8,
              }}
            >
              {here ? 'you are here' : `stop ${String(focused + 1).padStart(2, '0')}`}
            </div>
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 34,
                fontWeight: 700,
                letterSpacing: -0.7,
                lineHeight: 1.12,
                color: theme.text,
              }}
            >
              {modules[focused].label}
            </div>
            {modules[focused].sub ? (
              <div style={{fontFamily: theme.fontBody, fontSize: 22, lineHeight: 1.3, color: theme.textMuted, marginTop: 8}}>
                {modules[focused].sub}
              </div>
            ) : null}
            {/* The stem, so the card is attached to its stop and not floating. */}
            <div
              style={{
                position: 'absolute',
                left: Math.max(14, Math.min(CARD_W - 16, nodeX(focused) - cardLeft - 1)),
                top: '100%',
                width: 2.5,
                height: 26 * lift,
                background: withAlpha(theme.accent, 0.6),
              }}
            />
          </div>
        ) : null}

        {/* The rail, and the progress drawn exactly over it. */}
        <div
          style={{
            position: 'absolute',
            left: inset,
            top: CARD_ZONE + dotR - 2,
            width: usable,
            height: 4,
            borderRadius: 2,
            background: withAlpha(theme.line, 0.2),
          }}
        />
        <div
          style={{
            position: 'absolute',
            left: inset,
            top: CARD_ZONE + dotR - 2,
            width: Math.max(0, progress - inset),
            height: 4,
            borderRadius: 2,
            background: withAlpha(theme.accent, 0.9),
          }}
        />

        {modules.map((m, i) => {
          const isFocus = i === focused;
          const done = ticked.has(i);
          const passed = focused >= 0 && i < focused;
          const pop = isFocus ? lift : done ? stamp : 0;
          return (
            <div key={i}>
              <div
                style={{
                  position: 'absolute',
                  left: nodeX(i) - dotR,
                  top: CARD_ZONE,
                  width: dotR * 2,
                  height: dotR * 2,
                  borderRadius: dotR,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: isFocus
                    ? withAlpha(theme.accent, 0.85)
                    : done
                      ? withAlpha(theme.accent, 0.55 * stamp + 0.2)
                      : passed
                        ? withAlpha(theme.surface, 0.95)
                        : withAlpha(theme.surface, 0.85),
                  border: `2.5px solid ${
                    isFocus ? withAlpha(theme.accent, 0.95) : done || passed ? withAlpha(theme.accent, 0.5) : theme.surfaceBorder
                  }`,
                  transform: `scale(${isFocus ? 1 + 0.16 * pop : done ? 0.92 + 0.08 * stamp : 1})`,
                  // The one glow: where the viewer is being told to look.
                  boxShadow: isFocus ? `0 0 ${30 * lift}px ${withAlpha(theme.accent, 0.4)}` : undefined,
                  zIndex: 2,
                }}
              >
                {done ? (
                  <svg width={dotR} height={dotR} viewBox="0 0 16 16" style={{opacity: stamp}}>
                    <path
                      d="M3 8.4 L6.4 11.8 L13 4.6"
                      fill="none"
                      stroke={theme.ink}
                      strokeWidth={2.6}
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                ) : (
                  <span
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: dotR > 26 ? 20 : 17,
                      fontWeight: 700,
                      color: isFocus ? theme.ink : theme.textMuted,
                    }}
                  >
                    {i + 1}
                  </span>
                )}
              </div>

              <div
                style={{
                  position: 'absolute',
                  left: nodeX(i) - usable / (modules.length * 2) - 24,
                  top: CARD_ZONE + dotR * 2 + 16,
                  width: usable / modules.length + 48,
                  textAlign: 'center',
                  fontFamily: theme.fontBody,
                  fontSize: modules.length > 6 ? 20 : 23,
                  lineHeight: 1.2,
                  color: isFocus ? theme.text : done || passed ? theme.textMuted : withAlpha(theme.textMuted, 0.6),
                  opacity: isFocus ? 1 : 0.85,
                }}
              >
                {m.label}
              </div>
            </div>
          );
        })}
      </div>
    </Stage>
  );
};
