import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

S.osa('tell application "Cursor" to activate')
time.sleep(2.5)
assert S.wait_front("Cursor")
S.move(960, 900, 0.4); time.sleep(1.0)
r = S.Rec("L2H_wrap", ("Cursor",))
r.start(settle=2.0)
time.sleep(2.5)
S.move(1500, 400, 1.4); time.sleep(5.0)          # the thread
S.drive.scroll(500, steps=20); time.sleep(4.0)
S.move(1500, 250, 1.2); time.sleep(5.0)
S.drive.scroll(-500, steps=20); time.sleep(4.0)
S.move(90, 200, 1.4); time.sleep(5.0)            # the files it made
S.drive.scroll(-120, steps=10); time.sleep(4.0)
S.move(90, 400, 1.2); time.sleep(5.0)
S.move(60, 120, 1.2); time.sleep(5.0)
S.move(700, 400, 1.3); time.sleep(6.0)
time.sleep(6.0)
r.stop()
