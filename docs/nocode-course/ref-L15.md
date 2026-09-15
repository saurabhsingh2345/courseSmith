# L15 — "Day 3: Hands-On with Cursor, Copilot, Codex & Agentic Vibe Coding" (8:27)

**Mostly screen recording.** First "yellow day" — a products day. Opens with ~2½ min
of ground rules on slides, then straight into the lab.

## The ground rules (0:00-2:30, slides)

Opens on a confession: "I feel a little bit guilty that yesterday was all me
talking. Today, less talking, more doing."

Four principles for the day:
1. **Everything today is optional.** You don't need to install all four products.
   Watch, pick one, stick with it — Cursor, Copilot, or Claude Code next week
2. **Your results will vary** — different models, cheaper models, free models
3. **Don't get frustrated.** Be patient; give the agent feedback; and the big
   trick: **"simplify, simplify, simplify"** — cut the scope, get something
   working, build from there. If stuck, delete and restart. If still stuck, skip
   that product and move to the next
4. **Have fun with it**

## The lab setup (2:30-8:27, footage)

| approx | on screen | what happens |
|---|---|---|
| 2:30 | Cursor | Reopen Cursor, back into the **`instant`** project from Day 1. File > New Window if the welcome screen doesn't show |
| 3:00 | course resources | Points at the resources doc (linked in the first videos and the welcome email) and the Node.js link in it |
| 3:20 | **nodejs.org**, live | Download page. Mac commands are the default; Windows via **Chocolatey**, or scroll to Windows > x64 for a normal installer |
| 4:20 | Cursor, View > Terminal | Where to actually run those commands. Then the **`+`** button for a new terminal |
| 5:00 | terminal | `node --version` — **"anything after 22 is great"** |
| 5:30 | terminal | A gentle terminal primer for people new to it — "nothing to be afraid of, even if it looks like something hung over from the 80s". **`pwd`** = print working directory, shows `/Users/ed/projects/instant`. Notes PCs use backslashes |
| 6:20 | terminal | **`cd ..`** up to `projects`, `pwd` again to prove it |
| 6:50 | terminal | **`git clone https://github.com/ed-donner/kanban.git`** — the starter repo for the day, also in the resources. Then `cd kanban`, `pwd` |
| 7:40 | Cursor | File > New Window > Open Project > navigate to `projects` > `kanban` > Open. Confirms by **KANBAN in block capitals** top-left |
| 8:27 | end | "It's time for us to start work" |

## For our version

- **This is the shape we want more of**: ~30% slides for framing, ~70% real screen.
- **We need our own starter repo.** His is `github.com/ed-donner/kanban` and it
  contains exactly one file — an `agents.md` (see `ref-L16.md`). Ours should be the
  same shape under our own account, so the clone step is real on camera.
- The terminal primer matters for our audience too — this is a no-code course and
  `pwd` / `cd ..` / `git clone` is the first command line most viewers meet. Keep
  it, and animate the directory tree alongside the terminal so the two views agree.
- His "simplify, simplify, simplify" is the most reusable piece of advice in the
  day. Make it a recurring card we bring back whenever a build goes wrong on camera.
