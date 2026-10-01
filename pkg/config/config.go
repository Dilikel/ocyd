// Copyright (C) 2026 Vladimir Shiryaev
// SPDX-License-Identifier: GPL-3.0-or-later

// Package config provides primitives for resolving paths and loading TOML configuration
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the application's root configuration structure.
type Config struct {
	Display DisplayConfig `toml:"display"`
	Device  DeviceConfig  `toml:"device"`
}

// DisplayConfig defines output settings such as temperature units and data sources.
type DisplayConfig struct {
	Unit   string `toml:"unit"`
	Source string `toml:"source"`
}

// DeviceConfig holds USB identification parameters for the target HID display.
type DeviceConfig struct {
	VendorID  uint16 `toml:"vendor_id"`
	ProductID uint16 `toml:"product_id"`
}

// DefaultPath returns the standard user-level configuration path (~/.config/ocyd/config.toml).
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}
	return filepath.Join(home, ".config", "ocyd", "config.toml"), nil
}

// Load reads and parses a TOML configuration file from the given file path.
func Load(path string) (Config, error) {
	//nolint:gosec // Config path is explicitly determined by application logic
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal toml config: %w", err)
	}
	return cfg, nil
}

// Save marshals the provided configuration into TOML format and saves it to a file.
//
// If the file at the specified path does not exist, it will be created with 0644
// permissions (read/write for owner, read-only for others). If the file already
// exists, it will be overwritten and truncated.
//
// It returns an error if the configuration fails to marshal or if the file write operation fails.
func Save(cfg Config, path string) error {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal toml config: %w", err)
	}

	err = os.WriteFile(path, data, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
