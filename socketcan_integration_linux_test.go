//go:build linux && socketcan_integration

package gocan

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestSocketCANIntegration_OpenSendClose(t *testing.T) {
	iface := os.Getenv("GOCAN_SOCKETCAN_IFACE")
	if iface == "" {
		t.Skip("set GOCAN_SOCKETCAN_IFACE to run SocketCAN integration tests")
	}

	bus, err := Open(SocketCAN(iface), WithReceiveMode(ModePolling), WithPollInterval(time.Millisecond))
	if err != nil {
		t.Fatalf("Open(SocketCAN(%q)) failed: %v", iface, err)
	}
	defer bus.Close()

	frame, err := NewFrame(0x123, []byte{1, 2, 3, 4})
	if err != nil {
		t.Fatalf("NewFrame failed: %v", err)
	}
	if err := bus.Send(context.Background(), frame); err != nil {
		t.Fatalf("Send failed: %v", err)
	}
}
