// Copyright (C) 2026 Vladimir Shiryaev
// SPDX-License-Identifier: GPL-3.0-or-later

// Command ocydctl updates the ocyd daemon configuration from the command line.
//
// It loads the configuration file from the default location, applies the values
// passed via flags, and saves the result. Flags that are not specified leave the
// corresponding settings unchanged. The daemon reads the configuration only on
// startup, so it must be restarted for the changes to take effect.
//
// Usage:
//
//	ocydctl [-unit C|F] [-source CPU|GPU] [-vendor ID] [-product ID]
//
// Flags:
//
//	-unit     temperature unit: C or F
//	-source   temperature source: CPU or GPU
//	-vendor   USB vendor ID of the display as a decimal number
//	-product  USB product ID of the display as a decimal number
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/Dilikel/ocyd/pkg/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("Configuration updated successfully!")
}

func run() error {
	var (
		unitFlag    string
		sourceFlag  string
		vendorFlag  string
		productFlag string
	)

	flag.StringVar(&unitFlag, "unit", "", "Temperature unit: C | F")
	flag.StringVar(&sourceFlag, "source", "", "Data source: CPU | GPU")
	flag.StringVar(&vendorFlag, "vendor", "", "USB Vendor ID as 16-bit integer: hex ('0x1a2c', '1a2c') or decimal ('6700')")
	flag.StringVar(&productFlag, "product", "", "USB Product ID as 16-bit integer: hex ('0x434d', '434d') or decimal ('17229')")

	flag.Parse()

	cfgPath, err := config.DefaultPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to get config value: %w", err)
	}

	if unitFlag != "" {
		switch unitFlag {
		case "C":
			cfg.Display.Unit = "C"
		case "F":
			cfg.Display.Unit = "F"
		default:
			return fmt.Errorf("invalid unit flag '%s'. Expected 'C' or 'F'", unitFlag)
		}
	}

	if sourceFlag != "" {
		switch sourceFlag {
		case "CPU":
			cfg.Display.Source = "CPU"
		case "GPU":
			cfg.Display.Source = "GPU"
		default:
			return fmt.Errorf("invalid source '%s'. Expected 'CPU' or 'GPU'", sourceFlag)
		}
	}

	if vendorFlag != "" {
		_, err := strconv.ParseUint(vendorFlag, 0, 16)
		if err != nil {
			return fmt.Errorf("invalid device vendor_id %q: must be a valid 16-bit integer (e.g. hex '0x1a2c'/'1a2c' or decimal '6700'): %w", vendorFlag, err)
		}
		cfg.Device.VendorID = vendorFlag
	}

	if productFlag != "" {
		_, err := strconv.ParseUint(productFlag, 0, 16)
		if err != nil {
			return fmt.Errorf("invalid device product_id %q: must be a valid 16-bit integer (e.g. hex '0x434d'/'434d' or decimal '17229'): %w", productFlag, err)
		}
		cfg.Device.ProductID = productFlag
	}

	err = config.Save(cfg, cfgPath)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}
