package cdn

import (
	"net"
)

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")

	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)

	if err != nil {
		return 0, err
	}

	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}

func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")

	if err != nil {
		return "", err
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}
