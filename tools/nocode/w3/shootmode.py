#!/usr/bin/env python3
"""Apply (or remove) the VS Code shoot settings.

    python3 shootmode.py on
    python3 shootmode.py off

teardown.sh restores his own settings at the end of a session, which is correct —
but it means the NEXT session starts with them. On 2026-08-31 a Day 4 capture ran
for five minutes against a light-theme window showing his GitHub avatar, this
session's name in the terminal tab, and the wrong model, because nobody put shoot
mode back. preflight.py now refuses to roll unless this has been applied.
"""

from __future__ import annotations

import json
import os
import re
import shutil
import sys

LIVE = os.path.expanduser("~/Library/Application Support/Code/User/settings.json")
BACKUP = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                      "restore", "vscode-settings.backup.json")

SHOOT = {
    "workbench.colorTheme": "Dark Modern",
    "window.autoDetectColorScheme": False,
    "workbench.startupEditor": "none",
    "editor.fontSize": 17,
    "terminal.integrated.fontSize": 15,
    "editor.minimap.enabled": False,
    "breadcrumbs.enabled": False,
    "editor.renderWhitespace": "none",
    "explorer.compactFolders": False,
    "window.commandCenter": False,
    "workbench.layoutControl.enabled": False,
    "workbench.tips.enabled": False,
    # His GitHub avatar is a photograph and it lives in the activity bar.
    "workbench.activityBar.location": "hidden",
    "telemetry.telemetryLevel": "off",
    "update.showReleaseNotes": False,
    "terminal.integrated.gpuAcceleration": "off",
    "extensions.ignoreRecommendations": True,
    "workbench.enableExperiments": False,
    "update.mode": "none",
    "git.autofetch": False,
    # The terminal must not inherit this session's identity.
    "terminal.integrated.env.osx": {
        "CLAUDECODE": None, "CLAUDE_CODE_CHILD_SESSION": None,
        "CLAUDE_CODE_SESSION_ID": None, "CLAUDE_CODE_ENTRYPOINT": None,
        "CLAUDE_CODE_EXECPATH": None, "CLAUDE_CODE_MESSAGING_SOCKET": None,
        "CLAUDE_CODE_MESSAGING_TOKEN": None, "CLAUDE_PID": None,
        "CLAUDE_EFFORT": None, "CLAUDE_JOB_DIR": None,
    },
}

MARKER = "workbench.activityBar.location"


def load() -> dict:
    return json.loads(re.sub(r"//.*", "", open(LIVE).read()))


def is_on() -> bool:
    try:
        d = load()
    except Exception:
        return False
    return d.get(MARKER) == "hidden" and d.get("workbench.colorTheme") == "Dark Modern"


def main() -> None:
    mode = sys.argv[1] if len(sys.argv) > 1 else "status"
    if mode == "status":
        print("shoot mode:", "ON" if is_on() else "OFF")
        return
    if mode == "on":
        if not os.path.exists(BACKUP):
            shutil.copy2(LIVE, BACKUP)
            print("backed up his settings")
        d = load()
        d.update(SHOOT)
        json.dump(d, open(LIVE, "w"), indent=4)
        print(f"shoot mode ON ({len(d)} keys)")
    elif mode == "off":
        if os.path.exists(BACKUP):
            shutil.copy2(BACKUP, LIVE)
            print("shoot mode OFF — his settings restored")
        else:
            print("no backup to restore")
    else:
        sys.exit(__doc__)


if __name__ == "__main__":
    main()
