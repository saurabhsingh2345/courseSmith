# WEEK 3 — CAPTURE PLANS

Five sessions, not four: **Capture A is split into A1 and A2** so no single
session exceeds ~2 hours. At 30 Mbit that is ~27 GB, which fits the 57 GB free
with room for the cuts. Cut and purge each session before rolling the next.

**Marker names are the contract.** `cut.py` gives a lecture every marker whose
label starts `w3-NN/`, and ends that lecture at the first marker belonging to a
different one. So mark where each beat *begins*; never mark an end.

```python
from rig import Session
s = Session("A1", allow=("Code", "Claude"))
s.start()
s.mark("w3-02/repo", "empty folder, first look")
...
s.stop()
```

## Rules that apply to every session

- `python3 tools/nocode/w3/preflight.py` first. It exits non-zero for a reason.
- Every terminal on camera: `PROMPT='%~ %# '` then `clear && printf '\033[3J'`.
- Work in `/Users/Shared/projects/control-tower`. Never `~`.
- **Quit the editor we are not filming.** A take was lost when Cursor stole
  focus mid-VS-Code shot.
- **Click into the pane and confirm by snapshot before typing.** Typing into
  Cursor without checking sent prompts to the agent twice, and it asked to edit
  `~/.zshrc`.
- Never on camera: Docker Desktop lists · Cursor Settings · VS Code Workspace
  Trust · the default prompt · `/plugin` or `/mcp` with anything personal
  installed · the GitHub avatar (blur in post).
- **Record from frame one.** Do not set up and then start rolling; the setup is
  the lecture.
- Waiting is filmed and ramped **12–14x** in post, never cut to black.

---

## CAPTURE A1 — Day 1 · one agent, properly

**Covers w3-02 … w3-06** (five lectures, 61 min cut, ~1h50m raw).
w3-01 is a slides lecture whose only footage is a teaser of the finished app,
so **it is picked up at the end of Capture D.**

State: VS Code + Claude Code, one repo, Docker not needed.

| marker | beat |
|---|---|
| `w3-02/empty` | The empty folder. `git init`. Nothing here yet |
| `w3-02/spec` | Paste `plan.md` — the Control Tower spec. Scroll it slowly; this is the document the whole week is built on |
| `w3-02/claudemd` | Write `claude.md`. **`@`-inline only `plan.md`; link everything else by path in backticks** so the agent pulls it in only when needed |
| `w3-02/first` | First Claude Code session in the repo. `/context` — show how little is loaded, and why that is the goal |
| `w3-03/why` | The routine we keep repeating by hand |
| `w3-03/make` | `.claude/commands/` — write the first slash command |
| `w3-03/run` | Run it. Then a second one that takes an argument |
| `w3-03/scope` | Project scope vs personal scope, and why a team wants project |
| `w3-04/why` | One agent, one context window, and what happens when it fills |
| `w3-04/make` | `/agents` — create a specialist. Its own context, its own tools |
| `w3-04/run` | Give it real work on the simulator. Watch it report back and vanish |
| `w3-04/explore` | The built-in explore sub-agent doing the same job for free |
| `w3-04/limit` | **You cannot talk to a sub-agent.** Set up Day 4 without naming it yet |
| `w3-05/why` | The rule we keep forgetting to apply |
| `w3-05/make` | `settings.json` hooks — fire a check on every edit |
| `w3-05/fire` | Break the rule on purpose and watch the hook catch it |
| `w3-05/events` | The other events, and one worked example that is not a linter |
| `w3-06/browse` | `/plugin` — the marketplace |
| `w3-06/install` | Install two at project scope. **Say out loud why we skip anything that spawns its own sub-agents** |
| `w3-06/restart` | Restart for them to take effect. `/context` again — see the cost |
| `w3-06/own` | Package our own slash command + hook + agent as a plugin |
| `w3-06/share` | Point it at a repo so a teammate gets the whole setup in one command |

**End on:** the repo has a spec, a memory file, commands, an agent, hooks and
plugins — and not one line of application code yet. That contrast is the
Day 1 close.

---

## CAPTURE A2 — Day 3 · agents off the command line

**Covers w3-12 … w3-14** (three lectures, 46 min cut, ~1h20m raw).
Shot after A1 because w3-12 needs a repo with history in it.

| marker | beat |
|---|---|
| `w3-12/rules` | The seven rules, **shot over the real repo**, not over bullets. Docs at every level · describe-and-link rather than `@`-inline · one team workflow · plugins then skills · honest tests · you are accountable · bite-sized chunks |
| `w3-12/at` | The `@` correction applied live to our own `claude.md` from w3-02 |
| `w3-12/tests` | The coverage trap: ask for 80% and read back the mock-everything slop it writes. **Then reject it on camera** |
| `w3-13/empty` | A brand-new empty folder |
| `w3-13/setup` | `uv init --bare`, add `claude-agent-sdk`. Name the rename: it is `claude-agent-sdk`, agent singular |
| `w3-13/code` | ~15 lines: prompt, allowed tools, options, `async for message in query(...)` |
| `w3-13/model` | Pick the model — **and warn against the expensive one in a loop, on a held frame** |
| `w3-13/run` | Run it. Files appear in the empty explorer |
| `w3-13/play` | Open the result in a browser and play it |
| `w3-13/point` | Not an agent framework. A way to drive Claude Code from code. **This is the seed of the orchestrator we build in w3-23** |
| `w3-14/open` | Claude desktop, the Cowork tab. Already signed in — **the sign-in flow is never filmed** |
| `w3-14/folder` | Point it at `/Users/Shared/projects/receipts` — **synthetic PDFs generated for this shoot, nothing real** |
| `w3-14/work` | The prompt, then it reading the spreadsheet skill and working through twelve files |
| `w3-14/result` | Open the spreadsheet. A skill, a task list and tools — the same machinery, no terminal |

---

## Verified for Day 3 (checked 2026-08-29, off camera)

- **`claude-agent-sdk` works on the Max subscription with no `ANTHROPIC_API_KEY`.**
  It drives the installed `claude` binary and inherits its auth. Nothing to buy,
  nothing to configure — confirmed with a one-turn round trip.
- **A naive `print(message)` is noisy.** The stream carries `HookEventMessage`,
  `SystemMessage(init)`, `AssistantMessage`, **`RateLimitEvent`** and
  `ResultMessage`. Ask the agent to print only the useful ones — and **keep
  `RateLimitEvent` on screen**, because it reports the five-hour window and ties
  straight into the cost thread running through the week.
- `ResultMessage` carries `duration_ms`, `num_turns` and cost fields: that is the
  number to read out at the end of the lecture.

## CAPTURE B — Day 2 · many agents, safely

**Covers w3-07 … w3-11** (five lectures, 61 min cut, ~1h40m raw).
State: Docker running, browser, phone on the desk, GitHub repo live.

| marker | beat |
|---|---|
| `w3-07/fear` | Why we have been approving every command, and what it costs |
| `w3-07/native` | `/sandbox` in Claude Code. What it does and does not stop |
| `w3-07/cursor` | The same idea in Cursor |
| `w3-08/build` | **The Sprites replacement.** A devcontainer for the repo, written on camera |
| `w3-08/up` | Bring it up. Show the agent inside it, with the repo mounted and nothing else reachable |
| `w3-08/yolo` | `--dangerously-skip-permissions` **inside the container**. The whole point of the day: the mode is only safe because of where it is running |
| `w3-08/prove` | Ask it to do something destructive outside the mount. It cannot |
| `w3-09/open` | claude.ai/code in the browser. Point it at the repo |
| `w3-09/task` | Hand it real work and leave. Come back to a branch |
| `w3-09/review` | Read the diff in the browser and merge |
| `w3-10/phone` | Claude on the phone, Code section. **Filmed off the panel, not a screen share** |
| `w3-10/amp` | The `&` prefix — hand a job off from the terminal and walk away |
| `w3-10/back` | Pick the same job up on the laptop |
| `w3-11/issue` | Open a GitHub issue describing a real change |
| `w3-11/tag` | Tag Claude on it. Watch the action start |
| `w3-11/pr` | The PR arrives. Review it properly and merge |
| `w3-11/close` | Five ways to run this thing without sitting in front of it |

---

## Verified for Day 4 (checked against the shipped binary 2026-08-29)

Agent teams is real in Claude Code **2.1.251** and still gated:

- env flag **`CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`**
- **`teammateMode`** accepts **`in-process`**, **`tmux`** and — new — **`iterm2`**.
  The reference lecture only knows the first two, so *"there is a third one now"*
  is a real, checkable update we can say on camera.
- **Verified live 2026-08-31** with the flag set: teammates are spawned with the
  **`Agent` tool** and coordinated through **`SendMessage`** / **`ListAgents`**,
  each with its own context window. **The session forms an implicit team — there
  is no separate "create a team" command**, which differs from the reference
  lecture's flow. Say that on camera; it is a real, checkable update.
- **`/teams`** and **`/list-agents`** both exist; `/list-agents` now lists live
  teammates, not just sub-agents.
- Teammates **use the leader's model** unless the spawn names one — so the
  "level the playing field" step the reference needs is no longer necessary.

## CAPTURE C — Day 4 · the team builds it

**Covers w3-15 … w3-20** (six lectures, 68 min cut, ~2h raw).
w3-15 is deck-heavy; its only footage is the settings and the docs table.

**One continuous run. Do not stop it.** This is the session that cannot be
repeated cheaply, so preflight twice.

| marker | beat |
|---|---|
| `w3-15/table` | The sub-agents vs teams comparison, **rebuilt in our kit and checked against the live docs on the day** |
| `w3-15/flag` | The settings that switch it on. If it has shipped GA, say so — better footage than his |
| `w3-16/audit` | `.claude/` audited on camera. `/plugin`, `/skills`, `/mcp` — three checks, three clean answers |
| `w3-16/claudemd` | Tighten `claude.md` for seven readers instead of one |
| `w3-16/git` | Commit the baseline. **Branch**, so coming back is one command |
| `w3-16/usage` | **`/usage` — write the number down. This is half the comparison in w3-24** |
| `w3-17/roster` | The prompt: the roster, and **why there is no code reviewer** — it would talk to everyone and drown the channel |
| `w3-17/launch` | Kick off. Accept-edits on. The exploring phase |
| `w3-17/spawn` | Teammates appear. The shared task list, with dependencies |
| `w3-18/flip` | Flipping between teammates. Toggling the task list |
| `w3-18/approve` | **The permissions lesson: approve-once vs approve-always, decided out loud, on two real prompts** |
| `w3-18/parallel` | Four agents live. Token counts climbing |
| `w3-18/surprise` | Whatever it does that we did not expect. **Do not script this** |
| `w3-19/tester` | The integration tester starts |
| `w3-19/finds` | What it finds. Read the failure properly |
| `w3-19/fixes` | Who fixes it, and whether that was the right agent. **Do not intervene** |
| `w3-19/done` | Teammates shut down, team cleaned up |
| `w3-20/usage` | **`/usage` again — the closing number** |
| `w3-20/run` | Start it |
| `w3-20/alive` | **Control Tower on screen. Hold it wide. Vans moving, ETAs counting, heat map shifting** |
| `w3-20/drive` | Drive it by hand: select a van, watch the route draw |
| `w3-20/chat` | Then by chat: reroute one, hold one, ask for the on-time rate |
| `w3-20/warts` | Whatever is broken. Say so plainly |
| `w3-20/commit` | Review the staging area on camera — **no `.env`, no `node_modules`** — commit, push |

---

## CAPTURE D — Day 5 · contenders, verdict, ship

**Covers w3-21 … w3-26** (six lectures, 65 min cut, ~2h raw),
**plus the w3-01 cold-open teaser.**

| marker | beat |
|---|---|
| `w3-21/reset` | **The fair test.** Delete the git-ignored build output and recreate it empty. Say why on camera, or the comparison is worthless |
| `w3-21/cursor` | Cursor, same brief, parallel agents. Same model where we can pick it |
| `w3-21/watch` | Watch it work. Where it is better and where it is worse |
| `w3-21/result` | Run it |
| `w3-22/reset` | Reset again |
| `w3-22/copilot` | Copilot's coding agent, same brief |
| `w3-22/result` | Run it |
| `w3-23/why` | Three tools, three opinions. What if we want our own? |
| `w3-23/build` | **Build a small orchestrator with the Agent SDK from w3-13** — spawn N agents, give each a slice, collect the results |
| `w3-23/run` | Run it on the same brief |
| `w3-23/point` | You are no longer choosing between other people's orchestrators. **This is the top of the ladder** |
| `w3-24/parade` | **All four apps on screen at once. Same data, same window size, same moment** |
| `w3-24/numbers` | Time and tokens for each, from our own `/usage` bookends |
| `w3-24/verdict` | **The verdict, written after this footage exists and never before** |
| `w3-25/docker` | Container, health check, env vars |
| `w3-25/deploy` | Deploy. **Neutral app name, no personal name in the URL** |
| `w3-25/live` | The public URL in the address bar, working |
| `w3-26/montage` | Pickups for the closing montage if anything is missing |
| `w3-01/teaser` | **The cold open for w3-01** — the finished Control Tower, shot last, used first |

---

## After every session

1. `python3 tools/nocode/w3/cut.py <cap> --list` — read the markers back before
   trusting them. Orphan markers mean a beat fell in a pause.
2. `python3 tools/nocode/w3/cut.py <cap>` — write the per-lecture clips.
3. Watch each cut **before writing a word of narration.**
4. Only then purge the raw segments, and only after the cuts are verified.
