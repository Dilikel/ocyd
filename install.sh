#!/usr/bin/env bash

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== Installing ocyd (Ocypus Display Daemon) ===${NC}\n"

if [ ! -f "./bin/ocyd" ]; then
  echo -e "${BLUE}[*] Executable binary not found. Running make build...${NC}"
  make build
fi

DEVICES_DIR="./install/devices"
if [ ! -d "$DEVICES_DIR" ]; then
  echo -e "${RED}[!] Error: Profile directory $DEVICES_DIR not found!${NC}"
  exit 1
fi

devices=()
for dir in "$DEVICES_DIR"/*; do
  if [ -d "$dir" ]; then
    devices+=("$(basename "$dir")")
  fi
done

if [ ${#devices[@]} -eq 0 ]; then
  echo -e "${RED}[!] Error: No device profiles found in $DEVICES_DIR!${NC}"
  exit 1
fi

echo "Select your Ocypus cooler model:"
select selected_device in "${devices[@]}"; do
  if [ -n "$selected_device" ]; then
    echo -e "\nSelected model: ${GREEN}$selected_device${NC}"
    break
  else
    echo -e "${RED}Invalid selection. Please try again.${NC}"
  fi
done

DEV_PATH="$DEVICES_DIR/$selected_device"
BASE_PATH="./install/base"

echo -e "\n${BLUE}[*] Installing daemon binary to /usr/local/bin...${NC}"
sudo install -Dm755 ./bin/ocyd /usr/local/bin/ocyd

if [ -f "$DEV_PATH/99-ocypus.rules" ]; then
  echo -e "${BLUE}[*] Installing udev rules...${NC}"
  sudo install -Dm644 "$DEV_PATH/99-ocypus.rules" /etc/udev/rules.d/99-ocypus.rules
  echo -e "${BLUE}[*] Reloading udev rules...${NC}"
  sudo udevadm control --reload-rules && sudo udevadm trigger
fi

CONFIG_DIR="$HOME/.config/ocyd"
mkdir -p "$CONFIG_DIR"

if [ -f "$CONFIG_DIR/config.toml" ]; then
  echo -e "${BLUE}[!] Existing configuration found at $CONFIG_DIR/config.toml. Skipping overwriting.${NC}"
else
  echo -e "${BLUE}[*] Creating default config in $CONFIG_DIR/config.toml...${NC}"
  cp "$DEV_PATH/config.toml" "$CONFIG_DIR/config.toml"
fi

SYSTEMD_USER_DIR="$HOME/.config/systemd/user"
mkdir -p "$SYSTEMD_USER_DIR"

if [ -f "$BASE_PATH/ocyd.service" ]; then
  echo -e "${BLUE}[*] Installing systemd user service...${NC}"
  cp "$BASE_PATH/ocyd.service" "$SYSTEMD_USER_DIR/ocyd.service"
  systemctl --user daemon-reload
  systemctl --user enable --now ocyd.service
fi

echo -e "\n${GREEN}=== Installation Completed Successfully! ===${NC}"
echo -e "The ocyd daemon has been started and enabled on user login."
echo -e "Check daemon status: ${BLUE}systemctl --user status ocyd.service${NC}"
