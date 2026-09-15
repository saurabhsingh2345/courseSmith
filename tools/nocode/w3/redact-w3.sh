#!/bin/bash
# Blur the two places `gh` puts a personal account name on screen.
#
# Run AFTER cutall.sh and BEFORE rendering w3-34 or w3-35.
# Boxes were read off frames of the raw take, not guessed. Originals are kept as
# <lecture>.orig.mp4, so a bad box is one `mv` away from being undone.
#
#   w3-35  `gh pr view` prints
#              Open - <username> wants to merge 1 commit into main from ...
#          and it STAYS on screen from the render until `gh pr diff` scrolls it
#          away. That is about twenty-three seconds of a legible personal name,
#          so the blur is a tight box on the name itself.
#
#   w3-34  `gh issue view 1` prints a header with the same name, but the body is
#          long enough that the whole thing renders and scrolls past it inside
#          about a tenth of a second. It is one to three frames, mid-scroll, and
#          unreadable - but "unreadable" is not the rule. The rule is that it is
#          not there. So the whole terminal is blurred for six tenths of a
#          second across the scroll, which is indistinguishable from the motion
#          already on screen.
set -u
cd "$(dirname "$0")"

echo "== w3-35: the pull request header"
python3 redact.py w3-35 --box 18.0 47.0 750 988 350 60

echo "== w3-34: the issue header, across the scroll"
python3 redact.py w3-34 --box 176.0 177.0 540 580 3280 1520

echo
echo "Now LOOK at both:"
echo "  ffmpeg -ss 25 -i cuts/w3-35.mp4 -frames:v 1 /tmp/chk35.png"
echo "  ffmpeg -ss 176.5 -i cuts/w3-34.mp4 -frames:v 1 /tmp/chk34.png"
