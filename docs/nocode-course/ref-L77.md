# L77 — "Week 3 Day 2: Cloud Sandbox Recap: YOLO Mode with Sprites.dev & GitHub PRs" (7:54)

**Screen recording, then the day's recap.** Short, and it closes Day 2.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | terminal | **`Ctrl+O`** — the week-two pro tip — to read the full conversation and confirm **most tests passed, some failed, and it accounted for them honestly** |
| 0:40 | terminal | Asks it to push. **It can't — the sprite isn't logged into GitHub.** It tells him to run `gh auth login` |
| 1:10 | terminal | `Ctrl+C` twice out of Claude, **`gh auth login` on the remote box spawns a browser on his local machine**, logs in |
| 1:40 | terminal | Back into `claude` — **but the conversation started over**. **`/resume`** puts him back where he was. It then branches, pushes, and opens a PR |
| 2:30 | GitHub | `market-data-review.md` — compare, create PR, **merge** |
| 3:10 | terminal | On a roll: *"switch to main and pull, then carry out all the fixes and improvements you documented in the review, **keep working until all tests pass**, then push your new branch"* |
| 4:00 | talk | **The best passage in Day 2**: *"it's just been the most amazing experience sitting here watching this happening. I haven't approved anything."* Fixing bugs one by one, re-running tests, reporting, branching, pushing — **"super productive AND secure… the best of both"** |
| 5:20 | GitHub | Another round of feedback, another PR, merge, confirm |
| 5:50 | talk | **It isn't Claude-specific** — use a sprite for Codex, OpenCode, whatever you picked |
| 6:10 | recap | **Blue / purple / yellow one more time**, in detail |
| 7:30 | close | **"You are 80 percent through."** Tomorrow: large codebases, plus "spicy stuff" |
| 7:54 | end | |

## For our version

- **"I haven't approved anything" is the thesis of Day 2** and it should be the
  line the lecture is built around. It is also the exact payoff of the
  approval-fatigue point from L72 — sandboxing is what makes YOLO honest.
- **Keep the GitHub-login failure.** It is a real friction on a fresh box and the
  fix (`gh auth login` remotely, browser opens locally) is genuinely useful.
- **`/resume` after the conversation resets is a small gift** to anyone who has
  lost a session. Keep it.
- **The blue/purple/yellow recall is the one graphic Day 2 needs**, and we build
  it once in L73. Everything else is screen.
- His "80% through" beat is computed from **his** position — ours uses ours.
- **Footage: ~85%.**
