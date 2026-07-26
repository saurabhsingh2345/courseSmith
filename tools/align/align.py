#!/usr/bin/env python3
"""Word-level alignment of a coursesmith voiceover using whisperX.

Usage:
    python align.py --audio voiceover.wav [--model small] [--language en]

Prints JSON to stdout:
    {"words": [{"word": "Hello", "startMs": 120, "endMs": 400}, ...]}

Diagnostics go to stderr; a non-zero exit means the Go engine falls back to
segment-level timing. Install (from tools/align/):
    uv sync            # or: pip install -e .
"""

from __future__ import annotations

import argparse
import json
import sys


def log(msg: str) -> None:
    print(f"align: {msg}", file=sys.stderr, flush=True)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--audio", required=True, help="path to voiceover.wav")
    parser.add_argument("--model", default="small", help="whisper model size")
    parser.add_argument("--language", default="en")
    args = parser.parse_args()

    # whisperX and its dependencies print progress to stdout; the Go engine
    # parses stdout as JSON, so route all library chatter to stderr and keep
    # the real stdout for the final payload only.
    payload_out = sys.stdout
    sys.stdout = sys.stderr

    try:
        # PyTorch >= 2.6 defaults torch.load to weights_only=True, which
        # rejects whisperX's pyannote VAD/alignment checkpoints (they pickle
        # omegaconf containers). The models come from trusted Hugging Face
        # repos pinned by whisperx, so restore the old behavior.
        import torch

        _orig_load = torch.load

        def _load(*args, **kwargs):  # noqa: ANN002, ANN003
            # Force, not setdefault: lightning passes weights_only=True
            # explicitly on torch >= 2.6.
            kwargs["weights_only"] = False
            return _orig_load(*args, **kwargs)

        torch.load = _load

        import whisperx  # noqa: PLC0415 — import late for a fast --help
    except ImportError as exc:  # pragma: no cover
        log(f"whisperx is not installed ({exc}); run `uv sync` in tools/align/")
        return 2

    device = "cpu"
    log(f"loading whisper model {args.model!r} ({device}, int8)")
    model = whisperx.load_model(
        args.model, device, compute_type="int8", language=args.language
    )
    audio = whisperx.load_audio(args.audio)

    log("transcribing")
    result = model.transcribe(audio, batch_size=8, language=args.language)

    log("aligning words")
    align_model, metadata = whisperx.load_align_model(
        language_code=args.language, device=device
    )
    aligned = whisperx.align(
        result["segments"], align_model, metadata, audio, device,
        return_char_alignments=False,
    )

    words = []
    for w in aligned.get("word_segments", []):
        # whisperX omits timestamps for words it could not align (e.g. digits);
        # skip those — the Go side interpolates across gaps.
        if "start" not in w or "end" not in w:
            continue
        text = w.get("word", "").strip()
        if not text:
            continue
        words.append(
            {
                "word": text,
                "startMs": int(round(w["start"] * 1000)),
                "endMs": int(round(w["end"] * 1000)),
            }
        )

    if not words:
        log("alignment produced no words")
        return 3

    json.dump({"words": words}, payload_out)
    payload_out.flush()
    return 0


if __name__ == "__main__":
    sys.exit(main())
