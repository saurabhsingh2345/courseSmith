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
 * Bottom reserve when a caption card is going to be drawn there.
 *
 * Stays at 120 — the value every scene's layout was tuned against — rather than
 * the 200 a caption line actually wants (64 margin + 40 padding + 42px/1.35 text,
 * plus pop headroom). Raising it looks like a fix and is a regression: a dozen
 * scenes capture STAGE_H at module scope, so a bigger reserve shrinks every one
 * of them, in the captions-OFF default where there is no card to make room for.
 * The under-reservation is a known, pre-existing compromise for the rare
 * captions-on video; making it worse for the common case to fix it is the wrong
 * trade, and correcting it properly means giving those scenes a runtime height.
 */
export const CAPTION_SAFE = 120;

/**
 * Bottom reserve when nothing is going to be drawn there — the frame margin, and
 * nothing more.
 *
 * Captions have been off by default since 2026-07-26, and the reserve stayed
 * unconditional: every frame in every default-configured video gave up a band at
 * the bottom to a caption card that was never rendered. On a 1080-line frame that
 * is 11% of the height held empty by a constant, which is a real part of why the
 * output reads as content floating high in an under-filled frame.
 *
 * Kept as a named constant rather than folded into SAFE_TOP because the two are
 * different ideas that happen to share a number, and a later change to the frame
 * margin should not silently change what a caption needs.
 */
export const NO_CAPTION_SAFE = 64;

/** Horizontal frame margin — content never runs to the edge. */
export const SAFE_X = 110;

/** Top frame margin. */
export const SAFE_TOP = 64;

/** Usable drawing box. Scenes should size against these, not 1920x1080. */
export const STAGE_W = FRAME_W - SAFE_X * 2;

/**
 * Usable height — unchanged, and deliberately not made conditional.
 *
 * Same reason the skin's `air` is a scale rather than padding: about a dozen
 * scenes capture this at module scope (`const BODY_H = STAGE_H - HEADER_H - 150`),
 * so it cannot vary per render without giving all of them a runtime height, which
 * would put a layout regression in every one. Scenes keep sizing against exactly
 * the height they were tuned against.
 *
 * What the conditional reserve below changes is the *padding*, and therefore where
 * the content block sits. That is the part worth having: a short composition now
 * centres in the whole frame instead of centring in a box with a permanent hole at
 * the bottom, which is what left everything riding high.
 */
export const STAGE_H = FRAME_H - SAFE_TOP - CAPTION_SAFE;

/**
 * Whether a caption card will be drawn over this frame, supplied once by
 * LessonVideo.
 *
 * A context for exactly the reason StageAirContext is one: Stage is composed into
 * by around thirty scene components, and a prop would work in twenty-nine of them.
 * Defaults to false — no captions — so a Stage rendered in a fixture or a test
 * gets the full frame.
 */
export const StageCaptionsContext = createContext(false);

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

/**
 * The skin's horizontal placement for the content block, supplied once by
 * LessonVideo.
 *
 * A context for the same reason air is one: fifty scene components render their
 * own `<Stage>` and not one of them passes an `align`, so a prop would work in
 * forty-nine of them and silently not in the fiftieth. Threading it through
 * every scene is also the thing a skin is supposed to avoid — the placement is a
 * house-style decision, not a property of any template's picture.
 *
 * Defaults to 'center', which is what every scene has always got, so a Stage
 * rendered outside a provider — a fixture, a Storybook entry, a baseline —
 * composes exactly as it did before this existed.
 */
export const StageAlignContext = createContext<React.CSSProperties['alignItems']>('center');

/**
 * The skin's VERTICAL placement, supplied once by LessonVideo.
 *
 * The companion to StageAlignContext, and it exists because fixing only the
 * horizontal axis made the vertical one worse. Thirty-seven scenes pass
 * `justify="center"`, so a short composition centres inside the drawing box and
 * leaves the bottom third of the frame empty — a problem this file already
 * describes ("content floating high in an under-filled frame"). Centred, that
 * emptiness is symmetric and reads as margin. Left-weighted, it reads as a
 * frame that did not get finished.
 *
 * Defaults to undefined, meaning "whatever the scene asked for" — so every skin
 * but `editorial` composes exactly as it always has.
 */
export const StageJustifyContext = createContext<React.CSSProperties['justifyContent'] | undefined>(undefined);

export const Stage: React.FC<{
  /**
   * Vertical placement of the content block within the stage box. A skin may
   * override it through StageJustifyContext; scenes keep passing what they
   * always passed.
   */
  justify?: React.CSSProperties['justifyContent'];
  /**
   * Horizontal placement. Omitted, it comes from StageAlignContext — the skin's
   * choice — which is 'center' for every skin but `editorial`. A scene passes it
   * explicitly only to opt out, the same way it does with `air`.
   */
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
}> = ({justify = 'center', align, air, children}) => {
  const contextAir = useContext(StageAirContext);
  const contextAlign = useContext(StageAlignContext);
  const contextJustify = useContext(StageJustifyContext);
  const placement = align ?? contextAlign;
  // Skin wins over the scene here, unlike `align`: the scenes all say the same
  // thing (`center`), so honouring them would mean the skin never applied.
  const flow = contextJustify ?? justify;
  const hasCaptions = useContext(StageCaptionsContext);
  // The reserve the *padding* uses. STAGE_H stays pessimistic (see its comment),
  // so a scene sized against it still fits; what this buys is that the content
  // block centres in the whole frame rather than in a box with a permanent
  // 120-pixel hole at the bottom.
  const bottomReserve = hasCaptions ? CAPTION_SAFE : NO_CAPTION_SAFE;
  // Clamped at zero on the low end, deliberately. A negative inset — scaling the
  // composition UP to fill more of the frame — was tried for the editorial skin
  // and reverted: the transform is about the centre, so anything above 1.0 grows
  // the content through SAFE_X and SAFE_TOP, and the frame margin this file
  // exists to guarantee stops being guaranteed. If filling the height is wanted,
  // it has to come from the scenes' own layout, not from here.
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
        paddingBottom: bottomReserve,
        display: 'flex',
        flexDirection: 'column',
        alignItems: placement,
        justifyContent: flow,
      }}
    >
      {inset === 0 ? (
        children
      ) : (
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: placement,
            justifyContent: flow,
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
