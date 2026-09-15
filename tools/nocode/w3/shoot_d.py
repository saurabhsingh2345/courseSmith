#!/usr/bin/env python3
"""CAPTURE D — Week 3 Day 5. Other ways to run a swarm, the verdict, and shipping.

    python3 shoot_d.py [w3-23]
    python3 shoot_d.py --only w3-25 --as D2

Runs after Day 4, on the repo the agent team built.

The reference course compares Claude agent teams against two third-party
orchestrators. We compare against tools he already pays for — Cursor — and
against one we build ourselves with the Agent SDK from day three. That is a
better lesson anyway: by the end he is not choosing between other people's
orchestrators, he can write one.
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

REPO = "/Users/Shared/projects/control-tower"
CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
PROFILE = "/Users/Shared/projects/.shoot-chrome"
s: rig.Session = None


def beat(label, note=""):
    s.mark(label, note)


def hold(x):
    time.sleep(x)


# Every prompt that builds something ends with this. `turn()` answers ANY card
# that appears with option 2, and a question card is not a permission card: on
# the first Day 5 take that typed "2" into "which part did Cursor build?" and
# chose an answer that was simply untrue, which would have put a fabricated
# claim into the lecture. The driver cannot tell the two cards apart, so the
# fix is to make sure no question card is ever raised.
NO_ASK = (" Do not ask me any questions — if a choice comes up, make it, and "
          "tell me what you chose and why. If evidence for something is "
          "missing, say it is missing rather than guessing.")


def wait_for(path: str, minutes: float = 12.0, label: str = "") -> bool:
    """Block until a file the agent was asked to write actually exists.

    `wait_idle` watches the elapsed-second counter, and a turn that stops to ask
    a question stops ticking — so the driver called it finished, sent /exit into
    the question card, and the next beat ran a command against a file that had
    never been written. Ground truth is the file.
    """
    deadline = time.time() + minutes * 60
    while time.time() < deadline:
        if os.path.exists(path):
            print(f"    · {label or path} exists after "
                  f"{minutes * 60 - (deadline - time.time()):.0f}s", flush=True)
            return True
        time.sleep(5)
    print(f"    !! {path} never appeared — filming what is actually there",
          flush=True)
    return False


def wait_process(pattern: str, minutes: float = 20.0, label: str = "") -> None:
    """Block while a shell command the driver started is still running.

    Give it a moment to appear first, or a slow start reads as already finished.
    """
    import subprocess as sp
    time.sleep(6)
    deadline = time.time() + minutes * 60
    t0 = time.time()
    while time.time() < deadline:
        if sp.run(["pgrep", "-f", pattern], capture_output=True).returncode != 0:
            print(f"    · {label or pattern} finished after "
                  f"{time.time() - t0:.0f}s", flush=True)
            return
        time.sleep(4)
    print(f"    !! {pattern} still running after {minutes:.0f}m", flush=True)


def cursor_ready() -> bool:
    """Is cursor-agent actually logged in?

    Its stored auth expires, and a logged-out run films an error message instead
    of a comparison. Check before spending a take on it.
    """
    import subprocess
    r = subprocess.run(["cursor-agent", "-p", "-f", "Reply with: OK"],
                       capture_output=True, text=True, timeout=180,
                       cwd="/Users/Shared/projects")
    out = (r.stdout + r.stderr).lower()
    return "authentication is invalid" not in out and "agent login" not in out


# ---------------------------------------------------------------- w3-21
def w3_21() -> None:
    if not cursor_ready():
        print("    !! SKIPPING w3-21 — cursor-agent is logged out. "
              "Run `cursor-agent login`, then: shoot_d.py --only w3-21 --as D2",
              flush=True)
        return
    beat("w3-21/fair", "the same brief, a different tool")
    W.run("git status --short | head -20", wait=4)
    hold(6)
    W.run("git checkout -b cursor-run 2>&1 | tail -2", wait=4)
    hold(4)

    beat("w3-21/cursor", "hand it to Cursor")
    W.run("cursor-agent -p -f 'Read plan.md. The simulator and its tests already "
          "exist under src/sim. Build the next piece only: the read API over the "
          "simulator that the front end will call. Write its tests. Do not touch "
          "src/sim.' 2>&1 | tail -40", wait=20)
    W.wait_idle(stable=12, timeout=1800, label="(cursor builds)")
    hold(8)

    beat("w3-21/read", "and read what came back")
    W.run("git status --short | head -20", wait=4)
    hold(8)


# ---------------------------------------------------------------- w3-23
def w3_23() -> None:
    beat("w3-23/why", "three tools, three opinions")
    W.run("claude --model sonnet", wait=9)
    W.wait_idle(stable=6, timeout=180, label="(launch)")
    W.turn(
        # Do NOT claim a build that did not happen. This prompt used to say the
        # project had also been built by Cursor; w3-21 skips itself whenever
        # cursor-agent is logged out, and the claim then went on camera as fact.
        "We have now had this project built by an agent team inside Claude Code. "
        "That is somebody else's idea of how to coordinate work, and it is a good "
        "one. Using the Agent SDK we met on day three, what would it take to "
        "write my own — something that takes a list of independent pieces, runs "
        "an agent on each, and collects the results?",
        timeout=900, label="(the case for our own)")
    hold(6)

    beat("w3-23/build", "so we write one")
    W.turn(
        "Write it as tools/orchestrator.mjs, in JavaScript, using the Agent SDK "
        "for TypeScript from npm. I know that makes it this repo's first "
        "dependency and that CLAUDE.md says zero dependencies — that is a "
        "deliberate exception for a tool, not for the app, so install it, and "
        "note the exception in CLAUDE.md. It should read a short list of tasks, "
        "run each one through the SDK with its own allowed tools and its own "
        "working brief, run them concurrently, and print a report at the end "
        "saying what each produced and how long it took. Give it three real "
        "small tasks against this repo — a README for the tools folder, a short "
        "CONTRIBUTING note, and a docs page describing the simulated data — so "
        "that running it actually does something. Pin every agent it spawns to "
        "the sonnet model explicitly. Keep it readable: I want to be able to "
        "explain every line of it. Then explain it to me." + NO_ASK,
        timeout=1800, stable=12, label="(build the orchestrator)")
    hold(8)
    wait_for(os.path.join(REPO, "tools/orchestrator.mjs"), 6, "orchestrator.mjs")
    W.turn("/exit", timeout=60, stable=4, label="(exit)")

    beat("w3-23/run", "and run it")
    W.run(f"cd {REPO} && ls tools && node tools/orchestrator.mjs 2>&1 | tail -40",
          wait=20)
    # NOT wait_idle. It compares a crop of Claude Code's elapsed-second counter,
    # and a plain `node` process makes no counter tick — so the screen is static,
    # the driver calls it finished, and every later keystroke queues into a shell
    # that is still busy. That silently voided a whole lecture: /usage and the
    # verdict turn were typed as text while the orchestrator was still running.
    # A process either exists or it does not; ask the OS.
    wait_process("tools/orchestrator.mjs", 20, "orchestrator")
    hold(14)


# ---------------------------------------------------------------- w3-24
def w3_24() -> None:
    beat("w3-24/numbers", "what each one cost")
    W.run("claude --model sonnet", wait=9)
    W.wait_idle(stable=6, timeout=180, label="(launch)")
    W.turn("/usage", timeout=240, stable=5, label="(usage)")
    hold(10)
    S.clear_mods(); S.key("escape"); hold(2)

    beat("w3-24/verdict", "and the honest comparison")
    W.turn(
        "Compare the two ways this project has actually been built: the agent "
        "team inside Claude Code, and the orchestrator we wrote ourselves with "
        "the Agent SDK. Only those two — we did not run Cursor against this "
        "repo, so say that plainly and leave it out rather than inferring what "
        "it might have done. For each one: what it was genuinely good at, what "
        "it cost, what it could not do, and when you would actually reach for "
        "it. Ground every claim in something in this repository — a commit, a "
        "file, a timestamp — and where you have no evidence, say so. Then say "
        "which one you would use for this project tomorrow, and why." + NO_ASK,
        timeout=1200, stable=10, label="(the verdict)")
    hold(10)
    W.turn("/exit", timeout=60, stable=4, label="(exit)")


# ---------------------------------------------------------------- w3-25
def w3_25() -> None:
    """Ship it — into a container, not just onto localhost.

    The first take of this lecture restarted the app and committed nothing,
    which is w3-20 again with a different title. Running on your own machine is
    not shipping. The thing that makes it shippable is that it builds into an
    image and runs with none of your machine in it.

    NEVER `docker ps` or `docker images` here — those list his real client work.
    `docker compose ps` is scoped to this project and is the safe one.
    """
    beat("w3-25/dockerfile", "what the team already wrote")
    W.run(f"cd {REPO} && bash scripts/stop.sh; cat Dockerfile", wait=8)
    hold(14)

    beat("w3-25/build", "build the image")
    W.run("docker build -t control-tower:v1 . 2>&1 | tail -25", wait=20)
    W.wait_idle(stable=12, timeout=900, label="(docker build)")
    hold(10)

    beat("w3-25/up", "run it with none of my machine in it")
    W.run("docker run -d --rm -p 8090:8080 --name control-tower control-tower:v1 "
          "&& sleep 4 && docker compose ps 2>/dev/null; "
          "curl -s -o /dev/null -w 'container answers %{http_code}\\n' "
          "http://localhost:8090/", wait=25)
    hold(12)

    beat("w3-25/alive", "Control Tower, from the image")
    W.run(f"'{CHROME}' --user-data-dir={PROFILE} --no-first-run "
          f"--no-default-browser-check --window-position=0,0 "
          f"--window-size=1920,1080 --app=http://localhost:8090 "
          f">/dev/null 2>&1 &", wait=14)
    hold(40)

    beat("w3-25/commit", "and put it somewhere")
    W.focus_shoot_window()
    hold(4)
    W.run("docker stop control-tower >/dev/null 2>&1; "
          "git add -A && git commit -q -m 'control tower v1' 2>&1 | tail -2; "
          "git log --oneline | head -5", wait=10)
    hold(12)


BEATS = [("w3-21", w3_21), ("w3-23", w3_23), ("w3-24", w3_24), ("w3-25", w3_25)]


def _arg(flag, default=None):
    return sys.argv[sys.argv.index(flag) + 1] if flag in sys.argv else default


def main() -> None:
    global s
    only = _arg("--only")
    name = _arg("--as", "D")
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
        for n, fn in todo:
            print(f"\n=== {n} ===", flush=True)
            fn()
    except Exception as exc:
        print(f"!! driver error: {exc!r}", flush=True)
    finally:
        s.stop()
        print(f"session ran {(time.time()-t0)/60:.1f} min", flush=True)


if __name__ == "__main__":
    main()
