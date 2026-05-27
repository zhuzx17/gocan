# gocan

> Go CAN / CAN FD 多后端库：Windows 走 PEAK-System PCANBasic，Linux 走 SocketCAN。

[![Go Reference](https://pkg.go.dev/badge/github.com/Crush251/gocan.svg)](https://pkg.go.dev/github.com/Crush251/gocan)
[![CI](https://github.com/Crush251/gocan/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Crush251/gocan/actions/workflows/ci.yml)
[![CodeQL](https://github.com/Crush251/gocan/actions/workflows/codeql.yml/badge.svg?branch=main)](https://github.com/Crush251/gocan/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Crush251/gocan)](https://goreportcard.com/report/github.com/Crush251/gocan)
[![codecov](https://codecov.io/gh/Crush251/gocan/branch/main/graph/badge.svg)](https://codecov.io/gh/Crush251/gocan)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`gocan` 提供统一的 Go `Bus` API，让程序能在 Windows PCANBasic 和 Linux SocketCAN
之间复用同一套 CAN / CAN FD 收发逻辑。Windows 后端用纯 Go（`syscall` 调用，无 CGO）
封装 `PCANBasic.dll`；Linux 后端直接使用内核 SocketCAN。

> ⚠️ **当前为预发布阶段**（`v0.x`），API 仍可能调整。本仓库目前处于**初始化阶段**，
> 完整设计见
> [docs/superpowers/specs/2026-05-22-gocan-design.md](docs/superpowers/specs/2026-05-22-gocan-design.md)。

---

## 为什么做这个

- **Windows** 端 PEAK 官方仅提供 C/C++、C#、Java、Python 等绑定，**没有 Go 封装**
- **Linux** 端事实标准是内核 SocketCAN，接口通常是 `can0` / `vcan0`
- 现有的 Python 桥接（`can-bridge-win`）需要打包 `PyInstaller`、引入 `python-can` 依赖，
  在嵌入到机器人控制等纯 Go 项目时既笨重又难以追踪问题

`gocan` 的目标是提供稳定的跨平台 CAN 抽象层：Windows 走 PCANBasic，Linux 走 SocketCAN，
上层业务尽量只依赖统一的 `Bus` / `Frame` API。

---

## 特性（v0.1 计划）

- ✅ Classical CAN 与 CAN FD 双标准支持
- ✅ 高层 `Bus` API：`Send` / `SendMany` / `Receive` / `ReadOne` / `TryRead` / `Status` / `Reset` / `SetFilter`
- ✅ 三种接收模式：`ModeAuto` / `ModePolling` / `ModeEvent`（Windows Event 驱动）
- ✅ 子包 `raw`：与 PCANBasic C API 1:1 对应的低层绑定
- ✅ 错误处理：位掩码语义 + `errors.Is` 哨兵
- ✅ Linux SocketCAN 后端：`Open(SocketCAN("can0"))` / `OpenFD(SocketCAN("vcan0"), "")` / `SetFilter`
- ✅ 通道发现：`LookupChannels()` 枚举当前平台可发现的 PCAN / SocketCAN 通道
- ✅ 完整的中文文档与 10 个示例

详细范围见 [设计文档](docs/superpowers/specs/2026-05-22-gocan-design.md)。

---

## 快速开始

```go
package main

import (
    "context"
    "log"

    "github.com/Crush251/gocan"
)

func main() {
    bus, err := gocan.Open(
        gocan.USBBus1,
        gocan.WithBitrate(gocan.Baud1M),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer bus.Close()

    frame, _ := gocan.NewFrame(0x123, []byte{0x01, 0x02, 0x03, 0x04})
    if err := bus.Send(context.Background(), frame); err != nil {
        log.Fatal(err)
    }
}
```

---

## 通道发现

```go
channels, err := gocan.LookupChannels()
if err != nil {
    log.Fatal(err)
}
for _, ch := range channels {
    log.Printf("%s %s up=%v fd=%v", ch.Backend, ch.Name, ch.Up, ch.FD)
}
```

## Linux SocketCAN

```go
bus, err := gocan.Open(gocan.SocketCAN("can0"))
if err != nil {
    log.Fatal(err)
}
defer bus.Close()
```

CAN FD：

```go
bus, err := gocan.OpenFD(gocan.SocketCAN("vcan0"), "")
```

Linux 上 bitrate 由系统配置，不由 `WithBitrate` 设置，例如：

```bash
sudo ip link set can0 type can bitrate 500000
sudo ip link set can0 up
```

## 系统要求

- Windows：Go 1.22+、已安装 PEAK PCAN 驱动、`PCANBasic.dll` 与 Go 程序架构匹配
- Linux：Go 1.22+、内核启用 SocketCAN、已创建并启用 `can0` / `vcan0` 等网络接口
- macOS：当前仅支持编译和纯逻辑测试，不支持真机通信

---

## 路线图

| 版本 | 主要内容 |
|---|---|
| v0.1.0 | Classical + FD 收发、Bus 完整高层、raw 子包基础 API、文档与示例（Windows 后端） |
| v0.2.0 | 最小 Linux SocketCAN 后端：`can0` / `vcan0` 的 Open、Send、Receive、Close；LookUpChannel、设备信息查询、Trace 后置 |
| v1.0.0 | API 冻结，进入严格兼容承诺 |

---

## 许可证

[MIT](LICENSE)
