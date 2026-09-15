import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

assert S.wait_front("Cursor")
S.move(960, 900, 0.4); time.sleep(1.0)
r = S.Rec("L2B_serve", ("Cursor",))
r.start(settle=2.0)
time.sleep(2.0)
S.move(90, 300, 1.2); time.sleep(3.0)          # the files on disk
S.move(60, 250, 1.0); time.sleep(2.5)
S.key('j', ('cmd',))                            # terminal
time.sleep(3.5)
S.move(700, 830, 1.2); time.sleep(2.0)
S.type_text("npm start", cps=8)
time.sleep(1.5)
S.key('return')
time.sleep(7.0)
S.move(700, 880, 1.0); time.sleep(4.0)
time.sleep(3.0)
r.stop()
S.snap(os.path.join(os.path.dirname(os.path.abspath(__file__)), "takes", "L2B_end.png"))
