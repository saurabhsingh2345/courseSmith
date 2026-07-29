import {interpolate, useCurrentFrame} from 'remotion';
import {ResolvedTheme} from '../theme/theme';
import {SAFE_X} from './Stage';

// Standing chrome: the furniture that sits on every frame of a skin that
// carries it, outside whatever the scene is drawing.
//
// This is the single component that most decides whether a set of clips reads
// as one production. The reference look derives almost all of its cohesion from
// three marks that never move — an eyebrow naming where in the argument you
// are, a takeaway line restating what the frame just proved, and a watermark —
// repeated identically across every beat of every video. Individually they are
// nearly invisible; together they are the difference between "a series" and
// "some videos by the same person".
//
// They are deliberately *small*. Chrome that competes with the headline stops
// being a frame and becomes content. Everything here is set in the mono face at
// caption size in muted ink, and the eye reads it only when it goes looking.
//
// The scene supplies `eyebrow` and `takeaway` as ordinary props, so a template
// that has nothing to say in those slots simply omits them and the corner stays
// clean. Nothing is invented on the scene's behalf: an eyebrow that read
// "SECTION 3" because the renderer counted scenes would be furniture pretending
// to be an argument.

/** Vertical inset of the chrome from the frame edges. */
const CHROME_TOP = 46;
const CHROME_BOTTOM = 44;

const mono = (theme: ResolvedTheme, size: number, tracking: number) =>
  ({
    fontFamily: theme.fontMono,
    fontSize: size,
    letterSpacing: tracking,
    textTransform: 'uppercase',
    fontWeight: 500,
    whiteSpace: 'nowrap',
  }) as const;

/**
 * The per-scene half: an eyebrow above the frame and a takeaway line below it.
 *
 * Rendered inside the scene's transition so both change with the beat and fade
 * with it, which is what makes the eyebrow read as a caption on *this* moment
 * rather than as a fixture of the video.
 */
export const SceneChrome: React.FC<{
  theme: ResolvedTheme;
  /** Where in the argument this beat sits — "NUMBER 2 / 3 · BANDWIDTH". */
  eyebrow?: string;
  /** What the frame just proved, in one line. */
  takeaway?: string;
}> = ({theme, eyebrow, takeaway}) => {
  const frame = useCurrentFrame();
  const enter = interpolate(frame, [0, 14], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  if (!eyebrow && !takeaway) return null;
  return (
    <>
      {eyebrow ? (
        <div
          style={{
            position: 'absolute',
            top: CHROME_TOP,
            left: 0,
            right: 0,
            textAlign: 'center',
            color: theme.textMuted,
            opacity: enter * 0.8,
            ...mono(theme, 17, 5.5),
          }}
        >
          {eyebrow}
        </div>
      ) : null}
      {takeaway ? (
        <div
          style={{
            position: 'absolute',
            bottom: CHROME_BOTTOM,
            left: SAFE_X,
            maxWidth: 1200,
            textAlign: 'left',
            color: theme.textMuted,
            // Dimmer than the eyebrow: this line is a footnote on the frame,
            // and the narration is already saying it out loud.
            opacity: enter * 0.55,
            ...mono(theme, 14, 3),
          }}
        >
          {takeaway}
        </div>
      ) : null}
    </>
  );
};

/**
 * The standing half: a mark in the corner that never changes.
 *
 * Rendered once at the top of the video rather than per scene, so it does not
 * cross-dissolve at every cut. A watermark that fades in and out with the beats
 * is a watermark the eye keeps noticing, which is the opposite of the job.
 */
export const Watermark: React.FC<{theme: ResolvedTheme}> = ({theme}) => {
  if (!theme.watermark) return null;
  return (
    <div
      style={{
        position: 'absolute',
        bottom: CHROME_BOTTOM,
        right: SAFE_X,
        color: theme.textMuted,
        opacity: 0.38,
        ...mono(theme, 14, 2),
      }}
    >
      {theme.watermark}
    </div>
  );
};
