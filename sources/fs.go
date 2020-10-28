package sources

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

type FileSystem struct {
	dir    string
	name   string
	filter glob.Glob
}

func NewFileSystem(u url.URL) (Source, error) {
	pattern := u.Query().Get("filter")

	var filter glob.Glob

	if pattern != "" {
		f, err := glob.Compile(pattern)

		if err != nil {
			return nil, err
		}

		filter = f
	}

	fullPath := filepath.Join(u.Host, u.Path)

	if !filepath.IsAbs(fullPath) {
		fp, err := filepath.Abs(fullPath)

		if err != nil {
			return nil, errors.Wrap(err, "get absolute path")
		}

		fullPath = fp
	}

	var dir string
	var name string

	// file
	if filepath.Ext(fullPath) != "" {
		dir = filepath.Dir(fullPath)
		name = filepath.Base(fullPath)
	} else {
		dir = fullPath
	}

	return &FileSystem{dir, name, filter}, nil
}

func (fs *FileSystem) Read(ctx context.Context) (<-chan File, <-chan Error) {
	onNext := make(chan File)
	onError := make(chan Error)

	go func() {
		fs.traverse(ctx, filepath.Join(fs.dir, fs.name), onNext, onError)

		close(onNext)
		close(onError)
	}()

	return onNext, onError
}

func (fs *FileSystem) Resolve(ctx context.Context, u url.URL) (<-chan File, <-chan Error) {
	onNext := make(chan File)
	onError := make(chan Error)

	go func() {
		defer func() {
			close(onNext)
			close(onError)
		}()

		fp, err := filepath.Abs(filepath.Join(fs.dir, filepath.Join(u.Host, u.Path)))

		if err != nil {
			onError <- NewErrorFrom(u.String(), err)

			return
		}

		fs.traverse(ctx, fp, onNext, onError)
	}()

	return onNext, onError
}

func (fs *FileSystem) traverse(ctx context.Context, path string, onNext chan<- File, onError chan<- Error) {
	fi, err := os.Stat(path)

	if err != nil {
		onError <- NewErrorFrom(path, err)

		return
	}

	// if only a single file was given
	if fi.Mode().IsRegular() {
		filename := path

		if !IsSupportedFile(path) {
			onError <- NewError(path, "invalid file")
		}

		// if not matched, skip the file
		if fs.filter != nil && !fs.filter.Match(filename) {
			return
		}

		fs.readFile(filename, onNext, onError)

		return
	}

	files, err := ioutil.ReadDir(path)

	if err != nil {
		onError <- NewErrorFrom(path, err)

		return
	}

	for _, file := range files {
		filename := filepath.Join(path, file.Name())

		if file.IsDir() {
			fs.traverse(ctx, filename, onNext, onError)

			continue
		}

		if !IsSupportedFile(file.Name()) {
			continue
		}

		// if not matched, skip the file
		if fs.filter != nil && !fs.filter.Match(filename) {
			continue
		}

		fs.readFile(filename, onNext, onError)
	}
}

func (fs *FileSystem) readFile(filename string, onNext chan<- File, onError chan<- Error) {
	content, err := ioutil.ReadFile(filename)

	if err != nil {
		onError <- NewErrorFrom(filename, err)

		return
	}

	onNext <- File{
		Source:  fs,
		Name:    filename,
		Content: content,
	}
}
