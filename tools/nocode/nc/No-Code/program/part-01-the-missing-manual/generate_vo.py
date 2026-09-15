#!/usr/bin/env python3
"""Generate Adam voiceover and slide sound effects via ElevenLabs."""

from __future__ import annotations

import json
import os
import sys
from pathlib import Path
import urllib.request

ROOT = Path(__file__).resolve().parent
SCRIPT = ROOT / "vo" / "script.json"
OUT = ROOT / "vo"
ENV = Path("/Users/ashish/Desktop/No-Code/.env")
DEFAULT_VOICE = "s3TPKV1kjDlVtZbl4Ksh"


def load_env() -> None:
    if not ENV.exists():
        return
    for line in ENV.read_text().splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        os.environ.setdefault(key.strip(), value.strip().strip('"').strip("'"))


def post_bytes(url: str, payload: dict, api_key: str) -> bytes:
    req = urllib.request.Request(
        url,
        data=json.dumps(payload).encode(),
        method="POST",
        headers={
            "xi-api-key": api_key,
            "Content-Type": "application/json",
            "Accept": "audio/mpeg",
        },
    )
    with urllib.request.urlopen(req) as res:
        return res.read()


def tts(text: str, dest: Path, api_key: str, voice_id: str, model_id: str) -> None:
    dest.write_bytes(
        post_bytes(
            f"https://api.elevenlabs.io/v1/text-to-speech/{voice_id}",
            {
                "text": text,
                "model_id": model_id,
                "voice_settings": {
                    "stability": 0.38,
                    "similarity_boost": 0.72,
                    "style": 0.42,
                    "use_speaker_boost": True,
                },
            },
            api_key,
        )
    )


def sfx(text: str, dest: Path, api_key: str, duration: float) -> None:
    dest.write_bytes(
        post_bytes(
            "https://api.elevenlabs.io/v1/sound-generation",
            {
                "text": text,
                "duration_seconds": duration,
                "prompt_influence": 0.55,
                "model_id": "eleven_text_to_sound_v2",
            },
            api_key,
        )
    )


def main() -> int:
    load_env()
    api_key = os.environ.get("ELEVENLABS_API_KEY", "").strip()
    if not api_key:
        print("Add ELEVENLABS_API_KEY to .env")
        return 1

    spec = json.loads(SCRIPT.read_text())
    voice_id = os.environ.get("ELEVENLABS_VOICE_ID", "").strip() or DEFAULT_VOICE
    model_id = spec.get("model_id", "eleven_multilingual_v2")
    OUT.mkdir(exist_ok=True)

    what = sys.argv[1] if len(sys.argv) > 1 else "all"

    if what in ("all", "vo"):
        for slide in spec["slides"]:
            dest = OUT / slide["file"]
            print(f"vo {slide['file']}")
            tts(slide["text"], dest, api_key, voice_id, model_id)

    if what in ("all", "sfx"):
        for item in spec.get("sfx", []):
            dest = OUT / item["file"]
            print(f"sfx {item['file']}")
            sfx(item["text"], dest, api_key, float(item["duration"]))

    print("done")
    return 0


if __name__ == "__main__":
    sys.exit(main())
