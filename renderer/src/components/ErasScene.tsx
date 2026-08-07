import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W, STAGE_H} from './Stage';
import {SceneHeader} from './SceneHeader';

// ErasScene: history as a band you can see the whole of at once.
//
// Every segment is on screen from the first frame, dim, and that is the
// argument. A history told as one card at a time is a list of facts; a history
// told as a band with one segment lit is a fact IN ITS PLACE — the viewer can
// see how much came before it and how much is still to come, and the shape of
// the span does the work that "and then, later" has to do in prose.
//
// Time runs left to right and the present is a real column at the right-hand
// end rather than an implied edge, so the band always ends somewhere the
// viewer actually lives. That column stays dark until the closing beat, which
// is what makes arriving at it feel like arriving.
//
// The arcs are the point of the template and they are drawn above the band, in
// their own reserved air, hopping segment to segment with the hand-off written
// at each apex. Dashed rather than solid, because a hand-off is an inheritance
// and not a pipe: the transistor did not send anything to the integrated
// circuit, it left it something. They enter left to right so the thread reads
// as a direction, and once drawn they stay, because the closer's claim is that
// the whole chain led here.
//
// The dates are mono and small, sitting above their segments like a ruler; the
// era names are display type inside the cards; the defining artifact only
// appears on the era in focus, so the frame never carries more than one line of
// prose at a time. One glow at most: the card lifting right now.

const BODY_W = Math.min(STAGE_W, 1620);
const NOW_W = 168;
const SEG_GAP = 14;
const TRACK_W = BODY_W - NOW_W - SEG_GAP;
const ARC_H = 152;
const WHEN_H = 42;
const CARD_H = 252;

type Era = {label: string; when: string; mark: string; carry: string};
type Thread = {from: number; to: number; carry: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'band' | 'era' | 'thread' | 'now';
  at?: number;
  lit: number[];
};

export const ErasScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const eras = (Array.isArray(props.eras) ? props.eras : []) as Era[];
  const threads = (Array.isArray(props.threads) ? props.threads : []) as Thread[];
  const carryNow = String(props.carryNow ?? '');
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (eras.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const segW = (TRACK_W - SEG_GAP * (eras.length - 1)) / eras.length;
  const cx = (i: number) => i * (segW + SEG_GAP) + segW / 2;
  const todayX = TRACK_W + SEG_GAP + NOW_W / 2;

  const lit = new Set(Array.isArray(step.lit) ? step.lit : []);
  const arrived = step.show === 'now';
  const focus = step.show === 'era' || arrived ? (step.at ?? -1) : -1;
  // Once the thread is drawn it stays: the closer's claim is that the whole
  // chain led here, so unpicking it would contradict the narration.
  const threaded = steps.slice(0, idx + 1).some((s) => s.show === 'thread' || s.show === 'now');
  const threading = step.show === 'thread';

  const bodyH = Math.min(STAGE_H - 170, ARC_H + WHEN_H + CARD_H + 22);

  const arcPath = (from: number, to: number): string => {
    const x1 = cx(from);
    const x2 = to >= eras.length ? todayX : cx(to);
    const hop = Math.min(ARC_H - 46, 96);
    const bend = (x2 - x1) * 0.42;
    return `M ${x1} ${ARC_H} C ${x1 + bend} ${ARC_H - hop}, ${x2 - bend} ${ARC_H - hop}, ${x2} ${ARC_H}`;
  };

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={20}
      />

      <div style={{width: BODY_W, height: bodyH, position: 'relative'}}>
        {/* The thread: hand-offs drawn as inheritances, not pipes. */}
        <svg
          width={BODY_W}
          height={ARC_H}
          viewBox={`0 0 ${BODY_W} ${ARC_H}`}
          style={{position: 'absolute', left: 0, top: 0, overflow: 'visible'}}
        >
          {threads.map((t, i) => {
            const draw = threading
              ? interpolate(sinceStep, [4 + i * 8, 24 + i * 8], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                })
              : threaded
                ? 1
                : 0;
            if (draw <= 0) return null;
            return (
              <path
                key={i}
                d={arcPath(t.from, t.to)}
                fill="none"
                stroke={withAlpha(theme.accent, 0.85)}
                strokeWidth={2.5}
                strokeLinecap="round"
                strokeDasharray="9 8"
                pathLength={1}
                // A second dash pattern would fight the first, so the reveal is
                // done by clipping the stroke's own length.
                style={{clipPath: `inset(0 ${(1 - draw) * 100}% 0 0)`}}
              />
            );
          })}
          {arrived && carryNow ? (
            <path
              d={arcPath(eras.length - 1, eras.length)}
              fill="none"
              stroke={theme.accent}
              strokeWidth={3}
              strokeLinecap="round"
              strokeDasharray="9 8"
              style={{
                clipPath: `inset(0 ${(1 -
                  interpolate(sinceStep, [4, 26], [0, 1], {
                    extrapolateLeft: 'clamp',
                    extrapolateRight: 'clamp',
                  })) *
                  100}% 0 0)`,
              }}
            />
          ) : null}
        </svg>

        {/* The hand-off text, set at each apex. */}
        {threads.map((t, i) => {
          const shown = threading
            ? interpolate(sinceStep, [12 + i * 8, 28 + i * 8], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              })
            : threaded
              ? 1
              : 0;
          if (shown <= 0) return null;
          const mid = (cx(t.from) + cx(t.to)) / 2;
          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: mid - (segW + SEG_GAP) / 2,
                top: 4,
                width: segW + SEG_GAP,
                textAlign: 'center',
                fontFamily: theme.fontBody,
                fontSize: 17,
                lineHeight: 1.2,
                color: theme.accentText,
                opacity: shown,
              }}
            >
              {t.carry}
            </div>
          );
        })}
        {arrived && carryNow ? (
          <div
            style={{
              position: 'absolute',
              left: (cx(eras.length - 1) + todayX) / 2 - (segW + SEG_GAP) / 2,
              top: 4,
              width: segW + SEG_GAP,
              textAlign: 'center',
              fontFamily: theme.fontBody,
              fontSize: 17,
              lineHeight: 1.2,
              color: theme.accentText,
              opacity: interpolate(sinceStep, [10, 26], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            {carryNow}
          </div>
        ) : null}

        {/* The dates: a ruler above the band. */}
        {eras.map((e, i) => (
          <div
            key={i}
            style={{
              position: 'absolute',
              left: i * (segW + SEG_GAP),
              top: ARC_H,
              width: segW,
              height: WHEN_H,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontFamily: theme.fontMono,
              fontSize: 21,
              letterSpacing: 1.4,
              color: i === focus ? theme.accentText : lit.has(i) ? theme.text : theme.textMuted,
            }}
          >
            {e.when}
          </div>
        ))}
        <div
          style={{
            position: 'absolute',
            left: TRACK_W + SEG_GAP,
            top: ARC_H,
            width: NOW_W,
            height: WHEN_H,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontFamily: theme.fontMono,
            fontSize: 21,
            letterSpacing: 1.4,
            color: arrived ? theme.accentText : withAlpha(theme.textMuted, 0.5),
          }}
        >
          today
        </div>

        {/* The band. Every segment present, one of them lit. */}
        {eras.map((e, i) => {
          const focused = i === focus;
          const seen = lit.has(i);
          const rise = focused
            ? spring({frame: sinceStep - 3, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 26})
            : 0;
          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: i * (segW + SEG_GAP),
                top: ARC_H + WHEN_H,
                width: segW,
                height: CARD_H,
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'flex-start',
                padding: 22,
                borderRadius: 14,
                background: withAlpha(theme.surface, focused ? 0.98 : seen ? 0.7 : 0.42),
                border: `1px solid ${focused ? withAlpha(theme.accent, 0.7) : theme.surfaceBorder}`,
                borderTop: `4px solid ${focused ? theme.accent : seen ? withAlpha(theme.accent, 0.4) : theme.surfaceBorder}`,
                transform: `translateY(${-8 * rise}px) scale(${1 + 0.03 * rise})`,
                // The one glow: the era lifting right now.
                boxShadow: focused && rise < 0.999 ? `0 0 44px ${withAlpha(theme.accent, 0.28)}` : undefined,
              }}
            >
              <div
                style={{
                  fontFamily: theme.fontDisplay,
                  fontSize: eras.length > 4 ? 27 : 32,
                  fontWeight: 700,
                  letterSpacing: -0.6,
                  lineHeight: 1.1,
                  color: focused || seen ? theme.text : theme.textMuted,
                }}
              >
                {e.label}
              </div>
              <div
                style={{
                  marginTop: 16,
                  fontFamily: theme.fontBody,
                  fontSize: eras.length > 4 ? 19 : 22,
                  lineHeight: 1.32,
                  color: theme.textMuted,
                  opacity: focused ? rise : seen ? 0.55 : 0,
                }}
              >
                {e.mark}
              </div>
            </div>
          );
        })}

        {/* Today: a real column, dark until the clip arrives at it. */}
        <div
          style={{
            position: 'absolute',
            left: TRACK_W + SEG_GAP,
            top: ARC_H + WHEN_H,
            width: NOW_W,
            height: CARD_H,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderRadius: 14,
            background: arrived ? withAlpha(theme.accent, 0.14) : withAlpha(theme.surface, 0.28),
            border: `1px dashed ${arrived ? theme.accent : theme.surfaceBorder}`,
            fontFamily: theme.fontDisplay,
            fontSize: 26,
            fontWeight: 700,
            color: arrived ? theme.accentText : withAlpha(theme.textMuted, 0.5),
            opacity: arrived
              ? interpolate(sinceStep, [0, 18], [0.5, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                })
              : 1,
          }}
        >
          you
        </div>
      </div>
    </Stage>
  );
};
