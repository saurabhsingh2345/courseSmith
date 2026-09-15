import sys, os, time, subprocess
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S
import Quartz as Q

S.osa('tell application "System Events" to set visible of every process whose name is not "Finder" to false')
time.sleep(2)
subprocess.run(["open", os.path.expanduser("~/Downloads/Cursor-darwin-arm64.dmg")])
time.sleep(14)
S.wait_front("Finder")
S.osa('''tell application "Finder"
    set b to bounds of front window
end tell''')
# park the disk-image window mid-panel
S.osa('tell application "Finder" to set bounds of front window to {430, 240, 1490, 836}')
time.sleep(2.0)
S.move(960, 900, 0.6)
time.sleep(1.5)
print("front:", S.frontmost())
S.snap(os.path.join(os.path.dirname(os.path.abspath(__file__)), "takes", "E_pre.png"))

r = S.Rec("E_drag", ("Finder",))
r.start(settle=1.5)
time.sleep(2.0)
S.move(700, 545, 1.4)          # onto the Cursor icon
time.sleep(2.0)
# drag it toward Applications, then set it back down - we are not reinstalling
sx, sy = S.P(700, 545)
Q.CGEventPost(Q.kCGHIDEventTap, Q.CGEventCreateMouseEvent(None, Q.kCGEventLeftMouseDown, (sx, sy), Q.kCGMouseButtonLeft))
time.sleep(0.3)
for i in range(1, 40):
    t = i / 39
    x = sx + (S.P(1220, 545)[0] - sx) * t
    e = Q.CGEventCreateMouseEvent(None, Q.kCGEventLeftMouseDragged, (x, sy), Q.kCGMouseButtonLeft)
    Q.CGEventSetFlags(e, 0)
    Q.CGEventPost(Q.kCGHIDEventTap, e)
    time.sleep(0.045)
time.sleep(1.2)
for i in range(1, 30):
    t = i / 29
    x = S.P(1220, 545)[0] + (sx - S.P(1220, 545)[0]) * t
    e = Q.CGEventCreateMouseEvent(None, Q.kCGEventLeftMouseDragged, (x, sy), Q.kCGMouseButtonLeft)
    Q.CGEventSetFlags(e, 0)
    Q.CGEventPost(Q.kCGHIDEventTap, e)
    time.sleep(0.04)
Q.CGEventPost(Q.kCGHIDEventTap, Q.CGEventCreateMouseEvent(None, Q.kCGEventLeftMouseUp, (sx, sy), Q.kCGMouseButtonLeft))
time.sleep(3.0)
r.stop()
