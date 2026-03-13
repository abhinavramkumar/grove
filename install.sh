#!/bin/sh
# grove installer
# Downloads and installs the grove binary from GitHub releases.
# Works for fresh installs and upgrades.
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/abhinavramkumar/grove/main/install.sh | sh
#   VERSION=v0.2.0 sh install.sh
#   INSTALL_DIR=~/.local/bin sh install.sh

set -e

GITHUB_REPO="abhinavramkumar/grove"
INSTALL_DIR="${GROVE_INSTALL_DIR:-${INSTALL_DIR:-/usr/local/bin}}"
BINARY_NAME="grove"

# Color output when stdout is a TTY or when explicitly requested (e.g. from grove update)
if [ -t 1 ] || [ -n "${GROVE_INSTALL_COLOR}" ] || [ -n "${FORCE_COLOR}" ]; then
  ESC=$(printf '\033')
  C_GREEN="${ESC}[0;32m"
  C_YELLOW="${ESC}[0;33m"
  C_RED="${ESC}[0;31m"
  C_BOLD="${ESC}[1m"
  C_RESET="${ESC}[0m"
else
  C_GREEN='' C_YELLOW='' C_RED='' C_BOLD='' C_RESET=''
fi

info()    { printf '%s%s%s\n' "${C_YELLOW}" "$1" "${C_RESET}"; }
success() { printf '%s%s%s\n' "${C_GREEN}"  "$1" "${C_RESET}"; }
err()     { printf '%s%s%s\n' "${C_RED}"    "$1" "${C_RESET}"; }

detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$ARCH" in
    x86_64)        ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)             err "Unsupported architecture: $ARCH"; exit 1 ;;
  esac

  case "$OS" in
    darwin|linux) ;;
    *) err "Unsupported OS: $OS"; exit 1 ;;
  esac

  PLATFORM="${OS}-${ARCH}"
}

get_version() {
  if [ -n "${VERSION:-}" ]; then
    echo "$VERSION"
    return
  fi

  if command -v curl >/dev/null 2>&1; then
    VER=$(curl -sI "https://github.com/$GITHUB_REPO/releases/latest" \
      | grep -i '^location:' | sed 's|.*/||' | tr -d '\r\n')
  elif command -v wget >/dev/null 2>&1; then
    VER=$(wget -qO- "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" \
      | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  else
    err "curl or wget is required"
    exit 1
  fi

  if [ -z "$VER" ]; then
    err "Could not determine latest version"
    exit 1
  fi
  echo "$VER"
}

install_binary() {
  VER="$1"
  ASSET="grove-${PLATFORM}.tar.gz"
  URL="https://github.com/${GITHUB_REPO}/releases/download/${VER}/${ASSET}"

  info "Downloading grove ${VER} for ${PLATFORM}..."

  TMP_DIR=$(mktemp -d)
  trap "rm -rf $TMP_DIR" EXIT

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "${TMP_DIR}/${ASSET}" "$URL"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "${TMP_DIR}/${ASSET}" "$URL"
  fi

  tar xzf "${TMP_DIR}/${ASSET}" -C "$TMP_DIR"
  chmod +x "${TMP_DIR}/grove-${PLATFORM}"

  if [ ! -d "$INSTALL_DIR" ]; then
    info "Creating directory $INSTALL_DIR..."
    sudo mkdir -p "$INSTALL_DIR"
  fi

  if [ -w "$INSTALL_DIR" ]; then
    mv "${TMP_DIR}/grove-${PLATFORM}" "${INSTALL_DIR}/${BINARY_NAME}"
  else
    info "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv "${TMP_DIR}/grove-${PLATFORM}" "${INSTALL_DIR}/${BINARY_NAME}"
  fi
}

main() {
  printf '%s%sgrove installer%s\n' "${C_BOLD}" "${C_YELLOW}" "${C_RESET}"

  detect_platform
  info "Detected platform: $PLATFORM"

  VER=$(get_version)
  info "Version: $VER"

  # Check if already installed at this version
  if command -v grove >/dev/null 2>&1; then
    CURRENT=$(grove --version 2>/dev/null | awk '{print $2}')
    VER_NUM=$(echo "$VER" | sed 's/^v//')
    if [ "$CURRENT" = "$VER_NUM" ] || [ "$CURRENT" = "$VER" ]; then
      success "grove ${VER} is already installed"
      exit 0
    fi
    info "Upgrading from ${CURRENT} to ${VER}..."
  fi

  install_binary "$VER"
  success "grove ${VER} installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

main "$@"
