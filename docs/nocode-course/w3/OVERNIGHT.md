# WEEK 3 — overnight session, 2026-08-29

## DELIVERED

**Day 1 is complete: 5 lectures, in `videos/nocode/vids/`.**

| file | covers |
|---|---|
| `w3-02_adam.mp4` | the brief, the first session, the memory file |
| `w3-03_adam.mp4` | slash commands |
| `w3-04_adam.mp4` | sub-agents |
| `w3-05_adam.mp4` | hooks |
| `w3-06_adam.mp4` | plugins, and packaging our own |

All five are **100% real footage, zero animation**, cut from one continuous
26-minute take of one repo, so they run continuously into each other. Narration
was written after sampling frames from each cut, and `assemble.py` reports any
segment where the words outrun the picture — there are none left. Audio is
−16.4 LUFS, peak −1.3 dBFS.

**w3-01, the week opener, is written and voiced** (11 slides in the sticker kit)
but cannot be finished yet: its cold open is the finished Control Tower, which
does not exist until Day 5.

**The app is real.** The sub-agent built a fleet simulator with **85 passing
tests** — non-overlapping regions, vehicles on real lanes, deterministic replay
from a seed, a map that draws offline — and packaged the day's commands, agent
and hook into `tools/control-tower-kit/` with a working plugin manifest.

### Runtime — decided

**He chose the dense honest version**: a Week 3 of roughly 2.5-3 hours where
every minute is real, rather than a padded 5h12m matching the reference.
**Never pad to hit Ed's runtime.** A lecture that will not honestly fill its slot
gets more teaching or a merge.

### Runtime, honestly

Day 1 first rendered at **13 minutes** against the reference's 72, and is being
re-rendered to about **22** by writing more teaching over the same picture.

Two calibrations that matter:
- **Adam measures 198 words per minute**, pauses included — measured off a
  finished render. A 150 wpm guess under-filled every lecture by a third.
- A lecture's ceiling is **its footage ÷ 0.92**. `fit.py` reports the room per
  segment before rendering; `assemble.py` prints `HELD ON LAST FRAME` if words
  outrun picture, so padding cannot ship quietly.

## What exists now

**The whole week is scripted and the pipeline is built and tested end to end.**

| piece | what it does |
|---|---|
| `preflight.py` | refuses to roll on light mode, no panel, another capture running, the advert unstubbed, low disk |
| `preroll.py` | closes stale editors, leaves one clean terminal, re-seats the window, asserts the repo's start state |
| `rig.py` | one long capture + a marker log; segment starts derived backwards from the kill time |
| `w3drive.py` | types only when the shoot window is really in front; answers permission prompts; refuses to type identity-leaking commands |
| `cut.py` | markers → one clip per lecture, stitched across pauses and across captures |
| `frames.py` | samples the cut so narration is written to what is actually there |
| `assemble.py` | ramps footage to fit the narration, two-pass loudnorm, delivers |
| `record_deck.py` | records a slide deck in the sticker kit |
| `shoot_a1/a1b/a2/b/c.py` | all five captures, scripted |
| `teardown.sh` | restores every setting the shoot changed |

## Five defects found by building it — each one would have shipped

1. **`/status` prints his email, organisation and username.** Caught on camera;
   that take was deleted. It also **shadows a custom command of the same name**
   and leaves a **modal that silently swallows every later prompt** — which is
   why one take logged five happy lectures against an empty repo. Custom commands
   are now namespaced `/ct-*` and a guard refuses to type `/status` at all.
2. **Idle detection was blind to the spinner**, so the driver moved on after ~9
   seconds while the agent was still working, and the next prompt was queued
   rather than sent. It now watches the working-status strip, whose elapsed-second
   counter ticks every second. A turn that used to "finish" in 9s correctly takes 39s.
3. **Fullscreen put VS Code on its own Space**, so keystrokes went to whatever was
   frontmost on the panel — they arrived in the agent session as chat messages.
   Now a plain 1920x1080 window with the menu bar auto-hidden.
4. **`alimiter` clips by default.** `limit=0.82` measured **0.0 dBFS** because its
   `level` option auto-levels back to full scale. Fixed to a real two-pass
   loudnorm plus `level=disabled`: **−16.3 LUFS, peak −1.5 dBFS**.
5. **VS Code's Save As appends the detected extension**, producing `plan.md.md`
   and sending the agent off fixing a filename instead of reading the brief.

6. **The idle strip has to reach the bottom edge.** A *background* agent reports
   on its own line UNDER the input box, so a strip that stopped short declared a
   40-minute build finished after 97 seconds, with nothing written. (The build
   did finish — quietly, during the next lecture.)
7. **`/exit` then `claude` is not a restart.** A newly created sub-agent stayed
   "not found" afterwards, and Claude Code itself concluded its own restart
   advice was wrong. A genuinely fresh process in the same folder finds it
   immediately, so the fix is to kill the terminal and open a new one.

## Two facts worth putting on camera

- **`teammateMode` now accepts `iterm2`** as well as `in-process` and `tmux`, and
  **teammates inherit the leader's model** — so the reference's "level the playing
  field" step is obsolete. Checked against the shipped binary.
- **The Agent SDK works on the Max subscription with no API key.** Its stream also
  carries a `RateLimitEvent`, which ties straight into the cost thread.

## The honest finding about runtime

Our lectures come out **much shorter than the reference**, because he types for
ten minutes what we ask for in one sentence. The answer is to teach more per
lecture — read what the agent wrote, make it justify its choices — or to merge
lectures. **Not to pad.** `assemble.py` prints `HELD ON LAST FRAME` if narration
ever outruns footage, so padding cannot ship unnoticed.

## Day 2 as shot (2026-08-29)

| lecture | footage | verdict |
|---|---|---|
| w3-07 sandboxing | **7.4 min** | good — the native sandbox turn alone ran 5.4 min of real work |
| w3-08 the container sandbox | **9.8 min** | **the strongest Day 2 lecture.** The agent wrote a Dockerfile and compose file with reasoning in the comments — pinned to the Node major in `package.json`, slim base, only git and certs, *"no compilers, no ssh client, no curl"*, the CLI baked into the image so the container is disposable, and one bind mount described as *"the one and only host path exposed. Not $HOME, not ~/.ssh, not ~/.aws"*. Three images built and the isolation proof ran |
| w3-09 Claude on the web | 1.7 min | **unusable** — sign-in wall, see below |
| w3-11 GitHub issue to PR | 1.4 min | **too thin.** The `gh` commands return instantly; a real lecture needs the Action actually tagged and a PR coming back, which is a bigger setup |

**Two production bridges**, neither changing teaching content: the agent named its
compose file `docker-compose.dev.yml` and its service `dev`, while the shoot
script calls plain `docker compose` and a service `agent`. Linked the filename,
renamed the service, verified `docker compose config` resolves.

**A private GitHub repo was created** at `saurabhsingh2345/control-tower` so
w3-11 could run at all — private, empty, no push, deletable.

## Still needing him

**`cursor-agent` is logged out** (checked 2026-08-31): *"Your stored
authentication is invalid. Please run 'agent login'."* That blocks **w3-21**, the
Cursor contender on Day 5. One command from him fixes it:

```
cursor-agent login
```

Without it, Day 5 still works — our own Agent-SDK orchestrator, the verdict and
shipping all run — but the comparison has two entries instead of three.



**w3-09 (Claude Code on the web) and w3-10 (on a phone) both need you.**
The shoot uses a throwaway Chrome profile on purpose — his own window would put
tabs, bookmarks and a profile picture on camera — and that profile is not signed
in. I will not type credentials, so w3-09 filmed a sign-in wall. Options, his
call:
- he signs the shoot profile in himself, once
  (`/Users/Shared/projects/.shoot-chrome`), and I refilm w3-09; or
- we drop both and give Day 2 the extra time to the container sandbox and the
  GitHub automation, which both work unattended today.

w3-11 (GitHub issue to PR) is unaffected — it runs through the `gh` CLI.



- **w3-10 (Claude Code on a phone)** — a phone cannot be driven from here. Every
  other lecture in the week is unblocked.
- **43% of the weekly Claude limit was already used** before Day 4's four builds.
  Worth watching even on Max.
