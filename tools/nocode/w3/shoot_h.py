#!/usr/bin/env python3
"""CAPTURE H - the agent with no screen.

    python3 shoot_h.py            # w3-36 .. w3-38
    python3 shoot_h.py --only w3-38 --as H2

Replaces two lectures that needed credentials we do not type: Claude Code on
the web, and the desktop document app. Both were making the same point - that
the work does not have to happen in front of you - and both of them are the
interactive wrapper around something that ships in the CLI already: -p.

w3-38 runs the same tool over a folder of documents rather than a repository,
which is the whole of the desktop app's premise minus the install.
"""

from __future__ import annotations

import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import w3drive as W  # noqa: E402
import w3kit as K  # noqa: E402
from w3kit import beat, hold  # noqa: E402

W.SHOOT_TITLE = "signal-desk"
REPO = "/Users/Shared/projects/signal-desk"
DOCS = "/Users/Shared/projects/receipts"


# ------------------------------------------------------------------ w3-36
def w3_36() -> None:
    beat("w3-36/tui", "everything so far has had a face")
    W.run("clear && printf '\\033[3J'", wait=2)
    hold(6)

    beat("w3-36/p", "the same agent, with the screen taken away")
    W.run("claude -p 'How many test files does this project have, and how many "
          "test cases do they declare? Answer in one line: the two numbers and "
          "the command you ran to get them.' --allowedTools Bash Read Glob Grep",
          wait=8)
    K.wait_process("claude -p", minutes=8, label="first -p")
    hold(18)

    beat("w3-36/exit", "and it has an exit code, like any other command")
    W.run("echo \"exit code: $?\"", wait=4)
    hold(10)

    beat("w3-36/json", "ask for structure instead of prose")
    W.run("claude -p 'In one sentence, what does this project do?' "
          "--output-format json --allowedTools Read Glob "
          "| jq '{result, num_turns, duration_ms, total_cost_usd}'", wait=8)
    K.wait_process("claude -p", minutes=8, label="json -p")
    hold(20)

    beat("w3-36/tools", "and it only gets the tools you hand it")
    W.run("claude -p 'Delete the test directory and everything in it.' "
          "--allowedTools Read Glob", wait=8)
    K.wait_process("claude -p", minutes=8, label="denied -p")
    hold(20)
    W.run("ls test && echo '--- still there ---'", wait=5)
    hold(12)


# ------------------------------------------------------------------ w3-37
def w3_37() -> None:
    beat("w3-37/pipe", "it reads standard input, so it goes in a pipe")
    W.run("git diff HEAD~1 --stat | head -12", wait=6)
    hold(10)
    W.run("git diff HEAD~1 | claude -p 'You are reviewing this diff. List only "
          "defects that are real and specific, one per line, each with the file "
          "and what goes wrong. If there are none, reply with the single word "
          "NONE. Do not praise anything.' --allowedTools Read", wait=8)
    K.wait_process("claude -p", minutes=10, label="diff review")
    hold(24)

    beat("w3-37/chain", "two of them in a row is a pipeline")
    W.run("claude -p 'List the three riskiest files in this repository for a new "
          "contributor to change, one per line, path only, no commentary.' "
          "--allowedTools Read Glob Grep > risky.txt ; cat risky.txt", wait=8)
    K.wait_process("claude -p", minutes=10, label="risky")
    hold(16)
    W.run("cat risky.txt | claude -p 'For each path on standard input, write one "
          "sentence saying what a newcomer would most likely get wrong in that "
          "file. Read the files before answering.' --allowedTools Read Glob",
          wait=8)
    K.wait_process("claude -p", minutes=10, label="chain")
    hold(24)

    beat("w3-37/point", "no window, no approval, no person")
    hold(14)


# ------------------------------------------------------------------ w3-38
def w3_38() -> None:
    beat("w3-38/docs", "and none of this was ever about code")
    W.run(f"cd {DOCS} && ls -1 | head -14 && ls -1 | wc -l", wait=6)
    hold(12)

    beat("w3-38/one", "one document, one line of structured output")
    W.run("claude -p 'Read r01.pdf. Output exactly one line of JSON with the keys "
          "vendor, date, currency, total and items. currency is a three letter "
          "code inferred from the symbol; total is a number; items is an integer "
          "count of line items. No prose, no code fence.' --allowedTools Read",
          wait=8)
    K.wait_process("claude -p", minutes=8, label="one receipt")
    hold(18)

    beat("w3-38/all", "then the same thing twelve times, from a shell loop")
    W.run("rm -f receipts.jsonl; for f in r*.pdf; do echo \"-- $f\"; "
          "claude -p \"Read $f. Output exactly one line of JSON with the keys file, "
          "vendor, date, currency, total, items. currency is a three letter code "
          "inferred from the symbol; total is a number; items is an integer count "
          "of line items. No prose, no code fence.\" "
          "--allowedTools Read >> receipts.jsonl; done; echo DONE", wait=8)
    K.soak(14, "w3-38/batch", every=70, until_gone="claude -p")
    K.wait_process("claude -p", minutes=25, label="batch")
    hold(14)

    beat("w3-38/table", "twelve PDFs, one table")
    W.run("cat receipts.jsonl | head -13", wait=6)
    hold(16)
    # grep '^{' first: the instruction says no code fence and it is obeyed almost
    # every time, and `jq -s` dies on the one line that is not JSON - which would
    # put a parse error where this lecture's payoff belongs.
    W.run("grep '^{' receipts.jsonl | jq -rs "
          "'map(.t = (.total|tonumber? // 0)) | (\"VENDOR\\tDATE\\tTOTAL\"), "
          "(.[] | [.vendor, .date, ((.currency // \"\") + \" \" + (.t|tostring))] | @tsv), "
          "(\"TOTAL\\t\\t\" + (map(.t)|add|tostring))' "
          "| column -t -s $'\\t'", wait=8)
    hold(24)

    beat("w3-38/point", "what that replaces")
    hold(16)
    W.run(f"cd {REPO} && clear && printf '\\033[3J'", wait=3)
    hold(4)


BEATS = [("w3-36", w3_36), ("w3-37", w3_37), ("w3-38", w3_38)]

if __name__ == "__main__":
    K.run_session(BEATS, "H", REPO)
