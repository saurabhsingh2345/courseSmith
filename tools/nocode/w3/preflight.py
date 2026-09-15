#!/usr/bin/env python3
"""Run before every Week 3 capture. Nothing is changed — it only reports.

A Week 3 capture is up to three hours long. Anything wrong at minute one is
wrong for the whole block, so every known way a take has been ruined before is
checked here first.
"""

from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))

HOME = os.path.expanduser("~")
ok, bad = [], []


def chk(cond: bool, good: str, worse: str) -> None:
    (ok if cond else bad).append(good if cond else worse)


def sh(*a) -> str:
    return subprocess.run(a, capture_output=True, text=True).stdout.strip()


# --- the panel -------------------------------------------------------------
import stage as S  # noqa: E402
chk(S.CAP_INDEX is not None, f"capture index {S.CAP_INDEX} found",
    "NO 16:9 panel — plug in the external monitor")

# --- appearance ------------------------------------------------------------
style = sh("defaults", "read", "-g", "AppleInterfaceStyle")
chk(style == "Dark", "macOS is in Dark mode",
    "macOS is in LIGHT mode — this is what caused the Day-1 bright-footage complaint")

# --- the shell prompt prints his username ----------------------------------
zshrc = os.path.join(HOME, ".zshrc")
chk(os.path.exists(zshrc), "~/.zshrc present (never cat it — it holds live keys)",
    "~/.zshrc missing")
print("   NOTE: set PROMPT='%~ %# ' then `clear && printf '\\033[3J'` in every "
      "terminal that will be on camera")

# --- the kickbacks.ai advert ----------------------------------------------
sl = os.path.join(HOME, ".kickbacks", "vibe-ads-statusline.mjs")
if os.path.exists(sl):
    body = open(sl).read()
    locked = "uchg" in sh("ls", "-lO", sl)
    stubbed = "NEUTRALISED FOR COURSE RECORDING" in body
    chk(stubbed and locked,
        "kickbacks statusline is the locked no-op stub",
        f"kickbacks statusline {'is not stubbed' if not stubbed else 'lost its uchg lock'}"
        " — the advert will render on camera")
else:
    ok.append("no kickbacks statusline on disk")
cj = os.path.join(HOME, ".claude.json")
if os.path.exists(cj):
    raw = open(cj).read()
    chk("kickbacks" not in raw.lower() or "statusLine" not in raw,
        "no kickbacks statusLine in ~/.claude.json",
        "kickbacks key is back in ~/.claude.json — strip it before rolling")

# --- nobody else is filming ------------------------------------------------
ff = sh("pgrep", "-fl", "avfoundation")
chk(not ff, "no other capture running",
    f"ANOTHER CAPTURE IS RUNNING — do not kill it, find the session:\n     {ff}")

# --- shooting location -----------------------------------------------------
chk(os.path.isdir("/Users/Shared/projects"),
    "/Users/Shared/projects exists (shoot from here, never ~)",
    "create /Users/Shared/projects — shooting from ~ puts his username in every pwd")

# --- room ------------------------------------------------------------------
free = shutil.disk_usage("/").free / 1e9
chk(free > 30, f"{free:.0f} GB free (a 2-hour session at 30 Mbit needs ~27 GB)",
    f"only {free:.0f} GB free — cut and purge the previous capture before rolling")

# --- tools -----------------------------------------------------------------
for t in ("ffmpeg", "ffprobe", "claude", "cursor-agent", "docker"):
    chk(shutil.which(t) is not None, f"{t} on PATH", f"{t} NOT on PATH")

# --- model pinned to sonnet ------------------------------------------------
import json as _j
_st = "/Users/Shared/projects/control-tower/.claude/settings.json"
try:
    _m = _j.load(open(_st)).get("model")
except Exception:
    _m = None
chk(_m == "sonnet",
    "shoot repo pinned to sonnet",
    f"shoot repo model is {_m!r} — must be 'sonnet'. Opus/Fable burned his "
    "allowance once already; set \"model\": \"sonnet\" in the repo settings")

# --- VS Code shoot mode ----------------------------------------------------
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import shootmode  # noqa: E402
chk(shootmode.is_on(),
    "VS Code is in shoot mode (dark, activity bar hidden, env scrubbed)",
    "VS Code is NOT in shoot mode — run `python3 tools/nocode/w3/shootmode.py on`."
    " teardown restores his settings, so this must be re-applied every session")

# --- who is in front -------------------------------------------------------
front = S.frontmost_fast()
print(f"\n   frontmost app: {front!r}")
if front not in ("Finder", "Terminal", "iTerm2", None):
    print("   ^ if that is not you launching this, HE IS AT THE MACHINE — do not roll")

print("\n".join("  ok   " + s for s in ok))
if bad:
    print("\n".join("  FAIL " + s for s in bad))
    print(f"\n{len(bad)} problem(s). Fix before rolling.")
    sys.exit(1)
print("\nclear to roll.")
