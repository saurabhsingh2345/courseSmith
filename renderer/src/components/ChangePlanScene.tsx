import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, seat, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// ChangePlanScene is a change plan with the scroll taken out of it.
//
// The rail of files stands still down the left and one file's bullets are open on
// the right. Nothing here scrolls and nothing reflows — which is the entire
// difference from the artifact it draws, where forty lines of monospace go past and
// the viewer never knows how much is left.
//
// Two things earn their pixels.
//
// THE VERDICT DOT, and its colour. Green adds, amber edits, red deletes, and a
// hollow ring means untouched. It is the only colour on the frame, so a viewer can
// read the shape of the whole change — mostly edits with one new file — before a
// word is spoken, which is the thing a wall of text cannot do at any size.
//
// THE HOLLOW RING for `unchanged`, rather than a grey dot. A grey dot in a column
// of coloured ones reads as a row that failed to load; a ring reads as deliberate,
// which is exactly what an untouched file is. That row is the most reassuring thing
// a plan contains and it has to look chosen.
//
// The open panel is on the right and its height is fixed to the rail's, so a file
// with one bullet and a file with four leave the composition identical. A panel
// that grew would move the rail, and the rail standing still is the whole idea.

const BLOCK_W = Math.min(STAGE_W, 1640);
const RAIL_W = 660;
const GAP = 56;
/** Between rail rows. */
const RAIL_GAP = 10;

type File = {
  path: string;
  summary: string;
  verdict?: string;
  edits?: string[];
};

type Step = {
  startMs: number;
  endMs: number;
  show: 'rail' | 'file' | 'all';
  at?: number;
  done?: number;
};

const verdictColour = (theme: ResolvedTheme, verdict?: string): string => {
  switch (verdict) {
    case 'add':
      return '#2ea043';
    case 'delete':
      return theme.accentLimit;
    case 'unchanged':
      return theme.textMuted;
    default:
      return theme.accentQuantity;
  }
};

/** The dot, or the ring that means "deliberately left alone". */
const VerdictMark: React.FC<{theme: ResolvedTheme; verdict?: string; lit: boolean}> = ({
  theme,
  verdict,
  lit,
}) => {
  const c = verdictColour(theme, verdict);
  const hollow = verdict === 'unchanged';
  return (
    <div
      style={{
        width: 18,
        height: 18,
        flex: 'none',
        borderRadius: 999,
        background: hollow ? 'transparent' : lit ? c : withAlpha(c, 0.45),
        border: hollow ? `3px solid ${lit ? c : withAlpha(c, 0.5)}` : 'none',
      }}
    />
  );
};

export const ChangePlanScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const files = (Array.isArray(props.files) ? props.files : []) as File[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  const closer = String(props.closer ?? '');
  if (files.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const since = ((nowMs - step.startMs) / 1000) * FPS;

  const open = step.show === 'file' ? (step.at ?? -1) : -1;
  const land = interpolate(since, [2, 16], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  // The rail sets the height and the panel matches it, rather than both being
  // measured against a constant.
  //
  // The constant was the bug: a fixed 560 body with three 96-pixel rows left the
  // rail filling half its column and the panel filling all of it, so the two sides
  // of a two-column composition were visibly different objects. Rows get a
  // comfortable fixed height, the rail is however tall that makes it, and the panel
  // is told to match — which also means a plan with six files gets a taller frame
  // than one with two, in the direction the content actually goes.
  const rowH = files.length > 4 ? 88 : 104;
  const bodyH = files.length * rowH + (files.length - 1) * RAIL_GAP;
  const current = open >= 0 ? files[open] : undefined;

  return (
    <Stage justify="center">
      <div style={{width: BLOCK_W}}>
        <SceneHeader
          theme={theme}
          title={String(props.title ?? '')}
          emphasis={props.emphasis as string | undefined}
          emphasisRole={props.emphasisRole as string | undefined}
          size="compact"
          marginBottom={40}
        />

        <div style={{display: 'flex', gap: GAP, alignItems: 'flex-start'}}>
          {/* The rail. */}
          <div style={{width: RAIL_W, flex: 'none', display: 'flex', flexDirection: 'column', gap: RAIL_GAP}}>
            {files.map((f, i) => {
              const lit = open === i;
              // A row the clip has already walked stays legible rather than going
              // back to sleep: the rail is a record of progress, not a spotlight.
              const walked = (step.done ?? 0) > i;
              const enter = interpolate(frame, [3 + i * 5, 20 + i * 5], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              });
              return (
                <div
                  key={i}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 18,
                    height: rowH,
                    padding: '0 24px',
                    borderRadius: 16,
                    background: lit ? theme.surface : 'transparent',
                    border: `1px solid ${lit ? theme.surfaceBorder : 'transparent'}`,
                    boxShadow: lit ? seat(theme, 'resting') : 'none',
                    opacity: enter,
                    transform: `translateX(${(1 - enter) * -12}px)`,
                  }}
                >
                  <VerdictMark theme={theme} verdict={f.verdict} lit={lit || walked} />
                  <div style={{minWidth: 0, flex: 1}}>
                    <div
                      style={{
                        fontFamily: theme.fontMono,
                        fontSize: rowH > 80 ? 27 : 24,
                        fontWeight: 500,
                        color: lit || walked ? theme.text : theme.textMuted,
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                      }}
                    >
                      {f.path}
                    </div>
                    <div
                      style={{
                        marginTop: 3,
                        fontFamily: theme.fontBody,
                        fontSize: rowH > 80 ? 21 : 19,
                        color: theme.textMuted,
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                      }}
                    >
                      {f.summary}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          {/* The open file. Fixed height, so a one-bullet file and a four-bullet
              file leave the rail in exactly the same place. */}
          <div
            style={{
              flex: 1,
              minWidth: 0,
              height: bodyH,
              borderRadius: 22,
              padding: '38px 40px',
              background: theme.surface,
              border: `1px solid ${theme.surfaceBorder}`,
              boxShadow: seat(theme, 'resting'),
              display: 'flex',
              flexDirection: 'column',
            }}
          >
            {current ? (
              <>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 14,
                    opacity: land,
                  }}
                >
                  <VerdictMark theme={theme} verdict={current.verdict} lit />
                  <div
                    style={{
                      fontFamily: theme.fontBody,
                      fontSize: 19,
                      fontWeight: 600,
                      letterSpacing: 2.6,
                      textTransform: 'uppercase',
                      color: verdictColour(theme, current.verdict),
                    }}
                  >
                    {current.verdict ?? 'edit'}
                  </div>
                </div>
                <div
                  style={{
                    marginTop: 16,
                    fontFamily: theme.fontDisplay,
                    fontSize: 40,
                    fontWeight: 700,
                    letterSpacing: -0.8,
                    lineHeight: 1.2,
                    color: theme.text,
                    opacity: land,
                  }}
                >
                  {current.summary}
                </div>
                <div style={{marginTop: 28, display: 'flex', flexDirection: 'column', gap: 18}}>
                  {(current.edits ?? []).map((e, i) => {
                    // Each bullet arrives a beat behind the one above it, which is
                    // the only stagger on the frame and the only thing that says
                    // this panel was written rather than switched on.
                    const on = interpolate(since, [6 + i * 5, 22 + i * 5], [0, 1], {
                      extrapolateLeft: 'clamp',
                      extrapolateRight: 'clamp',
                    });
                    return (
                      <div
                        key={i}
                        style={{
                          display: 'flex',
                          gap: 16,
                          opacity: on,
                          transform: `translateY(${(1 - on) * 8}px)`,
                        }}
                      >
                        <div
                          style={{
                            fontFamily: theme.fontMono,
                            fontSize: 28,
                            color: withAlpha(theme.text, 0.35),
                            flex: 'none',
                          }}
                        >
                          −
                        </div>
                        <div
                          style={{
                            fontFamily: theme.fontBody,
                            fontSize: 28,
                            lineHeight: 1.35,
                            color: theme.text,
                          }}
                        >
                          {e}
                        </div>
                      </div>
                    );
                  })}
                  {(current.edits ?? []).length === 0 ? (
                    <div
                      style={{
                        fontFamily: theme.fontBody,
                        fontSize: 28,
                        lineHeight: 1.4,
                        color: theme.textMuted,
                        opacity: land,
                      }}
                    >
                      Nothing to do here — and that is worth saying out loud.
                    </div>
                  ) : null}
                </div>
              </>
            ) : (
              // The rail beat and the closing beat. Deliberately not empty: an
              // empty panel beside a full rail reads as a frame that failed, so it
              // holds the count first and the conclusion last.
              <div
                style={{
                  flex: 1,
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'center',
                }}
              >
                <div
                  style={{
                    fontFamily: theme.fontDisplay,
                    fontSize: 120,
                    fontWeight: 800,
                    letterSpacing: -4,
                    lineHeight: 1,
                    color: theme.accentText,
                  }}
                >
                  {files.length}
                </div>
                <div
                  style={{
                    marginTop: 10,
                    fontFamily: theme.fontBody,
                    fontSize: 30,
                    color: theme.textMuted,
                  }}
                >
                  file{files.length === 1 ? '' : 's'} touched
                </div>
                {step.show === 'all' && closer ? (
                  <div
                    style={{
                      marginTop: 34,
                      fontFamily: theme.fontDisplay,
                      fontSize: 36,
                      fontWeight: 700,
                      letterSpacing: -0.6,
                      lineHeight: 1.3,
                      color: theme.text,
                      opacity: land,
                    }}
                  >
                    {closer}
                  </div>
                ) : null}
              </div>
            )}
          </div>
        </div>
      </div>
    </Stage>
  );
};
