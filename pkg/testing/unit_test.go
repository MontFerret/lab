package testing_test

import (
	"context"
	"errors"
	stdtesting "testing"
	"time"

	"github.com/MontFerret/ferret/v2"
	labruntime "github.com/MontFerret/lab/v2/pkg/runtime"
	"github.com/MontFerret/lab/v2/pkg/sources"
	testing2 "github.com/MontFerret/lab/v2/pkg/testing"
)

func TestUnitExpectedFailureCompatibility(t *stdtesting.T) {
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
		t.Run(test.name, func(t *stdtesting.T) {
			testCase, err := testing2.New(testing2.Options{
				File: sources.File{
					Name:    test.filename,
					Content: []byte("RETURN 1"),
				},
				Timeout: time.Second,
			})
			if err != nil {
				t.Fatalf("expected no construction error, got %v", err)
			}

			rt := labruntime.AsFunc(func(_ context.Context, _ ferret.Source, _ map[string]any) ([]byte, error) {
				return []byte(`1`), test.runtimeErr
			})

			err = testCase.Run(context.Background(), rt, testing2.NewParams())
			if test.wantErr == "" && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if test.wantErr != "" && (err == nil || err.Error() != test.wantErr) {
				t.Fatalf("expected error %q, got %v", test.wantErr, err)
			}
		})
	}
}
