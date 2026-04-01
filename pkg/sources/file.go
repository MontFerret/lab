package sources

import (
	"context"
	"net/url"
)

type File struct {
	Source  Source
	Name    string
	Content []byte
}

func (f File) Resolve(ctx context.Context, u *url.URL) (onNext <-chan File, onError <-chan Error) {
	q := u.Query()
	q.Set("from", f.Name)
	u.RawQuery = q.Encode()

	return f.Source.Resolve(ctx, u)
}
