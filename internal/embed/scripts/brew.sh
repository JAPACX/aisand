#!/bin/bash
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok()   { echo -e "${GREEN}[✓]${NC} $1"; }
info() { echo -e "${YELLOW}[ ]${NC} $1"; }

# ── Install Homebrew ─────────────────────────────────────────────────────────
export PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:$PATH"

if command -v brew &>/dev/null; then
  ok "Homebrew already installed ($(brew --version | head -1))"
else
  info "Installing system dependencies..."
  sudo apt-get update -qq
  sudo apt-get install -y -qq curl git build-essential procps file
  ok "System dependencies ready"

  info "Installing Homebrew..."
  NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  export PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:$PATH"
  ok "Homebrew installed ($(brew --version | head -1))"
fi
