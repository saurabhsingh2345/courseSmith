import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// BridgeScene: what you already have, and the one thing it cannot answer yet.
//
// The hand-off between two lessons is the moment a course either coheres or
// becomes a playlist, and it is almost always narrated and never drawn. So it
// gets a picture with exactly two sides and a physical span between them: the
// lesson behind on the left with its established results stacked, the lesson
// ahead on the right, and a chevron pointing across the gap.
//
// The left column is the load-bearing idea. Its items do not simply appear —
// they TICK and then step back, dimming as they are carried, because a thing
// carried forward is a thing you no longer have to hold. Watching the left
// side quiet down as the right side fills is the whole emotional shape of a
// prerequisite, and it costs nothing but an opacity ramp.
//
// The gap renders as a dashed outline slot: an empty rectangle with the
// question set inside it, drawn in the same size and position the answer will
// occupy. That is deliberate. A dashed box is the universal sign for "this is
// where something goes", so when the closer lands the slot does not get
// replaced, it gets FILLED — same geometry, solid border, accent ground. The
// viewer sees a socket accept a plug rather than one card swap for another.
//
// One glow maximum: the filled slot on the closing beat. The chevron brightens
// but never lights, because the chevron is a road sign and road signs do not
// glow.

const BOARD_W = Math.min(STAGE_W, 1500);
const CHEVRON_W = 118;
const COL_W = Math.floor((BOARD_W - CHEVRON_W) / 2) - 26;
const ITEM_H = 74;

type Step = {
  startMs: number;
  endMs: number;
  show: 'back' | 'carry' | 'gap' | 'ahead';
  at?: number;
  carried: number[];
  gapOpen: boolean;
  arrived: boolean;
};

export const BridgeScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const from = String(props.from ?? '');
  const to = String(props.to ?? '');
  const gap = String(props.gap ?? '');
  const established = (Array.isArray(props.established) ? props.established : []) as string[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (established.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const carrying = step.show === 'carry' ? step.at ?? -1 : -1;
  const carried = new Set(Array.isArray(step.carried) ? step.carried : []);
  const gapOpen = Boolean(step.gapOpen);
  const arrived = Boolean(step.arrived);

  const tick = spring({frame: sinceStep - 2, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 24});
  const land = arrived
    ? spring({frame: sinceStep - 4, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 28})
    : 0;
  const slotOpen = gapOpen
    ? step.show === 'gap'
      ? spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.55}, durationInFrames: 26})
      : 1
    : 0;
  // The chevron leans across as soon as anything has been carried, and firms
  // up when the closer lands.
  const span = carried.size > 0 || gapOpen ? 1 : interpolate(sinceStep, [6, 22], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});

  const columnLabel = (text: string, colour: string): React.ReactNode => (
    <div
      style={{
        fontFamily: theme.fontMono,
        fontSize: 15,
        letterSpacing: 3.2,
        textTransform: 'uppercase',
        color: colour,
        marginBottom: 10,
      }}
    >
      {text}
    </div>
  );

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

      <div style={{width: BOARD_W, display: 'flex', alignItems: 'flex-start', gap: 26}}>
        {/* Behind: what the last lesson left standing. */}
        <div style={{width: COL_W, flexShrink: 0}}>
          {columnLabel('behind you', theme.textMuted)}
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 38,
              fontWeight: 700,
              letterSpacing: -0.8,
              lineHeight: 1.1,
              color: carried.size === established.length ? theme.textMuted : theme.text,
              marginBottom: 22,
            }}
          >
            {from}
          </div>
          <div style={{display: 'flex', flexDirection: 'column', gap: 12}}>
            {established.map((item, i) => {
              const isTicking = i === carrying;
              const done = carried.has(i);
              const pop = isTicking ? tick : 1;
              return (
                <div
                  key={i}
                  style={{
                    minHeight: ITEM_H,
                    display: 'flex',
                    alignItems: 'center',
                    gap: 14,
                    paddingInline: 18,
                    paddingBlock: 12,
                    borderRadius: 12,
                    background: withAlpha(theme.surface, done ? 0.45 : 0.85),
                    border: `2px solid ${isTicking ? withAlpha(theme.accent, 0.6) : done ? withAlpha(theme.accent, 0.24) : theme.surfaceBorder}`,
                    // Carried items step back: the work is done, it is not
                    // gone, and it is no longer the subject.
                    opacity: done && !isTicking ? 0.5 : 1,
                    transform: `translateX(${isTicking ? 8 * pop : 0}px)`,
                  }}
                >
                  <div
                    style={{
                      width: 28,
                      height: 28,
                      flexShrink: 0,
                      borderRadius: 14,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      background: done ? withAlpha(theme.accent, 0.9 * (isTicking ? pop : 1)) : 'transparent',
                      border: `2px solid ${done ? withAlpha(theme.accent, 0.9) : theme.surfaceBorder}`,
                      transform: `scale(${done ? 0.8 + 0.2 * (isTicking ? pop : 1) : 1})`,
                    }}
                  >
                    <svg width={16} height={16} viewBox="0 0 16 16" style={{opacity: done ? (isTicking ? pop : 1) : 0}}>
                      <path
                        d="M3 8.4 L6.4 11.8 L13 4.6"
                        fill="none"
                        stroke={theme.ink}
                        strokeWidth={2.4}
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                  </div>
                  <span
                    style={{
                      fontFamily: theme.fontBody,
                      fontSize: 24,
                      lineHeight: 1.25,
                      color: done && !isTicking ? theme.textMuted : theme.text,
                      textDecoration: 'none',
                    }}
                  >
                    {item}
                  </span>
                </div>
              );
            })}
          </div>
        </div>

        {/* The span itself. */}
        <div
          style={{
            width: CHEVRON_W,
            flexShrink: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            paddingTop: 132,
          }}
        >
          <svg width={CHEVRON_W} height={96} viewBox="0 0 118 96" style={{overflow: 'visible'}}>
            <line
              x1={4}
              y1={48}
              x2={4 + 84 * span}
              y2={48}
              stroke={withAlpha(theme.accent, arrived ? 0.85 : 0.4)}
              strokeWidth={3}
              strokeLinecap="round"
              strokeDasharray="3 9"
            />
            <path
              d="M76 26 L106 48 L76 70"
              fill="none"
              stroke={withAlpha(theme.accent, arrived ? 0.95 : 0.45 + 0.3 * span)}
              strokeWidth={5}
              strokeLinecap="round"
              strokeLinejoin="round"
              transform={`translate(${(1 - span) * -18}, 0)`}
            />
          </svg>
        </div>

        {/* Ahead: the lesson, and the socket its answer drops into. */}
        <div style={{width: COL_W, flexShrink: 0}}>
          {columnLabel('ahead', arrived ? theme.accentText : theme.textMuted)}
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 38,
              fontWeight: 700,
              letterSpacing: -0.8,
              lineHeight: 1.1,
              color: arrived ? theme.text : theme.textMuted,
              marginBottom: 22,
              opacity: 0.45 + 0.55 * Math.max(span, arrived ? 1 : 0),
            }}
          >
            {to}
          </div>

          {/* Same geometry either way: a socket, then a socket with something
              in it. Nothing is replaced, so nothing pops. */}
          <div
            style={{
              minHeight: 168,
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              gap: 12,
              padding: 26,
              borderRadius: 16,
              background: arrived ? withAlpha(theme.accent, 0.12 * land) : 'transparent',
              // Same width and radius either way: only the stroke style and
              // the ground change, so the socket is filled rather than swapped.
              border: `2.5px ${arrived ? 'solid' : 'dashed'} ${
                arrived ? withAlpha(theme.accent, 0.45 + 0.45 * land) : withAlpha(theme.line, 0.45 * slotOpen)
              }`,
              opacity: slotOpen,
              transform: `translateY(${(1 - slotOpen) * 14}px) scale(${arrived ? 0.985 + 0.015 * land : 1})`,
              // The one glow, on the one beat that earns it.
              boxShadow: arrived ? `0 0 ${38 * land}px ${withAlpha(theme.accent, 0.3)}` : undefined,
            }}
          >
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 14,
                letterSpacing: 3,
                textTransform: 'uppercase',
                color: arrived ? theme.accentText : withAlpha(theme.textMuted, 0.8),
              }}
            >
              {arrived ? 'answered here' : 'open question'}
            </div>
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 32,
                fontWeight: 600,
                letterSpacing: -0.5,
                lineHeight: 1.2,
                color: arrived ? theme.text : theme.textMuted,
              }}
            >
              {gap}
            </div>
          </div>
        </div>
      </div>
    </Stage>
  );
};
