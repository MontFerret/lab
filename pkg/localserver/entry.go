package localserver

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	entryAliasExp = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)
	entryPortExp  = regexp.MustCompile(`:(\d+)$`)
)

type (
	Entry struct {
		Alias string
		Path  string
		Port  int
	}

	Entries []Entry

	EntryParseOptions struct {
		EntryName     string
		AliasName     string
		PortName      string
		DuplicateName string
		DefaultAlias  func(path string) string
	}
)

func ParseEntries(bindings []string, opts EntryParseOptions) (Entries, error) {
	entries := make(Entries, 0, len(bindings))
	aliases := make(map[string]struct{}, len(bindings))

	for _, binding := range bindings {
		entry, err := ParseEntry(binding, opts)
		if err != nil {
			return nil, err
		}

		if _, found := aliases[entry.Alias]; found {
			return nil, errors.Errorf("duplicate %s %q", duplicateName(opts), entry.Alias)
		}

		aliases[entry.Alias] = struct{}{}
		entries = append(entries, entry)
	}

	return entries, nil
}

func ParseEntry(binding string, opts EntryParseOptions) (Entry, error) {
	if binding == "" {
		return Entry{}, errors.Errorf("invalid %s \"\"", entryName(opts))
	}

	pathPart, alias, port, err := SplitEntryBinding(binding, opts)
	if err != nil {
		return Entry{}, err
	}

	if pathPart == "" {
		return Entry{}, errors.Errorf("invalid %s %q", entryName(opts), binding)
	}

	hasExplicitAlias := alias != ""

	if hasExplicitAlias && !IsValidAlias(alias) {
		return Entry{}, errors.Errorf("invalid %s %q", aliasName(opts), alias)
	}

	if !hasExplicitAlias {
		alias = defaultAlias(pathPart, opts)
	}

	if !IsValidAlias(alias) {
		return Entry{}, errors.Errorf("invalid %s %q", aliasName(opts), alias)
	}

	return Entry{
		Alias: alias,
		Path:  pathPart,
		Port:  port,
	}, nil
}

func SplitEntryBinding(binding string, opts EntryParseOptions) (string, string, int, error) {
	base := binding
	port := 0

	if match := entryPortExp.FindStringSubmatch(binding); match != nil {
		value, err := strconv.Atoi(match[1])
		if err != nil {
			return "", "", 0, errors.Wrapf(err, "invalid %s %q", portName(opts), match[1])
		}

		if value <= 0 || value > 65535 {
			return "", "", 0, errors.Errorf("invalid %s %q", portName(opts), match[1])
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

	if entryPortExp.MatchString(pathPart) {
		return "", "", 0, errors.Errorf("invalid %s %q: use <path>@<alias>:<port>", entryName(opts), binding)
	}

	if alias == "" {
		return "", "", 0, errors.Errorf("invalid %s %q", aliasName(opts), alias)
	}

	return pathPart, alias, port, nil
}

func IsValidAlias(alias string) bool {
	return entryAliasExp.MatchString(alias)
}

func defaultAlias(path string, opts EntryParseOptions) string {
	if opts.DefaultAlias != nil {
		return opts.DefaultAlias(path)
	}

	return filepath.Base(filepath.Clean(path))
}

func entryName(opts EntryParseOptions) string {
	if opts.EntryName != "" {
		return opts.EntryName
	}

	return "local server entry"
}

func aliasName(opts EntryParseOptions) string {
	if opts.AliasName != "" {
		return opts.AliasName
	}

	return "local server alias"
}

func portName(opts EntryParseOptions) string {
	if opts.PortName != "" {
		return opts.PortName
	}

	return "local server port"
}

func duplicateName(opts EntryParseOptions) string {
	if opts.DuplicateName != "" {
		return opts.DuplicateName
	}

	return aliasName(opts)
}
