#!/usr/bin/env python3
"""L16 take E — the long one. His L16 walks agents.md for ~7 minutes, so short
takes cannot carry it. This is a single slow pass with the preview open beside
the source, pausing on each section long enough to cut a slide out of it."""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Cursor",)


def reset():
    """settings tab away, agents.md open, preview to the side"""
    S.wait_front("Cursor"); S.clear_mods()
    S.key('w', ('cmd',)); time.sleep(1.5)          # close Cursor Settings
    S.click(67, 103, 1.0); time.sleep(2.5)         # agents.md
    S.move(700, 400, 0.6); time.sleep(0.8)
    S.click(1411, 52, 0.8); time.sleep(4.0)        # preview to the side
    S.move(700, 500, 0.8); time.sleep(1.0)
    S.key('home', ()) if False else None
    for _ in range(8):
        S.scroll(400); time.sleep(0.25)            # back to the top
    time.sleep(2.0)


def walk():
    """pause on each section, and point at the lines that matter"""
    holds = [
        # (scroll clicks before the hold, seconds to hold, where to park the mouse)
        (0,  14, (600, 200)),    # the h1 and Business requirements
        (2,  16, (600, 350)),    # one board / five columns / card fields
        (2,  15, (600, 300)),    # drag and drop, add, delete, nothing beyond
        (2,  14, (600, 320)),    # the priority: a good-looking interface
        (2,  15, (600, 300)),    # Technical details
        (2,  14, (600, 320)),    # Colour scheme
        (2,  16, (600, 300)),    # Strategy
        (2,  16, (600, 320)),    # Coding standards
        (2,  15, (600, 300)),    # Working agreement
    ]
    for clicks, hold, park in holds:
        for _ in range(clicks):
            S.scroll(-200); time.sleep(0.9)
        S.move(*park, 0.9)
        time.sleep(hold)


if __name__ == "__main__":
    reset()
    print("rolling long take", flush=True)
    r = S.Rec("L16_E_walk", guard=GUARD)
    r.start(settle=2.0)
    try:
        walk()
    finally:
        r.stop()
