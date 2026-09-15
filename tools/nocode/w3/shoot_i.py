#!/usr/bin/env python3
"""CAPTURE I - what to do when it is wrong.

    python3 shoot_i.py            # w3-39 .. w3-41
    python3 shoot_i.py --only w3-40 --as I2

Nothing in this capture is staged. It runs against whatever the orchestrator
actually left behind in signal-desk, which is why it is written as a set of
questions rather than a set of outcomes: find the failure, get it diagnosed
from evidence rather than from a guess, and find out when the repository
stopped being correct.

If the suite is green when this rolls, w3-39 says so on camera and the lecture
becomes the same technique applied to the worst thing REVIEW.md found. What it
must never do is manufacture a bug so the lesson lands.
"""

from __future__ import annotations

import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import w3drive as W  # noqa: E402
import w3kit as K  # noqa: E402
from w3kit import beat, hold, NO_ASK  # noqa: E402

W.SHOOT_TITLE = "signal-desk"
REPO = "/Users/Shared/projects/signal-desk"


# ------------------------------------------------------------------ w3-39
def w3_39() -> None:
    beat("w3-39/state", "start from evidence, not from a feeling")
    W.run("git log --oneline | head -10", wait=6)
    hold(10)
    W.run("npm test", wait=10)
    K.wait_process("node --test", minutes=10, label="suite")
    hold(20)

    beat("w3-39/ask", "hand over the evidence, not the conclusion")
    W.run("claude --model sonnet", wait=8)
    W.trust()
    hold(4)
    W.turn(
        "Run npm test yourself and read the whole output. If anything fails, "
        "take the first failure only and work out what is actually wrong: read "
        "the failing test, read the code it exercises, and say which of the two "
        "is wrong and why. Show me the evidence before you propose a change. If "
        "the suite is green, say so plainly and instead find the single "
        "strongest claim in REVIEW.md that you can check right now, check it, "
        "and tell me whether the review was right. Do not fix anything yet." + NO_ASK,
        timeout=1200, label="diagnose")
    hold(10)

    beat("w3-39/fix", "and only then change something")
    W.turn(
        "Now make the smallest change that fixes what you just described, and "
        "nothing else. Run the suite before and after so both are on screen. If "
        "the fix does not work, say that rather than trying a second thing on "
        "top of the first." + NO_ASK,
        timeout=1200, label="fix")
    hold(10)
    W.turn("/exit")
    W.run("git diff --stat && git diff | head -60", wait=8)
    hold(20)


# ------------------------------------------------------------------ w3-40
def w3_40() -> None:
    """The history the pipeline did not keep.

    This was going to be a bisect lecture. It is not, because the honest state of
    the repository will not support one: seven tasks produced two commits, so
    there is nothing to bisect across. That is a better lecture than a staged
    one - the defect is in our own runner, it is one line, and fixing it means
    putting back the dependency the fix pass removed, which is the exception we
    still have not written down.
    """
    beat("w3-40/log", "seven agents, and how many commits")
    W.run("git log --oneline | cat", wait=6)
    hold(10)
    W.run("git log --stat --oneline | head -30", wait=6)
    hold(18)

    beat("w3-40/cost", "what that costs, in commands that do not work")
    W.run("claude --model sonnet", wait=8)
    W.trust()
    hold(4)
    W.turn(
        "This repository was built by tools/orchestrator.mjs - seven tasks, run "
        "with dependencies, all of it landing as untracked files that I then "
        "committed in one go. Look at the history and tell me concretely what "
        "that costs. Name the git commands that would work if each task had "
        "committed separately and do not work now, and for each one say what "
        "question it would have answered. Do not change anything yet." + NO_ASK,
        timeout=900, label="cost")
    hold(8)

    beat("w3-40/fix", "and the smallest change that would have prevented it")
    W.turn(
        "Now show me the smallest change to tools/orchestrator.mjs that would "
        "make each task its own commit, with a message naming the task and its "
        "area. Make the change. Explain why it belongs in the runner rather than "
        "in each task's brief, and say what it would do about a task that fails "
        "halfway through writing files." + NO_ASK,
        timeout=1200, label="fix")
    hold(8)
    W.turn("/exit")
    W.run("git diff --stat && git diff tools/orchestrator.mjs | head -50", wait=8)
    hold(20)

    beat("w3-40/exception", "and the sentence that should have been there all along")
    W.run("claude --model sonnet", wait=8)
    hold(4)
    W.turn(
        "One more thing, and it is the one that cost us this morning. "
        "tools/orchestrator.mjs needs @anthropic-ai/claude-agent-sdk, and "
        "CLAUDE.md says zero dependencies - so the last agent to read that rule "
        "deleted the dependency and left the tool unable to run. The rule was "
        "incomplete, not wrong. Install the package again, and amend CLAUDE.md to "
        "record the exception precisely: which file needs it, why it is a "
        "development tool rather than part of the shipped app, and what stays "
        "true regardless - that the console still starts with no network and "
        "nothing beyond Node. Then show me the diff of CLAUDE.md." + NO_ASK,
        timeout=1200, label="exception")
    hold(8)
    W.turn("/exit")
    W.run("git diff CLAUDE.md | head -40", wait=8)
    hold(20)
    W.run("git add -A && git commit -q -m "
          "'orchestrator: one commit per task, and write the dependency exception down' "
          "&& git log --oneline | cat", wait=8)
    hold(12)


# ------------------------------------------------------------------ w3-41
def w3_41() -> None:
    beat("w3-41/two", "two whole products, two different machines")
    W.run("clear && printf '\\033[3J'; "
          "echo '  control-tower  - built by an interactive agent team'; "
          "echo '  signal-desk    - built by a program we wrote'", wait=6)
    hold(12)
    # Counted the same way on both sides, and by directory layout rather than by
    # a hardcoded src/ - the two repositories do not agree on where source lives,
    # and a comparison that measures them differently is worse than no numbers.
    W.run("for r in /Users/Shared/projects/control-tower "
          "/Users/Shared/projects/signal-desk; do echo \"== $(basename $r)\"; "
          "(cd $r && "
          "git log --oneline 2>/dev/null | wc -l | xargs echo '   commits:' ; "
          "find . -name '*.js' -not -path './node_modules/*' -not -path './.git/*' "
          "-not -path './tools/*' | wc -l | xargs echo '   js files:' ; "
          "find . -name '*.js' -not -path './node_modules/*' -not -path './.git/*' "
          "-not -path './tools/*' | xargs wc -l 2>/dev/null | tail -1 "
          "| awk '{print \"   lines:   \" $1}'); done", wait=10)
    hold(24)

    beat("w3-41/verdict", "and the honest verdict")
    W.run("claude --model sonnet", wait=8)
    hold(4)
    W.turn(
        "Two products in /Users/Shared/projects: control-tower, built by an "
        "interactive agent team with a person in the room, and signal-desk, "
        "built by tools/orchestrator.mjs with nobody watching. Read both "
        "repositories - the plans, the run logs, the review, the test output, "
        "the commit history - and write VERDICT.md in signal-desk comparing "
        "them on four things: what each one actually delivered against its own "
        "plan, where each one went wrong, what each cost in wall-clock time, and "
        "which of the two you would choose for what kind of work. Ground every "
        "claim in a file you read. Where the evidence is missing, say it is "
        "missing. Do not decide which is better in general - decide what each is "
        "for." + NO_ASK,
        timeout=1800, label="verdict")
    hold(10)
    W.turn("/exit")
    K.read_doc("VERDICT.md", pages=12, dwell=3.4)


BEATS = [("w3-39", w3_39), ("w3-40", w3_40), ("w3-41", w3_41)]

if __name__ == "__main__":
    K.run_session(BEATS, "I", REPO)
