---
name: Verified device
about: Add a cooler model that you tested successfully
title: '[Device]: '
labels: ['device-support']
assignees: []
---

## Device

- Cooler model:
- Vendor ID:
- Product ID:
- `lsusb` output:
- Linux distribution and version:
- Kernel version:
- ocyd version or commit:

## Test result

- [ ] The display is detected
- [ ] The temperature is displayed
- [ ] The temperature updates correctly over time
- [ ] CPU source tested
- [ ] GPU source tested

Describe what you tested, including the configured source and unit:

## Evidence

Please attach evidence that the device works. Drag files into this field or paste links:

- Video or photo of the display working:
- Additional logs, screenshots, or test files:

Do not include personal data, private paths, or unrelated full system logs.

## Logs

<details>
<summary>Service status and logs</summary>

Paste the relevant output here:

```text
systemctl --user status ocyd.service
journalctl --user -u ocyd.service --no-pager
```

</details>

## Suggested profile

If possible, include a proposed profile or udev rule in a pull request. Do not attach files containing secrets or personal data.
