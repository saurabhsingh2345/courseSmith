import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';
import {
  CardLabel,
  CardName,
  CardPanel,
  CardSubject,
  CardTile,
  cardColour,
} from './showroomCard';

// CardsScene is a row of named things, each wearing its own mark.
//
// The row is up from the first frame and never reflows. That is the one layout
// decision everything else follows from: these cards carry logos, and a logo the
// eye has already located is a logo it can stop looking for — so the frame's job
// after the opening beat is to move LIGHT around a fixed arrangement, never to
// move the arrangement. Nothing here changes width, changes place, or appears
// late. A lit card gains its colour and lifts a few pixels; the others lose
// theirs.
//
// EVERYTHING IN A CARD IS IN THE CARD. That is the correction this component was
// rewritten for. It used to draw a logo and a name, and float the note under the
// whole row in a shared box — which meant a card was two things, and two things
// is a sticker. The look this template copies never has a card with fewer than
// three: a mark, a name, and something that carries information. Here the third
// thing is the note, and the row's shared question (`ask`) is what turns the note
// into an answer worth waiting for.
//
// The slot is the trick worth naming. With `ask` set, every card shows the same
// small-caps label above a slot reading `? ? ?`, and the note lands in that slot
// on the card's beat. It costs no layout — the slot is the card's shape from frame
// one, so nothing moves when it fills — and it converts a row the viewer has
// already read into a row they are waiting on.

const ROW_W = Math.min(STAGE_W, 1660);

// The gap a connector sits in. Fixed rather than proportional: the connector is
// the same size whether there are two cards or five, so the space it needs is.
const GAP = 74;

// The art tile.
const TILE = 138;

// The card's floor height, which makes it portrait rather than landscape.
//
// Not decoration. A card taller than it is wide reads as an object — a tile, a
// product shot — and one wider than it is tall reads as a row of a table that
// happens to have a picture in it. The number is set against the FRAME rather
// than against the card: at 560 the finished block (heading, row, closing line)
// occupies about three quarters of the stage's height, which is what the look
// being copied does. At 470 it occupied half, and a composition using half the
// frame reads as a slide somebody has not finished laying out.
const CARD_MIN_H = 560;

type Item = CardSubject;

type Step = {
  startMs: number;
  endMs: number;
  show: 'row' | 'card' | 'all';
  at?: number;
  lit?: number[];
};

/**
 * What goes in the gap between two cards.
 *
 * Drawn rather than typed, and that is not a detail. `vs` used to be two mono
 * characters and an arrow used to be the character `→`, which meant the
 * template's entire argument — are these alternatives, a sequence, or just the
 * players — was carried by whatever those glyphs happen to look like in the
 * course's font. A drawn arrow has a stroke weight chosen against the frame, and
 * `vs` in a chip reads as a label on the gap rather than as text that fell
 * between two cards.
 */
const Connector: React.FC<{relation: string; theme: ResolvedTheme; opacity: number}> = ({
  relation,
  theme,
  opacity,
}) => {
  if (relation === 'none') {
    return <div style={{width: GAP, flex: 'none'}} />;
  }
  const ink = withAlpha(theme.text, 0.55);
  return (
    <div
      style={{
        width: GAP,
        flex: 'none',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        opacity,
      }}
    >
      {relation === 'then' ? (
        <svg width={GAP - 18} height={22} viewBox="0 0 56 22" aria-hidden>
          <path d="M2 11 H42" stroke={ink} strokeWidth={3} strokeLinecap="round" fill="none" />
          <path
            d="M36 4 L47 11 L36 18"
            stroke={ink}
            strokeWidth={3}
            strokeLinecap="round"
            strokeLinejoin="round"
            fill="none"
          />
        </svg>
      ) : (
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 24,
            fontWeight: 700,
            letterSpacing: 0.5,
            padding: '7px 15px',
            borderRadius: 999,
            color: withAlpha(theme.text, 0.62),
            background: withAlpha(theme.text, 0.055),
          }}
        >
          vs
        </div>
      )}
    </div>
  );
};

export const CardsScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const relation = String(props.relation ?? 'none');
  const closer = String(props.closer ?? '');
  const ask = String(props.ask ?? '');
  const items = (Array.isArray(props.items) ? props.items : []) as Item[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (items.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const current = step.show === 'card' ? (step.at ?? 0) : -1;
  // On `row` and `all` every card is lit, so nothing is singled out; on a card
  // beat exactly one is. Reading it off the step rather than re-deriving it
  // keeps the frame a function of one object.
  const litSet = new Set(step.lit ?? items.map((_, i) => i));

  // The row assembles once, on the opening beat, and then holds. Keyed to the
  // clip rather than to the step: an entrance that replayed on every beat would
  // be the reflow this layout exists to avoid.
  const raise = interpolate(frame, [0, 22], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const swap = interpolate(sinceStep, [2, 14], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  // Once a card's beat has passed its answer stays on the card. The row builds up
  // to a finished board rather than resetting to `? ? ?` behind the voice — the
  // closing beat has to show every answer at once, or the clip ends having proved
  // nothing was learned.
  const answered = (i: number): boolean => {
    if (!ask) return true;
    if (step.show === 'all') return true;
    if (i === current) return swap > 0.35;
    for (let s = 0; s <= idx; s++) {
      if (steps[s].show === 'card' && steps[s].at === i) return true;
    }
    return false;
  };

  const noteSize = items.length > 3 ? 23 : 27;
  // The slot is given a FIXED height, sized here for the longest note in the row
  // rather than left to grow around whichever note is currently showing.
  //
  // This is the reflow rule again, in the one place it is easy to miss. Every
  // other element on these cards is the same size on every card, so the row's
  // height is settled before the first frame; a slot that sized itself to its
  // contents would make the row's height a function of which card is lit, and a
  // four-word answer following an eleven-word one would shorten every card in the
  // row mid-sentence. Measured off the longest string, so the tallest state is
  // the only state.
  const cardTextW = (ROW_W - (items.length - 1) * GAP) / items.length - 52;
  const longest = Math.max(1, ...items.map((it) => (it.note ?? '').length));
  // ~0.5em per character is the working average for Inter at these sizes. It is
  // an estimate and it only has to be conservative: over-estimating the wrap
  // costs a few pixels of air, under-estimating clips a note.
  const noteLines = Math.max(1, Math.ceil(longest / Math.max(10, cardTextW / (noteSize * 0.5))));
  const slotH = Math.round(noteLines * noteSize * 1.34) + 28;

  return (
    <Stage justify="center">
      <div style={{width: ROW_W}}>
        <SceneHeader
          theme={theme}
          title={String(props.title ?? '')}
          emphasis={props.emphasis as string | undefined}
          emphasisRole={props.emphasisRole as string | undefined}
          size="compact"
          marginBottom={44}
        />

        <div style={{display: 'flex', alignItems: 'stretch'}}>
          {items.map((it, i) => {
            const lit = litSet.has(i);
            const single = current === i;
            const colour = cardColour(theme, it);
            // The stagger is the only thing in this scene that arrives late, and
            // it only happens once: the row builds left to right on the opener
            // so the count registers as a count.
            const on = interpolate(frame, [3 + i * 6, 21 + i * 6], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            });
            // A card lifts when it is the one being spoken about. A few pixels
            // and a deeper shadow, not a scale: scaling a card with a logo on it
            // resamples the logo, and the row stops looking machined.
            const lift = single ? swap * 10 : 0;
            const show = answered(i);
            return (
              <div key={i} style={{display: 'flex', flex: 1, minWidth: 0}}>
                {i > 0 ? <Connector relation={relation} theme={theme} opacity={on} /> : null}
                <CardPanel
                  theme={theme}
                  colour={colour}
                  selected={single}
                  style={{
                    flex: 1,
                    minWidth: 0,
                    minHeight: CARD_MIN_H,
                    padding: '40px 26px 46px',
                    opacity: on,
                    transform: `translateY(${(1 - on) * 20 - lift}px)`,
                  }}
                >
                  <div
                    style={{
                      // An auto margin above the mark and another above the
                      // footer, which is what actually distributes this card.
                      // Flexbox splits the leftover height evenly between the
                      // two, so the mark-and-name group sits centred in the upper
                      // part and the labelled slot sits at the foot with equal
                      // air above each. Pinning only the footer left the name
                      // high with a visible hole under it — all the free space
                      // went to one place because only one thing asked for it.
                      marginTop: 'auto',
                      marginBottom: 22,
                    }}
                  >
                    <CardTile theme={theme} subject={it} size={TILE} lit={lit} />
                  </div>

                  <CardName theme={theme} size={items.length > 3 ? 38 : 44} lit={lit}>
                    {it.title}
                  </CardName>

                  {/* The third thing on the card. With `ask` it is a labelled
                      slot that fills on the card's beat; without it the note
                      simply sits under the name. Either way it is inside the
                      card, and the card is never just a logo and a name. */}
                  {ask ? (
                    <>
                      <CardLabel theme={theme} lit={lit} style={{marginTop: 'auto', paddingTop: 24}}>
                        {ask}
                      </CardLabel>
                      <div
                        style={{
                          marginTop: 12,
                          width: '100%',
                          height: slotH,
                          borderRadius: 14,
                          padding: '14px 16px',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          // The empty slot is a filled well rather than an
                          // outline. An outlined box on an unanswered card reads
                          // as an input waiting for the viewer to type in it; a
                          // recessed one reads as a fact not yet given.
                          background: show ? 'transparent' : withAlpha(theme.text, 0.05),
                          fontFamily: theme.fontBody,
                          fontSize: show ? noteSize : 30,
                          lineHeight: 1.3,
                          fontWeight: show ? 500 : 700,
                          letterSpacing: show ? 0 : 5,
                          color: show
                            ? lit
                              ? theme.text
                              : theme.textMuted
                            : withAlpha(theme.text, 0.28),
                        }}
                      >
                        {show ? it.note : '? ? ?'}
                      </div>
                    </>
                  ) : (
                    <div
                      style={{
                        marginTop: 'auto',
                        paddingTop: 22,
                        fontFamily: theme.fontBody,
                        fontSize: noteSize,
                        lineHeight: 1.35,
                        fontWeight: 500,
                        color: lit ? theme.text : theme.textMuted,
                      }}
                    >
                      {it.note}
                    </div>
                  )}
                </CardPanel>
              </div>
            );
          })}
        </div>

        {/* The closing line under the finished row. One box, fixed height, so
            the row does not move when it appears. */}
        <div style={{minHeight: 92, marginTop: 40, textAlign: 'center', opacity: raise}}>
          {step.show === 'all' && closer ? (
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 40,
                fontWeight: 700,
                letterSpacing: -0.6,
                lineHeight: 1.3,
                color: theme.accentText,
                maxWidth: 1360,
                margin: '0 auto',
                opacity: swap,
                transform: `translateY(${(1 - swap) * 12}px)`,
              }}
            >
              {closer}
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
