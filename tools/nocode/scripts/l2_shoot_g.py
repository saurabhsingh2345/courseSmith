import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

S.osa('tell application "Google Chrome" to activate')
time.sleep(2.5)
assert S.wait_front("Google Chrome")
S.move(960, 950, 0.4); time.sleep(1.0)

r = S.Rec("L2G_test2", ("Google Chrome", "Chrome"))
r.start(settle=2.0)
time.sleep(1.8)
S.key('r', ('cmd',)); time.sleep(6.5)
S.move(1290, 705, 1.3); time.sleep(2.0)      # difficulty row
S.move(960, 950, 0.7); time.sleep(0.8)
S.key('return'); time.sleep(4.5)
seq = [('right',0.5),('space',0.2),('up',0.9),('space',0.2),('left',0.6),
       ('space',0.2),('up',0.8),('space',0.2),('right',0.7),('space',0.2),
       ('up',1.0),('space',0.2),('left',0.5),('space',0.2),('space',0.2)]
t0 = time.time(); i = 0
while time.time() - t0 < 32:
    k, h = seq[i % len(seq)]
    S.drive.key(k); time.sleep(h); i += 1
    if i % 24 == 0:
        S.drive.key('r'); time.sleep(0.4)
time.sleep(3.0)
r.stop()
