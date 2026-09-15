#!/usr/bin/env bash
# Pack everything the OTHER Mac needs to cut a no-code lecture film.
#
# Everything in here is either non-reproducible (the memory, the reference
# notes, the marker sheets, the .env) or small source. Footage, renders,
# node_modules and the .voice caches are deliberately left behind: they are
# gigabytes and the other machine does not need this machine's footage to
# shoot its own lectures.
#
#   ./pack.sh                 -> full kit    (~4 MB)
#   ./pack.sh --memory-only   -> just the brain, for a daily re-sync (~800 KB)
#   ./pack.sh --with-pack     -> full kit + No-Code.zip design source (~51 MB)
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
OUT="${OUT:-$HOME/Desktop}"
MEM="$HOME/.claude/projects/$(printf '%s' "$REPO" | tr '/_' '--')/memory"
MODE="${1:-full}"
STAMP="$(date +%Y%m%d-%H%M)"
STAGE="$(mktemp -d)/nocode-kit"
trap 'rm -rf "$(dirname "$STAGE")"' EXIT
mkdir -p "$STAGE"

say() { printf '  %s\n' "$*"; }

# ---- 1. the brain: Claude's memory -----------------------------------------
mkdir -p "$STAGE/memory"
rsync -a "$MEM/" "$STAGE/memory/"
say "memory            $(ls "$STAGE/memory" | wc -l | tr -d ' ') files"

# ---- 2. the reference program: transcripts, ref notes, the build plan ------
mkdir -p "$STAGE/docs"
rsync -a "$REPO/docs/nocode-course/" "$STAGE/docs/nocode-course/"
say "docs/nocode-course  PLAN + curriculum + ref-L01..L34 + transcripts"

# ---- 3. marker sheets: every deviation from his lectures, per film ---------
if [ -d "$REPO/videos/nocode/markers" ]; then
  mkdir -p "$STAGE/markers"
  rsync -a "$REPO/videos/nocode/markers/" "$STAGE/markers/"
  say "markers           $(ls "$STAGE/markers" | wc -l | tr -d ' ') sheets"
fi

# ---- 4. the secret: ElevenLabs key + the Adam voice id --------------------
# adam_lay.py reads this file by absolute path, so it has to land in the same
# place on the other machine. It is a secret: this tarball never goes to a
# public repo, a gist, or any chat that is not yours.
if [ -f "$REPO/tools/nocode/nc/No-Code/.env" ]; then
  mkdir -p "$STAGE/secret"
  cp "$REPO/tools/nocode/nc/No-Code/.env" "$STAGE/secret/No-Code.env"
  say "secret            No-Code.env (ELEVENLABS_API_KEY + VOICE_ID)"
fi

# ---- 5. the demo app the lectures film ------------------------------------
if [ -d "$HOME/arena" ]; then
  mkdir -p "$STAGE/arena"
  rsync -a "$HOME/arena/" "$STAGE/arena/"
  say "arena             the game the lectures build on screen"
fi

if [ "$MODE" != "--memory-only" ]; then
  # ---- 6. the deck: Remotion source for the animated half of each film ----
  mkdir -p "$STAGE/src"
  rsync -a "$REPO/renderer/src/nocode/"   "$STAGE/src/nocode/"
  rsync -a "$REPO/renderer/src/nocode01/" "$STAGE/src/nocode01/"
  say "renderer/src      nocode (kit.tsx + l03..l10) + nocode01"

  # ---- 7. the rig: voice, shoot-driving and assembly scripts -------------
  mkdir -p "$STAGE/scripts"
  rsync -a --include='*.py' --include='*.sh' --exclude='*' \
        "$REPO/tools/nocode/scripts/" "$STAGE/scripts/"
  [ -d "$REPO/tools/nocode/shoot06" ] && \
    rsync -a --include='*.py' --exclude='*' "$REPO/tools/nocode/shoot06/" "$STAGE/shoot06/"
  rsync -a "$REPO/tools/nocode/deck/" "$STAGE/deck/"
  rsync -a "$REPO/tools/nocode/narration/" "$STAGE/narration/" 2>/dev/null || true
  say "tools/nocode      $(ls "$STAGE/scripts" | wc -l | tr -d ' ') rig scripts + deck + narration"

  # ---- 8. this kit itself, so the other machine can pack back ------------
  mkdir -p "$STAGE/sync"
  rsync -a "$REPO/tools/nocode/sync/" "$STAGE/sync/"
fi

if [ "$MODE" = "--with-pack" ] && [ -f "$REPO/No-Code.zip" ]; then
  cp "$REPO/No-Code.zip" "$STAGE/"
  say "No-Code.zip       design source for the animation templates"
fi

cat > "$STAGE/MANIFEST.txt" <<EOF
no-code lecture kit
packed   $(date '+%Y-%m-%d %H:%M')  from $(scutil --get ComputerName 2>/dev/null || hostname)
mode     $MODE
source   $REPO

Shoot geometry of the packing machine — footage only intercuts if the other
machine matches this. Check it before rolling:
$(system_profiler SPDisplaysDataType 2>/dev/null | grep -E 'Display Type|Resolution|UI Looks like' | sed 's/^ */  /')
  model  $(system_profiler SPHardwareDataType 2>/dev/null | grep 'Model Identifier' | sed 's/.*: //')
  macOS  $(sw_vers -productVersion)

Apply with:  ./sync/unpack.sh <path-to-this-kit>
EOF

TAR="$OUT/nocode-kit-$STAMP.tar.gz"
tar -czf "$TAR" -C "$(dirname "$STAGE")" nocode-kit
printf '\n  -> %s  (%s)\n' "$TAR" "$(du -h "$TAR" | cut -f1)"
printf '  AirDrop it. It holds an API key, so keep it off anything public.\n'
