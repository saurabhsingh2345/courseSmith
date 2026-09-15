#!/usr/bin/env python3
"""L30 prep — VS Code on the PARENT projects folder, so the terminal can clone.

Two things the old prep_l18.py did are gone, and they were the whole complaint:

  1. `set visible of every process whose name is not "Finder" to false` — hid
     every one of his apps, which reads exactly as "you closed my VS Code".
  2. `tell application "Visual Studio Code" to quit` — closed the window he was
     working in, purely to defeat session restore. `code -n` already opens a
     NEW window, so the quit was never needed.

What replaces them: open a new window on the neutral folder, confirm by TITLE
that the window we are about to move is ours and not his, and take only that
window fullscreen on the shoot panel. His windows stay open, visible, and on
the other display, out of frame.
"""
import os, sys, time, subprocess
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

PROJ = "/Users/Shared/projects"
OUT = os.path.dirname(os.path.abspath(__file__))
CODE = "/usr/local/bin/code" if os.path.exists("/usr/local/bin/code") else "code"


def front_window_title(app="Code"):
    return S.osa(f'tell application "System Events" to tell process "{app}" '
                 'to get name of front window')


def main():
    # Desktop icons + dock only. No app hiding.
    S.desktop_quiet(True)
    S.close_finder_windows()

    before = S.osa('tell application "System Events" to get name of every process '
                   'whose visible is false and background only is false')
    print("hidden before (should be empty):", repr(before))

    subprocess.Popen([CODE, "--disable-workspace-trust", "-n", PROJ])
    time.sleep(18)
    S.wait_front("Code")
    time.sleep(2.5)
    S.clear_mods()

    title = front_window_title()
    print("front window title:", repr(title))
    if "projects" not in title.lower():
        # Never fullscreen or resize a window that is not the one we opened.
        raise SystemExit(f"REFUSING to touch window {title!r} — not our shoot window. "
                         "His window is frontmost; nothing was changed.")

    S.move_window_to_panel("Code")
    time.sleep(2)
    S.move(960, 620, 0.4)
    time.sleep(1.5)

    print("frontmost:", S.frontmost())
    print("hidden after (must still be empty):",
          repr(S.osa('tell application "System Events" to get name of every process '
                     'whose visible is false and background only is false')))
    print("snapshot:", S.snap(os.path.join(OUT, "l30_prep.png")))


if __name__ == "__main__":
    main()
