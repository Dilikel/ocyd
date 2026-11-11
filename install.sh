#!/usr/bin/env bash
# Copyright (C) 2026 Vladimir Shiryaev
# SPDX-License-Identifier: GPL-3.0-or-later
#
# ocyd installer.
#
# Works from both a git checkout (binaries in ./bin, built on demand) and an
# unpacked release tar.gz (binaries lie next to this script).
#
# Environment overrides (mostly for packaging and testing):
#   OCYD_BIN_DIR   where to install binaries      (default: /usr/bin)
#   OCYD_UDEV_DIR  where to install the udev rule (default: /etc/udev/rules.d)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

BIN_DIR="${OCYD_BIN_DIR:-/usr/bin}"
UDEV_DIR="${OCYD_UDEV_DIR:-/etc/udev/rules.d}"

UDEV_RULE="99-ocypus.rules"
UNIT_NAME="ocyd.service"
USER_UNIT_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
CONFIG_DIR="$HOME/.config/ocyd"
CONFIG_FILE="$CONFIG_DIR/config.toml"

DEFAULT_VID="1a2c"
DEFAULT_PID="434d"

REBUILD=0

if [[ -t 1 ]]; then
  C_RESET=$'\033[0m'
  C_BOLD=$'\033[1m'
  C_RED=$'\033[31m'
  C_GREEN=$'\033[32m'
  C_YELLOW=$'\033[33m'
  C_BLUE=$'\033[34m'
else
  C_RESET="" C_BOLD="" C_RED="" C_GREEN="" C_YELLOW="" C_BLUE=""
fi

step() { printf '\n%s==>%s %s%s%s\n' "$C_BLUE" "$C_RESET" "$C_BOLD" "$*" "$C_RESET"; }
info() { printf '    %s\n' "$*"; }
ok() { printf '    %s✓%s %s\n' "$C_GREEN" "$C_RESET" "$*"; }
warn() { printf '    %s!%s %s\n' "$C_YELLOW" "$C_RESET" "$*" >&2; }
die() {
  printf '%s✗ error:%s %s\n' "$C_RED" "$C_RESET" "$*" >&2
  exit 1
}

ask_yn() {
  local question="$1" default="${2:-y}" hint answer
  if [[ "$default" == "y" ]]; then hint="[Y/n]"; else hint="[y/N]"; fi
  while true; do
    read -r -p "    ${question} ${hint} " answer || answer=""
    answer="${answer:-$default}"
    case "${answer,,}" in
    y | yes | д | да) return 0 ;;
    n | no | н | нет) return 1 ;;
    *) echo "    Please answer y or n." ;;
    esac
  done
}

usage() {
  cat <<EOF
Usage: ./install.sh [--rebuild] [-h|--help]

  --rebuild   always rebuild binaries from source (needs Go, gcc, libudev)
  -h, --help  show this help
EOF
}

for arg in "$@"; do
  case "$arg" in
  --rebuild) REBUILD=1 ;;
  -h | --help)
    usage
    exit 0
    ;;
  *)
    usage >&2
    die "unknown argument: $arg"
    ;;
  esac
done

preflight() {
  if [[ "$(uname -s)" != "Linux" ]]; then
    die "ocyd only supports Linux."
  fi

  if [[ $EUID -eq 0 ]]; then
    die "do not run the installer as root. Run it as your regular user:
      the config and the systemd user unit belong to you, and sudo is
      requested only for the system files (binaries and the udev rule)."
  fi

  if ! command -v sudo >/dev/null 2>&1; then
    die "sudo is required to install system files (${BIN_DIR}, ${UDEV_DIR})."
  fi
}

BIN_SRC=""

have_binaries() {
  [[ -f "$1/ocyd" && -x "$1/ocyd" && -f "$1/ocydctl" && -x "$1/ocydctl" ]]
}

build_binaries() {
  command -v go >/dev/null 2>&1 ||
    die "Go is not installed and no prebuilt binaries were found.
      Install Go (1.26+) or download a release archive with ready binaries."
  command -v gcc >/dev/null 2>&1 ||
    die "gcc is required to build ocyd (CGO, HID). Install a C compiler."

  [[ -f "$SCRIPT_DIR/go.mod" ]] ||
    die "go.mod not found next to install.sh; cannot build from source."

  mkdir -p "$SCRIPT_DIR/bin"
  info "Building ocyd..."
  (cd "$SCRIPT_DIR" && go build -trimpath -ldflags "-s -w" -o bin/ocyd ./cmd/ocyd) ||
    die "failed to build ocyd. On Debian/Ubuntu install libudev-dev, on Arch systemd-libs."
  info "Building ocydctl..."
  (cd "$SCRIPT_DIR" && go build -trimpath -ldflags "-s -w" -o bin/ocydctl ./cmd/ocydctl) ||
    die "failed to build ocydctl."
  ok "Build finished: $SCRIPT_DIR/bin"
}

locate_binaries() {
  step "Looking for binaries"

  if [[ $REBUILD -eq 0 ]] && have_binaries "$SCRIPT_DIR"; then
    BIN_SRC="$SCRIPT_DIR"
    ok "Found prebuilt binaries in the project directory (release archive)"
    return
  fi

  if [[ $REBUILD -eq 0 ]] && have_binaries "$SCRIPT_DIR/bin"; then
    BIN_SRC="$SCRIPT_DIR/bin"
    ok "Found binaries in ./bin"
    return
  fi

  if [[ $REBUILD -eq 1 ]]; then
    info "Rebuild requested."
  else
    info "No binaries found, building from source."
  fi
  build_binaries
  BIN_SRC="$SCRIPT_DIR/bin"
}

install_binaries() {
  step "Installing binaries to ${BIN_DIR}"
  sudo install -Dm755 "$BIN_SRC/ocyd" "$BIN_DIR/ocyd"
  sudo install -Dm755 "$BIN_SRC/ocydctl" "$BIN_DIR/ocydctl"
  ok "$BIN_DIR/ocyd"
  ok "$BIN_DIR/ocydctl"
}

VID="$DEFAULT_VID"
PID="$DEFAULT_PID"
CUSTOM_IDS=0

normalize_id() {
  local id="${1,,}"
  id="${id#0x}"
  [[ "$id" =~ ^[0-9a-f]{4}$ ]] || return 1
  printf '%s' "$id"
}

choose_device() {
  step "Cooler USB IDs"
  info "The udev rule gives your user access to the cooler's HID device."
  info "Default: ${C_BOLD}${DEFAULT_VID}:${DEFAULT_PID}${C_RESET} (Ocypus Delta A62 EX Digital)."
  info "Other models may use different IDs. Find yours with: lsusb"
  echo

  if ask_yn "Use the default IDs ${DEFAULT_VID}:${DEFAULT_PID}?" y; then
    return
  fi

  local v p
  while true; do
    read -r -p "    Vendor ID  (4 hex digits, e.g. 1a2c): " v || die "no input"
    if v="$(normalize_id "$v")"; then break; fi
    warn "Invalid vendor ID, expected 4 hex digits."
  done
  while true; do
    read -r -p "    Product ID (4 hex digits, e.g. 434d): " p || die "no input"
    if p="$(normalize_id "$p")"; then break; fi
    warn "Invalid product ID, expected 4 hex digits."
  done
  VID="$v"
  PID="$p"
  CUSTOM_IDS=1
  ok "Using ${VID}:${PID}"
}

install_udev() {
  step "Installing udev rule"

  local template="$SCRIPT_DIR/packaging/udev/$UDEV_RULE"
  [[ -f "$template" ]] || die "udev rule template not found: $template"

  local tmp
  tmp="$(mktemp)"
  sed -e "s/${DEFAULT_VID}/${VID}/g" -e "s/${DEFAULT_PID}/${PID}/g" "$template" >"$tmp"

  sudo install -Dm644 "$tmp" "$UDEV_DIR/$UDEV_RULE"
  rm -f "$tmp"
  ok "$UDEV_DIR/$UDEV_RULE  (${VID}:${PID})"

  if command -v udevadm >/dev/null 2>&1; then
    sudo udevadm control --reload-rules || warn "failed to reload udev rules"
    sudo udevadm trigger --subsystem-match=hidraw --action=add || true
    ok "udev rules reloaded"
    info "If the cooler was already plugged in, reconnect it once."
  else
    warn "udevadm not found: reload udev rules or reboot manually."
  fi
}

apply_custom_ids() {
  if "$BIN_DIR/ocydctl" -vendor "0x${VID}" -product "0x${PID}" >/dev/null; then
    ok "Device IDs in config set to 0x${VID}:0x${PID}"
  else
    warn "could not update the config, edit $CONFIG_FILE manually."
  fi
}

install_config() {
  step "User configuration"

  local template="$SCRIPT_DIR/packaging/config/config.toml"

  if [[ -f "$CONFIG_FILE" ]]; then
    ok "Config already exists, keeping it: $CONFIG_FILE"
    [[ $CUSTOM_IDS -eq 1 ]] && apply_custom_ids
    return 0
  fi

  [[ -f "$template" ]] || die "config template not found: $template"
  mkdir -p "$CONFIG_DIR"
  install -m644 "$template" "$CONFIG_FILE"
  ok "Created $CONFIG_FILE"

  [[ $CUSTOM_IDS -eq 1 ]] && apply_custom_ids
  info "Change source/unit any time: ocydctl -source CPU|GPU -unit C|F"
  return 0
}

install_systemd() {
  step "systemd"

  local have_systemd=0 default=n
  if [[ -d /run/systemd/system ]] && command -v systemctl >/dev/null 2>&1; then
    have_systemd=1
    default=y
  else
    warn "systemd does not seem to be running on this system."
  fi

  if ! ask_yn "Install the systemd user service?" "$default"; then
    info "Skipping systemd. Start the daemon yourself with: ocyd"
    info "For other init systems (runit, OpenRC, ...) create a user-level service by hand."
    return 0
  fi

  [[ $have_systemd -eq 1 ]] || die "systemctl is not available, cannot install a systemd service."

  local unit="$SCRIPT_DIR/packaging/systemd/$UNIT_NAME"
  [[ -f "$unit" ]] || die "service file not found: $unit"

  mkdir -p "$USER_UNIT_DIR"
  install -m644 "$unit" "$USER_UNIT_DIR/$UNIT_NAME"
  ok "$USER_UNIT_DIR/$UNIT_NAME"

  if ! systemctl --user daemon-reload 2>/dev/null; then
    warn "could not talk to the systemd user instance (no login session?)."
    info "Log in normally and run:  systemctl --user enable --now $UNIT_NAME"
    return 0
  fi
  ok "systemd user units reloaded"

  if ask_yn "Enable autostart and start ocyd now?" y; then
    if systemctl --user enable --now "$UNIT_NAME"; then
      ok "ocyd is enabled and running"
      info "Status: systemctl --user status $UNIT_NAME"
      info "Logs:   journalctl --user -u $UNIT_NAME -f"
    else
      warn "failed to enable the service, check: systemctl --user status $UNIT_NAME"
    fi
  else
    info "Autostart not enabled. Start it manually when you need it:"
    info "  systemctl --user start $UNIT_NAME"
    info "Enable autostart later with:"
    info "  systemctl --user enable --now $UNIT_NAME"
  fi
  return 0
}

main() {
  printf '%socyd installer%s\n' "$C_BOLD" "$C_RESET"
  preflight
  locate_binaries
  choose_device
  install_binaries
  install_udev
  install_config
  install_systemd

  step "Done"
  info "Uninstall any time with ./uninstall.sh"
}

main
