#!/usr/bin/env python3
"""L15 takes — the lab setup: node check, the terminal primer, and the real
clone of our own starter repo.

Shot from /Users/Shared/projects so `pwd` prints a path with no username in it.
The clone URL does carry his GitHub handle — he approved that explicitly.
"""
import os, sys, time, subprocess
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Cursor",)
HERE = os.path.dirname(os.path.abspath(__file__))


def open_cursor_on_projects():
    S.osa('tell application "System Events" to set visible of every process '
          'whose name is not "Finder" to false')
    time.sleep(1.2)
    S.close_finder_windows()
    subprocess.run(["osascript", "-e", 'tell application "Cursor" to quit'],
                   capture_output=True)
    time.sleep(5)
    subprocess.Popen(["open", "-a", "Cursor", "/Users/Shared/projects"])
    time.sleep(14)
    S.wait_front("Cursor"); time.sleep(2)
    S.key('return'); time.sleep(2)          # trust prompt, if any
    S.clear_mods()
    S.move_window_to_panel("Cursor")
    time.sleep(2)
    S.key('b', ('cmd', 'alt')); time.sleep(1.5)   # agent panel away — full width
    S.move(960, 620, 0.4); time.sleep(1.2)
    return S.snap(os.path.join(HERE, "pre_l15.png"))


def open_terminal():
    """The default zsh prompt prints `user@host` on EVERY line — here that is
    his username and a hostname carrying the company name. Neutralise the
    prompt to the path only, then wipe the scrollback (plain `clear` leaves the
    old prompt lines in the buffer, so the 3J escape is required)."""
    S.key('j', ('cmd',)); time.sleep(3.0)
    S.clear_mods()
    S.move(700, 800, 0.6); time.sleep(1.0)
    S.type_text("PROMPT='%~ %# '", cps=30); time.sleep(0.5)
    S.key('return'); time.sleep(1.5)
    S.type_text("clear && printf '\\033[3J'", cps=30); time.sleep(0.5)
    S.key('return'); time.sleep(2.0)


def t_node():
    """node --version, and the gentle terminal primer"""
    S.move(700, 800, 0.9); time.sleep(1.5)
    S.type_text("node --version", cps=11); time.sleep(1.2)
    S.key('return'); time.sleep(3.5)
    S.type_text("pwd", cps=8); time.sleep(1.0)
    S.key('return'); time.sleep(3.5)
    time.sleep(2.0)


def t_clone():
    """the real clone, then into it"""
    S.type_text("git clone https://github.com/saurabhsingh2345/kanban.git", cps=15)
    time.sleep(1.5)
    S.key('return'); time.sleep(7.0)
    S.type_text("cd kanban", cps=9); time.sleep(0.8)
    S.key('return'); time.sleep(2.0)
    S.type_text("pwd", cps=8); time.sleep(0.8)
    S.key('return'); time.sleep(2.5)
    S.type_text("ls", cps=6); time.sleep(0.8)
    S.key('return'); time.sleep(3.5)
    time.sleep(2.0)


if __name__ == "__main__":
    print("pre-roll:", open_cursor_on_projects(), flush=True)
    open_terminal()
    for name, fn in (("L15_B_node", t_node), ("L15_C_clone", t_clone)):
        print(f"--- {name} ---", flush=True)
        r = S.Rec(name, guard=GUARD)
        r.start(settle=2.0)
        try:
            fn()
        finally:
            r.stop()
        time.sleep(1.5)
