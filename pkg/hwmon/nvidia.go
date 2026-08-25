package hwmon

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// NVMLSensor represents a GPU thermal sensor accessed via NVIDIA Management Library (NVML).
type NVMLSensor struct {
	Device nvml.Device
}

func newNVMLSensor() (NVMLSensor, error) {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return NVMLSensor{}, fmt.Errorf("failed to initialize NVML: %s", nvml.ErrorString(ret))
	}
	device, ret := nvml.DeviceGetHandleByIndex(0)
	if ret != nvml.SUCCESS {
		return NVMLSensor{}, fmt.Errorf("failed to get nvidia device handle: %s", nvml.ErrorString(ret))
	}
	return NVMLSensor{
		Device: device,
	}, nil
}

// Read queries the current GPU temperature in Celsius via NVML.
func (n NVMLSensor) Read() (int, error) {
	temp, ret := n.Device.GetTemperature(nvml.TEMPERATURE_GPU)
	if ret != nvml.SUCCESS {
		return 0, fmt.Errorf("failed to get GPU temperature: %s", nvml.ErrorString(ret))
	}
	return int(temp), nil
}

// Close shuts down the NVML context and releases system resources.
func (n NVMLSensor) Close() error {
	if ret := nvml.Shutdown(); ret != nvml.SUCCESS {
		return fmt.Errorf("failed to shutdown NVML: %s", nvml.ErrorString(ret))
	}
	return nil
}
