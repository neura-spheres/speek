#!/usr/bin/env bash
set -e

REPO="neura-spheres/speek"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  linux|darwin) ;;
  *)
    echo "Unsupported OS: $OS"
    echo "Windows users: run this in PowerShell:"
    echo '  irm https://raw.githubusercontent.com/neura-spheres/speek/main/install.ps1 | iex'
    exit 1
    ;;
esac

# Get latest release version
echo "Fetching latest Speek release..."
VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' \
  | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "Could not determine latest version. Check your internet connection."
  exit 1
fi

FILENAME="speek-${OS}-${ARCH}"
URL="https://github.com/$REPO/releases/download/$VERSION/${FILENAME}"

echo "Downloading Speek $VERSION for $OS/$ARCH..."
TMP=$(mktemp -d)
curl -sL "$URL" -o "$TMP/speek"
chmod +x "$TMP/speek"

echo "Installing to $INSTALL_DIR/speek..."
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/speek" "$INSTALL_DIR/speek"
else
  sudo mv "$TMP/speek" "$INSTALL_DIR/speek"
fi

rm -rf "$TMP"

echo ""
echo "Speek $VERSION installed!"
echo ""

# Auto-install VS Code extension if VS Code is found
if speek install-vscode 2>/dev/null; then
  :
else
  echo "VS Code not detected. If you install it later, run:"
  echo "  speek install-vscode"
fi

echo ""
echo "Try it:"
echo "  speek repl"
echo "  echo 'show \"Hello, world!\"' > hello.spk && speek run hello.spk"
