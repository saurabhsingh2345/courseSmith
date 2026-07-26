# What We Have — courseSmith

_A living snapshot of the project: what it is, everything that's built, what's
left, and where it's going. Last updated 2026-07-18 (evening: the 10x visual
overhaul — see §8)._

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
- Pages: Courses, Course detail, Lesson detail, Quiz editor, Ledger.
- Live run control over SSE (`/api/events`, `/api/run` POST/DELETE), feedback
  and regenerate endpoints, quiz-override editor, artifact serving, OpenAPI
  schema → typed client (`schema.d.ts`).

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
- **LLM layer** (`internal/llm/`): provider router (Groq + OpenAI-compatible),
  disk cache, per-provider token-bucket rate limiter w/ persisted state,
  quota-aware clean stops, retry logic, transcription.
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
```

---

## 5. What's left / rough edges

- **API keys not wired into this environment.** `GROQ_API_KEY` /
  `OPENAI_API_KEY` are not exported in the shell, so LLM-dependent stages and
  commands can't be exercised end-to-end here without the user exporting them.
- **Pace vs. voice mismatch.** Kokoro `af_heart` speaks ~160–195 wpm while the
  course targets 145 wpm, so the pace report flags several sections. It's
  informational today — no automatic tempo adjustment yet.
- **Only two lessons authored** in the sample course (`python-basics/01`,
  `/02`). Full multi-lesson runs (spaced repetition, bridges, analyze at scale)
  are lightly exercised.
- **Single content provider proven.** Router supports OpenAI-compatible
  providers but the content path is validated mainly against Groq
  llama-3.3-70b + gpt-4o-mini reviews.
- **Not a git repo.** No version history/CI beyond the site-deploy workflow;
  no release tagging.
- **Studio is early** (v0.1.0). Core run/monitor/quiz-edit flows exist; deeper
  editing (script/diagram authoring in-browser) is not there yet.
- **Graceful-degradation paths are less tested** than the happy path (no Node →
  ffmpeg fallback; no whisperX → segment timing; no Docker → unsandboxed host
  exec).

---

## 6. Where this is going

Natural next moves, roughly in priority order:

1. **Adaptive pacing** — let the audio stage retime narration (or nudge the
   voice) so `pace_wpm` is actually hit, closing the pace-report loop instead
   of just reporting it.
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
6. **Repo hygiene** — `git init`, CI for `make test` / `make lint`, and tagged
   releases so the pipeline itself is reproducible.
7. **Broader subjects** — the sandbox and verify gates are Python-first;
   generalizing execution verification to other languages would open up the
   engine beyond Python courses.

---

## 7. Repo map

```
cmd/coursesmith/     CLI commands (run, status, doctor, serve, analyze, …)
internal/
  pipeline/          the 15 stages + render/audio/align/quiz/ebook/bundle
  llm/               providers, router, rate limiter, cache, transcribe
  project/           course/lesson/state parsing, StageOrder
  config/            layered config
  studio/            Go JSON API + SSE + ledger + artifacts
prompts/             *.tmpl generation prompts + diagram_style exemplars
renderer/            Remotion (Node/React) video engine
studio/              React + Vite + Tailwind studio UI
site/                Hugo skeleton + course theme
sandbox/             Docker image for code verification + VHS demos
tools/align/         whisperX venv (word-level timing)
courses/             python-basics sample course
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
