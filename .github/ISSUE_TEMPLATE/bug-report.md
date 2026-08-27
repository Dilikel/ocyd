---
name: Bug report
about: Report a cooler, installation, or runtime problem
title: '[Bug]: '
labels: ['bug']
assignees: []
---

## Problem

Describe what happened and what you expected to happen.

## Device and environment

- Cooler model:
- Vendor ID:
- Product ID:
- `lsusb` output:
- Linux distribution and version:
- Kernel version:
- ocyd version or commit:
- Temperature source (`cpu` or `gpu`):
- Temperature unit (`C` or `F`):

## Reproduction steps

1.
2.
3.

## Checks

- [ ] The cooler is connected
- [ ] `~/.config/ocyd/config.toml` contains the correct vendor and product IDs
- [ ] `/etc/udev/rules.d/99-ocypus.rules` contains the same IDs
- [ ] udev rules were reloaded and the cooler was reconnected
- [ ] `systemctl --user restart ocyd.service` was tried

## Logs

Paste relevant output, preferably with `--no-pager`:

```text
systemctl --user status ocyd.service
journalctl --user -u ocyd.service --no-pager
```

<details>
<summary>Additional logs</summary>

```text
Paste additional relevant output here.
```

</details>

## Evidence

Attach screenshots, photos, or video when they help demonstrate the problem. Drag files into this field or paste links. Remove personal data and unrelated logs before posting.
