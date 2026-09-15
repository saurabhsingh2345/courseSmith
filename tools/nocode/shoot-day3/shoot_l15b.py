#!/usr/bin/env python3
"""L15 terminal takes, run only after the prompt has been neutralised and the
terminal focus VERIFIED by snapshot.

Hard lesson from the first attempt: typing into Cursor without confirming focus
sends the keystrokes to the AGENT, which queued the shell commands as tasks and
asked to edit the real ~/.zshrc. Always click into the terminal first.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Cursor",)
TERM = (600, 841)


def cmd(text, cps=11, wait=3.0):
    S.type_text(text, cps=cps)
    time.sleep(0.9)
    S.key('return')
    time.sleep(wait)


def t_primer():
    S.click(*TERM, 0.9); time.sleep(1.5)
    cmd("node --version", cps=10, wait=3.5)
    cmd("pwd", cps=7, wait=3.5)
    cmd("cd ..", cps=6, wait=2.5)
    cmd("pwd", cps=7, wait=3.5)
    cmd("cd projects", cps=10, wait=2.5)
    time.sleep(2.0)


def t_clone():
    S.click(*TERM, 0.9); time.sleep(1.2)
    cmd("git clone https://github.com/saurabhsingh2345/kanban.git", cps=16, wait=7.0)
    cmd("cd kanban", cps=9, wait=2.2)
    cmd("pwd", cps=7, wait=2.8)
    cmd("ls", cps=5, wait=3.5)
    time.sleep(2.0)


if __name__ == "__main__":
    S.wait_front("Cursor"); S.clear_mods()
    for name, fn in (("L15_B_primer", t_primer), ("L15_C_clone", t_clone)):
        print(f"--- {name} ---", flush=True)
        r = S.Rec(name, guard=GUARD)
        r.start(settle=2.0)
        try:
            fn()
        finally:
            r.stop()
        time.sleep(1.5)
