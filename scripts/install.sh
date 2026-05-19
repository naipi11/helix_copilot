#!/usr/bin/env bash
set -euo pipefail

REPO="naipi11/helix_copilot"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-latest}"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required" >&2
  exit 1
fi

if [ "$VERSION" = "latest" ]; then
  URL="https://github.com/$REPO/releases/latest/download/helix-copilot_${OS}_${ARCH}.tar.gz"
else
  URL="https://github.com/$REPO/releases/download/$VERSION/helix-copilot_${OS}_${ARCH}.tar.gz"
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$BIN_DIR"
echo "Downloading $URL"
curl -fsSL "$URL" -o "$tmp/helix-copilot.tar.gz"
tar -xzf "$tmp/helix-copilot.tar.gz" -C "$tmp"
install -m 0755 "$tmp/helix-copilot" "$BIN_DIR/helix-copilot"
if [ -f "$tmp/hx" ]; then
  install -m 0755 "$tmp/hx" "$BIN_DIR/hx"
fi

echo "Installed helix-copilot to $BIN_DIR"
if ! command -v node >/dev/null 2>&1; then
  echo "warning: Node.js is required for @github/copilot-language-server" >&2
fi
