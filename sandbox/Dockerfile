# coursesmith execution sandbox: verifies lesson code and records VHS
# terminal demos. Build from the repo root:
#
#   docker build -t coursesmith-sandbox sandbox/
#
# python3.12  — executes lesson code blocks (verify stage)
# vhs+ttyd    — records terminal demos (demos stage)
# ffmpeg      — vhs renders MP4s with it
FROM python:3.12-slim

# ttyd is not in Debian's repos; install the static binary from GitHub.
ARG TTYD_VERSION=1.7.7

# chromium: vhs renders the terminal through a headless browser, and the
# sandbox runs with --network none, so it must be preinstalled.
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates curl gnupg ffmpeg bash chromium fonts-dejavu-core \
    && arch="$(dpkg --print-architecture)" \
    && case "$arch" in \
         amd64) ttyd_arch=x86_64 ;; \
         arm64) ttyd_arch=aarch64 ;; \
         *) echo "unsupported arch: $arch" && exit 1 ;; \
       esac \
    && curl -fsSL -o /usr/local/bin/ttyd \
         "https://github.com/tsl0922/ttyd/releases/download/${TTYD_VERSION}/ttyd.${ttyd_arch}" \
    && chmod +x /usr/local/bin/ttyd \
    && mkdir -p /etc/apt/keyrings \
    && curl -fsSL https://repo.charm.sh/apt/gpg.key | gpg --dearmor -o /etc/apt/keyrings/charm.gpg \
    && echo "deb [signed-by=/etc/apt/keyrings/charm.gpg] https://repo.charm.sh/apt/ * *" > /etc/apt/sources.list.d/charm.list \
    && apt-get update && apt-get install -y --no-install-recommends vhs \
    && apt-get purge -y gnupg && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/*

# pytest — verifies practice-exercise solutions against their hidden tests
# (exercises stage).
RUN pip install --no-cache-dir pytest

# vhs runs as root inside the container; Chromium needs the sandbox off.
ENV VHS_NO_SANDBOX=true

WORKDIR /work
