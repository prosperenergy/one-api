#!/usr/bin/env bash
set -euo pipefail

# Colima + Docker setup helper for mac (Apple Silicon friendly)
# Run locally: bash scripts/colima-setup.sh

# 1) Install Homebrew (if missing) - non-interactive script may still ask for password
if ! command -v brew >/dev/null 2>&1; then
  echo "Homebrew not found. Installing Homebrew. This may prompt for your password." >&2
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  echo "After Homebrew install, ensure brew is on your PATH and re-run this script." >&2
  exit 0
fi

# 2) Install colima, docker, docker-compose
brew install colima docker docker-compose || true

# 3) Start colima (adjust cpu/memory as desired)
colima start --cpu 4 --memory 8 --disk 60

# 4) Confirm docker available
if ! command -v docker >/dev/null 2>&1; then
  echo "docker CLI not found after installing Colima. Please reopen your shell." >&2
  exit 1
fi

echo "Colima started. Run in repo root:"
echo "  docker compose pull --quiet || true"
echo "  docker compose up -d"
echo "Open UI: http://localhost:3000"
