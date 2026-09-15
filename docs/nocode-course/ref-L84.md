# L84 — "Week 3 Day 4: Setting Up Claude Code Agent Teams for Full-Stack Development" (10:29)

**Footage, start to finish.** The setup lecture: clean the project, install
plugins, tune `claude.md`, branch, flip the flag, and launch a seven-agent team.
Ends the moment the team spawns.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | VS Code | **Audit `.claude/` on camera.** `agents/` empty · `commands/` empty · `skills/` has **only the Cerebras skill** (kept) · `settings.json` empty · `settings.local.json` holds permissions · **sandbox off**. *"We want to keep everything as vanilla, as simple as possible"* before going crazy |
| 1:00 | VS Code | `.github/` has the actions — **tagging Claude is available but not used today**. `market-data-summary` is present from Day 3 |
| 1:30 | Claude | `/plugin` → **no plugins installed**. `/skills` → **one, Cerebras inference**. `/mcp` → **no MCP servers**. Three checks, three clean answers |
| 2:30 | Claude | **Installs three plugins, and explains the exclusion rule**: *"I want to avoid plugins that themselves spawn sub-agents or we're going to lose control."* **front-end design** (production-grade UI, kills the LLM look) · **Context7** (current API docs) · **Playwright** (browser + tests). **Rejects code-simplifier** — it would add a sub-agent. All installed at **project scope** |
| 4:30 | editor | **The `claude.md` edit, and the best teaching moment in the lecture.** He replaces `@plan.md` with: *the key document is `plan.md`, included below; the market-data component is already complete, summarised here with more detail in that folder; consult these docs only when required.* **`@` inserts the entire document; a path in backticks lets the agent decide** — *"a bit like the way we build skills files… progressively add to context when required"*. **This is L79's advice being applied one lecture later** |
| 6:00 | terminal | **Git housekeeping.** `git status` → `git commit -m "ready for teams"` → `git push`. *"A good baseline point we can use for going crazy"* |
| 6:40 | terminal | **`git checkout -b agent-teams`.** A branch to one side, so coming back is easy |
| 7:00 | editor | **`settings.json`.** Adds the experimental agent-teams env var and **`teammateMode: in-process`**. Catches **a missing comma that was also wrong on his slide** and says so |
| 7:40 | Claude | Launches Claude. **`/usage` first, and he writes the numbers down: 0% today, 8% of the week, 2%.** *"We'll see where this ends up being"* — the cost experiment set up honestly |
| 8:10 | Claude | **`/context`.** *"Gosh, there's lots of Playwright-related tools — I'm wondering whether that was a mistake"*, plus a lot of memory from the plan |
| 8:40 | Claude | **The prompt, pasted.** Create an agent team to build the entire project: **front-end engineer · back-end engineer · database engineer · LLM engineer** (using the Cerebras skill) · **integration tester** · **DevOps engineer** for the Docker container. All engineers write their own unit tests. Notes you can also let it choose its own team |
| 9:20 | talk | **The agent he deliberately left out.** He was tempted by a code-reviewer to check every agent's work — *"but that requires conversation with every single agent, more backwards and forwards… more noise than we would like"*. **Shows the reasoning, not just the decision** |
| 9:50 | Claude | Shift-tab through the modes → **accept edits on**. Kicks it off. *"I am a bit anxious, I have to admit."* **Shift-tab did not turn on delegate mode** — he notes the docs may be behind |
| 10:00 | Claude | **~5 minutes of exploring**, then the spawn. **DB engineer and front-end engineer are live**, DB already building a SQLite layer. The team panel: **back-end, LLM, DevOps and integration testing all "blocked by"** — *"a little dependency project plan going on"* |
| 10:20 | Claude | **Shift up/down** flips between **team lead → DB engineer → front-end engineer → hide**. *"They are all on the go"* |
| 10:29 | end | |

## For our version

- **100% footage. No slides at all.** Every beat here is a screen doing something.
  If a graphic appears in our cut it is a mistake.
- **The `@` vs backtick-path edit is the highest-value 90 seconds in Day 4.** Shoot
  it as a proper editor beat — the old line, the new line, both on screen — and
  **name L79 out loud**. It is the one place the advice lecture pays off visibly.
- **Keep `/usage` before and after.** He writes the numbers down; **we put them on
  screen as a lower-third at the start and again at the end of L85.** That turns a
  throwaway into the cost lesson the whole day needs.
- **Keep the rejected agents.** The code-simplifier plugin and the code-reviewer
  teammate, both rejected for the same reason — noise and lost control. Two
  rejections in one lecture is a pattern worth naming.
- **Our team roster can differ, and should.** Our project is bigger; the roster
  should match what we actually build. Keep the shape (vertical + horizontal
  split, one tester, one DevOps, no reviewer) and the count around seven.
- **Never film `/plugin` or `/mcp` with anything personal installed** — audit the
  real machine before rolling, and shoot from `/Users/Shared/projects`.
- **Verify the settings key on the day** — see L83. If agent teams have shipped
  non-experimental, the flag beat becomes *"you no longer need this"*, which is
  strictly better footage.
- **This lecture ends mid-air on purpose.** Do not resolve it. L85 is the payoff.
- **Footage: 100%.**
