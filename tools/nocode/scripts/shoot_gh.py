import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

assert S.wait_front("Cursor"), "Cursor not frontmost"
S.move(960, 700, 0.5); time.sleep(1.0)

r = S.Rec("GH_cursor", ("Cursor",))
r.start(settle=2.0)

# ---- G : the welcome screen (~29s) ----------------------------------------
time.sleep(3.0)
S.move(818, 425, 1.5)          # the Cursor mark
time.sleep(4.0)
S.move(838, 452, 1.0)          # Team / Settings line
time.sleep(4.0)
S.move(855, 512, 1.3)          # Open project
time.sleep(3.5)
S.move(1063, 512, 1.1)         # Clone repo
time.sleep(3.0)
S.move(855, 585, 1.1)          # Connect via SSH
time.sleep(3.0)
S.move(770, 672, 1.2)          # recent projects
time.sleep(3.5)
S.move(855, 512, 1.2)          # back to Open project
time.sleep(3.0)

# ---- H : pick a folder, make Instant (~50s) -------------------------------
S.click(855, 512, 0.3)
time.sleep(4.0)
S.key('g', ('cmd', 'shift'))   # go straight to the projects folder
time.sleep(2.0)
S.type_text("~/projects", cps=12)
time.sleep(1.5)
S.key('return')
time.sleep(4.0)
S.snap(os.path.join(os.path.dirname(os.path.abspath(__file__)), "takes", "H_dialog.png"))
time.sleep(4.0)
r.stop()
