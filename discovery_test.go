package gocan

import "testing"

func TestChannelBackends(t *testing.T) {
	if BackendPCAN != "pcan" {
		t.Fatalf("BackendPCAN = %q, want pcan", BackendPCAN)
	}
	if BackendSocketCAN != "socketcan" {
		t.Fatalf("BackendSocketCAN = %q, want socketcan", BackendSocketCAN)
	}
}

func TestLookupChannels_DoesNotPanic(t *testing.T) {
	channels, err := LookupChannels()
	if err != nil {
		t.Fatalf("LookupChannels failed: %v", err)
	}
	for _, ch := range channels {
		if ch.Channel == 0 {
			t.Fatalf("channel %q has zero handle", ch.Name)
		}
		if ch.Name == "" {
			t.Fatalf("channel %+v has empty name", ch)
		}
		if ch.Backend == "" {
			t.Fatalf("channel %+v has empty backend", ch)
		}
	}
}
