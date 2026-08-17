import {ResolvedTheme, seat, withAlpha} from '../theme/theme';
import {iconFor} from './icons';
import {BrandMark, cardArtTreatment} from './brandmark';

// The card the showroom family is built out of.
//
// Three templates draw the same object — a white panel with a real brand mark on
// it, a name, and something underneath that carries information — and they differ
// in how many of them are on screen and what the something is. Before this file
// existed the first of them had that panel inlined, and the second copy of it
// immediately drifted: a different corner radius, a different tile size, a
// shadow composed by hand. Two cards that are nearly the same object read worse
// than two that are obviously different ones, because the eye reports the
// mismatch without being able to say what it saw.
//
// So the panel, the tile, the mark and the seating live here, and a template
// supplies what goes below the name.

/** The art and identity of one card. Every field but `title` is optional. */
export type CardSubject = {
  title: string;
  note?: string;
  role?: string;
  icon?: string;
  mark?: string;
  image?: string;
  tint?: string;
};

/** Which semantic accent a card falls back to when its brand has no colour. */
export const roleColour = (theme: ResolvedTheme, role?: string): string => {
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
 * The colour a card is rimmed, tinted and (when there is no fetched mark)
 * lettered in: the brand's own when the fetch found one, its role's otherwise.
 *
 * That ordering is the point of carrying a tint at all. A lit Gemini card rimmed
 * in Gemini's blue-violet is Gemini being spoken about; the same card rimmed in
 * the course accent is a card being spoken about.
 */
export const cardColour = (theme: ResolvedTheme, subject: CardSubject): string =>
  cardArtTreatment(subject.tint, theme.surface).usable ?? roleColour(theme, subject.role);

/**
 * The mark on its tile.
 *
 * Exported separately from the card because the duel draws its two marks inside a
 * layout the card cannot own — but it must draw exactly this tile, at exactly this
 * size, or the two templates stop looking like one family.
 */
export const CardTile: React.FC<{
  theme: ResolvedTheme;
  subject: CardSubject;
  size: number;
  lit: boolean;
}> = ({theme, subject, size, lit}) => {
  const art = cardArtTreatment(subject.tint, theme.surface);
  const colour = art.usable ?? roleColour(theme, subject.role);
  const Icon = iconFor(subject.icon);
  // A favicon brings its own background, and it is almost always opaque white.
  //
  // Found in the first real render rather than in a fixture, because the fixtures
  // all use vector marks: ChatGPT falls back to the favicon service (Simple Icons
  // dropped OpenAI's marks), and the white bitmap landed on the pale tint every
  // other tile gets — a white square inside a coloured square, which reads as a
  // loading state. A bitmap gets the card's own surface behind it instead, so
  // whatever background it carries merges with the tile rather than sitting on it.
  const bitmap = !subject.mark && !!subject.image;
  return (
    <div
      style={{
        width: size,
        height: size,
        flex: 'none',
        // Proportional to the tile rather than fixed, so a hero tile and a row
        // tile are the same SHAPE at different sizes. A constant radius makes the
        // small one look rounder than the large one.
        borderRadius: size * 0.19,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        // Clipped, so a bitmap drawn to the tile's edge takes the tile's corners
        // instead of squaring them off.
        overflow: 'hidden',
        background: bitmap ? theme.surface : (art.plate ?? withAlpha(colour, lit ? 0.12 : 0.07)),
        border: `1px solid ${withAlpha(bitmap || art.plate ? theme.text : colour, lit ? 0.16 : 0.1)}`,
      }}
    >
      {subject.mark ? (
        <BrandMark
          path={subject.mark}
          tint={subject.tint}
          fallback={lit ? colour : theme.textMuted}
          surface={theme.surface}
          size={size * 0.54}
          lit={lit}
        />
      ) : subject.image ? (
        <img
          src={subject.image}
          alt=""
          style={{
            // Larger than a vector mark's 0.54. A favicon is padded by whoever
            // drew it — usually generously, for a 16-pixel tab — so the visible
            // glyph inside a 0.58 box is smaller than the vector next to it, and a
            // row that mixes the two reads as one logo that failed to load at
            // full size.
            width: size * 0.74,
            height: size * 0.74,
            objectFit: 'contain',
            // A bitmap made for a 16-pixel browser tab is going up eight times
            // its size. Crisp edges beat the browser's smeared interpolation.
            imageRendering: 'crisp-edges',
            opacity: lit ? 1 : 0.65,
          }}
        />
      ) : (
        <Icon size={size * 0.46} strokeWidth={1.7} color={lit ? colour : theme.textMuted} />
      )}
    </div>
  );
};

/**
 * The panel itself: surface, hairline, seating, and the selected state.
 *
 * `lit` is quiet-or-loud contents and `selected` is the one being spoken about
 * right now. They are separate because a row of three has one selected card and
 * two lit ones on its opening beat, and collapsing them would make the opener
 * pick a winner before the argument started.
 *
 * A dim card keeps its FULL opacity, which is the light-mode correction this
 * family cost most to learn. On the dark stage fading a card recedes it, because
 * what fades is a surface lighter than the ground behind it. On paper the card is
 * the brightest thing in the frame, so fading it fades it into the page: at 0.62
 * a row of three read as three cards that had failed to load. What recedes a card
 * here is its contents going quiet and its rim light going out.
 */
export const CardPanel: React.FC<{
  theme: ResolvedTheme;
  colour: string;
  selected: boolean;
  style?: React.CSSProperties;
  children: React.ReactNode;
}> = ({theme, colour, selected, style, children}) => (
  <div
    style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      textAlign: 'center',
      borderRadius: 24,
      background: theme.surface,
      border: `1.5px solid ${selected ? withAlpha(colour, 0.7) : theme.surfaceBorder}`,
      boxShadow: [
        seat(theme, selected ? 'lifted' : 'resting'),
        // A spread ring rather than the soft outer blur the dark stage uses: a
        // blurred halo on white is a grubby edge, while a hard ring at low alpha
        // reads as the card being selected.
        selected ? `0 0 0 5px ${withAlpha(colour, 0.18)}` : '',
      ]
        .filter(Boolean)
        .join(', '),
      ...style,
    }}
  >
    {children}
  </div>
);

/** The name, at whichever of the family's two sizes the template needs. */
export const CardName: React.FC<{
  theme: ResolvedTheme;
  children: React.ReactNode;
  size: number;
  lit: boolean;
}> = ({theme, children, size, lit}) => (
  <div
    style={{
      fontFamily: theme.fontDisplay,
      fontSize: size,
      fontWeight: 800,
      letterSpacing: -1,
      lineHeight: 1.1,
      color: lit ? theme.text : theme.textMuted,
    }}
  >
    {children}
  </div>
);

/** A small-caps label — the row's shared question, a bar's axis, a card's tier. */
export const CardLabel: React.FC<{
  theme: ResolvedTheme;
  children: React.ReactNode;
  lit?: boolean;
  style?: React.CSSProperties;
}> = ({theme, children, lit = true, style}) => (
  <div
    style={{
      fontFamily: theme.fontBody,
      fontSize: 17,
      fontWeight: 600,
      letterSpacing: 2.4,
      textTransform: 'uppercase',
      color: lit ? theme.textMuted : withAlpha(theme.textMuted, 0.55),
      ...style,
    }}
  >
    {children}
  </div>
);

/**
 * The pill under a name: a dot and a word or two.
 *
 * The dot is not decoration. Without it the pill is a rounded rectangle with text
 * in it, which on a white card is indistinguishable from a button — and a frame
 * with buttons on it reads as a screenshot of an interface rather than as a card
 * about a product. The dot makes it a status.
 */
export const CardPill: React.FC<{
  theme: ResolvedTheme;
  colour: string;
  lit: boolean;
  children: React.ReactNode;
}> = ({theme, colour, lit, children}) => (
  <div
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      gap: 10,
      padding: '9px 18px',
      borderRadius: 999,
      background: lit ? withAlpha(colour, 0.11) : withAlpha(theme.text, 0.05),
      fontFamily: theme.fontBody,
      fontSize: 24,
      fontWeight: 600,
      color: lit ? theme.text : theme.textMuted,
    }}
  >
    <span
      style={{
        width: 10,
        height: 10,
        borderRadius: 999,
        background: lit ? colour : withAlpha(theme.text, 0.28),
      }}
    />
    {children}
  </div>
);
