//go:build windows

package gocan

import (
	"fmt"
	"unsafe"

	"github.com/zhuzxdev/gocan/raw"
)

const (
	pcanChannelUnavailable uint32 = 0
	pcanChannelAvailable   uint32 = 1
	pcanChannelOccupied    uint32 = 2
)

func lookupChannels() ([]ChannelInfo, error) {
	if err := raw.EnsureLoaded(); err != nil {
		return nil, ErrDLLNotFound
	}

	channels := make([]ChannelInfo, 0)
	for _, candidate := range pcanDiscoveryChannels() {
		var condition uint32
		status := raw.GetValue(candidate.channel, raw.PCAN_CHANNEL_CONDITION,
			unsafe.Pointer(&condition), uint32(unsafe.Sizeof(condition)))
		if status != raw.PCAN_ERROR_OK || condition == pcanChannelUnavailable {
			continue
		}
		channels = append(channels, ChannelInfo{
			Channel: candidate.channel,
			Name:    candidate.name,
			Backend: BackendPCAN,
			Up:      condition == pcanChannelAvailable,
			FD:      pcanChannelSupportsFD(candidate.channel),
		})
	}
	return channels, nil
}

type pcanDiscoveryChannel struct {
	channel Channel
	name    string
}

func pcanDiscoveryChannels() []pcanDiscoveryChannel {
	channels := make([]pcanDiscoveryChannel, 0, 48)
	for i, ch := range []Channel{
		raw.PCAN_USBBUS1, raw.PCAN_USBBUS2, raw.PCAN_USBBUS3, raw.PCAN_USBBUS4,
		raw.PCAN_USBBUS5, raw.PCAN_USBBUS6, raw.PCAN_USBBUS7, raw.PCAN_USBBUS8,
		raw.PCAN_USBBUS9, raw.PCAN_USBBUS10, raw.PCAN_USBBUS11, raw.PCAN_USBBUS12,
		raw.PCAN_USBBUS13, raw.PCAN_USBBUS14, raw.PCAN_USBBUS15, raw.PCAN_USBBUS16,
	} {
		channels = append(channels, pcanDiscoveryChannel{channel: ch, name: fmt.Sprintf("PCAN_USBBUS%d", i+1)})
	}
	for i, ch := range []Channel{
		raw.PCAN_PCIBUS1, raw.PCAN_PCIBUS2, raw.PCAN_PCIBUS3, raw.PCAN_PCIBUS4,
		raw.PCAN_PCIBUS5, raw.PCAN_PCIBUS6, raw.PCAN_PCIBUS7, raw.PCAN_PCIBUS8,
		raw.PCAN_PCIBUS9, raw.PCAN_PCIBUS10, raw.PCAN_PCIBUS11, raw.PCAN_PCIBUS12,
		raw.PCAN_PCIBUS13, raw.PCAN_PCIBUS14, raw.PCAN_PCIBUS15, raw.PCAN_PCIBUS16,
	} {
		channels = append(channels, pcanDiscoveryChannel{channel: ch, name: fmt.Sprintf("PCAN_PCIBUS%d", i+1)})
	}
	for i, ch := range []Channel{
		raw.PCAN_LANBUS1, raw.PCAN_LANBUS2, raw.PCAN_LANBUS3, raw.PCAN_LANBUS4,
		raw.PCAN_LANBUS5, raw.PCAN_LANBUS6, raw.PCAN_LANBUS7, raw.PCAN_LANBUS8,
		raw.PCAN_LANBUS9, raw.PCAN_LANBUS10, raw.PCAN_LANBUS11, raw.PCAN_LANBUS12,
		raw.PCAN_LANBUS13, raw.PCAN_LANBUS14, raw.PCAN_LANBUS15, raw.PCAN_LANBUS16,
	} {
		channels = append(channels, pcanDiscoveryChannel{channel: ch, name: fmt.Sprintf("PCAN_LANBUS%d", i+1)})
	}
	return channels
}

func pcanChannelSupportsFD(ch Channel) bool {
	var features uint32
	status := raw.GetValue(ch, raw.PCAN_CHANNEL_FEATURES,
		unsafe.Pointer(&features), uint32(unsafe.Sizeof(features)))
	return status == raw.PCAN_ERROR_OK && features != 0
}

var _ = pcanChannelOccupied
