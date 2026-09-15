# L66 — "Week 3 Day 1: Setting Up the FiNALLY Project: Our Multi-Agent Trading App" (9:44)

**Almost entirely screen recording.** VS Code + terminal. This lecture sets up the
capstone that the whole of Week 3 builds.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | VS Code | "Bring up your favourite terminal with **Control backtick**" — into home, into `projects` |
| 0:20 | terminal | **`git clone`** his scaffolding repo. The project is **FiNALLY** — final week, and **"FiNance ALLY"** |
| 0:50 | slide/talk | What we're building: a **live trader workstation**, market data streaming, "feel like a high-end application", built by an army of agents |
| 1:20 | VS Code | Opens it as a project. Walks the tree: empty-ish README, **`claude.md`**, empty `backend/` `db/` `frontend/` `tests/`, and **`planning/plan.md`** — the one document that matters. A `.gitignore`, and a **`.env` with his OpenRouter key already in it** |
| 2:30 | `claude.md` | Very short. "All project documentation is in the planning directory… The key document is `plan.md`, **included in full here**" — using **`@planning/plan.md`** so it is forced into context every time |
| 3:10 | `plan.md` in preview | **The business requirements.** Vision: a stunning AI-powered trading workstation, streams live market data, simulated portfolio trading, an **LLM assistant that can analyse portfolios and execute trades**. "Modern Bloomberg terminal with an AI co-pilot" |
| 4:00 | the doc | **He admits how he wrote it**: a paragraph by hand, then iterated with **claude.ai the chatbot**, not Claude Code. Most of the rest is generated from Q&A with him |
| 4:30 | the doc | UX (Docker, browser to localhost, watch prices stream, chat with assistant) · visual design + colour scheme · architecture (SQLite, background task for market data) · **directory structure** · **boundaries between areas** — "agents need to know what they're responsible for and how they interact" |
| 6:00 | the doc | Env vars: OpenRouter key, plus **optional** market data — **"Massive, formerly known as Polygon"**. Optional because the app ships **simulated market data**, so it costs nothing to run |
| 6:50 | the doc | Market data over **SSE**, same streaming shape as an LLM response. **SQLite**, lazily created, persists between sessions |
| 7:30 | the doc | API endpoints · LLM integration via **Cerebras through OpenRouter** ("I love the way it's so fast") · **structured outputs**, as in week two |
| 8:00 | `.claude/` | He has **copied in the Cerebras skill from week two** — the skill just works by being in the folder |
| 8:20 | the doc | Front-end design/screens · **deployment as ONE Docker container**. "LLMs love to try and put things into multiple Docker containers… I came in and said no. **That's the kind of place where human involvement is so important**" |
| 9:10 | the doc | Testing: unit tests and end-to-end tests |
| 9:44 | end | "A juicy, great big project… run using multiple agents in controlled chaos" |

## For our version

- **This is the capstone brief, and he told us to make ours bigger and better.**
  Same shape — one `plan.md` that every agent converges on — but ours should be
  a genuinely more ambitious product. Keep his boundaries discipline: it is the
  thing that makes multi-agent work at all.
- **Our own starter repo, our own account**, exactly as we did for `kanban`.
  The clone is real on camera.
- **We already have the pattern for the brief** from W1's `agents.md`, and it
  worked (L17's plan traced every decision back to it). Scale it up, don't
  restyle it.
- **Simulated market data is the right call and we keep it** — it means a viewer
  pays nothing and the app still looks alive. Real data stays an optional key.
- Keep his **"could this be simpler?"** beat about the single Docker container.
  It is the best human-in-the-loop moment in the lecture.
- **Footage: ~90%.** Terminal clone, the tree, the file in preview, the skill
  folder. Almost nothing here needs animation — at most one diagram of the
  boundaries between backend/frontend/db, because that is what the agents will
  be divided along.
