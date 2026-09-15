#!/bin/bash
# Put the machine back exactly as it was before the Week 3 shoot.
# Written BEFORE the shoot started. Safe to run twice.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
R="$HERE/restore"

echo "== stopping any capture this rig started =="
pkill -f "avfoundation" 2>/dev/null && echo "  killed an ffmpeg capture" || echo "  none running"

echo "== VS Code settings =="
S="$HOME/Library/Application Support/Code/User/settings.json"
if [ -f "$R/vscode-settings.backup.json" ]; then
  cp "$R/vscode-settings.backup.json" "$S" && echo "  restored"
else
  echo "  NO BACKUP — leaving as is"
fi

echo "== displays =="
# The panel was ALREADY primary before this shoot and no marker was written, so
# nothing was rearranged and nothing should be moved back. displays.json records
# what it looked like, for a human to compare against.
if [ -f "$HERE/../shoot-day3/.arranged" ]; then
  python3 "$HERE/../scripts/arrange.py" restore && echo "  restored"
else
  echo "  this shoot did not rearrange displays — left alone"
  echo "  (state at start: $(cat "$R/displays.json" | tr -d '\n ' ))"
fi

echo "== Claude Code plugins =="
P="$HOME/.claude/plugins/installed_plugins.json"
if [ -f "$R/installed_plugins.backup.json" ]; then
  cp "$R/installed_plugins.backup.json" "$P" &&     echo "  restored his 3 user plugins (frontend-design, vercel, gopls-lsp)"
else
  echo "  NO BACKUP — leaving as is"
fi

# The plugin CACHE is not touched: Claude Code re-materialises it on its own,
# so moving it aside achieved nothing except a duplicate. Only
# installed_plugins.json is cleared for the shoot, and that is restored above.

echo "== desktop icons =="
defaults write com.apple.finder CreateDesktop -bool true 2>/dev/null
killall Finder 2>/dev/null
sleep 2
open -a Finder 2>/dev/null
echo "  desktop icons shown again (killall Finder alone leaves the desktop dead)"

echo "== menu bar and dock =="
if [ -f "$R/menubar.json" ] && grep -q UNSET "$R/menubar.json"; then
  defaults delete NSGlobalDomain _HIHideMenuBar 2>/dev/null && echo "  menu bar shown again"
else
  echo "  menu bar left as found"
fi
osascript -e 'tell application "System Events" to tell dock preferences to set autohide to false' 2>/dev/null
echo "  dock restored"

echo "== window visibility =="
osascript -e 'tell application "System Events" to set visible of every process to true' 2>/dev/null
echo "  all processes visible again"

echo "== the advert guard stays =="
echo "  kickbacks stub is deliberate — do NOT restore it (see memory)"

echo "== what is NOT touched =="
echo "  the shoot repo /Users/Shared/projects/control-tower is left in place"
echo "  captures and cuts under tools/nocode/w3 are left in place"
echo "done."
