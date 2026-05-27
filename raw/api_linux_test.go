//go:build linux

package raw

import "testing"

func TestLinuxCANFrameRoundTrip(t *testing.T) {
	in := &TPCANMsg{
		ID:      0x123,
		MsgType: PCAN_MESSAGE_RTR,
		Len:     3,
		Data:    [8]byte{1, 2, 3},
	}
	buf, status := encodeLinuxCANFrame(in)
	if status != PCAN_ERROR_OK {
		t.Fatalf("encode status = 0x%X, want OK", uint32(status))
	}

	var out TPCANMsg
	if status := decodeLinuxCANFrame(buf[:], &out); status != PCAN_ERROR_OK {
		t.Fatalf("decode status = 0x%X, want OK", uint32(status))
	}
	if out.ID != in.ID || out.MsgType != in.MsgType || out.Len != in.Len {
		t.Fatalf("decoded header = {ID:0x%X MsgType:0x%X Len:%d}, want {ID:0x%X MsgType:0x%X Len:%d}",
			out.ID, out.MsgType, out.Len, in.ID, in.MsgType, in.Len)
	}
	for i := 0; i < int(in.Len); i++ {
		if out.Data[i] != in.Data[i] {
			t.Fatalf("Data[%d] = %d, want %d", i, out.Data[i], in.Data[i])
		}
	}
}

func TestLinuxCANFrameRoundTripExtended(t *testing.T) {
	in := &TPCANMsg{
		ID:      0x1ABCDE,
		MsgType: PCAN_MESSAGE_EXTENDED | PCAN_MESSAGE_RTR,
		Len:     8,
		Data:    [8]byte{0, 1, 2, 3, 4, 5, 6, 7},
	}
	buf, status := encodeLinuxCANFrame(in)
	if status != PCAN_ERROR_OK {
		t.Fatalf("encode status = 0x%X, want OK", uint32(status))
	}

	var out TPCANMsg
	if status := decodeLinuxCANFrame(buf[:], &out); status != PCAN_ERROR_OK {
		t.Fatalf("decode status = 0x%X, want OK", uint32(status))
	}
	if out.ID != in.ID || out.MsgType != in.MsgType || out.Len != in.Len || out.Data != in.Data {
		t.Fatalf("decoded = %+v, want %+v", out, *in)
	}
}

func TestLinuxCANFrameRejectsInvalidLength(t *testing.T) {
	_, status := encodeLinuxCANFrame(&TPCANMsg{ID: 0x123, Len: 9})
	if status != PCAN_ERROR_ILLDATA {
		t.Fatalf("status = 0x%X, want ILLDATA", uint32(status))
	}
}

func TestLinuxCANFDFrameRoundTrip(t *testing.T) {
	in := &TPCANMsgFD{
		ID:      0x1ABCDEF,
		MsgType: PCAN_MESSAGE_EXTENDED | PCAN_MESSAGE_FD | PCAN_MESSAGE_BRS | PCAN_MESSAGE_ESI,
		DLC:     10,
	}
	for i := 0; i < 16; i++ {
		in.Data[i] = byte(i)
	}
	buf, status := encodeLinuxCANFDFrame(in)
	if status != PCAN_ERROR_OK {
		t.Fatalf("encode status = 0x%X, want OK", uint32(status))
	}

	var out TPCANMsgFD
	if status := decodeLinuxCANFDFrame(buf[:], &out); status != PCAN_ERROR_OK {
		t.Fatalf("decode status = 0x%X, want OK", uint32(status))
	}
	if out.ID != in.ID || out.MsgType != in.MsgType || out.DLC != in.DLC {
		t.Fatalf("decoded header = {ID:0x%X MsgType:0x%X DLC:%d}, want {ID:0x%X MsgType:0x%X DLC:%d}",
			out.ID, out.MsgType, out.DLC, in.ID, in.MsgType, in.DLC)
	}
	for i := 0; i < 16; i++ {
		if out.Data[i] != in.Data[i] {
			t.Fatalf("Data[%d] = %d, want %d", i, out.Data[i], in.Data[i])
		}
	}
}

func TestLinuxCANFDFrameRejectsInvalidDLC(t *testing.T) {
	_, status := encodeLinuxCANFDFrame(&TPCANMsgFD{ID: 0x123, MsgType: PCAN_MESSAGE_FD, DLC: 16})
	if status != PCAN_ERROR_ILLDATA {
		t.Fatalf("status = 0x%X, want ILLDATA", uint32(status))
	}
}

func TestSocketCANFilters_StandardExactID(t *testing.T) {
	filters, status := socketCANFilters(0x123, 0x123, PCAN_MESSAGE_STANDARD)
	if status != PCAN_ERROR_OK {
		t.Fatalf("status = 0x%X, want OK", uint32(status))
	}
	if len(filters) != 1 {
		t.Fatalf("len(filters) = %d, want 1", len(filters))
	}
	if filters[0].Id != 0x123 {
		t.Fatalf("filter id = 0x%X, want 0x123", filters[0].Id)
	}
	wantMask := uint32(0xC00007FF)
	if filters[0].Mask != wantMask {
		t.Fatalf("filter mask = 0x%X, want 0x%X", filters[0].Mask, wantMask)
	}
}

func TestSocketCANFilters_StandardRange(t *testing.T) {
	filters, status := socketCANFilters(0x100, 0x1FF, PCAN_MESSAGE_STANDARD)
	if status != PCAN_ERROR_OK {
		t.Fatalf("status = 0x%X, want OK", uint32(status))
	}
	if len(filters) != 1 {
		t.Fatalf("len(filters) = %d, want 1", len(filters))
	}
	if filters[0].Id != 0x100 {
		t.Fatalf("filter id = 0x%X, want 0x100", filters[0].Id)
	}
	wantMask := uint32(0xC0000700)
	if filters[0].Mask != wantMask {
		t.Fatalf("filter mask = 0x%X, want 0x%X", filters[0].Mask, wantMask)
	}
}

func TestSocketCANFilters_ExtendedRange(t *testing.T) {
	filters, status := socketCANFilters(0x1ABCDE0, 0x1ABCDEF, PCAN_MESSAGE_EXTENDED)
	if status != PCAN_ERROR_OK {
		t.Fatalf("status = 0x%X, want OK", uint32(status))
	}
	if len(filters) != 1 {
		t.Fatalf("len(filters) = %d, want 1", len(filters))
	}
	if filters[0].Id != 0x81ABCDE0 {
		t.Fatalf("filter id = 0x%X, want 0x81ABCDE0", filters[0].Id)
	}
	wantMask := uint32(0xDFFFFFF0)
	if filters[0].Mask != wantMask {
		t.Fatalf("filter mask = 0x%X, want 0x%X", filters[0].Mask, wantMask)
	}
}

func TestSocketCANFilters_SplitsUnalignedRange(t *testing.T) {
	filters, status := socketCANFilters(0x101, 0x103, PCAN_MESSAGE_STANDARD)
	if status != PCAN_ERROR_OK {
		t.Fatalf("status = 0x%X, want OK", uint32(status))
	}
	if len(filters) != 2 {
		t.Fatalf("len(filters) = %d, want 2", len(filters))
	}
	if filters[0].Id != 0x101 || filters[0].Mask != 0xC00007FF {
		t.Fatalf("filter[0] = {Id:0x%X Mask:0x%X}, want {Id:0x101 Mask:0xC00007FF}", filters[0].Id, filters[0].Mask)
	}
	if filters[1].Id != 0x102 || filters[1].Mask != 0xC00007FE {
		t.Fatalf("filter[1] = {Id:0x%X Mask:0x%X}, want {Id:0x102 Mask:0xC00007FE}", filters[1].Id, filters[1].Mask)
	}
}

func TestSocketCANFilters_RejectsInvalidRanges(t *testing.T) {
	tests := []struct {
		name   string
		fromID uint32
		toID   uint32
		mode   TPCANMessageType
	}{
		{"reversed", 0x200, 0x100, PCAN_MESSAGE_STANDARD},
		{"standard out of range", 0x800, 0x800, PCAN_MESSAGE_STANDARD},
		{"extended out of range", 0x20000000, 0x20000000, PCAN_MESSAGE_EXTENDED},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, status := socketCANFilters(tt.fromID, tt.toID, tt.mode)
			if status != PCAN_ERROR_ILLPARAMVAL {
				t.Fatalf("status = 0x%X, want ILLPARAMVAL", uint32(status))
			}
		})
	}
}

func TestLinuxInitializeRejectsPlainPCANHandle(t *testing.T) {
	status := Initialize(PCAN_USBBUS1, PCAN_BAUD_1M)
	if status != PCAN_ERROR_ILLOPERATION {
		t.Fatalf("Initialize(PCAN_USBBUS1) = 0x%X, want ILLOPERATION", uint32(status))
	}
}
