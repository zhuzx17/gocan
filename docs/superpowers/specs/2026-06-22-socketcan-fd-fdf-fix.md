# SocketCAN CAN FD 收发修复 — 设计 + 计划

- 日期：2026-06-22
- 范围：仅动 `raw/api_linux.go` + `raw/api_linux_test.go`；共 3 处逻辑修改 + 单元测试。
- 分支：`fix/socketcan-fd-fdf`

## 1. 背景

用户在真硬件（PEAK PCAN-USB FD 类 L30 设备）上发现 gocan Linux SocketCAN 后端的两个 bug：

1. **短 FD 帧收不到应答**：2 字节 payload 的使能帧发出去被设备忽略。
2. **长 FD 帧数据被截断**：48 字节位置帧发出去只有前 8 字节到达对端。

根因：`encodeLinuxCANFDFrame` 里 `buf[5]` 只在 BRS/ESI 时才置位；不带 BRS 的普通 FD 帧发出时 `buf[5]=0`，Linux 内核按 Classical CAN 处理（`can.h` 里的 `CANFD_FDF (0x04)` 位没置） → payload >8 字节被截断/丢弃。

同时 `ReadFD` / `readFDWithTimestamp` 的 `switch n { case 16: ...; case 72: ... }` 过于严格，中间态返回 `PCAN_ERROR_ILLDATA`，reader 静默丢帧。

## 2. 决策：CANFD_FDF 硬编码 0x04

**不新增用户可见 Option**。理由（用户已确认）：

- `CANFD_FDF` 是 Linux 内核 wire-level 协议识别位（`<linux/can.h>`），不是可配置特性
- 用户的 "FD vs Classical" 选择已由 `Open` vs `OpenFD` 表达；进入 `encodeLinuxCANFDFrame` 已隐含"发 FD 帧"
- 让 FDF 可选 = 允许"打开 FD bus 但发出的帧不宣称是 FD"的矛盾状态，等于把当前 bug 保留成 opt-out 陷阱

## 3. 不变量

1. 不新增任何公开 API / Option / 类型
2. 不改任何 `.go` 文件的公开签名（`Read` / `ReadFD` / `encodeLinuxCANFDFrame` / `decodeLinuxCANFDFrame` / `readFDWithTimestamp` 签名全部保留）
3. Windows PCAN 后端完全不受影响（走 DLL，不经这些函数）
4. 现有测试全部继续通过；新增测试锁定新行为

## 4. 修改点（3 处，全在 `raw/api_linux.go`）

### 4.1 常量块加 `linuxCANFDF`

`raw/api_linux.go:15-21` 常量块从：

```go
const (
    linuxCANFrameSize   = 16
    linuxCANFDFrameSize = 72

    linuxCANFDBRS = 0x01
    linuxCANFDESI = 0x02
)
```

改为：

```go
const (
    linuxCANFrameSize   = 16
    linuxCANFDFrameSize = 72

    // Linux <linux/can.h> CAN FD flag 位：
    //   CANFD_FDF (0x04) 标识"这一帧是 CAN FD"，raw socket 发送 FD 帧时必须置位，
    //   否则内核按 Classical CAN 处理，>8 字节 payload 会被截断。
    linuxCANFDF   = 0x04
    linuxCANFDBRS = 0x01
    linuxCANFDESI = 0x02
)
```

### 4.2 `encodeLinuxCANFDFrame` 无条件置 FDF

`raw/api_linux.go:458-481` 里，把：

```go
nativeEndian.PutUint32(buf[0:4], canID)
buf[4] = uint8(length)
if m.MsgType&PCAN_MESSAGE_BRS != 0 {
    buf[5] |= linuxCANFDBRS
}
if m.MsgType&PCAN_MESSAGE_ESI != 0 {
    buf[5] |= linuxCANFDESI
}
```

改为：

```go
nativeEndian.PutUint32(buf[0:4], canID)
buf[4] = uint8(length)
// FDF 位必须置起：raw socket 上没有这位内核按 Classical CAN 处理，
// >8 字节 payload 会被截断。BRS / ESI 按上层 Flags 可选叠加。
buf[5] = linuxCANFDF
if m.MsgType&PCAN_MESSAGE_BRS != 0 {
    buf[5] |= linuxCANFDBRS
}
if m.MsgType&PCAN_MESSAGE_ESI != 0 {
    buf[5] |= linuxCANFDESI
}
```

### 4.3 `ReadFD` 分支放宽

`raw/api_linux.go:203-219` 里，把：

```go
switch n {
case linuxCANFrameSize:
    var cm TPCANMsg
    if status := decodeLinuxCANFrame(buf[:linuxCANFrameSize], &cm); status != PCAN_ERROR_OK {
        return status
    }
    m.ID = cm.ID
    m.MsgType = cm.MsgType
    m.DLC = cm.Len
    copy(m.Data[:], cm.Data[:])
case linuxCANFDFrameSize:
    if status := decodeLinuxCANFDFrame(buf[:], m); status != PCAN_ERROR_OK {
        return status
    }
default:
    return PCAN_ERROR_ILLDATA
}
```

改为按 length + FDF flag 双条件分支：

```go
switch {
case n == linuxCANFrameSize:
    var cm TPCANMsg
    if status := decodeLinuxCANFrame(buf[:linuxCANFrameSize], &cm); status != PCAN_ERROR_OK {
        return status
    }
    m.ID = cm.ID
    m.MsgType = cm.MsgType
    m.DLC = cm.Len
    copy(m.Data[:], cm.Data[:])
case n == linuxCANFDFrameSize || (n >= linuxCANFrameSize && buf[5]&linuxCANFDF != 0):
    var fdBuf [linuxCANFDFrameSize]byte
    copy(fdBuf[:], buf[:n])
    if status := decodeLinuxCANFDFrame(fdBuf[:], m); status != PCAN_ERROR_OK {
        return status
    }
default:
    return PCAN_ERROR_ILLDATA
}
```

变化点：`case linuxCANFDFrameSize` → `case n == linuxCANFDFrameSize || (n >= linuxCANFrameSize && buf[5]&linuxCANFDF != 0)`。这允许非严格 72 字节但带 FDF 标志的读入被当作 FD 帧解码。因为 `decodeLinuxCANFDFrame` 要求 `len(buf) == 72`，把实际读到的 `n` 字节拷贝到 72 字节固定 buffer 后再调用（剩余字节零填充）。

### 4.4 `readFDWithTimestamp` 分支同步

`raw/api_linux.go:672-703` 的 `readFDWithTimestamp` 有一份相同的 `switch n { case 16: ...; case 72: ... }` 逻辑。同 §4.3 一样改。

### 4.5 `decodeLinuxCANFDFrame` 关于 length 上限的鲁棒性

当前 `decodeLinuxCANFDFrame` 中 `length := int(buf[4]); if length > 64 { return PCAN_ERROR_ILLDATA }`。这个已经足够——非法长度直接拒收，正常 FD 帧 length ≤ 64。不动。

## 5. 单元测试（`raw/api_linux_test.go` 追加）

### 5.1 `TestLinuxCANFDFrameEncodeSetsFDF`

验证 encode 后 `buf[5]` 一定包含 FDF 位（0x04），且 BRS/ESI 按 flag 叠加正确：

- 无 flag → `buf[5] == 0x04`
- 有 BRS → `buf[5] == 0x05`（0x04 | 0x01）
- 有 ESI → `buf[5] == 0x06`（0x04 | 0x02）
- 全有 → `buf[5] == 0x07`

### 5.2 `TestLinuxCANFDFrameRoundTripSmall`

2 字节 payload：encode → decode，数据字节 + DLC 都保留：

```go
in := &TPCANMsgFD{
    ID: 0x123, MsgType: PCAN_MESSAGE_FD, DLC: 2,
    Data: [64]byte{0xDE, 0xAD},
}
```

### 5.3 `TestLinuxCANFDFrameRoundTripLarge`

48 字节 payload（DLC=14）：encode → decode，全 48 字节保留，不被截断：

```go
in := &TPCANMsgFD{ID: 0x456, MsgType: PCAN_MESSAGE_FD, DLC: 14}
for i := 0; i < 48; i++ {
    in.Data[i] = byte(0xA0 + i&0x0F)
}
```

## 6. 集成测试（可选，本轮不新增）

现有 `socketcan_options_integration_linux_test.go` 有 `TestSocketCANIntegration_LoopbackRecvOwn` 用 Classical 帧。等本 PR 落地后可以在其中加一个 FD 变体（用 vcan0 + `OpenFD` + `WithLoopback + WithRecvOwnMsgs` + 48 字节 FD 帧），但不属于本 PR 的必要范围。

## 7. 验证 gate

```bash
go vet ./...
go test ./raw -race -v -run "LinuxCANFD"
go test ./... -race -timeout 120s
GOOS=linux   go build ./...
GOOS=windows go build ./...
GOOS=darwin  go build ./...
```

全部必须干净通过。

## 8. Commit 序列

按主题切 3 个 commit：

1. `fix(raw): set CANFD_FDF flag on all FD frame transmissions` — 常量 + `encodeLinuxCANFDFrame` 修复 + `TestLinuxCANFDFrameEncodeSetsFDF`
2. `fix(raw): accept FD frames of non-canonical length in ReadFD` — `ReadFD` + `readFDWithTimestamp` 分支放宽
3. `test(raw): add small and large payload round-trip tests for CAN FD` — 2 字节 + 48 字节 round-trip 测试

按 TDD 顺序：先写测试确认失败，再写实现让测试通过。

## 9. CHANGELOG 更新（可选，本 PR 一起做）

`CHANGELOG.md` 的 `[Unreleased]` 段追加：

```markdown
### Fixed

- Linux SocketCAN 后端发送 CAN FD 帧时未置 `CANFD_FDF` 标志，导致内核按
  Classical CAN 处理，>8 字节 payload 被截断/丢弃。修复后所有 FD 帧发送均
  正确带 FDF 位；BRS / ESI 仍按 `Frame.Flags` 可选叠加。
- `ReadFD` / `readFDWithTimestamp` 分支放宽：现在按"标准 FD 帧长度 或
  长度足够且携带 FDF 标志"识别 FD 帧，避免混合总线场景下静默丢帧。
```

## 10. 范围之外

- 不动 Windows PCAN 后端（不受此 bug 影响）
- 不动 `raw/api_linux.go` 之外的任何业务代码
- 不改 `NewFDFrame` / `Frame.Flags` / 任何公开 API 的语义
- 不加新公开 Option（`WithFDF` 之类都被明确排除）
- `docs/socketcan-options.md` / `options.md` 不动（无新 API 需要文档）
- 不动 `docs/superpowers/{specs,plans}/` 历史目录
