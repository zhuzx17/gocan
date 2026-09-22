package gocan

import "testing"

func TestMiniCANFDWindowsSetFilterDropsChannel(t *testing.T) {
	var (
		gotDevice uint32
		gotNumber int8
		gotType   int8
		gotID     uint32
		gotMask   uint32
		gotEnable int8
	)
	raw := func(device uint32, number, typ int8, id, mask uint32, enable int8) int32 {
		gotDevice, gotNumber, gotType = device, number, typ
		gotID, gotMask, gotEnable = id, mask, enable
		return 17
	}

	if status := miniCANFDWindowsSetFilter(raw, 7, 3, 0, 1, 0x123, 0x7ff, 1); status != 17 {
		t.Fatalf("status = %d, want 17", status)
	}
	if gotDevice != 7 || gotNumber != 0 || gotType != 1 || gotID != 0x123 || gotMask != 0x7ff || gotEnable != 1 {
		t.Fatalf("Windows filter args = (%d, %d, %d, %#x, %#x, %d)", gotDevice, gotNumber, gotType, gotID, gotMask, gotEnable)
	}
}
