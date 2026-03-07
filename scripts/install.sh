#!/bin/sh
set -e

REPO="nownow-labs/nownow"
BINARY="nownow"
INSTALL_DIR="/usr/local/bin"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS — on Windows use: irm https://now.ctx.st/install.ps1 | iex"; exit 1 ;;
esac

# Resolve existing install location (follows symlinks, e.g. homebrew)
UPGRADE=0
OLD_VERSION=""
if EXISTING=$(command -v "$BINARY" 2>/dev/null); then
  if [ "$OS" = "darwin" ]; then
    # macOS readlink doesn't support -f, use python3 or perl
    REAL_PATH=$(python3 -c "import os,sys;print(os.path.realpath(sys.argv[1]))" "$EXISTING" 2>/dev/null \
      || perl -MCwd -e 'print Cwd::realpath($ARGV[0])' "$EXISTING" 2>/dev/null \
      || echo "$EXISTING")
  else
    REAL_PATH=$(readlink -f "$EXISTING" 2>/dev/null || echo "$EXISTING")
  fi
  INSTALL_DIR=$(dirname "$REAL_PATH")
  OLD_VERSION=$("$BINARY" version 2>/dev/null | awk '{print $2}' || echo "unknown")
  UPGRADE=1
fi

# Get latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release"
  exit 1
fi

# Skip if already up to date (strip leading 'v' for comparison)
LATEST_CLEAN=$(echo "$LATEST" | sed 's/^v//')
if [ "$OLD_VERSION" = "$LATEST_CLEAN" ]; then
  echo "${BINARY} ${OLD_VERSION} is already up to date."
  exit 0
fi

if [ "$UPGRADE" = 1 ]; then
  echo "Upgrading ${BINARY} ${OLD_VERSION} -> ${LATEST} (${OS}/${ARCH})..."
else
  echo "Installing ${BINARY} ${LATEST} (${OS}/${ARCH})..."
fi

# Stop daemon if running (pid file at ~/.config/nownow/daemon.pid)
DAEMON_WAS_RUNNING=0
CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/nownow"
PID_FILE="${CONFIG_DIR}/daemon.pid"
if [ -f "$PID_FILE" ]; then
  DAEMON_PID=$(cat "$PID_FILE" 2>/dev/null || true)
  if [ -n "$DAEMON_PID" ] && kill -0 "$DAEMON_PID" 2>/dev/null; then
    echo "Stopping daemon (pid ${DAEMON_PID})..."
    kill "$DAEMON_PID" 2>/dev/null || true
    DAEMON_WAS_RUNNING=1
    # Wait up to 5s for graceful exit, then force kill
    for i in 1 2 3 4 5; do
      kill -0 "$DAEMON_PID" 2>/dev/null || break
      sleep 1
    done
    if kill -0 "$DAEMON_PID" 2>/dev/null; then
      kill -9 "$DAEMON_PID" 2>/dev/null || true
    fi
  fi
fi

# If daemon was stopped, ensure we restart it even if install fails
restart_daemon() {
  if [ "$DAEMON_WAS_RUNNING" = 1 ]; then
    TARGET_BIN=$(command -v "$BINARY" 2>/dev/null || echo "${INSTALL_DIR}/${BINARY}")
    if [ -x "$TARGET_BIN" ]; then
      echo "Restarting daemon..."
      "$TARGET_BIN" start >/dev/null 2>&1 || true
    fi
  fi
}
trap 'rm -rf "$TMPDIR" 2>/dev/null; restart_daemon' EXIT

URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}_${OS}_${ARCH}.tar.gz"

TMPDIR=$(mktemp -d)
curl -fsSL "$URL" -o "${TMPDIR}/${BINARY}.tar.gz"
tar -xzf "${TMPDIR}/${BINARY}.tar.gz" -C "$TMPDIR"

TARGET="${INSTALL_DIR}/${BINARY}"
if [ -w "$INSTALL_DIR" ] && ([ ! -e "$TARGET" ] || [ -w "$TARGET" ]); then
  cp "${TMPDIR}/${BINARY}" "$TARGET"
else
  echo "Need sudo to install to ${INSTALL_DIR}"
  sudo cp "${TMPDIR}/${BINARY}" "$TARGET"
fi

chmod +x "$TARGET"

# Verify
INSTALLED_VERSION=$("$TARGET" version 2>/dev/null | awk '{print $2}' || echo "unknown")
echo "Installed ${BINARY} ${INSTALLED_VERSION} to ${TARGET}"

if [ "$UPGRADE" = 0 ]; then
  echo ""
  echo "Next steps:"
  echo "  nownow login"
fi
# Daemon restart happens via EXIT trap
