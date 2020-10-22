package sources

import (
	"context"
	"github.com/hashicorp/go-retryablehttp"
	"io/ioutil"
	"net/url"
)

type HTTP struct {
	url url.URL
}

func NewHTTP(u url.URL) (Source, error) {
	return &HTTP{u}, nil
}

func (src *HTTP) Read(ctx context.Context) (<-chan File, <-chan Error) {
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

		req, err := retryablehttp.NewRequest("GET", src.url.String(), nil)

		if err != nil {
			onError <- NewErrorFrom(src.url.String(), err)

			return
		}

		res, err := retryClient.Do(req.WithContext(ctx))

		if err != nil {
			onError <- NewErrorFrom(src.url.String(), err)

			return
		}

		defer res.Body.Close()

		content, err := ioutil.ReadAll(res.Body)

		if err != nil {
			onError <- NewErrorFrom(src.url.String(), err)

			return
		}

		onNext <- File{
			Name:    src.url.String(),
			Content: content,
		}
	}()

	return onNext, onError
}

func (src *HTTP) Resolve(ctx context.Context, path string) (<-chan File, <-chan Error) {
	// src.url.ResolveReference()

	return nil, nil
}
