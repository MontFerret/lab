package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGoContextRunsFunction(t *testing.T) {
	pool := NewPool(2)
	var called atomic.Bool

	err := pool.GoContext(context.Background(), func() {
		called.Store(true)
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if !called.Load() {
		t.Fatal("expected function to be called")
	}
}

func TestGoContextReturnsErrorWhenContextCanceled(t *testing.T) {
	pool := NewPool(1)

	// Fill the pool slot.
	started := make(chan struct{})
	hold := make(chan struct{})

	err := pool.GoContext(context.Background(), func() {
		close(started)
		<-hold
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	<-started

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = pool.GoContext(ctx, func() {
		t.Fatal("function should not run when context is canceled")
	})
	if err == nil {
		t.Fatal("expected error when context is canceled")
	}
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	close(hold)
}

func TestGoContextReleasesSlotAfterFunctionCompletes(t *testing.T) {
	pool := NewPool(1)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := pool.GoContext(context.Background(), func() {
				time.Sleep(10 * time.Millisecond)
			})
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for all goroutines to complete")
	}
}

func TestGoContextRespectsPoolCapacity(t *testing.T) {
	pool := NewPool(2)
	var running atomic.Int32
	var maxRunning atomic.Int32

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := pool.GoContext(context.Background(), func() {
				cur := running.Add(1)
				for {
					old := maxRunning.Load()
					if cur <= old || maxRunning.CompareAndSwap(old, cur) {
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
				running.Add(-1)
			})
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		}()
	}

	wg.Wait()

	if got := maxRunning.Load(); got > 2 {
		t.Fatalf("expected at most 2 concurrent goroutines, saw %d", got)
	}
}
