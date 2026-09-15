#!/usr/bin/env python3
"""CAPTURE A1 — Week 3 Day 1. One continuous take, five lectures cut from it.

    python3 shoot_a1.py                       # the whole session
    python3 shoot_a1.py w3-04                 # resume from a lecture
    python3 shoot_a1.py --only w3-04 --as A1r # re-shoot one lecture into its own
                                              # capture; cut.py merges the two

Nothing here is narrated. Narration is written afterwards, to the footage that
actually exists — which is the whole reason the week is shot this way.

Everything is done by ASKING the agent, never by typing code. That is the
course: the student directs, the agent builds.
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
# The STUDENT-FACING brief. control-tower-spec.md is the production master and
# opens with notes about how the week is shot — it must never be on screen.
SPEC = os.path.join(HERE, "..", "..", "..", "docs", "nocode-course", "w3",
                    "plan.md")

s: rig.Session = None  # set in main


def beat(label: str, note: str = "") -> None:
    s.mark(label, note)


def hold(sec: float) -> None:
    """Let a frame breathe. Dead air on camera is ramped 12-14x in post."""
    time.sleep(sec)


def expect_new_code(minimum: int = 2) -> None:
    """Did the build actually produce application files?

    Naming the exact path is too brittle — the agent chooses the layout. What
    matters is that something outside .claude and the brief now exists.
    """
    found = []
    for root, dirs, files in os.walk(REPO):
        dirs[:] = [d for d in dirs if d not in (".git", ".claude", "node_modules")]
        for f in files:
            if f not in ("plan.md", "CLAUDE.md") and not f.startswith("."):
                found.append(os.path.relpath(os.path.join(root, f), REPO))
    if len(found) >= minimum:
        print(f"    ok  {len(found)} file(s) built, e.g. {found[:4]}", flush=True)
    else:
        print(f"    !! NOTHING WAS BUILT — only {found}", flush=True)


def expect(*rel: str) -> None:
    """Shout if a beat produced nothing.

    A modal panel left open by a built-in slash command swallows every prompt
    that follows it, so the driver sails on happily while the agent never hears
    a word. The log looked fine; the repo was empty. Check the ground truth.
    """
    # Give the write a moment to land: idle-detection can return while the
    # agent's last file write is still flushing, and a false alarm here is worse
    # than no alarm — it trains you to ignore the one that matters.
    for _ in range(12):
        missing = [r for r in rel if not os.path.exists(os.path.join(REPO, r))]
        if not missing:
            print(f"    ok  {', '.join(rel)}", flush=True)
            return
        time.sleep(1.0)
    print(f"    !! NOTHING WAS BUILT — missing {missing}", flush=True)


# ---------------------------------------------------------------- w3-02
def w3_02() -> None:
    beat("w3-02/empty", "the empty folder")
    W.clean_prompt()
    W.run("ls -a", wait=2.5)
    hold(3)

    beat("w3-02/git", "git init")
    W.run("git init", wait=3.0)
    hold(2)

    beat("w3-02/spec", "the brief lands")
    # Write it from the terminal rather than through Save As. VS Code's save
    # dialog appends the detected extension to whatever is typed, which produced
    # `plan.md.md` on the first take and sent the agent off fixing a filename
    # instead of reading the brief.
    subprocess.run(["pbcopy"], stdin=open(os.path.abspath(SPEC), "rb"), check=True)
    W.run("pbpaste > plan.md", wait=2.0)
    hold(2)
    W.run("open -a 'Visual Studio Code' plan.md", wait=3.5)
    hold(2)
    # Read it on camera. This document is what the whole week builds.
    S.clear_mods(); S.key("up", ("cmd",)); hold(1.5)   # top of file; "home" is not a mapped key
    for _ in range(14):
        S.scroll(-3, steps=6); hold(1.1)
    hold(2)

    beat("w3-02/first", "first Claude Code session in this repo")
    W.palette("Terminal: Focus Terminal", settle=2.0)
    W.run("claude --model sonnet", wait=8)
    W.trust()
    W.wait_idle(stable=6, timeout=120, label="(banner)")
    hold(3)

    beat("w3-02/context", "how little is loaded")
    W.turn("/context", timeout=180, label="(context)")
    hold(4)

    beat("w3-02/claudemd", "the memory file, written by the agent")
    W.turn(
        "Write claude.md for this project. Rules: include plan.md with an @ tag "
        "because every session needs the brief; refer to any other document by "
        "its path in backticks instead, so you only read it when the work "
        "actually needs it. Keep it under 30 lines. No emojis.",
        timeout=600, label="(claude.md)")
    hold(3)
    W.turn("/context", timeout=180, label="(context after)")
    hold(4)


# ---------------------------------------------------------------- w3-03
def w3_03() -> None:
    beat("w3-03/why", "the thing we keep asking for by hand")
    W.turn("Give me a one-paragraph status of this project: what exists, what "
           "does not, and what the obvious next step is.",
           timeout=420, label="(manual status)")
    hold(3)

    beat("w3-03/make", "turn it into one word")
    W.turn(
        "Create a slash command called ct-brief in .claude/commands. It should "
        "print what exists in the repo, what is still missing against plan.md, "
        "and the single next step. Keep the command file short.",
        timeout=600, label="(make /ct-brief)")
    expect(".claude/commands/ct-brief.md")
    hold(3)

    beat("w3-03/run", "run it")
    W.turn("/ct-brief", timeout=420, label="(run /ct-brief)")
    hold(4)

    beat("w3-03/arg", "a command that takes an argument")
    W.turn(
        "Now create a second slash command called ct-check that takes one argument: "
        "a requirement number from plan.md. It should report whether that "
        "requirement is built, partly built, or not started, and why.",
        timeout=600, label="(make /ct-check)")
    expect(".claude/commands/ct-check.md")
    hold(2)
    W.turn("/ct-check 1", timeout=420, label="(run /ct-check)")
    hold(4)

    beat("w3-03/scope", "where they live, and who gets them")
    W.turn("/exit", timeout=60, stable=4, label="(exit)")
    hold(1)
    W.run("ls -R .claude", wait=3.0)
    hold(4)


# ---------------------------------------------------------------- w3-04
def w3_04() -> None:
    beat("w3-04/why", "one agent, one context window")
    W.run("claude --model sonnet", wait=8)
    W.wait_idle(stable=6, timeout=120, label="(relaunch)")
    W.turn("/context", timeout=180, label="(context)")
    hold(4)

    beat("w3-04/make", "hire a specialist")
    W.turn(
        "Create a sub-agent called simulator. Its job is only the delivery "
        "simulator described in plan.md: vehicles, routes, ticks, exceptions. "
        "Give it a tight description so you know when to hand work to it, and "
        "only the tools it needs.",
        timeout=600, label="(make sub-agent)")
    hold(3)

    beat("w3-04/notyet", "and it is not there yet")
    # A brand new sub-agent is NOT addressable in the session that created it:
    # the agent registry is read at session start. This is a real gotcha and it
    # is better teaching to show it happening than to quietly restart first.
    W.turn(
        "Use the simulator sub-agent to build the simulator from plan.md.",
        timeout=300, stable=7, label="(agent not found)")
    hold(6)

    beat("w3-04/restart", "so we restart, and now it exists")
    W.turn("/exit", timeout=60, stable=4, label="(exit)")
    hold(2)
    # A genuinely cold start. `/exit` then `claude` in the SAME terminal left the
    # agent registry stale — the new sub-agent was still "not found" afterwards,
    # while a fresh process in the same folder finds it immediately. Killing the
    # terminal guarantees a new shell and a new Claude Code process.
    W.one_terminal()
    hold(2)
    W.run("claude --model sonnet", wait=10)
    W.wait_idle(stable=6, timeout=180, label="(cold start)")
    hold(3)

    beat("w3-04/run", "give it real work")
    W.turn(
        "Use the simulator sub-agent to build the simulator from plan.md: "
        "60 vehicles, six regions, a one-second tick, ETAs that recompute, and "
        "random exceptions. Deterministic when given a seed. Write it and its "
        "unit tests. Nothing else in the app yet.",
        timeout=2400, stable=12, label="(sub-agent builds)")
    expect_new_code()
    hold(6)

    beat("w3-04/report", "what came back, and what did not")
    W.turn("/context", timeout=180, label="(context after)")
    hold(4)

    beat("w3-04/explore", "the one you already have")
    W.turn("Use the explore sub-agent to tell me what is now in this repo and "
           "how the pieces fit together. Keep it short.",
           timeout=600, label="(explore)")
    hold(4)


# ---------------------------------------------------------------- w3-05
def w3_05() -> None:
    beat("w3-05/why", "the rule we keep forgetting")
    W.turn("Read plan.md and tell me the one rule about this project that is "
           "easiest to break by accident while writing code.",
           timeout=420, label="(the rule)")
    hold(3)

    beat("w3-05/make", "make the rule fire by itself")
    W.turn(
        "Add a hook to .claude/settings.json that runs after every file edit and "
        "fails loudly if the edited file contains an emoji or a real person's "
        "name. Write the check as a small script in the repo so I can read it.",
        timeout=900, label="(make hook)")
    expect(".claude/settings.json")
    hold(3)

    beat("w3-05/fire", "break it on purpose")
    W.turn("Add a line with an emoji to a scratch file called hook-test.txt so "
           "we can watch the hook catch it.",
           timeout=600, label="(fire hook)")
    hold(5)

    beat("w3-05/clean", "and clean up after ourselves")
    W.turn("Delete hook-test.txt.", timeout=300, label="(clean)")
    hold(3)

    beat("w3-05/events", "the other moments you can hook")
    W.turn("What other events can a hook fire on, and give me one useful example "
           "for this project that is not a linter.",
           timeout=420, label="(events)")
    hold(5)


# ---------------------------------------------------------------- w3-06
def w3_06() -> None:
    beat("w3-06/browse", "the marketplace")
    W.turn("/plugin", timeout=240, stable=5, label="(plugin browse)")
    hold(6)
    S.clear_mods(); S.key("escape"); hold(2)

    beat("w3-06/install", "install one, at project scope")
    W.turn("Install the frontend-design plugin for this project only. Do not "
           "install anything that spawns its own sub-agents — we would lose "
           "track of who is doing what.",
           timeout=600, label="(install)")
    hold(4)

    beat("w3-06/cost", "what it cost us in context")
    W.turn("/context", timeout=180, label="(context)")
    hold(4)

    beat("w3-06/own", "package our own setup")
    W.turn(
        "Package the two slash commands, the hook and the simulator sub-agent we "
        "built today into a plugin inside this repo, with a plugin manifest, so "
        "someone else can install the whole setup in one command. Explain the "
        "folder layout when you are done.",
        timeout=1200, label="(own plugin)")
    hold(5)

    beat("w3-06/close", "a whole workshop, and no app code yet")
    W.turn("/ct-brief", timeout=420, label="(final status)")
    hold(6)


BEATS = [("w3-02", w3_02), ("w3-03", w3_03), ("w3-04", w3_04),
         ("w3-05", w3_05), ("w3-06", w3_06)]


def _arg(flag: str, default=None):
    return sys.argv[sys.argv.index(flag) + 1] if flag in sys.argv else default


def main() -> None:
    global s
    only = _arg("--only")
    name = _arg("--as", "A1")
    positional = [a for a in sys.argv[1:] if a.startswith("w3-")]
    start = positional[0] if positional and not only else None

    todo = BEATS
    if only:
        todo = [(n, f) for n, f in BEATS if n == only]
        if not todo:
            sys.exit(f"unknown lecture {only}")
    elif start:
        idx = [i for i, (n, _) in enumerate(BEATS) if n == start]
        if not idx:
            sys.exit(f"unknown lecture {start}")
        todo = BEATS[idx[0]:]

    s = rig.Session(name, allow=("Code",))
    s.start(need_gb=25)
    t0 = time.time()
    try:
        for name, fn in todo:
            print(f"\n=== {name} ===", flush=True)
            fn()
    except Exception as exc:  # never lose the footage to a driver bug
        print(f"!! driver error: {exc!r}", flush=True)
    finally:
        s.stop()
        print(f"session ran {(time.time()-t0)/60:.1f} min", flush=True)


if __name__ == "__main__":
    main()
