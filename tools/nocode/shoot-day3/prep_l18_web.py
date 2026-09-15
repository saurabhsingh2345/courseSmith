#!/usr/bin/env python3
"""Stage Chrome on the finished Copilot build for the five checks.

Three things this has to get right, all learned the hard way:
  * a NEW window with ONE tab — his normal window has eleven tabs whose titles
    name Gmail, GitHub and WhatsApp, and fullscreen does not hide a tab strip.
  * pre-navigate with AppleScript BEFORE recording. Driving the address bar in
    fullscreen uses toolbar coordinates that are stale seconds after you
    measure them.
  * no app hiding and no quitting anything. Fullscreen on the shoot panel is
    what keeps the frame clean.
"""
import os, sys, time, subprocess
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "scripts"))
import stage as S

URL = "http://localhost:3002"
OUT = os.path.dirname(os.path.abspath(__file__))

S.osa('tell application "Google Chrome" to activate')
time.sleep(1.5)
S.osa(f'tell application "Google Chrome" to make new window')
time.sleep(1.5)
S.osa(f'tell application "Google Chrome" to set URL of active tab of front window to "{URL}"')
time.sleep(6)
S.wait_front("Google Chrome")
time.sleep(1.0)
S.clear_mods()

title = S.osa('tell application "Google Chrome" to get title of active tab of front window')
tabs = S.osa('tell application "Google Chrome" to get number of tabs of front window')
print("front window tabs:", tabs, "title:", repr(title))
if tabs.strip() not in ("1",):
    raise SystemExit(f"REFUSING: front Chrome window has {tabs} tabs, not 1")

S.move_window_to_panel("Google Chrome")
time.sleep(2.5)
S.move(960, 560, 0.5)
time.sleep(1.2)
print("frontmost:", S.frontmost())
print(S.snap(os.path.join(OUT, "l18_web_pre.png")))
