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

export const Stage: React.FC<{
  /** Vertical placement of the content block within the stage box. */
  justify?: React.CSSProperties['justifyContent'];
  /** Horizontal placement. Defaults to centred. */
  align?: React.CSSProperties['alignItems'];
  children: React.ReactNode;
}> = ({justify = 'center', align = 'center', children}) => (
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
    {children}
  </AbsoluteFill>
);
