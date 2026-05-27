//go:build !windows && !linux

package raw

import "unsafe"

// EnsureLoaded 在非 Windows 平台总是返回 nil；后续所有调用返回 PCAN_ERROR_ILLOPERATION。
//
// 设计意图：让 Linux/macOS 上的 lint / vet / 纯逻辑单测可以正常编译运行。
// 真实的 CAN 通信能力仅在 Windows + PCANBasic.dll 环境下提供。
func EnsureLoaded() error { return nil }

// Initialize 是 CAN_Initialize 的桩实现。
func Initialize(ch TPCANHandle, br TPCANBaudrate) TPCANStatus {
	return PCAN_ERROR_ILLOPERATION
}

// InitializeFD 是 CAN_InitializeFD 的桩实现。
func InitializeFD(ch TPCANHandle, bitrateFD string) TPCANStatus {
	return PCAN_ERROR_ILLOPERATION
}

// Uninitialize 是 CAN_Uninitialize 的桩实现。
func Uninitialize(ch TPCANHandle) TPCANStatus { return PCAN_ERROR_ILLOPERATION }

// Reset 是 CAN_Reset 的桩实现。
func Reset(ch TPCANHandle) TPCANStatus { return PCAN_ERROR_ILLOPERATION }

// GetStatus 是 CAN_GetStatus 的桩实现。
func GetStatus(ch TPCANHandle) TPCANStatus { return PCAN_ERROR_ILLOPERATION }

// Read 是 CAN_Read 的桩实现。
func Read(ch TPCANHandle, m *TPCANMsg, t *TPCANTimestamp) TPCANStatus {
	return PCAN_ERROR_ILLOPERATION
}

// ReadFD 是 CAN_ReadFD 的桩实现。
func ReadFD(ch TPCANHandle, m *TPCANMsgFD, t *TPCANTimestampFD) TPCANStatus {
	return PCAN_ERROR_ILLOPERATION
}

// Write 是 CAN_Write 的桩实现。
func Write(ch TPCANHandle, m *TPCANMsg) TPCANStatus { return PCAN_ERROR_ILLOPERATION }

// WriteFD 是 CAN_WriteFD 的桩实现。
func WriteFD(ch TPCANHandle, m *TPCANMsgFD) TPCANStatus { return PCAN_ERROR_ILLOPERATION }

// FilterMessages 是 CAN_FilterMessages 的桩实现。
func FilterMessages(ch TPCANHandle, fromID, toID uint32, mode TPCANMessageType) TPCANStatus {
	return PCAN_ERROR_ILLOPERATION
}

// GetValue 是 CAN_GetValue 的桩实现。
func GetValue(ch TPCANHandle, p TPCANParameter, buf unsafe.Pointer, n uint32) TPCANStatus {
	return PCAN_ERROR_ILLOPERATION
}

// SetValue 是 CAN_SetValue 的桩实现。
func SetValue(ch TPCANHandle, p TPCANParameter, buf unsafe.Pointer, n uint32) TPCANStatus {
	return PCAN_ERROR_ILLOPERATION
}

// GetErrorText 是 CAN_GetErrorText 的桩实现。
func GetErrorText(code TPCANStatus, lang uint16) (string, TPCANStatus) {
	return "", PCAN_ERROR_ILLOPERATION
}
