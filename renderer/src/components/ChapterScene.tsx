import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {iconFor} from './icons';

// ChapterScene is the break between two stretches of a course.
//
// It is furniture rather than a diagram, and the composition says so. The
// ordinal is enormous — 340px of outlined display type, the loudest thing on the
// frame by a wide margin — and everything else is small and set against it. That
// is the whole trick: two words and a number, read in half a second, on a frame
// with almost nothing in it.
//
// Three decisions carry it.
//
// The path is standing furniture. It is drawn once on the first beat and never
// redrawn; every beat after that only moves light along it. A break whose
// picture rebuilds itself is a break that takes as long to read as it does to
// watch, and this clip's whole job is to be over quickly.
//
// The ordinal is a cut-out, not a fill. It is drawn in the stage's own colour
// with an accent stroke around it, so it reads as a hole punched in the frame
// rather than as a very large word. A solid 340px numeral in the accent would be
// the only thing anybody saw; hollowed out it holds the corner of the eye while
// the section title stays the thing being read.
//
// And the stop the viewer is AT gets a halo that breathes, while everything
// behind it is a small filled disc with a tick and everything ahead is an open
// ring. Three states, three shapes — so the path is legible at a glance without
// reading a single label, which matters because nobody reads the labels on a
// twenty-second card.

const PATH_W = Math.min(STAGE_W, 1500);
// The path sits low and wide. A journey drawn across the middle of the frame
// competes with the ordinal; drawn along the bottom it reads as a footer, which
// is what it is.
const NODE_R = 15;
const HERE_R = 30;

type Stop = {label: string; icon?: string; note?: string; state: 'done' | 'here' | 'ahead'};
type Step = {startMs: number; endMs: number; show: 'path' | 'done' | 'here'; at?: number};

const StopIcon: React.FC<{name?: string; color: string; size: number}> = ({name, color, size}) => {
  const Icon = iconFor(name);
  return (
    <div style={{color, lineHeight: 0, display: 'flex'}}>
      <Icon size={size} strokeWidth={2.1} />
    </div>
  );
};

export const ChapterScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const path = String(props.path ?? '');
  const stops = (Array.isArray(props.stops) ? props.stops : []) as Stop[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const at = Number(props.at ?? 0);
  const ordinal = Number(props.ordinal ?? at + 1);
  const total = Number(props.total ?? stops.length);
  if (stops.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;
  const sinceStart = ((nowMs - steps[0].startMs) / 1000) * FPS;

  // Which stop the light is on. On a `done` beat it is the one being recalled;
  // otherwise it is where the viewer actually stands.
  const recalled = step.show === 'done' ? (step.at ?? 0) : -1;
  const opening = step.show === 'here';

  const enter = spring({frame: sinceStart, fps, config: {damping: 200, mass: 0.7}, durationInFrames: 22});
  // The marker travels the path on the opening beat rather than being there from
  // frame one. A position that was always there is a fact; a position arrived at
  // is a journey, and the second one is what this card is claiming.
  const travel = interpolate(sinceStart, [8, 40], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // A slow breath under the current stop. It is the one thing on the frame that
  // never settles, so the eye keeps coming back to where the viewer is.
  const breath = 0.5 + 0.5 * Math.sin((frame / fps) * 1.7);

  const gap = stops.length > 1 ? PATH_W / (stops.length - 1) : 0;
  const xOf = (i: number) => (stops.length > 1 ? i * gap : PATH_W / 2);
  const hereX = xOf(at);
  // How far along the whole path the traversed line has been drawn.
  const drawn = hereX * travel;

  const here = stops[at];
  const focus = recalled >= 0 ? stops[recalled] : here;
  const ahead = at + 1 < stops.length ? stops[at + 1] : undefined;

  return (
    <Stage justify="center">
      <div style={{width: PATH_W, position: 'relative'}}>
        {/* The ordinal, hollowed out of the stage. It is the largest object on
            the frame and it carries no information the title does not — which is
            the point: it is a place marker, and place markers are read at a
            glance or not at all. */}
        <div
          style={{
            position: 'absolute',
            right: 4,
            // Set beside the title rather than above it. The first pass hung it
            // off the top corner and left a hole the size of a quarter of the
            // frame between the two — a card this sparse has to be composed,
            // because there is nothing else in it to distract from a gap.
            top: -46,
            fontFamily: theme.fontDisplay,
            fontSize: 400,
            fontWeight: 800,
            lineHeight: 0.82,
            letterSpacing: -18,
            color: withAlpha(theme.accentQuantity, 0.1),
            WebkitTextStroke: `2px ${withAlpha(theme.accentQuantity, opening ? 0.55 : 0.28)}`,
            fontVariantNumeric: 'tabular-nums',
            opacity: enter,
            transform: `translateY(${(1 - enter) * 26}px)`,
          }}
        >
          {String(ordinal).padStart(2, '0')}
        </div>

        {/* The eyebrow: what run this is a break in. */}
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 19,
            letterSpacing: 5,
            textTransform: 'uppercase',
            color: theme.textMuted,
            opacity: enter,
            display: 'flex',
            alignItems: 'center',
            gap: 16,
          }}
        >
          <span
            style={{
              width: 46,
              height: 2,
              background: theme.accentQuantity,
              display: 'inline-block',
              transform: `scaleX(${enter})`,
              transformOrigin: 'left center',
            }}
          />
          {path}
        </div>

        {/* The subject: whichever stop the beat is about. On a look-back beat the
            card names the part being recalled, so the words and the light on the
            path are always about the same thing. */}
        <div style={{marginTop: 26, minHeight: 232, maxWidth: PATH_W * 0.66}}>
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 17,
              letterSpacing: 3.4,
              textTransform: 'uppercase',
              color: recalled >= 0 ? theme.textMuted : theme.accentText,
              opacity: interpolate(sinceStep, [0, 10], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            {recalled >= 0 ? 'Already behind you' : 'Starting now'}
          </div>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 22,
              marginTop: 14,
              opacity: interpolate(sinceStep, [2, 14], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
              transform: `translateY(${interpolate(sinceStep, [2, 16], [14, 0], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              })}px)`,
            }}
          >
            <div
              style={{
                width: 74,
                height: 74,
                borderRadius: 20,
                flexShrink: 0,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: withAlpha(recalled >= 0 ? theme.text : theme.accentQuantity, 0.12),
                border: `1px solid ${withAlpha(recalled >= 0 ? theme.text : theme.accentQuantity, 0.3)}`,
              }}
            >
              <StopIcon
                name={focus?.icon}
                color={recalled >= 0 ? theme.textMuted : theme.accentQuantity}
                size={36}
              />
            </div>
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 82,
                fontWeight: 800,
                letterSpacing: -2,
                lineHeight: 1.05,
                color: theme.text,
              }}
            >
              {focus?.label}
            </div>
          </div>
          {focus?.note ? (
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 29,
                lineHeight: 1.35,
                color: theme.textMuted,
                marginTop: 16,
                opacity: interpolate(sinceStep, [10, 24], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                }),
              }}
            >
              {focus.note}
            </div>
          ) : null}
        </div>

        {/* The path. Drawn once, then only lit. */}
        <div style={{position: 'relative', height: 120, marginTop: 58}}>
          <svg
            width={PATH_W}
            height={80}
            style={{position: 'absolute', left: 0, top: 0, overflow: 'visible'}}
          >
            {/* Everything still ahead: a dashed run the eye reads as unwalked. */}
            <line
              x1={0}
              y1={40}
              x2={PATH_W}
              y2={40}
              stroke={withAlpha(theme.text, 0.16)}
              strokeWidth={3}
              strokeDasharray="2 12"
              strokeLinecap="round"
              opacity={enter}
            />
            {/* Everything walked, drawn on. */}
            <line
              x1={0}
              y1={40}
              x2={Math.max(drawn, 1)}
              y2={40}
              stroke={theme.accentQuantity}
              strokeWidth={3}
              strokeLinecap="round"
              opacity={enter * 0.9}
            />
          </svg>

          {stops.map((s, i) => {
            const x = xOf(i);
            const isHere = i === at;
            const isRecalled = i === recalled;
            // A stop only lights once the drawn line has reached it, so the
            // marker and the ticks arrive together rather than the row being
            // finished before the line gets there.
            const reached = x <= drawn + 1;
            const c =
              s.state === 'ahead'
                ? withAlpha(theme.text, 0.22)
                : isHere
                  ? theme.accentQuantity
                  : withAlpha(theme.accentQuantity, isRecalled ? 0.95 : 0.5);
            const r = isHere ? HERE_R : NODE_R;
            return (
              <div key={i} style={{position: 'absolute', left: x, top: 40}}>
                {/* The halo under where the viewer is, breathing. */}
                {isHere ? (
                  <div
                    style={{
                      position: 'absolute',
                      left: -HERE_R * 3,
                      top: -HERE_R * 3,
                      width: HERE_R * 6,
                      height: HERE_R * 6,
                      borderRadius: '50%',
                      background: `radial-gradient(circle, ${withAlpha(
                        theme.accentQuantity,
                        0.3 * (0.5 + breath * 0.5) * travel,
                      )} 0%, transparent 68%)`,
                    }}
                  />
                ) : null}
                <div
                  style={{
                    position: 'absolute',
                    left: -r,
                    top: -r,
                    width: r * 2,
                    height: r * 2,
                    borderRadius: '50%',
                    boxSizing: 'border-box',
                    // Three shapes for three states, so the path reads without
                    // anybody reading a label: behind is solid, here is a ring
                    // with a bright core, ahead is hollow.
                    background: s.state === 'done' && reached ? c : theme.bgBottom,
                    border: `${isHere ? 4 : 3}px solid ${reached || s.state !== 'ahead' ? c : withAlpha(theme.text, 0.18)}`,
                    transform: `scale(${
                      isRecalled
                        ? 1 + 0.28 * interpolate(sinceStep, [0, 12], [0, 1], {
                            extrapolateLeft: 'clamp',
                            extrapolateRight: 'clamp',
                          })
                        : 1
                    })`,
                    opacity: enter,
                  }}
                />
                {isHere ? (
                  <div
                    style={{
                      position: 'absolute',
                      left: -8,
                      top: -8,
                      width: 16,
                      height: 16,
                      borderRadius: '50%',
                      background: theme.accentQuantity,
                      opacity: travel,
                    }}
                  />
                ) : null}
                {/* A tick inside the ones already behind. The filled disc alone
                    says "not here"; the tick says "done", and those are
                    different claims — this card's whole job is the second one. */}
                {s.state === 'done' && reached ? (
                  <svg
                    width={18}
                    height={18}
                    viewBox="0 0 18 18"
                    style={{position: 'absolute', left: -9, top: -9}}
                  >
                    <path
                      d="M 4 9.4 L 7.6 13 L 14 5.6"
                      fill="none"
                      stroke={theme.bgBottom}
                      strokeWidth={2.6}
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                ) : null}
                {/* Its name, under it. Only the walked ones and the next one are
                    legible; the rest are present and faint, which is what makes
                    the path feel longer than the part being talked about. */}
                <div
                  style={{
                    position: 'absolute',
                    left: 0,
                    // Off the largest node, not off this one, so every name in
                    // the row sits on one baseline. Measured from its own radius
                    // the current stop's label dropped 15px below its
                    // neighbours, which reads as a typesetting mistake rather
                    // than as emphasis.
                    top: HERE_R + 20,
                    transform: 'translateX(-50%)',
                    fontFamily: theme.fontMono,
                    fontSize: 16,
                    letterSpacing: 1.6,
                    textTransform: 'uppercase',
                    whiteSpace: 'nowrap',
                    color: isHere || isRecalled ? theme.text : theme.textMuted,
                    opacity: enter * (s.state === 'ahead' ? 0.4 : isHere || isRecalled ? 1 : 0.62),
                  }}
                >
                  {s.label}
                </div>
              </div>
            );
          })}

          {/* The count, pinned to the right of the path. Two numbers, and they
              are the whole reason somebody watches a card like this. */}
          <div
            style={{
              position: 'absolute',
              right: 0,
              top: -46,
              fontFamily: theme.fontMono,
              fontSize: 18,
              letterSpacing: 2.6,
              color: theme.textMuted,
              opacity: enter,
            }}
          >
            <span style={{color: theme.accentText, fontWeight: 700}}>{String(ordinal).padStart(2, '0')}</span>
            {` / ${String(total).padStart(2, '0')}`}
          </div>
        </div>

        {/* What opens after this one, on the handover beat only. The last thing
            said should point forwards, so the last thing drawn does too. */}
        {opening && ahead ? (
          <div
            style={{
              position: 'absolute',
              left: 0,
              bottom: -74,
              display: 'flex',
              alignItems: 'center',
              gap: 12,
              fontFamily: theme.fontMono,
              fontSize: 17,
              letterSpacing: 2.6,
              textTransform: 'uppercase',
              color: theme.textMuted,
              opacity: interpolate(sinceStep, [16, 32], [0, 0.85], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            <StopIcon name={ahead.icon} color={theme.textMuted} size={18} />
            {`then · ${ahead.label}`}
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
