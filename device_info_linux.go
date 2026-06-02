//go:build linux

package gocan

import (
	"errors"
	"net"

	"github.com/Crush251/gocan/raw"
)

func getDeviceInfo(ch Channel) (DeviceInfo, error) {
	ifaceName, ok := raw.SocketCANInterface(ch)
	if !ok || ifaceName == "" {
		return DeviceInfo{}, ErrIllHandle
	}

	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return DeviceInfo{}, err
		}
		return DeviceInfo{}, ErrIllHandle
	}
	return DeviceInfo{
		Channel:       ch,
		Name:          iface.Name,
		Backend:       BackendSocketCAN,
		HardwareName:  iface.Name,
		InterfaceName: iface.Name,
		Up:            iface.Flags&net.FlagUp != 0,
		FD:            true,
	}, nil
}
