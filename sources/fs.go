package sources

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

type FileSystem struct {
	path   string
	filter glob.Glob
}

func NewFileSystem(path string, pattern string) (Source, error) {
	var filter glob.Glob

	if pattern != "" {
		f, err := glob.Compile(pattern)

		if err != nil {
			return nil, err
		}

		filter = f
	}

	fullPath := path

	if !filepath.IsAbs(path) {
		fp, err := filepath.Abs(path)

		if err != nil {
			return nil, errors.Wrap(err, "get absolute path")
		}

		fullPath = fp
	}

	return &FileSystem{fullPath, filter}, nil
}

func (fs *FileSystem) Read(ctx context.Context) Stream {
	onNext := make(chan File)
	onError := make(chan error)

	go func() {
		if err := fs.traverse(ctx, fs.path, onNext); err != nil {
			onError <- err
		}

		close(onNext)
		close(onError)
	}()

	return NewStream(onNext, onError)
}

func (fs *FileSystem) traverse(ctx context.Context, path string, onFile chan<- File) error {
	fi, err := os.Stat(path)

	if err != nil {
		onFile <- File{
			Name:    path,
			Content: nil,
			Error:   err,
		}

		return nil
	}

	// if only a single file was given
	if fi.Mode().IsRegular() {
		filename := path

		if !IsSupportedFile(path) {
			onFile <- File{
				Name:    "file://" + filename,
				Content: nil,
				Error:   errors.New("invalid file"),
			}
		}

		// if not matched, skip the file
		if fs.filter != nil && !fs.filter.Match(filename) {
			return nil
		}

		onFile <- fs.readFile(filename)

		return nil
	}

	files, err := ioutil.ReadDir(path)

	if err != nil {
		return err
	}

	for _, file := range files {
		filename := filepath.Join(path, file.Name())

		if file.IsDir() {
			if err := fs.traverse(ctx, filename, onFile); err != nil {
				return err
			}

			continue
		}

		if !IsSupportedFile(file.Name()) {
			continue
		}

		// if not matched, skip the file
		if fs.filter != nil && !fs.filter.Match(filename) {
			continue
		}

		onFile <- fs.readFile(filename)
	}

	return nil
}

func (fs *FileSystem) readFile(filename string) File {
	content, err := ioutil.ReadFile(filename)

	if err != nil {
		return File{
			Name:    "file://" + filename,
			Content: nil,
			Error:   err,
		}
	}

	return File{
		Name:    "file://" + filename,
		Content: content,
	}
}
