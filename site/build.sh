#!/usr/bin/env sh
# Builds the public course site.
#
#   hugo            — the only required tool; produces a fully working site.
#   npx pagefind    — OPTIONAL post-build step that adds full-text search.
#                     If node/npx is unavailable the site still works; the
#                     search box simply stays hidden.
set -eu
cd "$(dirname "$0")"

hugo --gc

if command -v npx >/dev/null 2>&1; then
  npx --yes pagefind --site public
else
  echo "npx not found — skipping pagefind index (search will be hidden)." >&2
fi
