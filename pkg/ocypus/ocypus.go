// Package ocypus provides a low-level driver for interacting with the Ocypus
// CPU cooler display via USB HID.package ocypus
package ocypus

import (
	"fmt"
	"log/slog"

	"github.com/sstallion/go-hid"
)

// PacketSize defines the required fixed size of the USB HID report for Ocypus devices in bytes.
const PacketSize = 64

// Display manages an active USB HID connection to the Ocypus display.
// It maintains an internal frame buffer to avoid memory allocations during frequent updates.
type Display struct {
	dev *hid.Device
	buf [PacketSize]byte
}

// NewDisplay initializes the HID subsystem and opens the first matching USB device
// identified by the provided vendorID and productID. It sets up the static report header.
//
// The caller is responsible for calling Close on the returned Display when it is no longer needed.
func NewDisplay(vendorID, productID uint16) (Display, error) {
	if err := hid.Init(); err != nil {
		return Display{}, fmt.Errorf("ocypus: hid init failed: %w", err)
	}

	dev, err := hid.OpenFirst(vendorID, productID)
	if err != nil {
		return Display{}, fmt.Errorf("ocypus: open device failed (VID: 0x%04x, PID: 0x%04x): %w", vendorID, productID, err)
	}

	d := Display{
		dev: dev,
	}
	d.buf[0] = 0x07
	d.buf[1] = 0xff
	d.buf[2] = 0xff
	return d, nil
}

// SendTemp formats the given temperature value in degrees into 3 decimal digits
// and writes the complete 64-byte payload to the USB display.
//
// If temp is out of the supported range (0-999), it is automatically clamped.
func (d *Display) SendTemp(temp int) error {
	digits := decompose(temp)
	copy(d.buf[3:6], digits[:])

	n, err := d.dev.Write(d.buf[:])
	if err != nil {
		return fmt.Errorf("ocypus: write temp failed: %w", err)
	}
	if n != PacketSize {
		slog.Warn("partially written bytes to ocypus display",
			slog.Int("written", n),
			slog.Int("expected", PacketSize),
		)
	}
	return nil
}

// Close releases the underlying USB HID device handle and associated system resources.
// It is safe to call Close on an uninitialized or already closed Display.
func (d *Display) Close() error {
	if d.dev != nil {
		return d.dev.Close()
	}
	return nil
}

func decompose(temp int) [3]byte {
	if temp < 0 {
		temp = 0
	}
	if temp > 999 {
		temp = 999
	}
	return [3]byte{
		byte(temp / 100),
		byte((temp % 100) / 10),
		byte(temp % 10),
	}
}
