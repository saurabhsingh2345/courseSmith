#!/bin/bash
# Stop a running take WITHOUT destroying it.
#
# `pkill ffmpeg` leaves an MP4 with no moov atom and the whole file is lost.
# Send SIGINT to the shoot driver so SegRec.stop() runs and ffmpeg is asked to
# quit properly. Only force-kill if that fails, and expect damage if you do.
set -u
pid=$(pgrep -f "tools/nocode/w3/shoot_" | head -1)
if [ -z "$pid" ]; then echo "no shoot running"; else
  echo "interrupting shoot driver $pid"; kill -INT "$pid"
fi
for i in $(seq 1 30); do
  pgrep -f "avfoundation" >/dev/null || { echo "capture closed cleanly"; exit 0; }
  sleep 1
done
echo "capture still running after 30s — inspect before forcing"
