#!/usr/bin/env python3
"""L16 takes — the agents.md walkthrough and the Auto Run setting.

Separate takes so any one can be retried without losing the rest, and so each
can be retimed into its own slide window later.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

GUARD = ("Cursor",)
OUT = os.path.dirname(os.path.abspath(__file__))
# sidebar entries, in 1920x1080 point space on the shoot panel
AGENTS_MD = (67, 103)
README_MD = (72, 125)
EDITOR = (888, 456)


def take(name, body, settle=2.0):
    r = S.Rec(name, guard=GUARD)
    r.start(settle=settle)
    try:
        body()
    finally:
        ok = r.stop()
    return ok


def a_shortcuts():
    """the two panel toggles, on camera, a few times"""
    S.move(600, 500, 0.8); time.sleep(1.2)
    for _ in range(2):
        S.key('b', ('cmd',)); time.sleep(1.6)          # file tree away
        S.key('b', ('cmd',)); time.sleep(1.8)          # and back
    time.sleep(1.0)
    for _ in range(2):
        S.key('b', ('cmd', 'alt')); time.sleep(1.8)    # agent panel away
        S.key('b', ('cmd', 'alt')); time.sleep(2.0)    # and back
    time.sleep(1.5)


def b_one_file():
    """the joke: the whole repo is one file"""
    S.move(*AGENTS_MD, 1.2); time.sleep(1.6)
    S.move(*README_MD, 0.8); time.sleep(1.4)
    S.move(43, 80, 0.8); time.sleep(2.0)               # the KANBAN root
    time.sleep(2.0)


def c_open_file():
    """open agents.md and let the markdown structure read"""
    S.click(*AGENTS_MD, 1.0); time.sleep(3.0)
    S.move(700, 300, 1.0); time.sleep(2.0)
    for _ in range(3):
        S.scroll(-260); time.sleep(1.4)
    time.sleep(1.5)
    for _ in range(3):
        S.scroll(260); time.sleep(1.2)
    time.sleep(2.0)


def main():
    if not S.wait_front("Cursor"):
        raise SystemExit("Cursor not frontmost")
    S.clear_mods()
    print("frontmost:", S.frontmost())
    results = {}
    for name, fn in (("L16_A_shortcuts", a_shortcuts),
                     ("L16_B_onefile", b_one_file),
                     ("L16_C_file", c_open_file)):
        print(f"\n--- {name} ---", flush=True)
        results[name] = take(name, fn)
        time.sleep(1.5)
    print("\n", results)


if __name__ == "__main__":
    main()
