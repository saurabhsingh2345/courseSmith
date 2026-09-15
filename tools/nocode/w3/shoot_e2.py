#!/usr/bin/env python3
"""CAPTURE E2 - actually using the console, after the fix pass.

    python3 shoot_e2.py

w3-31's first look filmed Signal Desk running and moving, which is the important
half, but two of its three interactions missed: the rules panel's controls sit
higher than the guess, and the replay scrubber is below the chart rather than
inside it. Coordinates here were read off a frame of the running app.

These beats belong to w3-32, not w3-31, and that turned out to be the better
structure anyway: w3-31 is the raw first run and its defects, and w3-32 is the
console being properly used once they are fixed. Cut with `cut.py E3,E2 w3-32`.

Points are on the 1920x1080 panel. The app is `--app` Chrome at 0,0, so the
page's own origin is the screen's.
"""

from __future__ import annotations

import os
import subprocess
import sys
import time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import w3drive as W  # noqa: E402
from w3drive import S  # noqa: E402
import w3kit as K  # noqa: E402
from w3kit import beat, hold  # noqa: E402

W.SHOOT_TITLE = "signal-desk"
REPO = "/Users/Shared/projects/signal-desk"

# Read off a frame of the running console, not guessed.
ROW = lambda n: (200, 92 + 45 * n)          # noqa: E731  watchlist rows
SYMBOL = (1633, 101)
KIND = (1732, 141)
THRESHOLD = (1736, 180)
ADD = (1626, 220)
SCRUB_L = (400, 818)
SCRUB_M = (900, 818)
SCRUB_R = (1430, 818)
THEME = (1837, 28)
ALERT = (700, 940)


def serve() -> None:
    """Make sure the console is up before Chrome is pointed at it."""
    W.run("curl -s -o /dev/null -w 'up %{http_code}\\n' http://localhost:8081/ "
          "|| true", wait=4)
    if subprocess.run(["curl", "-s", "-o", "/dev/null",
                       "http://localhost:8081/"]).returncode != 0:
        W.run("(npm start >/tmp/sd3.log 2>&1 &) ; sleep 6 ; "
              "curl -s -o /dev/null -w 'up %{http_code}\\n' http://localhost:8081/",
              wait=14)


def w3_32b() -> None:
    beat("w3-32/using", "back to the console, this time to use it")
    serve()
    hold(4)
    K.browser("http://localhost:8081", wait=14)
    hold(18)

    beat("w3-32/sort", "the list is re-ordering itself")
    for i in range(6):
        S.move(200, 150 + i * 90, 1.2)
        hold(4.0)

    beat("w3-32/pick", "pick an instrument and watch the chart follow")
    S.click(*ROW(2), 1.0)
    hold(8)
    S.click(*ROW(7), 1.0)
    hold(8)

    beat("w3-32/rule", "write a rule in the sentence it understands")
    S.click(*SYMBOL, 1.0)
    hold(1.5)
    S.type_text("ALP", cps=6)
    hold(2.0)
    S.click(*THRESHOLD, 1.0)
    hold(1.2)
    S.type_text("400", cps=6)
    hold(2.0)
    S.click(*ADD, 1.0)
    hold(10)

    beat("w3-32/second", "and a second one, of a different kind")
    S.click(*SYMBOL, 1.0)
    hold(1.2)
    S.type_text("KLN", cps=6)
    hold(1.5)
    S.click(*KIND, 1.0)
    hold(1.5)
    S.key("down")
    hold(1.0)
    S.key("return")
    hold(1.5)
    S.click(*THRESHOLD, 1.0)
    hold(1.0)
    S.type_text("70", cps=6)
    hold(1.5)
    S.click(*ADD, 1.0)
    hold(12)

    beat("w3-32/wait", "now leave it alone and see whether anything fires")
    for i in range(10):
        S.move(760 + (i % 4) * 130, 300 + (i % 3) * 110, 1.3)
        hold(6.0)

    beat("w3-32/alert", "an alert, and what clicking one does")
    S.click(*ALERT, 1.0)
    hold(10)

    beat("w3-32/replay", "and rewind the whole console")
    S.move(*SCRUB_R, 1.2)
    hold(2)
    S.drag(SCRUB_R[0], SCRUB_R[1], SCRUB_M[0], SCRUB_M[1], glide=2.6)
    hold(10)
    S.drag(SCRUB_M[0], SCRUB_M[1], SCRUB_L[0], SCRUB_L[1], glide=2.2)
    hold(10)
    S.drag(SCRUB_L[0], SCRUB_L[1], SCRUB_R[0], SCRUB_R[1], glide=2.6)
    hold(8)

    beat("w3-32/theme", "and the light one, because the brief asked for both")
    S.click(*THEME, 1.0)
    hold(9)
    S.click(*THEME, 1.0)
    hold(6)

    beat("w3-32/honest", "and now the things that are wrong with it")
    for i in range(6):
        S.move(500 + (i % 3) * 220, 250 + (i % 4) * 120, 1.4)
        hold(5.0)
    K.close_browser()
    hold(4)


BEATS = [("w3-32", w3_32b)]

if __name__ == "__main__":
    K.run_session(BEATS, "E2", REPO)
