//go:build linux

package raw

func socketCANInterface(ch TPCANHandle) (string, bool) {
	socketCANRegistry.mu.Lock()
	defer socketCANRegistry.mu.Unlock()

	iface, ok := socketCANRegistry.names[ch]
	return iface, ok
}
