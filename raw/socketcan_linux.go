//go:build linux

package raw

func socketCANInterface(ch TPCANHandle) (string, bool) {
	return SocketCANInterface(ch)
}
