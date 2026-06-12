# 文档刷新与仓库重命名 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 module path `github.com/Crush251/gocan` 同步刷新到新仓库 `github.com/zhuzxdev/gocan`，并补齐双平台 quickstart + Option 总览页 + troubleshooting 平台对照段，**不改任何业务代码行为**。

**Architecture:** 任务分两条主线：(1) 机械字符串替换 module path（`go.mod` + 33 个 `.go` 文件 + 几个 `.md`）作为一个独立 commit，让后续 docs 改动从一个干净基线出发；(2) 文档刷新——拆 quickstart、新增 options.md、加 troubleshooting 对比段、更新 README/CHANGELOG。每个 doc 任务一个 commit，便于审查。

**Tech Stack:** Go 1.22；Markdown；shell（`grep` / `sed`）。无新依赖。

**Spec:** `docs/superpowers/specs/2026-06-12-docs-and-rename-design.md`

---

## File Structure

新增：

- `docs/quickstart-linux.md` — Linux 快速启动（vcan + 真接口、Classical + FD、3 个常用 Linux Option、BusGroup）
- `docs/quickstart-windows.md` — Windows 快速启动（PCAN 驱动安装、Classical + FD、接收模式、多通道）
- `docs/options.md` — 所有 `WithXxx` Option 集中速查 + 平台标注 + 详解 + 运行期方法

删除：

- `docs/quickstart.md` —— 拆为上面两份

修改（仅替换 `Crush251` → `zhuzxdev` / 添加内容，不改代码逻辑）：

- `go.mod` — module 行
- 33 个 `.go` 文件（19 个 source + 14 个 example）— `import` 路径
- `README.md` — badges + 快速开始入口段 + 内嵌 URL
- `CHANGELOG.md` — 替换旧引用 + 文末追加 `[Unreleased]` 条目
- `doc.go` — 包注释里的 import 示例
- `raw/doc.go` — 包注释里的 顶层包引用
- `docs/troubleshooting.md` — 新增「平台对照速查」段

不动（spec §2 不变量）：

- `docs/superpowers/{specs,plans}/` 整体只读（含旧 `Crush251` 引用是历史快照）
- `docs/socketcan-options.md` / `platform-support.md` / `can-fd.md` / `architecture.md` / `error-handling.md` / `hardware-test-setup.md`
- `scripts/setup-vcan.sh` / `justfile` / `LICENSE`
- 任何 `*.go` 业务/测试文件的实现或测试逻辑

---

## Task List

1. Module path 机械重命名
2. 删除旧 `docs/quickstart.md`
3. 新增 `docs/quickstart-linux.md`
4. 新增 `docs/quickstart-windows.md`
5. 新增 `docs/options.md`
6. `docs/troubleshooting.md` 加平台对照段
7. 更新 `README.md`（badges + 快速开始入口段）
8. 更新 `CHANGELOG.md`（旧引用 + `[Unreleased]` 条目）
9. 最终验证（grep 干净度 + 三平台构建 + 测试）

---

## Task 1: Module path 机械重命名

**Files:**
- Modify: `go.mod`
- Modify: 33 个 `.go` 文件（精确清单见 Step 1 的 grep）
- Modify: `doc.go`（package 注释里的 import 示例）
- Modify: `raw/doc.go`（package 注释里的顶层包引用）

- [ ] **Step 1: 列出所有命中文件（确认范围）**

```bash
cd /home/linkerhand/recode/Go_win_can
grep -rln "Crush251" --include="*.go" --include="*.mod" . 2>/dev/null | grep -v "^./.git/" | grep -v "^./docs/superpowers/" | sort
```

Expected: 34 行（1 个 `go.mod` + 33 个 `.go`）。在 `docs/superpowers/` 下若有命中应被排除。如果数量异常，停下检查。

- [ ] **Step 2: 修改 `go.mod`**

把 `go.mod` 第 1 行 `module github.com/Crush251/gocan` 改为：

```
module github.com/zhuzxdev/gocan
```

- [ ] **Step 3: 全量替换 .go 文件中的 import 路径**

```bash
cd /home/linkerhand/recode/Go_win_can
grep -rln "github.com/Crush251/gocan" --include="*.go" . 2>/dev/null \
  | grep -v "^./.git/" \
  | grep -v "^./docs/superpowers/" \
  | xargs sed -i 's|github.com/Crush251/gocan|github.com/zhuzxdev/gocan|g'
```

- [ ] **Step 4: 验证替换干净**

```bash
grep -rn "Crush251" --include="*.go" --include="*.mod" . 2>/dev/null | grep -v "^./.git/" | grep -v "^./docs/superpowers/"
```

Expected: **0 行**。

- [ ] **Step 5: 三平台构建 + 测试**

```bash
go vet ./...
go test ./... -race -timeout 120s
GOOS=linux   go build ./...
GOOS=windows go build ./...
GOOS=darwin  go build ./...
```

Expected: 全部 PASS / clean。

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "refactor: rename module path to github.com/zhuzxdev/gocan"
```

---

## Task 2: 删除旧 `docs/quickstart.md`

**Files:**
- Delete: `docs/quickstart.md`

- [ ] **Step 1: 删除文件**

```bash
git rm docs/quickstart.md
```

- [ ] **Step 2: 确认无残留引用（暂时允许 README/CHANGELOG 待 Task 7/8 修）**

```bash
grep -rn "quickstart\.md" --include="*.go" --include="*.md" . 2>/dev/null \
  | grep -v "^./.git/" \
  | grep -v "^./docs/superpowers/"
```

输出可能命中 `README.md` / `CHANGELOG.md` —— 这些会在 Task 7/8 替换为新文件。其他位置不应该有引用；如果 docs/ 里有其他文件引用 `quickstart.md`（不是 quickstart-linux/windows.md），停下处理。

- [ ] **Step 3: Commit**

```bash
git commit -m "docs: remove old monolithic quickstart.md (split incoming)"
```

---

## Task 3: 新增 `docs/quickstart-linux.md`

**Files:**
- Create: `docs/quickstart-linux.md`

- [ ] **Step 1: 创建文件**

写入下面完整内容到 `docs/quickstart-linux.md`：

````markdown
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
go get github.com/zhuzxdev/gocan@latest
```

## 3. Classical CAN 收发

最小可运行示例（自发自收，需要 `vcan0`）：

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/zhuzxdev/gocan"
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

	"github.com/zhuzxdev/gocan"
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

错误帧通过 `bus.Receive()` 投递，`Frame.ID` 中带 `CAN_ERR_FLAG` 与 mask 信息。完整位掩码见 [`docs/socketcan-options.md`](socketcan-options.md#33-witherrfilter)。

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
````

- [ ] **Step 2: 验证 markdown 链接相对路径**

```bash
# 确认引用的目标文件都存在
ls docs/socketcan-options.md docs/options.md docs/troubleshooting.md examples/11_busgroup_socketcan examples/13_socketcan_loopback examples/14_socketcan_advanced 2>&1
```

`docs/options.md` 在 Task 5 创建，此时不存在是正常的（CI 不会渲染 markdown 链接，发布到 GitHub 时 Task 5 commit 已落地）。其它文件应都存在。

- [ ] **Step 3: Commit**

```bash
git add docs/quickstart-linux.md
git commit -m "docs: add Linux quickstart with Classical and FD walkthroughs"
```

---

## Task 4: 新增 `docs/quickstart-windows.md`

**Files:**
- Create: `docs/quickstart-windows.md`

- [ ] **Step 1: 创建文件**

写入下面完整内容到 `docs/quickstart-windows.md`：

````markdown
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
````

- [ ] **Step 2: Commit**

```bash
git add docs/quickstart-windows.md
git commit -m "docs: add Windows quickstart with Classical and FD walkthroughs"
```

---

## Task 5: 新增 `docs/options.md`

**Files:**
- Create: `docs/options.md`

- [ ] **Step 1: 创建文件**

写入下面完整内容到 `docs/options.md`：

````markdown
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
````

- [ ] **Step 2: Commit**

```bash
git add docs/options.md
git commit -m "docs: add unified options.md with platform annotations"
```

---

## Task 6: `docs/troubleshooting.md` 加平台对照段

**Files:**
- Modify: `docs/troubleshooting.md`

- [ ] **Step 1: 读现有文件确认插入位置**

```bash
cat docs/troubleshooting.md
```

确认现有结构。新段「平台对照速查」应插在文件末尾（如果无明显小节边界）或紧接现有「常见错误 / 常见症状」类小节之后。

- [ ] **Step 2: 在文件末尾追加新段**

把下面内容追加到 `docs/troubleshooting.md` 末尾（追加前确认末尾有空行）：

```markdown

## 平台对照速查

下表把常见症状按 Windows / Linux 分别给出根因，方便跨平台排查。

| 症状 | Windows | Linux |
|---|---|---|
| `Open` 报 `ErrNoDriver` | `PCANBasic.dll` 加载失败 — 检查 dll 位置、`GOARCH` 与 dll 架构匹配、`PCANBASIC_DLL_PATH` 环境变量 | SocketCAN 模块未加载 — `sudo modprobe can can_raw vcan` |
| `Open` 报 `ErrIllParamValue` | 通道未连接（PCAN-USB 没插）/ 波特率与硬件不符 | 网络接口不存在 — `ip link show` 确认 `can0`/`vcan0` 已建并 up |
| 收不到帧 | 检查发送端是否真的发出 / 通道总线状态 BUSOFF / 滤波器太窄 | 同左 + `WithJoinFilters(true)` 误用导致 AND 滤波 / 没启用 `RecvOwnMsgs` 又想自发自收 |
| `WithLoopback` 编译报错 | 这是 Linux 专属 Option，Windows 上不存在 | — |
| FD 帧发送失败 | `OpenFD` 的 `fdBitrate` 字符串字段写错 — 见 [`docs/can-fd.md`](can-fd.md) | 接口未启用 FD — `sudo ip link set canX type can ... fd on` |
| 高吞吐丢帧 | `WithRxBufferSize` 调大 / 切到 `ModeEvent` | 同左 + `WithSocketBuffers` 调大（受 `net.core.rmem_max` 限制） |

按平台跑通入门：[`docs/quickstart-linux.md`](quickstart-linux.md) / [`docs/quickstart-windows.md`](quickstart-windows.md)。
所有 Option 速查见 [`docs/options.md`](options.md)。
```

- [ ] **Step 3: Commit**

```bash
git add docs/troubleshooting.md
git commit -m "docs: add platform comparison cheat sheet to troubleshooting"
```

---

## Task 7: 更新 `README.md`

**Files:**
- Modify: `README.md`

- [ ] **Step 1: 替换 badges**

打开 `README.md`，找到顶部 6 个 badge 行（`[![...]](...Crush251/gocan...)`）。把每个 URL 里的 `Crush251` 改为 `zhuzxdev`。Task 1 的全局替换不会动到 `.md`，所以这里手动改。

实际操作：

```bash
sed -i 's|github.com/Crush251/gocan|github.com/zhuzxdev/gocan|g' README.md
sed -i 's|/gh/Crush251/gocan|/gh/zhuzxdev/gocan|g' README.md
```

第二条命令处理 codecov badge URL（形式是 `https://codecov.io/gh/Crush251/gocan/...`）。

- [ ] **Step 2: 替换「快速开始」段**

在 `README.md` 中找到 `## 快速开始` 整段（以下一个 `##` 之前为止），替换为：

```markdown
## 快速开始

按你的平台跳转：

- 🐧 **Linux**（SocketCAN）→ [docs/quickstart-linux.md](docs/quickstart-linux.md)
- 🪟 **Windows**（PEAK PCAN）→ [docs/quickstart-windows.md](docs/quickstart-windows.md)

完整 Option 总览：[docs/options.md](docs/options.md)
```

可以用 `Edit` 工具替换或手工编辑。原段落里的 Go 代码块（`gocan.Open(gocan.USBBus1, ...)`）整体被替代——因为同样的代码在 `quickstart-windows.md` 里有更完整的版本。

- [ ] **Step 3: 验证残留**

```bash
grep -n "Crush251" README.md
```

Expected: 0 行。

```bash
grep -n "quickstart\.md" README.md
```

Expected: 0 行（旧 quickstart.md 引用应都被替换）。

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: update README badges and switch quickstart to per-platform entries"
```

---

## Task 8: 更新 `CHANGELOG.md`

**Files:**
- Modify: `CHANGELOG.md`

- [ ] **Step 1: 替换旧 URL 引用**

```bash
sed -i 's|github.com/Crush251/gocan|github.com/zhuzxdev/gocan|g' CHANGELOG.md
```

- [ ] **Step 2: 在文件顶部 `# Changelog` 标题与 `## [0.1.0]` 之间插入 `[Unreleased]` 段**

打开 `CHANGELOG.md`，在第二个空行（紧接 "版本号遵循..." 那段之后、`## [0.1.0] - 2026-05-22` 之前）插入：

```markdown
## [Unreleased]

### Changed

- 仓库迁移到 `github.com/zhuzxdev/gocan`，module path 同步更新。所有
  `import "github.com/Crush251/gocan"` 需改为 `import "github.com/zhuzxdev/gocan"`。
  GitHub 在一段时间内会保留旧 URL 的 redirect，但建议尽快更新。

### Added

- BusGroup：按业务名字管理多个 `*Bus`，提供合流接收 (`Receive() <-chan SourcedFrame`)、聚合关闭 (`*GroupCloseError`)、`Each` 遍历等方法。
- Linux SocketCAN 自定义参数：`WithLoopback` / `WithRecvOwnMsgs` / `WithErrFilter` / `WithJoinFilters` / `WithRecvTimestamp` / `WithSocketBuffers` / `WithRWTimeout`。
- 运行期方法：`(b *Bus) SetErrFilter(uint32)` / `SetJoinFilters(bool)`，Linux 真实生效，其它平台返回 `ErrNotSupported`。
- 4 个新示例：`examples/11_busgroup_socketcan` / `12_busgroup_fan_in` / `13_socketcan_loopback` / `14_socketcan_advanced`。
- 工具：`scripts/setup-vcan.sh` 一键创建 / 销毁 vcan 接口；`justfile` 别名 `vcan-up` / `vcan-down`。

### Documentation

- 拆分 `docs/quickstart.md` → `docs/quickstart-linux.md` + `docs/quickstart-windows.md`，
  各自独立讲清 Classical 与 CAN FD 的 5 分钟启动流程。
- 新增 `docs/options.md`：所有 `WithXxx` Option 集中速查表 + 平台标注 + 详细行为说明。
- 新增 `docs/socketcan-options.md`：Linux 专属 Option 深度阅读（每项对应 setsockopt 名 + 内核版本要求）。
- `docs/troubleshooting.md` 增加「平台对照速查」段：常见失败按 Windows / Linux 分别给出根因。
```

- [ ] **Step 3: 验证残留**

```bash
grep -n "Crush251" CHANGELOG.md
```

Expected: 0 行。

```bash
grep -n "quickstart\.md" CHANGELOG.md
```

Expected: 0 行（如有命中需替换为新文件名）。

- [ ] **Step 4: Commit**

```bash
git add CHANGELOG.md
git commit -m "docs: add [Unreleased] changelog entry covering rename and new docs"
```

---

## Task 9: 最终验证

**Files:** 无文件改动；只跑命令。

- [ ] **Step 1: 重命名干净度（全仓库扫）**

```bash
grep -rn "Crush251" \
  --include="*.go" --include="*.md" --include="*.mod" --include="*.yml" \
  . 2>/dev/null \
  | grep -v "^./.git/" \
  | grep -v "^./docs/superpowers/"
```

Expected: **0 行**。如有命中需修复。

- [ ] **Step 2: 三平台构建**

```bash
GOOS=linux   go build ./...
GOOS=windows go build ./...
GOOS=darwin  go build ./...
```

Expected: 无输出，全过。

- [ ] **Step 3: 测试 + race**

```bash
go vet ./...
go test ./... -race -timeout 120s
```

Expected: 全 PASS（包含 `TestSocketCANIntegration_*` 在 Linux + 有 vcan 时 PASS，否则 SKIP）。

- [ ] **Step 4: 示例编译**

```bash
for d in examples/*/; do GOOS=linux go build ./"$d" || exit 1; done

for d in examples/*/; do
  case "$d" in
    examples/13_*/|examples/14_*/) continue ;;  # Linux only build-tagged
  esac
  GOOS=windows go build ./"$d" || exit 1
done
```

Expected: 全 OK。

- [ ] **Step 5: Markdown 链接完整性（轻量手工）**

```bash
# 列出新增 / 改动的 .md 中所有相对链接
grep -hoE '\(([^)]*\.md[^)]*)\)' \
  README.md CHANGELOG.md \
  docs/quickstart-linux.md docs/quickstart-windows.md docs/options.md docs/troubleshooting.md \
  | sed 's/[()]//g' | sort -u
```

人工核对每条目标存在（带相对路径解析）：在 `docs/` 内引用 `socketcan-options.md` / `options.md` / `quickstart-*.md` / `troubleshooting.md` / `can-fd.md` 都应能 `ls docs/<那个文件>` 命中；README/CHANGELOG 引用的 `docs/...` 应能 `ls docs/...` 命中。

- [ ] **Step 6: 确认无未追踪改动**

```bash
git status -s
```

Expected: 仅看到任务过程之外的预先存在文件（如 11_busgroup_socketcan / 12_busgroup_fan_in 这些用户本地构建产物，与本任务无关）。本任务的所有改动应已 commit。

- [ ] **Step 7: 列出 PR commit 序列做最终确认**

```bash
git log main..HEAD --oneline
```

Expected: 8 个 commit（Task 1-8 各一个；Task 9 无 commit）。

如全部干净，准备 push + PR。

---

## Self-Review

- ✅ Spec coverage — §1（背景）：plan 概述覆盖；§2（不变量）：Task 1 Step 5 三平台构建 + Task 9 全套验证；§3（文件清单）：Task 1-8 各覆盖；§4（rename）：Task 1；§5（quickstart）：Task 3-4；§6（options.md）：Task 5；§7（README/CHANGELOG/troubleshooting/doc.go/raw/doc.go）：Task 1（doc.go + raw/doc.go 走全局 sed）+ Task 6 + 7 + 8；§8（测试）：Task 9；§10（范围之外）：plan 全程不动 `*.go` 业务/测试逻辑、`docs/superpowers/`、`scripts/`、`justfile`。
- ✅ Placeholder scan — 无 TBD/TODO/incomplete sections。
- ✅ Type/path consistency — 文档间链接 `docs/options.md` / `docs/quickstart-{linux,windows}.md` / `docs/socketcan-options.md` / `docs/troubleshooting.md` 名字在 README、CHANGELOG、各 quickstart、options.md、troubleshooting.md 间完全一致。
