# L68 — "Week 3 Day 1: Building Agents & Sub-Agents with Claude Code and Codex CLI" (11:52)

**Entirely screen recording.** The big topic of Day 1. **Ours uses Cursor where he
uses Codex** — see WEEK3.md rule 4.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | VS Code terminals | **"Multi-agents" is the boring part**: press `+`, type `claude`, again, again. Three Claude prompts. One builds the front end, one the back end, one the tests |
| 1:00 | talk | And he **warns against doing it now**: `plan.md` is not fleshed out enough, the agents would not know about each other, **"a fiasco would emerge"**. Boundaries are not yet real |
| 2:00 | browser | Installs the **Codex CLI** — "the equivalent OpenAI has to Claude Code". `npm` or Homebrew. Runs `codex`, notes the first run may push you through an OAuth login |
| 3:20 | terminal | The key move: a plain shell command — **`codex exec "please review the file planning/plan.md and write your feedback to planning/review.md"`** |
| 4:00 | `review.md` | Codex wrote the file, fast. He reads it, unimpressed by one item: *"Massive API refers to Polygon IO, but the name is a placeholder… **this is Codex 5.3, we expect more**"* |
| 5:00 | talk | **The point**: a powerful model, from a different vendor, running as a separate agent off one shell command. "This opens the door to lots of new possibilities" |
| 5:40 | talk | **Sub-agents proper.** You have used them already — Claude Code ships with them (an explore one, a plan one). They take a task, run it **in a separate context**, return the result. Sometimes on cheaper models like Haiku. **Context is the big win**, plus parallelism |
| 7:00 | `/agents` | The menu: "create new agent", a `code-simplifier` that came from a plugin, and the built-ins. **"But we're pros"** — make the file yourself |
| 7:40 | talk | Where sub-agents can live: project `.claude/`, home `.claude/`, passed as flags at launch, or **inside a plugin** |
| 8:10 | new file | **`.claude/agents/reviewer.md`**. Front matter: `name: reviewer`, `description: carry out a comprehensive review when requested`. **`tools` and `model` are optional — omit them and it inherits from whatever spawned it** |
| 9:00 | the file | Body: "You review the file `plan.md` and you write your feedback to `planning/review.md`" |
| 9:30 | terminal | Restart Claude, `/agents` — the reviewer is there. **"You don't call an agent with a slash command."** He asks in prose: *"use the reviewer agent to carry out a review"* |
| 10:30 | Claude working | The reviewer sub-agent runs, visible in the trace, writes `review.md` |
| 11:00 | `review.md` | A **false-positive security finding** — it insists `.env` is not gitignored. *"This often seems to happen with the sub-agents"* |
| 11:30 | talk | Two distinctions: **you cannot call it directly, it decides**; and the work happened **outside the main context**. Says the phrasing should be *"use the reviewer **sub-agent**"* |
| 11:52 | end | Teases doing the review with a different AI |

## For our version

- **Substitute throughout**: `cursor-agent -p -f "review planning/plan.md and write
  your feedback to planning/review.md"`. Same beat, same shape, and it keeps the
  "different vendor reviewing your work" point exactly intact.
- **Keep the multi-agent anticlimax.** "Just open three terminals" being a
  let-down, followed by "and here is why you should not do it yet", is honest and
  sets up boundaries as the real skill.
- **Keep the false-positive finding.** An agent confidently wrong about a
  security issue is the Week 1 through-line resurfacing, and it is on camera.
  Ours must use whatever our own run actually produces.
- **Footage: ~95%.** The one graphic worth it is context-with-a-sub-agent vs
  context-without — and we already built `stack` for that in W1 L11. Recall it
  for a few seconds, no new artwork.
