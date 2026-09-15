#!/usr/bin/env python3
"""L34 take D — the assistant, after it was made real.

Until an hour before this take, "move a card by asking" was decided by a regular
expression that ran before the model was ever called, and the API key never
reached the container — so every question that genuinely needed the model failed
while the one thing that did not need it worked perfectly. Both are fixed here.

The move is phrased INDIRECTLY on purpose — "the card about the flaky checkout
tests" is not the card's title. A pattern cannot resolve that; a model reading
the board can. That is the proof, and it is why the wording matters.
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

BOX, SEND = (1748, 974), (1748, 1049)
USER, PASS, GO = (958, 558), (958, 644), (958, 712)
OUT = (1861, 119)


def beat(msg, hold):
    print(f"  {msg}  (hold {hold}s)", flush=True)
    time.sleep(hold)


def ask(text, wait):
    S.click(*BOX, 0.9); time.sleep(0.5)
    S.type_text(text, cps=11); time.sleep(1.0)
    S.click(*SEND, 0.9)
    print(f"  ? {text}   (wait {wait}s)", flush=True)
    time.sleep(wait)


if __name__ == "__main__":
    S.osa('tell application "Google Chrome" to activate'); time.sleep(1.5)
    S.wait_front("Google Chrome"); time.sleep(0.8); S.clear_mods()

    r = S.SegRec("L34_D_demo", allow=("Google Chrome",))
    r.start(settle=2.5)
    try:
        beat("the board, and the assistant beside it", 13)
        ask("Summarise my project in two sentences.", 42)
        beat("holding on the answer", 18)
        ask("Please move the card about the flaky checkout tests into the In review column.", 45)
        beat("and it moved, on screen", 20)
        S.move(700, 400, 1.6)
        beat("where it landed", 10)
        S.click(*OUT, 1.2); time.sleep(2.5)
        S.key('r', ('cmd', 'shift')); time.sleep(5.0)
        beat("logged out and hard reloaded", 8)
        S.click(*USER, 0.9); time.sleep(0.4); S.type_text("admin", cps=7)
        S.click(*PASS, 0.9); time.sleep(0.4); S.type_text("admin123", cps=7)
        S.click(*GO, 1.0); time.sleep(4.5)
        beat("and the AI's move survived the session", 20)
        S.move(958, 460, 1.4)
        beat("final hold", 12)
    finally:
        r.stop()
