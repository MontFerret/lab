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

func NewFileSystem(path string) (*FileSystem, error) {
	return &FileSystem{
		path: path,
	}, nil
}

func (fs *FileSystem) SetFilter(pattern string) error {
	filter, err := glob.Compile(pattern)

	if err != nil {
		return err
	}

	fs.filter = filter

	return nil
}

func (fs *FileSystem) Read(ctx context.Context) Stream {
	onFile := make(chan File)
	onError := make(chan error)

	go func() {
		if err := fs.traverse(ctx, fs.path, onFile); err != nil {
			onError <- err
		}

		close(onFile)
		close(onError)
	}()

	return Stream{
		Files:  onFile,
		Errors: onError,
	}
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

		// if not matched, skip the file
		if fs.filter != nil && !fs.filter.Match(filename) {
			return nil
		}

		if !isFQLFile(path) {
			onFile <- File{
				Name:    filename,
				Content: nil,
				Error:   errors.New("invalid file"),
			}
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

		// if not matched, skip the file
		if fs.filter != nil && !fs.filter.Match(filename) {
			continue
		}

		if file.IsDir() {
			if err := fs.traverse(ctx, filename, onFile); err != nil {
				return err
			}

			continue
		}

		if !isFQLFile(file.Name()) {
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
			Name:    filename,
			Content: nil,
			Error:   err,
		}
	}

	return File{
		Name:    filename,
		Content: content,
	}
}
