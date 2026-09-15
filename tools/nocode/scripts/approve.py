#!/usr/bin/env python3
"""Find and click the agent's "yes" button, by colour, during a long build take.

Extracted from shoot_l18_run.py once it was clear all three remaining Day 3
builds need it. The problem it solves: an agentic build stalls on a permission
prompt, and an unattended take then films a frozen screen instead of a build.
On L18 that cost eleven minutes of the original take and produced nothing.

Two things this gets right that a naive version does not:

  * It finds blue RUNS row by row and merges vertically adjacent ones that
    overlap in x, rather than taking one bounding box over every blue pixel.
    Once files have been written there are TWO accent buttons on screen at once
    (the pending Allow, and a Keep on the changed-files row far to the right);
    their union is ~690pt wide, gets rejected as implausible, and the build then
    sits blocked forever.
  * The merge tolerance is 26pt, not a few pixels, because the button's own
    white LABEL splits the blue into a run above the text and a run below it.
    At 6pt those two fragments never join and nothing ever matches.

The click lands 30% across the button so it can never hit a dropdown chevron on
the right edge, which opens a menu instead of approving.
"""
import os, subprocess, time
from PIL import Image

# VS Code's primary-button blue on this machine's Dark Modern theme. Antigravity
# and Cursor are both VS Code forks and theme buttons from the same token, so the
# same value is the right starting guess for all three.
BTN = (16, 120, 207)
TOL = 8

# Not every approval button is that blue. The panel used for the Day 5 build
# renders its "Allow" in a desaturated teal instead, and a matcher that knew only
# the one colour sat looking at a pending confirmation for four minutes without
# clicking it. Match against a LIST, and give the second one a wider tolerance —
# it was measured off a JPEG still, so it carries compression error.
BTNS = [(BTN, TOL), ((45, 119, 161), 16)]


def _hit(px, x, y, cols):
    r, g, b = px[x, y]
    for (cr, cg, cb), t in cols:
        if abs(r-cr) <= t and abs(g-cg) <= t and abs(b-cb) <= t:
            return True
    return False


def snap_points(path="/tmp/_approve_scan.png"):
    """Screenshot of the shoot panel, normalised so 1 image pixel == 1 point."""
    subprocess.run(["screencapture", "-x", "-D", "1", path], capture_output=True)
    subprocess.run(["sips", "-Z", "1920", path, "-o", path], capture_output=True)
    return Image.open(path).convert("RGB")


def find_button(im, panel_x0=1130, btn=None, tol=None,
                min_w=46, max_w=320, min_h=6, max_h=60):
    """Topmost accent button to the right of panel_x0, as (x0, y0, x1, y1)."""
    cols = BTNS if btn is None else [(btn, TOL if tol is None else tol)]
    w, h = im.size
    px = im.load()
    runs = []
    for y in range(60, h - 30, 2):
        x = panel_x0
        while x < w - 4:
            if _hit(px, x, y, cols):
                x0 = x
                while x < w - 4:
                    if not _hit(px, x, y, cols):
                        break
                    x += 2
                if x - x0 >= 40:
                    runs.append((y, x0, x))
            else:
                x += 2

    boxes = []
    for y, x0, x1 in runs:
        for bx in boxes:
            if y - bx[3] <= 26 and not (x1 < bx[0] - 6 or x0 > bx[2] + 6):
                bx[0] = min(bx[0], x0); bx[2] = max(bx[2], x1); bx[3] = y
                break
        else:
            boxes.append([x0, y, x1, y])

    cands = [b for b in boxes
             if min_w <= b[2]-b[0] <= max_w and min_h <= b[3]-b[1] <= max_h]
    if not cands:
        return None
    cands.sort(key=lambda b: b[1])
    return tuple(cands[0])


def click_point(box, frac=0.30):
    x0, y0, x1, y1 = box
    return int(x0 + (x1 - x0) * frac), int((y0 + y1) / 2)


def fs_state(root, skip=(".git", "node_modules", ".next", ".agent")):
    """(file count, newest mtime) — the honest signal that a build is working."""
    n, newest = 0, 0.0
    for dirpath, dirs, files in os.walk(root):
        dirs[:] = [d for d in dirs if d not in skip]
        for f in files:
            try:
                m = os.path.getmtime(os.path.join(dirpath, f))
            except OSError:
                continue
            n += 1
            newest = max(newest, m)
    return n, newest


def run_build(stage, rec_name, proj, allow, prompt_box, prompt_text,
              out_dir, quiet=240.0, cap=3000.0, min_files=15, cps=9,
              approve=True):
    """Roll, type the prompt, approve whatever it asks, stop when the project
    stops changing. Returns (segments, approvals)."""
    S = stage
    S.wait_front(allow[0]); time.sleep(1.2); S.clear_mods()
    print("start fs:", fs_state(proj), flush=True)

    r = S.SegRec(rec_name, allow=allow)
    r.start(settle=3.0)
    clicks = 0
    try:
        # An empty prompt means RESUME: the agent is already mid-run (it stalled
        # on a confirmation this clicker could not see, which is how the teal
        # button in BTNS came to be measured). Roll the camera and take over the
        # approvals without saying anything, so the rest of the part is filmed
        # instead of being lost between takes.
        if prompt_text:
            S.click(*prompt_box, 1.1); time.sleep(1.6)
            S.type_text(prompt_text, cps=cps); time.sleep(2.2)
            S.key('return')
            print(f"prompted: {prompt_text!r}", flush=True)
        else:
            print("RESUME: no prompt typed, approving only", flush=True)

        t0 = time.time()
        last, last_change, i = fs_state(proj), time.time(), 0
        while time.time() - t0 < cap:
            # Claude Code's Auto mode never asks, so the clicker is switched
            # off there rather than left hunting for accent-coloured buttons it
            # might mistake for something else.
            box = find_button(snap_points()) if approve else None
            if box:
                cx, cy = click_point(box)
                clicks += 1
                subprocess.run(["cp", "/tmp/_approve_scan.png",
                                os.path.join(out_dir, f"{rec_name}_ok{clicks:02d}.png")],
                               capture_output=True)
                print(f"[{int(time.time()-t0):5d}s] APPROVE #{clicks} {box} -> ({cx},{cy})",
                      flush=True)
                S.click(cx, cy, 0.8)
                time.sleep(7.0)
            else:
                S.move(1250 + (i % 4)*150, 300 + (i % 5)*110, 1.2)
                time.sleep(7.0); i += 1

            now = fs_state(proj)
            if now != last:
                last, last_change = now, time.time()
                print(f"[{int(time.time()-t0):5d}s] files={now[0]}", flush=True)
            elif now[0] > min_files and time.time() - last_change > quiet:
                print(f"[{int(time.time()-t0):5d}s] BUILD QUIET files={now[0]}", flush=True)
                break
        time.sleep(5)
    finally:
        segs = r.stop()
        print(f"approvals: {clicks}", flush=True)
        print("final fs:", fs_state(proj), flush=True)
    return segs, clicks
