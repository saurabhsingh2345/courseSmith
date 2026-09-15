#!/usr/bin/env bash
# Apply a kit from pack.sh onto this Mac and get it ready to cut a lecture.
#
#   ./unpack.sh ~/Desktop/nocode-kit-20260828-1240.tar.gz
#   ./unpack.sh <kit> --repo ~/Desktop/enfec_subs/courseSmith
#   ./unpack.sh --check              # re-run the checks only
set -euo pipefail

GIT_URL="https://github.com/saurabhsingh2345/courseSmith.git"
KIT=""; REPO=""; CHECK_ONLY=0
while [ $# -gt 0 ]; do
  case "$1" in
    --check|--check-only) CHECK_ONLY=1 ;;
    --repo) REPO="$2"; shift ;;
    *) KIT="$1" ;;
  esac; shift
done
[ -n "$REPO" ] || REPO="$HOME/Desktop/enfec_subs/courseSmith"

ok()   { printf '  \033[32mok\033[0m    %s\n' "$*"; }
warn() { printf '  \033[33mTODO\033[0m  %s\n' "$*"; }
bad()  { printf '  \033[31mFAIL\033[0m  %s\n' "$*"; }
head_() { printf '\n\033[1m%s\033[0m\n' "$*"; }

if [ "$CHECK_ONLY" = 0 ]; then
  [ -n "$KIT" ] || { echo "usage: unpack.sh <kit.tar.gz|kit-dir> [--repo DIR]"; exit 2; }

  # ---- unwrap ------------------------------------------------------------
  if [ -d "$KIT" ]; then SRC="$KIT"; else
    TMP="$(mktemp -d)"; tar -xzf "$KIT" -C "$TMP"; SRC="$TMP/nocode-kit"
  fi
  [ -d "$SRC/memory" ] || { bad "not a kit: $SRC"; exit 1; }
  head_ "kit"; sed 's/^/  /' "$SRC/MANIFEST.txt" 2>/dev/null || true

  # ---- the repo ----------------------------------------------------------
  head_ "repo"
  if [ ! -d "$REPO/.git" ]; then
    mkdir -p "$(dirname "$REPO")"
    git clone "$GIT_URL" "$REPO"
    ok "cloned into $REPO"
  else
    git -C "$REPO" fetch --all --quiet && ok "repo present at $REPO"
  fi
  git -C "$REPO" checkout v10-atelier-and-the-101 --quiet 2>/dev/null \
    && ok "on v10-atelier-and-the-101" \
    || warn "could not switch branch — do it by hand"

  # ---- place the untracked source ---------------------------------------
  head_ "source"
  mkdir -p "$REPO/renderer/src" "$REPO/tools/nocode" "$REPO/docs" "$REPO/videos/nocode/markers"
  for d in nocode nocode01; do
    [ -d "$SRC/src/$d" ] && rsync -a "$SRC/src/$d/" "$REPO/renderer/src/$d/" && ok "renderer/src/$d"
  done
  for d in scripts shoot06 deck narration sync; do
    [ -d "$SRC/$d" ] && rsync -a "$SRC/$d/" "$REPO/tools/nocode/$d/" && ok "tools/nocode/$d"
  done
  chmod +x "$REPO/tools/nocode/sync/"*.sh 2>/dev/null || true
  rsync -a "$SRC/docs/nocode-course/" "$REPO/docs/nocode-course/" && ok "docs/nocode-course"
  [ -d "$SRC/markers" ] && rsync -a "$SRC/markers/" "$REPO/videos/nocode/markers/" && ok "videos/nocode/markers"
  [ -d "$SRC/arena" ] && { mkdir -p "$HOME/arena"; rsync -a "$SRC/arena/" "$HOME/arena/"; ok "~/arena"; }
  [ -f "$SRC/No-Code.zip" ] && cp "$SRC/No-Code.zip" "$REPO/" && ok "No-Code.zip"

  # ---- the secret adam_lay.py reads by absolute path --------------------
  if [ -f "$SRC/secret/No-Code.env" ]; then
    mkdir -p "$REPO/tools/nocode/nc/No-Code"
    cp "$SRC/secret/No-Code.env" "$REPO/tools/nocode/nc/No-Code/.env"
    chmod 600 "$REPO/tools/nocode/nc/No-Code/.env"
    ok "tools/nocode/nc/No-Code/.env  (ElevenLabs key + Adam voice id)"
  else
    warn "no .env in the kit — adam_lay.py will not speak without it"
  fi

  # ---- the memory, merged not clobbered --------------------------------
  head_ "Claude memory"
  MEM="$HOME/.claude/projects/$(printf '%s' "$REPO" | tr '/_' '--')/memory"
  mkdir -p "$MEM"
  if [ -f "$MEM/MEMORY.md" ] && ! cmp -s "$MEM/MEMORY.md" "$SRC/memory/MEMORY.md"; then
    cp "$SRC/memory/MEMORY.md" "$MEM/MEMORY.md.incoming"
    warn "MEMORY.md differs — incoming kept as MEMORY.md.incoming, merge the index by hand"
    rsync -a --exclude MEMORY.md "$SRC/memory/" "$MEM/"
  else
    rsync -a "$SRC/memory/" "$MEM/"
  fi
  ok "$MEM  ($(ls "$MEM" | wc -l | tr -d ' ') files)"

  # ---- deps ------------------------------------------------------------
  head_ "dependencies"
  command -v brew >/dev/null || warn "install Homebrew: https://brew.sh"
  for f in ffmpeg node; do
    command -v "$f" >/dev/null && ok "$f $("$f" --version 2>&1 | head -1 | cut -c1-40)" \
      || { warn "brew install $f"; }
  done
  if [ -d "$REPO/renderer" ] && [ ! -d "$REPO/renderer/node_modules" ]; then
    ( cd "$REPO/renderer" && npm install ) && ok "renderer/node_modules installed"
  else
    ok "renderer/node_modules present"
  fi
  python3 -m pip install --user --quiet pyobjc-framework-Quartz numpy requests 2>/dev/null \
    && ok "python: Quartz + numpy + requests" || warn "pip install --user pyobjc-framework-Quartz numpy requests"
fi

# =============================== checks ==================================
[ -n "${REPO:-}" ] || REPO="$HOME/Desktop/enfec_subs/courseSmith"
MEM="$HOME/.claude/projects/$(printf '%s' "$REPO" | tr '/_' '--')/memory"

head_ "checks"
[ -d "$MEM" ] && [ -f "$MEM/coursesmith-nocode-RESUME.md" ] \
  && ok "memory in place — a fresh Claude session will read RESUME first" \
  || bad "memory missing at $MEM"
[ -f "$REPO/docs/nocode-course/PLAN.md" ] && ok "reference plan + transcripts" || bad "docs/nocode-course missing"
[ -f "$REPO/renderer/src/nocode/kit.tsx" ] && ok "deck kit.tsx" || bad "renderer/src/nocode missing"
[ -f "$REPO/tools/nocode/scripts/adam_deck.py" ] && ok "voice rig" || bad "tools/nocode/scripts missing"

ENVF="$REPO/tools/nocode/nc/No-Code/.env"
if [ -f "$ENVF" ]; then
  VID="$(grep '^ELEVENLABS_VOICE_ID=' "$ENVF" | cut -d= -f2 | tr -d '"'"'"' ')"
  KEY="$(grep '^ELEVENLABS_API_KEY=' "$ENVF" | cut -d= -f2 | tr -d '"'"'"' ')"
  ok "voice id $VID  (must match the other machine, or Adam changes mid-course)"
  code=$(curl -s -o /dev/null -w '%{http_code}' -m 10 -H "xi-api-key: $KEY" \
         "https://api.elevenlabs.io/v1/voices/$VID" || true)
  [ "$code" = "200" ] && ok "ElevenLabs reachable and the key works" || bad "ElevenLabs said HTTP $code"
else
  bad "$ENVF missing — no voice"
fi

python3 -c 'import Quartz, numpy' 2>/dev/null && ok "pyobjc + numpy (drives Cursor/VS Code during a shoot)" \
  || bad "python deps missing — the shoot scripts cannot move the mouse"

head_ "screen — footage only intercuts if this matches the other Mac"
system_profiler SPDisplaysDataType 2>/dev/null | grep -E 'Display Type|Resolution|UI Looks like' | sed 's/^ */  /'
printf '  capture inputs (pick the built-in screen, index changes per machine):\n'
ffmpeg -f avfoundation -list_devices true -i "" 2>&1 | grep -E 'Capture screen' | sed 's/.*\] /    /' || true

head_ "grant by hand — System Settings > Privacy & Security"
warn "Screen Recording: your terminal app (and VS Code, if you shoot from it)"
warn "Accessibility:    same app — pyobjc CGEvents drive Cursor/VS Code, System Events does not"
warn "Cursor + VS Code installed, both on a DARK theme before rolling"
warn "Chrome signed in to Udemy if you need to pull another lecture transcript"
warn "no name, email or company visible in any frame"
printf '\nThen open a Claude session in %s and say: let'"'"'s finish that Udemy program\n' "$REPO"
