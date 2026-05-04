#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINDIR="${VOXTYPE_TUI_BINDIR:-$HOME/.local/bin}"

RED='\033[1;31m'
GREEN='\033[1;32m'
CYAN='\033[1;36m'
BOLD='\033[1m'
NC='\033[0m'

ok() {
  echo -e "${GREEN}[ok]${NC} $1"
}

info() {
  echo -e "${CYAN}[info]${NC} $1"
}

err() {
  echo -e "${RED}[error]${NC} $1" >&2
}

echo ""
echo -e "${CYAN}=== Voxtype TUI Setup ===${NC}"
echo -e "Source: ${BOLD}$SCRIPT_DIR${NC}"
echo -e "Install: ${BOLD}$BINDIR/voxtype-tui${NC}"

if ! command -v go >/dev/null 2>&1; then
  err "Go is required to build voxtype-tui"
  exit 1
fi

mkdir -p "$BINDIR"

info "Tidying Go module"
go -C "$SCRIPT_DIR" mod tidy

info "Running Go tests"
go -C "$SCRIPT_DIR" test ./...

info "Building voxtype-tui"
go -C "$SCRIPT_DIR" build -o "$BINDIR/voxtype-tui" ./cmd/voxtype-tui
chmod +x "$BINDIR/voxtype-tui"

ok "voxtype-tui installed"
