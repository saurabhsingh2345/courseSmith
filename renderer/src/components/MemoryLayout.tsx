import {AbsoluteFill, interpolate, useCurrentFrame} from 'remotion';
import {CodeTrace, TraceValue, TraceVar} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {MotionTokens, bezierEasing, resolveMotion} from '../theme/motion';

// MemoryLayout is the "stack & heap" companion to PythonExecutionViz: the left
// column is the call stack with each frame's variable slots; the right column is
// the heap, one card per mutable object (list / dict / instance). Arrows connect
// a variable slot to the object it references. Everything is driven by
// useCurrentFrame so it renders deterministically in `remotion render` — no
// WebGL, no requestAnimationFrame. New objects scale in and changed values pulse
// using the course motion tokens, matching the execution viz.
//
// Note: the trace records values, not object identity, so two variables bound to
// the same object are drawn as two cards (no alias sharing). That's a limitation
// of the trace schema, not this view.



// Canvas geometry (composition is 1920×1080). Fixed coordinates let the SVG
// connector layer compute arrow endpoints without measuring the DOM.
const STACK_X = 96;
const STACK_W = 620;
const HEAP_X = 1150;
const HEAP_W = 660;
const ROW_TOP = 250;
const SLOT_H = 72;
const SLOT_GAP = 18;
const CARD_GAP = 26;

// This scene positions everything absolutely against the raw 1920x1080 frame,
// so it can't compose into <Stage>. It honours the same reservation by hand:
// nothing may be laid out below FRAME_H - CAPTION_SAFE, where CaptionTrack
// draws. At the natural 90px pitch that is exactly 7 slots; past that the
// pitch compresses rather than running under the captions.
const FLOOR = 1080 - 200;
const NATURAL_PITCH = SLOT_H + SLOT_GAP;

/** Vertical pitch that keeps `n` slots between ROW_TOP and FLOOR. */
const slotPitch = (n: number): number =>
  n <= 1 ? NATURAL_PITCH : Math.min(NATURAL_PITCH, (FLOOR - ROW_TOP - SLOT_H) / (n - 1));

const isContainer = (v: TraceVar): boolean =>
  !!(v.value.items?.length || v.value.entries?.length || v.value.fields?.length);

const slotY = (i: number, count: number): number => ROW_TOP + i * slotPitch(count);

/** Height a heap card needs for the object it holds. */
const cardHeight = (v: TraceValue): number => {
  if (v.items && v.items.length) {
    const shown = Math.min(v.items.length, 12);
    const rows = Math.ceil(shown / 5);
    return 96 + rows * 52;
  }
  const rows = Math.min((v.entries?.length ?? v.fields?.length ?? 0) || 1, 6);
  return 96 + rows * 36;
};

/** Map the current frame to a step index, holding the last step to the end. */
const stepForFrame = (frame: number, durationInFrames: number, steps: number): number => {
  if (steps <= 1) return 0;
  const idx = Math.floor(
    interpolate(frame, [0, Math.max(1, durationInFrames)], [0, steps], {
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp',
    }),
  );
  return Math.min(steps - 1, Math.max(0, idx));
};

type CardLayout = {name: string; value: TraceValue; y: number; h: number; anchorY: number};

export const MemoryLayout: React.FC<{
  theme: ResolvedTheme;
  motion?: MotionTokens;
  durationInFrames: number;
  props: Record<string, unknown>;
}> = ({theme, motion, durationInFrames, props}) => {
  const frame = useCurrentFrame();
  const m = resolveMotion(motion);
  const trace = props.trace as CodeTrace | undefined;
  const title = String(props.title ?? 'Memory');

  if (!trace || trace.steps.length === 0) {
    return <AbsoluteFill />;
  }

  const steps = trace.steps;
  const framesPerStep = Math.max(1, durationInFrames / steps.length);
  const idx = stepForFrame(frame, durationInFrames, steps.length);
  const step = steps[idx];
  const prev = idx > 0 ? steps[idx - 1] : undefined;

  // Progress within the current step (0→1) for entrance/pulse animation.
  const stepStart = idx * framesPerStep;
  const localT = interpolate(frame, [stepStart, stepStart + framesPerStep], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const ease = bezierEasing(m.easing.entrance);
  const settle = ease(Math.min(1, localT)); // 0 at step start → 1 as it settles
  const pulse = 1 - settle;

  const prevByName = new Map((prev?.vars ?? []).map((v) => [v.name, v.value.repr]));
  const prevContainers = new Set(
    (prev?.vars ?? []).filter(isContainer).map((v) => v.name),
  );

  const containers = step.vars.filter(isContainer);

  // Lay heap cards out top-to-bottom, remembering each object's anchor point so
  // an arrow from its variable slot lands on the card header.
  const cards: CardLayout[] = [];
  let cursor = ROW_TOP;
  for (const v of containers) {
    const h = cardHeight(v.value);
    cards.push({name: v.name, value: v.value, y: cursor, h, anchorY: cursor + 42});
    cursor += h + CARD_GAP;
  }
  const cardByName = new Map(cards.map((c) => [c.name, c]));

  return (
    <AbsoluteFill
      style={{
        display: 'flex',
        flexDirection: 'column',
        padding: 64,
        fontFamily: theme.fontBody,
      }}
    >
      <div style={{fontFamily: theme.fontDisplay, fontSize: 40, fontWeight: 700, letterSpacing: -0.5, color: theme.text, marginBottom: 8}}>{title}</div>
      <div style={{display: 'flex', gap: 24, fontSize: 20, color: theme.textMuted, marginBottom: 4}}>
        <span>
          line <strong style={{color: theme.accent}}>{step.line}</strong>
        </span>
        <span>·</span>
        <span style={{fontFamily: theme.fontMono}}>{step.stack.join('  ›  ')}</span>
      </div>

      {/* Column headers */}
      <div style={{position: 'absolute', left: STACK_X, top: 190, width: STACK_W}}>
        <ColumnHeader label="Call stack" theme={theme} />
      </div>
      <div style={{position: 'absolute', left: HEAP_X, top: 190, width: HEAP_W}}>
        <ColumnHeader label="Heap" theme={theme} />
      </div>

      {/* Connector layer (behind the boxes) */}
      <svg
        width={1920}
        height={1080}
        style={{position: 'absolute', left: 0, top: 0, pointerEvents: 'none'}}
      >
        <defs>
          <marker id="mem-arrow" markerWidth="12" markerHeight="12" refX="9" refY="4" orient="auto">
            <path d="M0,0 L9,4 L0,8 Z" fill={theme.accent} />
          </marker>
        </defs>
        {step.vars.map((v, i) => {
          if (!isContainer(v)) return null;
          const card = cardByName.get(v.name);
          if (!card) return null;
          const x1 = STACK_X + STACK_W - 26;
          const y1 = slotY(i, step.vars.length) + SLOT_H / 2;
          const x2 = HEAP_X + 6;
          const y2 = card.anchorY;
          const isNew = !prevContainers.has(v.name);
          const draw = isNew ? settle : 1;
          const mx = (x1 + x2) / 2;
          const d = `M ${x1} ${y1} C ${mx} ${y1}, ${mx} ${y2}, ${x2} ${y2}`;
          return (
            <path
              key={v.name}
              d={d}
              fill="none"
              stroke={theme.accent}
              strokeWidth={3}
              strokeOpacity={0.55 * draw}
              markerEnd="url(#mem-arrow)"
              strokeDasharray={1000}
              strokeDashoffset={1000 * (1 - draw)}
            />
          );
        })}
      </svg>

      {/* Stack column: one slot per variable in the current frame */}
      {step.vars.map((v, i) => {
        const container = isContainer(v);
        const changed = prevByName.get(v.name) !== v.value.repr;
        const scale = changed ? 1 + 0.05 * pulse : 1;
        return (
          <div
            key={v.name}
            style={{
              position: 'absolute',
              left: STACK_X,
              top: slotY(i, step.vars.length),
              width: STACK_W,
              height: SLOT_H,
              display: 'flex',
              alignItems: 'center',
              gap: 14,
              padding: '0 22px',
              boxSizing: 'border-box',
              backgroundColor: theme.surface,
              border: `1px solid ${changed ? theme.accent : theme.surfaceBorder}`,
              borderRadius: 14,
              boxShadow: '0 10px 26px rgba(0,0,0,0.35)',
              transform: `scale(${scale})`,
              transformOrigin: 'left center',
            }}
          >
            <span style={{fontFamily: theme.fontMono, fontSize: 26, color: theme.text, fontWeight: 700}}>
              {v.name}
            </span>
            <span style={{fontSize: 17, color: theme.textMuted}}>{v.value.type}</span>
            <span
              style={{
                marginLeft: 'auto',
                fontFamily: theme.fontMono,
                fontSize: 24,
                color: container ? theme.accent : '#e6e6ec',
                fontWeight: container ? 700 : 400,
              }}
            >
              {container ? '●' : v.value.repr}
            </span>
          </div>
        );
      })}

      {/* Heap column: one card per mutable object */}
      {cards.map((card) => {
        const isNew = !prevContainers.has(card.name);
        const appear = isNew ? settle : 1;
        return (
          <div
            key={card.name}
            style={{
              position: 'absolute',
              left: HEAP_X,
              top: card.y,
              width: HEAP_W,
              height: card.h,
              boxSizing: 'border-box',
              backgroundColor: '#0e0e11',
              borderRadius: 16,
              padding: 20,
              color: '#f5f5f7',
              boxShadow: '0 20px 50px rgba(0,0,0,0.22)',
              opacity: appear,
              transform: `scale(${0.94 + 0.06 * appear})`,
              transformOrigin: 'left center',
            }}
          >
            <div style={{display: 'flex', alignItems: 'baseline', gap: 12, marginBottom: 12}}>
              <span style={{fontFamily: theme.fontMono, fontSize: 22, color: theme.accent, fontWeight: 700}}>
                {card.name}
              </span>
              <span style={{fontSize: 16, color: '#8b8b95', letterSpacing: 1, textTransform: 'uppercase'}}>
                {card.value.type}
              </span>
            </div>
            <HeapBody value={card.value} accent={theme.accent} />
          </div>
        );
      })}
    </AbsoluteFill>
  );
};

const ColumnHeader: React.FC<{label: string; theme: ResolvedTheme}> = ({label, theme}) => (
  <div
    style={{
      fontSize: 18,
      letterSpacing: 2,
      textTransform: 'uppercase',
      color: theme.accent,
      fontWeight: 700,
    }}
  >
    {label}
  </div>
);

const HEAP_MONO = '"JetBrains Mono", Menlo, Consolas, monospace';

/** Renders the contents of a heap object: list chips, or dict/object rows. */
const HeapBody: React.FC<{value: TraceValue; accent: string}> = ({value, accent}) => {
  if (value.items && value.items.length) {
    const shown = value.items.slice(0, 12);
    return (
      <div style={{display: 'flex', flexWrap: 'wrap', gap: 10}}>
        {shown.map((it, i) => (
          <div
            key={i}
            style={{
              minWidth: 56,
              padding: '8px 12px',
              borderRadius: 10,
              backgroundColor: '#1e1e23',
              border: '1px solid #33333b',
              textAlign: 'center',
            }}
          >
            <div style={{fontSize: 13, color: '#71717c'}}>{i}</div>
            <div style={{fontFamily: HEAP_MONO, fontSize: 22, color: '#e6e6ec'}}>{it.repr}</div>
          </div>
        ))}
        {value.items.length > shown.length ? (
          <div style={{alignSelf: 'center', color: '#71717c', fontSize: 20}}>
            +{value.items.length - shown.length}
          </div>
        ) : null}
      </div>
    );
  }
  const rows = value.entries ?? value.fields ?? [];
  if (rows.length) {
    return (
      <div style={{display: 'flex', flexDirection: 'column', gap: 8}}>
        {rows.slice(0, 6).map((r, i) => (
          <div key={i} style={{display: 'flex', gap: 12, fontFamily: HEAP_MONO, fontSize: 22}}>
            <span style={{color: accent}}>{r.key}</span>
            <span style={{color: '#71717c'}}>→</span>
            <span style={{color: '#e6e6ec'}}>{r.value.repr}</span>
          </div>
        ))}
        {rows.length > 6 ? <span style={{color: '#71717c', fontSize: 18}}>+{rows.length - 6} more</span> : null}
      </div>
    );
  }
  return <span style={{fontFamily: HEAP_MONO, fontSize: 22, color: '#e6e6ec'}}>{value.repr}</span>;
};
