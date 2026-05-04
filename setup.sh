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

go_version="$(go env GOVERSION)"
if [[ "$go_version" =~ ^go([0-9]+)\.([0-9]+) ]] && \
  ((BASH_REMATCH[1] < 1 || (BASH_REMATCH[1] == 1 && BASH_REMATCH[2] < 23))); then
  err "Go 1.23 or newer is required (found $go_version)"
  exit 1
fi

mkdir -p "$BINDIR"

info "Running Go tests"
go -C "$SCRIPT_DIR" test ./...

info "Building voxtype-tui"
version="$(git -C "$SCRIPT_DIR" describe --tags --always --dirty 2>/dev/null || printf 'dev')"
go -C "$SCRIPT_DIR" build \
  -ldflags "-X main.version=$version" \
  -o "$BINDIR/voxtype-tui" \
  ./cmd/voxtype-tui
chmod +x "$BINDIR/voxtype-tui"

ok "voxtype-tui installed"
