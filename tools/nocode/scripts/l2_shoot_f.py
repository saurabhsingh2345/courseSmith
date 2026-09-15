import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

P2 = ("Please add a minimap in the corner so I can see where the bot is, "
      "and also make the difficulty harder.")

S.osa('tell application "Cursor" to activate')
time.sleep=getattr(time,'sleep')
time.sleep(2.5)
assert S.wait_front("Cursor"), "Cursor not frontmost"
S.move(960, 900, 0.4); time.sleep(1.0)

r = S.Rec("L2F_iter2", ("Cursor",))
r.start(settle=2.0)
time.sleep(2.5)
S.move(1550, 990, 1.4); time.sleep(2.5)
S.click(1550, 990, 0.4); time.sleep(1.8)
S.type_text(P2, cps=22)
time.sleep(2.5)
S.key('return')
time.sleep(3.0)
S.move(1500, 450, 1.4)
time.sleep(8.0)
for _ in range(20):
    time.sleep(10.0)
r.stop()
