#!/usr/bin/env bash
set -e

REPO="fdrolshagen/jetter"
INSTALL_DIR="/usr/local/bin"
VERSION="${1:-latest}"

echo " > Installing jetter ($VERSION)..."

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
esac

if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -sL \
    -H "Accept: application/vnd.github.v3+json" \
    -H "User-Agent: install-script" \
    "https://api.github.com/repos/$REPO/tags" \
    | grep '"name":' \
    | head -n 1 \
    | cut -d '"' -f4)
fi

BINARY_NAME="jetter-$VERSION-$OS"
if [[ "$ARCH" != "amd64" ]]; then
  BINARY_NAME="$BINARY_NAME-$ARCH"
fi

if [[ "$OS" == "windows" ]]; then
  BINARY_NAME="$BINARY_NAME.exe"
fi

BINARY_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

echo " > Downloading binary..."
curl -L --fail "$BINARY_URL" -o "$BINARY_NAME"

echo " > Downloading checksums..."
curl -L --fail "$CHECKSUM_URL" -o checksums.txt

echo " > Verifying checksum..."
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum --ignore-missing -c <(grep " ${BINARY_NAME}$" checksums.txt)
elif command -v shasum >/dev/null 2>&1; then
  grep " ${BINARY_NAME}$" checksums.txt | shasum -a 256 -c -
else
  echo " !>  Warning: no checksum utility found. Skipping verification."
fi

chmod +x "$BINARY_NAME"
echo " > Installing to $INSTALL_DIR..."
sudo mv "$BINARY_NAME" "$INSTALL_DIR/jetter"

echo " > jetter $VERSION installed successfully!"
echo " > Restart your shell and run 'jetter --help'"
