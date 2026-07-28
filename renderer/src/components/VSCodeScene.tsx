import {useEffect, useMemo, useState} from 'react';
import {continueRender, delayRender, interpolate, random, spring, useCurrentFrame, useVideoConfig} from 'remotion';
import {codeToTokens, ThemedToken} from 'shiki';
import {
  Blocks,
  Bug,
  ChevronRight,
  FileCode2,
  Files,
  FolderOpen,
  GitBranch,
  Search,
  X,
} from 'lucide-react';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {SceneHeader} from './SceneHeader';
import {Stage} from './Stage';
import {CHROME, type Chrome} from './editorChrome';

// VSCodeScene is the "VS Code walkthrough": a synthesized editor — activity
// bar, file tree, tabs, minimap, integrated terminal, status bar — whose code
// evolves through timed steps.
//
// With `intro` on (the snippet path) it plays the full opening: the window
// scales up out of nothing, the file is picked out of the tree, its tab slides
// in, and only then does the cursor start typing. Mid-lesson (`intro` off) the
// editor is simply already there, because cutting to a fresh window every time
// a lesson shows code reads as a restart.
//
// A step marked `run` opens the integrated terminal under the editor: the
// command types itself, then the output streams in line by line. That output
// comes from the verify stage — it is what the interpreter really printed.
//
// Everything is frame-driven (research: the only deterministic approach; real
// editors and screen capture are wall-clock). The chrome is VS Code-like but
// carries no Microsoft branding.

// Window geometry.
//
// The proportions matter more than the raw size: a 1680-wide window holding
// six lines of code is a letterbox strip, and it reads as small no matter how
// much of the frame it covers. 1520 × ~560 is roughly a real editor window, so
// the same code reads noticeably larger.
//
// The editor pane is a constant eight lines. It does not resize between steps
// (a window that changes shape mid-clip reads as a jump cut) and it does not
// hug short files (a three-line pane looks like a screenshot of a mistake —
// real editors show empty rows below the code).
//
// Height budget, against Stage's 896 usable: 91 header + 46 title bar +
// 382 editor pane + 250 terminal + 38 status = 807, putting the window's
// bottom edge at y=871 — clear of the caption card's top at y=919. Snippets
// render with captions on, so that clearance is not optional; Stage's
// CAPTION_SAFE assumes captions are off and does not reserve enough on its
// own.
const WINDOW_W = 1520;
const EDITOR_LINES = 7;
const LINE_H = 46;
const FONT_SIZE = 29;
const TERMINAL_H = 250;
const TAB_BAR_H = 46;
const EDITOR_PAD_TOP = 14;
// Breathing room under the last line. Without it the code butts against the
// status bar and the pane reads as clipped rather than as scrolled.
const EDITOR_PAD_BOTTOM = 12;
const SIDEBAR_W = 226;
const ACTIVITY_W = 58;
const MINIMAP_W = 78;
// Minimap metrics. The character width is set so a ~44-character line — about
// the widest this pane shows before it scrolls — just fills the column.
const MINIMAP_PAD = 8;
const MINIMAP_CHAR_W = 1.35;
const MINIMAP_BLOCK_H = 4;
const MINIMAP_LINE_H = 7;
const MINIMAP_LINES = 44;
const GUTTER_W = 74;
const TYPING_PORTION = 0.7; // of the first step's window

// Opening choreography. The window scales up from the scene's first frame;
// the file is then picked out of the tree and its tab opens, timed backwards
// from the first keystroke so the gesture lands right before typing however
// long the intro narration ran.
const INTRO = {
  windowInFrames: 18,
  // Frames before the first keystroke that each phase begins/ends.
  treeHoverBefore: [26, 20],
  treeClickBefore: [20, 12],
  tabInBefore: [14, 3],
  // The pointer withdraws between the click and the first keystroke — it has
  // done its job, and the hand has moved to the keyboard.
  pointerLeaveBefore: [12, 2],
} as const;

// Terminal choreography, in frames from the moment the step's run begins.
const RUN = {
  drawerFrames: 14,
  commandStart: 9,
  framesPerChar: 1.5,
  outputGap: 9,
  framesPerOutputLine: 2.5,
  // Chrome inside the drawer: the panel tab row plus its padding.
  tabsH: 44,
  padH: 14,
  lineHeight: 1.42,
  minFont: 15,
  maxFont: 26,
} as const;

// Every chrome label (tree rows, tabs, title bar) clips rather than overflows.
// Flex children default to min-width:auto, so a long file name would otherwise
// push past its panel and paint on top of the editor text.
/**
 * The selection wash behind a file-tree row.
 *
 * An alpha suffix on the primary rather than a white or black overlay: white
 * over the light sidebar is invisible and black over the dark one is a hole,
 * and the row is *selected*, which is a thing the brand colour should say.
 */
const selectedTint = (primary: string, amount: number): string =>
  amount <= 0
    ? 'transparent'
    : `${primary}${Math.round(Math.min(1, amount) * 38)
        .toString(16)
        .padStart(2, '0')}`;

const ELLIPSIS: React.CSSProperties = {
  minWidth: 0,
  overflow: 'hidden',
  whiteSpace: 'nowrap',
  textOverflow: 'ellipsis',
};

/**
 * The mouse pointer, during the opening gesture.
 *
 * The scene already timed a hover and a click on the file it is about to open —
 * the row lit, then it selected — and drew no cursor at all, so the highlight
 * moved on its own like a haunted menu. A pointer is the difference between
 * "the UI is animating" and "somebody is using this".
 *
 * It renders *inside* the row it is clicking, positioned relatively, rather
 * than at computed window coordinates. The tree is a flex column with padding;
 * any absolute position for this would be a second copy of that layout, and
 * would part company with it the first time a font size changed.
 */
const MousePointer: React.FC<{
  /** 0-1 as the pointer travels in. */
  arrive: number;
  /** 0-1 as the button goes down. */
  press: number;
  /** 0-1 as it leaves once the work is done. */
  leave: number;
  color: string;
  shadow: string;
}> = ({arrive, press, leave, color, shadow}) => {
  if (arrive <= 0 || leave >= 1) return null;
  // It comes in from below-right — the direction a hand on a trackpad moves
  // from — and eases to the row. The press is a small dip, not a scale: a
  // cursor that grows when it clicks is a cursor nobody has ever seen.
  const dx = (1 - arrive) * 120 + leave * 40;
  const dy = (1 - arrive) * 90 + leave * 60;
  return (
    <div
      style={{
        position: 'absolute',
        left: 128,
        top: 16,
        opacity: Math.min(arrive, 1 - leave),
        transform: `translate(${dx}px, ${dy + press * 2}px)`,
        pointerEvents: 'none',
      }}
    >
      {/* The click ring: a single soft pulse under the tip as the button goes
          down, which is what makes the selection read as caused. */}
      {press > 0 && (
        <div
          style={{
            position: 'absolute',
            left: -13,
            top: -9,
            width: 34,
            height: 34,
            borderRadius: 17,
            border: `2px solid ${color}`,
            opacity: Math.max(0, 0.55 - press * 0.55),
            transform: `scale(${0.4 + press * 1.1})`,
          }}
        />
      )}
      <svg width={26} height={32} viewBox="0 0 26 32" style={{filter: `drop-shadow(0 2px 3px ${shadow})`}}>
        <path d="M2 1 L2 24 L8 18.5 L12 27.5 L16 25.5 L12 17 L20 16.5 Z" fill={color} stroke="#0b0b0d" strokeWidth={1.6} strokeLinejoin="round" />
      </svg>
    </div>
  );
};

type WalkStep = {
  code: string;
  atMs: number;
  output?: string;
  /** Execute the file during this step (opens the terminal drawer). */
  run?: boolean;
  /** When the terminal opens; defaults to the step's own start. */
  runAtMs?: number;
  /** The command the terminal types, e.g. "python3 loops.py". */
  command?: string;
};

/** Frames from the scene's start for an absolute scene-graph timestamp. */
const framesFrom = (sceneStartMs: number, atMs: number): number =>
  Math.round(((atMs - sceneStartMs) / 1000) * FPS);
type TokenLine = ThemedToken[];

/** Per-character stream of one step's tokens, for the typing phase. */
type Char = {ch: string; color: string; line: number};
const flatten = (lines: TokenLine[], fallback: string): Char[] => {
  const chars: Char[] = [];
  lines.forEach((line, li) => {
    for (const token of line) {
      for (const ch of token.content) {
        chars.push({ch, color: token.color ?? fallback, line: li});
      }
    }
    chars.push({ch: '\n', color: '', line: li});
  });
  chars.pop();
  return chars;
};

/**
 * Fallback reveal frames, for a scene graph with no `keystrokes`.
 *
 * The schedule is Go's now (internal/pipeline/typing.go) — it models a real
 * rhythm and it is what the keystroke click track is generated from, so the two
 * cannot drift. This remains only so an older scene graph, or a lesson whose
 * walkthrough predates the field, still types instead of appearing all at once.
 * It is uniform-with-jitter, which is exactly the thing Go replaced.
 */
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

/** Leading-space depth of a line, in 4-space indent units. */
const indentDepth = (line: string): number => {
  const spaces = line.length - line.trimStart().length;
  return Math.floor(spaces / 4);
};

// --- suggestions ------------------------------------------------------------
//
// The completion popup is the single strongest "somebody is coding here"
// signal an editor gives off, and this scene had nothing like it: characters
// simply appeared, which is what a text animation does, not what an editor does.
//
// The candidates come from **the file's own identifiers**, not from an invented
// vocabulary of Python builtins. That is not a shortcut — it is what a real
// editor's word-based completion actually does, it means the suggestions are
// always relevant to the code on screen, and it means this cannot go stale
// against a language it was never taught. It is also the honest reading of "the
// editor knows about the code inside it".

const IDENT = /[A-Za-z_][A-Za-z0-9_]*/g;

/** Every identifier in the finished buffer, longest-first, de-duplicated. */
const identifiersIn = (code: string): string[] => {
  const seen = new Set<string>();
  for (const m of code.matchAll(IDENT)) {
    if (m[0].length >= 3) seen.add(m[0]);
  }
  return [...seen];
};

/** The word fragment immediately before the caret, if any. */
const fragmentAt = (line: string): string => {
  const m = /[A-Za-z_][A-Za-z0-9_]*$/.exec(line);
  return m ? m[0] : '';
};

/**
 * Up to `max` completions for `fragment`.
 *
 * The exact match is kept rather than filtered out: a real editor still lists
 * the word you have finished typing, and dropping it made the popup vanish for
 * one frame at the end of every word — a flicker that read as a glitch.
 */
const suggestionsFor = (fragment: string, ids: string[], max = 4): string[] => {
  if (fragment.length < 2) return [];
  const lower = fragment.toLowerCase();
  return ids
    .filter((id) => id.toLowerCase().startsWith(lower))
    .sort((a, b) => a.length - b.length || a.localeCompare(b))
    .slice(0, max);
};

export const VSCodeScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  durationInFrames: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, durationInFrames, props}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const nowMs = sceneStartMs + (frame / FPS) * 1000;

  const title = String(props.title ?? '');
  const file = String(props.file ?? 'main.py');
  const language = String(props.language ?? 'python');
  const project = String(props.project ?? 'workspace');
  const files = Array.isArray(props.files) ? (props.files as string[]) : [file];
  const steps = Array.isArray(props.steps) ? (props.steps as WalkStep[]) : [];
  const intro = props.intro === true;

  const ui = CHROME[theme.mode];
  // The "this one is active" rail, on the tab's top edge and the activity bar.
  // The accent is a saturated yellow picked to sit on the dark stage; on the
  // light chrome it is a highlighter stroke nobody can see, so light mode marks
  // the active item with the primary instead.
  const marker = theme.mode === 'light' ? theme.primary : theme.accent;

  const [tokens, setTokens] = useState<TokenLine[][] | null>(null);
  const [handle] = useState(() => delayRender('vscode-highlight'));

  useEffect(() => {
    let cancelled = false;
    Promise.all(
      steps.map((s) =>
        // The syntax theme follows the chrome. Dark-plus tokens on a white
        // editor are a specific kind of unreadable: the comment green and the
        // string orange are both picked to sit on #1e1e1e.
        codeToTokens(s.code, {lang: language as 'python', theme: ui.shiki})
          .then((r) => r.tokens)
          .catch(() =>
            s.code.split('\n').map<TokenLine>((ln) => [{content: ln, color: ui.token, offset: 0}]),
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
  }, [language, handle, ui.shiki]);

  const stepIdx = useMemo(() => {
    let idx = 0;
    steps.forEach((s, i) => {
      if (nowMs >= s.atMs) {
        idx = i;
      }
    });
    return idx;
  }, [steps, nowMs]);

  // When the first character lands. The planner supplies it (typeAtMs) so the
  // opening gesture fills exactly the gap the intro narration left; without it
  // the scene starts typing at once, which is the lesson path.
  const typeStartFrame =
    typeof props.typeAtMs === 'number'
      ? Math.max(0, framesFrom(sceneStartMs, props.typeAtMs))
      : 0;

  // Typing is measured from step 0's own start, so the scene-relative
  // typeAtMs becomes a delay inside that step.
  const typeDelay = steps.length
    ? Math.max(0, typeStartFrame - framesFrom(sceneStartMs, steps[0].atMs))
    : 0;

  // Typing plan for step 0 (chars + reveal frames), computed once.
  //
  // `reveal` is measured in frames from the start of step 0, which is how the
  // draw loop below consumes it. When Go supplied a keystroke schedule that is
  // simply a conversion; otherwise it falls back to the local estimate.
  const typing = useMemo(() => {
    if (!tokens || tokens.length === 0) {
      return null;
    }
    const chars = flatten(tokens[0], ui.token);
    const schedule = Array.isArray(props.keystrokes) ? (props.keystrokes as number[]) : null;
    if (schedule && schedule.length === chars.length) {
      // Absolute scene-graph ms -> frames since typing began, which is the base
      // the draw loop below uses (`framesIntoStep - typeDelay`). Measuring from
      // step 0's start instead would double-count the intro, and the file would
      // finish a second and a half early.
      //
      // Go owns the rhythm — pauses at line ends, free indentation, auto-closed
      // brackets — and the same numbers drive the keystroke click track, so the
      // sound lands on the character rather than near it.
      return {
        chars,
        reveal: schedule.map((ms) => framesFrom(sceneStartMs, ms) - typeStartFrame),
      };
    }
    const step0EndMs = steps.length > 1 ? steps[1].atMs : sceneStartMs + (durationInFrames / FPS) * 1000;
    const windowFrames = Math.max(
      1,
      Math.round(((step0EndMs - steps[0].atMs) / 1000) * FPS) - typeDelay,
    );
    const typingFrames = Math.max(1, Math.floor(windowFrames * TYPING_PORTION));
    return {chars, reveal: charRevealFrames(chars, typingFrames)};
  }, [tokens, steps, sceneStartMs, durationInFrames, typeDelay, typeStartFrame, props.keystrokes, ui.token]);

  if (!tokens || tokens.length === 0) {
    return null;
  }

  const step = steps[stepIdx];
  const stepStartFrame = Math.round(((step.atMs - sceneStartMs) / 1000) * FPS);
  const framesIntoStep = frame - stepStartFrame;

  // --- opening choreography -------------------------------------------------
  // Phases are timed backwards from the first keystroke, so a long intro beat
  // holds on the open window rather than stretching the gesture into slow
  // motion. Clamped to non-negative in case the intro is very short.
  const before = ([from, to]: readonly [number, number]) => {
    if (!intro) {
      return 1;
    }
    const start = Math.max(0, typeStartFrame - from);
    const end = Math.max(start + 1, typeStartFrame - to);
    return interpolate(frame, [start, end], [0, 1], {
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp',
    });
  };
  const windowIn = intro
    ? spring({frame, fps, config: {damping: 200, mass: 0.7}, durationInFrames: INTRO.windowInFrames})
    : 1;
  const treeHover = before(INTRO.treeHoverBefore);
  const treeClick = before(INTRO.treeClickBefore);
  const tabIn = before(INTRO.tabInBefore);
  // The pointer's exit: it leaves over the frames between the click landing and
  // the first keystroke. A cursor still sitting on the file tree while the code
  // types itself is a cursor nobody is holding.
  const pointerLeave = before(INTRO.pointerLeaveBefore);

  // --- editor content -------------------------------------------------------
  const stepTokens = tokens[Math.min(stepIdx, tokens.length - 1)];
  let visibleLines: TokenLine[] = stepTokens;
  let cursorLine = -1;
  let cursorOn = false;
  let typingDone = true;
  if (stepIdx === 0 && typing) {
    const framesTyping = framesIntoStep - typeDelay;
    let visibleCount = 0;
    while (visibleCount < typing.chars.length && typing.reveal[visibleCount] <= framesTyping) {
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
    cursorLine = lines.length - 1;
    cursorOn = !typingDone && frame % 16 < 9;
  }

  // Completions for whatever word is half-typed at the caret, drawn from the
  // finished buffer's own identifiers. Only while typing: a popup hanging over
  // finished code is a popup nobody dismissed.
  //
  // Not memoised, deliberately. This sits below the early return above, so a
  // hook here changes the hook count between renders — which is exactly the
  // React invariant it broke the first time. A regex over a few hundred
  // characters per frame is not worth a second early-return branch.
  const bufferIdents = identifiersIn(step.code);
  const fragment =
    stepIdx === 0 && !typingDone && cursorLine >= 0
      ? fragmentAt((visibleLines[cursorLine] ?? []).map((t) => t.content).join(''))
      : '';
  const suggestions = fragment ? suggestionsFor(fragment, bufferIdents) : [];

  const changed = stepIdx > 0 ? changedLines(steps[stepIdx - 1].code, step.code) : new Set<number>();
  const flash = interpolate(framesIntoStep, [0, 8, 34], [0, 1, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const totalLines = visibleLines.length;
  const activeLine = stepIdx === 0 ? cursorLine : Math.min(...(changed.size ? [...changed] : [0]));

  const col = (() => {
    if (stepIdx !== 0 || !typing) return 1;
    const lastLine = visibleLines[visibleLines.length - 1];
    return (lastLine?.reduce((n, t) => n + t.content.length, 0) ?? 0) + 1;
  })();

  const editorLines = EDITOR_LINES;
  const editorPaneH = editorLines * LINE_H + TAB_BAR_H + EDITOR_PAD_TOP + EDITOR_PAD_BOTTOM;

  // --- terminal drawer ------------------------------------------------------
  // A step with output but no explicit run is the lesson path: the terminal
  // simply belongs to the step, opening as it begins.
  const runs = step.run === true || (step.run === undefined && Boolean(step.output));
  const runAtMs = step.runAtMs ?? step.atMs;
  const framesIntoRun = frame - framesFrom(sceneStartMs, runAtMs);
  const drawer =
    runs && framesIntoRun >= 0
      ? spring({
          frame: framesIntoRun,
          fps,
          config: {damping: 200, mass: 0.55},
          durationInFrames: RUN.drawerFrames,
        })
      : 0;
  const terminalH = TERMINAL_H * drawer;

  const command = step.command ?? `python3 ${file}`;
  const commandChars = Math.max(
    0,
    Math.floor((framesIntoRun - RUN.commandStart) / RUN.framesPerChar),
  );
  const typedCommand = command.slice(0, commandChars);
  const commandDone = commandChars >= command.length;
  const commandEndFrame = RUN.commandStart + command.length * RUN.framesPerChar;
  const outputLines = (step.output ?? '').trimEnd().split('\n');
  // The drawer is a fixed height, so long output has to be typeset to fit
  // rather than clipped. Clipping loses the last line — which, in a clip built
  // around running the code, is usually the punchline.
  const terminalFont = Math.max(
    RUN.minFont,
    Math.min(
      RUN.maxFont,
      Math.floor((TERMINAL_H - RUN.tabsH - RUN.padH) / ((outputLines.length + 1) * RUN.lineHeight)),
    ),
  );
  const visibleOutput = commandDone
    ? Math.max(
        0,
        Math.floor((framesIntoRun - commandEndFrame - RUN.outputGap) / RUN.framesPerOutputLine),
      )
    : 0;
  // The accent glow behind the window swells as the program runs — the one
  // moment in the clip that deserves emphasis.
  const runGlow = drawer * interpolate(framesIntoRun, [0, 24], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const scroll = Math.max(
    0,
    Math.min((activeLine < 0 ? totalLines : activeLine) - (editorLines - 3), totalLines - editorLines),
  );

  const ACTIVITY: {Icon: typeof Files; active?: boolean}[] = [
    {Icon: Files, active: true},
    {Icon: Search},
    {Icon: GitBranch},
    {Icon: Bug},
    {Icon: Blocks},
  ];

  const codeLines = step.code.split('\n');

  return (
    <Stage>
      <SceneHeader theme={theme} title={title} size="compact" marginBottom={26} />
      <div style={{width: WINDOW_W, maxWidth: '100%', position: 'relative'}}>
        {/* Accent bloom behind the window; swells while the program runs. */}
        <div
          style={{
            position: 'absolute',
            inset: -60,
            borderRadius: 80,
            background: `radial-gradient(60% 55% at 50% 82%, ${theme.accent}, transparent 70%)`,
            opacity: 0.05 + runGlow * 0.16,
            filter: 'blur(60px)',
          }}
        />
        <div
          style={{
            position: 'relative',
            borderRadius: 16,
            overflow: 'hidden',
            border: `1px solid ${ui.outline}`,
            boxShadow: ui.shadow,
            backgroundColor: ui.bg,
            fontFamily: theme.fontMono,
            opacity: windowIn,
            transform: `translateY(${(1 - windowIn) * 38}px) scale(${0.955 + windowIn * 0.045})`,
          }}
        >
          {/* Title bar */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              height: 46,
              padding: '0 18px',
              backgroundColor: ui.chrome,
              borderBottom: `1px solid ${ui.border}`,
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
                color: ui.dim,
                opacity: tabIn,
                ...ELLIPSIS,
              }}
            >
              {file} — {project}
            </div>
            <div style={{width: 60}} />
          </div>
          <div style={{display: 'flex', height: editorPaneH + terminalH}}>
            {/* Activity bar */}
            <div
              style={{
                width: ACTIVITY_W,
                flexShrink: 0,
                backgroundColor: ui.chrome,
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
                    borderLeft: `2px solid ${active ? marker : 'transparent'}`,
                    width: '100%',
                    display: 'flex',
                    justifyContent: 'center',
                  }}
                >
                  <Icon size={25} color={active ? ui.text : ui.faint} strokeWidth={1.8} />
                </div>
              ))}
            </div>
            {/* File tree */}
            <div
              style={{
                width: SIDEBAR_W,
                flexShrink: 0,
                backgroundColor: ui.sidebar,
                borderRight: `1px solid ${ui.border}`,
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
                  color: ui.dim,
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
                const isTarget = f === file;
                // The opening gesture: the row lights under a hover, then the
                // click lands and it becomes the selected file.
                const selected = isTarget ? Math.max(treeHover * 0.45, treeClick) : 0;
                return (
                  <div
                    key={f}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 9,
                      padding: '6px 16px 6px 34px',
                      color: selected > 0.6 ? ui.text : ui.dim,
                      backgroundColor: selectedTint(theme.primary, selected),
                      borderLeft: `2px solid ${
                        selected > 0.6 ? marker : 'transparent'
                      }`,
                      // Without this the row grows to fit the name and the
                      // text spills across the editor column underneath.
                      minWidth: 0,
                      // So the pointer can sit on this row without knowing
                      // where the row is.
                      position: 'relative',
                    }}
                  >
                    <FileCode2
                      size={19}
                      color={selected > 0.6 ? ui.fileIcon : ui.faint}
                      strokeWidth={1.8}
                      style={{flexShrink: 0}}
                    />
                    <span style={ELLIPSIS}>{f}</span>
                    {isTarget && intro && (
                      <MousePointer
                        arrive={treeHover}
                        press={treeClick}
                        leave={pointerLeave}
                        color={ui.text}
                        shadow="rgba(0,0,0,0.5)"
                      />
                    )}
                  </div>
                );
              })}
            </div>
            {/* Editor column */}
            <div style={{flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0}}>
              {/* Tab bar — the tab slides in as the file opens */}
              <div style={{display: 'flex', height: TAB_BAR_H, backgroundColor: ui.chrome, minWidth: 0, flexShrink: 0}}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 9,
                    padding: '0 20px',
                    backgroundColor: ui.bg,
                    borderTop: `2px solid ${marker}`,
                    color: ui.text,
                    fontFamily: theme.fontBody,
                    fontSize: 20,
                    minWidth: 0,
                    opacity: tabIn,
                    transform: `translateX(${(1 - tabIn) * -22}px)`,
                  }}
                >
                  <FileCode2 size={19} color={ui.fileIcon} strokeWidth={1.8} style={{flexShrink: 0}} />
                  <span style={ELLIPSIS}>{file}</span>
                  <X size={17} color={ui.faint} strokeWidth={2.2} style={{flexShrink: 0}} />
                </div>
              </div>
              {/* Editor + minimap */}
              <div
                style={{
                  display: 'flex',
                  height: editorPaneH - TAB_BAR_H,
                  flexShrink: 0,
                  overflow: 'hidden',
                  paddingTop: EDITOR_PAD_TOP,
                  opacity: tabIn,
                }}
              >
                <div style={{flex: 1, overflow: 'hidden', minWidth: 0}}>
                  <div style={{transform: `translateY(${-scroll * LINE_H}px)`}}>
                    {visibleLines.map((line, li) => {
                      const isChanged = changed.has(li);
                      const isActive = li === activeLine && !typingDone;
                      const text = line.map((t) => t.content).join('');
                      const depth = indentDepth(codeLines[li] ?? text);
                      return (
                        <div
                          key={li}
                          style={{
                            display: 'flex',
                            height: LINE_H,
                            position: 'relative',
                            backgroundColor: isChanged
                              ? `rgba(255, 212, 59, ${0.16 * flash})`
                              : isActive
                                ? ui.activeLine
                                : 'transparent',
                          }}
                        >
                          <span
                            style={{
                              width: GUTTER_W,
                              flexShrink: 0,
                              textAlign: 'right',
                              paddingRight: 24,
                              fontSize: 22,
                              lineHeight: `${LINE_H}px`,
                              color: isActive ? ui.dim : ui.faint,
                              userSelect: 'none',
                            }}
                          >
                            {li + 1}
                          </span>
                          <span style={{position: 'relative', fontSize: FONT_SIZE, lineHeight: `${LINE_H}px`, whiteSpace: 'pre'}}>
                            {/* Indent guides — the detail that separates a
                                real editor from a code card. */}
                            {Array.from({length: depth}, (_, d) => (
                              <span
                                key={d}
                                style={{
                                  position: 'absolute',
                                  left: d * 4 * (FONT_SIZE * 0.6) + 1,
                                  top: 6,
                                  bottom: 6,
                                  width: 1,
                                  backgroundColor: ui.indent,
                                }}
                              />
                            ))}
                            {line.map((t, ti) => (
                              <span key={ti} style={{color: t.color ?? ui.token}}>
                                {t.content}
                              </span>
                            ))}
                            {li === cursorLine && !typingDone ? (
                              <span
                                style={{
                                  display: 'inline-block',
                                  width: 3,
                                  height: 32,
                                  marginLeft: 1,
                                  verticalAlign: 'text-bottom',
                                  backgroundColor: cursorOn ? theme.accent : 'transparent',
                                }}
                              />
                            ) : null}
                            {/* The completion popup, hanging off the caret so
                                it follows the text without any column
                                arithmetic. */}
                            {li === cursorLine && !typingDone && suggestions.length > 0 && (
                              <span
                                style={{
                                  position: 'absolute',
                                  left: 0,
                                  top: LINE_H - 4,
                                  zIndex: 5,
                                  display: 'flex',
                                  flexDirection: 'column',
                                  minWidth: 300,
                                  backgroundColor: ui.chrome,
                                  border: `1px solid ${ui.border}`,
                                  borderRadius: 6,
                                  boxShadow: '0 10px 26px rgba(0,0,0,0.45)',
                                  overflow: 'hidden',
                                  fontSize: 20,
                                }}
                              >
                                {suggestions.map((s, si) => (
                                  <span
                                    key={s}
                                    style={{
                                      display: 'flex',
                                      alignItems: 'center',
                                      gap: 10,
                                      padding: '5px 12px',
                                      // The first row is the selected one, the
                                      // way it is the moment the list opens.
                                      backgroundColor: si === 0 ? selectedTint(theme.primary, 1) : 'transparent',
                                      color: si === 0 ? ui.text : ui.dim,
                                    }}
                                  >
                                    <Blocks size={17} color={theme.accent} strokeWidth={2} style={{flexShrink: 0}} />
                                    {/* The typed prefix is highlighted, which
                                        is how a real list shows you why these
                                        entries are the matches. */}
                                    <span style={{whiteSpace: 'pre'}}>
                                      <span style={{color: theme.accent, fontWeight: 700}}>
                                        {s.slice(0, fragment.length)}
                                      </span>
                                      {s.slice(fragment.length)}
                                    </span>
                                  </span>
                                ))}
                              </span>
                            )}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                </div>
                {/* Minimap.
                    One block per *token*, coloured by that token, rather than
                    one grey bar per line. It costs nothing — the tokens are
                    already here for the editor — and it is the difference
                    between a miniature of this code and a row of ticks that
                    would look the same whatever the file said. */}
                <div
                  style={{
                    width: MINIMAP_W,
                    flexShrink: 0,
                    position: 'relative',
                    backgroundColor: ui.minimap,
                    borderLeft: `1px solid ${ui.border}`,
                    padding: `2px ${MINIMAP_PAD}px 0`,
                    overflow: 'hidden',
                  }}
                >
                  {/* The viewport: which slice of the file the pane is
                      showing. Only drawn when the file is actually longer than
                      the pane — on a file that fits, a box around the whole
                      miniature says nothing and reads as a stray outline. */}
                  {totalLines > editorLines && (
                    <div
                      style={{
                        position: 'absolute',
                        left: 0,
                        right: 0,
                        top: 2 + scroll * MINIMAP_LINE_H,
                        height: editorLines * MINIMAP_LINE_H,
                        backgroundColor: ui.activeLine,
                        borderTop: `1px solid ${ui.border}`,
                        borderBottom: `1px solid ${ui.border}`,
                      }}
                    />
                  )}
                  {visibleLines.slice(0, MINIMAP_LINES).map((line, i) => (
                    <div
                      key={i}
                      style={{
                        position: 'relative',
                        display: 'flex',
                        gap: 1,
                        height: MINIMAP_BLOCK_H,
                        marginBottom: MINIMAP_LINE_H - MINIMAP_BLOCK_H,
                      }}
                    >
                      {changed.has(i) && (
                        <div
                          style={{
                            position: 'absolute',
                            left: -MINIMAP_PAD,
                            top: -1,
                            bottom: -1,
                            width: 3,
                            backgroundColor: theme.accent,
                            opacity: flash * 0.9 + 0.35,
                          }}
                        />
                      )}
                      {line.map((t, ti) => {
                        // Whitespace keeps its width and paints nothing, which
                        // is what gives the miniature its indentation — the
                        // single cue that makes it read as code.
                        const blank = t.content.trim() === '';
                        return (
                          <div
                            key={ti}
                            style={{
                              width: t.content.length * MINIMAP_CHAR_W,
                              backgroundColor: blank ? 'transparent' : (t.color ?? ui.token),
                              opacity: blank ? 0 : 0.7,
                              borderRadius: 1,
                            }}
                          />
                        );
                      })}
                    </div>
                  ))}
                </div>
              </div>
              {/* Integrated terminal: slides up, types the command, streams
                  the output the interpreter really produced. */}
              <div
                style={{
                  height: terminalH,
                  overflow: 'hidden',
                  backgroundColor: ui.panel,
                  borderTop: `1px solid ${drawer > 0.05 ? ui.border : 'transparent'}`,
                  flexShrink: 0,
                }}
              >
                <div
                  style={{
                    padding: `0 24px ${RUN.padH}px`,
                    height: TERMINAL_H,
                    display: 'flex',
                    flexDirection: 'column',
                  }}
                >
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 26,
                      height: RUN.tabsH,
                      flexShrink: 0,
                      fontFamily: theme.fontBody,
                      fontSize: 15,
                      letterSpacing: 1.8,
                      fontWeight: 600,
                      textTransform: 'uppercase',
                    }}
                  >
                    {['Problems', 'Output', 'Terminal'].map((tab) => {
                      const active = tab === 'Terminal';
                      return (
                        <span
                          key={tab}
                          style={{
                            color: active ? ui.text : ui.faint,
                            borderBottom: `2px solid ${active ? marker : 'transparent'}`,
                            paddingBottom: 6,
                          }}
                        >
                          {tab}
                        </span>
                      );
                    })}
                  </div>
                  <div
                    style={{
                      fontSize: terminalFont,
                      lineHeight: RUN.lineHeight,
                      color: ui.terminalText,
                      paddingTop: 8,
                    }}
                  >
                    <div style={{display: 'flex', alignItems: 'baseline', gap: 10}}>
                      <ChevronRight
                        size={terminalFont * 0.82}
                        color={ui.prompt}
                        strokeWidth={3}
                        style={{flexShrink: 0}}
                      />
                      <span style={{whiteSpace: 'pre'}}>
                        {typedCommand}
                        {!commandDone && framesIntoRun >= RUN.commandStart ? (
                          <span
                            style={{
                              display: 'inline-block',
                              width: Math.round(terminalFont * 0.5),
                              height: terminalFont,
                              verticalAlign: 'text-bottom',
                              backgroundColor: frame % 16 < 9 ? theme.accent : 'transparent',
                            }}
                          />
                        ) : null}
                      </span>
                    </div>
                    {outputLines.slice(0, visibleOutput).map((ln, i) => (
                      <div
                        key={i}
                        style={{
                          whiteSpace: 'pre-wrap',
                          opacity: interpolate(
                            framesIntoRun,
                            [
                              commandEndFrame + RUN.outputGap + i * RUN.framesPerOutputLine,
                              commandEndFrame + RUN.outputGap + i * RUN.framesPerOutputLine + 4,
                            ],
                            [0, 1],
                            {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'},
                          ),
                        }}
                      >
                        {ln}
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
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
