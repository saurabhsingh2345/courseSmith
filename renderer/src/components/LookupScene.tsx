import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// LookupScene: one question, walked along a chain of stations.
//
// The composition is a left-to-right path because that is the only reading
// order the eye brings for free, and the claim of the clip is entirely about
// order: this station, then that one, then the last. The asker sits at the
// hard left and never moves — it is the origin AND the destination, which is
// what makes the return arc read as a return rather than as a sixth station.
//
// The moving thing is the QUESTION, not the data. A beginner's wrong model is
// that the answer was sitting somewhere and got fetched, so the card carries
// the key in mono, and what changes about it as it travels is that it collects
// stamps — one per station that has spoken. By the last hop the card is
// visibly heavier than it was, which is the picture of an answer being
// assembled rather than retrieved.
//
// Two arcs, deliberately on opposite sides of the rail. The hit returns
// UNDERNEATH, solid and lit, because it retraces ground already walked. The
// cache shortcut goes OVER the top, dashed, because it is the road not taken —
// it exists only because the long one was, and putting it on the same side
// would read as one more leg of the same journey. Both are drawn by animating
// a normalised dash along the path, so the arc arrives as a stroke rather than
// appearing whole.
//
// One glow maximum: the answer chip landing in the asker on the hit beat.

const BOARD_W = Math.min(STAGE_W, 1580);
const BOARD_H = 500;
const ASKER_W = 240;
const ASKER_GAP = 76;
const STATION_GAP = 26;
const STATION_TOP = 178;
const STATION_H = 180;
const CARD_TOP = 74;
const CARD_H = 76;

type Hop = {where: string; gives: string; miss: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'ask' | 'hop' | 'hit' | 'cache';
  at?: number;
  visited: number[];
  answered: boolean;
  cached: boolean;
};

export const LookupScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const keyText = String(props.key ?? '');
  const answer = String(props.answer ?? '');
  const hops = (Array.isArray(props.hops) ? props.hops : []) as Hop[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (hops.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const travel = spring({frame: sinceStep, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 28});
  const land = spring({frame: sinceStep - 8, fps, config: {damping: 12, mass: 0.5}, durationInFrames: 24});

  const laneX = ASKER_W + ASKER_GAP;
  const laneW = BOARD_W - laneX;
  const stationW = Math.min(300, Math.floor((laneW - STATION_GAP * (hops.length - 1)) / hops.length));
  const stationCx = (i: number): number => laneX + i * (stationW + STATION_GAP) + stationW / 2;
  const askerCx = ASKER_W / 2;
  const lastCx = stationCx(hops.length - 1);
  const stationCy = STATION_TOP + STATION_H / 2;

  const visited = new Set(step.visited ?? []);
  const current = step.show === 'hop' ? step.at ?? -1 : -1;

  // Where the card is standing, and where it came from, so the travel between
  // two stations is a slide rather than a jump.
  const restingCx = current >= 0 ? stationCx(current) : step.answered ? askerCx : visited.size > 0 ? stationCx(visited.size - 1) : askerCx;
  const fromCx = current > 0 ? stationCx(current - 1) : current === 0 ? askerCx : restingCx;
  const cardCx = fromCx + (restingCx - fromCx) * travel;

  const arcProgress = (on: boolean): number =>
    on ? interpolate(sinceStep, [4, 30], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'}) : 1;
  const hitDraw = step.show === 'hit' ? arcProgress(true) : step.answered ? 1 : 0;
  const cacheDraw = step.show === 'cache' ? arcProgress(true) : step.cached ? 1 : 0;

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

      <div style={{position: 'relative', width: BOARD_W, height: BOARD_H}}>
        {/* The rail the stations sit on. */}
        <div
          style={{
            position: 'absolute',
            left: askerCx,
            top: stationCy - 1,
            width: lastCx - askerCx,
            height: 2,
            background: withAlpha(theme.line, 0.3),
          }}
        />

        {/* The two arcs: the return underneath, the shortcut over the top. */}
        <svg width={BOARD_W} height={BOARD_H} style={{position: 'absolute', left: 0, top: 0, overflow: 'visible'}}>
          {hitDraw > 0 ? (
            <>
              <path
                d={`M ${lastCx} ${STATION_TOP + STATION_H} Q ${(lastCx + askerCx) / 2} ${BOARD_H + 40} ${askerCx} ${STATION_TOP + STATION_H}`}
                fill="none"
                stroke={theme.accentQuantity}
                strokeWidth={3.5}
                strokeLinecap="round"
                pathLength={1}
                strokeDasharray={`${hitDraw} 1`}
              />
              <circle
                cx={askerCx}
                cy={STATION_TOP + STATION_H}
                r={6}
                fill={theme.accentQuantity}
                opacity={interpolate(hitDraw, [0.85, 1], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})}
              />
            </>
          ) : null}
          {cacheDraw > 0 ? (
            <path
              d={`M ${askerCx} ${STATION_TOP} Q ${(lastCx + askerCx) / 2} ${-90} ${lastCx} ${STATION_TOP}`}
              fill="none"
              stroke={withAlpha(theme.accent, 0.85)}
              strokeWidth={3}
              strokeLinecap="round"
              strokeDasharray={`${cacheDraw * 0.02} 0.014`}
              pathLength={1}
            />
          ) : null}
        </svg>

        {/* The asker: origin, destination, and where the answer lands. */}
        <div
          style={{
            position: 'absolute',
            left: 0,
            top: STATION_TOP,
            width: ASKER_W,
            height: STATION_H,
            borderRadius: 16,
            border: `2px solid ${step.answered ? theme.accentQuantity : theme.surfaceBorder}`,
            background: withAlpha(theme.surface, 0.7),
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 12,
            padding: 16,
          }}
        >
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 14,
              letterSpacing: 3,
              textTransform: 'uppercase',
              color: theme.textMuted,
            }}
          >
            you
          </div>
          {step.answered ? (
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 26,
                fontWeight: 700,
                color: theme.ink,
                background: theme.accentQuantity,
                borderRadius: 10,
                paddingInline: 14,
                paddingBlock: 8,
                textAlign: 'center',
                transform: `scale(${0.85 + 0.15 * (step.show === 'hit' ? land : 1)})`,
                // The one glow: the answer arriving home.
                boxShadow: step.show === 'hit' ? `0 0 30px ${withAlpha(theme.accentQuantity, 0.55)}` : undefined,
              }}
            >
              {answer}
            </div>
          ) : (
            <div style={{fontFamily: theme.fontBody, fontSize: 22, color: theme.textMuted, textAlign: 'center'}}>
              asking
            </div>
          )}
        </div>

        {/* The stations. */}
        {hops.map((hop, i) => {
          const stamped = visited.has(i);
          const isCurrent = i === current;
          const x = laneX + i * (stationW + STATION_GAP);
          const pop = isCurrent ? land : 1;
          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: x,
                top: STATION_TOP,
                width: stationW,
                height: STATION_H,
                borderRadius: 16,
                border: `2px solid ${isCurrent ? theme.accent : stamped ? withAlpha(theme.accent, 0.4) : theme.surfaceBorder}`,
                background: isCurrent ? withAlpha(theme.accent, 0.12) : withAlpha(theme.surface, 0.6),
                opacity: isCurrent ? 1 : stamped ? 0.86 : 0.5,
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center',
                gap: 10,
                padding: 18,
                transform: `translateY(${(1 - pop) * 8}px)`,
              }}
            >
              <div
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 13,
                  letterSpacing: 2.6,
                  textTransform: 'uppercase',
                  color: isCurrent ? theme.accentText : theme.textMuted,
                }}
              >
                {`step ${i + 1}`}
              </div>
              <div
                style={{
                  fontFamily: theme.fontDisplay,
                  fontSize: stationW < 240 ? 24 : 28,
                  fontWeight: 700,
                  letterSpacing: -0.4,
                  color: isCurrent || stamped ? theme.text : theme.textMuted,
                  lineHeight: 1.15,
                }}
              >
                {hop.where}
              </div>
              <div
                style={{
                  fontFamily: theme.fontBody,
                  fontSize: 20,
                  lineHeight: 1.3,
                  color: theme.textMuted,
                  opacity: stamped ? 1 : 0,
                  transform: `translateY(${(stamped ? 1 - pop : 1) * 6}px)`,
                }}
              >
                {hop.gives}
              </div>
              {hop.miss && isCurrent ? (
                <div
                  style={{
                    fontFamily: theme.fontMono,
                    fontSize: 15,
                    color: withAlpha(theme.accentLimit, 0.95),
                    opacity: pop,
                  }}
                >
                  {`miss: ${hop.miss}`}
                </div>
              ) : null}
            </div>
          );
        })}

        {/* The travelling card: the key, plus a stamp for every station that
            has spoken. */}
        <div
          style={{
            position: 'absolute',
            left: cardCx - 150,
            top: CARD_TOP,
            width: 300,
            height: CARD_H,
            borderRadius: 12,
            border: `2px solid ${theme.accent}`,
            background: withAlpha(theme.surface, 0.95),
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 8,
            opacity: step.cached ? 0.35 : 1,
            boxShadow: `0 10px 24px ${withAlpha(theme.ink, 0.45)}`,
          }}
        >
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 24,
              fontWeight: 700,
              color: theme.text,
              maxWidth: 272,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {keyText}
          </div>
          <div style={{display: 'flex', gap: 6, height: 8}}>
            {hops.map((_, i) => (
              <div
                key={i}
                style={{
                  width: visited.has(i) ? 22 : 8,
                  height: 8,
                  borderRadius: 4,
                  background: visited.has(i) ? theme.accentQuantity : withAlpha(theme.line, 0.35),
                }}
              />
            ))}
          </div>
        </div>

        {/* The shortcut's label, riding its arc. */}
        {cacheDraw > 0 ? (
          <div
            style={{
              position: 'absolute',
              left: (askerCx + lastCx) / 2 - 140,
              top: 24,
              width: 280,
              textAlign: 'center',
              fontFamily: theme.fontMono,
              fontSize: 16,
              letterSpacing: 2.6,
              textTransform: 'uppercase',
              color: theme.accentText,
              opacity: cacheDraw,
            }}
          >
            cached, next time
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
