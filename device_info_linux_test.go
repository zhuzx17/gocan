//go:build linux

package gocan

import (
	"errors"
	"testing"
)

func TestGetDeviceInfo_RejectsUnknownSocketCANHandle(t *testing.T) {
	_, err := GetDeviceInfo(USBBus1)
	if !errors.Is(err, ErrIllHandle) {
		t.Fatalf("GetDeviceInfo(USBBus1) error = %v, want ErrIllHandle", err)
	}
}

func TestGetDeviceInfo_ReturnsSocketCANInterface(t *testing.T) {
	channels, err := LookupChannels()
	if err != nil {
		t.Fatalf("LookupChannels failed: %v", err)
	}
	if len(channels) == 0 {
		t.Skip("no SocketCAN interface found")
	}

	info, err := GetDeviceInfo(channels[0].Channel)
	if err != nil {
		t.Fatalf("GetDeviceInfo failed: %v", err)
	}
	if info.Channel != channels[0].Channel {
		t.Fatalf("Channel = 0x%X, want 0x%X", info.Channel, channels[0].Channel)
	}
	if info.Name != channels[0].Name || info.InterfaceName != channels[0].Name {
		t.Fatalf("name/interface = %q/%q, want %q", info.Name, info.InterfaceName, channels[0].Name)
	}
	if info.Backend != BackendSocketCAN {
		t.Fatalf("Backend = %q, want %q", info.Backend, BackendSocketCAN)
	}
}
