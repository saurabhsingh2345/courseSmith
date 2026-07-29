import {createContext, useContext} from 'react';
import {AbsoluteFill} from 'remotion';

// Stage is the one safe drawing area every scene composes into.
//
// CaptionTrack is an AbsoluteFill sibling pinned to the bottom of the frame,
// so it sits *on top of* whatever a scene painted there. Scenes used to centre
// their content in the full 1080 and collided with it — a five-item points
// slide bottoms out around y=933 while the caption card's top edge is at ~919.
// Reserving the band here, once, makes that collision impossible rather than
// something each scene has to remember.
//
// The same reservation doubles as the frame-edge margin, so "fill the stage"
// has a single definition every scene can size against.

export const FRAME_W = 1920;
export const FRAME_H = 1080;

/**
 * Bottom reserve. With on-screen captions off (the default since 2026-07-26)
 * this is plain compositional breathing room. If a course turns karaoke
 * captions back on (style.captions: on), restore 200 — one caption line is
 * 161px tall (64 margin + 40 padding + 42px/1.35 text) plus pop headroom —
 * or tall scenes will run under the caption card again.
 */
export const CAPTION_SAFE = 120;

/** Horizontal frame margin — content never runs to the edge. */
export const SAFE_X = 110;

/** Top frame margin. */
export const SAFE_TOP = 64;

/** Usable drawing box. Scenes should size against these, not 1920x1080. */
export const STAGE_W = FRAME_W - SAFE_X * 2;
export const STAGE_H = FRAME_H - SAFE_TOP - CAPTION_SAFE;

/**
 * The skin's stage inset, supplied once by LessonVideo.
 *
 * A context rather than a prop because roughly thirty scene components compose
 * into Stage, and threading a value through every one of them would mean the
 * skin worked in twenty-nine of them and silently did not in the thirtieth.
 * The default is 0, so a Stage rendered outside a provider — a fixture, a
 * Storybook entry, a test — behaves exactly as it always did.
 */
export const StageAirContext = createContext(0);

export const Stage: React.FC<{
  /** Vertical placement of the content block within the stage box. */
  justify?: React.CSSProperties['justifyContent'];
  /** Horizontal placement. Defaults to centred. */
  align?: React.CSSProperties['alignItems'];
  /**
   * Extra inset as a fraction of the drawing box, from the theme's skin.
   *
   * Zero — the default, and what every existing scene gets — fills the stage,
   * which is the right call for a diagram that has to stay legible on a phone.
   * The broadcast skin instead leaves a seventh of the box empty on every side
   * so one small precise picture sits in a lot of nothing, which is most of why
   * the reference look reads as composed rather than as packed.
   *
   * Applied here rather than by each scene so a template cannot forget it, and
   * so the two compositions stay one definition of "the stage" apart.
   *
   * Omitted, it comes from StageAirContext. A scene passes it explicitly only
   * to opt out — a full-bleed screen recording, say, where insetting the frame
   * would letterbox the thing being recorded.
   */
  air?: number;
  children: React.ReactNode;
}> = ({justify = 'center', align = 'center', air, children}) => {
  const contextAir = useContext(StageAirContext);
  const inset = Math.max(0, Math.min(0.3, air ?? contextAir));
  // Air is a *scale*, not extra padding, and that is a deliberate correction.
  //
  // Padding is the obvious implementation and it does nothing: almost every
  // scene sizes its content against the STAGE_W constant at module scope
  // (`const CARD_W = Math.min(STAGE_W, 1700)`), so a fatter padding leaves a
  // fixed-width card exactly as wide as it was and merely overflows the box.
  // Making it work that way would have meant editing all two dozen scenes to
  // size against a runtime width — which is the opposite of a skin being
  // additive, and would have put a layout regression in every one of them.
  //
  // Scaling the content instead reaches fixed widths and fluid ones alike, and
  // it keeps each scene's internal proportions intact: the composition it was
  // designed with, set slightly further back on the stage.
  const scale = 1 - inset * 2;
  return (
    <AbsoluteFill
      style={{
        paddingTop: SAFE_TOP,
        paddingLeft: SAFE_X,
        paddingRight: SAFE_X,
        paddingBottom: CAPTION_SAFE,
        display: 'flex',
        flexDirection: 'column',
        alignItems: align,
        justifyContent: justify,
      }}
    >
      {inset === 0 ? (
        children
      ) : (
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: align,
            justifyContent: justify,
            width: '100%',
            height: '100%',
            transform: `scale(${scale})`,
            // Scale about the centre, so the air is shared between the top and
            // the bottom rather than all falling below the content.
            transformOrigin: 'center center',
          }}
        >
          {children}
        </div>
      )}
    </AbsoluteFill>
  );
};
