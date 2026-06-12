package gocan

import "github.com/zhuzxdev/gocan/raw"

// SocketCAN 返回 Linux SocketCAN 网络接口对应的通道句柄。
//
// 典型用法：Open(SocketCAN("can0")) 或 OpenFD(SocketCAN("vcan0"), "")。
// 空接口名会返回 PCAN_NONEBUS，Open/OpenFD 随后会返回参数错误。
func SocketCAN(iface string) Channel {
	return raw.SocketCANHandle(iface)
}
