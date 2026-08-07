import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W, STAGE_H} from './Stage';
import {SceneHeader} from './SceneHeader';

// VersusScene: two contenders across a spine, and one line of advice.
//
// The composition is a single vertical axis with everything hung off it. The
// contenders' names sit at the top, one either side, set right-aligned and
// left-aligned so they lean into the gap between them; the dimension labels
// ride the axis itself in small mono caps; the cells face each other across it.
// A comparison drawn as two independent columns reads as two lists that happen
// to be adjacent — the shared axis is what makes a row a row, so it is the
// strongest line in the frame.
//
// Nothing about the layout moves once the clip starts. The row area is
// reserved at full height from the first frame, so the opener is the two names
// alone above a lot of deliberate emptiness, and each row fills into a slot
// that was already there. That is the difference between rows landing and rows
// pushing: pushing would re-centre the whole composition on every beat, which
// is the single most restless thing a comparison can do.
//
// The two cells of a row do not arrive together. They are staggered by a few
// frames, and the lead alternates down the table — left, right, left — so the
// eye is walked across the spine rather than being asked to take in both sides
// at once. It is a small thing and it is most of why the table reads as a
// sequence of comparisons instead of a block of text appearing.
//
// Colour is the verdict, said quietly and early. A row's winning cell takes a
// tint in the semantic role of its side — accentQuantity on the left,
// accentRival on the right — so by the time the strip lands the viewer has
// already felt the shape of the answer and the words only have to name the
// case. Even rows take no tint at all, which is what makes the tinted ones
// mean something.

const BODY_W = Math.min(STAGE_W, 1560);
const SPINE_W = 210;
const CELL_W = Math.floor((BODY_W - SPINE_W) / 2);
const ROW_H = 76;
const ROW_GAP = 9;
const NAMES_H = 108;
const VERDICT_H = 116;

type Row = {dim: string; leftVal: string; rightVal: string; edge: 'left' | 'right' | 'even'};
type Step = {
  startMs: number;
  endMs: number;
  show: 'face' | 'row' | 'verdict';
  at?: number;
  landed: number[];
};

export const VersusScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const left = String(props.left ?? '');
  const right = String(props.right ?? '');
  const verdict = String(props.verdict ?? '');
  const leftWins = Number(props.leftWins ?? 0);
  const rightWins = Number(props.rightWins ?? 0);
  const rows = (Array.isArray(props.rows) ? props.rows : []) as Row[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (rows.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const landed = new Set(Array.isArray(step.landed) ? step.landed : []);
  const arriving = step.show === 'row' ? (step.at ?? -1) : -1;
  const judged = step.show === 'verdict';

  // The names are large while they are the whole picture and settle back once
  // the table starts filling. Interpolating from the PREVIOUS step's target
  // rather than switching means the change is a movement, not a cut.
  const nameSize = (s: Step) => (s.show === 'face' ? 1 : 0.62);
  const from = idx > 0 ? nameSize(steps[idx - 1]) : nameSize(step);
  const ease = interpolate(sinceStep, [0, 18], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const grand = from + (nameSize(step) - from) * ease;

  const rowsH = rows.length * ROW_H + (rows.length - 1) * ROW_GAP;
  const bodyH = Math.min(STAGE_H - 190, NAMES_H + 34 + rowsH + 30 + VERDICT_H);

  const verdictIn = judged
    ? spring({frame: sinceStep - 3, fps, config: {damping: 14, mass: 0.55}, durationInFrames: 28})
    : 0;

  const nameStyle = (side: 'left' | 'right'): React.CSSProperties => ({
    width: CELL_W + SPINE_W / 2 - 26,
    textAlign: side === 'left' ? 'right' : 'left',
    fontFamily: theme.fontDisplay,
    fontSize: Math.round(40 + 42 * grand),
    fontWeight: 800,
    letterSpacing: -1.4,
    lineHeight: 1.02,
    color: side === 'left' ? theme.accentQuantity : theme.accentRival,
  });

  const cell = (r: Row, i: number, side: 'left' | 'right'): React.ReactNode => {
    const wins = r.edge === side;
    const tint = side === 'left' ? theme.accentQuantity : theme.accentRival;
    // The lead alternates down the table, so the eye is walked across the
    // spine rather than shown both sides at once.
    const leads = i % 2 === 0 ? 'left' : 'right';
    const delay = side === leads ? 3 : 9;
    const enter =
      i === arriving
        ? spring({frame: sinceStep - delay, fps, config: {damping: 13, mass: 0.55}, durationInFrames: 26})
        : 1;
    return (
      <div
        style={{
          width: CELL_W,
          height: ROW_H,
          display: 'flex',
          alignItems: 'center',
          justifyContent: side === 'left' ? 'flex-end' : 'flex-start',
          paddingLeft: side === 'left' ? 26 : 30,
          paddingRight: side === 'left' ? 30 : 26,
          borderRadius: side === 'left' ? '12px 0 0 12px' : '0 12px 12px 0',
          background: wins ? withAlpha(tint, 0.16) : withAlpha(theme.surface, 0.7),
          border: `1px solid ${wins ? withAlpha(tint, 0.5) : theme.surfaceBorder}`,
          opacity: enter,
          transform: `translateX(${(1 - enter) * (side === 'left' ? -26 : 26)}px)`,
        }}
      >
        <span
          style={{
            fontFamily: theme.fontBody,
            fontSize: 25,
            lineHeight: 1.22,
            textAlign: side === 'left' ? 'right' : 'left',
            color: wins ? theme.text : theme.textMuted,
            fontWeight: wins ? 600 : 400,
          }}
        >
          {side === 'left' ? r.leftVal : r.rightVal}
        </span>
      </div>
    );
  };

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

      <div style={{width: BODY_W, height: bodyH, position: 'relative', display: 'flex', flexDirection: 'column'}}>
        {/* The spine: the strongest line in the frame, because it is what
            makes a row a row. */}
        <div
          style={{
            position: 'absolute',
            left: BODY_W / 2 - 1,
            top: NAMES_H - 18,
            width: 2,
            height: 34 + rowsH + 18,
            background: `linear-gradient(180deg, ${withAlpha(theme.line, 0)}, ${withAlpha(theme.line, 0.6)} 12%, ${withAlpha(theme.line, 0.6)} 88%, ${withAlpha(theme.line, 0)})`,
          }}
        />

        {/* The contenders. */}
        <div style={{height: NAMES_H, display: 'flex', alignItems: 'center'}}>
          <div style={nameStyle('left')}>{left}</div>
          <div
            style={{
              width: 52,
              textAlign: 'center',
              fontFamily: theme.fontMono,
              fontSize: 18,
              letterSpacing: 2,
              textTransform: 'uppercase',
              color: theme.textMuted,
            }}
          >
            vs
          </div>
          <div style={nameStyle('right')}>{right}</div>
        </div>

        {/* The rows. The area is reserved from the first frame, so a row
            lands into a slot rather than pushing the composition around. */}
        <div style={{height: rowsH, marginTop: 34, display: 'flex', flexDirection: 'column', gap: ROW_GAP}}>
          {rows.map((r, i) => {
            if (!landed.has(i)) {
              return <div key={i} style={{height: ROW_H}} />;
            }
            const label =
              i === arriving
                ? interpolate(sinceStep, [0, 14], [0, 1], {
                    extrapolateLeft: 'clamp',
                    extrapolateRight: 'clamp',
                  })
                : 1;
            return (
              <div key={i} style={{height: ROW_H, display: 'flex', alignItems: 'center'}}>
                {cell(r, i, 'left')}
                <div
                  style={{
                    width: SPINE_W,
                    textAlign: 'center',
                    fontFamily: theme.fontMono,
                    fontSize: 15,
                    letterSpacing: 2.2,
                    textTransform: 'uppercase',
                    color: theme.textMuted,
                    opacity: label,
                  }}
                >
                  {r.dim}
                </div>
                {cell(r, i, 'right')}
              </div>
            );
          })}
        </div>

        {/* The verdict: the only part of the clip that says what to DO. */}
        <div
          style={{
            height: VERDICT_H,
            marginTop: 30,
            display: 'flex',
            alignItems: 'center',
            gap: 30,
            paddingLeft: 34,
            paddingRight: 34,
            borderRadius: 16,
            background: withAlpha(theme.surface, 0.9),
            borderTop: `3px solid ${theme.accent}`,
            border: `1px solid ${theme.surfaceBorder}`,
            opacity: verdictIn,
            transform: `translateY(${(1 - verdictIn) * 18}px)`,
          }}
        >
          <div
            style={{
              flexShrink: 0,
              fontFamily: theme.fontMono,
              fontSize: 20,
              letterSpacing: 1,
              color: theme.textMuted,
              whiteSpace: 'nowrap',
            }}
          >
            <span style={{color: theme.accentQuantity, fontWeight: 700}}>{leftWins}</span>
            {' — '}
            <span style={{color: theme.accentRival, fontWeight: 700}}>{rightWins}</span>
          </div>
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 33,
              fontWeight: 600,
              lineHeight: 1.22,
              letterSpacing: -0.4,
              color: theme.text,
            }}
          >
            {verdict}
          </div>
        </div>
      </div>
    </Stage>
  );
};
