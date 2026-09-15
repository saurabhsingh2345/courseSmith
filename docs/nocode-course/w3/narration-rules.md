# WEEK 3 — HOW THE NARRATION IS WRITTEN

Written 2026-08-29, from his three standing notes: *say what is on screen*,
*be fluid and engaging*, *more footage, less animation*.

## The order of work is not negotiable

1. Shoot.
2. Cut.
3. **Watch the cut.** `python3 tools/nocode/w3/frames.py w3-NN` samples a frame
   at every beat and every 20 seconds.
4. **Then** write the narration, against those frames.

Written the other way round, narration describes what we *expected*. That is how
a lecture once described a failure that never happened on our screen. The
transcript of the reference course tells us **what to teach**. It never supplies
a sentence.

## What a line of narration has to do

Every sentence sits over a specific picture. It must be true of that picture.

- **Name what just appeared.** "The file list on the left just grew by one."
- **Say what it means before saying what it is.** "That number is the whole
  reason we did this" beats "this is the context percentage."
- **Read a number out loud when it is on screen.** If `/usage` shows 8%, say
  eight percent. The frame is the evidence; use it.
- **Never describe something that is not in frame.** If it needs a diagram, that
  is a slide, and it goes before the cut to footage — never after.

## Fluid, not flat

- **Vary the sentence length.** A long explaining sentence, then a short one that
  lands it. Monotone pacing is what makes a screencast feel like homework.
- **Ask the question the student is already asking**, then answer it. "So why not
  just open six terminals? Because they cannot talk to each other."
- **Say the honest thing about what is on screen.** If the agent picks a clumsy
  approach, say so. Watching an instructor let a suboptimal choice play out is
  worth more than a clean demo.
- **No throat-clearing.** Not "now what we are going to do is". Just do it.
- **No emojis, no exclamation marks, no "amazing".** The screen carries the
  excitement; the voice stays level.

## Waiting

An agent thinking for four minutes is ramped 12–14x, never cut to black. The
narration over a ramp does not narrate the wait — it uses the time to explain
what the agent is doing and why, so the ramp reads as a summary rather than a
gap.

## Length follows the footage, not a target

We ask an agent to do in one sentence what the reference instructor types by
hand, so **our lectures are naturally shorter than his.** The response to that is
to **teach more per lecture, never to pad**. A dense five minutes beats a padded
twelve. Where a lecture will not honestly fill its slot, the fix is a richer
shoot — more beats, harder questions — or merging it with its neighbour.

Freeze-framing to fill time is a failure, and `assemble.py` prints
**HELD ON LAST FRAME** whenever it happens so it cannot pass unnoticed.
