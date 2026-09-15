import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

assert S.wait_front("Google Chrome")
S.move(960, 950, 0.4); time.sleep(1.0)

r = S.Rec("L2C_play", ("Google Chrome", "Chrome"))
r.start(settle=2.0)
time.sleep(2.0)
# the title screen, then the controls, then difficulty
S.move(600, 495, 1.5); time.sleep(4.0)
S.move(1290, 560, 1.4); time.sleep(4.0)        # controls panel
S.move(1163, 705, 1.2); time.sleep(2.5)        # ROOKIE
S.click(1163, 705, 0.3); time.sleep(2.5)
S.move(540, 788, 1.3); time.sleep(3.0)         # the enter button
S.move(960, 950, 0.8); time.sleep(1.0)
S.key('return'); time.sleep(4.0)               # drop in

def play(seconds):
    seq = [('right',0.55),('space',0.2),('up',0.9),('space',0.2),('space',0.2),
           ('left',0.7),('up',0.8),('space',0.2),('right',0.5),('space',0.2),
           ('up',1.1),('space',0.2),('left',0.6),('space',0.2),('space',0.2),
           ('right',0.8),('up',0.7),('space',0.2),('space',0.2),('down',0.5)]
    t0 = time.time()
    i = 0
    while time.time() - t0 < seconds:
        k, hold = seq[i % len(seq)]
        S.drive.key(k)
        time.sleep(hold)
        i += 1
        if i % 26 == 0:
            S.drive.key('r'); time.sleep(0.5)     # reload

play(38)
time.sleep(3.0)
# whatever the outcome, take it again
S.key('return'); time.sleep(3.0)
play(26)
time.sleep(4.0)
r.stop()
S.snap(os.path.join(os.path.dirname(os.path.abspath(__file__)), "takes", "L2C_end.png"))
