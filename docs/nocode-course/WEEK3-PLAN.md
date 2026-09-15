# WEEK 3 — EXECUTION PLAN

Decided 2026-08-29. Supersedes the per-lecture approach in `WEEK3.md`.
Beat sheets `ref-L65.md` … `ref-L94.md` remain the **reference**, not the script.

## The four decisions

| | chosen | consequence |
|---|---|---|
| shape | **Same arc, our cut** | 26 lectures, not 30. Same 5-day arc, same 5h12m. Lecture breaks land where OUR footage changes |
| method | **One build, many cuts** | 4 long continuous captures. Every lecture is an *edit*, not a shoot |
| orchestrators | **Claude agent teams only** | GSD, Gastown, Sprites.dev and OpenClaw are OUT. Days 2, 4 and 5 are re-spined around tools we own |
| capstone | **A different product** | Not his trading workstation. See "Control Tower" below |

## What "no-code pro" means for this week

Week 1 taught building. Week 2 taught working like a professional. **Week 3 teaches
directing** — the student never writes a line, they set up the conditions and
supervise many agents at once. Every tool on camera is one we already pay for and
can set up in minutes:

**In:** Claude Code (+ sub-agents, hooks, plugins, agent teams, Agent SDK) ·
Cursor · GitHub Copilot · Docker · GitHub Actions · Claude on web + mobile ·
Claude Desktop / Cowork · OpenRouter · ElevenLabs.

**Out, and why:** **Gastown** (8 parallel Opus agents — the biggest token spend in
the program) · **GSD** (5 hours of wall clock to prove a negative) ·
**Sprites.dev** (another account, and Docker gives us the same lesson) ·
**OpenClaw** (needs a Telegram bot and grants a hobby beta full machine access).

**Nothing is lost by cutting them.** Every teaching point they carried has a
replacement built from something we own:

| his beat | our replacement | why it is not worse |
|---|---|---|
| Sprites.dev cloud sandbox | **Docker devcontainer sandbox** | Same lesson — YOLO mode without fear — with no signup and no bill |
| GSD spec-driven orchestration | **The Agent SDK orchestrator we write on Day 5** | The student *builds* the disciplined runner instead of installing someone else's, and it pays off Day 3 |
| Gastown 8-agent swarm | **Three real contenders: Claude teams · Cursor · Copilot** | A three-way comparison of tools they actually have beats a tour of one they never will |
| OpenClaw sidekick | **Cowork** (already in the week) | Same "agents beyond code" point, no bot token, no security exposure |

---

## THE CAPSTONE — "CONTROL TOWER"

A live logistics operations console. Same density and motion as his trading
terminal, but a **map** makes it visually distinct, all data is simulated so it
costs nothing, and "reroute this van" is a far more legible AI action for a
no-code audience than "buy three shares."

**On screen:**
- **Left** — live fleet list, ~60 vehicles grouped by lane and region, status
  colour, ETA counting down
- **Centre** — a live map with vehicles moving along routes, plus the main chart
  (on-time % across the day)
- **Right** — an SLA **heat map** by region: box size = shipment value, colour =
  delay risk. This is the frame that reads instantly at a glance
- **Bottom** — an exceptions ticker: breakdowns, weather holds, customs
- **Chat** — an AI dispatcher: *"reroute VAN-14 around the closure"*, *"which
  lanes are at risk this afternoon?"*, *"hold the Leeds shipment and tell me why"*
- Dark **and light** theme, Docker container, deployed live at the end

**Why it builds zero-shot:** it is the same component set his did — ticker, heat
map, chart, chat, websocket feed — so the agents are on well-trodden ground.

**Data:** a simulator we generate. No API, no key, no cost. It is a logistics
sim, so nobody expects it to be real — which removes the live-data question
entirely.

**One-line swaps if you'd rather:** *Grid Console* (energy: plants, load curve,
price ticker) or *Signal Desk* (newsroom: live headlines, entity heat map,
sentiment). Say the word before Capture A and nothing else in the plan changes.

---

## THE 26 LECTURES

Runtimes are targets. Per-day totals match his exactly so the section length holds.

### Day 1 — The pro tier of one agent (6 lectures, 72:00)

| # | title | len | shape |
|---|---|---|---|
| w3-01 | Welcome to expert: what changes when you stop typing | 11:00 | **slides 60%** — the one deck-heavy opener. Ladder, the week, the project reveal |
| w3-02 | The brief: setting up Control Tower | 12:00 | footage 90% — repo, `claude.md`, `plan.md`, the spec |
| w3-03 | Slash commands: a routine in one word | 12:00 | footage 100% |
| w3-04 | Sub-agents: a team of specialists inside one agent | 13:00 | footage 95% |
| w3-05 | Hooks: making your rules fire by themselves | 11:00 | footage 95% |
| w3-06 | Plugins and marketplaces: packaging your setup | 13:00 | footage 100% |

### Day 2 — Many agents, safely (5 lectures, 61:00)

| # | title | len | shape |
|---|---|---|---|
| w3-07 | Why sandboxing is the unlock | 10:00 | footage 70% — native `/sandbox` in Claude Code and Cursor |
| w3-08 | **YOLO without the fear: the container sandbox** | 13:00 | footage 95% — **replaces Sprites.dev.** Devcontainer, then skip-permissions *inside* it |
| w3-09 | Claude Code on the web | 12:00 | footage 100% |
| w3-10 | Claude Code on your phone, and the `&` prefix | 12:00 | footage 100% |
| w3-11 | Tagging Claude on a GitHub issue, all the way to a PR | 14:00 | footage 100% |

### Day 3 — Agents off the command line (3 lectures, 46:00)

| # | title | len | shape |
|---|---|---|---|
| w3-12 | Working in a big codebase: the seven rules | 14:00 | **footage 60%** — his "vegetables" lecture, but shot over real files instead of bullet slides |
| w3-13 | Driving Claude from code: the Agent SDK | 17:00 | footage 90% — empty folder → a working game. **Sets up Day 5's orchestrator** |
| w3-14 | Agents beyond code: Cowork | 15:00 | footage 90% — synthetic receipts only |

### Day 4 — The team builds it (6 lectures, 68:00)

| # | title | len | shape |
|---|---|---|---|
| w3-15 | Sub-agents vs agent teams: what actually changes | 11:00 | **slides 70%** — the second and last deck-heavy lecture |
| w3-16 | Setting up the team: clean house, pick the roster | 11:00 | footage 100% |
| w3-17 | Launch: seven agents, one shared task list | 12:00 | footage 100% |
| w3-18 | Watching a swarm work: approvals, blocks, tokens | 11:00 | footage 100% |
| w3-19 | The integration tester finds the bugs | 11:00 | footage 100% |
| w3-20 | First run: Control Tower is alive | 12:00 | footage 100% — **the showcase lecture of the program** |

### Day 5 — Three ways to run a swarm, and shipping (6 lectures, 65:00)

| # | title | len | shape |
|---|---|---|---|
| w3-21 | Contender two: Cursor's parallel agents, same brief | 12:00 | footage 100% |
| w3-22 | Contender three: Copilot's coding agent | 10:00 | footage 100% |
| w3-23 | **Roll your own orchestrator with the Agent SDK** | 11:00 | footage 95% — the Gastown replacement, built not installed |
| w3-24 | The verdict: four builds side by side | 12:00 | footage 90% — one graphic: time, tokens, what each started from |
| w3-25 | Ship it: Docker, deploy, live on the internet | 10:00 | footage 90% |
| w3-26 | Wrap: from no-code to agentic director | 10:00 | **montage 100% of our own footage** — no bullet slides |

**Totals:** 312 min = **5h12m** across 26 lectures. **Deck-heavy lectures: two
(w3-01, w3-15).** Everything else is 90%+ footage. Week average ≈ **91% footage.**

---

## THE PIPELINE — one build, many cuts

### Four captures

| capture | covers | cut footage | raw to shoot | machine state |
|---|---|---|---|---|
| **A** | Day 1 + Day 3 (9 lectures) | ~118 min | ~3h | VS Code + Claude Code on the Control Tower repo; Cowork segment at the end |
| **B** | Day 2 (5 lectures) | ~61 min | ~1.5h | Docker, browser, phone, GitHub |
| **C** | Day 4 (6 lectures) | ~68 min | ~2h | One continuous agent-team run, start to running app |
| **D** | Day 5 (6 lectures) | ~65 min | ~2.5h | Cursor, Copilot, our orchestrator, the four-way, deploy |

**~9 hours of raw capture for the whole week.** That is the number the plan turns on.

### The ten steps, in order

1. **Preflight** — `tools/nocode/shoot-w3/preflight.sh`: dark theme on, kickbacks
   ad still gone, `PROMPT='%~ %# '`, shooting from `/Users/Shared/projects`, the
   16:9 display picked by aspect (not by hardcoded id), no other shoot session
   running, disk space, frontmost-app guard armed.
2. **Shot plan, not a script** — `docs/nocode-course/w3/capture-A.md` … `-D.md`:
   the beats in order, what I click, what must never be on screen. Written and
   reviewed *before* rolling.
3. **Capture** — avfoundation 3840×2160, one continuous file per session.
4. **Marker log** — at every beat boundary, append `date +%s` + a label to
   `markers.txt`. **This is what makes cutting cheap.** Without it, 3 hours of
   footage is unsearchable.
5. **Cut** — `tools/nocode/scripts/cut.py` reads `markers.txt` + the lecture map
   and emits one source clip per lecture. Never cut past the end of a take.
6. **Watch the cut, then write** — narration is written *to our frames*. His
   transcript guides **what to teach**; it never supplies a sentence. Result
   narration only ever after the run. This is the W1 L17 rule and it applies to
   w3-19, w3-20, w3-21, w3-24 hardest.
7. **Voice** — `program/part-w3-NN/vo/script.json` → `generate_vo.py`, Adam,
   `playbackRate 0.9`, ellipses as pacing marks.
8. **Deck** — only the slides that lecture actually needs, in the new sticker kit.
9. **Assemble** — deck section (recorded via `export_mp4.sh`) hard-cuts to
   footage and, per the kit's own rule, **does not go back**.
10. **Master and deliver** — two-pass loudnorm to −16 LUFS + `alimiter=limit=0.82`,
    then `videos/nocode/vids/w3-NN_adam.mp4`. Delivery is the last step of every
    lecture, not a staging folder.

---

## THE DECK — new sticker kit, ported to `program/`

`No-Code 4.zip` is unzipped into `program/` and the style rule into
`.cursor/rules/illustration-style.mdc`.

**The look:** black field `#070707` · Archivo Black display + IBM Plex Mono ·
body `#f5c518` yellow · **hard red `#e53935` offset down-right ~12px, no blur** ·
blue `#3b82f6` only as a small accent · thick black outlines, stroke 6–8.

**How the deck runs:** each `.slide` carries `data-vo="vo/NN.mp3"` and
`data-enter="whoosh|hit|rise"`; **slides advance when the voice clip ends**, so
slide timing follows narration automatically — the gap formula is not needed for
deck sections at all. `?record=1&wait=1800` auto-plays for capture.

**Three rules carried over from the kit, and they match your brief exactly:**
1. *"Reference frames are a look, not a kit to stamp on every slide."* Draw a
   subject only when that line needs it. **One hero mark per slide, or a pair for
   a versus.** No icon walls.
2. The shot list convention: slides for the opening beats, then **a hard cut to
   screen recording — and after that cut, do not go back to slides.**
3. No soft shadows, no gradients, no thin line icons, no photos.

**Bonus we should use:** the kit has `data-ask` interactive pause cards
(single / multi / quiz / code) with a 30-second countdown and their own
pause/resume VO. Ed has nothing like it. **One per day, at the point where a
student would otherwise coast.**

**Subjects to draw for Week 3** (originals, in each part's `images/`): a control
tower · a swarm of workers around one lead · a padlocked container · a phone with
an agent in it · a merge queue · a task list with a shared pen.

---

## COST AND RISK

**ElevenLabs:** 312 min × 1,023 chars/min ≈ **319,000 characters**, roughly **$52**
at the flat ~$10 per finished hour we measured. Purge `.voice` caches after each
part — the cache goes stale silently when text changes.

**Claude tokens — the real number.** One agent-team build ran him ~8% of a weekly
quota. We run **four builds** (teams, Cursor, Copilot, our own orchestrator) plus
retakes. **This is the one line item I need a cap on before Capture C.**

**Everything else is free:** Docker, Cursor, Copilot, GitHub Actions, the
simulator.

### Risks, and what each costs

| risk | mitigation |
|---|---|
| A bad capture costs a whole block, not one lecture | Marker log + beats ordered so a failure can be re-picked mid-session rather than restarting |
| Agent teams is experimental — the flag may be renamed or gone | Verify on the day. If it shipped GA, the beat becomes *"you no longer need this"*, which is better footage |
| Cursor / Copilot parallel-agent features may differ from expectation | **Pre-check both before Capture D**, not during. Fallback: two contenders instead of three, said plainly |
| Our build fails where his succeeded | Keep it in. A failure we recover from on camera is the most valuable footage in the week |

### Never on camera
Docker Desktop lists · Cursor Settings · VS Code Workspace Trust · the default
zsh prompt · any URL, sprite name or app name carrying a personal name · the
GitHub avatar (blur in post) · `~/.zshrc` in any form.

---

## SETTLED (2026-08-29)

1. **Claude usage — no cap.** He is on **Claude Max**, so all four builds run
   without stopping to ask. **`/usage` still goes on screen at both ends of every
   build**, because w3-24's verdict is worthless without our own numbers.
2. **Claude Code on web + mobile is covered by Max** — w3-09 and w3-10 are safe.
   Confirm both load on the shoot day before rolling; no fallback needed.
3. **Capstone confirmed: CONTROL TOWER** — delivery vans on a live map.
4. **Ending: "go build."** No LinkedIn, no rating ask, no links. Nothing personal
   of his on screen, ever. w3-26 ends on the montage and one line.

## A structural finding about runtime (2026-08-29)

**Our lectures are naturally shorter than his, and that is not a defect.**

The reference instructor types everything by hand: nine minutes of a lecture is
nine minutes of keystrokes. We ask an agent, and the same work is a sentence.
Measured on the first Day 1 captures, a beat that takes him minutes takes us
seconds.

Two honest responses, and only two:

1. **Teach more per lecture.** Harder questions, read the artifacts the agent
   produced, make it justify its choices. This is what the depth pass
   (`shoot_a1b.py`) does, and it is genuinely better teaching — reading what came
   back is the actual skill in a no-code course.
2. **Fewer, denser lectures.** If a lecture will not honestly fill its slot,
   merge it with its neighbour rather than pad it.

**What must not happen is padding.** `assemble.py` prints `HELD ON LAST FRAME`
whenever narration outruns footage, so it cannot pass unnoticed.

**Expect the week to land shorter than 5h12m.** A realistic figure will only be
known once Day 1 is cut and measured; the number to report is the honest one, not
the one that matches the reference.

## Order of work

1. Port the deck kit ✅ done
2. Write the Control Tower spec (`plan.md`) — the artifact every capture builds
3. `preflight.sh` + the marker-log rig + `cut.py`
4. `capture-A.md` shot plan → **shoot Capture A**
5. Cut 9 lectures, narrate to the frames, voice, deck, assemble, deliver
6. Repeat for B, C, D
