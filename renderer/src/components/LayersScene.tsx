import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W, STAGE_H} from './Stage';
import {SceneHeader} from './SceneHeader';

// LayersScene: strata, and the one line that is not like the others.
//
// The bars run the full width of the composition and sit 2px apart, which is a
// deliberate rejection of the usual stack drawing. Boxes with margins between
// them read as separate objects that happen to be arranged vertically; bars
// that nearly touch read as ONE body with divisions in it, which is what a
// stack actually is. The eye should have to look for the seams.
//
// That is what makes the boundary work. Every seam is two pixels of background
// except one, which opens to thirty and carries a dashed rule with a name at
// its right-hand end. There is no other special treatment anywhere in the
// picture, so the single wide gap does all the arguing: this line is different,
// and everything above it is a different world from everything below.
//
// The crossing is the payoff and it is drawn as compression, not as a journey.
// A packet crossing an encapsulation boundary, or a call crossing into the
// kernel, does not simply move — it gets wrapped, checked, made smaller in the
// terms of the layer receiving it. So the chip shrinks as it passes through the
// rule, and its three stacked bands close up into one. After it has crossed,
// the rule stays lit, because a boundary that has been used is a fact the rest
// of the clip can lean on.
//
// Focus is a shift in weight rather than a spotlight: the focused bar gains
// contrast and slides a few pixels right, so the stack stays readable as a
// whole while one band is being discussed.

const STACK_W = Math.min(STAGE_W, 1520);
const BODY_H = STAGE_H - 170;
const SEAM = 2;
const BOUNDARY_SEAM = 30;
const CHIP_W = 132;
const CHIP_H = 44;

type Stratum = {label: string; holds: string; above: boolean};
type Step = {
  startMs: number;
  endMs: number;
  show: 'stack' | 'stratum' | 'cross' | 'whole';
  at?: number;
  lit: number[];
  crossed: boolean;
};

export const LayersScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const strata = (Array.isArray(props.strata) ? props.strata : []) as Stratum[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const boundary = Number(props.boundary ?? -1);
  const boundaryLabel = String(props.boundaryLabel ?? '');
  if (strata.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const n = strata.length;
  const hasBoundary = boundary >= 0 && boundary < n - 1;
  const seams = SEAM * (n - 1) + (hasBoundary ? BOUNDARY_SEAM - SEAM : 0);
  const barH = Math.max(70, Math.min(126, Math.floor((BODY_H - seams) / n)));

  // Where each bar starts, walked once so the boundary seam is in one place.
  const tops: number[] = [];
  let y = 0;
  for (let i = 0; i < n; i += 1) {
    tops.push(y);
    y += barH + (hasBoundary && i === boundary ? BOUNDARY_SEAM : SEAM);
  }
  // The loop above adds a seam after the last bar too; the stack ends at it.
  const stackH = y - SEAM;
  const ruleY = hasBoundary ? tops[boundary] + barH + BOUNDARY_SEAM / 2 : -1;

  const focused = new Set(step.lit ?? []);
  const current = step.show === 'stratum' && typeof step.at === 'number' ? step.at : -1;
  const allLit = step.show === 'whole';
  const crossing = step.show === 'cross';
  const ruleHot = crossing || Boolean(step.crossed) || allLit;

  const stackIn = interpolate(frame, [0, 20], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  // The crossing: one continuous compression through the rule.
  const cross = crossing
    ? spring({frame: sinceStep - 4, fps, config: {damping: 14, mass: 0.6}, durationInFrames: 30})
    : 0;
  const chipY = hasBoundary
    ? interpolate(cross, [0, 1], [tops[boundary] + barH / 2 - CHIP_H / 2, tops[boundary + 1] + barH / 2 - CHIP_H / 2])
    : 0;
  const chipScale = interpolate(cross, [0, 1], [1, 0.62], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const bandSpread = interpolate(cross, [0, 1], [1, 0.2], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

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

      <div style={{width: STACK_W, height: stackH, position: 'relative', opacity: stackIn}}>
        {strata.map((s, i) => {
          const isCurrent = i === current;
          const seen = focused.has(i) || allLit;
          const pop = isCurrent
            ? spring({frame: sinceStep, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 24})
            : 1;
          const weight = isCurrent ? 1 : seen ? 0.72 : 0.4;
          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: 0,
                top: tops[i],
                width: STACK_W,
                height: barH,
                boxSizing: 'border-box',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '0 30px',
                borderRadius: 6,
                background: withAlpha(theme.surface, 0.3 + 0.65 * weight),
                border: `1px solid ${isCurrent ? withAlpha(theme.accent, 0.7) : theme.surfaceBorder}`,
                transform: `translateX(${isCurrent ? pop * 14 : 0}px)`,
                // The one glow, and only on the band being discussed.
                boxShadow: isCurrent ? `0 0 30px ${withAlpha(theme.accent, 0.22 * pop)}` : undefined,
              }}
            >
              <div style={{display: 'flex', alignItems: 'baseline', gap: 18}}>
                <span
                  style={{
                    fontFamily: theme.fontMono,
                    fontSize: 13,
                    letterSpacing: 2,
                    color: withAlpha(theme.textMuted, 0.75),
                  }}
                >
                  {s.above ? 'above' : hasBoundary ? 'below' : String(i + 1).padStart(2, '0')}
                </span>
                <span
                  style={{
                    fontFamily: theme.fontDisplay,
                    fontSize: barH > 100 ? 34 : 29,
                    fontWeight: 700,
                    letterSpacing: -0.6,
                    color: isCurrent ? theme.accentText : seen ? theme.text : withAlpha(theme.text, 0.5),
                  }}
                >
                  {s.label}
                </span>
              </div>
              <span
                style={{
                  fontFamily: theme.fontBody,
                  fontSize: barH > 100 ? 23 : 20,
                  color: theme.textMuted,
                  opacity: 0.35 + 0.65 * weight,
                  textAlign: 'right',
                  maxWidth: STACK_W * 0.46,
                }}
              >
                {s.holds}
              </span>
            </div>
          );
        })}

        {/* The one seam that is not like the others. */}
        {hasBoundary ? (
          <div style={{position: 'absolute', left: 0, top: ruleY - 1, width: STACK_W}}>
            <div
              style={{
                height: 0,
                borderTop: `2px dashed ${ruleHot ? theme.accentLimit : withAlpha(theme.accentLimit, 0.45)}`,
              }}
            />
            <div
              style={{
                position: 'absolute',
                right: 0,
                top: -12,
                padding: '3px 12px',
                borderRadius: 5,
                background: theme.background,
                fontFamily: theme.fontMono,
                fontSize: 15,
                letterSpacing: 1.4,
                textTransform: 'uppercase',
                color: ruleHot ? theme.accentLimit : withAlpha(theme.accentLimit, 0.6),
              }}
            >
              {boundaryLabel}
            </div>
          </div>
        ) : null}

        {/* The payload, compressing as it goes through. */}
        {crossing && hasBoundary ? (
          <div
            style={{
              position: 'absolute',
              left: STACK_W * 0.3,
              top: chipY,
              width: CHIP_W,
              height: CHIP_H,
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              gap: 4 * bandSpread,
              padding: 6,
              boxSizing: 'border-box',
              borderRadius: 8,
              background: withAlpha(theme.accent, 0.16),
              border: `2px solid ${theme.accent}`,
              transform: `scale(${chipScale})`,
              boxShadow: `0 0 26px ${withAlpha(theme.accent, 0.4)}`,
            }}
          >
            {[0, 1, 2].map((b) => (
              <div
                key={b}
                style={{
                  height: 5,
                  borderRadius: 3,
                  background: b === 1 ? theme.accent : withAlpha(theme.accent, 0.5),
                  width: `${100 - b * 12 * bandSpread}%`,
                }}
              />
            ))}
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
