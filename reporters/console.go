package reporters

import (
	"context"
	"errors"
	"io"

	"github.com/hako/durafmt"
	"github.com/rs/zerolog"

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
		var evt *zerolog.Event

		if res.Error != nil {
			evt = c.logger.Error().Err(res.Error)
		} else {
			evt = c.logger.Info()
		}

		evt = evt.
			Str("File", res.Filename).
			Str("Duration", durafmt.ParseShort(res.Duration).InternationalString()).
			Uint64("Attempts", res.Attempts).
			Uint64("Times", res.Times)

		if res.Error != nil {
			evt.Msg("Failed")
		} else {
			evt.Msg("Passed")
		}
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	case sum := <-stream.Summary:
		var event *zerolog.Event

		if !sum.HasErrors() {
			event = c.logger.Info()
		} else {
			event = c.logger.Error()
		}

		event.
			Int("Passed", sum.Passed).
			Int("Failed", sum.Failed).
			Str("Duration", durafmt.ParseShort(sum.Duration).InternationalString()).
			Msg("Done")

		if sum.HasErrors() {
			return errors.New("has errors")
		}

		return nil
	}
}
