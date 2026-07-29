import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';

// GaugeScene is a marked line and the things measured against it.
//
// The composition is a single horizontal track with the ceiling drawn across it
// as a dashed rule. Bars fill from the left and stop where they stop. That is
// the entire vocabulary, and keeping it that small is what makes the one moment
// that matters — a fill crossing the rule — read instantly.
//
// Three decisions carry it.
//
// The ceiling is drawn on the track and stays there for the whole clip, rather
// than appearing beside each bar. A line that moves is a line the eye has to
// re-find every beat, and the claim the template makes is that ONE limit
// governs everything.
//
// A bar that overruns is drawn past the rule rather than clipped at it. Clipping
// would say "this is full"; overrunning says "this does not fit", which is a
// different and much more useful statement. The overrun segment is drawn in the
// limit colour so the failing part is visually separable from the part that
// would have fit.
//
// Bars already run stay on screen, dimmed. The gauge is cumulative because the
// verdict is a comparison, and a comparison you have to remember is one nobody
// makes.

const TRACK_W = Math.min(STAGE_W, 1420);
const TRACK_H = 54;
const ROW_GAP = 30;

type Bar = {label: string; value: number; note?: string; fits: boolean; frac: number};
type Ceiling = {value: number; label: string; note?: string; frac: number};
type Step = {startMs: number; endMs: number; show: 'ceiling' | 'bar' | 'verdict'; at?: number};

export const GaugeScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const unit = String(props.unit ?? '');
  const ceiling = props.ceiling as Ceiling | undefined;
  const bars = (Array.isArray(props.bars) ? props.bars : []) as Bar[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (!ceiling || bars.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;
  const verdict = step.show === 'verdict';

  // Which bars have been run by now. On the verdict every bar is up; otherwise
  // it is every bar whose beat has already passed, so the picture accumulates.
  const runIndex = new Map<number, number>();
  steps.slice(0, idx + 1).forEach((s) => {
    if (s.show === 'bar' && s.at !== undefined && !runIndex.has(s.at)) {
      runIndex.set(s.at, s.startMs);
    }
  });
  const shown = verdict ? bars.map((_, i) => i) : [...runIndex.keys()];
  const current = step.show === 'bar' ? (step.at ?? 0) : -1;

  const ceilingIn = spring({
    frame: ((nowMs - steps[0].startMs) / 1000) * FPS,
    fps,
    config: {damping: 200, mass: 0.7},
    durationInFrames: 20,
  });

  return (
    <Stage justify="center">
      <div style={{width: TRACK_W, display: 'flex', flexDirection: 'column'}}>
        {/* The ceiling's identity, set above the track where the rule lands. */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'flex-start',
            marginBottom: 18,
            marginLeft: `${ceiling.frac * 100}%`,
            transform: 'translateX(-50%)',
            opacity: ceilingIn,
            whiteSpace: 'nowrap',
          }}
        >
          <div style={{textAlign: 'center'}}>
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 17,
                letterSpacing: 3.4,
                textTransform: 'uppercase',
                color: theme.accentLimit,
              }}
            >
              {ceiling.label}
            </div>
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 38,
                fontWeight: 800,
                color: theme.text,
                marginTop: 4,
                fontVariantNumeric: 'tabular-nums',
              }}
            >
              {ceiling.value}
              <span style={{fontFamily: theme.fontMono, fontSize: 19, color: theme.textMuted, marginLeft: 5}}>
                {unit}
              </span>
            </div>
          </div>
        </div>

        {/* The rule itself, drawn the full height of the stack of tracks. */}
        <div style={{position: 'relative'}}>
          <div
            style={{
              position: 'absolute',
              left: `${ceiling.frac * 100}%`,
              top: -8,
              bottom: -8,
              width: 0,
              borderLeft: `3px dashed ${theme.accentLimit}`,
              opacity: ceilingIn * 0.85,
              zIndex: 2,
            }}
          />

          {bars.map((bar, i) => {
            const isShown = shown.includes(i);
            const isCurrent = i === current;
            // Each bar fills from the frame its own beat started, so a bar that
            // ran three beats ago is not re-animating under the current one.
            const startedAt = runIndex.get(i);
            const since =
              startedAt === undefined ? 0 : ((nowMs - startedAt) / 1000) * FPS;
            const fill = isShown
              ? interpolate(since, [2, 24], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                })
              : 0;
            // Verdict brings everything to full without re-running the fill.
            const grown = verdict ? 1 : fill;
            const w = bar.frac * grown * 100;
            const overFrac = Math.max(0, bar.frac - ceiling.frac);
            const overW = Math.min(overFrac, Math.max(0, bar.frac * grown - ceiling.frac)) * 100;
            const good = theme.accentQuantity;
            const bad = theme.accentLimit;

            return (
              <div
                key={i}
                style={{
                  marginBottom: i === bars.length - 1 ? 0 : ROW_GAP,
                  opacity: isShown ? (isCurrent || verdict ? 1 : 0.42) : 0.14,
                }}
              >
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'baseline',
                    marginBottom: 9,
                  }}
                >
                  <div
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: 19,
                      letterSpacing: 2.2,
                      textTransform: 'uppercase',
                      color: isCurrent || verdict ? theme.text : theme.textMuted,
                    }}
                  >
                    {bar.label}
                  </div>
                  <div
                    style={{
                      fontFamily: theme.fontDisplay,
                      fontSize: 26,
                      fontWeight: 700,
                      color: bar.fits ? good : bad,
                      opacity: isShown ? 1 : 0,
                      fontVariantNumeric: 'tabular-nums',
                    }}
                  >
                    {bar.value}
                    <span style={{fontFamily: theme.fontMono, fontSize: 15, color: theme.textMuted, marginLeft: 4}}>
                      {unit}
                    </span>
                  </div>
                </div>

                <div
                  style={{
                    position: 'relative',
                    height: TRACK_H,
                    borderRadius: 5,
                    background: withAlpha(theme.text, 0.06),
                    overflow: 'visible',
                  }}
                >
                  {/* The part that would have fit. */}
                  <div
                    style={{
                      position: 'absolute',
                      left: 0,
                      top: 0,
                      bottom: 0,
                      width: `${w}%`,
                      borderRadius: 5,
                      background: bar.fits ? good : withAlpha(good, 0.5),
                    }}
                  />
                  {/* The overrun, drawn past the rule in the limit colour. A bar
                      clipped at the line would say "full"; one that runs past
                      says "does not fit", which is the actual claim. */}
                  {!bar.fits && overW > 0 ? (
                    <div
                      style={{
                        position: 'absolute',
                        left: `${ceiling.frac * 100}%`,
                        top: 0,
                        bottom: 0,
                        width: `${overW}%`,
                        borderRadius: '0 5px 5px 0',
                        background: bad,
                      }}
                    />
                  ) : null}
                </div>

                {bar.note && (isCurrent || verdict) ? (
                  <div
                    style={{
                      fontFamily: theme.fontBody,
                      fontSize: 22,
                      color: theme.textMuted,
                      marginTop: 10,
                      opacity: interpolate(since, [18, 30], [0, 1], {
                        extrapolateLeft: 'clamp',
                        extrapolateRight: 'clamp',
                      }),
                    }}
                  >
                    {bar.note}
                  </div>
                ) : null}
              </div>
            );
          })}
        </div>

        {/* The ceiling's reason, held under the track until the line is set and
            then left alone — it explains the rule, not any one bar. */}
        {ceiling.note && step.show === 'ceiling' ? (
          <div
            style={{
              fontFamily: theme.fontBody,
              fontSize: 25,
              color: theme.textMuted,
              textAlign: 'center',
              marginTop: 34,
              opacity: interpolate(sinceStep, [14, 28], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            {ceiling.note}
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
