# L30 — "Day 5: Setting Up a Full-Stack Project with GitHub Copilot & FastAPI" (11:45)

**100% screen recording.** Sets up the commercial MVP — and this time by the rules.
The tool is **GitHub Copilot** ("or at least I am — you can use whichever one you
want").

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | github.com > avatar > **Copilot Settings** | He is on **Copilot Pro**; you may be on free with its request quota. **1.4% of premium requests used**, resets monthly. Shows it's enabled for github.com, CLI and IDE, and where to toggle models. **"Most importantly, keep an eye on usage"** |
| 1:20 | the repo | The starter repo is **`pm`** — private at recording time, public by release, linked in resources. Green **Code** button > **HTTPS** > copy |
| 2:00 | VS Code | **Ctrl backtick** for a terminal. `cd projects`, `pwd`, then **`git clone …/pm.git`** (he's already cloned it). File > New Window > Open > `pm` > Open. **"Do you trust the author? Yes, you trust me, right?"** |
| 3:20 | **why an existing repo** | Pre-empts the objection that he didn't start from scratch: **agents are great at building from absolute nothing** — scaffolding, files, step by step. **It is harder when you start with something**, and hardest with a large legacy project (covered later). So this project deliberately mixes **inherited code plus new work** |
| 4:20 | the inheritance | **The inherited code is the Codex Kanban front end from Day 3** — "the really cool one". He asks you to **pretend it wasn't vibe coded**, that another team built that MVP front end |
| 4:50 | the mission | Turn it into a real application: **proper front end, back end, database, API** — persistent project-management software. **And it will itself have an AI feature** — "doesn't everything have an AI feature these days?" — a **chat-with-my-project** sidebar that can answer questions **and make changes to the board** |
| 6:00 | Copilot panel recap | The **agent mode dropdown**, the **model selector** and **Manage Models** — including the note that this is where you'd wire up **Ollama running a model locally, for free**, if your machine is strong enough. Plus the **context indicator** and **premium request usage**, with **Manage paid premium requests** opening GitHub to set budget limits |
| 7:30 | **the file tree** | `frontend/` — the existing Kanban MVP. `backend/` — **empty except a placeholder**. `scripts/` — **empty placeholder**. `.env` — copied across from the last project (**"be careful with that .env file"**). `.gitignore` — the defaults, including `.env`, "which you never want to check into Git" |

## His `agents.md` for this project (8:00-11:45)

Right-click > **Open Preview**. He is upfront that it is **scrappy and written just
now** — "rather than you sitting there while I type it."

**Requirements** — a project-management MVP web app: user signs in; sees a **Kanban
board** for their project; **fixed columns that can be renamed**; cards can be
moved; **an AI chat feature in the sidebar**; the AI can **create, edit or move one
or more cards**.

**MVP limitations** — a **single hard-coded user and password**, but **the database
should support many users for the future**; **one board per user**; **runs locally
in a Docker container**, not deployed.

**Technical decisions** — and the point he makes about them matters: *"you don't
need to know what to put here, because you could just ask the AI."* He has opinions,
so he states them:
- **Next.js** front end (already have it)
- **Python back end with FastAPI** — "Python is what I do"
- **Everything packaged into a Docker container**
- **`uv` as the package manager** — "these agents don't like to do that, they like
  to use pip. I like uv because I love uv"
- **OpenRouter** for AI calls, key in `.env` at the project root, model named
  explicitly (**append `:free` for the free variant**)
- **SQLite** local database — noting you could instead put **Supabase** keys in
  `.env` and it would work fine

**Starting point** — a live edit worth keeping: he wrote "current state", then
**renames it to "starting point" on camera** because *"it's only going to get
confusing when it sees that's not the current state later on."*

**Colour scheme** and **coding standards** — much as before, plus a fourth he
highlights: **"when hitting issues, always identify root cause. Do not guess. Prove
with evidence, then fix the root cause."**

**Working documentation** — all planning and execution docs live in `docs/`, and
**"please review `docs/plan.md` before proceeding."**

## For our version

- **"Inherit something" is a real teaching decision and he justifies it well.**
  Ours needs the same setup: a starter repo containing our Day 3 Kanban front end,
  an empty back end, a `.env`, a `.gitignore` and an `agents.md`.
- **The renaming of "current state" to "starting point" is a tiny, perfect moment**
  — an expert catching an ambiguity a model would trip over. Keep it.
- The fourth coding standard is **the L18 debugging rule promoted into the file
  itself.** Our cut should make that connection explicit — the thing you had to
  type by hand on Day 3 is now written down once.
- Same PII rule: neutral GitHub account, no real usage dashboards showing a
  personal plan.
