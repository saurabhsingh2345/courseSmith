import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

S.osa('tell application "System Events" to set visible of every process whose name is not "Google Chrome" to false')
time.sleep(2.0)
S.osa('tell application "Google Chrome" to activate')
time.sleep(2.5)
assert S.wait_front("Google Chrome")
S.key('l', ('cmd',)); time.sleep(1.0)
S.type_text("file:///Users/enfecsolutions/arena/index.html", cps=26)
time.sleep(0.8)
S.key('return'); time.sleep(8.0)
S.move(960, 950, 0.5); time.sleep(1.5)

r = S.Rec("L2I_arena", ("Google Chrome", "Chrome"))
r.start(settle=2.0)
time.sleep(3.5)
S.move(960, 400, 1.6); time.sleep(5.0)         # the ARENA wordmark
S.move(880, 640, 1.3); time.sleep(4.0)         # the control hints
S.move(958, 735, 1.3); time.sleep(3.0)         # ENTER ARENA
S.click(958, 735, 0.4)
time.sleep(5.0)
seq = [('right',0.5),('space',0.2),('up',0.9),('space',0.2),('left',0.6),
       ('space',0.2),('up',0.8),('space',0.2),('right',0.7),('space',0.2),
       ('up',1.0),('space',0.2),('left',0.5),('space',0.2),('space',0.2),
       ('up',0.7),('space',0.2),('right',0.6),('space',0.2),('space',0.2)]
t0 = time.time(); i = 0
while time.time() - t0 < 66:
    k, h = seq[i % len(seq)]
    S.drive.key(k); time.sleep(h); i += 1
time.sleep(5.0)
r.stop()
