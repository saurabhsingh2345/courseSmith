#!/usr/bin/env python3
"""L33 take D — going and looking, which is the half of part 7 that matters.

Part 7 wired the board to the database. The agent says it works and, unlike
L32, it says it checked in a real browser — and it left the evidence behind: a
card called "browser test card" that it created while verifying, sitting in the
backlog. That card is the proof, and the take holds on it.

Then the check it cannot do for me: drag a card, log out, log back in, and see
whether the move survived a session that no longer exists.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

CARD = (582, 384)          # "Rate limit the public API", in READY
DROP = (958, 560)          # IN PROGRESS, below the two cards already there
USER, PASS, GO = (958, 558), (958, 644), (958, 712)
OUT = (1861, 119)
EVIDENCE = (204, 574)      # the card it left behind


def beat(msg, hold):
    print(f"  {msg}  (hold {hold}s)", flush=True)
    time.sleep(hold)


if __name__ == "__main__":
    S.osa('tell application "Google Chrome" to activate'); time.sleep(1.5)
    S.wait_front("Google Chrome"); time.sleep(0.8); S.clear_mods()

    r = S.SegRec("L33_D_check", allow=("Google Chrome",))
    r.start(settle=2.5)
    try:
        beat("the board, now coming out of the database", 14)
        S.move(*EVIDENCE, 1.6)
        beat("the card it left behind while testing", 13)
        S.move(*CARD, 1.2)
        beat("about to move one", 5)
        S.drag(CARD[0], CARD[1], DROP[0], DROP[1], glide=2.0, hold=0.8)
        beat("dropped in progress", 14)
        S.click(*OUT, 1.2); time.sleep(2.5)
        beat("logged out — the session is gone", 10)
        S.key('r', ('cmd', 'shift')); time.sleep(5.0)
        beat("and a hard reload, so nothing is cached", 8)
        S.click(*USER, 0.9); time.sleep(0.4); S.type_text("admin", cps=7)
        S.click(*PASS, 0.9); time.sleep(0.4); S.type_text("admin123", cps=7)
        S.click(*GO, 1.0); time.sleep(4.0)
        beat("and the card is where I put it", 20)
        S.move(958, 480, 1.5)
        beat("final hold", 12)
    finally:
        r.stop()
