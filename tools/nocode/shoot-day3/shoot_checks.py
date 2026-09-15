#!/usr/bin/env python3
"""Run the five checks by hand, in Chrome, on any of the four builds.

Usage: shoot_checks.py <take_name> <url> [geometry.json]

The five checks are the spine of Day 3 — the same five, in the same order, on
every build, so that if one app is genuinely worse than another the film can
point at where. They are run by hand on camera because every one of these agents
now writes and runs its own tests and then reports success, and a tool that
marks its own exam has demonstrated consistency, not correctness.

Geometry differs per build (each agent lays the board out its own way), so the
card and button coordinates come from a small json probed off a still before
rolling rather than being hardcoded. Defaults match the Copilot build.

Holds are long on purpose: a check that flicks past in two seconds teaches
nothing, and the deck's cut retimes each clip to its narration window anyway.
"""
import json, os, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S

DEFAULT = {
    "col1_card1":  [362, 528],     # the card we drag out of column one
    "col3_drop":   [959, 660],     # empty space in column three
    "col2_card1":  [661, 528],     # the card we reorder
    "col2_below":  [661, 700],
    "del_card":    [1258, 528],    # hover this to reveal its delete control
    "del_button":  [1366, 490],
    "col1_title":  [270, 441],     # click to rename
    "add_button":  [1641, 348],
    "form_title":  [960, 515],
    "form_desc":   [960, 583],
    "form_submit": [1613, 655],
    "new_name":    "Backlog",
    "new_title":   "Check it myself",
    "new_desc":    "A claim is not a demonstration.",
}


def beat(msg, hold):
    print(f"  {msg}  (hold {hold}s)", flush=True)
    time.sleep(hold)


def main():
    take, url = sys.argv[1], sys.argv[2]
    g = dict(DEFAULT)
    if len(sys.argv) > 3 and os.path.exists(sys.argv[3]):
        g.update(json.load(open(sys.argv[3])))

    S.wait_front("Google Chrome"); time.sleep(0.8); S.clear_mods()
    S.key('r', ('cmd',))                 # reset the board; state is in memory
    time.sleep(4.5)
    S.move(960, 300, 0.6)

    r = S.SegRec(take, allow=("Google Chrome",))
    r.start(settle=2.5)
    try:
        beat("open on the finished board", 9)

        S.drag(*g["col1_card1"], *g["col3_drop"], glide=1.8, hold=0.7)
        beat("check 1: drag between columns", 11)

        S.drag(*g["col2_card1"], *g["col2_below"], glide=1.4, hold=0.7)
        beat("check 2: reorder within a column", 11)

        S.move(*g["del_card"], 1.0); time.sleep(1.2)
        S.click(*g["del_button"], 0.9)
        beat("check 3: delete a card", 11)

        S.click(*g["col1_title"], 0.9); time.sleep(1.0)
        S.key('a', ('cmd',)); time.sleep(0.5)
        S.type_text(g["new_name"], cps=7); time.sleep(1.0)
        S.key('return')
        beat("check 4: rename a column", 11)

        S.click(*g["add_button"], 0.9); time.sleep(1.6)
        S.click(*g["form_title"], 0.8); time.sleep(0.7)
        S.type_text(g["new_title"], cps=8); time.sleep(0.8)
        S.click(*g["form_desc"], 0.8); time.sleep(0.7)
        S.type_text(g["new_desc"], cps=8); time.sleep(1.2)
        S.click(*g["form_submit"], 0.9)
        beat("check 5: add a card with a description", 13)

        S.move(960, 700, 1.2)
        beat("hold on the finished board", 12)
    finally:
        r.stop()


if __name__ == "__main__":
    main()
