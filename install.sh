#!/bin/sh
set -e

REPO="internet-kid/tenote"
BINARY="tenote"
INSTALL_DIR="/usr/local/bin"

# ── detect OS ──────────────────────────────────────────────────────────────────
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux"  ;;
  Darwin) OS="darwin" ;;
  *)
    echo "Unsupported OS: $OS"
    echo "Download manually from: https://github.com/$REPO/releases"
    exit 1
    ;;
esac

# ── detect architecture ────────────────────────────────────────────────────────
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)          ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    echo "Download manually from: https://github.com/$REPO/releases"
    exit 1
    ;;
esac

# ── resolve version ────────────────────────────────────────────────────────────
if [ -z "$VERSION" ]; then
  VERSION="$(curl -sfL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\(.*\)".*/\1/')"
fi

if [ -z "$VERSION" ]; then
  echo "Could not determine latest version. Set VERSION manually:"
  echo "  VERSION=v1.0.0 sh install.sh"
  exit 1
fi

echo "Installing $BINARY $VERSION ($OS/$ARCH)..."

# ── download and install ───────────────────────────────────────────────────────
ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"
TMP="$(mktemp -d)"

curl -sfL "$URL" -o "$TMP/$ARCHIVE"
tar -xzf "$TMP/$ARCHIVE" -C "$TMP"
rm "$TMP/$ARCHIVE"

# ── pick install location ──────────────────────────────────────────────────────
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
elif command -v sudo >/dev/null 2>&1; then
  sudo mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
else
  # Fallback: install to ~/.local/bin
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
  mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
  echo ""
  echo "Installed to $INSTALL_DIR/$BINARY"
  echo "Make sure $INSTALL_DIR is in your PATH:"
  echo '  export PATH="$HOME/.local/bin:$PATH"'
  rm -rf "$TMP"
  exit 0
fi

rm -rf "$TMP"
echo "Installed to $INSTALL_DIR/$BINARY"
echo "Run: $BINARY"
