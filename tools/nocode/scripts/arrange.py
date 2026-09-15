#!/usr/bin/env python3
"""Make the 16:9 external panel primary for the shoot, and put it back after."""
import sys, time
import Quartz as Q

# Display ids are assigned by the OS and are NOT stable — they were 1 and 2 on
# the day-one shoots and 1 and 3 on 2026-08-28, which made a hardcoded arrange
# silently do nothing. Find them by shape instead: the shoot panel is the 16:9
# one, the built-in Retina is not.
def pick():
    d = displays()
    ext = [i for i, (x, y, w, h) in d.items() if abs(w / h - 16 / 9) < 0.01]
    if not ext:
        raise SystemExit(f"no 16:9 display found in {d}")
    e = ext[0]
    b = [i for i in d if i != e]
    if not b:
        raise SystemExit("only one display attached — nothing to arrange")
    return b[0], e

def displays():
    err, ids, cnt = Q.CGGetActiveDisplayList(8, None, None)
    out = {}
    for d in ids[:cnt]:
        b = Q.CGDisplayBounds(d)
        out[d] = (b.origin.x, b.origin.y, b.size.width, b.size.height)
    return out

def arrange(pairs):
    err, cfg = Q.CGBeginDisplayConfiguration(None)
    for did, (x, y) in pairs.items():
        Q.CGConfigureDisplayOrigin(cfg, did, int(x), int(y))
    Q.CGCompleteDisplayConfiguration(cfg, Q.kCGConfigurePermanently)
    time.sleep(3)

def _marker(on):
    """Leave a marker so teardown knows whether it should restore displays at
    all. Without it teardown ran `restore` unconditionally and would have moved
    his primary display on a day when nothing was ever rearranged."""
    import os
    p = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                     "..", "shoot-day3", ".arranged")
    if on:
        open(p, "w").write("arranged\n")
    elif os.path.exists(p):
        os.remove(p)


if __name__ == "__main__":
    mode = sys.argv[1]
    print("before:", displays())
    BUILTIN, EXTERNAL = pick()
    print(f"builtin={BUILTIN} external={EXTERNAL}")
    if mode == "shoot":
        # external at the origin makes it primary; built-in parked to its right
        arrange({EXTERNAL: (0, 0), BUILTIN: (1920, 0)})
        _marker(True)
    else:
        arrange({BUILTIN: (0, 0), EXTERNAL: (-1920, -85)})
        _marker(False)
    print("after: ", displays())
