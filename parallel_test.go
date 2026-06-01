package cinc

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestParallelForEach_AllRunBounded(t *testing.T) {
	const n = 50
	items := make([]int, n)
	for i := range items {
		items[i] = i
	}
	var running, maxRunning atomic.Int32
	var mu sync.Mutex
	seen := make(map[int]bool)

	err := parallelForEach(context.Background(), items, func(_ context.Context, it int) error {
		cur := running.Add(1)
		for { // record the high-water mark of concurrent workers
			m := maxRunning.Load()
			if cur <= m || maxRunning.CompareAndSwap(m, cur) {
				break
			}
		}
		time.Sleep(time.Millisecond) // encourage overlap
		mu.Lock()
		seen[it] = true
		mu.Unlock()
		running.Add(-1)
		return nil
	})
	if err != nil {
		t.Fatalf("parallelForEach: %v", err)
	}
	if len(seen) != n {
		t.Fatalf("processed %d items, want %d", len(seen), n)
	}
	if got := maxRunning.Load(); got > maxParallelTransfers {
		t.Fatalf("max concurrency %d exceeded cap %d", got, maxParallelTransfers)
	}
	if got := maxRunning.Load(); got < 2 {
		t.Fatalf("expected real parallelism, max concurrent was %d", got)
	}
}

func TestParallelForEach_FirstErrorStopsEarly(t *testing.T) {
	items := make([]int, 100)
	var processed atomic.Int32
	sentinel := errors.New("boom")

	err := parallelForEach(context.Background(), items, func(_ context.Context, _ int) error {
		processed.Add(1)
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want sentinel", err)
	}
	if p := processed.Load(); p == int32(len(items)) {
		t.Errorf("processed all %d items despite an early error; expected early stop", p)
	}
}

func TestParallelForEach_Empty(t *testing.T) {
	err := parallelForEach(context.Background(), []int{}, func(context.Context, int) error {
		t.Fatal("fn should not be called for an empty slice")
		return nil
	})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
}

func TestParallelForEach_PrecanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var processed atomic.Int32

	err := parallelForEach(ctx, []int{1, 2, 3}, func(context.Context, int) error {
		processed.Add(1)
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if p := processed.Load(); p != 0 {
		t.Errorf("processed %d items with a pre-canceled context, want 0", p)
	}
}
