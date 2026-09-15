#!/usr/bin/env python3
"""CAPTURE J - the second week of a codebase.

    python3 shoot_j.py                 # w3-42 .. w3-44
    python3 shoot_j.py --only w3-43 --as J2

Everything in Week 3 so far starts from an empty folder, and that is the one
thing almost nobody actually does. This capture answers the question the week
otherwise leaves open: you have a repository you did not write, it has tests
that pass, and you have to change it without breaking it.

It runs against control-tower as the agent team left it - a real repository with
a real suite - so nothing here is staged. The feature is chosen to be small
enough to land in one sitting and awkward enough to touch three layers: the
simulator that owns the data, the rules that decide what is an exception, and
the console that has to show it.

If the suite is already red when this rolls, w3-42 says so on camera and fixing
it becomes the first job. What it must never do is pretend the repository was
clean when it was not.
"""

from __future__ import annotations

import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import w3drive as W  # noqa: E402
import w3kit as K  # noqa: E402
from w3kit import beat, hold, NO_ASK  # noqa: E402

W.SHOOT_TITLE = "control-tower"
REPO = "/Users/Shared/projects/control-tower"


# ------------------------------------------------------------------ w3-42
def w3_42() -> None:
    """Read it before you touch it."""
    beat("w3-42/cold", "what you actually have, before any opinion about it")
    W.run("git log --oneline | head -15", wait=6)
    hold(12)
    W.run("git status && ls", wait=6)
    hold(10)

    beat("w3-42/suite", "the only honest starting point is a green run")
    W.run("npm test 2>&1 | tail -25", wait=8)
    K.wait_process("node --test", minutes=12, label="baseline suite")
    hold(22)

    beat("w3-42/map", "hand it the repository, not a description of it")
    W.run("claude --model sonnet", wait=8)
    W.trust()
    hold(4)
    W.turn(
        "You have never seen this repository before and you are about to change "
        "it. Do not change anything yet. Read it - the brief, the memory file, "
        "the source, the tests - and write MAP.md at the root containing: what "
        "this product is in two sentences; the four or five parts it is built "
        "from and which directory each one lives in - src holds api, chat, risk, "
        "sim and ui; how data gets from where it "
        "is produced to where it is drawn; where the tests are and what they "
        "actually cover; and the three places a newcomer is most likely to break "
        "something. Cite a real path for every claim. Where you are guessing, "
        "say you are guessing." + NO_ASK,
        timeout=1500, label="map")
    hold(12)
    W.turn("/exit")
    K.read_doc("MAP.md", pages=10, dwell=3.4)


# ------------------------------------------------------------------ w3-43
def w3_43() -> None:
    """A real feature, on code somebody else wrote."""
    beat("w3-43/brief", "small enough for one sitting, wide enough to hurt")
    W.run("claude --model sonnet", wait=8)
    W.trust()
    hold(4)
    W.turn(
        "Add driver shifts to Control Tower. Every vehicle gets a driver with a "
        "shift that has a start and an end, and a vehicle still running after "
        "its shift ends becomes a fifth kind of exception, alongside the four "
        "already in src/sim/exceptions.js. Read those four first, because this "
        "one does not work the way they do: they fire at random against a "
        "weight, and a shift breach is not random at all - it is a fact about "
        "the clock. Decide how that fits the shape that is already there, and "
        "tell me what you decided and why before you write it. Follow CLAUDE.md; "
        "the rules in it are not advice. Add tests for the new behaviour, and do "
        "not change an existing test to make your code pass. Run the whole suite "
        "when you are done and show me the output." + NO_ASK,
        timeout=2400, label="feature")
    hold(15)

    beat("w3-43/diff", "read the change, not the summary of the change")
    W.turn("/exit")
    W.run("git diff --stat", wait=8)
    hold(18)
    W.run("git diff -- src tests | head -120", wait=8)
    hold(30)

    beat("w3-43/alive", "and does the product still run")
    W.run("node src/server.js >/dev/null 2>&1 &", wait=8)
    hold(8)
    K.browser("http://127.0.0.1:8080", wait=18)
    hold(30)
    K.close_browser()


# ------------------------------------------------------------------ w3-44
def w3_44() -> None:
    """The review that catches what a green suite does not."""
    beat("w3-44/green", "green is a claim, not a proof")
    W.run("npm test 2>&1 | tail -20", wait=8)
    K.wait_process("node --test", minutes=12, label="suite after the feature")
    hold(20)

    beat("w3-44/review", "review it as somebody who did not write it")
    W.run("claude --model sonnet", wait=8)
    W.trust()
    hold(4)
    W.turn(
        "Review the driver-shift change as a reviewer who did not write it and "
        "does not trust it. Read the diff against the previous commit. Find the "
        "cases the new tests do not cover - the boundary at the exact shift end, "
        "a shift that crosses midnight, a vehicle already held by one of the "
        "other four exceptions when its shift ends - and say "
        "for each one whether the code is right, wrong, or untested. Do not fix "
        "anything yet. If the change is genuinely sound, say so and give me the "
        "single weakest thing about it anyway." + NO_ASK,
        timeout=1500, label="review")
    hold(15)

    beat("w3-44/tighten", "one finding, closed properly")
    W.turn(
        "Take the strongest finding from your own review and close it: write the "
        "test that would have caught it first, watch it fail, then make it pass. "
        "Show both runs. Change nothing else." + NO_ASK,
        timeout=1500, label="tighten")
    hold(12)
    W.turn("/exit")
    W.run("git diff --stat && npm test 2>&1 | tail -12", wait=8)
    K.wait_process("node --test", minutes=12, label="final suite")
    hold(22)


BEATS = [("w3-42", w3_42), ("w3-43", w3_43), ("w3-44", w3_44)]

if __name__ == "__main__":
    K.run_session(BEATS, "J", REPO)
