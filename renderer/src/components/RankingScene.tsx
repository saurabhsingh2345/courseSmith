import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// RankingScene is an ordered board with rows that slide.
//
// The composition is a single column of rows, absolutely positioned by rank.
// When an arrival lands, every row's target position changes and each one
// animates from where it was to where it now is. Nothing is redrawn in a new
// place, and that is the entire point: a board that cuts between two orderings
// is two boards, and the viewer has to diff them to see what happened. A board
// whose rows travel shows what happened without anyone having to look for it.
//
// Three decisions carry it.
//
// The orderings come from Go (rankingOrder), one per beat, precomputed. The
// renderer never sorts. Two implementations of "which rows are visible and in
// what order" — one deciding the validator's answer and one deciding what is
// drawn — would eventually disagree about a tie, and the disagreement would be
// invisible until a clip shipped with a row in the wrong place.
//
// An arriving row enters from the right rather than fading in place. Fading in
// at its final position makes the arrival look like it was always there and the
// others moved around it; entering from outside says something came in, which is
// what the narrator is saying.
//
// A row pushed off the bottom leaves rather than vanishing. The eviction is
// usually half the point — a sorted set with a cap, a top-five that just lost
// somebody — so the last row exits downward and dims instead of being cut.

/** The board's own box. */
const BOARD_W = Math.min(STAGE_W, 1080);
const ROW_H = 66;
const ROW_GAP = 8;
const PITCH = ROW_H + ROW_GAP;

type Entry = {label: string; value: number; note?: string; role: string; arrival: boolean};
type Step = {
  startMs: number;
  endMs: number;
  show: 'board' | 'insert' | 'read';
  at?: number;
  entered?: number;
  order: number[];
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
      return theme.text;
  }
};

export const RankingScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const metric = String(props.metric ?? '');
  const unit = String(props.unit ?? '');
  const entries = (Array.isArray(props.entries) ? props.entries : []) as Entry[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (entries.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const prevOrder = idx > 0 ? steps[idx - 1].order : step.order;
  // The re-sort. Deliberately unhurried — the whole clip exists for this
  // movement, and a row that snaps into place is a row nobody watched travel.
  const t = interpolate(sinceStep, [4, 26], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // The arriving row lands AFTER the others have parted, not alongside them.
  // Running both at once put the new row on top of the one it was displacing
  // for most of the transition, which reads as a rendering fault rather than as
  // an insertion — the eye cannot tell two overlapping rows from one broken one.
  const enterT = interpolate(sinceStep, [16, 34], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  // Only the opening beat fades the board up. Recomputing this per step made
  // every re-sort start from a board at zero opacity, so the whole picture
  // flashed on each arrival instead of holding still while one row moved.
  const boardIn =
    idx === 0
      ? interpolate(sinceStep, [0, 18], [0, 1], {
          extrapolateLeft: 'clamp',
          extrapolateRight: 'clamp',
        })
      : 1;

  const posOf = (order: number[], e: number) => order.indexOf(e);
  const rows = Math.max(prevOrder.length, step.order.length);

  // The arrival being spoken about, for the caption under the board. On the
  // closing beat the last arrival stays up, so the frame does not end on a
  // picture with nothing naming it.
  const spokenIndex =
    step.entered !== undefined
      ? step.entered
      : steps
          .slice(0, idx + 1)
          .map((s) => s.entered)
          .filter((v): v is number => v !== undefined)
          .pop();
  const spoken = spokenIndex !== undefined ? entries[spokenIndex] : undefined;

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

      {/* What the board is ordered by. It sits above the column because a
          ranked list whose axis is unstated is a list of names. */}
      <div
        style={{
          width: BOARD_W,
          display: 'flex',
          justifyContent: 'space-between',
          fontFamily: theme.fontMono,
          fontSize: 15,
          letterSpacing: 3,
          textTransform: 'uppercase',
          color: theme.textMuted,
          opacity: 0.75,
          marginBottom: 14,
          paddingInline: 4,
        }}
      >
        <span>rank</span>
        <span>{unit ? `${metric} · ${unit}` : metric}</span>
      </div>

      <div style={{width: BOARD_W, height: rows * PITCH, position: 'relative'}}>
        {entries.map((entry, e) => {
          const prev = posOf(prevOrder, e);
          const cur = posOf(step.order, e);
          if (prev < 0 && cur < 0) return null;

          const entering = prev < 0 && cur >= 0;
          const leaving = prev >= 0 && cur < 0;
          // A leaving row keeps travelling in the direction it was going, one
          // place past the bottom, so the exit reads as being pushed off rather
          // than as a glitch.
          const fromY = prev >= 0 ? prev * PITCH : cur * PITCH;
          const toY = cur >= 0 ? cur * PITCH : rows * PITCH;
          const y = interpolate(t, [0, 1], [fromY, toY]);

          const opacity = entering
            ? enterT
            : leaving
              ? 1 - t
              : boardIn;
          // Entering rows come in from the right; the offset is spent by the
          // time the row reaches its rank.
          const x = entering ? (1 - enterT) * 70 : 0;

          const isCurrent = step.entered === e;
          const colour = roleColour(theme, entry.role);
          // A row on its way off carries no rank at all. It cannot keep the one
          // it held — the row rising into that place is taking the same number,
          // and two rows both labelled 05 reads as the board having two fifth
          // places rather than as one leaving. It cannot take the one it is
          // going to either, because it is going nowhere.
          const rank = cur >= 0 ? String(cur + 1).padStart(2, '0') : '';

          return (
            <div
              key={e}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                height: ROW_H,
                transform: `translate3d(${x}px, ${y}px, 0)`,
                opacity,
                display: 'flex',
                alignItems: 'center',
                gap: 22,
                paddingInline: 22,
                borderRadius: 12,
                background: isCurrent ? withAlpha(colour, 0.12) : theme.surface,
                border: `1px solid ${isCurrent ? withAlpha(colour, 0.55) : theme.surfaceBorder}`,
                boxShadow: isCurrent ? `0 0 34px ${withAlpha(colour, 0.22)}` : undefined,
              }}
            >
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 20,
                  color: isCurrent ? colour : theme.textMuted,
                  fontVariantNumeric: 'tabular-nums',
                  width: 44,
                }}
              >
                {rank}
              </span>
              <span
                style={{
                  flex: 1,
                  fontFamily: theme.fontBody,
                  fontSize: 27,
                  fontWeight: isCurrent ? 700 : 500,
                  color: isCurrent ? theme.text : theme.text,
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                }}
              >
                {entry.label}
              </span>
              <span
                style={{
                  fontFamily: theme.fontDisplay,
                  fontSize: 29,
                  fontWeight: 700,
                  color: isCurrent ? colour : theme.textMuted,
                  fontVariantNumeric: 'tabular-nums',
                }}
              >
                {entry.value.toLocaleString()}
              </span>
            </div>
          );
        })}
      </div>

      {spoken?.note ? (
        <div
          style={{
            marginTop: 30,
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
          {spoken.note}
        </div>
      ) : null}
    </Stage>
  );
};
