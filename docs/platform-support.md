# 平台支持

## 支持矩阵

| 平台 | 架构 | 真机 | 编译 | 单元测试 |
|---|---|---|---|---|
| Windows | amd64 | ✓ | ✓ | ✓ |
| Windows | 386 | ✓（驱动支持） | ✓ | ✓ |
| Linux | amd64/arm64 等 | ✓（SocketCAN） | ✓ | ✓（无硬件单测 + fake adapter） |
| macOS | any | ✗ | ✓ | ✓（fake adapter） |

> "真机"意味着 `Open()` 能成功返回可工作的 `Bus`。Windows 使用 PCANBasic 通道，
> Linux 使用 SocketCAN 网络接口；macOS 当前仅保留编译桩。

## DLL 加载策略

Windows 上库通过 `golang.org/x/sys/windows.LazyDLL` 加载 `PCANBasic.dll`。
加载顺序：

1. 环境变量 `PCANBASIC_DLL_PATH` 指向的路径（绝对或相对）
2. 默认值 `"PCANBasic.dll"`，按 Windows 标准 DLL 搜索路径解析（exe 同目录 → System32 → PATH）

```bash
# 自定义 DLL 位置
set PCANBASIC_DLL_PATH=C:\PEAK\PCAN-Basic API\x64\PCANBasic.dll
your-app.exe
```

加载失败时所有后续调用返回 `PCAN_ERROR_NODRIVER`（对应 `ErrNoDriver`）。
`sync.Once` 保证同一进程内只尝试加载一次。

## 32 位 vs 64 位

PCAN-USB 驱动同时提供 x86 和 x64 两份 DLL，路径分别是：

- `C:\Program Files\PEAK-System\PCAN-Basic API\x86\PCANBasic.dll`
- `C:\Program Files\PEAK-System\PCAN-Basic API\x64\PCANBasic.dll`

**Go 程序的位数必须和 DLL 一致**（`GOARCH=amd64` → x64 DLL）。否则加载时报 `0xC000007B`。

## 通道发现

`LookupChannels()` 返回当前平台可发现的通道列表：

- Windows：扫描内置 PCAN USB/PCI/LAN 通道，使用 `PCAN_CHANNEL_CONDITION` 判断是否可用
- Linux：枚举网络接口中 `can*` / `vcan*` / `slcan*` / `vxcan*` 名称的 SocketCAN 接口
- macOS：返回空列表，不报错

返回的 `ChannelInfo` 包含 `Channel`、`Name`、`Backend`、`Up`、`FD` 字段，可直接把 `Channel` 传给 `Open` / `OpenFD`。

`GetDeviceInfo(ch)` 返回指定通道的更详细信息：

- Windows：查询硬件名称、设备号、控制器号、通道特性等 PCAN 参数；缺少 DLL 时返回 `ErrDLLNotFound`
- Linux：返回 SocketCAN 网络接口名称、up 状态和 FD 能力标记
- macOS：返回 `ErrNotSupported`

## Linux SocketCAN 行为

Linux 上使用内核 SocketCAN，不需要 `PCANBasic.dll`：

- `Open(SocketCAN("can0"))` 打开 Classical CAN raw socket
- `OpenFD(SocketCAN("vcan0"), "")` 打开 CAN FD raw socket，并启用 `CAN_RAW_FD_FRAMES`
- `WithBitrate` 在 Linux 后端不生效；bitrate 由 `ip link set can0 type can bitrate ...` 配置
- `ModeEvent` 不支持；`ModeAuto` 会降级到 `ModePolling`
- `SetFilter` / `ResetFilter` 使用 SocketCAN `CAN_RAW_FILTER` 实现，支持标准帧和扩展帧 ID 范围过滤
- 其他 PCAN 参数 GetValue/SetValue 在 SocketCAN 后端暂不支持

示例：

```bash
sudo modprobe can
sudo modprobe can_raw
sudo modprobe vcan
sudo ip link add dev vcan0 type vcan
sudo ip link set vcan0 up
```

```go
bus, err := gocan.Open(gocan.SocketCAN("vcan0"))
```

CI 默认只跑不依赖硬件的单元测试；真实 SocketCAN 通信测试应在具备 `vcan0` 或真机 CAN 接口的 Linux 环境中手动执行：

```bash
GOCAN_SOCKETCAN_IFACE=vcan0 go test -tags=socketcan_integration ./...
```

## macOS 行为

macOS 上：

- `raw.EnsureLoaded()` 直接返回 `nil`（无事可做）
- `raw.Initialize / InitializeFD / Read / Write / ...` 全部返回 `PCAN_ERROR_ILLOPERATION`
- 顶层 `Open / OpenFD` 因此返回 `*Error{Code: PCAN_ERROR_ILLOPERATION}`，即 `errors.Is(err, ErrIllOperation)` 为 true
- 单元测试通过 fake adapter 注入：跑 `go test ./...` 能通过
