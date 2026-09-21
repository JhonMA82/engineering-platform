#!/bin/sh
# eng installer — remote install and remote update from GitHub Releases.
#
# Install (latest):
#   curl -fsSL https://github.com/JhonMA82/engineering-platform/releases/latest/download/install.sh | sh
# Update to latest (same command — idempotent):
#   curl -fsSL https://github.com/JhonMA82/engineering-platform/releases/latest/download/install.sh | sh
# Pin a version:
#   curl -fsSL .../releases/download/v1.4.0/install.sh | sh -s -- --version 1.4.0
#
# Env overrides: ENG_GITHUB_REPO (fork slug), ENG_VERSION (X.Y.Z or latest),
# ENG_INSTALL_DIR (default ~/.local/bin), ENG_NO_MODIFY_PATH=1.
#
# Every binary is verified against the release checksums.txt before it
# touches the install dir. No sudo, no package manager, no surprises.
set -eu

REPO="${ENG_GITHUB_REPO:-JhonMA82/engineering-platform}"
VERSION="${ENG_VERSION:-latest}"
INSTALL_DIR="${ENG_INSTALL_DIR:-$HOME/.local/bin}"
BIN_NAME="eng"

usage() {
  echo "usage: install.sh [--version X.Y.Z|latest] [--dir DIR] [--repo OWNER/NAME]" >&2
}

while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="${2:?--version needs a value}"; shift 2 ;;
    --dir) INSTALL_DIR="${2:?--dir needs a value}"; shift 2 ;;
    --repo) REPO="${2:?--repo needs a value}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "install.sh: unknown flag $1" >&2; usage; exit 2 ;;
  esac
done

need() {
  command -v "$1" >/dev/null 2>&1 || { echo "install.sh: need '$1' on PATH" >&2; exit 1; }
}
need curl
need uname
if command -v sha256sum >/dev/null 2>&1; then
  SHA256="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  SHA256="shasum -a 256"
else
  echo "install.sh: need 'sha256sum' or 'shasum' on PATH" >&2; exit 1
fi

OS="$(uname -s)"
ARCH="$(uname -m)"
case "$OS/$ARCH" in
  Linux/x86_64|Linux/amd64) ASSET="eng-linux-amd64" ;;
  Linux/aarch64|Linux/arm64) ASSET="eng-linux-arm64" ;;
  Darwin/arm64) ASSET="eng-darwin-arm64" ;;
  *) echo "install.sh: unsupported platform $OS/$ARCH (releases publish linux/amd64, linux/arm64, darwin/arm64, windows/amd64)" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  # Follow the /releases/latest redirect — no API token, no rate limit.
  EFFECTIVE="$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest")"
  TAG="${EFFECTIVE##*/}"
  [ -n "$TAG" ] || { echo "install.sh: could not resolve latest release" >&2; exit 1; }
else
  case "$VERSION" in
    v*) TAG="$VERSION" ;;
    *) TAG="v$VERSION" ;;
  esac
fi

BASE="https://github.com/$REPO/releases/download/$TAG"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT INT TERM

echo "installing $BIN_NAME $TAG ($ASSET) from $REPO..."
curl -fsSL --retry 3 -o "$TMP/$ASSET" "$BASE/$ASSET"
curl -fsSL --retry 3 -o "$TMP/checksums.txt" "$BASE/checksums.txt"

WANT="$(awk -v a="$ASSET" '$2==a {print $1}' "$TMP/checksums.txt")"
[ -n "$WANT" ] || { echo "install.sh: checksums.txt does not cover $ASSET" >&2; exit 1; }
GOT="$( (cd "$TMP" && $SHA256 "$ASSET") | awk '{print $1}')"
if [ "$GOT" != "$WANT" ]; then
  echo "install.sh: checksum mismatch for $ASSET (download corrupted or tampered)" >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"
# Atomic replace in the same dir: temp + rename, never a half-written binary.
cp "$TMP/$ASSET" "$INSTALL_DIR/.$BIN_NAME.new"
chmod +x "$INSTALL_DIR/.$BIN_NAME.new"
mv -f "$INSTALL_DIR/.$BIN_NAME.new" "$INSTALL_DIR/$BIN_NAME"

echo "installed $INSTALL_DIR/$BIN_NAME ($TAG)"
"$INSTALL_DIR/$BIN_NAME" version || true
if command -v "$BIN_NAME" >/dev/null 2>&1; then
  ON_PATH="$(command -v "$BIN_NAME")"
  if [ "$ON_PATH" != "$INSTALL_DIR/$BIN_NAME" ]; then
    echo "note: 'eng' on PATH is $ON_PATH (the new binary is $INSTALL_DIR/$BIN_NAME)"
  fi
else
  case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *)
      echo "add $INSTALL_DIR to your PATH, e.g.:"
      echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
      ;;
  esac
fi
