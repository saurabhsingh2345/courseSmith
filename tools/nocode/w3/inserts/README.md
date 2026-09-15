# Week 3's animation layer

## Why Week 3 had none

Week 1's lectures are Remotion decks. `renderer/src/nocode/l03.tsx` … `l34.tsx`
is one deck per lecture, and the screen recordings are mounted *inside* them as
`shot` slides — `l17.tsx` is 48 slides across 9m41s, so the picture changes
roughly every twelve seconds and animation is the spine of the lecture.

Week 3 was cut the other way round. `../assemble.py` lays narration over one
continuous 4K screen capture and ramps the footage to fit the words; it contains
no reference to a deck, a slide or Remotion, and there is no `w3-*.tsx` in
`renderer/src/nocode/`. So Week 3 has no animation in it at all, and that is why
25 of its 35 lectures sit above 70% of their runtime on an unchanged picture
(`../../qa/holds_w3.log`).

It was not meant to end that way. Every `program/part-w3-*/shot-list.txt` still
specifies `SLIDE ANIMATION` before `SCREEN RECORDING`, and `../record_deck.py`
exists to render those decks. But 25 of the 26 `slides.html` files are two-slide
stubs, only `w3-01` was ever authored (11 slides) and recorded (169s), and no
assembly script ever reads `../decks/`. The deck path was started and abandoned
under the delivery deadline.

## What this does instead

Rather than rebuild 35 lectures, it lays short animated **cutaways** over the
stretches where the delivered picture does not move — the same templates Week 1
uses, out of `renderer/src/nocode/kit.tsx`, so the two weeks match by
construction rather than by taste.

Two rules the whole approach depends on:

1. **The narration is never touched.** An insert is exactly as long as the hole
   it fills, the audio stream is copied rather than re-encoded, and runtime out
   equals runtime in to the frame. No re-voice, no re-shoot, no re-sync.
2. **Only dead frames get covered.** A card either visualises the concept the
   narration is explaining, or re-renders at a readable size something that is
   on the frame but illegible. It never hides footage the narration is asking
   you to read.

## The pipeline

```
python3 source.py w3-NN --remap     # dead spans + sync leads of the current source
python3 brief.py w3-NN              # what is SAID over each dead span. Author from this.
vi plans/w3-NN.json                 # the cutaways
python3 render.py w3-NN --dry       # refuse a bad plan before a 4K render
python3 batch.py w3-NN              # render + lay + before/after, in one call
python3 verify.py w3-NN             # the honest result
python3 lay.py w3-NN --deliver      # moves the master aside into orig/
```

`batch.py w3-NN w3-MM ...` or `--all` does a run of them; a failure stops that
lecture and moves on, so a long run tells you which ones still need work rather
than losing the ones that succeeded.

`lay.py` reads from `orig/` when it exists, so a second pass builds on the
footage as it was shot and never lays a card on top of a card.

## Read the result off the SHARP detector, not holds.py

`holds.py` graded the original problem correctly and its "before" numbers are the
ones the whole week was measured on, so keep using it for that. **It is the wrong
metric for a finished cut**, and wrong in one direction: a few lines of terminal
text barely move the mean of a 64x36 luma, and terminal text is most of what
these lectures contain.

`w3-39` finished reads **25% frozen with a 42-second hold** by `holds.py`. Run
`deadzones.py` over the same file and there is **not one dead span of 12 seconds
or more anywhere in it** - inside that "42-second hold", two commands and their
output print. `w3-20` finished reads 43% with a 64-second hold; that hold is the
live delivery map, sixty vans moving, which the coarse detector cannot see either.

So `verify.py` reports both, and the column that matters is `dead >=12s` on the
finished cut. Zero is the target, and it is what "no frozen frames" actually
means. Quoting the coarse figure as the result understates the work considerably
and invites a second pass over footage that is already fine.

## Two traps that cost real time

**`holds.py`'s detector is too coarse to plan against.** It grades a lecture on a
64x36 luma, which is the right question for "is this picture moving?" and the
wrong one for "may I cover these seconds?". It called w3-36's first 148 seconds
one unbroken hold; three commands and two answers land inside that window,
because a few lines of terminal text in a 4K frame barely move the mean of a
64x36 luma. `deadzones.py` therefore uses a 320x180 grid and **counts the cells
that moved** rather than averaging the frame — a couple of words light up a
handful of cells hard, while encoder noise moves many cells barely, and
averaging cannot tell those apart.

The first cut of `w3-03` was planned against the coarse map and covered
89.0–96.5s, where the next request is being typed on camera. `render.py`'s
`check()` now refuses a plan whose windows leave the measured dead spans, which
is the guard that would have caught it.

**Compare frames, not seconds.** An mp4 reports a duration one or two frames
longer than its frame count, because the container gives the last frame a
duration of its own. Checking seconds rejected a render that was exactly right.

**Never copy a master.** `batch.py` used to copy each one into `orig/` before
working on it; nine lectures in that filled a 460 GB disk and truncated four
copies mid-write, which then produced maps claiming a lecture had ZERO dead
spans (w3-27, w3-34, w3-41). `--deliver` now MOVES the master aside with
`os.replace`, which is free on one filesystem, and `batch.py` deletes the 4K
strips after laying. Cross-check any map against `../../qa/deadzones_w3.log`,
which was scanned straight off `vids/`, before authoring from it.

**One input per insert will get ffmpeg OOM-killed.** Four lectures died with
SIGKILL because the filter graph opened one 4K decoder per cutaway - nineteen at
once for w3-17. `render.py` now renders a lecture's cutaways as ONE strip and
`lay.py` trims them back out by frame offset, so the graph has two inputs
whatever the card count.

**`Term` defaults its prompt to `arena % `**, from the old game project. Every
`term` card must set `"prompt": ""` or it puts the wrong project name in front
of every command. Caught by rendering stills of a finished cut, not by reading
the plan.

## Authoring notes

- Aim for roughly a third of runtime animated. `w3-03` is 34% across four
  cutaways in 2m45 — about one every forty seconds, which reads as a rhythm
  rather than as a slide deck with clips in it.
- 6–8s per card. Under 2s is a flash, and `check()` refuses it. `myth` needs
  four seconds before its second half lands; `term` types at 42 chars/sec.
- Mine the existing catalog before building a template. There are 31 in
  `kit.tsx` and 105 more in `internal/pipeline/snippet_*.go`. Watch the
  rows-overuse trap — `rows` is the easy answer and rarely the right one.
- Write the card from the narration at that timestamp, never from the lecture's
  topic. `../narration/w3-NN.json` is the ground truth; the `anchor` field is
  the second the words land on.

## Two more traps, 2026-09-07

**The masters are variable frame rate.** `assemble.py`'s ramp drops a frame
every half-second or so (w3-13: 16,841 frames over 581 s, 588 gaps), so frame
index N is NOT at N/30 s. The concat version of `lay.py` trimmed the master by
frame index and every card landed early while the video stream came out eight
seconds shorter than the audio - and `format=duration` still reported the
runtime unchanged because it reports the longer stream. `lay.py` now runs
`fps=30` on the master before any trim, and both `lay.py` and `deliver.py` gate
the **video stream's** duration against the master. Verify a lay by comparing
frames against the master BY TIME, never by index.

**Run everything here with `/usr/bin/python3`.** The default `python3` is 3.14
without numpy, and holds/deadzones/align/flash all need it; batch.py, verify.py
and deliver.py call them through `sys.executable`, so the wrong interpreter
fails the remap step before anything renders.

**Strips are chunked** past ~120 s of cutaways (`render.py` `CHUNK_FRAMES`) -
one Remotion composition over ~140 s exited 1 on seven lectures. The manifest
records the chunk of each insert next to its offset.
