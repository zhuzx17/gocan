# Windows 快速启动

5 分钟在 Windows 上跑通 gocan 第一帧。涵盖 PCAN 驱动安装、Classical CAN、CAN FD、接收模式与多通道。

## 1. 安装 PEAK PCAN 驱动

下载并安装 PEAK-System 官方驱动：<https://www.peak-system.com/quick/DrvSetup>

安装后在 Windows 设备管理器里应能看到 `PCAN-USB` / `PCAN-USB FD` 等节点。

## 2. 放置 `PCANBasic.dll`

PCAN 驱动安装包自带 `PCANBasic.dll`，路径通常为：

- 32 位：`C:\Program Files\PEAK-System\PCAN-Basic API\x86\PCANBasic.dll`
- 64 位：`C:\Program Files\PEAK-System\PCAN-Basic API\x64\PCANBasic.dll`

**Go 程序的位数必须和 DLL 一致**：`GOARCH=amd64` → x64 dll；`GOARCH=386` → x86 dll。否则 `Open` 会返回 `ErrNoDriver`（错误码 `0xC000007B`）。

加载顺序：

1. 与可执行文件同目录
2. `PATH` 中
3. 环境变量 `PCANBASIC_DLL_PATH` 指向的绝对路径（最高优先级）

最简单做法：把 `PCANBasic.dll` 复制到可执行文件旁边。

## 3. 安装库

```cmd
go get github.com/zhuzxdev/gocan@latest
```

## 4. Classical CAN 收发

最小可运行示例（需要插着真 PCAN-USB，对端有发送或者你两路自连）：

```go
package main

import (
	"context"
	"log"

	"github.com/zhuzxdev/gocan"
)

func main() {
	bus, err := gocan.Open(
		gocan.USBBus1,
		gocan.WithBitrate(gocan.Baud500K),
	)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer bus.Close()

	frame, _ := gocan.NewFrame(0x123, []byte{0x01, 0x02, 0x03, 0x04})
	if err := bus.Send(context.Background(), frame); err != nil {
		log.Fatalf("send: %v", err)
	}
	log.Printf("sent: id=0x%X data=%X", frame.ID, frame.Data)
}
```

或运行仓库自带示例：

```cmd
go run ./examples/01_send_classical -channel=USBBus1
```

预期输出：

```
sent: id=0x123 data=01020304
```

可在 PEAK 自带的 PCAN-View 程序里观察发出的帧。

## 5. CAN FD 收发

Windows PCAN 的 FD 比特率通过 `fdBitrate` 字符串配置（与 Linux SocketCAN 不同），传给 `OpenFD`。常用 80 MHz 时钟下 500K 名义 / 2M 数据的字符串：

```
f_clock=80000000,nom_brp=10,nom_tseg1=12,nom_tseg2=3,nom_sjw=1,data_brp=4,data_tseg1=7,data_tseg2=2,data_sjw=1
```

完整字段表见 [`docs/can-fd.md`](can-fd.md)。

```go
package main

import (
	"context"
	"log"

	"github.com/zhuzxdev/gocan"
)

func main() {
	fdBitrate := "f_clock=80000000,nom_brp=10,nom_tseg1=12,nom_tseg2=3,nom_sjw=1,data_brp=4,data_tseg1=7,data_tseg2=2,data_sjw=1"

	bus, err := gocan.OpenFD(gocan.USBBus1, fdBitrate)
	if err != nil {
		log.Fatalf("open fd: %v", err)
	}
	defer bus.Close()

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12} // 12 字节是 FD 合法离散长度
	frame, _ := gocan.NewFDFrame(0x456, data, true /*brs*/, false /*extended*/)
	if err := bus.Send(context.Background(), frame); err != nil {
		log.Fatalf("send fd: %v", err)
	}
	log.Printf("sent fd: id=0x%X dlc=%d brs=%v", frame.ID, len(frame.Data), frame.Has(gocan.FlagBRS))
}
```

## 6. 接收模式

`WithReceiveMode` 控制 reader goroutine 等待数据的策略：

| 模式 | 含义 | CPU | 延迟 |
|---|---|---|---|
| `ModeAuto`（默认） | 尝试 Event；硬件不支持时降级 Polling | 低（Event 时） | 低（Event 时） |
| `ModePolling` | 用 `WithPollInterval` 设置的间隔轮询 | 取决于间隔 | ≥ 间隔 |
| `ModeEvent` | Windows 事件驱动，强制阻塞等待 | 最低 | 最低 |

实际生产代码通常**保持默认 `ModeAuto`** 即可。`WithPollInterval(d)` 仅在 Polling 模式（含 Auto 降级）时有意义，默认 1ms。

```go
bus, err := gocan.Open(
    gocan.USBBus1,
    gocan.WithReceiveMode(gocan.ModeEvent), // 强制 Event；不支持时 Open 失败
)
```

## 7. 多通道：两路 PCAN-USB

同时连两台 PCAN-USB，使用 `BusGroup` 合流：

```go
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"

	"github.com/zhuzxdev/gocan"
)

func main() {
	g := gocan.NewBusGroup(0)
	defer func() {
		if err := g.Close(); err != nil {
			var gce *gocan.GroupCloseError
			if errors.As(err, &gce) {
				for name, e := range gce.Causes {
					log.Printf("close %s: %v", name, e)
				}
			}
		}
	}()

	if _, err := g.Add("primary", gocan.USBBus1, gocan.WithBitrate(gocan.Baud500K)); err != nil {
		log.Fatalf("add primary: %v", err)
	}
	if _, err := g.Add("secondary", gocan.USBBus2, gocan.WithBitrate(gocan.Baud500K)); err != nil {
		log.Fatalf("add secondary: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case sf := <-g.Receive():
			log.Printf("[%s] id=0x%X data=%X", sf.Source, sf.Frame.ID, sf.Frame.Data)
		}
	}
}
```

不用 `BusGroup` 的旧式手写 `sync.WaitGroup` 写法见 `examples/06_multi_channel/`。

## 8. 下一步

- Event 模式底层细节：`examples/03_receive_event/`
- 接 slog：`examples/09_with_logger/`
- Option 速查 + 平台对照：[`docs/options.md`](options.md)
- FD 比特率字段：[`docs/can-fd.md`](can-fd.md)
- 故障排查：[`docs/troubleshooting.md`](troubleshooting.md)
