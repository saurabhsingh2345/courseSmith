#!/usr/bin/env python3
"""L15 take D — the payoff shot: the cloned project open in Cursor, KANBAN in
the sidebar, agents.md sitting there as the only thing in it."""
import os, sys, time, subprocess
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Cursor",)

subprocess.run(["osascript", "-e", 'tell application "Cursor" to quit'], capture_output=True)
time.sleep(5)
subprocess.Popen(["open", "-a", "Cursor", "/Users/Shared/projects/kanban"])
time.sleep(14)
S.wait_front("Cursor"); time.sleep(2)
S.key('return'); time.sleep(2)
S.clear_mods()
S.move_window_to_panel("Cursor")
time.sleep(2)
S.move(960, 620, 0.4); time.sleep(1.5)
print("pre-roll:", S.snap(os.path.join(os.path.dirname(os.path.abspath(__file__)), "pre_l15d.png")))

r = S.Rec("L15_D_open", guard=GUARD)
r.start(settle=2.0)
try:
    S.move(120, 90, 1.2); time.sleep(2.5)      # the KANBAN root in the tree
    S.move(67, 103, 0.9); time.sleep(2.0)      # agents.md
    S.click(67, 103, 0.5); time.sleep(3.5)     # open it
    S.move(700, 300, 1.0); time.sleep(3.0)
    time.sleep(2.0)
finally:
    r.stop()
