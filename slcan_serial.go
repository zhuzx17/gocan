package gocan

import "go.bug.st/serial"

func openSerialPort(name string, baud int) (slcanPort, error) {
	return serial.Open(name, &serial.Mode{
		BaudRate: baud,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	})
}
