#!/usr/bin/env python3
"""L18 take B — approve the plan and record Copilot building.

Copilot asked "Ready to start Phase 1?" rather than just going, which is itself
a difference from Cursor worth keeping on camera.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Code",)
PROMPT = (1770, 971)

if __name__ == "__main__":
    S.wait_front("Code"); S.clear_mods()
    r = S.Rec("L18_B_build", guard=GUARD)
    r.start(settle=2.5)
    try:
        # read the tail of the plan, then approve
        S.move(1760, 700, 1.2); time.sleep(3.0)
        S.click(*PROMPT, 1.0); time.sleep(1.8)
        S.type_text("yes, go ahead and build it", cps=11); time.sleep(2.0)
        S.key('return'); time.sleep(6.0)
        for i in range(62):
            S.move(1690 + (i % 4) * 80, 360 + (i % 5) * 95, 1.2)
            time.sleep(10.0)
    finally:
        r.stop()
