# coursesmith

An AI course production engine. Point it at a markdown lesson outline and it
compiles a broadcast-quality video lesson with **zero manual recording** —
animated title cards, self-typing code scenes, real executed terminal demos,
diagrams that appear on the exact spoken word, karaoke captions, an
interactive quiz, and a Hugo web page — at near-zero cost using free-tier
APIs and open-source tools.

```
lesson.md ─▶ script ─▶ verify ─▶ review ─▶ visuals ─▶ quiz ─▶ demos ─▶ audio ─▶ align ─▶ captions ─▶ scenegraph ─▶ render ─▶ hugo
             (Groq)   (docker    (LLM      (SVG +     (MCQ,   (VHS     (Kokoro  (whisperX (word-level  (lesson-      (Remotion  (page
                       executes   quality   vision QA) exec-   real     TTS)     word      WebVTT)      video.json)   1080p30)   bundle)
                       all code)  gate)                checked) terminal)         timing)
```

Every stage is idempotent and resumable: outputs land in the lesson's
`generated/` directory, input hashes are recorded in `state.json`, and
unchanged stages are skipped on re-runs. Every LLM response is cached on
disk, so re-running a pipeline costs **zero API calls** unless something
actually changed.

**The quality moat:** every Python code block is executed for real in a
Docker sandbox — the course never shows output Python didn't actually
produce, broken code fails the build, and terminal demos are genuine
recordings of python3 running (faking output with `echo` is rejected at
generation time).

## Requirements

| What | Why | Install |
| --- | --- | --- |
| Go 1.22+ | build the CLI | <https://go.dev/dl/> |
| ffmpeg | audio processing + probing | `brew install ffmpeg` / `apt install ffmpeg` |
| Docker | code verification sandbox + VHS demos | <https://docker.com> then `docker build -t coursesmith-sandbox sandbox/` |
| Node 18+ | Remotion video rendering | <https://nodejs.org> then `cd renderer && npm install` |
| Kokoro TTS server | local, free voiceover synthesis | see below |
| Groq API key (free) | scripts, quizzes, diagrams, tapes | <https://console.groq.com/keys> |
| OpenAI API key | quality + vision reviews (gpt-4o-mini) | <https://platform.openai.com/api-keys> |
| whisperX (optional) | word-level caption/scene sync | `cd tools/align && uv sync` |
| Hugo (optional) | build the course website | `brew install hugo` |

Missing optional pieces degrade gracefully: no whisperX → segment-level
timing estimates (with a warning); no Node → ffmpeg slide-show fallback
video; no Docker → code runs unsandboxed on the host (warned) and demos
require a host `vhs` install.

## Setup

### 1. Build

```sh
git clone <this repo> && cd coursesmith
make build            # → bin/coursesmith
```

### 2. Set API keys

```sh
export GROQ_API_KEY=gsk_...      # free tier: 30 req/min, ~1k req/day
export OPENAI_API_KEY=sk-...     # only gpt-4o-mini is used
```

### 3. Start the Kokoro TTS server

The easiest way (CPU, no GPU needed):

```sh
docker run -p 8880:8880 ghcr.io/remsky/kokoro-fastapi-cpu:latest
```

Alternative without Docker, via the `kokoro-onnx` Python package behind any
OpenAI-compatible `/v1/audio/speech` server (e.g. [Kokoro-FastAPI]
run from source):

```sh
git clone https://github.com/remsky/Kokoro-FastAPI && cd Kokoro-FastAPI
uv run --extra cpu python -m uvicorn api.src.main:app --port 8880
```

If your server runs elsewhere, point the engine at it:

```sh
export KOKORO_URL=http://otherhost:8880/v1
```

[Kokoro-FastAPI]: https://github.com/remsky/Kokoro-FastAPI

### 4. Check your setup

```sh
./bin/coursesmith doctor
```

Every failed check prints the exact command that fixes it.

## Make your first lesson

A ready-to-run example course ships in `courses/python-basics/`. Run its
first lesson end to end:

```sh
./bin/coursesmith run python-basics/01
./bin/coursesmith status python-basics
```

Outputs land in
`courses/python-basics/lessons/01-what-is-python/generated/`:

```
script.json        narration script with diagram/demo cues
verification.json  every code block's real, sandbox-executed output
reviews/           critic scores, visual QA verdicts, TTS accuracy (WER)
diagrams/*.svg     standalone branded SVG diagrams
diagrams/attempts/ every generation attempt + screenshot, for audit
quiz.json          interactive multiple-choice quiz (answers exec-checked)
demos/*.mp4        VHS terminal recordings of real python3 sessions
demos/manifest.json  demo ids, paths, durations for the scene graph
audio/*.wav        per-section synthesized narration
voiceover.wav      full lesson voiceover (long silences compressed out)
alignment.json     word-level timestamps + per-section spans and WER
captions.vtt       WebVTT captions grouped from the word alignment
lesson-video.json  the scene graph — full render input for Remotion
final.mp4          the rendered lesson video (1920x1080, 30fps)
state.json         pipeline state (which stages are done, input hashes)
```

Or scaffold a brand-new course:

```sh
./bin/coursesmith init "My Course"
$EDITOR courses/my-course/course.yaml
$EDITOR courses/my-course/lessons/01-welcome/lesson.md
./bin/coursesmith run my-course
```

### The render engine

The `render` stage compiles `lesson-video.json` (the scene graph) with
Remotion into `final.mp4`: an animated title card with staggered learning
outcomes, code that types itself character-by-character with realistic
jitter and Shiki syntax highlighting, diagrams whose SVG groups reveal
sequentially, VHS terminal demos framed in a styled window, callout
arrows/circles landing on the exact spoken word, and karaoke captions with
the current word emphasized.

```sh
coursesmith preview python-basics/01   # open the lesson in Remotion Studio
                                       # for hot-reload visual editing
coursesmith run python-basics/01 --concurrency 4   # parallel render tabs
```

No Node installed? The render stage falls back to the v1 ffmpeg slide
assembly. A screen recording dropped at `lessons/<id>/recording.mp4` is
used by that fallback path.

### Publish the website

The `hugo` stage emits each lesson into `site/content/` as a page bundle
(video player with captions, inline diagrams, interactive quiz). Preview:

```sh
hugo server -s site        # http://localhost:1313
hugo -s site --minify      # production build → site/public/
```

## Writing lessons

`lessons/<nn-slug>/lesson.md` is YAML front-matter plus a markdown outline:

```markdown
---
title: What is Python?
outcomes:                   # shown on the animated title card
  - Run your first line of Python
diagrams:
  - id: memory-model
    prompt: "3 variables in memory: x=5, name='Alice', flag=True ..."
callouts:                   # overlays anchored to spoken phrases
  - section: your-first-line
    shape: circle           # or arrow
    x: 0.5                  # fraction of frame width
    y: 0.4
    label: quotes matter!
    at: "inside the quotes" # appears when the narrator says this
style:                      # optional per-lesson overrides
  tone: extra playful
---

## A language for talking to computers
- key point one

[DIAGRAM: memory-model]

## Your first line

```python
print("Hello, world!")
```

```output
Hello, world!
```

[DEMO: open the Python REPL and print a greeting]
```

- Each `##` heading becomes a narrated video section.
- `[DIAGRAM: id]` markers place diagrams declared in front-matter.
- `[DEMO: description]` markers become real VHS terminal recordings —
  the LLM writes a `.tape` script that actually runs python3 in the sandbox.
- ` ```python ` blocks become self-typing code scenes; an adjacent
  ` ```output ` block is a *claim* that the verify stage checks against
  what Python actually prints (and corrects if wrong).
- Section ids are the heading slugs (`## Your first line` →
  `your-first-line`), used by callouts and the scene graph.

## Configuration

Settings merge in this order (later wins, field by field):
**defaults < course.yaml < lesson front-matter < CLI flags**.

```yaml
# course.yaml
name: "Python Basics"
slug: python-basics
description: ...

style:
  voice: af_heart              # Kokoro voice id — or a weighted blend:
                               #   voice: af_bella(2)+af_sky(1)
  voice_speed: 1.0             # speaking-rate multiplier (0.5–2.0)
  tone: warm, encouraging teacher
  pace_wpm: 145                # auto-pace: when the rendered lesson misses
                               # this ±15%, align writes tts_speed.json and
                               # the next run re-synthesizes at the corrected
                               # rate (bounded 0.75–1.35)
  audience: absolute beginners
  language: en

branding:
  colors: {primary: "#306998", accent: "#ffd43b", background: "#ffffff"}
  fonts:                       # video type stack (Google Fonts bundled by
    display: Space Grotesk     # the renderer: Space Grotesk, Inter,
    body: Inter                # JetBrains Mono, Sora, IBM Plex Sans)
    mono: JetBrains Mono
  diagram_style: clean, flat, rounded corners

pipeline:
  llm_content: groq/llama-3.3-70b-versatile   # provider/model
  llm_review: openai/gpt-4o-mini
  review_threshold: 8                         # 1-10 quality gate
  captions_model: whisper-large-v3

audio:                       # voiceover post-production (all optional)
  section_pause_ms: 700      # silence between sections
  paragraph_pause_ms: 350    # silence between narration paragraphs
  crossfade_ms: 50           # fade length at every audio join
  target_lufs: -16           # two-pass loudness normalization target
  music_bed: false           # mix assets/music/*.mp3 under the voice
  music_duck_db: -18         # how far the bed sits below the voice
```

`style.pronunciations` extends the built-in speech dictionary — written
form → spoken form, applied to narration before TTS:

```yaml
style:
  pronunciations:
    Groq: "grock"
    NumPy: "num pie"     # overrides/extends ~40 built-in Python terms
```

## CLI reference

```
coursesmith init <course-name>          scaffold a new course
coursesmith run <course>[/<lesson>]     run the pipeline (resumable)
    --stage <name>                      run one stage only
    --force                             re-run even if inputs are unchanged
    --concurrency <n>                   parallel Remotion render tabs
coursesmith preview <course>/<lesson>   open the lesson in Remotion Studio
coursesmith status <course>             lesson × stage table (done/stale/pending)
coursesmith doctor                      check ffmpeg, docker, node, Kokoro,
                                        whisperX, keys, templates
coursesmith serve [--addr host:port]    JSON API of project state (studio stub)
coursesmith audition <course>           render a sample paragraph in every
                                        matching Kokoro voice + HTML player page
    --choose <voice>                    write the picked voice to course.yaml
coursesmith analyze <course>            concept graph (concepts.json + .svg),
                                        dependency violations (= error),
                                        terminology drift, narrative bridges
coursesmith export-review <course>      one markdown doc per lesson for a
                                        human SME (script + diagrams + quiz +
                                        mistakes + exercises + all QA flags)
coursesmith compile-course <course>     join every rendered lesson into one
                                        course.mp4 + YouTube chapter file
```

Stages, in order: `script verify trace review storyboard visuals quiz
quiz-strategy mistakes exercises demos audio align captions chapters
scenegraph render hugo`.

## The video design system

Lesson videos render on a dark editorial canvas derived from the course's
brand colours (`internal/pipeline/videotheme.go`): a hue-tinted gradient,
drifting accent glows, film grain, Space Grotesk display type. Scene types:

- **title** — animated intro/heading cards with icon-chip outcomes.
- **points** — the storyboard stage's visual beats: keyword phrases with
  icons that pop in on the exact spoken word. Sections without any declared
  visual get these automatically — no more heading-only dead screens.
- **code** — self-typing editor window (chrome, line numbers, dark-plus
  tokens, verified-output drawer); with a trace it becomes the execution viz
  or memory layout.
- **walkthrough** — a synthesized VS Code editor (activity bar, file tree,
  tabs, minimap, status bar). Any outline section with **two or more** Python
  blocks becomes one: each block is a timed step; step 1 types itself, later
  steps flash the changed lines.
- **diagram** — `svg`, `d3`, `d2` (D2 language compiled in-process, sketch
  aesthetic), `mermaid`, or `excalidraw` kinds, all vision-QA gated, all
  generated with the narration of their cueing section in the prompt so the
  picture matches what's being said.
- Captions are page-based karaoke (3–5 words, spring pop on the active word)
  with LLM-marked keywords held in the accent colour.

### Editing the generated video: video-plan.yaml

Drop `video-plan.yaml` next to a `lesson.md` to retarget scenes without
touching the pipeline (editing it re-stales the scenegraph stage):

```yaml
edits:
  - scene: 2                 # index in lesson-video.json's scenes[]
    template: grid           # swap the template variant (points: rows|grid)
  - scene: 4
    props: {title: "A better on-screen title"}
  - scene: 5
    skip: true               # drop it; the previous scene covers its span
```

Archetypes pick default template variants per scene type (the `playful`
animation style renders points as a card grid, for example).

### Other voices (cloned voice included)

Any OpenAI-compatible `/v1/audio/speech` server works — set `TTS_URL`
(alias: `KOKORO_URL`). For your own cloned voice, run
[Chatterbox Turbo](https://github.com/travisvn/chatterbox-tts-api) (MIT,
Apple-Silicon MPS supported, clones from a 5–20s sample, `[laugh]`/`[sigh]`
emotion tags) and point `TTS_URL` at it; whisperX alignment keeps working
unchanged.

## Word-level sync (whisperX)

The `align` stage produces `alignment.json`: a timestamp for every spoken
word, plus per-section spans. This is what lets diagrams appear on the
exact word that references them, callouts land on spoken phrases, and
captions highlight word-by-word. Install once:

```sh
cd tools/align && uv sync     # CPU int8; no GPU needed
```

The stage also compresses awkward silences (>1.5s between words) out of the
voiceover, and writes a TTS accuracy report
(`reviews/tts_accuracy.json`): sections where the aligned transcript
diverges from the script by more than 5% word error rate are flagged as
likely Kokoro misreads. Without whisperX, Groq Whisper segments are used
with estimated word timing (quality warning printed; set
`COURSESMITH_ALIGN` to point at a custom aligner command).

## How costs stay near zero

- **Disk cache** — every LLM response is stored in `.coursesmith/cache/`
  keyed by a hash of the full request. Unchanged inputs → cache hit → zero
  API calls, zero rate-limit tokens.
- **Rate limiting** — a token bucket per provider (Groq: 28/min, 950/day —
  under the free-tier 30/min, ~1000/day) with state persisted across
  restarts, so an interrupted run can never blow the daily budget.
- **Quota-aware stops** — when a daily budget runs out mid-course, the run
  stops cleanly and tells you when to re-run; completed work is cached and
  skipped on resume.
- **Consistent system prompts** — per-course system prompts stay
  byte-identical across calls so Groq's prompt caching kicks in.
- **The only paid tokens** are gpt-4o-mini reviews (fractions of a cent per
  lesson).

## The quality gates

**Execution verification (hard gate).** The `verify` stage runs every
Python block from lesson.md and the script in the Docker sandbox
(python3.12, 5s timeout, no network). Real stdout is recorded in
`verification.json` and injected into review prompts as ground truth;
claimed ` ```output ` blocks that don't match reality are corrected. A
block that errors **fails the pipeline** — a lesson with broken code cannot
be published. Unchanged blocks (by hash) skip re-execution. Quiz questions
containing code are executed too: if the code prints one of the options,
`answer_index` must point at it.

**LLM review (soft gate).** Every generated artifact (script, quiz) is
scored 1-10 by the review model on technical accuracy, clarity, engagement,
and pacing. Below `review_threshold`, the generator re-runs with the
critique injected (max 2 retries) and the best-scoring draft wins. All
reviews persist to `generated/reviews/`.

**Visual QA (soft gate).** Each diagram is screenshotted in headless
Chromium and shown to the vision model with a checklist: overlapping text,
clipped elements, unreadable contrast, misleading layout. Issues trigger
regeneration with the critique (max 2 rounds); every attempt (SVG + PNG) is
kept in `generated/diagrams/attempts/`. Three exemplar SVGs in
`prompts/diagram_style/` are injected as few-shot references so all
diagrams share one visual language — replace them to restyle a whole
course.

**Real-execution linting.** Generated VHS tapes are rejected if they use
`echo` or never run `python3`; `vhs validate` failures get one retry with
the error. Engine-owned settings (dimensions, fonts, theme colors, typing
speed) are injected by the engine, never written by the LLM.

**Three-pass script review.** The `review` stage runs three sequential
passes: *accuracy* (factual claims extracted, code-checkable ones executed
in the sandbox, the rest judged with citations to Python docs terminology),
*pedagogy* (concept ordering, cognitive load ≈1 new concept per 90s,
concrete-before-abstract, worked examples), and *tone* (matches
`style.tone`, no condescension, no filler). The weighted score (accuracy
50%, pedagogy 35%, tone 15%) gates against `review_threshold`; failing
drafts regenerate with all three critiques (max 2 rounds). Records land in
`generated/reviews/script-multipass-round-N.json`.

**Quiz QA.** Every quiz needs ≥1 recall, application, debugging, and
prediction question (tagged in `quiz.json`); prediction answers are checked
against real execution. Distractors are scored as misconceptions — weak
ones trigger one regeneration — and the review model role-plays 10 cold
students per question: >90% simulated success flags `too_easy`, <30%
`too_hard` (`reviews/quiz_distractors.json`, `reviews/quiz_difficulty.json`).
From lesson 4 on, quizzes include 1-2 spaced-repetition review questions
drawn from earlier lessons' `concepts.json`.

**Mistakes & exercises (hard gates).** `mistakes.json` documents the top 3
beginner errors with the *actual traceback* from running the broken code in
the sandbox — code that doesn't fail is rejected. `exercises/` ships two
practice exercises (starter + hidden pytest tests + solution); the solution
must pass the tests and the starter must fail them, both proven by
execution.

## Broadcast-quality audio

The `audio` stage rewrites narration for speech before synthesis
(`tts_prep.go`): code tokens get spoken forms ("`__init__`" → "dunder
init", "PyPI" → "pie pee eye"), numbers and operators become words ("3.11"
→ "three point eleven", "!=" → "not equal to"), 30+-word sentences split at
clause boundaries, and `*emphasis*` spans become micro-pauses. What was
actually sent to the TTS persists as `tts_script.json`.

Sections synthesize per paragraph and join with configurable pauses
(700 ms between sections, 350 ms between paragraphs) and 50 ms crossfades —
no audible joins. The joined voiceover runs through a documented mastering
chain (`highpass 70 Hz → de-esser → gentle 2.5:1 compression`) and two-pass
`loudnorm` to −16 LUFS / −1.5 dBTP; before/after stats persist to
`reviews/loudness.json`. An optional CC0 music bed
(`courses/<slug>/assets/music/*.mp3`, `audio.music_bed: true`) is ducked
under the voice via sidechain compression with fades at the lesson bounds.

The `align` stage judges WER against BOTH the written narration and the
spoken text (whisper inverse-normalizes speech, so either alone produces
false misreads). Sections >5% WER list their exact misreads; misreads the
pronunciation dictionary knows get auto-fixed via `tts_fixes.json`, which
re-stales the audio stage. `reviews/pace.json` flags sections outside
`pace_wpm` ±15%. The `chapters` stage emits `chapters.json`, a
YouTube-format `chapters.txt`, and a timestamped `transcript.md`.

## Human review loop

`coursesmith export-review <course>` writes one markdown document per
lesson (script, inline diagrams, quiz with answers, mistakes with real
tracebacks, exercises, every QA flag). The reviewer answers in
`courses/<slug>/review-notes.yaml` (lesson → section → note); the next
pipeline run injects unresolved notes into script generation and marks them
`resolved: true`.

## Prompt tuning

Prompts are Go `text/template` files in `prompts/`, not hardcoded:

```
script.tmpl            outline → narration script
review_rubric.tmpl     the critic's scoring rubric
diagram_svg.tmpl       diagram spec → standalone SVG
diagram_visual_qa.tmpl the vision inspector's checklist
quiz.tmpl              outline + narration → quiz
demo_tape.tmpl         [DEMO] marker → VHS tape body
diagram_style/*.svg    few-shot style exemplars for all diagrams
```

Editing a template (or an exemplar) automatically marks dependent stages
stale.

## Development

```sh
make build   # bin/coursesmith
make test    # go test ./...
make lint    # golangci-lint run
```

Layout: `cmd/coursesmith` (CLI) · `internal/llm` (providers, router, rate
limiter, cache) · `internal/pipeline` (the 8 stages) · `internal/project`
(course/lesson/state parsing) · `internal/config` (layered config) ·
`prompts/` (templates) · `site/` (Hugo skeleton + course theme).

Phase 3 will add a TypeScript/React studio UI on top of `coursesmith serve`.

## Troubleshooting

| Symptom | Fix |
| --- | --- |
| `GROQ_API_KEY is not set` | `export GROQ_API_KEY=...` (free at console.groq.com/keys) |
| `cannot reach the Kokoro TTS server` | start the docker container, or set `KOKORO_URL` |
| `ffmpeg not found` | `brew install ffmpeg` / `apt install ffmpeg` |
| `sandbox image missing` | `docker build -t coursesmith-sandbox sandbox/` |
| `code block(s) failed to execute` | the lesson's Python is broken — fix it; the report names each block |
| `renderer dependencies missing` | `cd renderer && npm install` |
| `whisperX aligner failed` | `cd tools/align && uv sync`; falls back to Groq segments meanwhile |
| WER warnings in `tts_accuracy.json` | Kokoro misread those sections — rephrase tricky words in the outline |
| `daily request budget exhausted` | wait for the printed reset time; re-run — everything done is cached |
| stage keeps re-running | its inputs changed — check `status`; `state.json` records the hashes |
| review scores stuck below threshold | read `generated/reviews/*.json`, improve the outline, re-run `--force` |
