import sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import stage as S

PROMPT = ("Please build a website for a 3D first-person shooter game in an arena "
          "against one computer opponent controlled by arrow keys and space bar to shoot.")

assert S.wait_front("Cursor"), "Cursor not frontmost"
S.key('escape'); time.sleep(0.6)
S.key('escape'); time.sleep(1.2)
S.move(960, 900, 0.6); time.sleep(1.5)

r = S.Rec("JKLM_build", ("Cursor",))
r.start(settle=2.0)

# ---- J : the three panes, then the agent (~122s) ---------------------------
time.sleep(3.0)
S.move(140, 82, 1.5)           # INSTANT in the tree
time.sleep(5.0)
S.move(150, 300, 1.3)          # the empty file tree
time.sleep(6.0)
S.move(745, 460, 1.6)          # middle, the editor
time.sleep(6.0)
S.move(745, 560, 1.1)
time.sleep(5.0)
S.move(1550, 100, 1.6)         # right, the agent panel
time.sleep(6.0)
S.key('b', ('cmd',)); time.sleep(3.5)      # tree away
S.key('b', ('cmd',)); time.sleep(3.5)      # and back
S.move(1550, 150, 1.2)
time.sleep(6.0)
S.move(1250, 100, 1.3)         # the New Agent tab
time.sleep(5.0)
S.move(1550, 250, 1.2)
time.sleep(8.0)
S.move(900, 500, 1.5)
time.sleep(8.0)
S.move(1550, 101, 1.4)
time.sleep(9.0)
S.move(1200, 300, 1.3)
time.sleep(9.0)
S.move(1550, 130, 1.2)
time.sleep(10.0)
S.move(1000, 600, 1.4)
time.sleep(9.0)
S.move(1550, 101, 1.3)
time.sleep(8.0)

# ---- K : the model (~33s) --------------------------------------------------
S.move(1337, 213, 1.4)         # the model chip
time.sleep(4.0)
S.click(1337, 213, 0.3)
time.sleep(3.5)
S.move(1400, 243, 1.0)         # Fast
time.sleep(3.0)
S.move(1400, 268, 1.0)         # Effort
time.sleep(3.0)
S.move(1418, 301, 1.0)         # Model
time.sleep(2.0)
S.click(1418, 301, 0.3)
time.sleep(3.0)
S.move(1600, 275, 1.1)         # Cursor Grok 4.6
time.sleep(3.0)
S.move(1595, 357, 1.2)         # Claude Opus 5
time.sleep(4.0)
S.click(1595, 357, 0.3)        # pick it
time.sleep(4.0)

# ---- L : the request (~19s) -----------------------------------------------
S.click(1550, 101, 1.0)        # into the prompt box
time.sleep(2.0)
S.type_text(PROMPT, cps=24)
time.sleep(3.0)
S.key('return')
time.sleep(2.0)

# ---- M : it builds. Let it run; post ramps this down. ---------------------
for _ in range(30):
    time.sleep(10.0)
r.stop()
