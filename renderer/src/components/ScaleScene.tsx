import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {iconFor} from './icons';

// ScaleScene is worlds inside worlds, with the camera pulling back.
//
// Every other way of drawing a thousand-fold gap fails the same way: one mark at
// full size and the rest at nothing. Nesting does not, because it never asks two
// quantities to share a scale — it asks the viewer to watch one become a speck
// inside the next, which is a thing the eye is extremely good at reading and
// needs no axis at all.
//
// The camera is a single number. `cam` is a continuous position on the ladder,
// eased between rungs, and every frame's size is BASE * NEST^(i - cam). That is
// the whole engine. It means nothing re-flows when the camera moves — the boxes
// are always in the same relationship to each other, and only the viewport
// changes — so the pull-back is one continuous move rather than a sequence of
// layouts, which is why this template is one scene for the whole clip.
//
// NEST is a fixed visual ratio and NOT the real one, and that is deliberate. Drawn
// truly proportionally, a rung a thousand times the last would put the previous
// world at a third of a pixel: honest, and a picture of nothing. So the geometry
// is constant and the true multiplier is *written* — on the frame that is
// arriving and on the rail. The eye gets the containment, the reader gets the
// number, and neither is asked to do the other's job.
//
// The current world wears viewfinder brackets. They cost four short lines and
// they do the work a caption would otherwise have to: they say which of the five
// boxes on screen is the one being talked about, without adding a word.

const BOX_W = Math.min(STAGE_W, 1690);
const BOX_H = 812;
// The nest is a square well left of centre; the rail reads down the right.
const NEST = 3.4;
/**
 * The nesting ratio on the closing beat.
 *
 * Pulling the camera further back was the obvious way to get every rung in
 * frame, and it produced the worst frame in the clip: at 3.4x per level a
 * four-rung ladder puts the innermost world at 2.5% of the outermost, so
 * "all four at once" rendered as one box with a speck in it and two rungs that
 * were, truthfully, sub-pixel.
 *
 * So the closing frame COMPRESSES instead. The camera holds on the largest rung
 * and the spacing tightens, which fits every world on screen at a readable size.
 * That is not a lie about the data for the same reason the 3.4 is not: the
 * geometry was never the true ratio, the written multipliers always were, and
 * they are all on the rail beside it.
 */
const NEST_TIGHT = 1.9;
const BASE = 596;
const NEST_CX = 470;
const NEST_CY = BOX_H / 2;
const RAIL_X = 940;

type Level = {label: string; value: number; display: string; icon?: string; note?: string; times?: string};
type Step = {startMs: number; endMs: number; show: 'level' | 'whole'; at?: number};

export const ScaleScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const unit = String(props.unit ?? '');
  const span = typeof props.span === 'string' ? props.span : '';
  const levels = (Array.isArray(props.levels) ? props.levels : []) as Level[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const n = levels.length;
  if (n === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;
  const sinceStart = ((nowMs - steps[0].startMs) / 1000) * FPS;
  const whole = step.show === 'whole';

  // Where the camera rests on each beat. `whole` stays on the largest rung and
  // compresses the ladder instead of pulling further back — see NEST_TIGHT.
  const restOf = (s: Step): number => (s.show === 'whole' ? n - 1 : (s.at ?? 0));
  const from = idx === 0 ? restOf(step) : restOf(steps[idx - 1]);
  const to = restOf(step);
  const beatFrames = ((step.endMs - step.startMs) / 1000) * FPS;
  const move = interpolate(sinceStep, [2, Math.max(16, beatFrames * 0.55)], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const eased = move * move * (3 - 2 * move);
  const cam = from + (to - from) * eased;
  // Eased alongside the camera, so the ladder draws together as the last move
  // lands rather than snapping tight on the cut.
  const tightFrom = idx > 0 && steps[idx - 1].show === 'whole' ? NEST_TIGHT : NEST;
  const tightTo = whole ? NEST_TIGHT : NEST;
  const nest = tightFrom + (tightTo - tightFrom) * eased;

  const current = whole ? n - 1 : (step.at ?? 0);
  const level = levels[current];
  const enter = spring({frame: sinceStart, fps, config: {damping: 200, mass: 0.8}, durationInFrames: 24});

  return (
    <Stage justify="center">
      <div style={{width: BOX_W, height: BOX_H, position: 'relative', opacity: enter}}>
        {/* The nest. Drawn largest first so the small worlds paint on top of the
            big ones they sit inside. */}
        {levels
          .map((l, i) => ({l, i, side: BASE * Math.pow(nest, i - cam)}))
          .filter(({side}) => side > 6 && side < 5600)
          .sort((a, b) => b.side - a.side)
          .map(({l, i, side}) => {
            const f = side / BASE;
            const isCurrent = i === current;
            const past = i < current;
            const c = isCurrent
              ? theme.accentQuantity
              : past
                ? withAlpha(theme.accentQuantity, 0.6)
                : withAlpha(theme.text, 0.16);
            const Icon = iconFor(l.icon);
            const pad = Math.max(6, side * 0.045);
            return (
              <div
                key={i}
                style={{
                  position: 'absolute',
                  left: NEST_CX,
                  top: NEST_CY,
                  width: side,
                  height: side,
                  marginLeft: -side / 2,
                  marginTop: -side / 2,
                  boxSizing: 'border-box',
                  borderRadius: Math.max(6, side * 0.055),
                  // A floor of 1.4px, not 1px. The worlds already visited are
                  // the evidence the camera has moved, and at a third of the
                  // frame size a hairline scaled off `f` came out under a pixel
                  // and simply vanished — leaving the current world looking
                  // empty rather than looking like the fourth one.
                  border: `${Math.max(1.4, Math.min(3, 2.6 * Math.min(f, 1.2)))}px solid ${c}`,
                  background: isCurrent
                    ? withAlpha(theme.accentQuantity, 0.05)
                    : past
                      ? withAlpha(theme.accentQuantity, 0.03)
                      : 'transparent',
                  boxShadow: isCurrent ? `0 0 60px ${withAlpha(theme.accentQuantity, 0.16)}` : 'none',
                  overflow: 'hidden',
                }}
              >
                {/* The world's own chip, which shrinks away with it. This is the
                    part that makes the pull-back legible: the words you were
                    reading a second ago are now too small to read, and that IS
                    the multiplier. */}
                {f > 0.1 ? (
                  <div
                    style={{
                      position: 'absolute',
                      left: pad,
                      top: pad,
                      display: 'flex',
                      alignItems: 'center',
                      gap: 10 * Math.min(f, 1),
                      color: isCurrent ? theme.accentQuantity : theme.textMuted,
                    }}
                  >
                    <Icon size={Math.max(8, 30 * Math.min(f, 1.15))} strokeWidth={2} />
                    <div
                      style={{
                        fontFamily: theme.fontMono,
                        fontSize: Math.max(7, 20 * Math.min(f, 1.15)),
                        letterSpacing: 2 * Math.min(f, 1),
                        textTransform: 'uppercase',
                        whiteSpace: 'nowrap',
                      }}
                    >
                      {l.display}
                    </div>
                  </div>
                ) : null}
                {/* The subject's own object, drawn large and faint in the corner
                    the nested worlds never reach — they sit in the middle third,
                    so everything outside it is free. Without it the current
                    world is an empty box with a chip in one corner, which is
                    exactly what it looked like. */}
                {isCurrent ? (
                  <div
                    style={{
                      position: 'absolute',
                      left: side * 0.09,
                      bottom: side * 0.09,
                      color: withAlpha(theme.accentQuantity, 0.3),
                      lineHeight: 0,
                    }}
                  >
                    <Icon size={side * 0.2} strokeWidth={1.1} />
                  </div>
                ) : null}
                {/* Viewfinder brackets on whichever world is the subject. Four
                    short lines doing a caption's job. */}
                {isCurrent
                  ? ([
                      {top: pad, left: pad, bt: true, bl: true},
                      {top: pad, right: pad, bt: true, br: true},
                      {bottom: pad, left: pad, bb: true, bl: true},
                      {bottom: pad, right: pad, bb: true, br: true},
                    ] as const).map((corner, k) => (
                      <div
                        key={`br-${k}`}
                        style={{
                          position: 'absolute',
                          width: Math.max(10, side * 0.075),
                          height: Math.max(10, side * 0.075),
                          top: 'top' in corner ? corner.top : undefined,
                          bottom: 'bottom' in corner ? corner.bottom : undefined,
                          left: 'left' in corner ? corner.left : undefined,
                          right: 'right' in corner ? corner.right : undefined,
                          borderTop: 'bt' in corner ? `3px solid ${theme.accentQuantity}` : undefined,
                          borderBottom: 'bb' in corner ? `3px solid ${theme.accentQuantity}` : undefined,
                          borderLeft: 'bl' in corner ? `3px solid ${theme.accentQuantity}` : undefined,
                          borderRight: 'br' in corner ? `3px solid ${theme.accentQuantity}` : undefined,
                          opacity: 0.9,
                        }}
                      />
                    ))
                  : null}
              </div>
            );
          })}

        {/* The readout. Fixed size, because it is the one thing that must stay
            legible while everything else is busy changing size. */}
        <div
          style={{
            position: 'absolute',
            left: RAIL_X,
            top: 0,
            width: BOX_W - RAIL_X,
            height: BOX_H,
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center',
          }}
        >
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 17,
              letterSpacing: 4.2,
              textTransform: 'uppercase',
              color: theme.textMuted,
            }}
          >
            {whole ? `all ${n} · measured in ${unit}` : `step ${current + 1} of ${n} · ${unit}`}
          </div>

          <div
            style={{
              marginTop: 18,
              display: 'flex',
              alignItems: 'baseline',
              gap: 18,
              opacity: interpolate(sinceStep, [2, 16], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 92,
                fontWeight: 800,
                letterSpacing: -3,
                lineHeight: 1,
                color: theme.text,
                fontVariantNumeric: 'tabular-nums',
              }}
            >
              {level?.display}
            </div>
            {/* The true multiplier, which the geometry deliberately does not try
                to draw — one step of it while climbing, and the whole span on
                the closing frame, which is the figure the clip is actually
                about and the one no single rung carries. */}
            {whole ? (
              span ? (
                <div
                  style={{
                    fontFamily: theme.fontMono,
                    fontSize: 26,
                    fontWeight: 700,
                    color: theme.accentText,
                    border: `1px solid ${withAlpha(theme.accentQuantity, 0.45)}`,
                    background: withAlpha(theme.accentQuantity, 0.1),
                    borderRadius: 999,
                    padding: '6px 16px',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {`×${span} end to end`}
                </div>
              ) : null
            ) : level?.times ? (
              <div
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 26,
                  fontWeight: 700,
                  color: theme.accentText,
                  border: `1px solid ${withAlpha(theme.accentQuantity, 0.45)}`,
                  background: withAlpha(theme.accentQuantity, 0.1),
                  borderRadius: 999,
                  padding: '6px 16px',
                  whiteSpace: 'nowrap',
                  transform: `scale(${
                    0.7 +
                    0.3 *
                      spring({
                        frame: sinceStep - 8,
                        fps,
                        config: {damping: 13, mass: 0.5},
                        durationInFrames: 20,
                      })
                  })`,
                }}
              >
                {`×${level.times}`}
              </div>
            ) : null}
          </div>

          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 40,
              fontWeight: 700,
              letterSpacing: -0.8,
              color: theme.text,
              marginTop: 10,
            }}
          >
            {level?.label}
          </div>
          {/* Held open whether or not there is a line in it. The column is
              centred, so a note that comes and goes would walk the whole rail up
              and down the frame between beats. */}
          <div
            style={{
              minHeight: 76,
              marginTop: 12,
              maxWidth: 640,
              fontFamily: theme.fontBody,
              fontSize: 26,
              lineHeight: 1.35,
              color: theme.textMuted,
              opacity:
                level?.note && !whole
                  ? interpolate(sinceStep, [12, 26], [0, 1], {
                      extrapolateLeft: 'clamp',
                      extrapolateRight: 'clamp',
                    })
                  : 0,
            }}
          >
            {level?.note}
          </div>

          {/* The rail: the whole ladder, biggest at the top, with the real
              multipliers in the gutter between rungs. It is the readable
              artifact — the nest gives the feeling, this gives the facts. */}
          <div
            style={{
              width: '100%',
              marginTop: 46,
              paddingTop: 26,
              borderTop: `1px solid ${withAlpha(theme.text, 0.14)}`,
            }}
          >
            {levels
              .map((l, i) => ({l, i}))
              .reverse()
              .map(({l, i}) => {
                const on = whole || i <= current;
                const isCurrent = i === current && !whole;
                return (
                  <div key={i}>
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 14,
                        padding: '7px 0',
                        opacity: on ? 1 : 0.34,
                      }}
                    >
                      <div
                        style={{
                          width: isCurrent ? 14 : 8,
                          height: isCurrent ? 14 : 8,
                          borderRadius: '50%',
                          flexShrink: 0,
                          background: on ? theme.accentQuantity : withAlpha(theme.text, 0.3),
                        }}
                      />
                      <div
                        style={{
                          fontFamily: theme.fontMono,
                          fontSize: 19,
                          letterSpacing: 1.4,
                          color: isCurrent ? theme.text : theme.textMuted,
                          width: 132,
                          fontVariantNumeric: 'tabular-nums',
                        }}
                      >
                        {l.display}
                      </div>
                      <div
                        style={{
                          fontFamily: theme.fontBody,
                          fontSize: 21,
                          color: isCurrent ? theme.text : theme.textMuted,
                          whiteSpace: 'nowrap',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                        }}
                      >
                        {l.label}
                      </div>
                    </div>
                    {/* The step DOWN from this rung to the next one listed, in
                        the gutter between them. Set on the row it belongs to,
                        a multiplier reads as a property of that rung — which is
                        the one thing it is not: it is the distance between two. */}
                    {l.times ? (
                      <div
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: 12,
                          paddingLeft: 4,
                          opacity: on ? 0.9 : 0.3,
                        }}
                      >
                        <div
                          style={{
                            width: 1,
                            height: 16,
                            marginLeft: 3,
                            background: withAlpha(theme.text, 0.28),
                          }}
                        />
                        <div
                          style={{
                            fontFamily: theme.fontMono,
                            fontSize: 15,
                            letterSpacing: 1.6,
                            color: on ? theme.accentText : theme.textMuted,
                          }}
                        >
                          {`×${l.times}`}
                        </div>
                      </div>
                    ) : null}
                  </div>
                );
              })}
          </div>
        </div>
      </div>
    </Stage>
  );
};
