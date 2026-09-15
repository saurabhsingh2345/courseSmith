#!/usr/bin/env python3
"""L16 take E, second attempt.

First attempt overscrolled: agents.md is 69 lines and I scrolled 3200px, so
half the take was an empty editor. The viewport shows ~40 lines, so the whole
traverse is roughly 650px. Source only — take D already covers the preview, and
the source column is wider and more readable without it.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Cursor",)


def reset():
    S.wait_front("Cursor"); S.clear_mods()
    S.key('p', ('cmd', 'shift')); time.sleep(1.6)
    S.type_text("View: Close All Editors", cps=20); time.sleep(1.8)
    S.key('return'); time.sleep(2.5)
    S.click(67, 103, 1.0); time.sleep(3.0)      # agents.md, alone in the tab bar
    S.move(700, 500, 0.8); time.sleep(0.8)
    for _ in range(6):
        S.scroll(400); time.sleep(0.2)          # ensure top
    time.sleep(1.5)
    return S.snap(os.path.join(os.path.dirname(os.path.abspath(__file__)), "pre_walk.png"))


def walk():
    holds = [16, 15, 15, 14, 15, 14, 16, 16, 15]   # seconds per section
    for i, hold in enumerate(holds):
        if i:
            S.scroll(-85); time.sleep(1.1)          # ~4 lines at a time
        S.move(620, 300 + (i % 3) * 90, 0.9)
        time.sleep(hold)


if __name__ == "__main__":
    print("pre-roll snapshot:", reset())
    r = S.Rec("L16_E_walk", guard=GUARD)
    r.start(settle=2.0)
    try:
        walk()
    finally:
        r.stop()
