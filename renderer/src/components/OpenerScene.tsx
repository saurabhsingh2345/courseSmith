import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {SAFE_X, Stage} from './Stage';

// OpenerScene is a title page: the subject set enormous as the ground of the
// frame, and the two lines you actually read sitting small and solid on it.
//
// The whole composition rests on one number — the ink the big type is set at — and
// it took getting wrong twice to find. At 30% the words are legible, so the eye
// reads them, and then reads the promise, and the frame is two competing headlines.
// At 4% they vanish and the frame is a small line of text on an empty page. At 12%
// they are shape rather than language: the viewer takes them in the way they take
// in a watermark, and the only thing they READ is the solid line. That is the
// effect, and it is why the validator insists on four or more words up there —
// texture needs area, and three words at this size is a logo with a caption.
//
// It is the one scene in the catalog set in a serif, and the one that does not
// use SceneHeader. Both are deliberate: a shared header would put this title at
// label size in the house sans, which is the frame this template was written to
// replace.
//
// Nothing here moves except opacity. A title page that slides or scales is a lower
// third, and the reference this borrows from holds the frame dead still for ten
// seconds — which is what makes it read as a printed page rather than as an
// animation waiting to finish.

type Step = {
  startMs: number;
  endMs: number;
  show: 'ground' | 'promise' | 'mark';
  promise?: boolean;
  mark?: boolean;
};

/**
 * The ink the big type is set at.
 *
 * 0.32, and it started at 0.12 on the reasoning in the file header — that at 30%
 * the words compete with the promise. On the reference frame that was true because
 * a PERSON stood in front of the type, breaking it up and adding contrast of their
 * own. On an empty page it is not: 12% ink across a near-white sheet does not read
 * as texture, it reads as a title that failed to load, and the frame looks washed
 * out rather than composed.
 *
 * At 0.32 the words are unmistakably there and still clearly the ground — the
 * promise below them is at full ink, which is a three-fold difference, and that gap
 * is what keeps the hierarchy rather than the absolute value.
 */
const GROUND_INK = 0.32;

/**
 * Optical weight for a single-weight face.
 *
 * Instrument Serif ships one weight, 400, so `fontWeight: 800` on it does nothing
 * at all — the browser has no bolder cut to reach for and (in a headless render)
 * will not synthesise one. The two levers that DO work on a display serif are ink,
 * above, and stroke: painting a hairline of the same colour around each glyph
 * thickens the stems without touching the letterforms.
 *
 * Kept proportional to the size rather than fixed. 0.9% of the cap height is about
 * two pixels at 250pt, which reads as a heavier cut; past about 2% the counters in
 * `e` and `a` start closing up and it reads as a smudge.
 */
const GROUND_STROKE = 0.009;

export const OpenerScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const ground = String(props.ground ?? '');
  const kicker = String(props.kicker ?? '');
  const promise = String(props.promise ?? '');
  const byline = String(props.byline ?? '');
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (!ground || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];

  // The ground fades up once, over about a second, and then never changes again.
  const up = interpolate(frame, [0, 30], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const since = ((nowMs - step.startMs) / 1000) * FPS;
  const land = interpolate(since, [0, 16], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // A line that landed on an earlier beat is fully up; the one landing now fades.
  const inkFor = (isNow: boolean, shown: boolean) => (shown ? (isNow ? land : 1) : 0);

  const words = ground.split(/\s+/).filter(Boolean);
  // The size is chosen from the word count rather than fixed, because the point is
  // that the type FILLS the frame: nine words at the size four words want would run
  // off the page, and four words at the size nine want would leave a band of empty
  // paper across the middle. Measured against a 1700px measure at these weights.
  const groundSize = words.length <= 5 ? 250 : words.length <= 7 ? 205 : 178;

  return (
    <Stage justify="center" align="stretch">
      <div style={{position: 'relative', width: '100%', height: '100%'}}>
        {/* The ground. Absolutely placed and vertically centred so it is the page
            rather than a block in a column — and pointer-events are irrelevant in
            a render, but the layering is not: everything solid is drawn after it. */}
        <div
          style={{
            position: 'absolute',
            inset: 0,
            display: 'flex',
            alignItems: 'center',
            opacity: up,
          }}
        >
          <div
            style={{
              fontFamily: theme.fontSerif,
              fontSize: groundSize,
              // 400 because that is the only cut this face has. The weight comes
              // from the stroke below; see GROUND_STROKE.
              fontWeight: 400,
              WebkitTextStroke: `${(groundSize * GROUND_STROKE).toFixed(2)}px ${withAlpha(theme.text, GROUND_INK)}`,
              // Tight. A serif at this size with normal leading breaks into
              // separate lines of unrelated words; at 0.92 the lines lock into one
              // block, which is what makes it read as a mass rather than as text.
              lineHeight: 0.92,
              letterSpacing: -4,
              color: withAlpha(theme.text, GROUND_INK),
              // 88% of the box, not 100%.
              //
              // Full-bleed is what read as "stretched": a nine-word title set edge
              // to edge puts seven words on a line that spans the whole frame and
              // two on the next, so the block is a wide band with a short tail and
              // the eye reads the first line as horizontally pulled. Pulling the
              // measure in breaks the lines nearer the middle, which makes the
              // block a mass rather than a banner — and a mass is what a title page
              // wants.
              width: '88%',
            }}
          >
            {ground}
          </div>
        </div>

        {/* The things that are actually read. Bottom-left, against the frame's own
            margin — the one asymmetric placement in this family, and it is what
            leaves the upper two thirds free for the ground to be a page. */}
        <div
          style={{
            position: 'absolute',
            left: 0,
            bottom: 0,
            display: 'flex',
            flexDirection: 'column',
            gap: 22,
            maxWidth: 1180,
          }}
        >
          {kicker ? (
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 22,
                fontWeight: 600,
                letterSpacing: 6,
                textTransform: 'uppercase',
                color: theme.accentText,
                opacity: inkFor(step.show === 'promise', Boolean(step.promise)),
              }}
            >
              {kicker}
            </div>
          ) : (
            // No kicker: a short accent rule takes its place.
            //
            // Not decoration filling a hole. The kicker was doing structural work —
            // it marked where the read-this zone began, so the promise had something
            // to hang from. Strip it and the promise floats in the lower left with
            // nothing establishing that corner, which on a page that is otherwise
            // one huge pale word reads as a stray caption. A 90-pixel rule in the
            // accent says "the frame starts here" in one mark instead of three words.
            <div
              style={{
                width: 90,
                height: 4,
                borderRadius: 2,
                background: theme.accentText,
                opacity: inkFor(step.show === 'promise', Boolean(step.promise)),
              }}
            />
          )}

          <div
            style={{
              fontFamily: theme.fontSerif,
              // Large for a line of body copy and small next to the ground, which
              // is the ratio doing the work: about a fifth of the big type, so the
              // two never read as the same voice at the same volume.
              fontSize: 62,
              fontWeight: 400,
              lineHeight: 1.16,
              letterSpacing: -1,
              color: theme.text,
              opacity: inkFor(step.show === 'promise', Boolean(step.promise)),
            }}
          >
            {promise}
          </div>

          {byline ? (
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 18,
                opacity: inkFor(step.show === 'mark', Boolean(step.mark)),
              }}
            >
              {/* A short rule before the byline. The one piece of furniture on the
                  frame, and it is here because a name on its own under a sentence
                  reads as another sentence — the rule says "this is a signature". */}
              <div style={{width: 54, height: 2, background: withAlpha(theme.text, 0.35)}} />
              <div
                style={{
                  fontFamily: theme.fontBody,
                  fontSize: 26,
                  fontWeight: 500,
                  letterSpacing: 1.5,
                  color: theme.textMuted,
                }}
              >
                {byline}
              </div>
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};

/** Exported for the fixture's benefit: the margin the ground is set against. */
export const OPENER_MARGIN = SAFE_X;
