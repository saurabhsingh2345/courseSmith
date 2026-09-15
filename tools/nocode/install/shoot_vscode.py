#!/usr/bin/env python3
"""The VS Code installation take. One continuous recording, beats marked live.

Coordinates are in PANEL POINTS (1920x1080) and were all measured off a
rehearsal pass, not guessed. Each beat sits on screen for slightly longer than
its narration needs, so assemble.py always speeds footage up a little and never
freezes on a last frame.

Run it and leave the machine alone. The guard kills the take if anything other
than Chrome, Finder or the Dock comes to the front.
"""
import os, subprocess, sys, time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S            # noqa: E402
import drive                 # noqa: E402
from take import Take        # noqa: E402

PROFILE = os.path.join(HERE, ".shootprofile")
CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"

# --- measured coordinates -----------------------------------------------------
ADDR      = (600, 63)         # the pinned toolbar's address bar
HERO      = (960, 447)        # "Download for macOS" on the landing page
NAV_DL    = (1665, 126)       # the blue Download button in the navigation bar
WIN_BTN   = (622, 530); WIN_X64 = (648, 597)
DEB_BTN   = (840, 530); RPM_BTN = (1000, 530)
MAC_BTN   = (1267, 530)
INTEL     = (1213, 597); ARM = (1287, 597); UNIV = (1360, 597)
DL_ICON   = (1824, 63)        # the downloads icon in the toolbar
DL_ITEM   = (1669, 115)       # the item in the bubble that opens itself
DL_ROW    = (1640, 155)       # the item in the re-opened history panel
APP_ICON  = (220, 521)        # Visual Studio Code, inside the disk image window
APPS_DIR  = (459, 518)        # the Applications shortcut beside it
DMG_TITLE = (340, 340)        # the disk image window's title bar
VSC_IN_APPS = (1341, 694)     # where the copy lands in the Applications window


def prep():
    """Blank tab, fullscreen on the panel, empty download list."""
    subprocess.run(["pkill", "-f", f"user-data-dir={PROFILE}"], capture_output=True)
    time.sleep(2.5)
    subprocess.Popen([CHROME, "--new-window", f"--user-data-dir={PROFILE}",
                      "--no-first-run", "--no-default-browser-check",
                      "--disable-features=Translate", "--hide-crash-restore-bubble",
                      "--window-position=0,0", "--window-size=1920,1080", "about:blank"],
                     stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    time.sleep(8)
    S.wait_front("Google Chrome")
    S.key('f', ('cmd', 'ctrl')); time.sleep(4.5); S.clear_mods()
    S.osa('tell application "System Events" to tell process "Google Chrome" to '
          'click menu item "Always Show Toolbar in Full Screen" of menu "View" of menu bar 1')
    time.sleep(1.5); S.clear_mods()
    S.move(960, 700, 0.6)
    print("prepped:", S.osa('tell application "Google Chrome" to get bounds of window 1'))


def main():
    prep()
    t = Take("vscode", allow=("Google Chrome", "Finder", "Dock"))
    t.start(settle=3.0)

    # 1. the empty browser, and the address typed in by hand
    t.hold(7, "address")
    S.click(*ADDR, 1.0); time.sleep(2.0)
    S.type_text("code.visualstudio.com", cps=5); time.sleep(2.5)
    S.key('return'); time.sleep(6)

    # 2. the landing page: read it, then head for the full download page
    t.mark("landing")
    S.move(*HERO, 1.4); time.sleep(5)
    for _ in range(3):
        S.scroll(-650, steps=22); time.sleep(4.5)
    time.sleep(3)
    for _ in range(3):
        S.scroll(650, steps=22); time.sleep(2.5)
    time.sleep(3)
    S.move(*HERO, 1.0); time.sleep(5)
    S.move(*NAV_DL, 1.4); time.sleep(4)

    # 3. every platform side by side
    t.mark("downloadpage")
    S.click(*NAV_DL, 0.6); time.sleep(6)
    S.move(*WIN_BTN, 1.6); time.sleep(5)
    S.move(*WIN_X64, 1.0); time.sleep(5)
    S.move(*DEB_BTN, 1.6); time.sleep(4)
    S.move(*RPM_BTN, 1.0); time.sleep(4)
    S.move(*MAC_BTN, 1.8); time.sleep(5)

    # 4. the one choice worth understanding
    t.hold(5, "chip")
    S.move(*INTEL, 1.2); time.sleep(7)
    S.move(*ARM, 1.0); time.sleep(7)
    S.move(*UNIV, 1.0); time.sleep(7)
    S.move(*ARM, 1.0); time.sleep(6)

    # 5. the download itself
    t.mark("download")
    S.click(*ARM, 0.5); time.sleep(13)
    S.move(*DL_ITEM, 1.4); time.sleep(7)
    S.move(960, 420, 1.6); time.sleep(6)
    S.move(*DL_ICON, 1.4); time.sleep(5)

    # 6. what a disk image actually is
    t.mark("mount")
    S.click(*DL_ICON, 0.5); time.sleep(2.5)
    S.click(*DL_ROW, 0.9); time.sleep(10)
    S.move(*APP_ICON, 1.6); time.sleep(5)
    S.move(*APPS_DIR, 1.2); time.sleep(4)

    # 7. the install: one drag
    t.mark("drag")
    S.move(*APP_ICON, 1.2); time.sleep(6)
    S.drag(*APP_ICON, *APPS_DIR, glide=3.0, hold=0.9, settle=1.0)
    time.sleep(20)

    # 8. the copy landing in the real Applications folder
    t.mark("landed")
    S.move(*VSC_IN_APPS, 1.8); time.sleep(9)
    S.move(1200, 640, 1.4); time.sleep(6)
    S.move(*VSC_IN_APPS, 1.2); time.sleep(9)

    # 9. eject the delivery box
    t.mark("eject")
    S.move(*DMG_TITLE, 1.8); time.sleep(3)
    S.click(*DMG_TITLE, 0.4); time.sleep(2.5)
    S.key('e', ('cmd',)); time.sleep(1.0); S.clear_mods(); time.sleep(4)
    S.move(1100, 560, 1.6); time.sleep(9)

    # 10. and it is installed
    t.mark("closing")
    S.move(*VSC_IN_APPS, 1.6); time.sleep(10)
    S.move(1050, 500, 2.0); time.sleep(8)
    S.move(*VSC_IN_APPS, 1.8); time.sleep(12)

    ok = t.stop()
    print("TAKE OK" if ok else "TAKE FAILED")


if __name__ == "__main__":
    main()
