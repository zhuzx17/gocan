//go:build linux

package gocan

import (
	"net"
	"sort"
	"strings"
)

func lookupChannels() ([]ChannelInfo, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	channels := make([]ChannelInfo, 0)
	for _, iface := range interfaces {
		if !isSocketCANInterface(iface.Name) {
			continue
		}
		channels = append(channels, ChannelInfo{
			Channel: SocketCAN(iface.Name),
			Name:    iface.Name,
			Backend: BackendSocketCAN,
			Up:      iface.Flags&net.FlagUp != 0,
			FD:      true,
		})
	}
	sort.Slice(channels, func(i, j int) bool { return channels[i].Name < channels[j].Name })
	return channels, nil
}

func isSocketCANInterface(name string) bool {
	return strings.HasPrefix(name, "can") || strings.HasPrefix(name, "vcan") ||
		strings.HasPrefix(name, "slcan") || strings.HasPrefix(name, "vxcan")
}
