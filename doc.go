// Package gocan 是 Go CAN/CAN FD 多后端库。
//
// 顶层包提供 idiomatic Go 高层 API（Bus / Frame / Option / Error），
// 大多数应用直接使用本包即可。Windows 支持 PCANBasic 和 CANable 2.0 SLCAN-FD，
// Linux 支持 SocketCAN；串口 SLCAN 后端也可在 Linux/macOS 使用。
// 需要 PCAN 特殊功能或希望进一步定制时，可使用子包 github.com/zhuzx17/gocan/raw。
//
// # 快速开始
//
//	bus, err := gocan.Open(gocan.USBBus1, gocan.WithBitrate(gocan.Baud1M))
//	if err != nil { log.Fatal(err) }
//	defer bus.Close()
//
//	f, _ := gocan.NewFrame(0x123, []byte{1, 2, 3})
//	_ = bus.Send(context.Background(), f)
//
// 详见 README 与 docs/quickstart-linux.md / docs/quickstart-windows.md。
//
// # 平台
//
// Windows 真机可使用 PCANBasic 或 SLCAN；Linux 可使用 SocketCAN 或 SLCAN；
// macOS 可使用串口 SLCAN，PCAN API 在非 Windows 平台是编译桩。
//
// # 并发
//
// Bus 内部使用单 reader goroutine 独占底层 Read，
// 调用方可在多个 goroutine 中并发 Send / Status / Reset / Receive；
// Close 是幂等的。
package gocan
