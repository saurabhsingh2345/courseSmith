import {interpolate, useCurrentFrame} from 'remotion';
import {CodeTrace, TraceValue} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {MotionTokens, bezierEasing, resolveMotion} from '../theme/motion';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_H} from './Stage';

// PythonExecutionViz is the "Python Tutor" scene: on the left the code with the
// executing line highlighted, on the right a live variable panel, and below it
// the program's captured stdout and the call stack. It steps through the
// trace's snapshots across the scene's duration, pulsing each variable as its
// value changes (timing/easing from the course motion tokens).

const valueText = (v: TraceValue): string => v.repr;

/** Map the current frame to a step index, holding the last step to the end. */
const stepForFrame = (frame: number, durationInFrames: number, steps: number): number => {
  if (steps <= 1) return 0;
  const idx = Math.floor(interpolate(frame, [0, Math.max(1, durationInFrames)], [0, steps], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  }));
  return Math.min(steps - 1, Math.max(0, idx));
};

export const PythonExecutionViz: React.FC<{
  theme: ResolvedTheme;
  motion?: MotionTokens;
  durationInFrames: number;
  props: Record<string, unknown>;
}> = ({theme, motion, durationInFrames, props}) => {
  const frame = useCurrentFrame();
  const m = resolveMotion(motion);
  const trace = props.trace as CodeTrace | undefined;
  const title = String(props.title ?? '');

  if (!trace || trace.steps.length === 0) {
    return null;
  }

  const MONO = theme.fontMono;

  const steps = trace.steps;
  const framesPerStep = Math.max(1, durationInFrames / steps.length);
  const idx = stepForFrame(frame, durationInFrames, steps.length);
  const step = steps[idx];
  const prev = idx > 0 ? steps[idx - 1] : undefined;

  // How far into the current step we are (0→1), for the value-change pulse.
  const stepStart = idx * framesPerStep;
  const localT = interpolate(frame, [stepStart, stepStart + framesPerStep], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const ease = bezierEasing(m.easing.entrance);
  const pulse = 1 - ease(Math.min(1, localT)); // 1 at step start → 0 as it settles

  const prevByName = new Map((prev?.vars ?? []).map((v) => [v.name, valueText(v.value)]));

  const lines = trace.lines.length > 0 ? trace.lines : trace.code.split('\n');
  const atError = trace.error && idx === steps.length - 1;

  // Short traces get big type so a two-line program doesn't rattle around a
  // panel scaled for twenty lines.
  const codeFont = lines.length <= 8 ? 34 : lines.length <= 16 ? 28 : 24;
  const stateFont = lines.length <= 8 ? 28 : 24;

  return (
    <Stage>
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={28} />
      <div
        style={{
          display: 'flex',
          gap: 30,
          alignItems: 'center',
          maxHeight: STAGE_H - (title ? 118 : 0),
          minHeight: 0,
          width: '100%',
        }}
      >
        {/* Code pane */}
        <div
          style={{
            flex: '1 1 58%',
            backgroundColor: '#16191f',
            border: `1px solid ${theme.surfaceBorder}`,
            borderRadius: 18,
            padding: '28px 8px',
            maxHeight: 830,
            fontFamily: MONO,
            fontSize: codeFont,
            lineHeight: 1.55,
            overflow: 'hidden',
            boxShadow: '0 34px 90px rgba(0,0,0,0.5)',
          }}
        >
          {lines.map((ln, i) => {
            const active = i + 1 === step.line;
            return (
              <div
                key={i}
                style={{
                  display: 'flex',
                  gap: 20,
                  padding: '3px 24px',
                  backgroundColor: active ? withAlpha(theme.accent, 0.22) : 'transparent',
                  borderLeft: `4px solid ${active ? theme.accent : 'transparent'}`,
                }}
              >
                <span style={{color: '#565f6e', width: 40, textAlign: 'right', userSelect: 'none'}}>
                  {i + 1}
                </span>
                <span style={{color: active ? '#f5f5f7' : '#b8b8c0', whiteSpace: 'pre'}}>
                  {ln || ' '}
                </span>
              </div>
            );
          })}
        </div>

        {/* State pane */}
        <div style={{flex: '1 1 42%', display: 'flex', flexDirection: 'column', gap: 18, minHeight: 0}}>
          <Panel title="Variables" theme={theme}>
            {step.vars.length === 0 ? (
              <Muted>— none yet —</Muted>
            ) : (
              step.vars.map((v) => {
                const changed = prevByName.get(v.name) !== valueText(v.value);
                const scale = changed ? 1 + 0.06 * pulse : 1;
                const glow = changed ? pulse : 0;
                return (
                  <div
                    key={v.name}
                    style={{
                      display: 'flex',
                      alignItems: 'baseline',
                      gap: 12,
                      padding: '6px 10px',
                      borderRadius: 10,
                      transform: `scale(${scale})`,
                      transformOrigin: 'left center',
                      backgroundColor: withAlpha(theme.accent, 0.12 * glow),
                    }}
                  >
                    <span style={{fontFamily: MONO, fontSize: stateFont, color: theme.text, fontWeight: 700}}>
                      {v.name}
                    </span>
                    <span style={{color: theme.textMuted, fontSize: stateFont - 6}}>{v.value.type}</span>
                    <span style={{fontFamily: MONO, fontSize: stateFont, color: theme.accent, marginLeft: 'auto'}}>
                      {renderValue(v.value)}
                    </span>
                  </div>
                );
              })
            )}
          </Panel>

          <Panel title={`Call stack · ${step.func}`} theme={theme}>
            <span style={{fontFamily: MONO, fontSize: stateFont - 2, color: theme.textMuted}}>
              {step.stack.join('  ›  ')}
            </span>
          </Panel>

          <Panel title="Output" theme={theme}>
            <pre
              style={{
                margin: 0,
                fontFamily: MONO,
                fontSize: stateFont - 2,
                color: '#e6edf3',
                whiteSpace: 'pre-wrap',
                minHeight: '1.4em',
              }}
            >
              {step.stdout || ''}
            </pre>
            {atError ? (
              <div style={{marginTop: 10, color: '#ef4444', fontFamily: MONO, fontSize: 20}}>
                {trace.error}
              </div>
            ) : null}
          </Panel>
        </div>
      </div>
    </Stage>
  );
};

const Panel: React.FC<{title: string; theme: ResolvedTheme; grow?: boolean; children: React.ReactNode}> = ({
  title,
  theme,
  grow,
  children,
}) => (
  <div
    style={{
      backgroundColor: theme.surface,
      border: `1px solid ${theme.surfaceBorder}`,
      borderRadius: 18,
      padding: 22,
      display: 'flex',
      flexDirection: 'column',
      gap: 8,
      flex: grow ? 1 : '0 0 auto',
      minHeight: 0,
      overflow: 'hidden',
    }}
  >
    <div style={{fontSize: 18, letterSpacing: 2, textTransform: 'uppercase', color: theme.accent, fontWeight: 700}}>
      {title}
    </div>
    {children}
  </div>
);

const Muted: React.FC<{children: React.ReactNode}> = ({children}) => (
  <span style={{color: 'rgba(255,255,255,0.4)', fontStyle: 'italic', fontSize: 22}}>{children}</span>
);

/** Compact rendering of a value: containers show a couple of elements inline. */
const renderValue = (v: TraceValue): string => {
  if (v.items && v.items.length > 0) {
    return v.repr;
  }
  if (v.entries && v.entries.length > 0) {
    return v.repr;
  }
  return v.repr;
};

/** Overlay a hex color at a given alpha, for highlight backgrounds. */
const withAlpha = (hex: string, alpha: number): string => {
  const h = hex.replace('#', '');
  if (h.length !== 6) return hex;
  const r = parseInt(h.slice(0, 2), 16);
  const g = parseInt(h.slice(2, 4), 16);
  const b = parseInt(h.slice(4, 6), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
};
