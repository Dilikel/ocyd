// Package hwmon provides discovery and reading capabilities for Linux sysfs thermal metrics.
package hwmon

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type SencorTarget struct {
	Drivers []string
	Labels  []string
}

var (
	CPUTatger = SencorTarget{
		Drivers: []string{
			"k10temp", "coretemp", "zenpower", "amd_energy",
		},
		Labels: []string{
			"Tctl", "Tdie", "Package id 0", "Package id 1",
		},
	}
	GPUTarget = SencorTarget{
		Drivers: []string{
			"amdgpu", "nouveau", "nvidia", "i915", "xe",
		},
		Labels: []string{
			"edge", "junction", "hotspot", "GPU",
		},
	}
)

// FindSensorPath scans /sys/class/hwmon and returns the absolute sysfs path (string)
// to a matching thermal sensor input file based on the provided SensorTarget rules.
// It returns a non-nil error if no matching driver or sensor file is found.
func FindSensorPath(target SencorTarget) (string, error) {
	paths, err := filepath.Glob("/sys/class/hwmon/hwmon*/name")
	if err != nil {
		return "", fmt.Errorf("failed to glob hwmon: %w", err)
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		driverName := string(bytes.TrimSpace(content))
		if slices.Contains(target.Drivers, driverName) {
			dir := filepath.Dir(path)
			if targetPath, ok := findPathByLabels(dir, target.Labels); ok {
				return targetPath, nil
			}
			fallbackPath := dir + "/temp1_input"
			if _, err := os.Stat(fallbackPath); err == nil {
				return fallbackPath, nil
			}
		}
	}

	return "", fmt.Errorf("thermal sensor file not found for specified target")
}

func findPathByLabels(dir string, labels []string) (string, bool) {
	paths, err := filepath.Glob(filepath.Join(dir, "temp*_label"))
	if err != nil {
		return "", false
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		labelName := string(bytes.TrimSpace(content))
		if slices.Contains(labels, labelName) {
			inputPath := strings.TrimSuffix(path, "_label") + "_input"
			if _, err := os.Stat(inputPath); err == nil {
				return inputPath, true
			}
		}
	}
	return "", false
}

// ReadTemp accepts an absolute sysfs file path (string) to a thermal sensor and returns
// the current temperature converted to integer degrees Celsius (int).
// It returns a non-nil error if reading the file fails or if parsing fails.
func ReadTemp(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read temp file: %w", err)
	}
	temp, err := strconv.Atoi(string(bytes.TrimSpace(data)))
	if err != nil {
		return 0, fmt.Errorf("failed to convert temp to int: %w", err)
	}
	return temp / 1000, nil
}
