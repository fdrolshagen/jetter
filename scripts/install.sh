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

FILENAME="jetter_${OS}_${ARCH}"
BINARY_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

echo $BINARY_URL

TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

echo " > Downloading binary..."
curl -L --fail "$BINARY_URL" -o "$FILENAME"

echo " > Downloading checksums..."
curl -L --fail "$CHECKSUM_URL" -o checksums.txt

echo " > Verifying checksum..."
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum --ignore-missing -c <(grep " ${FILENAME}$" checksums.txt)
elif command -v shasum >/dev/null 2>&1; then
  grep " ${FILENAME}$" checksums.txt | shasum -a 256 -c -
else
  echo " !>  Warning: no checksum utility found. Skipping verification."
fi

chmod +x "$FILENAME"
echo "Installing to $INSTALL_DIR (sudo may be required)..."
sudo mv "$FILENAME" "$INSTALL_DIR/jetter"

echo " > jetter $VERSION installed successfully!"
echo "You can now run 'jetter --help'"
