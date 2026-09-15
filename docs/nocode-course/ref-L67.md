# L67 — "Week 3 Day 1: Creating Custom Slash Commands in Claude Code" (10:53)

**Entirely screen recording.** The first pro feature, and the simplest.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | VS Code | Two places a slash command can live: **`.claude/` in the project**, or `.claude/` in your **home directory** for every project. He does the project |
| 0:30 | file tree | Make a folder — **`.claude/commands/`**. Each markdown file in it becomes a slash command |
| 1:00 | new file | **`doc-review.md`**. Notes the convention: lowercase with hyphens |
| 1:20 | the file | The whole body is the prompt: *"Review the documentation file in the planning folder called **`$ARGUMENTS`** and add questions, clarifications, or feedback to a new section at the end"* — plus **"along with any opportunities to simplify"** |
| 2:00 | terminal | `claude`. **"Opus 4.6 is here"**. Types **`/doc-review plan.md`** — the command is really there, and `plan.md` lands in `$ARGUMENTS` |
| 2:40 | Claude working | Reads the file, searches project state, finds the scaffolding empty, writes a review section |
| 3:30 | `plan.md` | **Tip: close an open preview first — it caches.** Reads the new "Document review" section |
| 4:00 | the review | Real findings: an orphaned "Brand colors" header, a docs-vs-planning inconsistency, a script naming mistake — **"that was my mistake, not Claude's"** |
| 5:30 | the review | **He disagrees with most of the simplifications**: keep the Massive API, keep `user_id` on all tables for future multi-user, keep Playwright. Accepts one: **don't stream the chat response, just return it** |
| 7:00 | talk | *"This is where you have to exercise some discretion… this is where you add value and where you need to be the boss."* |
| 7:40 | `/context` | Shows the conversation history **is** in context — because a command is just a prompt. **This is the setup for sub-agents**: a sub-agent would have done this work outside the context and the main Claude would only discover the new section in the file |
| 8:40 | Claude | Feeds his edits back: "I've updated the comments… removed the simplifications I disagree with. Please now review the remaining issues and incorporate the solutions throughout `plan.md`" |
| 9:40 | wrap | **The second way to make a slash command: a skill.** Skills give you one automatically — `/cerebras-inference` is there because the week-two skill exists. "Most of the time, people just focus on skills now" |
| 10:53 | end | And if `.claude/` is checked into git, **everyone on the team gets the commands** |

## For our version

- **The `$ARGUMENTS` reveal is the beat.** Type `/doc-review` and watch it
  autocomplete — that lands better than any diagram.
- **Keep the disagreement.** Him rejecting most of the agent's simplifications is
  the most valuable 90 seconds in the lecture and it is the same "be the boss"
  spine as W1. Ours must disagree with something **real** that our own agent
  actually proposes — written after the shoot, per
  [[coursesmith-narration-matches-picture]].
- **The `/context` beat is the hinge into sub-agents.** We already built the
  `stack` graphic for context in W1 L11 — **recall it here for three seconds**,
  then go straight back to footage.
- **Footage: ~95%.** The only graphic worth having is command-vs-skill-vs-subagent
  (which goes in context, which does not) — and that lands better at the end of
  L69 than here.
