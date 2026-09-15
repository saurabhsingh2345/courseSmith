#!/usr/bin/env python3
"""CAPTURE G - the loop that runs on somebody else's computer.

    python3 shoot_g.py            # w3-33 .. w3-35
    python3 shoot_g.py --only w3-34 --as G2

This is the replacement for the lecture that needed the hosted Claude Code
GitHub Action. The Action is a webhook and an install button; the thing worth
teaching is the loop underneath it - an issue is a brief, a branch is a
sandbox, a pull request is a review, and CI is the only opinion that is not
negotiable. Every step of that runs from the terminal with the gh CLI, on a
real repository, with a real Actions run.

The repository is created inside an organisation rather than under a personal
account, because gh prints owner/repo on almost every command and a personal
account name is a person's name.
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
SLUG = "heftyradius/signal-desk"


# ------------------------------------------------------------------ w3-33
def w3_33() -> None:
    beat("w3-33/local", "everything so far has been on this laptop")
    W.run("git log --oneline | head -8 && echo '---' && git status --short | head -5",
          wait=5)
    hold(10)

    beat("w3-33/create", "put it where a team would keep it")
    W.run(f"gh repo create {SLUG} --private --source=. --remote=origin --push", wait=8)
    K.wait_process("gh repo create", minutes=6, label="repo create")
    hold(10)
    W.run(f"gh repo view {SLUG} --json name,visibility,defaultBranchRef "
          f"--jq '{{repo:.name, visibility:.visibility, branch:.defaultBranchRef.name}}'",
          wait=6)
    hold(10)

    beat("w3-33/ci", "the only opinion that is not negotiable")
    W.run("claude --model sonnet", wait=8)
    W.trust()
    hold(4)
    W.turn(
        "Write .github/workflows/ci.yml for this repository. It runs on push to "
        "main and on every pull request. One job on ubuntu latest: check out the "
        "code, set up Node 24, and run npm test. The app has no dependencies, so "
        "do not add an install step it does not need, and do not add a cache for "
        "a lockfile that may not exist. Keep it short enough to read in one "
        "screen, and explain what each step is for as you write it." + NO_ASK,
        timeout=600, label="ci")
    hold(6)
    W.turn("/exit")
    K.read_doc(".github/workflows/ci.yml", pages=4, dwell=3.4)

    beat("w3-33/push", "and let a computer that is not mine run the tests")
    W.run("git add -A && git commit -q -m 'ci: run the suite on every push and pull request' "
          "&& git push -q origin main && echo pushed", wait=10)
    K.wait_process("git push", minutes=5, label="push")
    hold(8)
    W.run("gh run list --limit 3", wait=8)
    hold(10)

    beat("w3-33/green", "watch it go green")
    # Name the run rather than letting `gh run watch` pick: with no argument it
    # errors out when the run has already finished, and an error message is not
    # what this beat is for.
    W.run("gh run watch $(gh run list --limit 1 --json databaseId "
          "--jq '.[0].databaseId') --exit-status", wait=10)
    K.wait_process("gh run watch", minutes=15, label="ci run")
    hold(14)
    W.run("gh run list --limit 3", wait=6)
    hold(12)


# ------------------------------------------------------------------ w3-34
def w3_34() -> None:
    beat("w3-34/pick", "an issue is a brief with an address")
    W.run("claude --model sonnet", wait=8)
    hold(4)
    W.turn(
        "Read REVIEW.md and, if it exists, FIXED.md. Pick the single most "
        "worthwhile thing that is still not done or still wrong in this "
        "repository - something small enough to fix in one change and real "
        "enough to matter. Then file it as a GitHub issue with gh issue create: "
        "a title a stranger could scan, and a body that says what is wrong, "
        "where, how to see it, and what done looks like. File the issue and stop. "
        "Do not fix anything." + NO_ASK,
        timeout=600, label="issue")
    hold(6)
    W.turn("/exit")

    beat("w3-34/view", "what a good issue looks like")
    W.run("gh issue list", wait=6)
    hold(8)
    W.run("gh issue view 1", wait=6)
    hold(16)

    beat("w3-34/hand", "hand over the address, not the work")
    W.run("claude --model sonnet", wait=8)
    hold(4)
    W.turn(
        "Run gh issue view 1 and read it. Then close it properly: create a "
        "branch named for the issue, write a test that fails because of the bug "
        "and would pass once it is fixed, run npm test and watch it fail, make "
        "the fix, run npm test again, commit with a message that explains the "
        "change rather than restating the diff, push the branch, and open a pull "
        "request with gh pr create whose body closes the issue. If the fix turns "
        "out to be bigger than the issue described, say so in the pull request "
        "rather than quietly widening it." + NO_ASK,
        timeout=1800, label="fix")
    hold(8)
    W.turn("/exit")
    W.run("git log --oneline | head -5 && git branch --show-current", wait=6)
    hold(12)


# ------------------------------------------------------------------ w3-35
def w3_35() -> None:
    beat("w3-35/pr", "the pull request")
    W.run("gh pr list", wait=6)
    hold(8)
    W.run("gh pr view", wait=6)
    hold(16)

    beat("w3-35/diff", "read the diff, because this is the review")
    W.run("gh pr diff | head -80", wait=8)
    hold(22)

    beat("w3-35/ci", "and the part that does not care how good the explanation was")
    W.run("sleep 8; gh pr checks --watch --interval 10", wait=10)
    K.wait_process("gh pr checks", minutes=15, label="pr checks")
    hold(14)

    beat("w3-35/merge", "merge it")
    W.run("gh pr merge --squash --delete-branch", wait=10)
    K.wait_process("gh pr merge", minutes=6, label="merge")
    hold(10)
    W.run("git checkout -q main && git pull -q && git log --oneline | head -6", wait=8)
    hold(12)
    W.run("gh issue list --state all --limit 3", wait=6)
    hold(12)


BEATS = [("w3-33", w3_33), ("w3-34", w3_34), ("w3-35", w3_35)]

if __name__ == "__main__":
    K.run_session(BEATS, "G", REPO)
