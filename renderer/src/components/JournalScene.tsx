import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// JournalScene is an append-only file: a numbered gutter, lines arriving at the
// bottom, and a cursor that later walks back down from the top.
//
// The panel is drawn at full height from the first frame, with every line's slot
// present and empty. That is deliberate and it is the whole reason the picture
// teaches: a panel that grew as lines arrived would say "the file gets bigger",
// which is true and uninteresting. A panel of fixed height with lines filling
// downward says "this only ever appends", which is the property.
//
// Two decisions carry the rest.
//
// A line that has not been written yet is a dash in the gutter and nothing else.
// Not a greyed-out preview of its text: the viewer would read ahead, and the
// arrival of a line they have already read is not an arrival. `written` comes
// from Go per beat, so the renderer never counts backwards through the steps.
//
// The replay cursor is a filled bar behind the whole line rather than a caret.
// A caret says "the text is being typed here"; a bar says "this line is
// happening now", which is what a replay does — the line is not being written
// again, it is being executed again.

const PANEL_W = Math.min(STAGE_W, 1020);
const LINE_H = 46;
const GUTTER_W = 62;

type Entry = {text: string; note?: string; role: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'file' | 'append' | 'replay' | 'read';
  at?: number;
  written: number;
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

export const JournalScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const file = String(props.file ?? '');
  const writeLabel = String(props.writeLabel ?? 'appending');
  const replayLabel = String(props.replayLabel ?? 'replaying — top to bottom');
  const entries = (Array.isArray(props.entries) ? props.entries : []) as Entry[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (entries.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // Once a replay has started the panel is in its second half for good, so the
  // phase label does not flicker back on a trailing "read" beat.
  const replaying = steps.slice(0, idx + 1).some((s) => s.show === 'replay');
  const cursor = replaying
    ? steps
        .slice(0, idx + 1)
        .map((s) => (s.show === 'replay' ? s.at : undefined))
        .filter((v): v is number => v !== undefined)
        .pop()
    : undefined;

  const panelIn = interpolate(sinceStep, [0, 16], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const arriving = step.show === 'append' ? step.at : undefined;
  // The arriving line lands quickly — an append is not a thing that takes time,
  // and drawing it slowly would suggest otherwise.
  const arriveT = interpolate(sinceStep, [2, 14], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const cursorT = interpolate(sinceStep, [0, 12], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const note = cursor !== undefined ? entries[cursor]?.note : undefined;

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={28}
      />

      {/* The file's name and which half of the clip we are in. The phase label
          is the only thing on the frame that tells a viewer joining mid-clip
          whether lines are being written or read. */}
      <div
        style={{
          width: PANEL_W,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'baseline',
          fontFamily: theme.fontMono,
          fontSize: 15,
          letterSpacing: 2.6,
          textTransform: 'uppercase',
          marginBottom: 12,
          paddingInline: 4,
        }}
      >
        <span style={{color: theme.textMuted, opacity: 0.85}}>{file}</span>
        <span
          style={{
            color: replaying ? theme.accentQuantity : theme.textMuted,
            opacity: replaying ? 0.95 : 0.7,
          }}
        >
          {replaying ? replayLabel : writeLabel}
        </span>
      </div>

      <div
        style={{
          width: PANEL_W,
          background: theme.surface,
          border: `1px solid ${theme.surfaceBorder}`,
          borderRadius: 14,
          paddingBlock: 16,
          opacity: panelIn,
        }}
      >
        {entries.map((entry, i) => {
          const exists = i < step.written;
          const isArriving = arriving === i;
          const isCursor = cursor === i;
          const colour = roleColour(theme, entry.role);
          // An arriving line fades up in place; everything already written sits
          // at full strength, and a line the cursor has passed dims back so the
          // current one is the only lit row.
          const textOpacity = isArriving
            ? arriveT
            : !exists
              ? // The dash is drawn, faintly. At zero the slot is blank and the
                // panel reads as a short file with odd padding; at a quarter it
                // reads as a file with room in it, which is what an append-only
                // log always has.
                0.28
              : replaying && !isCursor
                ? 0.42
                : 1;

          return (
            <div
              key={i}
              style={{
                height: LINE_H,
                display: 'flex',
                alignItems: 'center',
                paddingInline: 18,
                position: 'relative',
              }}
            >
              {/* The replay bar, behind the line rather than beside it. */}
              {isCursor ? (
                <div
                  style={{
                    position: 'absolute',
                    left: 8,
                    right: 8,
                    top: 3,
                    bottom: 3,
                    borderRadius: 8,
                    background: withAlpha(colour, 0.16 * cursorT),
                    border: `1px solid ${withAlpha(colour, 0.5 * cursorT)}`,
                  }}
                />
              ) : null}
              <span
                style={{
                  width: GUTTER_W,
                  flexShrink: 0,
                  fontFamily: theme.fontMono,
                  fontSize: 19,
                  color: isCursor ? colour : theme.textMuted,
                  opacity: exists ? 0.9 : 0.35,
                  fontVariantNumeric: 'tabular-nums',
                  position: 'relative',
                }}
              >
                {String(i + 1).padStart(2, '0')}
              </span>
              <span
                style={{
                  fontFamily: theme.fontMono,
                  fontSize: 24,
                  color: isCursor ? colour : theme.text,
                  opacity: textOpacity,
                  whiteSpace: 'pre',
                  position: 'relative',
                }}
              >
                {/* An unwritten line is a dash, not greyed-out text: a viewer who
                    can read ahead is a viewer for whom the line never arrives. */}
                {exists ? entry.text : '—'}
              </span>
            </div>
          );
        })}
      </div>

      {note ? (
        <div
          style={{
            marginTop: 26,
            maxWidth: 1000,
            textAlign: 'center',
            fontFamily: theme.fontBody,
            fontSize: 24,
            color: theme.textMuted,
            opacity: interpolate(sinceStep, [8, 22], [0, 1], {
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
