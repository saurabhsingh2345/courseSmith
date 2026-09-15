#!/usr/bin/env python3
"""L18 continuation: keep recording the build that is ALREADY running.

The first script was going to type a second prompt ("go ahead and build it")
after the planning wait. It turned out not to be needed — Copilot planned and
then started implementing off the single `go ahead and plan`, exactly as the
brief's Strategy section tells it to. A second prompt mid-build would have
queued a redundant turn and muddied the story, so the first script was
interrupted before it fired and this one takes over the recording.

No input is sent to the editor at all here. It only records and waits for the
filesystem to go quiet.
"""
import os, sys, time, subprocess
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

PROJ = "/Users/Shared/projects/kanban"


def fs_state():
    n, newest = 0, 0.0
    for root, dirs, files in os.walk(PROJ):
        dirs[:] = [d for d in dirs if d not in (".git", "node_modules", ".next")]
        for f in files:
            try:
                m = os.path.getmtime(os.path.join(root, f))
            except OSError:
                continue
            n += 1
            newest = max(newest, m)
    return n, newest


if __name__ == "__main__":
    S.wait_front("Code"); time.sleep(0.8); S.clear_mods()
    r = S.SegRec("L18_Q_build", allow=("Code",))
    r.start(settle=1.5)
    t0 = time.time()
    last, last_change, i = fs_state(), time.time(), 0
    try:
        while time.time() - t0 < 2700:
            S.move(1250 + (i % 4) * 150, 300 + (i % 5) * 110, 1.3)
            time.sleep(8.0)
            i += 1
            now = fs_state()
            if now != last:
                last, last_change = now, time.time()
                print(f"[{int(time.time()-t0):5d}s] files={now[0]}", flush=True)
            elif time.time() - last_change > 120:
                print(f"[{int(time.time()-t0):5d}s] QUIET, files={now[0]}", flush=True)
                break
        time.sleep(6)
    finally:
        r.stop()
        print("final fs:", fs_state(), flush=True)
