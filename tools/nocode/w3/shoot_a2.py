#!/usr/bin/env python3
"""CAPTURE A2 — Week 3 Day 3, lectures w3-12 and w3-13.

    python3 shoot_a2.py [w3-13]

w3-14 (Cowork) is a separate script: it drives the Claude desktop app rather
than VS Code, so it needs its own window guard and its own frame checks.

Runs after A1, because w3-12 is shot over a repo that has history in it.
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

SDK_DIR = "/Users/Shared/projects/sdk-demo"
# A throwaway Chrome profile. Opening the game in his normal Chrome would put
# his tabs, bookmarks bar and profile picture on camera.
CHROME_PROFILE = "/Users/Shared/projects/.shoot-chrome"
CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
s: rig.Session = None


def beat(label, note=""):
    s.mark(label, note)


def hold(x):
    time.sleep(x)


# ---------------------------------------------------------------- w3-12
def w3_12() -> None:
    beat("w3-12/repo", "what a week of work looks like")
    W.run("claude --model sonnet", wait=8)
    W.wait_idle(stable=6, timeout=120, label="(launch)")
    W.turn("/ct-brief", timeout=420, label="(where we are)")
    hold(4)

    beat("w3-12/rules", "the seven rules, against this repo")
    W.turn(
        "You are teaching someone who has never worked in a large codebase. "
        "Using THIS repo as the example, walk through the seven habits that keep "
        "an agent effective as a project grows: documentation at every level; "
        "describing and linking a document instead of pasting it; one agreed "
        "workflow for the team; plugins first then skills; tests that survive a "
        "rewrite; a human who stays accountable; and work cut into pieces small "
        "enough to review. For each one, point at a real file we already have, "
        "or say plainly that we do not have it yet.",
        timeout=900, label="(seven rules)")
    hold(6)

    beat("w3-12/at", "fix our own memory file")
    W.turn(
        "Look at claude.md again. Is there anything in it that we pull into "
        "context every single session but only actually need occasionally? If "
        "so, change the @ tag to a linked path and tell me what that saves.",
        timeout=600, label="(the @ fix)")
    hold(5)

    beat("w3-12/tests", "the coverage trap, on purpose")
    W.turn(
        "Write me the kind of test suite a coverage target produces: aim for "
        "80% on the simulator and mock whatever gets in the way. Put it in "
        "tests/coverage_demo.py. I want to look at what that actually buys us.",
        timeout=900, label="(slop tests)")
    hold(5)

    beat("w3-12/reject", "and reject it")
    W.turn(
        "Now read that file back to me critically. Which of those tests would "
        "still pass if the simulator were reimplemented from scratch, and which "
        "would break for no good reason? Delete the ones that only exist to move "
        "the coverage number, and say why each one went.",
        timeout=900, label="(reject slop)")
    hold(6)
    W.turn("/exit", timeout=60, stable=4, label="(exit)")
    hold(2)


# ---------------------------------------------------------------- w3-13
def w3_13() -> None:
    beat("w3-13/empty", "a brand new empty folder")
    subprocess.run(["rm", "-rf", SDK_DIR])
    os.makedirs(SDK_DIR, exist_ok=True)
    # A fresh terminal, not just a `cd`. Typing into the old one straight after
    # Claude Code exits is unreliable — its TUI is still releasing the terminal,
    # the `cd` is swallowed, and everything after it runs in the WRONG FOLDER.
    # That put a Python venv inside the Control Tower repo on the first attempt.
    W.run(f"cd {SDK_DIR}", wait=2.0)
    W.clean_prompt()
    # Prove on camera that we really are somewhere else before installing into it.
    W.run("pwd && ls -a", wait=3.0)
    hold(4)

    beat("w3-13/setup", "three commands and a name worth getting right")
    # Every command carries the folder with it, so a swallowed keystroke can
    # never install into the wrong project again.
    W.run(f"cd {SDK_DIR} && uv init --bare", wait=8)
    hold(2)
    W.run(f"cd {SDK_DIR} && uv python pin 3.13", wait=8)
    hold(2)
    W.run(f"cd {SDK_DIR} && uv add claude-agent-sdk python-dotenv", wait=55)
    hold(5)

    beat("w3-13/ask", "we do not write this by hand either")
    W.run("claude --model sonnet", wait=8)
    W.wait_idle(stable=6, timeout=120, label="(launch)")
    W.trust()
    W.wait_idle(stable=6, timeout=120, label="(ready)")
    W.turn(
        "Write main.py in this folder using the claude-agent-sdk. It should hold "
        "one prompt and a list of allowed tools, build ClaudeAgentOptions from "
        "them, then loop over query() printing each message as it arrives. The "
        "prompt: build a small browser game of asteroids in plain HTML, CSS and "
        "JavaScript, and write the files into this directory including "
        "index.html. Keep main.py under 30 lines and explain each line to me "
        "afterwards as if I have never written Python.",
        timeout=900, label="(write main.py)")
    hold(6)

    beat("w3-13/model", "the expensive mistake to avoid")
    W.turn(
        "Which model will this use, and what would it cost me if I put this in a "
        "loop that runs all night? Give me the one-line change that makes it "
        "cheap, and tell me when the expensive one is actually worth it.",
        timeout=600, label="(model warning)")
    hold(6)
    W.turn("/exit", timeout=60, stable=4, label="(exit)")
    hold(2)

    beat("w3-13/run", "run it, and watch an empty folder fill")
    W.run(f"cd {SDK_DIR} && uv run main.py", wait=8)
    W.wait_idle(stable=12, timeout=1800, label="(sdk build)")
    hold(6)

    beat("w3-13/play", "open what it made")
    W.run(f"cd {SDK_DIR} && ls -la", wait=3)
    hold(4)
    W.run(f"'{CHROME}' --user-data-dir={CHROME_PROFILE} --no-first-run "
          f"--no-default-browser-check --new-window {SDK_DIR}/index.html "
          f">/dev/null 2>&1 &", wait=10)
    hold(14)
    W.focus_shoot_window()
    hold(3)

    beat("w3-13/point", "what this actually was")
    W.run(f"cd /Users/Shared/projects/control-tower", wait=1.5)
    W.clean_prompt()
    hold(3)


BEATS = [("w3-12", w3_12), ("w3-13", w3_13)]


def _arg(flag, default=None):
    return sys.argv[sys.argv.index(flag) + 1] if flag in sys.argv else default


def main() -> None:
    global s
    only = _arg("--only")
    name = _arg("--as", "A2")
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

    s = rig.Session(name, allow=("Code", "Google Chrome"))
    s.start(need_gb=20)
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
