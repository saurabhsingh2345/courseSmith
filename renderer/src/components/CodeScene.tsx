import {useEffect, useMemo, useState} from 'react';
import {
  continueRender,
  delayRender,
  interpolate,
  random,
  useCurrentFrame,
  useVideoConfig,
} from 'remotion';
import {codeToTokens, ThemedToken} from 'shiki';
import {ResolvedTheme} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage} from './Stage';

// CodeScene renders a self-typing editor window: chrome bar with traffic
// lights and a filename tab, line-number gutter, VS Code dark-plus tokens,
// then a verified-output console drawer once typing completes.

const TYPING_PORTION = 0.55;
// Sized so chrome + editor + output drawer + header still clear STAGE_H:
// 9*48+56 editor, ~74 chrome, ~130 drawer, ~95 header = 787 of 816.
const MAX_VISIBLE_LINES = 9;
const LINE_HEIGHT = 48;
const EDITOR_BG = '#16191f';
const CHROME_BG = '#0f1116';
const FALLBACK_TOKEN = '#d4d4d8';

type Char = {ch: string; color: string; line: number};

/** Flattens Shiki token lines into a per-character stream. */
const flatten = (lines: ThemedToken[][]): Char[] => {
  const chars: Char[] = [];
  lines.forEach((line, li) => {
    for (const token of line) {
      for (const ch of token.content) {
        chars.push({ch, color: token.color ?? FALLBACK_TOKEN, line: li});
      }
    }
    chars.push({ch: '\n', color: '', line: li});
  });
  chars.pop(); // no trailing newline
  return chars;
};

/**
 * Deterministic per-character reveal times with human jitter: each keystroke
 * takes base±40%, newlines pause a beat. Uses Remotion's seeded random so
 * every render is identical.
 */
const charRevealFrames = (chars: Char[], typingFrames: number): number[] => {
  const weights = chars.map((c, i) => {
    const jitter = 0.6 + 0.8 * random(`key-${i}`);
    return c.ch === '\n' ? 2.4 * jitter : jitter;
  });
  const total = weights.reduce((a, b) => a + b, 0) || 1;
  const times: number[] = [];
  let acc = 0;
  for (const w of weights) {
    acc += w;
    times.push((acc / total) * typingFrames);
  }
  return times;
};

/** Derives a plausible file name for the tab from the scene title. */
const fileNameFor = (title: string, language: string): string => {
  const ext = language === 'python' ? 'py' : language || 'txt';
  const slug = title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/^_+|_+$/g, '');
  return `${slug || 'main'}.${ext}`;
};

export const CodeScene: React.FC<{
  theme: ResolvedTheme;
  props: Record<string, unknown>;
}> = ({theme, props}) => {
  const frame = useCurrentFrame();
  const {durationInFrames} = useVideoConfig();

  const code = String(props.code ?? '');
  const language = String(props.language ?? 'python');
  const title = String(props.title ?? '');
  const output = String(props.output ?? '');
  const focusLines = Array.isArray(props.focusLines) ? (props.focusLines as number[]) : [];

  const [chars, setChars] = useState<Char[] | null>(null);
  const [handle] = useState(() => delayRender('shiki-highlight'));

  useEffect(() => {
    let cancelled = false;
    codeToTokens(code, {lang: language as 'python', theme: 'dark-plus'})
      .then((result) => {
        if (!cancelled) {
          setChars(flatten(result.tokens));
          continueRender(handle);
        }
      })
      .catch(() => {
        if (!cancelled) {
          // Highlighting failed (unknown language?) — fall back to plain text.
          setChars(
            code.split('').map((ch, i) => ({
              ch,
              color: FALLBACK_TOKEN,
              line: code.slice(0, i).split('\n').length - 1,
            })),
          );
          continueRender(handle);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [code, language, handle]);

  const typingFrames = Math.max(1, Math.floor(durationInFrames * TYPING_PORTION));
  const reveal = useMemo(
    () => (chars ? charRevealFrames(chars, typingFrames) : []),
    [chars, typingFrames],
  );

  if (!chars) {
    return null;
  }

  let visibleCount = 0;
  while (visibleCount < chars.length && reveal[visibleCount] <= frame) {
    visibleCount++;
  }
  const typingDone = visibleCount >= chars.length;
  const currentLine = visibleCount > 0 ? chars[Math.min(visibleCount, chars.length) - 1].line : 0;
  const totalLines = chars.length ? chars[chars.length - 1].line + 1 : 1;

  // Scroll so the line being typed stays in view.
  const scrollLines = Math.max(0, Math.min(currentLine - (MAX_VISIBLE_LINES - 4), totalLines - MAX_VISIBLE_LINES));

  // Rebuild visible lines from the character stream.
  const lines: Char[][] = [[]];
  for (let i = 0; i < visibleCount; i++) {
    if (chars[i].ch === '\n') {
      lines.push([]);
    } else {
      lines[lines.length - 1].push(chars[i]);
    }
  }

  const cursorOn = frame % 20 < 12;
  const dimmed = (line: number) =>
    focusLines.length > 0 && !focusLines.includes(line + 1) ? 0.32 : 1;

  const outputOpacity = output
    ? interpolate(frame, [typingFrames + 8, typingFrames + 22], [0, 1], {
        extrapolateLeft: 'clamp',
        extrapolateRight: 'clamp',
      })
    : 0;

  const shownLines = Math.min(totalLines, MAX_VISIBLE_LINES);

  return (
    <Stage>
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={30} />
      <div style={{width: 1520, maxWidth: '100%'}}>
        <div
          style={{
            borderRadius: 18,
            overflow: 'hidden',
            border: `1px solid ${theme.surfaceBorder}`,
            boxShadow: '0 44px 110px rgba(0, 0, 0, 0.55)',
            backgroundColor: EDITOR_BG,
          }}
        >
          {/* Chrome bar */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 10,
              padding: '16px 22px',
              backgroundColor: CHROME_BG,
              borderBottom: `1px solid ${theme.surfaceBorder}66`,
            }}
          >
            {['#ff5f57', '#febc2e', '#28c840'].map((c) => (
              <div key={c} style={{width: 16, height: 16, borderRadius: 8, backgroundColor: c}} />
            ))}
            <div
              style={{
                marginLeft: 18,
                padding: '8px 22px',
                borderRadius: 10,
                backgroundColor: EDITOR_BG,
                border: `1px solid ${theme.surfaceBorder}55`,
                color: theme.textMuted,
                fontFamily: theme.fontMono,
                fontSize: 22,
              }}
            >
              {fileNameFor(title, language)}
            </div>
          </div>
          {/* Editor body */}
          <div
            style={{
              display: 'flex',
              padding: '28px 0',
              overflow: 'hidden',
              maxHeight: shownLines * LINE_HEIGHT + 56,
            }}
          >
            {/* Gutter */}
            <div
              style={{
                width: 92,
                flexShrink: 0,
                textAlign: 'right',
                paddingRight: 26,
                fontFamily: theme.fontMono,
                fontSize: 26,
                lineHeight: `${LINE_HEIGHT}px`,
                color: theme.textMuted + '88',
                transform: `translateY(${-scrollLines * LINE_HEIGHT}px)`,
                userSelect: 'none',
              }}
            >
              {Array.from({length: totalLines}, (_, i) => (
                <div key={i} style={{height: LINE_HEIGHT, opacity: i < lines.length ? 1 : 0.25}}>
                  {i + 1}
                </div>
              ))}
            </div>
            <pre
              style={{
                margin: 0,
                flex: 1,
                fontFamily: theme.fontMono,
                fontSize: 32,
                lineHeight: `${LINE_HEIGHT}px`,
                transform: `translateY(${-scrollLines * LINE_HEIGHT}px)`,
              }}
            >
              {lines.map((line, li) => (
                <div key={li} style={{opacity: dimmed(li), minHeight: LINE_HEIGHT}}>
                  <span>
                    {line.map((c, ci) => (
                      <span key={ci} style={{color: c.color}}>
                        {c.ch}
                      </span>
                    ))}
                  </span>
                  {li === lines.length - 1 && !typingDone ? (
                    <span
                      style={{
                        display: 'inline-block',
                        width: 17,
                        height: 38,
                        marginLeft: 2,
                        verticalAlign: 'text-bottom',
                        backgroundColor: cursorOn ? theme.accent : 'transparent',
                      }}
                    />
                  ) : null}
                </div>
              ))}
            </pre>
          </div>
          {/* Output drawer */}
          {output ? (
            <div
              style={{
                opacity: outputOpacity,
                borderTop: `1px solid ${theme.surfaceBorder}66`,
                backgroundColor: CHROME_BG,
                padding: '20px 30px 26px',
              }}
            >
              <div style={{display: 'flex', alignItems: 'center', gap: 12, marginBottom: 12}}>
                <div style={{width: 11, height: 11, borderRadius: 6, backgroundColor: '#28c840'}} />
                <span
                  style={{
                    fontFamily: theme.fontBody,
                    fontSize: 19,
                    letterSpacing: 3,
                    fontWeight: 600,
                    textTransform: 'uppercase',
                    color: theme.textMuted,
                  }}
                >
                  Output · really executed
                </span>
              </div>
              <pre
                style={{
                  margin: 0,
                  whiteSpace: 'pre-wrap',
                  fontFamily: theme.fontMono,
                  fontSize: 29,
                  lineHeight: 1.45,
                  color: '#e6edf3',
                }}
              >
                {output.trimEnd()}
              </pre>
            </div>
          ) : null}
        </div>
      </div>
    </Stage>
  );
};
