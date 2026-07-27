import {useMemo} from 'react';
import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {Stage, STAGE_H, STAGE_W} from './Stage';
import {FIGURE_BOX, figureFor, type FigurePalette} from './artwork';

// IllustrationScene is kinetic typography with a figure beside it.
//
// Unlike the whiteboard and the flow diagram, this scene does *not* accumulate:
// one beat is one shot, and the clip cuts between them. That is the form. The
// other two templates are about a picture being built; this one is about a
// phrase landing, and a phrase cannot land on a stage that is already full of
// the last four.
//
// So the whole design budget goes into the arrival. The headline is set word by
// word rather than as a block, the one word the plan marks gets a highlight
// swept in behind it, and the figure assembles from its parts. After that
// everything holds — except the figure, which keeps a small idle running
// (artwork.tsx) so the shot is never actually still.

const ENTER = {
  /** Frames between one headline word and the next. */
  wordStagger: 3,
  /** How long the figure takes to assemble. */
  figureFrames: 22,
  /** The highlight sweeps only once its word has settled. */
  markDelay: 7,
  markFrames: 11,
  /** The supporting line comes in under a finished headline. */
  captionDelay: 16,
  captionFrames: 12,
} as const;

const GUTTER = 84;
const FIGURE_COL = 0.4;

type Segment = {text: string; mark: boolean};

/**
 * Splits a headline so the emphasised phrase is a single segment.
 *
 * Matching is done on a normalised copy but the slice comes out of the
 * original, so the headline keeps its own capitalisation and punctuation — the
 * model naming the phrase in lower case should not restyle the line.
 */
const splitHeadline = (headline: string, emphasis: string): Segment[] => {
  const phrase = emphasis.trim();
  if (!phrase) {
    return [{text: headline, mark: false}];
  }
  const norm = (s: string) => s.toLowerCase().replace(/[^a-z0-9 ]/g, '');
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
const headlineWords = (headline: string, emphasis: string): Segment[] =>
  splitHeadline(headline, emphasis).flatMap((seg) =>
    seg.text
      .split(/(\s+)/)
      .filter((w) => w.trim().length > 0)
      .map((w) => ({text: w, mark: seg.mark})),
  );

export const IllustrationScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const headline = String(props.headline ?? '');
  const emphasis = String(props.emphasis ?? '');
  const caption = String(props.caption ?? '');
  const flip = Boolean(props.flip);
  const Figure = figureFor(props.figure as string | undefined);

  const words = useMemo(() => headlineWords(headline, emphasis), [headline, emphasis]);

  // Every one of these is a token, and that is the whole of what makes the
  // figures work in both modes: `mass` flips to a mid-tone on paper so a body
  // is still a shape, and `ink` stays darker than it so the shaded faces stay
  // shading rather than becoming highlights.
  const palette: FigurePalette = {
    accent: theme.accent,
    primary: theme.primary,
    ink: theme.ink,
    soft: theme.mass,
    line: theme.line,
  };

  const figureSize = Math.min(STAGE_W * FIGURE_COL * 0.94, STAGE_H * 0.7, 520);

  // Long headlines step down rather than wrapping to three lines, which at this
  // size stops being a headline and starts being a paragraph.
  const chars = headline.length;
  const headlineSize = chars <= 18 ? 104 : chars <= 30 ? 88 : chars <= 44 ? 74 : 62;

  const figureBuild = interpolate(frame, [0, ENTER.figureFrames], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const lastWordFrame = (words.length - 1) * ENTER.wordStagger;
  const captionP = interpolate(
    frame,
    [lastWordFrame + ENTER.captionDelay, lastWordFrame + ENTER.captionDelay + ENTER.captionFrames],
    [0, 1],
    {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
  );

  const figure = (
    <div
      style={{
        position: 'relative',
        width: figureSize,
        height: figureSize,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      {/* A soft pool of the accent under the figure. Without it the artwork
          floats on the background instead of being lit by the scene. */}
      <div
        style={{
          position: 'absolute',
          inset: '-14%',
          borderRadius: '50%',
          background: `radial-gradient(circle, ${theme.accent}22 0%, ${theme.primary}14 42%, transparent 70%)`,
          opacity: figureBuild,
        }}
      />
      <svg
        width={figureSize}
        height={figureSize}
        viewBox={`0 0 ${FIGURE_BOX} ${FIGURE_BOX}`}
        style={{position: 'relative', overflow: 'visible'}}
      >
        <Figure build={figureBuild} t={frame / FPS} palette={palette} />
      </svg>
    </div>
  );

  const type = (
    <div style={{flex: '1 1 auto', minWidth: 0}}>
      <div
        style={{
          fontFamily: theme.fontDisplay,
          fontSize: headlineSize,
          fontWeight: 700,
          lineHeight: 1.1,
          letterSpacing: '-0.02em',
          color: theme.text,
          display: 'flex',
          flexWrap: 'wrap',
          // Row gap has to carry the descenders of a two-line headline; the
          // highlight sits behind the text and would otherwise clip into it.
          columnGap: '0.28em',
          rowGap: '0.14em',
        }}
      >
        {words.map((word, i) => {
          const since = frame - i * ENTER.wordStagger;
          const enter = spring({
            frame: since,
            fps,
            config: {damping: 200, mass: 0.7},
            durationInFrames: 15,
          });
          const markP = word.mark
            ? interpolate(
                since,
                [ENTER.markDelay, ENTER.markDelay + ENTER.markFrames],
                [0, 1],
                {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
              )
            : 0;
          return (
            <span
              key={i}
              style={{
                position: 'relative',
                display: 'inline-block',
                opacity: enter,
                transform: `translateY(${(1 - enter) * 30}px)`,
                // accentText, not accent: the emphasised word IS type, and a
                // brand yellow set on paper is unreadable.
                color: word.mark ? theme.accentText : theme.text,
              }}
            >
              {/* A marker stroke swept in under the word, left to right.
                  A filled highlight box was tried first and cannot work on a
                  dark stage: the accent at an opacity low enough to read text
                  through comes out as a muddy rectangle, and it looks like the
                  word has been redacted rather than emphasised. */}
              {word.mark && (
                <span
                  style={{
                    position: 'absolute',
                    left: '-0.05em',
                    // A two-word emphasis has to read as one stroke. Without
                    // bridging the column gap it comes out as two underlines
                    // with a hole between them, which looks like two separate
                    // emphases rather than one phrase.
                    right: words[i + 1]?.mark ? '-0.34em' : '-0.05em',
                    bottom: '0.06em',
                    height: '0.13em',
                    background: theme.accentText,
                    opacity: 0.85,
                    borderRadius: 999,
                    transform: `scaleX(${markP})`,
                    transformOrigin: 'left center',
                  }}
                />
              )}
              <span style={{position: 'relative'}}>{word.text}</span>
            </span>
          );
        })}
      </div>
      {caption && (
        <div
          style={{
            marginTop: 30,
            fontFamily: theme.fontBody,
            fontSize: 34,
            lineHeight: 1.45,
            fontWeight: 500,
            color: theme.textMuted,
            opacity: captionP,
            transform: `translateY(${(1 - captionP) * 14}px)`,
            maxWidth: '92%',
          }}
        >
          {caption}
        </div>
      )}
    </div>
  );

  return (
    <Stage>
      <div
        style={{
          display: 'flex',
          width: STAGE_W,
          alignItems: 'center',
          gap: GUTTER,
          // Alternating the side beat to beat is what keeps a run of these
          // shots from reading as one slide with the words swapped out.
          flexDirection: flip ? 'row-reverse' : 'row',
        }}
      >
        {/* The figure gets a fixed share of the width and centres inside it,
            so the type column starts at the same x on every shot however big
            the artwork ends up. */}
        <div
          style={{
            flex: `0 0 ${FIGURE_COL * 100}%`,
            display: 'flex',
            justifyContent: 'center',
          }}
        >
          {figure}
        </div>
        {type}
      </div>
    </Stage>
  );
};
