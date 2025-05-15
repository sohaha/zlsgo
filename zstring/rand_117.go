//go:build !go1.18
// +build !go1.18

package zstring

import (
	"sync"
	"time"
)

// rngPool is a pool of random number generators to reduce allocation overhead
var rngPool sync.Pool

// Uint32 generates a pseudorandom uint32 value using a simple xorshift algorithm.
// It initializes the generator state from the current time if needed.
func (r *ru) Uint32() uint32 {
	for r.x == 0 {
		x := time.Now().UnixNano()
		r.x = uint32((x >> 32) ^ x)
	}
	x := r.x
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	r.x = x
	return x
}

// RandUint32 returns a pseudorandom uint32 value.
// It uses a pool of generators to improve performance by reducing allocations.
func RandUint32() uint32 {
	v := rngPool.Get()
	if v == nil {
		v = &ru{}
	}
	r := v.(*ru)
	x := r.Uint32()
	rngPool.Put(r)
	return x
}
