#!/usr/bin/env python3
"""Save and restore every machine setting the install shoot touches.

    python3 state.py save     # snapshot, then apply the shoot settings
    python3 state.py restore  # put all of it back

Written before anything is changed, because this session runs INSIDE the VS Code
whose bundle gets renamed. If that kills the session, `restore` still knows
everything it needs from state.json.

Two things are deliberately NOT touched:
  * ~/Library/Application Support/Code — his real VS Code profile. The freshly
    installed copy gets a virgin one from VSCODE_PORTABLE instead, so a real
    first-run screen is filmed without going anywhere near his settings.
  * The running VS Code process. The bundle is RENAMED, not moved off-volume and
    not quit: a same-volume rename keeps the running process on its own inode.
"""
from __future__ import annotations
import json, os, subprocess, sys

HERE = os.path.dirname(os.path.abspath(__file__))
STATE = os.path.join(HERE, "state.json")
APP = "/Applications/Visual Studio Code.app"
ASIDE = "/Applications/.vscode-live-aside.app"
PORTABLE = "/Users/Shared/vscode-shoot-portable"


def sh(*c, **kw):
    return subprocess.run(c, capture_output=True, text=True, **kw)


def read_default(dom, key):
    r = sh("defaults", "read", dom, key)
    return r.stdout.strip() if r.returncode == 0 else None


def save():
    st = {
        "computer_name": sh("scutil", "--get", "ComputerName").stdout.strip(),
        "create_desktop": read_default("com.apple.finder", "CreateDesktop"),
        "dock_autohide": read_default("com.apple.dock", "autohide"),
        "app_renamed": False,
        "portable_set": False,
    }
    json.dump(st, open(STATE, "w"), indent=1)
    print("saved:", {k: v for k, v in st.items() if k != "computer_name"})

    # The Finder sidebar prints the computer name, and his is his own name.
    sh("scutil", "--set", "ComputerName", "Mac")
    print("computer name -> Mac")

    # A virgin profile for the copy we are about to install, so its first launch
    # is a real first launch. launchctl setenv reaches GUI double-click launches;
    # the already-running instance keeps the environment it started with.
    os.makedirs(os.path.join(PORTABLE, "user-data"), exist_ok=True)
    os.makedirs(os.path.join(PORTABLE, "extensions"), exist_ok=True)
    sh("launchctl", "setenv", "VSCODE_PORTABLE", PORTABLE)
    st["portable_set"] = True

    if os.path.exists(APP) and not os.path.exists(ASIDE):
        os.rename(APP, ASIDE)          # same volume: the live process is fine
        st["app_renamed"] = True
        print(f"live bundle renamed aside (dot-prefixed, so Finder hides it)")
    json.dump(st, open(STATE, "w"), indent=1)


def restore():
    st = json.load(open(STATE))
    # Anything newly installed at the canonical path is the copy we filmed
    # installing. Remove it and give him back the bundle he was running.
    if st.get("app_renamed") and os.path.exists(ASIDE):
        if os.path.exists(APP):
            sh("rm", "-rf", APP)
        os.rename(ASIDE, APP)
        print("live bundle restored")
    if st.get("portable_set"):
        sh("launchctl", "unsetenv", "VSCODE_PORTABLE")
        sh("rm", "-rf", PORTABLE)
        print("portable profile removed")
    # Apps renamed aside so their names stay off camera (his employer's name
    # was in one of them). Same-volume rename, so nothing was ever copied.
    for src, dst in st.get("apps_aside", []):
        if os.path.exists(dst):
            os.rename(dst, src)
            print(f"restored {os.path.basename(src)}")
    # The whole Dock plist is put back, which also restores autohide,
    # show-recents and the pinned VS Code icon in one move.
    bak = st.get("dock_plist")
    if bak and os.path.exists(bak):
        sh("cp", bak, os.path.expanduser("~/Library/Preferences/com.apple.dock.plist"))
        sh("defaults", "read", "com.apple.dock")      # reload from disk
        print("dock plist restored")
    if st.get("computer_name") and st.get("renamed_computer"):
        sh("scutil", "--set", "ComputerName", st["computer_name"])
        print("computer name restored")
    # These two were SET on this machine, not unset, so restore the values —
    # `defaults delete` would have left the dock behaving differently than before.
    cd = st.get("create_desktop")
    if cd is None:
        sh("defaults", "delete", "com.apple.finder", "CreateDesktop")
    else:
        sh("defaults", "write", "com.apple.finder", "CreateDesktop", "-bool",
           "true" if cd == "1" else "false")
    ah = st.get("dock_autohide")
    if ah is not None:
        sh("defaults", "write", "com.apple.dock", "autohide", "-bool",
           "true" if ah == "1" else "false")
    sh("killall", "Finder"); sh("killall", "Dock")
    print(f"finder/dock restored (CreateDesktop={cd} autohide={ah})")


if __name__ == "__main__":
    {"save": save, "restore": restore}[sys.argv[1]]()
