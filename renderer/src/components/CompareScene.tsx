import {useEffect, useState} from 'react';
import {continueRender, delayRender, interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {codeToTokens, ThemedToken} from 'shiki';
import {Check, Minus} from 'lucide-react';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage, STAGE_H, STAGE_W} from './Stage';
import {FIGURE_BOX, figureFor, type FigurePalette} from './artwork';

// CompareScene is two subjects in one frame, introduced in turn and judged.
//
// The only scene in the catalog that shows more than one thing at a time, and
// the reason it exists: cutting between two shots does not compare them,
// because the viewer has to hold the first in memory to weigh it against the
// second. Both columns are in the frame from the first frame; the beats only
// change which one is *lit*.
//
// Which means the layout must not move. A column that arrives by sliding in
// would shift the other one, and the two have to be measurable against each
// other — same width, same baseline, same type size — or the comparison is
// being made by the layout rather than by the content.

const HEADER_H = 104;
const GUTTER = 56;
const COL_W = Math.floor((Math.min(STAGE_W, 1560) - GUTTER) / 2);
const BODY_H = STAGE_H - HEADER_H - 150;

type Side = {label: string; note?: string; code?: string; figure?: string};
type Step = {startMs: number; endMs: number; show: string};

/** Shiki tokens for a column's code, or null while loading / on failure. */
const useTokens = (code: string | undefined, language: string) => {
  const [tokens, setTokens] = useState<ThemedToken[][] | null>(null);
  const [handle] = useState(() => delayRender('compare-highlight'));
  useEffect(() => {
    if (!code) {
      continueRender(handle);
      return;
    }
    let cancelled = false;
    codeToTokens(code, {lang: language as 'python', theme: 'dark-plus'})
      .then((r) => {
        if (!cancelled) {
          setTokens(r.tokens);
          continueRender(handle);
        }
      })
      .catch(() => {
        if (!cancelled) {
          // Unknown language: fall back to unhighlighted lines rather than an
          // empty column, which would read as "this side has no code".
          setTokens(code.split('\n').map((l) => [{content: l, color: undefined} as ThemedToken]));
          continueRender(handle);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [code, language, handle]);
  return tokens;
};

export const CompareScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const title = String(props.title ?? '');
  const language = String(props.language ?? 'python');
  const left = (props.left ?? {}) as Side;
  const right = (props.right ?? {}) as Side;
  const winner = String(props.winner ?? 'tie');
  const verdict = String(props.verdict ?? '');
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];

  // Both hooks run unconditionally and before any early return — the columns
  // may or may not carry code, and a conditional hook is a hook-count change.
  const leftTokens = useTokens(left.code, language);
  const rightTokens = useTokens(right.code, language);

  if (steps.length === 0) return null;

  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  let idx = steps.findIndex((s) => nowMs >= s.startMs && nowMs < s.endMs);
  if (idx < 0) idx = nowMs < steps[0].startMs ? 0 : steps.length - 1;
  const show = steps[idx].show;

  // Once introduced, always present. Derived by scanning rather than tracked,
  // because a frame renders on its own.
  const leftAt = steps.find((s) => s.show === 'left');
  const rightAt = steps.find((s) => s.show === 'right');
  const verdictAt = steps.find((s) => s.show === 'verdict');
  const leftIn = leftAt && nowMs >= leftAt.startMs;
  const rightIn = rightAt && nowMs >= rightAt.startMs;
  const decided = verdictAt ? nowMs >= verdictAt.startMs : false;

  const enter = (from: Step | undefined) =>
    from
      ? spring({
          frame: ((nowMs - from.startMs) / 1000) * FPS,
          fps,
          config: {damping: 200, mass: 0.6},
          durationInFrames: 18,
        })
      : 0;

  const verdictP = verdictAt
    ? interpolate(((nowMs - verdictAt.startMs) / 1000) * FPS, [0, 16], [0, 1], {
        extrapolateLeft: 'clamp',
        extrapolateRight: 'clamp',
      })
    : 0;

  const palette: FigurePalette = {
    accent: theme.accent,
    primary: theme.primary,
    ink: theme.ink,
    soft: theme.mass,
    line: theme.line,
  };

  // ONE code size and ONE panel height, shared by both columns.
  //
  // Sizing each column from its own content is the bug this template is least
  // able to afford: the left side would set 8 lines at 20px and the right side
  // one line at 28px, and the frame would be making the argument instead of the
  // code. Two things being weighed against each other have to be measured in
  // the same units, so the widest line and the tallest column in the *pair*
  // decide the type size and the box for both.
  const codeLines = [left.code, right.code]
    .filter(Boolean)
    .flatMap((c) => (c as string).split('\n'));
  const widest = Math.max(1, ...codeLines.map((l) => l.length));
  const tallest = Math.max(1, ...[left.code, right.code].map((c) => (c ? c.split('\n').length : 1)));
  const codeSize = Math.max(18, Math.min(30, ((COL_W - 56) / widest) * 1.72));
  const lineH = codeSize * 1.55;
  const hasFigure = !left.code || !right.code;
  // Height from the taller column's content, floored so a one-line comparison
  // is still a panel rather than a strip, and capped at the stage.
  const panelH = Math.min(
    BODY_H,
    Math.max(hasFigure ? 360 : 200, Math.round(tallest * lineH + 48)),
  );

  /** How lit a column is: dimmed while the other one is the subject. */
  const litness = (isLeft: boolean): number => {
    const present = isLeft ? leftIn : rightIn;
    if (!present) return 0;
    if (decided) {
      if (winner === 'tie') return 1;
      const won = (winner === 'left') === isLeft;
      return won ? 1 : 1 - verdictP * 0.55;
    }
    if (show === 'both') return 1;
    if (show === 'left') return isLeft ? 1 : 0.45;
    if (show === 'right') return isLeft ? 0.45 : 1;
    return 1;
  };

  const column = (side: Side, tokens: ThemedToken[][] | null, isLeft: boolean) => {
    const present = isLeft ? leftIn : rightIn;
    const arrive = enter(isLeft ? leftAt : rightAt);
    const won = decided && (winner === 'tie' || (winner === 'left') === isLeft);

    return (
      <div
        style={{
          width: COL_W,
          display: 'flex',
          flexDirection: 'column',
          gap: 16,
          opacity: present ? litness(isLeft) : 0,
          // Only a small settle. A column that slides in would move the other
          // one, and two things being compared have to hold still.
          transform: `translateY(${(1 - arrive) * 14}px)`,
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 12,
            fontFamily: theme.fontDisplay,
            fontSize: 34,
            fontWeight: 700,
            color: won ? theme.accentText : theme.text,
          }}
        >
          {decided && won && (
            <span style={{display: 'flex', opacity: verdictP}}>
              {winner === 'tie' ? (
                <Minus size={28} color={theme.accentText} strokeWidth={3} />
              ) : (
                <Check size={28} color={theme.accentText} strokeWidth={3} />
              )}
            </span>
          )}
          {side.label}
        </div>

        <div
          style={{
            height: panelH,
            borderRadius: 18,
            border: `2px solid ${decided && won ? theme.accent : theme.surfaceBorder}`,
            backgroundColor: theme.surface,
            padding: side.code ? '22px 26px' : 0,
            display: 'flex',
            alignItems: side.code ? 'flex-start' : 'center',
            justifyContent: 'center',
            overflow: 'hidden',
          }}
        >
          {side.code ? (
            <div
              style={{
                width: '100%',
                fontFamily: theme.fontMono,
                fontSize: codeSize,
                lineHeight: `${lineH}px`,
                whiteSpace: 'pre',
              }}
            >
              {(tokens ?? []).map((line, li) => (
                <div key={li}>
                  {line.length === 0 ? ' ' : null}
                  {line.map((t, ti) => (
                    <span key={ti} style={{color: t.color ?? theme.text}}>
                      {t.content}
                    </span>
                  ))}
                </div>
              ))}
            </div>
          ) : (
            <svg
              width={Math.min(panelH * 0.74, 300)}
              height={Math.min(panelH * 0.74, 300)}
              viewBox={`0 0 ${FIGURE_BOX} ${FIGURE_BOX}`}
              style={{overflow: 'visible'}}
            >
              {(() => {
                const F = figureFor(side.figure);
                return <F build={arrive} t={frame / FPS} palette={palette} />;
              })()}
            </svg>
          )}
        </div>

        {side.note && (
          <div
            style={{
              fontFamily: theme.fontMono,
              fontSize: 24,
              color: decided && won ? theme.accentText : theme.textMuted,
              letterSpacing: 0.4,
            }}
          >
            {side.note}
          </div>
        )}
      </div>
    );
  };

  // Centred, which is only safe because the verdict line below is always
  // rendered — at zero opacity before it lands — so its height is reserved and
  // nothing shifts when it arrives. Two columns being weighed against each
  // other must not move at the moment the answer appears.
  return (
    <Stage justify="center">
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={22} />
      <div style={{display: 'flex', gap: GUTTER, alignItems: 'stretch', position: 'relative'}}>
        {column(left, leftTokens, true)}
        {/* The divider. It draws down from the top as the second column
            arrives, which is the moment the frame becomes a comparison rather
            than one thing with space beside it. */}
        <div
          style={{
            position: 'absolute',
            left: COL_W + GUTTER / 2 - 1,
            top: 0,
            width: 2,
            height: `${(rightIn ? enter(rightAt) : 0) * 100}%`,
            backgroundColor: theme.surfaceBorder,
          }}
        />
        {column(right, rightTokens, false)}
      </div>

      {/* The verdict, across both columns. It is the only thing in the frame
          that belongs to neither side. */}
      <div
        style={{
          marginTop: 26,
          maxWidth: COL_W * 2 + GUTTER,
          fontFamily: theme.fontBody,
          fontSize: 30,
          lineHeight: 1.35,
          color: theme.text,
          opacity: verdictP,
          transform: `translateY(${(1 - verdictP) * 10}px)`,
        }}
      >
        {verdict}
      </div>
    </Stage>
  );
};
