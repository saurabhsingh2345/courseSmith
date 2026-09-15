#!/usr/bin/env python3
"""L21 take A — working in a loop, with real feedback on a real flaw.

ref-L21 opens by finishing L20's thread: its Add Card was a bare browser prompt,
so he sends feedback and gets a proper modal. Ours already built a proper modal
with a title and an optional description, so that thread does not exist for us.

But there IS a genuine flaw, and it is a better one to teach on because it is the
kind you catch by looking rather than by clicking: the first column has no left
padding so its cards touch the window edge, and the five columns do not fill the
viewport width, leaving dead space on the right. That reads as unfinished.

The feedback below is phrased in the three parts W3 names on screen — the exact
thing that is wrong, why it is wrong, and the standard being held to — so the
graphic and the footage are describing the same sentence.
"""
import os, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S
import approve as A

PROJ = "/Users/Shared/projects/kanban"
ALLOW = ("Antigravity IDE", "Electron")
PROMPT = (1500, 1000)

FEEDBACK = ("the board works but the layout is not finished. the first column has no "
            "left padding so its cards touch the edge of the window, and the five "
            "columns do not fill the width, so there is dead space on the right. i "
            "would not show this to anyone in this state. please make the columns fill "
            "the available width evenly, with the same padding on both edges.")

if __name__ == "__main__":
    S.osa('tell application "Antigravity IDE" to activate'); time.sleep(2.0)
    S.clear_mods()
    print("frontmost:", S.frontmost_fast(), flush=True)
    A.run_build(S, "L21_A_iterate", PROJ, ALLOW, PROMPT, FEEDBACK,
                HERE, quiet=240.0, cap=1800.0, approve=False, cps=13)
