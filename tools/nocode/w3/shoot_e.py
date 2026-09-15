#!/usr/bin/env python3
"""CAPTURE E — a second product, built by a program.

    python3 shoot_e.py            # w3-27 .. w3-32
    python3 shoot_e.py w3-30      # from here on
    python3 shoot_e.py --only w3-31 --as E2

The reference week compares Claude against two third-party orchestrators. Both
of ours are logged out on this machine and we do not type his credentials, so
the second way to run this work is one we WRITE: the Agent SDK orchestrator
from w3-23, grown into a real planner / worker-pool / reviewer pipeline and
pointed at a brief it has never seen.

Nothing about the outcome is scripted. Whatever the program builds — including
what it gets wrong — is the lecture, and the narration is written afterwards
from the frames.
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

W.SHOOT_TITLE = "signal-desk"

REPO = "/Users/Shared/projects/signal-desk"
SRC = "/Users/Shared/projects/control-tower"
BRIEF = "/Users/Shared/projects/.prep/signal-brief.md"
CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
PROFILE = "/Users/Shared/projects/.shoot-chrome"

s: rig.Session = None

NO_ASK = (" Do not ask me any questions - if a choice comes up, make it, and "
          "tell me what you chose and why. If evidence for something is "
          "missing, say it is missing rather than guessing.")


def beat(label, note=""):
    s.mark(label, note)


def hold(x):
    time.sleep(x)


def soak(minutes: float, label: str, every: float = 75.0,
         until_gone: str = "") -> None:
    """Film a long unattended run: approve prompts, keep marking, stay put.

    `until_gone` is a pgrep pattern. A fixed soak either cuts a slow build off
    or films eighteen minutes of a finished prompt, and both cost a lecture.
    """
    end = time.time() + minutes * 60
    n = 0
    while time.time() < end:
        time.sleep(min(every, max(5.0, end - time.time())))
        if until_gone and subprocess.run(
                ["pgrep", "-f", until_gone], capture_output=True).returncode != 0:
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


def wait_process(pattern: str, minutes: float = 45.0, label: str = "") -> None:
    time.sleep(8)
    deadline = time.time() + minutes * 60
    t0 = time.time()
    while time.time() < deadline:
        if subprocess.run(["pgrep", "-f", pattern], capture_output=True).returncode != 0:
            print(f"    . {label or pattern} finished after {time.time()-t0:.0f}s", flush=True)
            return
        time.sleep(5)
    print(f"    !! {pattern} still running after {minutes:.0f}m", flush=True)


def wait_for(path: str, minutes: float = 15.0, label: str = "") -> bool:
    deadline = time.time() + minutes * 60
    while time.time() < deadline:
        if os.path.exists(path):
            print(f"    . {label or path} exists", flush=True)
            return True
        time.sleep(5)
    print(f"    !! {path} never appeared", flush=True)
    return False


# ------------------------------------------------------------------ setup
def setup() -> None:
    W.focus_shoot_window()
    W.palette("View: Close All Editors", settle=2.0)
    W.one_terminal()
    S.osa('tell application "System Events" to tell process "Code"\n'
          '  set position of window 1 to {0, 0}\n'
          '  set size of window 1 to {1920, 1080}\n'
          'end tell')
    time.sleep(1.0)
    W.require_shoot_window()
    W.run(f"cd {REPO}", wait=1.0)
    W.clean_prompt()


def read_doc(path: str, pages: int = 8, dwell: float = 3.2) -> None:
    """Open a file, hide the terminal so it fills the frame, and scroll it.

    A document read in the top quarter of the screen with a terminal under it is
    unreadable at 1080p. Cmd+J is not used: the palette command means the same
    thing no matter what has focus.
    """
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


# ------------------------------------------------------------------ w3-27
def w3_27() -> None:
    beat("w3-27/empty", "a second empty folder")
    W.run("pwd && ls -la && node -v", wait=4)
    hold(8)

    beat("w3-27/git", "version control before anything else")
    W.run("git init -q && git status --short && echo 'clean'", wait=4)
    hold(6)

    beat("w3-27/brief", "the brief lands")
    # pbpaste, never a heredoc: a multi-line paste into zsh's line editor does
    # not submit its last line, and never a Save As either - VS Code appends the
    # detected extension and produced plan.md.md on a Day 1 take.
    subprocess.run(["pbcopy"], stdin=open(BRIEF, "rb"), check=True)
    W.run("pbpaste > BRIEF.md", wait=2.0)
    W.run("wc -l BRIEF.md", wait=3)
    hold(5)

    beat("w3-27/read", "read it, because everything downstream comes from it")
    read_doc("BRIEF.md", pages=7, dwell=3.6)

    beat("w3-27/claude", "hand it to an agent to turn into a plan")
    W.palette("Terminal: Focus Terminal", settle=1.5)
    W.run("claude --model sonnet", wait=8)
    W.trust()
    hold(5)

    beat("w3-27/plan", "a brief is not a plan")
    W.turn(
        "Read BRIEF.md. Turn it into plan.md: a build plan for this repository "
        "divided into areas that different people could own at the same time "
        "without treading on each other. For each area give it a name, the "
        "files it owns, what it must expose to the others, and what it must "
        "not touch. Put the simulator first because everything else consumes "
        "it. Also write CLAUDE.md with the conventions from the brief. Build "
        "nothing yet - only plan.md, CLAUDE.md and package.json. package.json is "
        "type module, engines node >= 24, and its scripts are start and test, "
        "test being node --test." + NO_ASK,
        timeout=900, label="plan")
    hold(6)

    beat("w3-27/planread", "read what it decided")
    W.turn("/exit")
    read_doc("plan.md", pages=12, dwell=3.2)


# ------------------------------------------------------------------ w3-28
def w3_28() -> None:
    beat("w3-28/tool", "the orchestrator we already wrote")
    W.run(f"mkdir -p tools && cp {SRC}/tools/orchestrator.mjs tools/ "
          f"&& wc -l tools/orchestrator.mjs", wait=4)
    hold(6)
    read_doc("tools/orchestrator.mjs", pages=6, dwell=3.0)

    beat("w3-28/grow", "eighty lines is a fan-out, not a pipeline")
    W.run("claude --model sonnet", wait=8)
    hold(4)
    W.turn(
        "Rewrite tools/orchestrator.mjs so it can build this repository from "
        "plan.md without a human in the loop. Three phases, each runnable on "
        "its own from the command line. "
        "plan: one agent reads plan.md and writes tasks.json - an array of "
        "tasks, each with an id, the area it belongs to, the files it owns, "
        "the tools it is allowed, a list of task ids it depends on, and a "
        "brief written for someone who cannot ask questions. "
        "run: execute tasks.json respecting dependencies, at most three "
        "agents at once, each its own query() session with only its allowed "
        "tools, appending a timestamped line to run.log every time a task "
        "starts, finishes or fails, and printing a table at the end. "
        "review: one agent with read-only tools inspects what was built "
        "against plan.md and writes REVIEW.md saying what is present, what is "
        "missing, and what is wrong - grounded in files it actually read. "
        "Use claude-sonnet-5. Keep it one file and keep it readable: this is a "
        "tool a person has to be able to debug at two in the morning." + NO_ASK,
        timeout=1200, label="orchestrator")
    hold(6)
    W.turn("/exit")

    beat("w3-28/read", "read the machine that will do the building")
    W.run("wc -l tools/orchestrator.mjs", wait=4)
    hold(6)
    read_doc("tools/orchestrator.mjs", pages=16, dwell=2.9)

    beat("w3-28/dep", "one dependency, and it is not part of the app")
    W.run("npm i @anthropic-ai/claude-agent-sdk", wait=8)
    wait_process("npm i @anthropic", minutes=6, label="npm install")
    hold(8)


# ------------------------------------------------------------------ w3-29
def w3_29() -> None:
    beat("w3-29/plan", "phase one: turn the plan into work")
    W.run("node tools/orchestrator.mjs plan", wait=10)
    wait_process("orchestrator.mjs plan", minutes=12, label="plan phase")
    hold(10)

    beat("w3-29/tasks", "what it decided the work is")
    W.run("node -e \"const t=require('fs').readFileSync('tasks.json','utf8');"
          "const a=JSON.parse(t);console.log(a.length+' tasks');"
          "for(const x of a)console.log(' -',x.id,'|',(x.area||''),'|',"
          "(x.dependsOn||[]).join(',')||'-')\" 2>&1 | head -30", wait=6)
    hold(12)

    beat("w3-29/run", "phase two: let it build")
    W.run("node tools/orchestrator.mjs run 2>&1 | tee build.out", wait=8)
    soak(40, "w3-29/build", every=75, until_gone="orchestrator.mjs run")
    wait_process("orchestrator.mjs run", minutes=45, label="build phase")
    hold(14)

    beat("w3-29/table", "what finished, what failed")
    hold(16)
    W.run("tail -30 run.log", wait=6)
    hold(12)


# ------------------------------------------------------------------ w3-30
def w3_30() -> None:
    beat("w3-30/tree", "what is actually on disk now")
    W.run("git status --short | head -30 && echo '---' && "
          "find src tests -type f 2>/dev/null | head -30", wait=6)
    hold(14)

    beat("w3-30/size", "how much of it there is")
    W.run("find . -name '*.js' -not -path './node_modules/*' | xargs wc -l | tail -20",
          wait=6)
    hold(12)

    beat("w3-30/tests", "does it pass its own tests")
    W.run("npm test", wait=10)
    wait_process("node --test", minutes=8, label="tests")
    hold(16)

    beat("w3-30/review", "phase three: an agent marks its own homework")
    W.run("node tools/orchestrator.mjs review", wait=10)
    wait_process("orchestrator.mjs review", minutes=15, label="review phase")
    hold(8)
    read_doc("REVIEW.md", pages=11, dwell=3.3)


# ------------------------------------------------------------------ w3-31
def w3_31() -> None:
    beat("w3-31/start", "the only question that matters")
    W.run("(npm start >/tmp/sd.log 2>&1 &) ; sleep 6 ; "
          "curl -s -o /dev/null -w 'it answers %{http_code}\\n' http://localhost:8081/",
          wait=16)
    hold(10)

    beat("w3-31/alive", "Signal Desk, first look")
    W.run(f"'{CHROME}' --user-data-dir={PROFILE} --no-first-run "
          f"--no-default-browser-check --window-position=0,0 "
          f"--window-size=1920,1080 --app=http://localhost:8081 "
          f">/dev/null 2>&1 &", wait=14)
    hold(25)

    beat("w3-31/watch", "watch it move before touching anything")
    for i in range(8):
        S.move(300 + (i % 4) * 120, 260 + (i % 5) * 90, 1.2)
        hold(4.0)

    beat("w3-31/pick", "pick an instrument")
    S.click(220, 300, 1.0)
    hold(7)
    S.click(220, 400, 1.0)
    hold(7)

    beat("w3-31/rules", "the rules panel")
    S.move(1650, 300, 1.2)
    hold(3)
    for _ in range(4):
        S.scroll(-240)
        hold(2.5)
    S.click(1650, 420, 1.0)
    hold(8)

    beat("w3-31/replay", "rewind the whole console")
    S.move(700, 760, 1.2)
    hold(2)
    S.drag(700, 760, 1300, 760, glide=2.4)
    hold(8)
    S.drag(1300, 760, 900, 760, glide=2.0)
    hold(10)

    beat("w3-31/honest", "and what is wrong with it")
    hold(14)


# ------------------------------------------------------------------ w3-32
def w3_32() -> None:
    beat("w3-32/back", "back to the terminal with a list")
    S.osa('tell application "Google Chrome" to quit')
    hold(3)
    W.focus_shoot_window()
    # The keyboard does not come back with the window: VS Code restores focus to
    # whatever had it last, and after a document read that is the explorer tree.
    W.palette("Terminal: Focus Terminal", settle=1.5)
    W.run("clear && printf '\\033[3J'", wait=2)
    hold(4)

    beat("w3-32/hand", "hand the defects to the same machine that built it")
    W.run("claude --model sonnet", wait=8)
    hold(4)
    W.turn(
        "Read REVIEW.md and plan.md. The console starts and serves on 8081. "
        "Work through REVIEW.md's list of what is missing or wrong, highest "
        "impact first, and fix it - actually edit the files, do not describe "
        "the fix. After each fix run npm test. Stop when the tests pass and "
        "every item you could reach is done, then write FIXED.md listing what "
        "you changed and what you deliberately left." + NO_ASK,
        timeout=2400, label="fix")
    hold(8)
    W.turn("/exit")

    beat("w3-32/verify", "did it hold")
    W.run("npm test", wait=10)
    wait_process("node --test", minutes=8, label="tests")
    hold(14)
    W.run("git add -A && git commit -q -m 'signal desk v1' && git log --oneline | head -5",
          wait=6)
    hold(10)

    beat("w3-32/again", "and look again")
    W.run("(npm start >/tmp/sd2.log 2>&1 &) ; sleep 6 ; "
          "curl -s -o /dev/null -w 'still answers %{http_code}\\n' http://localhost:8081/",
          wait=14)
    W.run(f"'{CHROME}' --user-data-dir={PROFILE} --no-first-run "
          f"--no-default-browser-check --window-position=0,0 "
          f"--window-size=1920,1080 --app=http://localhost:8081 "
          f">/dev/null 2>&1 &", wait=14)
    hold(30)
    for i in range(6):
        S.move(400 + (i % 3) * 200, 300 + (i % 4) * 110, 1.3)
        hold(4.5)
    S.osa('tell application "Google Chrome" to quit')
    hold(3)
    W.focus_shoot_window()
    W.palette("Terminal: Focus Terminal", settle=1.5)
    hold(4)


BEATS = [("w3-27", w3_27), ("w3-28", w3_28), ("w3-29", w3_29),
         ("w3-30", w3_30), ("w3-31", w3_31), ("w3-32", w3_32)]


def _arg(flag, default=None):
    return sys.argv[sys.argv.index(flag) + 1] if flag in sys.argv else default


def main() -> None:
    global s
    only = _arg("--only")
    name = _arg("--as", "E")
    positional = [a for a in sys.argv[1:] if a.startswith("w3-")]
    start = positional[0] if positional and not only else None

    todo = BEATS
    if only:
        todo = [(n, f) for n, f in BEATS if n == only]
        if not todo:
            sys.exit(f"unknown lecture {only}")
    elif start:
        idx = [i for i, (n, _) in enumerate(BEATS) if n == start]
        if not idx:
            sys.exit(f"unknown lecture {start}")
        todo = BEATS[idx[0]:]

    s = rig.Session(name, allow=("Code", "Google Chrome"))
    s.start(need_gb=20)
    t0 = time.time()
    try:
        setup()
        for nm, fn in todo:
            print(f"\n=== {nm} ===", flush=True)
            fn()
    except Exception as exc:
        print(f"!! driver error: {exc!r}", flush=True)
    finally:
        s.stop()
        print(f"session ran {(time.time()-t0)/60:.1f} min", flush=True)


if __name__ == "__main__":
    main()
