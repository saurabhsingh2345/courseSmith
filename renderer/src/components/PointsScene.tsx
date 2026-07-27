import {spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {MotionTokens, resolveMotion} from '../theme/motion';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_H} from './Stage';
import {iconFor} from './icons';

// PointsScene renders the storyboard's visual beats: keyword phrases that
// land on the exact narration word they belong to. Icon names come from the
// pipeline's closed vocabulary (storyboard.go).
//
// Items are laid out in full from frame 0 and merely *ghosted* until their
// cue, rather than popping in from nothing. Revealing one at a time left the
// slide showing a heading and a single line for seconds at a stretch — the
// composition never existed until the section was nearly over. Ghosting keeps
// the slide whole while still marking progress through it.
//
// Item metrics shrink with the item count so a long list still fits the stage
// box instead of growing down into the caption band.

// The icon vocabulary is shared with every scene that takes an icon name
// from the pipeline (see components/icons.ts).

type PointItem = {text: string; icon: string; atMs: number};

/**
 * Row metrics that keep `n` items inside the stage box under the header.
 * The header eats ~200px of STAGE_H, so the list gets the rest.
 */
const rowMetrics = (n: number) => {
  if (n <= 3) return {box: 84, icon: 42, font: 52, gap: 34, radius: 22};
  if (n === 4) return {box: 76, icon: 38, font: 45, gap: 26, radius: 20};
  if (n === 5) return {box: 68, icon: 34, font: 40, gap: 22, radius: 18};
  return {box: 60, icon: 30, font: 35, gap: 16, radius: 16};
};

/** Opacity of an item that has not reached its cue yet. */
const GHOST = 0.26;

export const PointsScene: React.FC<{
  theme: ResolvedTheme;
  motion?: MotionTokens;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, motion, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  resolveMotion(motion); // motion tokens reserved for per-archetype tuning

  const title = String(props.title ?? '');
  // Template variant: "rows" (default, vertical list) or "grid" (2-column
  // cards — the playful archetype's pick, also settable per scene in
  // video-plan.yaml).
  const template = String(props.template ?? 'rows');
  const items = Array.isArray(props.items) ? (props.items as PointItem[]) : [];
  const nowMs = sceneStartMs + (frame / FPS) * 1000;

  const revealed = items.filter((it) => nowMs >= it.atMs).length;
  const m = rowMetrics(items.length);

  const header = (
    <SceneHeader
      theme={theme}
      title={title}
      kicker={String(props.kicker ?? theme.courseName ?? '')}
      marginBottom={items.length > 4 ? 44 : 56}
    />
  );

  /** Reveal state for item i: ghosted before its cue, settling after it. */
  const reveal = (it: PointItem, i: number) => {
    const startFrame = Math.round(((it.atMs - sceneStartMs) / 1000) * FPS);
    const s =
      nowMs >= it.atMs
        ? spring({
            frame: frame - startFrame,
            fps,
            config: {damping: 26, stiffness: 210, mass: 0.7},
          })
        : 0;
    return {s, opacity: GHOST + (1 - GHOST) * s, isLatest: i === revealed - 1};
  };

  if (template === 'grid') {
    const cols = items.length <= 2 ? items.length || 1 : 2;
    return (
      <Stage>
        {header}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: `repeat(${cols}, minmax(380px, 620px))`,
            gap: items.length > 4 ? 22 : 30,
            justifyContent: 'center',
            maxHeight: STAGE_H,
          }}
        >
          {items.map((it, i) => {
            const {s, opacity, isLatest} = reveal(it, i);
            const Icon = iconFor(it.icon);
            return (
              <div
                key={i}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 22,
                  padding: `${Math.round(m.box * 0.36)}px 34px`,
                  borderRadius: m.radius,
                  backgroundColor: theme.surface,
                  border: `1.5px solid ${isLatest ? theme.accent : theme.surfaceBorder}`,
                  boxShadow: isLatest ? `0 0 54px ${theme.accent}2b` : '0 18px 44px rgba(0,0,0,0.3)',
                  opacity,
                  transform: `translateY(${(1 - s) * 10}px)`,
                }}
              >
                <Icon
                  size={m.icon}
                  color={isLatest ? theme.accent : theme.textMuted}
                  strokeWidth={2}
                  style={{flexShrink: 0}}
                />
                <div
                  style={{
                    fontFamily: theme.fontDisplay,
                    fontSize: m.font * 0.86,
                    fontWeight: 600,
                    letterSpacing: -0.4,
                    color: isLatest ? theme.text : theme.textMuted,
                  }}
                >
                  {it.text}
                </div>
              </div>
            );
          })}
        </div>
      </Stage>
    );
  }

  return (
    <Stage>
      {header}
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          gap: m.gap,
          width: 'fit-content',
          maxWidth: '100%',
          minWidth: 700,
        }}
      >
        {items.map((it, i) => {
          const {s, opacity, isLatest} = reveal(it, i);
          const Icon = iconFor(it.icon);
          return (
            <div
              key={i}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 26,
                opacity,
                // A short slide rather than the old 40px drop: the row already
                // occupies its slot as a ghost, so it settles instead of
                // arriving.
                transform: `translateX(${(1 - s) * -14}px)`,
              }}
            >
              <div
                style={{
                  width: m.box,
                  height: m.box,
                  borderRadius: m.radius,
                  flexShrink: 0,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  backgroundColor: theme.surface,
                  border: `1.5px solid ${isLatest ? theme.accent : theme.surfaceBorder}`,
                  boxShadow: isLatest ? `0 0 44px ${theme.accent}33` : 'none',
                }}
              >
                <Icon
                  size={m.icon}
                  color={isLatest ? theme.accent : theme.textMuted}
                  strokeWidth={2.2}
                />
              </div>
              <div
                style={{
                  fontFamily: theme.fontDisplay,
                  fontSize: m.font,
                  fontWeight: 600,
                  letterSpacing: -0.6,
                  lineHeight: 1.2,
                  color: isLatest ? theme.text : theme.textMuted,
                }}
              >
                {it.text}
              </div>
            </div>
          );
        })}
      </div>
    </Stage>
  );
};
