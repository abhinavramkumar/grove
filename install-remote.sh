#!/usr/bin/env bash
set -euo pipefail

REPO="abhinavramkumar/grove"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest release tag (or use VERSION env var)
if [ -n "${VERSION:-}" ]; then
    TAG="$VERSION"
else
    TAG="$(curl -sI "https://github.com/$REPO/releases/latest" | grep -i '^location:' | sed 's|.*/||' | tr -d '\r')"
fi

if [ -z "$TAG" ]; then
    echo "Error: could not determine latest release"
    exit 1
fi

BINARY="grove-${OS}-${ARCH}"
URL="https://github.com/$REPO/releases/download/${TAG}/${BINARY}.tar.gz"

echo "Downloading grove $TAG ($OS/$ARCH)..."
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -sL "$URL" -o "$TMP/grove.tar.gz"
tar xzf "$TMP/grove.tar.gz" -C "$TMP"

echo "Installing to $INSTALL_DIR/grove..."
mkdir -p "$INSTALL_DIR"
mv "$TMP/$BINARY" "$INSTALL_DIR/grove"
chmod +x "$INSTALL_DIR/grove"

echo "grove $TAG installed successfully"
