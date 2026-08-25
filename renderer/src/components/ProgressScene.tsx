import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';
import {CardLabel, roleColour} from './showroomCard';
import {iconFor} from './icons';

// ProgressScene is one road with NOW standing on it: finished work ticked behind
// you, coming work still faint ahead.
//
// The first cut of this drew two columns with a divider between them, and it was
// wrong in a way worth recording. Two columns say "here are two lists". They
// cannot say "this is one road and you are at this point on it", which is the
// entire claim the narration makes when it summarises a stretch of work — and the
// claim is what makes the frame encouraging rather than administrative. So the
// composition is a single rail with stations on it, and the rail is the only
// element that is never obscured.
//
// Stations alternate above and below the rail, and that is load-bearing rather
// than decorative. A row of eight chips in one line gives each of them an eighth
// of the frame, which is not enough width for a four-word label. Alternating
// means a chip only has to clear the neighbour TWO stations along, so it can run
// to nearly twice its own cell — which is what lets "Claude Desktop installed"
// sit on a frame that also has four things coming.
//
// What a beat moves is INK, never opacity. On paper a faded element fades into
// the page and stops being an element, so a not-yet station keeps its full
// surface and its hairline and lacks only dark text, a drawn tick, and its stem.
// This family has re-learned that rule once per template — see showroomCard.
//
// Three things animate, in the order the eye should take them. The rail draws
// left to right and carries on PAST the pin on the second beat, which is the only
// animation here that spans two beats and the reason the frame reads as one road
// rather than two states of a diagram. A tick strikes into each finished station,
// scaled in rather than faded, because a tick that fades up reads as a checkbox
// rendering and a tick that lands reads as something completed. The coming
// stations ink last. They are on the frame from the first moment, quiet: seeing
// that there IS more road, before hearing what is on it, is what makes the first
// beat feel like progress rather than like a stopping point.

const BLOCK_W = Math.min(STAGE_W, 1680);

/** The rail sits at the middle of this band, with stations above and below. */
const BAND_H = 540;
/** Gap between the rail and the near edge of a chip — the stem's length. */
const STEM = 66;

type Item = {label: string; icon?: string};

type Step = {startMs: number; endMs: number; show: 'sofar' | 'ahead' | 'all'};

/** A drawn tick that scales in rather than fading. */
const Tick: React.FC<{theme: ResolvedTheme; colour: string; size: number; on: number}> = ({
  theme,
  colour,
  size,
  on,
}) => {
  const Check = iconFor('check');
  return (
    <div
      style={{
        width: size,
        height: size,
        flex: 'none',
        borderRadius: 999,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: withAlpha(colour, 0.14 * on),
        border: `1px solid ${withAlpha(on > 0.5 ? colour : theme.text, on > 0.5 ? 0.32 : 0.12)}`,
        transform: `scale(${0.72 + on * 0.28})`,
      }}
    >
      <div style={{opacity: on}}>
        <Check size={size * 0.54} strokeWidth={2.6} color={colour} />
      </div>
    </div>
  );
};

/**
 * One station: a dot on the rail, a stem, and a chip above or below it.
 *
 * Absolutely positioned against the band rather than laid out in a column,
 * because the dot has to land exactly on the rail — a station whose dot is a
 * few pixels off makes the rail look bent.
 */
const Station: React.FC<{
  theme: ResolvedTheme;
  item: Item;
  done: boolean;
  above: boolean;
  colour: string;
  on: number;
}> = ({theme, item, done, above, colour, on}) => {
  const Icon = iconFor(item.icon ?? 'arrow');
  const inked = on > 0.02;
  const glyph = 54;

  const chip = (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 18,
        padding: '22px 30px',
        borderRadius: 18,
        background: theme.surface,
        border: `1px solid ${inked ? withAlpha(colour, 0.18 + 0.12 * on) : theme.surfaceBorder}`,
        boxShadow: inked ? `0 ${3 + 7 * on}px ${12 + 18 * on}px ${withAlpha(theme.text, 0.06 * on)}` : 'none',
        // The chip travels toward the rail as it inks — a station arriving at
        // the road rather than appearing beside it.
        transform: `translateY(${(1 - on) * (above ? -16 : 16)}px)`,
        // Nearly twice the cell, which is what alternating sides buys. See the
        // header: this is why the labels fit at all.
        maxWidth: '170%',
        width: 'max-content',
      }}
    >
      {done ? (
        <Tick theme={theme} colour={colour} size={glyph} on={on} />
      ) : (
        <div
          style={{
            width: glyph,
            height: glyph,
            flex: 'none',
            borderRadius: glyph * 0.28,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: withAlpha(colour, 0.06 + 0.06 * on),
            border: `1px solid ${withAlpha(inked ? colour : theme.text, inked ? 0.18 : 0.1)}`,
          }}
        >
          <Icon size={glyph * 0.5} strokeWidth={1.8} color={inked ? colour : theme.textMuted} />
        </div>
      )}
      <div
        style={{
          fontFamily: theme.fontBody,
          fontSize: 32,
          fontWeight: 600,
          lineHeight: 1.22,
          color: inked ? theme.text : theme.textMuted,
          whiteSpace: 'nowrap',
        }}
      >
        {item.label}
      </div>
    </div>
  );

  return (
    <div style={{position: 'relative', flex: 1, height: BAND_H}}>
      {/* The chip, hung off the rail on its own side. */}
      <div
        style={{
          position: 'absolute',
          left: '50%',
          transform: 'translateX(-50%)',
          display: 'flex',
          justifyContent: 'center',
          ...(above ? {bottom: BAND_H / 2 + STEM} : {top: BAND_H / 2 + STEM}),
        }}
      >
        {chip}
      </div>

      {/* The stem. It grows out of the rail toward the chip, so an un-inked
          station is a dot on the road with nothing yet attached to it. */}
      <div
        style={{
          position: 'absolute',
          left: '50%',
          width: 1,
          height: STEM * Math.max(on, 0.001),
          background: withAlpha(colour, 0.35),
          transform: 'translateX(-0.5px)',
          ...(above ? {bottom: BAND_H / 2} : {top: BAND_H / 2}),
        }}
      />

      {/* The dot, exactly on the rail. */}
      <div
        style={{
          position: 'absolute',
          top: BAND_H / 2,
          left: '50%',
          width: 13,
          height: 13,
          marginTop: -6.5,
          marginLeft: -6.5,
          borderRadius: 999,
          background: inked ? colour : theme.bgTop,
          border: `2px solid ${inked ? colour : withAlpha(theme.text, 0.18)}`,
        }}
      />
    </div>
  );
};

export const ProgressScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const done = (Array.isArray(props.done) ? props.done : []) as Item[];
  const next = (Array.isArray(props.next) ? props.next : []) as Item[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const nowLabel = String(props.now ?? 'you are here');
  if (done.length === 0 || next.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;

  // Each side is keyed to the beat that inks it and nothing else needs to know
  // which beat is current: once a side is in it stays in, so a held closing beat
  // is the same frame as the one before it.
  const beatStart = (show: Step['show']) => {
    const s = steps.find((st) => st.show === show);
    return s ? s.startMs : Infinity;
  };
  const sinceSofar = ((nowMs - beatStart('sofar')) / 1000) * FPS;
  const sinceAhead = ((nowMs - beatStart('ahead')) / 1000) * FPS;
  const sofarIn = Number.isFinite(sinceSofar) ? sinceSofar : -1e6;
  const aheadIn = Number.isFinite(sinceAhead) ? sinceAhead : -1e6;

  const railLeft = interpolate(sofarIn, [0, 26], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const railRight = interpolate(aheadIn, [0, 32], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const pinIn = interpolate(sofarIn, [8, 24], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  // One hue for the whole frame, taken from the headline's own emphasis role.
  // Reaching for theme.accent instead put a blue rail under an ochre underline,
  // which reads as two accents rather than one composition.
  const colour = roleColour(theme, (props.emphasisRole as string | undefined) ?? 'quantity');

  // Stations alternate above and below along the WHOLE rail, counted across both
  // sides, so the alternation does not reset at the pin — a reset would put two
  // chips on the same side of the rail next to each other, which is the collision
  // the alternation exists to avoid.
  const station = (item: Item, i: number, isDone: boolean, globalIndex: number) => {
    const on = spring({
      frame: (isDone ? sofarIn : aheadIn) - 8 - i * 6,
      fps,
      config: {damping: 200, mass: 0.6},
    });
    return (
      <Station
        key={`${isDone ? 'd' : 'n'}${i}`}
        theme={theme}
        item={item}
        done={isDone}
        above={globalIndex % 2 === 0}
        colour={colour}
        on={on}
      />
    );
  };

  return (
    <Stage justify="center">
      <div style={{width: BLOCK_W, display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
        <SceneHeader
          theme={theme}
          title={String(props.title ?? '')}
          emphasis={props.emphasis as string | undefined}
          emphasisRole={props.emphasisRole as string | undefined}
          marginBottom={16}
        />

        <div style={{position: 'relative', width: '100%', height: BAND_H}}>
          {/* The rail, under everything, in two halves growing out from the pin.
              Two halves rather than one bar because the left one belongs to the
              first beat and the right one to the second, and a single bar could
              only tell one of those two stories. The road ahead is the same
              colour, dashed: the same road, not a different kind of thing. */}
          <div style={{position: 'absolute', top: BAND_H / 2, left: 0, right: 0, height: 2, marginTop: -1}}>
            <div
              style={{
                position: 'absolute',
                right: '50%',
                top: 0,
                height: '100%',
                width: `${railLeft * 50}%`,
                background: withAlpha(colour, 0.55),
              }}
            />
            {/* The road ahead exists before it is inked. Without this the
                coming stations sit as dots on nothing for the whole first beat,
                which reads as a rendering fault rather than as road not yet
                travelled. */}
            <div
              style={{
                position: 'absolute',
                left: '50%',
                top: 0,
                height: '100%',
                width: '50%',
                background: `repeating-linear-gradient(to right, ${withAlpha(theme.text, 0.1)} 0 13px, transparent 13px 25px)`,
              }}
            />
            <div
              style={{
                position: 'absolute',
                left: '50%',
                top: 0,
                height: '100%',
                width: `${railRight * 50}%`,
                background: `repeating-linear-gradient(to right, ${withAlpha(colour, 0.42)} 0 13px, transparent 13px 25px)`,
              }}
            />
          </div>

          <div style={{display: 'flex', alignItems: 'flex-start', width: '100%', height: BAND_H}}>
            {done.map((it, i) => station(it, i, true, i))}

            {/* NOW. The only element that names the present, and the only one
                set in the display face — it is a label on the road, not a
                station on it, so it gets no chip and no stem. */}
            <div style={{position: 'relative', flex: 'none', width: 132, height: BAND_H}}>
              {/* The label clears the rail rather than straddling it: set on the
                  line it reads as a caption the road runs through. The dot is
                  positioned independently so it lands exactly ON the rail — the
                  first cut centred label and dot together as one column, which
                  put the label above the line and the dot below it. */}
              <div
                style={{
                  position: 'absolute',
                  bottom: BAND_H / 2 + 20,
                  left: '50%',
                  transform: `translateX(-50%) scale(${0.9 + pinIn * 0.1})`,
                  opacity: pinIn,
                  fontFamily: theme.fontDisplay,
                  fontSize: 19,
                  fontWeight: 700,
                  letterSpacing: 1.6,
                  textTransform: 'uppercase',
                  color: colour,
                  whiteSpace: 'nowrap',
                }}
              >
                {nowLabel}
              </div>
              <div
                style={{
                  position: 'absolute',
                  top: BAND_H / 2,
                  left: '50%',
                  width: 22,
                  height: 22,
                  marginTop: -11,
                  marginLeft: -11,
                  borderRadius: 999,
                  background: colour,
                  opacity: pinIn,
                  transform: `scale(${0.85 + pinIn * 0.15})`,
                  boxShadow: `0 0 0 ${9 * pinIn}px ${withAlpha(colour, 0.13)}`,
                }}
              />
            </div>

            {next.map((it, i) => station(it, i, false, done.length + i))}
          </div>

          {/* The two side labels, at the outer ends of the rail rather than over
              a column: they name the two directions of the road, and putting
              them over the stations would make them column headings again. */}
          <CardLabel
            theme={theme}
            lit={sofarIn >= 0}
            style={{position: 'absolute', left: 0, top: BAND_H / 2 + 22}}
          >
            so far
          </CardLabel>
          <CardLabel
            theme={theme}
            lit={aheadIn >= 0}
            style={{position: 'absolute', right: 0, top: BAND_H / 2 + 22}}
          >
            next
          </CardLabel>
        </div>
      </div>
    </Stage>
  );
};
