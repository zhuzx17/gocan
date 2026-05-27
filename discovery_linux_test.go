//go:build linux

package gocan

import "testing"

func TestIsSocketCANInterface(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"can0", true},
		{"vcan0", true},
		{"slcan0", true},
		{"vxcan0", true},
		{"eth0", false},
		{"lo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSocketCANInterface(tt.name); got != tt.want {
				t.Fatalf("isSocketCANInterface(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
