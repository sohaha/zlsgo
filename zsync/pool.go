//go:build go1.18
// +build go1.18

package zsync

import "sync"

type Pool[T any] struct {
	p sync.Pool
}

func NewPool[T any](n func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any {
				return n()
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

func (p *Pool[T]) Put(x T) {
	p.p.Put(x)
}
