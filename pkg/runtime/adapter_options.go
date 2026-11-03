package runtime

import (
	"errors"
	"fmt"
	"maps"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

func validateAdapterOptionNames(options map[string]any, adapter string, supported ...string) error {
	for _, name := range slices.Sorted(maps.Keys(options)) {
		if !slices.Contains(supported, name) {
			accepted := "no adapter options"

			if len(supported) > 0 {
				accepted = strings.Join(supported, ", ")
			}

			return fmt.Errorf("runtime option %q is not supported by %s runtimes (supported: %s); use --param for FQL inputs", name, adapter, accepted)
		}
	}

	return nil
}

func validateHTTPURL(u *url.URL) error {
	if u.Opaque != "" || u.Hostname() == "" || strings.HasSuffix(u.Host, ":") {
		return errors.New("invalid HTTP runtime target: expected an HTTP/HTTPS URL with a host and optional port")
	}

	if strings.Contains(u.Host, ":") && !strings.HasPrefix(u.Host, "[") {
		if _, _, err := net.SplitHostPort(u.Host); err != nil {
			return errors.New("invalid HTTP runtime target: IPv6 addresses must be bracketed")
		}
	}

	if port := u.Port(); port != "" {
		number, err := strconv.ParseUint(port, 10, 16)
		if err != nil || number == 0 {
			return errors.New("invalid HTTP runtime target: port must be from 1 to 65535")
		}
	}

	return nil
}

func wireConnectTimeout(options map[string]any) (time.Duration, error) {
	if err := validateAdapterOptionNames(options, "Wire", "connect-timeout"); err != nil {
		return 0, err
	}

	value, exists := options["connect-timeout"]
	if !exists {
		return DefaultConnectTimeout, nil
	}

	text, ok := value.(string)
	if !ok {
		return 0, errors.New("runtime option \"connect-timeout\" must be a JSON duration string, such as \"5s\"")
	}

	timeout, err := time.ParseDuration(text)
	if err != nil || timeout <= 0 {
		return 0, errors.New("runtime option \"connect-timeout\" must be a positive duration, such as \"5s\"")
	}

	return timeout, nil
}
