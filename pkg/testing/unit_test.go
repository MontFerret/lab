package testing_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MontFerret/api"
	"github.com/MontFerret/lab/v2/internal/testutil/apiruntime"
	"github.com/MontFerret/lab/v2/pkg/sources"
	labtesting "github.com/MontFerret/lab/v2/pkg/testing"
)

func TestUnitExpectedFailureCompatibility(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		runtimeErr error
		wantErr    string
	}{
		{
			name:     "regular unit succeeds",
			filename: "test.fql",
		},
		{
			name:       "legacy expected failure accepts runtime error",
			filename:   "test.fail.fql",
			runtimeErr: errors.New("runtime failed"),
		},
		{
			name:     "legacy expected failure rejects success",
			filename: "test.fail.fql",
			wantErr:  "expected to fail",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testCase, err := labtesting.New(labtesting.Options{
				File: sources.File{
					Name:    test.filename,
					Content: []byte("RETURN 1"),
				},
				Timeout: time.Second,
			})
			if err != nil {
				t.Fatalf("expected no construction error, got %v", err)
			}

			rt := &apiruntime.Runtime{RunFunc: func(_ context.Context, _ api.Source, _ ...api.SessionOption) (api.Output, error) {
				return &apiruntime.Output{Data: []byte(`1`), Err: test.runtimeErr}, nil
			}}

			err = testCase.Run(context.Background(), newOneShotExecutor(t, rt), labtesting.NewParams())
			if test.wantErr == "" && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if test.wantErr != "" && (err == nil || err.Error() != test.wantErr) {
				t.Fatalf("expected error %q, got %v", test.wantErr, err)
			}
		})
	}
}
