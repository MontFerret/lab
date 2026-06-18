package mockserver

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/MontFerret/lab/v2/pkg/localserver"
)

type (
	Entry   = localserver.Entry
	Entries = localserver.Entries
)

func ParseEntries(bindings []string) (Entries, error) {
	entries, err := localserver.ParseEntries(bindings, entryParseOptions())
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if err := validateSpecPath(entry.Path); err != nil {
			return nil, err
		}
	}

	return entries, nil
}

func ParseEntry(binding string) (Entry, error) {
	entry, err := localserver.ParseEntry(binding, entryParseOptions())
	if err != nil {
		return Entry{}, err
	}

	if err := validateSpecPath(entry.Path); err != nil {
		return Entry{}, err
	}

	return entry, nil
}

func validateSpecPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("mock API spec %q does not exist", path)
		}

		return errors.Wrapf(err, "inspect mock API spec %q", path)
	}

	if info.IsDir() {
		return errors.Errorf("mock API spec %q is not a file", path)
	}

	return nil
}

func entryParseOptions() localserver.EntryParseOptions {
	return localserver.EntryParseOptions{
		EntryName:     "mock API entry",
		AliasName:     "mock API alias",
		PortName:      "mock API port",
		DuplicateName: "mock API alias",
		DefaultAlias: func(path string) string {
			base := filepath.Base(filepath.Clean(path))
			ext := filepath.Ext(base)

			return strings.TrimSuffix(base, ext)
		},
	}
}
