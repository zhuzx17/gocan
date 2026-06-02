package gocan

// DeviceInfo 描述一个 CAN/CAN FD 通道对应的设备信息。
type DeviceInfo struct {
	Channel          Channel
	Name             string
	Backend          ChannelBackend
	HardwareName     string
	InterfaceName    string
	DeviceNumber     uint32
	ControllerNumber uint32
	Features         uint32
	Up               bool
	FD               bool
}

// GetDeviceInfo 查询指定通道的设备信息。
func GetDeviceInfo(ch Channel) (DeviceInfo, error) {
	return getDeviceInfo(ch)
}
