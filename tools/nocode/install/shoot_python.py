#!/usr/bin/env python3
"""The Python installation take: browser, then the .pkg installer.

Two takes with the authentication cut out between them. Take A ends with the
cursor resting on Install; the password prompt is macOS's, shows the account's
real name, and is never filmed. Take B picks up on the progress bar.

Installer buttons are located through the accessibility API and then clicked
with a real cursor glide, so the coordinates are exact and the pointer is still
visible doing the work.
"""
import os, subprocess, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S
from take import Take


ADDR = (600, 93)
NAV_DL = (683, 305)        # "Downloads" in python.org's navigation bar
DROP_MAC = (932, 784)      # the macOS 3.14.7 button inside that dropdown
YELLOW = (585, 428)        # "Download Python 3.14.7"
WIN_LINK = (859, 468)      # "Windows" in the different-OS line
DL_ICON = (1824, 93)
DL_ROW = (1668, 143)


def axc(name, sheet=False):
    """Centre of a named Installer button, in panel points."""
    tgt = (f'button "{name}" of sheet 1 of window 1' if sheet
           else f'button "{name}" of window 1')
    p = S.osa(f'tell application "System Events" to tell process "Installer" to get position of {tgt}')
    z = S.osa(f'tell application "System Events" to tell process "Installer" to get size of {tgt}')
    if not p or not z:
        return None
    x, y = [float(v) for v in p.split(", ")]
    w, h = [float(v) for v in z.split(", ")]
    return (x + w / 2, y + h / 2)


def has(name):
    r = S.osa('tell application "System Events" to tell process "Installer" to '
              f'get (exists button "{name}" of window 1)')
    return r == "true"


def buttons():
    return S.osa('tell application "System Events" to tell process "Installer" to '
                 'get name of every button of window 1')


def press(name, sheet=False, glide=1.2, wait=2.0):
    c = axc(name, sheet)
    if not c and not sheet:
        c = axc(name, True)          # some panes put it in a sheet
    if not c:
        print(f"  !! no button {name!r}; window has: {buttons()!r}"); return False
    S.click(*c, glide); time.sleep(wait)
    return True


PROFILE = os.path.join(HERE, ".shootprofile")
CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"


def prep():
    """A normal 1920x1080 window at the origin. NOT fullscreen: the fullscreen
    toggle landed inconsistently and shifted every measured coordinate."""
    subprocess.run(["pkill", "-f", f"user-data-dir={PROFILE}"], capture_output=True)
    time.sleep(3)
    import sqlite3
    db = os.path.join(PROFILE, "Default", "History")
    c = sqlite3.connect(db); c.executescript(
        "DELETE FROM downloads_url_chains; DELETE FROM downloads; DELETE FROM downloads_slices;")
    c.commit(); c.close()
    for f in ("python-3.14.7-macos11.pkg",):
        p = os.path.expanduser(f"~/Downloads/{f}")
        if os.path.exists(p): os.remove(p)
    subprocess.Popen([CHROME, "--new-window", f"--user-data-dir={PROFILE}",
                      "--no-first-run", "--no-default-browser-check",
                      "--disable-features=Translate", "--hide-crash-restore-bubble",
                      "--window-position=0,0", "--window-size=1920,1080", "about:blank"],
                     stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    time.sleep(8)
    S.wait_front("Google Chrome")
    S.osa('tell application "System Events" to set visible of process "Code" to false')
    S.move(960, 700, 0.6)
    print("prepped:", S.osa('tell application "Google Chrome" to get bounds of window 1'))


def take_a():
    prep()
    t = Take("pythonA", allow=("Google Chrome", "Finder", "Installer", "Dock"))
    t.start(settle=3.0)

    # 1. python.org, typed in
    t.hold(5, "address")
    S.click(*ADDR, 1.0); time.sleep(0.8)
    S.key('a', ('cmd',)); S.clear_mods(); time.sleep(0.6)
    S.type_text("python.org", cps=5); time.sleep(2.0)
    S.key('return'); time.sleep(7)

    # 2. the site, and the Downloads menu that already knows the platform
    t.mark("homepage")
    S.move(960, 500, 1.4); time.sleep(4)
    S.move(*NAV_DL, 1.4); time.sleep(3)
    S.move(*DROP_MAC, 1.6); time.sleep(7)
    S.move(*NAV_DL, 1.2); time.sleep(3)

    # 3. the download page and the release table
    t.mark("downloadspage")
    S.click(*NAV_DL, 0.5); time.sleep(6)
    S.move(*YELLOW, 1.4); time.sleep(6)
    S.scroll(-450, steps=20); time.sleep(6)
    S.move(960, 700, 1.6); time.sleep(5)
    S.scroll(450, steps=20); time.sleep(4)
    S.move(*WIN_LINK, 1.4); time.sleep(6)
    S.move(*YELLOW, 1.2); time.sleep(4)

    # 4. the .pkg
    t.mark("download")
    S.click(*YELLOW, 0.5); time.sleep(11)
    S.move(1668, 143, 1.4); time.sleep(6)
    S.move(*DL_ICON, 1.2); time.sleep(4)

    # 5. the installer opens
    t.mark("intro")
    S.click(*DL_ICON, 0.5); time.sleep(2.5)
    S.click(*DL_ROW, 0.9); time.sleep(9)
    S.wait_front("Installer", timeout=20)
    print("  installer front:", S.frontmost())
    S.move(760, 500, 1.6); time.sleep(5)      # the step list down the left
    press("Continue", wait=3.0)

    # 6. the Read Me nobody reads
    t.mark("readme")
    S.move(1100, 600, 1.4); time.sleep(4)
    S.scroll(-260, steps=16); time.sleep(6)
    S.scroll(-260, steps=16); time.sleep(6)
    press("Continue", wait=3.0)

    # 7. the licence, and the sheet that makes you mean it
    t.mark("license")
    S.move(1100, 560, 1.4); time.sleep(6)
    press("Continue", wait=2.5)
    press("Agree", sheet=True, wait=3.0)

    # some Macs put a Destination Select pane in here; walk past it
    if not has("Install"):
        press("Continue", wait=3.0)

    # 8. the last screen before it works
    t.mark("installtype")
    S.move(1000, 560, 1.6); time.sleep(6)
    c = axc("Customize")
    if c: S.move(*c, 1.4); time.sleep(6)
    inst = axc("Install")
    S.move(*inst, 1.6); time.sleep(8)
    print("  install button at", inst)
    t.stop()
    return inst


def wait_for_auth():
    """Click Install, then let him authenticate. The prompt is never filmed."""
    print("\n>>> CLICKING INSTALL — AUTHENTICATE WHEN macOS ASKS <<<\n", flush=True)
    S.clear_mods(); import drive; drive.click()
    time.sleep(2.5)
    seen = False
    for i in range(200):
        up = subprocess.run(["pgrep", "-x", "SecurityAgent"], capture_output=True).returncode == 0
        if up: seen = True
        elif seen:
            print(f"  authenticated after ~{i}s", flush=True); return True
        if os.path.exists("/Library/Frameworks/Python.framework/Versions/3.14"):
            print("  install underway", flush=True); return True
        time.sleep(1)
    print("  !! no authentication seen"); return False


def take_b():
    t = Take("pythonB", allow=("Installer", "Finder", "Dock"))
    t.start(settle=1.5)
    t.mark("progress")
    # sit on the progress bar until the summary pane turns up
    for i in range(180):
        if S.osa('tell application "System Events" to tell process "Installer" to '
                 'get (exists button "Close" of window 1)') == "true":
            print(f"  summary after ~{i}s", flush=True); break
        time.sleep(1)
    time.sleep(2)
    t.mark("summary")
    S.move(1100, 500, 1.8); time.sleep(8)
    # the installer opens the Python folder itself; look around it
    S.osa('tell application "Finder" to activate'); time.sleep(2)
    b = S.osa('tell application "Finder" to get bounds of front window')
    print("  python folder window:", b)
    if b:
        x1, y1, x2, y2 = [float(v) for v in b.split(", ")]
        S.move((x1 + x2) / 2, y1 + 90, 1.6); time.sleep(7)
        S.move(x1 + 120, y1 + 160, 1.4); time.sleep(7)
        S.move((x1 + x2) / 2 - 60, y1 + 160, 1.4); time.sleep(8)
    t.stop()


if __name__ == "__main__":
    take_a()
    wait_for_auth()
    take_b()
