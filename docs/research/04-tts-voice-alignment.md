# Research: open-source TTS, voice cloning, prosody control, alignment (2025–2026)

_Web-research report, 2026-07-18. Mac/Apple-Silicon feasibility and commercial licensing
are first-class filters._

## Arena context
Open-weight vs commercial Elo gap narrowed to ~81 (2026). **Kokoro (~1056 Elo) is still
remarkably competitive per-parameter** — the wins elsewhere are cloning + expressiveness,
not raw naturalness. Chatterbox beat ElevenLabs 63.75% in blind preference.

## Key models (commercial-safe, Mac-viable in bold)

- **Kokoro-82M (current; Apache 2.0)** — no cloning; 54 voices. UNDEREXPLOITED CONTROL via
  Kokoro-FastAPI: `speed` 0.25–4x; **voice blending** `af_bella(2)+af_sky(1)`; **exact IPA
  pronunciation** `[Kokoro](/kˈOkəɹO/)`; **per-word stress** `[word](+2)`; **native
  word-level timestamps via `/dev/captioned_speech`** (could replace whisperX for Kokoro
  audio — exact by construction). CPU real-time; MLX/CoreML ports exist.
- **Chatterbox + Chatterbox Turbo (Resemble AI; MIT)** — THE headline upgrade. Zero-shot
  cloning from 5–20s; `exaggeration` + `cfg_weight` knobs; Turbo (350M, Dec 2025) adds
  `[laugh] [sigh] [chuckle]`… tags in the cloned voice. **MPS support** via
  devnen/Chatterbox-TTS-Server (`device: mps`); travisvn/chatterbox-tts-api = drop-in
  OpenAI-compatible `/v1/audio/speech`. PerTh watermark built in. 23-language variant.
- **NeuTTS Air (Neuphonic; Apache 2.0, 748M)** — GGUF/llama.cpp, **real-time on CPU**,
  instant cloning from ~3s. Best lightweight Mac cloning; quality below Chatterbox.
- **Orpheus 3B (Apache 2.0)** — trained emotion tags `<laugh> <sigh>`…; zero-shot cloning;
  GGUF + Metal via LM Studio/llama.cpp + SNAC decoder; community OpenAI-compatible servers.
- **Dia-1.6B (Apache 2.0)** — dialogue-native, nonverbals; MPS works; voice identity
  unstable without fixed audio prompt — awkward for consistent narration.
- **Zonos (Apache 2.0)** — 8-dim emotion vector + rate + pitch (most SSML-like parametric
  surface); CUDA-centric, painful on macOS.
- Maya1 (Apache 2.0, 3B): voice design by text description + 20+ tags; no sample cloning;
  ~16GB-GPU class.

## License traps (avoid for commercial course content)
F5-TTS (weights CC-BY-NC), Fish/OpenAudio S1-mini (CC-BY-NC-SA), XTTS-v2 (CPML, Coqui
dead — unpurchasable), Voxtral TTS (CC-BY-NC), Spark-TTS (NC weights), IndexTTS-2
(weights need Bilibili permission), MegaTTS3 (voice encoder withheld), Higgs (NC-ish),
VibeVoice (code pulled by Microsoft), Piper (now GPL-3.0, quality < Kokoro).

## Prosody control idioms (no open model does real W3C SSML)
- Inline trained tags: Orpheus, Maya1, Chatterbox Turbo.
- Scalar knobs: Chatterbox exaggeration/cfg; Zonos vectors.
- Reference-audio emotion: IndexTTS-2 (separate emotion vs timbre refs), Dia, F5.
- **Phoneme/stress markup: Kokoro/Misaki is the BEST here** (IPA + stress levels).
- Pauses: universally weak → handle at pipeline level (per-segment synthesis + spliced
  measured silence — already how courseSmith works).

## Audio post
- **resemble-enhance (MIT)**: denoise + bandwidth-extend to 44.1kHz — closest open "Adobe
  Podcast Enhance"; CPU-slow but batchable.
- **DeepFilterNet (MIT/Apache, Rust)**: real-time noise suppression — clean cloning
  reference samples.
- Mastering: existing ffmpeg loudnorm chain is correct; keep.

## Forced alignment
- Benchmark (Interspeech 2024, TIMIT @20ms): **MFA 72.8% vs WhisperX 52.7%** word accuracy.
- **Crucial insight: the script is known — ASR is wasted work; pure forced alignment
  suffices.**
- **ctc-forced-aligner** (MMS-300M CTC): text+audio → ms word boundaries; INT8 ~340MB,
  fast on CPU; 1130+ languages. Best drop-in Mac upgrade.
- **Kokoro `/dev/captioned_speech`**: timestamps from synthesis itself — zero-cost, exact,
  eliminates alignment for Kokoro narration.
- MFA: most accurate, heavy workflow. NeMo NFA: CUDA/Linux-oriented. stable-ts: same
  accuracy class as whisperX.

## Recommended pipeline fit
| Slot | Server | Backend |
|---|---|---|
| Primary upgrade | chatterbox-tts-api / Chatterbox-TTS-Server (MPS) | **Chatterbox Turbo** — user's cloned voice + emotion tags, drop-in /v1/audio/speech |
| Keep | Kokoro-FastAPI :8880 | Kokoro — fast/deterministic; exploit blending + IPA + stress + speed + native timestamps |
| Lightweight cloning | llama.cpp + GGUF | NeuTTS Air |
| Expressive alt | Orpheus-FastAPI | Orpheus 3B |

Alignment: Kokoro → native captioned_speech timestamps; cloned-voice audio →
ctc-forced-aligner with whisperX fallback.
