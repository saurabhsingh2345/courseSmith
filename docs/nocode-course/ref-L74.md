# L74 — "Week 3 Day 2: Setting Up Claude Code Sandbox & GitHub Integration" (13:01)

**Entirely screen recording**, and the longest setup lecture in Week 3. Two halves:
`/sandbox` working in 3 minutes, then ~9 minutes of GitHub plumbing.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | VS Code | Housekeeping — he has **uninstalled the plugin** from Day 1 and deleted the files. `.claude/` is empty **except the Cerebras skill**, which he keeps |
| 0:30 | `/sandbox` | "Sandbox disabled". Enter to configure |
| 1:00 | the menu | Three options: no sandbox · **(2)** sandbox allowing bash with regular permissions · **(1) a sandboxed YOLO** — *"commands try to run in the sandbox automatically, attempts to run outside fall back to regular permissions, **explicit ask/deny rules always respected**"*. He picks **1** |
| 1:50 | the menu | **Overrides** (fallback vs strict) and **Config** (allow/deny lists) — "a few things in there that look like they probably should be denied, which is just as well" |
| 2:30 | talk | What it can do unattended: **bash, and read/write files freely**. **Web searches still ask** |
| 3:00 | terminal | A real task: research the **Massive (formerly Polygon)** market data API, design the market data interface, and design a **market data simulator** — three documents into `planning/` |
| 4:00 | Claude working | Runs at length with no permission prompts. Then **it does ask** — for the **Context7** plugin he'd forgotten he had, which is a better source than web search |
| 5:00 | result | Three documents written. **"I haven't reviewed them because I'm not going to — we're going to be much more trusting this week"** |
| 5:30 | talk | Read the sandbox docs for granular config and **the security concerns**. Notes **native Windows support "coming soon"** |
| 6:00 | browser | **`claude.ai/code`** — "code with Claude anywhere". **Connect to GitHub** → authorize → **install the Claude GitHub app** into selected repositories |
| 7:20 | Claude Code | **`/install-github-app`**. It complains the **GitHub CLI** is missing and prints the exact install command per OS |
| 8:00 | terminal | `brew install gh`, then **`gh auth login`** — browser opens, a code is printed in the terminal, paste it in. Also does `git add .` and `git commit` because **everything must be checked in** |
| 9:30 | Claude Code | `/install-github-app` again, select the repo, **Configure** in the browser, save, back to the terminal, press enter |
| 10:30 | the key step | **"Select the GitHub workflows to install"** — space to tick both, enter. Then **a long-lived token that comes with your Claude subscription**, and authorize |
| 11:30 | GitHub | It opens a **pull request that adds the workflows**. Press **Create pull request**, then **Merge**, then confirm |
| 12:20 | repo | The result: **`.github/workflows/claude.yml` and `claude-code-review.yml`** now exist. "You've really bridged Claude with your GitHub repo" |
| 13:01 | end | |

## For our version

- **This is a setup lecture and setup lectures age fastest.** He says so himself
  ("Anthropic is changing this from time to time"). Ours must be **dated on
  screen** and shot on the day.
- **The `/sandbox` half is the strong half** and it is 100% footage: the menu, the
  choice, then a long unattended run that only stops for a web search. **Hold on
  the run** — an agent working with no prompts is the whole promise of the day.
- **Keep "I'm not going to review them, we're being more trusting this week."**
  It is an honest admission of the trade and it sets up the week's chaos/control
  spine.
- **The GitHub half is nine minutes of plumbing.** Beat-for-beat we keep it, but
  this is where **2x-4x ramping** earns its place (W1 house rule: ramp waits at
  2-3x, not 10x) — the token paste and browser round-trips are dead air otherwise.
- **PII: the GitHub flow puts his account, repo list and email on screen.** Ours
  runs on his handle (already approved for the kanban repo) but the **repo picker
  shows every repository** — same class of leak as the VS Code Workspace Trust
  page. **Select "only selected repositories" and keep the list off-frame, or
  blur it.** See [[coursesmith-no-pii-in-renders]].
- **Blocked until confirmed**: this needs a Claude plan that includes Claude Code
  on the web.
