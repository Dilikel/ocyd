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
