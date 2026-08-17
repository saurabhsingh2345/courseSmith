import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';
import {iconFor} from './icons';

// CardsScene is a row of named things, each wearing its own mark.
//
// The row is up from the first frame and never reflows. That is the one layout
// decision everything else follows from: these cards carry logos, and a logo the
// eye has already located is a logo it can stop looking for — so the frame's job
// after the opening beat is to move LIGHT around a fixed arrangement, never to
// move the arrangement. Nothing here changes width, changes place, or appears
// late. A lit card lifts a few pixels and gains its colour; the others lose
// theirs. The note under the row is the only thing that swaps.
//
// The mark comes down three ways and the component treats them as three
// different objects rather than one image slot.
//
// `mark` is SVG path data on a 24x24 grid — a real brand mark with its colour
// stripped in Go, painted here in the card's own accent. This is the good case
// and the reason the fetch bothers extracting geometry instead of inlining the
// file: a logo that takes the theme sits on the stage like everything else,
// while one that arrives pre-coloured is a sticker somebody put on the frame.
//
// `image` is a data URI that has to keep its own colours — a favicon, a photo.
// It gets a light plate behind it rather than the accent wash, because a
// full-colour bitmap on a tinted tile reads as a colour cast.
//
// Neither: the Lucide glyph, drawn exactly like a mark. That is the floor, and
// it is why a failed fetch is a slightly plainer card rather than a hole.
//
// The connector is drawn once between each pair and it is the template's whole
// argument in two characters. `vs` is set in mono at the row's midline; `then`
// is an arrow; `none` leaves the gap empty, which is a claim too — that these
// things are simply the players.

const ROW_W = Math.min(STAGE_W, 1660);

// The card, and everything on it, is proportioned against this. 340 leaves the
// composition — header, row, note, closer — landing around 80% of the stage,
// which is what stops a two-card row reading as a diagram floating in dark.
//
// The contents are CENTRED in it rather than stacked from the top. A card is a
// fixed-height object holding two things of variable size, and top-aligning them
// left a third of every card empty — which on a row of five reads as five cards
// that failed to load rather than as deliberate air.
const CARD_H = 340;

// The gap a connector sits in. Fixed rather than proportional: "vs" is the same
// size whether there are two cards or five, so the space it needs is too.
const GAP = 76;

// The art tile. Big enough that a brand mark is identified rather than squinted
// at, short of the size where the logo becomes the card and the words under it
// become a caption.
const TILE = 118;

type Item = {
  title: string;
  note?: string;
  role?: string;
  icon?: string;
  mark?: string;
  image?: string;
};

type Step = {
  startMs: number;
  endMs: number;
  show: 'row' | 'card' | 'all';
  at?: number;
  lit?: number[];
};

const roleColour = (theme: ResolvedTheme, role?: string): string => {
  switch (role) {
    case 'limit':
      return theme.accentLimit;
    case 'rival':
      return theme.accentRival;
    case 'quantity':
      return theme.accentQuantity;
    default:
      return theme.accentText;
  }
};

/**
 * The card's art, in whichever of the three forms arrived.
 *
 * All three land on the same 24x24-equivalent box so a row that mixes a fetched
 * brand mark with a fallback glyph still reads as one set of cards — which it
 * will, most of the time, because not everything worth putting in a row has a
 * logo in an icon set.
 */
const CardArt: React.FC<{item: Item; colour: string; lit: boolean}> = ({item, colour, lit}) => {
  if (item.mark) {
    return (
      <svg viewBox="0 0 24 24" width={TILE * 0.52} height={TILE * 0.52} aria-hidden>
        <path d={item.mark} fill={colour} />
      </svg>
    );
  }
  if (item.image) {
    return (
      <img
        src={item.image}
        alt=""
        style={{
          width: TILE * 0.56,
          height: TILE * 0.56,
          objectFit: 'contain',
          // A bitmap made for a browser tab is going up eight times its size.
          // Crisp edges beat the smeared interpolation the browser would
          // otherwise apply, which is the difference between a small logo and a
          // blurred one.
          imageRendering: 'crisp-edges',
          opacity: lit ? 1 : 0.72,
        }}
      />
    );
  }
  const Icon = iconFor(item.icon);
  return <Icon size={TILE * 0.46} strokeWidth={1.7} color={colour} />;
};

export const CardsScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const relation = String(props.relation ?? 'none');
  const closer = String(props.closer ?? '');
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

  const note = current >= 0 ? items[current]?.note : undefined;
  const connector = relation === 'versus' ? 'vs' : relation === 'then' ? '→' : '';

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
            const colour = roleColour(theme, it.role);
            // The stagger is the only thing in this scene that arrives late, and
            // it only happens once: the row builds left to right on the opener
            // so the count registers as a count.
            const on = interpolate(frame, [3 + i * 6, 21 + i * 6], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            });
            // A card lifts when it is the one being spoken about. Three pixels
            // and a deeper shadow, not a scale: scaling a card with a logo on it
            // resamples the logo, and the row stops looking machined.
            const lift = single ? swap * 10 : 0;
            return (
              <div key={i} style={{display: 'flex', flex: 1, minWidth: 0}}>
                {i > 0 ? (
                  <div
                    style={{
                      width: GAP,
                      flex: 'none',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontFamily: theme.fontMono,
                      fontSize: relation === 'then' ? 44 : 32,
                      letterSpacing: relation === 'then' ? 0 : 1,
                      // The connector is the template's whole argument and it
                      // gets two characters to make it, so it is set at the same
                      // weight as a dimmed card rather than below one — quieter
                      // than that and the row reads as cards with gaps.
                      color: withAlpha(theme.text, 0.58),
                      opacity: on,
                    }}
                  >
                    {connector}
                  </div>
                ) : null}
                <div
                  style={{
                    flex: 1,
                    minWidth: 0,
                    height: CARD_H,
                    padding: '34px 28px',
                    borderRadius: 18,
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    textAlign: 'center',
                    // The surface, not a tint of the accent. A row of five
                    // accent-washed cards is five coloured panels and the lit
                    // one has nowhere to go; keeping the card neutral means
                    // colour itself is the signal.
                    background: theme.surface,
                    border: `1px solid ${lit ? withAlpha(colour, 0.5) : theme.surfaceBorder}`,
                    // Three layers, and on a near-black stage the order of
                    // importance is the reverse of what it would be on paper.
                    //
                    // A drop shadow is what lifts an object off a light page and
                    // it does almost nothing here: black on near-black is black.
                    // What actually seats these cards is the INSET hairline
                    // along the top edge — a rim light, the one-pixel highlight a
                    // real object catches from a stage lit from above — plus the
                    // surface being a shade lighter than the ground behind it.
                    // The drop shadow is still cast, softly and in real black
                    // rather than the theme's ink, because it darkens the stage
                    // immediately under each card and that gradient is what
                    // stops a row of five reading as one long panel.
                    boxShadow: [
                      'inset 0 1px 0 rgba(255,255,255,0.07)',
                      single ? '0 28px 64px rgba(0,0,0,0.62)' : '0 20px 46px rgba(0,0,0,0.5)',
                      single ? `0 0 0 4px ${withAlpha(colour, 0.14)}` : '',
                    ]
                      .filter(Boolean)
                      .join(', '),
                    // 0.6, not 0.46. A card the eye cannot read is a card that
                    // has left the row, and the row's whole job is that all of
                    // them stay countable while one of them is being spoken
                    // about — receded, not removed.
                    opacity: on * (lit ? 1 : 0.6),
                    transform: `translateY(${(1 - on) * 20 - lift}px)`,
                  }}
                >
                  <div
                    style={{
                      width: TILE,
                      height: TILE,
                      flex: 'none',
                      borderRadius: 22,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      marginBottom: 26,
                      // A vector mark sits on a wash of its own colour; a
                      // full-colour bitmap sits on a neutral plate, because a
                      // tinted tile under a coloured icon reads as a cast.
                      background: it.image
                        ? withAlpha(theme.text, 0.07)
                        : withAlpha(colour, lit ? 0.14 : 0.08),
                      border: `1px solid ${withAlpha(it.image ? theme.text : colour, lit ? 0.3 : 0.16)}`,
                    }}
                  >
                    <CardArt item={it} colour={lit ? colour : theme.textMuted} lit={lit} />
                  </div>

                  <div
                    style={{
                      fontFamily: theme.fontDisplay,
                      fontSize: items.length > 3 ? 40 : 48,
                      fontWeight: 800,
                      letterSpacing: -1,
                      lineHeight: 1.1,
                      color: lit ? theme.text : theme.textMuted,
                    }}
                  >
                    {it.title}
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* The line under the row. One box, fixed height, whichever of the two
            things is due — the card's note while a card is lit, the closer on
            the last beat. Two stacked boxes would move the row every time one of
            them appeared, which is the reflow this layout refuses. */}
        <div style={{minHeight: 118, marginTop: 40, textAlign: 'center', opacity: raise}}>
          {note ? (
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 38,
                lineHeight: 1.35,
                color: theme.text,
                maxWidth: 1260,
                margin: '0 auto',
                opacity: swap,
                transform: `translateY(${(1 - swap) * 12}px)`,
              }}
            >
              {items[current].note}
            </div>
          ) : step.show === 'all' && closer ? (
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 42,
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
