package reporters_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/MontFerret/lab/v2/pkg/reporters"
	"github.com/MontFerret/lab/v2/pkg/runner"
)

func TestReportersRenderDeprecationWarningOnce(t *testing.T) {
	const warning = "'.fail.fql' expected-failure tests are deprecated; use a YAML test with 'expect.error' instead"

	tests := []struct {
		name       string
		newReport  func(*bytes.Buffer) reporters.Reporter
		passMarker string
		doneMarker string
	}{
		{
			name:       "console",
			newReport:  func(out *bytes.Buffer) reporters.Reporter { return reporters.NewConsole(out) },
			passMarker: "Passed",
			doneMarker: "Done",
		},
		{
			name:       "simple",
			newReport:  func(out *bytes.Buffer) reporters.Reporter { return reporters.NewSimple(out) },
			passMarker: "PASS",
			doneMarker: "DONE",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			progress := make(chan runner.Result, 1)
			summary := make(chan runner.Summary, 1)
			progress <- runner.Result{
				Times:    2,
				Attempts: 2,
				Filename: "test.fail.fql",
				Warning:  warning,
			}
			close(progress)
			summary <- runner.Summary{Passed: 1}
			close(summary)

			var out bytes.Buffer
			reporter := test.newReport(&out)
			if err := reporter.Report(context.Background(), runner.Stream{Progress: progress, Summary: summary}); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			output := out.String()
			if strings.Count(output, warning) != 1 {
				t.Fatalf("expected warning once, got output %q", output)
			}

			if !strings.Contains(output, test.passMarker) || !strings.Contains(output, test.doneMarker) {
				t.Fatalf("expected passing progress and summary, got output %q", output)
			}
		})
	}
}
