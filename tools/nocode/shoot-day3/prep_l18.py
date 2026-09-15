#!/usr/bin/env python3
"""L18 prep: stage VS Code on the neutral kanban-copilot folder, fullscreen on
the shoot panel. Workspace trust is disabled on the command line so the trust
page (which shows the full path) is never on camera."""
import os, sys, time, subprocess
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

PROJ = "/Users/Shared/projects/kanban-copilot"
OUT = os.path.dirname(os.path.abspath(__file__))

S.desktop_quiet(True)
S.osa('tell application "System Events" to set visible of every process '
      'whose name is not "Finder" to false')
time.sleep(1.5)
S.close_finder_windows()

subprocess.run(["osascript", "-e", 'tell application "Visual Studio Code" to quit'],
               capture_output=True)
time.sleep(5)
subprocess.Popen(["/usr/local/bin/code" if os.path.exists("/usr/local/bin/code")
                  else "code", "--disable-workspace-trust", "-n", PROJ])
time.sleep(16)
S.wait_front("Code")
time.sleep(2)
S.clear_mods()

S.move_window_to_panel("Code")
time.sleep(2)
S.move(960, 620, 0.4)
time.sleep(1.5)
print("frontmost:", S.frontmost())
print("snapshot:", S.snap(os.path.join(OUT, "l18_prep.png")))
