#!/usr/bin/env python3
"""L17 take C — the five checks, run live in Cursor's built-in browser.

Same five checks every build gets: board loads with data, drag between columns,
reorder inside a column, add and delete a card, and would you show it to anyone.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S
import Quartz as Q

GUARD = ("Cursor",)
L = Q.kCGMouseButtonLeft


def drag(x1, y1, x2, y2, hold=0.5, steps=26):
    """dnd-kit needs a press, a small jiggle past its activation threshold,
    then real movement — a straight teleport does not register."""
    S.move(x1, y1, 0.9); time.sleep(0.5)
    Q.CGEventPost(Q.kCGHIDEventTap,
        Q.CGEventCreateMouseEvent(None, Q.kCGEventLeftMouseDown, (x1, y1), L))
    time.sleep(hold)
    for d in (4, 9, 15):
        Q.CGEventPost(Q.kCGHIDEventTap,
            Q.CGEventCreateMouseEvent(None, Q.kCGEventLeftMouseDragged, (x1, y1 + d), L))
        time.sleep(0.08)
    for i in range(1, steps + 1):
        t = i / steps
        e = 4*t*t*t if t < 0.5 else 1 - pow(-2*t + 2, 3) / 2
        Q.CGEventPost(Q.kCGHIDEventTap, Q.CGEventCreateMouseEvent(
            None, Q.kCGEventLeftMouseDragged,
            (x1 + (x2 - x1) * e, y1 + (y2 - y1) * e), L))
        time.sleep(0.035)
    time.sleep(0.45)
    Q.CGEventPost(Q.kCGHIDEventTap,
        Q.CGEventCreateMouseEvent(None, Q.kCGEventLeftMouseUp, (x2, y2), L))
    time.sleep(1.4)


if __name__ == "__main__":
    S.wait_front("Cursor"); S.clear_mods()
    S.click(1648, 53, 0.8); time.sleep(2.5)      # close the agent panel — all 5 columns
    S.move(900, 600, 0.6); time.sleep(1.5)
    r = S.Rec("L17_C_checks", guard=GUARD)
    r.start(settle=2.5)
    try:
        # 1. the board, loaded, with data already in it
        S.move(900, 300, 1.2); time.sleep(4.0)
        S.move(1500, 400, 1.2); time.sleep(3.5)
        # 2. drag between columns — Backlog -> Ready
        drag(560, 400, 900, 480)
        time.sleep(3.0)
        # 3. reorder inside a column — In Progress
        drag(1240, 400, 1240, 620)
        time.sleep(3.0)
        # 4. add a card
        S.click(560, 1120, 1.0); time.sleep(1.2)
        S.type_text("Check the five checks", cps=13); time.sleep(1.0)
        S.click(560, 1175, 0.8); time.sleep(1.0)
        S.type_text("Every build gets the same list.", cps=15); time.sleep(1.2)
        S.move(460, 1232, 0.8); time.sleep(0.8)
        S.click(460, 1232, 0.3); time.sleep(3.0)
        # 5. delete a card
        S.move(660, 375, 1.0); time.sleep(1.5)
        S.click(660, 375, 0.3); time.sleep(3.0)
        S.move(900, 700, 1.0); time.sleep(3.5)
    finally:
        r.stop()
