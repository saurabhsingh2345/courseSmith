# L83 — "Week 3 Day 4: Claude Code Agent Teams — Swarms and Orchestration" (11:37)

**Slides.** Opens Day 4, the biggest day in the course. The theory lecture that
earns the next five.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | slide | **The rollercoaster callback lands.** *"At the start of our three weeks I told you to expect a rollercoaster… we've reached the very, very top, you're looking down, and we're about to take the plunge."* A purple day. Core skills |
| 0:50 | slide | **Stages 7 and 8 of the Steve Yegge chart** — the one from Week 2. **7 = swarms** (tons of agents, chaos). **8 = orchestration** (reel it in, structure, some in charge of others, LLMs managing the process) |
| 1:40 | slide | And the honest qualifier: **it is not binary.** Everything today sits somewhere on a continuum between wild swarm and rigid orchestration, *"and where you are depends on the choices you make and how crazy you want to go"* |
| 2:20 | slide | **What agent teams are.** Multiple Claude Code instances, coordinated. **One is the team lead** — assigns tasks, manages collaboration. The rest are **teammates**: they work independently, **they can message each other**, and **you can talk to any of them directly without going through the lead** |
| 3:20 | slide | **The distinction that matters — teammates vs sub-agents.** You cannot send a message to a sub-agent. It exists to serve the main Claude Code, it does not work independently, it lives only for its task |
| 4:10 | slide | **Three use cases.** (1) Parallel research — Wikipedia, Stack Overflow, three other places, come back with everything. (2) **Independent modules** — you could just open several terminals, and they *could* talk through `.md` files, but *"it would be a little bit flaky"*; teams give a real channel. (3) **Horizontal split by layer** — front-end, back-end, LLM engineer |
| 6:00 | slide | **The balance to strike**, and it is the load-bearing idea: separate enough responsibilities that they work independently, **enough communication to collaborate, not so much that they grind down** in backwards-and-forwards churn |
| 6:50 | slide | **Anthropic's comparison table**, read across. Sub-agents: own context, returned to caller, communicate only upward, main agent manages all work, focused tasks, **summarise back so they are token-efficient**. Agent teams: **full independent context, not returned**, message each other, **one shared task list managed by the lead**, best for complex collaborative work, **and they can be expensive** |
| 8:20 | slide | The last bullet, quoted: **sub-agents for quick focused workers that report back; agent teams when teammates need to share findings, challenge each other and coordinate on their own** |
| 8:50 | slide | **The five steps.** (1) `settings.json` — enable the experimental agent-teams flag, and **`teammateMode`**. Two modes: **in-process** works everywhere, all agents in one screen, **shift-up/down to flip**; **tmux** splits the screen per agent but is Mac/Linux only and needs extra installs or iTerm2. He uses in-process |
| 10:10 | slide | (2) Prompt: *"create an agent team to…"*. (3) **Shift-tab into delegate mode** so the lead doesn't do the work itself. (4) **Shift up/down** to flip between teammates. (5) *"ask the X teammate to shut down"*, and **"clean up the team"** to stop everything. Also: **prompt it to wait for teammates before proceeding** — Anthropic flags that as a common failure |
| 11:00 | slide | **Three warnings.** **Invest in `claude.md`** — it loads into every agent's context, so it is your one shot at starting them all correctly. **Expect it to be pricey** — free models will not be reliable; *"watch your costs, stop it if you become uncomfortable"*. **Be willing to rerun the whole thing** — it may go off the rails, git back and try again |
| 11:37 | end | *"Let's go to VS Code. Let's try Claude agent teams for real."* |

## For our version

- **This is a slides lecture and it has to be — but it is the one where our
  graphics earn their keep.** Two are worth building well: the **swarm↔orchestration
  continuum** with today's four tools placed on it, and the **sub-agents vs teams
  table** built column by column rather than pasted.
- **Do not screenshot Anthropic's table.** Rebuild it in our kit, and **check every
  row against the live docs on the shoot day** — this is exactly the class of fact
  that went stale on us in W1 L11.
- **Reuse the Yegge ladder from Week 2 rather than drawing a new one**, and light
  stages 7 and 8. The recall is the point.
- **"They can message each other and you can talk to any of them" is the single
  sentence that separates this from everything before it.** Give it its own frame
  with the arrows drawn.
- **Keep all three warnings, especially the money one.** We are about to spend real
  tokens on camera in L84–L85; the student needs the warning *before* they copy it.
- **Verify the settings key and `teammateMode` values on the day.** Experimental
  flags rename. If it has graduated out of experimental by our shoot, **say so on
  camera** — that is a better beat than his.
- **Footage: ~20%** — the exception lecture. Compensate by making L84–L88 nearly
  all footage.
