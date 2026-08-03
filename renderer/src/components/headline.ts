// Splitting a headline around the phrase it emphasises.
//
// The reference look puts exactly one coloured span in every headline — "A
// $4000 BOX BUILT FOR AI", "64 ACCELERATORS MINIMUM" — and that one span is
// most of why the type reads as designed rather than as set. It is also the
// cheapest possible effect: no layout, no motion, one colour on a slice of a
// string.
//
// This lived inside IllustrationScene, which needed it for a marker stroke
// drawn under part of its heading. It is here now because the shared header
// wants the same split for a different paint, and two implementations of "which
// characters did they mean" would eventually disagree about a headline with an
// em-dash in it.
//
// The phrase is matched, never described. Go validates that the emphasis is a
// literal substring of the heading (see containsPhrase in snippet_illustration.go)
// so by the time a scene graph reaches the renderer the match is guaranteed to
// exist — the fallbacks below are for hand-written props and old scene graphs,
// not for the normal path.

export type Segment = {text: string; mark: boolean};

/** Lower-case alphanumerics and spaces. Mirrors phraseKey in Go. */
const norm = (s: string) => s.toLowerCase().replace(/[^a-z0-9 ]/g, '');

/**
 * Splits a headline so the emphasised phrase is a single segment.
 *
 * Matching is done on a normalised copy but the slice comes out of the
 * original, so the headline keeps its own capitalisation and punctuation — the
 * model naming the phrase in lower case should not restyle the line.
 */
export const splitHeadline = (headline: string, emphasis: string): Segment[] => {
  const phrase = emphasis.trim();
  if (!phrase) {
    return [{text: headline, mark: false}];
  }
  const at = norm(headline).indexOf(norm(phrase));
  if (at < 0 || norm(phrase) === '') {
    return [{text: headline, mark: false}];
  }
  // norm() only ever drops characters, so walking the original in step with the
  // normalised index recovers the true offsets.
  let seen = 0;
  let start = -1;
  let end = -1;
  for (let i = 0; i < headline.length; i++) {
    if (seen === at && start < 0) start = i;
    if (seen === at + norm(phrase).length && end < 0) end = i;
    if (norm(headline[i]).length > 0) seen++;
  }
  if (start < 0) return [{text: headline, mark: false}];
  if (end < 0) end = headline.length;
  return [
    {text: headline.slice(0, start), mark: false},
    {text: headline.slice(start, end), mark: true},
    {text: headline.slice(end), mark: false},
  ].filter((s) => s.text.length > 0);
};

/** Segments flattened to words, each keeping whether it is emphasised. */
export const headlineWords = (headline: string, emphasis: string): Segment[] =>
  splitHeadline(headline, emphasis).flatMap((seg) =>
    seg.text
      .split(/(\s+)/)
      .filter((w) => w.trim().length > 0)
      .map((w) => ({text: w, mark: seg.mark})),
  );
