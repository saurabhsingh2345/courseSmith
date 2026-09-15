#!/usr/bin/env python3
"""Speak this part with Adam. Reads the key from the rig's existing .env."""
import json, os, sys, urllib.request, time
from pathlib import Path

ROOT = Path(__file__).resolve().parent
ENV = Path("/Users/enfecsolutions/Desktop/enfec_subs/courseSmith/tools/nocode/nc/No-Code/.env")
SPEC = ROOT / "vo" / "script.json"


def env():
    d = {}
    for line in ENV.read_text().splitlines():
        if "=" in line and not line.strip().startswith("#"):
            k, v = line.split("=", 1)
            d[k.strip()] = v.strip().strip('"').strip("'")
    return d


def main():
    e = env()
    key, voice = e["ELEVENLABS_API_KEY"], e["ELEVENLABS_VOICE_ID"]
    spec = json.loads(SPEC.read_text())
    for s in spec["slides"]:
        dest = ROOT / "vo" / s["file"]
        if dest.exists() and dest.stat().st_size > 2000:
            # The cache is keyed on the TEXT, not the filename: editing a line
            # and reusing the name used to ship the old take.
            stamp = ROOT / "vo" / (s["file"] + ".txt")
            if stamp.exists() and stamp.read_text() == s["text"]:
                print("cached", s["file"]); continue
        body = json.dumps({
            "text": s["text"],
            "model_id": spec.get("model_id", "eleven_multilingual_v2"),
            "voice_settings": {"stability": 0.42, "similarity_boost": 0.8,
                               "style": 0.15, "use_speaker_boost": True},
        }).encode()
        req = urllib.request.Request(
            f"https://api.elevenlabs.io/v1/text-to-speech/{voice}",
            data=body, method="POST",
            headers={"xi-api-key": key, "Content-Type": "application/json",
                     "Accept": "audio/mpeg"})
        for attempt in range(5):
            try:
                with urllib.request.urlopen(req, timeout=120) as r:
                    dest.write_bytes(r.read())
                break
            except Exception as exc:
                if attempt == 4:
                    raise
                print(f"  retry {attempt+1} after {exc}")
                time.sleep(2 ** attempt)
        (ROOT / "vo" / (s["file"] + ".txt")).write_text(s["text"])
        print("wrote", s["file"], dest.stat().st_size // 1024, "KB")


if __name__ == "__main__":
    main()
