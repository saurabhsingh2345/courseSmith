import {AbsoluteFill, interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';

// WaypointScene is a chapter opening that also says where you are in the piece.
//
// Two decisions, both aimed at the same failure: a long video where the viewer
// silently loses the thread and leaves.
//
// THE COMPOSITION IS ASYMMETRIC. A centred headline with air around it is a
// slide, and ten of them in half an hour is a slideshow. Here the type sits on a
// hard left axis with the ordinal above and a rule under it — an editorial page
// rather than a title card. It is the same content in a register that can stand
// being seen ten times.
//
// THE SPINE IS PERIPHERAL BY CONSTRUCTION. Every chapter of the piece runs down
// the right edge, but at an ink level where only the lit row is properly readable.
// That is deliberate: at full strength it becomes a contents page competing with
// the headline, and the point is not for anyone to READ the list. The point is
// that the eye takes in a texture with a bright mark two-thirds of the way down
// and knows, without deciding to look, that most of the video is behind them.
// A viewer who can see the end coming stays for it.
//
// Nothing scales here. See the note in SessionScene: continuous scale on text
// glitters, and this card is mostly type.

type Step = {
  startMs: number;
  endMs: number;
  show: 'arrive' | 'promise' | 'spine';
  promise?: boolean;
  spine?: boolean;
};

type Props = {
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: {
    ordinal?: string;
    line?: string;
    promise?: string;
    stops?: string[];
    at?: number;
    steps?: Step[];
  };
};

const AXIS_X = 190;
const MEASURE = 980;
const SPINE_X = 1400;

/** Title size, stepped by length against the measure. */
const sizeFor = (chars: number): number => {
  if (chars <= 14) return 132;
  if (chars <= 26) return 106;
  return 84;
};

export const WaypointScene = ({theme, sceneStartMs, props}: Props) => {
  const frame = useCurrentFrame();
  const ms = (frame / FPS) * 1000 + sceneStartMs;
  const steps = props.steps ?? [];
  const stops = props.stops ?? [];
  const at = props.at ?? 0;
  const line = props.line ?? '';

  const step = steps.find((s) => ms >= s.startMs && ms < s.endMs) ?? steps[steps.length - 1];
  const sceneStart = steps[0]?.startMs ?? sceneStartMs;

  const arrive = (from: number, dur = 440) =>
    interpolate(ms, [from, from + dur], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});

  const inA = arrive(sceneStart);
  const promiseStep = steps.find((s) => s.show === 'promise');
  const spineStep = steps.find((s) => s.show === 'spine');
  const promiseIn = step?.promise && promiseStep ? arrive(promiseStep.startMs) : 0;
  // The spine comes up with the card when no beat is dedicated to it — a chapter
  // card that never shows the arc has given up the only thing this template adds.
  const spineIn = spineStep ? (step?.spine ? arrive(spineStep.startMs) : 0) : inA;

  const size = sizeFor(line.length);

  return (
    <AbsoluteFill style={{background: theme.surface}}>
      {/* The chapter, on the axis. */}
      <div
        style={{
          position: 'absolute',
          left: AXIS_X,
          top: 300,
          width: MEASURE,
        }}
      >
        {props.ordinal ? (
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 24,
              letterSpacing: 5,
              color: theme.accentText,
              opacity: inA,
              marginBottom: 26,
            }}
          >
            {props.ordinal}
          </div>
        ) : null}

        <div
          style={{
            fontFamily: theme.fontSerif,
            fontSize: size,
            lineHeight: 1.05,
            letterSpacing: -1.8,
            color: theme.text,
            opacity: inA,
            transform: `translateY(${Math.round((1 - inA) * 14)}px)`,
          }}
        >
          {line}
        </div>

        <div
          style={{
            height: 1,
            background: withAlpha(theme.text, 0.22),
            marginTop: 40,
            // The rule draws itself in from the axis, which is the one piece of
            // motion the card needs: it reads as a page being set.
            width: `${Math.round(inA * 100)}%`,
          }}
        />

        {props.promise ? (
          <div
            style={{
              fontFamily: theme.fontBody,
              fontSize: 34,
              lineHeight: 1.4,
              color: theme.textMuted,
              marginTop: 34,
              opacity: promiseIn,
              transform: `translateY(${Math.round((1 - promiseIn) * 10)}px)`,
            }}
          >
            {props.promise}
          </div>
        ) : null}
      </div>

      {/* The arc. */}
      <div
        style={{
          position: 'absolute',
          left: SPINE_X,
          top: 250,
          width: 400,
          opacity: spineIn,
        }}
      >
        {stops.map((s, i) => {
          const done = i < at;
          const here = i === at;
          // Rows stagger in, so the spine assembles downward rather than
          // appearing as a block of small type.
          const rowIn = interpolate(
            ms,
            [(spineStep?.startMs ?? sceneStart) + i * 55, (spineStep?.startMs ?? sceneStart) + i * 55 + 260],
            [0, 1],
            {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
          );
          return (
            <div
              key={i}
              style={{
                display: 'flex',
                alignItems: 'baseline',
                gap: 14,
                height: 46,
                paddingLeft: here ? 14 : 0,
                borderLeft: `2px solid ${here ? theme.accent : 'transparent'}`,
                opacity: rowIn * (here ? 1 : done ? 0.42 : 0.2),
              }}
            >
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 17,
                  color: here ? theme.accentText : theme.text,
                  width: 20,
                }}
              >
                {done ? '✓' : here ? '▸' : '·'}
              </span>
              <span
                style={{
                  fontFamily: here ? theme.fontBody : theme.fontMono,
                  fontSize: here ? 26 : 20,
                  fontWeight: here ? 700 : 400,
                  color: theme.text,
                }}
              >
                {s}
              </span>
            </div>
          );
        })}
      </div>
    </AbsoluteFill>
  );
};
