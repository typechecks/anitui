#!/bin/sh
set -e

# anitui install script
# downloads the latest binary for your platform

REPO="typechecks/anitui"
INSTALL_DIR="/usr/local/bin"

if [ "$(id -u)" -ne 0 ]; then
    echo "Please run as root (use sudo)"
    exit 1
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Detecting latest version..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo "Could not find latest release tag"
    exit 1
fi

BINARY_NAME="anitui_${OS}_${ARCH}"
[ "$OS" = "darwin" ] && BINARY_NAME="anitui_darwin_${ARCH}"

URL="https://github.com/$REPO/releases/download/$LATEST_TAG/${BINARY_NAME}.tar.gz"

echo "Downloading anitui $LATEST_TAG for $OS/$ARCH..."
curl -L "$URL" -o /tmp/anitui.tar.gz

echo "Installing to $INSTALL_DIR..."
rm -f "$INSTALL_DIR/anitui.old"
tar -xzf /tmp/anitui.tar.gz -C /tmp
mv /tmp/anitui "$INSTALL_DIR/anitui"
chmod +x "$INSTALL_DIR/anitui"

rm /tmp/anitui.tar.gz
echo "Done! Run 'anitui' to start."
