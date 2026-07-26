# courseSmith 10x plan — world-class course videos

_2026-07-18. Driven by: "bad course, not good looking, visuals not aware of the context,
want VS Code walkthrough, template format, editing templates into a bigger video, voice in
my control, better captions — everything 10x better."_

Research backing every tool choice: `docs/research/01-motion-video-templates.md`,
`02-captions-context-visuals.md`, `03-vscode-walkthrough.md`, `04-tts-voice-alignment.md`.

---

## The audit (what's actually wrong, verified from rendered frames of lesson 01)

1. **Dead screens dominate.** Sections without a declared `[DIAGRAM]`/`[DEMO]` render as a
   bare heading + subtitle on a white void for their whole duration (frames at 60s, 120s).
   The script's cue vocabulary is only `diagram|demo|pause`, and diagrams must be
   pre-declared by the human author — the AI cannot add visuals per narration beat.
2. **No design language.** System fonts, flat white background, no gradients/texture/depth,
   hardcoded grays, 3-color theme, plain `<h1>` title cards.
3. **Visuals aren't context-aware by construction.** `diagram_svg.tmpl` receives only the
   lesson TITLE + a one-line prompt — never the narration or section content.
4. **Code scenes waste the canvas** — no window chrome, no line numbers, light shiki theme
   in a dark box, huge empty panels.
5. **Captions** are one static dark pill, color-only word highlight.
6. **Voice control** = a single course-wide Kokoro voice id. No speed, no per-lesson
   override surfaced, no pronunciation/stress usage, pace report is informational only.
7. **Transitions**: one uniform 500ms fade. Renderer has zero animation/font packages.
8. **No VS Code walkthrough scene. No editable template/composition layer.**

## Design principles

- **Go owns tokens** (established architecture): the extended video theme (fonts, surfaces,
  caption style, transition style) lives in Go, flows through `lesson-video.json`, renderer
  consumes. Extend, don't fork.
- **Frame-determinism is law**: everything derives from `useCurrentFrame()`; seeded
  @remotion/noise is the only "random". Never Framer Motion (Class C — verified
  incompatible).
- **Dark, cohesive, editorial look**: Space Grotesk (display) + Inter (body) + JetBrains
  Mono (code) via @remotion/google-fonts; layered gradient background + 4% grain;
  glass/surface cards; one accent per course.
- **Every section gets a visual.** No scene may show only a heading.

---

## Workstreams

### W1 — Design system & renderer visual overhaul (the "not good looking" fix)
Deps: `@remotion/google-fonts`, `@remotion/transitions`, `@remotion/layout-utils`,
`@remotion/paths`, `@remotion/noise`, `@remotion/shapes` (all version-locked to remotion
4.0.x installed version).
- Extend Go `SceneTheme` → rich `VideoTheme`: mode (dark default), bg gradient stops,
  surface/border/muted colors, font trio, grain amount; TS mirror in types.ts (drift-guard
  test like motion tokens).
- `SceneBackground`: theme gradient + two noise-drifting accent blobs + feTurbulence grain
  overlay (fixed seed), used by every scene.
- Rebuild `TitleCard`: display type with fitText, accent geometry (shapes/paths draw-on),
  staggered outcome rows with icon chips, course kicker, bottom progress strip.
- Rebuild `CodeScene`: editor window chrome (traffic lights, filename tab, line numbers),
  Shiki `dark-plus`, auto-sized panel, output as attached console drawer with success tick.
- `SectionTransition v2`: per-archetype transition style tokens (fade/slide/wipe scale)
  from Motion tokens.
- Heading-only cards get kinetic typography treatment (per-word spring reveal) — and W2
  makes them rare.
Verify: `remotion still` set + visual_regression `--update` after eyeball pass; tsc; go test.

### W2 — Context-aware visuals & no dead screens
- **Storyboard pass** (new stage `storyboard`, LLM, cached like all stages): input = script
  sections + narration; output = per-section visual beats: `points` scenes (3-5 keyword
  bullets with Lucide icon names + at_word timings), stat/analogy callouts, and
  auto-suggested diagrams for sections that have none. Hard gate: every section ends with
  ≥1 non-heading visual.
- New scene type `points`: bullets reveal on the exact spoken word (existing alignment),
  icon chips, kinetic emphasis. Renders instead of bare heading cards.
- **Context into diagram prompts**: `diagram_*.tmpl` gain the section narration + adjacent
  section titles + audience/tone so the picture matches what's being said.
- **D2 diagram kind** (`kind: d2`, MPL-2.0 pure-Go library — imported directly into the
  pipeline binary, no binary install): LLM emits D2 source; compile = validation (self-
  repair loop like mermaid); sketch mode + dark theme matches design system; same
  vision-QA gate. This is the "good-looking diagrams" workhorse.
- Icon assets: vendor Lucide SVGs (stroke-based → draw-on via @remotion/paths).
Verify: run storyboard+visuals on lesson 01; stills show zero heading-only scenes.

### W3 — Captions 10x
- Page-based karaoke: 3–5 words/page (~1200ms combine window), active word spring scale
  1.1 + accent color, pill with backdrop blur, safe-area lower third.
- **Emphasis pass**: LLM marks ≤1 keyword per page (accent color even when inactive) —
  runs inside captions stage, cached, output `captions_emphasis.json`.
- Caption style tokens in theme (position, size, mode: `karaoke|pages|off`).
Verify: stills at word boundaries; captions.vtt unchanged (web path unaffected).

### W4 — VS Code walkthrough scene
- New scene type `walkthrough` + `VSCodeScene` renderer component: synthesized VS Code
  chrome (title bar, activity bar, file tree w/ material-icon-theme SVGs, tab bar, editor,
  status bar, minimap), Shiki `dark-plus` tokens, per-step code evolution (typing +
  line-highlight steps interpolated by frame), cursor with frame-parity blink.
- Go side: lesson front-matter `walkthrough:` block (files + steps referencing section
  code blocks) or auto-derived: multi-block sections become a walkthrough (block N = step
  N). Steps timed to narration via at_word cues from the storyboard pass.
- Secondary tier (deferred): openvscode-server + Playwright per-step stills for realism
  b-roll.
Verify: dedicated demo composition + stills at typing/highlight/step-transition frames.

### W5 — Voice in the user's control
- `style.voice_speed` (0.5–2.0) → Kokoro `speed` param.
- **Auto-pace**: close the pace-report loop — when a section's measured wpm misses
  `pace_wpm` ±15%, compute per-lesson speed correction (bounded) on re-run.
- Voice blending passes through (`af_bella(2)+af_sky(1)` is a valid voice string —
  document + validate against /v1/audio/voices).
- Per-lesson voice/speed via front-matter (config layering already supports style).
- Generic `TTS_URL` (alias of KOKORO_URL) + docs for **Chatterbox Turbo** (MIT, MPS,
  OpenAI-compatible servers) as the cloned-voice engine: user records 10–20s, points
  courseSmith at chatterbox-tts-api, keeps whisperX alignment (Kokoro-native timestamps
  don't apply to other engines).
- `audition` already picks voices; extend to preview blends & speeds.
Verify: unit tests + one section re-synth at 1.15x with WER check.

### W6 — Template format & composing bigger videos
- **Template registry**: every scene carries `template` (e.g. `title.hero`,
  `points.iconGrid`, `code.editor`, `diagram.build`); registry maps template+archetype →
  component variant + motion overrides. Archetypes (existing F workstream) select default
  template sets — this is the "template format".
- **`video-plan.yaml` override layer** (like quiz overrides): per-lesson editable file to
  reorder scenes, swap a scene's template, pin durations, insert custom scenes (title
  cards, image boards) — merged by the scenegraph stage. This is "editing the template".
- **`coursesmith compile-course`**: stitches lesson finals + generated chapter interstitial
  cards into one long course video (ffmpeg concat + rendered interstitials) — "a bigger
  video".
Verify: plan file round-trip test; compile a 2-lesson course video.

## Execution order
W1 → W3 → W2 → W4 → W6 → W5 (visual foundation first; captions ride on it; storyboard
depends on new scene types; walkthrough on the chrome/design system; templates formalize
what exists; voice is independent and last).

## Deferred (documented, not in this pass)
- FLUX.2-klein-4B illustrations via headless ComfyUI (needs model download + ComfyUI
  service; big win for hero images later).
- Qwen2.5-VL-7B local vision judge (replace gpt-4o-mini vision).
- ctc-forced-aligner / Kokoro captioned_speech alignment swap.
- resemble-enhance mastering insert; Pexels b-roll stage; Remotion license check for
  Enfec team size; openvscode-server realism tier.
