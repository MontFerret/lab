package cdn

import (
	"github.com/pkg/errors"
	"net"
)

func getLocalIPAddress() (string, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return "", err
	}

	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()

		if err != nil {
			return "", err
		}

		// handle err
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)

			if !ok {
				continue
			}

			v4 := ipnet.IP.To4()

			if ipnet.IP.IsLoopback() || v4 == nil {
				continue
			}

			return v4.String(), nil
		}
	}

	return "", errors.New("unable to detect local IP address")
}
