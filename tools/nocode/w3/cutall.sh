#!/bin/bash
# Cut every top-up capture, in one batch, after all the shooting is done.
#
#   ./cutall.sh
#
# cut.py re-encodes at about one times realtime on this machine, so cutting an
# hour of capture costs an hour of the hardware encoder. That is fine at the end
# of the day and expensive in the middle of one: while it runs, no capture can
# roll without dropping frames. Hence one batch, last.
#
# w3-31 is cut from E and E2 together - the second pass appends to the first,
# which is what cut.py's merge-on-lecture-id is for. w3-32 comes only from E3,
# because E's own w3-32 filmed a terminal that never received a keystroke.
set -u
cd "$(dirname "$0")"

echo "== E: the second product, first pass"
python3 cut.py E w3-27 w3-28 w3-29 w3-30 2>&1 | tail -6

echo "== E: the console, first look"
python3 cut.py E w3-31 2>&1 | tail -3

echo "== E3 + E2: the fix pass, then the console actually used"
if [ -f captures/capture-E2.json ]; then
  python3 cut.py E3,E2 w3-32 2>&1 | tail -3
else
  python3 cut.py E3 w3-32 2>&1 | tail -3
fi

for c in G H I; do
  if [ -f "captures/capture-$c.json" ]; then
    echo "== $c"
    python3 cut.py "$c" 2>&1 | tail -8
  fi
done

echo
echo "== tail frames, for eyes-on before anything is rendered"
python3 tails.py --force 2>&1 | tail -4
df -h / | tail -1
