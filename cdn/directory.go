package cdn

import (
	"github.com/pkg/errors"
	"path/filepath"
	"regexp"
	"strconv"
)

var bindingExp = regexp.MustCompile("^(?P<path>[A-Za-z0-9./\\\\]+)(?P<port>:\\d+)?(?P<name>@\\w+)?")

type Directory struct {
	Name string
	Path string
	Port int
}

func NewDirectoryFrom(binding string) (Directory, error) {
	matches := bindingExp.FindStringSubmatch(binding)

	if matches == nil {
		return Directory{}, errors.Errorf(
			"invalid directory binding format. expected binding in format path{:port?}{@name?}, but got %s",
			binding,
		)
	}

	path := matches[1]
	portStr := matches[2]
	name := matches[3]
	var port int

	if portStr != "" {
		portStr = portStr[1:] // remove ':' character
		p, err := strconv.Atoi(portStr)

		if err != nil {
			return Directory{}, err
		}

		port = p
	} else {
		p, err := GetFreePort()

		if err != nil {
			return Directory{}, err
		}

		port = p
	}

	if name == "" {
		name = filepath.Base(path)

		if name == "" {
			name = "default"
		}
	} else {
		name = name[1:] // remove '@' character
	}

	return Directory{
		Path: path,
		Name: name,
		Port: port,
	}, nil
}
