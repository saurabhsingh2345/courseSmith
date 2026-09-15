#!/usr/bin/env python3
"""L30 take — the real clone, the project shape, and the brief.

PII: VS Code's Welcome tab lists Recent folders, and his include
`~/Desktop/enfec_subs` (his company) and `~/Desktop/self`. The tab is closed
before rolling. The activity-bar avatar is his photograph as usual and is masked
in the cut by cut18.py.

The clone is genuine — the local copy was deleted and the repo lives at the
neutral org, so the URL on camera carries no personal handle.
"""
import os, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S

ALLOW = ("Code",)
REPO = "https://github.com/heftyradius/pm.git"


def beat(msg, hold):
    print(f"  {msg}  (hold {hold}s)", flush=True)
    time.sleep(hold)


if __name__ == "__main__":
    # verify the WINDOW, not just the app — his courseSmith window took the
    # keystrokes last time because VS Code was frontmost but the wrong window was
    title = S.wait_front_window("Code", "projects")
    print("front window:", title, flush=True)
    time.sleep(1.0); S.clear_mods()
    # close the Welcome TAB (its Recent list shows his personal folders) without
    # risking closing the window: only press cmd+W if a tab is actually open
    if "welcome" in title.lower():
        S.key('w', ('cmd',))
        time.sleep(1.5)
        S.wait_front_window("Code", "projects")
    S.key('grave', ('ctrl',))     # integrated terminal
    time.sleep(3.0)
    S.wait_front_window("Code", "projects")   # still our window before typing

    r = S.SegRec("L30_A_clone", allow=ALLOW)
    r.start(settle=2.5)
    try:
        beat("open on an empty project folder", 6)
        S.type_text("cd /Users/Shared/projects", cps=16); time.sleep(0.6)
        S.key('return'); time.sleep(1.2)
        S.type_text("pwd", cps=14); time.sleep(0.5)
        S.key('return')
        beat("pwd — a neutral path, no username in it", 7)

        S.type_text(f"git clone {REPO}", cps=15); time.sleep(0.8)
        S.key('return')
        beat("the clone itself", 14)

        S.type_text("cd pm && ls -a", cps=16); time.sleep(0.6)
        S.key('return')
        beat("what came down", 12)
        S.type_text("cat .gitignore | head -4", cps=16); time.sleep(0.6)
        S.key('return')
        beat(".env is the first line of .gitignore", 12)
    finally:
        r.stop()
