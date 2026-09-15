#!/usr/bin/env python3
"""L19 take — Claude Code plans and builds the same brief, build three of four.

Claude Code stands in for Codex here: he approved that substitution because Codex
needs a paid ChatGPT plan and this one has a free path in, more of our audience
already has it, and it makes the identical point — a command line agent living
inside the editor, which is the bridge into next week's CLI material.

The prompt is `go ahead and plan`, character for character what L17 typed into
Cursor and L18 typed into Copilot. If the prompt changes, the comparison is
worthless, and the comparison is the whole point of the day.
"""
import os, sys, time
HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, "..", "scripts"))
import stage as S
import approve as A

PROJ = "/Users/Shared/projects/kanban"
PROMPT_BOX = (1450, 985)          # the Claude Code input; override with argv
ALLOW = ("Code",)


if __name__ == "__main__":
    box = PROMPT_BOX
    if len(sys.argv) >= 3:
        box = (int(sys.argv[1]), int(sys.argv[2]))
    print("prompt box:", box)
    # approve=False: the panel is on Auto (full access), the mode this lecture
    # sets on camera, so there are no permission prompts to click and a clicker
    # would only risk hitting something else.
    A.run_build(S, "L19_A_build", PROJ, ALLOW, box, "go ahead and plan",
                HERE, quiet=240.0, cap=3000.0, approve=False)
