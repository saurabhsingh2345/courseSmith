# L70 — "Week 3 Day 1: Claude Code Hooks: Auto-Trigger Reviews with Events & Commands" (9:07)

**Entirely screen recording.** The feature he explicitly says most people don't need.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | talk | Framing, and it is unusually honest: **"most of the time you don't need this"**. Get the gist, and when the moment comes, go read the docs |
| 0:40 | talk | **A hook is an event triggering something you fixed in advance.** Events: about to make a tool call, about to run a shell command, finishing work |
| 1:20 | talk | Three worked examples: force **`uv run` instead of `python`** on every shell call · **notify you when it stops for permission** so you don't come back two hours later to find it waiting · **do something when Claude finishes**. Notes **Ralph loops are implemented this way** — on stop, send a prompt saying "I don't think you've finished, try again" |
| 3:00 | new file | **`.claude/settings.json`** — where hooks live |
| 3:20 | talk | **Three things a hook can do**: run a **command** (a shell command), send a **prompt**, or spawn an **agent** |
| 4:00 | `/hooks` | The interactive menu. The events: `PreToolUse`, `PostToolUse`, `Notification`, `UserPromptSubmit`, `SessionStart`, **`Stop`** — and later he spots more: `SubagentStart`/`Stop`, **`PreCompact`** ("shove in extra reminders just before it compacts"), `SessionEnd`, permission, setup, task-completed |
| 5:00 | talk | **Command is the reliable one.** Prompt and agent hooks have permission constraints — they can't write files — and "take a lot of experimenting to get predictable" |
| 5:40 | settings.json | The `Stop` hook runs **`codex exec "review the changes since the last commit and write results to planning/review.md"`**. He also deletes the earlier command and sub-agent **so there is only one way to review** |
| 6:40 | terminal | Test: *"please make a concise readme.md for the project"*. Claude writes it, he approves the edit, **the stop hook fires**, Codex runs behind the scenes and writes `review.md` |
| 8:00 | result | **Claude has no idea the review happened.** The review notices the removed sub-agent and flags it as a possible workflow regression |
| 8:40 | talk | Recap: don't use hooks until you need them; `settings.json` or the menu; event → hook; command is the predictable one |
| 9:07 | end | |

## For our version

- **`cursor-agent -p -f` replaces `codex exec`** in the Stop hook.
- **Keep "you probably don't need this".** An instructor telling you to skip a
  feature is rare and it buys trust for everything he does insist on.
- **The payoff is that Claude never knows the review ran.** That is a footage
  beat — the trace ends, the hook fires, a file appears that the agent has no
  memory of. Hold on the file tree as `review.md` appears.
- **The Ralph-loop reveal is a genuine callback** to W1 L13, where we built the
  `loops` animation. **Recall `loops` for a few seconds** — one of the very few
  places in Week 3 where animation is the right answer.
- Also worth keeping: **`PreCompact`** as a hook event, because it connects
  straight to W1 L11's compaction lecture.
- **Footage: ~90%.**
