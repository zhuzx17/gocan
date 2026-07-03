# Linux 快速启动

5 分钟在 Linux 上跑通 gocan 第一帧。涵盖 SocketCAN 接口准备、Classical CAN、CAN FD、常用 Option 与多通道。

## 1. 准备 SocketCAN 接口

### 1.1 真实 CAN 接口

如果你的硬件是 USB-CAN 适配器（如 PEAK PCAN-USB FD、Innomaker、Kvaser 等内核已支持的设备）：

```bash
sudo ip link set can0 type can bitrate 500000
sudo ip link set can0 up
ip link show can0   # 期望 state UP
```

### 1.2 虚拟 vcan（无硬件，开发 / 测试）

```bash
# 仓库自带脚本（推荐）
sudo ./scripts/setup-vcan.sh up vcan0 vcan1

# 或者手动
sudo modprobe vcan
sudo ip link add vcan0 type vcan
sudo ip link set vcan0 up
```

## 2. 安装库

```bash
go get github.com/zhuzx17/gocan@latest
```

## 3. Classical CAN 收发

最小可运行示例（自发自收，需要 `vcan0`）：

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/zhuzx17/gocan"
)

func main() {
	bus, err := gocan.Open(
		gocan.SocketCAN("vcan0"),
		gocan.WithLoopback(true),
		gocan.WithRecvOwnMsgs(true), // 让本 socket 收到自己发的帧
	)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer bus.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	frame, _ := gocan.NewFrame(0x123, []byte{0x01, 0x02, 0x03})
	if err := bus.Send(ctx, frame); err != nil {
		log.Fatalf("send: %v", err)
	}

	got, err := bus.ReadOne(ctx)
	if err != nil {
		log.Fatalf("read: %v", err)
	}
	log.Printf("rx id=0x%X data=%X", got.ID, got.Data)
}
```

预期输出：

```
rx id=0x123 data=010203
```

也可以在另一终端用 `candump` 验证：

```bash
candump vcan0   # 在另一个 terminal 运行，观察发出的帧
```

## 4. CAN FD 收发

Linux SocketCAN 的 FD 比特率由 `ip link` 配置，不在 `OpenFD` 的 `fdBitrate` 参数里传。

```bash
# 真实接口启用 FD
sudo ip link set can0 down
sudo ip link set can0 type can bitrate 500000 dbitrate 2000000 fd on
sudo ip link set can0 up

# vcan 接口直接支持 FD，无需特殊配置
```

代码里 `OpenFD` 的 `fdBitrate` 传空字符串即可：

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/zhuzx17/gocan"
)

func main() {
	bus, err := gocan.OpenFD(
		gocan.SocketCAN("vcan0"), "",
		gocan.WithLoopback(true),
		gocan.WithRecvOwnMsgs(true),
	)
	if err != nil {
		log.Fatalf("open fd: %v", err)
	}
	defer bus.Close()

	// 12 字节 + BRS（FD 离散长度之一）
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	frame, err := gocan.NewFDFrame(0x456, data, true /*brs*/, false /*extended*/)
	if err != nil {
		log.Fatalf("frame: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := bus.Send(ctx, frame); err != nil {
		log.Fatalf("send fd: %v", err)
	}

	got, _ := bus.ReadOne(ctx)
	log.Printf("rx id=0x%X dlc=%d data=%X brs=%v", got.ID, len(got.Data), got.Data, got.Has(gocan.FlagBRS))
}
```

## 5. 常用 Linux 专属 Option

### 5.1 自发自收回归（loopback + recv-own-msgs）

```go
gocan.WithLoopback(true)     // 默认就是 true，显式声明便于排查
gocan.WithRecvOwnMsgs(true)  // 关键：让自己看到自己发的帧
```

### 5.2 订阅总线错误帧

```go
gocan.WithErrFilter(
    gocan.CANErrBusOff |
    gocan.CANErrTxTimeout |
    gocan.CANErrCrtl,
)
```

错误帧通过 `bus.Receive()` 投递；具体错误位与诊断字段位于内核 `can_frame` 的 ID 与 data 字段中，详见 [`docs/socketcan-options.md`](socketcan-options.md#33-witherrfilter)。

### 5.3 内核纳秒时间戳

```go
gocan.WithRecvTimestamp(gocan.RxTimestampNano)
```

启用后 `Frame.TimestampMicros` 由内核写入，精度比 `time.Now()` 合成方案高，且无 reader 调度抖动。

更多 Linux 专属 Option（`WithJoinFilters` / `WithSocketBuffers` / `WithRWTimeout`）见 [`docs/socketcan-options.md`](socketcan-options.md)。

## 6. 多通道：BusGroup

`BusGroup` 一次管多个 SocketCAN 接口，合流接收 + 一行收尾：

```go
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"

	"github.com/zhuzx17/gocan"
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

	if _, err := g.Add("front", gocan.SocketCAN("vcan0")); err != nil {
		log.Fatalf("add front: %v", err)
	}
	if _, err := g.Add("rear", gocan.SocketCAN("vcan1")); err != nil {
		log.Fatalf("add rear: %v", err)
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

完整示例见 `examples/11_busgroup_socketcan/`。

## 7. 下一步

- 自动化回归测试模板：`examples/13_socketcan_loopback/`
- 完整调优演示：`examples/14_socketcan_advanced/`
- Option 速查 + 平台对照：[`docs/options.md`](options.md)
- Linux 选项深度阅读：[`docs/socketcan-options.md`](socketcan-options.md)
- 故障排查：[`docs/troubleshooting.md`](troubleshooting.md)
