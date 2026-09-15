#!/usr/bin/env python3
"""L21 take B — checking the fix, which is the half of the loop people skip.

The lecture's point is that feedback is only half of working in a loop: you send
it, and then you go and look. This take is the looking. It reloads the board and
holds on it long enough to compare against the frame from before the fix, then
runs two of the five checks again to show the fix did not break anything else —
which is the other thing people skip.
"""
import os, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S


def beat(msg, hold):
    print(f"  {msg}  (hold {hold}s)", flush=True)
    time.sleep(hold)


if __name__ == "__main__":
    S.osa('tell application "Google Chrome" to activate'); time.sleep(1.5)
    S.wait_front("Google Chrome"); time.sleep(0.6); S.clear_mods()
    S.key('r', ('cmd',)); time.sleep(5.0)
    S.move(960, 300, 0.6)

    r = S.SegRec("L21_B_verify", allow=("Google Chrome",))
    r.start(settle=2.5)
    try:
        beat("the board after the fix", 16)
        S.move(960, 600, 1.4)
        beat("holding on the full width", 12)
        # two checks again, to show the fix broke nothing
        S.drag(168, 185, 880, 500, glide=1.8, hold=0.7)
        beat("drag still works", 11)
        S.click(32, 144, 0.9); time.sleep(1.0)
        S.key('a', ('cmd',)); time.sleep(0.5)
        S.type_text("Shipped", cps=7); time.sleep(0.9)
        S.key('return')
        beat("rename still works", 11)
        S.move(960, 700, 1.2)
        beat("final hold", 10)
    finally:
        r.stop()
