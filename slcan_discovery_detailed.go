//go:build windows || linux

package gocan

import (
	"strings"

	"go.bug.st/serial/enumerator"
)

func lookupSLCANPorts() ([]SLCANPortInfo, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}
	result := make([]SLCANPortInfo, 0, len(ports))
	for _, port := range ports {
		if port == nil {
			continue
		}
		vid := strings.ToUpper(port.VID)
		pid := strings.ToUpper(port.PID)
		result = append(result, SLCANPortInfo{
			Name:         port.Name,
			VID:          vid,
			PID:          pid,
			SerialNumber: port.SerialNumber,
			Product:      port.Product,
			CANable2:     port.IsUSB && vid == "16D0" && pid == "117E",
		})
	}
	return result, nil
}
