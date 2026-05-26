# Changelog

本项目的所有显著变更将记录在此文件。

文件格式遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循 [语义化版本 2.0.0](https://semver.org/lang/zh-CN/)。

## [0.1.0] - 2026-05-22

首个公开版本：Windows 上的 PEAK PCANBasic.dll Go 封装。

### Added

**底层绑定（`raw` 子包）**

- Windows 通过 `golang.org/x/sys/windows` 直接 syscall 加载 `PCANBasic.dll`，零 cgo
- 非 Windows 平台 stub 实现：编译通过、API 返回 `PCAN_ERROR_ILLOPERATION`
- 完整类型映射：`TPCANHandle` / `TPCANStatus` / `TPCANBaudrate` / `TPCANMessageType` / `TPCANParameter` / `TPCANMsg` / `TPCANMsgFD` / `TPCANTimestamp(FD)`
- 13 个 PCAN API 函数绑定：`Initialize` / `InitializeFD` / `Uninitialize` / `Read` / `ReadFD` / `Write` / `WriteFD` / `GetStatus` / `Reset` / `FilterMessages` / `GetValue` / `SetValue` / `GetErrorText`
- 完整常量：USB/PCI/LAN 通道、波特率、错误码、消息类型、参数键、过滤器模式
- DLL 路径可通过环境变量 `PCANBASIC_DLL_PATH` 覆盖

**顶层 idiomatic API**

- `Bus`：表示一个已初始化的通道；`Open` / `OpenFD` 构造，`Close` 释放（幂等）
- `Frame`：统一 Classical/FD 表示，附带 `FrameFlags`（Extended/Remote/FD/BRS/ESI）、`TimestampMicros`、`ReceivedAt`
- 构造器：`NewFrame` / `NewExtendedFrame` / `NewRemoteFrame` / `NewFDFrame`，含 ID 范围、数据长度、FD 离散长度校验
- 发送：`Send(ctx, frame)` / `SendMany(ctx, frames)`；批量失败返回 `*SendManyError`（含失败 index、深拷贝帧、内部错误）
- 接收：`Receive() <-chan Frame` 流式、`ReadOne(ctx)` 阻塞取一帧、`TryRead()` 非阻塞取一帧、`Errors() <-chan error` 异步错误
- 状态与恢复：`Status()`、`Reset()`、`StatusHas(s, bit)`
- 过滤器：`SetFilter(idMin, idMax, mode)` / `ResetFilter()`
- 选项：`WithBitrate` / `WithReceiveMode` / `WithPollInterval` / `WithRxBufferSize` / `WithErrBufferSize` / `WithLogger`

**接收模式**

- `ModePolling`：周期轮询底层 Read，跨平台
- `ModeEvent`：Windows 上 `CreateEvent` + `WaitForMultipleObjects`，低延迟低 CPU
- `ModeAuto`（默认）：尝试 Event，失败降级 Polling 并通过 Logger 提示

**错误模型**

- `*Error{Code, Op, Msg}` 保留原始 PCAN 位掩码；`Is(target)` 支持位匹配，允许同一错误同时匹配多个哨兵
- 完整哨兵错误：`ErrBusClosed` / `ErrFDNotSupportedOnBus` / `ErrIDOutOfRange` / `ErrDataTooLong` / `ErrInvalidFDLength` / `ErrRemoteOnFD` / `ErrNotSupported` / `ErrDLLNotFound` / `ErrQueueEmpty` / `ErrQueueOverrun` / `ErrQueueXmtFull` / `ErrBusLight` / `ErrBusHeavy` / `ErrBusPassive` / `ErrBusOff` / `ErrNotInitialized` / `ErrIllHandle` / `ErrIllParamType` / `ErrIllParamValue` / `ErrIllOperation` / `ErrNoDriver` / `ErrUnknown`
- `SendManyError` 实现 `Unwrap`，`errors.Is/As` 自然穿透

**并发与生命周期**

- 单 reader goroutine 独占底层 Read，多 goroutine 可安全并发 Send/Status/Reset/Receive
- Close 流程：标记 closed → 关闭 closing → 唤醒 reader → 等 rxCh 关闭 → 释放 event 句柄 → Uninitialize → 关闭 errCh
- `sync.Once` 保证 Close 幂等

**示例（`examples/`）**

10 个独立示例，每个 ≤ 100 行带中文头注释：发送 Classical / Polling 接收 / Event 接收 / 发送 FD / 接收 FD / 多通道 / 过滤器 / Status+Reset / slog 适配 / 直接调 raw 包

**文档（`docs/`）**

7 篇：`quickstart` / `architecture` / `can-fd` / `error-handling` / `platform-support` / `hardware-test-setup` / `troubleshooting`

**测试与 CI**

- fake adapter + 错误注入 adapter 覆盖所有公开 API
- 测试覆盖率 80%+，`-race` 通过
- GitHub Actions：Linux runner 跑 vet + golangci-lint + race test + 跨平台编译（windows/amd64+386）；Windows runner 跑 vet + 普通 test

[0.1.0]: https://github.com/Crush251/can_go/releases/tag/v0.1.0
