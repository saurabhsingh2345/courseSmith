#!/usr/bin/env python3
"""L18 take — Copilot plans and builds the Kanban app, for real, in one take.

Why this replaces L18_A/L18_B: those two takes are 15 minutes of a near-empty
editor with the VS Code watermark filling the frame, and L18_B sat eleven of its
twelve minutes waiting for per-command permission that never came. Nothing was
built — `kanban-copilot/` contains only agents.md and a readme.

What is different here:
  * agents.md is OPEN, so the middle of the frame carries the brief instead of a
    logo, and the chat panel is dragged out to 800pt so the agent is legible.
  * `chat.tools.autoApprove` is already true, so the build actually runs.
  * SegRec pauses rather than kills when he switches app, so a glance at another
    window costs a cut and not the whole twenty-minute build.
  * The wait ends when the FILESYSTEM goes quiet, not on a guessed timer.

The prompt is `go ahead and plan`, character for character what L17 typed into
Cursor, because the lecture's claim is that only the harness changes.
"""
import os, sys, time, subprocess
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

PROJ = "/Users/Shared/projects/kanban"
PROMPT_BOX = (1400, 971)
ALLOW = ("Code",)


def fs_state():
    """(file count, newest mtime) over the project, ignoring git and modules."""
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


def jiggle(seconds, step=9.0):
    """Keep the cursor alive so the frame is not a freeze-frame, and report."""
    t0 = time.time()
    i = 0
    while time.time() - t0 < seconds:
        S.move(1250 + (i % 4) * 150, 300 + (i % 5) * 110, 1.3)
        time.sleep(step)
        i += 1


def wait_until_quiet(quiet=100.0, cap=2100.0):
    """Return when the project stops changing for `quiet` seconds."""
    last = fs_state()
    last_change = time.time()
    t0 = time.time()
    i = 0
    while time.time() - t0 < cap:
        S.move(1250 + (i % 4) * 150, 300 + (i % 5) * 110, 1.3)
        time.sleep(8.0)
        i += 1
        now = fs_state()
        if now != last:
            last, last_change = now, time.time()
            print(f"[{int(time.time()-t0):5d}s] files={now[0]}", flush=True)
        elif time.time() - last_change > quiet:
            print(f"[{int(time.time()-t0):5d}s] quiet {quiet:.0f}s, files={now[0]}",
                  flush=True)
            return True
    print("cap reached", flush=True)
    return False


if __name__ == "__main__":
    S.wait_front("Code"); time.sleep(1.2); S.clear_mods()
    print("start fs:", fs_state(), flush=True)

    r = S.SegRec("L18_P_build", allow=ALLOW)
    r.start(settle=3.0)
    try:
        S.click(*PROMPT_BOX, 1.1); time.sleep(1.6)
        S.type_text("go ahead and plan", cps=9); time.sleep(2.2)
        S.key('return')
        print("prompted, planning...", flush=True)
        jiggle(170)

        S.click(*PROMPT_BOX, 1.0); time.sleep(1.4)
        S.type_text("go ahead and build it", cps=9); time.sleep(2.0)
        S.key('return')
        print("prompted build, waiting on filesystem...", flush=True)
        time.sleep(20)
        wait_until_quiet()
        jiggle(25)
    finally:
        segs = r.stop()
        print("final fs:", fs_state(), flush=True)
