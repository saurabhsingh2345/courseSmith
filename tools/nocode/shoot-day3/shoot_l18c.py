#!/usr/bin/env python3
"""L18 take C — the build, with auto-approve on.

First attempt stalled 11 minutes on `mkdir` waiting for per-command permission,
which is a real difference from Cursor and stays in the lecture. This take is
the build itself, with `chat.tools.autoApprove` set so it can run unattended.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Code",)
PROMPT = (1770, 971)

if __name__ == "__main__":
    S.wait_front("Code"); time.sleep(1.5); S.clear_mods()
    r = S.Rec("L18_C_build", guard=GUARD)
    r.start(settle=2.5)
    try:
        S.click(*PROMPT, 1.0); time.sleep(2.0)
        S.type_text("go ahead and build it", cps=10); time.sleep(2.0)
        S.key('return'); time.sleep(6.0)
        for i in range(80):
            S.move(1690 + (i % 4) * 80, 360 + (i % 5) * 95, 1.2)
            time.sleep(10.0)
    finally:
        r.stop()
