#!/usr/bin/env python3
"""Take 2: the last three beats of the VS Code film, re-shot.

Take 1's Applications window opened further left than the rehearsal's, so the
disk image window sat on top of it and the cursor spent the last minute hovering
empty desktop. The Applications window is now 1100x700 and Finder places it at
(0,168), which fills the frame properly; the disk image window ends up behind it,
so ejecting moved into the narration as advice instead of an action.

Starts from exactly the state take 1 is in at 208.79s: disk image window open at
(100,328,580,680), nothing installed, cursor on the app icon. So the splice is
invisible.
"""
import os, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S
from take import Take

APP_ICON = (220, 521)
APPS_DIR = (459, 518)
VSC = (452, 790)          # where the copy lands in the 1100x700 Applications window

t = Take("vscode2", allow=("Finder", "Dock"))
t.start(settle=3.0)

# 7. the install: one drag
t.hold(6, "drag")
S.drag(*APP_ICON, *APPS_DIR, glide=3.0, hold=0.9, settle=1.0)
time.sleep(26)                       # copy sheet, then Applications opens itself

# 8. the copy landing in the real Applications folder
t.mark("landed")
S.move(*VSC, 1.8); time.sleep(10)
S.move(700, 700, 1.6); time.sleep(6)
S.move(*VSC, 1.4); time.sleep(11)

# 9. and it is installed
t.mark("closing")
S.move(620, 500, 2.2); time.sleep(11)
S.move(*VSC, 2.0); time.sleep(13)
S.move(900, 620, 2.4); time.sleep(10)
S.move(*VSC, 1.8); time.sleep(12)

ok = t.stop()
print("TAKE OK" if ok else "TAKE FAILED")
print("installed:", os.path.exists("/Applications/Visual Studio Code.app"))
