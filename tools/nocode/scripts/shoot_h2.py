import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

assert S.wait_front("Cursor"), "Cursor not frontmost"
r = S.Rec("H2_folder", ("Cursor",))
r.start(settle=2.0)
time.sleep(2.5)
S.move(900, 400, 1.3)          # over the projects contents
time.sleep(3.5)
S.move(756, 731, 1.4)          # New Folder
time.sleep(2.5)
S.click(756, 731, 0.3)
time.sleep(2.5)
S.type_text("Instant", cps=6)
time.sleep(2.0)
S.key('return')                # Create
time.sleep(3.5)
S.move(1341, 731, 1.4)         # Open
time.sleep(2.5)
S.click(1341, 731, 0.3)
time.sleep(10.0)               # the project opens
S.move(300, 300, 1.2)
time.sleep(4.0)
r.stop()
S.snap(os.path.join(os.path.dirname(os.path.abspath(__file__)), "takes", "J_pre.png"))
