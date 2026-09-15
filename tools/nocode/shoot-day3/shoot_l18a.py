#!/usr/bin/env python3
"""L18 take A — the same four-word prompt, in Copilot's agent mode.

Same brief, same prompt, same five checks as L17. The only thing that changes
between the four builds is the harness.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Code",)
PROMPT = (1770, 971)

if __name__ == "__main__":
    S.wait_front("Code"); S.clear_mods()
    r = S.Rec("L18_A_prompt", guard=GUARD)
    r.start(settle=2.5)
    try:
        # the tree: one file, same as every other build starts with
        S.move(110, 68, 1.2); time.sleep(2.5)
        S.move(1770, 1009, 1.2); time.sleep(2.5)     # Agent mode / model row
        S.click(*PROMPT, 1.0); time.sleep(2.0)
        S.type_text("go ahead and plan", cps=9); time.sleep(2.5)
        S.key('return'); time.sleep(5.0)
        for i in range(16):
            S.move(1700 + (i % 3) * 70, 400 + (i % 4) * 90, 1.2)
            time.sleep(8.0)
    finally:
        r.stop()
