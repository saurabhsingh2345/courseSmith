# L75 — "Week 3 Day 2: 5 Ways to Run Claude Code Remotely: Cloud, Web, Mobile & GitHub" (11:47)

**Entirely screen recording.** The payoff for L74's setup, and the best lecture of
Day 2. Five routes, demonstrated.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | Claude Code | **Way 1 — the `&` prefix.** Any instruction prefixed with `&` runs in the cloud instead of locally. **It fails for him — a live Anthropic production outage.** He says so and moves on |
| 1:20 | talk | **`/tasks`** lists everything running remotely. His is empty because of the outage. Notes **the full conversation context is sent to the cloud** so it continues the work already done |
| 2:20 | terminal | **Way 2 — `claude --remote "<command>"`** from the shell. **Also fails**: *"unable to create remote session, thank you Anthropic"* — down for a couple of days |
| 3:10 | browser | **Way 3 — `claude.ai/code`.** First run asks for network access: **none / trusted / full** — trusted is the usual pick. This is your **cloud environment**. Pick a repo, type a command |
| 4:00 | web | Toy first: "what is two plus two" — *"trillions of floating point calculations… in order to calculate two plus two"* |
| 4:40 | web | Then real work: read everything in `planning/`, **design the market data back end in detail**, write `market-data-design.md` with code snippets |
| 5:40 | talk | **The honest report.** Claude Code on the web is **slower** — ten minutes to write a document — **and flaky**: *"it just seemed to sort of stop and be like, all right, that's it, I'm out."* It is **a research preview**. He asked twice and was patient |
| 7:00 | GitHub | The result arrives as a **pull request**. Compare and pull request → create → **merge** → the document is on main. **1,490 lines**, written entirely in the cloud |
| 8:40 | phone | **Way 4 — the Claude mobile app.** Navigation: chats / projects / artifacts / **code**. Tap it, see the same sessions, **New session**, same interface. *"You should pick some moment to do it, like when you're out at dinner"* |
| 9:50 | GitHub | **Way 5, "a drum roll please" — a GitHub issue tagged `@claude`.** New issue: *"build complete market data back end… with full unit tests"*, tag **`@claude`**, create |
| 10:40 | GitHub | **Seconds later, unprompted: "Claude bot… Claude Code is working"**, a to-do list appears in the GitHub UI. *"You saw my hands were up here, I was gesturing — you knew I didn't do anything"* |
| 11:30 | talk | The scale idea: set up **a lot of issues**, like last week's Jira tickets, and have **many Claude Codes working them in parallel** |
| 11:47 | end | |

## For our version

- **Way 5 is the money shot of Day 2** and it is pure footage: tag an issue, take
  your hands off the keyboard, and watch a bot start working in the GitHub UI.
  **Hold that frame.** No animation anywhere near it.
- **Keep the two failures.** Ways 1 and 2 failing on a live outage is exactly the
  honesty this program trades on, and it is on camera. If ours works, we say ours
  worked and note his didn't — that is the "narration matches the picture" rule.
- **Keep "slower and flaky, it's a research preview."** Ten minutes for a document
  is a real number a viewer needs before they build a workflow on it.
- **The phone shot is genuinely charming** and we should film it properly rather
  than describe it — but **his phone would show his accounts**. Ours needs a clean
  device or a tight crop. Flag as a PII surface.
- **Footage: ~95%.** The only card worth showing is the five routes as a list, and
  even that can be a title over the first shot rather than a slide.
- **Blocked until confirmed**: needs a Claude plan covering web **and** mobile.
