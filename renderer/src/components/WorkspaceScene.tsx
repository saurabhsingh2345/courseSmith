import {useEffect, useMemo, useState} from 'react';
import {continueRender, delayRender, interpolate, useCurrentFrame} from 'remotion';
import {codeToTokens, ThemedToken} from 'shiki';
import {Bug, ChevronRight, FileCode2, Files, FolderOpen, GitBranch, Search} from 'lucide-react';
import {FPS} from '../types';
import {ResolvedTheme} from '../theme/theme';
import {FRAME_H, FRAME_W} from './Stage';
import {CHROME, type Chrome} from './editorChrome';

// WorkspaceScene is a multi-file project, shot like a screen recording.
//
// The `walkthrough` scene beside it is a *window*: an editor floating on the
// design system's stage, sized so eight lines of one file read large. This one
// is the opposite decision on purpose — the editor fills the frame, there is no
// stage behind it, and the camera moves. It is imitating a capture of somebody
// working, because the thing it has to show is not a snippet but a program:
// which file a function lives in, which file imports it, and the terminal
// agreeing with both.
//
// Two consequences follow from "screen recording" and they are the whole
// design:
//
//   1. **One scene for the whole clip.** A scene per beat would remount the
//      editor every few seconds, and a remount is a cut. Nobody cuts in the
//      middle of a screen capture. The beats are timed steps inside one
//      continuous editor instead.
//   2. **The camera moves rather than the layout.** Everything is laid out once
//      at frame size and a transform picks the part of it worth looking at.
//      Panning to the tree and zooming into a function are the two gestures a
//      real screencast uses to direct attention, and they cost nothing here
//      because the content underneath never re-flows.
//
// The model never supplies a coordinate. It names a subject — `tree`, `code`,
// `terminal` — and the geometry below turns that into a camera. Same bargain
// the story template's staging makes, for the same reason: a model handed x/y
// frames the gap between two panels.
//
// Everything is frame-driven; nothing reads a clock.

const ACTIVITY_W = 62;
const SIDEBAR_W = 300;
const TAB_BAR_H = 52;
const TITLE_H = 44;
const STATUS_H = 40;
const TERMINAL_H = 300;
const LINE_H = 40;
const FONT_SIZE = 25;
const GUTTER_W = 86;
const MINIMAP_W = 96;
const EDITOR_PAD_TOP = 14;

/** The editor's own coordinate space: the whole frame, always. */
const BOARD = {w: FRAME_W, h: FRAME_H};

/** Where the editor pane sits inside it — the camera targets are built on it. */
const PANE = {
  x: ACTIVITY_W + SIDEBAR_W,
  y: TITLE_H + TAB_BAR_H,
  w: BOARD.w - ACTIVITY_W - SIDEBAR_W,
  h: BOARD.h - TITLE_H - TAB_BAR_H - STATUS_H,
};

/**
 * How tall the code area is: the longest file in the project, plus headroom.
 *
 * A real editor fills its pane and leaves empty rows under a short file, and
 * imitating that here put five hundred blank pixels between a four-line program
 * and the terminal proving it works — the two things the shot is about, held
 * apart by nothing. Sized from the longest file rather than the current one so
 * the layout does not jump when the clip switches tabs.
 */
const codeAreaH = (files: {code: string}[]): number => {
  const longest = files.reduce((n, f) => Math.max(n, f.code.trimEnd().split('\n').length), 1);
  return Math.min(PANE.h, (longest + 2) * LINE_H + EDITOR_PAD_TOP);
};

/** How many lines that area shows before the file has to scroll. */
const visibleLines = (h: number): number => Math.max(1, Math.floor((h - EDITOR_PAD_TOP) / LINE_H));

/**
 * How long a camera move takes, and how far it goes.
 *
 * Slower and shallower than instinct suggests. A screencast zoom that snaps is
 * a jump cut with extra steps, and one that goes past about 1.5x stops reading
 * as "look here" and starts reading as a different shot.
 */
const CAM = {
  frames: 26,
  wide: 1,
  code: 1.34,
  tree: 1.5,
  tabs: 1.55,
  terminal: 1.18,
} as const;

type Cam = {x: number; y: number; zoom: number};

type ProjectFile = {path: string; code: string};
type Step = {
  startMs: number;
  endMs: number;
  file: string;
  through: number;
  focus: string;
  run: boolean;
  caption?: string;
};

type TokenLine = ThemedToken[];

/**
 * The camera for a focus, given where the typing currently is.
 *
 * Clamped so the frame never leaves the editor: at zoom 1.5 the visible half-
 * width is a third of the board, and a target near an edge would otherwise
 * show background beside a window that is supposed to be full-screen.
 */
const camFor = (focus: string, activeLineY: number, terminalTop: number, terminalOpen: boolean): Cam => {
  const clamp = (c: Cam): Cam => {
    const halfW = BOARD.w / (2 * c.zoom);
    const halfH = BOARD.h / (2 * c.zoom);
    return {
      zoom: c.zoom,
      x: Math.min(Math.max(c.x, halfW), BOARD.w - halfW),
      y: Math.min(Math.max(c.y, halfH), BOARD.h - halfH),
    };
  };
  switch (focus) {
    case 'tree':
      return clamp({x: ACTIVITY_W + SIDEBAR_W / 2, y: BOARD.h * 0.4, zoom: CAM.tree});
    case 'tabs':
      return clamp({x: PANE.x + PANE.w * 0.32, y: TITLE_H + TAB_BAR_H / 2, zoom: CAM.tabs});
    case 'terminal':
      // Frames the code *and* the terminal together. The payoff of this
      // template is the two of them agreeing, and a shot of the output alone
      // is a screenshot of some text.
      return clamp({x: PANE.x + PANE.w * 0.44, y: terminalTop - 40, zoom: CAM.terminal});
    case 'code':
      // Follows the line being written, which is what makes the zoom read as
      // attention rather than as a crop.
      return clamp({x: PANE.x + PANE.w * 0.42, y: activeLineY, zoom: CAM.code});
    default:
      // `wide` sits slightly high when the terminal is open, so the drawer
      // does not push the code out of the top of frame.
      return {x: BOARD.w / 2, y: BOARD.h / 2 - (terminalOpen ? 40 : 0), zoom: CAM.wide};
  }
};

const lerpCam = (a: Cam, b: Cam, t: number): Cam => ({
  x: a.x + (b.x - a.x) * t,
  y: a.y + (b.y - a.y) * t,
  zoom: a.zoom + (b.zoom - a.zoom) * t,
});

/** Smoothstep — a camera that starts and stops abruptly reads as a glitch. */
const smooth = (p: number): number => {
  const c = Math.max(0, Math.min(1, p));
  return c * c * (3 - 2 * c);
};

export const WorkspaceScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const frame = useCurrentFrame();
  const nowMs = sceneStartMs + (frame / FPS) * 1000;
  const ui: Chrome = CHROME[theme.mode];

  const projectName = String(props.project ?? 'project');
  const command = String(props.command ?? '');
  const output = String(props.output ?? '');
  const files = (Array.isArray(props.files) ? props.files : []) as ProjectFile[];
  const steps = (Array.isArray(props.steps) ? props.steps : []) as Step[];

  const [tokens, setTokens] = useState<Record<string, TokenLine[]> | null>(null);
  const [handle] = useState(() => delayRender('workspace-highlight'));

  useEffect(() => {
    let cancelled = false;
    Promise.all(
      files.map((f) =>
        codeToTokens(f.code, {lang: 'python', theme: ui.shiki})
          .then((r) => [f.path, r.tokens] as const)
          .catch(
            () =>
              [
                f.path,
                f.code.split('\n').map<TokenLine>((ln) => [{content: ln, color: ui.token, offset: 0}]),
              ] as const,
          ),
      ),
    ).then((all) => {
      if (!cancelled) {
        setTokens(Object.fromEntries(all));
        continueRender(handle);
      }
    });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [handle, ui.shiki]);

  const stepIdx = useMemo(() => {
    let idx = 0;
    steps.forEach((s, i) => {
      if (nowMs >= s.startMs) {
        idx = i;
      }
    });
    return idx;
  }, [steps, nowMs]);

  if (!tokens || steps.length === 0 || files.length === 0) {
    return null;
  }

  const step = steps[stepIdx];
  const prev = steps[Math.max(0, stepIdx - 1)];
  const activePath = step.file;
  const activeTokens = tokens[activePath] ?? [];

  // How much of the active file exists. `through` counts lines; 0 is the whole
  // file, which is what a finished file and an untyped one both want.
  const shownLines = step.through > 0 ? Math.min(step.through, activeTokens.length) : activeTokens.length;
  // Within the step, the last revealed line types itself in rather than
  // appearing whole — the one detail that makes this read as somebody working
  // instead of a slideshow of diffs.
  const stepFrames = Math.max(1, ((step.endMs - step.startMs) / 1000) * FPS);
  const intoStep = frame - ((step.startMs - sceneStartMs) / 1000) * FPS;
  const grew = shownLines > (prev.through > 0 ? prev.through : shownLines) || stepIdx === 0;
  const typeP = grew
    ? interpolate(intoStep, [0, stepFrames * 0.55], [0, 1], {
        extrapolateLeft: 'clamp',
        extrapolateRight: 'clamp',
      })
    : 1;
  const fromLine = stepIdx === 0 || prev.file !== activePath ? 0 : Math.min(prev.through || shownLines, shownLines);
  const typedLines = Math.round(fromLine + (shownLines - fromLine) * typeP);
  const activeLine = Math.max(0, typedLines - 1);

  // --- terminal -------------------------------------------------------------
  const runStep = steps.find((s) => s.run);
  const runFrom = runStep ? ((runStep.startMs - sceneStartMs) / 1000) * FPS : Infinity;
  const drawer = interpolate(frame - runFrom, [0, 16], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  // The drawer takes whatever the code area left, rather than a fixed height.
  // At a fixed 300 a four-line program left a band of empty editor under the
  // terminal that the camera could not frame its way out of — and empty
  // *terminal* reads as a terminal, where empty editor reads as a mistake.
  const openTerminalH = Math.max(TERMINAL_H, PANE.h - codeAreaH(files));
  const terminalH = openTerminalH * drawer;
  const cmdChars = Math.max(0, Math.floor((frame - runFrom - 10) / 1.4));
  const typedCommand = command.slice(0, cmdChars);
  const cmdDone = cmdChars >= command.length;
  const outLines = output ? output.split('\n') : [];
  const shownOut = cmdDone
    ? Math.max(0, Math.floor((frame - runFrom - 10 - command.length * 1.4 - 8) / 2.5))
    : 0;

  // --- scrolling ------------------------------------------------------------
  // A file longer than the pane scrolls to keep the line being written in
  // frame, held a few lines off the bottom the way an editor does. Without it
  // the camera dutifully zooms to a line that is not on screen.
  const codeH = codeAreaH(files);
  const perScreen = visibleLines(codeH);
  const scroll = Math.max(
    0,
    Math.min(activeLine - (perScreen - 3), activeTokens.length - perScreen),
  );

  // --- camera ---------------------------------------------------------------
  const terminalTop = PANE.y + codeH;
  const lineY = PANE.y + EDITOR_PAD_TOP + (activeLine - scroll) * LINE_H + LINE_H / 2;
  const target = camFor(step.focus, lineY, terminalTop, drawer > 0.4);
  const prevTarget = camFor(prev.focus, lineY, terminalTop, drawer > 0.4);
  const camP = smooth(interpolate(intoStep, [0, CAM.frames], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  }));
  const cam = stepIdx === 0 ? target : lerpCam(prevTarget, target, camP);
  // Move the world under a fixed lens: scale about the origin, then translate
  // the focus point to the middle of the frame.
  const camTransform = `translate(${FRAME_W / 2 - cam.x * cam.zoom}px, ${FRAME_H / 2 - cam.y * cam.zoom}px) scale(${cam.zoom})`;


  const openTabs = files.map((f) => f.path);

  return (
    <div style={{position: 'absolute', inset: 0, overflow: 'hidden', backgroundColor: ui.chrome}}>
      <div
        style={{
          position: 'absolute',
          width: BOARD.w,
          height: BOARD.h,
          transform: camTransform,
          transformOrigin: '0 0',
          fontFamily: theme.fontMono,
          backgroundColor: ui.bg,
        }}
      >
        {/* Title bar */}
        <div
          style={{
            height: TITLE_H,
            display: 'flex',
            alignItems: 'center',
            padding: '0 18px',
            backgroundColor: ui.chrome,
            borderBottom: `1px solid ${ui.border}`,
          }}
        >
          {['#ff5f57', '#febc2e', '#28c840'].map((c) => (
            <div key={c} style={{width: 13, height: 13, borderRadius: 7, backgroundColor: c, marginRight: 9}} />
          ))}
          <div
            style={{
              flex: 1,
              textAlign: 'center',
              fontFamily: theme.fontBody,
              fontSize: 18,
              color: ui.dim,
            }}
          >
            {activePath} — {projectName}
          </div>
          <div style={{width: 60}} />
        </div>

        <div style={{display: 'flex', height: BOARD.h - TITLE_H - STATUS_H}}>
          {/* Activity bar */}
          <div
            style={{
              width: ACTIVITY_W,
              backgroundColor: ui.chrome,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              paddingTop: 18,
              gap: 26,
            }}
          >
            {[Files, Search, GitBranch, Bug].map((Icon, i) => (
              <div
                key={i}
                style={{
                  width: '100%',
                  display: 'flex',
                  justifyContent: 'center',
                  borderLeft: `2px solid ${i === 0 ? theme.primary : 'transparent'}`,
                }}
              >
                <Icon size={26} color={i === 0 ? ui.text : ui.faint} strokeWidth={1.8} />
              </div>
            ))}
          </div>

          {/* File tree — every file in the project, the open one lit */}
          <div
            style={{
              width: SIDEBAR_W,
              backgroundColor: ui.sidebar,
              borderRight: `1px solid ${ui.border}`,
              paddingTop: 14,
              fontFamily: theme.fontBody,
              fontSize: 20,
            }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 9,
                padding: '6px 18px',
                color: ui.dim,
                fontWeight: 600,
                textTransform: 'uppercase',
                fontSize: 15,
                letterSpacing: 1.4,
              }}
            >
              <FolderOpen size={18} color={theme.primary} strokeWidth={2} />
              {projectName}
            </div>
            {files.map((f) => {
              const on = f.path === activePath;
              return (
                <div
                  key={f.path}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 9,
                    padding: '7px 18px 7px 36px',
                    color: on ? ui.text : ui.dim,
                    backgroundColor: on ? `${theme.primary}26` : 'transparent',
                    borderLeft: `2px solid ${on ? theme.primary : 'transparent'}`,
                  }}
                >
                  <FileCode2 size={18} color={on ? ui.fileIcon : ui.faint} strokeWidth={1.8} />
                  {f.path}
                </div>
              );
            })}
          </div>

          {/* Editor column */}
          <div style={{flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0}}>
            {/* Tabs — one per file, so switching files is visible */}
            <div style={{display: 'flex', height: TAB_BAR_H, backgroundColor: ui.chrome, flexShrink: 0}}>
              {openTabs.map((pth) => {
                const on = pth === activePath;
                return (
                  <div
                    key={pth}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 9,
                      padding: '0 22px',
                      backgroundColor: on ? ui.bg : 'transparent',
                      borderTop: `2px solid ${on ? theme.primary : 'transparent'}`,
                      borderRight: `1px solid ${ui.border}`,
                      color: on ? ui.text : ui.dim,
                      fontFamily: theme.fontBody,
                      fontSize: 19,
                    }}
                  >
                    <FileCode2 size={17} color={on ? ui.fileIcon : ui.faint} strokeWidth={1.8} />
                    {pth}
                  </div>
                );
              })}
            </div>

            {/* Code */}
            <div style={{display: 'flex', height: codeH, overflow: 'hidden', paddingTop: EDITOR_PAD_TOP, flexShrink: 0}}>
              <div style={{flex: 1, minWidth: 0, transform: `translateY(${-scroll * LINE_H}px)`}}>
                {activeTokens.slice(0, typedLines).map((line, i) => {
                  const isActive = i === activeLine;
                  return (
                    <div
                      key={i}
                      style={{
                        display: 'flex',
                        height: LINE_H,
                        backgroundColor: isActive ? ui.activeLine : 'transparent',
                      }}
                    >
                      <span
                        style={{
                          width: GUTTER_W,
                          flexShrink: 0,
                          textAlign: 'right',
                          paddingRight: 26,
                          fontSize: 20,
                          lineHeight: `${LINE_H}px`,
                          color: isActive ? ui.dim : ui.faint,
                        }}
                      >
                        {i + 1}
                      </span>
                      <span style={{fontSize: FONT_SIZE, lineHeight: `${LINE_H}px`, whiteSpace: 'pre'}}>
                        {line.map((t, ti) => (
                          <span key={ti} style={{color: t.color ?? ui.token}}>
                            {t.content}
                          </span>
                        ))}
                        {isActive && typeP < 1 && (
                          <span
                            style={{
                              display: 'inline-block',
                              width: 3,
                              height: 27,
                              marginLeft: 2,
                              verticalAlign: 'text-bottom',
                              backgroundColor: frame % 16 < 9 ? theme.accent : 'transparent',
                            }}
                          />
                        )}
                      </span>
                    </div>
                  );
                })}
              </div>
              {/* Minimap, same miniature-of-the-tokens idea as the window scene */}
              <div
                style={{
                  width: MINIMAP_W,
                  flexShrink: 0,
                  backgroundColor: ui.minimap,
                  borderLeft: `1px solid ${ui.border}`,
                  padding: '4px 10px 0',
                }}
              >
                {activeTokens.slice(0, typedLines).map((line, i) => (
                  <div key={i} style={{display: 'flex', gap: 1, height: 4, marginBottom: 3}}>
                    {line.map((t, ti) => (
                      <div
                        key={ti}
                        style={{
                          width: t.content.length * 1.5,
                          backgroundColor: t.content.trim() === '' ? 'transparent' : (t.color ?? ui.token),
                          opacity: t.content.trim() === '' ? 0 : 0.7,
                          borderRadius: 1,
                        }}
                      />
                    ))}
                  </div>
                ))}
              </div>
            </div>

            {/* Terminal */}
            <div
              style={{
                height: terminalH,
                overflow: 'hidden',
                backgroundColor: ui.panel,
                borderTop: `1px solid ${drawer > 0.05 ? ui.border : 'transparent'}`,
                flexShrink: 0,
              }}
            >
              <div style={{padding: '0 26px 14px', height: openTerminalH, display: 'flex', flexDirection: 'column'}}>
                <div
                  style={{
                    display: 'flex',
                    gap: 28,
                    height: 46,
                    alignItems: 'center',
                    fontFamily: theme.fontBody,
                    fontSize: 15,
                    letterSpacing: 1.8,
                    fontWeight: 600,
                    textTransform: 'uppercase',
                    flexShrink: 0,
                  }}
                >
                  {['Problems', 'Output', 'Terminal'].map((tab) => {
                    const on = tab === 'Terminal';
                    return (
                      <span
                        key={tab}
                        style={{
                          color: on ? ui.text : ui.faint,
                          borderBottom: `2px solid ${on ? theme.primary : 'transparent'}`,
                          paddingBottom: 6,
                        }}
                      >
                        {tab}
                      </span>
                    );
                  })}
                </div>
                <div style={{fontSize: 23, lineHeight: 1.45, color: ui.terminalText, paddingTop: 8}}>
                  <div style={{display: 'flex', alignItems: 'baseline', gap: 10}}>
                    <ChevronRight size={19} color={ui.prompt} strokeWidth={3} style={{flexShrink: 0}} />
                    <span style={{whiteSpace: 'pre'}}>
                      {typedCommand}
                      {!cmdDone && frame > runFrom + 10 && (
                        <span
                          style={{
                            display: 'inline-block',
                            width: 11,
                            height: 22,
                            verticalAlign: 'text-bottom',
                            backgroundColor: frame % 16 < 9 ? theme.accent : 'transparent',
                          }}
                        />
                      )}
                    </span>
                  </div>
                  {outLines.slice(0, shownOut).map((ln, i) => (
                    <div key={i} style={{whiteSpace: 'pre-wrap'}}>
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
            height: STATUS_H,
            display: 'flex',
            alignItems: 'center',
            gap: 26,
            padding: '0 22px',
            backgroundColor: theme.primary,
            color: 'rgba(255,255,255,0.92)',
            fontFamily: theme.fontBody,
            fontSize: 18,
          }}
        >
          <span style={{display: 'flex', alignItems: 'center', gap: 7}}>
            <GitBranch size={16} strokeWidth={2.2} /> main
          </span>
          <span style={{marginLeft: 'auto'}}>Ln {activeLine + 1}</span>
          <span>Spaces: 4</span>
          <span>UTF-8</span>
          <span>Python</span>
        </div>
      </div>

      {/* The caption rides the *frame*, not the world — a line of type that
          zoomed with the camera would be unreadable on the beats that matter
          most, which are exactly the zoomed ones. */}
      {step.caption && (
        <div
          style={{
            position: 'absolute',
            left: 0,
            right: 0,
            bottom: 54,
            textAlign: 'center',
            fontFamily: theme.fontBody,
            fontSize: 32,
            fontWeight: 600,
            color: '#ffffff',
            textShadow: '0 2px 18px rgba(0,0,0,0.75)',
          }}
        >
          <span style={{background: 'rgba(10,16,26,0.72)', padding: '10px 26px', borderRadius: 12}}>
            {step.caption}
          </span>
        </div>
      )}
    </div>
  );
};
