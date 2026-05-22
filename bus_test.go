package pcanbasic

import (
	"context"
	"errors"
	"testing"

	"github.com/Crush251/pcanbasic_go/raw"
)

// openTest 是测试专用入口，注入 fake adapter 打开 Classical 通道。
func openTest(t *testing.T, adapt rawAdapter, opts ...Option) *Bus {
	t.Helper()
	bus, err := openWith(adapt, USBBus1, false, "", opts...)
	if err != nil {
		t.Fatalf("openWith: %v", err)
	}
	return bus
}

// openTestFD 是测试专用入口，注入 fake adapter 打开 FD 通道。
func openTestFD(t *testing.T, adapt rawAdapter, opts ...Option) *Bus {
	t.Helper()
	bus, err := openWith(adapt, USBBus1, true, "f_clock=80000000", opts...)
	if err != nil {
		t.Fatalf("openWith FD: %v", err)
	}
	return bus
}

func TestOpen_InitFailureMaps(t *testing.T) {
	f := newFakeAdapter()
	f.initializeReturn = raw.PCAN_ERROR_INITIALIZE
	_, err := openWith(f, USBBus1, false, "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotInitialized) {
		t.Errorf("expected ErrNotInitialized, got %v", err)
	}
}

func TestOpenFD_InitFailureMaps(t *testing.T) {
	f := newFakeAdapter()
	f.initializeFDReturn = raw.PCAN_ERROR_ILLPARAMVAL
	_, err := openWith(f, USBBus1, true, "bad bitrate string")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrIllParamValue) {
		t.Errorf("expected ErrIllParamValue, got %v", err)
	}
}

func TestSend_OK(t *testing.T) {
	f := newFakeAdapter()
	bus := openTest(t, f)
	defer bus.Close()

	frame, _ := NewFrame(0x123, []byte{1, 2, 3})
	if err := bus.Send(context.Background(), frame); err != nil {
		t.Fatal(err)
	}
	if f.writeCalls != 1 {
		t.Errorf("writeCalls = %d, want 1", f.writeCalls)
	}
	if f.lastWrittenMsg.ID != 0x123 || f.lastWrittenMsg.Len != 3 {
		t.Errorf("bad msg: %+v", f.lastWrittenMsg)
	}
}

func TestSend_ExtendedFlag(t *testing.T) {
	f := newFakeAdapter()
	bus := openTest(t, f)
	defer bus.Close()

	frame, _ := NewExtendedFrame(0x1FFF, []byte{1})
	if err := bus.Send(context.Background(), frame); err != nil {
		t.Fatal(err)
	}
	if f.lastWrittenMsg.MsgType&raw.PCAN_MESSAGE_EXTENDED == 0 {
		t.Error("expected EXTENDED bit on msg")
	}
}

func TestSend_RemoteFlag(t *testing.T) {
	f := newFakeAdapter()
	bus := openTest(t, f)
	defer bus.Close()

	frame, _ := NewRemoteFrame(0x123, 4, false)
	if err := bus.Send(context.Background(), frame); err != nil {
		t.Fatal(err)
	}
	if f.lastWrittenMsg.MsgType&raw.PCAN_MESSAGE_RTR == 0 {
		t.Error("expected RTR bit on msg")
	}
}

func TestSend_AfterClose(t *testing.T) {
	f := newFakeAdapter()
	bus := openTest(t, f)
	bus.Close()
	frame, _ := NewFrame(0x1, nil)
	err := bus.Send(context.Background(), frame)
	if !errors.Is(err, ErrBusClosed) {
		t.Errorf("got %v, want ErrBusClosed", err)
	}
}

func TestSend_CtxCancel(t *testing.T) {
	f := newFakeAdapter()
	bus := openTest(t, f)
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	frame, _ := NewFrame(0x1, nil)
	err := bus.Send(ctx, frame)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("got %v, want context.Canceled", err)
	}
}

func TestSend_PCANWriteError(t *testing.T) {
	f := newFakeAdapter()
	f.writeReturn = raw.PCAN_ERROR_QXMTFULL
	bus := openTest(t, f)
	defer bus.Close()

	frame, _ := NewFrame(0x1, nil)
	err := bus.Send(context.Background(), frame)
	if !errors.Is(err, ErrQueueXmtFull) {
		t.Errorf("got %v, want ErrQueueXmtFull", err)
	}
}

func TestClose_Idempotent(t *testing.T) {
	f := newFakeAdapter()
	bus := openTest(t, f)
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatalf("second Close should not error: %v", err)
	}
	if f.uninitializeCalls != 1 {
		t.Errorf("uninitializeCalls = %d, want 1", f.uninitializeCalls)
	}
}

func TestSend_FDFrameOnClassicalBus(t *testing.T) {
	f := newFakeAdapter()
	bus := openTest(t, f)
	defer bus.Close()

	frame, _ := NewFDFrame(0x1, []byte{1, 2}, false, false)
	err := bus.Send(context.Background(), frame)
	if !errors.Is(err, ErrFDNotSupportedOnBus) {
		t.Errorf("got %v, want ErrFDNotSupportedOnBus", err)
	}
}

func TestSend_ClassicalFrameOnFDBus(t *testing.T) {
	f := newFakeAdapter()
	bus := openTestFD(t, f)
	defer bus.Close()

	frame, _ := NewFrame(0x1, []byte{1, 2})
	if err := bus.Send(context.Background(), frame); err != nil {
		t.Fatalf("Classical frame on FD bus should be allowed: %v", err)
	}
	if f.writeCalls != 1 {
		t.Errorf("expected Write to be used for Classical frame on FD bus, got writeCalls=%d",
			f.writeCalls)
	}
}

func TestSend_FDFrameOnFDBus(t *testing.T) {
	f := newFakeAdapter()
	bus := openTestFD(t, f)
	defer bus.Close()

	frame, _ := NewFDFrame(0x1, []byte{1, 2}, true, false)
	if err := bus.Send(context.Background(), frame); err != nil {
		t.Fatal(err)
	}
	if f.writeFDCalls != 1 {
		t.Errorf("writeFDCalls = %d, want 1", f.writeFDCalls)
	}
	if f.lastWrittenMsgFD.MsgType&raw.PCAN_MESSAGE_BRS == 0 {
		t.Error("expected BRS bit set on FD msg")
	}
	if f.lastWrittenMsgFD.MsgType&raw.PCAN_MESSAGE_FD == 0 {
		t.Error("expected FD bit set on FD msg")
	}
}

func TestSendMany_AllOK(t *testing.T) {
	f := newFakeAdapter()
	bus := openTest(t, f)
	defer bus.Close()

	frames := []Frame{}
	for i := 0; i < 5; i++ {
		fr, _ := NewFrame(uint32(0x100+i), []byte{byte(i)})
		frames = append(frames, fr)
	}
	if err := bus.SendMany(context.Background(), frames); err != nil {
		t.Fatal(err)
	}
	if f.writeCalls != 5 {
		t.Errorf("writeCalls = %d, want 5", f.writeCalls)
	}
}

func TestSendMany_PartialFailure(t *testing.T) {
	f := newFakeAdapter()
	f.writeSequence = []raw.TPCANStatus{
		raw.PCAN_ERROR_OK,
		raw.PCAN_ERROR_OK,
		raw.PCAN_ERROR_QXMTFULL,
	}
	bus := openTest(t, f)
	defer bus.Close()

	frames := []Frame{}
	for i := 0; i < 5; i++ {
		fr, _ := NewFrame(uint32(0x100+i), []byte{byte(i)})
		frames = append(frames, fr)
	}
	err := bus.SendMany(context.Background(), frames)
	if err == nil {
		t.Fatal("expected error")
	}
	var sme *SendManyError
	if !errors.As(err, &sme) {
		t.Fatalf("expected *SendManyError, got %T", err)
	}
	if sme.Index != 2 {
		t.Errorf("Index = %d, want 2", sme.Index)
	}
	if !errors.Is(err, ErrQueueXmtFull) {
		t.Error("expected errors.Is(err, ErrQueueXmtFull)")
	}
}

func TestSendMany_CtxCancel(t *testing.T) {
	f := newFakeAdapter()
	bus := openTest(t, f)
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	frames := []Frame{}
	for i := 0; i < 5; i++ {
		fr, _ := NewFrame(uint32(0x100+i), nil)
		frames = append(frames, fr)
	}
	err := bus.SendMany(ctx, frames)
	var sme *SendManyError
	if !errors.As(err, &sme) {
		t.Fatalf("expected *SendManyError, got %v", err)
	}
	if sme.Index != 0 {
		t.Errorf("Index = %d, want 0", sme.Index)
	}
	if !errors.Is(sme.Err, context.Canceled) {
		t.Errorf("Err = %v, want context.Canceled", sme.Err)
	}
}

func TestStatus_OK(t *testing.T) {
	f := newFakeAdapter()
	f.statusReturn = raw.PCAN_ERROR_OK
	bus := openTest(t, f)
	defer bus.Close()
	s, err := bus.Status()
	if err != nil {
		t.Fatal(err)
	}
	if s != StatusOK {
		t.Errorf("status = 0x%X, want OK", uint32(s))
	}
}

func TestStatus_AfterClose(t *testing.T) {
	f := newFakeAdapter()
	bus := openTest(t, f)
	bus.Close()
	_, err := bus.Status()
	if !errors.Is(err, ErrBusClosed) {
		t.Errorf("got %v, want ErrBusClosed", err)
	}
}

func TestReset_OK(t *testing.T) {
	f := newFakeAdapter()
	bus := openTest(t, f)
	defer bus.Close()
	if err := bus.Reset(); err != nil {
		t.Fatal(err)
	}
	if f.resetCalls != 1 {
		t.Errorf("resetCalls = %d, want 1", f.resetCalls)
	}
}

func TestReset_AfterClose(t *testing.T) {
	f := newFakeAdapter()
	bus := openTest(t, f)
	bus.Close()
	if err := bus.Reset(); !errors.Is(err, ErrBusClosed) {
		t.Errorf("got %v, want ErrBusClosed", err)
	}
}

// dataLenToDLC / dlcToDataLen 是 FD 长度编码相关的内部函数。
func TestDataLenDLC_Roundtrip(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 12, 16, 20, 24, 32, 48, 64} {
		dlc := dataLenToDLC(n)
		back := dlcToDataLen(dlc)
		if back != n {
			t.Errorf("len=%d → dlc=%d → back=%d", n, dlc, back)
		}
	}
}
