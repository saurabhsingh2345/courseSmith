#!/usr/bin/env python3
"""L16 take D — raw markdown beside rendered markdown, which is the shot that
makes "this is how the agent reads it" land. Driven through the command palette
rather than a context menu, because a blind right-click menu is not reliable."""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Cursor",)


def d_preview():
    """Cursor overrides Cmd+K with its own inline edit box, so the chord does not
    work — the editor-toolbar button does. Tooltip confirms it is
    "Open Preview to the Side"."""
    S.move(700, 400, 0.8); time.sleep(1.2)
    S.move(1411, 52, 1.1); time.sleep(1.8)        # hover, so the tooltip reads
    S.click(1411, 52, 0.3); time.sleep(4.0)       # raw left, rendered right
    S.move(1150, 450, 1.0); time.sleep(2.5)
    for _ in range(4):
        S.scroll(-220); time.sleep(1.6)
    time.sleep(2.5)


if __name__ == "__main__":
    S.wait_front("Cursor"); S.clear_mods()
    r = S.Rec("L16_D_preview", guard=GUARD)
    r.start(settle=2.0)
    try:
        d_preview()
    finally:
        r.stop()
