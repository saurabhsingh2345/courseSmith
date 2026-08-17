import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';
import {AppWindow, WINDOW_BAR_H, windowDim, windowText} from './AppWindow';

// PatchScene draws one change at a time, big enough to read.
//
// The measurements are the design here, so they are stated rather than tuned by
// eye. A unified diff in the tool this borrows from is set around 14 logical
// pixels; this sets it at 30, which is a little over twice the size, and that one
// number is the whole reason the template exists. Everything else follows from
// affording it: the window is 1500 wide, which at 30px mono is about 78 columns,
// and Go clamps lines at 62 so there is room for the gutter and the sign.
//
// What moves: exactly one thing per beat. The removed line is struck through where
// it sits and dims; the added line arrives under it, sliding up a few pixels as it
// fades in. Nothing else on the frame changes. A diff where the whole block
// re-flows on every beat is the scrolling wall this replaces.
//
// The tally under the window is not decoration either. "How big is this change" is
// the question a raw diff makes you scroll to the end to answer, and it is the one
// that decides whether anybody reads the diff at all — so it is on screen the whole
// time and it grows as hunks land.

const WINDOW_W = Math.min(STAGE_W, 1500);
const CODE = 30;
const ROW_H = Math.round(CODE * 1.5);
/** Padding above and below the rows, inside the window body. */
const BODY_PAD = 26;

/**
 * The window's height, measured from the TALLEST hunk in the patch.
 *
 * Fixed at 560 first, and 200 pixels of empty black under a three-line hunk is
 * not a neutral choice — a dark panel with nothing in the bottom third reads as
 * output that has been cut off, which is the opposite of what this frame is
 * claiming. Measured per hunk instead would be worse: the window would change
 * height between beats, and the whole family's rule is that the arrangement holds
 * still while the light moves. So it is sized once, for the largest hunk, and
 * every smaller one leaves a little air.
 */
const windowHeight = (hunks: Hunk[]): number => {
  const rows = Math.max(
    ...hunks.map(
      (h) => (h.context?.length ?? 0) + (h.before?.length ?? 0) + (h.after?.length ?? 0),
    ),
    3,
  );
  return WINDOW_BAR_H + BODY_PAD * 2 + rows * ROW_H;
};

type Hunk = {
  at?: number;
  before?: string[];
  after?: string[];
  context?: string[];
  note?: string;
};

type Step = {
  startMs: number;
  endMs: number;
  show: 'file' | 'hunk' | 'tally';
  at?: number;
  landed?: number;
  added?: number;
  removed?: number;
};

/** One line of the diff, with its gutter number and its sign. */
const Line: React.FC<{
  theme: ResolvedTheme;
  n?: number;
  sign?: '+' | '-';
  text: string;
  /** 0..1 — how far this line has arrived. Context lines are always 1. */
  on: number;
  struck?: boolean;
}> = ({theme, n, sign, text, on, struck}) => {
  const band =
    sign === '+' ? 'rgba(46, 160, 67, 0.16)' : sign === '-' ? 'rgba(248, 81, 73, 0.15)' : 'transparent';
  const ink = sign === '+' ? '#7ee787' : sign === '-' ? '#ffa198' : '#d7dae1';
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        height: ROW_H,
        background: band,
        opacity: on,
        transform: `translateY(${(1 - on) * 10}px)`,
      }}
    >
      <div
        style={{
          ...windowText(theme, CODE),
          width: 108,
          flex: 'none',
          textAlign: 'right',
          paddingRight: 20,
          color: windowDim(),
        }}
      >
        {n ?? ''}
      </div>
      <div
        style={{
          ...windowText(theme, CODE),
          width: 34,
          flex: 'none',
          color: sign ? ink : 'transparent',
        }}
      >
        {sign ?? ' '}
      </div>
      <div
        style={{
          ...windowText(theme, CODE),
          color: struck ? withAlpha(ink, 0.75) : ink,
          // The strikethrough is the template's one true animation: it says the
          // line is GONE rather than just dimmer, which is the difference between
          // reading a diff and watching a change.
          textDecoration: struck ? 'line-through' : 'none',
          textDecorationThickness: struck ? 3 : undefined,
        }}
      >
        {text}
      </div>
    </div>
  );
};

/** A `+A −D` chip pair. */
const Tally: React.FC<{theme: ResolvedTheme; added: number; removed: number; files: number}> = ({
  theme,
  added,
  removed,
  files,
}) => (
  <div style={{display: 'flex', alignItems: 'center', gap: 18}}>
    <div
      style={{
        fontFamily: theme.fontBody,
        fontSize: 26,
        fontWeight: 600,
        color: theme.textMuted,
      }}
    >
      {files} file{files === 1 ? '' : 's'}
    </div>
    <div style={{width: 6, height: 6, borderRadius: 999, background: withAlpha(theme.text, 0.25)}} />
    <div style={{fontFamily: theme.fontMono, fontSize: 28, fontWeight: 700, color: '#2ea043'}}>
      +{added}
    </div>
    <div style={{fontFamily: theme.fontMono, fontSize: 28, fontWeight: 700, color: '#d1242f'}}>
      −{removed}
    </div>
  </div>
);

export const PatchScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const path = String(props.path ?? '');
  const closer = String(props.closer ?? '');
  const hunks = (Array.isArray(props.hunks) ? props.hunks : []) as Hunk[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (hunks.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const since = ((nowMs - step.startMs) / 1000) * FPS;

  // Which hunk is on screen. On the opening and closing beats it is the first and
  // the last, so the window is never empty and never mid-air.
  const shown = step.show === 'hunk' ? (step.at ?? 0) : step.show === 'tally' ? hunks.length - 1 : 0;
  const hunk = hunks[Math.max(0, Math.min(hunks.length - 1, shown))];

  // The change arrives over about half a second. Slower than the light changes
  // elsewhere in this family: the viewer is reading, not registering.
  const applied =
    step.show === 'file'
      ? 0
      : interpolate(since, [4, 20], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});
  const struck = applied > 0.3;

  const at = hunk.at ?? 1;
  const context = hunk.context ?? [];
  const before = hunk.before ?? [];
  const after = hunk.after ?? [];

  return (
    <Stage justify="center">
      <div style={{width: WINDOW_W}}>
        <SceneHeader
          theme={theme}
          title={String(props.title ?? '')}
          emphasis={props.emphasis as string | undefined}
          emphasisRole={props.emphasisRole as string | undefined}
          size="compact"
          marginBottom={38}
        />

        <AppWindow
          theme={theme}
          title={path}
          badge="◆"
          width={WINDOW_W}
          height={windowHeight(hunks)}
          // The rail is how far through the change we are, which is a real fact
          // about this clip rather than a borrowed decoration.
          progress={(step.landed ?? 0) / hunks.length}
        >
          <div style={{padding: `${BODY_PAD}px 0`}}>
            {context.map((l, i) => (
              <Line key={`c${i}`} theme={theme} n={at + i} text={l} on={1} />
            ))}
            {before.map((l, i) => (
              <Line
                key={`b${i}`}
                theme={theme}
                n={at + context.length + i}
                sign="-"
                text={l}
                on={1}
                struck={struck}
              />
            ))}
            {after.map((l, i) => (
              <Line
                key={`a${i}`}
                theme={theme}
                n={at + context.length + before.length + i}
                sign="+"
                text={l}
                on={applied}
              />
            ))}
          </div>
        </AppWindow>

        {/* The note and the tally share one row: the reason on the left, the size
            on the right. Fixed height so the window never moves when the note
            changes length between hunks. */}
        <div
          style={{
            marginTop: 30,
            minHeight: 96,
            display: 'flex',
            alignItems: 'flex-start',
            justifyContent: 'space-between',
            gap: 40,
          }}
        >
          <div
            style={{
              flex: 1,
              fontFamily: theme.fontBody,
              fontSize: 30,
              lineHeight: 1.35,
              fontWeight: 500,
              color: theme.text,
              opacity: step.show === 'file' ? 0 : applied,
            }}
          >
            {step.show === 'tally' && closer ? closer : hunk.note}
          </div>
          <Tally
            theme={theme}
            files={1}
            added={step.added ?? 0}
            removed={step.removed ?? 0}
          />
        </div>
      </div>
    </Stage>
  );
};
