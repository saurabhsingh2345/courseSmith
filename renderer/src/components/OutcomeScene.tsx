import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// OutcomeScene: the promise, itemised, with a counter that fills as it is kept.
//
// "By the end of this lesson you will be able to…" is the most skipped
// sentence in online education, and it is skipped because it is delivered as
// prose. Here it is a LEDGER: a hard left axis, one ability stamping onto it
// per beat, each with its payoff line set underneath in muted type. Two sizes,
// two weights, one column. The viewer can count the promises without being
// told how many there are.
//
// Left-aligned rather than centred, and that is the whole layout decision. A
// centred list of four items reads as a poster; a left rail with items landing
// against it reads as a record being written, which is what a promise you can
// be held to should look like. The rail runs the full height and brightens
// only up to the last landed card, so the ledger visibly grows.
//
// The count chip in the top right is the honesty mechanism. It shows landed /
// total against a filling bar, so the closing "contract" beat — where Go lights
// every ability at once — reads as the counter completing rather than as more
// cards appearing. That beat is the handshake, and it should feel like one.
//
// Cards STAMP: they arrive slightly oversize and settle, from a spring with a
// little more snap than the payoff line that follows them. The two-stage
// arrival is deliberate — the skill lands, then the reason for it catches up.
//
// One glow maximum: the card landing on this beat.

const LEDGER_W = Math.min(STAGE_W, 1400);
const RAIL_X = 5;

type Ability = {skill: string; payoff: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'promise' | 'ability' | 'contract';
  at?: number;
  lit: number[];
};

export const OutcomeScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const lesson = String(props.lesson ?? '');
  const abilities = (Array.isArray(props.abilities) ? props.abilities : []) as Ability[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (abilities.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const landing = step.show === 'ability' ? step.at ?? -1 : -1;
  const lit = new Set(Array.isArray(step.lit) ? step.lit : []);
  const contract = step.show === 'contract';

  const stamp = spring({frame: sinceStep - 2, fps, config: {damping: 12, mass: 0.5}, durationInFrames: 22});
  const seal = contract
    ? spring({frame: sinceStep - 2, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 28})
    : 0;
  // The counter eases rather than jumping, so the bar reads as filling.
  const fillTarget = abilities.length > 0 ? lit.size / abilities.length : 0;
  const prevSize = idx > 0 ? new Set(steps[idx - 1].lit ?? []).size : 0;
  const prevFill = abilities.length > 0 ? prevSize / abilities.length : 0;
  const fill = interpolate(sinceStep, [2, 20], [prevFill, fillTarget], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const cardH = abilities.length > 3 ? 108 : 122;
  const ledgerH = abilities.length * cardH + (abilities.length - 1) * 14;
  const lastLit = lit.size > 0 ? Math.max(...lit) : -1;
  const railFill = lastLit < 0 ? 0 : ((lastLit + 1) * cardH + lastLit * 14) / ledgerH;

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

      <div style={{width: LEDGER_W}}>
        {/* Lesson name on the left, the count chip on the right. */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: 22,
            paddingBottom: 16,
            borderBottom: `2px solid ${withAlpha(theme.line, 0.2)}`,
          }}
        >
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 16,
              letterSpacing: 3.4,
              textTransform: 'uppercase',
              color: theme.textMuted,
            }}
          >
            {lesson ? `${lesson} — you will be able to` : 'you will be able to'}
          </div>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 14,
              paddingInline: 16,
              paddingBlock: 8,
              borderRadius: 10,
              background: withAlpha(theme.surface, 0.85),
              border: `1.5px solid ${contract ? withAlpha(theme.accent, 0.4 + 0.35 * seal) : theme.surfaceBorder}`,
            }}
          >
            <span
              style={{
                fontFamily: theme.fontMono,
                fontSize: 22,
                fontWeight: 700,
                color: fillTarget >= 1 ? theme.accentQuantity : theme.text,
              }}
            >
              {lit.size}/{abilities.length}
            </span>
            <span style={{width: 96, height: 8, borderRadius: 4, background: withAlpha(theme.mass, 0.14), overflow: 'hidden', display: 'block'}}>
              <span
                style={{
                  display: 'block',
                  width: `${Math.max(0, Math.min(1, fill)) * 100}%`,
                  height: '100%',
                  borderRadius: 4,
                  background: theme.accentQuantity,
                }}
              />
            </span>
          </div>
        </div>

        <div style={{position: 'relative', paddingLeft: 30}}>
          {/* The rail: the ledger's spine, lit only as far as it is written. */}
          <div
            style={{
              position: 'absolute',
              left: RAIL_X,
              top: 0,
              width: 4,
              height: ledgerH,
              borderRadius: 2,
              background: withAlpha(theme.line, 0.18),
            }}
          />
          <div
            style={{
              position: 'absolute',
              left: RAIL_X,
              top: 0,
              width: 4,
              height: ledgerH * railFill,
              borderRadius: 2,
              background: withAlpha(theme.accent, 0.85),
            }}
          />

          <div style={{display: 'flex', flexDirection: 'column', gap: 14}}>
            {abilities.map((a, i) => {
              const isLanding = i === landing;
              const on = lit.has(i);
              const pop = isLanding ? stamp : contract ? seal : 1;
              const scale = on ? (isLanding ? 1.03 - 0.03 * pop : 1) : 0.99;
              return (
                <div
                  key={i}
                  style={{
                    minHeight: cardH,
                    display: 'flex',
                    flexDirection: 'column',
                    justifyContent: 'center',
                    gap: 8,
                    paddingInline: 24,
                    paddingBlock: 16,
                    borderRadius: 14,
                    background: on ? withAlpha(theme.surface, 0.9) : withAlpha(theme.surface, 0.3),
                    border: `2px solid ${
                      isLanding
                        ? withAlpha(theme.accent, 0.4 + 0.4 * pop)
                        : contract && on
                          ? withAlpha(theme.accent, 0.24 + 0.24 * seal)
                          : on
                            ? withAlpha(theme.accent, 0.22)
                            : withAlpha(theme.surfaceBorder, 0.5)
                    }`,
                    opacity: on ? 1 : 0.34,
                    transform: `translateX(${on ? 0 : -8}px) scale(${scale})`,
                    // The one glow: the card being stamped right now.
                    boxShadow: isLanding ? `0 0 ${32 * pop}px ${withAlpha(theme.accent, 0.26)}` : undefined,
                  }}
                >
                  <div
                    style={{
                      fontFamily: theme.fontDisplay,
                      fontSize: 34,
                      fontWeight: 700,
                      letterSpacing: -0.7,
                      lineHeight: 1.14,
                      color: on ? theme.text : theme.textMuted,
                    }}
                  >
                    {a.skill}
                  </div>
                  <div
                    style={{
                      fontFamily: theme.fontBody,
                      fontSize: 23,
                      lineHeight: 1.28,
                      color: theme.textMuted,
                      // The reason catches up a beat behind the promise.
                      opacity: on ? (isLanding ? interpolate(pop, [0.35, 1], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'}) : 1) : 0,
                    }}
                  >
                    {a.payoff}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </Stage>
  );
};
