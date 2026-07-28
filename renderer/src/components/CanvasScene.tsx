import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_W} from './Stage';
import {iconFor} from './icons';

// CanvasScene is an automation wired across a builder's canvas.
//
// The whole chain is on screen from the first frame and every card keeps its
// place — what moves is which card is *current*, and, on the closing beat, a
// token carrying the payload from the first card to the last. This is the same
// choice TimelineScene makes about its future stops and for the same reason: a
// workflow whose remaining steps are invisible is a list, and the reason to draw
// a canvas is that you can see where the record is going before it gets there.
//
// The dotted grid, the ports on each card's edges and the wires between them are
// not decoration — they are the three marks that make a viewer recognise this as
// a builder rather than as a row of slides, and every no-code tool a lesson
// could be about draws all three.

const COL_W = Math.min(STAGE_W, 1700);
const PANEL_PAD = 46;
const CARD_H = 232;
const PORT_R = 7;
// The token rides above the row rather than on the wire, because a pill wide
// enough to read the payload in is wider than the gap between two cards — put
// it on the wire and it lands on whichever card it is passing. TOKEN_LIFT is
// how far above the card tops it sits, and the panel is sized to hold it.
const TOKEN_LIFT = 62;
const TOKEN_H = 46;
// Panel height: the row, plus room for the token above it and the same margin
// below so the cards sit centred. Sizing this to the stage instead left a band
// of empty grid top and bottom that read as a layout that had lost its content.
const PANEL_H = CARD_H + (TOKEN_LIFT + TOKEN_H) * 2;

type Node = {app?: string; title: string; kind: string; icon?: string; note?: string};
type Step = {startMs: number; endMs: number; at?: number; run?: boolean};

/** Column gap, widened when there is room. Five cards need every pixel. */
const gapFor = (n: number) => (n <= 3 ? 76 : n === 4 ? 56 : 42);

/** Smoothstep, for a token that leaves and arrives rather than sliding at a constant rate. */
const smooth = (x: number) => x * x * (3 - 2 * x);
const clamp01 = (x: number) => Math.min(1, Math.max(0, x));

export const CanvasScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const title = String(props.title ?? '');
  const payload = String(props.payload ?? '');
  const nodes = (Array.isArray(props.nodes) ? props.nodes : []) as Node[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];

  if (nodes.length === 0 || steps.length === 0) return null;

  const n = nodes.length;
  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const running = step.run === true;
  const current = running ? n - 1 : (step.at ?? 0);
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // Geometry. Cards sit in one row inside the panel; everything else is
  // measured off their centres.
  const inner = COL_W - PANEL_PAD * 2;
  const gap = gapFor(n);
  const cardW = (inner - gap * (n - 1)) / n;
  const cardY = (PANEL_H - CARD_H) / 2;
  const wireY = cardY + CARD_H / 2;
  const leftOf = (i: number) => PANEL_PAD + i * (cardW + gap);
  const centreOf = (i: number) => leftOf(i) + cardW / 2;

  // The run. A lead-in and a tail keep the token from starting on the beat's
  // first frame and finishing on its last, so the fire reads as an event rather
  // than as something already in progress when the beat cut to it.
  const runP = running
    ? clamp01((nowMs - step.startMs) / Math.max(1, step.endMs - step.startMs))
    : 0;
  const travel = clamp01((runP - 0.12) / 0.68);
  let tokenX = centreOf(0);
  if (running && n > 1) {
    const t = travel * (n - 1);
    const seg = Math.min(n - 2, Math.floor(t));
    const frac = t - seg;
    // Dwell on each card before moving off it — the pause is where the viewer
    // reads what that step did.
    const e = smooth(clamp01((frac - 0.38) / 0.62));
    tokenX = centreOf(seg) + (centreOf(seg + 1) - centreOf(seg)) * e;
  }
  /** 1 while the token is over card i, falling off either side of it. */
  const hotness = (i: number) =>
    running ? clamp01(1 - Math.abs(tokenX - centreOf(i)) / (cardW * 0.55)) : 0;
  const passed = (i: number) => running && tokenX >= centreOf(i) - 2;

  // The lift a card gets when the narration arrives on it.
  const lift = spring({frame: sinceStep, fps, config: {damping: 200, mass: 0.7}, durationInFrames: 20});

  const kindTint = (kind: string) =>
    kind === 'trigger' || kind === 'output' ? theme.accent : theme.primary;

  const active = nodes[current];

  return (
    <Stage justify="center">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={16} />

      {/* The canvas itself: a dotted grid inside a hairline frame. */}
      <div
        style={{
          width: COL_W,
          height: PANEL_H,
          position: 'relative',
          borderRadius: 30,
          border: `1px solid ${withAlpha(theme.surfaceBorder, 0.9)}`,
          backgroundColor: withAlpha(theme.surface, 0.42),
          backgroundImage: `radial-gradient(${withAlpha(theme.surfaceBorder, 0.85)} 1.6px, transparent 1.6px)`,
          backgroundSize: '30px 30px',
          backgroundPosition: '18px 18px',
          overflow: 'hidden',
        }}
      >
        {/* A soft pool of accent under whatever the beat is about, so the eye
            lands before it has read anything. */}
        <div
          style={{
            position: 'absolute',
            left: centreOf(current) - 300,
            top: wireY - 300,
            width: 600,
            height: 600,
            borderRadius: 300,
            background: `radial-gradient(circle, ${withAlpha(theme.accent, running ? 0.1 : 0.14)} 0%, transparent 62%)`,
          }}
        />

        {/* Wires. Drawn under the cards so the ports appear to sit on them. */}
        <svg width={COL_W} height={PANEL_H} style={{position: 'absolute', left: 0, top: 0}}>
          {nodes.slice(0, -1).map((_, i) => {
            const x1 = leftOf(i) + cardW;
            const x2 = leftOf(i + 1);
            const live = running ? tokenX >= x1 : i + 1 <= current;
            return (
              <g key={i}>
                <line
                  x1={x1}
                  y1={wireY}
                  x2={x2}
                  y2={wireY}
                  stroke={theme.line}
                  strokeWidth={3}
                  strokeLinecap="round"
                  strokeDasharray="7 9"
                  opacity={0.4}
                />
                <line
                  x1={x1}
                  y1={wireY}
                  x2={live ? x2 : x1}
                  y2={wireY}
                  stroke={theme.accent}
                  strokeWidth={3.5}
                  strokeLinecap="round"
                />
                {/* The chevron says which way it flows even in a still frame. */}
                <path
                  d={`M ${(x1 + x2) / 2 - 5} ${wireY - 7} L ${(x1 + x2) / 2 + 6} ${wireY} L ${(x1 + x2) / 2 - 5} ${wireY + 7}`}
                  fill="none"
                  stroke={live ? theme.accent : theme.line}
                  strokeWidth={3}
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  opacity={live ? 1 : 0.5}
                />
              </g>
            );
          })}
        </svg>

        {nodes.map((node, i) => {
          const reached = running || i <= current;
          const isCurrent = !running && i === current;
          const hot = hotness(i);
          const done = passed(i);
          const Icon = iconFor(node.icon);
          const tint = kindTint(node.kind);
          // Two things raise a card: the narration arriving on it, and the token
          // reaching it. They never happen at once, so one transform carries both.
          const raise = (isCurrent ? lift * 10 : 0) + hot * 12;
          const litness = isCurrent || hot > 0.15 ? 1 : reached ? 0.82 : 0.3;

          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: leftOf(i),
                top: cardY,
                width: cardW,
                height: CARD_H,
                borderRadius: 22,
                backgroundColor: theme.surface,
                border: `${isCurrent || hot > 0.3 ? 2 : 1}px ${reached ? 'solid' : 'dashed'} ${
                  isCurrent || hot > 0.3 ? theme.accent : withAlpha(theme.surfaceBorder, 0.95)
                }`,
                boxShadow:
                  isCurrent || hot > 0.15
                    ? `0 18px 46px ${withAlpha(theme.ink, 0.55)}, 0 0 0 6px ${withAlpha(theme.accent, 0.1 + hot * 0.08)}`
                    : `0 8px 22px ${withAlpha(theme.ink, 0.28)}`,
                opacity: litness,
                transform: `translateY(${-raise}px)`,
                padding: '22px 22px 18px',
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden',
              }}
            >
              {/* Chip + app name. */}
              <div style={{display: 'flex', alignItems: 'center', gap: 13}}>
                <div
                  style={{
                    width: 54,
                    height: 54,
                    borderRadius: 16,
                    flexShrink: 0,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: `linear-gradient(145deg, ${withAlpha(tint, 0.28)}, ${withAlpha(tint, 0.1)})`,
                    border: `1px solid ${withAlpha(tint, 0.42)}`,
                  }}
                >
                  <Icon size={27} color={tint} strokeWidth={2.1} />
                </div>
                <div style={{minWidth: 0}}>
                  {node.app ? (
                    <div
                      style={{
                        fontFamily: theme.fontMono,
                        fontSize: 15,
                        letterSpacing: 1.6,
                        textTransform: 'uppercase',
                        color: theme.textMuted,
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                      }}
                    >
                      {node.app}
                    </div>
                  ) : null}
                  <div
                    style={{
                      marginTop: node.app ? 5 : 0,
                      fontFamily: theme.fontBody,
                      fontSize: 13,
                      fontWeight: 700,
                      letterSpacing: 1.4,
                      textTransform: 'uppercase',
                      color: tint,
                    }}
                  >
                    {node.kind}
                  </div>
                </div>
              </div>

              <div
                style={{
                  marginTop: 18,
                  fontFamily: theme.fontDisplay,
                  fontSize: n >= 5 ? 25 : 28,
                  fontWeight: 650,
                  lineHeight: 1.22,
                  letterSpacing: -0.3,
                  color: theme.text,
                }}
              >
                {node.title}
              </div>

              {/* The tick the token leaves behind. Reserved space either way, so
                  a card does not resize the moment it is reached. */}
              <div
                style={{
                  marginTop: 'auto',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  height: 22,
                  opacity: done ? 1 : 0,
                }}
              >
                <div
                  style={{
                    width: 20,
                    height: 20,
                    borderRadius: 10,
                    backgroundColor: theme.accent,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  <svg width={12} height={12} viewBox="0 0 12 12">
                    <path
                      d="M2.5 6.3 L4.8 8.6 L9.5 3.6"
                      fill="none"
                      stroke={theme.ink}
                      strokeWidth={2.2}
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                </div>
                <span
                  style={{
                    fontFamily: theme.fontMono,
                    fontSize: 15,
                    letterSpacing: 0.6,
                    color: theme.accentText,
                  }}
                >
                  done
                </span>
              </div>
            </div>
          );
        })}

        {/* Ports, drawn after the cards so they sit on the card edge and the
            wire appears to plug into them. */}
        {nodes.map((_, i) => {
          const reached = running || i <= current;
          return (
            <div key={`p${i}`}>
              {i > 0 ? <Port x={leftOf(i)} y={wireY} on={reached} theme={theme} /> : null}
              {i < n - 1 ? <Port x={leftOf(i) + cardW} y={wireY} on={reached} theme={theme} /> : null}
            </div>
          );
        })}

        {/* The payload, on the closing beat. The tail is what makes it read as
            "the record is at THIS card" rather than as a caption drifting over
            the row — without it the pill and the cards are two things moving
            past each other. */}
        {running && payload ? (
          <div
            style={{
              position: 'absolute',
              left: tokenX,
              top: cardY - TOKEN_LIFT - TOKEN_H,
              transform: `translateX(-50%) scale(${interpolate(runP, [0, 0.1], [0.7, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              })})`,
              opacity: interpolate(runP, [0, 0.08], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
            }}
          >
            <div
              style={{
                height: TOKEN_H,
                display: 'flex',
                alignItems: 'center',
                padding: '0 21px',
                borderRadius: 999,
                whiteSpace: 'nowrap',
                backgroundColor: theme.accent,
                color: theme.ink,
                fontFamily: theme.fontMono,
                fontSize: 21,
                fontWeight: 700,
                letterSpacing: 0.3,
                boxShadow: `0 12px 34px ${withAlpha(theme.accent, 0.34)}`,
              }}
            >
              {payload}
            </div>
            <div style={{width: 2, height: TOKEN_LIFT - 9, backgroundColor: withAlpha(theme.accent, 0.75)}} />
            <div
              style={{
                width: 9,
                height: 9,
                borderRadius: 5,
                backgroundColor: theme.accent,
              }}
            />
          </div>
        ) : null}
      </div>

      {/* One line under the canvas: the current card's note while building, and
          what is travelling while it runs. Only one is ever up, so the frame
          never holds two sentences to read. */}
      <div
        style={{
          marginTop: 26,
          width: COL_W,
          textAlign: 'center',
          fontFamily: theme.fontBody,
          fontSize: 29,
          lineHeight: 1.35,
          color: theme.textMuted,
          opacity: interpolate(sinceStep, [0, 12], [0, 1], {
            extrapolateLeft: 'clamp',
            extrapolateRight: 'clamp',
          }),
        }}
      >
        {running ? (payload ? `Running it with one ${payload.toLowerCase()}.` : '') : (active?.note ?? '')}
      </div>
    </Stage>
  );
};

/** The plug a wire lands in. Two per card, at the height the wire runs. */
const Port: React.FC<{x: number; y: number; on: boolean; theme: ResolvedTheme}> = ({x, y, on, theme}) => (
  <div
    style={{
      position: 'absolute',
      left: x - PORT_R,
      top: y - PORT_R,
      width: PORT_R * 2,
      height: PORT_R * 2,
      borderRadius: PORT_R,
      backgroundColor: on ? theme.accent : theme.bgBottom,
      border: `2px solid ${on ? theme.accent : theme.line}`,
      opacity: on ? 1 : 0.55,
    }}
  />
);
