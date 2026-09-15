#!/usr/bin/env python3
"""Put the machine into shoot condition for the install films.

    python3 vs_stage.py

Everything here is undone by `state.py restore`. Read that first if a take is
abandoned half way.
"""
import os, subprocess, sys, time

HERE = os.path.dirname(os.path.abspath(__file__))
SCRIPTS = os.path.join(HERE, "..", "scripts")
PROFILE = os.path.join(HERE, ".shootprofile")
CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"


def osa(s):
    return subprocess.run(["osascript", "-e", s], capture_output=True, text=True).stdout.strip()


# 1. the 16:9 panel becomes primary, so Chrome cannot clamp its window back onto
#    the built-in and fullscreen lands where we are pointing the camera
print(subprocess.run([sys.executable, os.path.join(SCRIPTS, "arrange.py"), "shoot"],
                     capture_output=True, text=True).stdout)
time.sleep(2)

sys.path.insert(0, SCRIPTS)
import stage as S  # noqa: E402  (probes the capture device — never while rolling)

# 2. his desktop is covered in client folders; hide the icons and the dock
S.desktop_quiet(True)

# 3. and hide every other app's windows, so nothing of his is behind a
#    non-fullscreen Finder window on the panel
osa('''tell application "System Events"
        repeat with p in (every process whose name is not "Finder" and ¬
                         name is not "Google Chrome" and visible is true)
            try
                set visible of p to false
            end try
        end repeat
    end tell''')

# 4. a brand new Chrome profile: no tabs of his, no bookmarks bar, no extensions,
#    no profile avatar, and an empty download list
subprocess.run(["pkill", "-f", f"user-data-dir={PROFILE}"], capture_output=True)
time.sleep(1.5)
subprocess.Popen([CHROME, "--new-window", f"--user-data-dir={PROFILE}",
                  "--no-first-run", "--no-default-browser-check",
                  "--disable-features=Translate", "--hide-crash-restore-bubble",
                  "--start-maximized", "about:blank"],
                 stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
time.sleep(9)
print("chrome frontmost:", S.wait_front("Google Chrome"))
S.key('f', ('cmd', 'ctrl'))     # fullscreen on the panel
time.sleep(4)
# pin the toolbar, or it slides in and out mid-shot
osa('tell application "System Events" to tell process "Google Chrome" to '
    'click menu item "Always Show Toolbar in Full Screen" of menu "View" of menu bar 1')
time.sleep(1.5)
S.clear_mods()
S.move(960, 600, 0.5)
print("frontmost:", S.frontmost())
print("bounds:", osa('tell application "Google Chrome" to get bounds of window 1'))
S.snap(os.path.join(HERE, "takes", "stage.png"))
print("staged")
