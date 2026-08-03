import {interpolate, useCurrentFrame} from 'remotion';
import {FPS} from '../types';
import {ResolvedTheme, withAlpha} from '../theme/theme';
import {Stage, STAGE_W} from './Stage';
import {SceneHeader} from './SceneHeader';

// ForkScene is two processes over one memory, and the page that splits.
//
// The layout is a row of page cells with the two process boxes above it, one
// left and one right, both connected down to the same row. That symmetry is the
// claim: before anything is written, the two sides are looking at *the same
// cells*, not at two copies that happen to match.
//
// A write draws a new cell BELOW the row, on the writer's side, and re-points
// only that side's connector to it. The original cell stays exactly where it
// was, still connected to the other side. Drawing it that way — rather than
// splitting the row into two rows — is what keeps the frame able to say the two
// things the template exists for at once: this page diverged, and those five
// did not.
//
// The still-shared cells are never dimmed. It is tempting, because the copied
// page is what the narrator is talking about, but dimming them would say the
// sharing stopped mattering the moment one page broke — which is the opposite of
// the argument. They stay at full strength and the copy is what lights up.

const PAGE_W = 128;
const PAGE_H = 74;
const PAGE_GAP = 12;
const BOX_W = 230;
const LAYOUT_W = Math.min(STAGE_W, 1240);

type Page = {label?: string};
type Step = {
  startMs: number;
  endMs: number;
  show: 'shared' | 'write' | 'read';
  at?: number;
  by?: string;
  note?: string;
  role?: string;
  copied: Record<string, string>;
};

const roleColour = (theme: ResolvedTheme, role?: string): string => {
  switch (role) {
    case 'limit':
      return theme.accentLimit;
    case 'rival':
      return theme.accentRival;
    default:
      return theme.accentQuantity;
  }
};

export const ForkScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();

  const parent = String(props.parent ?? 'parent');
  const child = String(props.child ?? 'child');
  const origin = String(props.origin ?? '');
  const originNote = String(props.originNote ?? '');
  const pages = (Array.isArray(props.pages) ? props.pages : []) as Page[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];
  if (pages.length === 0 || steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const step = steps[idx];
  const sinceStep = ((nowMs - step.startMs) / 1000) * FPS;

  const copied = step.copied ?? {};
  const colour = roleColour(theme, step.role);
  const enter = interpolate(sinceStep, [0, 18], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // The copy appears after a beat's pause, so the viewer sees the write land on
  // the shared page before the page splits under it. Splitting instantly reads
  // as "there were always two", which is the thing being disproved.
  const split = interpolate(sinceStep, [8, 26], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const rowW = pages.length * (PAGE_W + PAGE_GAP) - PAGE_GAP;
  const note = step.show === 'write' || step.show === 'read' ? step.note : undefined;

  /** The horizontal centre of page i, in the layout's own coordinates. */
  const pageCentreX = (i: number) =>
    (LAYOUT_W - rowW) / 2 + i * (PAGE_W + PAGE_GAP) + PAGE_W / 2;

  /**
   * Where each side's connector lands.
   *
   * A side that has written is reading its own copy, so its leg moves off the
   * row's middle and onto that column. A side that has not is still reading the
   * shared row, and its leg stays where it was — which is the point: nothing
   * about the other process changed.
   */
  const legXFor = (side: string) => {
    const at = Object.keys(copied).find((k) => copied[k] === side);
    return at === undefined ? LAYOUT_W / 2 : pageCentreX(Number(at));
  };
  const parentLegX = legXFor('parent');
  const childLegX = legXFor('child');

  const box = (label: string, sub: string, lit: boolean) => (
    <div
      style={{
        width: BOX_W,
        padding: '16px 18px',
        borderRadius: 12,
        textAlign: 'center',
        background: lit ? withAlpha(colour, 0.1 * split) : theme.surface,
        border: `1px solid ${lit ? withAlpha(colour, 0.28 + 0.4 * split) : theme.surfaceBorder}`,
      }}
    >
      <div
        style={{
          fontFamily: theme.fontMono,
          fontSize: 20,
          letterSpacing: 2.6,
          textTransform: 'uppercase',
          color: lit ? colour : theme.text,
        }}
      >
        {label}
      </div>
      {sub ? (
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 14,
            color: theme.textMuted,
            marginTop: 5,
          }}
        >
          {sub}
        </div>
      ) : null}
    </div>
  );

  return (
    <Stage justify="center">
      <SceneHeader
        theme={theme}
        title={String(props.title ?? '')}
        emphasis={props.emphasis as string | undefined}
        emphasisRole={props.emphasisRole as string | undefined}
        size="compact"
        marginBottom={24}
      />

      {origin ? (
        <div
          style={{
            fontFamily: theme.fontMono,
            fontSize: 15,
            letterSpacing: 2.4,
            textTransform: 'uppercase',
            color: theme.textMuted,
            opacity: enter * 0.75,
            marginBottom: 18,
          }}
        >
          {origin}
          {originNote ? ` · ${originNote}` : ''}
        </div>
      ) : null}

      <div style={{width: LAYOUT_W, opacity: enter}}>
        {/* The two sides. */}
        <div style={{display: 'flex', justifyContent: 'space-between', marginBottom: 12}}>
          {box(parent, '', Object.values(copied).includes('parent'))}
          {box(child, '', Object.values(copied).includes('child'))}
        </div>

        {/* The connectors. Each side's leg runs to the column it is reading: the
            row's middle while everything is shared, and the copied column once
            that side has written.

            This is the part of the picture that carries the mechanism. Drawing
            both legs to the row and putting the copy underneath separately —
            which is what this did first — says "there is a copy" without ever
            saying the writer now reads it, which is the whole claim. */}
        <svg width={LAYOUT_W} height={46} style={{display: 'block'}}>
          {([
            [parentLegX, Object.values(copied).includes('parent')],
            [childLegX, Object.values(copied).includes('child')],
          ] as const).map(([x, lit], i) => (
            <path
              key={i}
              d={`M ${i === 0 ? BOX_W / 2 : LAYOUT_W - BOX_W / 2} 0 L ${
                i === 0 ? BOX_W / 2 : LAYOUT_W - BOX_W / 2
              } 22 L ${x} 22 L ${x} 46`}
              fill="none"
              stroke={lit ? withAlpha(colour, 0.35 + 0.5 * split) : theme.surfaceBorder}
              strokeWidth={2}
            />
          ))}
        </svg>

        {/* The shared memory. */}
        <div style={{display: 'flex', justifyContent: 'center'}}>
          <div style={{width: rowW, display: 'flex', gap: PAGE_GAP}}>
            {pages.map((pg, i) => {
              const by = copied[String(i)];
              const isCopied = by !== undefined;
              return (
                <div key={i} style={{width: PAGE_W}}>
                  <div
                    style={{
                      height: PAGE_H,
                      borderRadius: 10,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      textAlign: 'center',
                      paddingInline: 8,
                      // Still-shared pages keep full strength: dimming them
                      // would say the sharing stopped mattering, which is the
                      // opposite of the argument.
                      background: theme.surface,
                      border: `1px solid ${theme.surfaceBorder}`,
                      fontFamily: theme.fontMono,
                      fontSize: 15,
                      color: theme.textMuted,
                    }}
                  >
                    {pg.label ?? ''}
                  </div>
                  {/* The copy, below the original and on the writer's side. */}
                  {isCopied ? (
                    <div style={{opacity: split}}>
                      {/* The stub joining the original to its copy, so the leg
                          coming down from the writer continues through the row
                          rather than stopping at it. */}
                      <div
                        style={{
                          width: 2,
                          height: 14,
                          margin: '0 auto',
                          background: withAlpha(colour, 0.55),
                        }}
                      />
                      <div
                        style={{
                          fontFamily: theme.fontMono,
                          fontSize: 11,
                          letterSpacing: 1.6,
                          textTransform: 'uppercase',
                          color: colour,
                          textAlign: 'center',
                          marginBottom: 5,
                        }}
                      >
                        {by} copy
                      </div>
                      <div
                        style={{
                          height: PAGE_H,
                          borderRadius: 10,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          textAlign: 'center',
                          paddingInline: 8,
                          background: withAlpha(colour, 0.14),
                          border: `1px solid ${withAlpha(colour, 0.6)}`,
                          fontFamily: theme.fontMono,
                          fontSize: 15,
                          color: colour,
                          transform: `translateY(${(1 - split) * -14}px)`,
                        }}
                      >
                        {pg.label ?? ''}
                      </div>
                    </div>
                  ) : null}
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {note ? (
        <div
          style={{
            marginTop: 30,
            maxWidth: 1040,
            textAlign: 'center',
            fontFamily: theme.fontBody,
            fontSize: 25,
            color: theme.textMuted,
            opacity: interpolate(sinceStep, [12, 26], [0, 1], {
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
