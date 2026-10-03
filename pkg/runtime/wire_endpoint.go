package runtime

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// DefaultConnectTimeout bounds Wire transport establishment and its logical handshake.
const DefaultConnectTimeout = 5 * time.Second

func isWireType(name string) bool {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "") == "wire"
}

func wireAddress(endpoint string) (string, error) {
	const prefix = "tcp://127.0.0.1:"
	if endpoint == "" {
		return "", fmt.Errorf("wire runtime requires --runtime-endpoint tcp://127.0.0.1:<port>")
	}

	invalid := func() (string, error) {
		return "", fmt.Errorf("invalid Wire endpoint %q: use tcp://127.0.0.1:<port> with a port from 1 to 65535", endpoint)
	}

	if !strings.HasPrefix(endpoint, prefix) {
		return invalid()
	}

	port := strings.TrimPrefix(endpoint, prefix)
	if port == "" {
		return invalid()
	}

	for _, char := range port {
		if char < '0' || char > '9' {
			return invalid()
		}
	}

	number, err := strconv.ParseUint(port, 10, 16)
	if err != nil || number == 0 {
		return invalid()
	}

	return net.JoinHostPort("127.0.0.1", strconv.FormatUint(number, 10)), nil
}

func validateWireOptions(opts Options) error {
	if opts.FSPolicy.hasSettings() {
		return errors.New("filesystem policy options are not supported by Wire runtimes; configure the runtime host")
	}

	if opts.HTTPPolicy.hasSettings() {
		return errors.New("HTTP policy options are not supported by Wire runtimes; configure the runtime host")
	}

	if opts.BinaryFlags != nil {
		return errors.New("binary flags are only supported by binary runtimes")
	}

	for _, key := range []string{"headers", "cookies", "path", "flags"} {
		if _, exists := opts.Params[key]; exists {
			return fmt.Errorf("runtime parameter %q is not supported by Wire runtimes", key)
		}
	}

	if _, err := wireAddress(opts.Endpoint); err != nil {
		return err
	}

	if opts.ConnectTimeout <= 0 {
		return errors.New("--runtime-connect-timeout must be positive")
	}

	return nil
}
