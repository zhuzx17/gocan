# Option 参数总览

所有 `WithXxx` Option 函数 + 平台标注。打 ✓ 表示当前平台真实生效；标"被忽略"或"✗"表示该 Option 在该平台**没有效果**或**不存在**。

Linux 专属 Option 用 `//go:build linux` 隔离 —— 在 Windows / Darwin 编译期就找不到符号。

## 1. 速查表

### 1.1 跨平台 Option（`gocan` 包，所有平台可用）

| Option | 默认值 | 作用 | Linux | Windows |
|---|---|---|:-:|:-:|
| `WithBitrate(Bitrate)` | `Baud1M` | Classical CAN 波特率 | 被忽略（由 `ip link` 设） | ✓ |
| `WithReceiveMode(ReceiveMode)` | `ModeAuto` | reader goroutine 等待策略 | 仅 `Polling` 生效 | `Auto`/`Polling`/`Event` |
| `WithPollInterval(time.Duration)` | `1ms` | Polling 模式轮询间隔 | ✓ | ✓（Polling 时） |
| `WithRxBufferSize(int)` | `1024` | 接收 channel 容量 | ✓ | ✓ |
| `WithErrBufferSize(int)` | `16` | 错误 channel 容量 | ✓ | ✓ |
| `WithLogger(Logger)` | `noopLogger{}` | 注入日志接口 | ✓ | ✓ |

### 1.2 Linux 专属 Option（`//go:build linux`）

| Option | 默认值 | 作用 | 内核要求 |
|---|---|---|---|
| `WithLoopback(bool)` | 内核默认 `true` | `CAN_RAW_LOOPBACK` | 3.6+ |
| `WithRecvOwnMsgs(bool)` | 内核默认 `false` | `CAN_RAW_RECV_OWN_MSGS` | 3.6+ |
| `WithErrFilter(uint32)` | 不设置 | `CAN_RAW_ERR_FILTER` | 3.6+ |
| `WithJoinFilters(bool)` | 不设置（OR） | `CAN_RAW_JOIN_FILTERS` | **4.1+** |
| `WithRecvTimestamp(RxTimestamp)` | `RxTimestampNone` | `SO_TIMESTAMP*` | 取决于 mode |
| `WithSocketBuffers(rcv, snd int)` | 不设置 | `SO_RCVBUF` / `SO_SNDBUF` | always |
| `WithRWTimeout(read, write Duration)` | 不设置 | `SO_RCVTIMEO` / `SO_SNDTIMEO` | always |

### 1.3 Windows 专属 Option

> 暂无。Windows PCAN 后端常见参数（`PCAN_LISTEN_ONLY`、`PCAN_ALLOW_*_FRAMES`、`PCAN_BUSOFF_AUTORESET`）将在后续 PR 补全。

## 2. 跨平台 Option 详解

### 2.1 `WithBitrate(b Bitrate) Option`

- **类型**：`gocan.Bitrate`（取值 `gocan.Baud1M` / `Baud500K` / `Baud250K` / `Baud125K` / `Baud100K` / `Baud50K` / `Baud20K` / `Baud10K` / `Baud5K`）
- **Windows**：在 `Open` 时通过 `CAN_Initialize` 应用到 PCAN 驱动
- **Linux**：被忽略 —— SocketCAN 比特率由内核 netlink 配置（`ip link set canX type can bitrate ...`），应用进程不参与
- **OpenFD 时**：在两个平台都被忽略；FD 比特率走 `OpenFD` 的 `fdBitrate` 字符串（Win）或 `ip link` 的 `dbitrate` 参数（Linux）

```go
bus, err := gocan.Open(gocan.USBBus1, gocan.WithBitrate(gocan.Baud500K))
```

### 2.2 `WithReceiveMode(m ReceiveMode) Option`

| `ReceiveMode` | 含义 | 平台行为 |
|---|---|---|
| `ModeAuto`（默认） | 尝试 Event；不支持降级 Polling 并通过 Logger 提示 | Win 有 Event；Linux 直接降级到 Polling |
| `ModePolling` | 周期轮询 | 跨平台 |
| `ModeEvent` | Windows Event Handle 阻塞 | Linux 调用直接报 setup 失败，Open 返回错误 |

```go
gocan.WithReceiveMode(gocan.ModePolling)
```

### 2.3 `WithPollInterval(d time.Duration) Option`

- **默认**：`1 * time.Millisecond`
- **生效条件**：`ModePolling`（含 `ModeAuto` 降级到 Polling 后）
- **非正值**：被忽略（保留默认）
- 调小 → 延迟低、CPU 占用高；调大 → 反之

```go
gocan.WithPollInterval(500 * time.Microsecond)
```

### 2.4 `WithRxBufferSize(n int) Option`

- **默认**：`1024`
- **作用**：`bus.Receive()` 返回的 channel 容量
- **满了**：reader 会丢帧并继续；高吞吐场景调大可减少丢失
- **非正值**：被忽略

```go
gocan.WithRxBufferSize(4096)
```

### 2.5 `WithErrBufferSize(n int) Option`

- **默认**：`16`
- **作用**：`bus.Errors()` 异步错误 channel 容量
- **满了**：直接丢弃（错误是"提示性"通道，不保证完整性）
- **非正值**：被忽略

```go
gocan.WithErrBufferSize(64)
```

### 2.6 `WithLogger(l Logger) Option`

- **默认**：`noopLogger{}`，不打印任何东西
- **作用**：注入 Logger 接口实现，库内部把 reader 失败、Event 降级、Close 异常等异步信息写到这里
- **传入 nil**：被忽略
- 接 slog 的最小桥接见 `examples/09_with_logger/`

```go
gocan.WithLogger(myLogger)
```

## 3. Linux 专属 Option 详解

每项链到 [`docs/socketcan-options.md`](socketcan-options.md) 对应章节做深度阅读。

### 3.1 `WithLoopback(enabled bool) Option`

控制 `CAN_RAW_LOOPBACK`。默认（不调）= 内核默认 `true`，本机其他 socket 能看到本进程发出的帧。**关闭后本机无法做自发自收回归**。

详见 [`socketcan-options.md` §3.1](socketcan-options.md#31-withloopback)。

### 3.2 `WithRecvOwnMsgs(enabled bool) Option`

控制 `CAN_RAW_RECV_OWN_MSGS`。开启后本 socket 能收到自己发出的帧。需配合 `WithLoopback(true)`，配合 vcan 即可单进程做收发回归测试。

详见 [`socketcan-options.md` §3.2](socketcan-options.md#32-withrecvownmsgs)。

### 3.3 `WithErrFilter(mask uint32) Option`

把内核错误帧（`CAN_ERR_FRAME`）转成业务可读流。`mask` 是 `CANErrBusOff | CANErrTxTimeout | ...` 的 OR。完整位掩码：`CANErrTxTimeout` / `CANErrLostArb` / `CANErrCrtl` / `CANErrProt` / `CANErrTrx` / `CANErrAck` / `CANErrBusOff` / `CANErrBusError` / `CANErrRestarted` / `CANErrMaskAll`。

```go
gocan.WithErrFilter(gocan.CANErrBusOff | gocan.CANErrTxTimeout)
```

详见 [`socketcan-options.md` §3.3](socketcan-options.md#33-witherrfilter)。

### 3.4 `WithJoinFilters(and bool) Option`

控制 `CAN_RAW_JOIN_FILTERS`：多个 `SetFilter` 范围之间是 AND 还是 OR。默认 OR；`true` 改为 AND（全部命中才接收）。**需要内核 ≥ 4.1**——更老内核 setsockopt 返回 `ENOPROTOOPT`，会被映射为 `*Error{Op:"setsockopt(CAN_RAW_JOIN_FILTERS)"}`。

详见 [`socketcan-options.md` §3.4](socketcan-options.md#34-withjoinfilters)。

### 3.5 `WithRecvTimestamp(mode RxTimestamp) Option`

启用内核时间戳，结果写入 `Frame.TimestampMicros`：

| `RxTimestamp` 值 | 含义 |
|---|---|
| `RxTimestampNone` | 不启用，沿用 `time.Now()` 合成时间戳（默认）|
| `RxTimestampSecond` | `SO_TIMESTAMP`：μs 级 |
| `RxTimestampNano` | `SO_TIMESTAMPNS`：ns 级 |
| `RxTimestampHardware` | `SO_TIMESTAMPING + RX_HARDWARE`：硬件时间戳；不支持时降级到 NS |

启用后 SocketCAN 后端从 `read(2)` 切到 `recvmsg(2)+cmsg`，每帧多一次小拷贝；不启用时性能与之前一致。

详见 [`socketcan-options.md` §3.5](socketcan-options.md#35-withrecvtimestamp)。

### 3.6 `WithSocketBuffers(rcvBytes, sndBytes int) Option`

`SO_RCVBUF` / `SO_SNDBUF`，单位字节。任一非正值则跳过该方向。**实际上限受 `net.core.rmem_max` / `wmem_max`** 限制；调大可能需要 `sysctl -w net.core.rmem_max=8388608`。

```go
gocan.WithSocketBuffers(2*1024*1024, 1*1024*1024)
```

详见 [`socketcan-options.md` §3.6](socketcan-options.md#36-withsocketbuffers)。

### 3.7 `WithRWTimeout(read, write time.Duration) Option`

`SO_RCVTIMEO` / `SO_SNDTIMEO`。零值（默认）= 不设置 = 阻塞读写。当前 reader goroutine 用 polling 循环 + 短读，read timeout 主要影响异常路径下的退出延迟。

```go
gocan.WithRWTimeout(500*time.Millisecond, 0)
```

详见 [`socketcan-options.md` §3.7](socketcan-options.md#37-withrwtimeout)。

## 4. 运行期方法

仅 Linux 真实生效，其他平台返回 `ErrNotSupported`：

```go
func (b *Bus) SetErrFilter(mask uint32) error
func (b *Bus) SetJoinFilters(and bool) error
```

其他参数（`Loopback` / `Buffer` / `Timeout` / `Timestamp` 模式）变更涉及 reader goroutine 协调，超出当前实现范围；需要时整体 `Close` 后重新 `Open`。

## 5. 错误处理

任意 Option / 运行期方法触发的 setsockopt / Initialize 失败 → `*gocan.Error{Op: "...", Code: ..., Msg: ...}`。

常见 `errors.Is` 命中：

| 错误 | 触发 |
|---|---|
| `ErrIllParamValue` | 参数值不合法（如 `WithBitrate` 传不识别的常量；Linux 上接口不存在） |
| `ErrNoDriver` | Windows 上 `PCANBasic.dll` 未加载 / 架构不匹配 |
| `ErrBusClosed` | 在已 `Close` 的 `*Bus` 上调用运行期方法 |
| `ErrNotSupported` | 在非 Linux 平台调用 `SetErrFilter` / `SetJoinFilters` |

## 6. 与现有文档的关系

- [`docs/socketcan-options.md`](socketcan-options.md)：Linux 专属 Option 深度阅读
- [`docs/can-fd.md`](can-fd.md)：`fdBitrate` 字符串字段表（不在 Option 范畴）
- [`docs/quickstart-linux.md`](quickstart-linux.md) / [`docs/quickstart-windows.md`](quickstart-windows.md)：5 分钟跑通第一帧
- [`docs/troubleshooting.md`](troubleshooting.md)：故障排查与平台对照
