// The editor's palette, shared by the two code templates.
//
// Extracted from VSCodeScene when the workspace template arrived: two scenes
// drawing the same editor in two slightly different greys is the kind of drift
// nobody notices until they are side by side in one course.

/**
 * The editor's own palette.
 *
 * Deliberately *not* the design system's tokens. An editor painted in `surface`
 * and `mass` is a courseSmith panel with code in it, and the whole point of
 * this template is that it reads as the tool the viewer actually uses — the
 * credibility comes from it looking like the thing on their second monitor.
 *
 * What it does take from the theme is the mode, the primary (status bar) and
 * the accent (the active marker and the caret). A dark editor on a paper stage
 * is a hole punched in the page, and that is exactly what this scene used to be
 * in light mode, because every value below was a hard-coded dark literal.
 */
export type Chrome = {
  /** The editor surface itself. */
  bg: string;
  /** Title bar, tab bar and activity bar. */
  chrome: string;
  sidebar: string;
  /** The terminal drawer. */
  panel: string;
  /** Hairlines between regions. */
  border: string;
  /** The window's outer edge. */
  outline: string;
  shadow: string;
  /** Chrome text at full strength — the active tab, the selected file. */
  text: string;
  /** Inactive chrome text. */
  dim: string;
  /** Line numbers and inactive icons. */
  faint: string;
  /** The row the cursor is on. */
  activeLine: string;
  indent: string;
  /** The minimap's own column, so it reads as part of the editor. */
  minimap: string;
  /** The colour code takes when highlighting has not resolved. */
  token: string;
  /** The file-type icon in the tree and on the tab. */
  fileIcon: string;
  /** Terminal body text. */
  terminalText: string;
  /** The shell prompt caret. */
  prompt: string;
  /** Which Shiki theme the code is tokenised with. */
  shiki: 'dark-plus' | 'light-plus';
};

export const CHROME: Record<'dark' | 'light', Chrome> = {
  dark: {
    bg: '#1a1e26',
    chrome: '#12151b',
    sidebar: '#161a21',
    panel: '#101319',
    border: 'rgba(255,255,255,0.07)',
    outline: 'rgba(255,255,255,0.10)',
    // Deep and soft. On the dark stage a shadow is nearly invisible anyway —
    // what actually separates the window from the background here is the
    // outline, which is why there is one.
    shadow: '0 44px 110px rgba(0,0,0,0.62)',
    text: '#e8ecf3',
    dim: 'rgba(255,255,255,0.42)',
    faint: 'rgba(255,255,255,0.26)',
    activeLine: 'rgba(255,255,255,0.045)',
    indent: 'rgba(255,255,255,0.09)',
    minimap: 'rgba(255,255,255,0.025)',
    token: '#d4d4d8',
    fileIcon: '#4fc1ff',
    terminalText: '#d6dde6',
    prompt: '#3ddc84',
    shiki: 'dark-plus',
  },
  light: {
    // Paper white against the stage's tinted near-white. The window has to be
    // *lighter* than the page it sits on, or it reads as a dent rather than a
    // panel — which is why the stage is tinted at 98.5% rather than pure white
    // in the first place.
    bg: '#ffffff',
    chrome: '#eceef2',
    sidebar: '#f3f4f7',
    // Darker than the editor, not lighter. At #f8f9fb the drawer and the
    // editor were one continuous white area and the terminal read as more
    // file rather than as a panel that had opened under it.
    panel: '#eef0f4',
    border: 'rgba(16,24,40,0.09)',
    outline: 'rgba(16,24,40,0.14)',
    // On paper the shadow is the only thing holding the window off the page,
    // so it is doing real work here and is tuned to be seen.
    shadow: '0 28px 70px rgba(16,24,40,0.16), 0 4px 14px rgba(16,24,40,0.08)',
    text: '#1f2430',
    dim: 'rgba(31,36,48,0.58)',
    faint: 'rgba(31,36,48,0.36)',
    activeLine: 'rgba(31,36,48,0.045)',
    indent: 'rgba(31,36,48,0.12)',
    minimap: 'rgba(31,36,48,0.035)',
    token: '#24292f',
    fileIcon: '#2c7ad6',
    terminalText: '#232a35',
    // The dark stage's terminal green is near-invisible on a white panel.
    prompt: '#1a8f4c',
    shiki: 'light-plus',
  },
};
