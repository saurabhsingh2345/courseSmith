import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_W} from './Stage';

// MockupScene is a page assembling itself inside a device frame.
//
// Every block is drawn from the design system rather than screenshotted, which
// is what lets a lesson about a builder outlive that builder's next redesign —
// and what keeps a page legible at 1080p, where a real capture of a 1440px
// editor is a field of unreadable 9px labels.
//
// The blocks that have not landed yet are NOT drawn. This is the opposite of the
// choice TimelineScene makes about its future stops, and deliberately: a
// timeline's subject is a run you should see the end of, while a page's subject
// is the page, and drawing the un-built sections as ghosts would mean the
// viewer never once sees the screen as it will actually look. The layer list
// beside the frame carries what is still to come, which is exactly the division
// of labour a real builder makes.
//
// Nothing here is positioned by the model. It declares a stack of kinds and the
// renderer owns every pixel, which is the same bargain the story template's
// staging makes and the reason a mockup cannot come out overlapping itself.

const ROW_W = 1580;
const ROW_H = 640;
const LAYERS_W = 400;
const GUTTER = 30;
const DEVICE_W = ROW_W - LAYERS_W - GUTTER;
const CHROME_H = 50;
const PAGE_PAD = 22;
const BLOCK_GAP = 14;
const PHONE_W = 348;

/** Nominal block heights, before the page is fitted to the viewport. */
const BLOCK_H: Record<string, number> = {
  header: 58,
  hero: 170,
  text: 96,
  image: 150,
  grid: 150,
  button: 70,
  input: 78,
  list: 128,
  footer: 52,
};

type Block = {kind: string; label: string; text?: string; note?: string};
type Step = {startMs: number; endMs: number; at?: number; whole?: boolean};

export const MockupScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const title = String(props.title ?? '');
  const screen = String(props.screen ?? '');
  const phone = props.device === 'phone';
  const blocks = (Array.isArray(props.blocks) ? props.blocks : []) as Block[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];

  if (blocks.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const whole = step.whole === true;
  const current = whole ? blocks.length - 1 : (step.at ?? 0);
  const built = current + 1;
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const frameW = phone ? PHONE_W : DEVICE_W;
  const pageW = frameW - PAGE_PAD * 2;
  // Fit the page to the viewport. A five-block landing page is taller than any
  // device frame that also fits on a slide, so the stack is scaled rather than
  // clipped — a page whose footer is off-screen is not the page being described.
  const avail = ROW_H - CHROME_H - PAGE_PAD * 2;
  const nominal =
    blocks.reduce((sum, b) => sum + (BLOCK_H[b.kind] ?? BLOCK_H.text), 0) +
    BLOCK_GAP * (blocks.length - 1);
  const fit = Math.min(1, avail / nominal);

  const wire = {
    fill: withAlpha(theme.mass, 0.12),
    bar: withAlpha(theme.mass, 0.3),
    strong: withAlpha(theme.mass, 0.58),
  };

  const active = blocks[current];

  return (
    <Stage justify="center">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={18} />

      <div style={{display: 'flex', gap: GUTTER, width: ROW_W, height: ROW_H}}>
        {/* The device. In phone mode the frame narrows but the column does not,
            so the composition holds its shape across both devices instead of
            collapsing toward the middle. */}
        <div style={{width: DEVICE_W, display: 'flex', justifyContent: 'center'}}>
          <div
            style={{
              width: frameW,
              height: ROW_H,
              borderRadius: phone ? 44 : 22,
              backgroundColor: theme.surface,
              border: `1px solid ${theme.surfaceBorder}`,
              boxShadow: `0 26px 70px ${withAlpha(theme.ink, 0.55)}`,
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column',
            }}
          >
            {phone ? (
              <div
                style={{
                  height: CHROME_H,
                  flexShrink: 0,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <div
                  style={{
                    width: 132,
                    height: 26,
                    borderRadius: 999,
                    backgroundColor: theme.bgBottom,
                  }}
                />
              </div>
            ) : (
              <div
                style={{
                  height: CHROME_H,
                  flexShrink: 0,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 18,
                  padding: '0 20px',
                  borderBottom: `1px solid ${theme.surfaceBorder}`,
                  backgroundColor: withAlpha(theme.bgBottom, 0.55),
                }}
              >
                <div style={{display: 'flex', gap: 9}}>
                  {[0, 1, 2].map((d) => (
                    <span
                      key={d}
                      style={{
                        width: 11,
                        height: 11,
                        borderRadius: 6,
                        backgroundColor: theme.line,
                        opacity: 0.45,
                      }}
                    />
                  ))}
                </div>
                <div
                  style={{
                    flex: 1,
                    height: 28,
                    borderRadius: 999,
                    backgroundColor: withAlpha(theme.mass, 0.07),
                    display: 'flex',
                    alignItems: 'center',
                    padding: '0 15px',
                    fontFamily: theme.fontMono,
                    fontSize: 15,
                    color: theme.textMuted,
                  }}
                >
                  {screen ? `${screen.toLowerCase().replace(/\s+/g, '-')}` : ''}
                </div>
              </div>
            )}

            {/* The page. */}
            <div
              style={{
                flex: 1,
                padding: PAGE_PAD,
                backgroundColor: theme.bgTop,
                display: 'flex',
                flexDirection: 'column',
                gap: BLOCK_GAP * fit,
                overflow: 'hidden',
              }}
            >
              {blocks.slice(0, built).map((block, i) => {
                const isCurrent = !whole && i === current;
                // Each block lands once, on its own beat; after that it is part
                // of the page and holds still.
                const land =
                  i === current
                    ? spring({
                        frame: sinceStep,
                        fps,
                        config: {damping: 200, mass: 0.6},
                        durationInFrames: 18,
                      })
                    : 1;
                return (
                  <div
                    key={i}
                    style={{
                      height: (BLOCK_H[block.kind] ?? BLOCK_H.text) * fit,
                      flexShrink: 0,
                      opacity: land,
                      transform: `translateY(${(1 - land) * 16}px) scale(${0.97 + land * 0.03})`,
                      borderRadius: 12,
                      outline: isCurrent ? `2px solid ${theme.accent}` : 'none',
                      outlineOffset: 3,
                      boxShadow: isCurrent ? `0 0 0 8px ${withAlpha(theme.accent, 0.12)}` : 'none',
                    }}
                  >
                    <BlockArt
                      block={block}
                      theme={theme}
                      wire={wire}
                      lit={isCurrent}
                      w={pageW}
                      h={(BLOCK_H[block.kind] ?? BLOCK_H.text) * fit}
                    />
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* The layer list. It is what every builder in this category shares, and
            it carries what has not been built yet so the page itself never has
            to draw a ghost of itself. */}
        <div
          style={{
            width: LAYERS_W,
            // The sidebar hugs its rows rather than matching the frame. Five
            // layers stretched down a 640px panel is a list with a void under
            // it, and an inspector that floats beside the canvas is what these
            // builders actually look like anyway.
            alignSelf: 'flex-start',
            borderRadius: 22,
            backgroundColor: withAlpha(theme.surface, 0.55),
            border: `1px solid ${theme.surfaceBorder}`,
            padding: 20,
            display: 'flex',
            flexDirection: 'column',
            gap: 10,
          }}
        >
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 15,
              letterSpacing: 2.4,
              textTransform: 'uppercase',
              color: theme.textMuted,
              padding: '2px 6px 12px',
              borderBottom: `1px solid ${theme.surfaceBorder}`,
            }}
          >
            {`Layers · ${blocks.length}`}
          </div>
          {blocks.map((block, i) => {
            const done = i < built;
            const isCurrent = !whole && i === current;
            return (
              <div
                key={i}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 13,
                  padding: '13px 14px',
                  borderRadius: 13,
                  backgroundColor: isCurrent ? withAlpha(theme.accent, 0.13) : 'transparent',
                  borderLeft: `4px solid ${isCurrent ? theme.accent : 'transparent'}`,
                  opacity: done ? 1 : 0.34,
                }}
              >
                <span
                  style={{
                    width: 28,
                    height: 28,
                    flexShrink: 0,
                    borderRadius: 8,
                    backgroundColor: done ? withAlpha(theme.accent, 0.2) : withAlpha(theme.line, 0.18),
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontFamily: theme.fontMono,
                    fontSize: 15,
                    color: done ? theme.accentText : theme.textMuted,
                  }}
                >
                  {i + 1}
                </span>
                <span
                  style={{
                    flex: 1,
                    fontFamily: theme.fontBody,
                    fontSize: 23,
                    fontWeight: isCurrent ? 600 : 500,
                    color: theme.text,
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                  }}
                >
                  {block.label}
                </span>
                <span
                  style={{
                    fontFamily: theme.fontMono,
                    fontSize: 13,
                    letterSpacing: 1.2,
                    textTransform: 'uppercase',
                    color: theme.textMuted,
                    flexShrink: 0,
                  }}
                >
                  {block.kind}
                </span>
              </div>
            );
          })}
        </div>
      </div>

      <div
        style={{
          marginTop: 22,
          width: ROW_W,
          textAlign: 'center',
          fontFamily: theme.fontBody,
          fontSize: 28,
          lineHeight: 1.35,
          color: theme.textMuted,
          opacity: interpolate(sinceStep, [0, 12], [0, 1], {
            extrapolateLeft: 'clamp',
            extrapolateRight: 'clamp',
          }),
        }}
      >
        {whole ? '' : (active?.note ?? '')}
      </div>
    </Stage>
  );
};

type Wire = {fill: string; bar: string; strong: string};

/** A filler rule — what a wireframe uses where real copy would go. */
const Rule: React.FC<{w: number | string; h: number; c: string; r?: number}> = ({w, h, c, r}) => (
  <div style={{width: w, height: h, borderRadius: r ?? h / 2, backgroundColor: c, flexShrink: 0}} />
);

/**
 * One block's drawing.
 *
 * Every kind is a different picture rather than a labelled rectangle, which is
 * the whole reason this template can stand in for a screen recording: a viewer
 * has to recognise a nav bar as a nav bar without reading the layer list.
 */
const BlockArt: React.FC<{
  block: Block;
  theme: ResolvedTheme;
  wire: Wire;
  lit: boolean;
  w: number;
  h: number;
}> = ({block, theme, wire, lit, w, h}) => {
  const box: React.CSSProperties = {
    width: '100%',
    height: '100%',
    borderRadius: 12,
    display: 'flex',
    overflow: 'hidden',
  };
  const tint = lit ? withAlpha(theme.accent, 0.1) : wire.fill;

  switch (block.kind) {
    case 'header':
      return (
        <div style={{...box, alignItems: 'center', gap: 14, padding: '0 16px', backgroundColor: tint}}>
          <div
            style={{
              width: h * 0.44,
              height: h * 0.44,
              borderRadius: 8,
              backgroundColor: theme.accent,
              flexShrink: 0,
            }}
          />
          {block.text ? (
            <span
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: h * 0.34,
                fontWeight: 700,
                color: theme.text,
              }}
            >
              {block.text}
            </span>
          ) : (
            <Rule w={110} h={h * 0.16} c={wire.strong} />
          )}
          <div style={{flex: 1}} />
          <Rule w={72} h={9} c={wire.bar} />
          <Rule w={58} h={9} c={wire.bar} />
          <Rule w={66} h={9} c={wire.bar} />
        </div>
      );

    case 'hero':
      return (
        <div
          style={{
            ...box,
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: h * 0.09,
            backgroundColor: tint,
          }}
        >
          {block.text ? (
            <span
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: h * 0.24,
                fontWeight: 700,
                letterSpacing: -0.5,
                color: theme.text,
                textAlign: 'center',
              }}
            >
              {block.text}
            </span>
          ) : (
            <Rule w={w * 0.5} h={h * 0.16} c={wire.strong} />
          )}
          <Rule w={w * 0.34} h={h * 0.07} c={wire.bar} />
          {/* The hero's call to action is a placeholder, not the accent. A page
              that also declares a `button` block would otherwise put two yellow
              pills in the frame competing to be the thing you click — the
              accent belongs to the block the clip is actually about. */}
          <div
            style={{
              marginTop: h * 0.05,
              width: w * 0.16,
              height: h * 0.17,
              borderRadius: 999,
              backgroundColor: wire.strong,
            }}
          />
        </div>
      );

    case 'text':
      return (
        <div
          style={{
            ...box,
            flexDirection: 'column',
            justifyContent: 'center',
            gap: h * 0.13,
            padding: '0 18px',
          }}
        >
          <Rule w="100%" h={h * 0.11} c={wire.bar} />
          <Rule w="92%" h={h * 0.11} c={wire.bar} />
          <Rule w="64%" h={h * 0.11} c={wire.bar} />
        </div>
      );

    case 'image':
      return (
        <div style={{...box, alignItems: 'flex-end', justifyContent: 'center', backgroundColor: tint}}>
          <svg width={w * 0.34} height={h * 0.62} viewBox="0 0 120 70" style={{marginBottom: h * 0.16}}>
            <circle cx={88} cy={18} r={9} fill={wire.bar} />
            <path d="M4 66 L40 22 L66 52 L84 34 L116 66 Z" fill={wire.bar} />
          </svg>
        </div>
      );

    case 'grid':
      return (
        <div style={{...box, gap: 16}}>
          {[0, 1, 2].map((c) => (
            <div
              key={c}
              style={{
                flex: 1,
                borderRadius: 11,
                backgroundColor: tint,
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden',
              }}
            >
              <div style={{height: '52%', backgroundColor: withAlpha(theme.mass, 0.19)}} />
              <div
                style={{
                  flex: 1,
                  padding: '12px 14px',
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'center',
                  gap: 9,
                }}
              >
                <Rule w="72%" h={9} c={wire.strong} />
                <Rule w="94%" h={7} c={wire.bar} />
              </div>
            </div>
          ))}
        </div>
      );

    case 'button':
      return (
        <div style={{...box, alignItems: 'center', justifyContent: 'center'}}>
          <div
            style={{
              minWidth: w * 0.2,
              height: h * 0.62,
              padding: '0 30px',
              borderRadius: 999,
              backgroundColor: theme.accent,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontFamily: theme.fontBody,
              fontSize: h * 0.28,
              fontWeight: 700,
              color: theme.ink,
            }}
          >
            {block.text ?? ''}
          </div>
        </div>
      );

    case 'input':
      return (
        <div style={{...box, alignItems: 'center'}}>
          <div
            style={{
              width: '100%',
              height: h * 0.66,
              borderRadius: 11,
              border: `2px solid ${lit ? theme.accent : withAlpha(theme.mass, 0.24)}`,
              backgroundColor: withAlpha(theme.mass, 0.05),
              display: 'flex',
              alignItems: 'center',
              padding: '0 18px',
              gap: 12,
            }}
          >
            {block.text ? (
              <span style={{fontFamily: theme.fontBody, fontSize: h * 0.27, color: theme.textMuted}}>
                {block.text}
              </span>
            ) : (
              <Rule w={180} h={9} c={wire.bar} />
            )}
            {lit ? <Rule w={2} h={h * 0.34} c={theme.accent} r={1} /> : null}
          </div>
        </div>
      );

    case 'list':
      return (
        <div style={{...box, flexDirection: 'column', justifyContent: 'center'}}>
          {[0, 1, 2].map((r) => (
            <div
              key={r}
              style={{
                flex: 1,
                display: 'flex',
                alignItems: 'center',
                gap: 15,
                padding: '0 16px',
                borderTop: r ? `1px solid ${withAlpha(theme.mass, 0.09)}` : 'none',
              }}
            >
              <div
                style={{
                  width: h * 0.17,
                  height: h * 0.17,
                  borderRadius: 7,
                  backgroundColor: wire.fill,
                  flexShrink: 0,
                }}
              />
              <Rule w="46%" h={9} c={wire.bar} />
              <div style={{flex: 1}} />
              <Rule w={54} h={9} c={wire.fill} />
            </div>
          ))}
        </div>
      );

    case 'footer':
      return (
        <div
          style={{
            ...box,
            alignItems: 'center',
            justifyContent: 'center',
            gap: 26,
            borderTop: `1px solid ${withAlpha(theme.mass, 0.12)}`,
            borderRadius: 0,
          }}
        >
          <Rule w={64} h={8} c={wire.bar} />
          <Rule w={52} h={8} c={wire.bar} />
          <Rule w={72} h={8} c={wire.bar} />
        </div>
      );

    default:
      return <div style={{...box, backgroundColor: tint}} />;
  }
};
