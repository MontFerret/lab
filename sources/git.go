package sources

import (
	"context"
	"io/ioutil"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type Git struct {
	url string
}

func NewGit(url string) *Git {
	return &Git{url}
}

func (g Git) Read(ctx context.Context) Stream {
	onFile := make(chan File)
	onError := make(chan error)

	go func() {
		defer func() {
			close(onFile)
			close(onError)
		}()

		r, err := git.CloneContext(ctx, memory.NewStorage(), nil, &git.CloneOptions{
			URL: g.url,
		})

		if err != nil {
			onError <- err

			return
		}

		ref, err := r.Head()

		if err != nil {
			onError <- err

			return
		}

		commit, err := r.CommitObject(ref.Hash())

		if err != nil {
			onError <- err

			return
		}

		files, err := commit.Files()

		if err != nil {
			onError <- err

			return
		}

		err = files.ForEach(func(f *object.File) error {
			if !isFQLFile(f.Name) {
				return nil
			}

			reader, err := f.Reader()

			if err != nil {
				onFile <- File{
					Name:  f.Name,
					Error: err,
				}

				return nil
			}

			defer reader.Close()

			content, err := ioutil.ReadAll(reader)

			if err != nil {
				onFile <- File{
					Name:  f.Name,
					Error: err,
				}

				return nil
			}

			onFile <- File{
				Name:    f.Name,
				Content: content,
				Error:   nil,
			}

			return nil
		})

		if err != nil {
			onError <- err

			return
		}
	}()

	return Stream{
		Files:  onFile,
		Errors: onError,
	}
}
