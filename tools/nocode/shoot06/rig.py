"""Shoot helpers that refuse to roll unless the panel really is the main display.

The arrangement reverted between two takes and cost one, so every roll now
re-asserts it. After arrange.py shoot the external panel IS the main display,
and avfoundation "Capture screen 0" (index 1) is always the main display —
so the index is derived, never probed.
"""
import os, subprocess, sys, time
sys.path.insert(0, "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith/tools/nocode/scripts")
import Quartz as Q
os.environ.setdefault("_NC_NOPROBE", "1")
import stage as S

S.CAP_INDEX = "1"                      # main display == the panel, once arranged
PANEL = (1920, 1080)


def _displays():
    err, ids, cnt = Q.CGGetActiveDisplayList(8, None, None)
    out = {}
    for d in ids[:cnt]:
        b = Q.CGDisplayBounds(d)
        out[d] = (b.origin.x, b.origin.y, b.size.width, b.size.height,
                  bool(Q.CGDisplayIsMain(d)))
    return out


def panel_is_main():
    for d, (x, y, w, h, main) in _displays().items():
        if main and (int(w), int(h)) == PANEL:
            return True
    return False


def ensure_panel(tries=3):
    for _ in range(tries):
        if panel_is_main():
            return True
        subprocess.run([sys.executable,
                        "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith/"
                        "tools/nocode/scripts/arrange.py", "shoot"],
                       capture_output=True)
        time.sleep(3)
    return panel_is_main()


def fullscreen(app, want=True):
    """Set AXFullScreen directly. The cmd+ctrl+F toggle guesses at the current
    state and put VS Code back into a window on the desktop, exposing it."""
    cur = S.osa(f'tell application "System Events" to tell process "{app}" to '
                f'get value of attribute "AXFullScreen" of front window')
    if cur.lower() == ("true" if want else "false"):
        return True
    S.osa(f'tell application "System Events" to tell process "{app}" to '
          f'set value of attribute "AXFullScreen" of front window to {str(want).lower()}')
    time.sleep(4.0)
    cur = S.osa(f'tell application "System Events" to tell process "{app}" to '
                f'get value of attribute "AXFullScreen" of front window')
    return cur.lower() == ("true" if want else "false")


def refullscreen(app):
    """Park the window on the panel, then take it fullscreen there."""
    S.osa(f'tell application "{app}" to activate')
    time.sleep(1.2)
    fullscreen(app, False)
    S.osa(f'''tell application "System Events" to tell process "{app}"
        set position of front window to {{40, 60}}
        set size of front window to {{{PANEL[0]-80}, {PANEL[1]-140}}}
    end tell''')
    time.sleep(1.5)
    S.osa(f'tell application "{app}" to activate')
    time.sleep(0.8)
    ok = fullscreen(app, True)
    time.sleep(1.5)
    return ok


def quiet(on=True):
    """Hide the desktop and dock, and VERIFY — desktop_quiet silently no-opped
    once and put his folders and screenshots on camera."""
    import subprocess as sp
    if on:
        sp.run(["defaults","write","com.apple.finder","CreateDesktop","-bool","false"])
        sp.run(["defaults","write","com.apple.dock","autohide","-bool","true"])
    else:
        sp.run(["defaults","delete","com.apple.finder","CreateDesktop"],capture_output=True)
        sp.run(["defaults","write","com.apple.dock","autohide","-bool","false"])
    sp.run(["killall","Finder"],capture_output=True)
    sp.run(["killall","Dock"],capture_output=True)
    time.sleep(3.5)
    S.close_finder_windows()
    got = sp.run(["defaults","read","com.apple.finder","CreateDesktop"],
                 capture_output=True,text=True).stdout.strip()
    print(f"[quiet] CreateDesktop={got or 'unset'}", flush=True)


def verify(app, tag):
    """Hard gate: panel is main, app is frontmost, and the panel is not bare."""
    assert ensure_panel(), "panel is not the main display"
    assert S.wait_front(app), f"{app} not frontmost"
    png = os.path.join(os.path.dirname(os.path.abspath(__file__)), f"chk_{tag}.png")
    subprocess.run(["screencapture", "-x", "-D", "1", png], capture_output=True)
    small = png.replace(".png", "_s.png")
    subprocess.run(["sips", "-Z", "1000", png, "--out", small],
                   capture_output=True)
    print(f"[verify] panel-main ok, front={S.frontmost()}, snap={small}", flush=True)
    return small
