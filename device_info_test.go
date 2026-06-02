package gocan

import "testing"

func TestDeviceInfoZeroValue(t *testing.T) {
	var info DeviceInfo
	if info.Channel != 0 || info.Name != "" || info.Backend != "" || info.Up || info.FD {
		t.Fatalf("zero DeviceInfo should be empty, got %+v", info)
	}
}
