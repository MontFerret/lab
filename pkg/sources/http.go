package sources

import (
	"context"
	"io"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
)

type HTTP struct {
	url *url.URL
}

func NewHTTP(u *url.URL) (Source, error) {
	return &HTTP{u}, nil
}

func (src *HTTP) Read(ctx context.Context) (<-chan File, <-chan Error) {
	return src.call(ctx, src.url)
}

func (src *HTTP) Resolve(ctx context.Context, u *url.URL) (<-chan File, <-chan Error) {
	return src.call(ctx, u)
}

func (src *HTTP) call(ctx context.Context, u *url.URL) (<-chan File, <-chan Error) {
	onNext := make(chan File)
	onError := make(chan Error)

	go func() {
		defer func() {
			close(onNext)
			close(onError)
		}()

		retryClient := retryablehttp.NewClient()
		retryClient.RetryMax = 10
		retryClient.Logger = nil

		req, err := retryablehttp.NewRequest("GET", u.String(), nil)

		if err != nil {
			onError <- NewErrorFrom(u.String(), err)

			return
		}

		res, err := retryClient.Do(req.WithContext(ctx))

		if err != nil {
			onError <- NewErrorFrom(u.String(), err)

			return
		}

		defer res.Body.Close()

		content, err := io.ReadAll(res.Body)

		if err != nil {
			onError <- NewErrorFrom(u.String(), err)

			return
		}

		onNext <- File{
			Source:  src,
			Name:    u.String(),
			Content: content,
		}
	}()

	return onNext, onError
}
