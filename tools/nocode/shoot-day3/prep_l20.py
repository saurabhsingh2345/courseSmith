#!/usr/bin/env python3
"""L20 prep — Antigravity IDE on the fresh clone.

Antigravity IDE 1.107.0 is a VS Code fork (`/Applications/Antigravity IDE.app`,
dataFolderName `.antigravity-ide`), so everything learned about staging VS Code
applies: open a new window on the neutral path, confirm by title before touching
anything, take only that window fullscreen, and hide nothing.

Two things are specific to this one:

  * **First launch may want a Google sign-in, and that is his to do, not mine.**
    This script stops and says so rather than going anywhere near a credential
    screen. Run it, and if it reports a sign-in wall, hand back.
  * **ref-L20's central claim is out of date and must be re-checked here.** The
    reference says Antigravity has not adopted `agents.md` and makes you copy the
    brief into `.agent/rules/strategy.md`. This build's bundle carries an
    `AGENTS.md` settings tab, a `USE_AGENT_MD` flag and an
    `[InstructionsContextComputer] AGENTS.md files added:` log line, so it looks
    to have adopted it. Confirm on camera before narrating either way. Note our
    file is lowercase `agents.md` against its uppercase `AGENTS.md`, which only
    matches because macOS is case-insensitive — worth saying out loud.
"""
import os, subprocess, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

APP = "/Applications/Antigravity IDE.app"
PROC = "Antigravity IDE"
PROJ = "/Users/Shared/projects/kanban"
OUT = os.path.dirname(os.path.abspath(__file__))


def main():
    if not os.path.isdir(APP):
        raise SystemExit(f"{APP} is not installed")
    S.desktop_quiet(True)
    S.close_finder_windows()
    print("hidden before (must be empty):",
          repr(S.osa('tell application "System Events" to get name of every process '
                     'whose visible is false and background only is false')))

    # its CLI, if the fork shipped one, opens a folder directly
    cli = os.path.join(APP, "Contents/Resources/app/bin/antigravity-ide")
    if os.path.exists(cli):
        subprocess.Popen([cli, "-n", PROJ])
    else:
        subprocess.Popen(["open", "-a", APP, PROJ])
    time.sleep(22)

    front = S.frontmost()
    print("frontmost:", front)
    if PROC not in front:
        print("NOTE: Antigravity is not frontmost yet — it may be updating on "
              "first launch. Waiting once more.")
        time.sleep(20)
        front = S.frontmost()
        print("frontmost:", front)

    snap = S.snap(os.path.join(OUT, "l20_first_launch.png"))
    print("snapshot:", snap)
    print()
    print("STOP AND LOOK AT THAT SNAPSHOT BEFORE GOING FURTHER.")
    print("If it shows a Google sign-in screen, do not drive it — that is his")
    print("credential. Hand back, let him sign in, then re-run this script.")


if __name__ == "__main__":
    main()
