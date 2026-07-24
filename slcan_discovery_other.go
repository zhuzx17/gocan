//go:build !windows && !linux

package gocan

import "go.bug.st/serial"

func lookupSLCANPorts() ([]SLCANPortInfo, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	result := make([]SLCANPortInfo, 0, len(ports))
	for _, name := range ports {
		result = append(result, SLCANPortInfo{Name: name})
	}
	return result, nil
}
