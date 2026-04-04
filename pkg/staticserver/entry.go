package staticserver

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	serveAliasExp = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)
	servePortExp  = regexp.MustCompile(`:(\d+)$`)
)

type (
	ServeEntry struct {
		Alias string
		Path  string
		Port  int
	}

	ServeEntries []ServeEntry
)

func ParseServeEntries(bindings []string) (ServeEntries, error) {
	entries := make(ServeEntries, 0, len(bindings))
	aliases := make(map[string]struct{}, len(bindings))

	for _, binding := range bindings {
		entry, err := ParseServeEntry(binding)
		if err != nil {
			return nil, err
		}

		if _, found := aliases[entry.Alias]; found {
			return nil, errors.Errorf("duplicate static alias %q", entry.Alias)
		}

		aliases[entry.Alias] = struct{}{}
		entries = append(entries, entry)
	}

	return entries, nil
}

func ParseServeEntry(binding string) (ServeEntry, error) {
	if binding == "" {
		return ServeEntry{}, errors.New("invalid serve entry \"\"")
	}

	pathPart, alias, port, err := splitServeBinding(binding)
	if err != nil {
		return ServeEntry{}, err
	}

	if pathPart == "" {
		return ServeEntry{}, errors.Errorf("invalid serve entry %q", binding)
	}

	hasExplicitAlias := alias != ""

	if hasExplicitAlias && !serveAliasExp.MatchString(alias) {
		return ServeEntry{}, errors.Errorf("invalid serve alias %q", alias)
	}

	info, err := os.Stat(pathPart)
	if err != nil {
		if os.IsNotExist(err) {
			return ServeEntry{}, errors.Errorf("served directory %q does not exist", pathPart)
		}

		return ServeEntry{}, errors.Wrapf(err, "inspect served directory %q", pathPart)
	}

	if !info.IsDir() {
		return ServeEntry{}, errors.Errorf("served directory %q is not a directory", pathPart)
	}

	if !hasExplicitAlias {
		alias = filepath.Base(filepath.Clean(pathPart))
	}

	if !serveAliasExp.MatchString(alias) {
		return ServeEntry{}, errors.Errorf("invalid serve alias %q", alias)
	}

	return ServeEntry{
		Alias: alias,
		Path:  pathPart,
		Port:  port,
	}, nil
}

func splitServeBinding(binding string) (string, string, int, error) {
	base := binding
	port := 0

	if match := servePortExp.FindStringSubmatch(binding); match != nil {
		value, err := strconv.Atoi(match[1])
		if err != nil {
			return "", "", 0, errors.Wrapf(err, "invalid serve port %q", match[1])
		}

		if value <= 0 || value > 65535 {
			return "", "", 0, errors.Errorf("invalid serve port %q", match[1])
		}

		port = value
		base = strings.TrimSuffix(binding, match[0])
	}

	aliasIdx := strings.LastIndex(base, "@")
	if aliasIdx < 0 {
		return base, "", port, nil
	}

	pathPart := base[:aliasIdx]
	alias := base[aliasIdx+1:]

	if servePortExp.MatchString(pathPart) {
		return "", "", 0, errors.Errorf("invalid serve entry %q: use <path>@<alias>:<port>", binding)
	}

	if alias == "" {
		return "", "", 0, errors.Errorf("invalid serve alias %q", alias)
	}

	return pathPart, alias, port, nil
}
