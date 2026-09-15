#!/usr/bin/env python3
"""Keep recording L19 after the first take's quiet detector fires.

The quiet detector watches files inside the project, and Claude Code's browser
verification phase writes its scratch scripts to /private/tmp instead — so the
project looks idle for minutes while the most interesting part of the build is
still happening. Rather than loosen the detector and risk filming a genuinely
finished screen for four minutes, the first take stops on schedule and this one
picks the finale up.

Sends no input. Records, waits for the project to go quiet for a good while, and
stops. `SegRec` will not overwrite the earlier take's segments — see the
_seg_path fix from 2026-08-28.
"""
import os, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S
import approve as A

PROJ = "/Users/Shared/projects/kanban"

if __name__ == "__main__":
    S.wait_front("Code"); time.sleep(0.8); S.clear_mods()
    r = S.SegRec("L19_B_tail", allow=("Code",))
    r.start(settle=1.5)
    t0 = time.time()
    last, last_change, i = A.fs_state(PROJ), time.time(), 0
    try:
        while time.time() - t0 < 1800:
            S.move(1250 + (i % 4) * 150, 300 + (i % 5) * 110, 1.3)
            time.sleep(8.0); i += 1
            now = A.fs_state(PROJ)
            if now != last:
                last, last_change = now, time.time()
                print(f"[{int(time.time()-t0):5d}s] files={now[0]}", flush=True)
            elif time.time() - last_change > 300:
                print(f"[{int(time.time()-t0):5d}s] QUIET files={now[0]}", flush=True)
                break
        time.sleep(6)
    finally:
        r.stop()
        print("final fs:", A.fs_state(PROJ), flush=True)
