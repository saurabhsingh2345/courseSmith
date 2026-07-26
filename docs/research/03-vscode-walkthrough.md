# Research: VS Code walkthrough scenes (FREE/OSS, 2025–2026)

_Web-research report, 2026-07-18. Evaluated against: Go emits JSON scene graph → React
renders in Remotion (frame-deterministic, no wall-clock)._

## Synthesized editor UIs in React

- **Code Hike v1 (`codehike`, MIT)** — the de-facto standard. `highlight()` (Shiki-based)
  → `HighlightedCode` tokens + annotations from comments (`// !mark(1:5)`); annotations map
  to YOUR React components (marks, focus, diff, tooltips, line numbers).
  **Killer feature:** `getStartingSnapshot()` + `calculateTransitions()` return **pure
  data** (per-token move/enter/leave with from/to x,y + color) that you interpolate with
  `useCurrentFrame()` — fully frame-deterministic; exactly what the official Remotion
  template does. RSC not required: precompute `HighlightedCode` and pass as prop.
  Integration effort: LOW.
- **shiki-magic-move** (MIT): FLIP + CSS transitions = wall-clock as shipped; `Precompiled`
  variant + core diff machine could be re-rendered frame-driven, but that rebuilds what
  Code Hike already provides as data. Only if its diff quality wins.
- **Monaco**: async workers, virtualized viewport, wall-clock cursor/scroll — bad fit for
  deterministic frames. Skip. **CodeMirror 6**: live-editor scheduling, no token motion —
  skip. **react-live**: irrelevant.

## Purpose-built tools

- **remotion-dev/template-code-hike — verified, official** (~215★). Steps as files →
  `<Sequence>` per step, token transitions interpolated per frame. Primary building block.
- **CodeVideo** (MIT, ~43 repos, tiny adoption): JSON **action stream**
  (`file-explorer-create-file`, `editor-type`, `terminal-run`, mouse moves) drives a
  virtual IDE (explorer + editor + terminal) as a reducer. Don't take the dependency —
  **steal the action-schema design** for the Go scene graph.
- Motion Canvas `Code` node: great primitives but unmaintained (community fork "Canvas
  Commons"), separate non-React ecosystem. Skip.
- Terminal recorders: keep VHS (MIT). asciinema is GPL + wall-clock player (its .cast JSON
  is a usable event source though).

## Chrome kits

- **No mature "VS Code chrome" React kit exists** — build it: title bar, activity bar,
  file tree, tabs, breadcrumbs, minimap (scaled render of same tokens), status bar.
  ~1–2 days, full scene-graph + motion-token control.
- **File icons: `material-icon-theme` npm (MIT)** — the actual VS Code Material Icon Theme
  SVGs + `generateManifest()` filename→icon mapping. Ideal for a JSON-driven file tree.
- **Theme: Shiki with real VS Code theme JSON** (`dark-plus` bundled) — token-for-token
  pixel-plausible VS Code.
- Caution: Code - OSS is MIT but Microsoft branding/logo is not — "VS Code-like" chrome is
  fine, don't reproduce the logo.

## Driving REAL VS Code (secondary tier only)

- openvscode-server (MIT, cleanest base) / code-server (MIT). Microsoft's server +
  vscode.dev: proprietary, prohibited — do not use.
- **Demo Time** (estruyf/vscode-demo-time, MIT): scripted demo steps from `.demo/demo.json`
  — best way to drive a real VS Code per step.
- Playwright + openvscode-server headless: reproducible **per-step stills**, not smooth
  video (cursor blink/scroll are wall-clock; per-frame screenshots brutally slow).
- Playwright video: Chromium bitrate hardcoded 1 Mbit/s — visibly mediocre. ffmpeg
  x11grab+Xvfb: real-time, dropped frames, fragile. Both rejected for "world-class".

## Recommendation

**Primary: synthesized "FakeVSCode" scene kit** — custom VS Code chrome components +
Code Hike v1 token transitions, driven by the Go JSON scene graph:
Go writes `{files[], activeTab, cursor, steps[{code, annotations, durationFrames}]}`;
tokens computed via Shiki/Code Hike (runtime with delayRender, or precomputed); React
renders chrome + editor pane; typing = per-frame slice of token stream with frame-parity
block cursor; scrolling = interpolated translateY; minimap = same tokens at 10% scale;
icons from material-icon-theme; theme dark-plus.

**Secondary (optional realism B-roll): openvscode-server + Playwright + Demo Time-style
steps captured as per-step PNG stills**, composited in Remotion with Ken Burns moves.
