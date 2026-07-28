import {interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_W} from './Stage';
import {iconFor} from './icons';

// BreakdownScene is a path whose current stage is open.
//
// The whole path is on screen from the first frame — collapsed rows for the
// stages not being talked about, one tall panel for the one that is. That
// accordion is the template: a viewer can see how far in they are and how much
// is left, while still getting the detail of where they are standing, and no
// other shape in the catalog carries both at once.
//
// The row heights are sprung rather than switched, so a stage opening reads as
// the list making room for it. Cutting between two static layouts would show the
// same information and lose the thing that makes it legible, which is that you
// can watch where the space came from.
//
// The closing beat is not simply "everything collapsed". Every row grows a little
// and shows its items as chips, so the last frame is the whole path *with its
// contents* — which is the frame somebody screenshots, and the reason to have
// walked it.

const COL_W = Math.min(STAGE_W, 1620);
const ROW_GAP = 12;
const H_COLLAPSED = 74;
// Sized to what an open phase actually contains — head, detail line, one row of
// item cards — rather than to a round number. The first guess was 292 and left a
// band of empty panel under the items that read as a row still loading.
const H_OPEN = 240;
// On the summary the head grows to fill the whole row, so the detail line below
// it is out of frame rather than sliced in half by the row's overflow.
const H_SUMMARY = 112;

type Item = {name: string; note?: string; icon?: string};
type Phase = {title: string; detail: string; items: Item[]};
type Step = {startMs: number; endMs: number; show: string; at?: number; item?: number};

export const BreakdownScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const title = String(props.title ?? '');
  const phases = (Array.isArray(props.phases) ? props.phases : []) as Phase[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];

  if (phases.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const whole = step.show === 'whole';
  const open = whole ? -1 : (step.at ?? 0);
  const spotlit = step.show === 'item' ? (step.item ?? -1) : -1;
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  // How far the accordion has moved into this beat's shape. Sprung, so the row
  // that opens looks like it pushed the others apart.
  const move = spring({frame: sinceStep, fps, config: {damping: 200, mass: 0.8}, durationInFrames: 22});
  const prev = steps[Math.max(0, idx - 1)];
  const prevOpen = prev.show === 'whole' ? -1 : (prev.at ?? 0);
  const prevWhole = prev.show === 'whole';
  const heightOf = (i: number) => {
    const to = whole ? H_SUMMARY : i === open ? H_OPEN : H_COLLAPSED;
    const from = prevWhole ? H_SUMMARY : i === prevOpen ? H_OPEN : H_COLLAPSED;
    return from + (to - from) * move;
  };

  return (
    <Stage justify="center">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={16} />

      <div style={{width: COL_W, display: 'flex', flexDirection: 'column', gap: ROW_GAP}}>
        {phases.map((phase, i) => {
          const isOpen = !whole && i === open;
          const reached = whole || i <= open;
          const h = heightOf(i);
          return (
            <div
              key={i}
              style={{
                height: h,
                flexShrink: 0,
                overflow: 'hidden',
                borderRadius: 20,
                backgroundColor: isOpen ? withAlpha(theme.surface, 0.9) : withAlpha(theme.surface, 0.5),
                border: `${isOpen ? 2 : 1}px solid ${isOpen ? theme.accent : theme.surfaceBorder}`,
                boxShadow: isOpen ? `0 20px 52px ${withAlpha(theme.ink, 0.5)}` : 'none',
                opacity: reached ? 1 : 0.42,
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              {/* The row head, present at every height. */}
              <div
                style={{
                  height: whole ? H_SUMMARY : H_COLLAPSED,
                  flexShrink: 0,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 20,
                  padding: '0 24px',
                }}
              >
                <span
                  style={{
                    width: 42,
                    height: 42,
                    flexShrink: 0,
                    borderRadius: 12,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    backgroundColor: reached ? withAlpha(theme.accent, 0.2) : withAlpha(theme.line, 0.16),
                    fontFamily: theme.fontMono,
                    fontSize: 20,
                    fontWeight: 700,
                    color: reached ? theme.accentText : theme.textMuted,
                  }}
                >
                  {i + 1}
                </span>
                <span
                  style={{
                    fontFamily: theme.fontDisplay,
                    fontSize: isOpen ? 36 : 30,
                    fontWeight: 700,
                    letterSpacing: -0.5,
                    color: theme.text,
                    whiteSpace: 'nowrap',
                  }}
                >
                  {phase.title}
                </span>
                {/* On the summary the chips take this space; otherwise a count
                    badge says how much is inside without opening it. */}
                {whole ? (
                  <div style={{display: 'flex', gap: 10, flexWrap: 'nowrap', overflow: 'hidden'}}>
                    {phase.items.map((it) => (
                      <span
                        key={it.name}
                        style={{
                          padding: '7px 15px',
                          borderRadius: 999,
                          border: `1px solid ${theme.surfaceBorder}`,
                          backgroundColor: withAlpha(theme.bgBottom, 0.5),
                          fontFamily: theme.fontBody,
                          fontSize: 21,
                          color: theme.textMuted,
                          whiteSpace: 'nowrap',
                        }}
                      >
                        {it.name}
                      </span>
                    ))}
                  </div>
                ) : null}
                <div style={{flex: 1}} />
                {!whole ? (
                  <span
                    style={{
                      fontFamily: theme.fontMono,
                      fontSize: 16,
                      letterSpacing: 1.4,
                      textTransform: 'uppercase',
                      color: theme.textMuted,
                      flexShrink: 0,
                    }}
                  >
                    {`${phase.items.length} items`}
                  </span>
                ) : null}
              </div>

              {/* The detail, revealed by the row's own height rather than by an
                  opacity fade — the panel is not appearing over the list, the
                  list is making room for it. */}
              <div style={{padding: '0 24px 22px', minWidth: 0, opacity: whole ? 0 : 1}}>
                <div
                  style={{
                    fontFamily: theme.fontBody,
                    fontSize: 26,
                    lineHeight: 1.3,
                    color: theme.textMuted,
                    marginBottom: 18,
                  }}
                >
                  {phase.detail}
                </div>
                <div style={{display: 'flex', gap: 14}}>
                  {phase.items.map((it, j) => {
                    const lit = isOpen && j === spotlit;
                    const Icon = iconFor(it.icon);
                    return (
                      <div
                        key={j}
                        style={{
                          flex: 1,
                          minWidth: 0,
                          display: 'flex',
                          alignItems: 'center',
                          gap: 13,
                          padding: '14px 16px',
                          borderRadius: 15,
                          backgroundColor: lit ? withAlpha(theme.accent, 0.12) : withAlpha(theme.bgBottom, 0.55),
                          border: `${lit ? 2 : 1}px solid ${lit ? theme.accent : theme.surfaceBorder}`,
                          transform: `translateY(${lit ? -move * 6 : 0}px)`,
                          opacity: isOpen && spotlit >= 0 && !lit ? 0.5 : 1,
                        }}
                      >
                        <div
                          style={{
                            width: 42,
                            height: 42,
                            flexShrink: 0,
                            borderRadius: 12,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            background: `linear-gradient(145deg, ${withAlpha(theme.accent, 0.26)}, ${withAlpha(
                              theme.primary,
                              0.16,
                            )})`,
                            border: `1px solid ${withAlpha(theme.accent, 0.32)}`,
                          }}
                        >
                          <Icon size={22} color={theme.accent} strokeWidth={2.1} />
                        </div>
                        <div style={{minWidth: 0}}>
                          <div
                            style={{
                              fontFamily: theme.fontDisplay,
                              fontSize: 23,
                              fontWeight: 650,
                              color: theme.text,
                              whiteSpace: 'nowrap',
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                            }}
                          >
                            {it.name}
                          </div>
                          {it.note ? (
                            <div
                              style={{
                                marginTop: 3,
                                fontFamily: theme.fontBody,
                                fontSize: 18,
                                lineHeight: 1.22,
                                color: theme.textMuted,
                              }}
                            >
                              {it.note}
                            </div>
                          ) : null}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* A progress read on the path, so "how much is left" is answerable at a
          glance rather than by counting rows. */}
      <div
        style={{
          marginTop: 20,
          width: COL_W,
          display: 'flex',
          alignItems: 'center',
          gap: 18,
          opacity: interpolate(sinceStep, [0, 10], [0.6, 1], {
            extrapolateLeft: 'clamp',
            extrapolateRight: 'clamp',
          }),
        }}
      >
        <span
          style={{
            fontFamily: theme.fontMono,
            fontSize: 17,
            letterSpacing: 2,
            textTransform: 'uppercase',
            color: theme.textMuted,
          }}
        >
          {whole ? `${phases.length} phases` : `Phase ${open + 1} of ${phases.length}`}
        </span>
        <div style={{flex: 1, height: 6, borderRadius: 3, backgroundColor: withAlpha(theme.line, 0.22)}}>
          <div
            style={{
              width: `${((whole ? phases.length : open + 1) / phases.length) * 100}%`,
              height: 6,
              borderRadius: 3,
              backgroundColor: theme.accent,
            }}
          />
        </div>
      </div>
    </Stage>
  );
};
