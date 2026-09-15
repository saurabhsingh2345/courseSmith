#!/usr/bin/env python3
"""L17 take D — the interactions, with coordinates read off a 1:1 snapshot.

First attempt grabbed the description text and selected it instead of dragging
the card. Grab the TITLE line, and give dnd-kit its activation jiggle.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S
import Quartz as Q

GUARD = ("Cursor",); L = Q.kCGMouseButtonLeft


def drag(x1, y1, x2, y2, hold=0.55, steps=30):
    S.move(x1, y1, 1.0); time.sleep(0.6)
    Q.CGEventPost(Q.kCGHIDEventTap,
        Q.CGEventCreateMouseEvent(None, Q.kCGEventLeftMouseDown, (x1, y1), L))
    time.sleep(hold)
    for d in (3, 7, 12, 18):
        Q.CGEventPost(Q.kCGHIDEventTap, Q.CGEventCreateMouseEvent(
            None, Q.kCGEventLeftMouseDragged, (x1, y1 + d), L)); time.sleep(0.09)
    for i in range(1, steps + 1):
        t = i / steps
        e = 4*t*t*t if t < 0.5 else 1 - pow(-2*t + 2, 3) / 2
        Q.CGEventPost(Q.kCGHIDEventTap, Q.CGEventCreateMouseEvent(
            None, Q.kCGEventLeftMouseDragged,
            (x1 + (x2 - x1) * e, y1 + (y2 - y1) * e), L)); time.sleep(0.04)
    time.sleep(0.6)
    Q.CGEventPost(Q.kCGHIDEventTap,
        Q.CGEventCreateMouseEvent(None, Q.kCGEventLeftMouseUp, (x2, y2), L))
    time.sleep(2.0)


if __name__ == "__main__":
    S.wait_front("Cursor"); S.clear_mods()
    r = S.Rec("L17_D_drag", guard=GUARD)
    r.start(settle=2.5)
    try:
        S.move(760, 400, 1.2); time.sleep(3.0)
        # check two — drag between columns, grabbing the TITLE line
        drag(443, 311, 794, 270)
        time.sleep(3.5)
        # check three — reorder inside In Progress
        drag(1110, 335, 1110, 480)
        time.sleep(3.5)
        # check four — delete
        S.move(582, 309, 1.2); time.sleep(1.8)
        S.click(582, 309, 0.3); time.sleep(3.5)
        S.move(900, 620, 1.0); time.sleep(3.0)
    finally:
        r.stop()
