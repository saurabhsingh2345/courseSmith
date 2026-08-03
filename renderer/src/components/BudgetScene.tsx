import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';
import {counted} from './counted';

// BudgetScene is one bar the width of the pot, eaten from the left.
//
// The bar is drawn at full width on the first frame and never changes size. That
// is the design: claims fill in from the left as coloured segments and the
// remainder is whatever the bar still shows empty. A bar that GREW as claims
// landed would be `costing` — accumulation toward a total — and the whole point
// here is that the ceiling was there from the start and is not moving.
//
// Two decisions carry it.
//
// Segment widths are fractions of the POT, computed in Go, not of the claimed
// total. Re-normalising as claims landed would make the first segment appear to
// shrink each time a new one arrived, which says the opposite of what is
// happening — nothing gives any of it back.
//
// The overrun is drawn past the right end of the bar rather than clipped at it.
// Clipping says "full"; overrunning says "this did not fit", and a budget that
// busts is the punchline the reference clips keep landing on. The remainder
// figure goes to the limit colour and the label says "over budget", so the
// closing frame states the failure rather than a suspiciously round zero.

const BAR_W = Math.min(STAGE_W, 1300);
const BAR_H = 76;

type Claim = {amount: number; label: string; note?: string; role: string; frac: number};
type Step = {
  startMs: number;
  endMs: number;
  show: 'pot' | 'claim' | 'remainder';
  at?: number;
  taken: number[];
  left: number;
};

const roleColour = (theme: ResolvedTheme, role: string): string => {
  switch (role) {
    case 'quantity':
      return theme.accentQuantity;
    case 'limit':
      return theme.accentLimit;
    case 'rival':
      return theme.accentRival;
    default:
      return theme.textMuted;
  }
};

/**
 * How solid a segment of this role is.
 *
 * A neutral claim is context — the weights, the driver, the things everyone
 * already expects to pay for. Drawn at full strength they were the brightest
 * marks on the frame, which put the loudest colour on the least interesting
 * bites and left the one that actually squeezes the budget looking subordinate
 * to them. Same fault the occupancy grid had, same fix: the roles that carry
 * meaning keep the top of the range.
 */
const roleSolidity = (role: string): number => (role === 'neutral' ? 0.4 : 0.88);

/** Trailing zeros are noise on a figure the clip is about. */
const trim = (n: number): string =>
  Number.isInteger(n) ? String(n) : String(Number(n.toFixed(2)));

export const BudgetScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const pot = Number(props.pot ?? 0);
  const unit = String(props.unit ?? '');
  const potLabel = String(props.potLabel ?? '');
  const remainderLabel = String(props.remainderLabel ?? 'left');
  const claims = (Array.isArray(props.claims) ? props.claims : []) as Claim[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (pot <= 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const taken = new Set(step.taken ?? []);
  const enter = interpolate(sinceStep, [0, 18], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // The arriving segment grows in. Everything already taken is at full width, so
  // only the current bite moves.
  const bite = interpolate(sinceStep, [3, 22], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const isRemainder = step.show === 'remainder';
  const left = step.left;
  const over = left < 0;
  // On the closing beat the remainder counts down to its value, so the number
  // lands with the sentence rather than being read before it.
  const remainderT = isRemainder
    ? interpolate(sinceStep, [2, 26], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : 1;

  const spoken = step.at !== undefined ? claims[step.at] : undefined;
  const note = spoken?.note;

  // Segments laid left to right in declaration order, each offset by the widths
  // before it that have also been taken.
  let offset = 0;
  const segments = claims.map((c, i) => {
    const shown = taken.has(i);
    const isCurrent = step.at === i;
    const seg = {claim: c, index: i, shown, isCurrent, offset};
    if (shown) offset += c.frac;
    return seg;
  });

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={30}
      />

      {/* The pot's identity, above the bar. It stays on screen for the whole
          clip: the remainder only means anything against the thing it is left
          out of. */}
      <div
        style={{
          width: BAR_W,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'baseline',
          marginBottom: 12,
          opacity: enter,
        }}
      >
        <span
          style={{
            fontFamily: theme.fontMono,
            fontSize: 15,
            letterSpacing: 2.8,
            textTransform: 'uppercase',
            color: theme.textMuted,
          }}
        >
          {potLabel}
        </span>
        <span
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 30,
            fontWeight: 800,
            color: theme.text,
            fontVariantNumeric: 'tabular-nums',
          }}
        >
          {trim(pot)}
          <span style={{fontFamily: theme.fontMono, fontSize: 17, color: theme.textMuted, marginLeft: 5}}>
            {unit}
          </span>
        </span>
      </div>

      {/* The bar. Full width from the first frame — it is the pot, and the pot
          does not move. */}
      <div
        style={{
          width: BAR_W,
          height: BAR_H,
          borderRadius: 10,
          position: 'relative',
          background: theme.surface,
          border: `1px solid ${theme.surfaceBorder}`,
          opacity: enter,
        }}
      >
        {segments.map(({claim, index, shown, isCurrent, offset: from}) => {
          if (!shown) return null;
          const w = claim.frac * (isCurrent ? bite : 1);
          const colour = roleColour(theme, claim.role);
          return (
            <div
              key={index}
              style={{
                position: 'absolute',
                top: 0,
                bottom: 0,
                left: `${from * 100}%`,
                width: `${w * 100}%`,
                background: withAlpha(colour, roleSolidity(claim.role)),
                borderRight: `1px solid ${theme.bgTop}`,
                borderTopLeftRadius: from === 0 ? 10 : 0,
                borderBottomLeftRadius: from === 0 ? 10 : 0,
              }}
            />
          );
        })}
      </div>

      {/* The current claim, or the remainder. One slot rather than two, because
          the closing beat should not have a claim's caption still under it. */}
      {isRemainder ? (
        <div style={{marginTop: 34, textAlign: 'center'}}>
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 96,
              fontWeight: 800,
              lineHeight: 1,
              color: over ? theme.accentLimit : theme.accentRival,
              fontVariantNumeric: 'tabular-nums',
            }}
          >
            {counted(trim(Math.abs(left)), true, remainderT)}
            <span
              style={{
                fontFamily: theme.fontMono,
                fontSize: 30,
                fontWeight: 400,
                color: theme.textMuted,
                marginLeft: 10,
              }}
            >
              {unit}
            </span>
          </div>
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 18,
              letterSpacing: 3,
              textTransform: 'uppercase',
              color: over ? theme.accentLimit : theme.textMuted,
              marginTop: 12,
            }}
          >
            {remainderLabel}
          </div>
        </div>
      ) : spoken ? (
        <div style={{marginTop: 30, textAlign: 'center'}}>
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 52,
              fontWeight: 800,
              lineHeight: 1,
              color: roleColour(theme, spoken.role),
              fontVariantNumeric: 'tabular-nums',
            }}
          >
            −{counted(trim(spoken.amount), true, bite)}
            <span
              style={{
                fontFamily: theme.fontMono,
                fontSize: 20,
                fontWeight: 400,
                color: theme.textMuted,
                marginLeft: 8,
              }}
            >
              {unit}
            </span>
            <span
              style={{
                fontFamily: theme.fontMono,
                fontSize: 20,
                fontWeight: 400,
                color: theme.textMuted,
                marginLeft: 16,
              }}
            >
              {spoken.label}
            </span>
          </div>
          {note ? (
            <div
              style={{
                fontFamily: theme.fontBody,
                fontSize: 24,
                color: theme.textMuted,
                marginTop: 14,
                maxWidth: 1040,
              }}
            >
              {note}
            </div>
          ) : null}
        </div>
      ) : null}
    </Stage>
  );
};
