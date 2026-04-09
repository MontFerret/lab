package runner

import (
	"context"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	ferretsource "github.com/MontFerret/ferret/v2/pkg/source"

	labruntime "github.com/MontFerret/lab/v2/pkg/runtime"
	"github.com/MontFerret/lab/v2/pkg/sources"
	testing2 "github.com/MontFerret/lab/v2/pkg/testing"
)

type singleFileSource struct {
	file sources.File
}

func (s singleFileSource) Read(_ context.Context) (<-chan sources.File, <-chan sources.Error) {
	onNext := make(chan sources.File, 1)
	onError := make(chan sources.Error)

	onNext <- s.file
	close(onNext)
	close(onError)

	return onNext, onError
}

func (s singleFileSource) Resolve(_ context.Context, _ *url.URL) (<-chan sources.File, <-chan sources.Error) {
	onNext := make(chan sources.File)
	onError := make(chan sources.Error)

	close(onNext)
	close(onError)

	return onNext, onError
}

func TestRunnerStopsDuringTimesIntervalWhenContextCanceled(t *testing.T) {
	var calls atomic.Int32
	firstCall := make(chan struct{}, 1)

	rt := labruntime.AsFunc(func(_ context.Context, _ *ferretsource.Source, _ map[string]interface{}) ([]byte, error) {
		if calls.Add(1) == 1 {
			firstCall <- struct{}{}
		}

		return []byte(`1`), nil
	})

	r, err := New(Options{
		Runtime:       rt,
		Times:         2,
		TimesInterval: 10,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	stream := r.Run(NewContext(ctx, testing2.NewParams()), singleFileSource{
		file: sources.File{
			Name:    "test.fql",
			Content: []byte("RETURN 1"),
		},
	})

	select {
	case <-firstCall:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first runtime invocation")
	}

	start := time.Now()
	cancel()

	select {
	case res, ok := <-stream.Progress:
		if !ok {
			t.Fatal("progress channel closed before delivering result")
		}

		if res.Error != nil {
			t.Fatalf("expected no error, got %v", res.Error)
		}

		if res.Times != 1 {
			t.Fatalf("expected a single completed run, got %d", res.Times)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for canceled run to finish")
	}

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("expected cancellation to stop interval wait promptly, took %s", elapsed)
	}

	select {
	case sum, ok := <-stream.Summary:
		if !ok {
			t.Fatal("summary channel closed before delivering summary")
		}

		if sum.Passed != 1 || sum.Failed != 0 {
			t.Fatalf("unexpected summary: %+v", sum)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for summary")
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("expected exactly one runtime invocation, got %d", got)
	}
}
