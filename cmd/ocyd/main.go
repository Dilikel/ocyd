// Package main provides the executable entry point for the ocyd daemon.
//
// ocyd is a lightweight system daemon designed to fetch real-time CPU or GPU
// thermal metrics from Linux sysfs (/sys/class/hwmon) or NVML lib and push them over USB HID
// to Ocypus CPU cooler displays (such as the Ocypus Delta A62EX).
//
// # Configuration
//
// The daemon automatically resolves and loads its configuration from ~/.config/ocyd/config.toml.
//
// # Signal Handling
//
// ocyd handles SIGINT (Ctrl+C) and SIGTERM (systemd stop) signals gracefully.
// Upon receiving a shutdown signal, it safely closes hardware handles and releases
// the USB HID interface before exiting.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dilikel/ocyd/pkg/config"
	"github.com/Dilikel/ocyd/pkg/hwmon"
	"github.com/Dilikel/ocyd/pkg/ocypus"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("daemon terminated with error", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("daemon stopped gracefully")
}

func run(ctx context.Context) error {
	cfgPath, err := config.DefaultPath()
	if err != nil {
		return fmt.Errorf("failed to resolve default config path: %w", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	slog.Info("configuration loaded",
		slog.String("path", cfgPath),
		slog.String("source", cfg.Display.Source),
		slog.String("unit", cfg.Display.Unit),
		slog.String("vendor_id", fmt.Sprintf("0x%04x", cfg.Device.VendorID)),
		slog.String("product_id", fmt.Sprintf("0x%04x", cfg.Device.ProductID)),
	)

	var reader hwmon.ThermalReader
	switch cfg.Display.Source {
	case "cpu":
		r, err := hwmon.NewCPUReader()
		if err != nil {
			return fmt.Errorf("failed to initialize CPU thermal reader: %w", err)
		}
		reader = r
	case "gpu":
		r, err := hwmon.NewGPUReader()
		if err != nil {
			return fmt.Errorf("failed to initialize GPU thermal reader: %w", err)
		}
		reader = r
	default:
		return fmt.Errorf("invalid display source %q: must be 'cpu' or 'gpu'", cfg.Display.Source)
	}

	defer func() {
		if err := reader.Close(); err != nil {
			slog.Error("failed to close thermal reader", slog.Any("error", err))
		}
	}()

	var convert func(int) int
	switch cfg.Display.Unit {
	case "C":
		convert = func(c int) int { return c }
	case "F":
		convert = hwmon.CToF
	default:
		return fmt.Errorf("invalid display unit %q: must be 'C' or 'F'", cfg.Display.Unit)
	}

	display, err := ocypus.NewDisplay(cfg.Device.VendorID, cfg.Device.ProductID)
	if err != nil {
		return fmt.Errorf("failed to initialize display: %w", err)
	}
	defer func() {
		if err := display.Close(); err != nil {
			slog.Error("failed to close display device", slog.Any("error", err))
		}
	}()

	slog.Info("ocyd daemon started successfully")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down ocyd daemon...")
			return ctx.Err()
		case <-ticker.C:
			temp, err := reader.Read()
			if err != nil {
				slog.Error("failed to read temperature", slog.Any("error", err))
				continue
			}

			val := convert(temp)
			if err := display.SendTemp(val); err != nil {
				slog.Error("failed to send temperature to display",
					slog.Int("raw_temp", temp),
					slog.Int("converted_val", val),
					slog.Any("error", err),
				)
				continue
			}

			slog.Debug("temperature updated",
				slog.Int("temp", val),
				slog.String("unit", cfg.Display.Unit),
			)
		}
	}
}
