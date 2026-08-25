import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, seat, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {iconFor} from './icons';

// RundownScene is a numbered row that lights one card at a time.
//
// Every card is on screen from the first frame, numbered and dim. That is the
// design decision the whole template turns on: a rundown that reveals its items
// one by one is a list again, and the viewer loses the one piece of information
// the format exists to give them, which is how far through they are.
//
// The numbers are set very large and behind the label rather than beside it —
// a ghosted 01 / 02 / 03 filling each card. It reads as an index rather than as
// a bullet, and at this size the row is legible as "three things" from across a
// room, before any word has been read.
//
// The detail line lives INSIDE the card, and it is there from the first frame.
//
// It used to sit under the row, on the argument that a card which grew when it lit
// would reflow the row and the eye would lose its place. That reasoning was right
// about reflow and wrong about where to put the line: a fixed slot inside the card,
// filled from the start and merely brightened on the beat, does not reflow either —
// and it means a paused frame is a row of finished cards rather than a row of
// labels with one caption underneath. The band under the row was also 110 pixels
// held empty for four beats out of five.
//
// What a beat moves is INK, never opacity. A card faded to 0.5 on a dark stage
// recedes, because what fades is a panel lighter than the ground behind it; on
// paper the card is the brightest thing in the frame, so the same fade bleaches it
// into the page and a row of four reads as one card and three that failed to load.
// The surface and its hairline come from the theme's own tokens, which are derived
// per mode, so the card is a card in both and neither needs a branch here.

const COL_W = Math.min(STAGE_W, 1620);

// 420, not 300.
//
// Everything in this file was proportioned against a stage that no longer
// exists. With the caption band no longer reserved unconditionally there are 952
// vertical pixels to compose in, and a 300-tall row plus its title and detail
// line came to about 550 of them — the row sat as a band across the middle with a
// third of the frame empty dark below it, which is most of why a rundown read as
// a slide rather than as a shot.
//
// Sized so the whole composition lands around 75% of the stage: enough that the
// frame is filled, short of the edge-to-edge packing that would leave the numbers
// nowhere to breathe.
const CARD_H = 420;

type Item = {label: string; detail?: string; icon?: string};

/** The card's glyph. iconFor returns a component, so it is rendered rather than
 *  called for markup. */
const RowIcon: React.FC<{name?: string; color: string; lit: boolean}> = ({name, color, lit}) => {
  const Icon = iconFor(name);
  return (
    <div style={{position: 'relative', color, opacity: lit ? 1 : 0.75, lineHeight: 0}}>
      <Icon size={38} strokeWidth={1.9} />
    </div>
  );
};
type Step = {startMs: number; endMs: number; show: 'promise' | 'item' | 'all'; at?: number};

export const RundownScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const promise = String(props.promise ?? '');
  const items = (Array.isArray(props.items) ? props.items : []) as Item[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (items.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;
  const all = step.show === 'all';
  const current = step.show === 'item' ? (step.at ?? 0) : -1;

  const enter = spring({
    frame: ((nowMs - steps[0].startMs) / 1000) * FPS,
    fps,
    config: {damping: 200, mass: 0.7},
    durationInFrames: 20,
  });

  return (
    <Stage justify="center">
      <div style={{width: COL_W, opacity: enter}}>
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 68,
            fontWeight: 800,
            letterSpacing: -1.6,
            lineHeight: 1.12,
            color: theme.text,
            textAlign: 'center',
            marginBottom: 56,
          }}
        >
          {promise}
        </div>

        <div style={{display: 'flex', gap: 24}}>
          {items.map((it, i) => {
            const lit = all || i === current;
            // Cards enter on a stagger on the opening beat, so the row
            // assembles left to right rather than appearing all at once.
            const on = interpolate(frame, [4 + i * 5, 20 + i * 5], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            });
            // The lit card lifts four pixels on its own beat. Small on purpose:
            // the seat shadow under it does the work, and a card that jumps reads
            // as an interface responding to a click rather than as a shot cutting.
            const rise =
              i === current
                ? interpolate(sinceStep, [0, 14], [0, 1], {
                    extrapolateLeft: 'clamp',
                    extrapolateRight: 'clamp',
                  })
                : 0;
            const c = lit ? theme.accentQuantity : theme.textMuted;
            return (
              <div
                key={i}
                style={{
                  flex: 1,
                  position: 'relative',
                  height: CARD_H,
                  padding: '38px 32px',
                  borderRadius: 16,
                  overflow: 'hidden',
                  background: theme.surface,
                  border: `1.5px solid ${lit ? withAlpha(theme.accentQuantity, 0.6) : theme.surfaceBorder}`,
                  boxShadow: [
                    seat(theme, lit ? 'lifted' : 'resting'),
                    // A hard ring rather than a blurred halo: on white a blurred
                    // glow is a grubby edge, while a spread ring at low alpha
                    // reads as the card being the one under discussion.
                    lit ? `0 0 0 5px ${withAlpha(theme.accentQuantity, 0.16)}` : '',
                  ]
                    .filter(Boolean)
                    .join(', '),
                  opacity: on,
                  transform: `translateY(${(1 - on) * 18 - rise * 4}px)`,
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'space-between',
                }}
              >
                {/* The index, set large at the top of the card with the label
                    below it — the arrangement the reference uses, and the one
                    that keeps the numeral out of the label's way. A ghosted
                    numeral behind the text collided with any label longer than
                    two words, which is most of them. */}
                <div style={{display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between'}}>
                  <div
                    style={{
                      fontFamily: theme.fontDisplay,
                      fontSize: 116,
                      fontWeight: 800,
                      lineHeight: 0.9,
                      letterSpacing: -5,
                      color: lit ? theme.accentQuantity : withAlpha(theme.text, 0.22),
                      fontVariantNumeric: 'tabular-nums',
                    }}
                  >
                    {String(i + 1).padStart(2, '0')}
                  </div>
                  <RowIcon name={it.icon} color={c} lit={lit} />
                </div>
                <div style={{display: 'flex', flexDirection: 'column', gap: 18}}>
                  <div
                    style={{
                      position: 'relative',
                      fontFamily: theme.fontDisplay,
                      fontSize: 40,
                      fontWeight: 700,
                      lineHeight: 1.2,
                      letterSpacing: -0.6,
                      color: lit ? theme.text : theme.textMuted,
                    }}
                  >
                    {it.label}
                  </div>
                  {it.detail ? (
                    <div
                      style={{
                        fontFamily: theme.fontBody,
                        fontSize: 26,
                        fontWeight: 500,
                        lineHeight: 1.36,
                        color: lit ? theme.text : theme.textMuted,
                      }}
                    >
                      {it.detail}
                    </div>
                  ) : null}
                </div>
              </div>
            );
          })}
        </div>

      </div>
    </Stage>
  );
};
