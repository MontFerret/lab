package reporters

import (
	"context"
	"github.com/rs/zerolog"
	"io"

	"github.com/MontFerret/lab/runner"
)

type Console struct {
	logger zerolog.Logger
}

func NewConsole(out io.Writer) *Console {
	return &Console{
		zerolog.New(zerolog.ConsoleWriter{Out: out}).With().Timestamp().Logger(),
	}
}

func (c *Console) Report(ctx context.Context, stream runner.Stream) error {
	var done bool
	var err error

	for {
		if done {
			break
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		case res := <-stream.Progress:
			if res.Error != nil {
				c.logger.Error().
					Err(res.Error).
					Str("file", res.File).
					Str("duration", res.Duration.String()).
					Msg("Failed")
			} else {
				c.logger.Info().
					Str("file", res.File).
					Str("duration", res.Duration.String()).
					Msg("Passed")
			}
		case sum := <-stream.Summary:
			var event *zerolog.Event

			if sum.Failed == 0 {
				event = c.logger.Info()
			} else {
				event = c.logger.Error()
			}

			event.
				Timestamp().
				Int("passed", sum.Passed).
				Int("failed", sum.Failed).
				Str("duration", sum.Duration.String()).
				Msg("Completed")

			done = true
		case e := <-stream.Error:
			err = e
			done = true
		}
	}

	return err
}
