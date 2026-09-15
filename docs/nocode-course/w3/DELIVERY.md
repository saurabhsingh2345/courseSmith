# Week 3 — what to upload

Files are in `videos/nocode/vids/`. **The `w3-NN` numbers are production numbers
with gaps in them** — six lectures were planned, blocked, and replaced by
material that was actually shot — so upload in the order below and use the
position column, not the filename. `ORDER.md` is the same list with the
reasoning attached.

**Thirty-five lectures. All shot, all cut, all written, all voiced, all
rendered.** Nothing in the week is waiting on anything.

| # | file | title | day |
|---|---|---|---|
| 1 | `w3-01_adam.mp4` | Welcome to expert: what changes when you stop typing | 1 |
| 2 | `w3-02_adam.mp4` | The brief: setting up Control Tower | 1 |
| 3 | `w3-03_adam.mp4` | Slash commands: a routine in one word | 1 |
| 4 | `w3-04_adam.mp4` | Sub-agents: a team of specialists inside one agent | 1 |
| 5 | `w3-05_adam.mp4` | Hooks: making your rules fire by themselves | 1 |
| 6 | `w3-06_adam.mp4` | Plugins and marketplaces: packaging your setup | 1 |
| 7 | `w3-07_adam.mp4` | Why sandboxing is the unlock | 2 |
| 8 | `w3-08_adam.mp4` | YOLO without the fear: the container sandbox | 2 |
| 9 | `w3-12_adam.mp4` | Working in a big codebase: the seven rules | 2 |
| 10 | `w3-13_adam.mp4` | Driving Claude from code: the Agent SDK | 2 |
| 11 | `w3-36_adam.mp4` | No screen at all: the print flag | 2 |
| 12 | `w3-37_adam.mp4` | Agents in a pipeline | 2 |
| 13 | `w3-38_adam.mp4` | A stack of documents, not a codebase | 2 |
| 14 | `w3-15_adam.mp4` | Sub-agents vs agent teams: what actually changes | 3 |
| 15 | `w3-16_adam.mp4` | Setting up the team: clean house, pick the roster | 3 |
| 16 | `w3-17_adam.mp4` | Launch: seven agents, one shared task list | 3 |
| 17 | `w3-18_adam.mp4` | What the swarm actually did: the evidence in the repo | 3 |
| 18 | `w3-19_adam.mp4` | The integration pass: are these tests worth anything? | 3 |
| 19 | `w3-20_adam.mp4` | First run: Control Tower is alive | 3 |
| 20 | `w3-23_adam.mp4` | Roll your own orchestrator with the Agent SDK | 4 |
| 21 | `w3-24_adam.mp4` | The verdict on the two ways of running a team | 4 |
| 22 | `w3-27_adam.mp4` | A second brief, and nobody in the chair | 4 |
| 23 | `w3-28_adam.mp4` | Turning the orchestrator into a real pipeline | 4 |
| 24 | `w3-29_adam.mp4` | The run: agents building with no terminal to watch | 4 |
| 25 | `w3-30_adam.mp4` | What a program built | 4 |
| 26 | `w3-31_adam.mp4` | First run, and what is wrong with it | 4 |
| 27 | `w3-32_adam.mp4` | The fix pass | 4 |
| 28 | `w3-33_adam.mp4` | Somewhere other than your laptop: the repo and its CI | 5 |
| 29 | `w3-34_adam.mp4` | An issue is a brief with an address | 5 |
| 30 | `w3-35_adam.mp4` | The pull request | 5 |
| 31 | `w3-39_adam.mp4` | When it is wrong: diagnosing from evidence | 5 |
| 32 | `w3-40_adam.mp4` | The history the pipeline did not keep | 5 |
| 33 | `w3-41_adam.mp4` | Two whole products, honestly compared | 5 |
| 34 | `w3-25_adam.mp4` | Ship it: into a container, with none of your machine in it | 5 |
| 35 | `w3-26_adam.mp4` | Wrap: from no-code to agentic director | 5 |

Confirm runtimes against the rendered files — `python3 week.py --live` prints
the delivered length of every one of them next to what it was cut from.

## Two orderings that are not negotiable

**w3-24 sits before w3-27, not after w3-32.** Its verdict is about the
eighty-two-line orchestrator, and it closes by saying that script cannot edit an
existing file or run the tests, so use the team instead. w3-28 is the lecture
that fixes exactly those two things. Played the other way round the verdict
reads as a stale claim rather than as the setup for the next lecture.

**w3-25 and w3-26 close the week**, after the Day 5 material, because w3-26 is a
montage built out of the week's own cuts and it ends on the container.

## What is in these files that was not in the last cut of them

1. **Deeper narration on the same footage.** The twenty original lectures were
   cut from 129 minutes and ran 83, because thin writing made `assemble.py` ramp
   the picture past real work at about 1.55x. Same footage, same cuts, more said
   over them, nothing padded.
2. **Fifteen new lectures** from five new captures — a second product built by
   an orchestrator the student writes, the issue-to-merged-pull-request loop on
   a real repository with a real Actions run, three headless lectures, and a
   debugging day.
3. **Every frame swept for PII.** See below.

## PII: swept, not spot-checked

`autoredact.py` OCRs every frame of every cut at four-second intervals, looks
for the username, the machine name and the company, re-samples every second
around anything it finds, and blurs the line it sits on for exactly the seconds
it is legible. Fifteen regions across nine lectures.

The ones that mattered were not the ones anybody had looked for. `gh` printing
an account name in an issue header was known and hand-redacted. What the sweep
found was **the shell echoing absolute paths** — `/Users/<him>/.local/share/...`
in w3-07, w3-15 and w3-23, an `ls -la` whose owner column is his username nine
times over in w3-13, and the prompt reverting to `<user>@<machine>` after Claude
Code exits in w3-08, w3-16, w3-27, w3-28, w3-32, w3-33, w3-39 and w3-40.

**w3-07, w3-13, w3-15 and w3-16 had already shipped with it in frame.** Re-render
those four whatever else changes.

Re-run it after any new capture:

```
python3 autoredact.py --plan        # every cut, report boxes, change nothing
python3 autoredact.py               # blur everything that needs it
```

It blurs onto the cut as it stands rather than rebuilding from `.orig.mp4`, so
it can be run twice and stacks with the hand-tuned boxes in `redact-w3.sh`.

## Not shot, and why

**Claude Code on a phone.** A phone cannot be driven from this rig. The point —
that the work continues when you are not at the keyboard — is carried by w3-36
to w3-38, where it runs with no screen attached at all.

**Copilot's coding agent.** No extension and no subscription on this machine.
w3-41 compares the two things we actually built instead of three things one of
which we did not.

**Cursor.** Signed out, CLI and IDE both; the IDE returns *Authentication error*
on an agent turn. w3-27 to w3-32 replace it with a second product built by an
orchestrator the student writes, which pays off w3-13 and w3-23 rather than
advertising a competitor.
