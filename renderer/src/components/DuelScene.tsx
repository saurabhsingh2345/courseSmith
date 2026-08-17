import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';
import {
  CardLabel,
  CardName,
  CardPanel,
  CardPill,
  CardSubject,
  CardTile,
  cardColour,
} from './showroomCard';

// DuelScene is two named things face to face with one measured bar each.
//
// The whole frame is a single sentence: these two, on this axis, and here is the
// gap. Everything on a card serves one clause of it — the mark says which thing,
// the pill says on what terms, the note says what it is, the bar says how much.
//
// The bars are the reason this is not just a two-card row, and they are drawn
// against a SHARED track rather than each being scaled to its own value. That is
// the whole point: a bar is a length the eye compares to the length beside it, and
// two bars normalised independently would both run the full width and say nothing.
// The track is the same width on both cards, at the same place on both cards, so
// the difference between the fills IS the claim.
//
// They fill once, on their own beat, and stay filled. A bar that re-animated every
// time the beat changed would say the measurement was in doubt.

const ROW_W = Math.min(STAGE_W, 1420);

// The gap between the two cards. Wider than the cards template's, because what
// sits in it here is the connector AND the whole argument — two things facing each
// other need to look like they are facing each other rather than touching.
const GAP = 96;

const TILE = 152;

const CARD_MIN_H = 620;

/** The bar's track, in card-relative terms. */
const BAR_H = 20;

type Side = CardSubject & {
  tag?: string;
  score?: number;
};

type Step = {
  startMs: number;
  endMs: number;
  show: 'pair' | 'card' | 'bars' | 'call';
  at?: number;
  lit?: number[];
  bars?: boolean;
};

/**
 * One side's measurement.
 *
 * `fill` is 0 before the measuring beat and animates to the score on it, so the
 * component never has to know which beat it is in — the timeline already decided
 * (see duelScenes) and passes the state down.
 */
const Bar: React.FC<{
  theme: ResolvedTheme;
  colour: string;
  score: number;
  fill: number;
  lit: boolean;
}> = ({theme, colour, score, fill, lit}) => {
  const pct = Math.max(0, Math.min(100, score)) * fill;
  // A measured zero has to look different from an unmeasured track, and it did
  // not. Found in the first real clip: the model picked "monthly cost" as the
  // axis, which puts a free tier at 0, and a 0% fill is an empty well — the exact
  // frame the same card showed thirty seconds earlier, before anything had been
  // measured. The viewer cannot tell "this costs nothing" from "the bar has not
  // filled yet", and one of those is the whole point of the beat.
  //
  // So a fill that has happened is never narrower than the bar is tall: at its
  // minimum it is the round end-cap of a bar and nothing more. That is honest —
  // it adds no length the eye would read as quantity, it just says the bar starts
  // here and goes nowhere. A zero that is genuinely zero should look like zero,
  // not like missing data.
  const measured = fill > 0.02;
  return (
    <div
      style={{
        width: '100%',
        height: BAR_H,
        borderRadius: 999,
        // The empty track is a recessed well in the card, the same well the cards
        // template's unanswered slot uses. One idea for "there is a place here and
        // nothing in it yet" across the family.
        background: withAlpha(theme.text, 0.07),
        overflow: 'hidden',
      }}
    >
      <div
        style={{
          width: `${pct}%`,
          minWidth: measured ? BAR_H : 0,
          height: '100%',
          borderRadius: 999,
          background: lit ? colour : withAlpha(colour, 0.45),
        }}
      />
    </div>
  );
};

export const DuelScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const sides = (Array.isArray(props.sides) ? props.sides : []) as Side[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const axis = String(props.axis ?? '');
  const verdict = String(props.verdict ?? '');
  const pick = Number(props.pick ?? 0);
  if (sides.length !== 2 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const litSet = new Set(step.lit ?? [0, 1]);
  const current = step.show === 'card' ? (step.at ?? 0) : step.show === 'call' ? pick : -1;

  const settle = interpolate(sinceStep, [2, 14], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // The fill. On the measuring beat it runs from nothing to the score over about
  // three quarters of a second, slower than the light changes elsewhere in the
  // family because it is the one movement on this frame the viewer is meant to
  // WATCH rather than register. On every beat after it, it is simply 1.
  const fill = step.show === 'bars' ? settle : step.bars ? 1 : 0;

  const on = interpolate(frame, [0, 20], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <Stage justify="center">
      <div style={{width: ROW_W}}>
        <SceneHeader
          theme={theme}
          title={String(props.title ?? '')}
          emphasis={props.emphasis as string | undefined}
          emphasisRole={props.emphasisRole as string | undefined}
          size="compact"
          marginBottom={46}
        />

        <div style={{display: 'flex', alignItems: 'stretch', gap: GAP, position: 'relative'}}>
          {sides.map((s, i) => {
            const lit = litSet.has(i);
            const selected = current === i;
            const colour = cardColour(theme, s);
            const enter = interpolate(frame, [3 + i * 7, 23 + i * 7], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            });
            return (
              <CardPanel
                key={i}
                theme={theme}
                colour={colour}
                selected={selected}
                style={{
                  flex: 1,
                  minWidth: 0,
                  minHeight: CARD_MIN_H,
                  padding: '44px 34px 40px',
                  opacity: enter,
                  transform: `translateY(${(1 - enter) * 22 - (selected ? settle * 10 : 0)}px)`,
                }}
              >
                <div style={{marginTop: 'auto', marginBottom: 24}}>
                  <CardTile theme={theme} subject={s} size={TILE} lit={lit} />
                </div>

                <CardName theme={theme} size={52} lit={lit}>
                  {s.title}
                </CardName>

                {s.tag ? (
                  <div style={{marginTop: 18}}>
                    <CardPill theme={theme} colour={colour} lit={lit}>
                      {s.tag}
                    </CardPill>
                  </div>
                ) : null}

                <div
                  style={{
                    marginTop: 20,
                    fontFamily: theme.fontBody,
                    fontSize: 26,
                    lineHeight: 1.35,
                    fontWeight: 500,
                    color: lit ? theme.text : theme.textMuted,
                  }}
                >
                  {s.note}
                </div>

                {/* The measurement, pinned to the foot of the card. Both cards put
                    it at the same height because both cards are the same height
                    and both pin it the same way — which is what makes the two
                    lengths comparable at a glance. */}
                <div style={{marginTop: 'auto', paddingTop: 30, width: '100%'}}>
                  <CardLabel theme={theme} lit={lit} style={{marginBottom: 14, textAlign: 'left'}}>
                    {axis}
                  </CardLabel>
                  <Bar
                    theme={theme}
                    colour={colour}
                    score={Number(s.score ?? 0)}
                    fill={fill}
                    lit={lit}
                  />
                </div>
              </CardPanel>
            );
          })}

          {/* The connector, centred in the gap between the two cards. Absolute
              rather than a third flex child so the two cards stay exactly the
              same width — a chip in the flow would take its own space out of one
              of them, and two cards of different widths is a duel that has
              already picked a winner. */}
          <div
            style={{
              position: 'absolute',
              left: '50%',
              top: '50%',
              transform: 'translate(-50%, -50%)',
              fontFamily: theme.fontDisplay,
              fontSize: 30,
              fontWeight: 700,
              padding: '10px 20px',
              borderRadius: 999,
              color: withAlpha(theme.text, 0.62),
              background: theme.surface,
              border: `1.5px solid ${theme.surfaceBorder}`,
              opacity: on,
            }}
          >
            vs
          </div>
        </div>

        {/* The call. Fixed-height box so nothing above it moves when it lands. */}
        <div style={{minHeight: 96, marginTop: 38, textAlign: 'center'}}>
          {step.show === 'call' && verdict ? (
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 42,
                fontWeight: 700,
                letterSpacing: -0.6,
                lineHeight: 1.3,
                color: theme.accentText,
                maxWidth: 1300,
                margin: '0 auto',
                opacity: settle,
                transform: `translateY(${(1 - settle) * 12}px)`,
              }}
            >
              {verdict}
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
