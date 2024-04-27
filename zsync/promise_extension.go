package zsync

import (
	"context"
	"strings"
)

func PromiseAllContext[T any](ctx context.Context, promises ...*Promise[T]) *Promise[[]T] {
	return NewPromiseContext(ctx, func() (res []T, err error) {
		if len(promises) == 0 {
			return
		}

		res = make([]T, len(promises))
		for index := range promises {
			value, err := promises[index].Done()
			if err != nil {
				return nil, err
			}
			res[index] = value
		}

		return
	})
}

func PromiseAll[T any](promises ...*Promise[T]) *Promise[[]T] {
	return PromiseAllContext(context.Background(), promises...)
}

func PromiseRaceContext[T any](ctx context.Context, promises ...*Promise[T]) *Promise[T] {
	return NewPromiseContext(ctx, func() (res T, err error) {
		if len(promises) == 0 {
			return
		}

		valC := make(chan T, len(promises))
		errC := make(chan error, len(promises))
		for index := range promises {
			go func(index int) {
				value, err := promises[index].Done()
				if err != nil {
					errC <- err
					return
				}
				valC <- value
			}(index)
		}

		select {
		case res = <-valC:
		case err = <-errC:
		case <-ctx.Done():
			err = ctx.Err()
		}

		return
	})
}

func PromiseRace[T any](promises ...*Promise[T]) *Promise[T] {
	return PromiseRaceContext(context.Background(), promises...)
}

type AggregateError struct {
	Errors []error
}

func (ae *AggregateError) Error() string {
	errStrings := make([]string, len(ae.Errors))

	for i, err := range ae.Errors {
		errStrings[i] = err.Error()
	}

	return "All promises were rejected: " + strings.Join(errStrings, ", ")
}

func PromiseAnyContext[T any](ctx context.Context, promises ...*Promise[T]) *Promise[T] {
	return NewPromiseContext(ctx, func() (res T, err error) {
		if len(promises) == 0 {
			return
		}

		valC := make(chan T, len(promises))
		errC := make(chan error, len(promises))
		for index := range promises {
			go func(index int) {
				value, err := promises[index].Done()
				if err != nil {
					errC <- err
					return
				}
				valC <- value
			}(index)
		}

		errs := make([]error, 0, len(promises))
	hander:
		select {
		case res = <-valC:
			return
		case e := <-errC:
			errs = append(errs, e)
			if len(errs) == len(promises) {
				err = &AggregateError{
					Errors: errs,
				}
				return
			}
			goto hander
		case <-ctx.Done():
			err = ctx.Err()
			return
		}
	})
}

func PromiseAny[T any](promises ...*Promise[T]) *Promise[T] {
	return PromiseAnyContext(context.Background(), promises...)
}
