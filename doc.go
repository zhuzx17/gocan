// Package gocan 是 PEAK-System PCANBasic.dll 的 Go 封装库（Windows 专用）。
//
// 顶层包提供 idiomatic Go 高层 API（Bus / Frame / Option / Error），
// 大多数应用直接使用本包即可。需要 PCAN 特殊功能或希望进一步定制时，
// 可使用子包 github.com/Crush251/gocan/raw。
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
// 详见 README 与 docs/quickstart.md。
//
// # 平台
//
// v0.1 真机仅支持 Windows；Linux/macOS 上代码可编译、单元测试可运行，
// 但 Open/OpenFD 会返回 ErrIllOperation（底层 PCAN_ERROR_ILLOPERATION）。
//
// # 并发
//
// Bus 内部使用单 reader goroutine 独占底层 Read，
// 调用方可在多个 goroutine 中并发 Send / Status / Reset / Receive；
// Close 是幂等的。
package gocan
