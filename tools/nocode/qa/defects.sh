#!/bin/bash
V=/Users/enfecsolutions/Desktop/enfec_subs/courseSmith/videos/nocode/vids
for f in "$V"/*_adam.mp4; do
  b=$(basename "$f" _adam.mp4)
  d=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$f")
  log=$(ffmpeg -nostdin -v info -i "$f" \
     -af "silencedetect=noise=-45dB:d=2.5,astats=metadata=1:reset=0" \
     -f null - 2>&1)
  # total silence
  sil=$(echo "$log" | grep -o 'silence_duration: [0-9.]*' | awk '{s+=$2} END{printf "%.1f", s}')
  nsil=$(echo "$log" | grep -c 'silence_duration')
  maxsil=$(echo "$log" | grep -o 'silence_duration: [0-9.]*' | awk '{if($2>m)m=$2} END{printf "%.1f", m}')
  peak=$(echo "$log" | grep 'Peak level dB' | tail -1 | awk '{print $NF}')
  rms=$(echo "$log" | grep 'RMS level dB' | tail -1 | awk '{print $NF}')
  echo "$b|dur=$d|silence_total=${sil:-0}|silence_events=$nsil|silence_max=${maxsil:-0}|peak=$peak|rms=$rms"
done
echo DEFDONE
