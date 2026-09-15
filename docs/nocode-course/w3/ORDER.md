# Week 3 — the order after the top-up

Production numbers have gaps in them (six lectures could not be shot) and the
top-up captures continue from `w3-27`. **Upload in the order below, not by
filename.** Positions are renumbered; the production id is what is on disk.

## Day 1 — the workshop

| # | id | title |
|---|---|---|
| 1 | w3-01 | Welcome to expert: what changes when you stop typing |
| 2 | w3-02 | The brief: setting up Control Tower |
| 3 | w3-03 | Slash commands: a routine in one word |
| 4 | w3-04 | Sub-agents: a team of specialists inside one agent |
| 5 | w3-05 | Hooks: making your rules fire by themselves |
| 6 | w3-06 | Plugins and marketplaces: packaging your setup |

## Day 2 — the boundary, and the agent with no screen

| # | id | title |
|---|---|---|
| 7 | w3-07 | Why sandboxing is the unlock |
| 8 | w3-08 | YOLO without the fear: the container sandbox |
| 9 | w3-12 | Working in a big codebase: the seven rules |
| 10 | w3-13 | Driving Claude from code: the Agent SDK |
| 11 | **w3-36** | **No screen at all: `claude -p`** |
| 12 | **w3-37** | **Agents in a pipeline** |
| 13 | **w3-38** | **A stack of documents, not a codebase** |

w3-36 to w3-38 replace *Claude Code on the web* and *Cowork*. Both of those
were making the point that the work does not have to happen in front of you,
and both are wrappers around `-p`, which ships in the CLI and needs no login.

## Day 3 — the team

| # | id | title |
|---|---|---|
| 14 | w3-15 | Sub-agents vs agent teams: what actually changes |
| 15 | w3-16 | Setting up the team: clean house, pick the roster |
| 16 | w3-17 | Launch: seven agents, one shared task list |
| 17 | w3-18 | What the swarm actually did: the evidence in the repo |
| 18 | w3-19 | The integration pass: are these tests worth anything? |
| 19 | w3-20 | First run: Control Tower is alive |

## Day 4 — writing the orchestrator, and using it

| # | id | title |
|---|---|---|
| 20 | w3-23 | Roll your own orchestrator with the Agent SDK |
| 21 | w3-24 | The verdict on the two ways of running a team |
| 22 | **w3-27** | **A second brief, and nobody in the chair** |
| 23 | **w3-28** | **Turning the orchestrator into a real pipeline** |
| 24 | **w3-29** | **The run: agents building with no terminal to watch** |
| 25 | **w3-30** | **What a program built** |
| 26 | **w3-31** | **First run, and what is wrong with it** |
| 27 | **w3-32** | **The fix pass** |

**w3-24 has to sit here, before w3-28, not after it.** Its verdict is about the
eighty-two-line orchestrator, and its closing line is that our script cannot
edit an existing file or run the tests, so use the team instead. w3-28 is the
lecture where we fix exactly those two things. Played the other way round, the
verdict reads as stale rather than as the thing that set up the next lecture.

w3-27 to w3-32 replace the Cursor lecture. Cursor is signed out on this
machine and we never type his credentials, so the second way to run this work
is one the student writes rather than installs — which pays off w3-13 and
w3-23 instead of advertising a competitor.

## Day 5 — the loop, the failure, the verdict, the ship

| # | id | title |
|---|---|---|
| 28 | **w3-33** | **Somewhere other than your laptop: the repo and its CI** |
| 29 | **w3-34** | **An issue is a brief with an address** |
| 30 | **w3-35** | **The pull request** |
| 31 | **w3-39** | **When it is wrong: diagnosing from evidence** |
| 32 | **w3-40** | **The history the pipeline did not keep** |
| 33 | **w3-41** | **Two whole products, honestly compared** |
| 34 | w3-25 | Ship it: into a container, with none of your machine in it |
| 35 | w3-26 | Wrap: from no-code to agentic director |

w3-33 to w3-35 replace the GitHub Action lecture. The Action is a webhook and
an install button; the loop underneath it — issue, branch, test, pull request,
CI, merge — runs from the terminal on a real repository with a real Actions
run, and teaches the part that transfers.

## Still not shot, and why

**Claude Code on a phone.** A phone cannot be driven from this rig at all. The
point it was making - that the work continues when you are not at the keyboard -
is carried by w3-36 to w3-38, where the work runs with no screen attached to it
at all.

**A container lecture beyond w3-08 and w3-25.** An agent running *inside* the
dev container needs to authenticate there, interactively, on first run - and we
do not type credentials. w3-08 proves the isolation and w3-25 ships from the
image; running an unattended agent inside one needs him for sixty seconds.

**Copilot coding agent.** No extension and no subscription on this machine.
w3-41 is the comparison lecture instead, and it compares two things we
actually built rather than three things one of which we did not.

## Continuity that has to move together

Any change to the list above lands in four places, and they were all rewritten
once already when six lectures were cut:

1. **w3-01's deck** — the FIVE DAYS slide lists what the week actually
   contains. It has to be re-voiced and re-recorded whenever this list moves.
2. **Forward references inside narration.** `grep -ri 'tomorrow\|next we\|day
   four\|day five' narration/` after every change.
3. **`montage.py`'s SHOTS list** — w3-26 is built from the week's own cuts and
   must show the week that ships.
4. **This file and `DELIVERY.md`.**

## What goes stale when w3-27 to w3-32 ship

Two lines that are true of the week shipping today stop being true the moment
the second product ships with it. Both are in narration that is already voiced,
so changing them is cheap but not free - check them before that upload, not
after:

1. **w3-24's verdict** and **w3-26's wrap** both say the orchestrator "cannot
   edit a file or run a test". That was true of the eighty-two-line version and
   is the exact defect **w3-28** fixes. Once w3-28 is in the week, both need a
   clause acknowledging it - or w3-24 must sit immediately before w3-27, which
   is what `ORDER.md` already specifies, so that the verdict reads as the setup
   for the fix rather than as a stale claim.
2. **w3-01's deck**, slides 3, 5 and 11 - see `DECK-EDITS.md`. It currently
   promises one product and five days; the finished week has two products and a
   pull-request loop in it.

Neither is a problem for today's upload, because a lecture that under-promises
is fine and one that contradicts the next lecture is not.
