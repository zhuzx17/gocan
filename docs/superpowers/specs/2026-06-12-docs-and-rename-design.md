# 文档刷新与仓库重命名 — 设计文档

- 日期：2026-06-12
- 范围：仓库迁移到 `github.com/zhuzxdev/gocan` 的同步更新；新增双平台 quickstart 与 Option 总览；不修改任何业务代码行为。
- 路径：路径 B（体系化文档刷新版），用户已敲定
- 分支：`docs/rename-and-platform-quickstart`

## 1. 背景与动机

仓库已经迁移到 `github.com/zhuzxdev/gocan`：GitHub 端在 push 时已经返回新地址提示。本地 `git remote` 已更新。但代码内部的 module path、所有 examples 的 import、README badges、CHANGELOG、docs 内嵌的 URL，仍指向旧 `Crush251/gocan`，需要同步刷新。

同时这次刷新也是补齐文档的好时机：
- 现有 `docs/quickstart.md` 是 Windows 专用，Linux 用户没有平行的入门页
- Option 数量从 v0.1 的 6 个增加到现在 13 个（6 跨平台 + 7 Linux 专属），需要一页可索引的总览
- 跨平台用户在 Windows 上调用 `WithLoopback` 等会编译报错，缺一份"这些 Option 在哪个平台可用"的速查

## 2. 不变量

1. **不改业务代码行为**：所有 `*.go` 业务文件、测试文件只改 import 路径，不改一行实现/测试逻辑
2. **历史只读**：`docs/superpowers/specs/` 与 `docs/superpowers/plans/` 整体不动，即便其中含 `Crush251` 引用——它们是历史快照
3. **不新增依赖**
4. **三平台编译保持干净**：`GOOS=linux/windows/darwin go build ./...` 全过
5. **现有 14 个 examples 功能与运行方式不变**

## 3. 文件清单

```
gocan/
├── go.mod                                # 改: module github.com/zhuzxdev/gocan
├── README.md                             # 改: badges + 快速开始段 + 内嵌 URL
├── CHANGELOG.md                          # 改: 替换旧引用 + 追加 [Unreleased] 条目
├── doc.go                                # 改: 包注释里的 import 路径示例
├── docs/
│   ├── quickstart.md                     # 删除（拆为下面两份）
│   ├── quickstart-linux.md               # 新增
│   ├── quickstart-windows.md             # 新增（重写自旧 quickstart.md）
│   ├── options.md                        # 新增
│   ├── troubleshooting.md                # 改: 增加「平台对照速查」段
│   ├── socketcan-options.md              # 不动
│   ├── platform-support.md               # 不动
│   ├── can-fd.md                         # 不动（被 quickstart 引用）
│   ├── architecture.md                   # 不动
│   ├── error-handling.md                 # 不动
│   ├── hardware-test-setup.md            # 不动
│   └── superpowers/                      # 历史只读，整体不动
└── examples/
    ├── 01_send_classical/main.go         # 改: import 路径
    ├── 02_receive_polling/main.go        # 改: import 路径
    ├── 03_receive_event/main.go          # 改: import 路径
    ├── 04_send_fd/main.go                # 改: import 路径
    ├── 05_receive_fd/main.go             # 改: import 路径
    ├── 06_multi_channel/main.go          # 改: import 路径
    ├── 07_filter/main.go                 # 改: import 路径
    ├── 08_status_and_reset/main.go       # 改: import 路径
    ├── 09_with_logger/main.go            # 改: import 路径
    ├── 10_using_raw/main.go              # 改: import 路径
    ├── 11_busgroup_socketcan/main.go     # 改: import 路径
    ├── 12_busgroup_fan_in/main.go        # 改: import 路径
    ├── 13_socketcan_loopback/main.go     # 改: import 路径
    └── 14_socketcan_advanced/main.go     # 改: import 路径
```

变更分类汇总：

| 类别 | 文件数 | 操作 |
|---|---|---|
| 删除 | 1 | `docs/quickstart.md` |
| 新增 | 3 | `docs/quickstart-linux.md` / `docs/quickstart-windows.md` / `docs/options.md` |
| 改 module path | 1 | `go.mod` |
| 改 import 路径 | 14 | `examples/**/main.go` |
| 改 doc URL/include | 5+ | `README.md` / `CHANGELOG.md` / `doc.go` / `docs/troubleshooting.md` / `docs/can-fd.md` 等含 `Crush251` 字样的 docs |

## 4. Module path 重命名

### 4.1 精确替换规则

机械字符串替换（在 §3 列出的"不动"作用域之外）：

```
github.com/Crush251/gocan  →  github.com/zhuzxdev/gocan
```

### 4.2 安全策略

1. **先扫，后改**：
   ```bash
   grep -rn "Crush251" \
     --include="*.go" --include="*.md" --include="*.mod" --include="*.yml" \
     . | grep -v "^./.git/" | grep -v "^./docs/superpowers/"
   ```
   过目所有命中确认没有意外内容（如硬编码外链）。
2. **历史目录显式排除**：`docs/superpowers/specs/` 与 `docs/superpowers/plans/` 整体不动。
3. **CI 配置确认**：`grep -rn "Crush251" .github/` 检查 GitHub Actions workflow 中是否有写死的路径（按默认模板用 `${{ github.repository }}` 占位变量则无需改）。

### 4.3 执行顺序

```
1. go.mod         module 行
2. examples/      14 个 import
3. README.md      badges + import 示例 + git clone URL
4. doc.go         包注释里的 import 示例
5. CHANGELOG.md   替换既有引用 + 文末新增 [Unreleased] 条目（§7.2）
6. docs/*.md      除 superpowers/ 之外含 Crush251 的文件（troubleshooting / quickstart-* 新建时直接写新名）
7. go vet ./... + GOOS=linux/windows/darwin go build ./... 三平台验证
```

### 4.4 风险与回滚

- 若 `go vet` / `build` 失败，最常见原因是漏改某个 import；§4.2 的 `grep` 应返回 0 行（除 `docs/superpowers/` 内）才算干净。
- 不保留 `replace` 指令。无下游 / redirect 维护成本。
- GitHub release tag 历史不动。

### 4.5 单元 / 集成测试不受影响

所有测试 import 用相对当前 module 的解析，模块改名后照常编译。

## 5. 两份 quickstart 大纲

### 5.1 `docs/quickstart-linux.md`（新增，~200 行）

```
# Linux 快速启动

1. 准备 SocketCAN 接口
   1.1 真实 CAN 接口（如 can0 / 通过 USB-CAN 设备）
       sudo ip link set can0 type can bitrate 500000
       sudo ip link set can0 up
   1.2 虚拟 vcan（开发 / 测试用，无需硬件）
       直接：sudo ./scripts/setup-vcan.sh up vcan0 vcan1
       手动：sudo modprobe vcan && sudo ip link add vcan0 type vcan && sudo ip link set vcan0 up

2. 安装库
   go get github.com/zhuzxdev/gocan@latest

3. Classical CAN 收发（最小可运行示例 ~25 行）
   - 完整 main.go：Open(SocketCAN("vcan0")) + NewFrame + Send + ReadOne
   - 预期输出
   - 在另一终端用 candump vcan0 验证

4. CAN FD 收发
   - 说明 Linux SocketCAN 的 FD bitrate 由 ip link 配置：
     sudo ip link set can0 down
     sudo ip link set can0 type can bitrate 500000 dbitrate 2000000 fd on
     sudo ip link set can0 up
   - OpenFD(SocketCAN("can0"), "") 即可（fdBitrate 传空）
   - 完整 main.go：FD 帧（DLC=12 / data 12 字节 / BRS）

5. 常用 Linux 专属选项一览
   引用 docs/socketcan-options.md，仅展示 3 项最常用：
     - WithLoopback + WithRecvOwnMsgs（自发自收回归测试）
     - WithErrFilter（订阅总线错误帧）
     - WithRecvTimestamp（内核纳秒时间戳）

6. 多通道：BusGroup
   30 行示例：开 vcan0/vcan1，按名字 Add，主循环 for sf := range g.Receive()

7. 下一步
   - examples/13_socketcan_loopback / examples/14_socketcan_advanced
   - docs/options.md 完整参数总览
   - docs/troubleshooting.md 故障排查
```

### 5.2 `docs/quickstart-windows.md`（新增，重写自旧 quickstart.md，~200 行）

```
# Windows 快速启动

1. 安装 PCAN 驱动
   下载 https://www.peak-system.com/quick/DrvSetup
   安装后设备管理器应能看到 PCAN-USB 节点

2. 放置 PCANBasic.dll
   驱动包自带，标准搜索路径：与 exe 同目录 → PATH → 环境变量 PCANBASIC_DLL_PATH（最高优先级）
   GOARCH=amd64 → 用 x64\PCANBasic.dll；GOARCH=386 → 用 x86\PCANBasic.dll

3. 安装库
   go get github.com/zhuzxdev/gocan@latest

4. Classical CAN 收发（最小可运行示例 ~25 行）
   - 完整 main.go：Open(USBBus1, WithBitrate(Baud500K)) + NewFrame + Send + ReadOne
   - 预期输出
   - 用 PCAN-View 或第二台 PCAN-USB 验证

5. CAN FD 收发
   - PCAN FD 用 fdBitrate 字符串配置（与 Linux 不同）：
     fdBitrate := "f_clock=80000000,nom_brp=10,nom_tseg1=12,nom_tseg2=3,nom_sjw=1,data_brp=4,data_tseg1=7,data_tseg2=2,data_sjw=1"
     bus, err := gocan.OpenFD(gocan.USBBus1, fdBitrate)
   - 完整 main.go：发一帧 BRS 加速 FD 帧
   - 链接到 docs/can-fd.md（fdBitrate 完整字段表）

6. 接收模式
   ModeAuto / ModePolling / ModeEvent 三种简短对照
   Windows 上推荐用默认 ModeAuto（驱动支持时自动启用 Event，CPU 占用最低）

7. 多通道：两路 PCAN-USB
   引用 examples/06_multi_channel + 30 行 BusGroup 替代写法

8. 下一步
   - examples/03_receive_event / examples/09_with_logger
   - docs/options.md 完整参数总览
   - docs/troubleshooting.md 故障排查
```

### 5.3 `docs/quickstart.md` 删除

直接 `git rm`。所有现有引用（`README.md` / `CHANGELOG.md` 等）会一并指向新文件名。

## 6. `docs/options.md`（新增，~300 行）

### 6.1 速查表

#### 6.1.1 跨平台 Option（gocan 包）

| Option | 默认值 | 作用 | Linux | Windows |
|---|---|---|:-:|:-:|
| `WithBitrate(Bitrate)` | `Baud1M` | Classical CAN 波特率 | ✗（被忽略，由 ip link 设）| ✓ |
| `WithReceiveMode(ReceiveMode)` | `ModeAuto` | reader goroutine 等待策略 | 仅 Polling 生效 | Auto/Polling/Event |
| `WithPollInterval(time.Duration)` | `1ms` | Polling 模式轮询间隔 | ✓ | ✓（Polling 模式时）|
| `WithRxBufferSize(int)` | `1024` | 接收 channel 容量 | ✓ | ✓ |
| `WithErrBufferSize(int)` | `16` | 错误 channel 容量 | ✓ | ✓ |
| `WithLogger(Logger)` | `noopLogger{}` | 注入日志接口 | ✓ | ✓ |

#### 6.1.2 Linux 专属 Option（`//go:build linux`）

| Option | 默认值 | 作用 | 内核要求 |
|---|---|---|---|
| `WithLoopback(bool)` | 内核默认 true | `CAN_RAW_LOOPBACK` | 3.6+ |
| `WithRecvOwnMsgs(bool)` | 内核默认 false | `CAN_RAW_RECV_OWN_MSGS` | 3.6+ |
| `WithErrFilter(uint32)` | 不设置 | `CAN_RAW_ERR_FILTER` | 3.6+ |
| `WithJoinFilters(bool)` | 不设置（OR）| `CAN_RAW_JOIN_FILTERS` | **4.1+** |
| `WithRecvTimestamp(RxTimestamp)` | `RxTimestampNone` | `SO_TIMESTAMP*` | 取决于 mode |
| `WithSocketBuffers(rcv, snd int)` | 不设置 | `SO_RCVBUF` / `SO_SNDBUF` | always |
| `WithRWTimeout(read, write Duration)` | 不设置 | `SO_RCVTIMEO` / `SO_SNDTIMEO` | always |

#### 6.1.3 Windows 专属 Option

> 暂无。Windows PCAN 后端常见参数（`PCAN_LISTEN_ONLY`、`PCAN_ALLOW_*_FRAMES`、`PCAN_BUSOFF_AUTORESET`）将在后续 PR 补全，详见 `docs/superpowers/specs/2026-06-08-busgroup-and-socketcan-options-design.md` §12。

### 6.2 跨平台 Option 详解（每项 1 个小节）

每节包含：函数签名、平台行为差异、最小代码片段、非正值/零值处理。

- `WithBitrate`：Win 走 `CAN_Initialize`；Linux 忽略，由内核 netlink 配
- `WithReceiveMode`：三模式表 + Linux Auto/Event 降级到 Polling
- `WithPollInterval`：仅 Polling 模式有意义；非正值忽略；调小 → 延迟低 / CPU 高
- `WithRxBufferSize`：满了 reader 丢帧（reader.go:42）；非正值忽略
- `WithErrBufferSize`：errCh "提示性"通道，满了直接丢
- `WithLogger`：传 nil 被忽略

### 6.3 Linux 专属 Option 详解

每项一节，每节链接到 `docs/socketcan-options.md` 对应章节做深度阅读。

### 6.4 运行期方法

```go
func (b *Bus) SetErrFilter(mask uint32) error
func (b *Bus) SetJoinFilters(and bool) error
```

仅 Linux 真实生效，其他平台返回 `ErrNotSupported`。

### 6.5 错误处理

任意 Option / 运行期方法触发的 setsockopt / Initialize 失败 → `*gocan.Error{Op: "...", Code: ..., Msg: ...}`。`errors.Is(err, ErrIllParamValue)` / `ErrNoDriver` / `ErrBusClosed` 命中关系一行表。

### 6.6 与现有文档的关系

- `docs/socketcan-options.md`：Linux 专属深度阅读，本页是简表入口
- `docs/can-fd.md`：fdBitrate 字符串字段表（不在 Option 范畴）
- `docs/quickstart-{linux,windows}.md`：5 分钟跑通第一帧

## 7. 其他文件改动

### 7.1 `README.md`

只改三段：

1. 顶部 6 个 badges + pkg.go.dev 链接 → `Crush251` 替换为 `zhuzxdev`
2. 「快速开始」段替换为双平台入口：
   ```markdown
   ## 快速开始

   按你的平台跳转：

   - 🐧 **Linux**（SocketCAN）→ [docs/quickstart-linux.md](docs/quickstart-linux.md)
   - 🪟 **Windows**（PEAK PCAN）→ [docs/quickstart-windows.md](docs/quickstart-windows.md)

   完整 Option 总览：[docs/options.md](docs/options.md)
   ```
3. 「快速开始」段之后的「通道发现」/「Linux SocketCAN」/「系统要求」段：保留原内容，只替换 `Crush251` → `zhuzxdev` 和 `git clone` URL。

不动：徽章下面的项目介绍段、为什么做这个段、特性段、路线图段、许可证段（除已替换的 URL 之外）。

### 7.2 `CHANGELOG.md`

文末追加新条目：

```markdown
## [Unreleased]

### Changed
- 仓库迁移到 `github.com/zhuzxdev/gocan`，module path 同步更新。所有
  `import "github.com/Crush251/gocan"` 需改为 `import "github.com/zhuzxdev/gocan"`。
  GitHub 在一段时间内会保留旧 URL 的 redirect，但建议尽快更新。

### Documentation
- 拆分 `docs/quickstart.md` → `docs/quickstart-linux.md` + `docs/quickstart-windows.md`，
  各自独立讲清 Classical 与 CAN FD 的 5 分钟启动流程。
- 新增 `docs/options.md`：所有 `WithXxx` Option 集中速查表 + 平台标注 + 详细行为说明。
- `docs/troubleshooting.md` 增加平台对比段：常见失败按 Windows / Linux 分别给出根因。
```

CHANGELOG 内已有的 `Crush251` 引用（如有）一并替换。

### 7.3 `docs/troubleshooting.md` 平台对比段

在适当位置（"常见错误"段之后）插入：

```markdown
## 平台对照速查

| 症状 | Windows | Linux |
|---|---|---|
| `Open` 报 `ErrNoDriver` | `PCANBasic.dll` 加载失败 — 检查 dll 位置、GOARCH 与 dll 架构匹配、`PCANBASIC_DLL_PATH` 环境变量 | SocketCAN 模块未加载 — `sudo modprobe can can_raw vcan` |
| `Open` 报 `ErrIllParamValue` | 通道未连接（PCAN-USB 没插）/ 波特率与硬件不符 | 网络接口不存在 — `ip link show` 确认 can0/vcan0 已建并 up |
| 收不到帧 | 检查发送端是否真的发出 / 通道总线状态 BUSOFF / 滤波器太窄 | 同左 + `WithJoinFilters(true)` 误用导致 AND 滤波 / 没启用 `RecvOwnMsgs` 又想自发自收 |
| `WithLoopback` 编译报错 | 这是 Linux 专属 Option，Windows 上不存在 | — |
| FD 帧发送失败 | `OpenFD` 的 `fdBitrate` 字符串字段写错 — 见 `docs/can-fd.md` | 接口未启用 FD — `ip link set can0 type can ... fd on` |
| 高吞吐丢帧 | `WithRxBufferSize` 调大 / 切到 `ModeEvent` | 同左 + `WithSocketBuffers` 调大（受 `net.core.rmem_max` 限制） |
```

### 7.4 `doc.go`

包注释里第一个代码块的 import 路径更新；其它不动。

## 8. 测试与验证

### 8.1 重命名干净度

```bash
grep -rn "Crush251" \
  --include="*.go" --include="*.md" --include="*.mod" --include="*.yml" \
  . | grep -v "^./.git/" | grep -v "^./docs/superpowers/"
```

期望：**0 行**。

### 8.2 构建与测试

```bash
go vet ./...
go test ./... -race -timeout 120s
GOOS=linux   go build ./...
GOOS=windows go build ./...
GOOS=darwin  go build ./...
```

### 8.3 示例编译

```bash
for d in examples/*/; do GOOS=linux go build ./"$d" || exit 1; done
for d in examples/*/; do
  case "$d" in
    examples/13_*/|examples/14_*/) continue ;;  # Linux only
  esac
  GOOS=windows go build ./"$d" || exit 1
done
```

### 8.4 链接完整性（轻量手工验证）

- 新增 / 修改的 .md 里的相对路径 `[xxx](path)` 用 `grep + ls` 验证目标存在
- `https://github.com/zhuzxdev/...` 链接用 `grep` 确认无残留 `Crush251`

### 8.5 CI 配置

```bash
grep -rn "Crush251" .github/
```

期望：0 行；如有，单独提交修复。

## 9. 工作量估算

| 模块 | 变更行数 | 文件数 |
|---|---|---|
| Module rename（go.mod + 14 examples + README + doc.go + CHANGELOG + docs/*） | ~80 | ~22 |
| `docs/quickstart-linux.md` 新增 | ~200 | 1 |
| `docs/quickstart-windows.md` 新增（重写自旧） | ~200 | 1 |
| `docs/options.md` 新增 | ~300 | 1 |
| `docs/troubleshooting.md` 增段 | ~30 | 1 |
| `CHANGELOG.md` 新条目 | ~15 | 1 |
| `README.md` 双 quickstart 入口段 | ~15 | 1 |
| 删除旧 `docs/quickstart.md` | -80 | 1 |
| **合计** | **~760 行净增** | ~28 文件 |

按主题切 7~9 个 commit、1 个 PR 完成。

## 10. 范围之外（明确不做）

- 改任何 `*.go` 业务/测试文件的实现或测试逻辑（仅改 import 路径）
- 改任何 spec/plan 历史文件（`docs/superpowers/{specs,plans}/` 整体只读）
- 新增 examples 文件
- 补 Windows PCAN 专属 Option（仍延后到独立 PR）
- 改 `scripts/setup-vcan.sh` 或 `justfile`
- 改 GitHub release tag / 现有 PR 标题等远端历史
