# ocyd

Linux daemon for sending live CPU or GPU temperature to Ocypus cooler displays over USB HID.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Latest release](https://img.shields.io/github/v/release/Dilikel/ocyd?include_prereleases)](https://github.com/Dilikel/ocyd/releases)

[Read this in Russian](README.ru.md) · [Report a bug or request support](https://github.com/Dilikel/ocyd/issues/new/choose)

---

## Status

ocyd is early alpha software (`v0.1.0-alpha`). Hardware support is still being expanded.

| Model                 | USB IDs         | Status                 |
| --------------------- | --------------- | ---------------------- |
| Ocypus Delta A62EX    | `0x1a2c:0x434d` | Tested                 |
| Other Ocypus displays | Unknown         | Please test and report |

---

## Requirements

- Linux with `systemd` and user services enabled.
- `libudev` development files. On Arch Linux: `sudo pacman -S systemd-libs`; on Ubuntu/Debian: `sudo apt install libudev-dev`.
- Go `1.26+` and `make` when building from source.

## Install

```bash
git clone https://github.com/Dilikel/ocyd.git
cd ocyd
./install.sh
```

The installer builds the daemon, lets you choose a device profile, installs the binary and udev rules, creates `~/.config/ocyd/config.toml`, and enables the user service. It may ask for `sudo` while installing system files.

## Configuration

The default file is `~/.config/ocyd/config.toml`:

```toml
[display]
source = "cpu" # "cpu" or "gpu"
unit = "C"     # "C" or "F"

[device]
vendor_id = 0x1a2c
product_id = 0x434d
```

To use another display, update the installed configuration and udev rule as described below. The daemon reads the configuration when it starts.

## Service and logs

```bash
systemctl --user status ocyd.service
journalctl --user -u ocyd.service -f
```

Stop or restart it with `systemctl --user stop ocyd.service` or `systemctl --user restart ocyd.service`.

## Testing your own cooler

Connect the cooler and find its USB IDs:

```bash
lsusb
```

The output may identify the cooler as a keyboard. For example, the Delta A62EX appears as:

```text
Bus 001 Device 005: ID 1a2c:434d China Resource Semico Co., Ltd USB Gaming Keyboard
```

Other coolers may also be detected as a keyboard. In this line, `1a2c` is the **vendor ID** and `434d` is the **product ID**. Use the IDs shown for your own device.

After finding your IDs, update these two files:

1. Device configuration: `~/.config/ocyd/config.toml`

   ```toml
   [device]
   vendor_id = 0x1234
   product_id = 0x5678
   ```

2. udev permissions: `/etc/udev/rules.d/99-ocypus.rules`

   ```text
   SUBSYSTEM=="hidraw", ATTRS{idVendor}=="1234", ATTRS{idProduct}=="5678", MODE="0666"
   ```

Replace `1234` and `5678` with your own vendor and product IDs in **both files**. Keep the IDs in the same format as `lsusb`: four hexadecimal characters without the `0x` prefix in the udev rule. The TOML configuration uses the `0x` prefix.

Changing the udev rule requires `sudo` because the file belongs to root.

After changing the udev rule, reload it and reconnect the cooler:

```bash
sudo udevadm control --reload-rules
sudo udevadm trigger
```

Reconnect the cooler, restart the daemon, and check the service logs:

```bash
systemctl --user restart ocyd.service
journalctl --user -u ocyd.service -f
```

## Testing another model

Only the Delta A62EX has been verified so far. If you have another Ocypus display, please try it and open an [issue](https://github.com/Dilikel/ocyd/issues/new/choose), even if it does not work yet. Include:

- exact cooler model and Linux distribution;
- USB vendor/product IDs from `lsusb`;
- whether the display is detected and whether the temperature updates;
- relevant output from `systemctl --user status ocyd.service` and `journalctl --user -u ocyd.service`.

Hardware profiles and pull requests are welcome. Please do not include personal data or full system logs unrelated to the device.

## Build and development

```bash
make build
make test
make all
```

`make all` formats, vets, lints, tests, and builds the project.

## Roadmap

- Add more tested device profiles.
- Expand unit and hardware-mocking tests.
- Add `ocydctl` for convenient daemon management.
- Publish an AUR package.

## License

ocyd is available under the [MIT License](LICENSE).
