import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
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
// The detail line lives under the row rather than inside the card, because a
// card that grows when it lights would reflow the whole row and the eye would
// lose its place. The row is fixed furniture; only brightness moves.

const COL_W = Math.min(STAGE_W, 1620);
const CARD_H = 300;

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
            fontSize: 60,
            fontWeight: 800,
            letterSpacing: -1.4,
            lineHeight: 1.12,
            color: theme.text,
            textAlign: 'center',
            marginBottom: 52,
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
            const c = lit ? theme.accentQuantity : theme.textMuted;
            return (
              <div
                key={i}
                style={{
                  flex: 1,
                  position: 'relative',
                  height: CARD_H,
                  padding: '30px 28px',
                  borderRadius: 16,
                  overflow: 'hidden',
                  background: lit ? withAlpha(theme.accentQuantity, 0.11) : withAlpha(theme.text, 0.04),
                  border: `1px solid ${lit ? withAlpha(theme.accentQuantity, 0.42) : withAlpha(theme.text, 0.1)}`,
                  opacity: on * (lit || current < 0 ? 1 : 0.5),
                  transform: `translateY(${(1 - on) * 18}px)`,
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
                      fontSize: 92,
                      fontWeight: 800,
                      lineHeight: 0.9,
                      letterSpacing: -4,
                      color: lit ? theme.accentQuantity : withAlpha(theme.text, 0.22),
                      fontVariantNumeric: 'tabular-nums',
                    }}
                  >
                    {String(i + 1).padStart(2, '0')}
                  </div>
                  <RowIcon name={it.icon} color={c} lit={lit} />
                </div>
                <div
                  style={{
                    position: 'relative',
                    fontFamily: theme.fontDisplay,
                    fontSize: 34,
                    fontWeight: 700,
                    lineHeight: 1.2,
                    letterSpacing: -0.5,
                    color: lit ? theme.text : theme.textMuted,
                  }}
                >
                  {it.label}
                </div>
              </div>
            );
          })}
        </div>

        {/* The detail, under the row rather than inside the card — a card that
            grew when it lit would reflow the row and the eye would lose its
            place. */}
        <div style={{minHeight: 92, marginTop: 34, textAlign: 'center'}}>
          {current >= 0 && items[current]?.detail ? (
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 31,
                lineHeight: 1.4,
                color: theme.textMuted,
                maxWidth: 1180,
                margin: '0 auto',
                opacity: interpolate(sinceStep, [3, 17], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                }),
                transform: `translateY(${
                  (1 -
                    interpolate(sinceStep, [3, 17], [0, 1], {
                      extrapolateLeft: 'clamp',
                      extrapolateRight: 'clamp',
                    })) *
                  12
                }px)`,
              }}
            >
              {items[current].detail}
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
