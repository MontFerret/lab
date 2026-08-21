package testing_test

import (
	"context"
	"errors"
	"strings"
	stdtesting "testing"
	"time"

	ferretsource "github.com/MontFerret/ferret/v2/pkg/source"

	labruntime "github.com/MontFerret/lab/v2/pkg/runtime"
	"github.com/MontFerret/lab/v2/pkg/sources"
	testing2 "github.com/MontFerret/lab/v2/pkg/testing"
)

func TestSuiteRunUsesAssertParams(t *stdtesting.T) {
	testCase, err := testing2.New(testing2.Options{
		File: sources.File{
			Name: "suite.yaml",
			Content: []byte(`
query:
  text: RETURN 1
  params:
    phase: "query"
assert:
  text: RETURN true
  params:
    phase: "assert"
`),
		},
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var (
		queryPhase  string
		assertPhase string
		dataPhase   string
		callCount   int
	)

	rt := labruntime.AsFunc(func(_ context.Context, _ *ferretsource.Source, params map[string]interface{}) ([]byte, error) {
		callCount++

		switch callCount {
		case 1:
			queryPhase, _ = params["phase"].(string)
			return []byte(`1`), nil
		case 2:
			assertPhase, _ = params["phase"].(string)

			lab, _ := params["lab"].(map[string]interface{})
			data, _ := lab["data"].(map[string]interface{})
			query, _ := data["query"].(map[string]interface{})
			queryParams, _ := query["params"].(map[string]interface{})
			dataPhase, _ = queryParams["phase"].(string)

			return []byte(`true`), nil
		default:
			t.Fatalf("expected exactly two runtime calls, got %d", callCount)
			return nil, nil
		}
	})

	if err := testCase.Run(context.Background(), rt, testing2.NewParams()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if callCount != 2 {
		t.Fatalf("expected 2 runtime calls, got %d", callCount)
	}

	if queryPhase != "query" {
		t.Fatalf("expected query params to use query manifest, got %q", queryPhase)
	}

	if assertPhase != "assert" {
		t.Fatalf("expected assert params to use assert manifest, got %q", assertPhase)
	}

	if dataPhase != "query" {
		t.Fatalf("expected assertion data context to retain query params, got %q", dataPhase)
	}
}

func TestSuiteExpectedError(t *stdtesting.T) {
	tests := []struct {
		name       string
		expect     string
		runtimeErr error
		wantErr    string
	}{
		{
			name:       "any error",
			expect:     "error: {}",
			runtimeErr: errors.New("runtime failed"),
		},
		{
			name:    "any error but query succeeds",
			expect:  "error: {}",
			wantErr: "expected query to fail, but it completed successfully",
		},
		{
			name:       "matching error",
			expect:     "error:\n    contains: expected Array",
			runtimeErr: errors.New("expected Array, got string"),
		},
		{
			name:       "non-matching error",
			expect:     "error:\n    contains: expected Array",
			runtimeErr: errors.New("network unavailable"),
			wantErr:    `expected error containing "expected Array", got: network unavailable`,
		},
		{
			name:    "matching error but query succeeds",
			expect:  "error:\n    contains: expected Array",
			wantErr: "expected query to fail, but it completed successfully",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *stdtesting.T) {
			testCase, err := testing2.New(testing2.Options{
				File: sources.File{
					Name:    "suite.yaml",
					Content: []byte("query:\n  text: RETURN 1\nexpect:\n  " + test.expect + "\n"),
				},
				Timeout: time.Second,
			})
			if err != nil {
				t.Fatalf("expected no construction error, got %v", err)
			}

			calls := 0
			rt := labruntime.AsFunc(func(_ context.Context, _ *ferretsource.Source, _ map[string]interface{}) ([]byte, error) {
				calls++

				return []byte(`1`), test.runtimeErr
			})

			err = testCase.Run(context.Background(), rt, testing2.NewParams())
			if test.wantErr == "" && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if test.wantErr != "" && (err == nil || err.Error() != test.wantErr) {
				t.Fatalf("expected error %q, got %v", test.wantErr, err)
			}

			if calls != 1 {
				t.Fatalf("expected exactly one runtime call, got %d", calls)
			}
		})
	}
}

func TestSuiteExpectedErrorDoesNotAcceptQueryResolutionFailure(t *stdtesting.T) {
	testCase, err := testing2.New(testing2.Options{
		File: sources.File{
			Name: "suite.yaml",
			Content: []byte(`
query:
  ref: "%"
expect:
  error: {}
`),
		},
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("expected no construction error, got %v", err)
	}

	rt := labruntime.AsFunc(func(_ context.Context, _ *ferretsource.Source, _ map[string]interface{}) ([]byte, error) {
		t.Fatal("runtime must not run when query resolution fails")

		return nil, nil
	})

	err = testCase.Run(context.Background(), rt, testing2.NewParams())
	if err == nil || !strings.Contains(err.Error(), "resolve query script") {
		t.Fatalf("expected a query resolution error, got %v", err)
	}
}

func TestSuiteExpectedErrorValidation(t *stdtesting.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name: "assert conflict",
			content: `
query:
  text: RETURN 1
assert:
  text: RETURN true
expect:
  error: {}
`,
			wantErr: "expect.error cannot be combined with assert",
		},
		{
			name: "empty assert conflict",
			content: `
query:
  text: RETURN 1
assert: {}
expect:
  error: {}
`,
			wantErr: "expect.error cannot be combined with assert",
		},
		{
			name: "empty expect",
			content: `
query:
  text: RETURN 1
expect: {}
`,
			wantErr: "assert: ref or text must have value",
		},
		{
			name: "null error",
			content: `
query:
  text: RETURN 1
expect:
  error: null
`,
			wantErr: "assert: ref or text must have value",
		},
		{
			name: "malformed error expectation",
			content: `
query:
  text: RETURN 1
expect:
  error: invalid
`,
			wantErr: "cannot unmarshal",
		},
		{
			name: "unknown error expectation fields",
			content: `
query:
  text: RETURN 1
expect:
  error:
    suffix: unavailable
    prefix: network
`,
			wantErr: `expect.error contains unsupported fields "prefix", "suffix"`,
		},
		{
			name: "misspelled contains field",
			content: `
query:
  text: RETURN 1
expect:
  error:
    contians: expected Array
`,
			wantErr: `expect.error contains unsupported field "contians"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *stdtesting.T) {
			testCase, err := testing2.NewSuite(testing2.Options{
				File: sources.File{
					Name:    "suite.yaml",
					Content: []byte(test.content),
				},
				Timeout: time.Second,
			})
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("expected error containing %q, got %v", test.wantErr, err)
			}

			if testCase != nil {
				t.Fatal("expected construction failure to return no test case")
			}
		})
	}
}

func TestSuiteExpectedErrorValidationIsScoped(t *stdtesting.T) {
	testCase, err := testing2.NewSuite(testing2.Options{
		File: sources.File{
			Name: "suite.yaml",
			Content: []byte(`
unsupported: true
query:
  text: RETURN 1
expect:
  unsupported: true
  error: {}
`),
		},
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("expected unknown fields outside expect.error to remain supported, got %v", err)
	}

	if testCase == nil {
		t.Fatal("expected a constructed suite")
	}
}
