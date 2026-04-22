#!/bin/bash
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok()   { echo -e "${GREEN}[✓]${NC} $1"; }
info() { echo -e "${YELLOW}[ ]${NC} $1"; }

# ── Install opencode ─────────────────────────────────────────────────────────
if command -v opencode &>/dev/null; then
  ok "opencode already installed ($(opencode --version 2>/dev/null || echo 'unknown version'))"
else
  info "Installing opencode..."
  curl -fsSL https://opencode.ai/install | bash
  ok "opencode installed"
fi

# ── Config symlinks ──────────────────────────────────────────────────────────
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
