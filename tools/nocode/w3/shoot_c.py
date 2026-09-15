#!/usr/bin/env python3
"""CAPTURE C — Week 3 Day 4. The agent team builds Control Tower.

    python3 shoot_c.py [w3-17]

This is the session that cannot be repeated cheaply, so it is written to keep
filming through anything: permission prompts are answered, the teammate panes
are cycled on a timer so the parallel work is actually on screen, and a beat is
marked whenever the picture changes even though nobody is watching.

The build's outcome is NOT scripted. Whatever it does — including failing — is
the lecture. Result narration gets written afterwards, from the footage.
"""

from __future__ import annotations

import os
import subprocess
import sys
import time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import rig  # noqa: E402
import w3drive as W  # noqa: E402
from w3drive import S  # noqa: E402

REPO = "/Users/Shared/projects/control-tower"
s: rig.Session = None


def beat(label, note=""):
    s.mark(label, note)


def hold(x):
    time.sleep(x)


def watch(minutes: float, label: str, flip_every: float = 55.0) -> None:
    """Film a long build: answer prompts, cycle teammate panes, keep marking.

    Sitting on one pane for forty minutes films a wall of text. Cycling gives
    the cut somewhere to go, and it is also the only way to SHOW that several
    agents are working at once — which is the entire claim of the lecture.
    """
    end = time.time() + minutes * 60
    n = 0
    while time.time() < end:
        time.sleep(min(flip_every, max(5.0, end - time.time())))
        try:
            if W.prompt_showing():
                print("    · approving", flush=True)
                W.choose("2")
                hold(1.5)
                continue
            W.focus_shoot_window()
            S.clear_mods()
            S.key("down" if n % 2 else "up", ("shift",))
            n += 1
            if n % 4 == 0:
                beat(f"{label}/pane{n//4}", "cycling the teammates")
        except W.WrongWindow as exc:
            print(f"    !! {exc}", flush=True)
            time.sleep(5)


# ---------------------------------------------------------------- w3-15
def w3_15() -> None:
    beat("w3-15/limit", "what a sub-agent cannot do")
    W.run("claude --model sonnet", wait=8)
    W.wait_idle(stable=6, timeout=120, label="(launch)")
    W.turn(
        "We built a sub-agent on day one. In plain words, and with our own repo "
        "as the example: what can a sub-agent NOT do that would be useful right "
        "now, if I wanted six of them building different parts of Control Tower "
        "at the same time?",
        timeout=600, label="(the limit)")
    hold(6)

    beat("w3-15/teams", "the thing that can")
    W.turn(
        "Now explain agent teams against that answer. Cover: each teammate has "
        "its own full context, they can message each other, they share one task "
        "list, and I can talk to any of them directly. Then tell me honestly "
        "when a sub-agent is still the better choice.",
        timeout=600, label="(teams vs subagents)")
    hold(6)

    beat("w3-15/flag", "it is still experimental, and here is the switch")
    W.turn(
        "What exactly do I have to put in settings for agent teams to work in "
        "this version, and what are the valid values for teammate mode? Check "
        "what this installed version actually supports rather than what you "
        "remember.",
        timeout=600, label="(the flag)")
    hold(6)
    W.turn("/exit", timeout=60, stable=4, label="(exit)")
    hold(2)


# ---------------------------------------------------------------- w3-16
def w3_16() -> None:
    beat("w3-16/audit", "clean house before going loud")
    W.run("claude --model sonnet", wait=8)
    W.wait_idle(stable=6, timeout=120, label="(launch)")
    for cmd in ("/plugin", "/mcp"):
        W.turn(cmd, timeout=240, stable=5, label=f"({cmd})")
        hold(5)
        S.clear_mods(); S.key("escape"); hold(2)

    beat("w3-16/claudemd", "a memory file seven agents will read")
    W.turn(
        "Seven separate agents are about to read CLAUDE.md, each with its own "
        "context. Rewrite it for that audience: the conventions they must all "
        "share, what is already built, and what is off limits. Keep it short — "
        "every word costs seven times what it used to.",
        timeout=900, label="(claude.md for a team)")
    hold(5)

    beat("w3-16/usage", "write this number down")
    W.turn("/usage", timeout=240, stable=5, label="(usage before)")
    hold(8)
    S.clear_mods(); S.key("escape"); hold(2)
    W.turn("/exit", timeout=60, stable=4, label="(exit)")

    beat("w3-16/git", "a branch to come back to")
    W.run("git add -A && git commit -q -m 'ready for the team' && git log --oneline -1", wait=5)
    hold(3)
    W.run("git checkout -b agent-team", wait=3)
    hold(3)

    beat("w3-16/settings", "the switch itself")
    # MERGE, never overwrite. settings.json already holds the content hook from
    # day one and the project-scoped plugin; a heredoc here would silently delete
    # both and the rest of the week would be filmed without them.
    W.run("claude --model sonnet", wait=9)
    W.wait_idle(stable=6, timeout=180, label="(launch)")
    W.turn(
        "Add two things to .claude/settings.json without touching anything that "
        "is already in it: the environment variable that switches on experimental "
        "agent teams, and the teammate mode. Check what values this installed "
        "version actually accepts rather than assuming, then show me the whole "
        "file so I can see the hook is still there.",
        timeout=600, label="(merge the flag)")
    hold(6)
    W.turn("/exit", timeout=60, stable=4, label="(exit)")
    W.run("cat .claude/settings.json", wait=3)
    hold(8)


# ---------------------------------------------------------------- w3-17
def w3_17() -> None:
    beat("w3-17/roster", "who is on the team, and who is deliberately not")
    W.run("claude --model sonnet", wait=9)
    W.wait_idle(stable=6, timeout=180, label="(launch with teams on)")
    hold(4)

    beat("w3-17/kick", "one instruction, seven agents")
    W.say(
        "Create an agent team to build the whole of Control Tower from plan.md. "
        "Teammates: a simulator engineer, a back-end engineer for the API and the "
        "live feed, a front-end engineer for the fleet list and the map, a second "
        "front-end engineer for the heat map and the chart, a chat engineer for "
        "the dispatcher, and an integration tester. Every engineer writes their "
        "own unit tests. No code reviewer — I do not want six conversations "
        "about style while they are trying to build. Wait for your teammates to "
        "finish their tasks before you start doing the work yourself.")
    hold(3)

    beat("w3-17/explore", "it reads the ground first")
    watch(6, "w3-17")

    beat("w3-17/spawn", "the team appears")
    watch(6, "w3-17")


# ---------------------------------------------------------------- w3-18
def w3_18() -> None:
    """What the team actually did.

    The build has already finished, so this is not a live watch — it is the
    forensics, which is the more useful lecture anyway: the evidence that six
    agents worked in parallel is in the repository, not in a scrolling pane.
    """
    beat("w3-18/shape", "what six agents left behind")
    W.run("git status --short | wc -l; echo ---; git status --short | head -25", wait=6)
    hold(10)

    beat("w3-18/split", "who owned what")
    W.run("find src -type f | sort", wait=5)
    hold(10)

    beat("w3-18/ask", "and ask the lead what happened")
    W.run("claude --model sonnet", wait=9)
    W.wait_idle(stable=6, timeout=180, label="(launch)")
    W.turn(
        "The team you spawned has finished. Walk me through what actually "
        "happened: which teammate owned which files, what ran in parallel and "
        "what had to wait, and where any two of them touched the same thing. "
        "Be specific to this repo, and tell me anything that looks wrong.",
        timeout=900, label="(forensics)")
    hold(10)

    beat("w3-18/cost", "and what it cost")
    W.turn("/usage", timeout=240, stable=5, label="(usage after)")
    hold(10)
    S.clear_mods(); S.key("escape"); hold(2)
    W.turn("/exit", timeout=60, stable=4, label="(exit)")


# ---------------------------------------------------------------- w3-19
def w3_19() -> None:
    beat("w3-19/tests", "so does it work")
    W.run("npm test 2>&1 | tail -30", wait=45)
    hold(12)

    beat("w3-19/read", "read the failures honestly")
    W.run("claude --model sonnet", wait=9)
    W.wait_idle(stable=6, timeout=180, label="(launch)")
    W.turn(
        "Run the test suite and tell me the truth about it. How many tests, how "
        "many pass, and — the question that matters — would these tests fail if "
        "the behaviour were wrong? Pick the two weakest tests in the suite and "
        "show me why they are weak. Do not fix anything yet.",
        timeout=900, label="(judge the tests)")
    hold(12)

    beat("w3-19/fix", "then fix what is actually broken")
    W.turn(
        "Now fix only the things that are genuinely broken — code, not tests — "
        "and re-run. If a test is wrong rather than the code, say so and change "
        "the test, but tell me which you did and why.",
        timeout=1500, stable=10, label="(fix)")
    hold(12)
    W.turn("/exit", timeout=60, stable=4, label="(exit)")


# ---------------------------------------------------------------- w3-20
def w3_20() -> None:
    beat("w3-20/start", "start it")
    W.run("ls scripts 2>/dev/null; cat package.json | head -20", wait=6)
    hold(8)
    W.run("(bash scripts/start.sh >/tmp/ct.log 2>&1 &) ; sleep 8; tail -5 /tmp/ct.log", wait=20)
    hold(8)

    beat("w3-20/alive", "Control Tower")
    W.run("'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' "
          "--user-data-dir=/Users/Shared/projects/.shoot-chrome --no-first-run "
          "--no-default-browser-check --window-position=0,0 "
          "--window-size=1920,1080 --app=http://localhost:8080 "
          ">/dev/null 2>&1 &", wait=14)
    hold(50)

    beat("w3-20/warts", "and what is wrong with it")
    W.focus_shoot_window()
    hold(6)
    W.run("git add -A && git commit -q -m 'agent team v1' && git log --oneline | head -3",
          wait=8)
    hold(10)


BEATS = [("w3-15", w3_15), ("w3-16", w3_16), ("w3-17", w3_17),
         ("w3-18", w3_18), ("w3-19", w3_19), ("w3-20", w3_20)]


def _arg(flag, default=None):
    return sys.argv[sys.argv.index(flag) + 1] if flag in sys.argv else default


def main() -> None:
    global s
    name = _arg("--as", "C")
    positional = [a for a in sys.argv[1:] if a.startswith("w3-")]
    start = positional[0] if positional else None
    todo = BEATS
    if start:
        idx = [i for i, (n, _) in enumerate(BEATS) if n == start]
        if not idx:
            sys.exit(f"unknown lecture {start}")
        todo = BEATS[idx[0]:]

    s = rig.Session(name, allow=("Code", "Google Chrome"))
    s.start(need_gb=25)
    t0 = time.time()
    try:
        for name, fn in todo:
            print(f"\n=== {name} ===", flush=True)
            fn()
    except Exception as exc:
        print(f"!! driver error: {exc!r}", flush=True)
    finally:
        s.stop()
        print(f"session ran {(time.time()-t0)/60:.1f} min", flush=True)


if __name__ == "__main__":
    main()
