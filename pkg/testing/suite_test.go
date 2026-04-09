package testing_test

import (
	"context"
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
