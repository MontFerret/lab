package sources

import (
	"context"
	"errors"
	"io/ioutil"
	"net/url"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gobwas/glob"
)

type Git struct {
	mu     sync.Mutex
	repo   *git.Repository
	url    *url.URL
	filter glob.Glob
}

func NewGit(u *url.URL) (Source, error) {
	var filter glob.Glob

	pattern := u.Query().Get("filter")

	if pattern != "" {
		f, err := glob.Compile(pattern)

		if err != nil {
			return nil, err
		}

		filter = f
	}

	src := new(Git)
	src.url = u
	src.filter = filter

	return src, nil
}

func NewGitFrom(repo *git.Repository, filter glob.Glob) (Source, error) {
	if repo == nil {
		return nil, errors.New("missed repo")
	}

	src := new(Git)
	src.repo = repo
	src.filter = filter

	return src, nil
}

func (g *Git) Read(ctx context.Context) (<-chan File, <-chan Error) {
	onNext := make(chan File)
	onError := make(chan Error)

	go func() {
		defer func() {
			close(onNext)
			close(onError)
		}()

		commit, err := g.getCommit(ctx)

		if err != nil {
			onError <- NewErrorFrom(g.url.String(), err)

			return
		}

		files, err := commit.Files()

		if err != nil {
			onError <- NewErrorFrom(g.url.String(), err)

			return
		}

		defer files.Close()

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
				onError <- NewErrorFrom(f.Name, err)

				return nil
			}

			defer reader.Close()

			content, err := ioutil.ReadAll(reader)

			if err != nil {
				onError <- NewErrorFrom(f.Name, err)

				return nil
			}

			onNext <- File{
				Name:    f.Name,
				Content: content,
			}

			return nil
		})

		if err != nil {
			onError <- NewErrorFrom(g.url.String(), err)

			return
		}
	}()

	return onNext, onError
}

func (g *Git) Resolve(ctx context.Context, fileName string) (<-chan File, <-chan Error) {
	onNext := make(chan File)
	onError := make(chan Error)

	go func() {
		defer func() {
			close(onNext)
			close(onError)
		}()

		commit, err := g.getCommit(ctx)

		if err != nil {
			onError <- NewErrorFrom(fileName, err)

			return
		}

		file, err := commit.File(fileName)

		if err != nil {
			onError <- NewErrorFrom(fileName, err)

			return
		}

		contents, err := file.Contents()

		if err != nil {
			onError <- NewErrorFrom(fileName, err)

			return
		}

		onNext <- File{
			Name:    file.Name,
			Content: []byte(contents),
		}
	}()

	return onNext, onError
}

func (g *Git) getCommit(ctx context.Context) (*object.Commit, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.repo == nil {
		r, err := git.CloneContext(ctx, memory.NewStorage(), nil, &git.CloneOptions{
			URL: g.url.String(),
		})

		if err != nil {
			return nil, err
		}

		g.repo = r
	}

	ref, err := g.repo.Head()

	if err != nil {
		return nil, err
	}

	return g.repo.CommitObject(ref.Hash())
}
