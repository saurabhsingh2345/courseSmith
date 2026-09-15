#!/usr/bin/env python3
"""Record the Antigravity build that is already running. Sends no input.

The first L20 attempt never captured a frame: its guard allowed only "Electron"
(the AppleScript process name) while SegRec polls NSWorkspace, which reports
"Antigravity IDE". By the time that was found, the plan had been approved and the
build was under way — so this records from here, and the plan beat is picked up
afterwards by scrolling the conversation back (it persists) in a separate take.
"""
import os, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S
import approve as A

PROJ = "/Users/Shared/projects/kanban"
ALLOW = ("Antigravity IDE", "Electron")

if __name__ == "__main__":
    print("frontmost(fast):", S.frontmost_fast(), flush=True)
    r = S.SegRec("L20_A_build", allow=ALLOW)
    r.start(settle=2.0)
    t0 = time.time()
    last, last_change, i = A.fs_state(PROJ), time.time(), 0
    try:
        while time.time() - t0 < 3200:
            S.move(1250 + (i % 4) * 150, 300 + (i % 5) * 110, 1.3)
            time.sleep(8.0); i += 1
            now = A.fs_state(PROJ)
            if now != last:
                last, last_change = now, time.time()
                print(f"[{int(time.time()-t0):5d}s] files={now[0]}", flush=True)
            elif now[0] > 15 and time.time() - last_change > 300:
                print(f"[{int(time.time()-t0):5d}s] BUILD QUIET files={now[0]}", flush=True)
                break
        time.sleep(6)
    finally:
        r.stop()
        print("final fs:", A.fs_state(PROJ), flush=True)
