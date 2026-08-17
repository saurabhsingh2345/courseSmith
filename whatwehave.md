# What We Have — courseSmith

_A living snapshot of the project: what it is, everything that's built, what's
left, and where it's going. Last updated 2026-07-31 — the snippet catalog is
**32 templates** in six browsable categories (§9–§11, §14), there is a house-style
axis (**skins**) independent of branding and mode (§10), **combos** cut one
video from several templates on a single timeline (§12), and the studio has a
nav rail and a working light mode (§13). Content generation is **OpenAI-only**
by default._

---

## 1. What this is

**courseSmith is an AI course-production engine.** You point it at a markdown
lesson outline and it compiles a broadcast-quality video lesson with **zero
manual recording** — animated title cards, self-typing code scenes, real
executed terminal demos, diagrams that appear on the exact spoken word, karaoke
captions, an interactive quiz, a companion ebook, and a publishable website — at
near-zero cost using free-tier APIs and open-source tooling.

The whole thing is a Go CLI (`bin/coursesmith`) driving a 15-stage,
idempotent, resumable pipeline, plus a Node/Remotion render engine, a Hugo
site, and a React studio UI.

**Two shorter shapes sit on the same spine.** A **snippet** (§9–§11) is a
prompt plus one of 32 visual templates, planned into a standalone clip; a
**combo** (§12) is an ordered run of segments, each with its own template, cut
onto one timeline. Both reuse the whole pipeline below the plan stage, so they
inherit the same quality moat with no second engine.

```
lesson.md ─▶ script ─▶ verify ─▶ review ─▶ visuals ─▶ quiz ─▶ mistakes ─▶ exercises ─▶ demos
                │         │          │          │        │                                  │
             (Groq)   (docker    (3-pass    (SVG +    (MCQ,   (real tracebacks,       (VHS real
                       executes  accuracy/  vision    exec-    proven-solvable          terminal
                       all code) pedagogy/  QA)       checked) exercises)               sessions)
                                 tone)
    ─▶ audio ─▶ align ─▶ captions ─▶ chapters ─▶ scenegraph ─▶ render ─▶ hugo
        │         │         │           │            │            │        │
     (Kokoro   (whisperX  (WebVTT   (YouTube     (lesson-      (Remotion  (page
      TTS +    word       from      chapters +   video.json)   1080p30)   bundle)
      master-  timing +   align)    transcript)
      ing)     WER QA)
```

**The quality moat:** every Python code block is executed for real in a Docker
sandbox. The course never shows output Python didn't actually produce, broken
code fails the build, and terminal demos are genuine recordings of `python3`
running (faking output with `echo` is rejected at generation time).

**The cost moat:** every LLM response is cached on disk keyed by request hash.
Re-running a pipeline costs **zero API calls** unless something actually
changed. Rate limits are enforced per provider with state persisted across
restarts, so an interrupted run can never blow the daily budget.

---

## 2. The pipeline — all 15 stages

Every stage is idempotent: outputs land in the lesson's `generated/` directory,
input hashes are recorded in `state.json`, and unchanged stages are skipped on
re-run. Stages go stale automatically when their inputs — including prompt
templates — change.

| # | Stage | What it produces | Engine / gate |
|---|-------|------------------|---------------|
| 1 | **script** | `script.json` — narration with diagram/demo cues | Groq LLM |
| 2 | **verify** | `verification.json` — real sandbox output of every code block | Docker (hard gate: broken code fails the build) |
| 3 | **review** | `reviews/script-multipass-*.json` — 3-pass critique | LLM soft gate (accuracy 50% / pedagogy 35% / tone 15%) |
| 4 | **visuals** | `diagrams/*.svg` (+ `.mmd`/`.excalidraw.json` source, `d3` `.json`) + `attempts/` | svg/mermaid/excalidraw/d3 gen + headless-Chromium vision QA |
| 5 | **quiz** | `quiz.json` — recall/application/debugging/prediction MCQs | LLM + exec-checked answers + distractor/difficulty QA |
| 6 | **mistakes** | `mistakes.json` — top-3 beginner errors w/ real tracebacks | Docker (hard gate: code that doesn't fail is rejected) |
| 7 | **exercises** | `exercises/` — starter + hidden pytest + solution | Docker (hard gate: solution passes, starter fails) |
| 8 | **demos** | `demos/*.mp4` + `manifest.json` | VHS real terminal recordings (echo-faking rejected) |
| 9 | **audio** | `audio/*.wav`, `voiceover.wav` | Kokoro TTS + tts_prep + mastering chain |
| 10 | **align** | `alignment.json` — word-level timestamps + per-section WER | whisperX (falls back to Groq Whisper segments) |
| 11 | **captions** | `captions.vtt` — WebVTT grouped from the alignment | — |
| 12 | **chapters** | `chapters.json/.txt`, `transcript.md` | YouTube-format chapters |
| 13 | **scenegraph** | `lesson-video.json` — full render input | — |
| 14 | **render** | `final.mp4` (1920×1080, 30fps) | Remotion (falls back to ffmpeg slideshow) |
| 15 | **hugo** | Hugo page bundle in `site/content/` | video player + captions + quiz |

---

## 3. Feature inventory (built & working)

### Content generation
- **Outline → narration script** from a markdown lesson with YAML front-matter
  (title, outcomes, diagrams, callouts, per-lesson style overrides).
- **Three-pass script review** — accuracy (claims extracted, code-checkable
  ones executed, rest judged with citations), pedagogy (concept ordering,
  ~1 new concept per 90s, concrete-before-abstract), and tone. Weighted score
  gates against `review_threshold`; failing drafts regenerate with all three
  critiques injected (max 2 rounds).
- **Interactive quiz** with a required taxonomy (≥1 recall, application,
  debugging, prediction question each). Prediction answers checked against real
  execution. Distractors scored as misconceptions; a review model role-plays
  10 cold students per question to flag `too_easy` (>90% success) / `too_hard`
  (<30%). From lesson 4 on, quizzes include 1–2 spaced-repetition questions
  drawn from earlier lessons' concept graphs.
- **Common mistakes** — top-3 beginner errors, each with the *actual traceback*
  from running the broken code in the sandbox.
- **Practice exercises** — two per lesson (starter + hidden pytest tests +
  solution), solvability proven by execution.

### Visual & video
- **Multi-format diagrams**, one per declared `kind`, all gated by the same
  headless-Chromium vision QA (overlapping text, clipped elements, contrast,
  misleading layout — regenerates on issues, every attempt kept for audit):
  - `svg` — the model draws freehand SVG (three exemplar SVGs act as few-shot
    style anchors so a whole course shares one visual language).
  - `d3` — the model emits a structured node-link spec; the renderer lays it out
    and animates it (nodes pop in, edges draw on).
  - `mermaid` — the model writes Mermaid syntax; the visuals stage compiles it to
    SVG with embedded Mermaid.js in the headless browser.
  - `excalidraw` — the model emits a small element list; embedded Rough.js
    (Excalidraw's own engine) compiles it to a hand-drawn SVG.
  The `mermaid`/`excalidraw` kinds compile *source* to a self-contained SVG
  server-side, so they flow through the exact same vision-QA gate and inline-SVG
  renderer as `svg` — the model authors an easy-to-validate spec, and a compile
  failure is fed back to it as a correction. Both JS libraries are embedded in
  the binary (offline, version-pinned). Without a browser, these kinds degrade
  to a freehand SVG of the same request.
- **Remotion render engine** (`renderer/`): animated title card with staggered
  outcomes, self-typing code (jitter + Shiki highlighting), sequentially
  revealed diagram groups, VHS demos in a styled window, callout arrows/circles
  landing on the exact spoken word, karaoke captions. Components:
  `TitleCard`, `CodeScene`, `DiagramScene`, `TerminalScene`, `CalloutLayer`,
  `CaptionTrack`, `SectionTransition`.
- **VHS terminal demos** — real `python3` sessions, engine-owned dimensions /
  fonts / theme so the LLM only writes the commands; `vhs validate` failures
  get one retry.
- **ffmpeg fallback** slideshow render when Node isn't available.

### Audio
- **tts_prep** speech rewriting: code tokens → spoken forms (`__init__` →
  "dunder init"), numbers/operators → words, long sentences split at clauses,
  `*emphasis*` → micro-pauses. What was sent to TTS persists as
  `tts_script.json`.
- **Kokoro TTS** per-paragraph synthesis, joined with configurable pauses +
  50ms crossfades.
- **Mastering chain**: highpass 70Hz → de-esser → 2.5:1 compression → two-pass
  loudnorm to −16 LUFS / −1.5 dBTP; before/after stats in `reviews/loudness.json`.
- **Optional CC0 music bed**, sidechain-ducked under the voice.
- **Speech dictionary** (~40 built-in Python terms) extendable per-course via
  `style.pronunciations`.

### Sync & QA
- **whisperX word-level alignment** → diagrams/callouts/captions land on the
  exact spoken word. Compresses awkward silences (>1.5s) out of the voiceover.
- **Dual-reference WER** — scores each section against both written narration
  and the spoken tts_script (whisper inverse-normalizes speech, so either alone
  yields false misreads), with token expansion + bigram merge. Achieved 0.0%
  WER on lesson 01.
- **Auto-fix loop** — misreads the pronunciation dictionary knows get corrected
  via `tts_fixes.json`, which re-stales the audio stage.
- **Pace report** (`reviews/pace.json`) flags sections outside `pace_wpm` ±15%.

### Cross-course intelligence
- **`analyze`** — builds a concept graph (`concepts.json` + `.svg`), detects
  dependency-order violations (error), terminology drift, and suggests
  narrative bridges between lessons.

### Distribution
- **Hugo website** — each lesson as a page bundle (player + captions + inline
  diagrams + interactive quiz).
- **`ebook`** — one print-styled HTML document (cover, TOC, per-lesson
  transcripts + diagrams + quizzes, answer-key appendix) → PDF via headless
  Chromium.
- **`bundle`** — builds the static site and zips it (videos included) so a
  course runs from a USB stick over `file://` with no server.
- **GitHub Actions** `deploy-site.yml` workflow.

### Studio UI (`studio/`, React + Vite + Tailwind)
- Served by `coursesmith serve` (Go JSON API in `internal/studio/`).
- Pages: Compose, Snippets, Combos, Courses, Course/Lesson detail and editors,
  Quiz editor + strategy, Templates, Library, Results gallery, Adaptive
  config, Showcase, Generation, Ledger — behind a grouped, collapsible nav
  rail (§13).
- Live run control over SSE (`/api/events`, `/api/run` POST/DELETE), feedback
  and regenerate endpoints, quiz-override editor, artifact serving, OpenAPI
  schema → typed client (`schema.d.ts`).
- **Light and dark**, for real: the ink ramp inverts (§13), so one class name
  is correct in both themes.

### Adaptive learning (`coursesmith-tutor`, workstream D)
- **Adaptive-learning microservice** (`cmd/coursesmith-tutor`, `:8765`,
  opt-in, CORS-enabled): `/bkt/estimate` (Bayesian Knowledge Tracing mastery),
  `/fsrs/schedule` (spaced-repetition interval), `/irt/calibrate` (stub),
  `/health`. All math lives in `internal/adaptive` — one source of truth shared
  by the service, a typed Go `Client`, and an in-process `Mock`.
- **Offline fitters** (`tools/tutor/`, dependency-free Python): `fit_bkt.py`
  (max-likelihood BKT param fit from pooled response logs) and
  `calibrate_irt.py` (1PL difficulty + point-biserial discrimination), with a
  documented pyBKT/py-irt/FSRS upgrade path.
- **Studio wiring**: `useTutorAPI` hook + `AdaptiveQuizDemo` on the Showcase
  page — simulate quiz answers and watch live BKT mastery ("You've mastered
  N%"), the difficulty recommendation, and the FSRS next-review date update.
  No student data is persisted yet (interfaces + real math, mock-free).

### Human-in-the-loop
- **`export-review`** — one markdown doc per lesson (script, diagrams, quiz w/
  answers, mistakes w/ tracebacks, exercises, every QA flag) for an SME.
- **`review-notes.yaml`** loop — reviewer notes get injected into the next
  script generation and marked resolved.
- **`audition`** — renders a sample paragraph in every matching Kokoro voice +
  an HTML player page; `--choose` writes the pick to `course.yaml`.

### Platform
- **LLM layer** (`internal/llm/`): provider router (any OpenAI-compatible
  backend; `openai/gpt-4o-mini` is the content default and Groq is no longer
  wired in by default), disk cache, per-provider token-bucket rate limiter w/
  persisted state, quota-aware clean stops, retry logic, transcription.
  `completeWithRepairRounds` is the shared correction loop every template's
  validator feeds.
- **Layered config** (`internal/config/`): defaults < course.yaml < lesson
  front-matter < CLI flags.
- **Prompts as templates** (`prompts/*.tmpl`) — editable, and editing one
  auto-stales dependent stages.
- **`doctor`** — checks ffmpeg, docker, node, Kokoro, whisperX, keys, templates;
  every failure prints the exact fix command.
- **`status`** — lesson × stage table (done / stale / pending).
- Well-tested: ~19k LOC of Go with test files across pipeline, llm, project,
  config, studio; ~3.2k LOC of studio TS/TSX with vitest.

---

## 4. CLI surface

```
coursesmith init <name>              scaffold a new course
coursesmith run <course>[/<lesson>]  run the pipeline (resumable)
    --stage <name> | --force | --concurrency <n>
coursesmith preview <course>/<lesson>  open in Remotion Studio (hot reload)
coursesmith status <course>          stage table
coursesmith doctor                   environment checks + fixes
coursesmith serve [--addr ...]       studio JSON API + UI
coursesmith audition <course>        sample every voice  [--choose <voice>]
coursesmith analyze <course>         concept graph + dependency/terminology QA
coursesmith export-review <course>   SME review docs
coursesmith build-site               hugo (+ pagefind) build
coursesmith bundle                   offline zip (videos included)
coursesmith ebook                    print-styled PDF companion
coursesmith compile-course <course>  join lesson finals into course.mp4
coursesmith snippet templates        the template catalog, grouped by category
coursesmith snippet new <prompt>     plan + render a clip
    --template (required) | --seconds | --mode | --captions | --skin | --plan-only
coursesmith snippet run <id>         re-run one (up-to-date stages skipped)
coursesmith snippet list             every snippet and its state
coursesmith combo direct <subject>    subject → argument, parts, looks, runtimes
    --minutes | --skin | --captions | --mode
    --run goes straight through; otherwise writes combo.yaml and stops
coursesmith combo new <title>         --segment template:prompt (repeatable)
coursesmith combo run <id>            plan every segment, one timeline, one render
coursesmith combo list | show <id>    inspect a combo and its segments
coursesmith combo segment <combo> <seg>  edit one segment (template/prompt/skip)
```

---

## 5. What's left / rough edges

- **Only two lessons authored** in the sample course (`python-basics/01`,
  `/02`). Full multi-lesson runs (spaced repetition, bridges, analyze at scale)
  are lightly exercised.
- **One content provider in practice.** The router still supports any
  OpenAI-compatible backend, but the default and every shipped course manifest
  now pin `openai/gpt-4o-mini` — Groq is no longer wired in anywhere by
  default, and nothing else has been validated end to end at this scale.
- **A combo is planned per segment, so a long one is a lot of calls.** Nine
  segments is nine planning calls plus enrichment; plan-only is the studio's
  default at that size for exactly that reason, but there is no cost estimate
  in the UI before the button is pressed.
- **The recast fallback hides a bad cast.** A segment whose template cannot be
  planned is recast as `illustration` and the run continues — which is right,
  but it means a combo can finish with a segment nobody chose. It is logged and
  `combo.yaml` keeps the original pick; it is not surfaced in the studio.
- **No release tagging.** It is a git repo now, with history, PR merges, and
  five GitHub Actions workflows (`quality-gates`, `visual-regression`,
  `accessibility`, `learning-science`, `deploy-site`) — but nothing is versioned
  or tagged, so there is no "known-good build" to point anyone at.
- **Studio is early** (v0.1.0). Core run/monitor/quiz-edit flows exist; deeper
  editing (script/diagram authoring in-browser) is not there yet.
- **Graceful-degradation paths are less tested** than the happy path (no Node →
  ffmpeg fallback; no whisperX → segment timing; no Docker → unsandboxed host
  exec).

---

## 6. Where this is going

Natural next moves, roughly in priority order:

1. **Combos past phase 4** — a cheap edit path that patches a rendered segment's
   props without re-planning its narration (the shape `video-plan.yaml` already
   has for lessons), and surfacing recast segments in the studio.
2. **Full-course authoring at scale** — write out the rest of `python-basics`,
   then stress the cross-lesson features (concept graph, bridges, spaced
   repetition) on a real 10–15 lesson course.
3. **Deeper studio** — in-browser script/diagram/quiz editing with live
   regenerate, diffing drafts, and the review-notes loop surfaced in the UI.
4. **More providers & models** — validate the router against other
   OpenAI-compatible backends and local models for a fully offline generation
   path (Kokoro + whisperX are already local).
5. **Publishing polish** — richer Hugo theme, search (pagefind) tuning, LMS/SCORM
   or YouTube-upload export alongside the ebook/bundle outputs.
6. **Repo hygiene** — tagged releases so the pipeline itself is reproducible
   (`git init` and CI are done: `quality-gates` runs `go vet` / `go test`, the
   renderer and studio typechecks, vitest and the animation-timing gate, and
   `visual-regression` enforces the baselines).
7. **Broader subjects** — the sandbox and verify gates are Python-first;
   generalizing execution verification to other languages would open up the
   engine beyond Python courses.

---

## 7. Repo map

```
cmd/coursesmith/     CLI commands (run, status, doctor, serve, snippet, combo, …)
internal/
  pipeline/          the 15 stages + render/audio/align/quiz/ebook/bundle
                     snippet*.go — the 32 templates, their prompts + validators
                     combo*.go    — casting, per-segment planning, assembly
                     videotheme.go, typing.go — theme/skins, keystroke rhythm
  llm/               providers, router, rate limiter, cache, transcribe
  project/           course/lesson/state parsing, Stage/Snippet/Combo orders
  config/            layered config
  studio/            Go JSON API + SSE + ledger + artifacts
prompts/             *.tmpl generation prompts + diagram_style exemplars
renderer/            Remotion (Node/React) video engine — a scene per template
studio/              React + Vite + Tailwind studio UI (rail + light/dark)
site/                Hugo skeleton + course theme
sandbox/             Docker image for code verification + VHS demos
tools/align/         whisperX venv (word-level timing)
test/                visual-regression baselines + gallery-preview builder
docs/                plans + tool research
courses/             python-basics sample course
.github/workflows/   quality-gates, visual-regression, a11y, deploy-site
```

**Build:** `make build` · **Test:** `make test` · **Lint:** `make lint`

---

## 8. The 10x visual overhaul (2026-07-18, second pass)

Driven by the verdict "the course doesn't look good, visuals aren't
context-aware." Full plan: `docs/plan-10x.md`; tool research (4 web surveys):
`docs/research/`.

**Design system.** Go derives a rich dark-editorial video theme from the three
course colours (`internal/pipeline/videotheme.go`: hue-tinted gradient stops,
surfaces, text tones, grain, type stack; `branding.fonts` overrides). The
renderer resolves it (`renderer/src/theme/theme.ts`), loads Google Fonts
deterministically (Space Grotesk / Inter / JetBrains Mono + Sora, IBM Plex
Sans), and every scene sits on `SceneBackground` (gradient + noise-drifting
accent glows + feTurbulence grain + vignette — all `useCurrentFrame`-driven).
TitleCard, CodeScene (editor chrome, line numbers, dark-plus tokens,
verified-output drawer), DiagramScene (light card framing), exec-viz,
MemoryLayout, D3 and captions all restyled to the tokens.

**No dead screens.** New `storyboard` stage (between review and visuals):
one cached LLM call plans 2–5 keyword beats per section (closed icon
vocabulary, at_word timings). Heading-only sections become `points` scenes —
phrases with Lucide icon chips that pop in on the exact spoken word
(rows/grid template variants).

**Context-aware visuals.** All diagram prompts (svg/d3/d2/mermaid/excalidraw)
now receive the narration of the section that cues them + the audience. New
`d2` diagram kind: pure-Go D2 (MPL-2.0) compiled in-process, sketch mode,
compile-errors feed the model's self-repair, same vision-QA gate.

**VS Code walkthrough.** New `walkthrough` scene (`VSCodeScene.tsx`):
synthesized editor — activity bar, file tree of all code-bearing sections,
tabs, minimap, status bar with live Ln/Col. Sections with 2+ Python blocks
become one (block N = timed step N; step 1 types, later steps flash changed
lines). Frame-deterministic by construction (research ruled out Monaco /
screen capture / Playwright video).

**Captions.** Page-based karaoke (3–5 words/page), spring pop + accent on the
active word, backdrop-blur pill; `captions` stage runs an LLM emphasis pass
(`captions_emphasis.json` → `SceneGraph.CaptionEmphasis`) marking ≤1 keyword
per page that stays accent-coloured.

**Voice control.** `style.voice_speed` (Kokoro speed param), voice blends
documented (`af_bella(2)+af_sky(1)`), and the pace loop is CLOSED: align
writes `tts_speed.json` (bounded 0.75–1.35, composes with voice_speed) when
the lesson misses `pace_wpm` ±15%; its appearance re-stales audio. `TTS_URL`
(engine-neutral alias of KOKORO_URL) + documented Chatterbox Turbo path for
cloned voices.

**Templates & bigger videos.** Archetype animation styles now select scene
template variants (`sceneTemplates` in archetypes.go; playful → points grid).
Per-lesson `video-plan.yaml` (lesson dir, re-stales scenegraph) edits scenes:
template swap / props patch / skip-with-span-absorption. New
`coursesmith compile-course <course>` joins all lesson finals into
`course.mp4` (lossless concat) + `course-chapters.txt` (YouTube format).

**Verification.** All Go suites green (incl. new tests: videotheme, caption
emphasis, D2 compile determinism, storyboard/walkthrough scenegraph paths,
video-plan, speed-fix); renderer + studio tsc clean; studio vitest 84 pass;
visual-regression baselines re-recorded (0px-deterministic); animation gate
updated for the ambient background and passing; before/after stills eyeballed
for every new scene type.

---

## 9. Snippets — the short-form surface (2026-07-27)

Driven by the pivot to a SaaS creators actually use: a course lesson is a
document first and a video second, which is the right shape for a 12-minute
lesson and the wrong shape for the 30-second clip a creator wants for a landing
page. A **snippet** inverts it — the prompt *is* the input, a **template**
decides what the screen looks like, and one LLM call produces the narration and
the visual spec together.

**The spine.** New `plan` stage (`internal/pipeline/snippet.go`) turns
`snippet.yaml` (prompt + template + target runtime) into `snippet-plan.json`,
and from it the `script.json` and `lesson.md` the rest of the pipeline already
expects. `project.SnippetStageOrder` is then the ordinary video path:
`plan → verify → audio → align → captions → chapters → scenegraph → render`.
Nothing downstream knows a snippet was not hand-written, so a clip inherits the
whole quality moat — real executed code, word-accurate whisperX timing, the
design system — with no second engine.

On disk a snippet is a lesson directory inside a synthetic course
(`.coursesmith/snippets/lessons/<id>/`), which is what lets the existing stage
machinery, state tracking, studio artifact serving, and `/api/lessons/...`
routes work on it with no special cases.

**The template catalog** (`snippet_templates.go`). A template owns the prompt
that plans the clip, the rules that plan must satisfy, and the mapping from a
planned-and-timed clip onto renderer scenes. Shared code owns timing, theming,
captions and the scene-graph envelope, so no template can drift from the design
system — a template only decides what fills the frame.

- **`vscode`** (`snippet_vscode.go`) — an editor opens, the file is picked out
  of the tree, code types itself in, and the integrated terminal slides up and
  runs it. The terminal output is not written by the model: the plan's code goes
  through the ordinary verify stage and the scene shows what the interpreter
  really printed. A beat that runs unverified code fails the build.

  One rule a plan can break silently, so it is checked
  (`checkBufferCarriesForward`): a beat's `code` is the **whole file as of that
  beat**, not the lines the beat adds. The prompt says so and models still hand
  back a diff — a beat defining the variables followed by a beat holding only
  the `print()` calls. It reads fine as a plan and fails twice downstream:
  verify executes each buffer state alone, so the second dies on a `NameError`;
  and the editor types whatever the buffer says, so a buffer that dropped its
  history would wipe the file mid-thought and retype — the jump cut this
  template exists to avoid. The test is "most of the previous buffer survives",
  not "is a prefix of", because editing a few lines is legitimate and reads
  well; replacing all of them does not.

**Enforced length.** Models systematically under-write to a seconds target (they
have no clock) — a 45s request came back as 58 words, half a video. The word
budget is now a gate, not a suggestion: a plan outside 75–130% of
`target_sec × pace_wpm / 60` is rejected and regenerated with the miss quoted
back.

**Scene work.** `VSCodeScene` gained the full choreography: window scales up,
tree hover → click → tab slide-in timed *backwards* from the first keystroke (so
a long intro beat holds on the open window instead of stretching the gesture),
a terminal drawer that grows the window and types its command before streaming
output line by line, indent guides, an active-line band, and a run-time accent
bloom. The window was re-proportioned to 1520×~560 (a 1680-wide window holding
six lines of code reads as a letterbox strip no matter how much frame it
covers), the terminal typesets its output to fit rather than clipping the last
line, and the whole thing stays clear of the caption card. `TitleCard` now
renders its subtitle in intro mode — it was being silently dropped, so every
snippet's hook went missing.

**The editor has its own palette, and it has two.** Every colour in
`VSCodeScene` used to be a hard-coded dark literal, so in light mode the scene
was a dark editor punched into a paper page. There is a `Chrome` record per
mode now — surfaces, hairlines, chrome text, the active-line band, the minimap
column, the shell prompt — plus the Shiki theme, because dark-plus tokens on a
white editor are their own kind of unreadable (its comment green and string
orange are both picked to sit on `#1e1e1e`).

These are deliberately **not** design-system tokens. An editor painted in
`surface` and `mass` is a courseSmith panel with code in it; the credibility of
this template comes from looking like the tool on the viewer's second monitor.
What it does take from the theme is the mode, the primary (status bar, and the
active-item rail in light mode — the accent is a saturated yellow picked for
the dark stage and is a highlighter stroke nobody can see on light chrome) and
the accent (the caret, and the rail on dark).

Two smaller things that were wrong for the same reason nobody had looked: the
**minimap** was one grey bar per line, which would look identical whatever the
file said — it draws one block per *token* now, coloured by that token, with
indentation preserved and a viewport box that only appears when the file is
actually longer than the pane. And the **demo had no run step**, so the
terminal drawer — half of what this template is for — had no composition and no
baseline. It has both now, in light mode, which is where a regression in it
would otherwise have hidden.

**Runtimes go down to 10 seconds, and the beat count comes from the budget.**
The shared beat range was a fixed 3-7, and every snippet prompt also calibrates
a beat at "about forty words" — so a three-beat floor is a 120-word floor. A
20-second clip's *ceiling* is 89 words. The two rules contradicted each other
and no plan could satisfy both; in the field this showed up as a plan walking
128 → 114 → 96 → 93 words across three correction rounds and failing, converging
on a floor the instructions themselves imposed.

`beatBounds` derives the count from the narration budget instead, and hands the
prompt the per-beat number that budget actually affords — so the number the
model is told to write and the number it is scored against are the same number
at every runtime, not just above 45 seconds:

| runtime | budget @175wpm | beats | words/beat |
|---|---|---|---|
| 10s | 21-44 | 2 | 14 |
| 20s | 43-89 | 2 | 29 |
| 45s | 98-203 | 3 | 43 |
| 120s | 262-542 | 7 | 50 |

The beat floor is two rather than three: two beats is one cut, which is the
least that still makes this a film rather than a held shot, and it is what
makes a ten-second clip expressible at all. `story` keeps its own 8-14 range
and recomputes its own per-beat number against it — inheriting the shared
arithmetic gave it a figure that multiplied straight past its own ceiling.

Two smaller fixes fell out of the same investigation. The word-budget error
said "rewrite with fuller sentences" in **both** directions, so a plan 40 words
over its ceiling was told three times running to write more; the advice now
points the way the plan has to move. And the calibration paragraph in all seven
prompts no longer quotes a fixed forty words — it quotes the runtime's own
arithmetic.

**Downloads are named after their titles, not `final.mp4`.** On disk every
lesson's video is `final.mp4` and that name is a contract — compile
concatenates it, the Hugo page embeds it, chapters splits it. It is also the
worst possible thing to have six of in a Downloads folder. The studio's
artifact route rebuilds the name on the way out (`downloadName`):

    Python Fundamentals 01 The print function making Python say things on screen.mp4
    Python Fundamentals 01 The print function … - 03 printing multiple items.mp4
    Applications of Python.mp4

It is built from the **titles** — the words the user sees in the studio — and
not from the directory names, and for snippets that distinction is the entire
point. A snippet's directory id is a slug of its *prompt* (`uniqueSnippetID`),
so the clip titled "Applications of Python" lives in a folder called
`hand-drawn-whiteboard-animation-illustrating-pyt-2`. Naming the download after
the directory swaps one unhelpful filename for another. A lesson leads with its
course and number so a folder sorts into teaching order; a snippet is one clip
with one title, and its course is always "snippets", so its title is the whole
name.

Spaces and capitals survive — reading as the title *is* the improvement, and
"Applications of Python.mp4" is worth more than "applications-of-python.mp4".
What does not survive is the set Windows rejects (`/ \ : * ? " < > |`), dropped
rather than replaced so "language?.mp4" becomes "language.mp4" and not
"language-.mp4"; titles are model-written prose and really do contain question
marks and colons. Over-long names lose whole *words* from the middle, keeping
the head that says which lesson and the tail that says which part.

It ships as `Content-Disposition: inline` — **not** `attachment`, because the
same URL is the `<video>` element's src and the download link's href, and
`attachment` stops the player ever showing a frame. Browsers take the filename
from the header in preference to an anchor's `download` attribute, so the one
header names the file on every path out: the UI, a right-click, `curl`. The
same string is also returned as the artifact's `download_name` so the UI passes
it explicitly rather than reimplementing the rule and drifting from it.

**Surfaces.** `coursesmith snippet templates | new | run | list`, plus
`/api/snippet-templates`, `/api/snippets` (CRUD) and a Studio **Snippets** page:
template gallery, prompt, runtime picker, one button, and a player + beat
breakdown for the finished clip.

- **`whiteboard`** (`snippet_whiteboard.go`) — one board that fills in as the
  narrator talks: a hand-drawn box is sketched, its icon wipes in, its label is
  written, and an arrow reaches across from an idea already on the board. The
  board never wipes; the accumulation is the form.

  The model does not draw. It picks *what* goes on the board — a short label, an
  icon from the closed vocabulary, and which earlier item it follows from — and
  the renderer draws it on a grid it owns. Letting an LLM author freehand SVG is
  what made the `visuals` stage need a vision-QA gate and a repair loop; this
  trade gives up nothing that reads on screen.

  Strokes are generated analytically (`renderer/src/components/sketch.ts`), which
  buys three things an imported drawing cannot: exact path length for a
  `stroke-dashoffset` draw-on, the pen's position at any frame for the marker
  that leads it, and ownership of the layout. Boxes are **double-stroked** with
  separate seeds — a single wobbled outline still reads as a `rect` element; two
  passes read as a pen. The wobble's harmonics are whole cycles so the function
  is periodic, which makes the overshoot land exactly on the stroke's start
  instead of as a floating tick. The grid is **serpentine**: rows alternate
  direction so consecutive items are always adjacent, because a straight
  left-to-right second row put the connector on a long diagonal straight through
  the boxes in between.

**Enforced length, revisited.** The floor is enforced (75% of
`target_sec × pace_wpm / 60`) because undershooting is fatal — the visuals are
timed to the voice. The *ceiling* is loose (155%), and that is a measured
decision, not laziness: a 135% ceiling was tried and the model produced 184-185
words in three consecutive correction rounds on a topic it judged to need that
much, ignoring the stated target entirely. Rejecting those plans bought failed
generations, not shorter clips. Runtimes are therefore documented and labelled
as approximate, and the finished duration is always reported.

**Repair rounds.** `completeWithRepairRounds` generalizes the one-shot repair
loop, and the snippet planner uses three. It also carries *every* rejection
forward rather than just the latest: with several independent numeric rules, a
single-round loop quoting one error made models oscillate — attempt one missed
the item count, attempt two fixed the count and blew the word budget, and the
call failed having never seen both constraints at once.

- **`flow`** (`snippet_flow.go`) — labelled boxes in layered columns, edges
  curving between them, and once the graph is up, **tokens visibly moving along
  those edges** while the narration traces a path through it. Beats declare a
  `focus` set; everything outside it dims and its traffic stops, so one diagram
  narrates the happy path, then the failing path, without ever being redrawn.
  That is the whole reason this is not a static Mermaid render: a systems
  diagram is about movement, which a picture can only imply.

  The model declares a DAG (nodes with a kind and the nodes that feed them) and
  **Go ranks it** — longest-path layering, in `flowScenes`. That split is
  deliberate: ranking is graph logic that wants Go unit tests, placement is a
  function of the stage box (`renderer/src/components/flow.ts`), and neither can
  quietly depend on the other. Edges are S-curves with horizontal control pulls,
  so every edge leaves a box going right and arrives going right and the eye
  never has to work out which end is which.

  Two rules are enforced rather than suggested, and both make the template
  distinct: the graph **must fork or join** (a straight chain wastes a layered
  layout and is what `whiteboard` already draws well), and each focused beat
  **must focus a different, strictly smaller set** (focusing everything dims
  nothing; two identical sets narrate the same picture twice). Depth is capped
  at 4 columns and width at 4 rows, because that is what leaves a readable label
  at 1080p.

- **`illustration`** (`snippet_illustration.go`) — kinetic typography. A short
  phrase lands word by word in 60-100px type, one word of it takes a marker
  stroke swept in underneath, and a flat-vector figure assembles beside it. The
  figure changes sides every beat.

  This is the one template that **does not accumulate**, and that is the design
  rather than an omission. The board and the diagram are both about a picture
  being built and staying built; this one is about a phrase landing, and a
  phrase cannot land on a stage still holding the last four. One beat is one
  shot, and the clip cuts. That makes it the template for the parts of an
  explainer that are rhetoric rather than architecture — the hook, the turn,
  the payoff — which the other two are bad at.

  **The figures are drawn, not imported** (`renderer/src/components/artwork.tsx`).
  Bundling a CC0 set (unDraw and friends) was the obvious alternative and it
  loses the only thing that matters here: a downloaded illustration is a single
  flat blob of paths, so a rocket's flame cannot flicker and a gear cannot turn.
  Owning the geometry means every figure has *parts*, and every figure keeps a
  continuous idle running after it assembles — the flame licks, the gears mesh
  and counter-rotate, the clock sweeps, packets run the network's spokes. A
  figure that assembles and then freezes is a slide no matter how good the
  entrance was. It also means the artwork speaks the design system's palette
  instead of being recoloured towards it, and there is no third-party asset
  licence to track. A hundred and one figures now, closed vocabulary, `spark` as
  the fallback, drift-tested against Go the same way the icon vocabulary is.

  The drawings live in themed modules under `artwork/` — tech, product, science,
  nature, work, abstract — over a shared `kit` that owns the box, the two clocks
  and the staggered entrance. One file holding a hundred figures was a file
  nobody could find anything in, and factoring the boilerplate out is what keeps
  each figure down to the part that is actually interesting: its mechanism. What
  stays in `artwork.tsx` is the vocabulary itself, because that is the thing Go
  mirrors.

  Two rules earn their keep. The emphasis must **occur in its own beat's
  heading** — it is a stroke drawn under part of the headline, so a phrase that
  is not there has nothing to underline and the shot silently loses its accent;
  matching is on letters and digits only, because the two fields come from
  different places in one reply and models are inconsistent about echoing case
  and punctuation. And **at most two beats may share a figure**, because a run
  of shots on one drawing is a still image with the text changing, which is
  exactly what this format must not be.

  `FigureSheet` (a development composition, no baseline) renders the whole
  vocabulary on one frame. It exists because the figures are the part that
  cannot be checked by reading: a geometrically fine path can still read as a
  letter, or vanish because it was painted in the shading colour on a dark
  stage. Both happened.

- **`cast`** (`snippet_cast.go`) — a character explains it. Same
  one-beat-one-shot shape as `illustration`, and the difference is the whole
  reason it exists: an object shows what a thing *is*, a person shows how to
  *feel* about it. A hand to the chin is "hold on", a shrug is
  "nobody's sure", a raised finger is "here it is" — the register an explainer
  opens and closes on, which no diagram can reach.

  **The character is Open Peeps** (CC0, Pablo Stanley, via the MIT `react-peeps`
  package), composed in `renderer/src/components/cast.tsx`. It replaced a
  skeleton rig that drew a person from eleven joint angles, and the trade is the
  opposite of the one `artwork.tsx` makes for objects — deliberately. The rig
  could do *anything*: any pose was eleven numbers and poses interpolated. What
  it could not do was look like a person somebody drew, and next to the
  illustration template's artwork it read as the cheap thing on the stage.
  Nobody knows what a "correct" rocket looks like, so a built one is just a
  rocket; everybody knows what a person looks like. **Objects are worth owning,
  people are worth importing.**

  The cost is that the vocabulary is no longer ours to invent — a pose exists
  exactly when somebody drew it *and* can be coloured. `wave`, `celebrate`,
  `defeated` and `walk` went because no drawing of them exists; `explain`,
  `coffee` and `phone` went for the colour reason below. Six remain: `idle`,
  `point`, `think`, `shrug`, `confident`, `reading`. The register `defeated`
  carried moved to the face as `sad`, where it reads better than a slump ever
  did, and the eight expressions are now where most of the range lives. Not
  offering a name the artwork cannot satisfy is the same rule the snippet
  runtimes follow, and for the same reason: a pose Go allows and the artwork
  lacks renders as somebody standing still through the beat that was meant to
  be its punchline.

  **The offset table** is what makes ten drawings behave like one character.
  Every pose is framed around the head axis rather than around its own bounds,
  with a fixed vertical crop, so the head is in the same place at the same size
  in every shot and only the arms change — a shrug simply gets a wider frame,
  and the scene lets it overhang into the gutter rather than shrinking the
  person. The numbers in it were measured off the rendered artwork (rasterise
  each part, scan for its ink) rather than guessed. Each pose also carries a
  `headPlay`: how freely the head may breathe on it. `think` rests a hand
  against the jaw, and a head that drifts a few pixels off its own knuckles is
  an error the eye catches instantly.

  What survives from the rig is the thing that actually made it look alive.
  Breathing, a head tilt and a blink run over the artwork, and a pose change
  between beats dissolves and settles rather than cutting. The blink is a face
  swap to `EyesClosed` for an eighth of a second — Open Peeps has no eyelid
  layer, and at that duration nobody resolves the mouth changing too.

  **Four colours, and which busts are usable follows from them.** A peep is
  three layers — body, hair, face — each carrying its own pair, so skin, hair,
  garment and ink can each be their own value. The first version passed one pair
  to all three, painting skin, hands and shirt one flat colour: it rendered, it
  was legible, and it looked like a sticker rather than a person. That was the
  single biggest thing wrong with the character and it was wiring, not the
  artwork.

  The catch is that the body layer's `backgroundColor` paints the hands *and*
  any garment the illustrator filled with it, and the two cannot be separated.
  Every bust was classified by rendering it in two loud colours and counting
  pixels; the split is clean at about 55%. On a stroke-filled bust background is
  free to be a skin tone and everything lands. On a background-filled one,
  keeping the hands right dresses the character in their own skin and keeping
  the garment right hides their hands in their shirt. **That is why POSES is the
  list it is** — and why `explain`, the most useful gesture in the set, is not
  in it. Pose and outfit stay fused, so the neckline and pattern still change
  between beats; what the restriction buys is that the *colour* does not, which
  is what the eye actually tracks as "the same person".

  **The cast is eight presenters** (`CASTS`), each a skin tone, a hair colour
  and a hairstyle, picked deterministically from the clip's seed so two snippets
  are visibly presented by two different people and no frame disagrees with its
  neighbour about who is on screen. Hair colours are held above about 20%
  lightness because the hair's outer edge is the silhouette against the dark
  stage. The garment is the theme primary for everyone: that is where the brand
  lives, and people are told apart by face and hair anyway.

  One enforced rule: **no two consecutive beats may share a pose**. A character
  holding still across two beats is a photograph with the text changing beside
  it. It is checked on *normalized* names, so two beats that both fall back to
  `idle` still collide.

  `CastSheet` (a development composition, no baseline) renders every pose and
  every expression on one frame. A character fails differently from a figure:
  the artwork is somebody else's and always renders, so the failures are ones
  only an eye catches — a head drifted off the hand resting against it, a pose
  clipped by its frame, a face saying something other than the word it is filed
  under. It is also what caught the poop emoji drawn on the laptop lid of the
  `Geek` bust, which is why there is no `typing` pose.

- **`story`** (`snippet_story.go`) — the long one: a directed piece of one to
  two minutes, staged shot by shot. Eight to fourteen beats, its own 90-second
  default runtime (the shared 45s cannot fund a beat floor of eight without
  writing every beat at the ten-word minimum).

  **The plan is two LLM calls, and that is the whole design.** The writer and
  the director are different jobs, and one call doing both does neither: asked
  to invent narration *and* stage it, a model spends its attention on the words
  and stages everything identically — fourteen beats of "person left, object
  right", which is a slideshow with a presenter in it. So call one writes the
  script and nothing else (the prompt is tested for not even mentioning staging
  vocabulary). Call two is handed the finished script, every beat at once in
  order, and does only direction. It can see the arc, which is the point: you
  cannot decide *this* is the beat to push in on unless you can see the beats
  either side of it. The two calls cache independently, so tuning the director
  prompt does not re-pay for the script.

  The director picks from closed vocabularies — five **stagings** (hero, duo,
  object, pair, empty) and six **camera moves** (hold, push, pull, pan, rise,
  drift) — and the renderer owns every coordinate. Rules, all enforced: no cut
  may repeat *both* staging and camera (repeating one is fine and is how a scene
  gets built); at least a third of shots carry the presenter; at least half move
  the camera; at least three stagings across the piece.

  **The camera is real.** `renderer/src/components/camera.ts` lays the shot out
  on a world 1.34× the frame and points a camera (x, y, zoom) at part of it,
  eased over the shot's *own* duration — which is why a long beat gets a slow
  move. The backdrop tracks at 0.42×, because depth on a flat stage is entirely
  differential motion: matched to the camera it is painted on the front element,
  unmoving it is wallpaper behind a cutout. Amplitudes are deliberately small (a
  14% push over eight seconds); the first pass used roughly double and every
  shot read as a zoom effect rather than as a camera.

  Two staging bugs worth remembering: a character stands on the floor line by
  its **soles** (`CAST_FEET`), not by the bottom of its drawing box — that box
  carries headroom for `celebrate`'s raised arms and using it directly leaves
  the figure visibly hovering. And a shot with no presenter must not advance the
  pose the next one eases from, or the character teleports across the
  intervening object shots.

**The gallery decides for you.** Each template card carries a real frame from
that template, downscaled from its visual-regression baseline
(`test/template_previews.mjs` → `studio/public/template-previews/`), so the
picture on the card is by construction what the template renders — a hand-drawn
thumbnail would drift the first time a layout changed and nothing would compare
them. Picking a template also fills the prompt box with that template's own
example, because the moment after someone reads what a template does is the
worst time to leave them inventing a prompt that suits it. It only overwrites
text that came from an example, checked against the whole example set rather
than a dirty flag — so browsing keeps swapping the demo, and one typed
character makes the box theirs. A Go test fails if a registered template has no
preview.

- **`data`** (`snippet_data.go`) — real numbers on one chart or world map.
  Thirteen kinds, because the shape of a dataset is not a style choice: parts of
  a whole, a value over a sequence, and two variables against each other are
  three different claims, and drawing any of them as bars states the wrong one.

  One number per label: `bars` (horizontal, the only orientation where a real
  label fits unrotated), `line`, `area`, `donut`, `waffle`, `treemap`, `funnel`,
  `gauge`, `kpi`, `map`. Several numbers per label, declared as named `series`:
  `stackedbars`, `groupedbars`, `scatter`. `series` means the *parts* of a label
  on the bar kinds and the two *axes* on a scatter — one field rather than two,
  because both are "one number per named dimension".

  Every kind is written against one context object, which is what keeps thirteen
  of them honest with each other: `dim`, `tint` and `grow` are computed once, so
  a highlight looks the same on a treemap tile, a scatter dot and a country. A
  kind that had to invent its own idea of emphasis would be a second design
  sharing a file with the first.

  What each kind is allowed to claim is checked, not suggested. A `funnel` whose
  values rise is rejected — drawn from data that widens it is a picture of a
  funnel with numbers written on it, which is worse than no chart because it
  looks like it means something. `donut` and `waffle` must total above zero,
  since both assert their values are shares of one thing. Several kinds cap
  points below the shared ceiling (5 for `waffle`, `gauge` and `kpi`, 6 for
  `funnel` and `groupedbars`, 8 for `treemap`) because the drawing runs out of
  room before the reader does. And a point whose `values` row is short of the
  declared series is an error rather than a pad: a stacked bar missing its third
  segment renders perfectly and states a total that is not the total.

  The chart is declared **once for the whole clip** and the beats only move the
  *emphasis* around it. A chart per beat was the obvious alternative and it
  fails the same way a new diagram every eight seconds does: the viewer never
  gets past reading the axes. One chart that stays put is what lets the second
  mention of a bar mean something, because it is the same bar. That is why
  `Chart` sits on `SnippetPlan` rather than on a beat — the only place in the
  catalog where visual state is not per-beat, and it is deliberate: the dataset
  is a property of the clip.

  Enforced: at least half the beats highlight something (a chart nobody points
  at is a screenshot with narration over it), and no two consecutive beats
  highlight the same set, compared order-independently.

  **The map** is `world-atlas` countries-110m — Natural Earth, public domain,
  108KB — projected once at module load with d3-geo (`geo.ts`); projecting 176
  countries per frame would be pure waste since only the fill changes.
  Antarctica is dropped: it is never the subject, and it costs a fifth of the
  box height plus the fit, leaving the inhabited world floating in the upper
  two thirds. Countries without data are still drawn, because a map of only the
  highlighted countries is a set of floating shapes nobody can place.

  Natural Earth's names are cartographer's names — "United States of America",
  "Dem. Rep. Congo", "Bosnia and Herz.", and it still says "Macedonia" — and no
  model writes those. `countries.go` carries the canonical list plus an alias
  table, and a drift test compares both against the real TopoJSON. It has
  already earned itself twice: it caught an alias pointing at a country the
  atlas does not have, and Antarctica being accepted by Go while the renderer
  refused to draw it.

- **`workspace`** (`snippet_workspace.go`, `WorkspaceScene.tsx`) — the `vscode`
  template's bigger sibling, and the difference is the *presentation*, not the
  code. `vscode` is a window floating on the brand stage, which is the right
  frame for eight lines that read large and sit still. This one fills the
  screen, has no stage behind it, and **moves**: it zooms into the function
  being written, pans to the tree when a file opens, pulls back for the wide
  shot, and drops to the terminal for the payoff. Reach for it when the clip
  should feel like a screen recording, when the code is longer than a window
  holds, or when there is more than one file.

  Two things follow from "screen recording" and they are the whole design. **One
  scene spans the clip** — a scene per beat remounts the editor every few
  seconds, and a remount is a cut; nobody cuts in the middle of a capture, so
  the beats are timed steps inside one continuous editor (the same shape `data`
  uses, for the same reason). And **the camera moves rather than the layout**:
  everything lays out once at frame size and a transform picks the part worth
  looking at, so panning and zooming cost nothing because the content never
  re-flows. A file longer than the pane scrolls, and the camera follows the line
  being written. The code area is sized from the *longest* file in the project
  plus headroom, so switching tabs does not make the layout jump — and a
  four-line program does not sit five hundred blank pixels above the terminal
  proving it works.

  The model never supplies a coordinate. It names a **subject** — `wide`,
  `code`, `tree`, `tabs`, `terminal` — and the geometry turns that into a
  camera. Same bargain the story template's staging makes, for the same reason:
  a model handed x/y frames the gap between two panels. A drift test keeps the
  Go vocabulary and the renderer's switch equal. Bounds: 1–4 files (four,
  because a fifth tab is a tab nobody looks at), ≤40 lines each (the cap is
  reading time, not screen space, since the pane scrolls), and a 30s minimum
  runtime — a zoom that has to finish in a second is a cut, so below that the
  template's reason for existing does not fit. Default 60s.

  **Multi-file needed real execution.** Both sandbox runners piped a single
  script through stdin, so `import greet` was a guaranteed
  `ModuleNotFoundError`. New `CodeRunner.RunProject(files, entry)` writes the
  set to a directory and runs the entry point there: under Docker a
  **read-only** bind mount (the program has no business editing its own source,
  and a clip that quietly rewrote the file it just showed would be lying about
  what ran), a temp dir on the host runner, and paths refused if they climb out
  of it — they come from a model, and Docker would contain that where the host
  path would not.

  The project runs **inside the planner's correction loop**, not downstream in
  verify, which executes each fenced block alone where the import fails. A
  program that raises comes back to the model as its own traceback; one that
  prints nothing is rejected, because the terminal is the payoff. That is what
  lets this template be pointed at something complicated — the loop that catches
  it is execution, not taste. Also enforced: every beat names a file the project
  has, some beat sets `run`, and the *first* beat does not (running the program
  before any of it is on screen).

  Several files are supported, **not required**. Insisting on a second produced
  exactly the thing nobody wants: a module invented to satisfy a validator.

**Cross-template field guards.** `SnippetBeat` is the union of what every
template needs, so `beatFields` declares ownership once per template and
`rejectForeignBeatFields` fails loudly when a plan sets a field its template
ignores. Each template hand-checking the others was quadratic and rotting.

**Per-frame render budget.** `RemotionRenderer.FrameTimeoutMs` (default 180s,
up from Remotion's 30s). A frame measures ~100ms, but a busy machine missed the
deadline once and aborted an otherwise-finished clip. The budget was raised
rather than the scenes made cheaper, because the scenes are not the problem.

**Light and dark.** `style.mode` picks which set of lightness targets the
branding hue runs through (`videotheme.go`). Both modes emit the same token
names, so no scene asks which mode it is in — a scene that has to branch on mode
is a scene with a colour hardcoded in it. Three tokens exist because the
polarity flip breaks the others: `mass`/`ink` (artwork body fill and its
shading, which flip together so shading stays shading), and `accentText` — the
accent walked down in lightness until it is legible as *type*, since a brand
accent is chosen for a dark stage and a saturated yellow is very nearly the
luminance of paper. The pairs are asserted around the hue circle in both modes
(`videotheme_contrast_test.go`) rather than eyeballed, because they are derived
by formula and the first sight of a bad pair would otherwise be a finished
video. Captions and mode are both per-snippet: CLI flags and Studio controls.

**Every template has a picture of itself.** `Root.tsx` carries a development
composition per scene type, and the visual-regression suite records a baseline
from each — fifteen at the time of this pass, **51 now**, all
0px-deterministic, with the gallery previews downscaled from them. Two exist
because a scene cannot be
checked by reading. There is a **light-mode VS Code twin**, since the editor
carries its own palette and is the one scene whose light mode cannot be
inferred from any other baseline; and `FigureSheet` / `CastSheet` render the
whole artwork and pose vocabularies on one frame.

Re-recording them closed two gaps that had been invisible. The VS Code demo had
no run step, so the terminal drawer — half of what that template is for — had no
composition and no baseline. And the cast and story fixtures still named poses
retired when the character moved to Open Peeps; those fall back to `idle`, so
those beats had been rendering as a motionless character and the baselines had
locked it in. A baseline is only a regression test for what it actually
exercises.

**That first catalog was eight templates** — `vscode`, `workspace`,
`whiteboard`, `flow`, `illustration`, `cast`, `story`, `data` — spanning code,
hand-drawn board, systems diagram, kinetic type, presenter, directed short, and
data. It is 29 now (§11), sitting on a house-style layer that did not exist
then (§10).

---

## 10. The house style — skins, surfaces, transitions, sound (2026-07-28)

Four explainer videos were supplied as look-and-feel references, and most of
what made them cohere was never templates. It was a house style. So this is an
axis **independent of branding and of light/dark mode**, added before the
templates that need it. The reference analysis, and the fourteen touchpoints an
eleventh reference template would need, are in
`docs/research/02-reference-visual-system.md`.

**Skins.** `style.skin` picks `default` (unchanged), `broadcast` (near-black
stage, standing chrome, large uppercase headlines, content set back in air) or
`minimal` (flat charcoal, one accent, no furniture). Every skin derives in both
polarities and right round the hue circle. `editorial` (§14) and `showroom` (§15)
came later; `showroom` is the one exception to the polarity rule and §15 says why.

They are **additive by construction**: `deriveVideoTheme` runs exactly as it
always did and a skin overrides only the tokens it disagrees with. A course that
never mentions `skin` gets a byte-identical scene graph, `omitempty` keeps the
new keys out of its JSON so no recorded config fingerprint moves, and every
pre-existing visual baseline passed with zero pixels differing.

**Semantic accents — `quantity`, `limit`, `rival` — are deliberately not
branding.** A bar that overruns its ceiling is red whatever the course is
branded with, because the colour states what the picture *means*; running the
anchor through the brand hue would make a green-branded course draw its failure
state in green. They derive for every skin, including the default.

The quantity role needed a hue rotation to survive light mode. Gold walked down
to AA on paper lands on `#8d6d0b` — khaki — so a gauge's bars, a myth's
replacement line and a rundown's lit card all read as mud rather than as the
deliberate colour of the role. Rotating toward amber lands on `#a45c09`, a burnt
orange that reads as chosen; chroma survives the drop in lightness where
yellow's does not. **The hue moves only on paper.** Found by rendering all ten
v1 templates in light mode rather than by trusting the contrast test, which
passed the whole time: 4.5:1 says a colour is readable, not that it is still the
colour you meant.

**Air is a scale, not padding.** Nearly every scene sizes against the `STAGE_W`
constant at module scope, so a fatter padding leaves a fixed-width card exactly
as wide and merely overflows the box. It arrives via `StageAirContext`, so no
scene component needed editing.

**Four backdrops, derived from what is standing on them.** One softly glowing
dot grid used to sit behind everything — a house style, and also the same shot
every time, actively fighting half its content (a dot grid behind a chart
competes with the chart's own gridlines; an even wash behind a character is the
opposite of a stage). Now: `paper` for the whiteboard (a horizontal rule and
flatter light, so the board has a top and a bottom — horizontal only, since a
full grid under a hand-drawn board reads as graph paper), `blueprint` for flow
(fine squares, a heavier line every fifth), `spotlight` for cast and story (one
pool of light, deeper vignette), `clean` for data (nothing at all). The variant
is **derived from the scene types in the video**, not passed in: a snippet is
one template start to finish so its scenes agree, while a lesson mixes title,
code, diagram, terminal and points, and mixed content keeps the neutral
backdrop. They differ by degree — glow, field, vignette and grain multipliers
over one shared canvas — so a variant reads as the same room lit differently
rather than as a second design system.

**Three transitions, and only three.** Every cut used to be the same
rise-and-dissolve. `push` goes to `illustration` and `cast`, which alternate the
figure's side every beat — moving both shots the same way reinforces what the
layout is already doing; the displacement is small on purpose, since a push that
crosses a 1920px frame reads as a slideshow control. `cut` goes to `story`,
because a rising cross-dissolve between two deliberately framed shots is the
transition a film would never use; it takes the motion language's `fast` window
and nothing moves. Everything else keeps `rise`. Derived from the scene types
the same way the backdrop is.

**Typing has a rhythm, and Go owns it** (`internal/pipeline/typing.go`).
Characters used to be evenly spaced with random jitter, which reads as a
teleprompter. People type a word in a burst, stop at the end of a line, stop
longer before the body of a block, and never type the indentation — the editor
inserts it. That is modelled as relative weights normalised to whatever window
the beat got: a newline costs 3.5 characters, a newline ending in `:` or `{`
costs 5, four spaces of indent cost almost nothing, and `)` costs almost nothing
when it is the closer an editor would have auto-inserted. Jitter is a hash of
the character index, not a PRNG, because three processes have to compute the
same answer. The scene graph carries `keystrokes` — one absolute millisecond per
character — and the renderer keeps its old estimate only as a fallback.

**It moved out of the renderer because of sound.** Every character the editor
types now makes a click, **synthesised rather than sampled** — a recording of a
keyboard is one keyboard, with a room, a mic, a licence, and an identical
waveform every time. A press is a transient: 20ms of noise under a fast attack
and exponential decay over a low sine, with deterministic per-keystroke pitch
and level. Newlines get a bigger, lower key (0.108 against 0.064) because Enter
is the beat a listener registers as punctuation. **The level is the whole
feature** — a click loud enough to sit on a −16 LUFS voice makes a clip
unwatchable, and it is exactly the sort of thing that survives review because
reviews are read rather than listened to. Peak is 0.09 of full scale, and a test
drives sixty keystrokes 1ms apart so their decays pile up, failing above 0.5.
Generated in the scene graph's finishing pass *after* the video plan, so an edit
that retimes a scene retimes its sound with it, and read back with the
pipeline's independently-written `wavDuration` parser so agreeing on the header
is evidence rather than one file believing its own arithmetic.

**One art vocabulary instead of two.** `FIGURES` is ~100 hand-built drawings
that each have parts, a staggered assembly and a continuous idle; `ICONS` is 43
flat single-stroke glyphs that do nothing. `illustration`, `cast` and `story`
drew from the first — `whiteboard` and `flow` drew from the second, putting a
motionless line drawing in a box while a hundred animated figures sat unused in
the next file. Both draw figures now, Go's vocabularies and drift tests moved
with them, and `points` keeps the glyphs because a small chip beside a phrase is
what an icon is *for*. **The fix was room, not detail:** figures are designed
against a 200-unit box and were being drawn at 58px on a 450×272 board item, so
a third of that turns a 2px bar into half a pixel. At 148px and 132px every
mechanism reads and not one figure had to change. Sixteen new figures in a
`learn` module (question, chalkboard, insight, timer, certificate, answer,
library, highlighter, signpost, foundation, progress, discussion, study, steps,
graduate, bookmark) — the vocabulary had been built for explaining systems, and
had a load balancer but no question. **117 total.**

**The whiteboard grew shapes and a visible marker.** Every item was the same
rounded rectangle — the right default and the wrong only option. Four now, each
meaning what a person at a board would mean: `box` is a component, `circle` is
an actor or a moment, `cloud` is deliberately vague (the internet, everything
else), `sticky` is an aside and gets five times the seeded tilt. The prompt caps
non-box shapes at one or two per board, since their whole value is standing out.
The fill is the outline's own path rather than a rect behind it, which matters
the moment a shape is not a rectangle. The cloud needed **rectifying** — a plain
sine scallop alternates in and out, so the inward halves cut back past the body
and every bump ends in a point: it drew a star. And the marker is a **drawn
pen** rather than a soft accent bead, because a bead says something is happening
and a pen says a person is drawing this. Drawing it exposed a real bug:
`roughSticky` sampled every edge with the same point count regardless of length,
so `penAt()` (which indexes by point) and the draw-on (which advances by arc
length) disagreed and the marker floated a corner away from its own stroke —
during the one moment the marker exists for.

**Flow nodes stopped being identical rectangles.** Kind showed only as a 4px
colour stripe, which is a legend nobody can read without being told what it
means. There is a silhouette per kind now — a store is a cylinder, a queue has
slots down its trailing edge, a client has a title bar, an external system is
dashed — drawn inside the same rect the ranking assigned, so layout is
untouched.

**The VS Code editor got a pointer and a completion list.** The scene already
timed a hover and a click on the file it opens and drew no cursor at all, so the
highlight moved on its own like a haunted menu. The pointer arrives from
below-right, presses with a single soft ring, and withdraws between the click
and the first keystroke because the hand has moved to the keyboard; it renders
*inside* the row it clicks rather than at computed window coordinates, since the
tree is a flex column and any absolute position would be a second copy of that
layout waiting to part company with it. The **completion popup** is the
strongest "somebody is coding here" signal an editor gives off; candidates come
from the file's own identifiers rather than a table of Python builtins, which is
what a real editor's word-based completion actually does and cannot go stale
against a language nobody taught it. The exact match stays in the list —
filtering it out made the popup vanish for one frame at the end of every word.
Both were invisible to every baseline (the fixtures set neither `intro` nor
`typeAtMs`, so the whole opening was switched off in one composition and
happened at negative frames in the other), which is why `VSCodeIntroViz` exists.

---

## 11. The catalog at 29 templates (2026-07-28)

Twenty-one templates were added on top of the original eight, in three waves:
two rounding out the original plan, a set written for no-code and vibe-coding
courses, and ten built against the reference clips.

**Each earns its place by the rule that distinguishes it**, not by looking
different. Where a new template overlaps an existing one, the error message
names the one to use instead.

| template | the frame | the rule that earns it |
|---|---|---|
| `anatomy` | one line of text, taken apart | a part is a **literal substring** of the subject; Go finds it and resolves rune spans, so a callout can never land on the wrong characters |
| `timeline` | a spine that fills in as it is walked | monotonic — walking back is narrating a diagram; two milestones is a before/after and `compare` does it better |
| `compare` | two columns, both in frame from the start | the `both` beat is required: describing two things separately then announcing a winner has compared nothing. A tie is a first-class answer |
| `quiz` | ask, wait, then tell | **the gap is the feature** — ≥1 beat between ask and reveal, all of it `think`, and nothing may explain an option early. Every option needs an explanation, not just the right one |
| `canvas` | app cards on a dotted grid, wired, then fired | the first card is the trigger and the only one; forward only; the last beat runs a real payload |
| `promptloop` | the vibe-coding conversation | turns strictly alternate, it ends on an answer, and ≥2 prompts — one ask and one answer is a demo, not a loop. **No code on screen**, deliberately |
| `mockup` | a screen assembling itself | built downward, ≤1 header and ≤1 footer; un-built blocks are *not* drawn (the layer list carries what is still to come) |
| `stack` | tiers of tools, and where the handoff is | the walk goes **down and only down**; every tier states what it is for; no product appears in two layers |
| `spec` | criteria written first, checked last | the beats write the criteria with nothing ticked and one closing beat checks them all. A criterion may be **missed**; a sheet where nothing was met is rejected |
| `showcase` | a tool's card, honestly | the limitations column is **enforced twice** — on the card and spoken in a beat. The one validator in the catalog defending the viewer rather than the layout |
| `breakdown` | a path whose phases open | an item beat may only spotlight something in the phase currently open; no re-opening, no walking back. **12 beats** — the case `MaxBeats` was added for |
| `metric` | one figure at a time, counting up | every number needs a unit and a label, and not everything may be `neutral` |
| `gauge` | a bar against a marked ceiling | the ceiling is set first, and nothing may exceed 4× it |
| `verdict` | a ruling and its asterisk | at least one *narrated* condition under which the call is wrong |
| `decision` | an axis of tiers | bounds ascend and the last band is open-ended, so the partition is total |
| `myth` | a belief struck through and replaced | the correction may not be a bare negation of the claim |
| `analogy` | a picture mapped part by part | nothing maps to nothing, and it must say where it breaks |
| `rundown` | a numbered promise | a promise naming N is backed by exactly N cards |
| `trace` | a system caught in the act | every step states the value after it, and the value must actually move |
| `costing` | a bill built line by line | the total equals the sum, and one line must be a cost nobody budgets |
| `constellation` | one idea and its properties | every spoke carries the relation word joining it to the centre |

**A per-template beat ceiling.** `maxSnippetBeats` was a hard 7, and
words-per-beat advice is derived from the budget divided by the beat count — so
past ~140 seconds the prompt told the model to write 75-word beats while the
validator rejected anything over 60. At 180s the writable window was 393–420
words: every beat had to land within 7% of the ceiling. The same contradiction
`beatBounds` documents at the short end, sitting unfixed at the long end,
unnoticed because nothing had asked for three minutes. Templates may now raise
their own ceiling with `MaxBeats`, and the beat count has a floor set by the
words as well as by the ideal, so the advice is never something the validator
will refuse.

**Normalize before validate.** Templates declare `Owns` / `OwnsPlan`; a plan is
normalized before it is validated, foreign payloads are stripped rather than
rejected, and a payload filed under the wrong name is migrated instead of
refused. A rule whose fix is "cut it to four words" teaches the model nothing,
so spending a correction round to say so is a round not spent on the clip. What
survives into validation is what only the model can fix — the line sits at
arithmetic-and-spelling vs. a claim: clamping a label is a repair, quietly
rewriting a winner is a different act.

**Six categories, because 29 in one alphabetical list is a wall.** The problem
is not length, it is that the list is sorted by something nobody knows when they
arrive — somebody opening the gallery is not thinking "I want the gauge
template", they are thinking "I need to show whether this fits". So the grouping
is by the **job**, not the mechanism: sorting by what is on screen (charts here,
editors there) would have been easier and useless. `gauge` and `metric` sit
together because both answer "how much"; `trace` and `flow` sit together because
both answer "how does this work", though one draws state and the other
structure.

```
Numbers & scale        how big, how much, how long — and whether it fits
Ideas & mental models  explaining, or replacing what someone believes
Systems & process      how it works, how it is built, in what order
Code & screens         anything whose subject is on a screen
Choices & verdicts     weighing options and saying what to do
Presenting & pacing    the shape of the delivery rather than the content
```

Category is **required** and `registerSnippetTemplate` panics without one: a
catalog that *can* grow an uncategorised entry does, and it lands in whatever
bucket the UI keeps for leftovers, which is where templates go to never be used
again. A test fails any category holding more than a third of the catalog.
Templates also carry `Since` ("v1" for the ten reference-clip templates) — a
fact rather than a status, so unlike a "new" badge it stays true when the next
batch lands. The gallery groups and filters (matching title and description as
well as name, so "does it fit" finds `gauge`); `/api/snippet-templates` emits in
category order, since the gallery keeps no copy of the vocabulary and arrival
order *is* heading order.

**Every description was rewritten to say when to reach for it.** They used to
say what was on screen and stop, which answers the wrong question — somebody
browsing 29 templates knows what they want to *say*. Each now states the frame
and then the occasion: *"A bar filling toward a marked ceiling. Reach for it
whenever the question is whether something fits — memory, budget, a latency
target."* Same copy in the gallery, the CLI and the caster's catalog, so the
model choosing templates reads the same guidance a person does.

**Thin briefs are enriched before planning.** The planner is good at turning a
rich brief into a clip and bad at inventing the facts a thin one leaves out, and
when it fails at the second job it does not fail gently. Enrichment reads what
the chosen template cannot be filled without — the same list the caster is
given, in the same words — and writes the fuller brief a person would have
written. It runs *before* planning rather than as a retry, because the thin
prompts that fail loudly are a fraction of the thin prompts that quietly make a
mediocre clip, and it never fails the pipeline: no rewrite means the original
prompt is used.

**A worked example outranks the prose above it.** "Plan has no beats" after
three correction rounds turned out to be the `metric` and `gauge` prompts
nesting `beats` *inside* their template object in the worked example, where
`SnippetPlan` does not read it. The model returned exactly what it was shown.
Nothing caught it — the templates' tests build plans in Go and the render tests
use fixtures, so no test had ever read a prompt's example.
`TestPromptExamplesPutBeatsAtTheTopLevel` now parses the worked example in all
29 prompts, matching by balancing braces rather than by line.

**Cross-template field guards** (`beatFields` / `rejectForeignBeatFields`) and
the shared normalizer are what keep 32 templates from becoming 32 dialects, and
**every template has a picture of itself**: 51 visual-regression baselines, all
0px-deterministic, and 29 gallery previews downscaled from them. A Go test fails
if a registered template has no preview.

---

## 12. Combos — one video cut from several templates (2026-07-29)

A snippet is one template start to finish, which is right for thirty seconds and
wrong for ten minutes: nothing holds attention through ten minutes of the same
picture. **A combo is an ordered run of segments, each with its own template,
rendered onto one timeline.**

**It is not several clips stitched together**, and that is the decision that
matters. There is one narration, one TTS pass and one alignment across the whole
piece. Stitching would give a seam at every join, loudness drifting between
segments, and a supercut rather than a video.

What makes that cheap is that alignment spans are already absolute
milliseconds. A template is handed the slice of spans covering its own beats —
already timed against the finished audio — and lays out scenes exactly as it
would in a snippet. **Assembly is slicing, not arithmetic, and no template knows
it is in a combo.** The renderer needed no change at all: `LessonVideo` has
always dispatched per scene on type.

Segments are planned separately, each through its own template's prompt. One
call for the whole combo would have been fewer round trips and worse in every
other way — each prompt carries its own vocabulary, bounds and enforced shape,
and merging them would either drop those or produce a document no model follows
to the end. Per-segment planning also means a segment that fails its validator
fails alone. `IsCombo` branches ahead of `IsSnippet` in the two stages that
differ — plan and scenegraph — and everything between them (verify, audio,
align, captions, chapters, render) is the shared path. Verify is kept whenever
*any* segment shows code.

**`combo.yaml` is the edit surface**, built for the editing that comes after the
first watch rather than retrofitted for it: segments carry stable ids generated
once and never renumbered, so an edit stays addressed to the same segment when a
neighbour moves; `skip: true` drops a segment from the cut without deleting the
prompt that produced it (the commonest edit after a first watch is "lose that
bit" and the second is "put it back"); and the file is a stage input, so editing
it re-stales exactly the stages that depend on it. Saving prunes zero-valued
keys through a generic tree — `config.Config` has no `omitempty` tags and cannot
have them, since a course manifest legitimately records an explicit zero, and a
file people are told to edit should be readable when they get there. Empty lists
survive: `segments: []` is a useful thing to see.

**Directing is the only genuinely new thinking; planning, assembly and rendering
were all reuse.** `coursesmith combo direct "<subject>" --minutes 5` is the whole
surface in one command. A creator makes four decisions — subject, length, theme,
captions — and four stages in a fixed order make the rest:

```
substance  what is actually known about this subject, and what is not
outline    what the piece ARGUES, divided into parts. Cannot see the catalog.
cast       which template holds each part. Cannot change the parts.
write      each part planned through its template's own writer, then the whole
           thing read back by the critic and the misfits re-planned
```

**The order is the design.** Casting used to do the first two jobs at once, and
that is why finished pieces contained segments that did not belong: asked to
divide a topic *and* pick looks in one breath, a model does not weigh them
equally — the catalog is concrete and eighty-one items long, the argument is
abstract and has to be invented, so the looks win and the division of the topic
comes out as a by-product of which templates sounded appealing. Now the outline
call cannot see the catalog at all, and the caster receives fixed parts and may
only choose how each is shown. Two calls that can each override the other are one
call with extra steps.

**Templates have bios** (`snippet_bio.go`). The gallery copy answers "is this the
look I want?", which is a person's question; a bio answers "can I fill this?",
which is the director's. Each declares the material it must be filled with, the
subject it is wrongly reached for, which arc roles it can carry, and whether it
needs real figures. That last flag replaced a hardcoded paragraph in the cast
prompt that named four templates by hand and was silent about the other
seventy-seven. Three rules are now checked rather than requested: a template
outside the theme's pool, a template that cannot carry its part's role, and a
data-hungry template cast over material containing no digit.

**The theme decides the pool** (`combo_pool.go`). The replica batch assumes the
broadcast stage and the foundations batch assumes the editorial left axis, so a
piece that mixes families freely changes production partway through — which is
what "that clip did not belong" usually turns out to mean. `default` and
`minimal` cast from the core catalog, `broadcast` adds replica, `editorial` adds
foundations, `showroom` adds showroom. Narrowing beats asking a model to hold a consistency rule it cannot
see.

**The critic is the only pass that sees the whole piece** (`combo_critic.go`).
Template validators check a plan against its own rules and the review gate scores
it against a rubric; both look at one clip with the others out of frame, so
neither can see a segment that repeats what segment three established,
contradicts it, or is true and does not advance the argument. One call reads
every segment's narration in order with the piece's angle in hand and returns
only what is wrong; those are re-planned through their own template with the
criticism attached. It never fails the run — a plan the critic dislikes still
renders — and it caps repairs at four, saying so out loud, because past that the
defect is the outline rather than the segments.

Rhythm is still enforced rather than requested: a template may not follow itself,
and none may appear more than three times. The catalog the caster reads is
**rendered from the live registry**, not written into the prompt file, so a
template added today is castable today.

`cast` writes `combo.yaml` and stops by default, because the cast is a structural
decision worth reading before nine planning calls are spent on it — and because
it writes exactly the file a person would have written, changing a pick is
editing one line. From the brief *"why two users buying the last item at the
same time oversells your stock"* it returned **myth → trace → breakdown →
decision → verdict** — open on the belief the viewer arrives with, show the race,
close on a ruling — and all five planned cleanly into 28 beats and 955 words.

**A miscast segment no longer kills the combo.** "Non-empty" is the only thing a
validator can check about the material field, so a look chosen because the name
fits, for a part with nothing to put in it, still gets through — `gauge` cast on
"how vibe coding lets users build by communicating ideas with AI" failed an
entire eight-segment combo. Seven good segments and one that cannot be planned is
a video with a hole, not a failed video, so a segment whose template cannot be
planned is **recast as `illustration`** — the one look with no data requirement
at all — and the run continues. The log says which and why, and `combo.yaml`
keeps the original choice. The caster prompt also now names the four templates
that cannot be planned from a subject alone (`gauge`, `metric`, `costing`,
`trace`): prevention and recovery, since neither is sufficient alone.

**The studio page is two halves.** Building a combo is choosing an *order* —
which look carries which part of the argument — so the builder is a list you add
to and reorder rather than a grid you pick one thing from; that is the real
difference from the snippets page, where the gallery is primary because a
snippet is one decision. The template picker is a grouped native select rather
than the card gallery, because here you are choosing for the fourth segment of
nine and the choice has to sit inline beside its prompt.

The editor is why the page exists rather than the CLI being enough. **Edits are
staged locally and applied together**, because every edit here moves the
narration — swapping a template re-plans the segment, dropping one takes its
words out of the read — so each costs a full rebuild. Batching turns four edits
into one run, and the banner says so before the button is pressed. Edited
segments are ringed and dropped ones dimmed. PATCH takes **pointers** for every
field, since a plain string cannot distinguish "set the prompt to empty" from
"leave the prompt alone", and it writes `combo.yaml` and stops — running is a
separate call, because only the user knows when they have finished editing.
Plan-only defaults **on** here and off for snippets: planning a combo is one call
per segment and rendering is minutes.

Look controls — dark/light, captions, skin — are lifted verbatim from the
snippets page so the two screens cannot drift, and they apply to the **whole
combo** rather than per segment: a piece that changed polarity or caption style
partway through would read as several videos stitched together, which is the one
thing a combo is built not to be.

Not yet built: a cheap props-only edit path (the `video-plan.yaml` shape).

---

## 13. Studio, pacing, and OpenAI-only (2026-07-29)

**A rail instead of a link row.** The nav was nine text links in the header. It
could not grow, and it read as nine equal things when the studio is really three
groups: what you make on, what you configure, what you inspect. Now a grouped,
icon-railed sidebar, collapsible to 16px (persisted, because it is a workspace
preference), with an off-canvas drawer below `sm`. `App.tsx` is the route table
again; the shell lives in `layout/StudioLayout.tsx`, and `store/studioStore.ts`
holds what outlives a route behind selector subscriptions so collapsing the
sidebar does not re-render a Remotion player mid-frame.

**Light mode existed and nothing could reach it.** `applyTheme` resolved every
token per mode and `preferredMode` read a stored choice; no control ever called
them, because of the ink ramp — 640 usages of a *fixed* dark scale, so a toggle
would have flipped the shell and left every page black.

So **the ramp inverts**. It is ordered by distance from the reader rather than
by lightness — 950 is the page behind everything, 100 is the brightest thing
written on it — which is what lets one `text-ink-100` be correct in both themes
instead of 640 class names needing a `dark:` variant. It is emitted as CSS
variables from the same `tokens.ts`, with the dark values as the `var()`
fallbacks in the Tailwind config so the first paint is right before React boots.
A terminal is not chrome, so log panes opt out with `.surface-dark`, written as
twelve `--ink-N: var(--ink-dark-N)` mappings against a second always-dark ramp
rather than a copy of the palette in CSS. The markdown editors deliberately do
not opt out: an editor in a light theme is light.

Measuring the ramp found two real contrast defects. `text-ink-500` — the single
most-used colour in the studio — was **2.3:1** on the page and ink-400 was
4.1:1; both are body text and both now pass. And 28 places wrote hints and empty
states at ink-600, **1.6:1**, near-invisible; those move to 500 rather than
raising the step, which would have dragged `border-ink-600` and `bg-ink-600`
with it. `ink.test.ts` asserts AA in both modes, the ramp's ordering, and that
the fallbacks match.

**Templates became a picker and a detail panel.** It was three lists of names,
which answered neither question a reader has. Now: the palette as swatches, the
motion language demonstrated *with itself* (real stagger, easing and duration
off the Go-owned tokens), and Apply, which writes the archetype to a course
through the endpoint the course editor already uses. There is no save button for
the archetype itself and there should not be — `/api/archetypes` is a GET of a
Go registry and the motion values are drift-guarded against `motion.go`, so a
slider here would be a control with nothing behind it.

**No more Groq.** The default `llm_content` is `openai/gpt-4o-mini`, the combos
course pins it explicitly the way the snippets course already did, and the
new-course scaffold no longer hands people Groq. That missing pin is why the
first real combo ran on Groq at all: snippets named a model, combos inherited
whatever the global default happened to be. A side benefit — gpt-4o-mini *is* in
the ledger's price table, so runs now cost what the ledger says instead of being
silently recorded at zero. Four router tests broke, and each was a coupling
worth removing rather than a number to update: they read their model out of
`config.Defaults()` while testing routing, caching and missing-key messages, so
sharing a model between content and review made the routing test stop proving
anything. They pin their own providers now.

**New courses ship the house voice speed.** The scaffold writes `voice_speed:
0.9` alongside `pace_wpm`, with a note that the align stage multiplies the two —
so slowing the voice moves the pace target with it instead of reading as under
pace. New coverage for the parts unit tests cannot reason about: `insertSilence`
against **real ffmpeg** (`anullsrc` as an in-graph source, concat's format
matching), `applySentencePauses` keeping audio and both timestamp tracks in
step, and `effectivePaceWPM` scaling the target with voice speed. The TTS test
servers had been decoding into `map[string]string`, so the numeric `speed` field
failed the whole body decode and nothing had ever asserted that 0.9 reaches the
server.

**Studio pages** now: Compose, Snippets, Combos, Courses (+ course/lesson
editors), Quiz editor and strategy, Templates, Library, Results gallery,
Adaptive config, Showcase, Generation, Ledger.


---

## 14. Three templates for a course, not a question (2026-07-31)

The catalog at 29 could answer almost any single question — how big, how it
works, why the obvious belief is wrong — and had nothing at all for the shape a
*course* has. Everything in it was about the subject. Nothing was about the
viewer's position in a run of lessons, nothing could draw something that comes
back round, and nothing could show two quantities more than four times apart.

Three templates, each earned by a rule rather than by a look, and each written
to carry very little text: **`chapter`**, **`cycle`**, **`scale`**. All three
derive in both polarities and have a light-mode baseline of their own, because
all three are structured almost entirely out of `accentQuantity` and light mode
is where a fixture that forgot the semantic accents would show up.

| template | category | the frame | the rule that earns it |
|---|---|---|---|
| `chapter` | Presenting & pacing | a 400px hollow ordinal, the section starting now, and the path it sits on | the marker only moves **forward**, and the clip **ends on what opens next** |
| `cycle` | Systems & process | a closed ring, stages on it, a comet running the arc | the loop must **name what changes each lap**, and the return is the last beat |
| `scale` | Numbers & scale | worlds nested inside worlds, the camera pulling back | every rung at least **4× the last**, strictly ascending |

**`chapter` is furniture, and that is what separates it from `timeline`.** Both
draw a path with stops on it. A timeline walks the milestones of a *subject*; a
chapter break is punctuation between two stretches of teaching, so the path is a
strip along the bottom and the ordinal is the loudest thing on the frame. The
numeral is drawn hollow — stage-coloured fill, accent stroke — because a solid
380px figure in the accent is the only thing anybody would see, and the section
title has to stay the thing being read.

The two enforced rules are one rule from either end. Looking back may only reach
stops already behind the viewer, and even the looking back runs forwards; and
`here` is the last beat, always. A break that closes by summarising what just
finished leaves the viewer at a stopping point, and the words that carry
somebody across a gap are the ones about what is on the other side of it. The
one concession: at the first stop no look-back is demanded, because there is
nothing behind and requiring one is asking the model to invent it.

**`cycle` exists because `flow` cannot draw a ring at all** — its validator
*requires* a fork or a join, so a loop drawn there becomes a chain, which states
the one thing that is not true about a loop. The rule that earns it is
`changes`: a ring whose second pass is identical to its first is a wheel
spinning, and the template refuses to be planned without naming what each lap
leaves behind. That line lands in the hub on the closing beat, inside a ring
that has just closed.

Two implementation notes worth keeping. The traversed arc is emitted in
**quarter-circle pieces** rather than as one `A` command: a single arc cannot
express a full circle — its start and end points coincide, so nothing is drawn —
and at the last beat that path *is* a full circle. The first version used the
large-arc flag and hung a loop off the top of the ring on exactly the frame the
template exists to deliver. And the comet is drawn on **its own layer above the
stages**, because the discs had to be made opaque (the 10px lit arc otherwise
runs straight through the icon) and the one moving object on the stage must
never end up behind a standing one. It also parks a shade short of each stage,
so the light waits at the disc's edge rather than under it.

**`scale` picks up exactly where `gauge` gives up.** A gauge caps at 4× its
ceiling and says so in its own error, because past that the line is a hairline
and the picture states nothing — but the interesting quantities in teaching are
almost never within 4× of each other. So this frame does not compare lengths at
all: it nests, and the camera pulls back a level at a time. The geometry is a
fixed 3.4× and **not** the true ratio, which is deliberate and stated: drawn
proportionally a thousand-fold rung puts the last world at a third of a pixel.
The eye gets the containment, the rail gets the numbers, and `gauge`'s rejection
message now names this template by name.

The closing beat **compresses instead of pulling back further**, which was the
one genuinely wrong frame in the first pass: at 3.4× a four-rung ladder renders
"all four at once" as one box with a speck in it. The camera holds on the
largest rung and the spacing tightens to 1.9×, so every world is legible, and
the frame carries the end-to-end span (`×40 billion`) — the figure the clip is
actually about and the one no single rung carries.

**Sixteen new icons**, because the vocabulary of 44 was built for architecture
diagrams: it had a load balancer and nothing for "you are three parts into a
course". Added in three groups — a journey and its landmarks (`compass`, `map`,
`signpost`, `milestone`, `trophy`, `graduate`), things that come back round
(`orbit`, `refresh`, `recycle`, `sprout`, `infinity`), and things at a size
(`ruler`, `mountain`, `atom`, `telescope`, `city`). Both sides of the mirror,
guarded by `TestIconVocabularyInSync`.

**All 53 baselines were re-recorded, and 44 of them had been stale since
`2ef06f5`.** That commit made Stage's bottom reserve conditional on captions,
which moves every centred composition up by 28 pixels, and the baselines were
never re-taken — so `visual-regression` had been red against templates nobody
had touched. The drift was verified as a pure vertical translation before
re-recording rather than assumed. They are 0px-deterministic again, and the
three new templates contribute six of them: a beat that hands over and a beat
that looks back for `chapter`, the closed ring and a mid-walk for `cycle`, one
rung and the compressed ladder for `scale`, plus a light-mode twin each.

**The animation gate was red for the same reason, and it is a better test now.**
`test/animation_timing.mjs` asserts a reveal builds and then settles, and it
reads total ink to do it — a complete description of the frame until `d64f816`
gave every scene a camera that creeps closer for eighteen seconds. Content keeps
growing after the last node lands, so "nothing more appears after the reveal
settles" went red on a scene nobody had touched. Measured on D3Viz between the
settle frame and eighty frames later: camera running, ink +3.35% and 4% of
pixels moved; camera parked, ink −0.01% and **zero** pixels different. The
reveal was correct the whole time.

So the four reveal checks render with the camera parked, and a fifth asserts the
camera is still alive — the same late frame with it running must differ from the
parked one (measured 3.7%; the floor is 1%). That check is not padding: the
camera's own commit records that its first implementation produced no motion at
all, and parking the camera to fix the first four would have removed the only
frames in the suite that would ever notice. Verified by zeroing the token and
watching it fail. One trap worth recording — input props must go to
`selectComposition` **and** `renderStill`; passed to the latter alone they are
silently dropped, and the parked frames come back byte-identical to the live
ones, which reads as a dead camera rather than as a dropped override.

## 15. The showroom — a light skin, and cards that wear real logos (2026-08-17)

`cards` shipped in v8 as the only stage in the pipeline that fetches something off
the open web: a brand mark from Simple Icons, so a row of products could be
identified rather than read. It came out looking like the catalog and not like the
thing it was copying, and the reason was one decision made in the wrong direction.
`svgPathData` took the geometry out of the fetched document and **threw the brand
colour away** — the CDN serves `fill="#D97757"` right there in the file — so the
mark could be repainted in the course accent and obey the theme like everything
else on the frame. Tidy, and it discarded the only thing the fetch was for. Half
of what a viewer recognises about a logo is its colour. Gemini's mark in gold is a
four-pointed star; in `#8E75B2` it is Gemini.

Keeping the colour is only possible on a light ground, which is why this is a skin
before it is three templates.

**`showroom` is the fourth skin and the only one that overrides the mode axis.**
The other three are *treatments* — how much light is on the backdrop, how loud the
type is, whether there is standing chrome — and a treatment is orthogonal to
polarity, so each owes a dark version and a paper version. This one is a specific
published look: a neutral near-white ground with pure-white cards seated by cast
shadow. Every part of that *is* the paper; a dark "showroom" would be one of the
three that already exist. Neutral rather than hue-tinted, because a tint at 96%
lightness is a visible cast and a cast under three cards each wearing its own
brand colour is a fourth colour arguing with three.

**Elevation became three tokens, and that is the load-bearing part.** "Lift this
off the surface" is not one effect, it is two opposite ones, and every scene in
the catalog had been picking the dark answer with a literal. On near-black a cast
shadow does almost nothing — black on near-black is black — and what seats a card
is the *rim*, the one-pixel highlight along its top edge. On paper the rim does
nothing instead (white on white) and the shadow does all of it. So `shadow`,
`shadowStrength` and `rim` derive per mode (`deriveElevation`) and `seat()` in
`theme/theme.ts` emits all three layers every time, letting the tokens turn off
the ones that do not apply. **No polarity check anywhere in a scene.**

Three things the light frame taught that the dark stage never could:

- **A dim card must keep its full opacity.** On the stage, fading a card recedes
  it, because what fades is a surface lighter than the ground. On paper the card is
  the brightest thing in the frame, so fading it fades it *into the page* — at 0.62
  a row of three read as three cards that had failed to load. What recedes a card
  here is its contents going quiet and its rim light going out.
- **The glow has to be a hard ring, not a blur.** A soft outer halo on white is a
  grubby edge; `0 0 0 5px` at 0.18 alpha reads as selection.
- **Less air than the dark skins, not more.** 0.07 was tried first. Insetting a
  small diagram on the broadcast stage buys composition because the surrounding
  near-black is doing nothing either way; on paper the surround is a bright sheet,
  and pulling the cards back from it makes them look small on a large empty page.
  0.03, and `sheet` — a new surface with the accent glows at zero, because at 0.55
  they are invisible on a dark stage and the brightest thing in the frame on paper.

**`cards` was rewritten around one rule: everything in a card is in the card.** It
used to draw a logo and a name and float the note under the whole row in a shared
box, which made a card two things, and two things is a sticker — the exact failure
its own file header forbids. Measured against the reference, no card there ever has
fewer than three: a mark, a name, and something carrying information. The new third
element is the note, and `ask` is what makes it worth waiting for — one optional
label on the *spec* rather than per card, set small-caps on every card above a slot
reading `? ? ?` until that card's beat. It costs no layout, because the slot is the
card's shape from frame one. The slot's height is computed from the longest note in
the row so the tallest state is the only state; sized to its contents it would make
the row's height a function of which card is lit.

**Two new templates, because one card is not a family.** `duel` is the two-up the
reference frame actually is: two products, a pill each, and one bar each against a
*shared* track — normalise them independently and both run full width and say
nothing. Its hard rule is that the two scores must differ by at least 12, and the
error says what to do instead: if they are genuinely level on this axis, that is
not the axis the choice turns on, and `versus` compares across five dimensions.
The pick is **allowed to be the shorter bar** and that is deliberately unvalidated,
because "the free tier is worse and is still right for most people" is the most
useful thing the template can say. `spotlight` is the asymmetric one — a hero card
left, claims landing one at a time right — and it is the fifteen-second counterpart
to `showcase`'s seventy. Its rows *land* rather than ghost-then-brighten, which is
the one place the family departs from the row's "no reflow" discipline: a count is
worth showing in advance, a claim is not.

Catalog 45 templates. `combo_pool.go` gains `showroom: {core, showroom}`, the
studio gains a fifth gallery, and one correction landed with them — the hero card
was `selected` in the first draft, and a brand-coloured rim around the only card on
screen has nothing to compare against, so in peach it read as a warning. It gets
the deeper seating instead: elevation rather than colour.
