package gocan

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"
)

func TestSLCANFrameCodec(t *testing.T) {
	fd12 := make([]byte, 12)
	for i := range fd12 {
		fd12[i] = byte(i)
	}
	fd64 := make([]byte, 64)
	for i := range fd64 {
		fd64[i] = byte(0x80 + i)
	}
	tests := []struct {
		name   string
		frame  Frame
		record string
	}{
		{
			name:   "classical standard",
			frame:  Frame{ID: 0x123, Data: []byte{0xDE, 0xAD, 0xBE}},
			record: "t1233DEADBE\r",
		},
		{
			name:   "classical extended",
			frame:  Frame{ID: 0x1ABCDE, Data: []byte{1, 2}, Flags: FlagExtended},
			record: "T001ABCDE20102\r",
		},
		{
			name:   "remote standard",
			frame:  Frame{ID: 0x321, Data: make([]byte, 8), Flags: FlagRemote},
			record: "r3218\r",
		},
		{
			name:   "FD standard no BRS",
			frame:  Frame{ID: 0x456, Data: fd12, Flags: FlagFD},
			record: "d4569000102030405060708090A0B\r",
		},
		{
			name:   "FD extended BRS 64 bytes",
			frame:  Frame{ID: 0x1ABCDE, Data: fd64, Flags: FlagFD | FlagBRS | FlagExtended},
			record: "B001ABCDEF808182838485868788898A8B8C8D8E8F909192939495969798999A9B9C9D9E9FA0A1A2A3A4A5A6A7A8A9AAABACADAEAFB0B1B2B3B4B5B6B7B8B9BABBBCBDBEBF\r",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := encodeSLCANFrame(tt.frame)
			if err != nil {
				t.Fatalf("encodeSLCANFrame: %v", err)
			}
			if string(record) != tt.record {
				t.Fatalf("record = %q, want %q", record, tt.record)
			}
			got, err := parseSLCANFrame(record[:len(record)-1])
			if err != nil {
				t.Fatalf("parseSLCANFrame: %v", err)
			}
			if got.ID != tt.frame.ID || got.Flags != tt.frame.Flags || !bytes.Equal(got.Data, tt.frame.Data) {
				t.Fatalf("round trip = %+v, want %+v", got, tt.frame)
			}
		})
	}
}

func TestSLCANFrameCodecRejectsUnsupportedFrames(t *testing.T) {
	tests := []struct {
		name  string
		frame Frame
		want  error
	}{
		{"ESI", Frame{ID: 1, Flags: FlagFD | FlagESI}, ErrSLCANESINotSupported},
		{"remote FD", Frame{ID: 1, Flags: FlagFD | FlagRemote}, ErrRemoteOnFD},
		{"standard ID", Frame{ID: 0x800}, ErrIDOutOfRange},
		{"classical length", Frame{ID: 1, Data: make([]byte, 9)}, ErrDataTooLong},
		{"FD length", Frame{ID: 1, Data: make([]byte, 10), Flags: FlagFD}, ErrInvalidFDLength},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encodeSLCANFrame(tt.frame)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestParseSLCANFrameRejectsMalformedRecords(t *testing.T) {
	for _, record := range []string{
		"",
		"x1230",
		"tGGG0",
		"t8000",
		"t1239",
		"t1232AA",
		"r1231A",
		"d1239AA",
		"D200000000",
	} {
		t.Run(record, func(t *testing.T) {
			_, err := parseSLCANFrame([]byte(record))
			if !errors.Is(err, ErrSLCANProtocol) {
				t.Fatalf("parse %q error = %v, want ErrSLCANProtocol", record, err)
			}
		})
	}
}

func TestParseSLCANFrameAcceptsCANableRemotePlaceholder(t *testing.T) {
	frame, err := parseSLCANFrame([]byte("r1234DEADBEEF"))
	if err != nil {
		t.Fatalf("parseSLCANFrame: %v", err)
	}
	if frame.ID != 0x123 || frame.Flags != FlagRemote || len(frame.Data) != 4 {
		t.Fatalf("frame = %+v", frame)
	}
}

func TestOpenSLCANFDConfiguresCANable2(t *testing.T) {
	port := newFakeSLCANPort()
	var gotName string
	var gotBaud int
	opener := func(name string, baud int) (slcanPort, error) {
		gotName, gotBaud = name, baud
		return port, nil
	}
	bus, err := openSLCANWith(opener, "COM7", SLCANBitrate500K, SLCANDataBitrate5M, true,
		WithSLCANSerialBaud(230400),
		WithSLCANSilent(true),
		WithSLCANAutoRetransmit(false),
		WithReceiveMode(ModePolling),
	)
	if err != nil {
		t.Fatalf("openSLCANWith: %v", err)
	}
	if gotName != "COM7" || gotBaud != 230400 {
		t.Fatalf("open args = %q %d, want COM7 230400", gotName, gotBaud)
	}
	if got := port.written(); got != "C\rS6\rY5\rM1\rA0\rO\r" {
		t.Fatalf("configuration = %q", got)
	}
	if port.timeout != slcanReadTimeout || !port.inputReset || !port.outputReset {
		t.Fatalf("serial setup timeout=%v inputReset=%v outputReset=%v", port.timeout, port.inputReset, port.outputReset)
	}
	if err := bus.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if got := port.written(); got != "C\rS6\rY5\rM1\rA0\rO\rC\r" {
		t.Fatalf("configuration + close = %q", got)
	}
}

func TestSLCANBusSendReceiveAndReset(t *testing.T) {
	port := newFakeSLCANPort()
	opener := func(string, int) (slcanPort, error) { return port, nil }
	bus, err := openSLCANWith(opener, "COM8", SLCANBitrate500K, SLCANDataBitrate2M, true,
		WithReceiveMode(ModePolling), WithPollInterval(time.Millisecond))
	if err != nil {
		t.Fatalf("openSLCANWith: %v", err)
	}
	defer bus.Close()

	frame, _ := NewFDFrame(0x123, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, true, false)
	if err := bus.Send(context.Background(), frame); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got := port.written(); got != "C\rS6\rY2\rM0\rA1\rO\rb1239000102030405060708090A0B\r" {
		t.Fatalf("writes = %q", got)
	}

	port.push([]byte("D001ABCDE9"))
	port.push([]byte("000102030405060708090A0B\r"))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	got, err := bus.ReadOne(ctx)
	if err != nil {
		t.Fatalf("ReadOne: %v", err)
	}
	if got.ID != 0x1ABCDE || got.Flags != FlagFD|FlagExtended || !bytes.Equal(got.Data, frame.Data) || got.ReceivedAt.IsZero() {
		t.Fatalf("received = %+v", got)
	}

	if err := bus.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if got := port.written(); !bytes.HasSuffix([]byte(got), []byte("C\rO\r")) {
		t.Fatalf("reset command missing from %q", got)
	}
	if _, err := bus.Status(); !errors.Is(err, ErrNotSupported) {
		t.Fatalf("Status error = %v, want ErrNotSupported", err)
	}
	if err := bus.SetFilter(0, 1, FilterStandard); !errors.Is(err, ErrNotSupported) {
		t.Fatalf("SetFilter error = %v, want ErrNotSupported", err)
	}
}

func TestOpenSLCANValidation(t *testing.T) {
	opener := func(string, int) (slcanPort, error) {
		t.Fatal("opener must not be called")
		return nil, nil
	}
	for _, tc := range []struct {
		name string
		port string
		nom  SLCANBitrate
		data SLCANDataBitrate
		fd   bool
		opts []Option
	}{
		{name: "empty port", port: " ", nom: SLCANBitrate500K},
		{name: "nominal bitrate", port: "COM1", nom: 10},
		{name: "data bitrate", port: "COM1", nom: SLCANBitrate500K, data: 3, fd: true},
		{name: "event mode", port: "COM1", nom: SLCANBitrate500K, opts: []Option{WithReceiveMode(ModeEvent)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := openSLCANWith(opener, tc.port, tc.nom, tc.data, tc.fd, tc.opts...)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

type fakeSLCANPort struct {
	mu sync.Mutex

	writes      bytes.Buffer
	readChunks  [][]byte
	timeout     time.Duration
	inputReset  bool
	outputReset bool
	closed      bool
}

func newFakeSLCANPort() *fakeSLCANPort { return &fakeSLCANPort{} }

func (p *fakeSLCANPort) Read(dst []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return 0, io.ErrClosedPipe
	}
	if len(p.readChunks) == 0 {
		return 0, nil
	}
	chunk := p.readChunks[0]
	n := copy(dst, chunk)
	if n == len(chunk) {
		p.readChunks = p.readChunks[1:]
	} else {
		p.readChunks[0] = chunk[n:]
	}
	return n, nil
}

func (p *fakeSLCANPort) Write(src []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return 0, io.ErrClosedPipe
	}
	return p.writes.Write(src)
}

func (p *fakeSLCANPort) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	return nil
}

func (p *fakeSLCANPort) ResetInputBuffer() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.inputReset = true
	p.readChunks = nil
	return nil
}

func (p *fakeSLCANPort) ResetOutputBuffer() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.outputReset = true
	p.writes.Reset()
	return nil
}

func (p *fakeSLCANPort) SetReadTimeout(timeout time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.timeout = timeout
	return nil
}

func (p *fakeSLCANPort) push(chunk []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.readChunks = append(p.readChunks, append([]byte(nil), chunk...))
}

func (p *fakeSLCANPort) written() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.writes.String()
}
