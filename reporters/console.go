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

	for !done {
		select {
		case <-ctx.Done():
			return context.Canceled
		case res, ok := <-stream.Progress:
			if !ok {
				done = true
				break
			}

			if res.Error != nil {
				c.logger.Error().
					Err(res.Error).
					Str("File", res.Filename).
					Str("Duration", res.Duration.String()).
					Msg("Failed")
			} else {
				c.logger.Info().
					Str("File", res.Filename).
					Str("Duration", res.Duration.String()).
					Msg("Passed")
			}
		case sum, ok := <-stream.Summary:
			if !ok {
				done = true
				break
			}

			var event *zerolog.Event

			if sum.Failed == 0 && sum.Error == nil {
				event = c.logger.Info()
			} else {
				event = c.logger.Error()
			}

			event.
				Timestamp().
				Int("Passed", sum.Passed).
				Int("Failed", sum.Failed).
				Str("Duration", sum.Duration.String())

			if sum.Error != nil {
				event = event.Str("Error", sum.Error.Error())
			}

			event.Msg("Done")

			done = true
		case e := <-stream.Error:
			err = e
			done = true
		}
	}

	return err
}
