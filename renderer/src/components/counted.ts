// Numbers that arrive by counting.
//
// A figure that snaps to its final value is a figure the eye reads before the
// voice has said it, so the reveal lands on nothing. A figure that counts up
// arrives with the sentence. `metric` established the rule and `gauge` needs the
// same one, so it lives here rather than in either of them — a formatting rule
// copied into two files is a formatting rule that will disagree with itself.
//
// The one hard constraint, and the reason this is not just `value * t`: the
// decimal places of the TARGET are preserved the whole way up. Without that
// "2.8" flickers between "3" and "2.80" on its way to itself, which reads as a
// glitch rather than as a count.

/**
 * The figure part-way through its count-up, as text.
 *
 * `countsUp` false — or a value that is not a plain number — returns the value
 * whole and unanimated. That is deliberate and not a fallback: animating a range
 * like "313K–577K" toward a single intermediate would have the clip state a
 * number that is false for a third of a second, and a wrong number on screen is
 * worse than a static one. For `metric` the Go side decides (Metric.countsUp);
 * for `gauge` the value is typed as a number, so it always counts.
 */
export const counted = (value: string, countsUp: boolean, t: number): string => {
  if (!countsUp || t >= 1) return value;
  const target = Number(value);
  if (!Number.isFinite(target)) return value;
  const dot = value.indexOf('.');
  const places = dot < 0 ? 0 : value.length - dot - 1;
  return (target * t).toFixed(places);
};
