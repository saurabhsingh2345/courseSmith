#!/usr/bin/env python3
"""L17 take B — read the plan briefly, press Build, and record the build.

The build is the lecture. Recorded long and sped up in post; his version runs
about six minutes of agent work and narrates over it.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Cursor",)
BUILD = (1836, 680)

if __name__ == "__main__":
    S.wait_front("Cursor"); S.clear_mods()
    r = S.Rec("L17_B_build", guard=GUARD)
    r.start(settle=2.5)
    try:
        # read down the plan document a little
        S.move(760, 500, 1.0); time.sleep(2.5)
        for _ in range(5):
            S.scroll(-200); time.sleep(2.2)
        time.sleep(2.0)
        S.scroll(900); time.sleep(2.0)
        # then press Build
        S.move(*BUILD, 1.2); time.sleep(2.0)
        S.click(*BUILD, 0.3); time.sleep(5.0)
        # and watch it work
        for i in range(58):
            S.move(1450 + (i % 4) * 110, 340 + (i % 5) * 90, 1.2)
            time.sleep(10.0)
    finally:
        r.stop()
