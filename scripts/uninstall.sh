#!/bin/sh
# eng uninstaller — removes the eng binary installed by install.sh.
#
# Remote:
#   curl -fsSL https://github.com/JhonMA82/engineering-platform/releases/latest/download/uninstall.sh | sh
# Local:
#   sh scripts/uninstall.sh [--dir DIR]
#
# Env overrides: ENG_INSTALL_DIR (default ~/.local/bin).
# Removes exactly one resolved path and reports it; never touches
# projects, workspaces or ~/.engineering state (project-local by design).
set -eu

INSTALL_DIR="${ENG_INSTALL_DIR:-$HOME/.local/bin}"
BIN_NAME="eng"

while [ $# -gt 0 ]; do
  case "$1" in
    --dir) INSTALL_DIR="${2:?--dir needs a value}"; shift 2 ;;
    -h|--help) echo "usage: uninstall.sh [--dir DIR]" >&2; exit 0 ;;
    *) echo "uninstall.sh: unknown flag $1" >&2; echo "usage: uninstall.sh [--dir DIR]" >&2; exit 2 ;;
  esac
done

TARGET="$INSTALL_DIR/$BIN_NAME"
if [ ! -e "$TARGET" ]; then
  echo "uninstall: $TARGET not present — nothing to remove."
  exit 0
fi
rm -f "$TARGET"
echo "uninstalled $TARGET"
