package staticserver

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/MontFerret/lab/v2/pkg/localserver"
)

type (
	ServeEntry   = localserver.Entry
	ServeEntries = localserver.Entries
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

	pathPart, alias, port, err := localserver.SplitEntryBinding(binding, serveEntryParseOptions())
	if err != nil {
		return ServeEntry{}, err
	}

	if pathPart == "" {
		return ServeEntry{}, errors.Errorf("invalid serve entry %q", binding)
	}

	hasExplicitAlias := alias != ""
	if hasExplicitAlias && !localserver.IsValidAlias(alias) {
		return ServeEntry{}, errors.Errorf("invalid serve alias %q", alias)
	}

	if err := validateServeEntryPath(pathPart); err != nil {
		return ServeEntry{}, err
	}

	if !hasExplicitAlias {
		alias = filepath.Base(filepath.Clean(pathPart))
	}

	if !localserver.IsValidAlias(alias) {
		return ServeEntry{}, errors.Errorf("invalid serve alias %q", alias)
	}

	return ServeEntry{
		Alias: alias,
		Path:  pathPart,
		Port:  port,
	}, nil
}

func validateServeEntryPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("served directory %q does not exist", path)
		}

		return errors.Wrapf(err, "inspect served directory %q", path)
	}

	if !info.IsDir() {
		return errors.Errorf("served directory %q is not a directory", path)
	}

	return nil
}

func serveEntryParseOptions() localserver.EntryParseOptions {
	return localserver.EntryParseOptions{
		EntryName:     "serve entry",
		AliasName:     "serve alias",
		PortName:      "serve port",
		DuplicateName: "static alias",
		DefaultAlias: func(path string) string {
			return filepath.Base(filepath.Clean(path))
		},
	}
}
