#!/bin/bash
set -euo pipefail
VO="/Users/ashish/Desktop/No-Code/program/part-01-the-missing-manual/vo"
OUT="/Users/ashish/Desktop/No-Code/program/part-01-the-missing-manual/recordings"
RAW="$OUT/_work/part-01-raw.mov"
AUD="$OUT/_work/part-01-audio.m4a"
MP4="$OUT/part-01.mp4"
mkdir -p "$OUT/_work"

# Voice starts: each clip + 280ms gap (matches slides.js)
# 0, 8.407, 16.768, 28.332, 40.269, 52.995
ffmpeg -y \
  -i "$VO/01.mp3" -i "$VO/02.mp3" -i "$VO/03.mp3" -i "$VO/04.mp3" -i "$VO/05.mp3" -i "$VO/06.mp3" \
  -i "$VO/sfx-hit.mp3" -i "$VO/sfx-whoosh.mp3" -i "$VO/sfx-stamp.mp3" -i "$VO/sfx-rise.mp3" \
  -filter_complex "\
    [0:a]adelay=0|0[v0];\
    [1:a]adelay=8407|8407[v1];\
    [2:a]adelay=16768|16768[v2];\
    [3:a]adelay=28332|28332[v3];\
    [4:a]adelay=40269|40269[v4];\
    [5:a]adelay=52995|52995[v5];\
    [6:a]volume=0.38,adelay=0|0[h0];\
    [6:a]volume=0.38,adelay=16768|16768[h1];\
    [7:a]volume=0.38,adelay=8407|8407[w0];\
    [7:a]volume=0.38,adelay=28332|28332[w1];\
    [7:a]volume=0.38,adelay=40269|40269[w2];\
    [8:a]volume=0.40,adelay=30132|30132[st];\
    [9:a]volume=0.38,adelay=40969|40969[r0];\
    [9:a]volume=0.38,adelay=52995|52995[r1];\
    [v0][v1][v2][v3][v4][v5][h0][h1][w0][w1][w2][st][r0][r1]amix=inputs=14:normalize=0:dropout_transition=0:duration=longest\
  " -c:a aac -b:a 192k "$AUD"

osascript -e 'tell application "System Events" to set visible of process "Cursor" to false' || true

/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome \
  --autoplay-policy=no-user-gesture-required \
  --new-window \
  --start-fullscreen \
  "http://127.0.0.1:8765/part-01-the-missing-manual/slides.html?record=1&wait=1800" &
CHROME=$!
sleep 1.2

rm -f "$RAW"
screencapture -v -x -D 1 "$RAW" &
CAP=$!
sleep 68
kill -INT "$CAP" 2>/dev/null || true
wait "$CAP" || true

# Trim the wait before slides start, mux clean audio
ffmpeg -y -ss 1.6 -i "$RAW" -i "$AUD" -map 0:v:0 -map 1:a:0 \
  -c:v libx264 -pix_fmt yuv420p -preset fast -crf 18 \
  -c:a aac -b:a 192k -shortest \
  "$MP4"

ls -lh "$MP4"
ffprobe -v error -show_entries format=duration,size -of default=noprint_wrappers=1 "$MP4"
