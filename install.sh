#!/bin/sh
set -e

REPO="mabd-dev/reposcan"
BINARY_NAME="reposcan"

# ── 1. Detect OS ────────────────────────────────────────────────────────────
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)
    echo "error: unsupported operating system: $OS" >&2
    echo "       supported: linux, darwin" >&2
    exit 1
    ;;
esac

# ── 2. Detect Arch ──────────────────────────────────────────────────────────
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "error: unsupported architecture: $ARCH" >&2
    echo "       supported: amd64, arm64 (darwin only)" >&2
    exit 1
    ;;
esac

# ── 3. Guard unsupported combos ─────────────────────────────────────────────
if [ "$OS" = "linux" ] && [ "$ARCH" = "arm64" ]; then
  echo "error: linux/arm64 is not supported yet." >&2
  exit 1
fi

# ── 4. Fetch latest release version from GitHub API ─────────────────────────
echo "Fetching latest release..."
LATEST_URL="https://api.github.com/repos/${REPO}/releases/latest"

if command -v curl >/dev/null 2>&1; then
  VERSION="$(curl -fsSL "$LATEST_URL" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
elif command -v wget >/dev/null 2>&1; then
  VERSION="$(wget -qO- "$LATEST_URL" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
else
  echo "error: curl or wget is required to install reposcan." >&2
  exit 1
fi

if [ -z "$VERSION" ]; then
  echo "error: could not determine the latest release version." >&2
  exit 1
fi

echo "Latest version: $VERSION"

# ── 5. Build download URL ────────────────────────────────────────────────────
ASSET="${BINARY_NAME}-${VERSION}-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

echo "Downloading $ASSET..."

TMP_DIR="$(mktemp -d)"
TMP_BIN="${TMP_DIR}/${BINARY_NAME}"
trap 'rm -rf "$TMP_DIR"' EXIT

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$DOWNLOAD_URL" -o "$TMP_BIN"
else
  wget -qO "$TMP_BIN" "$DOWNLOAD_URL"
fi

chmod +x "$TMP_BIN"

# ── 6. Find a writable directory on $PATH ────────────────────────────────────
find_install_dir() {
  # Prefer ~/.local/bin (no sudo needed), then /usr/local/bin
  CANDIDATES="$HOME/.local/bin /usr/local/bin /usr/bin"
  for DIR in $CANDIDATES; do
    if echo "$PATH" | grep -q "$DIR" && [ -w "$DIR" ]; then
      echo "$DIR"
      return
    fi
  done

  # ~/.local/bin is on PATH but doesn't exist yet — create it
  LOCAL_BIN="$HOME/.local/bin"
  if echo "$PATH" | grep -q "$LOCAL_BIN"; then
    mkdir -p "$LOCAL_BIN"
    echo "$LOCAL_BIN"
    return
  fi

  echo ""
}

INSTALL_DIR="$(find_install_dir)"

if [ -z "$INSTALL_DIR" ]; then
  echo "error: could not find a writable directory on your \$PATH." >&2
  echo "       Add ~/.local/bin to your PATH and re-run, or install manually:" >&2
  echo "       sudo mv $TMP_BIN /usr/local/bin/${BINARY_NAME}" >&2
  exit 1
fi

mv "$TMP_BIN" "${INSTALL_DIR}/${BINARY_NAME}"

echo ""
echo "reposcan $VERSION installed to ${INSTALL_DIR}/${BINARY_NAME}"
echo "Run 'reposcan --help' to get started."
