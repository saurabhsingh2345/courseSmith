import {contrastRatio, luminance, withAlpha} from '../theme/theme';

// Painting a fetched brand mark, and the one decision that makes it look real.
//
// The pipeline hands the renderer a logo as path data plus the brand's own hex
// (see snippet_cards_art.go). Painting it in that hex is the entire reason for
// fetching it — a four-pointed star in the course accent is a shape, the same
// star in Gemini's blue-violet is Gemini — but a brand colour is chosen to sit on
// the brand's own site, not on this card, and some of them are invisible here.
// JavaScript's yellow on white is a 1.35:1 mark: technically painted, actually
// absent.
//
// So there are two lockups, and which one a brand gets is decided by measurement
// rather than by taste.
//
// MARK ON A WASH, when the brand colour reads as ink on the card: the mark in its
// own colour, on a tile tinted a few percent of the same. This is the common case
// and the better-looking one — the tile is barely there and the logo is the
// object.
//
// MARK ON A PLATE, when it does not: the tile filled SOLID in the brand colour,
// with the mark knocked out in whichever of ink or paper contrasts with it. This
// is not a compromise, it is what the brand itself does — Anthropic ships a black
// mark on peach, JavaScript a black J-S on yellow — so the failing case lands on
// the more authentic lockup rather than on a fallback.
//
// The threshold is 3:1 against the card, which is the ratio below which a large
// solid shape stops being reliably distinguishable from what it sits on. Above it
// the wash; below it the plate.

const READABLE_ON_CARD = 3;

export type ArtTreatment = {
  /**
   * The colour to paint the mark and rim the card in — the brand's own, when it
   * is legible against the card. Undefined when there is no tint at all, which
   * is the caller's cue to fall back to its role colour.
   */
  usable?: string;
  /**
   * A solid tile fill. Set only for the plate lockup, so `plate ? ... : ...` is
   * the whole test a caller needs.
   */
  plate?: string;
  /** What the mark is painted in when it is on a plate. */
  onPlate?: string;
};

/**
 * Which lockup this brand colour gets on this card surface.
 *
 * Exported alongside BrandMark because the card needs the answer for more than
 * the mark: the tile fill, the lit border and the glow are all the brand colour
 * too, and they have to agree with what the logo did.
 */
export const cardArtTreatment = (tint: string | undefined, surface: string): ArtTreatment => {
  if (!tint) return {};
  if (contrastRatio(tint, surface) >= READABLE_ON_CARD) {
    return {usable: tint};
  }
  // On a plate, the mark takes whichever end of the range the plate is furthest
  // from. Computed rather than assumed white: a pale brand colour needs a dark
  // mark, and half the brands that fail the test above fail it for being pale.
  const onPlate = luminance(tint) > 0.35 ? '#111111' : '#ffffff';
  return {usable: tint, plate: tint, onPlate};
};

/**
 * A fetched brand mark, drawn on the 24x24 grid the pipeline guarantees.
 *
 * `lit` fades the mark rather than recolouring it. A dimmed card in a row is
 * saying "not this one, yet", and the honest way to say that about a logo is less
 * of it — repaint it grey and the viewer stops recognising the thing the row is
 * about, which is the same mistake as never fetching the colour in the first
 * place.
 */
export const BrandMark: React.FC<{
  path: string;
  tint?: string;
  /** The colour to use when no brand colour survived the fetch. */
  fallback: string;
  surface: string;
  size: number;
  lit: boolean;
}> = ({path, tint, fallback, surface, size, lit}) => {
  const art = cardArtTreatment(tint, surface);
  const colour = art.plate ? art.onPlate! : (art.usable ?? fallback);
  return (
    <svg viewBox="0 0 24 24" width={size} height={size} aria-hidden>
      <path d={path} fill={lit ? colour : withAlpha(colour, 0.5)} />
    </svg>
  );
};
