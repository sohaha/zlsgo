package zsync

import (
	"context"
	"sync"
	"time"
)

// MergeContext merges multiple contexts into a single one.
// The resulting context is canceled when any of the input contexts is canceled,
// and its deadline is the earliest deadline of all input contexts.
// Values from all contexts are accessible, with values from earlier contexts
// in the list taking precedence over later ones when keys conflict.
func MergeContext(ctxs ...context.Context) Context {
	if len(ctxs) == 0 {
		return context.Background()
	}

	mc := mergeContext{
		ctxs:      ctxs,
		doneCh:    make(chan struct{}),
		doneIndex: -1,
	}
	go mc.monitor()

	return &mc
}

// Context is an interface that extends the standard context.Context interface.
// It provides all the functionality of the standard context with potential
// additional methods specific to the zsync package.
type Context interface {
	context.Context
}

// mergeContext is an implementation of Context that merges multiple contexts.
// It tracks which context was canceled first and propagates values from all contexts.
type mergeContext struct {
	err       error            // The error from the first canceled context
	doneCh    chan struct{}    // Channel that is closed when any context is canceled
	ctxs      []context.Context // The merged contexts
	doneIndex int              // Index of the first context that was canceled
}

// Deadline returns the earliest deadline of all merged contexts.
// If none of the merged contexts has a deadline, it returns a zero time and false.
func (mc *mergeContext) Deadline() (time.Time, bool) {
	dl := time.Time{}
	for _, ctx := range mc.ctxs {
		thisDL, ok := ctx.Deadline()
		if ok {
			if dl.IsZero() {
				dl = thisDL
			} else if thisDL.Before(dl) {
				dl = thisDL
			}
		}
	}
	return dl, !dl.IsZero()
}

// Done returns a channel that is closed when any of the merged contexts is done.
func (mc *mergeContext) Done() <-chan struct{} {
	return mc.doneCh
}

// Err returns the error from the first context that was canceled,
// or nil if no context has been canceled yet.
func (mc *mergeContext) Err() error {
	return mc.err
}

// Value returns the value associated with the key in any of the merged contexts.
// It checks contexts in the order they were provided to MergeContext,
// returning the first non-nil value found.
func (mc *mergeContext) Value(key any) any {
	for _, ctx := range mc.ctxs {
		if v := ctx.Value(key); v != nil {
			return v
		}
	}
	return nil
}

// monitor is an internal method that watches all contexts and closes
// the done channel when any context is canceled.
func (mc *mergeContext) monitor() {
	winner := multiselect(mc.ctxs)

	mc.doneIndex = winner
	mc.err = mc.ctxs[winner].Err()

	close(mc.doneCh)
}

// multiselect waits for any of the given contexts to be done and returns
// the index of the first context that was canceled.
// It returns -1 if no context was canceled (which should not happen in practice).
func multiselect(ctxs []context.Context) int {
	res := make(chan int)

	count := len(ctxs)
	if count == 1 {
		<-ctxs[0].Done()
		return 0
	}

	var wg sync.WaitGroup
	wg.Add(count)

	for i, ctx := range ctxs {
		go func(i int, ctx context.Context) {
			defer wg.Done()
			<-ctx.Done()
			if ctx.Err() != nil {
			}
			res <- i
		}(i, ctx)
	}

	go func() {
		wg.Wait()
		close(res)
	}()

	return <-res
}
