#!/usr/bin/env python3
"""The Python installation film: three takes, no Finder and no file dialog.

The first attempt filmed Installer's Open dialog, which listed his real
Downloads folder — a payslip, a resume, client spreadsheets. So:

  * the shoot browser downloads into /Users/Shared/pyshoot/Downloads, which
    holds exactly one file, and nothing else is ever shown;
  * the package is opened between takes, so the film cuts from the download
    straight to the installer window, which is what he asked for;
  * his own Chrome windows are minimised — they are the same application as the
    shoot window, so the frontmost-app guard cannot tell them apart, and one of
    them was sitting on the panel taking the clicks.

Take A  browser, ending on the finished download
Take B  the installer's panes, ending with the cursor on Install
   (authentication happens here, off camera — macOS names the account)
Take C  the progress bar and the success pane
"""
import os, subprocess, sys, time, glob
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S
import drive
from take import Take

PROFILE = os.path.join(HERE, ".shootprofile")
CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
DLDIR = "/Users/Shared/pyshoot/Downloads"
PKG = os.path.join(DLDIR, "python-3.14.7-macos11.pkg")

ADDR = (600, 93); NAV_DL = (683, 305); DROP_MAC = (932, 784)
YELLOW = (585, 428); WIN_LINK = (859, 468); BUBBLE = (1668, 143)
WIN = None            # the shoot window's id, so his windows are never touched


def q(fmt):
    return S.osa(f'tell application "Google Chrome" to {fmt}')


def prep():
    for f in glob.glob(os.path.join(DLDIR, "*")):
        os.remove(f)
    # his windows out of the frame: same app, so the guard cannot see them
    S.osa('''tell application "System Events" to tell process "Google Chrome"
        repeat with w in windows
            try
                if name of w contains "Claude" then set value of attribute "AXMinimized" of w to true
            end try
        end repeat
    end tell''')
    subprocess.run(["pkill", "-f", f"user-data-dir={PROFILE}"], capture_output=True)
    time.sleep(3)
    subprocess.Popen([CHROME, "--new-window", f"--user-data-dir={PROFILE}",
                      "--no-first-run", "--no-default-browser-check",
                      "--disable-features=Translate", "--hide-crash-restore-bubble",
                      "--window-position=0,0", "--window-size=1920,1080", "about:blank"],
                     stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    time.sleep(9)
    global WIN
    n = int(q("get count of windows") or 0)
    for i in range(1, n + 1):
        if q(f"get URL of active tab of window {i}") == "about:blank":
            WIN = q(f"get id of window {i}"); break
    S.osa('tell application "Google Chrome" to activate'); time.sleep(1)
    q(f"set index of window id {WIN} to 1"); time.sleep(1)
    S.osa('tell application "System Events" to set visible of process "Code" to false')
    S.move(960, 700, 0.6)
    print(f"shoot window id={WIN} bounds={q(f'get bounds of window id {WIN}')}")


def take_a():
    t = Take("pyA", allow=("Google Chrome", "Finder", "Installer", "Dock"))
    t.start(settle=3.0)
    t.hold(5, "address")
    S.click(*ADDR, 1.0); time.sleep(0.8)
    S.key('a', ('cmd',)); S.clear_mods(); time.sleep(0.6)
    S.type_text("python.org", cps=5); time.sleep(2.0)
    S.key('return'); time.sleep(7)

    t.mark("homepage")
    S.move(960, 520, 1.4); time.sleep(4)
    S.move(*NAV_DL, 1.4); time.sleep(3)
    S.move(*DROP_MAC, 1.6); time.sleep(7)
    S.move(*NAV_DL, 1.2); time.sleep(3)

    t.mark("downloadspage")
    S.click(*NAV_DL, 0.5); time.sleep(6)
    S.move(*YELLOW, 1.4); time.sleep(6)
    S.scroll(-450, steps=20); time.sleep(6)
    S.move(960, 720, 1.6); time.sleep(5)
    S.scroll(450, steps=20); time.sleep(4)
    S.move(*WIN_LINK, 1.4); time.sleep(6)
    S.move(*YELLOW, 1.2); time.sleep(4)

    t.mark("download")
    S.click(*YELLOW, 0.5)
    for i in range(60):
        time.sleep(1)
        if os.path.exists(PKG) and not glob.glob(os.path.join(DLDIR, "*.crdownload")):
            print(f"  downloaded in ~{i}s"); break
    time.sleep(6)
    S.move(*BUBBLE, 1.4); time.sleep(8)
    S.move(1400, 500, 1.6); time.sleep(4)

    open_pkg()                                  # no Finder, no file dialog

    t.hold(4, "intro")
    S.move(760, 420, 1.6); time.sleep(6)
    S.move(900, 500, 1.2); time.sleep(4)
    press("Continue", wait=3.0)

    t.mark("readme")
    S.move(1000, 480, 1.4); time.sleep(5)
    S.scroll(-200, steps=14); time.sleep(6)
    S.scroll(-200, steps=14); time.sleep(6)
    press("Continue", wait=3.0)

    t.mark("license")
    S.move(1000, 470, 1.4); time.sleep(7)
    press("Continue", wait=2.5)
    press("Agree", sheet=True, wait=3.5)
    if not axc("Install"):
        press("Continue", wait=3.0)

    t.mark("installtype")
    S.move(950, 470, 1.6); time.sleep(6)
    c = axc("Customize")
    if c: S.move(*c, 1.4); time.sleep(6)
    inst = axc("Install")
    if inst: S.move(*inst, 1.6); time.sleep(9)
    t.stop()
    return inst


def axc(name, sheet=False):
    tgt = (f'button "{name}" of sheet 1 of window 1' if sheet
           else f'button "{name}" of window 1')
    p = S.osa(f'tell application "System Events" to tell process "Installer" to get position of {tgt}')
    z = S.osa(f'tell application "System Events" to tell process "Installer" to get size of {tgt}')
    if not p or not z or "," not in p:
        return None
    x, y = [float(v) for v in p.split(", ")]; w, h = [float(v) for v in z.split(", ")]
    return (x + w / 2, y + h / 2)


def press(name, sheet=False, glide=1.3, wait=2.5):
    c = axc(name, sheet) or axc(name, not sheet)
    if not c:
        print(f"  !! no button {name!r}"); return False
    S.click(*c, glide); time.sleep(wait); return True


def open_pkg():
    subprocess.run(["open", "-a", "Installer", PKG], capture_output=True)
    for _ in range(30):
        time.sleep(1)
        if S.osa('tell application "System Events" to tell process "Installer" to '
                 'get name of every window') == "Install Python":
            break
    # centre it on the panel
    S.osa('tell application "System Events" to tell process "Installer" to '
          'set position of window 1 to {650, 316}')
    time.sleep(1.5)
    print("  installer at", S.osa('tell application "System Events" to tell process "Installer" to get position of window 1'))


def wait_for_auth(inst):
    print("\n>>> CLICKING INSTALL — PLEASE AUTHENTICATE ON THE PANEL <<<\n", flush=True)
    S.clear_mods(); drive.click()
    time.sleep(2)
    seen = False
    for i in range(240):
        up = subprocess.run(["pgrep", "-x", "SecurityAgent"], capture_output=True).returncode == 0
        if up: seen = True
        elif seen:
            print(f"  authenticated after ~{i}s", flush=True); return True
        if os.path.exists("/Library/Frameworks/Python.framework/Versions/3.14"):
            print("  install underway", flush=True); return True
        time.sleep(1)
    return False


def take_c():
    t = Take("pyB", allow=("Installer", "Finder", "Dock"))
    t.start(settle=1.5)
    t.mark("progress")
    for i in range(240):
        if S.osa('tell application "System Events" to tell process "Installer" to '
                 'get (exists button "Close" of window 1)') == "true":
            print(f"  summary after ~{i}s", flush=True); break
        time.sleep(1)
    time.sleep(2)
    t.mark("summary")
    S.move(1000, 460, 1.8); time.sleep(9)
    S.move(860, 520, 1.4); time.sleep(8)
    S.move(1050, 430, 1.6); time.sleep(9)
    t.stop()


if __name__ == "__main__":
    prep()
    inst = take_a()
    if inst and wait_for_auth(inst):
        take_c()
    else:
        print("!! authentication did not happen; take C skipped")
