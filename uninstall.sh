#!/usr/bin/env bash

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== Uninstalling ocyd (Ocypus Display Daemon) ===${NC}\n"

SYSTEMD_SERVICE="$HOME/.config/systemd/user/ocyd.service"
if [ -f "$SYSTEMD_SERVICE" ]; then
  echo -e "${BLUE}[*] Stopping and disabling systemd user service...${NC}"
  systemctl --user stop ocyd.service 2>/dev/null || true
  systemctl --user disable ocyd.service 2>/dev/null || true
  rm -f "$SYSTEMD_SERVICE"
  systemctl --user daemon-reload
  systemctl --user reset-failed 2>/dev/null || true
  echo -e "${GREEN}[+] systemd service removed.${NC}"
fi

if [ -f "/usr/local/bin/ocyd" ] || [ -f "/usr/local/bin/ocydctl" ]; then
  echo -e "${BLUE}[*] Removing binaries from /usr/local/bin...${NC}"
  sudo rm -f /usr/local/bin/ocyd /usr/local/bin/ocydctl
  echo -e "${GREEN}[+] Binaries removed.${NC}"
fi

UDEV_RULE="/etc/udev/rules.d/99-ocypus.rules"
if [ -f "$UDEV_RULE" ]; then
  echo -e "${BLUE}[*] Removing udev rules...${NC}"
  sudo rm -f "$UDEV_RULE"
  echo -e "${BLUE}[*] Reloading udev rules...${NC}"
  sudo udevadm control --reload-rules && sudo udevadm trigger
  echo -e "${GREEN}[+] udev rules removed.${NC}"
fi

CONFIG_DIR="$HOME/.config/ocyd"
if [ -d "$CONFIG_DIR" ]; then
  read -p "Do you want to delete configuration files in $CONFIG_DIR? [y/N]: " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf "$CONFIG_DIR"
    echo -e "${GREEN}[+] Configuration directory removed.${NC}"
  else
    echo -e "${BLUE}[*] Configuration files kept in $CONFIG_DIR${NC}"
  fi
fi

echo -e "\n${GREEN}=== Uninstall Completed Successfully! ===${NC}"
