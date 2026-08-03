import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// MultiplyScene builds one sentence across the frame: a figure, a row of glyphs,
// and a product.
//
// The glyph row is the reason this is not `metric` with two numbers. "× 8" is an
// assertion the viewer has to take on trust; eight identical marks appearing one
// after another is the multiplication happening in front of them, and by the
// third or fourth mark they are already ahead of the narration. That is the
// whole effect, so the row is drawn even when the count is large — a texture of
// sixty-four marks still reads as "far too many", which is the point.
//
// Two decisions carry the rest.
//
// The per-unit figure does not move or shrink when the product arrives. It stays
// exactly where it was set, at the same size, and the product lands beneath it.
// Replacing it — or scaling it down to make room — would take away the comparison
// the clip spent two beats setting up.
//
// The product counts up from the per-unit figure rather than from zero. Counting
// from zero is the ordinary reveal and it wastes the setup: starting at 14.5 and
// running to 116 is the multiplication drawn as motion, and the viewer sees the
// distance rather than just the destination.

const LAYOUT_W = Math.min(STAGE_W, 1180);
const GLYPH = 30;
const GLYPH_GAP = 9;

type Step = {startMs: number; endMs: number; show: 'unit' | 'count' | 'total' | 'caveat'};

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

/** Trailing zeros are noise on a figure the clip is about. */
const trim = (n: number): string =>
  Number.isInteger(n) ? String(n) : String(Number(n.toFixed(2)));

export const MultiplyScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const unitValue = Number(props.unitValue ?? 0);
  const unit = String(props.unit ?? '');
  const unitLabel = String(props.unitLabel ?? '');
  const unitNote = String(props.unitNote ?? '');
  const count = Number(props.count ?? 0);
  const countLabel = String(props.countLabel ?? '');
  const total = Number(props.total ?? 0);
  const totalLabel = String(props.totalLabel ?? '');
  const totalNote = String(props.totalNote ?? '');
  const caveat = String(props.caveat ?? '');
  const role = String(props.role ?? 'limit');
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (unitValue <= 0 || count <= 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const reached = (show: Step['show']) =>
    steps.slice(0, idx + 1).some((s) => s.show === show);
  const showCount = reached('count');
  const showTotal = reached('total');
  const showCaveat = reached('caveat');

  const colour = roleColour(theme, role);
  const enter = interpolate(sinceStep, [0, 16], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // The glyphs land one after another. Fast, but sequential — a row that appears
  // at once is "× 8" again, which is the assertion this picture replaces.
  const fill = step.show === 'count' ? interpolate(sinceStep, [2, 30], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  }) : 1;
  const productT = step.show === 'total'
    ? interpolate(sinceStep, [2, 28], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'})
    : 1;

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

      <div style={{width: LAYOUT_W, textAlign: 'center'}}>
        {/* The per-unit figure. Set once and never moved: the comparison the
            product makes is against this, exactly where the viewer left it. */}
        <div
          style={{
            fontFamily: theme.fontDisplay,
            fontSize: 78,
            fontWeight: 800,
            lineHeight: 1,
            color: theme.text,
            fontVariantNumeric: 'tabular-nums',
            opacity: enter,
          }}
        >
          {trim(unitValue)}
          <span
            style={{
              fontFamily: theme.fontMono,
              fontSize: 26,
              fontWeight: 400,
              color: theme.textMuted,
              marginLeft: 8,
            }}
          >
            {unit}
          </span>
        </div>
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 16,
            letterSpacing: 2.8,
            textTransform: 'uppercase',
            color: theme.textMuted,
            marginTop: 10,
            opacity: enter,
          }}
        >
          {unitLabel}
        </div>
        {unitNote && !showCount ? (
          <div
            style={{
              fontFamily: theme.fontBody,
              fontSize: 23,
              color: theme.textMuted,
              marginTop: 16,
              opacity: enter,
            }}
          >
            {unitNote}
          </div>
        ) : null}

        {/* The glyph row: the multiplication, happening. */}
        {showCount ? (
          <div style={{marginTop: 30}}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'center',
                flexWrap: 'wrap',
                gap: GLYPH_GAP,
                maxWidth: 900,
                marginInline: 'auto',
              }}
            >
              {Array.from({length: count}, (_, i) => {
                const on = interpolate(fill, [i / count, (i + 1) / count], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                });
                return (
                  <div
                    key={i}
                    style={{
                      width: GLYPH,
                      height: GLYPH * 1.25,
                      borderRadius: 4,
                      background: withAlpha(colour, 0.16 + 0.34 * on),
                      border: `1px solid ${withAlpha(colour, 0.25 + 0.55 * on)}`,
                      transform: `scale(${0.7 + 0.3 * on})`,
                    }}
                  />
                );
              })}
            </div>
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 16,
                letterSpacing: 2.8,
                textTransform: 'uppercase',
                color: theme.textMuted,
                marginTop: 14,
              }}
            >
              × {count} — {countLabel}
            </div>
          </div>
        ) : null}

        {/* The product. */}
        {showTotal ? (
          <div style={{marginTop: 26}}>
            <div
              style={{
                fontFamily: theme.fontDisplay,
                fontSize: 116,
                fontWeight: 800,
                lineHeight: 1,
                color: colour,
                fontVariantNumeric: 'tabular-nums',
              }}
            >
              {/* Counted up from the per-unit figure, not from zero: the distance
                  travelled IS the multiplication. Interpolated here rather than
                  through counted(), which only ever ramps from zero. */}
              {trim(unitValue + (total - unitValue) * productT)}
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 34,
                  fontWeight: 400,
                  color: theme.textMuted,
                  marginLeft: 10,
                }}
              >
                {unit}
              </span>
            </div>
            {totalLabel ? (
              <div
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 17,
                  letterSpacing: 3,
                  textTransform: 'uppercase',
                  color: colour,
                  marginTop: 12,
                }}
              >
                {totalLabel}
              </div>
            ) : null}
            {totalNote ? (
              <div
                style={{
                  fontFamily: theme.fontBody,
                  fontSize: 24,
                  color: theme.textMuted,
                  marginTop: 14,
                }}
              >
                {totalNote}
              </div>
            ) : null}
          </div>
        ) : null}

        {/* The caveat: a chip rather than a figure, so it cannot be mistaken for
            another number in the arithmetic. */}
        {showCaveat && caveat ? (
          <div
            style={{
              display: 'inline-block',
              marginTop: 22,
              paddingInline: 14,
              paddingBlock: 8,
              borderRadius: 8,
              fontFamily: theme.fontMono,
              fontSize: 16,
              letterSpacing: 1.4,
              textTransform: 'uppercase',
              color: theme.accentLimit,
              background: withAlpha(theme.accentLimit, 0.1),
              border: `1px solid ${withAlpha(theme.accentLimit, 0.45)}`,
              opacity: interpolate(sinceStep, [2, 18], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              }),
            }}
          >
            {caveat}
          </div>
        ) : null}
      </div>
    </Stage>
  );
};
