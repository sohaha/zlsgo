package zsync

import (
	"context"
	"sync"
	"time"
)

// MergeContext merges multiple contexts into a single one
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

type Context interface {
	context.Context
}

type mergeContext struct {
	err       error
	doneCh    chan struct{}
	ctxs      []context.Context
	doneIndex int
}

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

func (mc *mergeContext) Done() <-chan struct{} {
	return mc.doneCh
}

func (mc *mergeContext) Err() error {
	return mc.err
}

func (mc *mergeContext) Value(key any) any {
	for _, ctx := range mc.ctxs {
		if v := ctx.Value(key); v != nil {
			return v
		}
	}
	return nil
}

func (mc *mergeContext) monitor() {
	winner := multiselect(mc.ctxs)

	mc.doneIndex = winner
	mc.err = mc.ctxs[winner].Err()

	close(mc.doneCh)
}

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
