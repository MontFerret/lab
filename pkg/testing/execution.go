package testing

import (
	"context"
	"errors"
)

func discardOutput(context.Context, []byte) error {
	return nil
}

func hasCleanupFailure(err error) bool {
	var failure *cleanupFailure

	return errors.As(err, &failure)
}

// A joined capability error must not hide an independent transport or execution
// failure. Ordinary contextual wrappers around ErrUnsupported still fall back.
func unsupportedCompilation(err error) bool {
	switch wrapped := err.(type) {
	case interface{ Unwrap() []error }:
		causes := wrapped.Unwrap()
		if len(causes) == 0 {
			return false
		}

		for _, cause := range causes {
			if !unsupportedCompilation(cause) {
				return false
			}
		}

		return true
	case interface{ Unwrap() error }:
		return unsupportedCompilation(wrapped.Unwrap())
	default:
		return errors.Is(err, errors.ErrUnsupported)
	}
}
