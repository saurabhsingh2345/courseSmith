#!/usr/bin/env python3
"""The Day 5 build — one continuous recorded session, cut across L31-L34.

His L31-L34 are four lectures of one build. Shooting them as four sessions would
mean four setups and four chances to lose continuity, so this records the whole
thing in segments and the cut splits it afterwards.

Usage: shoot_build.py <take_name> <prompt-file> [quiet_seconds]

The prompt comes from a file rather than the command line so that multi-line
prompts keep their shape and nothing is mangled by shell quoting.

Guards that matter here:
  * `wait_front_window` before every keystroke burst — VS Code has his window
    open too, and a `git clone` once went into it.
  * approve=False: `chat.tools.autoApprove` is on, and Copilot's own held
    categories are clicked by approve.find_button, which is colour-based and was
    written for exactly this panel.
"""
import os, subprocess, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S
import approve as A

PROJ = "/Users/Shared/projects/pm"
PROMPT_BOX = (1400, 958)
ALLOW = ("Code",)

if __name__ == "__main__":
    take = sys.argv[1]
    # '-' resumes a stalled run: roll, approve, wait, type nothing.
    prompt = '' if sys.argv[2] == '-' else open(sys.argv[2]).read().strip()
    quiet = float(sys.argv[3]) if len(sys.argv) > 3 else 300.0

    # `code -r <path>` focuses the existing window for that folder, and unlike
    # AXRaise it will cross Spaces — the build window is fullscreen on its own
    # Space, which is why raising it by title alone timed out and the guard
    # (correctly) refused to type.
    CODE = "/usr/local/bin/code" if os.path.exists("/usr/local/bin/code") else "code"
    subprocess.run([CODE, "-r", PROJ], capture_output=True)
    time.sleep(6)
    S.wait_front_window("Code", "pm")
    S.clear_mods()
    print("prompt:", (prompt[:90].replace("\n", " ") + " ...") if prompt
          else "(resume — approvals only)", flush=True)
    A.run_build(S, take, PROJ, ALLOW, PROMPT_BOX, prompt, HERE,
                quiet=quiet, cap=5400.0, approve=True, cps=22)
