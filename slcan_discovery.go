package gocan

import (
	"sort"
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
	ports, err := lookupSLCANPorts()
	if err != nil {
		return nil, err
	}
	sort.Slice(ports, func(i, j int) bool { return ports[i].Name < ports[j].Name })
	return ports, nil
}
