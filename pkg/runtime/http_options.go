package runtime

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"

	"golang.org/x/net/http/httpguts"
)

type httpOptions struct {
	Headers http.Header
	Path    string
	Cookies []http.Cookie
}

func newHTTPOptions(target string, options map[string]any) (httpOptions, error) {
	config := httpOptions{Headers: http.Header{
		"Content-Type":    {"application/json"},
		"Accept":          {"*/*"},
		"Accept-Charset":  {"utf-8"},
		"Accept-Encoding": {"gzip", "deflate"},
		"Cache-Control":   {"no-cache"},
	}}

	if err := validateAdapterOptionNames(options, "HTTP", "headers", "cookies", "path"); err != nil {
		return config, err
	}

	if value, exists := options["path"]; exists {
		path, ok := value.(string)
		if !ok {
			return config, errors.New("runtime option \"path\" must be a JSON string")
		}

		reference, err := url.Parse(path)
		if err != nil {
			return config, errors.New("runtime option \"path\" must be a valid URL reference")
		}

		if reference.Scheme != "" && reference.Scheme != "http" && reference.Scheme != "https" {
			return config, errors.New("runtime option \"path\" must resolve to an HTTP/HTTPS URL")
		}

		if reference.Scheme != "" || reference.Host != "" {
			if err := validateHTTPURL(reference); err != nil {
				return config, errors.New("runtime option \"path\" must resolve to a valid HTTP/HTTPS URL")
			}
		}

		config.Path = path
	}

	for _, name := range []string{"headers", "cookies"} {
		value, exists := options[name]
		if !exists {
			continue
		}

		object, ok := value.(map[string]any)
		if !ok {
			return config, fmt.Errorf("runtime option %q must be a JSON object of strings", name)
		}

		for _, key := range slices.Sorted(maps.Keys(object)) {
			text, ok := object[key].(string)
			if !ok {
				return config, fmt.Errorf("runtime option %q entry %q must be a string", name, key)
			}

			if name == "headers" {
				if !httpguts.ValidHeaderFieldName(key) || !httpguts.ValidHeaderFieldValue(text) {
					return config, fmt.Errorf("runtime option \"headers\" entry %q is not a valid HTTP header", key)
				}

				config.Headers.Add(key, text)

				continue
			}

			cookie := http.Cookie{Name: key, Value: text}
			if err := cookie.Valid(); err != nil {
				return config, fmt.Errorf("runtime option \"cookies\" entry %q is not a valid HTTP cookie", key)
			}

			cookie.Domain, cookie.Path = target, config.Path
			config.Cookies = append(config.Cookies, cookie)
		}
	}

	return config, nil
}
