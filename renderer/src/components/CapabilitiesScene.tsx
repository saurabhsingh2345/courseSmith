import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// CapabilitiesScene is a boundary with the denied things outside it.
//
// The subject sits in the middle inside a drawn line, and the capabilities are
// chips arranged around it — outside the line. That placement is the whole
// picture: a viewer reads "these are not in there" from the geometry before any
// label is read, which is what a list of features can never do.
//
// A granted capability crosses the line: its chip moves inward, changes from the
// limit colour to its own role colour, loses its ✗, and gains a connector to the
// subject. A denied one keeps its ✗ and stays outside.
//
// Two decisions carry the rest.
//
// The chips are laid out in two columns flanking the subject rather than in a
// ring. A ring is the more obvious drawing of "around", and it is worse: chips
// at the top and bottom of a circle are much further from the subject than the
// ones at the sides, so a ring encodes a distance difference that means nothing.
// Two columns give every capability the same standing.
//
// Denied chips are drawn in the limit colour, not greyed out. Grey reads as
// "not applicable"; red reads as "refused", and refused is what the frame means.
// It also keeps the closing beat — the one that lands on what is STILL shut —
// carrying colour rather than looking like an unfinished diagram.

const CHIP_W = 246;
const CHIP_H = 66;
const CHIP_GAP = 16;
const CORE_W = 320;
const LAYOUT_W = Math.min(STAGE_W, 1220);

type Item = {label: string; note?: string; role: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'sealed' | 'grant' | 'read';
  at?: number;
  granted: number[];
};

const roleColour = (theme: ResolvedTheme, role: string): string => {
  switch (role) {
    case 'limit':
      return theme.accentLimit;
    case 'rival':
      return theme.accentRival;
    case 'quantity':
      return theme.accentQuantity;
    default:
      return theme.textMuted;
  }
};

export const CapabilitiesScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const subject = String(props.subject ?? '');
  const subjectNote = String(props.subjectNote ?? '');
  const boundary = String(props.boundary ?? '');
  const granter = String(props.granter ?? '');
  const items = (Array.isArray(props.items) ? props.items : []) as Item[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (items.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const granted = new Set(step.granted ?? []);
  const enter = interpolate(sinceStep, [0, 18], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // The crossing. Slow enough to read as a decision rather than a state change:
  // something outside the line chose to pass this in.
  const cross = interpolate(sinceStep, [4, 24], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  // Split into two columns, left taking the first half. Two columns rather than
  // a ring so every capability is the same distance from the subject.
  const half = Math.ceil(items.length / 2);
  const columns: number[][] = [
    items.map((_, i) => i).slice(0, half),
    items.map((_, i) => i).slice(half),
  ];

  const spoken = step.at !== undefined ? items[step.at] : undefined;
  const note =
    spoken?.note ??
    // On the closing beat, lead with something still denied — that is the frame
    // the template wants screenshotted.
    (step.show === 'read'
      ? items.find((_, i) => !granted.has(i))?.note
      : undefined);

  const chip = (i: number, side: 'left' | 'right') => {
    const item = items[i];
    const isGranted = granted.has(i);
    const isCurrent = step.at === i;
    const colour = isGranted ? roleColour(theme, item.role) : theme.accentLimit;
    const t = isCurrent ? cross : 1;
    // Granted chips move toward the subject; denied ones stay out.
    const inward = isGranted ? (side === 'left' ? 1 : -1) * 26 * t : 0;
    return (
      <div
        key={i}
        style={{
          width: CHIP_W,
          height: CHIP_H,
          marginBottom: CHIP_GAP,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          paddingInline: 16,
          borderRadius: 10,
          position: 'relative',
          transform: `translateX(${inward}px)`,
          background: withAlpha(colour, isGranted ? 0.13 : 0.07),
          border: `1px solid ${withAlpha(colour, isGranted ? 0.55 : 0.34)}`,
        }}
      >
        {/* The wire into the boundary. Moving the chip inward was not enough on
            its own: the gap to the subject is two hundred pixels, so a
            twenty-six pixel shift read as a chip slightly out of line rather
            than as something that had crossed. The wire is what makes the grant
            legible, and only a granted chip has one. */}
        {isGranted ? (
          <div
            style={{
              position: 'absolute',
              top: '50%',
              [side === 'left' ? 'left' : 'right']: CHIP_W - 2,
              width: 128 * t,
              height: 2,
              background: withAlpha(colour, 0.5),
            }}
          />
        ) : null}
        <span
          style={{
            fontFamily: theme.fontMono,
            fontSize: 19,
            letterSpacing: 1.4,
            textTransform: 'uppercase',
            color: colour,
          }}
        >
          {item.label}
        </span>
        <span
          style={{
            fontFamily: theme.fontMono,
            fontSize: isGranted ? 15 : 22,
            color: colour,
            opacity: 0.9,
          }}
        >
          {/* A denied capability keeps its cross for the whole clip. Grey would
              read as "not applicable"; the cross reads as refused. */}
          {isGranted ? 'granted' : '✗'}
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
        marginBottom={22}
      />

      {granter ? (
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 15,
            letterSpacing: 2.4,
            textTransform: 'uppercase',
            color: theme.textMuted,
            opacity: enter * 0.7,
            marginBottom: 16,
          }}
        >
          granted by {granter}
        </div>
      ) : null}

      <div
        style={{
          width: LAYOUT_W,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          opacity: enter,
        }}
      >
        <div>{columns[0].map((i) => chip(i, 'left'))}</div>

        {/* The subject, inside the line. */}
        <div
          style={{
            width: CORE_W,
            padding: '30px 22px',
            borderRadius: 16,
            textAlign: 'center',
            background: theme.surface,
            // The boundary is a solid line, not dashed. A dashed border reads as
            // provisional, and the one thing this line is not is negotiable.
            border: `2px solid ${withAlpha(theme.accentRival, 0.6)}`,
            boxShadow: `0 0 44px ${withAlpha(theme.accentRival, 0.14)}`,
          }}
        >
          <div
            style={{
              fontFamily: theme.fontDisplay,
              fontSize: 32,
              fontWeight: 800,
              color: theme.text,
              letterSpacing: -0.4,
            }}
          >
            {subject}
          </div>
          {subjectNote ? (
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 15,
                color: theme.textMuted,
                marginTop: 6,
              }}
            >
              {subjectNote}
            </div>
          ) : null}
          {boundary ? (
            <div
              style={{
                marginTop: 14,
                display: 'inline-block',
                paddingInline: 10,
                paddingBlock: 4,
                borderRadius: 6,
                fontFamily: theme.fontMono,
                fontSize: 12,
                letterSpacing: 1.8,
                textTransform: 'uppercase',
                color: theme.accentRival,
                border: `1px solid ${withAlpha(theme.accentRival, 0.4)}`,
              }}
            >
              {boundary}
            </div>
          ) : null}
          {/* The count, which states the gap the template exists for. */}
          <div
            style={{
              marginTop: 16,
              fontFamily: theme.fontMono,
              fontSize: 14,
              letterSpacing: 1.6,
              textTransform: 'uppercase',
              color: theme.textMuted,
            }}
          >
            {granted.size} of {items.length} granted
          </div>
        </div>

        <div>{columns[1].map((i) => chip(i, 'right'))}</div>
      </div>

      {note ? (
        <div
          style={{
            marginTop: 26,
            maxWidth: 1040,
            textAlign: 'center',
            fontFamily: theme.fontBody,
            fontSize: 24,
            color: theme.textMuted,
            opacity: interpolate(sinceStep, [10, 24], [0, 1], {
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
