package pcanbasic

import (
	"sync"

	"github.com/Crush251/pcanbasic_go/raw"
)

// fakeAdapter 是测试用 rawAdapter 实现：可编程返回状态、记录调用、注入收到的帧。
//
// 所有字段加锁保护，便于测试时从多 goroutine 安全访问（如 reader + Send 并发场景）。
type fakeAdapter struct {
	mu sync.Mutex

	// 调用计数。
	initializeCalls   int
	initializeFDCalls int
	uninitializeCalls int
	writeCalls        int
	writeFDCalls      int
	resetCalls        int
	readCalls         int
	readFDCalls       int

	// 行为控制：单次固定返回值。
	initializeReturn   raw.TPCANStatus
	initializeFDReturn raw.TPCANStatus
	uninitializeReturn raw.TPCANStatus
	writeReturn        raw.TPCANStatus
	writeFDReturn      raw.TPCANStatus
	resetReturn        raw.TPCANStatus
	statusReturn       raw.TPCANStatus
	errorTextReturn    string

	// 自定义 Write 行为：按调用次数返回不同状态（用于 SendMany 测试）。
	writeSequence   []raw.TPCANStatus
	writeFDSequence []raw.TPCANStatus

	// 收到的最后一帧（便于断言写出内容）。
	lastWrittenMsg   *raw.TPCANMsg
	lastWrittenMsgFD *raw.TPCANMsgFD

	// 待派发的接收帧（reader 模式下从这里取）。
	rxQueue   []rxItem
	rxFDQueue []rxFDItem
}

type rxItem struct {
	msg raw.TPCANMsg
	ts  raw.TPCANTimestamp
}

type rxFDItem struct {
	msg raw.TPCANMsgFD
	ts  raw.TPCANTimestampFD
}

func newFakeAdapter() *fakeAdapter {
	return &fakeAdapter{
		initializeReturn:   raw.PCAN_ERROR_OK,
		initializeFDReturn: raw.PCAN_ERROR_OK,
		uninitializeReturn: raw.PCAN_ERROR_OK,
		writeReturn:        raw.PCAN_ERROR_OK,
		writeFDReturn:      raw.PCAN_ERROR_OK,
		resetReturn:        raw.PCAN_ERROR_OK,
		statusReturn:       raw.PCAN_ERROR_OK,
		errorTextReturn:    "fake error text",
	}
}

func (f *fakeAdapter) Initialize(ch raw.TPCANHandle, br raw.TPCANBaudrate) raw.TPCANStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.initializeCalls++
	return f.initializeReturn
}

func (f *fakeAdapter) InitializeFD(ch raw.TPCANHandle, b string) raw.TPCANStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.initializeFDCalls++
	return f.initializeFDReturn
}

func (f *fakeAdapter) Uninitialize(ch raw.TPCANHandle) raw.TPCANStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.uninitializeCalls++
	return f.uninitializeReturn
}

func (f *fakeAdapter) Read(ch raw.TPCANHandle, m *raw.TPCANMsg, t *raw.TPCANTimestamp) raw.TPCANStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.readCalls++
	if len(f.rxQueue) == 0 {
		return raw.PCAN_ERROR_QRCVEMPTY
	}
	item := f.rxQueue[0]
	f.rxQueue = f.rxQueue[1:]
	*m = item.msg
	*t = item.ts
	return raw.PCAN_ERROR_OK
}

func (f *fakeAdapter) ReadFD(ch raw.TPCANHandle, m *raw.TPCANMsgFD, t *raw.TPCANTimestampFD) raw.TPCANStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.readFDCalls++
	if len(f.rxFDQueue) == 0 {
		return raw.PCAN_ERROR_QRCVEMPTY
	}
	item := f.rxFDQueue[0]
	f.rxFDQueue = f.rxFDQueue[1:]
	*m = item.msg
	*t = item.ts
	return raw.PCAN_ERROR_OK
}

func (f *fakeAdapter) Write(ch raw.TPCANHandle, m *raw.TPCANMsg) raw.TPCANStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writeCalls++
	cp := *m
	f.lastWrittenMsg = &cp
	if len(f.writeSequence) > 0 {
		s := f.writeSequence[0]
		f.writeSequence = f.writeSequence[1:]
		return s
	}
	return f.writeReturn
}

func (f *fakeAdapter) WriteFD(ch raw.TPCANHandle, m *raw.TPCANMsgFD) raw.TPCANStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writeFDCalls++
	cp := *m
	f.lastWrittenMsgFD = &cp
	if len(f.writeFDSequence) > 0 {
		s := f.writeFDSequence[0]
		f.writeFDSequence = f.writeFDSequence[1:]
		return s
	}
	return f.writeFDReturn
}

func (f *fakeAdapter) GetStatus(ch raw.TPCANHandle) raw.TPCANStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.statusReturn
}

func (f *fakeAdapter) GetErrorText(code raw.TPCANStatus, lang uint16) (string, raw.TPCANStatus) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.errorTextReturn, raw.PCAN_ERROR_OK
}

func (f *fakeAdapter) Reset(ch raw.TPCANHandle) raw.TPCANStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.resetCalls++
	return f.resetReturn
}

// push 把一帧 Classical 报文塞进接收队列，供 reader 取走。
func (f *fakeAdapter) push(msg raw.TPCANMsg) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rxQueue = append(f.rxQueue, rxItem{msg: msg})
}

// pushFD 把一帧 FD 报文塞进接收队列，供 reader 取走。
func (f *fakeAdapter) pushFD(msg raw.TPCANMsgFD) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rxFDQueue = append(f.rxFDQueue, rxFDItem{msg: msg})
}
