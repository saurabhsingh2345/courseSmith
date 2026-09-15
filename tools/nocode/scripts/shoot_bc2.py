import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

assert S.wait_front("Google Chrome"), "Chrome is not frontmost - re-run prep_chrome.py"
S.close_finder_windows()
S.move(960, 620, 0.4)
time.sleep(1.0)

r = S.Rec("BC_browser", ("Google Chrome", "Chrome"))
r.start(settle=2.0)

# ---- B : bring up a browser, go to cursor.com  (~26s) ----------------------
time.sleep(2.0)
S.move(187, 64, 1.1)
time.sleep(0.9)
S.key('l', ('cmd',))
time.sleep(0.9)
S.type_text("cursor.com", cps=7)
time.sleep(1.5)
S.key('return')
time.sleep(6.5)
S.move(486, 285, 1.7)
time.sleep(4.5)
S.move(760, 330, 1.3)
time.sleep(3.2)

# ---- C : the download page  (~40s) ----------------------------------------
S.move(1565, 114, 1.6)
time.sleep(4.5)
S.move(1200, 210, 1.1)
time.sleep(2.5)
S.move(417, 360, 1.7)
time.sleep(5.5)
S.click(417, 360, 0.35)
time.sleep(8.0)
S.move(1000, 300, 1.3)
time.sleep(6.5)
S.drive.scroll(-240, steps=24)
time.sleep(5.0)
time.sleep(3.0)

r.stop()
