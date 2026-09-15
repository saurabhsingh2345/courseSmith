#!/usr/bin/env python3
"""Fix the three things that made the first L18 takes unusable:
   1. the editor was an empty VS Code watermark for 12 minutes -> open agents.md
   2. the chat panel was 16% of frame width and illegible -> widen the sash
   3. dismiss the /create-agent tip so it is not burned into every frame
"""
import os, sys, time
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

OUT = os.path.dirname(os.path.abspath(__file__))

S.wait_front("Code"); time.sleep(1.0); S.clear_mods()

# 1. open the brief in the editor so the middle of the frame carries content
S.click(110, 68, 0.7); time.sleep(2.0)

# 2. widen the chat panel: drag its sash left
S.drag(1628, 520, 1150, 520, glide=0.9)
time.sleep(1.5)

S.move(900, 600, 0.5); time.sleep(1.0)
print("frontmost:", S.frontmost())
print(S.snap(os.path.join(OUT, "l18v2_frame.png")))
