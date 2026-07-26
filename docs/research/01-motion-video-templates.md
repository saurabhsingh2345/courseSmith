# Research: FREE/OSS tooling for broadcast-quality Remotion course videos (2025–2026)

_Web-research report, 2026-07-18. Filtered against the hard Remotion rule: all animation
must derive from `useCurrentFrame()` — no rAF, no wall-clock, no unseeded random._

Integration classes used below:
- **Class A (native):** built on `useCurrentFrame()`/`interpolate()`/`spring()`. Zero risk.
- **Class B (seekable):** paused timeline + `seek(frame/fps)` per frame. Deterministic if used this way.
- **Class C (clock-driven):** internal ticker/rAF. Unusable in Remotion.

Licensing note: **Remotion itself is source-available, not OSS** — free for individuals and
companies ≤3 people (commercial included); 4+ person for-profit needs a Company License.

## 1. Official Remotion packages (version-lock all `@remotion/*` to the same exact version)

| Package | What | Value |
|---|---|---|
| @remotion/transitions | `<TransitionSeries>` + fade/slide/wipe/flip/clockWipe/iris, springTiming/linearTiming | THE composition primitive for a template system. Class A |
| @remotion/shapes | Parametric SVG shapes | Decorative geometry, progress pies. Class A |
| @remotion/animation-utils | `makeTransform()`, `interpolateStyles()` | Type-safe transforms. Class A |
| @remotion/layout-utils | `measureText()`, `fitText()`, `fillTextBox()` | Auto-fit LLM text — kills overflow bugs. Class A |
| @remotion/paths | `evolvePath`, `getPointAtLength`, `warpPath` | Draw-on diagram lines/arrows. Class A |
| @remotion/lottie | Frame-synced Lottie JSON | Pro micro-animations for free. Class A |
| @remotion/noise | Seeded simplex `noise2D/3D/4D` | The sanctioned "random": drift, sway, particles. Class A |
| @remotion/motion-blur | `<Trail>`, `<CameraMotionBlur>` | Cheapest "feels professional" upgrade. Class A |
| @remotion/google-fonts | Deterministic `loadFont()` | Typography token foundation |
| @remotion/skia | React Native Skia | GPU 2D effects; heavier setup. Class A |
| @remotion/three | `<ThreeCanvas>` (drive via useCurrentFrame, NOT R3F useFrame) | Shader/3D backgrounds; `uTime = frame/fps` uniform |
| @remotion/captions | `Caption` type + `createTikTokStyleCaptions()` | Word-timed caption pages + spring pop. Class A |
| @remotion/animated-emoji | Google animated emoji | Free production-value accents |

Official templates gallery: **Code Hike template** (animated code walkthroughs), TikTok
captions, Audiogram, R3F, Skia.

## 2. Community Remotion component libraries (template accelerators)

- **Remocn** (MIT, remocn.dev, `npx shadcn@latest add @remocn/<name>`): **110+ copy-paste
  components** — 30+ typography effects (typewriter, shimmer, matrix decode), transitions
  (grid pixelate, chromatic aberration, device mockups), environments (mesh gradients,
  grids, spotlights, confetti), **code editors, terminals, charts**, bento compositions.
  Code is vendored into the repo — no dep lock-in. Best single find.
- **remotion-scenes** (MIT): 201 scenes across 16 categories (text/shapes/transitions/
  backgrounds/particles/cinematic/layouts/lists/theme animations).
- **remotion-bits** (MIT): text reveals, gradient transitions, particles, charts; CLI + MCP.
- **Onda** (~70 components), **RemotionUI**, **remotion-transition-series** (superseded).
- **OpenMontage** (OSS agentic video system on Remotion) — closest analog to courseSmith; study.

## 3. Alternative frameworks

- **Motion Canvas** (MIT): generator-function animation; separate runtime, not React — skip.
- **Revideo** (MIT): Motion Canvas fork with headless "template + variables → render" API;
  maintenance risk (team pivoted to Midrender). Study its API shape only.
- **Theatre.js** (core Apache-2.0, studio AGPL dev-only): keyframe GUI → JSON, playback via
  `sequence.position = frame/fps` (Class B). Repo momentum uncertain.

## 4. Tweening engines

- **GSAP — verified 100% free since Apr 30 2025** (Webflow). SplitText, MorphSVG, DrawSVG,
  CustomEase etc. all free, commercial OK (source-available license). **Class B in Remotion**:
  build a **paused** timeline, `tl.seek(frame/fps)` every frame. Unlocks MorphSVG (diagram
  state morphs) and SplitText (kinetic type). Avoid ScrambleText (unseeded random).
- **anime.js v4** (MIT): `.seek(ms)` → Class B; fine secondary, less compelling now GSAP is free.
- **Framer Motion / Motion for React: Class C — NOT Remotion-compatible. Hard no.** Use `spring()`.
- **Rive**: MIT runtimes, `@remotion/rive` exists; authoring needs proprietary editor — optional.
- **Lottie**: lottie-web MIT; free sources = LottieFiles free tier, Lottielab, Lordicon free,
  **Glaxnimate** (fully OSS Lottie authoring).

## 5. Template-based video editing prior art (OSS)

Editly (MIT, unmaintained, JSON-spec reference), FFCreator (aging), OpenCut (45k★ CapCut
alt, interactive not headless), Motionity (browser editor), Revideo (real params→video).
**Takeaway: nothing OSS beats Remotion + Zod-schema'd scene components + TransitionSeries
as the template engine.** Pattern: scene template = React component + Zod schema; Go emits
JSON matching the schemas; master composition maps `scenes[]` into TransitionSeries.

## 6. Backgrounds & polish

- **@paper-design/shaders-react** (Apache-2.0, verified): ~28 WebGL2 shaders — MeshGradient,
  GrainGradient, Warp, Waves, NeuroNoise, Voronoi, Metaballs, GodRays. **Must be frame-locked**
  (speed=0 + frame-derived offset, or lift GLSL into @remotion/three with uTime=frame/fps).
- Raw GLSL via @remotion/three = gold standard deterministic backgrounds.
- **easy-mesh-gradient** (zero-dep, static CSS mesh) + slow @remotion/noise blob drift =
  zero-risk default. `whatamesh` needs freezing.
- **Grain:** SVG feTurbulence + feColorMatrix at 3–5% opacity (fixed seed = deterministic);
  masks H.264 banding in dark gradients. The cheapest broadcast-polish trick.

## 7. Free assets

- Icons: **Iconify** (MIT, 200k+ icons, one API), **Lucide** (stroke-based → pairs with
  evolvePath draw-on), **Phosphor** (7,700 × 6 weights), Heroicons.
- Illustrations: **unDraw** (free, no attribution, accent-color parameterizable),
  Storyset (attribution required), **3dicons.co (CC0 3D icons)**.
- Fonts (OFL via @remotion/google-fonts): **Sora + IBM Plex Sans (+ IBM Plex Mono)**;
  **Space Grotesk + Inter + JetBrains Mono** (de facto 2025 dev-content stack);
  Bitter/Roboto Slab + Source Sans 3 (editorial). Use fitText() at display sizes.
- Code/terminal: **Code Hike Remotion template** (Class A token-diff code animation — the
  state of the art); shiki-magic-move is Class C by default (CSS transitions) — prefer
  Code Hike inside Remotion. VHS (already used) for authentic terminal captures.

## 8. Recommended stack (condensed)

1. Template engine: Remotion + Zod scene schemas + @remotion/transitions; seed vocabulary
   from Remocn + remotion-scenes (both MIT, vendored).
2. Motion: useCurrentFrame + spring + animation-utils; GSAP via paused-seek bridge for
   SplitText/MorphSVG. Never Framer Motion.
3. Backgrounds: paper-shaders (frame-locked) or GLSL via @remotion/three; static mesh +
   feTurbulence grain as cheap default.
4. Polish: motion-blur on fast moves, noise drift, 3–5% grain.
5. Diagrams: paths.evolvePath draw-on + MorphSVG morphs.
6. Code: Code Hike template pattern. 7. Captions: @remotion/captions + spring pop.
8. Assets: Iconify/Lucide/Phosphor, unDraw, 3dicons, Lottie (free tiers/Glaxnimate);
   fonts Space Grotesk/Inter/JetBrains Mono or Sora/IBM Plex.

Risks: Framer Motion (hard no) · paper-shaders frame-lock needs a spike · Revideo/Theatre
maintenance drift · Remotion company license for 4+ person teams · Storyset attribution.
