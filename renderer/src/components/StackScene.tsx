import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_W} from './Stage';
import {iconFor} from './icons';

// StackScene is a set of tiers stacked down the frame.
//
// Every band is on screen from the first frame, because the claim a stack makes
// is about *arrangement* and an arrangement you cannot see the shape of is a
// list. What moves is which band is lit; the rest recede rather than vanish.
//
// Tools sit side by side inside their band, which is the part of the layout
// doing the most teaching for the least effort: two cards in one tier read as
// alternatives without a word of narration, and the same two in a flow diagram
// would read as two things a request passes through in turn.
//
// The connector between bands is drawn from the middle of the stack rather than
// down one edge. A stack's handoff is not a route — nothing is choosing a path —
// so a spine down the left would borrow the timeline's grammar for a claim this
// template is not making.

const COL_W = Math.min(STAGE_W, 1660);
const BODY_H = 660;
const RAIL_W = 300;
const LINK_H = 26;

type Tool = {name: string; icon?: string; note?: string};
type Layer = {name: string; role: string; tools: Tool[]};
type Step = {startMs: number; endMs: number; at?: number; whole?: boolean};

export const StackScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const title = String(props.title ?? '');
  const layers = (Array.isArray(props.layers) ? props.layers : []) as Layer[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];

  if (layers.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const whole = step.whole === true;
  const current = whole ? layers.length - 1 : (step.at ?? 0);
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const n = layers.length;
  const bandH = (BODY_H - LINK_H * (n - 1)) / n;
  const lift = spring({frame: sinceStep, fps, config: {damping: 200, mass: 0.7}, durationInFrames: 18});

  const active = layers[current];

  return (
    <Stage justify="center">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={16} />

      <div style={{width: COL_W}}>
        {layers.map((layer, i) => {
          const isCurrent = !whole && i === current;
          const reached = whole || i <= current;
          return (
            <div key={i}>
              <div
                style={{
                  height: bandH,
                  display: 'flex',
                  borderRadius: 22,
                  overflow: 'hidden',
                  backgroundColor: isCurrent ? withAlpha(theme.accent, 0.07) : withAlpha(theme.surface, 0.5),
                  border: `${isCurrent ? 2 : 1}px solid ${
                    isCurrent ? theme.accent : theme.surfaceBorder
                  }`,
                  boxShadow: isCurrent
                    ? `0 20px 50px ${withAlpha(theme.ink, 0.5)}, 0 0 0 7px ${withAlpha(theme.accent, 0.1)}`
                    : 'none',
                  opacity: reached ? 1 : 0.4,
                  transform: `translateX(${isCurrent ? lift * 10 : 0}px)`,
                }}
              >
                {/* The rail: which tier this is. */}
                <div
                  style={{
                    width: RAIL_W,
                    flexShrink: 0,
                    padding: '0 26px',
                    display: 'flex',
                    flexDirection: 'column',
                    justifyContent: 'center',
                    gap: 9,
                    borderRight: `1px solid ${isCurrent ? withAlpha(theme.accent, 0.35) : theme.surfaceBorder}`,
                    backgroundColor: withAlpha(theme.bgBottom, 0.4),
                  }}
                >
                  <span
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: 14,
                      letterSpacing: 2.2,
                      textTransform: 'uppercase',
                      color: isCurrent ? theme.accentText : theme.textMuted,
                    }}
                  >
                    {`Layer ${i + 1}`}
                  </span>
                  <span
                    style={{
                      fontFamily: theme.fontDisplay,
                      fontSize: n >= 4 ? 33 : 38,
                      fontWeight: 700,
                      letterSpacing: -0.5,
                      lineHeight: 1.1,
                      color: theme.text,
                    }}
                  >
                    {layer.name}
                  </span>
                </div>

                {/* The tools. Side by side means alternatives, which is the
                    comparison a viewer wants and the layout gives for free. */}
                <div style={{flex: 1, display: 'flex', gap: 16, padding: 16}}>
                  {layer.tools.map((tool, j) => {
                    const Icon = iconFor(tool.icon);
                    return (
                      <div
                        key={j}
                        style={{
                          flex: 1,
                          display: 'flex',
                          alignItems: 'center',
                          gap: 16,
                          padding: '0 20px',
                          borderRadius: 15,
                          backgroundColor: theme.surface,
                          // The cards inside the lit band pick up the accent
                          // too. Without it the band's border was the only lit
                          // thing and the tools read as belonging to the frame
                          // rather than to the tier.
                          border: `1px solid ${
                            isCurrent ? withAlpha(theme.accent, 0.42) : theme.surfaceBorder
                          }`,
                          minWidth: 0,
                        }}
                      >
                        <div
                          style={{
                            width: 52,
                            height: 52,
                            flexShrink: 0,
                            borderRadius: 15,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            background: `linear-gradient(145deg, ${withAlpha(theme.accent, 0.26)}, ${withAlpha(
                              theme.primary,
                              0.16,
                            )})`,
                            border: `1px solid ${withAlpha(theme.accent, 0.34)}`,
                          }}
                        >
                          <Icon size={26} color={theme.accent} strokeWidth={2.1} />
                        </div>
                        <div style={{minWidth: 0}}>
                          <div
                            style={{
                              fontFamily: theme.fontDisplay,
                              fontSize: 27,
                              fontWeight: 650,
                              color: theme.text,
                              whiteSpace: 'nowrap',
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                            }}
                          >
                            {tool.name}
                          </div>
                          {tool.note ? (
                            <div
                              style={{
                                marginTop: 4,
                                fontFamily: theme.fontBody,
                                fontSize: 19,
                                lineHeight: 1.25,
                                color: theme.textMuted,
                              }}
                            >
                              {tool.note}
                            </div>
                          ) : null}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* The handoff to the tier below. */}
              {i < n - 1 ? (
                <div
                  style={{
                    height: LINK_H,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  <svg width={26} height={LINK_H} viewBox="0 0 26 26">
                    <path
                      d="M13 2 L13 16 M6 12 L13 20 L20 12"
                      fill="none"
                      stroke={i + 1 <= current || whole ? theme.accent : theme.line}
                      strokeWidth={3}
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      opacity={i + 1 <= current || whole ? 1 : 0.45}
                    />
                  </svg>
                </div>
              ) : null}
            </div>
          );
        })}
      </div>

      <div
        style={{
          marginTop: 24,
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
        {whole ? '' : (active?.role ?? '')}
      </div>
    </Stage>
  );
};
