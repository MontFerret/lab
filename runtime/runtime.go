package runtime

import "context"

type (
	Options struct {
		RemoteURL string
		CDP       string
	}

	Runtime interface {
		Run(ctx context.Context, query string, params map[string]interface{}) ([]byte, error)
	}
)

func New(opts Options) Runtime {
	if opts.RemoteURL != "" {
		return NewRemote(opts.RemoteURL)
	}

	return NewNative(opts.CDP)
}
