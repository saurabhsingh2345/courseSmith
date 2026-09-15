#!/usr/bin/env python3
"""L20 take — Antigravity builds the same brief. Build four of four.

Setup verified before rolling, and three of the reference lecture's facts about
this tool have moved:

  * **It has adopted `agents.md`.** ref-L20 is built around the claim that it has
    not, and makes you copy the brief into `.agent/rules/strategy.md`. The
    shipped bundle carries an `AGENTS.md` settings tab beside `GEMINI.md`, a
    `USE_AGENT_MD` flag, and an `[InstructionsContextComputer] AGENTS.md files
    added:` log line. The take will show whether it picks the file up unaided.
  * **The model list is Gemini 3.7 Flash High by default**, not "Gemini 3 Pro" —
    and alongside Google's own models it offers Claude Sonnet 4.6, Claude Opus
    4.6 and GPT-OSS 120B. We shoot on the default, which is what a free account
    actually gets, and name it on camera.
  * **Auto Execution has TWO settings, not three** — Always Proceed and Request
    Review. His "Agent decides" middle option is gone from this build.

`Agent Auto-Fix Lints` was already On. `Auto Execution` is set to Always Proceed,
which is this tool's YOLO beat, chosen deliberately on a throwaway project.

approve=False: with Always Proceed there should be nothing to click, and
Antigravity is a VS Code fork whose accent buttons would otherwise tempt the
colour-matching clicker into pressing something else.
"""
import os, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S
import approve as A

PROJ = "/Users/Shared/projects/kanban"
PROMPT_BOX = (1400, 450)
# BOTH names are needed. `stage.frontmost()` asks System Events, which reports
# the PROCESS name ("Electron"); `stage.frontmost_fast()` asks NSWorkspace, which
# reports the DISPLAY name ("Antigravity IDE"). SegRec polls the fast one, so a
# guard of only ("Electron",) never matched and the recorder silently never
# started a single segment — the log looked fine and no take appeared.
ALLOW = ("Antigravity IDE", "Electron")


if __name__ == "__main__":
    S.wait_front("Electron"); time.sleep(0.8); S.clear_mods()
    S.key('escape'); time.sleep(1.0)          # dismiss the settings popover
    A.run_build(S, "L20_A_build", PROJ, ALLOW, PROMPT_BOX, "go ahead and plan",
                HERE, quiet=300.0, cap=3200.0, approve=False)
