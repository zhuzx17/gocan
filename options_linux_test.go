//go:build linux

package gocan

import "testing"

func TestWithLoopback_SetsField(t *testing.T) {
	cfg := newDefaultConfig()
	WithLoopback(false)(cfg)
	if cfg.linux.loopback == nil || *cfg.linux.loopback != false {
		t.Errorf("loopback = %v, want pointer to false", cfg.linux.loopback)
	}
}

func TestWithRecvOwnMsgs_SetsField(t *testing.T) {
	cfg := newDefaultConfig()
	WithRecvOwnMsgs(true)(cfg)
	if cfg.linux.recvOwnMsgs == nil || *cfg.linux.recvOwnMsgs != true {
		t.Errorf("recvOwnMsgs = %v, want pointer to true", cfg.linux.recvOwnMsgs)
	}
}

func TestWithErrFilter_SetsField(t *testing.T) {
	cfg := newDefaultConfig()
	WithErrFilter(CANErrBusOff | CANErrTxTimeout)(cfg)
	if cfg.linux.errFilter == nil {
		t.Fatal("errFilter is nil")
	}
	want := uint32(CANErrBusOff | CANErrTxTimeout)
	if *cfg.linux.errFilter != want {
		t.Errorf("errFilter = 0x%X, want 0x%X", *cfg.linux.errFilter, want)
	}
}
