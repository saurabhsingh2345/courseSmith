#!/usr/bin/env python3
"""CAPTURE A1b — depth pass for the Day 1 lectures that came out thin.

    python3 shoot_a1b.py w3-02 w3-03

Uses the SAME `w3-NN/` marker prefixes as A1, so cut.py appends this footage to
those lectures:  python3 postprod.py A1,A1b

Why this exists: asking an agent to do a thing takes twenty seconds, where the
reference instructor types for ten minutes. A first pass through a lecture is
therefore far shorter than the lecture needs. The fix is not a padded script —
it is to go back and actually READ what the agent wrote, which is teaching we
skipped the first time round and is the most useful footage in a no-code course.
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

s: rig.Session = None


def beat(label, note=""):
    s.mark(label, note)


def hold(x):
    time.sleep(x)


def read_file(path: str, pages: int = 10, pause: float = 2.2) -> None:
    """Open a file and actually read it on camera, slowly."""
    W.run(f"open -a 'Visual Studio Code' {path}", wait=3.5)
    W.focus_shoot_window()
    S.clear_mods()
    S.key("up", ("cmd",))
    hold(2.0)
    for _ in range(pages):
        S.scroll(-3, steps=6)
        hold(pause)
    hold(2)


# ---------------------------------------------------------------- w3-02
def w3_02() -> None:
    beat("w3-02/read-brief", "read the brief properly")
    read_file("plan.md", pages=16, pause=2.4)

    beat("w3-02/read-memory", "and the memory file it produced")
    read_file("CLAUDE.md", pages=6, pause=2.6)

    beat("w3-02/at-vs-link", "the one decision in that file")
    W.palette("Terminal: Focus Terminal", settle=2.0)
    W.run("claude --model sonnet", wait=8)
    W.wait_idle(stable=6, timeout=120, label="(launch)")
    W.turn(
        "In CLAUDE.md, plan.md is pulled in with an @ tag and other documents "
        "are referred to by path. Show me, concretely, what would change if I "
        "did it the other way round: what lands in your context on a fresh "
        "session, roughly how much of it, and what it costs me over a long day.",
        timeout=600, label="(@ vs link)")
    hold(6)

    beat("w3-02/history", "what the repo remembers")
    W.turn("/exit", timeout=60, stable=4, label="(exit)")
    W.run("git log --oneline --stat | head -40", wait=4)
    hold(8)


# ---------------------------------------------------------------- w3-03
def w3_03() -> None:
    beat("w3-03/read-command", "what a slash command actually is")
    read_file(".claude/commands/ct-brief.md", pages=5, pause=2.6)

    beat("w3-03/read-arg", "and one that takes an argument")
    read_file(".claude/commands/ct-check.md", pages=5, pause=2.6)

    beat("w3-03/explain", "it is a prompt in a file")
    W.palette("Terminal: Focus Terminal", settle=2.0)
    W.run("claude --model sonnet", wait=8)
    W.wait_idle(stable=6, timeout=120, label="(launch)")
    W.turn(
        "Explain what happens between me typing /ct-brief and you answering. Where "
        "does the file go, what replaces the argument, and why is this different "
        "from just pasting the same text every time?",
        timeout=600, label="(how it works)")
    hold(6)

    beat("w3-03/third", "one more, and this one earns its keep")
    W.turn(
        "Create a third command called ct-next. It should read plan.md, compare it "
        "with what is actually built, and propose the single next piece of work "
        "with a one-paragraph reason. Then run it.",
        timeout=900, label="(make /ct-next)")
    hold(4)
    W.turn("/ct-next", timeout=600, label="(run /ct-next)")
    hold(8)

    beat("w3-03/all", "three words that replace three paragraphs")
    W.run("ls -la .claude/commands", wait=3)
    hold(6)
    W.turn("/exit", timeout=60, stable=4, label="(exit)")
    hold(2)


BEATS = {"w3-02": w3_02, "w3-03": w3_03}


def main() -> None:
    global s
    want = [a for a in sys.argv[1:] if a.startswith("w3-")] or list(BEATS)
    s = rig.Session("A1b", allow=("Code",))
    s.start(need_gb=15)
    t0 = time.time()
    try:
        for name in want:
            if name not in BEATS:
                print(f"skip unknown {name}")
                continue
            print(f"\n=== {name} (depth) ===", flush=True)
            BEATS[name]()
    except Exception as exc:
        print(f"!! driver error: {exc!r}", flush=True)
    finally:
        s.stop()
        print(f"session ran {(time.time()-t0)/60:.1f} min", flush=True)


if __name__ == "__main__":
    main()
