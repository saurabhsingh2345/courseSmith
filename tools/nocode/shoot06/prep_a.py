import sys, os, time, subprocess
sys.path.insert(0, "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith/tools/nocode/scripts")
import stage as S
W = "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith/tools/nocode/shoot06"

S.desktop_quiet(True)
S.osa('tell application "System Events" to set visible of every process whose name is not "Finder" to false')
time.sleep(1.5)
S.close_finder_windows()

subprocess.run(["osascript","-e",'tell application "Cursor" to quit'], capture_output=True)
time.sleep(5)
subprocess.Popen(["/Applications/Cursor.app/Contents/Resources/app/bin/cursor",
                  os.path.expanduser("~/arena")],
                 stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
time.sleep(14)
S.wait_front("Cursor")
time.sleep(2)
S.key('f', ('cmd','ctrl'))          # fullscreen on the shoot panel
time.sleep(4.5)
S.clear_mods()
S.move(960, 640, 0.6)
time.sleep(2)
print("front:", S.frontmost())
S.snap(os.path.join(W, "pre_a.png"))
