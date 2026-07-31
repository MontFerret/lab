package runtime

import (
	"fmt"
	"net/url"

	"strings"
)

// Protocol identifies the command contract implemented by a binary
// runtime.
type Protocol string

const (
	ProtocolUnknown   Protocol = ""
	ProtocolCLI       Protocol = "cli"
	ProtocolCLIDirect Protocol = "cli-direct"
)

// ParseProtocol parses a supported runtime protocol.
func ParseProtocol(value string) (Protocol, error) {
	protocol := Protocol(strings.TrimSpace(strings.ToLower(value)))

	switch protocol {
	case ProtocolCLI, ProtocolCLIDirect:
		return protocol, nil
	default:
		return ProtocolUnknown, fmt.Errorf(
			"unknown runtime protocol %q; expected one of: %s, %s",
			value,
			ProtocolCLI,
			ProtocolCLIDirect,
		)
	}
}

// ParseProtocolFrom parses a supported runtime protocol from a URL query parameter.
func ParseProtocolFrom(u *url.URL, fallback Protocol) (Protocol, error) {
	name := strings.TrimSpace(u.Query().Get("protocol"))

	if name != "" {
		return ParseProtocol(name)
	}

	return fallback, nil
}

func (protocol Protocol) unsupported(feature string) error {
	return fmt.Errorf("runtime protocol %q does not support %s", protocol, feature)
}
