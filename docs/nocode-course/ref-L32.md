# L32 — "Day 5: Building a Kanban App with GitHub Copilot, Docker & FastAPI" (11:34)

**100% screen recording.** Parts 2 to 4 of the plan, and the lecture with the
sharpest LLM-behaviour lesson in the whole section.

## Part 2 closed out (0:00-3:00)

- **He makes it explain itself** about the `requirements.txt` from L31: *"can you
  explain why you needed a requirements.txt? I thought we were using uv."* Answer:
  uv still needs an input, so it used requirements. His verdict — **it didn't set up
  a uv project the way he would have, "that's not the best way of doing it for
  people that know uv, but it doesn't matter. It works for this purpose"**
- Opens the **Dockerfile** and reads it against the L29 explanation — **Python 3.12**,
  bringing in `requirements.txt`
- **"Please tell me how I test this part myself"** — deliberately making the agent
  hand him the verification steps, **"because we want to trust but verify at each step"**
- Runs **`scripts/start-mac.sh`** → **permission denied**. His note: *"if you have a
  problem like that, you just tell it the problem and it should fix it"*
- **localhost:8000** → `/health` returns **status ok**, `/api/hello` returns **hello
  from FastAPI**. "Part two is working. Thank you to our GitHub Copilot"
- Then the check most people skip: **"are all the success criteria for part two
  achieved?"** → **"Not fully."** Coverage wasn't measured; scripts weren't validated
  on PC or Linux. He waives the PC/Linux item and marks it complete

## Part 3 and **the 80% coverage trap** (3:00-7:30)

The best teaching moment in Day 5, and it is his own mistake:

> He set an **80% unit-test coverage target** in L31. Part 3 wrote the code quickly
> and then **bogged down for most of ten minutes building long, complicated tests**
> — *"not doing it in an intelligent way, not testing things that really matter,
> adding tests just for the sake of it, in order to achieve the objective."*
>
> **"This is a great example of a trap you can fall into with LLMs: they're so eager
> to follow the letter of the rules that they don't push back and say, you know
> what, I know you wanted 80% coverage, but we're not making productive use of time."**

He fixes it by **amending the plan mid-flight**: *going forwards, please update the
plan to only achieve 80% coverage if it's sensible to do so. Avoid adding
unnecessary tests just to hit 80%. Focus on valuable tests. **Not hitting 80% is
okay.***

Part 3 result: `start-mac` → localhost:8000 → **the Kanban front end is being
served** by the back end. "The good old Codex one that we like so much, Kanban
Studio." Approved.

## Reading diffs (7:30-8:30)

With **16 files changed** on screen he explains the review model: **green is added,
red is removed**; you can **keep or undo each diff individually**, or press **Keep**
once to accept everything. **"There's a school of thought that says you should look
through every single diff individually. We're not going to do that."** His actual
position: accept in bulk, **but stop for the sensitive ones** — "maybe when we first
use the AI aspect." And, quoting Karpathy again, **"watch it like a hawk"**.

## Part 4 — sign-in, and persistence proven (8:30-11:34)

- **Sign in with demo credentials** (user / password) → Kanban Studio, with a
  **logout button**
- Chrome offers to save the password — **"Google, of course, does not like that.
  I'm not going to add that to LastPass"**
- **Logs out and back in — the board state is maintained.** "Did you see that? That
  is good to see"

## The Git checkpoint he admits he skipped (10:30-11:34)

**"Something I said I was going to do every step and I didn't. Did you notice? Did
you stop me?"** — the Git snapshots. His justification: the first four parts were
low-risk and repeatable, **but databases are next, so it is a good point to start.**

On camera, with the commands named for beginners: **`git status`** to see what
changed → **`git add .`** to stage → **`git commit -m "part 4 complete"`**. Notes it
is **committed locally, not pushed**, and that this is the point he can return to.
And the fallback for anyone stuck: **ask the agent or ChatGPT for the basics.**

## For our version

- **The 80% coverage trap is the single most valuable five minutes in Day 5** —
  a specification that was followed perfectly and produced waste. It generalises
  far beyond testing, and it is the strongest argument in the section for staying
  in the loop. Give it a dedicated graphic: instruction in, letter-of-the-law out.
- **He asks the agent how to test its own work.** That is a reusable prompt for a
  no-code audience who cannot write tests. Card it.
- **The "did you notice I skipped it?" moment is genuinely charming and honest** —
  and it teaches checkpointing better than a rule would. Keep it.
- Git basics on camera (`status` / `add .` / `commit -m`) is our first real Git
  teaching in the course. Animate the three-stage model alongside the terminal.
