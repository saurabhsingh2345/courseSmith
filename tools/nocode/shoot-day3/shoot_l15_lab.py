#!/usr/bin/env python3
"""L15 — one long continuous lab take, which is how his lecture actually runs:
terminal open, node check, the pwd/cd primer, the real clone, into it, ls.
Deliberate teaching pace with real pauses, so slides can be cut anywhere."""
import os, sys, time, subprocess
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Cursor",)
HERE = os.path.dirname(os.path.abspath(__file__))
TERM = (600, 841)


def setup():
    subprocess.run(["osascript", "-e", 'tell application "Cursor" to quit'],
                   capture_output=True)
    time.sleep(5)
    subprocess.Popen(["open", "-a", "Cursor", "/Users/Shared/projects"])
    time.sleep(14)
    S.wait_front("Cursor"); time.sleep(2)
    S.key('return'); time.sleep(2)
    S.clear_mods()
    S.move_window_to_panel("Cursor")
    time.sleep(2)
    S.click(1595, 53, 1.0); time.sleep(2.0)    # close the agent panel — the
    S.clear_mods()                             # terminal gets the full width
    S.key('j', ('cmd',)); time.sleep(3.0)
    S.clear_mods()
    S.click(*TERM, 0.9); time.sleep(1.5)
    S.type_text("PROMPT='%~ %# '", cps=30); time.sleep(0.5)
    S.key('return'); time.sleep(1.3)
    S.type_text("clear && printf '\033[3J'", cps=30); time.sleep(0.5)
    S.key('return'); time.sleep(2.0)
    S.clear_mods()
    return S.snap(os.path.join(HERE, "pre_lab.png"))


def cmd(text, cps=10, wait=3.5, pause=1.4):
    time.sleep(pause)
    S.type_text(text, cps=cps)
    time.sleep(1.0)
    S.key('return')
    time.sleep(wait)


def lab():
    S.click(*TERM, 0.8); time.sleep(2.0)
    cmd("node --version", cps=9, wait=5.0, pause=2.0)
    cmd("pwd", cps=6, wait=5.0, pause=2.5)
    cmd("cd ..", cps=5, wait=3.5, pause=2.0)
    cmd("pwd", cps=6, wait=4.5, pause=1.6)
    cmd("cd projects", cps=9, wait=3.5, pause=1.8)
    cmd("git clone https://github.com/saurabhsingh2345/kanban.git",
        cps=15, wait=9.0, pause=2.5)
    cmd("cd kanban", cps=8, wait=3.5, pause=2.0)
    cmd("pwd", cps=6, wait=4.0, pause=1.5)
    cmd("ls", cps=4, wait=5.5, pause=1.8)
    time.sleep(3.0)


if __name__ == "__main__":
    print("pre-roll:", setup(), flush=True)
    r = S.Rec("L15_E_lab", guard=GUARD)
    r.start(settle=2.5)
    try:
        lab()
    finally:
        r.stop()
