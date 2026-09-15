#!/usr/bin/env python3
"""Drive Claude Code in the VS Code integrated terminal for a Week 3 capture.

Named w3drive, not drive: the day rig's stage.py imports its own module called
`drive`, and a file of that name on sys.path shadows it — stage.type_text then
fails with AttributeError halfway through a take.

Claude Code in a terminal answers on its own schedule, so an unattended take
needs a reliable "it has stopped working" signal. Polling for a fixed sleep
either films a frozen screen or cuts the agent off mid-answer.

wait_idle() watches a crop of the terminal and returns once the pixels have
stopped changing for `stable` seconds. The spinner, the streaming text and the
token counter all move while it is thinking, so stillness is a good proxy for
done — and it costs nothing when the agent finishes early.
"""

from __future__ import annotations

import os
import subprocess
import sys
import time

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S  # noqa: E402

from PIL import Image, ImageChops  # noqa: E402

SCAN = "/tmp/_w3_scan.png"
# Two crops, because one was not enough.
#
# TERM is the body of the terminal. STATUS is the thin strip just above the input
# box, where Claude Code prints its working line — "Twisting… (46s · 2.6k tokens)".
# That elapsed-second counter ticks every second while a turn is running, so it
# is a reliable "still working" signal. Watching only the body declared idle
# after 30 seconds while the agent was still going, and the next prompt was then
# QUEUED rather than sent — which is how a take recorded five lectures against a
# repo where nothing had been built.
TERM = (300, 300, 1900, 1040)
STATUS = (300, 850, 1900, 1078)



class WrongWindow(RuntimeError):
    pass


SHOOT_TITLE = "control-tower"


def require_shoot_window() -> None:
    """Refuse to type unless the shoot window is really in front.

    This exists because it already went wrong: VS Code was put in fullscreen,
    macOS moved it to its own Space, the panel showed a different Space, and
    every keystroke the driver sent landed in the terminal running the session
    instead — where they arrived as chat messages. Nothing was damaged, but a
    driver that types blind into whatever happens to be frontmost is not safe to
    leave running for two hours.
    """
    front = S.frontmost_fast()
    title = S.front_window_title("Code") or ""
    if front != "Code" or SHOOT_TITLE not in title:
        raise WrongWindow(
            f"frontmost={front!r} title={title!r} — expected Code / {SHOOT_TITLE}")


def focus_shoot_window(tries: int = 4) -> None:
    """Bring the shoot window back and confirm it. Raises if it cannot.

    Two ways of raising it, because one is not enough. System Events'
    `set frontmost` loses to an application that has just taken focus itself -
    a capture died on its first beat with Finder in front and VS Code's window
    correctly titled, which is exactly the state `set frontmost` cannot get out
    of. `activate` goes through the application rather than the accessibility
    layer and wins.
    """
    for _ in range(tries):
        try:
            require_shoot_window()
            return
        except WrongWindow:
            S.osa('tell application "Visual Studio Code" to activate')
            time.sleep(1.0)
            S.osa('tell application "System Events" to tell process "Code" '
                  'to set frontmost to true')
            time.sleep(1.2)
            S.osa('tell application "System Events" to tell process "Code"\n'
                  '  set position of window 1 to {0, 0}\n'
                  '  set size of window 1 to {1920, 1080}\n'
                  'end tell')
            time.sleep(1.0)
    require_shoot_window()


def _crop() -> tuple[Image.Image, Image.Image]:
    subprocess.run(["screencapture", "-x", "-D", "1", SCAN], capture_output=True)
    subprocess.run(["sips", "-Z", "1920", SCAN, "-o", SCAN], capture_output=True)
    im = Image.open(SCAN).convert("L")
    return im.crop(TERM), im.crop(STATUS)


def _diff(a: Image.Image, b: Image.Image, size=(160, 74)) -> int:
    return sum(ImageChops.difference(a, b).point(lambda p: 255 if p > 26 else 0)
               .resize(size).getdata()) // 255


def wait_idle(stable: float = 6.0, timeout: float = 900.0,
              poll: float = 1.5, quiet_px: int = 12, label: str = "") -> bool:
    """Block until the terminal stops changing. True if it settled, False on timeout."""
    t0 = time.time()
    last, still = _crop(), 0.0
    while time.time() - t0 < timeout:
        time.sleep(poll)
        now = _crop()
        body = _diff(last[0], now[0])
        # The status strip is small and its counter is high contrast, so it is
        # compared at full size with a tight threshold.
        status = _diff(last[1], now[1], size=(400, 34))
        last = now
        busy = body > quiet_px or status > 2
        still = 0.0 if busy else still + poll
        if still >= stable:
            print(f"    idle after {time.time()-t0:.0f}s {label}", flush=True)
            return True
    print(f"    !! TIMEOUT after {timeout:.0f}s {label}", flush=True)
    return False


def say(text: str, cps: int = 26, send: bool = True, settle: float = 0.8) -> None:
    """Type a prompt into Claude Code and send it."""
    _check_forbidden(text)
    focus_shoot_window()
    S.clear_mods()
    S.type_text(text, cps=cps)
    time.sleep(settle)
    if send:
        S.key("return")


def choose(n: str = "2", settle: float = 0.6) -> None:
    """Answer a permission prompt by number."""
    focus_shoot_window()
    S.clear_mods()
    S.type_text(n, cps=8)
    time.sleep(settle)
    S.key("return")


def run(cmd: str, wait: float = 1.2) -> None:
    """Run a shell line in the terminal (Claude Code not attached)."""
    focus_shoot_window()
    S.clear_mods()
    S.type_text(cmd, cps=30)
    S.key("return")
    time.sleep(wait)


def paste(path: str) -> None:
    """Put a file on the clipboard so a long document lands in one keystroke.

    Typing a 120-line spec character by character takes four minutes of screen
    time and reads as filler. The lecture wants the document being READ, not
    typed."""
    subprocess.run(["pbcopy"], stdin=open(path, "rb"), check=True)
    S.clear_mods()
    S.key("v", ("cmd",))
    time.sleep(0.8)


def clean_prompt(settle: float = 1.0) -> None:
    """Strip the username and hostname out of the zsh prompt, then wipe scrollback.

    The default prompt on this machine prints `enfecsolutions@Enfecs-MacBook-Pro-3`
    on every single line. His .zshrc sets it, so no VS Code env setting can win —
    it has to be done per terminal, after the shell has started.
    """
    run("PROMPT='%~ %# '", wait=0.6)
    run("clear && printf '\\033[3J'", wait=settle)


def palette(cmd: str, settle: float = 1.6) -> None:
    """Run a VS Code command by name. Safer than a chord that means different
    things depending on what has focus."""
    S.clear_mods()
    S.key("p", ("cmd", "shift"))
    time.sleep(1.0)
    S.type_text(cmd, cps=24)
    time.sleep(1.0)
    S.key("return")
    time.sleep(settle)


def one_terminal() -> None:
    """Leave exactly one clean terminal open.

    Do NOT do this with repeated Cmd+W. Cmd+W closes a terminal tab only while
    the terminal has focus; otherwise it closes editors and then the WINDOW. Six
    of them shut VS Code entirely mid-setup and put his browser on the panel.
    """
    palette("Terminal: Kill All Terminals", settle=2.0)
    palette("Terminal: Create New Terminal", settle=3.5)
    clean_prompt()


# Claude Code draws a wide amber rule above every permission prompt. Sampled off
# a real prompt at (248,192,56) spanning essentially the full width, which is
# what makes it distinguishable from ordinary syntax colour.
AMBER = (248, 192, 56)

# Built-in commands that put identity on screen. /status is the one that caught
# us: it renders "Login method / Organization / Email" and it also SHADOWS a
# custom command of the same name, so a project command called status silently
# opens the built-in panel instead. Our own commands are namespaced `/ct-*` so
# they cannot collide with anything Claude Code ships now or later.
FORBIDDEN = ("/status", "/config", "/login", "/logout", "/doctor",
             "/feedback", "/bug", "/upgrade", "/whoami")


def _check_forbidden(text: str) -> None:
    head = text.strip().split()[0].lower() if text.strip() else ""
    if head in FORBIDDEN:
        raise RuntimeError(
            f"{head} must never be filmed — it shows his email and organisation. "
            "Use a namespaced custom command (/ct-brief) or /usage instead.")


def prompt_showing(tol: int = 26, min_run: int = 200) -> bool:
    subprocess.run(["screencapture", "-x", "-D", "1", SCAN], capture_output=True)
    im = Image.open(SCAN).convert("RGB")
    w, h = im.size
    px = im.load()
    for y in range(0, h, 3):
        c = 0
        for x in range(0, w, 6):
            r, g, b = px[x, y]
            if (abs(r - AMBER[0]) < tol and abs(g - AMBER[1]) < tol
                    and abs(b - AMBER[2]) < tol):
                c += 1
                if c >= min_run:
                    return True
    return False


def turn(text: str, timeout: float = 900.0, stable: float = 7.0,
         answer: str = "2", rounds: int = 6, label: str = "") -> None:
    """Send one instruction and see it through, answering any permission prompt.

    An unattended take that stops on a prompt films a frozen screen. Option 2 is
    "yes, and don't ask again for this", which keeps a long session moving without
    ever granting more than the tool it is actually asking about.
    """
    say(text)
    if text.strip() in ("/exit", "/quit"):
        # Quitting never needs an answer, and the amber matcher false-positived
        # on the shutdown output six times in one take — each one typing a stray
        # "2" and a return into whatever had focus next.
        wait_idle(stable=stable, timeout=min(timeout, 90), label=label)
        # Then hand back a KNOWN-GOOD shell. Claude Code's TUI keeps hold of the
        # terminal for a moment after it exits, and the next line typed is
        # swallowed — which silently ran a whole install in the previous folder
        # once. A fresh terminal cannot swallow anything.
        one_terminal()
        return
    for _ in range(rounds):
        wait_idle(stable=stable, timeout=timeout, label=label)
        if not prompt_showing():
            if text.strip().startswith("/"):
                # A built-in like /usage or /context opens a modal that swallows
                # the next prompt. Close it before the driver types again.
                S.clear_mods()
                S.key("escape")
                time.sleep(1.0)
            return
        print("    · answering a permission prompt", flush=True)
        choose(answer)
        time.sleep(1.2)


def trust() -> None:
    """Answer Claude Code's folder safety check, which is arrows not numbers."""
    if prompt_showing():
        S.clear_mods()
        S.key("down")
        time.sleep(0.5)
        S.key("return")
        time.sleep(3.0)
