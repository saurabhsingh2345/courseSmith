#!/usr/bin/env python3
"""L19 prep — VS Code on the fresh clone, with the Claude Code panel forward.

Same rules as prep_l18_v2: a NEW window via `code -n`, confirmed by title before
anything is moved, and no app hiding and no quitting. His windows stay put.

The one L19-specific job is getting the right tab forward. The secondary sidebar
carries three tabs on this machine — CHAT, CODEX and CLAUDE CODE — and which one
is showing is whatever was showing last. The tab is found by pixel rather than by
a remembered coordinate, because the tab strip's layout moves with the panel
width, and the panel width is something this script changes.
"""
import os, sys, time, subprocess
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S
from PIL import Image

PROJ = "/Users/Shared/projects/kanban"
OUT = os.path.dirname(os.path.abspath(__file__))
CODE = "/usr/local/bin/code" if os.path.exists("/usr/local/bin/code") else "code"


def shot(path="/tmp/_l19probe.png"):
    subprocess.run(["screencapture", "-x", "-D", "1", path], capture_output=True)
    subprocess.run(["sips", "-Z", "1920", path, "-o", path], capture_output=True)
    return Image.open(path).convert("RGB")


def panel_edge(im, y=600):
    """x of the secondary sidebar's left edge, found by the divider highlight."""
    px = im.load()
    prev = None
    edge = None
    for x in range(700, 1900):
        c = px[x, y]
        if prev and sum(abs(a - b) for a, b in zip(c, prev)) > 10 and c[0] < 30:
            edge = x
        prev = c
    return edge


def main():
    S.desktop_quiet(True)
    S.close_finder_windows()
    hidden = S.osa('tell application "System Events" to get name of every process '
                   'whose visible is false and background only is false')
    print("hidden before (must be empty):", repr(hidden))

    subprocess.Popen([CODE, "--disable-workspace-trust", "-n", PROJ])
    time.sleep(18)
    S.wait_front("Code")
    time.sleep(2.5)
    S.clear_mods()

    title = S.osa('tell application "System Events" to tell process "Code" '
                  'to get name of front window')
    print("front window:", repr(title))
    if "kanban" not in title.lower():
        raise SystemExit(f"REFUSING to touch {title!r} — not our shoot window")

    S.move_window_to_panel("Code")
    time.sleep(2.5)
    S.move(900, 600, 0.5)
    time.sleep(1.5)
    print("edge before widen:", panel_edge(shot()))
    print("snapshot:", S.snap(os.path.join(OUT, "l19_pre.png")))


if __name__ == "__main__":
    main()
