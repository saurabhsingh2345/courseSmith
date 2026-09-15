#!/usr/bin/env python3
"""L34 take H — the shortcut that would have shipped.

The reference's honest section is "everything is in one module". That is not
true of ours — db.py and main.py are properly separated — so this films the
thing that IS true and is worse: the sign-in is entirely client-side, which
means the password is inside a JavaScript file the browser downloads, readable
by anyone who visits.

Nobody did anything wrong. Our own brief said one hard-coded user, MVP, and that
is what got built. The lecture's point is about knowing which of your own
simplifications mean "later" and which mean "this must never leave this laptop".

The chunk name is discovered at shoot time — it changes on every build.
"""
import os, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S

PROJ = "/Users/Shared/projects/pm"


def run(cmd, hold, cps=13):
    S.type_text(cmd, cps=cps); time.sleep(0.5)
    S.key('return')
    print(f"  $ {cmd}   (hold {hold}s)", flush=True)
    time.sleep(hold)


if __name__ == "__main__":
    CODE = "/usr/local/bin/code" if os.path.exists("/usr/local/bin/code") else "code"
    os.system(f'"{CODE}" -r {PROJ}')
    time.sleep(6)
    S.wait_front_window("Code", "pm")
    S.clear_mods()
    S.key('`', ('ctrl',)); time.sleep(3.0)
    S.type_text(f"cd {PROJ} && export PS1='pm $ ' && clear", cps=30)
    S.key('return'); time.sleep(2.0)

    r = S.SegRec("L34_H_honest", allow=("Code",))
    r.start(settle=2.5)
    try:
        run("curl -s localhost:8000/ | grep -o '/_next/static/chunks/[a-z0-9_-]*\\.js' | sort -u", 16)
        run("F=$(for f in $(curl -s localhost:8000/ | grep -o '/_next/static/chunks/[a-z0-9_-]*\\.js' "
            "| sort -u); do curl -s localhost:8000$f | grep -q admin123 && echo $f; done | head -1); "
            "echo \"the password is in: $F\"", 18)
        run("curl -s localhost:8000$F | grep -o '.\\{34\\}admin123.\\{34\\}'", 24)
        time.sleep(8)
    finally:
        r.stop()
