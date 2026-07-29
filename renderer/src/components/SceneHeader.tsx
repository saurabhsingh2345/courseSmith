import {interpolate, useCurrentFrame} from 'remotion';
import {ResolvedTheme} from '../theme/theme';

// SceneHeader is the single title treatment for the whole lesson.
//
// Scenes used to each invent their own: points slides had a kicker + 68px
// display title + accent rule, diagram scenes a bare 48px centred line, the
// walkthrough a 38px label shoved against the top-left corner. Cutting between
// them read as three different videos. One component, two sizes: `display` for
// scenes whose title *is* the composition, `compact` for scenes where a window
// or diagram is the subject and the title is a label above it.

export const SceneHeader: React.FC<{
  theme: ResolvedTheme;
  title: string;
  /** Eyebrow above the title. Omitted entirely when blank. */
  kicker?: string;
  size?: 'display' | 'compact';
  /** Space below the header, so callers control their own rhythm. */
  marginBottom?: number;
}> = ({theme, title, kicker, size = 'display', marginBottom}) => {
  const frame = useCurrentFrame();
  const enter = interpolate(frame, [2, 16], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  if (!title) {
    return null;
  }
  const display = size === 'display';

  // The broadcast skin sets the title very large, uppercase and unruled, and
  // hands the kicker to the standing chrome instead of drawing its own — the
  // eyebrow is pinned to the top of the *frame* there, not stacked on the
  // headline, and rendering both would put the same words on screen twice.
  //
  // Every existing template already passes a kicker, so they all inherit the
  // reference treatment here without being touched. That is the point of doing
  // this in the shared header rather than in each new template.
  if (theme.skin === 'broadcast') {
    return (
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          flexShrink: 0,
          marginBottom: marginBottom ?? (display ? 64 : 40),
          opacity: enter,
          transform: `translateY(${(1 - enter) * 12}px)`,
        }}
      >
        <div
          style={{
            fontFamily: theme.fontDisplay,
            // Large enough that a five-word headline is the loudest thing in
            // the frame by a wide margin, which is what lets the diagram below
            // it be small and still read as the subject.
            fontSize: display ? 84 : 52,
            fontWeight: 800,
            letterSpacing: display ? -1.5 : -0.8,
            lineHeight: 1.06,
            textTransform: 'uppercase',
            color: theme.text,
            textAlign: 'center',
            maxWidth: 1500,
          }}
        >
          {title}
        </div>
      </div>
    );
  }

  // The minimal skin keeps sentence case and drops the furniture entirely: no
  // kicker, no rule. The diagram is the frame and the title is a label on it.
  if (theme.skin === 'minimal') {
    return (
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          flexShrink: 0,
          marginBottom: marginBottom ?? (display ? 48 : 28),
          opacity: enter,
          transform: `translateY(${(1 - enter) * 12}px)`,
        }}
      >
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: display ? 50 : 34,
            fontWeight: 600,
            letterSpacing: -0.4,
            lineHeight: 1.2,
            color: theme.text,
            textAlign: 'center',
            maxWidth: 1400,
          }}
        >
          {title}
        </div>
      </div>
    );
  }

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        flexShrink: 0,
        marginBottom: marginBottom ?? (display ? 56 : 30),
        opacity: enter,
        transform: `translateY(${(1 - enter) * 16}px)`,
      }}
    >
      {kicker ? (
        <div
          style={{
            fontFamily: theme.fontBody,
            fontSize: display ? 22 : 17,
            letterSpacing: display ? 8 : 6,
            textTransform: 'uppercase',
            color: theme.accent,
            fontWeight: 600,
            marginBottom: display ? 16 : 11,
          }}
        >
          {kicker}
        </div>
      ) : null}
      <div
        style={{
          fontFamily: theme.fontDisplay,
          fontSize: display ? 64 : 40,
          fontWeight: 700,
          letterSpacing: display ? -1 : -0.5,
          lineHeight: 1.14,
          color: theme.text,
          textAlign: 'center',
          maxWidth: 1400,
        }}
      >
        {title}
      </div>
      <div
        style={{
          marginTop: display ? 22 : 15,
          width: display ? 116 : 72,
          height: display ? 6 : 4,
          borderRadius: 3,
          background: `linear-gradient(90deg, ${theme.accent}, ${theme.primary})`,
        }}
      />
    </div>
  );
};
