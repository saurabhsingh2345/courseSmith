# L69 — "Week 3 Day 1: Subagents vs Agent Teams in Claude Code with Codex Review" (8:41)

**Screen recording, then a talking close.** Finishes sub-agents, then explains
Agent Teams without using them (they land on Day 4).

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | VS Code | Deletes `review.md` "with its silly `.env` worry" |
| 0:20 | new agent file | **`codex-reviewer.md`**. The body is emphatic: *"You are using a different AI agent… You **must execute the following shell command**… **Do not review yourself**"* — with `codex exec` inline |
| 1:10 | terminal | *"Use your codex-reviewer sub-agent to carry out a review of planning/plan.md"* |
| 1:40 | Claude working | It reads `plan.md` itself first, then **does** fire the bash command. Codex runs inside a Claude Code sub-agent — *"feels very sub-agenty"* |
| 2:40 | `review.md` | Codex's feedback: strengths, risks, gaps. Several misses — worried about a skill that **is** defined, doesn't know Massive is the new Polygon. Verdict line: *"a strong blueprint that should enable parallel execution"* |
| 4:00 | edit | Generalises it: rename to **`change-reviewer.md`**, and review **all changes since the last commit** rather than one file. Says plainly you can drop Codex and let Claude do it — he keeps it "because I enjoy having a different LLM in the mix" |
| 5:20 | terminal | Restart, delete `review.md`, *"use the change-reviewer sub-agent to review changes since last commit"*. It shells out, writes the review |
| 6:20 | `/context` | **The payoff shot.** All that to-and-fro — working out what changed, deciding how to look, writing findings — **none of it in the main context**. The window is basically clean |
| 7:00 | talk | **Sub-agents vs Agent Teams.** A sub-agent takes *one* task, runs isolated, returns to the main Claude Code. Always that one relationship. (Can be given project-level memory, but that aside, it is one task.) |
| 7:50 | talk | **Agent Teams** assemble a *group* of Claude Codes working collaboratively. The x-factor: **they can talk to each other**, not just through the main agent — a tester agent giving feedback straight to the front-end and back-end agents. Long-running presence, challenge each other. **Experimental** at time of recording |
| 8:41 | end | |

## For our version

- **`cursor-agent -p -f` replaces `codex exec`** in both the `codex-reviewer` and
  `change-reviewer` files. Rename ours **`cursor-reviewer.md`** and
  **`change-reviewer.md`**.
- **Keep "do not review yourself" verbatim in spirit.** It is the whole trick —
  without it the agent quietly does the work itself, which is the goal-substitution
  failure from Week 1 in a new costume. If ours does that on camera, keep it.
- **The `/context` shot is the single best frame in Day 1** and it is pure
  footage. Hold it. Recall the W1 `stack` for three seconds against it, no more.
- **Sub-agents vs Agent Teams is the one genuinely conceptual beat in Day 1** and
  he only says it. This earns a graphic: one hub-and-spoke (sub-agent, one task,
  back to the centre) against a mesh (agents talking to each other). Build it
  once, **recall it on Day 4** when Agent Teams actually run.
- **Footage: ~85%**, the highest animation share of Day 1 and still mostly screen.
