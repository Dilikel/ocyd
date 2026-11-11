<div align="center">

# ocyd

**Show your CPU or GPU temperature on an Ocypus cooler display — on Linux.**

A tiny daemon that reads thermal sensors from sysfs / NVML and pushes the value to the cooler's screen over USB HID.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/Dilikel/ocyd?include_prereleases)](https://github.com/Dilikel/ocyd/releases)
![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)
![Platform](https://img.shields.io/badge/platform-Linux-lightgrey?logo=linux&logoColor=white)

[Русская версия](README.ru.md) · [Report a bug / request support](https://github.com/Dilikel/ocyd/issues/new/choose) · [Discussions](https://github.com/Dilikel/ocyd/discussions)

</div>

---

## About

Ocypus coolers have a small screen that can show the CPU or GPU temperature. The official app exists only for Windows, so on Linux the screen just stays dark. **ocyd** fills that gap: the protocol was reverse-engineered from the official app's USB traffic, and the daemon speaks it natively.

- **CPU or GPU** temperature — your choice
- **°C or °F**
- **Lightweight** — the daemon uses only about 10 MB of RAM
- **Wide sensor support** — AMD, Intel and NVIDIA (including the proprietary driver via NVML)
- **Configurable USB IDs** — nothing is hard-coded to a single model
- **systemd user service** — starts with your session, no root daemon
- **`ocydctl`** — change settings from the command line
- **Install / uninstall scripts** that do the boring parts for you

> [!NOTE]
> ocyd is **early alpha**. It works well on the author's hardware, but it has no unit tests yet and the code is not polished. Both are planned — see the [Roadmap](#roadmap).

## Supported devices

| Model                       | USB ID      | Status                    |
| --------------------------- | ----------- | ------------------------- |
| Ocypus Delta A62 EX Digital | `1a2c:434d` | ✅ Tested                 |
| Other Ocypus displays       | Unknown     | ❓ Please test and report |

The cooler does not have its own USB IDs: it identifies itself as a generic **"USB Gaming Keyboard"** (`China Resource Semico Co., Ltd`). Other models or hardware revisions **may use different IDs** — see [Using a different cooler](#using-a-different-cooler).

## Requirements

- Linux (`x86_64`)
- `sudo` — for installing the binaries and the udev rule
- `systemd` — optional; needed only for the bundled user service
- To build from source: Go `1.26+`, `gcc`, `make` (optional) and libudev headers
  - Arch Linux: `sudo pacman -S systemd-libs`
  - Debian / Ubuntu: `sudo apt install libudev-dev`
- For NVIDIA GPU temperature with the proprietary driver: the driver itself (NVML ships with it)

## Installation

### Option 1 — release archive (no build needed)

Download `ocyd_<version>_linux_amd64.tar.gz` from the [Releases](https://github.com/Dilikel/ocyd/releases) page, then:

```bash
tar -xzf ocyd_*_linux_amd64.tar.gz
cd ocyd_*_linux_amd64
./install.sh
```

The archive already contains `ocyd` and `ocydctl`; the installer copies them from the project directory.

### Option 2 — git checkout

```bash
git clone https://github.com/Dilikel/ocyd.git
cd ocyd
./install.sh
```

If there are no binaries in `./bin`, the installer **builds them for you**. To force a fresh build, run `./install.sh --rebuild`.

### What the installer does

Run it as your **regular user** (not root). It asks for `sudo` only for system files.

| Step                    | What happens                                                                                                                                           |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1. Binaries             | Uses binaries from the project directory (release archive) or from `./bin`; builds them if missing. Installs `ocyd` and `ocydctl` to `/usr/bin/`.      |
| 2. USB IDs              | Asks whether to use the default `1a2c:434d` (Ocypus Delta A62 EX Digital) or to enter your own.                                                        |
| 3. udev rule            | Installs `/etc/udev/rules.d/99-ocypus.rules` so your user can talk to the cooler's HID device, then reloads udev.                                      |
| 4. Config               | Creates `~/.config/ocyd/config.toml` if it does not exist (an existing config is never overwritten).                                                   |
| 5. systemd _(optional)_ | Asks whether you use systemd. If yes — installs `ocyd.service` to `~/.config/systemd/user/` and asks whether to **enable autostart and start it now**. |

If you answer "no" to systemd, just run `ocyd` yourself or wire it into your init system (runit, OpenRC, s6, ...) as a **user-level** service.

### The udev rule

```text
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="1a2c", ATTRS{idProduct}=="434d", MODE="0666"
```

`1a2c:434d` is the VID:PID of the **Ocypus Delta A62 EX Digital**. Your model may differ — the installer lets you enter your own IDs and writes them to the rule and to the config.

> [!WARNING]
> `MODE="0666"` makes the cooler's `hidraw` node readable and writable by every local user. For a device that only draws a number on a screen this is a reasonable trade-off, but you should know about it.

## Uninstallation

```bash
./uninstall.sh
```

Removes everything the installer put on the system: stops and disables the user service, deletes the unit, the binaries from `/usr/bin`, and the udev rule (and reloads udev). It asks separately whether to delete `~/.config/ocyd`, so your settings survive by default.

## Configuration

Config file: `~/.config/ocyd/config.toml`

```toml
[display]
source = 'CPU'   # 'CPU' or 'GPU'
unit   = 'C'     # 'C' or 'F'

[device]
vendor_id  = '0x1a2c'
product_id = '0x434d'
```

- Values are **case-sensitive**: `'CPU'`, `'GPU'`, `'C'`, `'F'`.
- `vendor_id` / `product_id` are strings and accept hex (`'0x1a2c'`, `'1a2c'`) or decimal (`'6700'`).
- The daemon reads the config **only on start** — restart it after any change.

### `ocydctl`

Edits the config for you. Flags you do not pass leave the corresponding values unchanged:

```bash
ocydctl -source GPU            # show GPU temperature
ocydctl -unit F                # Fahrenheit
ocydctl -vendor 0x1a2c -product 0x434d

systemctl --user restart ocyd.service
```

| Flag       | Values                | Description        |
| ---------- | --------------------- | ------------------ |
| `-source`  | `CPU`, `GPU`          | Temperature source |
| `-unit`    | `C`, `F`              | Temperature unit   |
| `-vendor`  | hex or decimal 16-bit | USB vendor ID      |
| `-product` | hex or decimal 16-bit | USB product ID     |

## Service and logs

```bash
systemctl --user status  ocyd.service     # status
systemctl --user restart ocyd.service     # restart (after config changes)
systemctl --user stop    ocyd.service     # stop
systemctl --user start   ocyd.service     # start
systemctl --user disable ocyd.service     # turn off autostart
systemctl --user enable --now ocyd.service  # turn on autostart and start

journalctl --user -u ocyd.service -f      # follow logs
```

## Using a different cooler

1. Plug the cooler in and find its IDs:

   ```bash
   lsusb
   ```

   The cooler usually looks like a keyboard:

   ```text
   Bus 001 Device 005: ID 1a2c:434d China Resource Semico Co., Ltd USB Gaming Keyboard
   ```

   Here `1a2c` is the **vendor ID** and `434d` is the **product ID**.

2. Easiest way: run `./install.sh` again and enter your IDs when asked — it updates both the udev rule and the config.

3. Or do it by hand:

   ```bash
   # config
   ocydctl -vendor 0xVVVV -product 0xPPPP

   # udev rule (VVVV and PPPP are 4 hex digits, no 0x here)
   sudo nano /etc/udev/rules.d/99-ocypus.rules
   #   SUBSYSTEM=="hidraw", ATTRS{idVendor}=="VVVV", ATTRS{idProduct}=="PPPP", MODE="0666"

   sudo udevadm control --reload-rules
   sudo udevadm trigger --subsystem-match=hidraw --action=add
   ```

4. Reconnect the cooler and restart the daemon:

   ```bash
   systemctl --user restart ocyd.service
   journalctl --user -u ocyd.service -f
   ```

If it works, please [tell us](https://github.com/Dilikel/ocyd/issues/new/choose) — we will add your model to the table above.

## How it works

**Sensors.** CPU temperature is read from `/sys/class/hwmon/hwmon*/temp*_input`. ocyd finds the right `hwmon` node by driver name and sensor label, falling back to `temp1_input`:

| Source | Drivers                                         | Labels                                         |
| ------ | ----------------------------------------------- | ---------------------------------------------- |
| CPU    | `k10temp`, `coretemp`, `zenpower`, `amd_energy` | `Tctl`, `Tdie`, `Package id 0`, `Package id 1` |
| GPU    | `amdgpu`, `nouveau`, `nvidia`, `i915`, `xe`     | `edge`, `junction`, `hotspot`, `GPU`           |

The proprietary NVIDIA driver does not expose a `hwmon` sensor, so for it ocyd falls back to the NVML library.

**Protocol.** Once per second the daemon sends one 64-byte HID report:

| Byte (0-based) | Value                                            |
| -------------- | ------------------------------------------------ |
| `0`            | Report ID `0x07`                                 |
| `1`, `2`       | `0xFF`, `0xFF`                                   |
| `3`, `4`, `5`  | Hundreds, tens and ones digit of the temperature |
| `6`–`63`       | `0x00`                                           |

## Troubleshooting

| Symptom                           | What to check                                                                                                                                  |
| --------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `failed to load config`           | `~/.config/ocyd/config.toml` is missing. Re-run `./install.sh` or copy `packaging/config/config.toml` there.                                   |
| `open device failed`              | The IDs in the config don't match your cooler, or the udev rule is missing / not applied. Check `lsusb`, reload udev and reconnect the cooler. |
| `invalid display source`          | Use upper case: `'CPU'` / `'GPU'`.                                                                                                             |
| `thermal sensor file not found`   | Your CPU/GPU driver isn't in the supported list. Open an issue with the output of `grep . /sys/class/hwmon/hwmon*/name`.                       |
| Service doesn't start after login | `systemctl --user status ocyd.service` and `journalctl --user -u ocyd.service`.                                                                |

## Testing another model

Only the Delta A62 EX Digital has been verified. If you own a different Ocypus display, please try it and open an [issue](https://github.com/Dilikel/ocyd/issues/new/choose) **even if it doesn't work**. Include:

- exact cooler model and Linux distribution;
- USB vendor/product IDs from `lsusb`;
- whether the display is detected and whether the temperature updates;
- relevant output of `systemctl --user status ocyd.service` and `journalctl --user -u ocyd.service`.

Please don't include personal data or unrelated system logs.

## Development

```bash
make build   # build ocyd and ocydctl into ./bin
make test    # go test ./...
make all     # fmt + vet + lint + test + build
```

Project layout:

```text
cmd/ocyd       daemon entry point
cmd/ocydctl    CLI for editing the config
pkg/config     TOML config loading and saving
pkg/hwmon      sysfs and NVML temperature readers
pkg/ocypus     HID driver for the cooler display
packaging/     udev rule, systemd unit, default config
```

## Roadmap

ocyd is young, and honest about it. Planned:

- [ ] **Unit tests** — config, sensor discovery (against a fake sysfs tree) and the HID driver (against a mock device).
- [ ] **Cleaner, better code** — refactoring, better error handling and structure as the project matures.
- [ ] More tested devices and device profiles.
- [ ] AUR packages: `ocyd` (build from source) and `ocyd-bin` (prebuilt release).

Contributions, hardware reports and code review are very welcome.

## License

ocyd is free software, released under the [GNU General Public License v3.0 or later](LICENSE).
