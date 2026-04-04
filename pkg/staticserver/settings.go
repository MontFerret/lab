package staticserver

import (
	"net"
	"strings"

	"github.com/pkg/errors"
)

const defaultHost = "127.0.0.1"

type Settings struct {
	BindHost      string
	AdvertiseHost string
}

func ResolveSettings(settings Settings) (Settings, error) {
	bindHost, err := normalizeHost(settings.BindHost, "serve-bind")
	if err != nil {
		return Settings{}, err
	}

	advertiseHost, err := normalizeHost(settings.AdvertiseHost, "serve-host")
	if err != nil {
		return Settings{}, err
	}

	if advertiseHost == "" {
		advertiseHost = defaultHost
	}

	if bindHost == "" {
		if settings.AdvertiseHost != "" {
			bindHost = wildcardBindHost(advertiseHost)
		} else {
			bindHost = defaultHost
		}
	}

	return Settings{
		BindHost:      bindHost,
		AdvertiseHost: advertiseHost,
	}, nil
}

func normalizeHost(value, option string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	original := value

	if host, port, err := net.SplitHostPort(value); err == nil {
		if port != "" {
			return "", errors.Errorf("invalid %s %q: must not include a port", option, original)
		}

		value = host
	}

	if strings.HasPrefix(value, "[") || strings.HasSuffix(value, "]") {
		if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			value = strings.TrimSuffix(strings.TrimPrefix(value, "["), "]")
		} else {
			return "", errors.Errorf("invalid %s %q", option, original)
		}
	}

	if value == "" {
		return "", errors.Errorf("invalid %s %q", option, original)
	}

	if strings.Count(value, ":") == 1 {
		return "", errors.Errorf("invalid %s %q: must not include a port", option, original)
	}

	return value, nil
}

func wildcardBindHost(host string) string {
	if ip := net.ParseIP(host); ip != nil && ip.To4() == nil {
		return "::"
	}

	return "0.0.0.0"
}
