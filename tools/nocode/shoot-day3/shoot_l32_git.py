#!/usr/bin/env python3
"""L32 take E — reading the history, finding the gap, and closing it.

The lecture's checkpoint section is not a complaint. The agent DID commit when
asked, so this take reads what it committed rather than what it skipped:

  * `git log --oneline` — the commits exist, one per part, as instructed
  * `git show --stat HEAD~1` — and part 3's commit contains part 2's Dockerfile,
    back end and scripts, because part 2 never got a checkpoint of its own. The
    work is safe; the RETURN POINT is gone.
  * `git status` — one back-end file is in no commit at all
  * `git add .` / `git commit` — closing it, with the three commands named

The prompt is flattened to `pm $` before anything is filmed. The default zsh
prompt on this machine prints the account name, and a username is PII the same
way a photograph is.
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

    # open the integrated terminal and make the prompt safe to film
    S.key('`', ('ctrl',)); time.sleep(3.0)
    S.type_text(f"cd {PROJ} && export PS1='pm $ ' && clear", cps=30)
    S.key('return'); time.sleep(2.0)

    r = S.SegRec("L32_E_git", allow=("Code",))
    r.start(settle=2.5)
    try:
        run("git log --oneline", 14)
        run("git show --stat HEAD~1", 20)
        run("git status", 14)
        run("git add .", 4)
        run('git commit -m "part 4 complete: front end served, sign-in working"', 10)
        run("git log --oneline", 14)
        time.sleep(6)
    finally:
        r.stop()
