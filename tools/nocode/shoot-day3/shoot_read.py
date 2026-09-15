#!/usr/bin/env python3
"""Record a long, continuous read-through of real project files.

Why this exists: the first cut of Day 5 stabbed into footage 21 times for two or
three seconds at a stretch, which reads as a glitch rather than as work. The
source course does these lectures as 100% screen recording — one take, narrated
over. This produces that: one continuous segment per lecture, long enough that
the cut never has to reach for a three-second clip.

Files are opened from a git WORKTREE of the real build, so every frame is the
actual project at the actual commit, and the live repo is never touched.

Usage: shoot_read.py <take_name> <worktree> <file:dwell> [file:dwell ...]
"""
import os, subprocess, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S

CODE = "/usr/local/bin/code" if os.path.exists("/usr/local/bin/code") else "code"


def open_file(root, rel):
    subprocess.run([CODE, "-r", os.path.join(root, rel)], capture_output=True)
    time.sleep(2.2)


def read_through(seconds, page_every=4.0):
    """Scroll steadily for `seconds`. Steady beats jumpy: the viewer is reading
    along with the narration, and a page that leaps is a page nobody read."""
    t0 = time.time()
    while time.time() - t0 < seconds:
        S.scroll(-3, steps=6)
        time.sleep(page_every)


if __name__ == "__main__":
    take, root, specs = sys.argv[1], sys.argv[2], sys.argv[3:]
    # The first attempt at this recorded a SMALL window sitting on the desktop,
    # and the desktop behind it was covered in his personal folders and a wall of
    # screenshots. Take deleted. Three things have to be true before rolling:
    # the desktop is empty, the window is fullscreen on the shoot panel, and the
    # frame has been looked at.
    S.desktop_quiet(True)
    subprocess.run([CODE, "-n", root], capture_output=True)
    time.sleep(7)
    want = os.path.basename(root.rstrip("/"))
    S.wait_front_window("Code", want)
    S.clear_mods()
    S.move_window_to_panel("Code"); time.sleep(2.5)
    S.key('b', ('cmd',)); time.sleep(1.0)      # sidebar away, code fills the frame
    S.key('j', ('cmd',)); time.sleep(0.8)      # bottom panel away
    S.key('b', ('cmd', 'alt')); time.sleep(0.8) # and the chat sidebar on the right
    S.snap(os.path.join(HERE, f"_frame_{take}.png"))
    print("frame written — check it before trusting this take", flush=True)

    r = S.SegRec(take, allow=("Code",))
    r.start(settle=3.0)
    try:
        for spec in specs:
            rel, dwell = spec.rsplit(":", 1)
            open_file(root, rel)
            print(f"  {rel}  ({dwell}s)", flush=True)
            time.sleep(2.0)
            read_through(float(dwell))
        time.sleep(4)
    finally:
        r.stop()
