//go:build !go1.18
// +build !go1.18

package zstring

var rngPool sync.Pool

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

// RandUint32 returns pseudorandom uint32
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
