import {useEffect, useMemo, useState} from 'react';
import {continueRender, delayRender, interpolate, random, useCurrentFrame} from 'remotion';
import {codeToTokens, ThemedToken} from 'shiki';
import {
  Blocks,
  Bug,
  FileCode2,
  Files,
  FolderOpen,
  GitBranch,
  Search,
} from 'lucide-react';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage} from './Stage';

// VSCodeScene is the "VS Code walkthrough": a synthesized editor — activity
// bar, file tree, tabs, minimap, status bar — whose code evolves through
// timed steps. Step 1 types itself in; later steps swap the buffer and flash
// the changed lines. Everything is frame-driven (research: the only
// deterministic approach; real editors and screen capture are wall-clock).
// The chrome is VS Code-like but carries no Microsoft branding.

const WINDOW_W = 1660;
// Bounded so title bar + editor + terminal drawer + status bar + the scene
// header still clear STAGE_H: 9*44+60 editor, 46 title, 38 status, ~160
// terminal, ~90 header = 700 of the 816 available.
const EDITOR_MAX_LINES = 9;
const LINE_H = 44;
const FONT_SIZE = 27;
const BG = '#1a1e26';
const CHROME = '#12151b';
const SIDEBAR = '#161a21';
const FALLBACK_TOKEN = '#d4d4d8';
const TYPING_PORTION = 0.7; // of the first step's window

// Every chrome label (tree rows, tabs, title bar) clips rather than overflows.
// Flex children default to min-width:auto, so a long file name would otherwise
// push past its panel and paint on top of the editor text.
const ELLIPSIS: React.CSSProperties = {
  minWidth: 0,
  overflow: 'hidden',
  whiteSpace: 'nowrap',
  textOverflow: 'ellipsis',
};

type WalkStep = {code: string; atMs: number; output?: string};
type TokenLine = ThemedToken[];

/** Per-character stream of one step's tokens, for the typing phase. */
type Char = {ch: string; color: string; line: number};
const flatten = (lines: TokenLine[]): Char[] => {
  const chars: Char[] = [];
  lines.forEach((line, li) => {
    for (const token of line) {
      for (const ch of token.content) {
        chars.push({ch, color: token.color ?? FALLBACK_TOKEN, line: li});
      }
    }
    chars.push({ch: '\n', color: '', line: li});
  });
  chars.pop();
  return chars;
};

const charRevealFrames = (chars: Char[], typingFrames: number): number[] => {
  const weights = chars.map((c, i) => {
    const jitter = 0.6 + 0.8 * random(`vsc-key-${i}`);
    return c.ch === '\n' ? 2.2 * jitter : jitter;
  });
  const total = weights.reduce((a, b) => a + b, 0) || 1;
  let acc = 0;
  return weights.map((w) => ((acc += w) / total) * typingFrames);
};

/** Lines in `cur` that differ from `prev` at the same index (simple, stable). */
const changedLines = (prev: string, cur: string): Set<number> => {
  const a = prev.split('\n');
  const b = cur.split('\n');
  const out = new Set<number>();
  for (let i = 0; i < b.length; i++) {
    if (a[i] !== b[i]) {
      out.add(i);
    }
  }
  return out;
};

export const VSCodeScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  durationInFrames: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, durationInFrames, props}) => {
  const frame = useCurrentFrame();
  const nowMs = sceneStartMs + (frame / FPS) * 1000;

  const title = String(props.title ?? '');
  const file = String(props.file ?? 'main.py');
  const language = String(props.language ?? 'python');
  const project = String(props.project ?? 'workspace');
  const files = Array.isArray(props.files) ? (props.files as string[]) : [file];
  const steps = Array.isArray(props.steps) ? (props.steps as WalkStep[]) : [];

  const [tokens, setTokens] = useState<TokenLine[][] | null>(null);
  const [handle] = useState(() => delayRender('vscode-highlight'));

  useEffect(() => {
    let cancelled = false;
    Promise.all(
      steps.map((s) =>
        codeToTokens(s.code, {lang: language as 'python', theme: 'dark-plus'})
          .then((r) => r.tokens)
          .catch(() =>
            s.code.split('\n').map<TokenLine>((ln) => [{content: ln, color: FALLBACK_TOKEN, offset: 0}]),
          ),
      ),
    ).then((all) => {
      if (!cancelled) {
        setTokens(all);
        continueRender(handle);
      }
    });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [language, handle]);

  const stepIdx = useMemo(() => {
    let idx = 0;
    steps.forEach((s, i) => {
      if (nowMs >= s.atMs) {
        idx = i;
      }
    });
    return idx;
  }, [steps, nowMs]);

  // Typing plan for step 0 (chars + reveal frames), computed once.
  const typing = useMemo(() => {
    if (!tokens || tokens.length === 0) {
      return null;
    }
    const chars = flatten(tokens[0]);
    const step0EndMs = steps.length > 1 ? steps[1].atMs : sceneStartMs + (durationInFrames / FPS) * 1000;
    const windowFrames = Math.max(1, Math.round(((step0EndMs - steps[0].atMs) / 1000) * FPS));
    const typingFrames = Math.max(1, Math.floor(windowFrames * TYPING_PORTION));
    return {chars, reveal: charRevealFrames(chars, typingFrames)};
  }, [tokens, steps, sceneStartMs, durationInFrames]);

  if (!tokens || tokens.length === 0) {
    return null;
  }

  const step = steps[stepIdx];
  const stepStartFrame = Math.round(((step.atMs - sceneStartMs) / 1000) * FPS);
  const framesIntoStep = frame - stepStartFrame;

  // Editor content: step 0 types in; later steps swap + flash changed lines.
  const stepTokens = tokens[Math.min(stepIdx, tokens.length - 1)];
  let visibleLines: TokenLine[] = stepTokens;
  let cursorLine = -1;
  let cursorOn = false;
  let typingDone = true;
  if (stepIdx === 0 && typing) {
    let visibleCount = 0;
    while (visibleCount < typing.chars.length && typing.reveal[visibleCount] <= framesIntoStep) {
      visibleCount++;
    }
    typingDone = visibleCount >= typing.chars.length;
    const lines: Char[][] = [[]];
    for (let i = 0; i < visibleCount; i++) {
      if (typing.chars[i].ch === '\n') {
        lines.push([]);
      } else {
        lines[lines.length - 1].push(typing.chars[i]);
      }
    }
    visibleLines = lines.map((ln) => ln.map((c) => ({content: c.ch, color: c.color, offset: 0})));
    // Merge adjacent same-color chars is cosmetic only; skip for determinism.
    cursorLine = lines.length - 1;
    cursorOn = !typingDone && frame % 18 < 11;
  }

  const changed = stepIdx > 0 ? changedLines(steps[stepIdx - 1].code, step.code) : new Set<number>();
  const flash = interpolate(framesIntoStep, [0, 8, 34], [0, 1, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const totalLines = visibleLines.length;
  const activeLine = stepIdx === 0 ? cursorLine : Math.min(...(changed.size ? [...changed] : [0]));
  const scroll = Math.max(0, Math.min((activeLine < 0 ? totalLines : activeLine) - (EDITOR_MAX_LINES - 4), totalLines - EDITOR_MAX_LINES));

  const col = (() => {
    if (stepIdx !== 0 || !typing) return 1;
    const lastLine = visibleLines[visibleLines.length - 1];
    return (lastLine?.reduce((n, t) => n + t.content.length, 0) ?? 0) + 1;
  })();

  // The window hugs the tallest step (+ a breathing row) instead of always
  // showing EDITOR_MAX_LINES — short snippets shouldn't float in a void.
  const tallestStep = Math.max(...steps.map((s) => s.code.split('\n').length), 1);
  const editorLines = Math.min(EDITOR_MAX_LINES, tallestStep + 1);

  const ACTIVITY: {Icon: typeof Files; active?: boolean}[] = [
    {Icon: Files, active: true},
    {Icon: Search},
    {Icon: GitBranch},
    {Icon: Bug},
    {Icon: Blocks},
  ];

  return (
    <Stage>
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={26} />
      <div style={{width: WINDOW_W, maxWidth: '100%'}}>
        <div
          style={{
            borderRadius: 16,
            overflow: 'hidden',
            border: `1px solid ${theme.surfaceBorder}`,
            boxShadow: '0 44px 110px rgba(0,0,0,0.6)',
            backgroundColor: BG,
            fontFamily: theme.fontMono,
          }}
        >
          {/* Title bar */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              height: 46,
              padding: '0 18px',
              backgroundColor: CHROME,
              borderBottom: '1px solid rgba(255,255,255,0.06)',
            }}
          >
            {['#ff5f57', '#febc2e', '#28c840'].map((c) => (
              <div key={c} style={{width: 14, height: 14, borderRadius: 7, backgroundColor: c, marginRight: 9}} />
            ))}
            <div
              style={{
                flex: 1,
                textAlign: 'center',
                fontFamily: theme.fontBody,
                fontSize: 19,
                color: 'rgba(255,255,255,0.45)',
                ...ELLIPSIS,
              }}
            >
              {file} — {project}
            </div>
            <div style={{width: 60}} />
          </div>
          <div style={{display: 'flex', height: editorLines * LINE_H + 46 + 14}}>
            {/* Activity bar */}
            <div
              style={{
                width: 62,
                backgroundColor: CHROME,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                paddingTop: 16,
                gap: 22,
              }}
            >
              {ACTIVITY.map(({Icon, active}, i) => (
                <div
                  key={i}
                  style={{
                    borderLeft: `2px solid ${active ? theme.accent : 'transparent'}`,
                    paddingLeft: 0,
                    width: '100%',
                    display: 'flex',
                    justifyContent: 'center',
                  }}
                >
                  <Icon size={27} color={active ? '#e8ecf3' : 'rgba(255,255,255,0.32)'} strokeWidth={1.8} />
                </div>
              ))}
            </div>
            {/* File tree */}
            <div
              style={{
                width: 264,
                backgroundColor: SIDEBAR,
                borderRight: '1px solid rgba(255,255,255,0.05)',
                padding: '14px 0',
                fontFamily: theme.fontBody,
                fontSize: 20,
              }}
            >
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 9,
                  padding: '5px 16px',
                  color: 'rgba(255,255,255,0.6)',
                  fontWeight: 600,
                  textTransform: 'uppercase',
                  fontSize: 16,
                  letterSpacing: 1.4,
                }}
              >
                <FolderOpen size={19} color={theme.accent} strokeWidth={2} style={{flexShrink: 0}} />
                <span style={ELLIPSIS}>{project}</span>
              </div>
              {files.map((f) => {
                const active = f === file;
                return (
                  <div
                    key={f}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 9,
                      padding: '6px 16px 6px 34px',
                      color: active ? '#e8ecf3' : 'rgba(255,255,255,0.44)',
                      backgroundColor: active ? 'rgba(255,255,255,0.07)' : 'transparent',
                      // Without this the row grows to fit the name and the
                      // text spills across the editor column underneath.
                      minWidth: 0,
                    }}
                  >
                    <FileCode2
                      size={19}
                      color={active ? '#4fc1ff' : 'rgba(255,255,255,0.36)'}
                      strokeWidth={1.8}
                      style={{flexShrink: 0}}
                    />
                    <span style={ELLIPSIS}>{f}</span>
                  </div>
                );
              })}
            </div>
            {/* Editor column */}
            <div style={{flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0}}>
              {/* Tab bar */}
              <div style={{display: 'flex', height: 46, backgroundColor: CHROME, minWidth: 0}}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 9,
                    padding: '0 24px',
                    backgroundColor: BG,
                    borderTop: `2px solid ${theme.accent}`,
                    color: '#e8ecf3',
                    fontFamily: theme.fontBody,
                    fontSize: 20,
                    minWidth: 0,
                  }}
                >
                  <FileCode2 size={19} color="#4fc1ff" strokeWidth={1.8} style={{flexShrink: 0}} />
                  <span style={ELLIPSIS}>{file}</span>
                </div>
              </div>
              {/* Editor + minimap */}
              <div style={{display: 'flex', flex: 1, overflow: 'hidden', paddingTop: 14}}>
                <div style={{flex: 1, overflow: 'hidden', minWidth: 0}}>
                  <div style={{transform: `translateY(${-scroll * LINE_H}px)`}}>
                    {visibleLines.map((line, li) => {
                      const isChanged = changed.has(li);
                      return (
                        <div
                          key={li}
                          style={{
                            display: 'flex',
                            height: LINE_H,
                            backgroundColor: isChanged ? `rgba(255, 212, 59, ${0.16 * flash})` : 'transparent',
                          }}
                        >
                          <span
                            style={{
                              width: 74,
                              flexShrink: 0,
                              textAlign: 'right',
                              paddingRight: 24,
                              fontSize: 22,
                              lineHeight: `${LINE_H}px`,
                              color: 'rgba(255,255,255,0.24)',
                              userSelect: 'none',
                            }}
                          >
                            {li + 1}
                          </span>
                          <span style={{fontSize: FONT_SIZE, lineHeight: `${LINE_H}px`, whiteSpace: 'pre'}}>
                            {line.map((t, ti) => (
                              <span key={ti} style={{color: t.color ?? FALLBACK_TOKEN}}>
                                {t.content}
                              </span>
                            ))}
                            {li === cursorLine && !typingDone ? (
                              <span
                                style={{
                                  display: 'inline-block',
                                  width: 14,
                                  height: 32,
                                  marginLeft: 1,
                                  verticalAlign: 'text-bottom',
                                  backgroundColor: cursorOn ? theme.accent : 'transparent',
                                }}
                              />
                            ) : null}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                </div>
                {/* Minimap: one bar per line, width ∝ length */}
                <div style={{width: 96, flexShrink: 0, padding: '4px 14px 0 6px', opacity: 0.55}}>
                  {steps[stepIdx].code.split('\n').slice(0, 44).map((ln, i) => (
                    <div
                      key={i}
                      style={{
                        height: 5,
                        marginBottom: 3,
                        width: `${Math.min(100, ln.trim().length * 3.4)}%`,
                        borderRadius: 2,
                        backgroundColor: changed.has(i) ? theme.accent : 'rgba(255,255,255,0.30)',
                      }}
                    />
                  ))}
                </div>
              </div>
            </div>
          </div>
          {/* Integrated terminal: the current step's really-executed output */}
          {step.output ? (
            <div
              style={{
                borderTop: '1px solid rgba(255,255,255,0.07)',
                backgroundColor: CHROME,
                padding: '14px 24px 18px',
                opacity: interpolate(framesIntoStep, [14, 30], [0, 1], {
                  extrapolateLeft: 'clamp',
                  extrapolateRight: 'clamp',
                }),
              }}
            >
              <div
                style={{
                  fontFamily: theme.fontBody,
                  fontSize: 16,
                  letterSpacing: 2.4,
                  fontWeight: 600,
                  textTransform: 'uppercase',
                  color: 'rgba(255,255,255,0.4)',
                  marginBottom: 8,
                }}
              >
                Terminal
              </div>
              <pre
                style={{
                  margin: 0,
                  whiteSpace: 'pre-wrap',
                  fontSize: 24,
                  lineHeight: 1.45,
                  color: '#d6dde6',
                  maxHeight: 150,
                  overflow: 'hidden',
                }}
              >
                <span style={{color: '#28c840'}}>$ </span>python3 {file}
                {'\n'}
                {step.output.trimEnd()}
              </pre>
            </div>
          ) : null}
          {/* Status bar */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 26,
              height: 38,
              padding: '0 20px',
              backgroundColor: theme.primary,
              color: 'rgba(255,255,255,0.92)',
              fontFamily: theme.fontBody,
              fontSize: 18,
            }}
          >
            <span style={{display: 'flex', alignItems: 'center', gap: 7}}>
              <GitBranch size={16} strokeWidth={2.2} /> main
            </span>
            <span style={{marginLeft: 'auto'}}>
              Ln {Math.max(1, (stepIdx === 0 ? cursorLine : activeLine) + 1)}, Col {col}
            </span>
            <span>Spaces: 4</span>
            <span>UTF-8</span>
            <span style={{textTransform: 'capitalize'}}>{language}</span>
          </div>
        </div>
      </div>
    </Stage>
  );
};
