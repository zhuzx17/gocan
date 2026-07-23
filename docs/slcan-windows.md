# Windows CANable 2.0 SLCAN-FD

gocan 支持 CANable 2.0 官方 `canable2-fw` 的扩展 Lawicel 协议。它通过 Windows
串口工作，不需要 `PCANBasic.dll`。

## 1. 固件和端口

设备必须刷入 [CANable 2.0 SLCAN-FD 固件](https://github.com/normaldotcom/canable2-fw)。
如果设备刷的是 candleLight/gs_usb 固件，它不会出现为 SLCAN 串口，不能使用本入口。

Windows 设备管理器会显示类似 `USB Serial Device (COM5)`。也可以从代码枚举：

```go
ports, err := gocan.LookupSLCANPorts()
if err != nil {
    log.Fatal(err)
}
for _, port := range ports {
    log.Printf("port=%s product=%s vid=%s pid=%s canable2=%v",
        port.Name, port.Product, port.VID, port.PID, port.CANable2)
}
```

官方固件的 USB VID:PID 是 `16D0:117E`。枚举函数会返回全部 USB 串口，兼容固件
即使使用其他 VID:PID 也可以手工选择。

## 2. 打开 CAN FD

```go
bus, err := gocan.OpenSLCANFD(
    "COM5",
    gocan.SLCANBitrate500K,
    gocan.SLCANDataBitrate2M,
    gocan.WithReceiveMode(gocan.ModePolling),
)
if err != nil {
    log.Fatal(err)
}
defer bus.Close()
```

`ModeAuto` 也可以使用，它在 SLCAN 后端等价于 Polling。`ModeEvent` 是 PCANBasic
专用机制，传给 SLCAN 会返回 `ErrNotSupported`。

只使用 Classical CAN 时：

```go
bus, err := gocan.OpenSLCAN("COM5", gocan.SLCANBitrate500K)
```

加入现有 `BusGroup` 做多通道合流时，使用 `AddSLCAN` 或 `AddSLCANFD`：

```go
bus, err := group.AddSLCANFD(
    "canable",
    "COM5",
    gocan.SLCANBitrate500K,
    gocan.SLCANDataBitrate2M,
)
```

## 3. 收发 CAN FD

```go
frame, err := gocan.NewFDFrame(
    0x456,
    []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
    true,  // BRS
    false, // standard 11-bit ID
)
if err != nil {
    log.Fatal(err)
}
if err := bus.Send(context.Background(), frame); err != nil {
    log.Fatal(err)
}

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
received, err := bus.ReadOne(ctx)
```

打开 FD Bus 后仍然可以发送和接收 Classical CAN 帧。FD 数据长度必须是
`0..8, 12, 16, 20, 24, 32, 48, 64` 之一。

## 4. 比特率预设

名义阶段支持：

| 常量 | 固件命令 | bit/s |
|---|---|---:|
| `SLCANBitrate10K` | `S0` | 10,000 |
| `SLCANBitrate20K` | `S1` | 20,000 |
| `SLCANBitrate50K` | `S2` | 50,000 |
| `SLCANBitrate100K` | `S3` | 100,000 |
| `SLCANBitrate125K` | `S4` | 125,000 |
| `SLCANBitrate250K` | `S5` | 250,000 |
| `SLCANBitrate500K` | `S6` | 500,000 |
| `SLCANBitrate750K` | `S7` | 750,000 |
| `SLCANBitrate1M` | `S8` | 1,000,000 |
| `SLCANBitrate83K3` | `S9` | 83,300 |

数据阶段只支持 `SLCANDataBitrate2M` (`Y2`) 和 `SLCANDataBitrate5M` (`Y5`)。

## 5. 打开时发送的命令

例如 500K / 2M 的默认 FD 配置会依次发送：

```text
C
S6
Y2
M0
A1
O
```

每条命令以 `\r` 结尾。固件要求先关闭通道再配置，配置完成后才能打开。
`WithSLCANSilent(true)` 发送 `M1`，`WithSLCANAutoRetransmit(false)` 发送 `A0`。

FD 报文严格使用官方扩展：`d/D` 表示无 BRS，`b/B` 表示有 BRS；DLC `9..F`
依次表示 12、16、20、24、32、48、64 字节。

官方固件上报接收到的 RTR 时可能在标准 `r/R + ID + DLC` 后附带 DLC 长度的占位
数据。gocan 同时接受标准形式和这个固件特有形式，对外都转换成正常的 `FlagRemote` 帧。

## 6. 固件限制

- 固件不对配置或发送命令返回 ACK/NACK。`OpenSLCANFD` 成功表示串口已打开且配置
  字节已写入，不代表设备验证了每条命令。
- 固件没有 SLCAN 硬件过滤命令，因此 `SetFilter` / `ResetFilter` 返回
  `ErrNotSupported`。
- 固件的错误寄存器响应不适合和帧流并发解析，因此 `Status` 返回
  `ErrNotSupported`；`Reset` 使用 `C`、`O` 重新启用通道。
- SLCAN-FD 字符格式没有 ESI 字段。发送带 `FlagESI` 的帧返回
  `ErrSLCANESINotSupported`。
- 高总线负载下 ASCII 十六进制编码和 USB CDC 带宽可能成为瓶颈；此时优先使用
  CANable candleLight/gs_usb 固件配合原生驱动，而不是 SLCAN。
