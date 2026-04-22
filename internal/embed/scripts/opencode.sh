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

# ── inotify limits (required for opencode FileWatcher) ───────────────────────
# opencode registers inotify watches on every subdirectory of the project.
# The default kernel limits (128 instances / 8192 watches) are too low and
# cause opencode to hang silently on startup with a blank screen.
info "Configuring inotify limits..."
SYSCTL_CONF=/etc/sysctl.d/99-opencode.conf
if [ ! -f "$SYSCTL_CONF" ]; then
  cat <<'EOF' | sudo tee "$SYSCTL_CONF" > /dev/null
# Raised for opencode FileWatcher — prevents blank-screen hang on startup
fs.inotify.max_user_instances=2048
fs.inotify.max_user_watches=524288
EOF
  sudo sysctl -p "$SYSCTL_CONF" 2>/dev/null || true
  ok "inotify limits set (instances=2048, watches=524288)"
else
  ok "inotify limits already configured"
fi

# ── Native cache dir (keep bun/opencode cache off the 9p mount) ──────────────
# Lima mounts the host home via 9p/sshfs. Writing bun package cache over 9p
# is very slow and can cause opencode's first launch to take minutes.
# We ensure ~/.cache/opencode is a real directory on the VM's native ext4 fs,
# not a symlink into the host mount.
if [ -L "$HOME/.cache/opencode" ]; then
  info "Replacing 9p symlink ~/.cache/opencode with native directory..."
  rm "$HOME/.cache/opencode"
  mkdir -p "$HOME/.cache/opencode"
  ok "~/.cache/opencode is now native (fast)"
elif [ ! -d "$HOME/.cache/opencode" ]; then
  mkdir -p "$HOME/.cache/opencode"
  ok "~/.cache/opencode created (native)"
else
  ok "~/.cache/opencode is already native"
fi

# Same for bun's own cache, which opencode uses internally
if [ -L "$HOME/.cache/bun" ]; then
  rm "$HOME/.cache/bun"
  mkdir -p "$HOME/.cache/bun"
  ok "~/.cache/bun is now native (fast)"
elif [ ! -d "$HOME/.cache/bun" ]; then
  mkdir -p "$HOME/.cache/bun"
fi

# ── Config symlinks (skipped if SKIP_SYMLINKS=1) ─────────────────────────────
if [ "${SKIP_SYMLINKS:-0}" = "1" ]; then
  ok "Skipping host config symlinks"
else
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
fi

echo ""
echo -e "${YELLOW}Note:${NC} The first launch of opencode may take 1-2 minutes"
echo "      while it downloads provider packages (bun deps). This is normal."
echo "      Subsequent launches will be fast."
echo ""
ok "opencode setup complete"
