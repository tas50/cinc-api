package cinc

import (
	"context"
	"sync"
)

// maxParallelTransfers bounds how many cookbook file transfers (pre-signed
// bookshelf uploads/downloads) run concurrently. This work is latency-bound and
// embarrassingly parallel, but fan-out is capped so a large cookbook does not
// open an unbounded number of connections.
const maxParallelTransfers = 8

// parallelForEach runs fn for every item using at most maxParallelTransfers
// concurrent workers. It returns the first error encountered, if any, and
// cancels the context handed to in-flight and not-yet-started work so the whole
// operation fails fast. If the parent context is canceled and no work errored,
// the context's error is returned.
func parallelForEach[T any](ctx context.Context, items []T, fn func(context.Context, T) error) error {
	if len(items) == 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sem := make(chan struct{}, maxParallelTransfers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for _, item := range items {
		if ctx.Err() != nil {
			break // an earlier worker failed (or the parent canceled); stop launching
		}
		sem <- struct{}{} // block until a worker slot frees up
		wg.Add(1)
		go func(item T) {
			defer wg.Done()
			defer func() { <-sem }()
			if ctx.Err() != nil {
				return
			}
			if err := fn(ctx, item); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
					cancel()
				}
				mu.Unlock()
			}
		}(item)
	}
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}
	return ctx.Err()
}
