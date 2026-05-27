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

func TestLinuxInitializeRejectsPlainPCANHandle(t *testing.T) {
	status := Initialize(PCAN_USBBUS1, PCAN_BAUD_1M)
	if status != PCAN_ERROR_ILLOPERATION {
		t.Fatalf("Initialize(PCAN_USBBUS1) = 0x%X, want ILLOPERATION", uint32(status))
	}
}
