import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage} from './Stage';
import {iconFor} from './icons';

// ConstellationScene is one idea in the middle with its properties around it.
//
// The centre is fixed and the spokes sit at angles Go computed, so four
// properties land on the compass points and five make a pentagon — the same map
// every render, rather than whatever a layout pass settles on.
//
// Two decisions make it read as a recap rather than as a diagram.
//
// The relation word rides on the connector, not in the node. "Redis —is→
// in-memory" is a sentence the eye completes on its own; the same two nodes
// joined by a bare line say only that they are related somehow, which is what
// separates a map of properties from a table of contents.
//
// The unlit spokes are drawn from the start, faint. The claim a closing map
// makes is that the idea has a *shape*, and a shape revealed one piece at a
// time is a list. By the final beat everything is bright and nothing has moved,
// which is what makes that frame the one people screenshot.

// Wide enough that a relation word sitting just outside the centre never
// reaches the node on a horizontal spoke, which is the tightest case.
const RADIUS = 340;
const CENTRE = 132;
const NODE_W = 250;

type Spoke = {rel: string; label: string; note?: string; icon?: string; angle: number};
type Step = {startMs: number; endMs: number; show: 'centre' | 'spoke' | 'whole'; at?: number};

const NodeIcon: React.FC<{name?: string; color: string; size: number}> = ({name, color, size}) => {
  const Icon = iconFor(name);
  return (
    <div style={{color, lineHeight: 0}}>
      <Icon size={size} strokeWidth={1.9} />
    </div>
  );
};

export const ConstellationScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const centre = String(props.centre ?? '');
  const spokes = (Array.isArray(props.spokes) ? props.spokes : []) as Spoke[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (spokes.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;
  const whole = step.show === 'whole';
  const current = step.show === 'spoke' ? (step.at ?? 0) : -1;

  // Which spokes have been lit by now, so the map accumulates.
  const litBefore = new Set<number>();
  steps.slice(0, idx + 1).forEach((s) => {
    if (s.show === 'spoke' && s.at !== undefined) litBefore.add(s.at);
  });

  const enter = spring({
    frame: ((nowMs - steps[0].startMs) / 1000) * FPS,
    fps,
    config: {damping: 200, mass: 0.7},
    durationInFrames: 20,
  });

  const accent = (i: number): string =>
    [theme.accentQuantity, theme.accentRival, theme.accentLimit][i % 3];

  return (
    <Stage justify="center">
      <div
        style={{
          position: 'relative',
          width: (RADIUS + NODE_W / 2) * 2,
          height: RADIUS * 2 + 150,
          opacity: enter,
        }}
      >
        {/* The connectors, drawn under the nodes. */}
        <svg
          width="100%"
          height="100%"
          style={{position: 'absolute', inset: 0, overflow: 'visible'}}
        >
          {spokes.map((s, i) => {
            const rad = (s.angle * Math.PI) / 180;
            const cx = (RADIUS + NODE_W / 2);
            const cy = RADIUS + 75;
            // The connector starts at the centre node's EDGE, not its middle:
            // a line drawn from the true centre runs straight through the name
            // sitting there, which reads as a strike-through rather than a link.
            const x0 = cx + Math.cos(rad) * (CENTRE + 6);
            const y0 = cy + Math.sin(rad) * (CENTRE + 6);
            const x = cx + Math.cos(rad) * RADIUS;
            const y = cy + Math.sin(rad) * RADIUS;
            const on = whole || litBefore.has(i);
            const isCurrent = i === current;
            const draw = isCurrent
              ? interpolate(sinceStep, [2, 18], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                })
              : on
                ? 1
                : 0;
            const c = accent(i);
            return (
              <g key={i}>
                <line
                  x1={x0}
                  y1={y0}
                  x2={x}
                  y2={y}
                  stroke={withAlpha(theme.text, 0.1)}
                  strokeWidth={2}
                />
                <line
                  x1={x0}
                  y1={y0}
                  x2={x0 + (x - x0) * draw}
                  y2={y0 + (y - y0) * draw}
                  stroke={c}
                  strokeWidth={isCurrent || whole ? 3 : 2}
                  opacity={on ? (isCurrent || whole ? 1 : 0.5) : 0}
                />
              </g>
            );
          })}
        </svg>

        {/* The relation words, riding on their connectors. This is what makes a
            spoke a property rather than a neighbouring topic. */}
        {spokes.map((s, i) => {
          const rad = (s.angle * Math.PI) / 180;
          const cx = RADIUS + NODE_W / 2;
          const cy = RADIUS + 75;
          const on = whole || litBefore.has(i);
          const isCurrent = i === current;
          return (
            <div
              key={`rel-${i}`}
              style={{
                position: 'absolute',
                // Hugged to the centre rather than centred on the span: at the
                // midpoint a word like SURVIVES runs under the horizontal
                // nodes, which sit only NODE_W/2 in from the spoke end.
                left: cx + Math.cos(rad) * (CENTRE + 40),
                top: cy + Math.sin(rad) * (CENTRE + 40),
                transform: 'translate(-50%, -50%)',
                fontFamily: theme.fontMono,
                fontSize: 15,
                letterSpacing: 1.8,
                textTransform: 'uppercase',
                color: theme.textMuted,
                background: theme.bgBottom,
                padding: '2px 8px',
                borderRadius: 4,
                whiteSpace: 'nowrap',
                opacity: on ? (isCurrent || whole ? 1 : 0.45) : 0,
              }}
            >
              {s.rel}
            </div>
          );
        })}

        {/* The spoke nodes. */}
        {spokes.map((s, i) => {
          const rad = (s.angle * Math.PI) / 180;
          const cx = RADIUS + NODE_W / 2;
          const cy = RADIUS + 75;
          const on = whole || litBefore.has(i);
          const isCurrent = i === current;
          const c = accent(i);
          return (
            <div
              key={`node-${i}`}
              style={{
                position: 'absolute',
                left: cx + Math.cos(rad) * RADIUS,
                top: cy + Math.sin(rad) * RADIUS,
                transform: 'translate(-50%, -50%)',
                width: NODE_W,
                padding: '16px 18px',
                borderRadius: 12,
                background: on ? withAlpha(c, isCurrent || whole ? 0.16 : 0.08) : withAlpha(theme.text, 0.03),
                border: `1px solid ${on ? withAlpha(c, isCurrent || whole ? 0.55 : 0.28) : withAlpha(theme.text, 0.08)}`,
                display: 'flex',
                alignItems: 'center',
                gap: 12,
                opacity: on ? 1 : 0.35,
              }}
            >
              <NodeIcon name={s.icon} color={on ? c : theme.textMuted} size={22} />
              <div
                style={{
                  fontFamily: theme.fontDisplay,
                  fontSize: 26,
                  fontWeight: 700,
                  lineHeight: 1.15,
                  color: on ? theme.text : theme.textMuted,
                }}
              >
                {s.label}
              </div>
            </div>
          );
        })}

        {/* The centre. */}
        <div
          style={{
            position: 'absolute',
            left: RADIUS + NODE_W / 2,
            top: RADIUS + 75,
            transform: 'translate(-50%, -50%)',
            width: CENTRE * 2,
            height: CENTRE * 2,
            borderRadius: '50%',
            background: withAlpha(theme.text, 0.06),
            border: `2px solid ${withAlpha(theme.text, 0.22)}`,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 10,
          }}
        >
          <NodeIcon name={String(props.centreIcon ?? '')} color={theme.text} size={32} />
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 40,
              fontWeight: 800,
              letterSpacing: -1,
              color: theme.text,
              textAlign: 'center',
              padding: '0 12px',
            }}
          >
            {centre}
          </div>
        </div>

        {/* The current spoke's line, under the map. */}
        {current >= 0 && spokes[current]?.note ? (
          <div
            style={{
              position: 'absolute',
              left: 0,
              right: 0,
              bottom: -18,
              textAlign: 'center',
              fontFamily: theme.fontBody,
              fontSize: 27,
              color: theme.textMuted,
              opacity: interpolate(sinceStep, [6, 20], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            {spokes[current].note}
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
