package apiruntime

import (
	"context"

	"github.com/MontFerret/api"
)

// Output supplies encoded data and an optional terminal failure to scripted runtimes.
// Nil Data represents absent content. An empty non-nil slice represents present content.
type Output struct {
	Data        []byte
	ContentType string
	Err         error
	ConsumeFunc func(context.Context, api.Consumer) error
	CloseFunc   func() error
}

func (o *Output) Metadata() api.Metadata {
	return api.Metadata{ContentType: o.ContentType, Length: int64(len(o.Data)), LengthKnown: true}
}

func (o *Output) Consume(ctx context.Context, consumer api.Consumer) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if o.ConsumeFunc != nil {
		return o.ConsumeFunc(ctx, consumer)
	}

	if o.Data != nil {
		if err := consumer(ctx, o.Data); err != nil {
			return err
		}
	}

	return o.Err
}

func (o *Output) Collect(ctx context.Context) (*api.Content, error) {
	var content *api.Content
	err := o.Consume(ctx, func(_ context.Context, data []byte) error {
		if content == nil {
			content = &api.Content{Metadata: o.Metadata()}
		}

		content.Data = append(content.Data, data...)

		return nil
	})

	return content, err
}

func (o *Output) Close() error {
	if o.CloseFunc != nil {
		return o.CloseFunc()
	}

	return nil
}
