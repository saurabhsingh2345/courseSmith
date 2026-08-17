import {ResolvedTheme, seat, withAlpha} from '../theme/theme';

// A macOS application window, on paper.
//
// Nine components in this catalog draw traffic lights, and every one of them
// draws them from its own literals — three hex codes, a size, a gap, repeated. It
// is the exact drift showroomCard.tsx was extracted to stop, one level up: the
// terminal, the editor, the browser and the shell are all the same object with
// different contents, and when two of them appear in one course the mismatch is
// visible without being nameable.
//
// This is that object, for the templates that came after it. The nine existing
// ones are deliberately NOT refactored onto it — they have pixel-exact baselines
// and the win would be tidiness bought with a day of re-recording — but nothing
// new should draw its own lights again.
//
// THE WINDOW IS DARK AND THE STAGE IS LIGHT, and that pairing is the point rather
// than an accident of the reference. A dark panel on paper reads as a real
// application sitting on a real desk: the shadow does the seating, the dark field
// makes the code the brightest thing inside its own frame, and the paper around it
// keeps the composition from being a black rectangle on a black stage. It is the
// one place the showroom skin's cards are not white.

/** The window's own palette. Dark regardless of the stage, and see above for why. */
const WINDOW = {
  body: '#15161b',
  bar: '#1f2027',
  border: '#2b2d36',
  title: '#9ba1ad',
  faint: '#6c7280',
};

export const TRAFFIC = ['#ff5f57', '#febc2e', '#28c840'] as const;

/**
 * The title bar's height, exported because contents have to size against it: a
 * window is a fixed box and the body gets whatever the bar leaves.
 */
export const WINDOW_BAR_H = 56;

export const windowPalette = () => WINDOW;

export const AppWindow: React.FC<{
  theme: ResolvedTheme;
  /** Centred in the title bar, in mono — a path, a session name, a host. */
  title: string;
  /**
   * A small glyph before the title, as the real thing puts a folder there. One
   * character; anything longer competes with the title.
   */
  badge?: string;
  width: number;
  height: number;
  /**
   * 0..1. A hairline rail along the bar's lower edge, which is how the tool this
   * borrows from shows a long task advancing. Omitted, no rail is drawn — an
   * empty rail on a window that is not working reads as a stalled one.
   */
  progress?: number;
  style?: React.CSSProperties;
  children: React.ReactNode;
}> = ({theme, title, badge, width, height, progress, style, children}) => (
  <div
    style={{
      width,
      height,
      flex: 'none',
      borderRadius: 14,
      overflow: 'hidden',
      background: WINDOW.body,
      border: `1px solid ${WINDOW.border}`,
      // The same seating every other object in this skin gets. A window that
      // composed its own shadow would sit at a different height on the page than
      // the cards beside it, which is the one thing a viewer notices instantly
      // and cannot name.
      boxShadow: seat(theme, 'lifted'),
      display: 'flex',
      flexDirection: 'column',
      ...style,
    }}
  >
    <div
      style={{
        height: WINDOW_BAR_H,
        flex: 'none',
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        padding: '0 20px',
        background: WINDOW.bar,
        borderBottom: `1px solid ${WINDOW.border}`,
        position: 'relative',
      }}
    >
      {TRAFFIC.map((c) => (
        <div key={c} style={{width: 13, height: 13, borderRadius: 999, background: c}} />
      ))}
      <div
        style={{
          position: 'absolute',
          left: 0,
          right: 0,
          textAlign: 'center',
          fontFamily: theme.fontMono,
          fontSize: 21,
          color: WINDOW.title,
          pointerEvents: 'none',
        }}
      >
        {badge ? <span style={{marginRight: 10, color: WINDOW.faint}}>{badge}</span> : null}
        {title}
      </div>
      {progress !== undefined ? (
        <div
          style={{
            position: 'absolute',
            left: 0,
            bottom: -1,
            height: 3,
            width: `${Math.max(0, Math.min(1, progress)) * 100}%`,
            background: theme.accent,
          }}
        />
      ) : null}
    </div>
    <div style={{flex: 1, minHeight: 0, position: 'relative'}}>{children}</div>
  </div>
);

/** Mono body text inside a window, at the one size these templates use. */
export const windowText = (theme: ResolvedTheme, size: number): React.CSSProperties => ({
  fontFamily: theme.fontMono,
  fontSize: size,
  lineHeight: 1.5,
  color: '#d7dae1',
  whiteSpace: 'pre',
});

/** A dim line — context in a diff, an unreached step. */
export const windowDim = (): string => withAlpha('#d7dae1', 0.42);
