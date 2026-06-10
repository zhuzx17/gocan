//go:build linux

package gocan

import "github.com/Crush251/gocan/raw"

// CAN 错误帧位掩码（与 raw 包对应常量等价）。
const (
	CANErrTxTimeout = raw.CANErrTxTimeout
	CANErrLostArb   = raw.CANErrLostArb
	CANErrCrtl      = raw.CANErrCrtl
	CANErrProt      = raw.CANErrProt
	CANErrTrx       = raw.CANErrTrx
	CANErrAck       = raw.CANErrAck
	CANErrBusOff    = raw.CANErrBusOff
	CANErrBusError  = raw.CANErrBusError
	CANErrRestarted = raw.CANErrRestarted
	CANErrMaskAll   = raw.CANErrMaskAll
)

// WithLoopback 设置 CAN_RAW_LOOPBACK（默认内核为 true，即本地回环开启）。
// 关闭后，本进程发出的帧不会被同主机其他 socket 看到。
func WithLoopback(enabled bool) Option {
	return func(c *config) {
		v := enabled
		c.linux.loopback = &v
	}
}

// WithRecvOwnMsgs 设置 CAN_RAW_RECV_OWN_MSGS。开启后会收到本 socket 自己发出的帧
// （需配合 WithLoopback(true)；通常用于自发自收的回归测试）。
func WithRecvOwnMsgs(enabled bool) Option {
	return func(c *config) {
		v := enabled
		c.linux.recvOwnMsgs = &v
	}
}

// WithErrFilter 启用 CAN_RAW_ERR_FILTER，只接收 mask 中位标记的错误帧类型。
// 参见 raw/can_err_linux.go 中的 CANErr* 常量。
func WithErrFilter(mask uint32) Option {
	return func(c *config) {
		v := mask
		c.linux.errFilter = &v
	}
}
