#!/usr/bin/env python3
"""L30 take B — the project shape and the brief, with the frame set up FIRST.

The previous attempt filmed two minutes of VS Code's Welcome screen, whose
Recent list shows `~/Desktop/enfec_subs` and `~/Desktop/self`. That take was
deleted. The lesson is the one lesson one already taught and I ignored here:
verify the frame with a still before rolling, never during.

The frame is now: Welcome closed, the tree expanded on pm, agents.md open. This
script only records and scrolls — it sets nothing up.

agents.md is ~95 lines and the viewport shows ~45, so the whole traverse is about
900px. Budget the scroll from the line count, not from a guess.
"""
import os, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S


def beat(msg, hold):
    print(f"  {msg}  (hold {hold}s)", flush=True)
    time.sleep(hold)


if __name__ == "__main__":
    S.wait_front_window("Code", "pm")
    S.clear_mods()
    r = S.SegRec("L30_B_brief", allow=("Code",))
    r.start(settle=2.5)
    try:
        beat("the shape of the project, and the brief at the top", 16)
        S.scroll(-170); time.sleep(1.0)
        beat("requirements", 14)
        S.scroll(-170); time.sleep(1.0)
        beat("limitations, and design for many users anyway", 14)
        S.scroll(-170); time.sleep(1.0)
        beat("starting point — inherited, and a demo", 15)
        S.scroll(-170); time.sleep(1.0)
        beat("technical decisions", 14)
        S.scroll(-170); time.sleep(1.0)
        beat("palette, and coding standards", 14)
        S.scroll(-170); time.sleep(1.0)
        beat("root cause, and say so rather than following it into a corner", 15)
        S.scroll(-170); time.sleep(1.0)
        beat("working documentation — read the plan first", 14)
    finally:
        r.stop()
