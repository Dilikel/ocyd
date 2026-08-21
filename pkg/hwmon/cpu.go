// Package hwmon provides discovery and reading capabilities for Linux sysfs thermal metrics.
package hwmon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var cpuDrivers = map[string]struct{}{
	"k10temp":    {},
	"coretemp":   {},
	"zenpower":   {},
	"amd_energy": {},
}

var cpuLabels = map[string]struct{}{
	"Tctl":         {},
	"Tdie":         {},
	"Package id 0": {},
	"Package id 1": {},
}

// CPUTempPath returns the absolute sysfs file path (string) to the primary CPU thermal sensor.
// It returns a non-nil error if the sysfs directory cannot be read or no matching driver is found.
func CPUTempPath() (string, error) {
	paths, err := filepath.Glob("/sys/class/hwmon/hwmon*/name")
	if err != nil {
		return "", fmt.Errorf("failed to glob hwmon: %w", err)
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		driverName := strings.TrimSpace(string(content))
		if _, ok := cpuDrivers[driverName]; ok {
			dir := filepath.Dir(path)
			if targetPath, ok := findPathByLabels(dir); ok {
				return targetPath, nil
			}

			fallbackPath := dir + "/temp1_input"
			if _, err := os.Stat(fallbackPath); err == nil {
				return fallbackPath, nil
			}
		}
	}
	return "", fmt.Errorf("cpu thermal sensor file not found")
}

func findPathByLabels(dir string) (string, bool) {
	paths, err := filepath.Glob(filepath.Join(dir, "temp*_label"))
	if err != nil {
		return "", false
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		labelText := strings.TrimSpace(string(content))
		if _, ok := cpuLabels[labelText]; ok {
			inputPath := strings.TrimSuffix(path, "_label") + "_input"
			if _, err := os.Stat(inputPath); err == nil {
				return inputPath, true
			}
		}
	}
	return "", false
}

// CPUTemp accepts an absolute sysfs file path (string) to a thermal sensor and returns
// the current CPU temperature converted to integer degrees Celsius (int).
// It returns a non-nil error if reading the file fails or if the content cannot be parsed as an integer.
func CPUTemp(path string) (int, error) {
	tempByte, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read cpu temp file: %w", err)
	}

	tempInt, err := strconv.Atoi(strings.TrimSpace(string(tempByte)))
	if err != nil {
		return 0, fmt.Errorf("failed to convert cpu temp to int: %w", err)
	}

	return tempInt / 1000, nil
}
