import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// MultiplexScene is a column of waiting sources and the one worker that serves
// them.
//
// The layout is deliberately asymmetric: a tall column of small identical chips
// on the left, a lot of empty space, and one box on the right. That imbalance IS
// the argument — the eye reads "many, and only that" before the narrator gets to
// the verb. A symmetric diagram with the worker the same size as the pool would
// draw the same information and make the opposite point.
//
// Two decisions carry the rest.
//
// Ready sources are connected to the worker by lines drawn at the same time, not
// in sequence. Drawing them one after another — even quickly — is the polling
// picture this template exists to refuse, and a viewer reads a stagger as an
// order whether or not one is meant.
//
// The worker's badge counts what it is holding this pass. It is the only number
// on the frame and it does the work the layout cannot: "3" beside a box marked
// "1 thread" is the whole claim in two glyphs.

const COL_W = 260;
const CHIP_H = 46;
const CHIP_GAP = 9;
const WORKER_W = 300;
const LAYOUT_W = Math.min(STAGE_W, 1140);

type Source = {label: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'pool' | 'round' | 'read';
  at?: number;
  ready?: number[];
  note?: string;
  role?: string;
};

const roleColour = (theme: ResolvedTheme, role?: string): string => {
  switch (role) {
    case 'limit':
      return theme.accentLimit;
    case 'rival':
      return theme.accentRival;
    case 'quantity':
      return theme.accentQuantity;
    default:
      return theme.accentQuantity;
  }
};

export const MultiplexScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const sourceKind = String(props.sourceKind ?? '');
  const worker = String(props.worker ?? '');
  const workerNote = String(props.workerNote ?? '');
  const sources = (Array.isArray(props.sources) ? props.sources : []) as Source[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (sources.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // On a closing "read" beat the last round stays lit, so the frame does not end
  // on an idle pool the narrator is describing in the past tense.
  const shown =
    step.show === 'round'
      ? step
      : [...steps.slice(0, idx + 1)].reverse().find((s) => s.show === 'round');
  const ready = step.show === 'pool' ? [] : (shown?.ready ?? []);
  const colour = roleColour(theme, shown?.role);

  const poolIn = interpolate(sinceStep, [0, 18], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // All the ready chips light on the same curve. A stagger would read as an
  // order, which is the polling picture this template refuses.
  const wake = interpolate(sinceStep, [3, 17], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const colH = sources.length * (CHIP_H + CHIP_GAP) - CHIP_GAP;
  const readySet = new Set(ready);

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

      <div
        style={{
          width: LAYOUT_W,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          opacity: poolIn,
        }}
      >
        {/* The pool. */}
        <div style={{width: COL_W}}>
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 14,
              letterSpacing: 3,
              textTransform: 'uppercase',
              color: theme.textMuted,
              opacity: 0.7,
              marginBottom: 12,
            }}
          >
            {sources.length} {sourceKind}
          </div>
          {sources.map((s, i) => {
            const isReady = readySet.has(i);
            const lit = isReady ? wake : 0;
            return (
              <div
                key={i}
                style={{
                  height: CHIP_H,
                  marginBottom: CHIP_GAP,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  paddingInline: 14,
                  borderRadius: 9,
                  background: isReady
                    ? withAlpha(colour, 0.14 * lit)
                    : theme.surface,
                  border: `1px solid ${
                    isReady ? withAlpha(colour, 0.25 + 0.4 * lit) : theme.surfaceBorder
                  }`,
                }}
              >
                <span
                  style={{
                    fontFamily: theme.fontMono,
                    fontSize: 19,
                    color: isReady ? colour : theme.textMuted,
                    opacity: isReady ? 0.6 + 0.4 * lit : 0.72,
                  }}
                >
                  {s.label}
                </span>
                {isReady ? (
                  <span
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: 12,
                      letterSpacing: 1.6,
                      textTransform: 'uppercase',
                      color: colour,
                      opacity: lit,
                    }}
                  >
                    ready
                  </span>
                ) : null}
              </div>
            );
          })}
        </div>

        {/* The wires. One SVG so every ready source connects on the same curve
            and no chip's line can arrive before another's. */}
        <svg
          width={LAYOUT_W - COL_W - WORKER_W}
          height={colH}
          style={{overflow: 'visible'}}
        >
          {sources.map((_, i) => {
            if (!readySet.has(i)) return null;
            const w = LAYOUT_W - COL_W - WORKER_W;
            const y = i * (CHIP_H + CHIP_GAP) + CHIP_H / 2 + 34;
            const midY = colH / 2 + 34;
            return (
              <path
                key={i}
                d={`M 0 ${y} C ${w * 0.45} ${y}, ${w * 0.55} ${midY}, ${w} ${midY}`}
                fill="none"
                stroke={colour}
                strokeWidth={2}
                opacity={0.22 + 0.5 * wake}
              />
            );
          })}
        </svg>

        {/* The worker. */}
        <div
          style={{
            width: WORKER_W,
            padding: '26px 22px',
            borderRadius: 14,
            textAlign: 'center',
            background: ready.length > 0 ? withAlpha(colour, 0.1 * wake) : theme.surface,
            border: `1px solid ${
              ready.length > 0 ? withAlpha(colour, 0.3 + 0.35 * wake) : theme.surfaceBorder
            }`,
          }}
        >
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 34,
              fontWeight: 800,
              color: theme.text,
              letterSpacing: -0.4,
            }}
          >
            {worker}
          </div>
          {workerNote ? (
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 16,
                letterSpacing: 2.4,
                textTransform: 'uppercase',
                color: theme.textMuted,
                marginTop: 8,
              }}
            >
              {workerNote}
            </div>
          ) : null}
          {/* The count, which is the whole claim in two glyphs. */}
          <div
            style={{
              marginTop: 16,
              fontFamily: theme.fontDisplay,
              fontSize: 58,
              fontWeight: 800,
              lineHeight: 1,
              color: colour,
              opacity: ready.length > 0 ? wake : 0.2,
              fontVariantNumeric: 'tabular-nums',
            }}
          >
            {ready.length}
          </div>
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 13,
              letterSpacing: 2,
              textTransform: 'uppercase',
              color: theme.textMuted,
              opacity: 0.7,
              marginTop: 4,
            }}
          >
            this pass
          </div>
        </div>
      </div>

      {shown?.note && step.show !== 'pool' ? (
        <div
          style={{
            marginTop: 30,
            maxWidth: 1020,
            textAlign: 'center',
            fontFamily: theme.fontBody,
            fontSize: 25,
            color: theme.textMuted,
            opacity: interpolate(sinceStep, [8, 22], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            }),
          }}
        >
          {shown.note}
        </div>
      ) : null}
    </Stage>
  );
};
