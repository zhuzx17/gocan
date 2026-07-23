package gocan

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultSLCANSerialBaud = 115200
	slcanReadTimeout       = 5 * time.Millisecond
	slcanMaxRecordSize     = 138
)

// SLCANBitrate is one of the nominal bitrate presets implemented by the
// CANable 2.0 SLCAN-FD firmware.
type SLCANBitrate uint8

const (
	SLCANBitrate10K SLCANBitrate = iota
	SLCANBitrate20K
	SLCANBitrate50K
	SLCANBitrate100K
	SLCANBitrate125K
	SLCANBitrate250K
	SLCANBitrate500K
	SLCANBitrate750K
	SLCANBitrate1M
	SLCANBitrate83K3
)

// SLCANDataBitrate is a CAN FD data-phase bitrate preset implemented by the
// CANable 2.0 firmware.
type SLCANDataBitrate uint8

const (
	SLCANDataBitrate2M SLCANDataBitrate = 2
	SLCANDataBitrate5M SLCANDataBitrate = 5
)

type slcanConfig struct {
	serialBaud     int
	silent         bool
	autoRetransmit bool
}

func defaultSLCANConfig() slcanConfig {
	return slcanConfig{
		serialBaud:     defaultSLCANSerialBaud,
		autoRetransmit: true,
	}
}

// WithSLCANSerialBaud sets the host serial line rate. CANable 2.0 uses USB CDC
// and normally ignores this value; 115200 is used by default.
func WithSLCANSerialBaud(baud int) Option {
	return func(c *config) {
		if baud > 0 {
			c.slcan.serialBaud = baud
		}
	}
}

// WithSLCANSilent selects listen-only mode before the CAN channel is opened.
func WithSLCANSilent(enabled bool) Option {
	return func(c *config) { c.slcan.silent = enabled }
}

// WithSLCANAutoRetransmit controls automatic retransmission. It defaults to on.
func WithSLCANAutoRetransmit(enabled bool) Option {
	return func(c *config) { c.slcan.autoRetransmit = enabled }
}

type slcanPort interface {
	io.ReadWriteCloser
	ResetInputBuffer() error
	ResetOutputBuffer() error
	SetReadTimeout(time.Duration) error
}

type slcanPortOpener func(name string, baud int) (slcanPort, error)

// OpenSLCAN opens a serial CANable 2.0 using the Classical CAN command set.
func OpenSLCAN(port string, bitrate SLCANBitrate, opts ...Option) (*Bus, error) {
	return openSLCANWith(openSerialPort, port, bitrate, 0, false, opts...)
}

// OpenSLCANFD opens a serial CANable 2.0 with its non-standard SLCAN-FD
// extensions. The returned Bus accepts both Classical CAN and CAN FD frames.
func OpenSLCANFD(port string, bitrate SLCANBitrate, dataBitrate SLCANDataBitrate, opts ...Option) (*Bus, error) {
	return openSLCANWith(openSerialPort, port, bitrate, dataBitrate, true, opts...)
}

func openSLCANWith(openPort slcanPortOpener, portName string, bitrate SLCANBitrate, dataBitrate SLCANDataBitrate, fd bool, opts ...Option) (*Bus, error) {
	portName = strings.TrimSpace(portName)
	if portName == "" {
		return nil, fmt.Errorf("open SLCAN: empty serial port: %w", ErrIllParamValue)
	}
	if bitrate > SLCANBitrate83K3 {
		return nil, fmt.Errorf("open SLCAN: nominal bitrate preset %d: %w", bitrate, ErrIllParamValue)
	}
	if fd && dataBitrate != SLCANDataBitrate2M && dataBitrate != SLCANDataBitrate5M {
		return nil, fmt.Errorf("open SLCAN: data bitrate preset %d: %w", dataBitrate, ErrIllParamValue)
	}

	cfg := newDefaultConfig()
	cfg.slcan = defaultSLCANConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.receiveMode == ModeEvent {
		return nil, fmt.Errorf("open SLCAN: event receive mode: %w", ErrNotSupported)
	}

	port, err := openPort(portName, cfg.slcan.serialBaud)
	if err != nil {
		return nil, fmt.Errorf("open SLCAN port %q: %w", portName, err)
	}
	cleanup := func(openErr error) (*Bus, error) {
		_ = port.Close()
		return nil, openErr
	}
	if err := port.SetReadTimeout(slcanReadTimeout); err != nil {
		return cleanup(fmt.Errorf("configure SLCAN port %q read timeout: %w", portName, err))
	}
	if err := port.ResetInputBuffer(); err != nil {
		return cleanup(fmt.Errorf("reset SLCAN port %q input: %w", portName, err))
	}
	if err := port.ResetOutputBuffer(); err != nil {
		return cleanup(fmt.Errorf("reset SLCAN port %q output: %w", portName, err))
	}

	backend := &slcanBackend{port: port, portName: portName}
	if err := backend.configure(bitrate, dataBitrate, fd, cfg.slcan); err != nil {
		return cleanup(err)
	}
	b := &Bus{
		slcan:   backend,
		cfg:     cfg,
		isFD:    fd,
		rxCh:    make(chan Frame, cfg.rxBufferSize),
		errCh:   make(chan error, cfg.errBufferSize),
		closing: make(chan struct{}),
	}
	b.startReader()
	return b, nil
}

type slcanBackend struct {
	port     slcanPort
	portName string

	writeMu   sync.Mutex
	readMu    sync.Mutex
	closeOnce sync.Once
	closed    atomic.Bool
	rxBuf     []byte
}

func (s *slcanBackend) configure(nominal SLCANBitrate, data SLCANDataBitrate, fd bool, cfg slcanConfig) error {
	var commands strings.Builder
	commands.WriteString("C\r")
	fmt.Fprintf(&commands, "S%d\r", nominal)
	if fd {
		fmt.Fprintf(&commands, "Y%d\r", data)
	}
	if cfg.silent {
		commands.WriteString("M1\r")
	} else {
		commands.WriteString("M0\r")
	}
	if cfg.autoRetransmit {
		commands.WriteString("A1\r")
	} else {
		commands.WriteString("A0\r")
	}
	commands.WriteString("O\r")
	if err := s.write([]byte(commands.String())); err != nil {
		return fmt.Errorf("configure SLCAN port %q: %w", s.portName, err)
	}
	return nil
}

func (s *slcanBackend) send(frame Frame) error {
	record, err := encodeSLCANFrame(frame)
	if err != nil {
		return err
	}
	if err := s.write(record); err != nil {
		return fmt.Errorf("write SLCAN port %q: %w", s.portName, err)
	}
	return nil
}

func (s *slcanBackend) reset() error {
	if err := s.write([]byte("C\rO\r")); err != nil {
		return fmt.Errorf("reset SLCAN port %q: %w", s.portName, err)
	}
	return nil
}

func (s *slcanBackend) write(data []byte) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if s.closed.Load() {
		return ErrBusClosed
	}
	return writeFull(s.port, data)
}

func writeFull(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}
		if n <= 0 || n > len(data) {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}

func (s *slcanBackend) readFrame() (Frame, error) {
	s.readMu.Lock()
	defer s.readMu.Unlock()
	for {
		if record, ok := s.nextRecord(); ok {
			if len(record) == 0 {
				continue
			}
			frame, err := parseSLCANFrame(record)
			if err != nil {
				return Frame{}, fmt.Errorf("read SLCAN port %q: %w", s.portName, err)
			}
			frame.ReceivedAt = time.Now()
			return frame, nil
		}
		if len(s.rxBuf) > slcanMaxRecordSize {
			s.rxBuf = nil
			return Frame{}, fmt.Errorf("read SLCAN port %q: record exceeds %d bytes: %w", s.portName, slcanMaxRecordSize, ErrSLCANProtocol)
		}

		var buf [256]byte
		n, err := s.port.Read(buf[:])
		if n > 0 {
			s.rxBuf = append(s.rxBuf, buf[:n]...)
		}
		if err != nil {
			if s.closed.Load() {
				return Frame{}, errQueueEmpty
			}
			return Frame{}, fmt.Errorf("read SLCAN port %q: %w", s.portName, err)
		}
		if n == 0 {
			return Frame{}, errQueueEmpty
		}
	}
}

func (s *slcanBackend) nextRecord() ([]byte, bool) {
	idx := bytes.IndexByte(s.rxBuf, '\r')
	if idx < 0 {
		return nil, false
	}
	record := bytes.TrimSpace(s.rxBuf[:idx])
	s.rxBuf = s.rxBuf[idx+1:]
	return record, true
}

func (s *slcanBackend) close() error {
	var closeErr error
	s.closeOnce.Do(func() {
		s.writeMu.Lock()
		s.closed.Store(true)
		commandErr := writeFull(s.port, []byte("C\r"))
		portErr := s.port.Close()
		s.writeMu.Unlock()
		if commandErr != nil {
			commandErr = fmt.Errorf("close SLCAN channel %q: %w", s.portName, commandErr)
		}
		if portErr != nil {
			portErr = fmt.Errorf("close SLCAN port %q: %w", s.portName, portErr)
		}
		closeErr = errors.Join(commandErr, portErr)
	})
	return closeErr
}

func encodeSLCANFrame(frame Frame) ([]byte, error) {
	if frame.Has(FlagESI) {
		return nil, ErrSLCANESINotSupported
	}
	if frame.Has(FlagFD) && frame.Has(FlagRemote) {
		return nil, ErrRemoteOnFD
	}
	limit := maxStdID
	idWidth := 3
	if frame.Has(FlagExtended) {
		limit = maxExtID
		idWidth = 8
	}
	if frame.ID > limit {
		return nil, ErrIDOutOfRange
	}

	prefix := byte('t')
	dlc := len(frame.Data)
	if frame.Has(FlagFD) {
		if !fdValidLengths[dlc] {
			return nil, ErrInvalidFDLength
		}
		if frame.Has(FlagBRS) {
			prefix = 'b'
		} else {
			prefix = 'd'
		}
	} else {
		if dlc > 8 {
			return nil, ErrDataTooLong
		}
		if frame.Has(FlagRemote) {
			prefix = 'r'
		}
	}
	if frame.Has(FlagExtended) {
		prefix -= 'a' - 'A'
	}

	payloadLen := len(frame.Data) * 2
	if frame.Has(FlagRemote) {
		payloadLen = 0
	}
	record := make([]byte, 0, 1+idWidth+1+payloadLen+1)
	record = append(record, prefix)
	record = strconv.AppendUint(record, uint64(frame.ID), 16)
	idStart := 1
	for len(record)-idStart < idWidth {
		record = append(record, 0)
		copy(record[idStart+1:], record[idStart:])
		record[idStart] = '0'
	}
	for i := idStart; i < len(record); i++ {
		if record[i] >= 'a' && record[i] <= 'f' {
			record[i] -= 'a' - 'A'
		}
	}
	dlcCode := dlc
	if frame.Has(FlagFD) {
		dlcCode = int(dataLenToDLC(dlc))
	}
	record = append(record, "0123456789ABCDEF"[dlcCode])
	if !frame.Has(FlagRemote) {
		encodedStart := len(record)
		record = append(record, make([]byte, hex.EncodedLen(len(frame.Data)))...)
		hex.Encode(record[encodedStart:], frame.Data)
		for i := encodedStart; i < len(record); i++ {
			if record[i] >= 'a' && record[i] <= 'f' {
				record[i] -= 'a' - 'A'
			}
		}
	}
	record = append(record, '\r')
	return record, nil
}

func parseSLCANFrame(record []byte) (Frame, error) {
	if len(record) < 5 {
		return Frame{}, ErrSLCANProtocol
	}
	prefix := record[0]
	var flags FrameFlags
	idWidth := 3
	switch prefix {
	case 't':
	case 'T':
		flags |= FlagExtended
		idWidth = 8
	case 'r':
		flags |= FlagRemote
	case 'R':
		flags |= FlagRemote | FlagExtended
		idWidth = 8
	case 'd':
		flags |= FlagFD
	case 'D':
		flags |= FlagFD | FlagExtended
		idWidth = 8
	case 'b':
		flags |= FlagFD | FlagBRS
	case 'B':
		flags |= FlagFD | FlagBRS | FlagExtended
		idWidth = 8
	default:
		return Frame{}, ErrSLCANProtocol
	}
	if len(record) < 1+idWidth+1 {
		return Frame{}, ErrSLCANProtocol
	}
	id, err := strconv.ParseUint(string(record[1:1+idWidth]), 16, 32)
	if err != nil {
		return Frame{}, fmt.Errorf("invalid SLCAN identifier: %w", ErrSLCANProtocol)
	}
	limit := uint64(maxStdID)
	if flags&FlagExtended != 0 {
		limit = uint64(maxExtID)
	}
	if id > limit {
		return Frame{}, fmt.Errorf("SLCAN identifier out of range: %w", ErrSLCANProtocol)
	}

	dlc, ok := hexNibble(record[1+idWidth])
	if !ok {
		return Frame{}, fmt.Errorf("invalid SLCAN DLC: %w", ErrSLCANProtocol)
	}
	dataLen := dlc
	if flags&FlagFD != 0 {
		dataLen = dlcToDataLen(uint8(dlc))
	} else if dlc > 8 {
		return Frame{}, fmt.Errorf("classical SLCAN DLC exceeds 8: %w", ErrSLCANProtocol)
	}
	payload := record[1+idWidth+1:]
	if flags&FlagRemote != 0 {
		// CANable 2.0 firmware appends DLC-sized placeholder bytes when it
		// reports a received RTR frame, although transmit commands omit them.
		if len(payload) != 0 && len(payload) != dataLen*2 {
			return Frame{}, fmt.Errorf("remote SLCAN placeholder length is %d, want 0 or %d: %w", len(payload), dataLen*2, ErrSLCANProtocol)
		}
		if len(payload) > 0 {
			placeholder := make([]byte, dataLen)
			if _, err := hex.Decode(placeholder, payload); err != nil {
				return Frame{}, fmt.Errorf("invalid remote SLCAN placeholder: %w", ErrSLCANProtocol)
			}
		}
		return Frame{ID: uint32(id), Data: make([]byte, dataLen), Flags: flags}, nil
	}
	if len(payload) != dataLen*2 {
		return Frame{}, fmt.Errorf("SLCAN payload length is %d, want %d: %w", len(payload), dataLen*2, ErrSLCANProtocol)
	}
	data := make([]byte, dataLen)
	if _, err := hex.Decode(data, payload); err != nil {
		return Frame{}, fmt.Errorf("invalid SLCAN payload: %w", ErrSLCANProtocol)
	}
	return Frame{ID: uint32(id), Data: data, Flags: flags}, nil
}

func hexNibble(b byte) (int, bool) {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0'), true
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10, true
	case b >= 'A' && b <= 'F':
		return int(b-'A') + 10, true
	default:
		return 0, false
	}
}
