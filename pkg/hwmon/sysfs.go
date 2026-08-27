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

// SysfsSensor represents a thermal sensor accessed via Linux sysfs interface (/sys/class/hwmon).
type SysfsSensor struct {
	Path string
}

// Read reads the raw temperature value from sysfs file and returns it in Celsius.
func (s SysfsSensor) Read() (int, error) {
	temp, err := readTemp(s.Path)
	if err != nil {
		return 0, fmt.Errorf("failed to get temperature from sysfs: %w", err)
	}
	return temp, nil
}

// Close implements ThermalReader interface for SysfsSensor (no-op for sysfs).
func (s SysfsSensor) Close() error {
	return nil
}

type sensorTarget struct {
	Drivers []string
	Labels  []string
}

var (
	cpuTarget = sensorTarget{
		Drivers: []string{
			"k10temp", "coretemp", "zenpower", "amd_energy",
		},
		Labels: []string{
			"Tctl", "Tdie", "Package id 0", "Package id 1",
		},
	}
	gpuTarget = sensorTarget{
		Drivers: []string{
			"amdgpu", "nouveau", "nvidia", "i915", "xe",
		},
		Labels: []string{
			"edge", "junction", "hotspot", "GPU",
		},
	}
)

func findSensorPath(target sensorTarget) (string, error) {
	paths, err := filepath.Glob("/sys/class/hwmon/hwmon*/name")
	if err != nil {
		return "", fmt.Errorf("failed to glob hwmon: %w", err)
	}
	for _, path := range paths {
		//nolint:gosec // Sysfs paths are safe hardware nodes scanned by internal logic
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
		//nolint:gosec // Sysfs paths are safe hardware nodes scanned by internal logic
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

func readTemp(path string) (int, error) {
	//nolint:gosec // Sysfs paths are safe hardware nodes scanned by internal logic
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
