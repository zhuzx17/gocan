//go:build !windows && !linux

package gocan

func lookupChannels() ([]ChannelInfo, error) {
	return []ChannelInfo{}, nil
}
