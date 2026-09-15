#!/usr/bin/env python3
"""Day 3 prep: hide everything, open Cursor on the neutral kanban folder,
fullscreen it on the shoot panel, and snapshot for eyes-on verification."""
import os, sys, time, subprocess
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

PROJ = "/Users/Shared/projects/kanban"
OUT = os.path.dirname(os.path.abspath(__file__))

S.osa('tell application "System Events" to set visible of every process '
      'whose name is not "Finder" to false')
time.sleep(1.5)
S.close_finder_windows()

subprocess.run(["osascript", "-e", 'tell application "Cursor" to quit'], capture_output=True)
time.sleep(5)
subprocess.Popen(["open", "-a", "Cursor", PROJ])
time.sleep(14)
S.wait_front("Cursor")
time.sleep(2)

# a new folder raises the trust prompt; return accepts the default (Trust)
S.key('return')
time.sleep(2)
S.clear_mods()

S.move_window_to_panel("Cursor")
time.sleep(2)
S.move(960, 620, 0.4)
time.sleep(1.5)
print("frontmost:", S.frontmost())
print("snapshot:", S.snap(os.path.join(OUT, "verify.png")))
