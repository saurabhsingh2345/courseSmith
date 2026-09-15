#!/usr/bin/env python3
"""L16 — the long agents.md walk, third attempt.

Attempt one overscrolled (3200px on a 69-line file) and half the take was an
empty editor. agents.md is 69 lines, the viewport shows ~40, so the entire
traverse is about 650px. Nine holds, ~75px apart.
"""
import os, sys, time, subprocess
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Cursor",)
HERE = os.path.dirname(os.path.abspath(__file__))


def setup():
    subprocess.run(["osascript", "-e", 'tell application "Cursor" to quit'],
                   capture_output=True)
    time.sleep(5)
    subprocess.Popen(["open", "-a", "Cursor", "/Users/Shared/projects/kanban"])
    time.sleep(14)
    S.wait_front("Cursor"); time.sleep(2)
    S.key('return'); time.sleep(2)
    S.clear_mods()
    S.move_window_to_panel("Cursor")
    time.sleep(2)
    S.click(67, 103, 1.0); time.sleep(3.0)     # agents.md
    S.move(700, 400, 0.7); time.sleep(1.0)
    for _ in range(6):
        S.scroll(400); time.sleep(0.2)         # ensure top
    time.sleep(1.5)
    return S.snap(os.path.join(HERE, "pre_walk2.png"))


def walk():
    holds = [15, 14, 15, 13, 14, 13, 15, 15, 14]
    for i, hold in enumerate(holds):
        if i:
            S.scroll(-75); time.sleep(1.2)
        S.move(640, 260 + (i % 3) * 110, 1.0)
        time.sleep(hold)


if __name__ == "__main__":
    print("pre-roll:", setup(), flush=True)
    r = S.Rec("L16_E_walk", guard=GUARD)
    r.start(settle=2.0)
    try:
        walk()
    finally:
        r.stop()
