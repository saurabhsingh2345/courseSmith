# Week 3 — the runtime push

**Goal (his, 2026-09-01):** Week 3 at 5–6 hours, matching the reference week's
5h11m, **screen recording only**, no padding. Today it is 83 min over 20
lectures.

## What is actually authenticated on this machine

Probed 2026-09-01 before planning anything.

| tool | state | evidence |
|---|---|---|
| Claude Code (Max) | **works** | `claude 2.1.252`, shoot repo pinned to sonnet |
| Claude Agent SDK | **works** | `tools/orchestrator.mjs` in control-tower |
| `gh` CLI | **works**, scopes `repo` + `workflow` | `gh auth status` |
| Docker | **works** | server 29.1.3 |
| Cursor CLI | **logged out** | `Authentication required` |
| Cursor IDE | **logged out** | agent turn returns *Authentication error* on camera |
| Antigravity | **logged out** | onboarding screen, *Awaiting Authentication* |
| Copilot | **absent** | no extension, no subscription |

So the "another Cursor project" cannot be Cursor. It becomes a second real
product built by a **different method** — our own Agent SDK orchestrator,
headless — which is a better lesson than a second vendor anyway: by the end the
student is not choosing between other people's orchestrators, they wrote one.

## The two levers

**A — reclaim what is already shot. +57 min, no camera.**
The 20 delivered lectures come from **128.8 min of cut footage** but deliver
**83.2 min**, because thin narration forces `assemble.py` to ramp the picture
1.5x past real work. Deepening the narration to fill the cut takes the week to
**~140 min** and is strictly better video: nothing races any more.

**B — new captures. +~175 min.**

| capture | lectures | what is on screen |
|---|---|---|
| **E** Signal Desk, part 1 | 3 | the brief, the orchestrator built for real, the run |
| **F** Signal Desk, part 2 | 3 | the evidence, first run, the honest defects |
| **G** GitHub | 3 | issue → agent → branch → PR → Actions green → merge |
| **H** headless | 3 | `claude -p`, JSON out, agents piped together, a batch job |
| **I** the box | 2 | an agent inside Docker it cannot leave, then ship |
| **J** re-cuts | 2 | the three-way verdict, the wrap montage |

## The honest answer on five to six hours

The reference week is 5h11m. This one will not reach it from the footage that
exists, and the reason is arithmetic rather than effort.

`week.py` prints the three numbers that settle it. **Footage** is what was shot.
**Ceiling** is what that footage can carry with somebody talking over every
second of it at the house rate and nothing ramped. **Written** is how much
narration exists.

After the top-up the cuts come to roughly **4.5 hours of ceiling** — 129 minutes
of the original week plus about 145 of new capture. Narrated to the last second
of it, Week 3 is a four-and-a-half hour week. Getting past that is not a writing
problem: it needs more footage, which means more capture days, because a lecture
cannot be longer than the thing it is describing.

The gap between what he types in ten minutes and what we ask for in one sentence
is real and it does not close. What closes it is more *distinct work on camera* -
more products, more failures, more tools - which is exactly what the top-up did
(a second product, a pull-request loop, three headless lectures, a debugging
day) and exactly what another day of it would do again.

## Where the five hours actually is

The footage supports the target; the voice and the writing are what is short.

| | minutes |
|---|---|
| cuts that exist after the top-up | ~190 new + 129 existing |
| what that footage can carry, narrated at natural pace | **~345 min, 5.75 h** |
| written so far | 104 (upgrades) + ~41 (second product) |
| **voiceable this month** | ~116 |

So the order of work is: **write everything to length, prefetch what the
allowance covers, render what is prefetched.** Writing costs nothing but time
and is the thing that turns shot footage into a lecture; the voice can be bought
in October against writing that is already finished. Doing it the other way
round - rendering short lectures now because the allowance is small - would
throw away footage that cannot be re-shot.

## What replaces each blocked lecture

| was blocked | replaced by | why it is the same lesson |
|---|---|---|
| w3-09 Claude Code on the web | **H1 headless `-p`** | the point was "it runs where you are not sitting" |
| w3-10 on a phone | **I1 an agent in a container you leave running** | same point, no hardware |
| w3-11 GitHub Action | **G1–G3 issue → PR → CI, driven by `gh`** | the hosted Action is a webhook; the loop is the lesson |
| w3-14 Cowork | **H3 a batch job over the receipts folder** | agents on documents, not code |
| w3-21 Cursor | **E1–E3 a second product, built by our orchestrator** | a second way to run the same kind of work |
| w3-22 Copilot | **J1 the verdict, three honest ways** | comparison without a subscription |

## What actually limits the runtime, and it is not the footage

Measured 2026-09-01, after the narration pass:

| | |
|---|---|
| voice allowance left | **48,143 characters** |
| resets | **2026-09-11** |
| re-render the 20 upgraded lectures | **27,172** |
| left after that | **~20,971** — about 19 minutes of new narration |

So the ceiling on what can be *delivered* today is roughly **104 minutes of
upgraded lectures plus one three-lecture block**, not five hours. The footage is
not the constraint and neither is the writing; a metered voice is.

Three things follow, and all three are done:

1. **`adam_lay.speak()` is content-addressed.** Every sentence is filed under a
   hash of its exact text in `tools/nocode/scripts/_voice/`, so a sentence is
   paid for once ever — wherever it later moves to, and in whichever lecture.
   1,218 existing clips were backfilled. Before this, inserting one sentence
   into a segment renumbered every clip after it and re-bought all of them.
2. **`quota.py --live`** prices a render against the live allowance without
   spending a character.
3. **`render.py`** renders in priority order and stops *before* a lecture it
   cannot finish, because a lecture that runs out of voice halfway is silent
   from that point and nothing tells you but watching it.

**Shoot everything, cut everything, write everything, render what fits.** The
footage is the only part that cannot be regenerated in October.

### The allowance is shared, and it moves while you are not looking

At 10:35 it read 48,143. At 11:30, with nothing local having called the API, it
read 39,819 — **8,324 characters gone in under an hour**. One key, two machines
(see the pack/unpack kit in `tools/nocode/sync`), one monthly limit.

So the buying is now separated from the rendering. **`prefetch.py`** speaks every
sentence of a lecture into the content store and stops there — network and disk
only, no ffmpeg, so it is safe to run while a capture is rolling, which is
exactly when there is time for it. Once a sentence is in the store it is paid
for permanently and `assemble.py` will not buy it again, whatever the allowance
has done in the meantime.

```
python3 prefetch.py --plan --upgrades    # price it, spend nothing
python3 prefetch.py --upgrades           # buy it, render later
```

**Buy before you write more.** The order that survives a shared key is: write
the narration, prefetch it, then render whenever the machine is free.

### Cutting is no longer the bottleneck either

`cut.py` re-encoded every slice with `h264_videotoolbox` at about one times
realtime — an hour of capture cost an hour of the same hardware encoder the next
take needs. The capture is written with `g=15`, a keyframe every half second,
and `PAD_IN` is 0.6s, so a stream copy's seek error is already inside the pad.
`cut.py` now copies by default (`--reencode` restores the old behaviour):
**54 minutes of footage cut in 2 seconds instead of 54 minutes.**

## Non-negotiable, unchanged

Build with sonnet. No name, email, company, `/status`, Cursor Settings,
`docker ps`, `docker images` in frame. Shoot from `/Users/Shared/projects`.
Check the last two seconds of every cut. `shootmode.py on` before,
`teardown.sh` after.
