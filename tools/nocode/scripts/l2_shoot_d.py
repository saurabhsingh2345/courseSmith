import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

P1 = "That's great. Please add some detail to the opponent so that it looks more like an enemy."

S.osa('tell application "Cursor" to activate')
time.sleep(2.5)
assert S.wait_front("Cursor"), "Cursor not frontmost"
S.move(960, 900, 0.4); time.sleep(1.0)

r = S.Rec("L2D_iter1", ("Cursor",))
r.start(settle=2.0)
time.sleep(2.5)
S.move(1550, 990, 1.4)          # the follow-up box
time.sleep(2.5)
S.click(1550, 990, 0.4)
time.sleep(1.8)
S.type_text(P1, cps=22)
time.sleep(2.5)
S.key('return')
time.sleep(3.0)
S.move(1500, 500, 1.4)
time.sleep(8.0)
for _ in range(17):
    time.sleep(10.0)
r.stop()
