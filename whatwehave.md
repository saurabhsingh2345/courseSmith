# What We Have — courseSmith

_A living snapshot of the project: what it is, everything that's built, what's
left, and where it's going. Last updated 2026-07-27 (snippets: the short-form
surface and its template catalog — see §9)._

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

**Next template** (planned, not built): data & maps — world-atlas TopoJSON +
d3-geo + Observable Plot.
