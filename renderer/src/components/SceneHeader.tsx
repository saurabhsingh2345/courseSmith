import {interpolate, useCurrentFrame} from 'remotion';
import {ResolvedTheme} from '../theme/theme';
import {splitHeadline} from './headline';

// SceneHeader is the single title treatment for the whole lesson.
//
// Scenes used to each invent their own: points slides had a kicker + 68px
// display title + accent rule, diagram scenes a bare 48px centred line, the
// walkthrough a 38px label shoved against the top-left corner. Cutting between
// them read as three different videos. One component, two sizes: `display` for
// scenes whose title *is* the composition, `compact` for scenes where a window
// or diagram is the subject and the title is a label above it.

/** Which semantic accent an emphasised phrase is painted in. */
export type EmphasisRole = 'quantity' | 'limit' | 'rival';

const emphasisColour = (theme: ResolvedTheme, role?: string): string => {
  switch (role) {
    case 'limit':
      return theme.accentLimit;
    case 'rival':
      return theme.accentRival;
    case 'quantity':
      return theme.accentQuantity;
    default:
      // No role stated is not a reason to paint nothing: the whole point of the
      // treatment is that one phrase is louder than the rest. The brand accent
      // is the honest default because it makes no claim about meaning, which is
      // exactly what an unroled emphasis is.
      return theme.accentText;
  }
};

/**
 * The headline, with at most one phrase in a semantic accent.
 *
 * Every headline in the reference look has exactly one coloured span, and it is
 * most of why the type reads as designed. The colour is chosen by what the
 * phrase is *doing* — the number under discussion, the ceiling it hits, the
 * alternative — never by taste, for the same reason the metric template colours
 * its figures by role. See videoskin.go.
 */
const Headline: React.FC<{
  theme: ResolvedTheme;
  title: string;
  emphasis?: string;
  emphasisRole?: string;
  style: React.CSSProperties;
}> = ({theme, title, emphasis, emphasisRole, style}) => {
  const segments = splitHeadline(title, emphasis ?? '');
  if (segments.length < 2) {
    return <div style={style}>{title}</div>;
  }
  const colour = emphasisColour(theme, emphasisRole);
  return (
    <div style={style}>
      {segments.map((seg, i) => (
        <span key={i} style={seg.mark ? {color: colour} : undefined}>
          {seg.text}
        </span>
      ))}
    </div>
  );
};

export const SceneHeader: React.FC<{
  theme: ResolvedTheme;
  title: string;
  /** Eyebrow above the title. Omitted entirely when blank. */
  kicker?: string;
  /**
   * A literal phrase of `title` to set in a semantic accent. Go guarantees it
   * occurs in the title (containsPhrase); anything that does not match leaves
   * the headline in one colour rather than failing.
   */
  emphasis?: string;
  /** What the emphasised phrase is doing: quantity | limit | rival. */
  emphasisRole?: string;
  size?: 'display' | 'compact';
  /** Space below the header, so callers control their own rhythm. */
  marginBottom?: number;
}> = ({theme, title, kicker, emphasis, emphasisRole, size = 'display', marginBottom}) => {
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
        <Headline
          theme={theme}
          title={title}
          emphasis={emphasis}
          emphasisRole={emphasisRole}
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
        />
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
        <Headline
          theme={theme}
          title={title}
          emphasis={emphasis}
          emphasisRole={emphasisRole}
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
        />
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
      <Headline
        theme={theme}
        title={title}
        emphasis={emphasis}
        emphasisRole={emphasisRole}
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
      />
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
