#!/usr/bin/env python3
"""L20 take B — the plan beat, filmed after the fact.

The first L20 recorder never captured a frame (its guard allowed only
"Electron", the AppleScript process name, while SegRec polls NSWorkspace, which
calls this app "Antigravity IDE"). By the time that was found the plan had been
approved. The conversation persists, so the three beats that matter are filmed
here by scrolling back to them — same screen, same words, a few minutes later:

  1. "based on the requirements in AGENTS.md" — the sentence that kills the
     reference lecture's whole middle section;
  2. the implementation plan it wrote before touching anything;
  3. the model list, with Anthropic's models and an open-weights model sitting in
     Google's own dropdown.
"""
import os, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S

ALLOW = ("Antigravity IDE", "Electron")
PANEL = (1500, 500)
MODEL_CHEV = (1327, 1028)


def beat(msg, hold):
    print(f"  {msg}  (hold {hold}s)", flush=True)
    time.sleep(hold)


if __name__ == "__main__":
    print("frontmost:", S.frontmost_fast(), flush=True)
    S.clear_mods()
    r = S.SegRec("L20_B_plan", allow=ALLOW)
    r.start(settle=2.5)
    try:
        beat("open on the finished build", 6)

        # scroll the conversation back to the first reply
        S.move(*PANEL, 0.6)
        for _ in range(14):
            S.scroll(700)
            time.sleep(0.55)
        beat("scrolled back to the AGENTS.md reply", 16)

        for _ in range(3):
            S.scroll(-260)
            time.sleep(0.6)
        beat("the plan it wrote before touching anything", 14)

        S.click(*MODEL_CHEV, 0.9)
        beat("the model list", 15)
        S.key('escape')
        time.sleep(1.2)
        beat("closed", 5)
    finally:
        r.stop()
