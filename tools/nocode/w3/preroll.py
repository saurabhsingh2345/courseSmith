#!/usr/bin/env python3
"""Put the shoot window into a known-good state, then assert it. Run before every take.

Each check here exists because it already cost a take:
  * a stale editor tab from a previous attempt sat in frame for five lectures
  * a second terminal tab showed in the panel header
  * the window drifted off the panel and his desktop files were in shot
  * the shell prompt printed his username and hostname
  * fullscreen moved VS Code to its own Space and keystrokes went elsewhere
"""

from __future__ import annotations

import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import w3drive as W  # noqa: E402
from w3drive import S  # noqa: E402

REPO = "/Users/Shared/projects/control-tower"


def main(expect_empty: bool = True) -> None:
    W.focus_shoot_window()

    # 1. no editors open at all — a stale tab is the most visible kind of mess
    W.palette("View: Close All Editors", settle=2.0)

    # 2. exactly one terminal, with a prompt that names no human
    W.one_terminal()

    # 3. the window covers the panel and is NOT fullscreen (fullscreen moves it
    #    to its own Space, which is how keystrokes ended up in another app)
    S.osa('tell application "System Events" to tell process "Code"\n'
          '  set position of window 1 to {0, 0}\n'
          '  set size of window 1 to {1920, 1080}\n'
          'end tell')

    W.require_shoot_window()

    # 4. the repo is in its intended starting state
    here = sorted(os.listdir(REPO))
    if expect_empty and here:
        print(f"  !! repo is not empty: {here}")
        print("     rm -rf its contents before rolling, or pass --dirty")
        sys.exit(1)

    # 5. nothing else is capturing
    if subprocess.run(["pgrep", "-f", "avfoundation"],
                      capture_output=True).returncode == 0:
        sys.exit("  !! a capture is already running — find the session, do not kill it")

    S.snap("/tmp/_w3_preroll.png", disp="1")
    print("preroll ok — window clean, one terminal, repo",
          "empty" if expect_empty else f"({len(here)} entries)")
    print("  frame saved to /tmp/_w3_preroll.png — look at it before rolling")


if __name__ == "__main__":
    main(expect_empty="--dirty" not in sys.argv)
