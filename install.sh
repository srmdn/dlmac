#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DLMAC_PATH="${SCRIPT_DIR}/dlmac"
WEB_HELPER_PATH="${SCRIPT_DIR}/dlmac-web"
web_helper_tmp=""

cleanup() {
  if [[ -n "$web_helper_tmp" ]]; then
    rm -f "$web_helper_tmp"
  fi
}

trap cleanup EXIT

echo "dlmac installer"
echo "==============="
echo ""

if [[ "$(uname)" != "Darwin" ]]; then
  echo "Error: dlmac is macOS only."
  exit 1
fi

if ! command -v brew &>/dev/null; then
  echo "Error: Homebrew not found."
  echo "Install it from: https://brew.sh"
  echo ""
  echo "Then run: brew install yt-dlp ffmpeg go"
  exit 1
fi

missing=()

if ! command -v yt-dlp &>/dev/null; then
  missing+=("yt-dlp")
fi

if ! command -v ffmpeg &>/dev/null; then
  missing+=("ffmpeg")
fi

if ! command -v go &>/dev/null; then
  missing+=("go")
fi

if [[ ${#missing[@]} -gt 0 ]]; then
  echo "Missing dependencies: ${missing[*]}"
  echo ""
  echo "Install with:"
  echo "  brew install ${missing[*]}"
  echo ""
  read -rp "Install now? [y/N] " answer
  if [[ "$answer" =~ ^[Yy]$ ]]; then
    brew install "${missing[@]}"
  else
    echo "Install manually and re-run install.sh."
    exit 1
  fi
fi

echo "Building local web helper..."
web_helper_tmp="$(mktemp "${WEB_HELPER_PATH}.tmp.XXXXXX")"

(
  cd "$SCRIPT_DIR"
  go build -o "$web_helper_tmp" ./cmd/dlmac-web
)

chmod 755 "$web_helper_tmp"
mv -f "$web_helper_tmp" "$WEB_HELPER_PATH"
web_helper_tmp=""

chmod +x "$DLMAC_PATH"

echo ""
echo "Done. Run ./dlmac to see usage."
echo "Built web helper: ${WEB_HELPER_PATH}"
echo "Add to PATH for global access (optional):"
echo "  export PATH=\"\$PATH:${SCRIPT_DIR}\""
