package sources

import (
	"context"
	"io/ioutil"

	"github.com/gobwas/glob"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type Git struct {
	url    string
	filter glob.Glob
}

func NewGit(url string, pattern string) (Source, error) {
	var filter glob.Glob

	if pattern != "" {
		f, err := glob.Compile(pattern)

		if err != nil {
			return nil, err
		}

		filter = f
	}

	return &Git{url, filter}, nil
}

func (g *Git) Read(ctx context.Context) Stream {
	onNext := make(chan File)
	onError := make(chan error)

	go func() {
		defer func() {
			close(onNext)
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
			if !IsSupportedFile(f.Name) {
				return nil
			}

			// if not matched, skip the file
			if g.filter != nil && !g.filter.Match(f.Name) {
				return nil
			}

			reader, err := f.Reader()

			if err != nil {
				onNext <- File{
					Name:  f.Name,
					Error: err,
				}

				return nil
			}

			defer reader.Close()

			content, err := ioutil.ReadAll(reader)

			if err != nil {
				onNext <- File{
					Name:  f.Name,
					Error: err,
				}

				return nil
			}

			onNext <- File{
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

	return NewStream(onNext, onError)
}
