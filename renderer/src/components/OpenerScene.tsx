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

/** The ink the big type is set at. See the file header — this is the whole design. */
const GROUND_INK = 0.12;

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
  const groundSize = words.length <= 5 ? 250 : words.length <= 7 ? 205 : 170;

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
              fontWeight: 400,
              // Tight. A serif at this size with normal leading breaks into
              // separate lines of unrelated words; at 0.92 the lines lock into one
              // block, which is what makes it read as a mass rather than as text.
              lineHeight: 0.92,
              letterSpacing: -4,
              color: withAlpha(theme.text, GROUND_INK),
              width: '100%',
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
          ) : null}

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
