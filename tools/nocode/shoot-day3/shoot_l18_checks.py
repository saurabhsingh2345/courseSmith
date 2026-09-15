#!/usr/bin/env python3
"""L18 take 3 — the five checks, run BY HAND on the finished Copilot build.

Why this take exists even though the agent already tested itself. It ran its
own Vitest suite and its own Playwright browser checks and reported all of them
passing, which is exactly the situation L17 warned about: the agent graded its
own homework. The lecture's claim is that you check it yourself, so the film has
to show someone checking it themselves.

All five mechanics were dry-run before rolling, so nothing here is a gamble:
drag between columns, reorder within a column, delete, rename a column, and add
a card all work. The board resets on reload because state is in memory, which is
what makes a clean take possible.

Holds are long on purpose. The house shape is a mean footage hold around 34
seconds, and a check that flicks past in two seconds teaches nothing.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

# board geometry at 1920x1080 points, fullscreen, fresh state
INBOX_CARD1 = (362, 528)        # Map the first release
PLAN_CARD1  = (661, 528)        # Write the product brief
INPROG_DROP = (959, 660)
REVIEW_CARD = (1258, 528)       # Polish the empty states
REVIEW_TRASH = (1366, 490)
INBOX_TITLE = (270, 441)
ADD_CARD    = (1641, 348)
F_TITLE     = (960, 515)
F_DESC      = (960, 583)
F_CREATE    = (1613, 655)


def beat(msg, hold):
    print(f"  {msg}  (hold {hold}s)", flush=True)
    time.sleep(hold)


if __name__ == "__main__":
    S.wait_front("Google Chrome"); time.sleep(0.8); S.clear_mods()
    S.key('r', ('cmd',))               # reset the board to seeded state
    time.sleep(4.0)
    S.move(960, 300, 0.6)

    r = S.SegRec("L18_S_checks", allow=("Google Chrome",))
    r.start(settle=2.5)
    try:
        beat("open on the finished board", 9)

        # 1 — drag between columns
        S.drag(*INBOX_CARD1, *INPROG_DROP, glide=1.8, hold=0.7)
        beat("check 1: drag between columns", 11)

        # 2 — reorder inside a column
        S.drag(*PLAN_CARD1, 661, 700, glide=1.4, hold=0.7)
        beat("check 2: reorder within a column", 11)

        # 3 — delete a card (hover reveals the trash)
        S.move(*REVIEW_CARD, 1.0); time.sleep(1.2)
        S.click(*REVIEW_TRASH, 0.9)
        beat("check 3: delete a card", 11)

        # 4 — rename a column
        S.click(*INBOX_TITLE, 0.9); time.sleep(1.0)
        S.key('a', ('cmd',)); time.sleep(0.5)
        S.type_text("Backlog", cps=7); time.sleep(1.0)
        S.key('return')
        beat("check 4: rename a column", 11)

        # 5 — add a card, with a description, which is the bit that matters
        S.click(*ADD_CARD, 0.9); time.sleep(1.6)
        S.click(*F_TITLE, 0.8); time.sleep(0.7)
        S.type_text("Check it myself", cps=8); time.sleep(0.8)
        S.click(*F_DESC, 0.8); time.sleep(0.7)
        S.type_text("A claim is not a demonstration.", cps=8); time.sleep(1.2)
        S.click(*F_CREATE, 0.9)
        beat("check 5: add a card with a description", 13)

        S.move(960, 700, 1.2)
        beat("hold on the finished board", 12)
    finally:
        r.stop()
