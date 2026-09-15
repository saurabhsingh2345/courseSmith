#!/usr/bin/env python3
"""L17 take A — the plan.

Plan mode is already selected (amber pill). Type only "go ahead and plan" —
deliberately saying nothing about WHAT, because agents.md is loaded into
context by default and that is the whole point of the beat.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Cursor",)
PROMPT = (1700, 100)

if __name__ == "__main__":
    S.wait_front("Cursor"); S.clear_mods()
    r = S.Rec("L17_A_plan", guard=GUARD)
    r.start(settle=2.5)
    try:
        S.click(*PROMPT, 1.0); time.sleep(2.0)      # click INTO the box first
        S.type_text("go ahead and plan", cps=9); time.sleep(2.5)
        S.key('return'); time.sleep(4.0)
        # let the plan generate, drifting the cursor so the frame is not dead
        for i in range(14):
            S.move(1500 + (i % 3) * 90, 380 + (i % 4) * 80, 1.1)
            time.sleep(7.0)
        # then read down it
        S.move(1650, 500, 0.8); time.sleep(1.5)
        for _ in range(6):
            S.scroll(-190); time.sleep(2.4)
        time.sleep(3.0)
    finally:
        r.stop()
