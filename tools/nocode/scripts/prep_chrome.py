import sys, os, time, subprocess
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

HERE = os.path.dirname(os.path.abspath(__file__))
prof = os.path.join(HERE, ".shootprofile")
subprocess.run(["pkill", "-f", f"user-data-dir={prof}"], capture_output=True)
time.sleep(2)
S.desktop_quiet(True)
subprocess.Popen(["/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
                  "--new-window", f"--user-data-dir={prof}",
                  "--no-first-run", "--no-default-browser-check",
                  "--disable-features=Translate", "--hide-crash-restore-bubble",
                  "about:blank"],
                 stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
time.sleep(8)
ok = S.wait_front("Google Chrome")
print("chrome frontmost:", ok)
print("bounds:", S.osa('tell application "Google Chrome" to get bounds of window 1'))
S.key('f', ('cmd', 'ctrl'))
time.sleep(4)
S.osa('tell application "System Events" to tell process "Google Chrome" to '
      'click menu item "Always Show Toolbar in Full Screen" of menu "View" of menu bar 1')
time.sleep(1.5)
S.move(960, 620, 0.5)
time.sleep(1.5)
print("frontmost now:", S.frontmost())
print("bounds now:", S.osa('tell application "Google Chrome" to get bounds of window 1'))
S.snap(os.path.join(HERE, "takes", "preflight.png"))
