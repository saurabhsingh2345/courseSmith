import {interpolate, useCurrentFrame} from 'remotion';
import {AbsoluteFill} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';

// TitleCardScene is the card that cuts between sections: one line, held, read.
//
// It is the counterpart to OpenerScene and almost every decision is inverted.
// An opener sets its subject at 32% ink as the GROUND of the frame, to be taken
// in as texture; this sets one short line at full ink in the middle of a bone
// card, to be read. An opener appears once, at the front. This appears four or
// five times in one video, which is the constraint that shapes everything else.
//
// THE CARD PAINTS ITS OWN GROUND. Every other scene composes onto the stage and
// inherits the skin's backdrop. This one fills the frame with the surface tone
// instead, and that is the entire reason it works in a long piece: cutting from a
// warm ground to a bone card and back is a rhythm a viewer feels as chapters.
// Inheriting the ground would make the card a slide with big text on it, which is
// the thing it exists to not be.
//
// THE SPLIT IS ONE LINE, TWO INKS. `line` solid, `tail` at muted — and they sit
// on the same baseline as one phrase rather than stacking. Stacked, they are two
// headlines and the viewer reads them as a list; inline, the second half is
// audibly subordinate to the first and the card makes an argument in four words.
//
// THE FACE CARRIES THE MODE. A section name is set in the display serif; a literal
// command is set in mono at nearly the same size. That single substitution is the
// clearest signal in the whole design language — serif means "here is an idea",
// mono means "here is something you type" — and it costs no words of narration to
// establish.
//
// The one thing that moves is a slow push. Unlike the opener, which holds dead
// still because a title page is a printed object, a section card is a cut inside a
// moving piece: a 2% scale over its whole life keeps it alive without ever
// announcing itself as an animation.

/**
 * Why the camera settles and then STOPS.
 *
 * A continuous scale on a layer with text in it re-rasterizes every glyph every
 * frame at a slightly different subpixel offset, and the eye reads that as the
 * type glittering. Measured on a held card: 2% of the frame's pixels changed
 * between consecutive frames with nothing supposed to be moving, 12,000 of them
 * by more than six levels. A 2% drift nobody consciously notices is not worth a
 * frame that will not sit still.
 *
 * So the move is: settle in, land on EXACTLY 1, and hold there. Everything that
 * makes the frame feel alive after that is content — a line landing, a row
 * lighting, a caret blinking — which is motion that means something anyway.
 * Where a held frame still wants a nudge, it is whole pixels of translate, which
 * resamples nothing.
 */

type Step = {
  startMs: number;
  endMs: number;
  show: 'line' | 'tail' | 'note' | 'stack';
  tail?: boolean;
  note?: boolean;
  stack?: boolean;
};

type Props = {
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: {
    line?: string;
    mono?: boolean;
    tail?: string;
    label?: string;
    note?: string;
    stack?: string[];
    stackNote?: string;
    steps?: Step[];
  };
};

/**
 * Display size for the line, stepped down as it gets longer.
 *
 * A fixed size cannot serve both "Scope" and a seven-word line: the first wants
 * to be enormous and the second has to fit on two lines without hyphenating. The
 * steps are measured against the 1500px measure this card allows itself, and they
 * are coarse on purpose — three sizes read as three deliberate treatments, where
 * a continuous fit function reads as text that shrank.
 */
const sizeFor = (chars: number, mono: boolean): number => {
  const base = mono ? 0.78 : 1; // mono sets wider at the same point size
  if (chars <= 10) return 210 * base;
  if (chars <= 20) return 165 * base;
  if (chars <= 34) return 128 * base;
  return 104 * base;
};

export const TitleCardScene = ({theme, sceneStartMs, props}: Props) => {
  const frame = useCurrentFrame();
  const ms = (frame / FPS) * 1000 + sceneStartMs;
  const steps = props.steps ?? [];
  const line = props.line ?? '';
  const tail = props.tail ?? '';
  const mono = Boolean(props.mono);

  const step = steps.find((s) => ms >= s.startMs && ms < s.endMs) ?? steps[steps.length - 1];
  const sceneStart = steps[0]?.startMs ?? sceneStartMs;
  const sceneEnd = steps[steps.length - 1]?.endMs ?? sceneStart + 4000;

  // The settle, then dead still. See the note above the component.
  const scale = interpolate(ms, [sceneStart, sceneStart + 700], [1.01, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  // Each part rises as it lands. Short, and eased at the end rather than the
  // start: type that decelerates into place reads as set, type that accelerates
  // reads as thrown.
  const arrive = (from: number) =>
    interpolate(ms, [from, from + 420], [0, 1], {
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp',
    });

  const lineIn = arrive(sceneStart);
  const tailStep = steps.find((s) => s.show === 'tail');
  const noteStep = steps.find((s) => s.show === 'note');
  const stackStep = steps.find((s) => s.show === 'stack');
  const tailIn = step?.tail && tailStep ? arrive(tailStep.startMs) : 0;
  const noteIn = step?.note && noteStep ? arrive(noteStep.startMs) : 0;
  const stackIn = step?.stack && stackStep ? arrive(stackStep.startMs) : 0;
  const stack = props.stack ?? [];

  const size = sizeFor(line.length + tail.length, mono);
  const face = mono ? theme.fontMono : theme.fontSerif;

  return (
    <AbsoluteFill style={{background: theme.surface}}>
      <AbsoluteFill
        style={{
          alignItems: 'center',
          justifyContent: 'center',
          transform: `scale(${scale})`,
        }}
      >
        <div style={{maxWidth: 1500, padding: '0 60px', textAlign: 'center'}}>
          {props.label ? (
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 22,
                letterSpacing: 3.4,
                textTransform: 'uppercase',
                color: theme.accentText,
                opacity: lineIn * 0.9,
                marginBottom: 34,
              }}
            >
              {props.label}
            </div>
          ) : null}

          <div
            style={{
              fontFamily: face,
              fontSize: size,
              lineHeight: 1.06,
              letterSpacing: mono ? -1 : -2,
              color: theme.text,
              opacity: lineIn,
              transform: `translateY(${(1 - lineIn) * 16}px)`,
            }}
          >
            {line}
            {tail ? (
              <span
                style={{
                  // The muted half. Painted with the text token at low alpha
                  // rather than with textMuted: on this card the two halves have
                  // to be the SAME colour at different strengths, and textMuted
                  // is a different hue that reads as a second voice.
                  color: withAlpha(theme.text, 0.3),
                  opacity: tailIn,
                  transform: `translateY(${(1 - tailIn) * 10}px)`,
                  display: 'inline-block',
                  marginLeft: size * 0.22,
                }}
              >
                {tail}
              </span>
            ) : null}
          </div>

          {stack.length ? (
            // The ladder. Every level but the last at low ink, so the order reads
            // before any of the words do: what is faded is what gets overridden.
            <div style={{marginTop: 52, opacity: stackIn, transform: `translateY(${(1 - stackIn) * 12}px)`}}>
              {stack.map((level, i) => {
                const last = i === stack.length - 1;
                return (
                  <div
                    key={i}
                    style={{
                      display: 'flex',
                      alignItems: 'baseline',
                      justifyContent: 'center',
                      gap: 22,
                      marginBottom: 10,
                    }}
                  >
                    <span
                      style={{
                        fontFamily: theme.fontSerif,
                        fontSize: 86,
                        lineHeight: 1.08,
                        letterSpacing: -1.4,
                        color: last ? theme.text : withAlpha(theme.text, 0.26),
                      }}
                    >
                      {level}
                    </span>
                    {last && props.stackNote ? (
                      <span style={{fontFamily: theme.fontMono, fontSize: 26, color: theme.textMuted}}>
                        {props.stackNote}
                      </span>
                    ) : null}
                  </div>
                );
              })}
            </div>
          ) : null}

          {props.note ? (
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 34,
                lineHeight: 1.4,
                color: theme.textMuted,
                opacity: noteIn,
                transform: `translateY(${(1 - noteIn) * 10}px)`,
                marginTop: 44,
              }}
            >
              {props.note}
            </div>
          ) : null}
        </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};
