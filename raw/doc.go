// Package raw 提供 PCANBasic.dll C API 的零开销 Go 绑定。
//
// 本包按 PCAN-Basic 官方头文件 PCANBasic.h 进行 1:1 映射，
// 适合需要 PCAN 特殊功能（Trace、Flash、设备信息查询、任意 Parameter 读写等）
// 或需要在 gocan 顶层包之外自行做更高层封装的用户。
//
// 大多数应用建议直接使用顶层 github.com/zhuzx17/gocan 包。
//
// # 平台
//
// 仅 Windows 真实工作；其他平台所有函数返回 PCAN_ERROR_ILLOPERATION，
// 由高层映射为 ErrNotSupported，便于在 Linux/macOS 上做 lint / vet / 单元测试。
//
// # DLL 加载
//
// 默认从进程当前目录及 Windows 标准 DLL 搜索路径加载 "PCANBasic.dll"。
// 可通过设置环境变量 PCANBASIC_DLL_PATH 指向具体 DLL 文件覆盖。
//
// 注意：Go 程序 GOARCH 必须与 DLL 位数一致（amd64 → 64 位 DLL；386 → 32 位 DLL）。
//
// # 稳定性
//
// 在 v1.0.0 之前，本包签名跟随 PCANBasic 官方头文件演进可能发生小幅调整；
// 顶层 gocan 包提供稳定的高层抽象。
package raw
