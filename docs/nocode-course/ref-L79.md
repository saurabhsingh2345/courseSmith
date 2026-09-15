# L79 — "Week 3 Day 3: Best Practices for Using Claude Code on Large Team Codebases" (10:34)

**Slides only.** The "vegetables" — his word. The one lecture in Week 3 that is
pure advice, and it is unusually good advice.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | slide | Where agents historically got criticised: great on a small greenfield project, **struggle inheriting a massive codebase**. Says it is **"less true now, past the point of inflection from last November"** |
| 0:50 | slide | Frames it: this is the **top half** of the six workflows — the 2025 control techniques, not the 2026 chaos ones |
| 1:20 | **1. Invest in agents.md / claude.md** | **Progressive disclosure** — documentation at *every* subdirectory level, reflecting the interface so an agent can work in that folder **without reading all the files**. And the balance: *"not so much detail that it consumes context, but enough that the right core functions are called"* |
| 3:00 | slide | **2. Structure docs to be navigable.** **`@`-tagging inserts a document in its entirety — usually not what you want.** Better: **describe what a document does and link to it**, so the agent decides whether it needs it. Summarise, then point |
| 4:20 | slide | **3. One consistent team workflow.** If some people tag Claude on GitHub issues and others use Jira tickets, *"things become a complete muddle"*. Agree one process. If you use plugins, everyone uses the same ones |
| 5:40 | slide | **4. Plugins first, then skills.** Add the right plugins (code simplification, feature-dev). Then build **domain-specific skills** that encode how *your* project does things — how to use your market data API, your frameworks |
| 7:00 | slide | **5. Tests — and the contrarian bit.** Have a robust suite, **but stop chasing coverage percentage**. LLMs will chase 80% *"to a fault"*, **over-mocking**, writing brittle tests that exercise every path. You want tests that **survive a reimplementation but break on a logic error**. Push back when you see mock-everything slop |
| 9:00 | slide | **6. You are accountable.** A culture of **human review** and of **rejecting coding-agent slop** — long files, over-defensive code. **The asymmetry problem**: generating code is now trivial, reviewing it is not, and the burden has shifted to the human |
| 10:00 | slide | **7. Bite-sized chunks.** Never *"refactor the whole codebase"* on a big project. Divvy it up so each piece can be **specified, tested and reviewed independently** |
| 10:20 | assignment | Clone a big open-source project in your field, find a TODO, have the agent do it. **And as an anti-test**: try "refactor everything" in a Ralph loop for 10 iterations *"and see what comes out the other end — it's probably not going to be pretty"* |
| 10:34 | end | |

## For our version

- **The `@` warning is the most immediately useful thing in the lecture** and it is
  a direct correction to L66, where his own `claude.md` `@`-inlines `plan.md`.
  **Keep that tension** — it is the honest version: inline the one document that
  must always be there, link everything else.
- **The testing point is the Week 1 through-line landing for the last time.** The
  80% coverage target that produces worthless tests is exactly what we wrote into
  W1 L13's X-block and what Copilot's own plan proposed in W1 L18. **Call the
  callback explicitly** — this is the payoff of a thread running the whole program.
- **The asymmetry problem deserves the one graphic in this lecture**: generation
  cost falling, review cost flat. Everything else is talk over recalled slides.
- **Keep the anti-test assignment.** Telling students to deliberately run the
  thing that fails, to see it fail, is better teaching than another success demo.
- **This is a slides lecture — the exception to rule 2.** Compensate by keeping it
  tight and recalling W1's `ladder` and `stack` rather than new art.
