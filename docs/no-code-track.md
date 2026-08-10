# The no-code track — real footage, automated

_Plan, 2026-07-31. How courseSmith produces a world-class no-code / vibe-coding
course without hand-editing video, and what has to be built to get there._

---

## 1. The problem in one sentence

Everything courseSmith renders today is **drawn**. A no-code course is about
**tools that exist on other people's screens**, and a viewer who has never seen
Lovable's prompt box will not believe a rectangle we drew that says "Lovable".

So the deliverable is not "add a no-code course". It is: **make real captured
footage a first-class asset in a pipeline built entirely on synthetic frames**,
without giving up idempotency, resumability, or the provenance discipline that
makes the rest of the engine trustworthy.

## 2. The moat translates — verify:python :: capture:tools

The Python course's quality moat is that **every code block really ran**. The
sandbox executes it, and output Python didn't produce cannot reach the screen.

A no-code course has no Python. Its moat is the same shape with a different
engine: **the tool really did that**. A clip claiming to show Lovable building
an app carries the URL it was captured from, the script that drove it, and the
timestamp — recorded by the driver, never written by a model.

That is the whole design. `capture` is to this course what `verify` is to
python-basics: a hard gate that turns "we say so" into "here is the recording".

## 3. Footage is not one problem — it is four

Lumping "screen recordings" together is why this looks impossible. Split by how
the pixels are produced and three of the four are already tractable:

| class | examples | automation | status |
|---|---|---|---|
| **A. Terminal** | Claude Code, `gh`, `supabase`, `vercel`, `npm` | VHS tapes | **phase 1** — `demos` stage, most of the way there |
| **B. Web UI** | Lovable, Bolt, v0, Supabase, Vercel, GitHub, Figma (web) | drive with `rod`, record via CDP screencast → ffmpeg | phase 2 |
| **C. Native desktop** | Cursor, Figma desktop | `ffmpeg -f avfoundation` + AppleScript keystrokes; human presses go | semi-auto, phase 3 |
| **D. Recreated** | the `vscode`, `promptloop`, `canvas`, `mockup` templates | already fully automated | **already built** |

Class A is the sleeper: **Claude Code is a CLI**, so a real agent session
editing real files is recordable by machinery this repo already has. The single
most credible shot in the course — an AI writing code that then really runs — is
also the cheapest one to make. "Most of the way there" rather than "done",
though: §6 phase 1 has what is actually in the way.

Class D is not a consolation prize. For "here is where the prompt box lives",
a recreated frame is *clearer* than a recording, never rots, and renders
deterministically. Real footage is for what is inherently **temporal**: an app
assembling itself, a diff landing, a deploy going green.

### 3a. Stills beat video for most of it

A screenshot pipeline is an order of magnitude more reliable than a video
pipeline — no frame timing, no encoder, no flake — and a zoomed still with a
generated callout teaches "this is the prompt box" *better* than four seconds
of video does. Ken Burns + annotation over a real screenshot is real footage.

**Rule: capture stills by default, video only where time is the subject.**
Reserve the video path for the build-in-progress, the agent editing, the deploy.

## 4. The architecture

### 4.1 Footage is a library, not a stage output

Capture is slow, flaky, auth-bound, rate-limited and occasionally needs a human.
Generation must stay fast and idempotent. So they do not live in the same run.

```
.coursesmith/footage/
  <clip-id>/
    clip.mp4 | frame-01.png …
    take.yaml        # the driver script — the thing that is re-runnable
    footage.json     # marks, origin, capturedAt, tool version, provenance
```

A clip is captured **once**, described once, catalogued, and any lesson, combo or
snippet can then cast it. `coursesmith footage capture <id>` re-runs one take;
`footage refresh --stale` re-runs everything older than N days.

### 4.2 Marks are what make footage editable by a machine

A raw 90-second Lovable build is unusable narration material. The driver emits a
**mark** at every step it performs:

```yaml
# take.yaml
tool: lovable
origin: https://lovable.dev
steps:
  - goto: /
  - mark: landing
  - type: {sel: "[data-testid=prompt]", text: "a habit tracker with streaks"}
  - mark: prompt-typed
  - click: "button[type=submit]"
  - waitFor: {sel: ".preview-frame", timeout: 180s}
  - mark: app-built
```

`footage.json` then carries `{landing: 0ms, prompt-typed: 4200ms, app-built:
96400ms}`. Downstream that is everything:

- narration can **reference a mark by name**, and the scene cuts there;
- the 92-second wait between `prompt-typed` and `app-built` is **speed-ramped**
  to six seconds automatically, because the engine knows it is dead air;
- a zoom target can be a selector's bounding box, captured at mark time.

This is word-level alignment applied to footage. Without marks, every clip needs
a human editor and the whole plan fails.

### 4.3 Footage flows into `substance` as a new provenance class

The existing fact sheet already has `given / sourced / derived / unverified`,
and only facts with a real origin may be rendered. Add:

- **`captured`** — observed in a recording, with the clip id, the mark, and the
  origin URL. *"Lovable produced a running app from one sentence in 96
  seconds"* is a captured fact; the clip is the citation.

This is the keystone. Casting already refuses templates whose material would
have to be invented — so once footage produces facts, the caster can *see* that
a `footage` segment is fundable, exactly the way it sees a `gauge` is not.

### 4.4 The gate

A clip's `origin` is written by the driver, never by a model. The gate refuses a
scene whose narration names a tool that does not match the clip's origin host.
Same posture as `echo`-faking rejection in the demos stage: the failure mode is
not malice, it is a model confidently captioning the wrong screenshot.

### 4.5 Freshness — the thing that kills courses like this

These tools redesign quarterly. A course full of 2026-07 recordings is wrong by
2026-10, and there is no compiler to tell you. So: `capturedAt` and the observed
tool version live in `footage.json`, `coursesmith footage doctor` lists clips
past their staleness window, and re-shooting is `footage refresh`, not a week
of a person's life. **This is the actual product advantage over a human course
creator** — not that the first cut is faster, but that the tenth is free.

## 5. Decisions (2026-07-31)

- **Terminal capture ships first.** It reuses the `demos` stage, needs no
  browser driving, and produces the two most credible shots in the course — a
  real agent editing real files, and a real deploy going green.
- **Real where time is the subject.** Recorded footage for the temporal
  (building, editing, deploying); stills for "here is where that lives";
  recreated frames where a drawn frame is clearer and cannot rot.

## 6. What gets built

### Phase 1 — terminal footage — **shipped 2026-07-31**

Reading `demos.go` first: the stage is **not** generically a terminal recorder.
Three things are welded to Python and all three have to give.

1. **`--network none`.** `DockerTapeRunner` runs the sandbox with networking
   off, which is correct for executing a stranger's Python and fatal for
   `claude`, `gh`, `vercel`, `supabase` — every one of which is a network
   client holding a credential. A tool capture therefore **cannot run in the
   isolation the Python path depends on**, and pretending otherwise is worse
   than admitting it. Tool tapes get their own runner, on the host, with the
   network up and the operator's real credentials, behind the same loud warning
   the existing `HostTapeRunner` already prints.

   This is a genuine security boundary, not a formality: a tape body is
   LLM-written, and it now runs unsandboxed against authenticated CLIs. The
   lint is what stands there — see (3) — and the allowed-binary list is
   engine-owned, not model-chosen.

2. **`lintTapeBody` hard-requires `python3`** and rejects any tape without it,
   so no tool tape can pass today. It splits by **capture kind**: `python`
   keeps its current rules exactly; `tool` requires the *declared* binary to
   appear in a `Type` line, keeps the `echo` ban, and keeps the
   never-type-a-program's-output rule, which matters more here — a model that
   types Vercel's success banner has forged a deploy.

3. **The prompt is Python's.** `prompts/capture_tape.tmpl` for the tool kind,
   carrying the same iron rule aimed at a different engine.

Then the new parts:

4. **`internal/pipeline/footage.go`** — `footage.json` per clip: id, tool,
   kind, `origin`, `capturedAt`, `toolVersion` (read by running the tool's own
   `--version` at capture time, so it is observed rather than declared),
   duration, marks.
5. **Marks, honestly.** VHS has no timestamp output, so a mark's offset is
   computed from the tape — typing speed × characters, plus explicit `Sleep`s.
   That model is exact for a scripted demo and wrong the moment a real command
   takes real time, which is precisely what `claude` and `vercel deploy` do. So
   the computed total is **asserted against the real mp4 duration**, and a clip
   whose model missed carries `approximate: true`, which disables speed-ramping
   and mark-accurate cutting for that clip rather than silently mistiming it.
   Same posture as the WER gate: measure the thing, and say so when it drifted.
6. **`captured` provenance** in `substance.go`, with the clip id as citation.
7. **`[CAPTURE: tool=… ]` marker** alongside `[DEMO: …]`, so a lesson declares
   which tool a clip is of and the lint knows what binary to demand.

Two things fell out of building it that were not in the spec and had to be:

8. **The script stage had to learn the marker.** `script.tmpl` turns outline
   markers into cues, and the scenegraph places a demo by matching that cue's
   description. A capture the script model has never heard of records a clip
   and then places nothing — the recording would exist on disk and never appear
   in the video. The capture marker now becomes an ordinary `demo` cue.
9. **The window title is the tool, not the sentence.** A capture renders through
   `TerminalScene`, which is right — `claude` and `vercel` really are terminal
   programs — but its title bar was being set to the demo's description. The
   credibility of a capture is that it looks like the thing on the viewer's
   second monitor, and that window says "Claude Code".

`doctor` gained a **separate** check for capture readiness, because the docker
sandbox satisfies `[DEMO]` and can never satisfy `[CAPTURE]`: a machine that
passes the demo check can still record nothing, and finding that out at
`doctor` costs nothing while finding it out nine stages in costs a run.

**What the first real recording changed.** Three things only showed up once a
real `git` capture ran through real VHS, and all three were invisible to a
fake-runner test:

- **`Output` must be quoted.** VHS's parser splits an unquoted argument on its
  path separators and reports the tail as an invalid command — so *every* tool
  capture, which writes to an absolute path back in the lesson dir, died at
  validate on line 1. There is a live test now that renders for real, because
  the mark model is arithmetic about a program we do not control.
- **A cheap `Wait` must not cost a clip its marks.** The model told to use one
  `Wait` used three, putting them after `git init` and `git status` too — where
  a wait returns instantly. All three marks were flagged approximate and were
  in fact perfect. The fix is a proof rather than a tolerance: total drift is
  the *sum* of every wait's overrun, so a clip that ran to its tape time is
  evidence that no individual wait blocked. Wait count stops mattering in that
  case, and only a clip that genuinely overran gets flagged.
- **The tape was appearing in its own recording.** VHS gives the recorded shell
  the working directory the tape was invoked from, so a `git status` demo
  listed `capture-1.tape` as an untracked file — on screen, in the finished
  video. The scratch root holds the tape now and its `work/` child is what gets
  recorded.

**What the first real *agent* capture changed.** Recording `git` proved the
path; recording `claude` proved the path was not enough. Three more, all from
the same run:

- **How to invoke a tool is engine knowledge, not content knowledge.** Claude
  Code's interactive UI asks whether you trust the folder it opened in, and a
  capture always starts in a fresh scratch directory — so the recording stalled
  on a question no keystroke in the tape was going to answer. `captureTool`
  carries an `Invocation` note now (`claude -p "…"`, `vercel --yes`), handed to
  the prompt. It sits beside the allowlist because both answer the same kind of
  question: what this binary is allowed to do and how it is driven.
- **`Wait` needs a timeout the engine sets.** VHS defaults to a few seconds,
  which is right for a scripted demo and turns every interesting capture into a
  failed take. Tool tapes get `Set WaitTimeout 10m`, and the stage's own render
  budget was raised past it so the engine cannot kill a take before VHS does.
- **A worked example outranks the prose above it — again.** The prompt
  described VHS directives at length and showed none, and the model returned
  the *shell session it was imagining*: bare `claude -p "…"` on its own line,
  no `Type`, no `Enter`. That is §11's lesson arriving a second time in a new
  file. The prompt carries a complete correct tape now, and the lint rejects any
  line that is not a VHS directive by name — because the old message
  ("the tape never runs claude") was true, unhelpful, and would not converge.

**The case the whole design was for, working.** The lesson-five capture: an
18.56s tape that recorded as 56.12s, because the agent really thought for
37.5 seconds. One `Wait`, so the entire overrun is attributable — the mark
before it sits at its computed 2.24s, and `feature-added` lands at 47.76s, the
frame showing the agent's own output with the real streak figures from the
fixture. Four marks, none approximate. That is the difference between footage a
machine can edit and footage that needs a person.

Measured against a real clip, the timing model is accurate to about 2%: a
16.32s tape recorded as 15.92s, with `commit-made` landing on the exact frame
showing `[main (root-commit)] Initial commit`. A character costs the typing
speed, a keypress costs the same, a `Sleep` costs its face value, and VHS trims
about 240ms off the tail — which shifts nothing, because marks are offsets from
the start.

**The course half is closed (2026-07-31).** A new **`capture` stage runs before
`script`**, and splitting by kind is what dissolved the cycle: a python demo
needs the verified code so it must follow verify (which follows script), while a
capture needs only its marker and a checked-in take. So captures record first
and the script stage takes `demos/captures.json` as input.

The script prompt now receives measured lines — *"A real recording of Claude
Code 2.1.220, 53 seconds long. It shows, in order: project open, feature added,
files changed, readme content."* — and is told it **may state those durations as
fact**, because they were measured, and must invent nothing else about the clip.

Two decisions in the plumbing: the capture stage writes `captures.json` and the
demos stage writes `manifest.json` by prepending them, so each file has exactly
one writer (two stages appending to one manifest is how it ends up half
overwritten); and the empty manifest is written anyway, because absence and
emptiness mean different things to staleness and a missing file would make every
run look like the captures had just changed.

**Still open for snippets and combos.** The `captured`
provenance class exists, validates and renders, but nothing *writes* a captured
fact — and reading the stage lists says why. `StageSubstance` appears only in
`SnippetStageOrder`; a course has no substance stage at all. `StageDemos`
appears only in `StageOrder`; a snippet has no capture stage at all. **The two
halves have never met**: captures live in the long-form path, provenance lives
in the short-form one.

So there are two different pieces of work behind the one open item, and they are
worth separating before either is started:

There is no capture stage in `SnippetStageOrder` to produce a fact from, so a
combo still cannot cite footage. That is a feature to build rather than a wire to
connect, and it is the last piece of §4.3.

### Phase 2a — web stills — **shipped 2026-07-31**

Stills only, and that is the plan's own recommendation rather than a shortcut:
no frame timing, no encoder, no screencast stream to lose, and for most of what
this course shows a zoomed frame with a callout teaches better than four seconds
of video. Video is phase 2b, for what is inherently temporal.

8. `capture_web.go` — a take runner on **`rod`**, already a dependency.
   `screenshot.go`'s header argues rod-over-Playwright for this exact case and
   the argument still holds.
9. **The take is checked in and written by a person**, and that is the one real
   asymmetry with the terminal path. A model can write a VHS tape because the
   tape drives *our* shell; nobody can invent `[data-testid=prompt]` for
   somebody else's DOM — they can only guess confidently and produce a take that
   screenshots the wrong element. So a take is `courses/<c>/takes/<n>.yaml`,
   authored once and maintained. That is not a workaround: it *is* §4.5's
   freshness mechanism. A checked-in take is the thing you re-run to discover
   that a product redesigned. Hand-recorded video has no such thing, which is
   exactly why it rots silently.
10. **The origin gate.** Every step must stay on the site's origin, checked
    before a browser launches. `footage.json` records the origin as evidence, so
    a take that wandered to another host would produce a clip whose stated
    provenance is false — the precise failure this track exists to prevent.
11. **Focus boxes, normalized.** A `shot` may name the element it is really
    about; its bounding box is recorded as 0..1 of the viewport, and
    `FootageScene` pushes in on it. Normalized rather than pixels because the
    renderer scales frames to the stage, and a pixel box would be right only at
    the resolution it was captured at. No focus means the frame holds — a push
    toward the middle of a screenshot is motion for its own sake.
12. **Auth is deliberately dull.** `coursesmith footage login <site>` opens a
    real window against the site using the same profile the capture will reuse;
    a person signs in, presses Enter, and every later capture runs headless.
    Nothing in the repo, nothing in an environment variable, no automation ever
    typing a password.
13. **`coursesmith footage list`** reports every clip, when it was shot, what
    tool version it caught, and which have passed the staleness window. This is
    §4.5 made operable — the compiler for a problem that otherwise has none.

Verified against a real browser rather than a fake: `capture_web_live_test.go`
serves its own page and drives it, asserting the frames are real PNGs at the 2×
device scale, that a focus selector resolves to a box in the right place, that a
shot with no focus carries none, and that a missing selector fails naming the
selector and the step. It uses a local page rather than a real product on
purpose — a test that depended on Lovable's markup would fail for reasons that
are not our bug.

### The marks finally did something — **2026-07-31**

Rendering lesson five end to end produced a finished 1920×1080 video with the
real Claude Code session in it, and looking at the frame exposed the defect the
marks were built for and nothing had used: **the clip was 53.3 seconds and the
scene slot was 20.8**, so the video cut away ten seconds in and the viewer never
saw the agent's output. The whole point of the shot, lost, in a build that
reported success.

Playing it faster everywhere would make the typing ridiculous. What the clip
holds is a few moments separated by dead air, and the marks say where. So
`PlanTerminalPacing` compresses the gaps and leaves the moments alone. On the
real clip:

| stretch | what is in it | rate |
|---|---|---|
| 0–2.2s | `ls`, then the command being typed | **1×** |
| 2.2–44.9s | the agent thinking | **4.19×** |
| 44.9–53.3s | its output, with the real streak figures | **1×** |

20.79s of playback into a 20.79s slot, with nothing dropped.

Two decisions worth keeping. **The plan is computed in Go and written into the
scene graph**, the way the motion tokens already are — `lesson-video.json`
records what was decided, a Go test checks the arithmetic against the real
numbers, and the renderer only plays what it is told. And **only exact marks are
cut points**: a clip whose timing could not be attributed gets a uniform
speed-up instead of a confident cut in the wrong place. That is the
`approximate` flag from phase 1 finally earning its place rather than being a
field nothing read.

### Phase 2b — web video — **shipped 2026-07-31**

Stills carry most of the course and §2a argues why. They cannot carry the one
shot the course exists for: a sentence going in and an application coming out.
Nobody believes that from two screenshots — the point is watching it happen,
including how long it takes.

A take gains `record` / `mark` / `stop`. Between record and stop, CDP
`Page.startScreencast` streams JPEG frames, `mark` stamps a moment, and ffmpeg's
concat demuxer turns the frames back into an mp4 whose pacing matches when they
really arrived. Stills and a recording can live in the same take; they answer
different questions.

**These marks are measured, not modelled.** The terminal path has to *infer*
where its marks fall, and the whole apparatus in `footage.go` exists to admit
when that inference cannot be trusted. Here we hold the clock: a frame arrives
when it arrives and a mark is stamped against the same clock. Web video marks
are never `approximate` — not a happy accident, but the difference between
driving the recorder and shelling out to one.

Screencast only sends a frame when something **changed**, which suits this
exactly: ninety seconds of a spinner is a handful of frames, so the recording
stays small and the timing lives in per-frame durations rather than in a wall of
identical frames.

Two bugs worth recording, both found by running it:

- **`EachEvent` returns a *wait* function, not a cancel.** Calling it to "stop
  recording" blocked for the full three-minute context, and the encode then ran
  against an already-expired context — surfacing as "ffmpeg: context deadline
  exceeded", which points at the wrong component entirely. A background
  subscription ends by cancelling a derived page context. 180s → 3.5s.
- **The frame slice was written from the event goroutine** and read from the
  step loop. Caught by `-race`, not by any assertion.

The live test asserts what duration alone cannot: **the first and last frames of
the clip must differ**, and the clip must not be essentially black. A recording
of the right length that caught nothing would pass every other check, and "black
video" is the classic screencast failure.

Pacing is shared with the terminal path — a web clip of an app being built runs
long for the same reason an agent session does, and its dead air is the same
dead air.

### The capture credit — pacing created an honesty problem, so pay for it

Compressing 53 seconds of real agent work into a 21-second slot and **saying
nothing** is a quiet misrepresentation of how long the tool took. In a course
whose entire moat is *the tool really did that*, every other claim is defended
by a recording; the length of the recording cannot be the one claim allowed to
drift. The pacing work in §2b/§1 introduced that, so it has to be paid for.

A condensed clip now states both numbers in frame:

```
Claude Code 2.1.220  ·  53s real, shown in 21s
```

Three properties earn it a place:

- **Every field is measured.** The tool name comes from the engine's own
  registry, the version was read from the binary at capture time, the durations
  from the recording and the scene. Nothing here can be written by a model —
  which makes this the first *rendered* captured fact, and the closest the
  course path gets to §4.3 without the stage reordering that path still needs.
- **It only claims what happened.** A clip that plays at real time says nothing
  about its speed, and a sub-second difference is rounding rather than a
  fast-forward. "12s real, shown in 12s" is noise wearing the costume of rigour.
- **A python demo gets none.** That is our own code running, not somebody
  else's product, so it makes no provenance claim at all.

The full version banner stays in `footage.json` and only the *display* is
shortened to the number — `claude --version` prints "2.1.220 (Claude Code)", so
the raw string renders as "Claude Code 2.1.220 (Claude Code)". That file is
evidence, and evidence should not be tidied.

### Selector fallbacks — how a take outlives a redesign

`selector:` takes a string or a **list tried in order**:

```yaml
selector:
  - "[data-testid=prompt]"      # the precise one
  - "textarea[placeholder]"     # the shape it probably has
  - "textarea"                  # the last resort
```

A single selector makes every redesign a broken take. A short ladder usually
survives one, and when it does not the error names **every alternative it
tried** — which is the difference between "fix this line" and "work out what
this page looks like now", a year later, for somebody who did not write it.

This is also what makes an unverified take useful rather than merely optimistic:
`first-build.yaml` cannot be checked against Lovable from here, but a ladder of
three guesses has a far better chance than one, and fails legibly when it misses.

### The narration had to be corrected too

Regenerating lesson five's script with the briefing produced: *"nothing is
staged and nothing is sped up"* — while the frame said **53s real, shown in
21s**. The narration contradicted the credit.

The claim came from the lesson outline, written before pacing existed. Three
lessons asserted some form of "nothing is sped up" and all three were corrected
to say what is actually true: the waiting is condensed, the frame says by how
much, and no step is cut. The regenerated script now reads *"The recording is
sped up, but the corner of the frame will show you how long it really took.
You'll see every step, just not every second of waiting."*

Worth keeping as a pattern: a feature that changes what the video *is* can
falsify prose written before it, and nothing but reading the output catches it.

### Takes are validated by the test suite

A take is the one artifact here a person writes by hand, against somebody else's
markup, and nothing exercises it until a capture runs — which needs a browser, a
login and several minutes. `TestShippedTakesAreValid` loads every
`courses/*/takes/*.yaml` and checks the site it names is still one we can drive,
so a typo in a step name fails in seconds rather than on the day somebody tries
to shoot the lesson.

### Phase 3 — desktop capture — **shipped 2026-07-31**

Cursor and Figma desktop, the apps with no DOM to drive.

**This one has a person in it, and that is the design rather than a shortfall.**
The terminal path generates its own script; the web path runs one somebody wrote
against selectors they can inspect. Neither is available here — a native app
exposes nothing to select, and driving it by simulated keystrokes means encoding
pixel positions and menu paths that break on the next release with nothing to
notice. So the engine does what a person is bad at (framing the window
identically every time, recording, cropping, timing, encoding) and the person
does what they are good at, which is using the application.

A take is a list of **beats**. The engine shows one at a time and stamps a mark
when the operator presses Enter — so these marks, like the web-video ones, are
measured rather than modelled.

That is slower than the other two paths and **no less repeatable**: the beats
are checked in, the window geometry is fixed, and a re-shoot is running the same
list again. A hand-recorded video offers none of that, which is §4.5's whole
argument.

Four decisions worth keeping:

- **Whole screen, then crop.** avfoundation captures a *display*, not a window;
  window-level capture needs ScreenCaptureKit and a second toolchain.
  Positioning the window to a known rectangle and cropping to it gets the same
  frames from tools already here. The window sits clear of the menu bar, because
  recording it would put the operator's clock, battery and unrelated app names
  into the course.
- **The Retina scale is measured, not assumed.** AppleScript works in points and
  the capture arrives in pixels. Assuming 2 is right on a laptop panel and wrong
  on an external monitor, and the failure is a video cropped to a quarter of the
  window with nothing on screen to say why. The scale comes from dividing the
  recording's real pixel width by the display's point width.
- **The device index is parsed, not hard-coded.** It counts *all* video devices:
  with no webcam the screen is 0, and on this machine it is 1 because the
  FaceTime camera took 0. Hard-coding it records somebody's face.
- **`run` refuses these; `footage shoot` does them.** A desktop capture blocks
  on a keypress, and a batch build that stopped halfway waiting for somebody is
  indistinguishable from a hang. The console check runs *before* the take is even
  read — whether anybody is at the keyboard is a property of the run, not of the
  take, and a missing-file error there would send somebody looking in the wrong
  place. (A test caught that ordering.)

`doctor` gets its own check, separate again, because the permission is granted
to the **terminal application** rather than to coursesmith — so the fix lives
somewhere the error message would never lead you, and without it ffmpeg simply
lists no screen rather than saying it was refused.

**Not verified end to end.** The take validation, crop arithmetic, device
parsing and the headless refusal are all tested; the recording itself needs a
real screen, real permissions and a person, so it has never been run here. That
is a real gap and it is the nature of the path rather than something to fix.

### Phase 3 — the hard ones
11. macOS window capture for Cursor / Figma desktop. Human-in-the-loop take.
12. `footage refresh` / staleness in `doctor`.

### Phase 4 — course-shaped
13. Two or three templates the no-code course needs and the catalog lacks:
    `handoff` (design → working UI), `ship` (the deploy moment), and something
    for the price/limit comparison that `showcase` handles one tool at a time.

Templates already carrying this subject well: `promptloop`, `canvas`, `mockup`,
`stack`, `spec`, `showcase`, `decision`, `verdict`, `costing`, `myth`.

## 6a. `nocode` — the fourth surface (2026-07-31)

A course was the wrong shape for what this track is actually for. courseSmith
has *surfaces*, and no-code needed to be one of them:

```
courses   a document first, a video second — chapters, quizzes, a site
snippets  one prompt, one template, one clip
combos    several templates cut onto one timeline
nocode    several segments, every one of them backed by something real
```

**The rule that earns it.** A combo about no-code tools is not this. A combo will
happily draw a person gesturing at a chart of numbers nobody measured, because
its contract is "each segment names material it can be filled with" and prose
counts as material. That is right for a combo and wrong here.

Every segment of a no-code piece must be backed by **evidence that exists on
disk** — a real capture, or a fact carrying provenance from the substance sheet:

```yaml
segments:
  - template: footage
    prompt: a habit tracker built from one sentence
    evidence: {kind: capture, capture: capture-1}
  - template: costing
    prompt: what running it actually costs
    evidence:
      kind: fact
      facts: ["Vercel's Hobby plan is free for personal projects"]
```

A segment that can name neither is refused **before a single planning call is
spent**. A combo discovers a hollow segment after the caster has gone and a
template's validator refuses the material; here it is a parse error.

That one rule produces everything else about the surface:

- **`NoCodeStageOrder` puts `capture` first** — the recordings have to exist
  before anything decides what to say about them. The argument, made mechanical.
- **The catalog is a subset.** `cast`, `story` and `illustration` are excluded,
  and the error says why: a drawn figure is the fastest way to fill a frame with
  no evidence behind it, and this surface exists to refuse that. It is not a
  judgement about the artwork — they are good templates in the wrong place.
- **Facts are written out, not referenced by index.** A fact sheet is
  regenerated; an index into it is a reference that silently retargets.

### The `footage` template

The keystone, and the first template in the catalog that draws nothing. The
frame is the recording, inside chrome that says where it came from, with the
capture credit stating the tool, its observed version and both durations.

**It is the one validator in the catalog that checks a plan against evidence on
disk rather than against its own shape.** The writer is narrating something it
cannot see — it gets the mark names and the durations and nothing else — so the
failure to defend against is a beat about a moment the recording does not
contain. `footage.json`'s marks are measured, which is the only reason the check
is possible; a beat naming an unknown mark is rejected with the real list quoted
back. Its prompt opens with **"YOU CANNOT SEE THE RECORDING"** for the same
reason.

Its gallery preview is the one that cannot show real content, and that is a
property of the template rather than a shortcut: what `footage` produces is
*your* recording, so a card showing somebody else's clip would misrepresent what
you get. The preview shows what the template actually contributes — the frame,
the origin, and the credit — around a deliberately neutral placeholder. It is
still a real render through the same visual-regression machinery as every other
card, so it cannot drift from the component.

## 7. The course

Shape: **a `courses/no-code/` course** — the combos path is right for a five-part
argument, and wrong for a thirteen-lesson curriculum that needs chapters,
quizzes, exercises and a site. Verify is skipped when a lesson has no code
blocks, so the course path already runs without Python.

Market note: **63% of vibe-coding users are not developers** — PMs, founders,
designers. That is the audience, and it argues for outcome-first lessons where
something works in the first ten minutes, which is what every well-reviewed
course in this space is doing.

| # | lesson | the real footage in it |
|---|---|---|
| 1 | Why you don't need to learn to code — and what that actually means | montage; the honest caveat |
| 2 | The map: what each tool is *for* | `stack` — mostly drawn |
| 3 | First build: one sentence → running app | **Lovable, web capture, the flagship clip** |
| 4 | Prompting is the new syntax | `promptloop` + real diffs |
| 5 | When you outgrow the browser: Cursor & Claude Code | **Claude Code via VHS** (cheap, real) |
| 6 | Design: Figma → real UI | Figma web capture + `handoff` |
| 7 | Data without SQL: Supabase | web capture, table → rows |
| 8 | Auth, secrets, and what will bite you | stills + `spec` |
| 9 | Ship it: Vercel, domains, going live | **`vercel` CLI via VHS**, deploy goes green |
| 10 | Glue: Zapier / Make / n8n | `canvas` (already built for this) |
| 11 | When it breaks: debugging with AI | real error, real fix, recorded |
| 12 | The honest limits: cost, lock-in, security, what still needs a dev | `costing`, `verdict` |
| 13 | Capstone: the whole flow end to end | the long capture, ramped |

Lesson 12 is not a downer, it is the credibility lesson — and it is what the
`showcase` template's enforced limitations column already exists to protect.

## 8. Risks, honestly

- **Legal.** Recording third-party SaaS UI in a *paid* course is contract law
  (their ToS) and trademark, not just fair use — and commercial use without
  transformation is where fair use gets thin. Mitigation: narrate over every
  clip (transformation), never imply affiliation, prefer each vendor's own
  press/brand kit for logos, capture only accounts we pay for, and read the ToS
  of the six tools that carry the course before shooting. Worth a lawyer's hour
  before launch, not after.
- **Flake.** Third-party UIs change selectors without notice. Marks fail loudly
  rather than producing a silently wrong clip; a failed take blocks its lesson,
  it does not corrupt it.
- **Auth and cost.** Every tool needs a paid seat and a logged-in profile.
  Budget it as a line item; it is the course's raw material.
- **Rot.** Covered in §4.5 — and it is the reason to build this rather than
  hire an editor.

## 9. The one-line pitch

Everyone else's no-code course is a person recording their screen, and it is
out of date the quarter after it ships. This one re-shoots itself.
