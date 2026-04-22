#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
NC='\033[0m'

ok() { echo -e "${GREEN}[✓]${NC} $1"; }
info() { echo "[ ] $1"; }

# Install opencode via brew
if brew list anomalyco/tap/opencode &>/dev/null 2>&1; then
  ok "opencode already installed"
else
  info "Installing opencode..."
  brew install anomalyco/tap/opencode
  ok "opencode installed"
fi

# Symlink ~/.config/opencode → host virtiofs mount
HOST_HOME=$(ls /Users 2>/dev/null | head -1)
if [ -n "$HOST_HOME" ]; then
  CONFIG_SRC="/Users/$HOST_HOME/.config/opencode"
  CONFIG_DST="$HOME/.config/opencode"
  if [ -L "$CONFIG_DST" ]; then
    ok "symlink ok: $CONFIG_DST"
  else
    mkdir -p "$(dirname "$CONFIG_DST")"
    ln -sf "$CONFIG_SRC" "$CONFIG_DST"
    ok "symlink created: $CONFIG_DST → $CONFIG_SRC"
  fi

  AUTH_SRC="/Users/$HOST_HOME/.local/share/opencode/auth.json"
  AUTH_DST="$HOME/.local/share/opencode/auth.json"
  if [ -L "$AUTH_DST" ]; then
    ok "symlink ok: $AUTH_DST"
  else
    mkdir -p "$(dirname "$AUTH_DST")"
    ln -sf "$AUTH_SRC" "$AUTH_DST"
    ok "symlink created: $AUTH_DST → $AUTH_SRC"
  fi
else
  echo "Warning: Could not detect host username from /Users"
fi

ok "opencode setup complete"
