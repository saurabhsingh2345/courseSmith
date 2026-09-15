# L86 — "Week 3 Day 4: GSD — Spec-Driven Design Meets Multi-Agent Orchestration" (13:22)

**The other end of the spectrum.** Reset the repo to nothing, install GSD, and
start the same project again — this time on rails. Mostly footage with a long
docs-reading stretch in the middle.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | VS Code | **The reset, and why it matters.** Even after `git checkout main`, **`frontend/` still has git-ignored files** — so he **deletes the whole folder and recreates it empty**, *"so we're definitely starting from scratch and not giving any hints away"*. A fair-test beat |
| 0:40 | editor | Checks `.claude/settings.json`: plugins are still there, **the experimental agent-teams flag is gone**. Quits the previous Claude Code. *"That is a wrap on that previous version"* |
| 1:20 | talk | **SDD — spec-driven design — recapped.** Tools that walk you through building a spec and then executing it. **`feature-dev` was a nod to that.** His read: it was a big deal with Claude Code and **is less so now — *"it feels like a slightly laborious approach"*** and you can give up more control now |
| 2:10 | browser | **GSD.** *"Stands for something which I will bring up and you can see visually, but I'm not going to say it"* — the repo says **"Get Shit Done"**. **Lightweight meta-prompting** — prompts that control a set of pre-built prompts. Context engineering + SDD. **Works with Claude Code, OpenCode and Gemini CLI** |
| 3:10 | browser | **Its founding problem: context rot** — quality degradation as the window fills. That is what it set out to solve |
| 3:40 | browser | **The testimonials, read out.** *"I've done SpecKit, OpenSpec and Taskmaster — this has produced the best results for me"* (he names all three as the traditional SDD tools worth looking up). *"By far my most powerful addition to Claude Code."* And the pitch: **"for people who can describe what they want and have it built correctly without pretending they're running a 50-person engineering org"** |
| 5:00 | browser | **GSD recommends `--dangerously-skip-permissions`** — *"stopping to approve 50 git commits defeats the purpose"*; the alternative is granular permissions. **He declines**: *"if I were to do that, I would do it on the sandbox as we established before, but I'm happy to be approving as we go"* |
| 6:00 | terminal | **`git checkout -b finally-gsd`** — a named branch to keep the two experiments apart |
| 6:30 | terminal | Runs the installer → **which runtime? Claude Code. Where? this project only.** Done. It creates **agents, commands, a state folder and hooks** — *"similar to installing a plugin, it's just put all of this stuff right here"* |
| 7:40 | talk | **His framing of GSD vs agent teams, and it's the thesis of Day 4:** *"similar to the agent teams concept we've just worked with, but more opinionated — more constructs designed to put you into a particular way of doing it"* |
| 8:20 | browser | **The command sequence.** `gsd:new-project` (spawns parallel agents to analyse stack and conventions) → **`gsd:discuss`** (shape it) → **`gsd:plan`** a phase → **`gsd:execute`** → **`gsd:verify`** → next milestone. Plus **a quick mode** that skips planning. Also `audit-milestone`, `complete-milestone`, `new-milestone` |
| 9:40 | browser | **Why it works, per the docs**: strong pre-written prompts, preset agents, **and a fixed markdown file structure that holds all the state** — `project.md`, research, requirements, roadmap, state, plan, summary, todos. *"An opinionated structure around having multiple agents work on your project"* |
| 11:00 | talk | Aside: **the author is an LA-based EDM musician**, and the GitHub stars show the following |
| 11:20 | Claude | Launches Claude. **`/context`** — lots of GSD in there already. Leaves the rest of the setup untouched |
| 11:50 | Claude | **`/usage`, written down again.** Last time 0 / 8 / 2. Now **8% used, 9% of the week, 2% Sonnet** — because he's been on Opus. **The baseline for the comparison** |
| 12:20 | terminal | `git add .` → `git commit -m "start of GSD process"`, so he can return to this exact point |
| 12:40 | Claude | **`/gsd:new-project`.** *Map the codebase first?* Yes. **Spawns four mapper agents in parallel** — and he names it: **this uses sub-agents, not agent teams.** *"This is an alternative to using Claude's agent teams"* |
| 13:00 | Claude | Mapping done, markdown files written. Then, before answering "what do you want to build?", he stops to **level the playing field: `/gsd:settings` → model quality = 1 (quality, not the recommended balanced)** so it uses Opus like the teams run did. *"I want to be fair about it."* Plan researcher, plan checker, execution verifier — all yes |
| 13:22 | end | |

## For our version

- **The reset beat is the fair-test beat — keep it and make it louder.** Deleting
  the git-ignored `frontend/` is the difference between a real comparison and a
  fake one. Put it on screen and say why.
- **Same for `/gsd:settings` → quality.** Two runs, same model, or the comparison
  in L91 means nothing. **These two beats are what make Day 4 a genuine
  experiment rather than two demos.**
- **The YOLO decision, third time asked, third time answered the same way**: not
  outside a sandbox. By now it should read as a rule, not a preference. **Ours
  runs in the Day 2 sandbox and we say so.**
- **The docs-reading stretch (3:00–11:00) is the risk in this lecture** — eight
  minutes of a browser scrolling. **Cut it to the four claims that matter**
  (meta-prompting · context rot · the file-structure state machine · the YOLO
  recommendation) and put the rest under a single held frame of the command
  sequence. This is where our runtime savings in Day 4 come from.
- **We must verify GSD still installs and still names its commands the same way**
  on shoot day. It is a fast-moving third-party project — if the flow changed,
  we shoot the new flow and say the docs moved.
- **Substitution: none.** GSD is CLI-agnostic and we run it on Claude Code, exactly
  as he does.
- **Footage: ~85%.**
