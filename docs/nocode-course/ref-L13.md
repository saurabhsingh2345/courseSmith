# L13 — "Day 2: The Evolution of AI Coding Workflows: From YOLO to Ralph Loops" (13:04)

**Slides only.** The other load-bearing framework of Week 1, alongside L7's eight
stages and L11's context diagram.

## The six workflows

Opens with a Frozen "let it go" joke carried over from L12. Frames workflow as
"how you organise your activities around this mysterious magical thing" — and
says workflow is what Karpathy's tweet was really about.

**2025 mindset — three levels of trust:**

| # | workflow | what it means |
|---|---|---|
| 1 | **Micromanagement** | Very specific `agents.md`, approve every change, frequently stop it, rewrite, restart. "This was my mode for most of last year" |
| 2 | **Plan → execute → review → test** | Use the agent's **plan mode** to produce to-do documentation, agree the plan, switch to execution. Execute **in phases**: build, code review ("make sure it's not making a mountain out of a molehill"), test, review results, mark complete, next phase |
| 3 | **Spec-driven development (SDD)** | Specify precisely, then let it go. There are even spec languages. The other name for it: **trust but verify** — verify at the end, not with per-phase code review |

**2026 mindset — three more:**

| # | workflow | what it means |
|---|---|---|
| 4 | **YOLO** | Around since 2025, but only recently used for real work rather than hobby projects — which is why he files it under 2026. No permissions, no approvals. "Set it going, go and have dinner, come back" |
| 5 | **Ralph Loops** | Invented by Australian developer **Geoffrey Huntley**, named for **Ralph Wiggum** — naive and optimistic, true to the pattern. An agent is already an LLM looping with tools; a Ralph Loop **wraps that whole loop in a bigger loop**: run it, test whether it has gone as far as it can, generate feedback on the gaps, fold that into the objectives, run the whole thing again — maybe **10 outer loops**. YOLO runs for an hour; **Ralph Loops run overnight** |
| 6 | **Multi-agents** | Many agents with roles — testing agents, feedback agents, manager agents in a hierarchy. **Swarms** and **orchestration**. "What people at the forefront are doing now in 2026" |

**The payoff**: the impressive shooter from L2 was this. Same prompt as the naive
one, dropped into a Ralph Loop, left running, **one-shot with no human feedback**
— the LLM generated its own feedback and re-ran ten times.

## Which approach is right

Not one answer — different approaches for different tasks. His split:

- **Mission-critical** — enterprise software, commercial SaaS, large codebases,
  highly innovative code (e.g. anything touching MCP servers, which current models
  handle badly because the training data is too new and the output isn't
  idiomatic) → **workflows 1-3**
- **MVP, prototype, pilot, greenfield from an empty directory**, with risk appetite,
  producing boilerplate — React apps, HTML, CRUD backends → **workflows 4-6**

He puts most of his own work in the top group. Yesterday's game was YOLO, then a
Ralph Loop for version two.

## The closing point

Aimed explicitly at people early in their careers: **your job is to deliver code
proven to work, and "the LLM wrote it" is no excuse.** Use them, but checking,
validating and picking the right approach for the task is your accountability.

## For our version

- **Two frameworks now stack**: L7's eight stages and L13's six workflows. Build
  them as **one graphic system** so the viewer sees they're two views of the same
  ladder — he says they are "somewhat analogous" and then draws them separately.
- **The Ralph Loop is the best animation opportunity in Week 1**: an inner loop
  spinning, then the camera pulling back to reveal the outer loop wrapping it,
  ten times. Then cut straight to our own overnight result.
- The mission-critical vs greenfield split should be a **two-column decision card**
  the viewer can pause on — it is the most directly actionable slide of the day.
- **Keep the accountability close.** It is the moral spine of the whole course and
  it costs 30 seconds.
