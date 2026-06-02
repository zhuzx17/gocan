//go:build !windows && !linux

package gocan

func getDeviceInfo(ch Channel) (DeviceInfo, error) {
	return DeviceInfo{}, ErrNotSupported
}
