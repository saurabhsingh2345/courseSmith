import {useMemo} from 'react';
import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {Stage, STAGE_H, STAGE_W} from './Stage';
import {FIGURE_BOX, figureFor, type FigurePalette} from './artwork';
import {headlineWords, type Segment} from './headline';

// SpineScene is the backbone: one narration, staged.
//
// Every other template in the catalog answers a question about content — how
// big is this number, which of these two should I pick, what does the code do.
// This one answers a question about a COURSE: the narration already exists, or
// is about to, and it needs a picture that matches it shot by shot. That is the
// intro, the hand-off between chapters, the aside, the outro — the connective
// material that is most of a course's running time and that no content-shaped
// template fits, because its subject is whatever the sentence happens to be.
//
// So the beat does not carry a chart or a diagram; it carries a SHOT. Twelve
// layouts, each one a way of arranging figures from the shared vocabulary
// around a line of type, and the planner picks per beat from what the narration
// is doing at that moment. A clip is a run of different shots rather than one
// picture being built, which is what makes it able to open, explain, turn, and
// close inside a single template.
//
// == The rail ==
//
// One device runs through all twelve: a segmented accent rail down the left
// edge, one segment per beat, filled to where the clip has got to. It is the
// reason the shots read as one piece rather than as twelve unrelated slides — a
// cut from an orbit diagram to a pull-quote is a hard cut, and without something
// that survives it the clip looks assembled from parts.
//
// It is segmented rather than continuous on purpose. A smooth bar says "43%
// through", which is a progress indicator; a row of ticks with some behind you,
// one lit and some ahead says "this is the fourth of nine things", which is what
// a spine actually is. It also does the job an intro normally spends a whole
// shot on: you can see the shape of the clip without being told it.
//
// == Filling the frame ==
//
// Every shot lays out into the full stage height rather than centring whatever
// it happens to be tall. The first version centred a short block and left the
// bottom third of every frame empty, which reads as a slide that was not
// finished — and it was invisible in code review because each shot looked
// correct on its own. BODY_H and the `minHeight` on the tile rows are what stop
// that: the composition is asked for a height, so it grows into one.

/** The shot vocabulary. Go mirrors this list (spineShots). */
export type SpineShot =
  | 'open'
  | 'chapter'
  | 'state'
  | 'pair'
  | 'row'
  | 'orbit'
  | 'steps'
  | 'recap'
  | 'aside'
  | 'focus'
  | 'quote'
  | 'close';

type SpineObject = {figure?: string; label?: string; detail?: string};

const ENTER = {
  /** Frames between one headline word and the next. */
  wordStagger: 3,
  /** How long a figure takes to assemble. */
  figure: 22,
  /** Frames between one object tile and the next. */
  tileStagger: 6,
  /** The marker stroke sweeps once its word has settled. */
  markDelay: 7,
  markFrames: 13,
  /** Supporting copy arrives under a finished headline. */
  captionDelay: 14,
  captionFrames: 12,
  /** The eyebrow leads everything — it is the smallest thing on screen. */
  noteFrames: 10,
  /** The last frames of a shot, over which it releases before the cut. */
  exit: 9,
} as const;

const RAIL_W = 4;
const RAIL_GUTTER = 52;

/** The height every shot composes into. See the header note on filling. */
const BODY_H = STAGE_H;

const clamp01 = (v: number) => Math.max(0, Math.min(1, v));

/** Linear 0→1 across [from, from+len] frames, clamped both ends. */
const ramp = (frame: number, from: number, len: number) =>
  interpolate(frame, [from, from + len], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

/**
 * The same window, eased out.
 *
 * Used for everything that MOVES — a tile travelling, a divider drawing, a
 * spoke leaving the centre. A linear ramp is fine for an opacity and wrong for
 * a position: it arrives at full speed and stops dead, which is the difference
 * between something landing and something being teleported one frame at a time.
 */
const eased = (frame: number, from: number, len: number) =>
  1 - Math.pow(1 - ramp(frame, from, len), 3);

const paletteFor = (theme: ResolvedTheme): FigurePalette => ({
  accent: theme.accent,
  primary: theme.primary,
  ink: theme.ink,
  soft: theme.mass,
  line: theme.line,
});

/**
 * A figure at a given size, lit by a pool of the accent.
 *
 * The pool is not decoration. A flat-vector figure dropped straight onto the
 * stage reads as clipart pasted over a background; the radial wash under it is
 * what makes it look lit by the same scene the type is in, and it costs one
 * div.
 */
const Art: React.FC<{
  theme: ResolvedTheme;
  figure?: string;
  size: number;
  build: number;
  t: number;
  /** Pool strength. Dropped to 0 for tiles, which have their own panel. */
  pool?: number;
}> = ({theme, figure, size, build, t, pool = 1}) => {
  const Figure = figureFor(figure);
  return (
    <div style={{position: 'relative', width: size, height: size}}>
      {pool > 0 && (
        <div
          style={{
            position: 'absolute',
            inset: '-16%',
            borderRadius: '50%',
            background: `radial-gradient(circle, ${theme.accent}${pool > 0.6 ? '22' : '14'} 0%, ${theme.primary}12 44%, transparent 70%)`,
            opacity: build,
          }}
        />
      )}
      <svg
        width={size}
        height={size}
        viewBox={`0 0 ${FIGURE_BOX} ${FIGURE_BOX}`}
        style={{position: 'relative', overflow: 'visible'}}
      >
        <Figure build={build} t={t} palette={paletteFor(theme)} />
      </svg>
    </div>
  );
};

/** Consecutive words sharing an emphasis state. See Headline. */
type Run = {mark: boolean; from: number; words: Segment[]};

const runsOf = (words: Segment[]): Run[] => {
  const out: Run[] = [];
  words.forEach((w, i) => {
    const last = out[out.length - 1];
    if (last && last.mark === w.mark) last.words.push(w);
    else out.push({mark: w.mark, from: i, words: [w]});
  });
  return out;
};

/**
 * The headline, set word by word with the emphasised phrase struck through.
 *
 * Shared across all twelve shots rather than per-layout: the type IS the
 * constant — what changes between shots is what surrounds it — and a second
 * copy of the word-stagger arithmetic is how two shots end up landing their
 * words at different speeds for no reason anybody chose.
 *
 * == Why the stroke is a background and not a bar ==
 *
 * It used to be an absolutely-positioned bar under each word, stretched to
 * `-0.34em` when the next word was also emphasised so a two-word phrase read as
 * one stroke. That bridge cannot know about a line break, and on a headline
 * that wrapped between the two words it drew a rule out into the empty right
 * margin of the first line. Painting it as a background on a run of words
 * instead makes the break the browser's problem: `box-decoration-break: clone`
 * gives each line fragment its own stroke, ending where the fragment ends.
 *
 * That also means the words themselves must sit in normal inline flow rather
 * than being flex items, which is why this is a block with real spaces in it.
 */
const Headline: React.FC<{
  theme: ResolvedTheme;
  text: string;
  emphasis: string;
  size: number;
  frame: number;
  fps: number;
  align?: 'left' | 'center';
  /** Frames to wait before the first word. */
  delay?: number;
}> = ({theme, text, emphasis, size, frame, fps, align = 'left', delay = 0}) => {
  const words = useMemo(() => headlineWords(text, emphasis), [text, emphasis]);
  const runs = useMemo(() => runsOf(words), [words]);

  const word = (w: Segment, i: number) => {
    const since = frame - delay - i * ENTER.wordStagger;
    const enter = spring({
      frame: since,
      fps,
      config: {damping: 200, mass: 0.7},
      durationInFrames: 15,
    });
    return (
      <span
        key={i}
        style={{
          display: 'inline-block',
          opacity: enter,
          transform: `translateY(${(1 - enter) * 28}px)`,
          // accentText rather than accent: an emphasised word is type, and the
          // brand accent set as type on paper is unreadable.
          color: w.mark ? theme.accentText : theme.text,
        }}
      >
        {w.text}
      </span>
    );
  };

  return (
    <div
      style={{
        fontFamily: theme.fontDisplay,
        fontSize: size,
        fontWeight: 700,
        lineHeight: 1.1,
        letterSpacing: '-0.022em',
        color: theme.text,
        textAlign: align,
      }}
    >
      {runs.map((run, r) => {
        const body = run.words.map((w, j) => (
          <span key={j}>
            {word(w, run.from + j)}
            {j < run.words.length - 1 ? ' ' : null}
          </span>
        ));
        if (!run.mark) {
          return (
            <span key={r}>
              {body}
              {r < runs.length - 1 ? ' ' : null}
            </span>
          );
        }
        // The stroke sweeps once the run's LAST word has settled, so a two-word
        // phrase is underlined as one gesture rather than in two halves.
        const settled = delay + (run.from + run.words.length - 1) * ENTER.wordStagger;
        const sweep = eased(frame, settled + ENTER.markDelay, ENTER.markFrames);
        return (
          <span key={r}>
            <span
              style={{
                backgroundImage: `linear-gradient(to top, ${theme.accentText} 0.1em, transparent 0.1em)`,
                backgroundRepeat: 'no-repeat',
                backgroundSize: `${sweep * 100}% 100%`,
                backgroundPosition: '0 0.02em',
                boxDecorationBreak: 'clone',
                WebkitBoxDecorationBreak: 'clone',
              }}
            >
              {body}
            </span>
            {r < runs.length - 1 ? ' ' : null}
          </span>
        );
      })}
    </div>
  );
};

/** The small letterspaced line above a headline: a chapter, a section, a cue. */
const Note: React.FC<{theme: ResolvedTheme; text: string; p: number; align?: 'left' | 'center'}> = ({
  theme,
  text,
  p,
  align = 'left',
}) => (
  <div
    style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: align === 'center' ? 'center' : 'flex-start',
      gap: 14,
      opacity: p,
      transform: `translateY(${(1 - p) * -10}px)`,
      marginBottom: 22,
    }}
  >
    <span
      style={{
        width: 30,
        height: 3,
        borderRadius: 999,
        background: theme.accent,
        transform: `scaleX(${p})`,
        transformOrigin: 'left center',
      }}
    />
    <span
      style={{
        fontFamily: theme.fontBody,
        fontSize: 24,
        fontWeight: 700,
        letterSpacing: '0.18em',
        textTransform: 'uppercase',
        color: theme.accentText,
      }}
    >
      {text}
    </span>
  </div>
);

const Caption: React.FC<{
  theme: ResolvedTheme;
  text: string;
  p: number;
  size?: number;
  align?: 'left' | 'center';
  maxWidth?: string;
}> = ({theme, text, p, size = 33, align = 'left', maxWidth = '90%'}) => (
  <div
    style={{
      marginTop: 28,
      fontFamily: theme.fontBody,
      fontSize: size,
      lineHeight: 1.45,
      fontWeight: 500,
      color: theme.textMuted,
      opacity: p,
      transform: `translateY(${(1 - p) * 14}px)`,
      maxWidth,
      textAlign: align,
      alignSelf: align === 'center' ? 'center' : undefined,
    }}
  >
    {text}
  </div>
);

/**
 * One object on a panel: the figure, its label, and an optional second line.
 *
 * Used by `pair`, `row`, `steps` and `recap`, which differ in how many there
 * are and how they are arranged rather than in what one of them looks like.
 * Keeping the tile one component is what stops a three-object row and a
 * two-object pair drifting into two different visual languages inside one clip.
 *
 * `travel` is the only per-shot licence: how far, and from which direction, the
 * tile comes in. A `pair` converging from the outside edges performs the
 * comparison; a `row` rising in place performs a list. Same tile, and the
 * arrival is what says which shot you are watching.
 */
const Tile: React.FC<{
  theme: ResolvedTheme;
  object: SpineObject;
  size: number;
  p: number;
  t: number;
  /** A number chip in the corner, for `steps`. */
  step?: number;
  /** A tick that lands after the tile, for `recap`. */
  ticked?: number;
  width?: number | string;
  travel?: {x?: number; y?: number};
  labelSize?: number;
}> = ({theme, object, size, p, t, step, ticked, width, travel, labelSize = 36}) => {
  const dx = (travel?.x ?? 0) * (1 - p);
  const dy = (travel?.y ?? 26) * (1 - p);
  return (
    <div
      style={{
        position: 'relative',
        width,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 20,
        padding: '38px 28px 34px',
        borderRadius: 24,
        background: `${theme.surface}cc`,
        border: `1px solid ${theme.surfaceBorder}`,
        opacity: p,
        transform: `translate(${dx}px, ${dy}px) scale(${0.96 + 0.04 * p})`,
      }}
    >
      {step !== undefined && (
        <span
          style={{
            position: 'absolute',
            top: -18,
            left: 24,
            minWidth: 36,
            height: 36,
            padding: '0 10px',
            borderRadius: 999,
            background: theme.accent,
            // Type sitting ON the accent, which is a bright fill in both modes:
            // the dark stage colour in dark mode, the dark text colour in light.
            // `text` alone is white on yellow in dark mode; `bgBottom` alone is
            // near-white on yellow in light.
            color: theme.mode === 'light' ? theme.text : theme.bgBottom,
            fontFamily: theme.fontDisplay,
            fontSize: 21,
            fontWeight: 800,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {step}
        </span>
      )}
      {ticked !== undefined && ticked > 0.01 && (
        <span
          style={{
            position: 'absolute',
            top: -18,
            right: 24,
            width: 38,
            height: 38,
            borderRadius: 999,
            background: theme.accent,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            transform: `scale(${0.4 + 0.6 * ticked})`,
            opacity: ticked,
          }}
        >
          <svg width={20} height={20} viewBox="0 0 20 20">
            <path
              d="M4 10.5 l4 4 8 -9"
              fill="none"
              stroke={theme.mode === 'light' ? theme.text : theme.bgBottom}
              strokeWidth={2.8}
              strokeLinecap="round"
              strokeLinejoin="round"
              pathLength={1}
              strokeDasharray={1}
              strokeDashoffset={1 - ticked}
            />
          </svg>
        </span>
      )}
      <Art theme={theme} figure={object.figure} size={size} build={p} t={t} pool={0.4} />
      {object.label && (
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: labelSize,
            fontWeight: 700,
            lineHeight: 1.15,
            letterSpacing: '-0.01em',
            color: theme.text,
            textAlign: 'center',
          }}
        >
          {object.label}
        </div>
      )}
      {object.detail && (
        <div
          style={{
            fontFamily: theme.fontBody,
            fontSize: 26,
            lineHeight: 1.4,
            fontWeight: 500,
            color: theme.textMuted,
            textAlign: 'center',
          }}
        >
          {object.detail}
        </div>
      )}
    </div>
  );
};

/**
 * The rail: where this beat sits in the clip.
 *
 * One segment per beat, drawn from `index` rather than animated within the shot,
 * so it is continuous ACROSS the cut — the segments behind are already full when
 * a new shot starts and only the current one grows. A rail that re-animated
 * every beat would read as twelve separate loading bars.
 *
 * Segmented rather than a single bar because the two say different things. A
 * continuous fill says "part way through"; ticks with some behind you, one lit
 * and some still ahead say "this is the fourth of nine", and being able to see
 * the shape of what is coming is the whole job of a spine.
 */
const Rail: React.FC<{
  theme: ResolvedTheme;
  index: number;
  total: number;
  height: number;
  /** Dimmed for the shots that want the frame to themselves. */
  quiet?: boolean;
  p: number;
}> = ({theme, index, total, height, quiet, p}) => {
  const steps = Math.max(total, 1);
  const gap = steps > 1 ? Math.min(9, height * 0.012) : 0;
  const seg = (height - gap * (steps - 1)) / steps;
  const headY = index * (seg + gap) + seg * p;
  return (
    <div
      style={{
        position: 'relative',
        width: RAIL_W,
        height,
        opacity: quiet ? 0.4 : 1,
        flex: `0 0 ${RAIL_W}px`,
      }}
    >
      {Array.from({length: steps}, (_, i) => {
        const top = i * (seg + gap);
        // Three states, and they have to be told apart at four pixels wide:
        // done is solid, current fills, ahead is barely there.
        const fill = i < index ? 1 : i === index ? p : 0;
        return (
          <div
            key={i}
            style={{
              position: 'absolute',
              top,
              left: 0,
              width: RAIL_W,
              height: seg,
              borderRadius: 999,
              background: `${theme.line}1f`,
              overflow: 'hidden',
            }}
          >
            <div
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                height: `${fill * 100}%`,
                background:
                  i === index
                    ? theme.accent
                    : `linear-gradient(${theme.primary}, ${theme.primary}cc)`,
              }}
            />
          </div>
        );
      })}
      {/* The head, which is the only part that moves during a shot — enough to
          say "still running" without pulling the eye off the type. */}
      <div
        style={{
          position: 'absolute',
          top: headY - 7,
          left: '50%',
          width: 14,
          height: 14,
          marginLeft: -7,
          borderRadius: '50%',
          background: theme.accent,
          boxShadow: `0 0 0 6px ${theme.accent}22`,
        }}
      />
    </div>
  );
};

export const SpineScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  /** Frames this shot is on screen for, used only for the release before a cut. */
  durationInFrames?: number;
  props: Record<string, unknown>;
}> = ({theme, durationInFrames, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const t = frame / FPS;

  const shot = (String(props.shot ?? 'state') || 'state') as SpineShot;
  const headline = String(props.headline ?? '');
  const emphasis = String(props.emphasis ?? '');
  const caption = String(props.caption ?? '');
  const note = String(props.note ?? '');
  const objects = Array.isArray(props.objects) ? (props.objects as SpineObject[]) : [];
  const index = Number(props.index ?? 0);
  const total = Number(props.total ?? 1);

  const figureBuild = ramp(frame, 0, ENTER.figure);
  const noteP = ramp(frame, 0, ENTER.noteFrames);
  const words = headlineWords(headline, emphasis).length;
  const headlineDelay = note ? 6 : 0;
  const afterHeadline = headlineDelay + Math.max(words - 1, 0) * ENTER.wordStagger;
  const captionP = ramp(frame, afterHeadline + ENTER.captionDelay, ENTER.captionFrames);

  // Long headlines step down rather than wrapping to three lines, which at this
  // size stops being a headline and becomes a paragraph. The scale differs per
  // shot because the column it lands in does: a `focus` headline has the whole
  // frame, a `row` headline has the strip above four tiles.
  const sizeFor = (base: number) => {
    const chars = headline.length;
    const scale = chars <= 18 ? 1 : chars <= 30 ? 0.86 : chars <= 44 ? 0.72 : 0.62;
    return Math.round(base * scale);
  };

  const bodyW = STAGE_W - RAIL_W - RAIL_GUTTER;
  // The floor a row of tiles grows into. Without it a three-tile row is as tall
  // as its tallest label and the bottom third of the frame is empty.
  const tileRowH = Math.round(BODY_H * 0.54);

  const tileP = (i: number) => eased(frame, 4 + i * ENTER.tileStagger, 18);

  const body = (() => {
    switch (shot) {
      // ── The opening. Eyebrow, the biggest type in the clip, one figure
      // holding the right third, and a rule that sweeps the full width under
      // the title. It is a title card that is still a shot: the figure is
      // doing its idle before the first sentence is over.
      case 'open': {
        const obj = objects[0];
        const ruleP = eased(frame, afterHeadline + 6, 18);
        return (
          <div style={{display: 'flex', alignItems: 'center', gap: 72, width: bodyW, height: '100%'}}>
            <div style={{flex: '1 1 auto', minWidth: 0}}>
              {note && <Note theme={theme} text={note} p={noteP} />}
              <Headline
                theme={theme}
                text={headline}
                emphasis={emphasis}
                size={sizeFor(126)}
                frame={frame}
                fps={fps}
                delay={headlineDelay}
              />
              <div
                style={{
                  marginTop: 34,
                  height: 3,
                  borderRadius: 999,
                  background: `linear-gradient(90deg, ${theme.accent}, ${theme.primary}00)`,
                  transform: `scaleX(${ruleP})`,
                  transformOrigin: 'left center',
                }}
              />
              {caption && <Caption theme={theme} text={caption} p={captionP} size={36} />}
            </div>
            {obj && (
              <div style={{flex: '0 0 34%', display: 'flex', justifyContent: 'center'}}>
                <Art
                  theme={theme}
                  figure={obj.figure}
                  size={Math.min(bodyW * 0.32, BODY_H * 0.66)}
                  build={figureBuild}
                  t={t}
                />
              </div>
            )}
          </div>
        );
      }

      // ── A section marker. The ordinal is the whole picture: set enormous and
      // half-transparent behind the title, it says "you are entering part
      // three" faster than any line of type can, and it is the one shot that
      // makes a spine clip legible as a *course* rather than as a video.
      case 'chapter': {
        // The beat index is NOT the chapter number and using it was a bug you
        // could read off the frame: a chapter marker sitting second in a clip
        // rendered "Part one" under a giant 2. It is the planner's to supply,
        // because only the planner knows what part of the course this is; with
        // no ordinal the shot simply drops the numeral.
        const ord = Number(props.ordinal ?? 0);
        const numP = eased(frame, 0, 20);
        return (
          <div
            style={{
              position: 'relative',
              width: bodyW,
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
            }}
          >
            {ord > 0 && (
              <div
                style={{
                  position: 'absolute',
                  right: '4%',
                  top: '50%',
                  transform: `translateY(-50%) translateX(${(1 - numP) * 60}px)`,
                  fontFamily: theme.fontDisplay,
                  fontSize: 460,
                  fontWeight: 800,
                  lineHeight: 0.8,
                  letterSpacing: '-0.06em',
                  color: theme.accent,
                  opacity: 0.13 * numP,
                }}
              >
                {ord}
              </div>
            )}
            <div style={{position: 'relative', maxWidth: '74%'}}>
              {note && <Note theme={theme} text={note} p={noteP} />}
              <Headline
                theme={theme}
                text={headline}
                emphasis={emphasis}
                size={sizeFor(104)}
                frame={frame}
                fps={fps}
                delay={headlineDelay}
              />
              {caption && <Caption theme={theme} text={caption} p={captionP} size={34} />}
              {objects.length > 0 && (
                <div style={{marginTop: 44, display: 'flex', gap: 18, flexWrap: 'wrap'}}>
                  {objects.map((o, i) => {
                    const p = tileP(i);
                    return (
                      <div
                        key={i}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: 14,
                          padding: '12px 26px 12px 14px',
                          borderRadius: 999,
                          border: `1px solid ${theme.surfaceBorder}`,
                          background: `${theme.surface}99`,
                          opacity: p,
                          transform: `translateY(${(1 - p) * 16}px)`,
                        }}
                      >
                        {/* Sixty rather than forty-four: a figure has parts, and
                            below about this size they merge into one smear that
                            is worse than no figure at all. */}
                        <Art theme={theme} figure={o.figure} size={60} build={p} t={t} pool={0} />
                        {o.label && (
                          <span
                            style={{
                              fontFamily: theme.fontBody,
                              fontSize: 25,
                              fontWeight: 600,
                              color: theme.textMuted,
                            }}
                          >
                            {o.label}
                          </span>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </div>
        );
      }

      // ── The workhorse. One claim, one picture, one supporting line. Most
      // beats in most clips are this, which is why it gets the plainest grid
      // in the set: it has to survive being seen six times in ninety seconds.
      case 'state': {
        const obj = objects[0];
        const flip = index % 2 === 1;
        return (
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 76,
              width: bodyW,
              height: '100%',
              flexDirection: flip ? 'row-reverse' : 'row',
            }}
          >
            {obj && (
              <div style={{flex: '0 0 36%', display: 'flex', justifyContent: 'center'}}>
                <Art
                  theme={theme}
                  figure={obj.figure}
                  size={Math.min(bodyW * 0.34, BODY_H * 0.68)}
                  build={figureBuild}
                  t={t}
                />
              </div>
            )}
            <div style={{flex: '1 1 auto', minWidth: 0}}>
              {note && <Note theme={theme} text={note} p={noteP} />}
              <Headline
                theme={theme}
                text={headline}
                emphasis={emphasis}
                size={sizeFor(106)}
                frame={frame}
                fps={fps}
                delay={headlineDelay}
              />
              {caption && <Caption theme={theme} text={caption} p={captionP} />}
              {obj?.label && (
                <div
                  style={{
                    marginTop: 26,
                    display: 'inline-block',
                    padding: '9px 18px',
                    borderRadius: 999,
                    border: `1px solid ${theme.surfaceBorder}`,
                    background: `${theme.surface}aa`,
                    fontFamily: theme.fontBody,
                    fontSize: 24,
                    fontWeight: 600,
                    letterSpacing: '0.02em',
                    color: theme.textMuted,
                    opacity: captionP,
                  }}
                >
                  {obj.label}
                </div>
              )}
            </div>
          </div>
        );
      }

      // ── Two things held against each other. The tiles converge from the
      // outside edges and the divider draws between them, so the shot performs
      // the comparison instead of presenting it already made.
      case 'pair': {
        const [a, b] = objects;
        const dividerP = eased(frame, 4 + ENTER.tileStagger, 16);
        return (
          <div style={{width: bodyW, height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center'}}>
            {note && <Note theme={theme} text={note} p={noteP} />}
            <Headline
              theme={theme}
              text={headline}
              emphasis={emphasis}
              size={sizeFor(80)}
              frame={frame}
              fps={fps}
              delay={headlineDelay}
            />
            <div
              style={{
                marginTop: 44,
                minHeight: tileRowH,
                display: 'flex',
                alignItems: 'stretch',
                justifyContent: 'center',
                gap: 44,
              }}
            >
              {a && (
                <Tile
                  theme={theme}
                  object={a}
                  size={252}
                  p={tileP(0)}
                  t={t}
                  width={'42%'}
                  travel={{x: -70, y: 0}}
                />
              )}
              <div
                style={{
                  width: 2,
                  alignSelf: 'stretch',
                  borderRadius: 999,
                  background: `linear-gradient(${theme.line}00, ${theme.line}55, ${theme.line}00)`,
                  transform: `scaleY(${dividerP})`,
                }}
              />
              {b && (
                <Tile
                  theme={theme}
                  object={b}
                  size={252}
                  p={tileP(1)}
                  t={t}
                  width={'42%'}
                  travel={{x: 70, y: 0}}
                />
              )}
            </div>
            {caption && (
              <Caption theme={theme} text={caption} p={captionP} align="center" maxWidth="78%" />
            )}
          </div>
        );
      }

      // ── A list that is a list. Three to five tiles arriving left to right,
      // which is the shot for "here is what this covers" — the one an intro
      // needs and the one a recap needs, at either end of the same clip.
      case 'row': {
        const n = Math.max(objects.length, 1);
        const size = n >= 5 ? 152 : n === 4 ? 178 : 208;
        return (
          <div style={{width: bodyW, height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center'}}>
            {note && <Note theme={theme} text={note} p={noteP} />}
            <Headline
              theme={theme}
              text={headline}
              emphasis={emphasis}
              size={sizeFor(80)}
              frame={frame}
              fps={fps}
              delay={headlineDelay}
            />
            <div
              style={{
                marginTop: 44,
                minHeight: tileRowH,
                display: 'flex',
                gap: 26,
                alignItems: 'stretch',
              }}
            >
              {objects.map((o, i) => (
                <Tile
                  key={i}
                  theme={theme}
                  object={o}
                  size={size}
                  p={tileP(i)}
                  t={t}
                  width={`${100 / n}%`}
                  labelSize={n >= 5 ? 30 : 36}
                />
              ))}
            </div>
            {caption && (
              <Caption theme={theme} text={caption} p={captionP} align="center" maxWidth="80%" />
            )}
          </div>
        );
      }

      // ── One idea and the things hanging off it. The connectors draw outward
      // from the centre, so the shot says "these belong to that" rather than
      // "here are six objects".
      case 'orbit': {
        const [centre, ...spokes] = objects;
        const n = Math.max(spokes.length, 1);
        // The box is sized from the HEIGHT and then checked against the width,
        // not the other way round: the spokes sit on a circle, so the binding
        // constraint is always the shorter axis, and sizing from the width is
        // what used to push the right-hand spoke into the frame margin.
        const box = Math.min(BODY_H * 0.86, bodyW * 0.5);
        const spokeSize = box * 0.17;
        // Room for a label under each spoke, so the ring does not have to be
        // shrunk to keep the type inside the box.
        const label = box * 0.16;
        const R = (box - spokeSize - label) / 2;
        const centreSize = box * 0.28;
        return (
          <div style={{display: 'flex', alignItems: 'center', gap: 60, width: bodyW, height: '100%'}}>
            <div style={{flex: '1 1 auto', minWidth: 0}}>
              {note && <Note theme={theme} text={note} p={noteP} />}
              <Headline
                theme={theme}
                text={headline}
                emphasis={emphasis}
                size={sizeFor(82)}
                frame={frame}
                fps={fps}
                delay={headlineDelay}
              />
              {caption && <Caption theme={theme} text={caption} p={captionP} size={31} />}
            </div>
            <div style={{flex: `0 0 ${box}px`, height: box, position: 'relative'}}>
              <svg
                width={box}
                height={box}
                style={{position: 'absolute', inset: 0, overflow: 'visible'}}
              >
                {spokes.map((_, i) => {
                  const a = (i / n) * Math.PI * 2 - Math.PI / 2;
                  const p = tileP(i);
                  const x = box / 2 + Math.cos(a) * R * p;
                  const y = box / 2 + Math.sin(a) * R * p;
                  return (
                    <line
                      key={i}
                      x1={box / 2}
                      y1={box / 2}
                      x2={x}
                      y2={y}
                      stroke={theme.line}
                      strokeOpacity={0.35}
                      strokeWidth={2}
                      strokeLinecap="round"
                    />
                  );
                })}
              </svg>
              {centre && (
                <div
                  style={{
                    position: 'absolute',
                    left: box / 2 - centreSize / 2,
                    top: box / 2 - centreSize / 2,
                  }}
                >
                  <Art
                    theme={theme}
                    figure={centre.figure}
                    size={centreSize}
                    build={figureBuild}
                    t={t}
                  />
                </div>
              )}
              {spokes.map((o, i) => {
                const a = (i / n) * Math.PI * 2 - Math.PI / 2;
                const p = tileP(i);
                // Each spoke rides out along its own connector, which is what
                // makes the ring assemble rather than appear.
                const travelled = R * p;
                return (
                  <div
                    key={i}
                    style={{
                      position: 'absolute',
                      // A fixed box wider than the figure, centred on the spoke:
                      // the label wraps inside it instead of running off the
                      // side of the ring, which `nowrap` used to let it do.
                      left: box / 2 + Math.cos(a) * travelled - (spokeSize + label) / 2,
                      top: box / 2 + Math.sin(a) * travelled - spokeSize / 2,
                      width: spokeSize + label,
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      opacity: p,
                      transform: `scale(${0.7 + 0.3 * p})`,
                    }}
                  >
                    <Art
                      theme={theme}
                      figure={o.figure}
                      size={spokeSize}
                      build={p}
                      t={t}
                      pool={0.4}
                    />
                    {o.label && (
                      <div
                        style={{
                          marginTop: 8,
                          fontFamily: theme.fontBody,
                          fontSize: 22,
                          fontWeight: 600,
                          lineHeight: 1.2,
                          color: theme.textMuted,
                          textAlign: 'center',
                        }}
                      >
                        {o.label}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        );
      }

      // ── An order of operations. Numbered, connected, and arriving in the
      // order they happen, because a sequence drawn all at once is a list. The
      // connector between two tiles draws before the tile it points at lands,
      // so the sequence pulls itself along rather than filling in.
      case 'steps': {
        const n = Math.max(objects.length, 1);
        const size = n >= 4 ? 158 : 194;
        return (
          <div style={{width: bodyW, height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center'}}>
            {note && <Note theme={theme} text={note} p={noteP} />}
            <Headline
              theme={theme}
              text={headline}
              emphasis={emphasis}
              size={sizeFor(80)}
              frame={frame}
              fps={fps}
              delay={headlineDelay}
            />
            <div
              style={{
                marginTop: 50,
                minHeight: tileRowH,
                display: 'flex',
                alignItems: 'stretch',
                justifyContent: 'center',
                gap: 0,
              }}
            >
              {objects.map((o, i) => (
                <div
                  key={i}
                  style={{display: 'flex', alignItems: 'stretch', flex: `1 1 ${100 / n}%`}}
                >
                  <Tile
                    theme={theme}
                    object={o}
                    size={size}
                    p={tileP(i)}
                    t={t}
                    step={i + 1}
                    width="100%"
                    labelSize={n >= 4 ? 31 : 36}
                  />
                  {i < objects.length - 1 && (
                    <div
                      style={{
                        flex: '0 0 40px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                      }}
                    >
                      {/* Drawn on the NEXT tile's clock but ahead of it, so the
                          arrow reaches before the thing it points at exists. */}
                      <svg width={26} height={20} viewBox="0 0 22 18">
                        <path
                          d="M2 9 h14 M12 4 l6 5 -6 5"
                          fill="none"
                          stroke={theme.accent}
                          strokeWidth={2.4}
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          pathLength={1}
                          strokeDasharray={1}
                          strokeDashoffset={1 - eased(frame, 4 + i * ENTER.tileStagger + 3, 12)}
                        />
                      </svg>
                    </div>
                  )}
                </div>
              ))}
            </div>
            {caption && (
              <Caption theme={theme} text={caption} p={captionP} align="center" maxWidth="80%" />
            )}
          </div>
        );
      }

      // ── Looking back. The same tiles as a `row`, with a tick landing on each
      // one after it arrives. A recap that merely lists what was covered is a
      // row; the tick is what makes it a recap, because it says these are done
      // rather than these are coming.
      case 'recap': {
        const n = Math.max(objects.length, 1);
        const size = n >= 4 ? 152 : 186;
        return (
          <div style={{width: bodyW, height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center'}}>
            {note && <Note theme={theme} text={note} p={noteP} />}
            <Headline
              theme={theme}
              text={headline}
              emphasis={emphasis}
              size={sizeFor(80)}
              frame={frame}
              fps={fps}
              delay={headlineDelay}
            />
            <div
              style={{
                marginTop: 46,
                minHeight: tileRowH,
                display: 'flex',
                gap: 26,
                alignItems: 'stretch',
              }}
            >
              {objects.map((o, i) => (
                <Tile
                  key={i}
                  theme={theme}
                  object={o}
                  size={size}
                  p={tileP(i)}
                  t={t}
                  ticked={eased(frame, 16 + i * ENTER.tileStagger * 1.6, 14)}
                  width={`${100 / n}%`}
                  labelSize={n >= 4 ? 31 : 36}
                />
              ))}
            </div>
            {caption && (
              <Caption theme={theme} text={caption} p={captionP} align="center" maxWidth="80%" />
            )}
          </div>
        );
      }

      // ── A parenthesis. Indented behind a bracket, set smaller and quieter
      // than the shots either side of it, and deliberately NOT filling the
      // frame — the whole point is that the clip has stepped sideways for a
      // moment and is about to step back. It is the one shot whose composition
      // is an interruption of the others rather than a variation on them.
      case 'aside': {
        const obj = objects[0];
        const barP = eased(frame, 0, 16);
        return (
          <div
            style={{
              width: bodyW,
              height: '100%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <div style={{display: 'flex', alignItems: 'stretch', gap: 38, maxWidth: '80%'}}>
              <div
                style={{
                  flex: '0 0 5px',
                  borderRadius: 999,
                  background: theme.accent,
                  opacity: 0.55,
                  transform: `scaleY(${barP})`,
                  transformOrigin: 'top center',
                }}
              />
              <div style={{display: 'flex', alignItems: 'center', gap: 44}}>
                <div style={{flex: '1 1 auto', minWidth: 0}}>
                  {note && <Note theme={theme} text={note} p={noteP} />}
                  <Headline
                    theme={theme}
                    text={headline}
                    emphasis={emphasis}
                    size={sizeFor(72)}
                    frame={frame}
                    fps={fps}
                    delay={headlineDelay}
                  />
                  {caption && <Caption theme={theme} text={caption} p={captionP} size={30} />}
                </div>
                {obj && (
                  <div style={{flex: '0 0 auto'}}>
                    <Art
                      theme={theme}
                      figure={obj.figure}
                      size={Math.min(bodyW * 0.19, BODY_H * 0.34)}
                      build={figureBuild}
                      t={t}
                      pool={0.5}
                    />
                  </div>
                )}
              </div>
            </div>
          </div>
        );
      }

      // ── A breath. One object, very large, and as little type as the beat can
      // get away with. This is the shot a course needs between two dense ones
      // and that no content template will ever produce, because its whole job
      // is to say less.
      case 'focus': {
        const obj = objects[0];
        return (
          <div
            style={{
              width: bodyW,
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            {obj && (
              <Art
                theme={theme}
                figure={obj.figure}
                size={Math.min(bodyW * 0.42, BODY_H * 0.56)}
                build={figureBuild}
                t={t}
              />
            )}
            <div style={{marginTop: 46, maxWidth: '80%'}}>
              <Headline
                theme={theme}
                text={headline}
                emphasis={emphasis}
                size={sizeFor(92)}
                frame={frame}
                fps={fps}
                align="center"
                delay={headlineDelay}
              />
            </div>
            {caption && (
              <Caption theme={theme} text={caption} p={captionP} align="center" maxWidth="66%" />
            )}
          </div>
        );
      }

      // ── The line that has to land on its own. No figure: an object beside a
      // punchline is the thing that stops it being a punchline. The marks are
      // set as a matched pair around the line rather than one floating above
      // it, which is what stops the shot reading as an unfinished slide.
      case 'quote': {
        const markP = eased(frame, 0, 14);
        const closeP = ramp(frame, afterHeadline + 4, 12);
        const mark = (p: number) => ({
          fontFamily: theme.fontDisplay,
          fontSize: 168,
          lineHeight: 0.62,
          fontWeight: 800,
          color: theme.accent,
          opacity: 0.42 * p,
          transform: `translateY(${(1 - p) * 12}px)`,
        });
        return (
          <div
            style={{
              width: bodyW,
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              alignItems: 'center',
            }}
          >
            <div style={{maxWidth: 1360}}>
              <div style={mark(markP)}>&ldquo;</div>
              <div style={{marginTop: 24}}>
                <Headline
                  theme={theme}
                  text={headline}
                  emphasis={emphasis}
                  size={sizeFor(104)}
                  frame={frame}
                  fps={fps}
                  delay={headlineDelay}
                />
              </div>
              <div style={{...mark(closeP), textAlign: 'right', marginTop: 18}}>&rdquo;</div>
              {caption && <Caption theme={theme} text={caption} p={captionP} size={32} />}
            </div>
          </div>
        );
      }

      // ── The exit. The rail is full, the figure is centred rather than off to
      // one side, and the caption is a next step rather than a supporting
      // point — a clip that ends on the workhorse layout does not feel ended.
      case 'close': {
        const obj = objects[0];
        return (
          <div
            style={{
              width: bodyW,
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              textAlign: 'center',
            }}
          >
            {note && <Note theme={theme} text={note} p={noteP} align="center" />}
            {obj && (
              <div style={{marginBottom: 44}}>
                <Art
                  theme={theme}
                  figure={obj.figure}
                  size={Math.min(bodyW * 0.28, BODY_H * 0.42)}
                  build={figureBuild}
                  t={t}
                />
              </div>
            )}
            <div style={{maxWidth: '84%'}}>
              <Headline
                theme={theme}
                text={headline}
                emphasis={emphasis}
                size={sizeFor(94)}
                frame={frame}
                fps={fps}
                align="center"
                delay={headlineDelay}
              />
            </div>
            {caption && (
              <div
                style={{
                  marginTop: 38,
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 14,
                  padding: '14px 26px',
                  borderRadius: 999,
                  border: `1px solid ${theme.accent}66`,
                  background: `${theme.accent}12`,
                  opacity: captionP,
                  transform: `translateY(${(1 - captionP) * 14}px)`,
                }}
              >
                <span
                  style={{
                    width: 9,
                    height: 9,
                    borderRadius: '50%',
                    background: theme.accent,
                  }}
                />
                <span
                  style={{
                    fontFamily: theme.fontBody,
                    fontSize: 29,
                    fontWeight: 600,
                    color: theme.text,
                  }}
                >
                  {caption}
                </span>
              </div>
            )}
          </div>
        );
      }

      default:
        return null;
    }
  })();

  // `focus`, `quote` and `aside` are the shots whose whole job is to hold the
  // frame still or to step out of it, so the rail steps back rather than
  // competing.
  const quiet = shot === 'focus' || shot === 'quote' || shot === 'aside';

  // The release. Over the last few frames the composition settles back a hair
  // and loses a little light, so the cut lands on something already leaving
  // rather than on something still holding at full strength. Without a duration
  // — a fixture, a still — there is nothing to release into, and it is skipped.
  const out = durationInFrames
    ? ramp(frame, durationInFrames - ENTER.exit, ENTER.exit)
    : 0;

  return (
    <Stage>
      <div
        style={{
          display: 'flex',
          alignItems: 'stretch',
          gap: RAIL_GUTTER,
          width: STAGE_W,
          height: BODY_H,
          opacity: 1 - out * 0.35,
          transform: `scale(${1 - out * 0.012})`,
        }}
      >
        <Rail
          theme={theme}
          index={index}
          total={total}
          height={BODY_H}
          quiet={quiet}
          p={clamp01(figureBuild)}
        />
        <div
          style={{
            flex: '1 1 auto',
            minWidth: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: quiet || shot === 'close' ? 'center' : 'flex-start',
          }}
        >
          {body}
        </div>
      </div>
    </Stage>
  );
};
