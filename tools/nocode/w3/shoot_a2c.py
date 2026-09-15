#!/usr/bin/env python3
"""Pickup for w3-13: the game reveal.

The re-shoot was stopped seconds before Chrome opened, so the lecture ends on a
terminal instead of on the thing it built. This films only that payoff and
appends it to the same lecture:

    python3 postprod-style:  cut.py A2b,A2c w3-13
"""
from __future__ import annotations

import os
import sys
import time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import rig  # noqa: E402
import w3drive as W  # noqa: E402
from w3drive import S  # noqa: E402

SDK = "/Users/Shared/projects/sdk-demo"
CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
PROFILE = "/Users/Shared/projects/.shoot-chrome"


def chrome_clean(on: bool) -> None:
    """Hide the menu bar and dock for the shot, and put them back after.

    The first attempt at this pickup filmed the payoff with the dock and menu bar
    on screen, which nothing else in the week has. Teardown restores them, so a
    pickup taken after a teardown has to set them again itself.
    """
    import subprocess
    if on:
        subprocess.run(["defaults", "write", "NSGlobalDomain",
                        "_HIHideMenuBar", "-bool", "true"])
    else:
        subprocess.run(["defaults", "delete", "NSGlobalDomain", "_HIHideMenuBar"],
                       capture_output=True)
    subprocess.run(["osascript", "-e",
                    "tell application \"System Events\" to tell dock preferences "
                    f"to set autohide to {str(on).lower()}"], capture_output=True)
    time.sleep(2)


def main() -> None:
    chrome_clean(True)
    s = rig.Session("A2d", allow=("Code", "Google Chrome"))
    s.start(need_gb=8)
    try:
        s.mark("w3-13/files", "what landed in the folder")
        W.run(f"cd {SDK} && ls -la", wait=4)
        time.sleep(6)

        s.mark("w3-13/game", "open what it made")
        W.run(f"'{CHROME}' --user-data-dir={PROFILE} --no-first-run "
              f"--no-default-browser-check --window-position=0,0 "
              f"--window-size=1920,1080 --app=file://{SDK}/index.html "
              f">/dev/null 2>&1 &", wait=12)
        time.sleep(34)          # let the game sit on screen
        s.mark("w3-13/close", "and that is the point")
        W.focus_shoot_window()
        time.sleep(6)
    except Exception as exc:
        print(f"!! {exc!r}", flush=True)
    finally:
        s.stop()
        chrome_clean(False)


if __name__ == "__main__":
    main()
