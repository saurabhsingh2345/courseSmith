#!/usr/bin/env python3
"""CAPTURE B — Week 3 Day 2. Many agents, safely.

    python3 shoot_b.py [w3-09]

Covers w3-07, w3-08, w3-09 and w3-11.

**w3-10 is deliberately absent.** It shows Claude Code on a phone, and a phone
cannot be driven from here — it has to be filmed off the panel by hand. Every
other Day 2 lecture is unblocked.

Docker appears only through the CLI. Docker Desktop's image and container lists
are his real client work and must never be on camera.
"""

from __future__ import annotations

import os
import sys
import time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import rig  # noqa: E402
import w3drive as W  # noqa: E402
from w3drive import S  # noqa: E402

CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
PROFILE = "/Users/Shared/projects/.shoot-chrome"
s: rig.Session = None


def beat(label, note=""):
    s.mark(label, note)


def hold(x):
    time.sleep(x)


def browse(url: str, wait: float = 10.0) -> None:
    W.run(f"'{CHROME}' --user-data-dir={PROFILE} --no-first-run "
          f"--no-default-browser-check --new-window '{url}' >/dev/null 2>&1 &",
          wait=wait)


# ---------------------------------------------------------------- w3-07
def w3_07() -> None:
    beat("w3-07/cost", "what approving everything has been costing us")
    W.run("claude --model sonnet", wait=8)
    W.wait_idle(stable=6, timeout=120, label="(launch)")
    W.turn(
        "Count how many times you have had to stop and ask me for permission "
        "while building this project so far, and what each of those pauses was "
        "actually protecting me from. Be honest about which ones were worth it.",
        timeout=600, label="(the cost of approving)")
    hold(6)

    beat("w3-07/native", "the sandbox that is already here")
    W.turn(
        "Explain the sandbox mode built into this version: what it stops, what it "
        "does NOT stop, and the specific case where it is not enough. Show me how "
        "to turn it on for this project.",
        timeout=600, label="(native sandbox)")
    hold(6)
    W.turn("/exit", timeout=60, stable=4, label="(exit)")
    hold(2)


# ---------------------------------------------------------------- w3-08
def w3_08() -> None:
    beat("w3-08/why", "a box of its own")
    W.run("claude --model sonnet", wait=8)
    W.wait_idle(stable=6, timeout=120, label="(launch)")
    W.turn(
        "I want to let an agent run without approving anything, and I want that "
        "to be safe rather than reckless. Build me a dev container for this "
        "project: a Dockerfile and a compose file that mount only this repo, with "
        "no access to the rest of my machine and no credentials passed in. "
        "Explain each line as you write it.",
        timeout=1200, label="(build the container)")
    hold(6)

    beat("w3-08/up", "bring it up")
    W.turn("/exit", timeout=60, stable=4, label="(exit)")
    W.run("docker compose build 2>&1 | tail -20", wait=90)
    hold(8)
    W.run("docker compose run --rm agent bash -lc 'pwd && ls -a && whoami'", wait=25)
    hold(10)

    beat("w3-08/prove", "and prove it cannot reach out")
    W.run("docker compose run --rm agent bash -lc "
          "'ls /Users 2>&1 | head -3; echo ---; ls ~ | head -5'", wait=20)
    hold(10)

    beat("w3-08/yolo", "now the dangerous flag is not dangerous")
    W.run("docker compose run --rm agent bash -lc 'echo would run: "
          "claude --dangerously-skip-permissions'", wait=8)
    hold(8)


# ---------------------------------------------------------------- w3-09
def w3_09() -> None:
    beat("w3-09/open", "the same agent, somewhere else")
    browse("https://claude.ai/code", wait=14)
    hold(18)

    beat("w3-09/task", "hand it work and walk away")
    hold(30)

    beat("w3-09/review", "come back to a branch")
    hold(25)
    W.focus_shoot_window()
    hold(3)


# ---------------------------------------------------------------- w3-11
def w3_11() -> None:
    beat("w3-11/issue", "describe the change as an issue")
    W.run("gh issue create --title 'Fleet list should sort late vehicles first' "
          "--body 'Late vehicles are hard to spot. Sort them to the top of the "
          "fleet list and mark them clearly.' 2>&1 | tail -3", wait=20)
    hold(8)

    beat("w3-11/tag", "hand it to the agent")
    W.run("gh issue list 2>&1 | head -10", wait=12)
    hold(10)

    beat("w3-11/pr", "and review what comes back")
    W.run("gh pr list 2>&1 | head -10", wait=12)
    hold(10)


BEATS = [("w3-07", w3_07), ("w3-08", w3_08), ("w3-09", w3_09), ("w3-11", w3_11)]


def main() -> None:
    global s
    start = sys.argv[1] if len(sys.argv) > 1 else None
    todo = BEATS
    if start:
        idx = [i for i, (n, _) in enumerate(BEATS) if n == start]
        if not idx:
            sys.exit(f"unknown lecture {start}")
        todo = BEATS[idx[0]:]

    s = rig.Session("B", allow=("Code", "Google Chrome"))
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
