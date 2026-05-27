package gocan

import "testing"

func TestSocketCAN_ReturnsStableHandle(t *testing.T) {
	ch1 := SocketCAN("vcan0")
	ch2 := SocketCAN("vcan0")
	if ch1 == 0 {
		t.Fatal("SocketCAN returned zero handle")
	}
	if ch1 != ch2 {
		t.Fatalf("SocketCAN returned different handles: 0x%X != 0x%X", ch1, ch2)
	}
}

func TestSocketCAN_DifferentInterfaces(t *testing.T) {
	ch1 := SocketCAN("can0")
	ch2 := SocketCAN("can1")
	if ch1 == ch2 {
		t.Fatalf("SocketCAN returned same handle for different interfaces: 0x%X", ch1)
	}
}

func TestSocketCAN_EmptyInterface(t *testing.T) {
	if got := SocketCAN(""); got != 0 {
		t.Fatalf("SocketCAN empty interface = 0x%X, want 0", got)
	}
}
