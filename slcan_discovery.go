package gocan

import (
	"sort"
	"strings"

	"go.bug.st/serial/enumerator"
)

// SLCANPortInfo describes a USB serial port that can be passed to OpenSLCAN or
// OpenSLCANFD. CANable2 is true for the official firmware VID:PID 16D0:117E.
type SLCANPortInfo struct {
	Name         string
	VID          string
	PID          string
	SerialNumber string
	Product      string
	CANable2     bool
}

// LookupSLCANPorts enumerates USB serial ports, including Windows COM ports.
// It does not hide non-CANable devices so callers can also use compatible
// SLCAN-FD firmware with a different USB identifier.
func LookupSLCANPorts() ([]SLCANPortInfo, error) {
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
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}
