import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {iconFor} from './icons';

// AnalogyScene is two columns with a mapping drawn between them.
//
// The familiar thing is on the left and the real thing on the right, with every
// correspondence a row that spans both. Both columns are on screen and complete
// from the first frame; walking a pair lights its row and draws the connector.
//
// Three decisions carry it.
//
// The connector is drawn per row rather than as one bundle of lines between two
// lists. A bundle reads as "these are related somehow"; a row that lights end to
// end reads as "this IS that", which is the claim being made.
//
// The two sides are typographically different — the familiar column is set in
// the body face, the real column in mono. That is doing real work: it stops the
// viewer having to remember which side is the metaphor, and it does it without
// a label repeated on every row.
//
// The `breaks` beat dims the entire mapping and puts the admission over it. The
// picture receding is the point: the clip is saying "stop using this now", and
// leaving the rows bright underneath would undercut it.

const COL_W = Math.min(STAGE_W, 1560);

type Pair = {from: string; to: string; note?: string};
type Step = {startMs: number; endMs: number; show: 'picture' | 'pair' | 'breaks'; at?: number};

const HeadIcon: React.FC<{name?: string; color: string}> = ({name, color}) => {
  const Icon = iconFor(name);
  return (
    <div style={{color, lineHeight: 0}}>
      <Icon size={34} strokeWidth={1.9} />
    </div>
  );
};

export const AnalogyScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const familiar = String(props.familiar ?? '');
  const real = String(props.real ?? '');
  const breaks = String(props.breaks ?? '');
  const pairs = (Array.isArray(props.pairs) ? props.pairs : []) as Pair[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (pairs.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;
  const onBreaks = step.show === 'breaks';
  const current = step.show === 'pair' ? (step.at ?? 0) : -1;

  const enter = spring({
    frame: ((nowMs - steps[0].startMs) / 1000) * FPS,
    fps,
    config: {damping: 200, mass: 0.7},
    durationInFrames: 20,
  });

  const left = theme.accentRival;
  const right = theme.accentQuantity;

  return (
    <Stage justify="center">
      <div style={{width: COL_W, opacity: enter, position: 'relative'}}>
        {/* The two headings, each over its own column. */}
        <div style={{display: 'flex', alignItems: 'center', marginBottom: 34, opacity: onBreaks ? 0.28 : 1}}>
          <div style={{flex: 1, display: 'flex', alignItems: 'center', gap: 14}}>
            <HeadIcon name={String(props.familiarIcon ?? '')} color={left} />
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 40,
                fontWeight: 800,
                letterSpacing: -0.8,
                color: theme.text,
              }}
            >
              {familiar}
            </div>
          </div>
          <div
            style={{
              width: 120,
              textAlign: 'center',
              fontFamily: theme.fontMono,
              fontSize: 16,
              letterSpacing: 3,
              textTransform: 'uppercase',
              color: theme.textMuted,
            }}
          >
            is really
          </div>
          <div style={{flex: 1, display: 'flex', alignItems: 'center', gap: 14, justifyContent: 'flex-end'}}>
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 40,
                fontWeight: 800,
                letterSpacing: -0.8,
                color: theme.text,
              }}
            >
              {real}
            </div>
            <HeadIcon name={String(props.realIcon ?? '')} color={right} />
          </div>
        </div>

        {/* The mapping. */}
        <div style={{opacity: onBreaks ? 0.11 : 1}}>
          {pairs.map((pr, i) => {
            const lit = i === current;
            const on = interpolate(frame, [4 + i * 5, 20 + i * 5], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            });
            // The connector draws across as the row lights, so the mapping is
            // an act rather than a static line that was always there.
            const draw = lit
              ? interpolate(sinceStep, [3, 20], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                })
              : 0;
            return (
              <div key={i} style={{marginBottom: 18}}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'stretch',
                    opacity: on * (lit || current < 0 ? 1 : 0.42),
                    transform: `translateY(${(1 - on) * 12}px)`,
                  }}
                >
                  <div
                    style={{
                      flex: 1,
                      padding: '18px 24px',
                      borderRadius: 10,
                      background: lit ? withAlpha(left, 0.14) : withAlpha(theme.text, 0.04),
                      border: `1px solid ${lit ? withAlpha(left, 0.42) : 'transparent'}`,
                      fontFamily: theme.fontBody,
                      fontSize: 29,
                      color: lit ? theme.text : theme.textMuted,
                    }}
                  >
                    {pr.from}
                  </div>

                  {/* The connector for this row only. */}
                  <div
                    style={{
                      width: 120,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      position: 'relative',
                    }}
                  >
                    <div
                      style={{
                        height: 2,
                        width: '100%',
                        background: withAlpha(theme.text, 0.12),
                      }}
                    />
                    <div
                      style={{
                        position: 'absolute',
                        left: 0,
                        height: 2,
                        width: `${draw * 100}%`,
                        background: right,
                      }}
                    />
                  </div>

                  <div
                    style={{
                      flex: 1,
                      padding: '18px 24px',
                      borderRadius: 10,
                      background: lit ? withAlpha(right, 0.14) : withAlpha(theme.text, 0.04),
                      border: `1px solid ${lit ? withAlpha(right, 0.42) : 'transparent'}`,
                      // Mono on the right, body on the left: the viewer never
                      // has to remember which side is the metaphor.
                      fontFamily: theme.fontMono,
                      fontSize: 25,
                      letterSpacing: 0.4,
                      color: lit ? theme.text : theme.textMuted,
                      textAlign: 'right',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'flex-end',
                    }}
                  >
                    {pr.to}
                  </div>
                </div>

                {lit && pr.note ? (
                  <div
                    style={{
                      fontFamily: theme.fontBody,
                      fontSize: 23,
                      color: theme.textMuted,
                      textAlign: 'center',
                      marginTop: 12,
                      opacity: interpolate(sinceStep, [16, 28], [0, 1], {
                        extrapolateLeft: 'clamp',
                        extrapolateRight: 'clamp',
                      }),
                    }}
                  >
                    {pr.note}
                  </div>
                ) : null}
              </div>
            );
          })}
        </div>

        {/* Where the picture stops working, over the receded mapping. */}
        {onBreaks && breaks ? (
          <div
            style={{
              position: 'absolute',
              inset: 0,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              opacity: interpolate(sinceStep, [4, 20], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 18,
                letterSpacing: 4.5,
                textTransform: 'uppercase',
                color: theme.accentLimit,
                marginBottom: 22,
              }}
            >
              Where it breaks
            </div>
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 54,
                fontWeight: 700,
                letterSpacing: -1,
                lineHeight: 1.2,
                color: theme.text,
                textAlign: 'center',
                maxWidth: 1200,
              }}
            >
              {breaks}
            </div>
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
