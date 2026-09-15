#!/usr/bin/env python3
"""L32 take D — the sign-in, and being precise about what it proves.

Shot AFTER the blank-page fix, because before it this page rendered nothing at
all. The take deliberately does the round trip twice: sign in, look at the board,
log out, sign back in. What that demonstrates is that the SESSION survives, and
nothing more — there is no database until part 5, so the board itself is still
the one the front end draws for itself. The narration says so rather than letting
a working demo imply more than it shows.

Credentials are the MVP's hard-coded pair, printed on the form as placeholders.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

USER = (958, 558)
PASS = (958, 644)
GO   = (958, 712)
OUT  = (1861, 119)


def beat(msg, hold):
    print(f"  {msg}  (hold {hold}s)", flush=True)
    time.sleep(hold)


if __name__ == "__main__":
    S.osa('tell application "Google Chrome" to activate'); time.sleep(1.5)
    S.wait_front("Google Chrome"); time.sleep(0.8); S.clear_mods()
    # start from signed out, whatever state the probe left behind
    S.click(*OUT, 0.8); time.sleep(2.0)

    r = S.SegRec("L32_D_web", allow=("Google Chrome",))
    r.start(settle=2.5)
    try:
        beat("the sign-in screen", 10)
        S.click(*USER, 0.9); time.sleep(0.5); S.type_text("admin", cps=6)
        beat("username", 2)
        S.click(*PASS, 0.9); time.sleep(0.5); S.type_text("admin123", cps=6)
        beat("password", 2)
        S.click(*GO, 1.0); time.sleep(3.0)
        beat("and there is the board", 16)
        S.move(960, 500, 1.6)
        beat("across the columns", 10)
        S.click(*OUT, 1.2); time.sleep(2.5)
        beat("logged out, back to the form", 9)
        S.click(*USER, 0.9); time.sleep(0.4); S.type_text("admin", cps=6)
        S.click(*PASS, 0.9); time.sleep(0.4); S.type_text("admin123", cps=6)
        S.click(*GO, 1.0); time.sleep(3.0)
        beat("and back in, exactly where I was", 18)
        S.move(960, 620, 1.4)
        beat("final hold", 10)
    finally:
        r.stop()
