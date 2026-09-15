import sys, os, time, subprocess
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

S.osa('tell application "System Events" to set visible of every process whose name is not "Finder" to false')
time.sleep(1.5)
S.close_finder_windows()
subprocess.run(["osascript", "-e", 'tell application "Cursor" to quit'], capture_output=True)
time.sleep(4)
subprocess.Popen(["open", "-a", "Cursor"])
time.sleep(12)
print("front:", S.frontmost())
S.wait_front("Cursor")
S.key('n', ('cmd', 'shift'))       # New Window -> the welcome screen
time.sleep(4)
S.key('f', ('cmd', 'ctrl'))
time.sleep(4)
S.move(960, 640, 0.5)
time.sleep(2)
print("front now:", S.frontmost())
S.snap(os.path.join(os.path.dirname(os.path.abspath(__file__)), "takes", "cursor_pre.png"))
