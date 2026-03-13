#!/usr/bin/env bash
set -euo pipefail

# Default install directory
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Parse --prefix flag
while [[ $# -gt 0 ]]; do
    case "$1" in
        --prefix) INSTALL_DIR="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Build
echo "Building grove..."
go build -o grove ./cmd/grove

# Install (handles upgrade by overwriting)
echo "Installing to $INSTALL_DIR/grove..."
mkdir -p "$INSTALL_DIR"
mv grove "$INSTALL_DIR/grove"

# Verify
if command -v grove &>/dev/null; then
    echo "grove installed successfully"
else
    echo "Installed to $INSTALL_DIR/grove (may need to add to PATH)"
fi
