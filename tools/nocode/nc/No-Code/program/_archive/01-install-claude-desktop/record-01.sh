#!/bin/bash
set -euo pipefail
OUT="/Users/ashish/Desktop/No-Code/course/recordings/01-install-claude-desktop.mov"
rm -f "$OUT"

screencapture -v -k -C -x -D 1 "$OUT" &
REC=$!
sleep 2

osascript -e 'tell application "System Events" to set visible of process "Cursor" to false' || true

open -na "Google Chrome" --args --new-window --start-fullscreen \
  "http://127.0.0.1:8765/lessons/01-install-claude-desktop/slides.html?record=1"

sleep 40

osascript <<'EOF'
tell application "Google Chrome"
  activate
  if (count of windows) is 0 then
    make new window
  end if
  set URL of active tab of front window to "https://claude.com/download"
end tell
EOF

sleep 8

osascript <<'EOF'
tell application "Google Chrome"
  execute front window's active tab javascript "const a=[...document.querySelectorAll('a,button')].find(x => /Download for macOS|darwin\\/universal/i.test((x.href||'') + x.textContent)); if (a) a.click();"
end tell
EOF

sleep 12

open -a Claude
sleep 10

kill -INT "$REC" || true
wait "$REC" || true
ls -lh "$OUT"
