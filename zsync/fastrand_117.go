//go:build !go1.18
// +build !go1.18

package zsync

import (
	"sync"
	"time"
)

type RNG struct {
	x uint32
}

var rngPool sync.Pool

func fastrand() uint32 {
	v := rngPool.Get()
	if v == nil {
		v = &RNG{}
	}
	r := v.(*RNG)
	x := r.Uint32()
	rngPool.Put(r)
	return x
}

func (r *RNG) Uint32() uint32 {
	for r.x == 0 {
		r.x = getRandom()
	}
	x := r.x
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	r.x = x
	return x
}

func getRandom() uint32 {
	x := time.Now().UnixNano()
	return uint32((x >> 32) ^ x)
}
