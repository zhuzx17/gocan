//go:build windows

package gocan

import (
	"strings"
	"unsafe"

	"github.com/zhuzxdev/gocan/raw"
)

func getDeviceInfo(ch Channel) (DeviceInfo, error) {
	if err := raw.EnsureLoaded(); err != nil {
		return DeviceInfo{}, ErrDLLNotFound
	}

	var condition uint32
	status := raw.GetValue(ch, raw.PCAN_CHANNEL_CONDITION,
		unsafe.Pointer(&condition), uint32(unsafe.Sizeof(condition)))
	if status != raw.PCAN_ERROR_OK || condition == pcanChannelUnavailable {
		return DeviceInfo{}, ErrIllHandle
	}

	info := DeviceInfo{
		Channel: ch,
		Name:    pcanChannelName(ch),
		Backend: BackendPCAN,
		Up:      condition == pcanChannelAvailable,
	}
	info.HardwareName = pcanStringValue(ch, raw.PCAN_HARDWARE_NAME)
	info.DeviceNumber = pcanUint32Value(ch, raw.PCAN_DEVICE_NUMBER)
	info.ControllerNumber = pcanUint32Value(ch, raw.PCAN_CONTROLLER_NUMBER)
	info.Features = pcanUint32Value(ch, raw.PCAN_CHANNEL_FEATURES)
	info.FD = info.Features != 0
	return info, nil
}

func pcanChannelName(ch Channel) string {
	for _, candidate := range pcanDiscoveryChannels() {
		if candidate.channel == ch {
			return candidate.name
		}
	}
	return "PCAN_UNKNOWN"
}

func pcanUint32Value(ch Channel, p raw.TPCANParameter) uint32 {
	var value uint32
	status := raw.GetValue(ch, p, unsafe.Pointer(&value), uint32(unsafe.Sizeof(value)))
	if status != raw.PCAN_ERROR_OK {
		return 0
	}
	return value
}

func pcanStringValue(ch Channel, p raw.TPCANParameter) string {
	var buf [256]byte
	status := raw.GetValue(ch, p, unsafe.Pointer(&buf[0]), uint32(len(buf)))
	if status != raw.PCAN_ERROR_OK {
		return ""
	}
	n := 0
	for n < len(buf) && buf[n] != 0 {
		n++
	}
	return strings.TrimSpace(string(buf[:n]))
}
