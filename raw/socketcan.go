package raw

import "sync"

const socketCANHandleBase TPCANHandle = 0xA000

var socketCANRegistry = struct {
	mu     sync.Mutex
	next   TPCANHandle
	byName map[string]TPCANHandle
	names  map[TPCANHandle]string
}{
	next:   socketCANHandleBase,
	byName: make(map[string]TPCANHandle),
	names:  make(map[TPCANHandle]string),
}

// SocketCANHandle 返回 Linux SocketCAN 网络接口对应的通道句柄。
func SocketCANInterface(ch TPCANHandle) (string, bool) {
	socketCANRegistry.mu.Lock()
	defer socketCANRegistry.mu.Unlock()

	iface, ok := socketCANRegistry.names[ch]
	return iface, ok
}

func SocketCANHandle(iface string) TPCANHandle {
	if iface == "" {
		return PCAN_NONEBUS
	}

	socketCANRegistry.mu.Lock()
	defer socketCANRegistry.mu.Unlock()

	if ch, ok := socketCANRegistry.byName[iface]; ok {
		return ch
	}
	ch := socketCANRegistry.next
	socketCANRegistry.next++
	socketCANRegistry.byName[iface] = ch
	socketCANRegistry.names[ch] = iface
	return ch
}
