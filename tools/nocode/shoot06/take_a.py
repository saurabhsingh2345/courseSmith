import sys, os, time
sys.path.insert(0, "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith/tools/nocode/scripts")
import stage as S
assert S.wait_front("Cursor"), "Cursor not frontmost"
S.move(960, 620, 0.6); time.sleep(1.0)

r = S.Rec("A_ide", ("Cursor",))
r.start(settle=2.0)
time.sleep(2.5)

# --- the three regions, named in the narration: files / code / agent -------
S.move(70, 102, 1.4);   time.sleep(2.6)     # file tree: index.html
S.move(73, 125, 0.9);   time.sleep(2.2)     # README.md
S.move(69, 102, 0.8);   time.sleep(1.6)
S.move(730, 300, 1.5);  time.sleep(2.0)     # into the code
S.drive.scroll(-320, steps=16); time.sleep(2.4)
S.drive.scroll(-320, steps=16); time.sleep(2.4)
S.drive.scroll(360, steps=16);  time.sleep(2.0)
S.move(1536, 300, 1.6); time.sleep(2.6)     # the agent panel
S.drive.scroll(-260, steps=14); time.sleep(2.4)
S.drive.scroll(300, steps=14);  time.sleep(2.2)
# --- the model, which the narration says is the part that matters ----------
S.move(1354, 1023, 1.3); time.sleep(3.4)
S.move(1536, 991, 1.0);  time.sleep(2.2)    # the prompt box
# --- back out to the whole window ------------------------------------------
S.move(960, 520, 1.6);  time.sleep(3.0)
r.stop()
