#!/bin/bash
# Restore everything the Day 3 shoot changes. Safe to run more than once.
# Written BEFORE the shoot on purpose — see coursesmith-shoot-teardown memory.
set -x
D="$(cd "$(dirname "$0")" && pwd)"

# 1. Cursor settings back to exactly what they were
cp "$D/cursor-settings.backup.json" "$HOME/Library/Application Support/Cursor/User/settings.json"

# 2. macOS appearance (was Light: AppleInterfaceStyle unset)
if [ "$(cat "$D/AppleInterfaceStyle.before")" = "UNSET(light)" ]; then
  defaults delete -g AppleInterfaceStyle 2>/dev/null
fi

# 3. Desktop icons (was UNSET, i.e. shown)
if [ "$(cat "$D/CreateDesktop.before")" = "UNSET" ]; then
  defaults delete com.apple.finder CreateDesktop 2>/dev/null
fi

# 4. Dock autohide (was 0)
defaults write com.apple.dock autohide -bool "$(cat "$D/dock-autohide.before" | sed 's/^0$/false/; s/^1$/true/')"

# 5. Cursor Run Mode — MUST be put back by hand, it is not in settings.json.
#    Cmd+Shift+J > Agents > Execution and Approvals > Run Mode
echo "*** SET CURSOR RUN MODE BACK TO: $(cat "$D/cursor-runmode.before") ***"
echo "*** Cmd+Shift+J > Agents > scroll to Execution and Approvals ***"

# 5b. VS Code settings back
cp "$D/vscode-settings.backup.json" "$HOME/Library/Application Support/Code/User/settings.json" 2>/dev/null

# 6. Displays back — ONLY if this shoot actually rearranged them.
#    On 2026-08-28 the external DELL was ALREADY primary before the shoot began
#    and arrange.py was never run, so an unconditional `restore` would have
#    moved his primary display to the built-in — a change he never had. The
#    marker file is written by `arrange.py shoot` and removed here.
if [ -f "$D/.arranged" ]; then
  python3 "$D/../scripts/arrange.py" restore && rm -f "$D/.arranged"
else
  echo "displays: arrange.py shoot was never run, leaving the arrangement alone"
fi

# 7. Relaunch Finder and Dock, then close the Finder windows killall reopens
killall Finder; killall Dock; sleep 2
osascript -e 'tell application "Finder" to close every window'

# 8. Unhide every app the prep hid
osascript -e 'tell application "System Events" to set visible of every process to true'

# 9. Kill any capture probe left running
pkill -f "_capprobe" 2>/dev/null
pkill -f "avfoundation" 2>/dev/null
set +x
echo "teardown complete"
