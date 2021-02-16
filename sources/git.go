package sources

import (
	"context"
	"errors"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gobwas/glob"
)

type Git struct {
	mu     sync.Mutex
	repo   *git.Repository
	url    string
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
	src.filter = filter

	switch u.Scheme {
	case "git+https":
		u.Scheme = "https"
		break
	case "git+http":
		u.Scheme = "http"
		break
	default:
		break
	}

	src.url = u.String()

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
			onError <- NewErrorFrom(g.url, err)

			return
		}

		files, err := commit.Files()

		if err != nil {
			onError <- NewErrorFrom(g.url, err)

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
				Source:  g,
				Name:    f.Name,
				Content: content,
			}

			return nil
		})

		if err != nil {
			onError <- NewErrorFrom(g.url, err)

			return
		}
	}()

	return onNext, onError
}

func (g *Git) Resolve(ctx context.Context, u *url.URL) (<-chan File, <-chan Error) {
	onNext := make(chan File)
	onError := make(chan Error)

	go func() {
		defer func() {
			close(onNext)
			close(onError)
		}()

		commit, err := g.getCommit(ctx)

		if err != nil {
			onError <- NewErrorFrom(u.String(), err)

			return
		}

		from := u.Query().Get("from")

		filename := filepath.Join(filepath.Join(ToDir(from), filepath.Join(u.Host, u.Path)))
		file, err := commit.File(filename)

		if err != nil {
			onError <- NewErrorFrom(filename, err)

			return
		}

		contents, err := file.Contents()

		if err != nil {
			onError <- NewErrorFrom(filename, err)

			return
		}

		onNext <- File{
			Source:  g,
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
			URL: g.url,
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
