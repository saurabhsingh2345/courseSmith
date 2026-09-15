# L93 — "Week 3 Day 5: Final Deployment, Course Wrap-Up & Coding Agent Best Practices" (9:28)

**Deploy in 15 minutes, then the closing argument.** Half footage, half the
best-practices summary the whole course has been assembling.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | browser | **"But wait, there's more."** He had **a 15-minute conversation with the agent about where to deploy**, agreed on **Fly.io — the company behind sprites.dev — as a natural fit for a container as-is**. It wrote a script, they ran it, **it's live** |
| 0:40 | browser | **The URL bar is the shot.** A public `.fly.dev` address. **AI chat works, portfolio there, live market data, env vars set correctly.** *"Deployment took all of 15 minutes, and you should do it too"* |
| 1:20 | slide | **The orchestrator scoreboard one last time.** *"My favourite experience is with Claude Agent Teams, and the winner of the prize in this case was probably Codex."* And the reason the sprite mattered: **you can run a coding agent in YOLO mode and feel completely secure** |
| 2:00 | slide | **The assignment.** Take the repo and give one of these a try. **Free or cheap model → expect a harder time, and something disciplined like GSD may be your route through.** Splashing out → challenge it to do something different. **Suggestion: add the users table and login from Week 2's legal SaaS.** *"Choose your own adventure — you've got a really robust specification"* |
| 3:00 | slide | Share it — Udemy Q&A, LinkedIn, tag him and he'll amplify. **And support each other's posts** |
| 3:40 | slide | **THE KARPATHY TWEET, RE-READ.** The one that opened the course. *"This time it's like I was speaking to you in a foreign language at the beginning, and now, once you've learned all the vocabulary, hopefully it all falls into place."* The full list — **agents, sub-agents, prompts, context, memory, modes, permissions, tools, plugins, skills, hooks, MCP, LSP, slash commands, workflows, IDEs** — and *"a powerful alien tool handed around, except it comes with no manual"* |
| 5:20 | talk | **And the honest concession.** *"The bad news is it's not like I've given you total transparency and clarity, because this world is moving so fast. When we look at things like Gastown it still feels completely bewildering for me, let alone for you."* But: **you've dimensioned the landscape**, you know plugins vs skills and when to use each, you've worked in the IDEs and the CLIs — **and you know things Karpathy didn't even mention: sandboxing, YOLO mode, Ralph loops, GSD** |
| 6:20 | slide | **The single most important thing: be willing to experiment.** *"There's not necessarily one right answer… if it all goes wrong, just simply start again"* |
| 6:50 | slide | **The six-workflow chart, final appearance.** **Top row / yellow** for mission-critical, large codebases, **code at the forefront that may not be in the training data** — markdown files, incremental, **SDD (GSD)**, trust-but-verify. **Purples** for MVPs, new builds, risk appetite, boilerplate — **YOLO but in a sandbox**, Ralph loops, **swarms and orchestration** |
| 7:40 | slide | **Choose based on**: project maturity, risk appetite, **your own skills** (seasoned pro → CLI; still getting used to it → IDEs), and plain preference |
| 8:00 | slide | **The order of operations.** **1. Plugins first** — official, popular, fit-for-project (feature-dev, code simplifier, front-end). **2. Then skills** for your project. **3. MCP where an MCP is exactly what you need** (Context7 — though there's a plugin too). **4. Trial and error, with an experimenter's mindset** |
| 8:40 | slide | **Git is your friend** — it is what makes throwing it away cheap. **Be disciplined about markdown docs** as your tracking layer, as in weeks 2 and 3 |
| 9:00 | slide | **Manage context proactively.** *"Do `/context` all the time. Don't wait for your context to compact. Manage it yourself. Write stuff out to markdown, write to `claude.md` or `agents.md`, then `/clear` and start fresh"* |
| 9:15 | slide | **And the last word: you own the quality of the code you push.** The **Jellyfin AI PR policy** callback. **Don't let it build slop** — no piles of extra test files, no extra readmes, **"and please, no emojis"**. *"Be ruthlessly on top of it. Own the code"* |
| 9:28 | end | |

## For our version

- **The deployment is real and it must be real for us too** — a live URL in the
  address bar is the proof the whole capstone needs. **Fly.io free tier is enough.**
  **NEEDED FROM YOU: nothing, unless you want the app to stay up** — I'll deploy it
  under a neutral name and tear it down after the shoot unless you say keep it.
- **The URL and app name must carry no identity.** His is `finally-ed.fly.dev`.
  Ours gets a project name, not a person's.
- **The Karpathy re-read is the emotional close and it is the best structural idea
  in the course** — the same text, twice, three weeks apart. **Keep it exactly.**
  Our version should reuse the *same slide* from Week 1 Day 1, unchanged, so the
  recall is literal.
- **Keep the concession that it's still bewildering.** After 95 lectures, an
  instructor admitting the ground is still moving is what makes the rest credible.
- **The best-practices run (6:50–9:28) is the most re-watched two minutes in a
  course like this.** Build it as **one clean checklist frame that assembles**,
  not five separate slides — and make it screenshot-friendly on purpose.
- **"No emojis" gets a hold.** It is the single most quoted line from the course
  and it costs one frame.
- **Substitution note:** his scoreboard line names Codex. Ours names **whatever our
  L92 parade actually decided**, and the two sentences must agree. **Write L93's
  narration after L92 is shot**, not before.
- **Footage: ~35%** — the deploy and the live URL are footage; the rest is the
  summary, and here the slides are the point.
