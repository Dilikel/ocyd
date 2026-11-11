#!/usr/bin/env bash
# Copyright (C) 2026 Vladimir Shiryaev
# SPDX-License-Identifier: GPL-3.0-or-later
#
# ocyd uninstaller: removes everything install.sh put on the system.
#
# Environment overrides (must match the ones used for install.sh):
#   OCYD_BIN_DIR   where binaries were installed   (default: /usr/bin)
#   OCYD_UDEV_DIR  where the udev rule was placed  (default: /etc/udev/rules.d)

set -euo pipefail

BIN_DIR="${OCYD_BIN_DIR:-/usr/bin}"
UDEV_DIR="${OCYD_UDEV_DIR:-/etc/udev/rules.d}"

UDEV_RULE="99-ocypus.rules"
UNIT_NAME="ocyd.service"
USER_UNIT_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
CONFIG_DIR="$HOME/.config/ocyd"

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

if [[ $EUID -eq 0 ]]; then
  die "do not run the uninstaller as root. Run it as your regular user:
      your user service and config are removed from your home directory,
      and sudo is requested only for the system files."
fi

command -v sudo >/dev/null 2>&1 || die "sudo is required to remove system files."

stop_service() {
  step "systemd user service"

  local unit_file="$USER_UNIT_DIR/$UNIT_NAME"
  local have_systemctl=0
  command -v systemctl >/dev/null 2>&1 && have_systemctl=1

  if [[ $have_systemctl -eq 1 ]]; then
    if systemctl --user cat "$UNIT_NAME" >/dev/null 2>&1; then
      systemctl --user disable --now "$UNIT_NAME" >/dev/null 2>&1 || true
      ok "Service stopped and disabled"
    else
      info "Service is not registered, nothing to stop."
    fi
  else
    info "systemctl not found, skipping service shutdown."
  fi

  if [[ -f "$unit_file" ]]; then
    rm -f "$unit_file"
    ok "Removed $unit_file"
    [[ $have_systemctl -eq 1 ]] && { systemctl --user daemon-reload 2>/dev/null || true; }
  fi

  rm -f "${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user/default.target.wants/$UNIT_NAME"

  if pgrep -u "$USER" -x ocyd >/dev/null 2>&1; then
    pkill -u "$USER" -x ocyd || true
    ok "Stopped a manually started ocyd process"
  fi
}

remove_binaries() {
  step "Binaries"

  local removed=0 dir name path
  for dir in "$BIN_DIR" /usr/local/bin; do
    for name in ocyd ocydctl; do
      path="$dir/$name"
      if [[ -e "$path" ]]; then
        sudo rm -f "$path"
        ok "Removed $path"
        removed=1
      fi
    done
  done
  [[ $removed -eq 1 ]] || info "No installed binaries found."
}

remove_udev() {
  step "udev rule"

  local rule="$UDEV_DIR/$UDEV_RULE"
  if [[ -e "$rule" ]]; then
    sudo rm -f "$rule"
    ok "Removed $rule"
    if command -v udevadm >/dev/null 2>&1; then
      sudo udevadm control --reload-rules || warn "failed to reload udev rules"
      ok "udev rules reloaded"
    fi
  else
    info "udev rule not found, nothing to remove."
  fi
}

remove_config() {
  step "User configuration"

  if [[ ! -d "$CONFIG_DIR" ]]; then
    info "No config directory found."
    return 0
  fi

  if ask_yn "Delete the config directory $CONFIG_DIR?" n; then
    rm -rf "$CONFIG_DIR"
    ok "Removed $CONFIG_DIR"
  else
    info "Config kept: $CONFIG_DIR"
  fi
  return 0
}

main() {
  printf '%socyd uninstaller%s\n' "$C_BOLD" "$C_RESET"
  echo
  info "This will remove the ocyd binaries, the systemd user service,"
  info "and the udev rule."
  echo
  ask_yn "Continue?" y || {
    info "Aborted."
    exit 0
  }

  stop_service
  remove_binaries
  remove_udev
  remove_config

  step "Done"
  info "ocyd has been removed. If the cooler was plugged in, reconnect it so udev"
  info "forgets the old permissions."
}

main
