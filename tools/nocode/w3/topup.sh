#!/bin/bash
# Run the remaining top-up captures back to back, cutting between them.
#
#   ./topup.sh              # E2, G, H, I
#   ./topup.sh G H          # just these
#
# Cutting is deliberately done BETWEEN captures and never during one: cut.py
# re-encodes with h264_videotoolbox, which is the same hardware encoder the
# capture is using, and sharing it drops frames in the take that is rolling.
#
# Nothing here is parallel and nothing here is retried. A capture that fails is
# a capture to look at, not one to run again blind - every rerun costs the
# repository state that the next lecture is filmed against.
set -u
cd "$(dirname "$0")"
STAGES=${*:-"E2 E3 G H I"}

for s in $STAGES; do
  case $s in
    E3) script="shoot_e.py --only w3-32" ;;   # the fix pass, re-shot
    E2) script=shoot_e2.py ;;
    G)  script=shoot_g.py ;;
    H)  script=shoot_h.py ;;
    I)  script=shoot_i.py ;;
    *)  echo "unknown stage $s"; exit 1 ;;
  esac

  echo
  echo "================ capture $s  ($script) ================"
  date
  # shellcheck disable=SC2086
  python3 -u $script --as "$s" 2>&1 | tee "/tmp/shoot_${s}.log"

  # Deliberately NOT cutting here. cut.py re-encodes with h264_videotoolbox,
  # the same hardware encoder the next capture will be using, and it runs at
  # about one times realtime - so cutting between stages would add an hour of
  # dead machine time to a day whose scarce resource is the screen. Everything
  # is cut in one batch at the end, by `cutall.sh`.
  df -h / | tail -1
  python3 -c "import json,sys; m=json.load(open('captures/capture-$s.json')); \
print('  %d segment(s), %.1f min, %d marker(s), %d orphan(s)' % (len(m['segments']), \
sum(x['dur'] for x in m['segments'])/60, len(m['markers']), \
sum(1 for x in m['markers'] if x.get('orphan'))))" 2>/dev/null || true
done

echo
echo "================ all stages done ================"
date
python3 quota.py --live 2>&1 | tail -5
