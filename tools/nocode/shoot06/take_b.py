import sys, os, time
sys.path.insert(0, "/Users/enfecsolutions/Desktop/enfec_subs/courseSmith/tools/nocode/scripts")
import stage as S
assert S.wait_front("Cursor"), "Cursor not frontmost"

r = S.Rec("B_ask", ("Cursor",))
r.start(settle=2.0)
time.sleep(2.0)

S.click(1536, 991, 1.2)                      # the prompt box
time.sleep(1.4)
S.type_text("add a small fps counter to the top left corner of the screen", cps=15)
time.sleep(2.2)
S.key('return')
time.sleep(3.0)
# let it actually work — real waiting, ramped in post
for _ in range(22):
    time.sleep(4.0)
S.move(1536, 500, 1.2); time.sleep(4.0)      # read what it did
S.drive.scroll(-240, steps=14); time.sleep(3.5)
S.move(730, 400, 1.4); time.sleep(4.0)       # the change, in the file
S.drive.scroll(-200, steps=12); time.sleep(3.5)
r.stop()
