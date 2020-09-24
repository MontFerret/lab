package sources

import (
	"context"
	"io/ioutil"

	"github.com/hashicorp/go-retryablehttp"
)

type HTTP struct {
	url string
}

func NewHTTP(url string) (Source, error) {
	return &HTTP{url}, nil
}

func (src *HTTP) Read(ctx context.Context) Stream {
	onNext := make(chan File)
	onError := make(chan error)

	go func() {
		defer func() {
			close(onNext)
			close(onError)
		}()

		retryClient := retryablehttp.NewClient()
		retryClient.RetryMax = 10
		retryClient.Logger = nil

		res, err := retryClient.Get(src.url)

		if err != nil {
			onError <- err

			return
		}

		defer res.Body.Close()

		content, err := ioutil.ReadAll(res.Body)

		if err != nil {
			onNext <- File{
				Name:  src.url,
				Error: err,
			}

			return
		}

		onNext <- File{
			Name:    src.url,
			Content: content,
		}
	}()

	return NewStream(onNext, onError)
}
