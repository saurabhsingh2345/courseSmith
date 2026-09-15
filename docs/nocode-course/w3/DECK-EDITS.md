# w3-01's deck — the edits the top-up requires

The opener promises the week. It was narrowed once already, when six lectures
could not be shot, and the top-up widens it again. **Three slides change, and
only those three get re-voiced** — `generate_vo.py` is keyed on the text, so an
unchanged slide costs nothing.

Nothing here may be recorded until the captures are cut and the lecture list in
`ORDER.md` is final. A deck that promises a lecture that does not exist is the
exact defect this file exists to prevent.

## Slide 03 — "on one real product"

Now two.

> So take stock of where you are. In week one you asked for things, and watched
> them appear. In week two you put a process around it. And this week, you run
> several agents at once, safely... and build two real products with them.

## Slide 05 — the five days

The current text stops at "we ship it" and never mentions the second product,
the headless runs or the pull-request loop.

> Five days to get there. Today, one agent, properly. Tomorrow, a box it cannot
> get out of... and the same agent with no screen at all, driven from a script.
> Day three, a team of six builds the product while you watch. Day four, you
> write your own orchestrator... and point it at a second product it has never
> seen. And day five, the loop that runs on somebody else's computer — an
> issue, a pull request, a green build — and an honest verdict on both ways of
> working.

## Slide 11 — "three things"

Add the fourth, because three of the new lectures are about it.

> By Friday you will be able to do four things. Direct a team of agents on one
> codebase without losing track of them. Contain them, so freedom is safe. Run
> them with nobody in the chair at all. And judge the result... telling a good
> one from a merely plausible one, and rejecting the difference.

## After re-recording

`record_deck.py w3-01` writes `decks/w3-01.mp4`; the lecture then goes through
the ordinary pipeline. The narration in `narration/w3-01.json` is timed against
the old clip lengths, so **re-run `fit.py w3-01` and `balance.py w3-01 --write
--trim` before assembling** — three longer slides move every anchor after them.
