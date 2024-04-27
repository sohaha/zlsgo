package zsync

import (
	"context"

	"github.com/sohaha/zlsgo/zutil"
)

type PromiseState uint8

type Promise[T any] struct {
	value   T
	done    chan struct{}
	reason  error
	timeout *zutil.Bool
	ctx     context.Context
}

func (p *Promise[T]) Catch(rejected func(error) (T, error)) *Promise[T] {
	return p.Then(nil, rejected)
}

func (p *Promise[T]) Finally(finally func()) *Promise[T] {
	_, _ = p.Done()
	finally()
	return p
}

func (p *Promise[T]) Then(fulfilled func(T) (T, error), rejected ...func(error) (T, error)) *Promise[T] {
	return NewPromiseContext[T](p.ctx, func() (T, error) {
		value, err := p.Done()
		if err == nil {
			return fulfilled(value)
		}

		if len(rejected) > 0 && rejected[0] != nil {
			return rejected[0](err)
		}

		return value, err
	})
}

func (p *Promise[T]) Done() (value T, reason error) {
	if p.timeout.Load() {
		return value, context.Canceled
	}

	select {
	case <-p.done:
	case <-p.ctx.Done():
		p.timeout.Store(true)
		return value, context.Canceled
	}

	return p.value, p.reason
}

func NewPromiseContext[T any](ctx context.Context, executor func() (T, error)) *Promise[T] {
	if executor == nil {
		return nil
	}

	p := &Promise[T]{done: make(chan struct{}, 1), ctx: ctx, timeout: zutil.NewBool(false)}

	go func() {
		value, err := executor()
		if err != nil {
			p.reason = err
		} else {
			p.value = value
		}
		close(p.done)
	}()

	return p
}

func NewPromise[T any](executor func() (T, error)) *Promise[T] {
	return NewPromiseContext(context.Background(), executor)
}
