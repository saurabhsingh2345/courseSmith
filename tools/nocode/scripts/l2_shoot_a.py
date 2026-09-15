import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

assert S.wait_front("Cursor"), "Cursor not frontmost"
S.close_finder_windows()
S.move(960, 900, 0.4); time.sleep(1.0)

r = S.Rec("L2A_files", ("Cursor",))
r.start(settle=2.0)
time.sleep(2.5)
# the thread: what it did
S.move(1550, 300, 1.4); time.sleep(4.0)
S.drive.scroll(400, steps=18); time.sleep(3.0)      # back up the thread
S.move(1500, 200, 1.1); time.sleep(3.5)
S.drive.scroll(-300, steps=16); time.sleep(3.0)
# the file tree
S.move(60, 82, 1.4); time.sleep(3.5)
S.click(38, 126, 1.0); time.sleep(2.5)              # expand js/
S.move(90, 200, 1.1); time.sleep(3.0)
S.drive.scroll(-140, steps=12); time.sleep(2.5)
S.move(90, 300, 1.0); time.sleep(3.0)
S.click(66, 236, 1.1); time.sleep(3.0)              # open a js file
S.move(700, 400, 1.3); time.sleep(4.0)
S.drive.scroll(-320, steps=18); time.sleep(3.5)
S.click(38, 104, 1.2); time.sleep(2.5)              # css/
S.move(90, 150, 1.0); time.sleep(2.5)
time.sleep(3.0)
r.stop()
S.snap(os.path.join(os.path.dirname(os.path.abspath(__file__)), "takes", "L2A_end.png"))
