package runtime

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"time"
)

// DefaultConnectTimeout bounds Wire transport establishment and its logical handshake.
const DefaultConnectTimeout = 5 * time.Second

func wireAddress(endpoint string) (string, error) {
	const prefix = "tcp://127.0.0.1:"
	invalid := errors.New("invalid Wire target: use wire+tcp://127.0.0.1:<port> with a port from 1 to 65535")

	if !strings.HasPrefix(endpoint, prefix) {
		return "", invalid
	}

	port := strings.TrimPrefix(endpoint, prefix)
	if port == "" {
		return "", invalid
	}

	for _, char := range port {
		if char < '0' || char > '9' {
			return "", invalid
		}
	}

	number, err := strconv.ParseUint(port, 10, 16)
	if err != nil || number == 0 {
		return "", invalid
	}

	return net.JoinHostPort("127.0.0.1", strconv.FormatUint(number, 10)), nil
}
