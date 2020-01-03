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
	for res := range stream.Progress {
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
	}

	select {
	case <-ctx.Done():
		break
	case sum := <-stream.Summary:
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
	}

	return nil
}
