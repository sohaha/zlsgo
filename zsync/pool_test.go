//go:build go1.18
// +build go1.18

package zsync

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestNewPool(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Run("base", func(tt *zlsgo.TestUtil) {
		pool := NewPool(func() int {
			return 1
		})
		tt.Equal(1, pool.Get())
	})

	tt.Run("pointer", func(tt *zlsgo.TestUtil) {
		type poolS struct{ V int }
		pool := NewPool(func() *poolS { return &poolS{} })
		v := pool.Get()
		tt.EqualTrue(v != nil)
		v.V = 1
		pool.Put(v)
		tt.EqualTrue(pool.Get() != nil)
	})

	tt.Run("slice", func(tt *zlsgo.TestUtil) {
		pool := NewPool(func() []byte { return make([]byte, 0, 8) })
		v := pool.Get()
		tt.EqualTrue(v != nil)
		pool.Put(v[:0])
		tt.EqualTrue(pool.Get() != nil)
	})
}
