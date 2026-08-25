// Package hwmon provides discovery and reading capabilities for Linux sysfs thermal metrics and NVIDIA NVML.
package hwmon

import (
	"fmt"
)

// ThermalReader defines an interface for reading hardware temperature metrics
// and releasing associated resources.
type ThermalReader interface {
	Read() (int, error)
	Close() error
}

// NewCPUReader searches sysfs for known CPU thermal drivers (e.g., k10temp, coretemp)
// and returns a ThermalReader for the detected sensor.
func NewCPUReader() (ThermalReader, error) {
	cpuPath, err := findSensorPath(cpuTarget)
	if err != nil {
		return nil, fmt.Errorf("failed to get cpu temp path: %w", err)
	}
	return SysfsSensor{
		Path: cpuPath,
	}, nil
}

// NewGPUReader attempts to find a GPU thermal sensor via sysfs first.
// If no sysfs sensor is found, it falls back to initializing an NVML sensor.
func NewGPUReader() (ThermalReader, error) {
	if gpuPath, err := findSensorPath(gpuTarget); err == nil {
		return SysfsSensor{
			Path: gpuPath,
		}, nil
	}

	if nvSensor, err := newNVMLSensor(); err == nil {
		return nvSensor, nil
	}
	return nil, fmt.Errorf("no supported GPU thermal sensor found via sysfs or NVML")
}

// CToF converts a temperature value from Celsius to Fahrenheit.
func CToF(celsius int) int {
	return (celsius * 9 / 5) + 32
}
