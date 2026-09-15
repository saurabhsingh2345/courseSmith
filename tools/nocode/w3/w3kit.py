#!/usr/bin/env python3
"""Shared beats for the Week 3 top-up captures (E onwards).

shoot_a1 through shoot_d each carry their own copy of these helpers, which was
fine for four scripts and is not fine for nine. Everything here was extracted
unchanged from those scripts, including the comments that explain why each one
exists - every single one of them cost a take.
"""

from __future__ import annotations

import os
import subprocess
import sys
import time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import rig  # noqa: E402
import w3drive as W  # noqa: E402
from w3drive import S  # noqa: E402

CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
PROFILE = "/Users/Shared/projects/.shoot-chrome"

# Every prompt that builds something carries this. turn() answers ANY card with
# option 2, and a question card is not a permission card: a Day 5 take answered
# "which part did Cursor build?" with an answer that was simply untrue.
NO_ASK = (" Do not ask me any questions - if a choice comes up, make it, and "
          "tell me what you chose and why. If evidence for something is "
          "missing, say it is missing rather than guessing.")

_s: rig.Session | None = None


def bind(session: rig.Session) -> None:
    global _s
    _s = session


def beat(label: str, note: str = "") -> None:
    _s.mark(label, note)


def hold(x: float) -> None:
    time.sleep(x)


def gone(pattern: str) -> bool:
    return subprocess.run(["pgrep", "-f", pattern], capture_output=True).returncode != 0


def wait_process(pattern: str, minutes: float = 30.0, label: str = "") -> None:
    """Block while a shell command the driver started is still running.

    wait_idle watches Claude Code's elapsed counter; a plain node or docker
    command never touches it and reads as idle instantly, so everything typed
    after it queues as text instead of running.
    """
    time.sleep(8)
    deadline = time.time() + minutes * 60
    t0 = time.time()
    while time.time() < deadline:
        if gone(pattern):
            print(f"    . {label or pattern} finished after {time.time()-t0:.0f}s", flush=True)
            return
        time.sleep(5)
    print(f"    !! {pattern} still running after {minutes:.0f}m", flush=True)


def wait_for(path: str, minutes: float = 15.0, label: str = "") -> bool:
    """Ground truth is the file, not the spinner."""
    deadline = time.time() + minutes * 60
    while time.time() < deadline:
        if os.path.exists(path):
            print(f"    . {label or path} exists", flush=True)
            return True
        time.sleep(5)
    print(f"    !! {path} never appeared - filming what is actually there", flush=True)
    return False


def soak(minutes: float, label: str, every: float = 75.0, until_gone: str = "") -> None:
    """Film a long unattended run: approve prompts, keep marking, stay put."""
    end = time.time() + minutes * 60
    n = 0
    while time.time() < end:
        time.sleep(min(every, max(5.0, end - time.time())))
        if until_gone and gone(until_gone):
            print(f"    . {until_gone} gone, ending soak early", flush=True)
            return
        n += 1
        try:
            if W.prompt_showing():
                W.choose("2")
                hold(1.5)
                continue
        except Exception as exc:
            print(f"    . soak: {exc!r}", flush=True)
        beat(f"{label}/t{n:02d}", f"{n * every / 60:.1f} min in")


def setup(repo: str) -> None:
    W.focus_shoot_window()
    W.palette("View: Close All Editors", settle=2.0)
    W.one_terminal()
    S.osa('tell application "System Events" to tell process "Code"\n'
          '  set position of window 1 to {0, 0}\n'
          '  set size of window 1 to {1920, 1080}\n'
          'end tell')
    time.sleep(1.0)
    W.require_shoot_window()
    W.run(f"cd {repo}", wait=1.0)
    W.clean_prompt()


def read_doc(path: str, pages: int = 8, dwell: float = 3.2) -> None:
    """Open a file, hide the terminal so it fills the frame, and scroll it."""
    W.run(f"open -a 'Visual Studio Code' {path}", wait=4.0)
    W.focus_shoot_window()
    W.palette("View: Toggle Panel Visibility", settle=1.6)
    S.move(960, 520, 0.8)
    time.sleep(2.0)
    for _ in range(pages):
        S.scroll(-300)
        time.sleep(dwell)
    time.sleep(3.0)
    W.palette("View: Toggle Panel Visibility", settle=1.6)
    W.palette("View: Close All Editors", settle=1.5)
    W.palette("Terminal: Focus Terminal", settle=1.5)


def browser(url: str, wait: float = 14.0) -> None:
    W.run(f"'{CHROME}' --user-data-dir={PROFILE} --no-first-run "
          f"--no-default-browser-check --window-position=0,0 "
          f"--window-size=1920,1080 --app={url} >/dev/null 2>&1 &", wait=wait)


def close_browser() -> None:
    """Quit the browser and put the keyboard back in the terminal.

    focus_shoot_window only guarantees the WINDOW is frontmost. When VS Code
    comes back from another app it restores focus to whatever element had it
    last, which after a document read is the explorer tree - and every command
    typed after that goes into type-ahead file search and vanishes. A whole
    lecture filmed a motionless terminal that way, with no error anywhere,
    because the driver has no way to tell that its keystrokes are landing
    somewhere harmless.
    """
    S.osa('tell application "Google Chrome" to quit')
    time.sleep(3)
    W.focus_shoot_window()
    W.palette("Terminal: Focus Terminal", settle=1.5)


def term() -> None:
    """Put the keyboard in the terminal. Cheap, and cheaper than a lost take."""
    W.focus_shoot_window()
    W.palette("Terminal: Focus Terminal", settle=1.2)


def arg(flag: str, default=None):
    return sys.argv[sys.argv.index(flag) + 1] if flag in sys.argv else default


def run_session(beats, name_default: str, repo: str, allow=("Code", "Google Chrome"),
                need_gb: float = 18.0) -> None:
    """Standard main(): --only / --as / a positional start lecture."""
    only = arg("--only")
    name = arg("--as", name_default)
    positional = [a for a in sys.argv[1:] if a.startswith("w3-")]
    start = positional[0] if positional and not only else None

    todo = beats
    if only:
        todo = [(n, f) for n, f in beats if n == only]
        if not todo:
            sys.exit(f"unknown lecture {only}")
    elif start:
        idx = [i for i, (n, _) in enumerate(beats) if n == start]
        if not idx:
            sys.exit(f"unknown lecture {start}")
        todo = beats[idx[0]:]

    s = rig.Session(name, allow=allow)
    bind(s)
    s.start(need_gb=need_gb)
    t0 = time.time()
    try:
        setup(repo)
        for nm, fn in todo:
            print(f"\n=== {nm} ===", flush=True)
            fn()
    except Exception as exc:
        print(f"!! driver error: {exc!r}", flush=True)
    finally:
        s.stop()
        print(f"session ran {(time.time()-t0)/60:.1f} min", flush=True)
