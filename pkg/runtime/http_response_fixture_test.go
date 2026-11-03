package runtime

import (
	"errors"
	"io"
	"net/http"
)

type (
	responseTransport struct {
		fn func(*http.Request) (*http.Response, error)
	}

	responseBody struct {
		reader   io.Reader
		readErr  error
		closeErr error
		closes   int
	}
)

func (t responseTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.fn(r)
}

func (b *responseBody) Read(p []byte) (int, error) {
	n, err := b.reader.Read(p)
	if err == io.EOF && b.readErr != nil {
		return n, b.readErr
	}

	return n, err
}

func (b *responseBody) Close() error {
	b.closes++

	if closer, ok := b.reader.(io.Closer); ok {
		return errors.Join(b.closeErr, closer.Close())
	}

	return b.closeErr
}
