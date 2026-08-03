import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// TableScene is a spec sheet that loses its even weighting.
//
// The sheet is drawn as a real sheet first: hairline rules between rows, labels
// left, values right in tabular figures, every line the same weight. That
// fidelity is the whole setup — a stylised approximation of a table would let the
// viewer off, because they would never have skimmed past this row in the way the
// clip is accusing them of.
//
// Then the weighting goes. The other rows drop to a third and the deciding row
// keeps its strength, gains a tinted plate and a rule in its role colour, and
// nothing moves. Nothing moving is deliberate: scaling the row up or pulling it
// out of the sheet would make it a callout, and a callout says "here is a fact".
// Leaving it exactly where it was printed says "it was always here", which is
// the accusation.
//
// The rows do not reflow when the sheet dims, so the buried row stays at the same
// pixel it occupied while the viewer was ignoring it. That is the only way the
// picture can claim it was in plain sight.

const SHEET_W = Math.min(STAGE_W, 1080);
const ROW_H = 68;

type Row = {label: string; value: string};
type Step = {startMs: number; endMs: number; show: 'sheet' | 'focus' | 'read'};

const roleColour = (theme: ResolvedTheme, role: string): string => {
  switch (role) {
    case 'quantity':
      return theme.accentQuantity;
    case 'rival':
      return theme.accentRival;
    case 'neutral':
      return theme.textMuted;
    default:
      return theme.accentLimit;
  }
};

export const TableScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const source = String(props.source ?? '');
  const rows = (Array.isArray(props.rows) ? props.rows : []) as Row[];
  const at = Number(props.at ?? -1);
  const note = String(props.note ?? '');
  const role = String(props.role ?? 'limit');
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (rows.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // Once the weighting is gone it does not come back.
  const focused = steps.slice(0, idx + 1).some((s) => s.show === 'focus');
  const colour = roleColour(theme, role);

  const enter = interpolate(sinceStep, [0, 18], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // The dimming, on the beat that does it. Unhurried: the viewer needs time to
  // notice that the row they are left with is one they had already read past.
  const strip = step.show === 'focus'
    ? interpolate(sinceStep, [3, 26], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : focused
      ? 1
      : 0;

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={26}
      />

      {source ? (
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 15,
            letterSpacing: 2.4,
            textTransform: 'uppercase',
            color: theme.textMuted,
            opacity: enter * 0.7,
            marginBottom: 14,
          }}
        >
          {source}
        </div>
      ) : null}

      <div
        style={{
          width: SHEET_W,
          borderRadius: 12,
          border: `1px solid ${theme.surfaceBorder}`,
          background: theme.surface,
          overflow: 'hidden',
          opacity: enter,
        }}
      >
        {rows.map((r, i) => {
          const isBuried = i === at;
          // Other rows recede; the buried one holds. Nothing moves or resizes:
          // the row has to stay at the pixel it occupied while it was being
          // ignored, or the picture cannot claim it was in plain sight.
          const dim = isBuried ? 1 : 1 - 0.66 * strip;
          return (
            <div
              key={i}
              style={{
                height: ROW_H,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                paddingInline: 26,
                borderTop: i === 0 ? undefined : `1px solid ${withAlpha(theme.textMuted, 0.14)}`,
                background: isBuried ? withAlpha(colour, 0.12 * strip) : undefined,
                // A left rule rather than a full border, so the sheet's own
                // ruling stays intact underneath it.
                boxShadow: isBuried
                  ? `inset ${4 * strip}px 0 0 0 ${withAlpha(colour, 0.9 * strip)}`
                  : undefined,
              }}
            >
              <span
                style={{
                  fontFamily: theme.fontBody,
                  fontSize: 24,
                  color: isBuried && strip > 0 ? colour : theme.text,
                  opacity: dim,
                }}
              >
                {r.label}
              </span>
              <span
                style={{
                  fontFamily: theme.fontDisplay,
                  fontSize: 28,
                  fontWeight: isBuried && strip > 0 ? 800 : 600,
                  color: isBuried && strip > 0 ? colour : theme.text,
                  opacity: dim,
                  fontVariantNumeric: 'tabular-nums',
                }}
              >
                {r.value}
              </span>
            </div>
          );
        })}
      </div>

      {focused && note ? (
        <div
          style={{
            marginTop: 28,
            maxWidth: 1000,
            textAlign: 'center',
            fontFamily: theme.fontBody,
            fontSize: 25,
            color: theme.textMuted,
            opacity: interpolate(sinceStep, [10, 26], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            }),
          }}
        >
          {note}
        </div>
      ) : null}
    </Stage>
  );
};
